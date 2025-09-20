"""Cache factory class."""

import constants
from models.config import ConversationCache
from cache.cache import Cache
from cache.sqlite_cache import SQLiteCache
from log import get_logger

logger = get_logger("cache.cache_factory")


class CacheFactory:
    """Cache factory class."""

    @staticmethod
    def conversation_cache(config: ConversationCache) -> Cache:
        """Create an instance of Cache based on loaded configuration.

        Returns:
            An instance of `Cache` (either `PostgresCache` or `InMemoryCache`).
        """
        logger.info("Creating cache instance of type %s", config.type)
        match config.type:
            case constants.CACHE_TYPE_MEMORY:
                return None
            case constants.CACHE_TYPE_SQLITE:
                return SQLiteCache(config.sqlite)
            case constants.CACHE_TYPE_POSTGRES:
                return None
            case _:
                raise ValueError(
                    f"Invalid cache type: {config.type}. "
                    f"Use '{constants.CACHE_TYPE_POSTGRES}' or "
                    f"'{constants.CACHE_TYPE_MEMORY}' options."
                )
