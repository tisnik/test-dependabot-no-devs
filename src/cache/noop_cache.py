"""No-operation cache implementation."""

from cache.cache import Cache
from models.cache_entry import CacheEntry
from models.responses import ConversationData
from log import get_logger
from utils.connection_decorator import connection

logger = get_logger("cache.noop_cache")


class NoopCache(Cache):
    """No-operation cache implementation."""

    def __init__(self) -> None:
        """Create a new instance of no-op cache."""

    def connect(self) -> None:
        """
        Mark the cache as connected and log the connection attempt.
        
        This no-op implementation logs "Connecting to storage" and performs no external connection.
        """
        logger.info("Connecting to storage")

    def connected(self) -> bool:
        """
        Indicates whether the cache is connected.
        
        Returns:
            `true` if the cache is connected, `false` otherwise.
        """
        return True

    def initialize_cache(self) -> None:
        """
        Prepare the cache for use. In this NoopCache implementation, this method performs no actions.
        """

    @connection
    def get(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> list[CacheEntry]:
        """
        Retrieve cache entries for a user's conversation while validating the identifiers; this noop implementation never stores or returns entries.
        
        Parameters:
            user_id (str): User identifier; validated as a UUID unless `skip_user_id_check` is True.
            conversation_id (str): Conversation identifier; validated as a UUID unless `skip_user_id_check` is True.
            skip_user_id_check (bool): If True, skip validation of `user_id`.
        
        Returns:
            list[CacheEntry]: An empty list (this cache does not persist entries).
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
        Validate the key for the given user and conversation and accept a cache entry without persisting it.
        
        Parameters:
            user_id (str): User identifier to validate.
            conversation_id (str): Conversation identifier scoped to the user.
            cache_entry (CacheEntry): The cache entry to insert or append; not stored by this implementation.
            skip_user_id_check (bool): If True, skip validation of the user_id.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)

    @connection
    def delete(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> bool:
        """
        Delete conversation history for a specific user and conversation.
        
        Parameters:
            user_id (str): User identifier; validated as a UUID unless skip_user_id_check is True.
            conversation_id (str): Conversation identifier; validated as a UUID.
            skip_user_id_check (bool): If True, skip validation of the user_id format.
        
        Returns:
            bool: `True` in all cases.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)
        return True

    @connection
    def list(
        self, user_id: str, skip_user_id_check: bool = False
    ) -> list[ConversationData]:
        """
        Validate the user ID and return an empty list of conversations (noop implementation).
        
        Parameters:
            user_id (str): User identification to validate.
            skip_user_id_check (bool): If True, skip validation of `user_id`.
        
        Returns:
            list[ConversationData]: An empty list.
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
        Set the topic summary for a conversation.
        
        Validates the user and conversation identifiers but performs no storage.
        
        Parameters:
            user_id (str): User identifier.
            conversation_id (str): Conversation identifier for the user.
            topic_summary (str): Topic summary text to associate with the conversation.
            skip_user_id_check (bool): If true, skip validation of the user_id.
        """
        # just check if user_id and conversation_id are UUIDs
        super().construct_key(user_id, conversation_id, skip_user_id_check)

    def ready(self) -> bool:
        """
        Indicates whether the cache is ready.
        
        Returns:
            True always.
        """
        return True