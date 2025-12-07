"""Handler for REST API call to provide info."""

import logging
from typing import Annotated, Any

from fastapi import APIRouter, Depends, HTTPException, Request, status
from llama_stack_client import APIConnectionError

from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.config import Action
from models.responses import InfoResponse
from version import __version__

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["info"])


get_info_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "name": "Service name",
        "service_version": "Service version",
        "llama_stack_version": "Llama Stack version",
    },
    500: {
        "detail": {
            "response": "Unable to connect to Llama Stack",
            "cause": "Connection error.",
        }
    },
}


@router.get("/info", responses=get_info_responses)
@authorize(Action.INFO)
async def info_endpoint_handler(
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    request: Request,
) -> InfoResponse:
    """
    Return service metadata: configured service name, application version, and Llama Stack version.
    
    Attempts to retrieve the Llama Stack version and returns an InfoResponse with:
    - name: configured service name
    - service_version: application version
    - llama_stack_version: retrieved Llama Stack version
    
    Raises:
        HTTPException: with status 500 and a detail object containing `response` and `cause` when unable to connect to Llama Stack.
    
    Returns:
        InfoResponse: Object with `name`, `service_version`, and `llama_stack_version`.
    """
    # Used only for authorization
    _ = auth

    # Nothing interesting in the request
    _ = request

    logger.info("Response to /v1/info endpoint")

    try:
        # try to get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()
        # retrieve version
        llama_stack_version_object = await client.inspect.version()
        llama_stack_version = llama_stack_version_object.version
        logger.debug("Service name: %s", configuration.configuration.name)
        logger.debug("Service version: %s", __version__)
        logger.debug("LLama Stack version: %s", llama_stack_version)
        return InfoResponse(
            name=configuration.configuration.name,
            service_version=__version__,
            llama_stack_version=llama_stack_version,
        )
    # connection to Llama Stack server
    except APIConnectionError as e:
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e