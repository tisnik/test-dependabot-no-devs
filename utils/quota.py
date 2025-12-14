"""Quota handling helper functions."""

import psycopg2
from fastapi import HTTPException

from log import get_logger
from models.responses import InternalServerErrorResponse, QuotaExceededResponse
from quota.quota_exceed_error import QuotaExceedError
from quota.quota_limiter import QuotaLimiter

logger = get_logger(__name__)


def consume_tokens(
    quota_limiters: list[QuotaLimiter],
    user_id: str,
    input_tokens: int,
    output_tokens: int,
) -> None:
    """
    Consume tokens from each provided quota limiter for the given user.
    
    Parameters:
        quota_limiters (list[QuotaLimiter]): QuotaLimiter instances to charge.
        user_id (str): Subject identifier whose quotas will be decreased.
        input_tokens (int): Number of input tokens to consume.
        output_tokens (int): Number of output tokens to consume.
    """
    # consume tokens all configured quota limiters
    for quota_limiter in quota_limiters:
        quota_limiter.consume_tokens(
            input_tokens=input_tokens,
            output_tokens=output_tokens,
            subject_id=user_id,
        )


def check_tokens_available(quota_limiters: list[QuotaLimiter], user_id: str) -> None:
    """
    Ensure every configured quota limiter reports available tokens for the given user.
    
    Raises:
        HTTPException: with status 500 if a database error occurs communicating with the quota backend, or with status 429 if the user's quota is exceeded.
    """
    try:
        # check available tokens using all configured quota limiters
        for quota_limiter in quota_limiters:
            quota_limiter.ensure_available_quota(subject_id=user_id)
    except psycopg2.Error as pg_error:
        message = "Error communicating with quota database backend"
        logger.error(message)
        response = InternalServerErrorResponse.database_error()
        raise HTTPException(**response.model_dump()) from pg_error
    except QuotaExceedError as e:
        logger.error("The quota has been exceeded")
        response = QuotaExceededResponse.from_exception(e)
        raise HTTPException(**response.model_dump()) from e


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