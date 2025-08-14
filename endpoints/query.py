"""Handler for REST API call to provide answer to query."""

from datetime import datetime, UTC
import json
import logging
import os
from pathlib import Path
from typing import Annotated, Any

from llama_stack_client import APIConnectionError
from llama_stack_client import AsyncLlamaStackClient  # type: ignore
from llama_stack_client.types import UserMessage, Shield  # type: ignore
from llama_stack_client.types.agents.turn_create_params import (
    ToolgroupAgentToolGroupWithArgs,
    Toolgroup,
)
from llama_stack_client.types.model_list_response import ModelListResponse

from fastapi import APIRouter, HTTPException, status, Depends

from auth import get_auth_dependency
from auth.interface import AuthTuple
from client import AsyncLlamaStackClientHolder
from configuration import configuration
import metrics
from models.responses import QueryResponse, UnauthorizedResponse, ForbiddenResponse
from models.requests import QueryRequest, Attachment
import constants
from utils.endpoints import check_configuration_loaded, get_agent, get_system_prompt
from utils.mcp_headers import mcp_headers_dependency, handle_mcp_headers_with_toolgroups
from utils.suid import get_suid

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["query"])
auth_dependency = get_auth_dependency()

query_response: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "response": "LLM answer",
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "User is not authorized",
        "model": ForbiddenResponse,
    },
    503: {
        "detail": {
            "response": "Unable to connect to Llama Stack",
            "cause": "Connection error.",
        }
    },
}


def is_transcripts_enabled() -> bool:
    """Check if transcripts is enabled.

    Returns:
        bool: True if transcripts is enabled, False otherwise.
    """
    return configuration.user_data_collection_configuration.transcripts_enabled


@router.post("/query", responses=query_response)
async def query_endpoint_handler(
    query_request: QueryRequest,
    auth: Annotated[AuthTuple, Depends(auth_dependency)],
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> QueryResponse:
    """
    Handle POST /query: route the user's query through the configured Llama Stack client and return the model response.
    
    Processes the incoming QueryRequest, selects an appropriate model/provider, obtains a response (and conversation id) from Llama Stack, increments call metrics, and optionally persists a transcript when transcript collection is enabled.
    
    Parameters:
        query_request (QueryRequest): The user's query payload and related options (documents, attachments, tool usage flags, etc.).
        mcp_headers (dict[str, dict[str, str]]): Optional MCP header mappings used to configure tool/network calls for the provider.
    
    Returns:
        QueryResponse: Contains the Llama Stack conversation_id and the model/agent textual response.
    
    Raises:
        HTTPException: Raised with status 500 when the service cannot connect to the Llama Stack backend (APIConnectionError).
    """
    check_configuration_loaded(configuration)

    llama_stack_config = configuration.llama_stack_configuration
    logger.info("LLama stack config: %s", llama_stack_config)

    user_id, _, token = auth

    try:
        # try to get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()
        model_id, provider_id = select_model_and_provider_id(
            await client.models.list(), query_request
        )
        response, conversation_id = await retrieve_response(
            client,
            model_id,
            query_request,
            token,
            mcp_headers=mcp_headers,
        )
        # Update metrics for the LLM call
        metrics.llm_calls_total.labels(provider_id, model_id).inc()

        if not is_transcripts_enabled():
            logger.debug("Transcript collection is disabled in the configuration")
        else:
            store_transcript(
                user_id=user_id,
                conversation_id=conversation_id,
                query_is_valid=True,  # TODO(lucasagomes): implement as part of query validation
                query=query_request.query,
                query_request=query_request,
                response=response,
                rag_chunks=[],  # TODO(lucasagomes): implement rag_chunks
                truncated=False,  # TODO(lucasagomes): implement truncation as part of quota work
                attachments=query_request.attachments or [],
            )

        return QueryResponse(conversation_id=conversation_id, response=response)

    # connection to Llama Stack server
    except APIConnectionError as e:
        # Update metrics for the LLM call failure
        metrics.llm_calls_failures_total.inc()
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e


def select_model_and_provider_id(
    models: ModelListResponse, query_request: QueryRequest
) -> tuple[str, str | None]:
    """
    Selects the Llama Stack model identifier and provider to use for a query.
    
    If the request contains both model and provider, those are used. Otherwise the function falls back to configured defaults; if those are not set it picks the first available model of type "llm" from the provided models list. The returned model identifier is formatted as "provider_id/model_id".
    
    Parameters:
        models: The available models listing returned by the models discovery API.
        query_request: The incoming query request which may contain optional `model` and `provider` overrides.
    
    Returns:
        A tuple (llama_stack_model_id, provider_id) where `llama_stack_model_id` is "provider_id/model_id".
    
    Raises:
        HTTPException (400) if no suitable LLM model is available or if the selected model/provider pair is not present in `models`.
    """
    # If model_id and provider_id are provided in the request, use them
    model_id = query_request.model
    provider_id = query_request.provider

    # If model_id is not provided in the request, check the configuration
    if not model_id or not provider_id:
        logger.debug(
            "No model ID or provider ID specified in request, checking configuration"
        )
        model_id = configuration.inference.default_model  # type: ignore[reportAttributeAccessIssue]
        provider_id = (
            configuration.inference.default_provider  # type: ignore[reportAttributeAccessIssue]
        )

    # If no model is specified in the request or configuration, use the first available LLM
    if not model_id or not provider_id:
        logger.debug(
            "No model ID or provider ID specified in request or configuration, "
            "using the first available LLM"
        )
        try:
            model = next(
                m
                for m in models
                if m.model_type == "llm"  # pyright: ignore[reportAttributeAccessIssue]
            )
            model_id = model.identifier
            provider_id = model.provider_id
            logger.info("Selected model: %s", model)
            return model_id, provider_id
        except (StopIteration, AttributeError) as e:
            message = "No LLM model found in available models"
            logger.error(message)
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail={
                    "response": constants.UNABLE_TO_PROCESS_RESPONSE,
                    "cause": message,
                },
            ) from e

    llama_stack_model_id = f"{provider_id}/{model_id}"
    # Validate that the model_id and provider_id are in the available models
    logger.debug("Searching for model: %s, provider: %s", model_id, provider_id)
    if not any(
        m.identifier == llama_stack_model_id and m.provider_id == provider_id
        for m in models
    ):
        message = f"Model {model_id} from provider {provider_id} not found in available models"
        logger.error(message)
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail={
                "response": constants.UNABLE_TO_PROCESS_RESPONSE,
                "cause": message,
            },
        )

    return llama_stack_model_id, provider_id


def _is_inout_shield(shield: Shield) -> bool:
    """
    Return True if the shield's identifier denotes an inout shield.
    
    Checks whether the shield's `identifier` string starts with the prefix `"inout_"`, which designates shields treated as both input and output.
    """
    return shield.identifier.startswith("inout_")


def is_output_shield(shield: Shield) -> bool:
    """
    Return True if the given Shield should be treated as an output-monitoring shield.
    
    A shield is considered an output shield when its identifier starts with "output_" or when it is an inout shield (identifier starts with "inout_").
    """
    return _is_inout_shield(shield) or shield.identifier.startswith("output_")


def is_input_shield(shield: Shield) -> bool:
    """
    Return True if the given shield should be treated as an input shield.
    
    A shield is considered an input shield when it is an "inout" shield (identifier starts with "inout_")
    or when it is not classified as an output shield.
    """
    return _is_inout_shield(shield) or not is_output_shield(shield)


async def retrieve_response(  # pylint: disable=too-many-locals
    client: AsyncLlamaStackClient,
    model_id: str,
    query_request: QueryRequest,
    token: str,
    mcp_headers: dict[str, dict[str, str]] | None = None,
) -> tuple[str, str]:
    """
    Retrieve a response from a Llama Stack LLM or agent for the given query.
    
    Builds/shares shields and MCP headers with the agent, optionally includes RAG toolgroups,
    executes a single agent turn with the user's query and documents, and returns the agent's
    final output and the conversation identifier.
    
    Parameters:
        model_id: Llama Stack model identifier string (format expected by the client).
        query_request: Request object containing the user query, optional attachments,
            conversation_id, documents, and no_tools flag.
        token: User bearer token used to populate MCP Authorization headers when needed.
        mcp_headers: Optional mapping of MCP server URL -> headers dict to forward to toolgroups.
            If not provided or empty and `token` is present, Authorization headers will be
            injected for configured MCP servers.
    
    Returns:
        A tuple (response_text, conversation_id) where `response_text` is the agent/LLM output
        converted to a string and `conversation_id` is the conversation identifier used/returned
        by the agent.
    """
    available_input_shields = [
        shield.identifier
        for shield in filter(is_input_shield, await client.shields.list())
    ]
    available_output_shields = [
        shield.identifier
        for shield in filter(is_output_shield, await client.shields.list())
    ]
    if not available_input_shields and not available_output_shields:
        logger.info("No available shields. Disabling safety")
    else:
        logger.info(
            "Available input shields: %s, output shields: %s",
            available_input_shields,
            available_output_shields,
        )
    # use system prompt from request or default one
    system_prompt = get_system_prompt(query_request, configuration)
    logger.debug("Using system prompt: %s", system_prompt)

    # TODO(lucasagomes): redact attachments content before sending to LLM
    # if attachments are provided, validate them
    if query_request.attachments:
        validate_attachments_metadata(query_request.attachments)

    agent, conversation_id, session_id = await get_agent(
        client,
        model_id,
        system_prompt,
        available_input_shields,
        available_output_shields,
        query_request.conversation_id,
        query_request.no_tools or False,
    )

    logger.debug("Conversation ID: %s, session ID: %s", conversation_id, session_id)
    # bypass tools and MCP servers if no_tools is True
    if query_request.no_tools:
        mcp_headers = {}
        agent.extra_headers = {}
        toolgroups = None
    else:
        # preserve compatibility when mcp_headers is not provided
        if mcp_headers is None:
            mcp_headers = {}
        mcp_headers = handle_mcp_headers_with_toolgroups(mcp_headers, configuration)
        if not mcp_headers and token:
            for mcp_server in configuration.mcp_servers:
                mcp_headers[mcp_server.url] = {
                    "Authorization": f"Bearer {token}",
                }

        agent.extra_headers = {
            "X-LlamaStack-Provider-Data": json.dumps(
                {
                    "mcp_headers": mcp_headers,
                }
            ),
        }

        vector_db_ids = [
            vector_db.identifier for vector_db in await client.vector_dbs.list()
        ]
        toolgroups = (get_rag_toolgroups(vector_db_ids) or []) + [
            mcp_server.name for mcp_server in configuration.mcp_servers
        ]
        # Convert empty list to None for consistency with existing behavior
        if not toolgroups:
            toolgroups = None

    response = await agent.create_turn(
        messages=[UserMessage(role="user", content=query_request.query)],
        session_id=session_id,
        documents=query_request.get_documents(),
        stream=False,
        toolgroups=toolgroups,
    )

    # Check for validation errors in the response
    steps = getattr(response, "steps", [])
    for step in steps:
        if step.step_type == "shield_call" and step.violation:
            # Metric for LLM validation errors
            metrics.llm_calls_validation_errors_total.inc()
            break

    return str(response.output_message.content), conversation_id  # type: ignore[union-attr]


def validate_attachments_metadata(attachments: list[Attachment]) -> None:
    """
    Validate attachment metadata in a request.
    
    Checks each Attachment's `attachment_type` and `content_type` against allowed values in
    constants. If any attachment has an invalid type or content type this function raises
    an HTTPException with status 422 and a detail containing `UNABLE_TO_PROCESS_RESPONSE`
    and a cause message.
    
    Parameters:
        attachments (list[Attachment]): List of attachments to validate.
    
    Raises:
        HTTPException: 422 Unprocessable Entity when an attachment's type or content type is invalid.
    """
    for attachment in attachments:
        if attachment.attachment_type not in constants.ATTACHMENT_TYPES:
            message = (
                f"Attachment with improper type {attachment.attachment_type} detected"
            )
            logger.error(message)
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail={
                    "response": constants.UNABLE_TO_PROCESS_RESPONSE,
                    "cause": message,
                },
            )
        if attachment.content_type not in constants.ATTACHMENT_CONTENT_TYPES:
            message = f"Attachment with improper content type {attachment.content_type} detected"
            logger.error(message)
            raise HTTPException(
                status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
                detail={
                    "response": constants.UNABLE_TO_PROCESS_RESPONSE,
                    "cause": message,
                },
            )


def construct_transcripts_path(user_id: str, conversation_id: str) -> Path:
    """
    Builds a filesystem path for storing transcripts for a given user and conversation.
    
    Both user_id and conversation_id are normalized (path-normalized and leading separators removed) to avoid producing absolute or traversal paths. The returned Path is rooted at the configured transcripts storage directory (configuration.user_data_collection_configuration.transcripts_storage) and appends the sanitized user_id and conversation_id.
     
    Parameters:
        user_id (str): User identifier used as a directory name (will be sanitized).
        conversation_id (str): Conversation identifier used as a directory name (will be sanitized).
    
    Returns:
        pathlib.Path: Path to the transcripts directory for the given user and conversation.
    """
    # these two normalizations are required by Snyk as it detects
    # this Path sanitization pattern
    uid = os.path.normpath("/" + user_id).lstrip("/")
    cid = os.path.normpath("/" + conversation_id).lstrip("/")
    file_path = (
        configuration.user_data_collection_configuration.transcripts_storage or ""
    )
    return Path(file_path, uid, cid)


def store_transcript(  # pylint: disable=too-many-arguments,too-many-positional-arguments
    user_id: str,
    conversation_id: str,
    query_is_valid: bool,
    query: str,
    query_request: QueryRequest,
    response: str,
    rag_chunks: list[str],
    truncated: bool,
    attachments: list[Attachment],
) -> None:
    """
    Store a transcript of a conversation to the configured local transcripts storage.
    
    Creates the transcripts directory for the given user and conversation if needed and writes a JSON file named with a unique id containing:
    - metadata (provider, model, user_id, conversation_id, UTC ISO8601 timestamp),
    - the redacted query, validation result, LLM response, RAG chunks, truncated flag,
    - serialized attachments.
    
    Parameters:
        user_id: User identifier (UUID) used to partition storage.
        conversation_id: Conversation identifier (UUID) used to partition storage.
        query_is_valid: Whether the query passed validation checks.
        query: Redacted user query text (attachments excluded).
        query_request: Original QueryRequest (used for provider/model metadata).
        response: LLM/agent response text to store.
        rag_chunks: List of RAG chunk strings included in the transcript.
        truncated: True if conversation history was truncated before the request.
        attachments: Attachments included in the request; each will be serialized via its model_dump method.
    
    Side effects:
        - Creates directories and writes a JSON file to disk under the configured transcripts path.
        - Logs the storage location on success.
    """
    transcripts_path = construct_transcripts_path(user_id, conversation_id)
    transcripts_path.mkdir(parents=True, exist_ok=True)

    data_to_store = {
        "metadata": {
            "provider": query_request.provider,
            "model": query_request.model,
            "user_id": user_id,
            "conversation_id": conversation_id,
            "timestamp": datetime.now(UTC).isoformat(),
        },
        "redacted_query": query,
        "query_is_valid": query_is_valid,
        "llm_response": response,
        "rag_chunks": rag_chunks,
        "truncated": truncated,
        "attachments": [attachment.model_dump() for attachment in attachments],
    }

    # stores feedback in a file under unique uuid
    transcript_file_path = transcripts_path / f"{get_suid()}.json"
    with open(transcript_file_path, "w", encoding="utf-8") as transcript_file:
        json.dump(data_to_store, transcript_file)

    logger.info("Transcript successfully stored at: %s", transcript_file_path)


def get_rag_toolgroups(
    vector_db_ids: list[str],
) -> list[Toolgroup] | None:
    """
    Return a RAG toolgroups list for the provided vector databases, or None if no databases are supplied.
    
    When `vector_db_ids` is non-empty the function returns a single Toolgroup configured for the built-in RAG knowledge search,
    with the list passed as the `vector_db_ids` argument. If `vector_db_ids` is empty, returns None.
    
    Parameters:
        vector_db_ids (list[str]): List of vector database identifiers to include in the RAG toolgroup.
    
    Returns:
        list[Toolgroup] | None: A one-element list containing the configured RAG Toolgroup, or None when `vector_db_ids` is empty.
    """
    return (
        [
            ToolgroupAgentToolGroupWithArgs(
                name="builtin::rag/knowledge_search",
                args={
                    "vector_db_ids": vector_db_ids,
                },
            )
        ]
        if vector_db_ids
        else None
    )
