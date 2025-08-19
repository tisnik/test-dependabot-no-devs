"""MCP headers handling."""

import json
import logging
from urllib.parse import urlparse

from fastapi import Request

from configuration import AppConfig


logger = logging.getLogger("app.endpoints.dependencies")


async def mcp_headers_dependency(request: Request) -> dict[str, dict[str, str]]:
    """
    Provide MCP headers for MCP servers as a FastAPI dependency.
    
    Delegates to extract_mcp_headers(request) and returns a mapping from MCP server URL to its headers.
    Returns:
        dict[str, dict[str, str]]: MCP headers mapping (empty if the header is missing or invalid)
    """
    return extract_mcp_headers(request)


def extract_mcp_headers(request: Request) -> dict[str, dict[str, str]]:
    """
    Extract MCP headers from the request's "MCP-HEADERS" HTTP header.
    
    Reads the "MCP-HEADERS" header, attempts to JSON-decode it, and returns the resulting mapping. If the header is missing, contains invalid JSON, or does not decode to a dictionary, an empty dict is returned and an error is logged.
    
    Args:
        request: FastAPI Request; the function reads the "MCP-HEADERS" header from request.headers.
    
    Returns:
        dict[str, dict[str, str]]: A mapping of MCP server URL (or toolgroup key) to its headers, or an empty dict on missing/invalid input.
    """
    mcp_headers_string = request.headers.get("MCP-HEADERS", "")
    mcp_headers = {}
    if mcp_headers_string:
        try:
            mcp_headers = json.loads(mcp_headers_string)
        except json.decoder.JSONDecodeError as e:
            logger.error("MCP headers decode error: %s", e)

        if not isinstance(mcp_headers, dict):
            logger.error(
                "MCP headers wrong type supplied (mcp headers must be a dictionary), "
                "but type %s was supplied",
                type(mcp_headers),
            )
            mcp_headers = {}
    return mcp_headers


def handle_mcp_headers_with_toolgroups(
    mcp_headers: dict[str, dict[str, str]], config: AppConfig
) -> dict[str, dict[str, str]]:
    """
    Convert MCP header keys that are toolgroup names into their corresponding MCP server URLs.
    
    Takes mcp_headers where each key is either a full HTTP/HTTPS URL or a toolgroup name. Keys that are valid HTTP/HTTPS URLs are preserved. Keys that are toolgroup names are resolved against config.mcp_servers; if a matching server with a non-empty URL is found, the entry is rekeyed to that server URL. Entries with unknown toolgroup names are omitted.
    
    Parameters:
        mcp_headers (dict[str, dict[str, str]]): Mapping from URL or toolgroup name to a headers dict.
    
    Returns:
        dict[str, dict[str, str]]: Mapping from MCP server URL to its headers.
    """
    converted_mcp_headers = {}

    for key, item in mcp_headers.items():
        key_url_parsed = urlparse(key)
        if key_url_parsed.scheme in ("http", "https") and key_url_parsed.netloc:
            # a valid url is supplied, deliver it as is
            converted_mcp_headers[key] = item
        else:
            # assume the key is a toolgroup name
            # look for toolgroups name in mcp_servers configuration
            # if the mcp server is not found, the mcp header gets ignored
            for mcp_server in config.mcp_servers:
                if mcp_server.name == key and mcp_server.url:
                    converted_mcp_headers[mcp_server.url] = item
                    break

    return converted_mcp_headers
