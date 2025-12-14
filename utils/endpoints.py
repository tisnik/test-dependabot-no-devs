"""Utility functions for endpoint handlers."""

from contextlib import suppress
from datetime import UTC, datetime
from typing import Any

from fastapi import HTTPException
from llama_stack_client._client import AsyncLlamaStackClient
from llama_stack_client.lib.agents.agent import AsyncAgent
from pydantic import AnyUrl, ValidationError

import constants
from app.database import get_session
from configuration import AppConfig, LogicError
from log import get_logger
from models.cache_entry import CacheEntry
from models.config import Action
from models.database.conversations import UserConversation
from models.requests import QueryRequest
from models.responses import (
    ForbiddenResponse,
    InternalServerErrorResponse,
    NotFoundResponse,
    ReferencedDocument,
    UnprocessableEntityResponse,
)
from utils.suid import get_suid
from utils.types import GraniteToolParser, TurnSummary

logger = get_logger(__name__)


def delete_conversation(conversation_id: str) -> bool:
    """Delete a conversation from the local database by its ID.

    Args:
        conversation_id (str): The unique identifier of the conversation to delete.

    Returns:
        bool: True if the conversation was deleted, False if it was not found.
    """
    with get_session() as session:
        db_conversation = (
            session.query(UserConversation).filter_by(id=conversation_id).first()
        )
        if db_conversation:
            session.delete(db_conversation)
            session.commit()
            logger.info("Deleted conversation %s from local database", conversation_id)
            return True
        logger.info(
            "Conversation %s not found in local database, it may have already been deleted",
            conversation_id,
        )
        return False


def retrieve_conversation(conversation_id: str) -> UserConversation | None:
    """
    Retrieve a user conversation by its ID from the database.
    
    Returns:
        The UserConversation if found, otherwise None.
    """
    with get_session() as session:
        return session.query(UserConversation).filter_by(id=conversation_id).first()


def validate_conversation_ownership(
    user_id: str, conversation_id: str, others_allowed: bool = False
) -> UserConversation | None:
    """Validate that the conversation belongs to the user.

    Validates that the conversation with the given ID belongs to the user with the given ID.
    If `others_allowed` is True, it allows conversations that do not belong to the user,
    which is useful for admin access.
    """
    with get_session() as session:
        conversation_query = session.query(UserConversation)

        filtered_conversation_query = (
            conversation_query.filter_by(id=conversation_id)
            if others_allowed
            else conversation_query.filter_by(id=conversation_id, user_id=user_id)
        )

        conversation: UserConversation | None = filtered_conversation_query.first()

        return conversation


def can_access_conversation(
    conversation_id: str, user_id: str, others_allowed: bool
) -> bool:
    """
    Determine whether a user may access a conversation.
    
    If `others_allowed` is True, access is permitted. If the conversation does not exist, access is permitted. If the conversation exists, access is permitted only when the conversation's owner matches `user_id`.
    
    Parameters:
        conversation_id (str): ID of the conversation to check.
        user_id (str): ID of the requesting user.
        others_allowed (bool): If True, bypasses ownership checks and allows access to conversations owned by others.
    
    Returns:
        bool: True if access is permitted, False otherwise.
    """
    if others_allowed:
        return True

    with get_session() as session:
        owner_user_id = (
            session.query(UserConversation.user_id)
            .filter(UserConversation.id == conversation_id)
            .scalar()
        )
        # If conversation does not exist, permissions check returns True
        if owner_user_id is None:
            return True

        # If conversation exists, user_id must match
        return owner_user_id == user_id


def check_configuration_loaded(config: AppConfig) -> None:
    """
    Raise an error if the configuration is not loaded.

    Args:
        config (AppConfig): The application configuration.

    Raises:
        HTTPException: If configuration is missing.
    """
    try:
        _ = config.configuration
    except LogicError as e:
        response = InternalServerErrorResponse.configuration_not_loaded()
        raise HTTPException(**response.model_dump()) from e


def get_system_prompt(query_request: QueryRequest, config: AppConfig) -> str:
    """
    Resolve which system prompt to use for a query.

    Precedence:
    1. If the request includes `system_prompt`, that value is returned (highest
       precedence).
    2. Else if the application configuration provides a customization
       `system_prompt`, that value is returned.
    3. Otherwise the module default `constants.DEFAULT_SYSTEM_PROMPT` is
       returned (lowest precedence).

    If configuration disables per-request system prompts
    (config.customization.disable_query_system_prompt) and the incoming
    `query_request` contains a `system_prompt`, an HTTP 422 Unprocessable
    Entity is raised instructing the client to remove the field.

    Parameters:
        query_request (QueryRequest): The incoming query payload; may contain a
        per-request `system_prompt`.
        config (AppConfig): Application configuration which may include
        customization flags and a default `system_prompt`.

    Returns:
        str: The resolved system prompt to apply to the request.
    """
    system_prompt_disabled = (
        config.customization is not None
        and config.customization.disable_query_system_prompt
    )
    if system_prompt_disabled and query_request.system_prompt:
        response = UnprocessableEntityResponse(
            response="System prompt customization is disabled",
            cause=(
                "This instance does not support customizing the system prompt in the "
                "query request (disable_query_system_prompt is set). Please remove the "
                "system_prompt field from your request."
            ),
        )
        raise HTTPException(**response.model_dump())

    if query_request.system_prompt:
        # Query taking precedence over configuration is the only behavior that
        # makes sense here - if the configuration wants precedence, it can
        # disable query system prompt altogether with disable_system_prompt.
        return query_request.system_prompt

    # profile takes precedence for setting prompt
    if (
        config.customization is not None
        and config.customization.custom_profile is not None
    ):
        prompt = config.customization.custom_profile.get_prompts().get("default")
        if prompt:
            return prompt

    if (
        config.customization is not None
        and config.customization.system_prompt is not None
    ):
        return config.customization.system_prompt

    # default system prompt has the lowest precedence
    return constants.DEFAULT_SYSTEM_PROMPT


def get_topic_summary_system_prompt(config: AppConfig) -> str:
    """
    Return the system prompt used for topic summaries, preferring a per-profile customization when available.
    
    Parameters:
        config (AppConfig): Application configuration from which to read customization/profile settings.
    
    Returns:
        str: The topic summary system prompt from the active custom profile if set, otherwise the default prompt.
    """
    # profile takes precedence for setting prompt
    if (
        config.customization is not None
        and config.customization.custom_profile is not None
    ):
        prompt = config.customization.custom_profile.get_prompts().get("topic_summary")
        if prompt:
            return prompt

    return constants.DEFAULT_TOPIC_SUMMARY_SYSTEM_PROMPT


def validate_model_provider_override(
    query_request: QueryRequest, authorized_actions: set[Action] | frozenset[Action]
) -> None:
    """
    Enforces RBAC for model/provider override fields in a QueryRequest.
    
    Raises:
        HTTPException: If `query_request` specifies `model` or `provider` and `authorized_actions` does not include `Action.MODEL_OVERRIDE`. The exception is created from `ForbiddenResponse.model_override()`.
    """
    if (query_request.model is not None or query_request.provider is not None) and (
        Action.MODEL_OVERRIDE not in authorized_actions
    ):
        response = ForbiddenResponse.model_override()
        raise HTTPException(**response.model_dump())


# # pylint: disable=R0913,R0917
def store_conversation_into_cache(
    config: AppConfig,
    user_id: str,
    conversation_id: str,
    cache_entry: CacheEntry,
    _skip_userid_check: bool,
    topic_summary: str | None,
) -> None:
    """
    Insert a conversation entry and optional topic summary into the configured conversation cache.
    
    If a conversation cache type is configured but the cache instance is not initialized, the function logs a warning and returns without persisting anything.
    
    Parameters:
        config (AppConfig): Application configuration that may contain conversation cache settings and instance.
        user_id (str): Owner identifier used as the cache key.
        conversation_id (str): Conversation identifier used as the cache key.
        cache_entry (CacheEntry): Entry to insert or append to the conversation history.
        _skip_userid_check (bool): When true, bypasses enforcing that the cache operation must match the user id.
        topic_summary (str | None): Optional topic summary to store alongside the conversation; ignored if None or empty.
    """
    if config.conversation_cache_configuration.type is not None:
        cache = config.conversation_cache
        if cache is None:
            logger.warning("Conversation cache configured but not initialized")
            return
        cache.insert_or_append(
            user_id, conversation_id, cache_entry, _skip_userid_check
        )
        if topic_summary and len(topic_summary) > 0:
            cache.set_topic_summary(
                user_id, conversation_id, topic_summary, _skip_userid_check
            )


# # pylint: disable=R0913,R0917,unused-argument
async def get_agent(
    client: AsyncLlamaStackClient,
    model_id: str,
    system_prompt: str,
    available_input_shields: list[str],
    available_output_shields: list[str],
    conversation_id: str | None,
    no_tools: bool = False,
) -> tuple[AsyncAgent, str, str]:
    """
    Create or reuse an AsyncAgent with session persistence for a conversation.
    
    If a conversation_id is provided the function will attempt to bind the agent to that conversation and return an existing session; otherwise it creates a new conversation and session.
    
    Parameters:
        model_id (str): Model identifier to instantiate the agent with.
        system_prompt (str): System/instruction prompt for the agent.
        available_input_shields (list[str]): Input shields to apply to the agent.
        available_output_shields (list[str]): Output shields to apply to the agent.
        conversation_id (str | None): Conversation identifier to reuse; if None a new conversation is created.
        no_tools (bool): When True, disables tool parsing for the agent.
    
    Returns:
        tuple[AsyncAgent, str, str]: (agent, conversation_id, session_id).
    
    Raises:
        HTTPException: 404 Not Found if a provided conversation_id is reused but no sessions exist for that conversation.
    """
    existing_agent_id = None
    if conversation_id:
        with suppress(ValueError):
            # agent_response = await client.agents.retrieve(agent_id=conversation_id)
            # existing_agent_id = agent_response.agent_id
            ...

    logger.debug("Creating new agent")
    # pylint: disable=unexpected-keyword-arg,no-member
    agent = AsyncAgent(
        client,  # type: ignore[arg-type]
        model=model_id,
        instructions=system_prompt,
        # type: ignore[call-arg]
        # input_shields=available_input_shields if available_input_shields else [],
        # type: ignore[call-arg]
        # output_shields=available_output_shields if available_output_shields else [],
        tool_parser=None if no_tools else GraniteToolParser.get_parser(model_id),
        enable_session_persistence=True,  # type: ignore[call-arg]
    )
    await agent.initialize()  # type: ignore[attr-defined]

    if existing_agent_id and conversation_id:
        logger.debug("Existing conversation ID: %s", conversation_id)
        logger.debug("Existing agent ID: %s", existing_agent_id)
        # orphan_agent_id = agent.agent_id
        agent._agent_id = conversation_id  # type: ignore[assignment]  # pylint: disable=protected-access
        # await client.agents.delete(agent_id=orphan_agent_id)
        # sessions_response = await client.agents.session.list(agent_id=conversation_id)
        # logger.info("session response: %s", sessions_response)
        try:
            # session_id = str(sessions_response.data[0]["session_id"])
            ...
        except IndexError as e:
            logger.error("No sessions found for conversation %s", conversation_id)
            response = NotFoundResponse(
                resource="conversation", resource_id=conversation_id
            )
            raise HTTPException(**response.model_dump()) from e
    else:
        # conversation_id = agent.agent_id
        # pylint: enable=unexpected-keyword-arg,no-member
        logger.debug("New conversation ID: %s", conversation_id)
        session_id = await agent.create_session(get_suid())
        logger.debug("New session ID: %s", session_id)

    return agent, conversation_id, session_id  # type: ignore[return-value]


async def get_temp_agent(
    client: AsyncLlamaStackClient,
    model_id: str,
    system_prompt: str,
) -> tuple[AsyncAgent, str, str]:
    """
    Create a temporary, non-persistent agent and a new session for one-off operations.
    
    This agent is initialized without session persistence, shields, or tools and is intended for short-lived tasks such as validation or summarization.
    
    Parameters:
        model_id (str): Identifier of the model to instantiate the agent with.
        system_prompt (str): System instructions to initialize the agent with.
    
    Returns:
        tuple: A tuple (agent, session_id, conversation_id):
            - agent (AsyncAgent): The initialized temporary agent.
            - session_id (str): The newly created session identifier.
            - conversation_id (str | None): The conversation identifier (always `None` for temporary agents).
    """
    logger.debug("Creating temporary agent")
    # pylint: disable=unexpected-keyword-arg,no-member
    agent = AsyncAgent(
        client,  # type: ignore[arg-type]
        model=model_id,
        instructions=system_prompt,
        # type: ignore[call-arg]  # Temporary agent doesn't need persistence
        # enable_session_persistence=False,
    )
    await agent.initialize()  # type: ignore[attr-defined]

    # Generate new IDs for the temporary agent
    # conversation_id = agent.agent_id
    conversation_id = None
    # pylint: enable=unexpected-keyword-arg,no-member
    session_id = await agent.create_session(get_suid())

    return agent, session_id, conversation_id  # type: ignore[return-value]


def create_rag_chunks_dict(summary: TurnSummary) -> list[dict[str, Any]]:
    """
    Convert a TurnSummary's RAG chunks into a list of dictionaries containing `content`, `source`, and `score`.
    
    Returns:
        list[dict[str, Any]]: A list where each dict has keys `content`, `source`, and `score` extracted from the summary's RAG chunks.
    """
    return [
        {"content": chunk.content, "source": chunk.source, "score": chunk.score}
        for chunk in summary.rag_chunks
    ]


def _process_http_source(
    src: str, doc_urls: set[str]
) -> tuple[AnyUrl | None, str] | None:
    """
    Validate and deduplicate an HTTP source and produce a document title.
    
    Parameters:
        src (str): The source URL string to process.
        doc_urls (set[str]): Set of already-seen source strings; the function will add `src` to this set when it is new.
    
    Returns:
        tuple[AnyUrl | None, str] | None: A tuple (validated_url, doc_title) when `src` was not previously seen:
            - `validated_url`: an `AnyUrl` instance if `src` is a valid URL, or `None` if validation failed.
            - `doc_title`: the last path segment of the URL or `src` if no path segment is present.
        Returns `None` if `src` was already present in `doc_urls`.
    """
    if src not in doc_urls:
        doc_urls.add(src)
        try:
            validated_url = AnyUrl(src)
        except ValidationError:
            logger.warning("Invalid URL in chunk source: %s", src)
            validated_url = None

        doc_title = src.rsplit("/", 1)[-1] or src
        return (validated_url, doc_title)
    return None


def _process_document_id(
    src: str,
    doc_ids: set[str],
    doc_urls: set[str],
    metas_by_id: dict[str, dict[str, Any]],
    metadata_map: dict[str, Any] | None,
) -> tuple[AnyUrl | None, str] | None:
    """
    Derives a validated document URL and a display title from a document identifier while deduplicating processed IDs and URLs.
    
    Parameters:
        src (str): Document identifier to process.
        doc_ids (set[str]): Set of already-seen document IDs; the function adds `src` to this set.
        doc_urls (set[str]): Set of already-seen document URLs; the function adds discovered URLs to this set to avoid duplicates.
        metas_by_id (dict[str, dict[str, Any]]): Mapping of document IDs to metadata dicts that may contain `docs_url` and `title`.
        metadata_map (dict[str, Any] | None): If provided (truthy), indicates metadata is available and enables metadata lookup; when falsy, metadata lookup is skipped.
    
    Returns:
        tuple[AnyUrl | None, str] | None: `(validated_url, doc_title)` where `validated_url` is a validated `AnyUrl` or `None` and `doc_title` is the chosen title string; returns `None` if the `src` or its URL was already processed.
    """
    if src in doc_ids:
        return None
    doc_ids.add(src)

    meta = metas_by_id.get(src, {}) if metadata_map else {}
    doc_url = meta.get("docs_url")
    title = meta.get("title")
    # Type check to ensure we have the right types
    if not isinstance(doc_url, (str, type(None))):
        doc_url = None
    if not isinstance(title, (str, type(None))):
        title = None

    if doc_url:
        if doc_url in doc_urls:
            return None
        doc_urls.add(doc_url)

    try:
        validated_doc_url = None
        if doc_url and doc_url.startswith("http"):
            validated_doc_url = AnyUrl(doc_url)
    except ValidationError:
        logger.warning("Invalid URL in metadata: %s", doc_url)
        validated_doc_url = None

    doc_title = title or (doc_url.rsplit("/", 1)[-1] if doc_url else src)
    return (validated_doc_url, doc_title)


def _add_additional_metadata_docs(
    doc_urls: set[str],
    metas_by_id: dict[str, dict[str, Any]],
) -> list[tuple[AnyUrl | None, str]]:
    """
    Collect additional referenced documents from a metadata mapping.
    
    Scans each metadata value in `metas_by_id` for a string `docs_url` and a string `title`. For entries
    where `docs_url` is a string not already present in `doc_urls` and `title` is a string, the function
    attempts to validate an HTTP(S) URL and adds the URL to `doc_urls` to prevent duplicates. Each
    result is a tuple of the validated `AnyUrl` (or `None` if the URL is absent or invalid) and the
    title string.
    
    Parameters:
        doc_urls (set[str]): Mutable set of URLs already recorded; new valid `docs_url` values will be
            added to this set to avoid duplicates.
        metas_by_id (dict[str, dict[str, Any]]): Mapping of metadata IDs to metadata dictionaries. Each
            metadata dictionary may contain `docs_url` and `title` keys.
    
    Returns:
        list[tuple[AnyUrl | None, str]]: List of tuples where the first element is a validated `AnyUrl`
        or `None` (if the docs URL is missing or invalid) and the second element is the document title.
    """
    additional_entries: list[tuple[AnyUrl | None, str]] = []
    for meta in metas_by_id.values():
        doc_url = meta.get("docs_url")
        title = meta.get("title")  # Note: must be "title", not "Title"
        # Type check to ensure we have the right types
        if not isinstance(doc_url, (str, type(None))):
            doc_url = None
        if not isinstance(title, (str, type(None))):
            title = None
        if doc_url and doc_url not in doc_urls and title is not None:
            doc_urls.add(doc_url)
            try:
                validated_url = None
                if doc_url.startswith("http"):
                    validated_url = AnyUrl(doc_url)
            except ValidationError:
                logger.warning("Invalid URL in metadata_map: %s", doc_url)
                validated_url = None

            additional_entries.append((validated_url, title))
    return additional_entries


def _process_rag_chunks_for_documents(
    rag_chunks: list,
    metadata_map: dict[str, Any] | None = None,
) -> list[tuple[AnyUrl | None, str]]:
    """
    Convert a list of RAG chunks into referenced document entries.
    
    Processes each chunk's source to produce a list of (doc_url, doc_title) pairs: it validates and deduplicates HTTP URLs, resolves document references by ID using optional metadata_map, skips empty or default sources, and includes additional referenced documents found in metadata_map.
    
    Parameters:
        rag_chunks (list): Iterable of RAG chunk objects; each chunk must provide a `source` attribute (e.g., an HTTP URL or a document ID).
        metadata_map (dict[str, Any] | None): Optional mapping of document IDs to metadata dictionaries used to resolve titles and document URLs.
    
    Returns:
        list[tuple[AnyUrl | None, str]]: Ordered list of tuples where the first element is a validated URL object or `None` (if no URL is available) and the second element is the document title.
    """
    doc_urls: set[str] = set()
    doc_ids: set[str] = set()

    # Process metadata_map if provided
    metas_by_id: dict[str, dict[str, Any]] = {}
    if metadata_map:
        metas_by_id = {k: v for k, v in metadata_map.items() if isinstance(v, dict)}

    document_entries: list[tuple[AnyUrl | None, str]] = []

    for chunk in rag_chunks:
        src = chunk.source
        if not src or src == constants.DEFAULT_RAG_TOOL:
            continue

        if src.startswith("http"):
            entry = _process_http_source(src, doc_urls)
            if entry:
                document_entries.append(entry)
        else:
            entry = _process_document_id(
                src, doc_ids, doc_urls, metas_by_id, metadata_map
            )
            if entry:
                document_entries.append(entry)

    # Add any additional referenced documents from metadata_map not already present
    if metadata_map:
        additional_entries = _add_additional_metadata_docs(doc_urls, metas_by_id)
        document_entries.extend(additional_entries)

    return document_entries


def create_referenced_documents(
    rag_chunks: list,
    metadata_map: dict[str, Any] | None = None,
    return_dict_format: bool = False,
) -> list[ReferencedDocument] | list[dict[str, str | None]]:
    """
    Create referenced documents from RAG chunks, optionally enriching entries with provided metadata.
    
    Processes RAG chunks into a deduplicated list of document references and returns them either
    as ReferencedDocument instances or as simple dictionaries suitable for serialization.
    
    Parameters:
        rag_chunks (list): List of RAG chunk entries containing source information.
        metadata_map (dict[str, Any] | None): Optional mapping of document IDs to metadata used
            to enrich or resolve document URLs and titles.
        return_dict_format (bool): If True, return a list of dictionaries with keys
            `doc_url` and `doc_title`; if False, return a list of ReferencedDocument objects.
    
    Returns:
        list[ReferencedDocument] | list[dict[str, str | None]]: Referenced documents as objects
        when `return_dict_format` is False, otherwise a list of dictionaries with `doc_url`
        (string or None) and `doc_title` (string or None).
    """
    document_entries = _process_rag_chunks_for_documents(rag_chunks, metadata_map)

    if return_dict_format:
        return [
            {
                "doc_url": str(doc_url) if doc_url else None,
                "doc_title": doc_title,
            }
            for doc_url, doc_title in document_entries
        ]
    return [
        ReferencedDocument(doc_url=doc_url, doc_title=doc_title)
        for doc_url, doc_title in document_entries
    ]


# Backward compatibility functions
def create_referenced_documents_with_metadata(
    summary: TurnSummary, metadata_map: dict[str, Any]
) -> list[ReferencedDocument]:
    """
    Create ReferencedDocument objects from a TurnSummary's RAG chunks, using the provided metadata map to resolve or enrich document URLs and titles.
    
    Parameters:
        summary (TurnSummary): Summary object containing `rag_chunks` to be processed.
        metadata_map (dict[str, Any]): Metadata keyed by document id used to derive or enrich document `doc_url` and `doc_title`.
    
    Returns:
        list[ReferencedDocument]: ReferencedDocument objects with `doc_url` and `doc_title` populated; `doc_url` may be `None` if no valid URL could be determined.
    """
    document_entries = _process_rag_chunks_for_documents(
        summary.rag_chunks, metadata_map
    )
    return [
        ReferencedDocument(doc_url=doc_url, doc_title=doc_title)
        for doc_url, doc_title in document_entries
    ]


def create_referenced_documents_from_chunks(
    rag_chunks: list,
) -> list[ReferencedDocument]:
    """
    Construct ReferencedDocument objects from a list of RAG chunks for the query endpoint.
    
    Parameters:
        rag_chunks (list): List of RAG chunk entries containing source and metadata information.
    
    Returns:
        list[ReferencedDocument]: ReferencedDocument instances created from the chunks; each contains `doc_url` (validated URL or `None`) and `doc_title`.
    """
    document_entries = _process_rag_chunks_for_documents(rag_chunks)
    return [
        ReferencedDocument(doc_url=doc_url, doc_title=doc_title)
        for doc_url, doc_title in document_entries
    ]


# pylint: disable=R0913,R0917,too-many-locals
async def cleanup_after_streaming(
    user_id: str,
    conversation_id: str,
    model_id: str,
    provider_id: str,
    llama_stack_model_id: str,
    query_request: QueryRequest,
    summary: TurnSummary,
    metadata_map: dict[str, Any],
    started_at: str,
    client: AsyncLlamaStackClient,
    config: AppConfig,
    skip_userid_check: bool,
    get_topic_summary_func: Any,
    is_transcripts_enabled_func: Any,
    store_transcript_func: Any,
    persist_user_conversation_details_func: Any,
    rag_chunks: list[dict[str, Any]] | None = None,
) -> None:
    """
    Perform post-streaming persistence and cleanup for a completed turn.
    
    This stores an optional transcript, optionally generates an initial topic summary for new conversations, builds referenced documents and a cache entry from the turn summary, writes the cache entry if configured, and persists conversation metadata.
    
    Parameters:
        user_id (str): ID of the user who made the request.
        conversation_id (str): Conversation identifier.
        model_id (str): Short model identifier used for logging/storage.
        provider_id (str): Provider identifier used for logging/storage.
        llama_stack_model_id (str): Full provider/model identifier used when invoking the Llama Stack client.
        query_request (QueryRequest): Original query request object.
        summary (TurnSummary): Turn summary containing LLM response, tool calls, and RAG chunks.
        metadata_map (dict[str, Any]): Mapping of document metadata used to enrich referenced documents.
        started_at (str): ISO timestamp when the request started.
        client (AsyncLlamaStackClient): Llama Stack client instance.
        config (AppConfig): Application configuration.
        skip_userid_check (bool): If true, bypasses user-id checks when storing cache entries.
        get_topic_summary_func (callable): Async function(query: str, client, llama_stack_model_id) -> str that generates a topic summary.
        is_transcripts_enabled_func (callable): Function() -> bool that indicates whether transcript storage is enabled.
        store_transcript_func (callable): Function(...) that persists a transcript; called with user_id, conversation_id, model_id, provider_id, query_is_valid, query, query_request, summary, rag_chunks, truncated, attachments.
        persist_user_conversation_details_func (callable): Function(user_id, conversation_id, model, provider_id, topic_summary) that persists conversation metadata.
        rag_chunks (list[dict[str, Any]] | None): Optional RAG chunks to include with the transcript; when None, no RAG chunks are attached (Agent API may supply this).
    """
    # Store transcript if enabled
    if not is_transcripts_enabled_func():
        logger.debug("Transcript collection is disabled in the configuration")
    else:
        # Prepare attachments
        attachments = query_request.attachments or []

        # Determine rag_chunks: use provided value or empty list
        transcript_rag_chunks = rag_chunks if rag_chunks is not None else []

        store_transcript_func(
            user_id=user_id,
            conversation_id=conversation_id,
            model_id=model_id,
            provider_id=provider_id,
            query_is_valid=True,
            query=query_request.query,
            query_request=query_request,
            summary=summary,
            rag_chunks=transcript_rag_chunks,
            truncated=False,
            attachments=attachments,
        )

    # Get the initial topic summary for the conversation
    topic_summary = None
    with get_session() as session:
        existing_conversation = (
            session.query(UserConversation).filter_by(id=conversation_id).first()
        )
        if not existing_conversation:
            # Check if topic summary should be generated (default: True)
            should_generate = query_request.generate_topic_summary

            if should_generate:
                logger.debug("Generating topic summary for new conversation")
                topic_summary = await get_topic_summary_func(
                    query_request.query, client, llama_stack_model_id
                )
            else:
                logger.debug("Topic summary generation disabled by request parameter")
                topic_summary = None

    completed_at = datetime.now(UTC).strftime("%Y-%m-%dT%H:%M:%SZ")

    referenced_documents = create_referenced_documents_with_metadata(
        summary, metadata_map
    )

    cache_entry = CacheEntry(
        query=query_request.query,
        response=summary.llm_response,
        provider=provider_id,
        model=model_id,
        started_at=started_at,
        completed_at=completed_at,
        referenced_documents=referenced_documents if referenced_documents else None,
    )

    store_conversation_into_cache(
        config,
        user_id,
        conversation_id,
        cache_entry,
        skip_userid_check,
        topic_summary,
    )

    persist_user_conversation_details_func(
        user_id=user_id,
        conversation_id=conversation_id,
        model=model_id,
        provider_id=provider_id,
        topic_summary=topic_summary,
    )