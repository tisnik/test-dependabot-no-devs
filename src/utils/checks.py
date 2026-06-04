"""Checks that are performed to configuration options."""

import os
from typing import Optional

from pydantic import FilePath


class InvalidConfigurationError(Exception):
    """Lightspeed configuration is invalid."""


def get_attribute_from_file(data: dict, file_name_key: str) -> Optional[str]:
    """
    Reads and returns the content of a file specified by a key in the given dictionary.
    
    If the key is present and its value is a valid file path, the file is opened with UTF-8 encoding, its content is read, and trailing whitespace is stripped. Returns `None` if the key is missing or its value is `None`.
    
    Parameters:
        data (dict): Dictionary containing file path values.
        file_name_key (str): Key whose value is the file path to read.
    
    Returns:
        Optional[str]: The stripped content of the file, or `None` if the key is not present or the value is `None`.
    """
    file_path = data.get(file_name_key)
    if file_path is not None:
        with open(file_path, encoding="utf-8") as f:
            return f.read().rstrip()
    return None


def file_check(path: FilePath, desc: str) -> None:
    """
    Verify that the given path points to a readable regular file.
    
    Raises:
        InvalidConfigurationError: If the path does not refer to a file or is not readable.
    """
    if not os.path.isfile(path):
        raise InvalidConfigurationError(f"{desc} '{path}' is not a file")
    if not os.access(path, os.R_OK):
        raise InvalidConfigurationError(f"{desc} '{path}' is not readable")
