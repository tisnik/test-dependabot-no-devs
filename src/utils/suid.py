"""Session ID utility functions."""

import uuid


def get_suid() -> str:
    """
    Generate a unique session identifier as an RFC 4122 UUID4 string.
    
    The returned value is the canonical UUID string (hex groups separated by hyphens).
    
    Returns:
        str: A UUID4 string suitable for use as a session identifier.
    """
    return str(uuid.uuid4())


def check_suid(suid: str) -> bool:
    """
    Determine whether a value is a valid UUID session identifier.
    
    Parameters:
        suid (str | bytes): UUID to validate — either a canonical UUID string or 16-byte UUID bytes.
    
    Returns:
        True if `suid` represents a valid UUID, False otherwise.
    """
    try:
        # accepts strings and bytes only
        uuid.UUID(suid)
        return True
    except (ValueError, TypeError):
        return False