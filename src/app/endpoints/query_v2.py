"""Handler for REST API call to provide answer to query using Response API."""

import logging
from typing import Annotated, Any, cast

from llama_stack_client import AsyncLlamaStackClient  # type: ignore
from llama_stack.apis.agents.openai_responses import (
    OpenAIResponseObject,
)

from fastapi import APIRouter, Request, Depends

from app.endpoints.query import (
    query_endpoint_handler_base,
    validate_attachments_metadata,
)
from constants import DEFAULT_RAG_TOOL
from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from configuration import configuration
import metrics
from models.config import Action
from models.requests import QueryRequest
from models.responses import (
    ForbiddenResponse,
    QueryResponse,
    ReferencedDocument,
    UnauthorizedResponse,
    QuotaExceededResponse,
)
from utils.endpoints import (
    get_system_prompt,
    get_topic_summary_system_prompt,
)
from utils.mcp_headers import mcp_headers_dependency
from utils.token_counter import TokenCounter
from utils.types import TurnSummary, ToolCallSummary

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["query_v2"])

query_v2_response: dict[int | str, dict[str, Any]] = {
    200: {
        "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
        "response": "LLM answer",
        "referenced_documents": [
            {
                "doc_url": "https://docs.openshift.com/"
                "container-platform/4.15/operators/olm/index.html",
                "doc_title": "Operator Lifecycle Manager (OLM)",
            }
        ],
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "Client does not have permission to access conversation",
        "model": ForbiddenResponse,
    },
    429: {
        "description": "The quota has been exceeded",
        "model": QuotaExceededResponse,
    },
    500: {
        "detail": {
            "response": "Unable to connect to Llama Stack",
            "cause": "Connection error.",
        }
    },
}


def _extract_text_from_response_output_item(output_item: Any) -> str:
    """
    Extract the assistant's message text from a Responses API output item.
    
    Parameters:
        output_item (Any): An output item from the Responses API (message-like object or dict).
    
    Returns:
        str: The concatenated assistant message text, or an empty string if the item is not an assistant message or contains no text.
    """
    if getattr(output_item, "type", None) != "message":
        return ""
    if getattr(output_item, "role", None) != "assistant":
        return ""

    content = getattr(output_item, "content", None)
    if isinstance(content, str):
        return content

    text_fragments: list[str] = []
    if isinstance(content, list):
        for part in content:
            if isinstance(part, str):
                text_fragments.append(part)
                continue
            text_value = getattr(part, "text", None)
            if text_value:
                text_fragments.append(text_value)
                continue
            refusal = getattr(part, "refusal", None)
            if refusal:
                text_fragments.append(refusal)
                continue
            if isinstance(part, dict):
                dict_text = part.get("text") or part.get("refusal")
                if dict_text:
                    text_fragments.append(str(dict_text))

    return "".join(text_fragments)


def _build_tool_call_summary(  # pylint: disable=too-many-return-statements,too-many-branches
    output_item: Any,
) -> ToolCallSummary | None:
    """
    Translate a Responses API output item into a ToolCallSummary when the item represents a tool invocation.
    
    Parameters:
        output_item (Any): An output item from a Responses API `response.output` entry.
    
    Returns:
        ToolCallSummary | None: A ToolCallSummary for supported tool item types (`function_call`, `file_search_call`, `web_search_call`, `mcp_call`, `mcp_list_tools`, `mcp_approval_request`); returns `None` if the item type is not a recognized tool output.
    """
    item_type = getattr(output_item, "type", None)

    if item_type == "function_call":
        parsed_arguments = getattr(output_item, "arguments", "")
        status = getattr(output_item, "status", None)
        if status:
            if isinstance(parsed_arguments, dict):
                args: Any = {**parsed_arguments, "status": status}
            else:
                args = {"arguments": parsed_arguments, "status": status}
        else:
            args = parsed_arguments

        call_id = getattr(output_item, "id", None) or getattr(
            output_item, "call_id", None
        )
        return ToolCallSummary(
            id=str(call_id),
            name=getattr(output_item, "name", "function_call"),
            args=args,
            response=None,
        )

    if item_type == "file_search_call":
        args = {
            "queries": list(getattr(output_item, "queries", [])),
            "status": getattr(output_item, "status", None),
        }
        results = getattr(output_item, "results", None)
        response_payload: Any | None = None
        if results is not None:
            # Store only the essential result metadata to avoid large payloads
            response_payload = {
                "results": [
                    {
                        "file_id": (
                            getattr(result, "file_id", None)
                            if not isinstance(result, dict)
                            else result.get("file_id")
                        ),
                        "filename": (
                            getattr(result, "filename", None)
                            if not isinstance(result, dict)
                            else result.get("filename")
                        ),
                        "score": (
                            getattr(result, "score", None)
                            if not isinstance(result, dict)
                            else result.get("score")
                        ),
                    }
                    for result in results
                ]
            }
        return ToolCallSummary(
            id=str(getattr(output_item, "id")),
            name=DEFAULT_RAG_TOOL,
            args=args,
            response=response_payload,
        )

    if item_type == "web_search_call":
        args = {"status": getattr(output_item, "status", None)}
        return ToolCallSummary(
            id=str(getattr(output_item, "id")),
            name="web_search",
            args=args,
            response=None,
        )

    if item_type == "mcp_call":
        parsed_arguments = getattr(output_item, "arguments", "")
        args = {"arguments": parsed_arguments}
        server_label = getattr(output_item, "server_label", None)
        if server_label:
            args["server_label"] = server_label
        error = getattr(output_item, "error", None)
        if error:
            args["error"] = error

        return ToolCallSummary(
            id=str(getattr(output_item, "id")),
            name=getattr(output_item, "name", "mcp_call"),
            args=args,
            response=getattr(output_item, "output", None),
        )

    if item_type == "mcp_list_tools":
        tool_names: list[str] = []
        for tool in getattr(output_item, "tools", []):
            if hasattr(tool, "name"):
                tool_names.append(str(getattr(tool, "name")))
            elif isinstance(tool, dict) and tool.get("name"):
                tool_names.append(str(tool.get("name")))
        args = {
            "server_label": getattr(output_item, "server_label", None),
            "tools": tool_names,
        }
        return ToolCallSummary(
            id=str(getattr(output_item, "id")),
            name="mcp_list_tools",
            args=args,
            response=None,
        )

    if item_type == "mcp_approval_request":
        parsed_arguments = getattr(output_item, "arguments", "")
        args = {"arguments": parsed_arguments}
        server_label = getattr(output_item, "server_label", None)
        if server_label:
            args["server_label"] = server_label
        return ToolCallSummary(
            id=str(getattr(output_item, "id")),
            name=getattr(output_item, "name", "mcp_approval_request"),
            args=args,
            response=None,
        )

    return None


async def get_topic_summary(  # pylint: disable=too-many-nested-blocks
    question: str, client: AsyncLlamaStackClient, model_id: str
) -> str:
    """
    Generate a concise topic summary for a question using the Responses API.
    
    Uses the configured topic-summary system prompt with the provided async Llama Stack client; returns the generated summary text trimmed, or an empty string on failure.
    
    Returns:
        str: Trimmed topic summary text if generation succeeded, otherwise an empty string.
    """
    topic_summary_system_prompt = get_topic_summary_system_prompt(configuration)

    try:
        # Use Responses API to generate topic summary
        response = await client.responses.create(
            input=question,
            model=model_id,
            instructions=topic_summary_system_prompt,
            stream=False,
            store=False,  # Don't store topic summary requests
        )
        response = cast(OpenAIResponseObject, response)

        # Extract text from response output
        summary_text = "".join(
            _extract_text_from_response_output_item(output_item)
            for output_item in response.output
        )

        return summary_text.strip() if summary_text else ""
    except Exception as e:  # pylint: disable=broad-exception-caught
        logger.warning("Failed to generate topic summary: %s", e)
        return ""  # Return empty string on failure


@router.post("/query", responses=query_v2_response)
@authorize(Action.QUERY)
async def query_endpoint_handler_v2(
    request: Request,
    query_request: QueryRequest,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> QueryResponse:
    """
    Handle POST /query requests using the Responses API and return a QueryResponse containing the conversation result.
    
    Returns:
        QueryResponse: Conversation ID, LLM-generated response text, tool call summaries, referenced documents, and token usage.
    """
    return await query_endpoint_handler_base(
        request=request,
        query_request=query_request,
        auth=auth,
        mcp_headers=mcp_headers,
        retrieve_response_func=retrieve_response,
        get_topic_summary_func=get_topic_summary,
    )


async def retrieve_response(  # pylint: disable=too-many-locals,too-many-branches,too-many-arguments
    client: AsyncLlamaStackClient,
    model_id: str,
    query_request: QueryRequest,
    token: str,
    mcp_headers: dict[str, dict[str, str]] | None = None,
    *,
    provider_id: str = "",
) -> tuple[TurnSummary, str, list[ReferencedDocument], TokenCounter]:
    """
    Retrieve a response for a user query from the Responses API and assemble a TurnSummary.
    
    Builds and sends a Responses API request (including system prompt, optional attachments, and configured tool groups such as RAG and MCP tools), then collects the assistant text, translated tool-call summaries, referenced documents, and token usage metrics from the response.
    
    Parameters:
        client: Async Llama Stack client used to call the Responses API.
        model_id: Identifier of the LLM model to use.
        query_request: QueryRequest containing the user's query, optional attachments, conversation_id, and tool flags.
        token: Authentication token used when constructing MCP tool headers.
        mcp_headers: Optional per-MCP-server headers to attach to MCP tools.
        provider_id: Optional provider identifier used when recording token usage metrics.
    
    Returns:
        tuple[TurnSummary, str, list[ReferencedDocument], TokenCounter]: 
            A 4-tuple with:
            - TurnSummary: assembled LLM text and tool-call summaries for this turn,
            - conversation ID assigned by the Responses API,
            - list of ReferencedDocument parsed from the response,
            - TokenCounter with extracted token usage metrics.
    """
    # TODO(ltomasbo): implement shields support once available in Responses API
    logger.info("Shields are not yet supported in Responses API. Disabling safety")

    # use system prompt from request or default one
    system_prompt = get_system_prompt(query_request, configuration)
    logger.debug("Using system prompt: %s", system_prompt)

    # TODO(lucasagomes): redact attachments content before sending to LLM
    # if attachments are provided, validate them
    if query_request.attachments:
        validate_attachments_metadata(query_request.attachments)

    # Prepare tools for responses API
    toolgroups: list[dict[str, Any]] | None = None
    if not query_request.no_tools:
        toolgroups = []
        # Get vector stores for RAG tools
        vector_store_ids = [
            vector_store.id for vector_store in (await client.vector_stores.list()).data
        ]

        # Add RAG tools if vector stores are available
        rag_tools = get_rag_tools(vector_store_ids)
        if rag_tools:
            toolgroups.extend(rag_tools)

        # Add MCP server tools
        mcp_tools = get_mcp_tools(configuration.mcp_servers, token, mcp_headers)
        if mcp_tools:
            toolgroups.extend(mcp_tools)
            logger.debug(
                "Configured %d MCP tools: %s",
                len(mcp_tools),
                [tool.get("server_label", "unknown") for tool in mcp_tools],
            )
        # Convert empty list to None for consistency with existing behavior
        if not toolgroups:
            toolgroups = None

    # Prepare input for Responses API
    # Convert attachments to text and concatenate with query
    input_text = query_request.query
    if query_request.attachments:
        for attachment in query_request.attachments:
            # Append attachment content with type label
            input_text += (
                f"\n\n[Attachment: {attachment.attachment_type}]\n{attachment.content}"
            )

    # Create OpenAI response using responses API
    create_kwargs: dict[str, Any] = {
        "input": input_text,
        "model": model_id,
        "instructions": system_prompt,
        "tools": cast(Any, toolgroups),
        "stream": False,
        "store": True,
    }
    if query_request.conversation_id:
        create_kwargs["previous_response_id"] = query_request.conversation_id

    response = await client.responses.create(**create_kwargs)
    response = cast(OpenAIResponseObject, response)

    logger.debug(
        "Received response with ID: %s, output items: %d",
        response.id,
        len(response.output),
    )

    # Return the response ID - client can use it for chaining if desired
    conversation_id = response.id

    # Process OpenAI response format
    llm_response = ""
    tool_calls: list[ToolCallSummary] = []

    for output_item in response.output:
        message_text = _extract_text_from_response_output_item(output_item)
        if message_text:
            llm_response += message_text

        tool_summary = _build_tool_call_summary(output_item)
        if tool_summary:
            tool_calls.append(tool_summary)

    logger.info(
        "Response processing complete - Tool calls: %d, Response length: %d chars",
        len(tool_calls),
        len(llm_response),
    )

    summary = TurnSummary(
        llm_response=llm_response,
        tool_calls=tool_calls,
    )

    # Extract referenced documents and token usage from Responses API response
    referenced_documents = parse_referenced_documents_from_responses_api(response)
    model_label = model_id.split("/", 1)[1] if "/" in model_id else model_id
    token_usage = extract_token_usage_from_responses_api(
        response, model_label, provider_id, system_prompt
    )

    if not summary.llm_response:
        logger.warning(
            "Response lacks content (conversation_id=%s)",
            conversation_id,
        )
    return (summary, conversation_id, referenced_documents, token_usage)


def parse_referenced_documents_from_responses_api(
    response: OpenAIResponseObject,  # pylint: disable=unused-argument
) -> list[ReferencedDocument]:
    """
    Parse referenced documents from OpenAI Responses API response.

    Args:
        response: The OpenAI Response API response object

    Returns:
        list[ReferencedDocument]: List of referenced documents with doc_url and doc_title
    """
    # TODO(ltomasbo): need to parse source documents from Responses API response.
    # The Responses API has a different structure than Agent API for referenced documents.
    # Need to extract from:
    # - OpenAIResponseOutputMessageFileSearchToolCall.results
    # - OpenAIResponseAnnotationCitation in message content
    # - OpenAIResponseAnnotationFileCitation in message content
    return []


def extract_token_usage_from_responses_api(
    response: OpenAIResponseObject,
    model: str,
    provider: str,
    system_prompt: str = "",  # pylint: disable=unused-argument
) -> TokenCounter:
    """
    Extract token counts from a Responses API response and update LLM token metrics.
    
    Parses the response's `usage` field (supports both dict and object shapes), sets `input_tokens` and `output_tokens` on a returned TokenCounter, and increments the LLM call and token Prometheus metrics when actual usage values are present. If no usable usage data exists, returns a TokenCounter with zero token counts but still records the LLM call metric.
    
    Parameters:
        response (OpenAIResponseObject): The Responses API response object that may contain a `usage` attribute.
        model (str): Model identifier used for metrics labeling.
        provider (str): Provider identifier used for metrics labeling.
        system_prompt (str): Unused; kept for compatibility.
    
    Returns:
        TokenCounter: Token usage with `input_tokens`, `output_tokens`, and `llm_calls` populated.
    """
    token_counter = TokenCounter()
    token_counter.llm_calls = 1

    # Extract usage from the response if available
    # Note: usage attribute exists at runtime but may not be in type definitions
    usage = getattr(response, "usage", None)
    if usage:
        try:
            # Handle both dict and object cases due to llama_stack inconsistency:
            # - When llama_stack converts to chat_completions internally, usage is a dict
            # - When using proper Responses API, usage should be an object
            # TODO: Remove dict handling once llama_stack standardizes on object type  # pylint: disable=fixme
            if isinstance(usage, dict):
                input_tokens = usage.get("input_tokens", 0)
                output_tokens = usage.get("output_tokens", 0)
            else:
                # Object with attributes (expected final behavior)
                input_tokens = getattr(usage, "input_tokens", 0)
                output_tokens = getattr(usage, "output_tokens", 0)
            # Only set if we got valid values
            if input_tokens or output_tokens:
                token_counter.input_tokens = input_tokens or 0
                token_counter.output_tokens = output_tokens or 0

                logger.debug(
                    "Extracted token usage from Responses API: input=%d, output=%d",
                    token_counter.input_tokens,
                    token_counter.output_tokens,
                )

                # Update Prometheus metrics only when we have actual usage data
                try:
                    metrics.llm_token_sent_total.labels(provider, model).inc(
                        token_counter.input_tokens
                    )
                    metrics.llm_token_received_total.labels(provider, model).inc(
                        token_counter.output_tokens
                    )
                except (AttributeError, TypeError, ValueError) as e:
                    logger.warning("Failed to update token metrics: %s", e)
                _increment_llm_call_metric(provider, model)
            else:
                logger.debug(
                    "Usage object exists but tokens are 0 or None, treating as no usage info"
                )
                # Still increment the call counter
                _increment_llm_call_metric(provider, model)
        except (AttributeError, KeyError, TypeError) as e:
            logger.warning(
                "Failed to extract token usage from response.usage: %s. Usage value: %s",
                e,
                usage,
            )
            # Still increment the call counter
            _increment_llm_call_metric(provider, model)
    else:
        # No usage information available - this is expected when llama stack
        # internally converts to chat_completions
        logger.debug(
            "No usage information in Responses API response, token counts will be 0"
        )
        # token_counter already initialized with 0 values
        # Still increment the call counter
        _increment_llm_call_metric(provider, model)

    return token_counter


def _increment_llm_call_metric(provider: str, model: str) -> None:
    """
    Increment the LLM call metric for the given provider and model, logging a warning on failure.
    
    Parameters:
        provider (str): Provider label to attach to the metric (e.g., "openai").
        model (str): Model label to attach to the metric (e.g., "gpt-4").
    """
    try:
        metrics.llm_calls_total.labels(provider, model).inc()
    except (AttributeError, TypeError, ValueError) as e:
        logger.warning("Failed to update LLM call metric: %s", e)


def get_rag_tools(vector_store_ids: list[str]) -> list[dict[str, Any]] | None:
    """
    Return a Responses API tool definition for retrieval-augmented generation (RAG).
    
    Returns:
        A list containing one `file_search` tool dictionary configured with the provided `vector_store_ids` and `max_num_results` set to 10, or `None` if `vector_store_ids` is empty.
    """
    if not vector_store_ids:
        return None

    return [
        {
            "type": "file_search",
            "vector_store_ids": vector_store_ids,
            "max_num_results": 10,
        }
    ]


def get_mcp_tools(
    mcp_servers: list,
    token: str | None = None,
    mcp_headers: dict[str, dict[str, str]] | None = None,
) -> list[dict[str, Any]]:
    """
    Construct Responses API tool definitions for configured MCP servers.
    
    Parameters:
        mcp_servers (list): Iterable of MCP server configuration objects with at least `name` and `url` attributes.
        token (str | None): Optional bearer token to attach as an Authorization header when per-server headers are not provided.
        mcp_headers (dict[str, dict[str, str]] | None): Optional mapping from server URL to headers to include for that server.
    
    Returns:
        list[dict[str, Any]]: A list of tool definition dictionaries for the Responses API, each containing `type`, `server_label`, `server_url`, `require_approval`, and optionally `headers`.
    """
    tools = []
    for mcp_server in mcp_servers:
        tool_def = {
            "type": "mcp",
            "server_label": mcp_server.name,
            "server_url": mcp_server.url,
            "require_approval": "never",
        }

        # Add authentication if headers or token provided (Response API format)
        headers = (mcp_headers or {}).get(mcp_server.url)
        if headers:
            tool_def["headers"] = headers
        elif token:
            tool_def["headers"] = {"Authorization": f"Bearer {token}"}
        tools.append(tool_def)
    return tools