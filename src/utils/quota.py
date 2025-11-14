"""Quota handling helper functions."""

import psycopg2

from fastapi import HTTPException, status

from quota.quota_limiter import QuotaLimiter
from quota.quota_exceed_error import QuotaExceedError

from log import get_logger

logger = get_logger(__name__)


def consume_tokens(
    quota_limiters: list[QuotaLimiter],
    user_id: str,
    input_tokens: int,
    output_tokens: int,
) -> None:
    """Consume tokens from cluster and/or user quotas.

    Args:
        quota_limiters: List of quota limiter instances to consume tokens from.
        user_id: Identifier of the user consuming tokens.
        input_tokens: Number of input tokens to consume.
        output_tokens: Number of output tokens to consume.

    Returns:
        None
    """
    # consume tokens all configured quota limiters
    for quota_limiter in quota_limiters:
        quota_limiter.consume_tokens(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            subject_id=user_id,
        )


def check_tokens_available(quota_limiters: list[QuotaLimiter], user_id: str) -> None:
    """Check if tokens are available for user.

    Args:
        quota_limiters: List of quota limiter instances to check.
        user_id: Identifier of the user to check quota for.

    Returns:
        None

    Raises:
        HTTPException: With status 500 if database communication fails,
            or status 429 if quota is exceeded.
    """
    try:
        # check available tokens using all configured quota limiters
        for quota_limiter in quota_limiters:
            quota_limiter.ensure_available_quota(subject_id=user_id)
    except psycopg2.Error as pg_error:
        message = "Error communicating with quota database backend"
        logger.error(message)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail={
                "response": message,
                "cause": str(pg_error),
            },
        ) from pg_error
    except QuotaExceedError as quota_exceed_error:
        message = "The quota has been exceeded"
        logger.error(message)
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail={
                "response": message,
                "cause": str(quota_exceed_error),
            },
        ) from quota_exceed_error


def get_available_quotas(
    quota_limiters: list[QuotaLimiter],
    user_id: str,
) -> dict[str, int]:
    """Get quota available from all quota limiters.

    Args:
        quota_limiters: List of quota limiter instances to query.
        user_id: Identifier of the user to get quotas for.

    Returns:
        Dictionary mapping quota limiter class names to available token counts.
    """
    available_quotas: dict[str, int] = {}

    # retrieve available tokens using all configured quota limiters
    for quota_limiter in quota_limiters:
        name = quota_limiter.__class__.__name__
        available_quota = quota_limiter.available_quota(user_id)
        available_quotas[name] = available_quota
    return available_quotas
