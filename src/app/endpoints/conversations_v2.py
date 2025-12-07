"""Handler for REST API calls to manage conversation history."""

import logging
from typing import Any

from fastapi import APIRouter, Depends, HTTPException, Request, status

from authentication import get_auth_dependency
from authorization.middleware import authorize
from configuration import configuration
from models.cache_entry import CacheEntry
from models.config import Action
from models.requests import ConversationUpdateRequest
from models.responses import (
    ConversationDeleteResponse,
    ConversationResponse,
    ConversationUpdateResponse,
    ConversationsListResponseV2,
    UnauthorizedResponse,
)
from utils.endpoints import check_configuration_loaded
from utils.suid import check_suid

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["conversations_v2"])


conversation_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "chat_history": [
            {
                "messages": [
                    {"content": "Hi", "type": "user"},
                    {"content": "Hello!", "type": "assistant"},
                ],
                "started_at": "2024-01-01T00:00:00Z",
                "completed_at": "2024-01-01T00:00:05Z",
                "provider": "provider ID",
                "model": "model ID",
            }
        ],
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    401: {
        "description": "Unauthorized: Invalid or missing Bearer token",
        "model": UnauthorizedResponse,
    },
    404: {
        "detail": {
            "response": "Conversation not found",
            "cause": "The specified conversation ID does not exist.",
        }
    },
}

conversation_delete_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "success": True,
        "message": "Conversation deleted successfully",
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    401: {
        "description": "Unauthorized: Invalid or missing Bearer token",
        "model": UnauthorizedResponse,
    },
    404: {
        "detail": {
            "response": "Conversation not found",
            "cause": "The specified conversation ID does not exist.",
        }
    },
}

conversations_list_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversations": [
            {
                "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
                "topic_summary": "This is a topic summary",
                "last_message_timestamp": "2024-01-01T00:00:00Z",
            }
        ]
    }
}

conversation_update_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "success": True,
        "message": "Topic summary updated successfully",
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    401: {
        "description": "Unauthorized: Invalid or missing Bearer token",
        "model": UnauthorizedResponse,
    },
    404: {
        "detail": {
            "response": "Conversation not found",
            "cause": "The specified conversation ID does not exist.",
        }
    },
}


@router.get("/conversations", responses=conversations_list_responses)
@authorize(Action.LIST_CONVERSATIONS)
async def get_conversations_list_endpoint_handler(
    request: Request,  # pylint: disable=unused-argument
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationsListResponseV2:
    """
    Retrieve all conversations belonging to the authenticated user.
    
    Returns:
        ConversationsListResponseV2: Response containing the list of conversations for the authenticated user.
    
    Raises:
        HTTPException: 404 if the conversation cache is not configured.
    """
    check_configuration_loaded(configuration)

    user_id = auth[0]

    logger.info("Retrieving conversations for user %s", user_id)

    skip_userid_check = auth[2]

    if configuration.conversation_cache is None:
        logger.warning("Converastion cache is not configured")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation cache is not configured",
                "cause": "Conversation cache is not configured",
            },
        )

    conversations = configuration.conversation_cache.list(user_id, skip_userid_check)
    logger.info("Conversations for user %s: %s", user_id, len(conversations))

    return ConversationsListResponseV2(conversations=conversations)


@router.get("/conversations/{conversation_id}", responses=conversation_responses)
@authorize(Action.GET_CONVERSATION)
async def get_conversation_endpoint_handler(
    request: Request,  # pylint: disable=unused-argument
    conversation_id: str,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationResponse:
    """
    Retrieve a conversation and its transformed chat history for the authenticated user.
    
    Parameters:
        conversation_id (str): Conversation identifier; validated for correct format.
    
    Returns:
        ConversationResponse: Response object containing `conversation_id` and `chat_history` where each entry is a transformed cache message.
    
    Raises:
        HTTPException: 400 if `conversation_id` is invalid.
        HTTPException: 404 if the conversation cache is not configured or the conversation is not found for the user.
    """
    check_configuration_loaded(configuration)
    check_valid_conversation_id(conversation_id)

    user_id = auth[0]
    logger.info("Retrieving conversation %s for user %s", conversation_id, user_id)

    skip_userid_check = auth[2]

    if configuration.conversation_cache is None:
        logger.warning("Converastion cache is not configured")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation cache is not configured",
                "cause": "Conversation cache is not configured",
            },
        )

    check_conversation_existence(user_id, conversation_id)

    conversation = configuration.conversation_cache.get(
        user_id, conversation_id, skip_userid_check
    )
    chat_history = [transform_chat_message(entry) for entry in conversation]

    return ConversationResponse(
        conversation_id=conversation_id, chat_history=chat_history
    )


@router.delete(
    "/conversations/{conversation_id}", responses=conversation_delete_responses
)
@authorize(Action.DELETE_CONVERSATION)
async def delete_conversation_endpoint_handler(
    request: Request,  # pylint: disable=unused-argument
    conversation_id: str,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationDeleteResponse:
    """
    Delete a conversation identified by `conversation_id` for the authenticated user.
    
    Validates configuration and conversation ID, ensures the conversation exists, and attempts to remove it from the configured conversation cache. Returns a ConversationDeleteResponse indicating whether the deletion was performed and includes a human-readable message.
    
    Raises:
        HTTPException: 400 if `conversation_id` is invalid.
        HTTPException: 404 if the conversation cache is not configured or the conversation does not exist.
    
    Returns:
        ConversationDeleteResponse: Object containing `conversation_id`, `success`, and `response` message.
    """
    check_configuration_loaded(configuration)
    check_valid_conversation_id(conversation_id)

    user_id = auth[0]
    logger.info("Deleting conversation %s for user %s", conversation_id, user_id)

    skip_userid_check = auth[2]

    if configuration.conversation_cache is None:
        logger.warning("Converastion cache is not configured")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation cache is not configured",
                "cause": "Conversation cache is not configured",
            },
        )

    check_conversation_existence(user_id, conversation_id)

    logger.info("Deleting conversation %s for user %s", conversation_id, user_id)
    deleted = configuration.conversation_cache.delete(
        user_id, conversation_id, skip_userid_check
    )

    if deleted:
        return ConversationDeleteResponse(
            conversation_id=conversation_id,
            success=True,
            response="Conversation deleted successfully",
        )
    return ConversationDeleteResponse(
        conversation_id=conversation_id,
        success=True,
        response="Conversation can not be deleted",
    )


@router.put("/conversations/{conversation_id}", responses=conversation_update_responses)
@authorize(Action.UPDATE_CONVERSATION)
async def update_conversation_endpoint_handler(
    conversation_id: str,
    update_request: ConversationUpdateRequest,
    auth: Any = Depends(get_auth_dependency()),
) -> ConversationUpdateResponse:
    """
    Update the stored topic summary for a conversation.
    
    Updates the conversation's topic summary in the configured conversation cache for the authenticated user and returns a success response on completion.
    
    Parameters:
        conversation_id (str): Identifier of the conversation to update.
        update_request (ConversationUpdateRequest): Request object containing the new `topic_summary`.
    
    Returns:
        ConversationUpdateResponse: Response indicating the conversation id, success status, and a descriptive message.
    
    Raises:
        HTTPException: 400 if `conversation_id` is invalid.
        HTTPException: 404 if the conversation cache is not configured or the conversation is not found.
    """
    check_configuration_loaded(configuration)
    check_valid_conversation_id(conversation_id)

    user_id = auth[0]
    logger.info(
        "Updating topic summary for conversation %s for user %s",
        conversation_id,
        user_id,
    )

    skip_userid_check = auth[2]

    if configuration.conversation_cache is None:
        logger.warning("Conversation cache is not configured")
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation cache is not configured",
                "cause": "Conversation cache is not configured",
            },
        )

    check_conversation_existence(user_id, conversation_id)

    # Update the topic summary in the cache
    configuration.conversation_cache.set_topic_summary(
        user_id, conversation_id, update_request.topic_summary, skip_userid_check
    )

    logger.info(
        "Successfully updated topic summary for conversation %s for user %s",
        conversation_id,
        user_id,
    )

    return ConversationUpdateResponse(
        conversation_id=conversation_id,
        success=True,
        message="Topic summary updated successfully",
    )


def check_valid_conversation_id(conversation_id: str) -> None:
    """
    Validate that `conversation_id` is a valid SUID/UUID-like identifier.
    
    Parameters:
        conversation_id (str): The conversation identifier to validate.
    
    Raises:
        HTTPException: With status 400 when `conversation_id` is not a valid SUID/UUID-like ID.
    """
    if not check_suid(conversation_id):
        logger.error("Invalid conversation ID format: %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "response": "Invalid conversation ID format",
                "cause": f"Conversation ID {conversation_id} is not a valid UUID",
            },
        )


def check_conversation_existence(user_id: str, conversation_id: str) -> None:
    """
    Verify that a conversation with the given ID exists for the specified user.
    
    If a conversation cache is configured, raises an HTTPException with status 404 when the conversation ID is not present for the user; otherwise does nothing.
    
    Parameters:
        user_id (str): ID of the user who owns the conversation.
        conversation_id (str): Conversation identifier to check.
    
    Raises:
        HTTPException: Raised with status 404 and a descriptive detail when the conversation is not found.
    """
    # checked already, but we need to make pyright happy
    if configuration.conversation_cache is None:
        return
    conversations = configuration.conversation_cache.list(user_id, False)
    conversation_ids = [conv.conversation_id for conv in conversations]
    if conversation_id not in conversation_ids:
        logger.error("No conversation found for conversation ID %s", conversation_id)
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail={
                "response": "Conversation not found",
                "cause": f"Conversation {conversation_id} could not be retrieved.",
            },
        )


def transform_chat_message(entry: CacheEntry) -> dict[str, Any]:
    """
    Convert a cached CacheEntry into the response-ready payload for API responses.
    
    The returned dictionary contains keys: `provider`, `model`, `messages`, `started_at`, and `completed_at`. `messages` is a list with a user message (`type`: "user", `content`: the entry query) followed by an assistant message (`type`: "assistant", `content`: the entry response). If the entry includes `referenced_documents`, they are added to the assistant message as a list of serialized document objects.
    
    Returns:
        response_payload (dict): Payload dictionary suitable for the ConversationResponse body:
            - provider: provider identifier from the cache entry
            - model: model identifier from the cache entry
            - messages: list[dict] with the user and assistant message objects (assistant may include `referenced_documents`)
            - started_at: timestamp when the interaction started
            - completed_at: timestamp when the interaction completed
    """
    user_message = {"content": entry.query, "type": "user"}
    assistant_message: dict[str, Any] = {"content": entry.response, "type": "assistant"}

    # If referenced_documents exist on the entry, add them to the assistant message
    if entry.referenced_documents is not None:
        assistant_message["referenced_documents"] = [
            doc.model_dump(mode="json") for doc in entry.referenced_documents
        ]

    return {
        "provider": entry.provider,
        "model": entry.model,
        "messages": [user_message, assistant_message],
        "started_at": entry.started_at,
        "completed_at": entry.completed_at,
    }