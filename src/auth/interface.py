"""Abstract base class for authentication methods."""

from abc import ABC, abstractmethod

from fastapi import Request


class AuthInterface(ABC):  # pylint: disable=too-few-public-methods
    """Base class for all authentication method implementations."""

    @abstractmethod
    async def __call__(self, request: Request) -> tuple[str, str, str]:
        """
        Validates a FastAPI request for authentication and authorization.
        
        Parameters:
            request (Request): The incoming FastAPI request to be authenticated.
        
        Returns:
            tuple[str, str, str]: A tuple containing authentication and authorization details, as defined by the implementation.
        """
