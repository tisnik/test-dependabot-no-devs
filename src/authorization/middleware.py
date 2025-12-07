"""Authorization middleware and decorators."""

import logging
from functools import wraps, lru_cache
from typing import Any, Callable, Tuple
from fastapi import HTTPException, status
from starlette.requests import Request

from authorization.resolvers import (
    AccessResolver,
    GenericAccessResolver,
    JwtRolesResolver,
    NoopAccessResolver,
    NoopRolesResolver,
    RolesResolver,
)
from models.config import Action
from configuration import configuration
import constants

logger = logging.getLogger(__name__)


@lru_cache(maxsize=1)
def get_authorization_resolvers() -> Tuple[RolesResolver, AccessResolver]:
    """
    Return the configured RolesResolver and AccessResolver based on authentication and authorization settings.
    
    The selection mirrors configuration: returns noop resolvers for NOOP/K8S/NOOP_WITH_TOKEN or when JWT role rules or authorization access rules are not set; returns JwtRolesResolver and GenericAccessResolver when JWK_TOKEN configuration provides role and access rules. The result is cached to avoid recomputing resolvers.
    
    Returns:
        tuple[RolesResolver, AccessResolver]: (roles_resolver, access_resolver) appropriate for the current configuration.
    """
    authorization_cfg = configuration.authorization_configuration
    authentication_config = configuration.authentication_configuration

    match authentication_config.module:
        case (
            constants.AUTH_MOD_NOOP
            | constants.AUTH_MOD_K8S
            | constants.AUTH_MOD_NOOP_WITH_TOKEN
        ):
            return (
                NoopRolesResolver(),
                NoopAccessResolver(),
            )
        case constants.AUTH_MOD_JWK_TOKEN:
            jwt_role_rules_unset = (
                len(
                    authentication_config.jwk_configuration.jwt_configuration.role_rules
                )
            ) == 0

            authz_access_rules_unset = len(authorization_cfg.access_rules) == 0

            if jwt_role_rules_unset or authz_access_rules_unset:
                return NoopRolesResolver(), NoopAccessResolver()

            return (
                JwtRolesResolver(
                    role_rules=(
                        authentication_config.jwk_configuration.jwt_configuration.role_rules
                    )
                ),
                GenericAccessResolver(authorization_cfg.access_rules),
            )

        case _:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail="Internal server error",
            )


async def _perform_authorization_check(
    action: Action, args: tuple[Any, ...], kwargs: dict[str, Any]
) -> None:
    """
    Perform the authorization check for an endpoint call and attach authorized actions to the request state.
    
    Performs role resolution and access verification for the supplied `action` using configured resolvers. Expects `kwargs` to contain an `auth` value from the authentication dependency; if a Request is present in `args` or `kwargs` its `state.authorized_actions` will be set to the set of actions the resolved roles are authorized to perform.
    
    Parameters:
        action (Action): The action to authorize.
        args (tuple[Any, ...]): Positional arguments passed to the endpoint; used to locate a Request instance if present.
        kwargs (dict[str, Any]): Keyword arguments passed to the endpoint; must include `auth` (authentication info) and may include `request`.
    
    Raises:
        HTTPException: with 500 Internal Server Error if `auth` is missing from `kwargs`.
        HTTPException: with 403 Forbidden if the resolved roles are not permitted to perform `action`.
    """
    role_resolver, access_resolver = get_authorization_resolvers()

    try:
        auth = kwargs["auth"]
    except KeyError as exc:
        logger.error(
            "Authorization only allowed on endpoints that accept "
            "'auth: Any = Depends(get_auth_dependency())'"
        )
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Internal server error",
        ) from exc

    # Everyone gets the everyone (aka *) role
    everyone_roles = {"*"}

    user_roles = await role_resolver.resolve_roles(auth) | everyone_roles

    if not access_resolver.check_access(action, user_roles):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail=f"Insufficient permissions for action: {action}",
        )

    authorized_actions = access_resolver.get_actions(user_roles)

    req: Request | None = None
    if "request" in kwargs and isinstance(kwargs["request"], Request):
        req = kwargs["request"]
    else:
        for arg in args:
            if isinstance(arg, Request):
                req = arg
                break
    if req is not None:
        req.state.authorized_actions = authorized_actions


def authorize(action: Action) -> Callable:
    """
    Create a decorator that enforces the specified authorization action on an endpoint.
    
    Parameters:
        action (Action): The action that the decorated endpoint must be authorized to perform.
    
    Returns:
        Callable: A decorator which, when applied to an endpoint function, performs the authorization check for the given action before invoking the function.
    """

    def decorator(func: Callable) -> Callable:
        """
        Wraps an endpoint function to perform an authorization check before invoking the original callable.
        
        Parameters:
            func (Callable): The function to wrap.
        
        Returns:
            Callable: A wrapper that performs authorization then calls `func`.
        """
        @wraps(func)
        async def wrapper(*args: Any, **kwargs: Any) -> Any:
            await _perform_authorization_check(action, args, kwargs)
            return await func(*args, **kwargs)

        return wrapper

    return decorator