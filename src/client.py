"""LLama stack client retrieval."""

import logging

from typing import Optional

from llama_stack.distribution.library_client import (
    AsyncLlamaStackAsLibraryClient,  # type: ignore
    LlamaStackAsLibraryClient,  # type: ignore
)
from llama_stack_client import AsyncLlamaStackClient, LlamaStackClient  # type: ignore
from models.config import LLamaStackConfiguration
from utils.types import Singleton


logger = logging.getLogger(__name__)


class LlamaStackClientHolder(metaclass=Singleton):
    """Container for an initialised LlamaStackClient."""

    _lsc: Optional[LlamaStackClient] = None

    def load(self, llama_stack_config: LLamaStackConfiguration) -> None:
        """
        Initializes and stores a Llama stack client instance based on the provided configuration.
        
        Depending on the configuration, this method initializes either a library-based or service-based Llama stack client. If `use_as_library_client` is set and a valid `library_client_config_path` is provided, it creates and initializes a library client. If the required config path is missing, a `ValueError` is raised. Otherwise, it creates a client configured to connect to a running Llama stack service.
        
        Raises:
            ValueError: If `use_as_library_client` is True but `library_client_config_path` is not set.
        """
        if llama_stack_config.use_as_library_client is True:
            if llama_stack_config.library_client_config_path is not None:
                logger.info("Using Llama stack as library client")
                client = LlamaStackAsLibraryClient(
                    llama_stack_config.library_client_config_path
                )
                client.initialize()
                self._lsc = client
            else:
                msg = "Configuration problem: library_client_config_path option is not set"
                logger.error(msg)
                # tisnik: use custom exception there - with cause etc.
                raise ValueError(msg)

        else:
            logger.info("Using Llama stack running as a service")
            self._lsc = LlamaStackClient(
                base_url=llama_stack_config.url, api_key=llama_stack_config.api_key
            )

    def get_client(self) -> LlamaStackClient:
        """
        Return the initialized LlamaStackClient instance.
        
        Returns:
            LlamaStackClient: The initialized client instance.
        
        Raises:
            RuntimeError: If the client has not been initialized via the `load` method.
        """
        if not self._lsc:
            raise RuntimeError(
                "LlamaStackClient has not been initialised. Ensure 'load(..)' has been called."
            )
        return self._lsc


class AsyncLlamaStackClientHolder(metaclass=Singleton):
    """Container for an initialised AsyncLlamaStackClient."""

    _lsc: Optional[AsyncLlamaStackClient] = None

    async def load(self, llama_stack_config: LLamaStackConfiguration) -> None:
        """
        Initializes the asynchronous Llama stack client based on the provided configuration.
        
        Depending on the configuration, this method initializes either an asynchronous library client (with explicit initialization) or a service client. Raises a ValueError if the configuration requires a library client but the config path is missing.
        """
        if llama_stack_config.use_as_library_client is True:
            if llama_stack_config.library_client_config_path is not None:
                logger.info("Using Llama stack as library client")
                client = AsyncLlamaStackAsLibraryClient(
                    llama_stack_config.library_client_config_path
                )
                await client.initialize()
                self._lsc = client
            else:
                msg = "Configuration problem: library_client_config_path option is not set"
                logger.error(msg)
                # tisnik: use custom exception there - with cause etc.
                raise ValueError(msg)
        else:
            logger.info("Using Llama stack running as a service")
            self._lsc = AsyncLlamaStackClient(
                base_url=llama_stack_config.url, api_key=llama_stack_config.api_key
            )

    def get_client(self) -> AsyncLlamaStackClient:
        """
        Return the initialized asynchronous Llama stack client instance.
        
        Returns:
            AsyncLlamaStackClient: The initialized asynchronous client.
        
        Raises:
            RuntimeError: If the client has not been initialized via the `load` method.
        """
        if not self._lsc:
            raise RuntimeError(
                "AsyncLlamaStackClient has not been initialised. Ensure 'load(..)' has been called."
            )
        return self._lsc
