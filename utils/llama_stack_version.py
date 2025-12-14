"""Check if the Llama Stack version is supported by the LCS."""

import logging
import re

from semver import Version

from llama_stack_client._client import AsyncLlamaStackClient


from constants import (
    MINIMAL_SUPPORTED_LLAMA_STACK_VERSION,
    MAXIMAL_SUPPORTED_LLAMA_STACK_VERSION,
)

logger = logging.getLogger("utils.llama_stack_version")


class InvalidLlamaStackVersionException(Exception):
    """Llama Stack version is not valid."""


async def check_llama_stack_version(
    client: AsyncLlamaStackClient,
) -> None:
    """
    Verify the connected Llama Stack version is within the configured supported range.
    
    Fetches the version from the provided client and validates it against the module-level minimal and maximal supported versions.
    
    Raises:
        InvalidLlamaStackVersionException: If the detected version is outside the supported range or cannot be parsed.
    """
    version_info = await client.inspect.version()
    compare_versions(
        version_info.version,
        MINIMAL_SUPPORTED_LLAMA_STACK_VERSION,
        MAXIMAL_SUPPORTED_LLAMA_STACK_VERSION,
    )


def compare_versions(version_info: str, minimal: str, maximal: str) -> None:
    """
    Check that a Llama Stack semantic version found in `version_info` is within the inclusive range [minimal, maximal].
    
    Extracts a `MAJOR.MINOR.PATCH` pattern from `version_info`, parses it as a semantic version, and verifies it is not less than `minimal` and not greater than `maximal`.
    
    Parameters:
        version_info (str): Text containing a semantic version (may include surrounding text); the first `X.Y.Z` pattern will be used.
        minimal (str): Minimum allowed semantic version (inclusive).
        maximal (str): Maximum allowed semantic version (inclusive).
    
    Raises:
        InvalidLlamaStackVersionException: If no version pattern can be extracted, the extracted version cannot be parsed, or the parsed version is outside the inclusive [minimal, maximal] range.
    """
    version_pattern = r"\d+\.\d+\.\d+"
    match = re.search(version_pattern, version_info)
    if not match:
        logger.warning(
            "Failed to extract version pattern from '%s'. Skipping version check.",
            version_info,
        )
        raise InvalidLlamaStackVersionException(
            f"Failed to extract version pattern from '{version_info}'. Skipping version check."
        )

    normalized_version = match.group(0)

    try:
        current_version = Version.parse(normalized_version)
    except ValueError as e:
        logger.warning("Failed to parse Llama Stack version '%s'.", version_info)
        raise InvalidLlamaStackVersionException(
            f"Failed to parse Llama Stack version '{version_info}'."
        ) from e

    minimal_version = Version.parse(minimal)
    maximal_version = Version.parse(maximal)
    logger.debug("Current version: %s", current_version)
    logger.debug("Minimal version: %s", minimal_version)
    logger.debug("Maximal version: %s", maximal_version)

    if current_version < minimal_version:
        raise InvalidLlamaStackVersionException(
            f"Llama Stack version >= {minimal_version} is required, but {current_version} is used"
        )
    if current_version > maximal_version:
        raise InvalidLlamaStackVersionException(
            f"Llama Stack version <= {maximal_version} is required, but {current_version} is used"
        )
    logger.info("Correct Llama Stack version : %s", current_version)