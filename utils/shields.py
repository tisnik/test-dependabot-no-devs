"""Utility functions for working with Llama Stack shields."""

import logging
from typing import Any

from llama_stack_client import AsyncLlamaStackClient

import metrics

logger = logging.getLogger(__name__)


async def get_available_shields(client: AsyncLlamaStackClient) -> list[str]:
    """
    Return the identifiers of shields available from the given Llama Stack client.
    
    Parameters:
        client (AsyncLlamaStackClient): Llama Stack client used to query available shields.
    
    Returns:
        list[str]: List of available shield identifiers; empty if no shields are available.
    """
    available_shields = [shield.identifier for shield in await client.shields.list()]
    if not available_shields:
        logger.info("No available shields. Disabling safety")
    else:
        logger.info("Available shields: %s", available_shields)
    return available_shields


def detect_shield_violations(output_items: list[Any]) -> bool:
    """
    Detect whether any item in `output_items` contains a shield refusal.
    
    Scans each item (expected to be objects or dict-like with a `type` attribute) and, for items where `type == "message"`, checks for a non-empty `refusal` attribute. If a refusal is found, increments the `metrics.llm_calls_validation_errors_total` counter and logs a warning.
    
    Parameters:
        output_items (list[Any]): Sequence of LLM output items to inspect; items are expected to expose `type` and optionally `refusal`.
    
    Returns:
        bool: `true` if a shield (refusal) violation was detected, `false` otherwise.
    """
    for output_item in output_items:
        item_type = getattr(output_item, "type", None)
        if item_type == "message":
            refusal = getattr(output_item, "refusal", None)
            if refusal:
                # Metric for LLM validation errors (shield violations)
                metrics.llm_calls_validation_errors_total.inc()
                logger.warning("Shield violation detected: %s", refusal)
                return True
    return False