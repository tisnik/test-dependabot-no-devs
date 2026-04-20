"""Handler for REST API calls to manage conversation history."""

import logging
from typing import Any

from llama_stack_client import APIConnectionError, NotFoundError

from fastapi import APIRouter, HTTPException, status, Depends

from client import LlamaStackClientHolder
from configuration import configuration
from models.responses import ConversationResponse, ConversationDeleteResponse
from auth import get_auth_dependency
from utils.endpoints import check_configuration_loaded
from utils.suid import check_suid

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["conversations"])
auth_dependency = get_auth_dependency()

conversation_id_to_agent_id: dict[str, str] = {}

conversation_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "session_data": {
            "session_id": "123e4567-e89b-12d3-a456-426614174000",
            "turns": [],
            "started_at": "2024-01-01T00:00:00Z",
        },
    },
    404: {
        "detail": {
            "response": "Conversation not found",
            "cause": "The specified conversation ID does not exist.",
        }
    },
    503: {
        "detail": {
            "response": "Unable to connect to Llama Stack",
            "cause": "Connection error.",
        }
    },
}

conversation_delete_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "success": True,
        "message": "Conversation deleted successfully",
    },
    404: {
        "detail": {
            "response": "Conversation not found",
            "cause": "The specified conversation ID does not exist.",
        }
    },
    503: {
        "detail": {
            "response": "Unable to connect to Llama Stack",
            "cause": "Connection error.",
        }
    },
}


def simplify_session_data(session_data: Any) -> list[dict[str, Any]]:
    """
    Convert full session data from llama-stack into a simplified list of conversation turns.
    
    Each turn in the returned list contains only the essential input and output messages (with roles renamed to "type") and associated timestamps.
    Returns a list of dictionaries, each representing a conversation turn with messages, `started_at`, and `completed_at` fields.
    """
    session_dict = session_data.model_dump()
    # Create simplified structure
    chat_history = []

    # Extract only essential data from each turn
    for turn in session_dict.get("turns", []):
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


@router.get("/conversations/{conversation_id}", responses=conversation_responses)
def get_conversation_endpoint_handler(
    conversation_id: str,
    _auth: Any = Depends(auth_dependency),
) -> ConversationResponse:
    """
    Retrieves a conversation's chat history by its ID.
    
    Validates the conversation ID, locates the associated agent, and fetches the session data from the Llama Stack service. Returns the conversation's simplified chat history in a structured response. Raises HTTP errors for invalid IDs, missing conversations, service unavailability, or unexpected failures.
    
    Returns:
        ConversationResponse: Contains the conversation ID and its chat history.
    """
    check_configuration_loaded(configuration)

    # Validate conversation ID format
    if not check_suid(conversation_id):
        logger.error("Invalid conversation ID format: %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "response": "Invalid conversation ID format",
                "cause": f"Conversation ID {conversation_id} is not a valid UUID",
            },
        )

    agent_id = conversation_id_to_agent_id.get(conversation_id)
    if not agent_id:
        logger.error("Agent ID not found for conversation %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "conversation ID not found",
                "cause": f"conversation ID {conversation_id} not found!",
            },
        )

    logger.info("Retrieving conversation %s", conversation_id)

    try:
        client = LlamaStackClientHolder().get_client()

        session_data = client.agents.session.retrieve(
            agent_id=agent_id, session_id=conversation_id
        )

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
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e
    except NotFoundError as e:
        logger.error("Conversation not found: %s", e)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation not found",
                "cause": f"Conversation {conversation_id} could not be retrieved: {str(e)}",
            },
        ) from e
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
def delete_conversation_endpoint_handler(
    conversation_id: str,
    _auth: Any = Depends(auth_dependency),
) -> ConversationDeleteResponse:
    """
    Deletes a conversation session by its ID.
    
    Validates the conversation ID format and checks for its existence. If valid, deletes the corresponding session using the Llama Stack client and returns a response indicating success. Raises HTTP errors for invalid IDs, missing conversations, service unavailability, or unexpected failures.
    
    Returns:
        ConversationDeleteResponse: Confirmation of successful deletion with the conversation ID.
    """
    check_configuration_loaded(configuration)

    # Validate conversation ID format
    if not check_suid(conversation_id):
        logger.error("Invalid conversation ID format: %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "response": "Invalid conversation ID format",
                "cause": f"Conversation ID {conversation_id} is not a valid UUID",
            },
        )
    agent_id = conversation_id_to_agent_id.get(conversation_id)
    if not agent_id:
        logger.error("Agent ID not found for conversation %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "conversation ID not found",
                "cause": f"conversation ID {conversation_id} not found!",
            },
        )
    logger.info("Deleting conversation %s", conversation_id)

    try:
        # Get Llama Stack client
        client = LlamaStackClientHolder().get_client()
        # Delete session using the conversation_id as session_id
        # In this implementation, conversation_id and session_id are the same
        client.agents.session.delete(agent_id=agent_id, session_id=conversation_id)

        logger.info("Successfully deleted conversation %s", conversation_id)

        return ConversationDeleteResponse(
            conversation_id=conversation_id,
            success=True,
            response="Conversation deleted successfully",
        )

    except APIConnectionError as e:
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e
    except NotFoundError as e:
        logger.error("Conversation not found: %s", e)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation not found",
                "cause": f"Conversation {conversation_id} could not be deleted: {str(e)}",
            },
        ) from e
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
