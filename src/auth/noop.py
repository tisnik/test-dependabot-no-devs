"""Manage authentication flow for FastAPI endpoints with no-op auth."""

import logging

from fastapi import Request

from constants import (
    DEFAULT_USER_NAME,
    DEFAULT_USER_UID,
    NO_USER_TOKEN,
    DEFAULT_VIRTUAL_PATH,
)
from auth.interface import AuthInterface

logger = logging.getLogger(__name__)


class NoopAuthDependency(AuthInterface):  # pylint: disable=too-few-public-methods
    """No-op AuthDependency class that bypasses authentication and authorization checks."""

    def __init__(self, virtual_path: str = DEFAULT_VIRTUAL_PATH) -> None:
        """
        Initialize the no-op authentication dependency with an optional virtual path.
        
        Parameters:
            virtual_path (str): The virtual path to associate with this authentication dependency. Defaults to a predefined constant.
        """
        self.virtual_path = virtual_path

    async def __call__(self, request: Request) -> tuple[str, str, str]:
        """
        Bypasses authentication and authorization, returning a user identity based on the request's query parameters or default values.
        
        If a `user_id` is provided in the request's query parameters, it is used; otherwise, a default user ID is returned. The username and token are always set to default values.
        
        Returns:
            tuple[str, str, str]: A tuple containing the user ID, username, and token.
        """
        logger.warning(
            "No-op authentication dependency is being used. "
            "The service is running in insecure mode intended solely for development purposes"
        )
        # try to extract user ID from request
        user_id = request.query_params.get("user_id", DEFAULT_USER_UID)
        logger.debug("Retrieved user ID: %s", user_id)
        return user_id, DEFAULT_USER_NAME, NO_USER_TOKEN
