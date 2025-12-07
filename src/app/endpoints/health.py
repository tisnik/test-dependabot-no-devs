"""Handlers for health REST API endpoints.

These endpoints are used to check if service is live and prepared to accept
requests. Note that these endpoints can be accessed using GET or HEAD HTTP
methods. For HEAD HTTP method, just the HTTP response code is used.
"""

import logging
from typing import Annotated, Any

from fastapi import APIRouter, Depends, Response, status
from llama_stack.providers.datatypes import HealthStatus

from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from models.config import Action
from models.responses import (
    LivenessResponse,
    ProviderHealthStatus,
    ReadinessResponse,
)

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["health"])


async def get_providers_health_statuses() -> list[ProviderHealthStatus]:
    """
    Retrieve the health status for each configured provider.
    
    Returns:
        list[ProviderHealthStatus]: A list of provider health entries. If providers cannot be queried or an error occurs, returns a single `ProviderHealthStatus` with `provider_id == "unknown"`, `status == HealthStatus.ERROR.value`, and a `message` describing the failure.
    """
    try:
        client = AsyncLlamaStackClientHolder().get_client()

        providers = await client.providers.list()
        logger.debug("Found %d providers", len(providers))

        health_results = [
            ProviderHealthStatus(
                provider_id=provider.provider_id,
                status=str(provider.health.get("status", "unknown")),
                message=str(provider.health.get("message", "")),
            )
            for provider in providers
        ]
        return health_results

    except Exception as e:  # pylint: disable=broad-exception-caught
        # eg. no providers defined
        logger.error("Failed to check providers health: %s", e)
        return [
            ProviderHealthStatus(
                provider_id="unknown",
                status=HealthStatus.ERROR.value,
                message=f"Failed to initialize health check: {str(e)}",
            )
        ]


get_readiness_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Service is ready",
        "model": ReadinessResponse,
    },
    503: {
        "description": "Service is not ready",
        "model": ReadinessResponse,
    },
}


@router.get("/readiness", responses=get_readiness_responses)
@authorize(Action.INFO)
async def readiness_probe_get_method(
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    response: Response,
) -> ReadinessResponse:
    """
    Assess service readiness by evaluating configured providers' health.
    
    Sets response.status_code to 503 if any provider reports an ERROR status; otherwise the response remains 200.
    Parameters:
        response (fastapi.Response): FastAPI response object used to set the HTTP status code.
    Returns:
        ReadinessResponse: Object with `ready` indicating overall readiness, `reason` explaining the outcome, and `providers` containing the list of unhealthy ProviderHealthStatus entries (empty when ready).
    """
    # Used only for authorization
    _ = auth

    logger.info("Response to /v1/readiness endpoint")

    provider_statuses = await get_providers_health_statuses()

    # Check if any provider is unhealthy (not counting not_implemented as unhealthy)
    unhealthy_providers = [
        p for p in provider_statuses if p.status == HealthStatus.ERROR.value
    ]

    if unhealthy_providers:
        ready = False
        unhealthy_provider_names = [p.provider_id for p in unhealthy_providers]
        reason = f"Providers not healthy: {', '.join(unhealthy_provider_names)}"
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
    else:
        ready = True
        reason = "All providers are healthy"

    return ReadinessResponse(ready=ready, reason=reason, providers=unhealthy_providers)


get_liveness_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "Service is alive",
        "model": LivenessResponse,
    },
    # HTTP_503_SERVICE_UNAVAILABLE will never be returned when unreachable
}


@router.get("/liveness", responses=get_liveness_responses)
@authorize(Action.INFO)
async def liveness_probe_get_method(
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> LivenessResponse:
    """
    Indicates whether the service is alive.
    
    Returns:
        LivenessResponse: Response object with `alive` set to `True`.
    """
    # Used only for authorization
    _ = auth

    logger.info("Response to /v1/liveness endpoint")

    return LivenessResponse(alive=True)