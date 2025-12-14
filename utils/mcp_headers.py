"""MCP headers handling."""

import json
import logging
from urllib.parse import urlparse

from fastapi import Request

from configuration import AppConfig


logger = logging.getLogger("app.endpoints.dependencies")


async def mcp_headers_dependency(request: Request) -> dict[str, dict[str, str]]:
    """
    Provide MCP headers extracted from the request for use with MCP servers.
    
    Returns:
        dict[str, dict[str, str]]: Mapping of MCP server URL (or toolgroup-converted URL) to its header dictionary,
        or an empty dict if the `MCP-HEADERS` header is missing or invalid.
    """
    return extract_mcp_headers(request)


def extract_mcp_headers(request: Request) -> dict[str, dict[str, str]]:
    """
    Extract MCP headers from the "MCP-HEADERS" request header.
    
    Parses the header value as JSON and returns it if it is a mapping of strings to header dictionaries; if the header is missing, invalid JSON, or the parsed value is not a dictionary, returns an empty dictionary.
    
    Returns:
        dict[str, dict[str, str]]: Mapping of MCP server URL or toolgroup name to its headers, or an empty dict on error or when the header is absent.
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
    """Process MCP headers by converting toolgroup names to URLs.

    This function takes MCP headers where keys can be either valid URLs or
    toolgroup names. For valid URLs (HTTP/HTTPS), it keeps them as-is. For
    toolgroup names, it looks up the corresponding MCP server URL in the
    configuration and replaces the key with the URL. Unknown toolgroup names
    are filtered out.

    Args:
        mcp_headers: Dictionary with keys as URLs or toolgroup names
        config: Application configuration containing MCP server definitions

    Returns:
        Dictionary with URLs as keys and their corresponding headers as values
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