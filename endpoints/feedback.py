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
    """
    Return whether user feedback collection is currently enabled based on configuration.
    
    Returns:
        bool: True if feedback collection is enabled; otherwise, False.
    """
    return configuration.user_data_collection_configuration.feedback_enabled


async def assert_feedback_enabled(_request: Request) -> None:
    """
    Ensures that feedback collection is enabled; raises an HTTP 403 error if it is not.
    
    Raises:
        HTTPException: If feedback collection is disabled.
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
    Processes a user feedback submission, storing the feedback and returning a confirmation response.
    
    Raises:
        HTTPException: Returns HTTP 500 if feedback storage fails.
        
    Returns:
        FeedbackResponse: Indicates successful receipt of feedback.
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
    Persist user feedback to a uniquely named JSON file in the configured local storage directory.
    
    Parameters:
        user_id (str): Unique identifier of the user submitting feedback.
        feedback (dict): Feedback data to be stored, merged with user ID and timestamp.
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
    Return the current enabled status of the feedback functionality.
    
    Returns:
        StatusResponse: Indicates whether feedback collection is enabled.
    """
    logger.debug("Feedback status requested")
    feedback_status_enabled = is_feedback_enabled()
    return StatusResponse(
        functionality="feedback", status={"enabled": feedback_status_enabled}
    )
