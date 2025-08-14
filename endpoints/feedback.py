"""Handler for REST API endpoint for user feedback."""

import logging
from typing import Annotated, Any
from pathlib import Path
import json
from datetime import datetime, UTC
from fastapi import APIRouter, Request, HTTPException, Depends, status

from auth import get_auth_dependency
from auth.interface import AuthTuple
from configuration import configuration
from models.responses import (
    FeedbackResponse,
    StatusResponse,
    UnauthorizedResponse,
    ForbiddenResponse,
)
from models.requests import FeedbackRequest
from utils.suid import get_suid

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/feedback", tags=["feedback"])
auth_dependency = get_auth_dependency()

# Response for the feedback endpoint
feedback_response: dict[int | str, dict[str, Any]] = {
    200: {"response": "Feedback received and stored"},
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "User is not authorized",
        "model": ForbiddenResponse,
    },
}


def is_feedback_enabled() -> bool:
    """Check if feedback is enabled.

    Returns:
        bool: True if feedback is enabled, False otherwise.
    """
    return configuration.user_data_collection_configuration.feedback_enabled


async def assert_feedback_enabled(_request: Request) -> None:
    """
    Dependency that ensures the feedback feature is enabled.
    
    This function is intended for use as a FastAPI dependency. It checks the current feature flag and raises an HTTP 403 Forbidden if feedback collection is disabled.
    
    Parameters:
        _request (Request): The incoming FastAPI request (unused, accepted for dependency injection).
    
    Raises:
        HTTPException: Always raised with status 403 if feedback is disabled.
    """
    feedback_enabled = is_feedback_enabled()
    if not feedback_enabled:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Forbidden: Feedback is disabled",
        )


@router.post("", responses=feedback_response)
def feedback_endpoint_handler(
    feedback_request: FeedbackRequest,
    auth: Annotated[AuthTuple, Depends(auth_dependency)],
    _ensure_feedback_enabled: Any = Depends(assert_feedback_enabled),
) -> FeedbackResponse:
    """
    Handle incoming feedback submissions from an authenticated user.
    
    Stores the provided feedback payload on disk associated with the authenticated user's ID.
    If storage fails, raises an HTTPException with status 500 and a detail object containing the error cause.
    
    Parameters:
        feedback_request: Payload containing the user's feedback (validated FeedbackRequest).
    
    Returns:
        FeedbackResponse indicating that the feedback was received.
    
    Raises:
        HTTPException: Raised with status 500 if storing the feedback fails.
    """
    logger.debug("Feedback received %s", str(feedback_request))

    user_id, _, _ = auth
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
    Store a user's feedback as a JSON file on the local filesystem.
    
    Writes a JSON object that includes the provided `user_id`, a UTC timestamp, and the fields from `feedback` into a uniquely named file within the configured feedback storage directory. The directory is created if missing; the filename is generated to be unique.
    
    Parameters:
        user_id (str): The user's identifier (expected to be a UUID string).
        feedback (dict): A JSON-serializable mapping containing the feedback payload.
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
    with open(feedback_file_path, "w", encoding="utf-8") as feedback_file:
        json.dump(data_to_store, feedback_file)

    logger.info("Feedback stored successfully at %s", feedback_file_path)


@router.get("/status")
def feedback_status() -> StatusResponse:
    """
    Return the current enabled/disabled status of the feedback feature.
    
    Queries configuration to determine whether user feedback collection is enabled and returns a StatusResponse with functionality set to "feedback" and a status object containing {"enabled": <bool>} indicating the feature state.
    
    Returns:
        StatusResponse: Status payload with keys:
            - functionality: "feedback"
            - status: {"enabled": bool}
    """
    logger.debug("Feedback status requested")
    feedback_status_enabled = is_feedback_enabled()
    return StatusResponse(
        functionality="feedback", status={"enabled": feedback_status_enabled}
    )
