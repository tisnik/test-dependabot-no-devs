"""Utility functions for processing Responses API output."""

from typing import Any


def extract_text_from_response_output_item(output_item: Any) -> str:
    """
    Extract text content from an assistant message output item.
    
    Parses a Responses API output item and returns concatenated text from assistant messages. If the item is not a message with role "assistant" or contains no text, returns an empty string.
    
    Parameters:
        output_item (Any): Output item expected to have attributes `type`, `role`, and `content`. `content` may be a string or a list of parts where parts can be strings, objects with a `text` attribute, objects with a `refusal` attribute, or dicts containing `text`/`refusal`.
    
    Returns:
        str: Concatenated text extracted from the assistant message, or an empty string if none is found.
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