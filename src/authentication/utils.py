"""Authentication utility functions."""

from fastapi import HTTPException
from starlette.datastructures import Headers


def extract_user_token(headers: Headers) -> str:
    """
    Extract the bearer token from the HTTP Authorization header.
    
    Parameters:
        headers (Headers): Incoming request headers from which the Authorization header will be read.
    
    Returns:
        str: The bearer token string extracted from the header.
    
    Raises:
        HTTPException: Raised with status_code=400 when the Authorization header is missing or not a valid Bearer token.
    """
    authorization_header = headers.get("Authorization")
    if not authorization_header:
        raise HTTPException(status_code=400, detail="No Authorization header found")

    scheme_and_token = authorization_header.strip().split()
    if len(scheme_and_token) != 2 or scheme_and_token[0].lower() != "bearer":
        raise HTTPException(
            status_code=400, detail="No token found in Authorization header"
        )

    return scheme_and_token[1]