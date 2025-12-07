"""In-memory cache implementation."""

from cache.cache import Cache
from models.cache_entry import CacheEntry
from models.config import InMemoryCacheConfig
from models.responses import ConversationData
from log import get_logger
from utils.connection_decorator import connection

logger = get_logger("cache.in_memory_cache")


class InMemoryCache(Cache):
    """In-memory cache implementation."""

    def __init__(self, config: InMemoryCacheConfig) -> None:
        """
        Initialize the InMemoryCache with the provided configuration.
        
        Parameters:
            config (InMemoryCacheConfig): Configuration options controlling cache behavior.
        """
        self.cache_config = config

    def connect(self) -> None:
        """
        Log the start of a storage connection for the in-memory cache; does not establish an external connection.
        
        This method records (via logger) that the cache is connecting to its storage backend. For the in-memory implementation this is a no-op with respect to network or persistent connections.
        """
        logger.info("Connecting to storage")

    def connected(self) -> bool:
        """
        Report whether the cache connection is currently available.
        
        Returns:
            True if the cache is available, False otherwise.
        """
        return True

    def initialize_cache(self) -> None:
        """
        No-op placeholder for cache initialization.
        
        This implementation performs no actions and exists only to satisfy the cache interface.
        """

    @connection
    def get(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> list[CacheEntry]:
        """
        Validate the provided identifiers and retrieve cache entries for a user's conversation.
        
        Parameters:
            skip_user_id_check (bool): If True, skip validation of `user_id`. Validation for `user_id` and `conversation_id` requires them to be well-formed UUIDs.
        
        Returns:
            list[CacheEntry]: An empty list.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)
        return []

    @connection
    def insert_or_append(
        self,
        user_id: str,
        conversation_id: str,
        cache_entry: CacheEntry,
        skip_user_id_check: bool = False,
    ) -> None:
        """
        Validate and construct the cache key for a user's conversation without storing data.
        
        This method verifies the provided `user_id` and `conversation_id` (via the base class key construction/validation) and performs no persistent storage or mutation.
        
        Parameters:
            user_id (str): Identifier for the user.
            conversation_id (str): Identifier for the conversation within the user scope.
            cache_entry (CacheEntry): The cache entry that would be stored (not persisted by this implementation).
            skip_user_id_check (bool): If true, skip additional user-id validation.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)

    @connection
    def delete(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> bool:
        """
        Validate the provided user and conversation identifiers and report deletion success.
        
        Parameters:
            user_id (str): User identifier to validate.
            conversation_id (str): Conversation identifier to validate.
            skip_user_id_check (bool): If True, skip validation of the `user_id` format.
        
        Returns:
            bool: `True` (operation is considered successful in all cases).
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)
        return True

    @connection
    def list(
        self, user_id: str, skip_user_id_check: bool = False
    ) -> list[ConversationData]:
        """
        Return the list of conversations for the specified user.
        
        Parameters:
            user_id (str): The user's identifier to list conversations for.
            skip_user_id_check (bool): If True, skip validation of `user_id`.
        
        Returns:
            list[ConversationData]: A list of conversation entries for the user; may be empty.
        """
        super()._check_user_id(user_id, skip_user_id_check)
        return []

    @connection
    def set_topic_summary(
        self,
        user_id: str,
        conversation_id: str,
        topic_summary: str,
        skip_user_id_check: bool = False,
    ) -> None:
        """
        Store the topic summary for a specific user's conversation.
        
        Parameters:
            user_id (str): The user's identifier (expected to be a UUID unless `skip_user_id_check` is True).
            conversation_id (str): The conversation identifier (expected to be a UUID).
            topic_summary (str): The summary text to associate with the conversation's topic.
            skip_user_id_check (bool): If True, skip validation of `user_id`.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)

    def ready(self) -> bool:
        """
        Report whether the cache is ready.
        
        Returns:
            True (`bool`): Always `True` for this in-memory cache implementation.
        """
        return True