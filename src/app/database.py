"""Database engine management."""

from pathlib import Path
from typing import Any

from sqlalchemy import create_engine, text
from sqlalchemy.engine.base import Engine
from sqlalchemy.orm import sessionmaker, Session
from log import get_logger, logging
from configuration import configuration
from models.database.base import Base
from models.config import SQLiteDatabaseConfiguration, PostgreSQLDatabaseConfiguration

logger = get_logger(__name__)

engine: Engine | None = None
session_local: sessionmaker | None = None


def get_engine() -> Engine:
    """
    Return the global database engine.
    
    Returns:
        Engine: The initialized SQLAlchemy Engine instance.
    
    Raises:
        RuntimeError: If the database engine has not been initialized; call initialize_database() first.
    """
    if engine is None:
        raise RuntimeError(
            "Database engine not initialized. Call initialize_database() first."
        )
    return engine


def create_tables() -> None:
    """
    Create all ORM tables defined on Base.metadata using the currently initialized engine.
    
    Raises:
        RuntimeError: If the global database engine is not initialized (call initialize_database() first).
    """
    Base.metadata.create_all(get_engine())


def get_session() -> Session:
    """
    Provide a new ORM Session bound to the configured engine.
    
    Returns:
        Session: A SQLAlchemy ORM Session instance bound to the initialized engine.
    
    Raises:
        RuntimeError: If the database has not been initialized; call initialize_database() first.
    """
    if session_local is None:
        raise RuntimeError(
            "Database session not initialized. Call initialize_database() first."
        )
    return session_local()


def _create_sqlite_engine(config: SQLiteDatabaseConfiguration, **kwargs: Any) -> Engine:
    """
    Create and return a SQLAlchemy Engine connected to the given SQLite database file.
    
    Parameters:
        config (SQLiteDatabaseConfiguration): Configuration containing `db_path` for the SQLite file.
        **kwargs: Additional keyword arguments forwarded to SQLAlchemy's create_engine.
    
    Returns:
        Engine: A SQLAlchemy Engine bound to the specified SQLite database file.
    
    Raises:
        FileNotFoundError: If the parent directory of `config.db_path` does not exist.
        RuntimeError: If engine creation fails.
    """
    if not Path(config.db_path).parent.exists():
        raise FileNotFoundError(
            f"SQLite database directory does not exist: {config.db_path}"
        )

    try:
        return create_engine(f"sqlite:///{config.db_path}", **kwargs)
    except Exception as e:
        logger.exception("Failed to create SQLite engine")
        raise RuntimeError(f"SQLite engine creation failed: {e}") from e


def _create_postgres_engine(
    config: PostgreSQLDatabaseConfiguration, **kwargs: Any
) -> Engine:
    """
    Create a SQLAlchemy Engine configured for a PostgreSQL database using the provided configuration.
    
    Builds a connection URL from the configuration and creates an Engine. If the configuration specifies a non-default namespace (a schema other than "public"), ensures that schema exists by creating it if necessary. If a CA certificate path is provided, the engine will be configured to use it for SSL.
    
    Parameters:
        config (PostgreSQLDatabaseConfiguration): Connection and database settings (user, password, host, port, db, ssl/gss options, optional namespace and ca_cert_path).
    
    Returns:
        Engine: A SQLAlchemy Engine connected to the configured PostgreSQL database.
    
    Raises:
        RuntimeError: If engine creation fails or if creating the specified schema fails.
    """
    postgres_url = (
        f"postgresql://{config.user}:{config.password.get_secret_value()}@"
        f"{config.host}:{config.port}/{config.db}"
        f"?sslmode={config.ssl_mode}&gssencmode={config.gss_encmode}"
    )

    is_custom_schema = config.namespace is not None and config.namespace != "public"

    connect_args = {}
    if is_custom_schema:
        connect_args["options"] = f"-csearch_path={config.namespace}"

    if config.ca_cert_path is not None:
        connect_args["sslrootcert"] = str(config.ca_cert_path)

    try:
        postgres_engine = create_engine(
            postgres_url, connect_args=connect_args, **kwargs
        )
    except Exception as e:
        logger.exception("Failed to create PostgreSQL engine")
        raise RuntimeError(f"PostgreSQL engine creation failed: {e}") from e

    if is_custom_schema:
        try:
            with postgres_engine.connect() as connection:
                connection.execute(
                    text(f'CREATE SCHEMA IF NOT EXISTS "{config.namespace}"')
                )
                connection.commit()
                logger.info("Schema '%s' created or already exists", config.namespace)
        except Exception as e:
            logger.exception("Failed to create schema '%s'", config.namespace)
            raise RuntimeError(
                f"Schema creation failed for '{config.namespace}': {e}"
            ) from e

    return postgres_engine


def initialize_database() -> None:
    """
    Initialize module-level database engine and session factory from the application's configuration.
    
    Reads configuration.database_configuration to determine the database type (SQLite or PostgreSQL), creates and assigns a module-level `engine`, and initializes `session_local` as a sessionmaker bound to that engine. The engine is configured to echo SQL when the logger is at DEBUG level and to use connection pre-ping. May raise RuntimeError if engine creation or required schema creation fails.
    """
    db_config = configuration.database_configuration

    global engine, session_local  # pylint: disable=global-statement

    # Debug print all SQL statements if our logger is at-least DEBUG level
    echo = bool(logger.isEnabledFor(logging.DEBUG))

    create_engine_kwargs = {
        "echo": echo,
        "pool_pre_ping": True,
    }

    match db_config.db_type:
        case "sqlite":
            logger.info("Initialize SQLite database")
            sqlite_config = db_config.config
            logger.debug("Configuration: %s", sqlite_config)
            assert isinstance(sqlite_config, SQLiteDatabaseConfiguration)
            engine = _create_sqlite_engine(sqlite_config, **create_engine_kwargs)
        case "postgres":
            logger.info("Initialize PostgreSQL database")
            postgres_config = db_config.config
            logger.debug("Configuration: %s", postgres_config)
            assert isinstance(postgres_config, PostgreSQLDatabaseConfiguration)
            engine = _create_postgres_engine(postgres_config, **create_engine_kwargs)

    session_local = sessionmaker(autocommit=False, autoflush=False, bind=engine)