"""Utility functions for endpoint handlers."""

from fastapi import HTTPException, status

import constants
from models.requests import QueryRequest
from configuration import AppConfig


def check_configuration_loaded(configuration: AppConfig) -> None:
    """
    Raises an HTTP 500 error if the application configuration is not loaded.
    
    If the provided configuration is None, an HTTPException is raised with a JSON detail indicating the configuration is missing.
    """
    if configuration is None:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={"response": "Configuration is not loaded"},
        )


def get_system_prompt(query_request: QueryRequest, configuration: AppConfig) -> str:
    """
    Determine the appropriate system prompt for a query request based on request data and application configuration.
    
    If custom system prompts are disabled in the configuration and the request includes one, raises an HTTP 422 error. Otherwise, returns the system prompt from the request if present, then from the configuration if available, or falls back to a default system prompt.
    
    Returns:
        str: The selected system prompt string.
    """
    system_prompt_disabled = (
        configuration.customization is not None
        and configuration.customization.disable_query_system_prompt
    )
    if system_prompt_disabled and query_request.system_prompt:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail={
                "response": (
                    "This instance does not support customizing the system prompt in the "
                    "query request (disable_query_system_prompt is set). Please remove the "
                    "system_prompt field from your request."
                )
            },
        )

    if query_request.system_prompt:
        # Query taking precedence over configuration is the only behavior that
        # makes sense here - if the configuration wants precedence, it can
        # disable query system prompt altogether with disable_system_prompt.
        return query_request.system_prompt

    if (
        configuration.customization is not None
        and configuration.customization.system_prompt is not None
    ):
        return configuration.customization.system_prompt

    # default system prompt has the lowest precedence
    return constants.DEFAULT_SYSTEM_PROMPT
