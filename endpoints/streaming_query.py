"""Handler for REST API call to provide answer to streaming query."""

import ast
import json
import re
import logging
from typing import Annotated, Any, AsyncIterator, Iterator

from llama_stack_client import APIConnectionError
from llama_stack_client import AsyncLlamaStackClient  # type: ignore
from llama_stack_client.types import UserMessage  # type: ignore

from llama_stack_client.lib.agents.event_logger import interleaved_content_as_str
from llama_stack_client.types.shared import ToolCall
from llama_stack_client.types.shared.interleaved_content_item import TextContentItem

from fastapi import APIRouter, HTTPException, Request, Depends, status
from fastapi.responses import StreamingResponse

from auth import get_auth_dependency
from auth.interface import AuthTuple
from client import AsyncLlamaStackClientHolder
from configuration import configuration
import metrics
from models.requests import QueryRequest
from utils.endpoints import check_configuration_loaded, get_agent, get_system_prompt
from utils.mcp_headers import mcp_headers_dependency, handle_mcp_headers_with_toolgroups

from app.endpoints.query import (
    get_rag_toolgroups,
    is_input_shield,
    is_output_shield,
    is_transcripts_enabled,
    store_transcript,
    select_model_and_provider_id,
    validate_attachments_metadata,
)

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["streaming_query"])
auth_dependency = get_auth_dependency()


METADATA_PATTERN = re.compile(r"\nMetadata: (\{.+})\n")


def format_stream_data(d: dict) -> str:
    """
    Format a dictionary as a Server-Sent Events (SSE) data frame.
    
    The input dictionary is JSON-encoded and prefixed with "data: ", followed by a blank line to terminate the SSE frame.
    
    Parameters:
        d (dict): JSON-serializable mapping to send as the event payload.
    
    Returns:
        str: An SSE-formatted string in the form "data: <json>\n\n".
    """
    data = json.dumps(d)
    return f"data: {data}\n\n"


def stream_start_event(conversation_id: str) -> str:
    """
    Create the Server-Sent Event (SSE) "start" frame containing the conversation identifier.
    
    Parameters:
        conversation_id (str): Conversation UUID to include in the event payload.
    
    Returns:
        str: SSE-formatted data frame (prefixed with "data: " and terminated by a blank line).
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
    Builds the final Server-Sent Event signaling stream completion, including referenced document summaries and token/quota placeholders.
    
    Parameters:
        metadata_map (dict): Mapping of document identifiers to metadata dictionaries. Entries that contain both "docs_url" and "title" will be included in the returned event's `referenced_documents` list.
    
    Returns:
        str: A single SSE-formatted data frame (prefixed with "data: " and terminated by a blank line) representing the "end" event. The event's `data` contains `referenced_documents`, and placeholder fields for `truncated`, `input_tokens`, `output_tokens`, and `available_quotas`.
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


def stream_build_event(chunk: Any, chunk_id: int, metadata_map: dict) -> Iterator[str]:
    """
    Convert a Llama Stack streaming chunk into one or more Server-Sent Event (SSE) frames.
    
    This generator inspects the incoming streaming `chunk` and dispatches it to the appropriate handler
    based on the chunk's event and step types, yielding SSE-formatted strings for each emitted event.
    It supports error, turn lifecycle, shield, inference, tool execution, and heartbeat events.
    
    Parameters:
        chunk: A streaming chunk object from the Llama Stack API. The function reads attributes such as
            `error`, `event.payload.event_type`, and `event.payload.step_type` to determine handling.
        chunk_id: Integer identifier for the chunk (used in emitted events).
        metadata_map: Mutable mapping used to accumulate extracted document metadata from tool responses;
            this map may be updated inside tool-execution handling.
    
    Returns:
        Iterator[str]: SSE data frames (as strings) produced for the given chunk.
    """
    if hasattr(chunk, "error"):
        yield from _handle_error_event(chunk, chunk_id)
        return

    event_type = chunk.event.payload.event_type
    step_type = getattr(chunk.event.payload, "step_type", None)

    if event_type in {"turn_start", "turn_awaiting_input"}:
        yield from _handle_turn_start_event(chunk_id)
    elif event_type == "turn_complete":
        yield from _handle_turn_complete_event(chunk, chunk_id)
    elif step_type == "shield_call":
        yield from _handle_shield_event(chunk, chunk_id)
    elif step_type == "inference":
        yield from _handle_inference_event(chunk, chunk_id)
    elif step_type == "tool_execution":
        yield from _handle_tool_execution_event(chunk, chunk_id, metadata_map)
    else:
        yield from _handle_heartbeat_event(chunk_id)


# -----------------------------------
# Error handling
# -----------------------------------
def _handle_error_event(chunk: Any, chunk_id: int) -> Iterator[str]:
    """
    Yield a Server-Sent Event (SSE) "error" frame containing the chunk id and its error message.
    
    Parameters:
        chunk: An object with an `error` mapping containing a `"message"` key; the value is used as the token in the SSE payload.
        chunk_id (int): Numeric identifier for the streaming chunk.
    
    Yields:
        str: A formatted SSE data frame (JSON encoded) representing an "error" event with fields `id` and `token`.
    """
    yield format_stream_data(
        {
            "event": "error",
            "data": {
                "id": chunk_id,
                "token": chunk.error["message"],
            },
        }
    )


# -----------------------------------
# Turn handling
# -----------------------------------
def _handle_turn_start_event(chunk_id: int) -> Iterator[str]:
    """
    Yield a Server-Sent Event 'token' frame indicating the start of a conversation turn.
    
    Yields a single SSE-formatted string that contains an event of type "token" with an empty token value and the provided chunk_id. This signals the client that a new turn has begun.
    """
    yield format_stream_data(
        {
            "event": "token",
            "data": {
                "id": chunk_id,
                "token": "",
            },
        }
    )


def _handle_turn_complete_event(chunk: Any, chunk_id: int) -> Iterator[str]:
    """
    Yield a Server-Sent Event signaling a completed turn with the final token text.
    
    Yields:
        An SSE-formatted string (via format_stream_data) containing:
          - event: "turn_complete"
          - data.id: the provided chunk_id
          - data.token: the final output message content extracted from the chunk (converted to string with interleaved_content_as_str)
    """
    yield format_stream_data(
        {
            "event": "turn_complete",
            "data": {
                "id": chunk_id,
                "token": interleaved_content_as_str(
                    chunk.event.payload.turn.output_message.content
                ),
            },
        }
    )


# -----------------------------------
# Shield handling
# -----------------------------------
def _handle_shield_event(chunk: Any, chunk_id: int) -> Iterator[str]:
    """
    Handle a completed shield step by yielding SSE `token` events that indicate whether a shield violation occurred.
    
    If the shield step has no violation, yields a `token` event with token "No Violation". If a violation exists, increments the LLM validation error metric and yields a `token` event containing a concise violation message (user-facing message plus metadata).
    
    Parameters:
        chunk: Streaming chunk object from the Llama Stack whose `event.payload` contains `event_type`, `step_type`, and `step_details.violation`.
        chunk_id (int): Sequential identifier for the chunk used in emitted event payloads.
    
    Yields:
        str: SSE-formatted JSON frames (via format_stream_data) representing `token` events.
    
    Side effects:
        - Increments `metrics.llm_calls_validation_errors_total` when a violation is present.
    """
    if chunk.event.payload.event_type == "step_complete":
        violation = chunk.event.payload.step_details.violation
        if not violation:
            yield format_stream_data(
                {
                    "event": "token",
                    "data": {
                        "id": chunk_id,
                        "role": chunk.event.payload.step_type,
                        "token": "No Violation",
                    },
                }
            )
        else:
            # Metric for LLM validation errors
            metrics.llm_calls_validation_errors_total.inc()
            violation = (
                f"Violation: {violation.user_message} (Metadata: {violation.metadata})"
            )
            yield format_stream_data(
                {
                    "event": "token",
                    "data": {
                        "id": chunk_id,
                        "role": chunk.event.payload.step_type,
                        "token": violation,
                    },
                }
            )


# -----------------------------------
# Inference handling
# -----------------------------------
def _handle_inference_event(chunk: Any, chunk_id: int) -> Iterator[str]:
    """
    Handle inference-related streaming chunks and yield Server-Sent Event (SSE) frames.
    
    Processes inference `chunk`s from the model's streaming API and yields formatted SSE data frames as strings. Supported behaviors:
    - On `step_start`: yield a "token" event with an empty token to signal the start of a generation step.
    - On `step_progress`:
      - If the delta is a tool call:
        - If the tool call is a string, yield a "tool_call" event with that string as the token.
        - If the tool call is a ToolCall object, yield a "tool_call" event with the ToolCall's `tool_name`.
      - If the delta is text, yield a "token" event containing the text delta.
    
    Parameters:
        chunk: The streaming chunk object containing `event.payload.event_type`, `payload.step_type`, and `payload.delta`.
        chunk_id: Integer identifier for the chunk; included in yielded event data.
    
    Returns:
        Iterator[str]: An iterator of SSE-formatted strings (each created by `format_stream_data`) representing incremental events.
    """
    if chunk.event.payload.event_type == "step_start":
        yield format_stream_data(
            {
                "event": "token",
                "data": {
                    "id": chunk_id,
                    "role": chunk.event.payload.step_type,
                    "token": "",
                },
            }
        )

    elif chunk.event.payload.event_type == "step_progress":
        if chunk.event.payload.delta.type == "tool_call":
            if isinstance(chunk.event.payload.delta.tool_call, str):
                yield format_stream_data(
                    {
                        "event": "tool_call",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": chunk.event.payload.delta.tool_call,
                        },
                    }
                )
            elif isinstance(chunk.event.payload.delta.tool_call, ToolCall):
                yield format_stream_data(
                    {
                        "event": "tool_call",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": chunk.event.payload.delta.tool_call.tool_name,
                        },
                    }
                )

        elif chunk.event.payload.delta.type == "text":
            yield format_stream_data(
                {
                    "event": "token",
                    "data": {
                        "id": chunk_id,
                        "role": chunk.event.payload.step_type,
                        "token": chunk.event.payload.delta.text,
                    },
                }
            )


# -----------------------------------
# Tool Execution handling
# -----------------------------------
# pylint: disable=R1702,R0912
def _handle_tool_execution_event(
    chunk: Any, chunk_id: int, metadata_map: dict
) -> Iterator[str]:
    """
    Handle tool-execution-related streaming chunks and yield SSE-formatted events.
    
    Processes a streaming chunk representing a tool execution step and emits one or more
    Server-Sent Events (SSE) frames (as JSON strings produced by format_stream_data)
    describing tool calls and tool responses. Behavior varies by step event type:
    
    - On "step_start": emits a "tool_call" event with an empty token to indicate the tool call began.
    - On "step_complete": for each recorded tool call emits a "tool_call" event with the tool name and arguments;
      for each tool response emits a "tool_call" event containing:
        - For "query_from_memory": a short message reporting bytes fetched from memory.
        - For "knowledge_search": a short summary (first line) and extracts any embedded document metadata
          (parsed via ast.literal_eval) into metadata_map keyed by document_id. Parsing failures are logged at debug level.
        - For any other tool: the response content converted to a string.
    
    Parameters:
        chunk: The streaming chunk object from the Llama Stack protocol containing event, payload,
            step_details, and tool response/call entries.
        chunk_id (int): An identifier for the chunk used in emitted event payloads.
        metadata_map (dict): Mutable mapping updated in-place with metadata extracted from
            knowledge search results (keys are document IDs).
    
    Returns:
        Iterator[str]: Yields SSE-formatted JSON strings describing tool calls and responses.
    """
    if chunk.event.payload.event_type == "step_start":
        yield format_stream_data(
            {
                "event": "tool_call",
                "data": {
                    "id": chunk_id,
                    "role": chunk.event.payload.step_type,
                    "token": "",
                },
            }
        )

    elif chunk.event.payload.event_type == "step_complete":
        for t in chunk.event.payload.step_details.tool_calls:
            yield format_stream_data(
                {
                    "event": "tool_call",
                    "data": {
                        "id": chunk_id,
                        "role": chunk.event.payload.step_type,
                        "token": {
                            "tool_name": t.tool_name,
                            "arguments": t.arguments,
                        },
                    },
                }
            )

        for r in chunk.event.payload.step_details.tool_responses:
            if r.tool_name == "query_from_memory":
                inserted_context = interleaved_content_as_str(r.content)
                yield format_stream_data(
                    {
                        "event": "tool_call",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": {
                                "tool_name": r.tool_name,
                                "response": f"Fetched {len(inserted_context)} bytes from memory",
                            },
                        },
                    }
                )

            elif r.tool_name == "knowledge_search" and r.content:
                summary = ""
                for i, text_content_item in enumerate(r.content):
                    if isinstance(text_content_item, TextContentItem):
                        if i == 0:
                            summary = text_content_item.text
                            newline_pos = summary.find("\n")
                            if newline_pos > 0:
                                summary = summary[:newline_pos]
                        for match in METADATA_PATTERN.findall(text_content_item.text):
                            try:
                                meta = ast.literal_eval(match)
                                if "document_id" in meta:
                                    metadata_map[meta["document_id"]] = meta
                            except Exception:  # pylint: disable=broad-except
                                logger.debug(
                                    "An exception was thrown in processing %s",
                                    match,
                                )

                yield format_stream_data(
                    {
                        "event": "tool_call",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": {
                                "tool_name": r.tool_name,
                                "summary": summary,
                            },
                        },
                    }
                )

            else:
                yield format_stream_data(
                    {
                        "event": "tool_call",
                        "data": {
                            "id": chunk_id,
                            "role": chunk.event.payload.step_type,
                            "token": {
                                "tool_name": r.tool_name,
                                "response": interleaved_content_as_str(r.content),
                            },
                        },
                    }
                )


# -----------------------------------
# Catch-all for everything else
# -----------------------------------
def _handle_heartbeat_event(chunk_id: int) -> Iterator[str]:
    """
    Yield a Server-Sent Event (SSE) heartbeat message for the given chunk.
    
    Yields:
        str: A single SSE-formatted data frame (JSON encoded) representing a "heartbeat" event with fields:
             - id: the provided chunk_id
             - token: the literal string "heartbeat"
    """
    yield format_stream_data(
        {
            "event": "heartbeat",
            "data": {
                "id": chunk_id,
                "token": "heartbeat",
            },
        }
    )


@router.post("/streaming_query")
async def streaming_query_endpoint_handler(
    _request: Request,
    query_request: QueryRequest,
    auth: Annotated[AuthTuple, Depends(auth_dependency)],
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> StreamingResponse:
    """
    Handle the POST /streaming_query endpoint and stream model responses as Server-Sent Events (SSE).
    
    Streams incremental events produced by a Llama Stack model back to the client. The returned StreamingResponse yields SSE frames for:
    - a "start" event containing the conversation id,
    - streaming events produced by the model (tokens, tool calls, shield notifications, turn lifecycle events, heartbeats, and errors),
    - a final "end" event summarizing referenced documents and token usage placeholders.
    
    Side effects:
    - Increments LLM call metrics (total and, on connection failure, failures).
    - Optionally persists a transcript after the stream completes when transcript collection is enabled.
    
    Parameters:
        query_request (QueryRequest): The user's query payload (includes the query text and flags such as `no_tools` and any attachments).
    
    Returns:
        StreamingResponse: An ASGI streaming response that emits SSE-formatted JSON events.
    
    Raises:
        HTTPException: Raised with status 500 if the service cannot connect to the Llama Stack backend (wraps APIConnectionError).
    """
    check_configuration_loaded(configuration)

    llama_stack_config = configuration.llama_stack_configuration
    logger.info("LLama stack config: %s", llama_stack_config)

    user_id, _user_name, token = auth

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
        metadata_map: dict[str, dict[str, Any]] = {}

        async def response_generator(turn_response: Any) -> AsyncIterator[str]:
            """
            Generate Server-Sent Events (SSE) frames from an asynchronous streaming model response.
            
            This async generator:
            - Yields an initial "start" SSE event with the current conversation identifier.
            - Iterates over chunks from `turn_response`, converts each chunk into one or more SSE events using `stream_build_event`, and yields them in order.
            - Tracks the most recent complete model response from `turn_complete` events and includes it when storing transcripts.
            - Yields a final "end" SSE event containing aggregated metadata.
            
            Side effects:
            - May store a transcript (via `store_transcript`) after streaming completes when transcripts are enabled.
            - Reads and updates outer-scope variables such as `conversation_id`, `metadata_map`, and `user_id`.
            
            Parameters:
            - turn_response: An asynchronous iterable of streaming chunks returned by the model/agent. Each chunk is transformed into SSE event strings.
            
            Returns:
            - An async iterator that yields SSE-formatted strings ready to be sent to clients.
            """
            chunk_id = 0
            complete_response = "No response from the model"

            # Send start event
            yield stream_start_event(conversation_id)

            async for chunk in turn_response:
                for event in stream_build_event(chunk, chunk_id, metadata_map):
                    if (
                        json.loads(event.replace("data: ", ""))["event"]
                        == "turn_complete"
                    ):
                        complete_response = json.loads(event.replace("data: ", ""))[
                            "data"
                        ]["token"]
                    chunk_id += 1
                    yield event

            yield stream_end_event(metadata_map)

            if not is_transcripts_enabled():
                logger.debug("Transcript collection is disabled in the configuration")
            else:
                store_transcript(
                    user_id=user_id,
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

        # Update metrics for the LLM call
        metrics.llm_calls_total.labels(provider_id, model_id).inc()

        return StreamingResponse(response_generator(response))
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


async def retrieve_response(
    client: AsyncLlamaStackClient,
    model_id: str,
    query_request: QueryRequest,
    token: str,
    mcp_headers: dict[str, dict[str, str]] | None = None,
) -> tuple[Any, str]:
    """
    Retrieve a streaming turn response from the Llama Stack client and the associated conversation ID.
    
    This prepares agent configuration (shields, system prompt, toolgroups, and MCP headers), validates attachments when present, and starts a streaming turn on the resolved agent.
    
    Parameters:
        model_id (str): Model identifier to use on the Llama Stack provider.
        query_request (QueryRequest): The client request object containing the user query, optional attachments,
            conversation_id, and flags such as `no_tools`.
        token (str): Optional bearer token used to populate MCP headers when not provided.
        mcp_headers (dict[str, dict[str, str]] | None): Optional mapping of MCP server URLs to header dicts
            that will be forwarded to the provider; may be augmented or normalized by the function.
    
    Returns:
        tuple[Any, str]: A pair of (streaming_response, conversation_id). `streaming_response` is the
        asynchronous streaming turn object returned by the agent (iterable/async iterable of chunks),
        and `conversation_id` is the identifier for the conversation created or reused.
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
        stream=True,
        toolgroups=toolgroups,
    )

    return response, conversation_id
