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
        Creates or returns the single instance of the class, ensuring the Singleton pattern is enforced.
        
        Returns:
            The sole instance of the class using this metaclass.
        """
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]


# See https://github.com/meta-llama/llama-stack-client-python/issues/206
class GraniteToolParser(ToolParser):
    """Workaround for 'tool_calls' with granite models."""

    def get_tool_calls(self, output_message: CompletionMessage) -> list[ToolCall]:
        """
        Retrieve the list of tool calls from a CompletionMessage if available.
        
        Returns:
            A list of ToolCall objects from the output_message, or an empty list if none are present.
        """
        if output_message and output_message.tool_calls:
            return output_message.tool_calls
        return []

    @staticmethod
    def get_parser(model_id: str) -> Optional[ToolParser]:
        """
        Return a GraniteToolParser instance if the model ID indicates a granite model; otherwise, return None.
        
        Parameters:
            model_id (str): The identifier of the model to check.
        
        Returns:
            Optional[ToolParser]: A GraniteToolParser instance if the model ID starts with "granite" (case-insensitive), or None otherwise.
        """
        if model_id and model_id.lower().startswith("granite"):
            return GraniteToolParser()
        return None
