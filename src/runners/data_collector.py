"""Data collector runner."""

import logging

from models.config import DataCollectorConfiguration
from services.data_collector import DataCollectorService

logger: logging.Logger = logging.getLogger(__name__)


def start_data_collector(configuration: DataCollectorConfiguration) -> None:
    """
    Starts the data collector service if enabled in the provided configuration.
    
    If data collection is disabled, the function logs this state and exits without running the service. Any exceptions raised during service execution are logged and re-raised.
    """
    logger.info("Starting data collector runner")

    if not configuration.enabled:
        logger.info("Data collection is disabled")
        return

    try:
        service = DataCollectorService()
        service.run()
    except Exception as e:
        logger.error(
            "Data collector service encountered an exception: %s", e, exc_info=True
        )
        raise
