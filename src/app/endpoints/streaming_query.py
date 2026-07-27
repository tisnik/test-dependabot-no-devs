"""Handler for REST API call to provide answer to streaming query."""

import json
import logging
import re
from typing import Any, AsyncIterator

from cachetools import TTLCache  # type: ignore

from llama_stack_client import APIConnectionError
from llama_stack_client.lib.agents.agent import AsyncAgent  # type: ignore
from llama_stack_client import AsyncLlamaStackClient  # type: ignore
from llama_stack_client.types.shared.interleaved_content_item import TextContentItem
from llama_stack_client.types import UserMessage  # type: ignore

from fastapi import APIRouter, HTTPException, Request, Depends, status
from fastapi.responses import StreamingResponse

from auth import get_auth_dependency
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.requests import QueryRequest
from utils.endpoints import check_configuration_loaded, get_system_prompt
from utils.common import retrieve_user_id
from utils.mcp_headers import mcp_headers_dependency, handle_mcp_headers_with_toolgroups
from utils.suid import get_suid
from utils.types import GraniteToolParser

from app.endpoints.conversations import conversation_id_to_agent_id
from app.endpoints.query import (
    get_rag_toolgroups,
    is_transcripts_enabled,
    store_transcript,
    select_model_id,
    validate_attachments_metadata,
)

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["streaming_query"])
auth_dependency = get_auth_dependency()

# Global agent registry to persist agents across requests
_agent_cache: TTLCache[str, AsyncAgent] = TTLCache(maxsize=1000, ttl=3600)


async def get_agent(
    client: AsyncLlamaStackClient,
    model_id: str,
    system_prompt: str,
    available_shields: list[str],
    conversation_id: str | None,
) -> tuple[AsyncAgent, str]:
    """
    Retrieve an existing AsyncAgent for the given conversation ID or create a new one with session persistence.
    
    If a conversation ID is provided and an agent exists in the cache, it is reused; otherwise, a new agent is created with the specified model, system prompt, and shields, and a new session is initialized. The agent is cached for future reuse.
    
    Returns:
        A tuple containing the AsyncAgent instance and the associated conversation ID.
    """
    if conversation_id is not None:
        agent = _agent_cache.get(conversation_id)
        if agent:
            logger.debug("Reusing existing agent with key: %s", conversation_id)
            return agent, conversation_id

    logger.debug("Creating new agent")
    agent = AsyncAgent(
        client,  # type: ignore[arg-type]
        model=model_id,
        instructions=system_prompt,
        input_shields=available_shields if available_shields else [],
        tool_parser=GraniteToolParser.get_parser(model_id),
        enable_session_persistence=True,
    )
    conversation_id = await agent.create_session(get_suid())
    _agent_cache[conversation_id] = agent
    conversation_id_to_agent_id[conversation_id] = agent.agent_id
    return agent, conversation_id


METADATA_PATTERN = re.compile(r"\nMetadata: (\{.+})\n")


def format_stream_data(d: dict) -> str:
    """
    Formats a dictionary as a Server-Sent Events (SSE) data message.
    
    Parameters:
        d (dict): The data to be serialized and sent as an SSE event.
    
    Returns:
        str: The SSE-formatted string containing the JSON-serialized data.
    """
    data = json.dumps(d)
    return f"data: {data}\n\n"


def stream_start_event(conversation_id: str) -> str:
    """
    Return an SSE-formatted start event containing the conversation ID.
    
    Parameters:
    	conversation_id (str): Unique identifier for the conversation.
    
    Returns:
    	str: Server-Sent Events (SSE) formatted string signaling the start of the stream.
    """
    return format_stream_data(
        {
            "event": "start",
            "data": {
                "conversation_id": conversation_id,
            },
        }
    )


def stream_end_event(metadata_map: dict) -> str:
    """
    Return a Server-Sent Events (SSE) formatted end event containing referenced document metadata and placeholders for token counts and quotas.
    
    Parameters:
        metadata_map (dict): A dictionary containing metadata entries, from which referenced documents are extracted.
    
    Returns:
        str: An SSE-formatted string representing the end of the data stream, including referenced documents and placeholder fields.
    """
    return format_stream_data(
        {
            "event": "end",
            "data": {
                "referenced_documents": [
                    {
                        "doc_url": v["docs_url"],
                        "doc_title": v["title"],
                    }
                    for v in filter(
                        lambda v: ("docs_url" in v) and ("title" in v),
                        metadata_map.values(),
                    )
                ],
                "truncated": None,  # TODO(jboos): implement truncated
                "input_tokens": 0,  # TODO(jboos): implement input tokens
                "output_tokens": 0,  # TODO(jboos): implement output tokens
            },
            "available_quotas": {},  # TODO(jboos): implement available quotas
        }
    )


def stream_build_event(chunk: Any, chunk_id: int, metadata_map: dict) -> str | None:
    """
    Processes a streaming chunk from the Llama Stack response and formats it as a Server-Sent Event (SSE) data string.
    
    If the chunk represents model token progress, returns a token event with the generated text. If the chunk indicates completion of a tool execution step, extracts and stores document metadata from tool responses and returns a token event with the tool name. Returns None if the chunk does not contain relevant event data.
    
    Parameters:
        chunk_id (int): The sequential identifier for the current chunk.
        metadata_map (dict): A dictionary to accumulate extracted document metadata.
    
    Returns:
        str | None: SSE-formatted event string if processable data is present, otherwise None.
    """
    # pylint: disable=R1702
    if hasattr(chunk.event, "payload"):
        if chunk.event.payload.event_type == "step_progress":
            if hasattr(chunk.event.payload.delta, "text"):
                text = chunk.event.payload.delta.text
                return format_stream_data(
                    {
                        "event": "token",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": text,
                        },
                    }
                )
        if (
            chunk.event.payload.event_type == "step_complete"
            and chunk.event.payload.step_details.step_type == "tool_execution"
        ):
            for r in chunk.event.payload.step_details.tool_responses:
                if r.tool_name == "knowledge_search" and r.content:
                    for text_content_item in r.content:
                        if isinstance(text_content_item, TextContentItem):
                            for match in METADATA_PATTERN.findall(
                                text_content_item.text
                            ):
                                meta = json.loads(match.replace("'", '"'))
                                metadata_map[meta["document_id"]] = meta
            if chunk.event.payload.step_details.tool_calls:
                tool_name = str(
                    chunk.event.payload.step_details.tool_calls[0].tool_name
                )
                return format_stream_data(
                    {
                        "event": "token",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": tool_name,
                        },
                    }
                )
    return None


@router.post("/streaming_query")
async def streaming_query_endpoint_handler(
    _request: Request,
    query_request: QueryRequest,
    auth: Any = Depends(auth_dependency),
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> StreamingResponse:
    """
    Handles POST requests to the /streaming_query endpoint, streaming LLM agent responses as Server-Sent Events (SSE).
    
    Receives a query request, manages agent session retrieval or creation, and streams generated tokens and metadata to the client in real time. Stores the full transcript after completion if transcript collection is enabled. Returns a StreamingResponse that yields SSE-formatted events for the start, each token, and the end of the stream.
    
    Raises:
        HTTPException: If unable to connect to the Llama Stack server.
    """
    check_configuration_loaded(configuration)

    llama_stack_config = configuration.llama_stack_configuration
    logger.info("LLama stack config: %s", llama_stack_config)

    _user_id, _user_name, token = auth

    try:
        # try to get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()
        model_id = select_model_id(await client.models.list(), query_request)
        response, conversation_id = await retrieve_response(
            client,
            model_id,
            query_request,
            token,
            mcp_headers=mcp_headers,
        )
        metadata_map: dict[str, dict[str, Any]] = {}

        async def response_generator(turn_response: Any) -> AsyncIterator[str]:
            """
            Asynchronously generates a stream of Server-Sent Events (SSE) for a model turn response.
            
            Yields SSE-formatted start, token, and end events as the model produces output tokens. Accumulates the complete response text and, if transcript storage is enabled, saves the transcript after streaming completes.
            
            Yields:
                str: SSE-formatted event strings for each stage of the streaming response.
            """
            chunk_id = 0
            complete_response = ""

            # Send start event
            yield stream_start_event(conversation_id)

            async for chunk in turn_response:
                if event := stream_build_event(chunk, chunk_id, metadata_map):
                    complete_response += json.loads(event.replace("data: ", ""))[
                        "data"
                    ]["token"]
                    chunk_id += 1
                    yield event

            yield stream_end_event(metadata_map)

            if not is_transcripts_enabled():
                logger.debug("Transcript collection is disabled in the configuration")
            else:
                store_transcript(
                    user_id=retrieve_user_id(auth),
                    conversation_id=conversation_id,
                    query_is_valid=True,  # TODO(lucasagomes): implement as part of query validation
                    query=query_request.query,
                    query_request=query_request,
                    response=complete_response,
                    rag_chunks=[],  # TODO(lucasagomes): implement rag_chunks
                    truncated=False,  # TODO(lucasagomes): implement truncation as part
                    # of quota work
                    attachments=query_request.attachments or [],
                )

        return StreamingResponse(response_generator(response))
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


async def retrieve_response(
    client: AsyncLlamaStackClient,
    model_id: str,
    query_request: QueryRequest,
    token: str,
    mcp_headers: dict[str, dict[str, str]] | None = None,
) -> tuple[Any, str]:
    """
    Asynchronously retrieves a streaming response from an LLM agent for a given user query, preparing agent session, toolgroups, and authorization headers as needed.
    
    Parameters:
        model_id (str): Identifier of the LLM model to use.
        query_request (QueryRequest): The user's query and associated metadata.
        token (str): User authentication token.
        mcp_headers (dict[str, dict[str, str]], optional): Additional headers for MCP servers.
    
    Returns:
        tuple: A tuple containing the streaming response iterator and the conversation ID.
    """
    available_shields = [shield.identifier for shield in await client.shields.list()]
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

    agent, conversation_id = await get_agent(
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

    logger.debug("Session ID: %s", conversation_id)
    vector_db_ids = [
        vector_db.identifier for vector_db in await client.vector_dbs.list()
    ]
    toolgroups = (get_rag_toolgroups(vector_db_ids) or []) + [
        mcp_server.name for mcp_server in configuration.mcp_servers
    ]
    response = await agent.create_turn(
        messages=[UserMessage(role="user", content=query_request.query)],
        session_id=conversation_id,
        documents=query_request.get_documents(),
        stream=True,
        toolgroups=toolgroups or None,
    )

    return response, conversation_id
