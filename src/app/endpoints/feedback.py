"""Handler for REST API endpoint for user feedback."""

import json
import logging
import threading
from datetime import UTC, datetime
from pathlib import Path
from typing import Annotated, Any

from fastapi import APIRouter, Depends, HTTPException, Request, status

from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from configuration import configuration
from models.config import Action
from models.requests import FeedbackRequest, FeedbackStatusUpdateRequest
from models.responses import (
    ErrorResponse,
    FeedbackResponse,
    FeedbackStatusUpdateResponse,
    ForbiddenResponse,
    StatusResponse,
    UnauthorizedResponse,
)
from utils.suid import get_suid

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/feedback", tags=["feedback"])
feedback_status_lock = threading.Lock()

# Response for the feedback endpoint
feedback_post_response: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Feedback received and stored",
        "model": FeedbackResponse,
    },
    401: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "Client does not have permission to access resource",
        "model": ForbiddenResponse,
    },
    500: {
        "description": "User feedback can not be stored",
        "model": ErrorResponse,
    },
}

feedback_put_response: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Feedback status successfully updated",
        "model": FeedbackStatusUpdateResponse,
    },
    401: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "Client does not have permission to access resource",
        "model": ForbiddenResponse,
    },
}

feedback_get_response: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Feedback status successfully retrieved",
        "model": StatusResponse,
    }
}


def is_feedback_enabled() -> bool:
    """
    Check if feedback is enabled.

    Return whether user feedback collection is currently enabled
    based on configuration.

    Returns:
        bool: True if feedback collection is enabled; otherwise, False.
    """
    return configuration.user_data_collection_configuration.feedback_enabled


async def assert_feedback_enabled(_request: Request) -> None:
    """
    Ensure feedback collection is enabled for the current request.
    
    Raises:
        HTTPException: with status code 403 and detail "Forbidden: Feedback is disabled" when feedback is disabled.
    """
    feedback_enabled = is_feedback_enabled()
    if not feedback_enabled:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Forbidden: Feedback is disabled",
        )


@router.post("", responses=feedback_post_response)
@authorize(Action.FEEDBACK)
async def feedback_endpoint_handler(
    feedback_request: FeedbackRequest,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    _ensure_feedback_enabled: Any = Depends(assert_feedback_enabled),
) -> FeedbackResponse:
    """
    Store submitted user feedback and return a confirmation.
    
    Parameters:
        feedback_request: The feedback payload received from the client.
        auth: Authentication tuple from the dependency; the first element is the submitting user's ID.
    
    Returns:
        FeedbackResponse: Confirmation response with `response` set to "feedback received".
    
    Raises:
        HTTPException: If persisting feedback fails (results in a 500 response).
    """
    logger.debug("Feedback received %s", str(feedback_request))

    user_id, _, _, _ = auth
    try:
        store_feedback(user_id, feedback_request.model_dump(exclude={"model_config"}))
    except Exception as e:
        logger.error("Error storing user feedback: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Error storing user feedback",
                "cause": str(e),
            },
        ) from e

    return FeedbackResponse(response="feedback received")


def store_feedback(user_id: str, feedback: dict) -> None:
    """
    Persist the provided feedback as a JSON file in the configured local storage.
    
    The stored object will include the provided `user_id` and a UTC timestamp in addition to the `feedback` contents. The storage location is taken from configuration.user_data_collection_configuration.feedback_storage; the function will create the directory if it does not exist.
    
    Parameters:
        user_id (str): Identifier of the user who submitted the feedback.
        feedback (dict): Feedback payload to persist; merged with `user_id` and a UTC timestamp.
    
    Raises:
        OSError: If writing the feedback file fails.
        IOError: If writing the feedback file fails.
    """
    logger.debug("Storing feedback for user %s", user_id)
    # Creates storage path only if it doesn't exist. The `exist_ok=True` prevents
    # race conditions in case of multiple server instances trying to set up storage
    # at the same location.
    storage_path = Path(
        configuration.user_data_collection_configuration.feedback_storage or ""
    )
    storage_path.mkdir(parents=True, exist_ok=True)

    current_time = str(datetime.now(UTC))
    data_to_store = {"user_id": user_id, "timestamp": current_time, **feedback}

    # stores feedback in a file under unique uuid
    feedback_file_path = storage_path / f"{get_suid()}.json"
    try:
        with open(feedback_file_path, "w", encoding="utf-8") as feedback_file:
            json.dump(data_to_store, feedback_file)
    except (OSError, IOError) as e:
        logger.error("Failed to store feedback at %s: %s", feedback_file_path, e)
        raise

    logger.info("Feedback stored successfully at %s", feedback_file_path)


@router.get("/status", responses=feedback_get_response)
def feedback_status() -> StatusResponse:
    """
    Report whether feedback collection is enabled.
    
    Returns:
        StatusResponse: Response with `functionality="feedback"` and `status` set to `{"enabled": <bool>}` indicating the current enabled state.
    """
    logger.debug("Feedback status requested")
    feedback_status_enabled = is_feedback_enabled()
    return StatusResponse(
        functionality="feedback", status={"enabled": feedback_status_enabled}
    )


@router.put("/status", responses=feedback_put_response)
@authorize(Action.ADMIN)
async def update_feedback_status(
    feedback_update_request: FeedbackStatusUpdateRequest,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> FeedbackStatusUpdateResponse:
    """
    Update the runtime feedback-enabled flag to the requested value and return metadata about the change.
    
    Parameters:
        feedback_update_request (FeedbackStatusUpdateRequest): Request containing the desired feedback-enabled value.
    
    Returns:
        FeedbackStatusUpdateResponse: Object with `previous_status`, `updated_status`, `updated_by` (user id), and `timestamp` (UTC).
    """
    user_id, _, _, _ = auth
    requested_status = feedback_update_request.get_value()

    with feedback_status_lock:
        previous_status = (
            configuration.user_data_collection_configuration.feedback_enabled
        )
        configuration.user_data_collection_configuration.feedback_enabled = (
            requested_status
        )
        updated_status = (
            configuration.user_data_collection_configuration.feedback_enabled
        )
        current_time = str(datetime.now(UTC))

    return FeedbackStatusUpdateResponse(
        status={
            "previous_status": previous_status,
            "updated_status": updated_status,
            "updated_by": user_id,
            "timestamp": current_time,
        }
    )