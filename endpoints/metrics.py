"""Handler for REST API call to provide metrics."""

from fastapi.responses import PlainTextResponse
from fastapi import APIRouter, Request
from prometheus_client import (
    generate_latest,
    CONTENT_TYPE_LATEST,
)

from metrics.utils import setup_model_metrics

router = APIRouter(tags=["metrics"])


@router.get("/metrics", response_class=PlainTextResponse)
async def metrics_endpoint_handler(_request: Request) -> PlainTextResponse:
    """
    Return the current Prometheus metrics in plaintext for the /metrics endpoint.
    
    Awaits the one-time asynchronous setup of model metrics, then returns the latest
    Prometheus metrics output formatted with the Prometheus text content type.
    Exceptions from setup or metrics generation propagate to the framework.
    
    Parameters:
        _request (Request): Incoming FastAPI request (unused; present to match the route handler signature).
    
    Returns:
        PlainTextResponse: HTTP response containing Prometheus plaintext metrics with the correct content type.
    """
    # Setup the model metrics if not already done. This is a one-time setup
    # and will not be run again on subsequent calls to this endpoint
    await setup_model_metrics()
    return PlainTextResponse(generate_latest(), media_type=CONTENT_TYPE_LATEST)
