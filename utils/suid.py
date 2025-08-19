"""Session ID utility functions."""

import uuid


def get_suid() -> str:
    """
    Return a new session ID as a UUID4 string.
    
    The value is a canonical RFC 4122 UUID (hex groups separated by hyphens) generated with uuid.uuid4().
    
    Returns:
        str: A UUID4 string suitable for use as a session identifier.
    """
    return str(uuid.uuid4())


def check_suid(suid: str) -> bool:
    """
    Return True if the given value is a valid UUID-based session ID, False otherwise.
    
    Parameters:
        suid (str | bytes): UUID value to validate — accepts a UUID string or its byte representation.
    
    Notes:
        Validation is performed by attempting to construct uuid.UUID(suid); invalid formats or types result in False.
    """
    try:
        # accepts strings and bytes only
        uuid.UUID(suid)
        return True
    except (ValueError, TypeError):
        return False
