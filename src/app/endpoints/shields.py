"""Handler for REST API call to list available shields."""

import logging
from typing import Annotated, Any

from fastapi import APIRouter, HTTPException, Request, status
from fastapi.params import Depends
from llama_stack_client import APIConnectionError

from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.config import Action
from models.responses import ShieldsResponse
from utils.endpoints import check_configuration_loaded

logger = logging.getLogger(__name__)
router = APIRouter(tags=["shields"])


shields_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "shields": [
            {
                "identifier": "lightspeed_question_validity-shield",
                "provider_resource_id": "lightspeed_question_validity-shield",
                "provider_id": "lightspeed_question_validity",
                "type": "shield",
                "params": {},
            }
        ]
    },
    500: {"description": "Connection to Llama Stack is broken"},
}


@router.get("/shields", responses=shields_responses)
@authorize(Action.GET_SHIELDS)
async def shields_endpoint_handler(
    request: Request,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> ShieldsResponse:
    """
    Handle requests to the /shields endpoint.

    Process GET requests to the /shields endpoint, returning a list of available
    shields from the Llama Stack service.

    Raises:
        HTTPException: If unable to connect to the Llama Stack server or if
        shield retrieval fails for any reason.

    Returns:
        ShieldsResponse: An object containing the list of available shields.
    """
    # Used only by the middleware
    _ = auth

    # Nothing interesting in the request
    _ = request

    check_configuration_loaded(configuration)

    llama_stack_configuration = configuration.llama_stack_configuration
    logger.info("Llama stack config: %s", llama_stack_configuration)

    try:
        # try to get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()
        # retrieve shields
        shields = await client.shields.list()
        s = [dict(s) for s in shields]
        return ShieldsResponse(shields=s)

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
    # any other exception that can occur during shield listing
    except Exception as e:
        logger.error("Unable to retrieve list of shields: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to retrieve list of shields",
                "cause": str(e),
            },
        ) from e
