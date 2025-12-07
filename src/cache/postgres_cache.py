"""PostgreSQL cache implementation."""

import json
import psycopg2

from cache.cache import Cache
from cache.cache_error import CacheError
from models.cache_entry import CacheEntry
from models.config import PostgreSQLDatabaseConfiguration
from models.responses import ConversationData, ReferencedDocument
from log import get_logger
from utils.connection_decorator import connection

logger = get_logger("cache.postgres_cache")


class PostgresCache(Cache):
    """Cache that uses PostgreSQL to store cached values.

    The cache itself lives stored in following table:

    ```
         Column            |              Type              | Nullable |
    -----------------------+--------------------------------+----------+
     user_id               | text                           | not null |
     conversation_id       | text                           | not null |
     created_at            | timestamp without time zone    | not null |
     started_at            | text                           |          |
     completed_at          | text                           |          |
     query                 | text                           |          |
     response              | text                           |          |
     provider              | text                           |          |
     model                 | text                           |          |
     referenced_documents  | jsonb                           |          |
    Indexes:
        "cache_pkey" PRIMARY KEY, btree (user_id, conversation_id, created_at)
        "timestamps" btree (created_at)
    ```
    """

    CREATE_CACHE_TABLE = """
        CREATE TABLE IF NOT EXISTS cache (
            user_id              text NOT NULL,
            conversation_id      text NOT NULL,
            created_at           timestamp NOT NULL,
            started_at           text,
            completed_at         text,
            query                text,
            response             text,
            provider             text,
            model                text,
            referenced_documents jsonb,
            PRIMARY KEY(user_id, conversation_id, created_at)
        );
        """

    CREATE_CONVERSATIONS_TABLE = """
        CREATE TABLE IF NOT EXISTS conversations (
            user_id                text NOT NULL,
            conversation_id        text NOT NULL,
            topic_summary          text,
            last_message_timestamp timestamp NOT NULL,
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
         WHERE user_id=%s AND conversation_id=%s
         ORDER BY created_at
        """

    INSERT_CONVERSATION_HISTORY_STATEMENT = """
        INSERT INTO cache(user_id, conversation_id, created_at, started_at, completed_at,
                          query, response, provider, model, referenced_documents)
        VALUES (%s, %s, CURRENT_TIMESTAMP, %s, %s, %s, %s, %s, %s, %s)
        """

    QUERY_CACHE_SIZE = """
        SELECT count(*) FROM cache;
        """

    DELETE_SINGLE_CONVERSATION_STATEMENT = """
        DELETE FROM cache
         WHERE user_id=%s AND conversation_id=%s
        """

    LIST_CONVERSATIONS_STATEMENT = """
        SELECT conversation_id, topic_summary, EXTRACT(EPOCH FROM last_message_timestamp) as last_message_timestamp
          FROM conversations
         WHERE user_id=%s
         ORDER BY last_message_timestamp DESC
    """

    INSERT_OR_UPDATE_TOPIC_SUMMARY_STATEMENT = """
        INSERT INTO conversations(user_id, conversation_id, topic_summary, last_message_timestamp)
        VALUES (%s, %s, %s, CURRENT_TIMESTAMP)
        ON CONFLICT (user_id, conversation_id)
        DO UPDATE SET topic_summary = EXCLUDED.topic_summary, last_message_timestamp = EXCLUDED.last_message_timestamp
        """

    DELETE_CONVERSATION_STATEMENT = """
        DELETE FROM conversations
         WHERE user_id=%s AND conversation_id=%s
        """

    UPSERT_CONVERSATION_STATEMENT = """
        INSERT INTO conversations(user_id, conversation_id, topic_summary, last_message_timestamp)
        VALUES (%s, %s, %s, CURRENT_TIMESTAMP)
        ON CONFLICT (user_id, conversation_id)
        DO UPDATE SET last_message_timestamp = EXCLUDED.last_message_timestamp
        """

    def __init__(self, config: PostgreSQLDatabaseConfiguration) -> None:
        """
        Initialize a Postgres-backed cache using the provided database configuration.
        
        Stores the configuration on the instance and establishes the PostgreSQL connection, initializing the cache schema.
        
        Parameters:
            config (PostgreSQLDatabaseConfiguration): Configuration used to connect to the PostgreSQL server and configure the cache.
        """
        self.postgres_config = config

        # initialize connection to DB
        self.connect()
        # self.capacity = config.max_entries

    # pylint: disable=W0201
    def connect(self) -> None:
        """
        Establish a psycopg2 connection using the instance PostgreSQL configuration and initialize the cache schema.
        
        Sets self.connection to a live connection, calls initialize_cache(), and enables autocommit on the connection. If an error occurs while connecting or initializing the schema, any opened connection is closed and the original exception is propagated.
        """
        logger.info("Connecting to storage")
        # make sure the connection will have known state
        # even if PostgreSQL is not alive
        self.connection = None
        config = self.postgres_config
        try:
            self.connection = psycopg2.connect(
                host=config.host,
                port=config.port,
                user=config.user,
                password=config.password.get_secret_value(),
                dbname=config.db,
                sslmode=config.ssl_mode,
                sslrootcert=config.ca_cert_path,
                gssencmode=config.gss_encmode,
            )
            self.initialize_cache()
        except Exception as e:
            if self.connection is not None:
                self.connection.close()
            logger.exception("Error initializing Postgres cache:\n%s", e)
            raise
        self.connection.autocommit = True

    def connected(self) -> bool:
        """
        Determine whether the PostgreSQL connection is alive and responsive.
        
        Returns:
            bool: `True` if the connection exists and a simple probe query succeeds, `False` otherwise.
        """
        if self.connection is None:
            logger.warning("Not connected, need to reconnect later")
            return False
        try:
            with self.connection.cursor() as cursor:
                cursor.execute("SELECT 1")
            logger.info("Connection to storage is ok")
            return True
        except (psycopg2.OperationalError, psycopg2.InterfaceError) as e:
            logger.error("Disconnected from storage: %s", e)
            return False

    def initialize_cache(self) -> None:
        """
        Ensure the PostgreSQL schema for the cache exists by creating required tables and index.
        
        Creates the cache and conversations tables and the cache-created_at index if they do not already exist. Commits the changes to the configured database connection.
        
        Raises:
            CacheError: If the internal database connection is not established.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("Initialize_cache: cache is disconnected")

        # cursor as context manager is not used there on purpose
        # any CREATE statement can raise it's own exception
        # and it should not interfere with other statements
        cursor = self.connection.cursor()

        logger.info("Initializing table for cache")
        cursor.execute(PostgresCache.CREATE_CACHE_TABLE)

        logger.info("Initializing table for conversations")
        cursor.execute(PostgresCache.CREATE_CONVERSATIONS_TABLE)

        logger.info("Initializing index for cache")
        cursor.execute(PostgresCache.CREATE_INDEX)

        cursor.close()
        self.connection.commit()

    @connection
    def get(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> list[CacheEntry]:
        """
        Retrieve the cached conversation history for a specific user and conversation.
        
        Each database row is converted into a CacheEntry. When present, the stored
        referenced_documents JSON is deserialized into a list of ReferencedDocument
        objects; if deserialization fails a warning is logged and referenced_documents
        will be set to None for that entry.
        
        Parameters:
            user_id (str): Identifier of the user owning the conversation.
            conversation_id (str): Identifier of the conversation for the given user.
            skip_user_id_check (bool): If True, skip validation of the user_id.
        
        Returns:
            list[CacheEntry]: Cache entries for the conversation ordered by creation
            timestamp (earliest first). The list is empty if no entries exist.
        
        Raises:
            CacheError: If the cache connection is not available.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("get: cache is disconnected")

        with self.connection.cursor() as cursor:
            cursor.execute(
                self.SELECT_CONVERSATION_HISTORY_STATEMENT, (user_id, conversation_id)
            )
            conversation_entries = cursor.fetchall()

            result = []
            for conversation_entry in conversation_entries:
                # Parse referenced_documents back into ReferencedDocument objects
                docs_data = conversation_entry[6]
                docs_obj = None
                if docs_data:
                    try:
                        docs_obj = [
                            ReferencedDocument.model_validate(doc) for doc in docs_data
                        ]
                    except (ValueError, TypeError) as e:
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
        Store a cache entry for the specified user's conversation and update the conversation's metadata.
        
        Parameters:
            user_id: Identifier of the user who owns the conversation.
            conversation_id: Identifier of the conversation for the given user.
            cache_entry: CacheEntry to persist; referenced_documents (if present) will be stored as JSON.
            skip_user_id_check: If True, skip any user-id validation performed by decorators or callers.
        
        Raises:
            CacheError: If the cache is disconnected or a database error occurs while persisting the entry.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("insert_or_append: cache is disconnected")

        try:
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

            # the whole operation is run in one transaction
            with self.connection.cursor() as cursor:
                cursor.execute(
                    PostgresCache.INSERT_CONVERSATION_HISTORY_STATEMENT,
                    (
                        user_id,
                        conversation_id,
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
                    PostgresCache.UPSERT_CONVERSATION_STATEMENT,
                    (user_id, conversation_id, None),
                )
                # commit is implicit at this point
        except psycopg2.DatabaseError as e:
            logger.error("PostgresCache.insert_or_append: %s", e)
            raise CacheError("PostgresCache.insert_or_append", e) from e

    @connection
    def delete(
        self, user_id: str, conversation_id: str, skip_user_id_check: bool = False
    ) -> bool:
        """
        Delete all cached entries for a specific user's conversation.
        
        Parameters:
            user_id: Identifier of the user.
            conversation_id: Identifier of the conversation for the given user.
            skip_user_id_check: If True, bypasses user-id validation checks.
        
        Returns:
            True if any cache rows were deleted, False otherwise.
        
        Raises:
            CacheError: If the cache is disconnected or a database error occurs.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("delete: cache is disconnected")

        try:
            with self.connection.cursor() as cursor:
                cursor.execute(
                    PostgresCache.DELETE_SINGLE_CONVERSATION_STATEMENT,
                    (user_id, conversation_id),
                )
                deleted = cursor.rowcount

                # Also delete conversation record for this conversation
                cursor.execute(
                    PostgresCache.DELETE_CONVERSATION_STATEMENT,
                    (user_id, conversation_id),
                )

                return deleted > 0
        except psycopg2.DatabaseError as e:
            logger.error("PostgresCache.delete: %s", e)
            raise CacheError("PostgresCache.delete", e) from e

    @connection
    def list(
        self, user_id: str, skip_user_id_check: bool = False
    ) -> list[ConversationData]:
        """
        List all conversations for a user.
        
        Parameters:
            user_id (str): The user's identifier.
            skip_user_id_check (bool): If True, skip validation of the provided user_id.
        
        Returns:
            list[ConversationData]: Conversation metadata entries containing `conversation_id`, `topic_summary`,
            and `last_message_timestamp` (seconds since epoch).
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("list: cache is disconnected")

        with self.connection.cursor() as cursor:
            cursor.execute(self.LIST_CONVERSATIONS_STATEMENT, (user_id,))
            conversations = cursor.fetchall()

        result = []
        for conversation in conversations:
            conversation_data = ConversationData(
                conversation_id=conversation[0],
                topic_summary=conversation[1],
                last_message_timestamp=float(conversation[2]),
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
        Set or update the topic summary for a user's conversation and refresh its last_message_timestamp.
        
        Parameters:
            user_id (str): ID of the user owning the conversation.
            conversation_id (str): Conversation identifier scoped to the user.
            topic_summary (str): Text to store as the conversation's topic summary.
            skip_user_id_check (bool): If True, bypass user_id validation checks.
        """
        if self.connection is None:
            logger.error("Cache is disconnected")
            raise CacheError("set_topic_summary: cache is disconnected")

        try:
            with self.connection.cursor() as cursor:
                cursor.execute(
                    self.INSERT_OR_UPDATE_TOPIC_SUMMARY_STATEMENT,
                    (user_id, conversation_id, topic_summary),
                )
        except psycopg2.DatabaseError as e:
            logger.error("PostgresCache.set_topic_summary: %s", e)
            raise CacheError("PostgresCache.set_topic_summary", e) from e

    def ready(self) -> bool:
        """Check if the cache is ready.

        Returns:
            True if the cache is ready, False otherwise.
        """
        return True