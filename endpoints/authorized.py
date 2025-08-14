"""Handler for REST API call to authorized endpoint."""

import logging
from typing import Any

from fastapi import APIRouter, Depends

from auth import get_auth_dependency
from models.responses import AuthorizedResponse, UnauthorizedResponse, ForbiddenResponse

logger = logging.getLogger(__name__)
router = APIRouter(tags=["authorized"])
auth_dependency = get_auth_dependency()


authorized_responses: dict[int | str, dict[str, Any]] = {
    200: {
        "description": "The user is logged-in and authorized to access OLS",
        "model": AuthorizedResponse,
    },
    400: {
        "description": "Missing or invalid credentials provided by client",
        "model": UnauthorizedResponse,
    },
    403: {
        "description": "User is not authorized",
        "model": ForbiddenResponse,
    },
}


@router.post("/authorized", responses=authorized_responses)
async def authorized_endpoint_handler(
    auth: Any = Depends(auth_dependency),
) -> AuthorizedResponse:
    """
    Return an AuthorizedResponse for the authenticated user.
    
    The function expects the dependency `auth` to provide a 3-tuple (user_id, username, token). The token is intentionally ignored and not included in the response. Returns an AuthorizedResponse containing the user's id and username.
    """
    # Ignore the user token, we should not return it in the response
    user_id, user_name, _ = auth
    return AuthorizedResponse(user_id=user_id, username=user_name)
