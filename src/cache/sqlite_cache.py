"""Cache that uses SQLite to store cached values."""

from time import time

import sqlite3
import json

from cache.cache import Cache
from cache.cache_error import CacheError
from models.cache_entry import CacheEntry
from models.config import SQLiteDatabaseConfiguration
from models.responses import ConversationData, ReferencedDocument
from log import get_logger
from utils.connection_decorator import connection

logger = get_logger("cache.sqlite_cache")


class SQLiteCache(Cache):
    """Cache that uses SQLite to store cached values.

    The cache itself is stored in following table:

    ```
         Column            |            Type             | Nullable |
    -----------------------+-----------------------------+----------+
     user_id               | text                        | not null |
     conversation_id       | text                        | not null |
     created_at            | int                         | not null |
     started_at            | text                        |          |
     completed_at          | text                        |          |
     query                 | text                        |          |
     response              | text                        |          |
     provider              | text                        |          |
     model                 | text                        |          |
     referenced_documents  | text                        |          |
    Indexes:
        "cache_pkey" PRIMARY KEY, btree (user_id, conversation_id, created_at)
        "cache_key_key" UNIQUE CONSTRAINT, btree (key)
        "timestamps" btree (updated_at)
    Access method: heap
    ```
    """

    CREATE_CACHE_TABLE = """
        CREATE TABLE IF NOT EXISTS cache (
            user_id              text NOT NULL,
            conversation_id      text NOT NULL,
            created_at           int NOT NULL,
            started_at           text,
            completed_at         text,
            query                text,
            response             text,
            provider             text,
            model                text,
            referenced_documents text,
            PRIMARY KEY(user_id, conversation_id, created_at)
        );
        """

    CREATE_CONVERSATIONS_TABLE = """
        CREATE TABLE IF NOT EXISTS conversations (
            user_id                text NOT NULL,
            conversation_id        text NOT NULL,
            topic_summary          text,
            last_message_timestamp int NOT NULL,
            PRIMARY KEY(user_id, conversation_id)
        );
        """

    CREATE_INDEX = """
        CREATE INDEX IF NOT EXISTS timestamps
            ON cache (created_at)
        """

    SELECT_CONVERSATION_HISTORY_STATEMENT = """
        SELECT query, response, provider, model, started_at, completed_at, referenced_documents
          FROM cache
         WHERE user_id=? AND conversation_id=?
         ORDER BY created_at
        """

    INSERT_CONVERSATION_HISTORY_STATEMENT = """
        INSERT INTO cache(user_id, conversation_id, created_at, started_at, completed_at,
                          query, response, provider, model, referenced_documents)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """

    QUERY_CACHE_SIZE = """
        SELECT count(*) FROM cache;
        """

    DELETE_SINGLE_CONVERSATION_STATEMENT = """
        DELETE FROM cache
         WHERE user_id=? AND conversation_id=?
        """

    LIST_CONVERSATIONS_STATEMENT = """
        SELECT conversation_id, topic_summary, last_message_timestamp
          FROM conversations
         WHERE user_id=?
         ORDER BY last_message_timestamp DESC
    """

    INSERT_OR_UPDATE_TOPIC_SUMMARY_STATEMENT = """
        INSERT OR REPLACE INTO conversations(user_id, conversation_id, topic_summary, last_message_timestamp)
        VALUES (?, ?, ?, ?)
        """

    DELETE_CONVERSATION_STATEMENT = """
        DELETE FROM conversations
         WHERE user_id=? AND conversation_id=?
        """

    UPSERT_CONVERSATION_STATEMENT = """
        INSERT INTO conversations(user_id, conversation_id, topic_summary, last_message_timestamp)
        VALUES (?, ?, ?, ?)
        ON CONFLICT (user_id, conversation_id)
        DO UPDATE SET last_message_timestamp = excluded.last_message_timestamp
        """

    def __init__(self, config: SQLiteDatabaseConfiguration) -> None:
        """
        Initialize the SQLiteCache with the provided database configuration and establish the database connection.
        
        Parameters:
            config (SQLiteDatabaseConfiguration): Configuration containing the SQLite database path and related settings used to create and open the connection.
        """
        self.sqlite_config = config

        # initialize connection to DB
        self.connect()
        # self.capacity = config.max_entries

    # pylint: disable=W0201
    def connect(self) -> None:
        """
        Establish a SQLite connection using the configured db_path, initialize the cache schema, and enable autocommit.
        
        Raises:
            sqlite3.Error: If the database cannot be opened or the cache schema cannot be initialized.
        """
        logger.info("Connecting to storage")
        # make sure the connection will have known state
        # even if SQLite is not alive
        self.connection = None
        config = self.sqlite_config
        try:
            self.connection = sqlite3.connect(database=config.db_path)
            self.initialize_cache()
        except sqlite3.Error as e:
            if self.connection is not None:
                self.connection.close()
            logger.exception("Error initializing SQLite cache:\n%s", e)
            raise
        self.connection.autocommit = True

    def connected(self) -> bool:
        """
        Return whether the SQLite connection used by the cache is alive.
        
        Returns:
            bool: `True` if a working connection to storage exists, `False` otherwise.
        """
        if self.connection is None:
            logger.warning("Not connected, need to reconnect later")
            return False
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("SELECT 1")
            logger.info("Connection to storage is ok")
            return True
        except sqlite3.Error as e:
            logger.error("Disconnected from storage: %s", e)
            return False
        finally:
            if cursor is not None:
                try:
                    cursor.close()
                except Exception:  # pylint: disable=broad-exception-caught
                    logger.warning("Unable to close cursor")

    def initialize_cache(self) -> None:
        """
        Initialize required SQLite schema for the cache.
        
        Creates the cache table, conversations table, and index if they do not already exist, and commits the changes.
        
        Raises:
            CacheError: if the database connection is not established.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("Initialize_cache: cache is disconnected")

        cursor = self.connection.cursor()

        logger.info("Initializing table for cache")
        cursor.execute(SQLiteCache.CREATE_CACHE_TABLE)

        logger.info("Initializing table for conversations")
        cursor.execute(SQLiteCache.CREATE_CONVERSATIONS_TABLE)

        logger.info("Initializing index for cache")
        cursor.execute(SQLiteCache.CREATE_INDEX)

        cursor.close()
        self.connection.commit()

    @connection
    def get(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> list[CacheEntry]:
        """
        Return the conversation history as a list of CacheEntry objects for the given user and conversation.
        
        Parameters:
            user_id (str): User identification.
            conversation_id (str): Conversation ID unique for the given user.
            skip_user_id_check (bool): If True, skip any user-id validation checks.
        
        Returns:
            list[CacheEntry]: Conversation history entries ordered by creation time; each entry's `referenced_documents` will be a list of ReferencedDocument objects or None if not present or deserialization failed.
        
        Raises:
            CacheError: If the cache connection is disconnected.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("get: cache is disconnected")

        cursor = self.connection.cursor()
        cursor.execute(
            self.SELECT_CONVERSATION_HISTORY_STATEMENT, (user_id, conversation_id)
        )
        conversation_entries = cursor.fetchall()
        cursor.close()

        result = []
        for conversation_entry in conversation_entries:
            docs_json_str = conversation_entry[6]
            docs_obj = None
            if docs_json_str:
                try:
                    docs_data = json.loads(docs_json_str)
                    docs_obj = [
                        ReferencedDocument.model_validate(doc) for doc in docs_data
                    ]
                except (json.JSONDecodeError, ValueError) as e:
                    logger.warning(
                        "Failed to deserialize referenced_documents for "
                        "conversation %s: %s",
                        conversation_id,
                        e,
                    )
            cache_entry = CacheEntry(
                query=conversation_entry[0],
                response=conversation_entry[1],
                provider=conversation_entry[2],
                model=conversation_entry[3],
                started_at=conversation_entry[4],
                completed_at=conversation_entry[5],
                referenced_documents=docs_obj,
            )
            result.append(cache_entry)

        return result

    @connection
    def insert_or_append(
        self,
        user_id: str,
        conversation_id: str,
        cache_entry: CacheEntry,
        skip_user_id_check: bool = False,
    ) -> None:
        """
        Insert a cache entry for the specified conversation and update the conversation's last message timestamp.
        
        Parameters:
            user_id (str): Identifier of the user owning the conversation.
            conversation_id (str): Identifier of the conversation for this user.
            cache_entry (CacheEntry): The cache entry to store.
            skip_user_id_check (bool): If True, skip any user-id validation checks before storing.
        
        Raises:
            CacheError: If the cache connection is not available.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("insert_or_append: cache is disconnected")

        cursor = self.connection.cursor()
        current_time = time()

        referenced_documents_json = None
        if cache_entry.referenced_documents:
            try:
                docs_as_dicts = [
                    doc.model_dump(mode="json")
                    for doc in cache_entry.referenced_documents
                ]
                referenced_documents_json = json.dumps(docs_as_dicts)
            except (TypeError, ValueError) as e:
                logger.warning(
                    "Failed to serialize referenced_documents for "
                    "conversation %s: %s",
                    conversation_id,
                    e,
                )

        cursor.execute(
            self.INSERT_CONVERSATION_HISTORY_STATEMENT,
            (
                user_id,
                conversation_id,
                current_time,
                cache_entry.started_at,
                cache_entry.completed_at,
                cache_entry.query,
                cache_entry.response,
                cache_entry.provider,
                cache_entry.model,
                referenced_documents_json,
            ),
        )

        # Update or insert conversation record with last_message_timestamp
        cursor.execute(
            self.UPSERT_CONVERSATION_STATEMENT,
            (user_id, conversation_id, None, current_time),
        )

        cursor.close()
        self.connection.commit()

    @connection
    def delete(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> bool:
        """
        Delete all cached entries and the conversation record for the given user and conversation.
        
        Parameters:
            user_id: Identifier of the user owning the conversation.
            conversation_id: Identifier of the conversation to delete.
            skip_user_id_check: If True, skip any user-id validation checks.
        
        Returns:
            True if any cache rows for the conversation were removed, False otherwise.
        
        Raises:
            CacheError: If the cache connection is not available.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("delete: cache is disconnected")

        cursor = self.connection.cursor()
        cursor.execute(
            self.DELETE_SINGLE_CONVERSATION_STATEMENT,
            (user_id, conversation_id),
        )
        deleted = cursor.rowcount > 0

        # Also delete conversation record for this conversation
        cursor.execute(
            self.DELETE_CONVERSATION_STATEMENT,
            (user_id, conversation_id),
        )

        cursor.close()
        self.connection.commit()
        return deleted

    @connection
    def list(
        self, user_id: str, skip_user_id_check: bool = False
    ) -> list[ConversationData]:
        """
        Retrieves all conversations for the specified user.
        
        Parameters:
            user_id (str): The user's identifier.
            skip_user_id_check (bool): If True, skip validation of the user_id.
        
        Returns:
            list[ConversationData]: A list of ConversationData objects, each containing
            conversation_id, topic_summary, and last_message_timestamp.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("list: cache is disconnected")

        cursor = self.connection.cursor()
        cursor.execute(self.LIST_CONVERSATIONS_STATEMENT, (user_id,))
        conversations = cursor.fetchall()
        cursor.close()

        result = []
        for conversation in conversations:
            conversation_data = ConversationData(
                conversation_id=conversation[0],
                topic_summary=conversation[1],
                last_message_timestamp=conversation[2],
            )
            result.append(conversation_data)

        return result

    @connection
    def set_topic_summary(
        self,
        user_id: str,
        conversation_id: str,
        topic_summary: str,
        skip_user_id_check: bool = False,
    ) -> None:
        """
        Set or update the topic summary for a specific user's conversation.
        
        Parameters:
            user_id (str): Identifier of the user who owns the conversation.
            conversation_id (str): Identifier of the conversation to update.
            topic_summary (str): New topic summary text to store for the conversation.
            skip_user_id_check (bool): If True, bypass any user-id validation checks.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("set_topic_summary: cache is disconnected")

        cursor = self.connection.cursor()
        cursor.execute(
            self.INSERT_OR_UPDATE_TOPIC_SUMMARY_STATEMENT,
            (user_id, conversation_id, topic_summary, time()),
        )
        cursor.close()
        self.connection.commit()

    def ready(self) -> bool:
        """Check if the cache is ready.

        Returns:
            True if the cache is ready, False otherwise.
        """
        return True