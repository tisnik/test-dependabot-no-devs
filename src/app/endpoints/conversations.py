"""Handler for REST API calls to manage conversation history."""

import logging
from typing import Any

from fastapi import APIRouter, Depends, HTTPException, Request, status
from llama_stack_client import APIConnectionError, NotFoundError

from app.database import get_session
from authentication import get_auth_dependency
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.config import Action
from models.database.conversations import UserConversation
from models.responses import (
    ConversationDeleteResponse,
    ConversationDetails,
    ConversationResponse,
    ConversationsListResponse,
    UnauthorizedResponse,
    NotFoundResponse,
    AccessDeniedResponse,
    BadRequestResponse,
    ServiceUnavailableResponse,
)
from utils.endpoints import (
    check_configuration_loaded,
    delete_conversation,
    can_access_conversation,
    retrieve_conversation,
)
from utils.suid import check_suid

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["conversations"])

conversation_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "model": ConversationResponse,
        "description": "Conversation retrieved successfully",
    },
    400: {
        "model": BadRequestResponse,
        "description": "Invalid request",
    },
    401: {
        "model": UnauthorizedResponse,
        "description": "Unauthorized: Invalid or missing Bearer token",
    },
    403: {
        "model": AccessDeniedResponse,
        "description": "Client does not have permission to access conversation",
    },
    404: {
        "model": NotFoundResponse,
        "description": "Conversation not found",
    },
    503: {
        "model": ServiceUnavailableResponse,
        "description": "Service unavailable",
    },
}

conversation_delete_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "model": ConversationDeleteResponse,
        "description": "Conversation deleted successfully",
    },
    400: {
        "model": BadRequestResponse,
        "description": "Invalid request",
    },
    401: {
        "model": UnauthorizedResponse,
        "description": "Unauthorized: Invalid or missing Bearer token",
    },
    403: {
        "model": AccessDeniedResponse,
        "description": "Client does not have permission to access conversation",
    },
    404: {
        "model": NotFoundResponse,
        "description": "Conversation not found",
    },
    503: {
        "model": ServiceUnavailableResponse,
        "description": "Service unavailable",
    },
}

conversations_list_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "model": ConversationsListResponse,
        "description": "List of conversations retrieved successfully",
    },
    401: {
        "model": UnauthorizedResponse,
        "description": "Unauthorized: Invalid or missing Bearer token",
    },
    503: {
        "model": ServiceUnavailableResponse,
        "description": "Service unavailable",
    },
}


def simplify_session_data(session_data: dict) -> list[dict[str, Any]]:
    """
    Convert a Llama Stack session dict into a simplified chat-history list.
    
    Each returned turn contains:
    - "messages": a list of message objects with "content" and "type" (the original role).
    - "started_at" and "completed_at" timestamps copied from the session turn.
    
    Parameters:
        session_data (dict): The full session dictionary returned by Llama Stack.
    
    Returns:
        list[dict[str, Any]]: A list of simplified turns. Each turn is a dict with keys
        "messages" (list of {"content": Any, "type": Any}), "started_at", and "completed_at".
    """
    # Create simplified structure
    chat_history = []

    # Extract only essential data from each turn
    for turn in session_data.get("turns", []):
        # Clean up input messages
        cleaned_messages = []
        for msg in turn.get("input_messages", []):
            cleaned_msg = {
                "content": msg.get("content"),
                "type": msg.get("role"),  # Rename role to type
            }
            cleaned_messages.append(cleaned_msg)

        # Clean up output message
        output_msg = turn.get("output_message", {})
        cleaned_messages.append(
            {
                "content": output_msg.get("content"),
                "type": output_msg.get("role"),  # Rename role to type
            }
        )

        simplified_turn = {
            "messages": cleaned_messages,
            "started_at": turn.get("started_at"),
            "completed_at": turn.get("completed_at"),
        }
        chat_history.append(simplified_turn)

    return chat_history


@router.get("/conversations", responses=conversations_list_responses)
@authorize(Action.LIST_CONVERSATIONS)
async def get_conversations_list_endpoint_handler(
    request: Request,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationsListResponse:
    """
    Retrieve a list of conversation summaries visible to the authenticated user.
    
    If the caller has the Action.LIST_OTHERS_CONVERSATIONS permission, include conversations from all users; otherwise include only conversations owned by the authenticated user.
    
    Returns:
        ConversationsListResponse: Contains a list of ConversationDetails with metadata (conversation_id, created_at, last_message_at, message_count, last_used_model, last_used_provider, topic_summary).
    """
    check_configuration_loaded(configuration)

    user_id = auth[0]

    logger.info("Retrieving conversations for user %s", user_id)

    with get_session() as session:
        try:
            query = session.query(UserConversation)

            filtered_query = (
                query
                if Action.LIST_OTHERS_CONVERSATIONS in request.state.authorized_actions
                else query.filter_by(user_id=user_id)
            )

            user_conversations = filtered_query.all()

            # Return conversation summaries with metadata
            conversations = [
                ConversationDetails(
                    conversation_id=conv.id,
                    created_at=conv.created_at.isoformat() if conv.created_at else None,
                    last_message_at=(
                        conv.last_message_at.isoformat()
                        if conv.last_message_at
                        else None
                    ),
                    message_count=conv.message_count,
                    last_used_model=conv.last_used_model,
                    last_used_provider=conv.last_used_provider,
                    topic_summary=conv.topic_summary,
                )
                for conv in user_conversations
            ]

            logger.info(
                "Found %d conversations for user %s", len(conversations), user_id
            )

            return ConversationsListResponse(conversations=conversations)

        except Exception as e:
            logger.exception(
                "Error retrieving conversations for user %s: %s", user_id, e
            )
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail={
                    "response": "Unknown error",
                    "cause": f"Unknown error while getting conversations for user {user_id}",
                },
            ) from e


@router.get("/conversations/{conversation_id}", responses=conversation_responses)
@authorize(Action.GET_CONVERSATION)
async def get_conversation_endpoint_handler(
    request: Request,
    conversation_id: str,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationResponse:
    """
    Retrieve a conversation's simplified chat history by conversation ID.
    
    Validates the ID and caller access, fetches the session from the Llama Stack backend,
    simplifies the session into chat turns, and returns a ConversationResponse containing
    the conversation ID and the simplified chat history.
    
    Returns:
        ConversationResponse: Response containing `conversation_id` and `chat_history` (list of simplified turns).
    
    Raises:
        HTTPException: on invalid conversation ID (400), access denied (403), conversation not found (404),
                       Llama Stack unavailable (503), or other internal errors (500).
    """
    check_configuration_loaded(configuration)

    # Validate conversation ID format
    if not check_suid(conversation_id):
        logger.error("Invalid conversation ID format: %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=BadRequestResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        )

    user_id = auth[0]
    if not can_access_conversation(
        conversation_id,
        user_id,
        others_allowed=(
            Action.READ_OTHERS_CONVERSATIONS in request.state.authorized_actions
        ),
    ):
        logger.warning(
            "User %s attempted to read conversation %s they don't have access to",
            user_id,
            conversation_id,
        )
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=AccessDeniedResponse(
                user_id=user_id,
                resource="conversation",
                resource_id=conversation_id,
                action="read",
            ).dump_detail(),
        )

    # If reached this, user is authorized to retreive this conversation
    conversation = retrieve_conversation(conversation_id)
    if conversation is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=NotFoundResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        )

    agent_id = conversation_id
    logger.info("Retrieving conversation %s", conversation_id)

    try:
        client = AsyncLlamaStackClientHolder().get_client()

        agent_sessions = (await client.agents.session.list(agent_id=agent_id)).data
        if not agent_sessions:
            logger.error("No sessions found for conversation %s", conversation_id)
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=NotFoundResponse(
                    resource="conversation", resource_id=conversation_id
                ).dump_detail(),
            )
        session_id = str(agent_sessions[0].get("session_id"))

        session_response = await client.agents.session.retrieve(
            agent_id=agent_id, session_id=session_id
        )
        session_data = session_response.model_dump()

        logger.info("Successfully retrieved conversation %s", conversation_id)

        # Simplify the session data to include only essential conversation information
        chat_history = simplify_session_data(session_data)

        return ConversationResponse(
            conversation_id=conversation_id,
            chat_history=chat_history,
        )

    except APIConnectionError as e:
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=ServiceUnavailableResponse(
                backend_name="Llama Stack", cause=str(e)
            ).dump_detail(),
        ) from e

    except NotFoundError as e:
        logger.error("Conversation not found: %s", e)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=NotFoundResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        ) from e

    except HTTPException:
        raise

    except Exception as e:
        # Handle case where session doesn't exist or other errors
        logger.exception("Error retrieving conversation %s: %s", conversation_id, e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unknown error",
                "cause": f"Unknown error while getting conversation {conversation_id} : {str(e)}",
            },
        ) from e


@router.delete(
    "/conversations/{conversation_id}", responses=conversation_delete_responses
)
@authorize(Action.DELETE_CONVERSATION)
async def delete_conversation_endpoint_handler(
    request: Request,
    conversation_id: str,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationDeleteResponse:
    """
    Delete a conversation and its corresponding Llama Stack session.
    
    Validates the conversation ID format and enforces access control; if authorized, attempts to delete the remote session (if any) and removes the local conversation record. Missing remote sessions are treated as a successful deletion.
    
    Parameters:
        request (Request): Incoming FastAPI request (used for auth state).
        conversation_id (str): Unique identifier of the conversation to delete.
        auth: Authentication dependency (omitted from detailed doc).
    
    Returns:
        ConversationDeleteResponse: Result containing `conversation_id`, `success` (True on successful deletion), and a human-readable `response` message.
    
    Raises:
        HTTPException: For invalid ID (400), forbidden access (403), not found (404), backend connection issues (503), or other internal errors (500).
    """
    check_configuration_loaded(configuration)

    # Validate conversation ID format
    if not check_suid(conversation_id):
        logger.error("Invalid conversation ID format: %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=BadRequestResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        )

    user_id = auth[0]
    if not can_access_conversation(
        conversation_id,
        user_id,
        others_allowed=(
            Action.DELETE_OTHERS_CONVERSATIONS in request.state.authorized_actions
        ),
    ):
        logger.warning(
            "User %s attempted to delete conversation %s they don't have access to",
            user_id,
            conversation_id,
        )
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=AccessDeniedResponse(
                user_id=user_id,
                resource="conversation",
                resource_id=conversation_id,
                action="delete",
            ).dump_detail(),
        )

    # If reached this, user is authorized to retreive this conversation
    conversation = retrieve_conversation(conversation_id)
    if conversation is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=NotFoundResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        )

    agent_id = conversation_id
    logger.info("Deleting conversation %s", conversation_id)

    try:
        # Get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()

        agent_sessions = (await client.agents.session.list(agent_id=agent_id)).data

        if not agent_sessions:
            # If no sessions are found, do not raise an error, just return a success response
            logger.info("No sessions found for conversation %s", conversation_id)
            return ConversationDeleteResponse(
                conversation_id=conversation_id,
                success=True,
                response="Conversation deleted successfully",
            )

        session_id = str(agent_sessions[0].get("session_id"))

        await client.agents.session.delete(agent_id=agent_id, session_id=session_id)

        logger.info("Successfully deleted conversation %s", conversation_id)

        delete_conversation(conversation_id=conversation_id)

        return ConversationDeleteResponse(
            conversation_id=conversation_id,
            success=True,
            response="Conversation deleted successfully",
        )

    except APIConnectionError as e:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=ServiceUnavailableResponse(
                backend_name="Llama Stack", cause=str(e)
            ).dump_detail(),
        ) from e

    except NotFoundError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=NotFoundResponse(
                resource="conversation", resource_id=conversation_id
            ).dump_detail(),
        ) from e

    except HTTPException:
        raise

    except Exception as e:
        # Handle case where session doesn't exist or other errors
        logger.exception("Error deleting conversation %s: %s", conversation_id, e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unknown error",
                "cause": f"Unknown error while deleting conversation {conversation_id} : {str(e)}",
            },
        ) from e