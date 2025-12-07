"""Configuration loader."""

import logging
from typing import Any, Optional

# We want to support environment variable replacement in the configuration
# similarly to how it is done in llama-stack, so we use their function directly
from llama_stack.core.stack import replace_env_vars

import yaml
from models.config import (
    AuthorizationConfiguration,
    Configuration,
    Customization,
    LlamaStackConfiguration,
    UserDataCollection,
    ServiceConfiguration,
    ModelContextProtocolServer,
    AuthenticationConfiguration,
    InferenceConfiguration,
    DatabaseConfiguration,
    ConversationCacheConfiguration,
    QuotaHandlersConfiguration,
)

from cache.cache import Cache
from cache.cache_factory import CacheFactory

from quota.quota_limiter import QuotaLimiter
from quota.quota_limiter_factory import QuotaLimiterFactory

logger = logging.getLogger(__name__)


class LogicError(Exception):
    """Error in application logic."""


class AppConfig:
    """Singleton class to load and store the configuration."""

    _instance = None

    def __new__(cls, *args: Any, **kwargs: Any) -> "AppConfig":
        """
        Ensure a single AppConfig instance exists and return it.
        
        Returns:
            The singleton AppConfig instance.
        """
        if not isinstance(cls._instance, cls):
            cls._instance = super().__new__(cls, *args, **kwargs)
        return cls._instance

    def __init__(self) -> None:
        """
        Set up the instance with no loaded configuration and cleared cached resources.
        
        Initializes:
        - _configuration: None until load_configuration or init_from_dict is called.
        - _conversation_cache: None until first access via the conversation_cache property.
        - _quota_limiters: empty list, populated lazily by the quota_limiters property.
        """
        self._configuration: Optional[Configuration] = None
        self._conversation_cache: Optional[Cache] = None
        self._quota_limiters: list[QuotaLimiter] = []

    def load_configuration(self, filename: str) -> None:
        """
        Load YAML configuration from the given file, substitute environment variables, and initialize the application configuration.
        
        Parameters:
            filename (str): Path to the YAML configuration file to load.
        """
        with open(filename, encoding="utf-8") as fin:
            config_dict = yaml.safe_load(fin)
            config_dict = replace_env_vars(config_dict)
            logger.info("Loaded configuration: %s", config_dict)
            self.init_from_dict(config_dict)

    def init_from_dict(self, config_dict: dict[Any, Any]) -> None:
        """
        Initialize the application's configuration from a dictionary.
        
        This replaces the current configuration with a new Configuration built from `config_dict`
        and clears any cached resources that depend on configuration (conversation cache and quota limiters).
        
        Parameters:
            config_dict (dict[Any, Any]): Dictionary containing configuration data compatible with the
                Configuration model's constructor.
        """
        # clear cached values when configuration changes
        self._conversation_cache = None
        self._quota_limiters = []
        # now it is possible to re-read configuration
        self._configuration = Configuration(**config_dict)

    @property
    def configuration(self) -> Configuration:
        """
        Access the loaded application configuration.
        
        Returns:
            configuration (Configuration): The parsed application configuration.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration

    @property
    def service_configuration(self) -> ServiceConfiguration:
        """
        Access the service-specific configuration section.
        
        Returns:
            ServiceConfiguration: the loaded service configuration.
        
        Raises:
            LogicError: if the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.service

    @property
    def llama_stack_configuration(self) -> LlamaStackConfiguration:
        """
        Get the Llama stack configuration from the loaded application configuration.
        
        Returns:
            LlamaStackConfiguration: The configured Llama stack settings.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.llama_stack

    @property
    def user_data_collection_configuration(self) -> UserDataCollection:
        """
        Get the user data collection configuration.
        
        Returns:
            UserDataCollection: The configured user data collection settings.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.user_data_collection

    @property
    def mcp_servers(self) -> list[ModelContextProtocolServer]:
        """
        Model Context Protocol (MCP) servers configuration.
        
        Returns:
            list[ModelContextProtocolServer]: The configured MCP servers.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.mcp_servers

    @property
    def authentication_configuration(self) -> AuthenticationConfiguration:
        """
        Access the authentication configuration section.
        
        Returns:
            AuthenticationConfiguration: The authentication configuration.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")

        return self._configuration.authentication

    @property
    def authorization_configuration(self) -> AuthorizationConfiguration:
        """
        Return the authorization configuration, falling back to a default no-op configuration if none is present.
        
        Returns:
            AuthorizationConfiguration: The configured authorization settings, or a default no-op `AuthorizationConfiguration` when the configuration does not include an authorization section.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")

        if self._configuration.authorization is None:
            return AuthorizationConfiguration()

        return self._configuration.authorization

    @property
    def customization(self) -> Optional[Customization]:
        """
        Access the optional customization configuration for the application.
        
        Returns:
            Optional[Customization]: The customization configuration if present, otherwise `None`.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.customization

    @property
    def inference(self) -> InferenceConfiguration:
        """
        Retrieve the inference configuration section from the loaded configuration.
        
        Returns:
            InferenceConfiguration: The inference configuration.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.inference

    @property
    def conversation_cache_configuration(self) -> ConversationCacheConfiguration:
        """
        Retrieve the conversation cache configuration section from the loaded application configuration.
        
        Returns:
            ConversationCacheConfiguration: The conversation cache configuration.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.conversation_cache

    @property
    def database_configuration(self) -> DatabaseConfiguration:
        """
        Get the application's database configuration.
        
        Returns:
            DatabaseConfiguration: The loaded database configuration section.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.database

    @property
    def quota_handlers_configuration(self) -> QuotaHandlersConfiguration:
        """
        Return the quota handlers configuration section.
        
        Returns:
            quota_handlers (QuotaHandlersConfiguration): The quota handlers configuration from the loaded configuration.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        return self._configuration.quota_handlers

    @property
    def conversation_cache(self) -> Cache:
        """
        Get the conversation cache, creating it from the loaded configuration if it has not been created yet.
        
        Returns:
            cache (Cache): The conversation cache instance.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        if self._conversation_cache is None:
            self._conversation_cache = CacheFactory.conversation_cache(
                self._configuration.conversation_cache
            )
        return self._conversation_cache

    @property
    def quota_limiters(self) -> list[QuotaLimiter]:
        """
        Provide the application's configured quota limiters.
        
        Builds and caches quota limiters from the loaded configuration if they have not been created yet.
        
        Returns:
            list[QuotaLimiter]: The list of configured quota limiters.
        
        Raises:
            LogicError: If the configuration has not been loaded.
        """
        if self._configuration is None:
            raise LogicError("logic error: configuration is not loaded")
        if not self._quota_limiters:
            self._quota_limiters = QuotaLimiterFactory.quota_limiters(
                self._configuration.quota_handlers
            )
        return self._quota_limiters


configuration: AppConfig = AppConfig()