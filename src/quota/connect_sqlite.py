"""SQLite connection handler."""

import sqlite3
from typing import Any

from log import get_logger
from models.config import SQLiteDatabaseConfiguration

logger = get_logger(__name__)


def connect_sqlite(config: SQLiteDatabaseConfiguration) -> Any:
    """Initialize connection to database."""
    logger.info("Connecting to SQLite storage")
    # make sure the connection will have known state
    # even if SQLite is not alive
    connection = None
    try:
        connection = sqlite3.connect(database=config.db_path)
        if connection is not None:
            connection.autocommit = True
        return connection
    except sqlite3.Error as e:
        logger.exception("Error initializing SQLite cache:\n%s", e)
        raise
