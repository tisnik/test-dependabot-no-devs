"""Session ID utility functions."""

import uuid


def get_suid() -> str:
    """
    Generate and return a unique session ID as a string using UUID version 4.
    
    Returns:
        str: A unique session ID in UUID4 string format.
    """
    return str(uuid.uuid4())


def check_suid(suid: str) -> bool:
    """
    Validate whether the provided string is a valid UUID-based session ID.
    
    Returns:
        bool: True if the input string is a valid UUID, False otherwise.
    """
    try:
        # accepts strings and bytes only
        uuid.UUID(suid)
        return True
    except (ValueError, TypeError):
        return False
