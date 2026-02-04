"""Session ID utility functions."""

import uuid


def get_suid() -> str:
    """
    Generate a canonical RFC 4122 UUID4 string to use as a session identifier.
    
    Returns:
        A UUID4 string in hyphen-separated hex group form (e.g. "550e8400-e29b-41d4-a716-446655440000").
    """
    return str(uuid.uuid4())


def check_suid(suid: str) -> bool:
    """
    Determine whether a string is a valid session or conversation ID.
    
    Accepts standard RFC 4122 UUID strings, 48-character hexadecimal llama-stack IDs, or the same hex ID prefixed with "conv_". Non-string inputs are considered invalid.
    
    Parameters:
        suid (str): Candidate ID to validate.
    
    Returns:
        bool: `true` if `suid` is a valid UUID, a 48-character hex string, or `"conv_"` + 48-character hex; `false` otherwise.
    """
    if not isinstance(suid, str):
        return False

    # Strip 'conv_' prefix if present
    hex_part = suid[5:] if suid.startswith("conv_") else suid

    # Check for 48-char hex string (llama-stack conversation ID format)
    if len(hex_part) == 48:
        try:
            int(hex_part, 16)
            return True
        except ValueError:
            return False

    # Check for standard UUID format
    try:
        uuid.UUID(suid)
        return True
    except (ValueError, TypeError):
        return False


def normalize_conversation_id(conversation_id: str) -> str:
    """
    Normalize a conversation ID for database storage.

    Strips the 'conv_' prefix if present to store just the UUID part.
    This keeps IDs shorter and database-agnostic.

    Args:
        conversation_id: The conversation ID, possibly with 'conv_' prefix.

    Returns:
        str: The normalized ID without 'conv_' prefix.

    Examples:
        >>> normalize_conversation_id('conv_abc123')
        'abc123'
        >>> normalize_conversation_id('550e8400-e29b-41d4-a716-446655440000')
        '550e8400-e29b-41d4-a716-446655440000'
    """
    if conversation_id.startswith("conv_"):
        return conversation_id[5:]  # Remove 'conv_' prefix
    return conversation_id


def to_llama_stack_conversation_id(conversation_id: str) -> str:
    """
    Convert a database conversation ID to llama-stack format by ensuring it starts with the 'conv_' prefix.
    
    Parameters:
        conversation_id (str): Conversation ID that may already include the 'conv_' prefix.
    
    Returns:
        str: The conversation ID with a leading 'conv_' prefix (unchanged if it was already present).
    """
    if not conversation_id.startswith("conv_"):
        return f"conv_{conversation_id}"
    return conversation_id