"""Cache that uses SQLite to store cached values."""

import json
import logging
from typing import Any
from time import time

import sqlite3

from cache.cache import Cache
from models.cache_entry import CacheEntry
from models.config import SQLiteDatabaseConfiguration
from log import get_logger
from utils.connection_decorator import connection

logger = get_logger("cache.sqlite_cache")


class SQLiteCache(Cache):
    """Cache that uses SQLite to store cached values.

    The cache itself is stored in following table:

    ```
         Column      |            Type             | Nullable |
    -----------------+-----------------------------+----------+
     user_id         | text                        | not null |
     conversation_id | text                        | not null |
     created_at      | int                         | not null |
     query           | text                        |          |
     response        | text                        |          |
    Indexes:
        "cache_pkey" PRIMARY KEY, btree (user_id, conversation_id, created_at)
        "cache_key_key" UNIQUE CONSTRAINT, btree (key)
        "timestamps" btree (updated_at)
    Access method: heap
    ```
    """

    CREATE_CACHE_TABLE = """
        CREATE TABLE IF NOT EXISTS cache (
            user_id         text NOT NULL,
            conversation_id text NOT NULL,
            created_at      int NOT NULL,
            query           text,
            response        text,
            PRIMARY KEY(user_id, conversation_id, created_at)
        );
        """

    CREATE_INDEX = """
        CREATE INDEX IF NOT EXISTS timestamps
            ON cache (created_at)
        """

    SELECT_CONVERSATION_HISTORY_STATEMENT = """
        SELECT query, response
          FROM cache
         WHERE user_id=? AND conversation_id=?
         ORDER BY created_at
        """

    INSERT_CONVERSATION_HISTORY_STATEMENT = """
        INSERT INTO cache(user_id, conversation_id, created_at, query, response)
        VALUES (?, ?, ?, ?, ?)
        """

    QUERY_CACHE_SIZE = """
        SELECT count(*) FROM cache;
        """

    DELETE_SINGLE_CONVERSATION_STATEMENT = """
        DELETE FROM cache
         WHERE user_id=? AND conversation_id=?
        """

    LIST_CONVERSATIONS_STATEMENT = """
        SELECT DISTINCT conversation_id
          FROM cache
         WHERE user_id=?
         ORDER BY created_at DESC
    """

    def __init__(self, config: SQLiteDatabaseConfiguration) -> None:
        """Create a new instance of SQLite cache."""
        self.sqlite_config = config

        # initialize connection to DB
        self.connect()
        # self.capacity = config.max_entries

    # pylint: disable=W0201
    def connect(self) -> None:
        """Initialize connection to database."""
        logger.info("Connecting to storage")
        # make sure the connection will have known state
        # even if SQLite is not alive
        self.connection = None
        config = self.sqlite_config
        self.connection = sqlite3.connect(database=config.db_path)
        try:
            self.initialize_cache()
        except Exception as e:
            self.connection.close()
            logger.exception("Error initializing SQLite cache:\n%s", e)
            raise
        self.connection.autocommit = True

    def connected(self) -> bool:
        """Check if connection to cache is alive."""
        if self.connection is None:
            logger.warning("Not connected, need to reconnect later")
            return False
        try:
            cursor = self.connection.cursor()
            cursor.execute("SELECT 1")
            logger.info("Connection to storage is ok")
            return True
        except Exception as e:
            logger.error("Disconnected from storage: %s", e)
            return False

    def initialize_cache(self) -> None:
        """Initialize cache - clean it up etc."""
        cursor = self.connection.cursor()

        logger.info("Initializing table for cache")
        cursor.execute(SQLiteCache.CREATE_CACHE_TABLE)

        logger.info("Initializing index for cache")
        cursor.execute(SQLiteCache.CREATE_INDEX)

        cursor.close()
        self.connection.commit()

    @connection
    def get(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> list[CacheEntry]:
        """Get the value associated with the given key.

        Args:
            user_id: User identification.
            conversation_id: Conversation ID unique for given user.
            skip_user_id_check: Skip user_id suid check.

        Returns:
            The value associated with the key, or None if not found.
        """
        cursor = self.connection.cursor()
        cursor.execute(
            self.SELECT_CONVERSATION_HISTORY_STATEMENT, (user_id, conversation_id)
        )
        conversation_entries = cursor.fetchall()
        cursor.close()

        result = []
        for conversation_entry in conversation_entries:
            result.append(
                CacheEntry(
                    query=conversation_entry[0],
                    response=conversation_entry[1],
                    provider="?",
                    model="?",
                )
            )

        return conversation_entries

    @connection
    def insert_or_append(
        self,
        user_id: str,
        conversation_id: str,
        cache_entry: CacheEntry,
        skip_user_id_check: bool = False,
    ) -> None:
        """Set the value associated with the given key.

        Args:
            user_id: User identification.
            conversation_id: Conversation ID unique for given user.
            cache_entry: The `CacheEntry` object to store.
            skip_user_id_check: Skip user_id suid check.

        """
        cursor = self.connection.cursor()
        cursor.execute(
            self.INSERT_CONVERSATION_HISTORY_STATEMENT,
            (
                user_id,
                conversation_id,
                time(),
                cache_entry.query,
                cache_entry.response,
            ),
        )
        cursor.close()
        self.connection.commit()

    @connection
    def delete(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> bool:
        """Delete conversation history for a given user_id and conversation_id.

        Args:
            user_id: User identification.
            conversation_id: Conversation ID unique for given user.
            skip_user_id_check: Skip user_id suid check.

        Returns:
            bool: True if the conversation was deleted, False if not found.

        """
        cursor = self.connection.cursor()
        cursor.execute(
            self.DELETE_SINGLE_CONVERSATION_STATEMENT,
            (user_id, conversation_id),
        )
        cursor.close()
        self.connection.commit()

    @connection
    def list(self, user_id: str, skip_user_id_check: bool = False) -> list[str]:
        """List all conversations for a given user_id.

        Args:
            user_id: User identification.
            skip_user_id_check: Skip user_id suid check.

        Returns:
            A list of conversation ids from the cache

        """
        cursor = self.connection.cursor()
        cursor.execute(self.LIST_CONVERSATIONS_STATEMENT, (user_id,))
        conversations = cursor.fetchall()
        cursor.close()

        return [conversation[0] for conversation in conversations]

    def ready(self) -> bool:
        """Check if the cache is ready.

        Returns:
            True if the cache is ready, False otherwise.
        """
        return True
