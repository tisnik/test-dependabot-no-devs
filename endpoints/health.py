"""Handlers for health REST API endpoints.

These endpoints are used to check if service is live and prepared to accept
requests. Note that these endpoints can be accessed using GET or HEAD HTTP
methods. For HEAD HTTP method, just the HTTP response code is used.
"""

import logging
from typing import Any

from llama_stack.providers.datatypes import HealthStatus

from fastapi import APIRouter, status, Response
from client import AsyncLlamaStackClientHolder
from models.responses import (
    LivenessResponse,
    ReadinessResponse,
    ProviderHealthStatus,
)

logger = logging.getLogger(__name__)
router = APIRouter(tags=["health"])


async def get_providers_health_statuses() -> list[ProviderHealthStatus]:
    """
    Retrieve health statuses for all configured providers.
    
    Queries the asynchronous client for the list of providers and maps each provider's reported health into a list of ProviderHealthStatus objects (provider_id, status, message). If querying providers fails (for example no providers are defined or the client call raises), returns a single ProviderHealthStatus with provider_id "unknown", status set to HealthStatus.ERROR.value, and a message describing the failure.
    
    Returns:
        list[ProviderHealthStatus]: A list of provider health statuses; on error a single-item list containing an "unknown" provider with an error status.
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
async def readiness_probe_get_method(response: Response) -> ReadinessResponse:
    """
    Return the service readiness status including any unhealthy providers.
    
    Queries provider health statuses and returns a ReadinessResponse where `ready` is True
    only if no provider reports HealthStatus.ERROR. If any providers are unhealthy, the
    response's HTTP status code is set to 503 (Service Unavailable) and the `reason`
    field lists the unhealthy provider IDs; `providers` contains those provider entries.
    """
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
    503: {
        "description": "Service is not alive",
        "model": LivenessResponse,
    },
}


@router.get("/liveness", responses=get_liveness_responses)
def liveness_probe_get_method() -> LivenessResponse:
    """
    Return the service liveness status.
    
    This probe always reports the service as alive (LivenessResponse.alive == True)
    and does not perform external or dynamic health checks.
    
    Returns:
        LivenessResponse: Response object with `alive` set to True.
    """
    return LivenessResponse(alive=True)
