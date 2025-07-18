"""Configuration loader."""

import logging
from typing import Any, Optional

import yaml
from models.config import (
    Configuration,
    Customization,
    LLamaStackConfiguration,
    UserDataCollection,
    ServiceConfiguration,
    ModelContextProtocolServer,
    AuthenticationConfiguration,
)

logger = logging.getLogger(__name__)


class AppConfig:
    """Singleton class to load and store the configuration."""

    _instance = None

    def __new__(cls, *args: Any, **kwargs: Any) -> "AppConfig":
        """
        Implements the singleton pattern by ensuring only one instance of AppConfig is created.
        
        Returns:
            AppConfig: The singleton instance of the AppConfig class.
        """
        if not isinstance(cls._instance, cls):
            cls._instance = super().__new__(cls, *args, **kwargs)
        return cls._instance

    def __init__(self) -> None:
        """
        Initializes the AppConfig instance with no loaded configuration.
        """
        self._configuration: Optional[Configuration] = None

    def load_configuration(self, filename: str) -> None:
        """
        Loads application configuration from a YAML file.
        
        Parameters:
            filename (str): Path to the YAML configuration file.
        """
        with open(filename, encoding="utf-8") as fin:
            config_dict = yaml.safe_load(fin)
            logger.info("Loaded configuration: %s", config_dict)
            self.init_from_dict(config_dict)

    def init_from_dict(self, config_dict: dict[Any, Any]) -> None:
        """
        Initialize the internal configuration using the provided dictionary.
        
        Parameters:
            config_dict (dict): A dictionary containing configuration data to instantiate the Configuration object.
        """
        self._configuration = Configuration(**config_dict)

    @property
    def configuration(self) -> Configuration:
        """
        Returns the loaded application configuration.
        
        Returns:
            Configuration: The complete configuration object.
        
        Raises:
            AssertionError: If the configuration has not been loaded.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration

    @property
    def service_configuration(self) -> ServiceConfiguration:
        """
        Returns the service configuration section from the loaded application configuration.
        
        Raises an AssertionError if the configuration has not been loaded.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.service

    @property
    def llama_stack_configuration(self) -> LLamaStackConfiguration:
        """
        Returns the Llama stack configuration from the loaded application configuration.
        
        Returns:
            LLamaStackConfiguration: The configuration settings for the Llama stack.
        
        Raises:
            AssertionError: If the configuration has not been loaded.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.llama_stack

    @property
    def user_data_collection_configuration(self) -> UserDataCollection:
        """
        Returns the user data collection configuration from the loaded application configuration.
        
        Returns:
            UserDataCollection: The user data collection configuration section.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.user_data_collection

    @property
    def mcp_servers(self) -> list[ModelContextProtocolServer]:
        """
        Returns the list of Model Context Protocol servers from the loaded configuration.
        
        Returns:
            List of ModelContextProtocolServer objects representing the configured MCP servers.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.mcp_servers

    @property
    def authentication_configuration(self) -> Optional[AuthenticationConfiguration]:
        """
        Returns the authentication configuration if available.
        
        Returns:
            Optional[AuthenticationConfiguration]: The authentication configuration, or None if not specified in the loaded configuration.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.authentication

    @property
    def customization(self) -> Optional[Customization]:
        """
        Returns the customization configuration if available.
        
        Returns:
            Optional[Customization]: The customization settings from the loaded configuration, or None if not specified.
        """
        assert (
            self._configuration is not None
        ), "logic error: configuration is not loaded"
        return self._configuration.customization


configuration: AppConfig = AppConfig()
