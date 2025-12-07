"""Transcript handling.

Transcripts are a log of individual query/response pairs that get
stored on disk for later analysis
"""

from datetime import UTC, datetime
import json
import logging
import os
from pathlib import Path
import hashlib

from configuration import configuration
from models.requests import Attachment, QueryRequest
from utils.suid import get_suid
from utils.types import TurnSummary

logger = logging.getLogger("utils.transcripts")


def _hash_user_id(user_id: str) -> str:
    """
    Compute a SHA-256 hexadecimal digest of the given user identifier.
    
    Returns:
        Hexadecimal SHA-256 digest of the provided `user_id`.
    """
    return hashlib.sha256(user_id.encode("utf-8")).hexdigest()


def construct_transcripts_path(user_id: str, conversation_id: str) -> Path:
    """
    Return filesystem path for storing transcripts for a user conversation.
    
    Constructs a path under the configured transcripts_storage using a SHA-256 hash
    of `user_id` as a directory component and a sanitized `conversation_id` as a
    subdirectory.
    
    Parameters:
        user_id (str): User identifier; its SHA-256 hash is used as the directory name.
        conversation_id (str): Conversation identifier; normalized to a safe path segment.
    
    Returns:
        Path: Filesystem path where transcripts for the specified user and conversation should be stored.
    """
    # these two normalizations are required by Snyk as it detects
    # this Path sanitization pattern
    hashed_user_id = _hash_user_id(user_id)
    uid = os.path.normpath("/" + hashed_user_id).lstrip("/")
    cid = os.path.normpath("/" + conversation_id).lstrip("/")
    file_path = (
        configuration.user_data_collection_configuration.transcripts_storage or ""
    )
    return Path(file_path, uid, cid)


def store_transcript(  # pylint: disable=too-many-arguments,too-many-positional-arguments,too-many-locals
    user_id: str,
    conversation_id: str,
    model_id: str,
    provider_id: str | None,
    query_is_valid: bool,
    query: str,
    query_request: QueryRequest,
    summary: TurnSummary,
    rag_chunks: list[dict],
    truncated: bool,
    attachments: list[Attachment],
) -> None:
    """
    Persist a transcript for a user conversation to the configured local transcripts storage.
    
    Parameters:
        user_id (str): User identifier (UUID); stored hashed for privacy.
        conversation_id (str): Conversation identifier (UUID) used to partition storage.
        model_id (str): Identifier of the model that produced the response.
        provider_id (str | None): Optional provider identifier for the model.
        query_is_valid (bool): Result of upstream query validation for this turn.
        query (str): The (possibly redacted) user query text.
        query_request (QueryRequest): Request metadata containing provider and model for the original query.
        summary (TurnSummary): Turn summary containing the LLM response and any tool call records.
        rag_chunks (list[dict]): Serialized RAG chunk dictionaries associated with the turn.
        truncated (bool): Whether conversation history was truncated when producing the response.
        attachments (list[Attachment]): Attachments referenced by the query; each will be serialized.
    
    Raises:
        OSError: If the transcript file cannot be written to disk.
    """
    transcripts_path = construct_transcripts_path(user_id, conversation_id)
    transcripts_path.mkdir(parents=True, exist_ok=True)

    hashed_user_id = _hash_user_id(user_id)

    data_to_store = {
        "metadata": {
            "provider": provider_id,
            "model": model_id,
            "query_provider": query_request.provider,
            "query_model": query_request.model,
            "user_id": hashed_user_id,
            "conversation_id": conversation_id,
            "timestamp": datetime.now(UTC).isoformat(),
        },
        "redacted_query": query,
        "query_is_valid": query_is_valid,
        "llm_response": summary.llm_response,
        "rag_chunks": rag_chunks,
        "truncated": truncated,
        "attachments": [attachment.model_dump() for attachment in attachments],
        "tool_calls": [tc.model_dump() for tc in summary.tool_calls],
    }

    # stores feedback in a file under unique uuid
    transcript_file_path = transcripts_path / f"{get_suid()}.json"
    try:
        with open(transcript_file_path, "w", encoding="utf-8") as transcript_file:
            json.dump(data_to_store, transcript_file)
    except (IOError, OSError) as e:
        logger.error("Failed to store transcript into %s: %s", transcript_file_path, e)
        raise

    logger.info("Transcript successfully stored at: %s", transcript_file_path)