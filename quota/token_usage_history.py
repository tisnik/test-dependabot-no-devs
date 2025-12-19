"""Class with implementation of storage for token usage history.

One table named `token_usage` is used to store statistic about token usage
history. Input and output token count are stored for each triple (user_id,
provider, model). This triple is also used as a primary key to this table.
"""

from log import get_logger
import logging
import sqlite3
from datetime import datetime
from typing import Optional

import psycopg2

from quota.connect_pg import connect_pg
from quota.connect_sqlite import connect_sqlite
from quota.sql import (
    CREATE_TOKEN_USAGE_TABLE,
    INIT_TOKEN_USAGE_FOR_USER,
    CONSUME_TOKENS_FOR_USER_PG,
    CONSUME_TOKENS_FOR_USER_SQLITE,
)

from models.config import QuotaHandlersConfiguration
from utils.connection_decorator import connection

logger = get_logger(__name__)


class TokenUsageHistory:
    """Class with implementation of storage for token usage history."""

    def __init__(self, configuration: QuotaHandlersConfiguration) -> None:
        """Initialize token usage history storage."""
        # store the configuration, it will be used
        # by reconnection logic later, if needed
        self.sqlite_connection_config = configuration.sqlite
        self.postgres_connection_config = configuration.postgres
        self.connection = None

        # initialize connection to DB
        self.connect()

    # pylint: disable=W0201
    def connect(self) -> None:
        """Initialize connection to database."""
        logger.info("Initializing connection to quota usage history database")
        if self.postgres_connection_config is not None:
            self.connection = connect_pg(self.postgres_connection_config)
        if self.sqlite_connection_config is not None:
            self.connection = connect_sqlite(self.sqlite_connection_config)

        try:
            self._initialize_tables()
        except Exception as e:
            self.connection.close()
            logger.exception("Error initializing quota usage history database:\n%s", e)
            raise

        self.connection.autocommit = True

    @connection
    def consume_tokens(
        self,
        user_id: str,
        provider: str,
        model: str,
        input_tokens: int,
        output_tokens: int,
    ) -> None:
        """Consume tokens by given user."""
        logger.info(
            "Token usage for user %s, provider %s and mode %s changed by %d, %d tokens",
            user_id,
            provider,
            model,
            input_tokens,
            output_tokens,
        )
        query_statement: str = ""
        if self.postgres_connection_config is not None:
            query_statement = CONSUME_TOKENS_FOR_USER_PG
        if self.sqlite_connection_config is not None:
            query_statement = CONSUME_TOKENS_FOR_USER_SQLITE

        # timestamp to be used
        updated_at = datetime.now()

        # it is not possible to use context manager there, because SQLite does
        # not support it
        cursor = self.connection.cursor()
        cursor.execute(
            query_statement,
            {
                "user_id": user_id,
                "provider": provider,
                "model": model,
                "input_tokens": input_tokens,
                "output_tokens": output_tokens,
                "updated_at": updated_at,
            },
        )
        cursor.close()

    def connected(self) -> bool:
        """Check if connection to quota usage history database is alive."""
        if self.connection is None:
            logger.warning("Not connected, need to reconnect later")
            return False
        cursor = None
        try:
            cursor = self.connection.cursor()
            cursor.execute("SELECT 1")
            logger.info("Connection to storage is ok")
            return True
        except (psycopg2.OperationalError, sqlite3.Error) as e:
            logger.error("Disconnected from storage: %s", e)
            return False
        finally:
            if cursor is not None:
                try:
                    cursor.close()
                except Exception:  # pylint: disable=broad-exception-caught
                    logger.warning("Unable to close cursor")

    def _initialize_tables(self) -> None:
        """Initialize tables used by quota limiter."""
        logger.info("Initializing tables for token usage history")
        cursor = self.connection.cursor()
        cursor.execute(CREATE_TOKEN_USAGE_TABLE)
        cursor.close()
        self.connection.commit()
