"""Handler for REST API calls to list and retrieve available providers."""

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
from models.responses import ProvidersListResponse, ProviderResponse
from utils.endpoints import check_configuration_loaded

logger = logging.getLogger(__name__)
router = APIRouter(tags=["providers"])


providers_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "providers": {
            "agents": [
                {
                    "provider_id": "meta-reference",
                    "provider_type": "inline::meta-reference",
                }
            ],
            "datasetio": [
                {"provider_id": "huggingface", "provider_type": "remote::huggingface"},
                {"provider_id": "localfs", "provider_type": "inline::localfs"},
            ],
            "inference": [
                {
                    "provider_id": "sentence-transformers",
                    "provider_type": "inline::sentence-transformers",
                },
                {"provider_id": "openai", "provider_type": "remote::openai"},
            ],
        }
    },
    500: {"description": "Connection to Llama Stack is broken"},
}

provider_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "api": "inference",
        "config": {"api_key": "********"},
        "health": {
            "status": "Not Implemented",
            "message": "Provider does not implement health check",
        },
        "provider_id": "openai",
        "provider_type": "remote::openai",
    },
    404: {"response": "Provider with given id not found"},
    500: {
        "response": "Unable to retrieve list of providers",
        "cause": "Connection to Llama Stack is broken",
    },
}


@router.get("/providers", responses=providers_responses)
@authorize(Action.LIST_PROVIDERS)
async def providers_endpoint_handler(
    request: Request,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> ProvidersListResponse:
    """
    Handle GET requests to list all available providers.

    Retrieves providers from the Llama Stack service, groups them by API type.

    Raises:
        HTTPException:
            - 500 if configuration is not loaded,
            - 500 if unable to connect to Llama Stack,
            - 500 for any unexpected retrieval errors.

    Returns:
        ProvidersListResponse: Object mapping API types to lists of providers.
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
        # retrieve providers
        providers = await client.providers.list()
        providers = [dict(p) for p in providers]
        return ProvidersListResponse(providers=group_providers(providers))

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
    # any other exception that can occur during model listing
    except Exception as e:
        logger.error("Unable to retrieve list of providers: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to retrieve list of providers",
                "cause": str(e),
            },
        ) from e


def group_providers(providers: list[dict[str, Any]]) -> dict[str, list[dict[str, Any]]]:
    """Group a list of providers by their API type.

    Args:
        providers: List of provider dictionaries. Each must contain
            'api', 'provider_id', and 'provider_type' keys.

    Returns:
        Mapping from API type to list of providers containing
        only 'provider_id' and 'provider_type'.
    """
    result: dict[str, list[dict[str, Any]]] = {}
    for provider in providers:
        result.setdefault(provider["api"], []).append(
            {
                "provider_id": provider["provider_id"],
                "provider_type": provider["provider_type"],
            }
        )
    return result


@router.get("/providers/{provider_id}", responses=provider_responses)
@authorize(Action.GET_PROVIDER)
async def get_provider_endpoint_handler(
    request: Request,
    provider_id: str,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
) -> ProviderResponse:
    """Retrieve a single provider by its unique ID.

    Raises:
        HTTPException:
            - 404 if provider with the given ID is not found,
            - 500 if unable to connect to Llama Stack,
            - 500 for any unexpected retrieval errors.

    Returns:
        ProviderResponse: A single provider's details including API, config, health,
        provider_id, and provider_type.
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
        # retrieve providers
        providers = await client.providers.list()
        p = [dict(p) for p in providers]
        match = next((item for item in p if item["provider_id"] == provider_id), None)
        if not match:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail={"response": f"Provider with id '{provider_id}' not found"},
            )
        return ProviderResponse(**match)

    # connection to Llama Stack server
    except HTTPException:
        raise
    except APIConnectionError as e:
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e
    # any other exception that can occur during model listing
    except Exception as e:
        logger.error("Unable to retrieve list of providers: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to retrieve list of providers",
                "cause": str(e),
            },
        ) from e
