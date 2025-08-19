"""Checks that are performed to configuration options."""

import os
from typing import Optional

from pydantic import FilePath


class InvalidConfigurationError(Exception):
    """Lightspeed configuration is invalid."""


def get_attribute_from_file(data: dict, file_name_key: str) -> Optional[str]:
    """
    Return the contents of a file whose path is stored in the given mapping.
    
    Looks up file_name_key in data; if a non-None value is found it is treated as a filesystem path, the file is opened with UTF-8 encoding, its full contents are returned with trailing whitespace removed. If the key is missing or maps to None, returns None.
    
    Parameters:
        data (dict): Mapping containing the file path under file_name_key.
        file_name_key (str): Key in `data` whose value is the path to the file.
    
    Returns:
        Optional[str]: File contents with trailing whitespace stripped, or None if the key is not present or is None.
    """
    file_path = data.get(file_name_key)
    if file_path is not None:
        with open(file_path, encoding="utf-8") as f:
            return f.read().rstrip()
    return None


def file_check(path: FilePath, desc: str) -> None:
    """
    Ensure the given path is an existing regular file and is readable.
    
    If the path is not a regular file or is not readable, raises InvalidConfigurationError
    with a message that includes `desc` and the offending path.
    
    Parameters:
        path (FilePath): Filesystem path to validate.
        desc (str): Short description of the value being checked; used in error messages.
    
    Raises:
        InvalidConfigurationError: If `path` does not point to a file or is not readable.
    """
    if not os.path.isfile(path):
        raise InvalidConfigurationError(f"{desc} '{path}' is not a file")
    if not os.access(path, os.R_OK):
        raise InvalidConfigurationError(f"{desc} '{path}' is not readable")
