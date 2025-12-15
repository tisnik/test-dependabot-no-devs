"""Integration tests for the /config endpoint."""

import pytest
from fastapi import HTTPException, Request, status

from app.endpoints.config import config_endpoint_handler
from authentication.interface import AuthTuple
from configuration import AppConfig


@pytest.mark.asyncio
async def test_config_endpoint_returns_config(
    test_config: AppConfig,
    test_request: Request,
    test_auth: AuthTuple,
) -> None:
    """
    Verify the /config endpoint returns the provided test configuration.
    
    Parameters:
        test_config (AppConfig): Fixture providing the expected configuration to be returned.
        test_request (Request): FastAPI request object used to call the endpoint.
        test_auth (AuthTuple): Authentication fixture used for the request.
    """
    response = await config_endpoint_handler(auth=test_auth, request=test_request)

    # Verify that response matches the real configuration
    assert response.configuration == test_config.configuration


@pytest.mark.asyncio
async def test_config_endpoint_returns_current_config(
    current_config: AppConfig,
    test_request: Request,
    test_auth: AuthTuple,
) -> None:
    """
    Verify the /config endpoint returns the application's current (root) configuration.
    
    Calls the config endpoint handler with the provided request and authentication and asserts the returned configuration equals current_config.configuration.
    """
    response = await config_endpoint_handler(auth=test_auth, request=test_request)

    # Verify that response matches the root configuration
    assert response.configuration == current_config.configuration


@pytest.mark.asyncio
async def test_config_endpoint_fails_without_configuration(
    test_request: Request,
    test_auth: AuthTuple,
) -> None:
    """
    Verify the /config endpoint raises an HTTP 500 error when no configuration is loaded.
    
    Asserts that calling the endpoint raises an HTTPException with status code 500 and that the exception detail's "response" contains the phrase "configuration is not loaded" (case-insensitive).
    
    Parameters:
    	test_request (Request): FastAPI request fixture
    	test_auth (AuthTuple): noop authentication fixture
    """

    # Verify that HTTPException is raised when configuration is not loaded
    with pytest.raises(HTTPException) as exc_info:
        await config_endpoint_handler(auth=test_auth, request=test_request)

    # Verify error details
    assert exc_info.value.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR
    assert isinstance(exc_info.value.detail, dict)
    assert (
        "configuration is not loaded" in exc_info.value.detail["response"].lower()
    )  # type: ignore