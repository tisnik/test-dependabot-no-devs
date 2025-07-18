"""Handler for REST API call to provide info."""

import logging
from typing import Any

from fastapi import APIRouter, Request

from version import __version__
from models.responses import InfoResponse

logger = logging.getLogger(__name__)
router = APIRouter(tags=["info"])


get_into_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "name": "Service name",
        "version": "Service version",
    },
}


@router.get("/info", responses=get_into_responses)
def info_endpoint_handler(_request: Request) -> InfoResponse:
    """Handle request to the /info endpoint."""
    return InfoResponse(name="foo", version=__version__)
