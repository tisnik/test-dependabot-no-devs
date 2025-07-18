"""Manage authentication flow for FastAPI endpoints with no-op auth."""

import logging

from fastapi import Request

from constants import (
    DEFAULT_USER_NAME,
    DEFAULT_USER_UID,
    DEFAULT_VIRTUAL_PATH,
)
from auth.interface import AuthInterface
from auth.utils import extract_user_token

logger = logging.getLogger(__name__)


class NoopWithTokenAuthDependency(
    AuthInterface
):  # pylint: disable=too-few-public-methods
    """No-op AuthDependency class that bypasses authentication and authorization checks."""

    def __init__(self, virtual_path: str = DEFAULT_VIRTUAL_PATH) -> None:
        """
        Initialize the no-op authentication dependency with an optional virtual path for authorization checks.
        
        Parameters:
            virtual_path (str): The virtual path used for authorization checks. Defaults to a predefined constant.
        """
        self.virtual_path = virtual_path

    async def __call__(self, request: Request) -> tuple[str, str, str]:
        """
        Simulates authentication for a FastAPI request by extracting a user token and user ID without enforcing real authentication or authorization.
        
        Parameters:
            request (Request): The incoming FastAPI request.
        
        Returns:
            tuple[str, str, str]: A tuple containing the user ID (from query parameters or a default), a default username, and the extracted user token.
        """
        logger.warning(
            "No-op with token authentication dependency is being used. "
            "The service is running in insecure mode intended solely for development purposes"
        )
        # try to extract user token from request
        user_token = extract_user_token(request.headers)
        # try to extract user ID from request
        user_id = request.query_params.get("user_id", DEFAULT_USER_UID)
        logger.debug("Retrieved user ID: %s", user_id)
        return user_id, DEFAULT_USER_NAME, user_token
