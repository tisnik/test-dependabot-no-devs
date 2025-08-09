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
    Format a dictionary as a Server-Sent Events (SSE) data string.
    
    Parameters:
        d (dict): The data to be formatted as an SSE event.
    
    Returns:
        str: The formatted SSE data string.
    """
    data = json.dumps(d)
    return f"data: {data}\n\n"


def stream_start_event(conversation_id: str) -> str:
    """
    Format a Server-Sent Events (SSE) start event containing the conversation ID.
    
    Parameters:
        conversation_id (str): Unique identifier for the conversation.
    
    Returns:
        str: SSE-formatted string representing the start event.
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
    Format and return the end event for a streaming response, including referenced document metadata and placeholder token counts.
    
    Parameters:
        metadata_map (dict): A mapping containing metadata about referenced documents.
    
    Returns:
        str: A Server-Sent Events (SSE) formatted string representing the end of the data stream.
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
    Processes a streaming chunk from Llama Stack and yields formatted Server-Sent Events (SSE) strings.
    
    Depending on the event and step type in the chunk, dispatches to the appropriate handler to generate SSE events for turn lifecycle, inference, tool execution, shield checks, errors, or heartbeat signals. Updates the provided metadata map with information extracted from tool execution events.
    
    Parameters:
        chunk: A streaming response chunk from Llama Stack containing event data.
        chunk_id (int): The sequential identifier for the current chunk.
        metadata_map (dict): A dictionary for accumulating metadata from tool execution events.
    
    Returns:
        Iterator[str]: Yields SSE-formatted event strings for the client.
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
    Yield a formatted Server-Sent Events (SSE) error event containing the error message from a streaming chunk.
    
    Parameters:
        chunk_id (int): The unique identifier for the current streaming chunk.
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
    Yield a Server-Sent Event (SSE) token event indicating the start of a new conversation turn.
    
    Parameters:
        chunk_id (int): The unique identifier for the current chunk.
    
    Yields:
        str: SSE-formatted token event with an empty token to signal turn start.
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
    Yields a Server-Sent Event (SSE) indicating the completion of a conversation turn, including the full output message content.
    
    Parameters:
        chunk_id (int): The unique identifier for the current chunk.
    
    Yields:
        str: SSE-formatted string containing the turn completion event and output message content.
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
    Processes a shield event chunk and yields a formatted SSE token event indicating shield validation results.
    
    Yields a "No Violation" token if no violation is detected, or a violation message if a shield violation occurs. Increments validation error metrics when violations are present.
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
    Yield formatted Server-Sent Events (SSE) strings for inference step events during streaming.
    
    Processes inference-related streaming chunks, yielding SSE events for step start, text token deltas, and tool call deltas. Supports both string and ToolCall object tool calls.
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
    Processes tool execution events from a streaming chunk and yields formatted Server-Sent Events (SSE) strings.
    
    Handles both tool call initiation and completion, including tool call arguments, responses, and summaries. Extracts and updates document metadata from knowledge search tool responses when present.
    
    Parameters:
        chunk_id (int): Unique identifier for the current streaming chunk.
        metadata_map (dict): Dictionary to be updated with document metadata extracted from tool responses.
    
    Yields:
        str: SSE-formatted event strings representing tool call events and responses.
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
    Yield a heartbeat event as a Server-Sent Event (SSE) for the given chunk ID.
    
    Parameters:
        chunk_id (int): The identifier for the current streaming chunk.
    
    Yields:
        str: SSE-formatted heartbeat event string.
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
    Handles POST requests to the /streaming_query endpoint, streaming language model responses as Server-Sent Events (SSE).
    
    This endpoint receives a query request, authenticates the user, selects the appropriate model and provider, and streams incremental response events from the Llama Stack backend to the client. Events include start, token updates, tool calls, turn completions, errors, and end-of-stream metadata. Optionally stores the conversation transcript if enabled in configuration.
    
    Returns:
        StreamingResponse: An HTTP streaming response yielding SSE-formatted events for the query lifecycle.
    
    Raises:
        HTTPException: Returns HTTP 500 if unable to connect to the Llama Stack server.
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
            Asynchronously generates a stream of Server-Sent Events (SSE) representing incremental responses from a language model turn.
            
            Yields start, token, tool call, turn completion, and end events as SSE-formatted strings. Collects the complete response for transcript storage if enabled.
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
    Asynchronously retrieves a streaming response and conversation ID from the Llama Stack agent for a given user query.
    
    This function configures input/output shields, system prompt, and tool usage based on the request and environment. It prepares the agent with appropriate headers and toolgroups, validates attachments if present, and initiates a streaming turn with the user's query and any provided documents.
    
    Parameters:
        model_id (str): Identifier of the model to use for the query.
        query_request (QueryRequest): The user's query and associated metadata.
        token (str): Authentication token for downstream services.
        mcp_headers (dict[str, dict[str, str]], optional): Multi-cluster proxy headers for tool integrations.
    
    Returns:
        tuple: A tuple containing the streaming response object and the conversation ID.
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
