"""Handler for REST API call to provide answer to query."""

import ast
import json
import logging
import re
from datetime import UTC, datetime
from typing import Annotated, Any, Optional, cast

from fastapi import APIRouter, Depends, HTTPException, Request, status
from litellm.exceptions import RateLimitError
from llama_stack_client import (
    APIConnectionError,
    AsyncLlamaStackClient,  # type: ignore
)
from llama_stack_client.lib.agents.event_logger import interleaved_content_as_str
from llama_stack_client.types import Shield, UserMessage  # type: ignore
from llama_stack_client.types.agents.turn import Turn
from llama_stack_client.types.agents.turn_create_params import (
    Toolgroup,
    ToolgroupAgentToolGroupWithArgs,
    Document,
)
from llama_stack_client.types.model_list_response import ModelListResponse
from llama_stack_client.types.shared.interleaved_content_item import TextContentItem
from llama_stack_client.types.tool_execution_step import ToolExecutionStep

import constants
import metrics
from app.database import get_session
from authentication import get_auth_dependency
from authentication.interface import AuthTuple
from authorization.middleware import authorize
from client import AsyncLlamaStackClientHolder
from configuration import configuration
from models.cache_entry import CacheEntry
from models.config import Action
from models.database.conversations import UserConversation
from models.requests import Attachment, QueryRequest
from models.responses import (
    ForbiddenResponse,
    QueryResponse,
    ReferencedDocument,
    ToolCall,
    UnauthorizedResponse,
    QuotaExceededResponse,
)
from utils.endpoints import (
    check_configuration_loaded,
    get_agent,
    get_topic_summary_system_prompt,
    get_temp_agent,
    get_system_prompt,
    store_conversation_into_cache,
    validate_conversation_ownership,
    validate_model_provider_override,
)
from utils.quota import (
    get_available_quotas,
    check_tokens_available,
    consume_tokens,
)
from utils.mcp_headers import handle_mcp_headers_with_toolgroups, mcp_headers_dependency
from utils.transcripts import store_transcript
from utils.types import TurnSummary
from utils.token_counter import extract_and_update_token_metrics, TokenCounter

logger = logging.getLogger("app.endpoints.handlers")
router = APIRouter(tags=["query"])

query_response: dict[int | str, dict[str, Any]] = {
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


def is_transcripts_enabled() -> bool:
    """Check if transcripts is enabled.

    Returns:
        bool: True if transcripts is enabled, False otherwise.
    """
    return configuration.user_data_collection_configuration.transcripts_enabled


def persist_user_conversation_details(
    user_id: str,
    conversation_id: str,
    model: str,
    provider_id: str,
    topic_summary: Optional[str],
) -> None:
    """
    Associate a conversation with a user in the database, creating it if missing or updating metadata if it exists.
    
    If no UserConversation with the given conversation_id exists, creates one with the provided model/provider, topic summary, and a message count of 1. If it exists, updates last used model and provider, sets last_message_at to the current UTC time, and increments message_count. Commits the change to the database.
    
    Parameters:
        user_id (str): The identifier of the user to associate with the conversation.
        conversation_id (str): The conversation identifier to create or update.
        model (str): The model identifier last used for this conversation.
        provider_id (str): The provider identifier last used for this conversation.
        topic_summary (Optional[str]): Optional topic summary to store for new conversations.
    """
    with get_session() as session:
        existing_conversation = (
            session.query(UserConversation).filter_by(id=conversation_id).first()
        )

        if not existing_conversation:
            conversation = UserConversation(
                id=conversation_id,
                user_id=user_id,
                last_used_model=model,
                last_used_provider=provider_id,
                topic_summary=topic_summary,
                message_count=1,
            )
            session.add(conversation)
            logger.debug(
                "Associated conversation %s to user %s", conversation_id, user_id
            )
        else:
            existing_conversation.last_used_model = model
            existing_conversation.last_used_provider = provider_id
            existing_conversation.last_message_at = datetime.now(UTC)
            existing_conversation.message_count += 1

        session.commit()


def evaluate_model_hints(
    user_conversation: UserConversation | None,
    query_request: QueryRequest,
) -> tuple[str | None, str | None]:
    """
    Determine which model and provider IDs to use for a query by preferring explicit request values and falling back to the user's last-used values when the request omits them.
    
    Parameters:
        user_conversation (UserConversation | None): The user's conversation metadata that may contain last_used_model and last_used_provider.
        query_request (QueryRequest): The incoming query request which may include explicit `model` and `provider` hints.
    
    Returns:
        tuple[str | None, str | None]: A tuple (model_id, provider_id) where each element is the chosen identifier or `None` if neither the request nor the user conversation specifies it.
    """
    model_id: str | None = query_request.model
    provider_id: str | None = query_request.provider

    if user_conversation is not None:
        if query_request.model is not None:
            if query_request.model != user_conversation.last_used_model:
                logger.debug(
                    "Model specified in request: %s, preferring it over user conversation model %s",
                    query_request.model,
                    user_conversation.last_used_model,
                )
        else:
            logger.debug(
                "No model specified in request, using latest model from user conversation: %s",
                user_conversation.last_used_model,
            )
            model_id = user_conversation.last_used_model

        if query_request.provider is not None:
            if query_request.provider != user_conversation.last_used_provider:
                logger.debug(
                    "Provider specified in request: %s, "
                    "preferring it over user conversation provider %s",
                    query_request.provider,
                    user_conversation.last_used_provider,
                )
        else:
            logger.debug(
                "No provider specified in request, "
                "using latest provider from user conversation: %s",
                user_conversation.last_used_provider,
            )
            provider_id = user_conversation.last_used_provider

    return model_id, provider_id


async def get_topic_summary(
    question: str, client: AsyncLlamaStackClient, model_id: str
) -> str:
    """
    Produce a concise topic summary for the given question using a temporary agent session.
    
    Parameters:
        question (str): The user question to summarize.
        client (AsyncLlamaStackClient): Llama Stack client used to create a temporary agent and run the turn.
        model_id (str): Identifier of the model to use for generating the summary.
    
    Returns:
        str: The summary text produced by the agent, or an empty string if no output content is present.
    """
    topic_summary_system_prompt = get_topic_summary_system_prompt(configuration)
    agent, session_id, _ = await get_temp_agent(
        client, model_id, topic_summary_system_prompt
    )
    response = await agent.create_turn(
        messages=[UserMessage(role="user", content=question)],
        session_id=session_id,
        stream=False,
        toolgroups=None,
    )
    response = cast(Turn, response)
    return (
        interleaved_content_as_str(response.output_message.content)
        if (
            getattr(response, "output_message", None) is not None
            and getattr(response.output_message, "content", None) is not None
        )
        else ""
    )


async def query_endpoint_handler_base(  # pylint: disable=R0914
    request: Request,
    query_request: QueryRequest,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    mcp_headers: dict[str, dict[str, str]],
    retrieve_response_func: Any,
    get_topic_summary_func: Any,
) -> QueryResponse:
    """
    Handle a query request by selecting a model, invoking the LLM/agent to produce a response, persisting conversation and transcript data as configured, and returning an assembled QueryResponse.
    
    Parameters:
        request: FastAPI request object (used for authorization context and request state).
        query_request: The incoming QueryRequest with the user's query and optional overrides.
        auth: Authentication tuple provided by dependency injection.
        mcp_headers: MCP header mappings provided by dependency.
        retrieve_response_func: Callable used to produce the LLM/agent response. Expected to accept (client, model_id, query_request, token, mcp_headers=..., provider_id=...) and return (TurnSummary, conversation_id, list[ReferencedDocument], TokenCounter).
        get_topic_summary_func: Async callable used to obtain an initial topic summary for new conversations. Expected to accept (question: str, client, model_id) and return a string summary.
    
    Returns:
        QueryResponse: The final response object including conversation_id, LLM-generated response, RAG chunks, tool calls, referenced documents, token usage, and available quotas.
    
    Raises:
        HTTPException: With status 500 if the Llama Stack service cannot be reached.
        HTTPException: With status 429 if a model's token quota is exceeded.
    """
    check_configuration_loaded(configuration)

    # Enforce RBAC: optionally disallow overriding model/provider in requests
    validate_model_provider_override(query_request, request.state.authorized_actions)

    # log Llama Stack configuration
    logger.info("Llama stack config: %s", configuration.llama_stack_configuration)

    user_id, _, _skip_userid_check, token = auth

    started_at = datetime.now(UTC).strftime("%Y-%m-%dT%H:%M:%SZ")
    user_conversation: UserConversation | None = None
    if query_request.conversation_id:
        logger.debug(
            "Conversation ID specified in query: %s", query_request.conversation_id
        )
        user_conversation = validate_conversation_ownership(
            user_id=user_id,
            conversation_id=query_request.conversation_id,
            others_allowed=(
                Action.QUERY_OTHERS_CONVERSATIONS in request.state.authorized_actions
            ),
        )

        if user_conversation is None:
            logger.warning(
                "Conversation %s not found for user %s",
                query_request.conversation_id,
                user_id,
            )
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail={
                    "response": "Conversation not found",
                    "cause": "The requested conversation does not exist.",
                },
            )
    else:
        logger.debug("Query does not contain conversation ID")

    try:
        check_tokens_available(configuration.quota_limiters, user_id)
        # try to get Llama Stack client
        client = AsyncLlamaStackClientHolder().get_client()
        llama_stack_model_id, model_id, provider_id = select_model_and_provider_id(
            await client.models.list(),
            *evaluate_model_hints(
                user_conversation=user_conversation, query_request=query_request
            ),
        )
        summary, conversation_id, referenced_documents, token_usage = (
            await retrieve_response_func(
                client,
                llama_stack_model_id,
                query_request,
                token,
                mcp_headers=mcp_headers,
                provider_id=provider_id,
            )
        )

        # Get the initial topic summary for the conversation
        topic_summary = None
        with get_session() as session:
            existing_conversation = (
                session.query(UserConversation).filter_by(id=conversation_id).first()
            )
            if not existing_conversation:
                topic_summary = await get_topic_summary_func(
                    query_request.query, client, llama_stack_model_id
                )
        # Convert RAG chunks to dictionary format once for reuse
        logger.info("Processing RAG chunks...")
        rag_chunks_dict = [chunk.model_dump() for chunk in summary.rag_chunks]

        if not is_transcripts_enabled():
            logger.debug("Transcript collection is disabled in the configuration")
        else:
            store_transcript(
                user_id=user_id,
                conversation_id=conversation_id,
                model_id=model_id,
                provider_id=provider_id,
                query_is_valid=True,  # TODO(lucasagomes): implement as part of query validation
                query=query_request.query,
                query_request=query_request,
                summary=summary,
                rag_chunks=rag_chunks_dict,
                truncated=False,  # TODO(lucasagomes): implement truncation as part of quota work
                attachments=query_request.attachments or [],
            )

        logger.info("Persisting conversation details...")
        persist_user_conversation_details(
            user_id=user_id,
            conversation_id=conversation_id,
            model=model_id,
            provider_id=provider_id,
            topic_summary=topic_summary,
        )

        completed_at = datetime.now(UTC).strftime("%Y-%m-%dT%H:%M:%SZ")

        cache_entry = CacheEntry(
            query=query_request.query,
            response=summary.llm_response,
            provider=provider_id,
            model=model_id,
            started_at=started_at,
            completed_at=completed_at,
            referenced_documents=referenced_documents if referenced_documents else None,
        )

        consume_tokens(
            configuration.quota_limiters,
            user_id,
            input_tokens=token_usage.input_tokens,
            output_tokens=token_usage.output_tokens,
        )

        store_conversation_into_cache(
            configuration,
            user_id,
            conversation_id,
            cache_entry,
            _skip_userid_check,
            topic_summary,
        )

        # Convert tool calls to response format
        logger.info("Processing tool calls...")
        tool_calls = [
            ToolCall(
                tool_name=tc.name,
                arguments=(
                    tc.args if isinstance(tc.args, dict) else {"query": str(tc.args)}
                ),
                result=(
                    {"response": tc.response}
                    if tc.response and tc.name != constants.DEFAULT_RAG_TOOL
                    else None
                ),
            )
            for tc in summary.tool_calls
        ]

        logger.info("Using referenced documents from response...")

        available_quotas = get_available_quotas(configuration.quota_limiters, user_id)

        logger.info("Building final response...")
        response = QueryResponse(
            conversation_id=conversation_id,
            response=summary.llm_response,
            rag_chunks=summary.rag_chunks if summary.rag_chunks else [],
            tool_calls=tool_calls if tool_calls else None,
            referenced_documents=referenced_documents,
            truncated=False,  # TODO: implement truncation detection
            input_tokens=token_usage.input_tokens,
            output_tokens=token_usage.output_tokens,
            available_quotas=available_quotas,
        )
        logger.info("Query processing completed successfully!")
        return response

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
    except RateLimitError as e:
        used_model = getattr(e, "model", "unknown")
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail={
                "response": "Model quota exceeded",
                "cause": f"The token quota for model {used_model} has been exceeded.",
            },
        ) from e


@router.post("/query", responses=query_response)
@authorize(Action.QUERY)
async def query_endpoint_handler(
    request: Request,
    query_request: QueryRequest,
    auth: Annotated[AuthTuple, Depends(get_auth_dependency())],
    mcp_headers: dict[str, dict[str, str]] = Depends(mcp_headers_dependency),
) -> QueryResponse:
    """
    Handle POST /query requests using the Agent API.
    
    Wrapper around query_endpoint_handler_base that supplies the Agent-specific
    retrieve_response and get_topic_summary callbacks.
    
    Returns:
        QueryResponse: The assembled response including conversation_id, LLM-generated
        response text, rag chunks, tool calls, referenced documents, token usage,
        and available quotas.
    """
    return await query_endpoint_handler_base(
        request=request,
        query_request=query_request,
        auth=auth,
        mcp_headers=mcp_headers,
        retrieve_response_func=retrieve_response,
        get_topic_summary_func=get_topic_summary,
    )


def select_model_and_provider_id(
    models: ModelListResponse, model_id: str | None, provider_id: str | None
) -> tuple[str, str, str]:
    """
    Choose the Llama Stack model identifier, model label, and provider identifier to use for a query.
    
    Parameters:
        models (ModelListResponse): Iterable of available models from the Llama Stack client.
        model_id (str | None): Optional model label supplied by the request or configuration (e.g., "gpt-4").
        provider_id (str | None): Optional provider identifier supplied by the request or configuration (e.g., "openai").
    
    Returns:
        tuple[str, str, str]: A tuple with
            - the combined Llama Stack model identifier in the form "provider/model",
            - the model label (the model part without the provider),
            - the provider identifier.
    
    Raises:
        HTTPException: If no LLM model is available among `models`, or if the resolved model/provider pair is not present in `models`.
    """
    # If model_id and provider_id are provided in the request, use them

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
            model_label = model_id.split("/", 1)[1] if "/" in model_id else model_id
            return model_id, model_label, provider_id
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

    return llama_stack_model_id, model_id, provider_id


def _is_inout_shield(shield: Shield) -> bool:
    """
    Determine if the shield identifier indicates an input/output shield.

    Parameters:
        shield (Shield): The shield to check.

    Returns:
        bool: True if the shield identifier starts with "inout_", otherwise False.
    """
    return shield.identifier.startswith("inout_")


def is_output_shield(shield: Shield) -> bool:
    """
    Return whether a Shield applies to output.
    
    Parameters:
        shield (Shield): The shield to classify.
    
    Returns:
        `true` if the shield's identifier starts with "output_" or "inout_", `false` otherwise.
    """
    return _is_inout_shield(shield) or shield.identifier.startswith("output_")


def is_input_shield(shield: Shield) -> bool:
    """
    Determine whether a shield applies to input monitoring.
    
    Returns:
        `True` if the shield monitors input or both input and output, `False` otherwise.
    """
    return _is_inout_shield(shield) or not is_output_shield(shield)


def parse_metadata_from_text_item(
    text_item: TextContentItem,
) -> Optional[ReferencedDocument]:
    """
    Extracts a referenced document from a TextContentItem's metadata block.
    
    Searches for a "Metadata: { ... }" block inside the item's text and returns the first metadata object that contains both the "docs_url" and "title" fields.
    
    Parameters:
        text_item (TextContentItem): The text content item to parse for metadata.
    
    Returns:
        ReferencedDocument: A ReferencedDocument with `doc_url` and `doc_title` if a valid metadata block is found; `None` otherwise.
    """
    docs: list[ReferencedDocument] = []
    if not isinstance(text_item, TextContentItem):
        return docs

    metadata_blocks = re.findall(
        r"Metadata:\s*({.*?})(?:\n|$)", text_item.text, re.DOTALL
    )
    for block in metadata_blocks:
        try:
            data = ast.literal_eval(block)
            url = data.get("docs_url")
            title = data.get("title")
            if url and title:
                return ReferencedDocument(doc_url=url, doc_title=title)
            logger.debug("Invalid metadata block (missing url or title): %s", block)
        except (ValueError, SyntaxError) as e:
            logger.debug("Failed to parse metadata block: %s | Error: %s", block, e)
    return None


def parse_referenced_documents(response: Turn) -> list[ReferencedDocument]:
    """
    Extract referenced document metadata from a Turn's RAG tool outputs.
    
    Returns:
        list[ReferencedDocument]: A list of ReferencedDocument objects discovered in the turn's RAG tool responses; empty list if none found.
    """
    docs = []
    for step in response.steps:
        if not isinstance(step, ToolExecutionStep):
            continue
        for tool_response in step.tool_responses:
            if tool_response.tool_name != constants.DEFAULT_RAG_TOOL:
                continue
            for text_item in tool_response.content:
                if not isinstance(text_item, TextContentItem):
                    continue
                doc = parse_metadata_from_text_item(text_item)
                if doc:
                    docs.append(doc)
    return docs


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
    Generate an agent/LLM response for the given query and return the response summary, conversation ID, any referenced documents discovered in tool outputs, and token usage.
    
    Parameters:
        client (AsyncLlamaStackClient): Llama Stack client used to create agents and list resources.
        model_id (str): Identifier of the model to use (may include provider prefix).
        query_request (QueryRequest): User query and related metadata (attachments, documents, conversation hints).
        token (str): Authentication token used when contacting configured MCP servers.
        mcp_headers (dict[str, dict[str, str]] | None): Optional per-MCP-server headers to forward to tool integrations.
        provider_id (str): Optional provider identifier associated with the chosen model.
    
    Returns:
        tuple[TurnSummary, str, list[ReferencedDocument], TokenCounter]: 
            A 4-tuple containing:
            - TurnSummary: summarized LLM/agent response and collected tool calls,
            - conversation_id (str): the conversation identifier used/returned by the agent,
            - referenced_documents (list[ReferencedDocument]): documents parsed from tool outputs,
            - TokenCounter: token usage information for the response.
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

    # TODO: LCORE-881 - Remove if Llama Stack starts to support these mime types
    documents: list[Document] = [
        (
            {"content": doc["content"], "mime_type": "text/plain"}
            if doc["mime_type"].lower() in ("application/json", "application/xml")
            else doc
        )
        for doc in query_request.get_documents()
    ]

    response = await agent.create_turn(
        messages=[UserMessage(role="user", content=query_request.query)],
        session_id=session_id,
        documents=documents,
        stream=False,
        toolgroups=toolgroups,
    )
    response = cast(Turn, response)

    summary = TurnSummary(
        llm_response=(
            interleaved_content_as_str(response.output_message.content)
            if (
                getattr(response, "output_message", None) is not None
                and getattr(response.output_message, "content", None) is not None
            )
            else ""
        ),
        tool_calls=[],
    )

    referenced_documents = parse_referenced_documents(response)

    # Update token count metrics and extract token usage in one call
    model_label = model_id.split("/", 1)[1] if "/" in model_id else model_id
    token_usage = extract_and_update_token_metrics(
        response, model_label, provider_id, system_prompt
    )

    # Check for validation errors in the response
    steps = response.steps or []
    for step in steps:
        if step.step_type == "shield_call" and step.violation:
            # Metric for LLM validation errors
            metrics.llm_calls_validation_errors_total.inc()
        if step.step_type == "tool_execution":
            summary.append_tool_calls_from_llama(step)

    if not summary.llm_response:
        logger.warning(
            "Response lacks output_message.content (conversation_id=%s)",
            conversation_id,
        )
    return (summary, conversation_id, referenced_documents, token_usage)


def validate_attachments_metadata(attachments: list[Attachment]) -> None:
    """
    Validate that each attachment's `attachment_type` and `content_type` are allowed.
    
    Raises:
        HTTPException: With status 422 and a detail payload when any attachment has an unsupported `attachment_type` or `content_type`.
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


def get_rag_toolgroups(
    vector_db_ids: list[str],
) -> list[Toolgroup] | None:
    """
    Return a list of RAG Tool groups if the given vector DB list is not empty.

    Generate a list containing a RAG knowledge search toolgroup if
    vector database IDs are provided.

    Parameters:
        vector_db_ids (list[str]): List of vector database identifiers to include in the toolgroup.

    Returns:
        list[Toolgroup] | None: A list with a single RAG toolgroup if
        vector_db_ids is non-empty; otherwise, None.
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