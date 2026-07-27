"""Authentication utility functions."""

from fastapi import HTTPException
from starlette.datastructures import Headers


def extract_user_token(headers: Headers) -> str:
    """
    Extracts the bearer token from the HTTP Authorization header in the provided headers.
    
    Raises an HTTPException with status 400 if the Authorization header is missing or does not contain a valid bearer token.
    
    Returns:
        str: The extracted bearer token.
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
