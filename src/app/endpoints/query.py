"""Handler for REST API call to provide answer to query."""

from datetime import datetime, UTC
import json
import logging
import os
from pathlib import Path
from typing import Any

from cachetools import TTLCache  # type: ignore

from llama_stack_client.lib.agents.agent import Agent
from llama_stack_client import APIConnectionError
from llama_stack_client import LlamaStackClient  # type: ignore
from llama_stack_client.types import UserMessage  # type: ignore
from llama_stack_client.types.agents.turn_create_params import (
    ToolgroupAgentToolGroupWithArgs,
    Toolgroup,
)
from llama_stack_client.types.model_list_response import ModelListResponse

from fastapi import APIRouter, HTTPException, status, Depends

from client import LlamaStackClientHolder
from configuration import configuration
from app.endpoints.conversations import conversation_id_to_agent_id
from models.responses import QueryResponse, UnauthorizedResponse, ForbiddenResponse
from models.requests import QueryRequest, Attachment
import constants
from auth import get_auth_dependency
from utils.common import retrieve_user_id
from utils.endpoints import check_configuration_loaded, get_system_prompt
from utils.mcp_headers import mcp_headers_dependency, handle_mcp_headers_with_toolgroups
from utils.suid import get_suid
from utils.types import GraniteToolParser

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["query"])
auth_dependency = get_auth_dependency()

# Global agent registry to persist agents across requests
_agent_cache: TTLCache[str, Agent] = TTLCache(maxsize=1000, ttl=3600)

query_response: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "response": "LLM ansert",
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
    """
    Returns True if transcript storage is enabled in the configuration; otherwise, returns False.
    """
    return not configuration.user_data_collection_configuration.transcripts_disabled


def get_agent(
    client: LlamaStackClient,
    model_id: str,
    system_prompt: str,
    available_shields: list[str],
    conversation_id: str | None,
) -> tuple[Agent, str]:
    """
    Retrieve an existing agent for the given conversation ID or create a new agent with session persistence.
    
    If a conversation ID is provided and an agent exists in the cache, returns the cached agent and conversation ID. Otherwise, creates a new agent configured with the specified model, system prompt, and shields, enables session persistence, generates a new session ID, and caches the agent for future use.
    
    Returns:
        A tuple containing the agent instance and its associated conversation ID.
    """
    if conversation_id is not None:
        agent = _agent_cache.get(conversation_id)
        if agent:
            logger.debug("Reusing existing agent with key: %s", conversation_id)
            return agent, conversation_id

    logger.debug("Creating new agent")
    # TODO(lucasagomes): move to ReActAgent
    agent = Agent(
        client,
        model=model_id,
        instructions=system_prompt,
        input_shields=available_shields if available_shields else [],
        tool_parser=GraniteToolParser.get_parser(model_id),
        enable_session_persistence=True,
    )
    conversation_id = agent.create_session(get_suid())
    _agent_cache[conversation_id] = agent
    conversation_id_to_agent_id[conversation_id] = agent.agent_id

    return agent, conversation_id


@router.post("/query", responses=query_response)
def query_endpoint_handler(
    query_request: QueryRequest,
    auth: Any = Depends(auth_dependency),
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> QueryResponse:
    """
    Processes a POST request to the /query endpoint, generating a response to a user query using a language model or agent.
    
    Authenticates the user, selects an appropriate model, retrieves a response from the Llama Stack, and optionally stores a transcript of the interaction. Returns the generated response and conversation ID. Raises an HTTP 500 error if unable to connect to the Llama Stack server.
    """
    check_configuration_loaded(configuration)

    llama_stack_config = configuration.llama_stack_configuration
    logger.info("LLama stack config: %s", llama_stack_config)

    _user_id, _user_name, token = auth

    try:
        # try to get Llama Stack client
        client = LlamaStackClientHolder().get_client()
        model_id = select_model_id(client.models.list(), query_request)
        response, conversation_id = retrieve_response(
            client,
            model_id,
            query_request,
            token,
            mcp_headers=mcp_headers,
        )

        if not is_transcripts_enabled():
            logger.debug("Transcript collection is disabled in the configuration")
        else:
            store_transcript(
                user_id=retrieve_user_id(auth),
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
        logger.error("Unable to connect to Llama Stack: %s", e)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": "Unable to connect to Llama Stack",
                "cause": str(e),
            },
        ) from e


def select_model_id(models: ModelListResponse, query_request: QueryRequest) -> str:
    """
    Selects and returns a model ID for processing a query based on the request or available models.
    
    If the request does not specify a model, selects the first available LLM model. Raises an HTTP 400 error if no suitable model is found or if the specified model and provider are not available.
    """
    model_id = query_request.model
    provider_id = query_request.provider

    # TODO(lucasagomes): support default model selection via configuration
    if not model_id:
        logger.info("No model specified in request, using the first available LLM")
        try:
            model = next(
                m
                for m in models
                if m.model_type == "llm"  # pyright: ignore[reportAttributeAccessIssue]
            ).identifier
            logger.info("Selected model: %s", model)
            return model
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

    logger.info("Searching for model: %s, provider: %s", model_id, provider_id)
    if not any(
        m.identifier == model_id and m.provider_id == provider_id for m in models
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

    return model_id


def retrieve_response(
    client: LlamaStackClient,
    model_id: str,
    query_request: QueryRequest,
    token: str,
    mcp_headers: dict[str, dict[str, str]] | None = None,
) -> tuple[str, str]:
    """
    Retrieves a response to a user query from a language model or agent, managing session context, toolgroups, and optional MCP headers.
    
    Parameters:
        model_id (str): The identifier of the language model to use.
        query_request (QueryRequest): The user's query and related request data.
        token (str): Authentication token for MCP servers.
        mcp_headers (dict[str, dict[str, str]], optional): Additional headers for MCP servers.
    
    Returns:
        tuple[str, str]: The generated response content and the conversation ID associated with the session.
    """
    available_shields = [shield.identifier for shield in client.shields.list()]
    if not available_shields:
        logger.info("No available shields. Disabling safety")
    else:
        logger.info("Available shields found: %s", available_shields)

    # use system prompt from request or default one
    system_prompt = get_system_prompt(query_request, configuration)
    logger.debug("Using system prompt: %s", system_prompt)

    # TODO(lucasagomes): redact attachments content before sending to LLM
    # if attachments are provided, validate them
    if query_request.attachments:
        validate_attachments_metadata(query_request.attachments)

    agent, conversation_id = get_agent(
        client,
        model_id,
        system_prompt,
        available_shields,
        query_request.conversation_id,
    )

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

    vector_db_ids = [vector_db.identifier for vector_db in client.vector_dbs.list()]
    toolgroups = (get_rag_toolgroups(vector_db_ids) or []) + [
        mcp_server.name for mcp_server in configuration.mcp_servers
    ]
    response = agent.create_turn(
        messages=[UserMessage(role="user", content=query_request.query)],
        session_id=conversation_id,
        documents=query_request.get_documents(),
        stream=False,
        toolgroups=toolgroups or None,
    )

    return str(response.output_message.content), conversation_id  # type: ignore[union-attr]


def validate_attachments_metadata(attachments: list[Attachment]) -> None:
    """
    Validates that each attachment in the request has an allowed type and content type.
    
    Raises:
        HTTPException: If any attachment has an invalid type or content type.
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
    Constructs a filesystem path for storing transcripts for a given user and conversation.
    
    Parameters:
        user_id (str): The user identifier.
        conversation_id (str): The conversation identifier.
    
    Returns:
        Path: The constructed path combining the transcripts storage directory, user ID, and conversation ID.
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
    Store a transcript of a query and its response to the local filesystem for a given user and conversation.
    
    The transcript includes metadata, the redacted query, validation status, LLM response, RAG chunks, truncation flag, and attachments. The data is saved as a JSON file under a directory path constructed from the user and conversation IDs.
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
    Generate a RAG knowledge search toolgroup if vector database IDs are provided.
    
    Parameters:
        vector_db_ids (list[str]): List of vector database IDs to include in the toolgroup.
    
    Returns:
        list[Toolgroup] | None: A list containing a RAG toolgroup if vector_db_ids is non-empty; otherwise, None.
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
