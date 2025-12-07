"""MCP headers handling."""

import json
import logging
from urllib.parse import urlparse

from fastapi import Request

from configuration import AppConfig


logger = logging.getLogger("app.endpoints.dependencies")


async def mcp_headers_dependency(request: Request) -> dict[str, dict[str, str]]:
    """
    Provide MCP headers extracted from the incoming request for use with MCP servers.
    
    Returns:
        dict[str, dict[str, str]]: Parsed MCP headers from the "MCP-HEADERS" request header, or an empty dict if the header is missing, not valid JSON, or not a JSON object.
    """
    return extract_mcp_headers(request)


def extract_mcp_headers(request: Request) -> dict[str, dict[str, str]]:
    """
    Parse the "MCP-HEADERS" request header and return it as a dictionary mapping server URLs or toolgroup names to header dictionaries.
    
    If the header is missing, contains invalid JSON, or the decoded value is not a dictionary, an empty dictionary is returned.
    
    Returns:
        dict[str, dict[str, str]]: Parsed MCP headers mapping, or an empty dict on missing/invalid header.
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
    Normalize MCP headers by resolving toolgroup names to MCP server URLs.
    
    Parameters:
        mcp_headers (dict[str, dict[str, str]]): Mapping where keys are either full MCP server URLs or toolgroup names and values are header dictionaries to send.
        config (AppConfig): Application configuration containing `mcp_servers`, each with `name` and `url` fields.
    
    Returns:
        dict[str, dict[str, str]]: A mapping with MCP server URLs as keys and their corresponding header dictionaries as values. Entries for unknown toolgroup names are omitted.
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