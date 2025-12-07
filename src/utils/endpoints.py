"""Utility functions for endpoint handlers."""

from contextlib import suppress
from typing import Any
from fastapi import HTTPException, status
from llama_stack_client._client import AsyncLlamaStackClient
from llama_stack_client.lib.agents.agent import AsyncAgent
from pydantic import AnyUrl, ValidationError

import constants
from models.cache_entry import CacheEntry
from models.requests import QueryRequest
from models.responses import ReferencedDocument
from models.database.conversations import UserConversation
from models.config import Action
from app.database import get_session
from configuration import AppConfig
from utils.suid import get_suid
from utils.types import TurnSummary
from utils.types import GraniteToolParser


from log import get_logger

logger = get_logger(__name__)


def delete_conversation(conversation_id: str) -> None:
    """
    Delete the UserConversation with the given ID from the local database.
    
    If a conversation with the ID exists, it is removed and the change is committed; if not, no change is made and a not-found condition is logged.
    
    Parameters:
        conversation_id (str): ID of the conversation to delete.
    """
    with get_session() as session:
        db_conversation = (
            session.query(UserConversation).filter_by(id=conversation_id).first()
        )
        if db_conversation:
            session.delete(db_conversation)
            session.commit()
            logger.info("Deleted conversation %s from local database", conversation_id)
        else:
            logger.info(
                "Conversation %s not found in local database, it may have already been deleted",
                conversation_id,
            )


def retrieve_conversation(conversation_id: str) -> UserConversation | None:
    """
    Retrieve a UserConversation by its ID.
    
    Returns:
        The UserConversation with the given ID, or `None` if no matching conversation exists.
    """
    with get_session() as session:
        return session.query(UserConversation).filter_by(id=conversation_id).first()


def validate_conversation_ownership(
    user_id: str, conversation_id: str, others_allowed: bool = False
) -> UserConversation | None:
    """
    Validate whether a conversation matches the provided identifiers.
    
    If `others_allowed` is True, returns any conversation with the given `conversation_id`; otherwise restricts to conversations owned by `user_id`.
    
    Parameters:
        others_allowed (bool): If True, allow conversations not owned by `user_id`.
    
    Returns:
        UserConversation | None: `UserConversation` if a matching conversation is found, `None` otherwise.
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
    
    Parameters:
        conversation_id (str): ID of the conversation to check.
        user_id (str): ID of the user requesting access.
        others_allowed (bool): If True, access to conversations owned by other users is permitted.
    
    Returns:
        `true` if the user may access the conversation, `false` otherwise.
    
    Notes:
        If `others_allowed` is False and the conversation does not exist, access is allowed.
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
    Ensure the application configuration object is present.

    Raises:
        HTTPException: HTTP 500 Internal Server Error with detail `{"response":
        "Configuration is not loaded"}` when `config` is None.
    """
    if config is None:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={"response": "Configuration is not loaded"},
        )


def get_system_prompt(query_request: QueryRequest, config: AppConfig) -> str:
    """
    Determine the system prompt to apply to a query using per-request, profile, configuration, and default precedence.
    
    Precedence (highest → lowest):
    1) `query_request.system_prompt` (unless per-request prompts are disabled),
    2) customization profile default prompt from `config.customization.custom_profile`,
    3) `config.customization.system_prompt`,
    4) `constants.DEFAULT_SYSTEM_PROMPT`.
    
    If `config.customization.disable_query_system_prompt` is true and `query_request.system_prompt` is provided, an HTTP 422 Unprocessable Entity is raised instructing removal of the field.
    
    Parameters:
        query_request (QueryRequest): The incoming query payload; may include `system_prompt`.
        config (AppConfig): Application configuration that may contain customization flags, a custom profile, or a configured system prompt.
    
    Returns:
        str: The resolved system prompt to use for the request.
    """
    system_prompt_disabled = (
        config.customization is not None
        and config.customization.disable_query_system_prompt
    )
    if system_prompt_disabled and query_request.system_prompt:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail={
                "response": (
                    "This instance does not support customizing the system prompt in the "
                    "query request (disable_query_system_prompt is set). Please remove the "
                    "system_prompt field from your request."
                )
            },
        )

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
    Resolve the system prompt to use for topic summaries.
    
    Checks for a topic summary prompt in the active custom profile and returns it if present; otherwise returns the module default prompt.
    
    Returns:
        str: The topic summary system prompt from the custom profile if available, otherwise the default prompt.
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
    Validate whether a request is allowed to override the model or provider under RBAC.
    
    Raises:
        HTTPException: with status 403 if `query_request` specifies `model` or `provider` and
        `Action.MODEL_OVERRIDE` is not present in `authorized_actions`.
    """
    if (query_request.model is not None or query_request.provider is not None) and (
        Action.MODEL_OVERRIDE not in authorized_actions
    ):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail={
                "response": (
                    "This instance does not permit overriding model/provider in the query request "
                    "(missing permission: MODEL_OVERRIDE). Please remove the model and provider "
                    "fields from your request."
                )
            },
        )


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
    Persist a cache entry and optional topic summary for a user's conversation when a conversation cache is configured.
    
    Parameters:
        config (AppConfig): Application configuration containing conversation cache settings and instance.
        user_id (str): ID of the user owning the conversation.
        conversation_id (str): ID of the conversation to update in the cache.
        cache_entry (CacheEntry): Entry to insert or append into the conversation cache.
        _skip_userid_check (bool): If True, bypass user ID checks performed by the cache when inserting/appending or setting the topic summary.
        topic_summary (str | None): Optional topic summary to store alongside the cache entry; ignored if None or empty.
    
    Notes:
        - If the configuration specifies a conversation cache type but the cache instance is not initialized, the function returns without storing anything.
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


# # pylint: disable=R0913,R0917
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
    Create or reuse an AsyncAgent with session persistence.

    Return the agent, conversation and session IDs.

    If a conversation_id is provided, the function attempts to retrieve the
    existing agent and, on success, rebinds a newly created agent instance to
    that conversation (deleting the temporary/orphan agent) and returns the
    first existing session_id for the conversation. If no conversation_id is
    provided or the existing agent cannot be retrieved, a new agent and session
    are created.

    Parameters:
        model_id (str): Identifier of the model to instantiate the agent with.
        system_prompt (str): Instructions/system prompt to initialize the agent with.

        available_input_shields (list[str]): Input shields to apply to the
        agent; empty list used if None/empty.

        available_output_shields (list[str]): Output shields to apply to the
        agent; empty list used if None/empty.

        conversation_id (str | None): If provided, attempt to reuse the agent
        for this conversation; otherwise a new conversation_id is created.

        no_tools (bool): When True, disables tool parsing for the agent (uses no tool parser).

    Returns:
        tuple[AsyncAgent, str, str]: A tuple of (agent, conversation_id, session_id).

    Raises:
        HTTPException: Raises HTTP 404 Not Found if an attempt to reuse a
        conversation succeeds in retrieving the agent but no sessions are found
        for that conversation.

    Side effects:
        - May delete an orphan agent when rebinding a newly created agent to an
          existing conversation_id.
        - Initializes the agent and may create a new session.
    """
    existing_agent_id = None
    if conversation_id:
        with suppress(ValueError):
            agent_response = await client.agents.retrieve(agent_id=conversation_id)
            existing_agent_id = agent_response.agent_id

    logger.debug("Creating new agent")
    agent = AsyncAgent(
        client,  # type: ignore[arg-type]
        model=model_id,
        instructions=system_prompt,
        input_shields=available_input_shields if available_input_shields else [],
        output_shields=available_output_shields if available_output_shields else [],
        tool_parser=None if no_tools else GraniteToolParser.get_parser(model_id),
        enable_session_persistence=True,
    )
    await agent.initialize()

    if existing_agent_id and conversation_id:
        logger.debug("Existing conversation ID: %s", conversation_id)
        logger.debug("Existing agent ID: %s", existing_agent_id)
        orphan_agent_id = agent.agent_id
        agent._agent_id = conversation_id  # type: ignore[assignment]  # pylint: disable=protected-access
        await client.agents.delete(agent_id=orphan_agent_id)
        sessions_response = await client.agents.session.list(agent_id=conversation_id)
        logger.info("session response: %s", sessions_response)
        try:
            session_id = str(sessions_response.data[0]["session_id"])
        except IndexError as e:
            logger.error("No sessions found for conversation %s", conversation_id)
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail={
                    "response": "Conversation not found",
                    "cause": f"Conversation {conversation_id} could not be retrieved.",
                },
            ) from e
    else:
        conversation_id = agent.agent_id
        logger.debug("New conversation ID: %s", conversation_id)
        session_id = await agent.create_session(get_suid())
        logger.debug("New session ID: %s", session_id)

    return agent, conversation_id, session_id


async def get_temp_agent(
    client: AsyncLlamaStackClient,
    model_id: str,
    system_prompt: str,
) -> tuple[AsyncAgent, str, str]:
    """
    Create a temporary, non-persistent AsyncAgent configured with the provided model and system prompt.
    
    This agent is created for short-lived operations and does not enable session persistence, shields, or tools.
    
    Parameters:
        client (AsyncLlamaStackClient): Client used to construct the agent.
        model_id (str): Identifier of the model to instantiate.
        system_prompt (str): System-level instructions provided to the agent.
    
    Returns:
        tuple[AsyncAgent, str, str]: A tuple containing the created agent, the new `session_id`, and the new `conversation_id` (in that order).
    """
    logger.debug("Creating temporary agent")
    agent = AsyncAgent(
        client,  # type: ignore[arg-type]
        model=model_id,
        instructions=system_prompt,
        enable_session_persistence=False,  # Temporary agent doesn't need persistence
    )
    await agent.initialize()

    # Generate new IDs for the temporary agent
    conversation_id = agent.agent_id
    session_id = await agent.create_session(get_suid())

    return agent, session_id, conversation_id


def create_rag_chunks_dict(summary: TurnSummary) -> list[dict[str, Any]]:
    """
    Create a serializable list of RAG chunk dictionaries for streaming responses.
    
    Parameters:
        summary (TurnSummary): TurnSummary containing the RAG chunks to convert.
    
    Returns:
        list[dict[str, Any]]: List of dictionaries where each dictionary has keys `content`, `source`, and `score`.
    """
    return [
        {"content": chunk.content, "source": chunk.source, "score": chunk.score}
        for chunk in summary.rag_chunks
    ]


def _process_http_source(
    src: str, doc_urls: set[str]
) -> tuple[AnyUrl | None, str] | None:
    """
    Validate and deduplicate an HTTP source URL and produce its display title.
    
    Parameters:
        src (str): The source URL string to process.
        doc_urls (set[str]): A set of already-seen source strings; the function will add `src` to this set when processing.
    
    Returns:
        tuple[AnyUrl | None, str] | None: A tuple `(validated_url, doc_title)` where `validated_url` is a validated `AnyUrl`
        or `None` if validation failed, and `doc_title` is the URL's last path segment (or the original `src` if none).
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
    Resolve a referenced document ID into a validated document URL and title, deduplicating by ID and URL.
    
    Parameters:
    	src (str): Document identifier to process.
    	doc_ids (set[str]): Set of already-seen document IDs; `src` will be added to this set.
    	doc_urls (set[str]): Set of already-seen document URLs; a discovered URL will be added to this set to prevent duplicates.
    	metas_by_id (dict[str, dict[str, Any]]): Mapping of document IDs to metadata dictionaries; used to look up `docs_url` and `title`.
    	metadata_map (dict[str, Any] | None): If provided, enables metadata lookup; if None, metadata lookup is skipped.
    
    Returns:
    	tuple[AnyUrl | None, str] | None: A pair of (validated AnyUrl or None, document title). Returns `None` if the ID or its URL was already processed. If a `docs_url` is present and starts with "http", it will be validated; invalid URLs yield `None` for the URL and produce a warning in logs.
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
    Collect additional referenced documents from a metadata map that are not already present.
    
    Iterates the provided metadata-by-id mapping and for each entry that contains a string `docs_url`
    and a string `title` (exact key "title"), adds a tuple of (validated_url, title) to the result
    when the URL is not already in `doc_urls`. The function mutates `doc_urls` by adding new
    string URLs that are accepted. For `docs_url` values that start with "http", an `AnyUrl` is
    constructed; if validation fails the tuple contains `None` for the URL and a warning is logged.
    
    Parameters:
        doc_urls (set[str]): Set of already-seen document URL strings; new valid URLs are added.
        metas_by_id (dict[str, dict[str, Any]]): Mapping from document id to its metadata dict.
    
    Returns:
        list[tuple[AnyUrl | None, str]]: A list of (validated_url_or_None, title) pairs for additional documents.
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
    Extract referenced documents from a list of RAG chunks, returning validated (doc_url, doc_title) pairs.
    
    Parameters:
        rag_chunks (list): Iterable of RAG chunk objects; each chunk must expose a `source` attribute.
        metadata_map (dict[str, Any] | None): Optional mapping of document IDs to metadata dicts used to resolve non-HTTP sources and to include additional referenced documents.
    
    Returns:
        list[tuple[AnyUrl | None, str]]: Ordered list of tuples where each tuple is (validated_doc_url or None, doc_title).
            `validated_doc_url` is an AnyUrl when the source could be validated as an HTTP(S) URL, or `None` when no URL is available.
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
    Produce referenced documents from RAG chunks, optionally enriching them with metadata.
    
    Parameters:
        rag_chunks (list): RAG chunks containing source references to convert into referenced documents.
        metadata_map (dict[str, Any] | None): Optional mapping of document metadata keyed by document ID used to enrich or add referenced documents.
        return_dict_format (bool): If True, return a list of dictionaries with keys `doc_url` and `doc_title`; otherwise return a list of `ReferencedDocument` objects.
    
    Returns:
        list[ReferencedDocument] | list[dict[str, str | None]]: A list of `ReferencedDocument` instances when `return_dict_format` is False; otherwise a list of dictionaries each containing:
            - `doc_url` (str | None): The document URL as a string, or `None` if unavailable.
            - `doc_title` (str | None): The document title, or `None` if unavailable.
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
    Produce ReferencedDocument objects from a TurnSummary's RAG chunks, using the provided metadata_map to enrich document references.
    
    Parameters:
        summary (TurnSummary): The turn summary containing RAG chunks to extract referenced documents from.
        metadata_map (dict[str, Any]): Mapping of document IDs to metadata used to resolve or enrich document URLs and titles.
    
    Returns:
        list[ReferencedDocument]: A list of ReferencedDocument objects with `doc_url` (which may be None) and `doc_title`.
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
    Produce ReferencedDocument objects extracted from RAG chunks for the query endpoint.
    
    Parameters:
        rag_chunks (list): Iterable of RAG chunk entries to scan for document references.
    
    Returns:
        list[ReferencedDocument]: A list of ReferencedDocument objects with `doc_url` and `doc_title` extracted from the chunks; `doc_url` may be `None` if no valid URL was found.
    """
    document_entries = _process_rag_chunks_for_documents(rag_chunks)
    return [
        ReferencedDocument(doc_url=doc_url, doc_title=doc_title)
        for doc_url, doc_title in document_entries
    ]