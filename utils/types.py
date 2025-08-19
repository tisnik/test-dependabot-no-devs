"""Common types for the project."""

from typing import Optional

from llama_stack_client.lib.agents.tool_parser import ToolParser
from llama_stack_client.types.shared.completion_message import CompletionMessage
from llama_stack_client.types.shared.tool_call import ToolCall


class Singleton(type):
    """Metaclass for Singleton support."""

    _instances = {}  # type: ignore

    def __call__(cls, *args, **kwargs):  # type: ignore
        """
        Return the single cached instance of the class, creating and caching it on first call.
        
        If an instance for `cls` does not yet exist in the class-level `_instances` mapping it will be constructed by delegating to the superclass `__call__` and stored. Subsequent calls return the same instance.
        
        Returns:
            object: The singleton instance for this class.
        """
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]


# See https://github.com/meta-llama/llama-stack-client-python/issues/206
class GraniteToolParser(ToolParser):
    """Workaround for 'tool_calls' with granite models."""

    def get_tool_calls(self, output_message: CompletionMessage) -> list[ToolCall]:
        """
        Return the `tool_calls` list from a CompletionMessage, or an empty list if none are present.
        
        If `output_message` is falsy or has no `tool_calls` attribute / it is empty, this returns an empty list rather than None.
        
        Parameters:
            output_message (CompletionMessage | None): Completion message potentially containing `tool_calls`.
        
        Returns:
            list[ToolCall]: The list of tool call entries extracted from `output_message`, or an empty list.
        """
        if output_message and output_message.tool_calls:
            return output_message.tool_calls
        return []

    @staticmethod
    def get_parser(model_id: str) -> Optional[ToolParser]:
        """
        Return a GraniteToolParser when the model identifier denotes a Granite model; otherwise return None.
        
        Parameters:
            model_id (str): Model identifier string checked case-insensitively. If it starts with "granite", a GraniteToolParser instance is returned.
        
        Returns:
            Optional[ToolParser]: GraniteToolParser for Granite models, or None if `model_id` is falsy or does not start with "granite".
        """
        if model_id and model_id.lower().startswith("granite"):
            return GraniteToolParser()
        return None
