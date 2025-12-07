"""Common types for the project."""

from typing import Any, Optional
import json
from llama_stack_client.lib.agents.event_logger import interleaved_content_as_str
from llama_stack_client.lib.agents.tool_parser import ToolParser
from llama_stack_client.types.shared.completion_message import CompletionMessage
from llama_stack_client.types.shared.tool_call import ToolCall
from llama_stack_client.types.tool_execution_step import ToolExecutionStep
from pydantic import BaseModel
from models.responses import RAGChunk
from constants import DEFAULT_RAG_TOOL


class Singleton(type):
    """Metaclass for Singleton support."""

    _instances = {}  # type: ignore

    def __call__(cls, *args, **kwargs):  # type: ignore
        """
        Get the cached singleton instance for this class, creating and caching it on first invocation.
        
        Returns:
            The singleton instance of the class.
        """
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]


# See https://github.com/meta-llama/llama-stack-client-python/issues/206
class GraniteToolParser(ToolParser):
    """Workaround for 'tool_calls' with granite models."""

    def get_tool_calls(self, output_message: CompletionMessage) -> list[ToolCall]:
        """
        Retrieve tool call entries from a CompletionMessage.
        
        Parameters:
            output_message (CompletionMessage | None): CompletionMessage that may contain `tool_calls`.
        
        Returns:
            list[ToolCall]: The list of `ToolCall` objects extracted from `output_message`, or an empty list if none are present.
        """
        if output_message and output_message.tool_calls:
            return output_message.tool_calls
        return []

    @staticmethod
    def get_parser(model_id: str) -> Optional[ToolParser]:
        """
        Get a GraniteToolParser for Granite model identifiers, or None.
        
        Parameters:
            model_id (str): Model identifier checked case-insensitively.
        
        Returns:
            Optional[ToolParser]: A GraniteToolParser when `model_id` is truthy and starts with "granite" (case-insensitive); otherwise `None`.
        """
        if model_id and model_id.lower().startswith("granite"):
            return GraniteToolParser()
        return None


class ToolCallSummary(BaseModel):
    """Represents a tool call for data collection.

    Use our own tool call model to keep things consistent across llama
    upgrades or if we used something besides llama in the future.
    """

    # ID of the call itself
    id: str
    # Name of the tool used
    name: str
    # Arguments to the tool call
    args: str | dict[Any, Any]
    response: str | None


class TurnSummary(BaseModel):
    """Summary of a turn in llama stack."""

    llm_response: str
    tool_calls: list[ToolCallSummary]
    rag_chunks: list[RAGChunk] = []

    def append_tool_calls_from_llama(self, tec: ToolExecutionStep) -> None:
        """
        Collects tool call summaries from a Llama ToolExecutionStep and appends them to this turn's tool_calls.
        
        Parameters:
            tec (ToolExecutionStep): Execution record containing `tool_calls` and `tool_responses`; calls and responses are matched by `call_id`. For each call, a ToolCallSummary is appended with the call id, tool name, arguments, and the response content (if any).
        
        Notes:
            If a call's tool name equals DEFAULT_RAG_TOOL and a textual response is present, RAG chunks are extracted from that response and appended to `rag_chunks`.
        """
        calls_by_id = {tc.call_id: tc for tc in tec.tool_calls}
        responses_by_id = {tc.call_id: tc for tc in tec.tool_responses}
        for call_id, tc in calls_by_id.items():
            resp = responses_by_id.get(call_id)
            response_content = (
                interleaved_content_as_str(resp.content) if resp else None
            )

            self.tool_calls.append(
                ToolCallSummary(
                    id=call_id,
                    name=tc.tool_name,
                    args=tc.arguments,
                    response=response_content,
                )
            )

            # Extract RAG chunks from knowledge_search tool responses
            if tc.tool_name == DEFAULT_RAG_TOOL and resp and response_content:
                self._extract_rag_chunks_from_response(response_content)

    def _extract_rag_chunks_from_response(self, response_content: str) -> None:
        """
        Parse a tool response string and append extracted RAG chunks to this TurnSummary's rag_chunks list.
        
        Attempts to parse `response_content` as JSON and extract chunks in either of two formats:
        - A dict containing a "chunks" list: each item's "content", "source", and "score" are used.
        - A top-level list of chunk objects: for dict items, "content", "source", and "score" are used; non-dict items are stringified into the chunk content.
        
        If JSON parsing fails or an unexpected structure/error occurs and `response_content` contains non-whitespace characters, the entire `response_content` is appended as a single RAGChunk with `source=DEFAULT_RAG_TOOL` and `score=None`. Empty or whitespace-only `response_content` is ignored.
        """
        try:
            # Parse the response to get chunks
            # Try JSON first
            try:
                data = json.loads(response_content)
                if isinstance(data, dict) and "chunks" in data:
                    for chunk in data["chunks"]:
                        self.rag_chunks.append(
                            RAGChunk(
                                content=chunk.get("content", ""),
                                source=chunk.get("source"),
                                score=chunk.get("score"),
                            )
                        )
                elif isinstance(data, list):
                    # Handle list of chunks
                    for chunk in data:
                        if isinstance(chunk, dict):
                            self.rag_chunks.append(
                                RAGChunk(
                                    content=chunk.get("content", str(chunk)),
                                    source=chunk.get("source"),
                                    score=chunk.get("score"),
                                )
                            )
            except json.JSONDecodeError:
                # If not JSON, treat the entire response as a single chunk
                if response_content.strip():
                    self.rag_chunks.append(
                        RAGChunk(
                            content=response_content,
                            source=DEFAULT_RAG_TOOL,
                            score=None,
                        )
                    )
        except (KeyError, AttributeError, TypeError, ValueError):
            # Treat response as single chunk on data access/structure errors
            if response_content.strip():
                self.rag_chunks.append(
                    RAGChunk(
                        content=response_content, source=DEFAULT_RAG_TOOL, score=None
                    )
                )