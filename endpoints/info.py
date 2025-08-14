"""Handler for REST API call to provide info."""

import logging
from typing import Any

from fastapi import APIRouter, Request

from configuration import configuration
from version import __version__
from models.responses import InfoResponse

logger = logging.getLogger(__name__)
router = APIRouter(tags=["info"])


get_info_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "name": "Service name",
        "version": "Service version",
    },
}


@router.get("/info", responses=get_info_responses)
def info_endpoint_handler(_request: Request) -> InfoResponse:
    """
    Return service information for the /info endpoint.
    
    Builds and returns an InfoResponse containing the service name (sourced from
    configuration.configuration.name) and the package version (from __version__).
    The `_request` parameter is accepted for compatibility with FastAPI handlers but
    is not used.
    Returns:
        InfoResponse: Response model with `name` and `version` fields.
    """
    return InfoResponse(name=configuration.configuration.name, version=__version__)
