"""Checks that are performed to configuration options."""

import os
import importlib
import importlib.util
from types import ModuleType
from typing import Optional
from pydantic import FilePath


class InvalidConfigurationError(Exception):
    """Lightspeed configuration is invalid."""


def get_attribute_from_file(data: dict, file_name_key: str) -> Optional[str]:
    """
    Retrieve file contents from a path stored in a mapping.
    
    Looks up file_name_key in data; if the value is not None, treats it as a filesystem path,
    reads the file using UTF-8 encoding, and returns its contents with trailing whitespace removed.
    If the key is missing or maps to None, returns None.
    
    Parameters:
        data (dict): Mapping containing the file path under file_name_key.
        file_name_key (str): Key in `data` whose value is the path to the file.
    
    Returns:
        Optional[str]: File contents with trailing whitespace removed, or None if the key is missing or maps to None.
    """
    file_path = data.get(file_name_key)
    if file_path is not None:
        with open(file_path, encoding="utf-8") as f:
            return f.read().rstrip()
    return None


def file_check(path: FilePath, desc: str) -> None:
    """
    Ensure the given path is an existing regular file and is readable.

    If the path is not a regular file or is not readable, raises
    InvalidConfigurationError.

    Parameters:
        path (FilePath): Filesystem path to validate.
        desc (str): Short description of the value being checked; used in error
        messages.

    Raises:
        InvalidConfigurationError: If `path` does not point to a file or is not
        readable.
    """
    if not os.path.isfile(path):
        raise InvalidConfigurationError(f"{desc} '{path}' is not a file")
    if not os.access(path, os.R_OK):
        raise InvalidConfigurationError(f"{desc} '{path}' is not readable")


def directory_check(
    path: FilePath, must_exists: bool, must_be_writable: bool, desc: str
) -> None:
    """
    Validate a directory path according to existence and writability requirements.
    
    Parameters:
        path (FilePath): Filesystem path to validate.
        must_exists (bool): If True, require that the path exists.
        must_be_writable (bool): If True, require that the existing path is writable.
        desc (str): Short description used in error messages to identify the value being checked.
    
    Raises:
        InvalidConfigurationError: If the path must exist but does not, if the path exists but is not a directory, or if writability is required but the path is not writable.
    """
    if not os.path.exists(path):
        if must_exists:
            raise InvalidConfigurationError(f"{desc} '{path}' does not exist")
        return
    if not os.path.isdir(path):
        raise InvalidConfigurationError(f"{desc} '{path}' is not a directory")
    if must_be_writable:
        if not os.access(path, os.W_OK):
            raise InvalidConfigurationError(f"{desc} '{path}' is not writable")


def import_python_module(profile_name: str, profile_path: str) -> ModuleType | None:
    """
    Import a Python module from a filesystem path and return the loaded module.

    Parameters:
        profile_name (str): Name to assign to the imported module.
        profile_path (str): Filesystem path to the Python source file; must end with `.py`.

    Returns:
        ModuleType | None: The loaded module on success; `None` if
        `profile_path` does not end with `.py`, if a module spec or loader
        cannot be created, or if importing/executing the module fails.
    """
    if not profile_path.endswith(".py"):
        return None
    spec = importlib.util.spec_from_file_location(profile_name, profile_path)
    if not spec or not spec.loader:
        return None
    module = importlib.util.module_from_spec(spec)
    try:
        spec.loader.exec_module(module)
    except (
        SyntaxError,
        ImportError,
        ModuleNotFoundError,
        NameError,
        AttributeError,
        TypeError,
        ValueError,
    ):
        return None
    return module


def is_valid_profile(profile_module: ModuleType) -> bool:
    """
    Determine whether a module provides a valid PROFILE_CONFIG containing system_prompts.
    
    A valid PROFILE_CONFIG is a dict that includes a 'system_prompts' key whose value is a dict and is not empty.
    
    Returns:
        True if PROFILE_CONFIG exists as a dict and contains a non-empty 'system_prompts' dict, False otherwise.
    """
    if not hasattr(profile_module, "PROFILE_CONFIG"):
        return False

    profile_config = getattr(profile_module, "PROFILE_CONFIG", {})
    if not isinstance(profile_config, dict):
        return False

    if not profile_config.get("system_prompts"):
        return False

    return isinstance(profile_config.get("system_prompts"), dict)