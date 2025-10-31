"""Log utilities."""

import logging
from rich.logging import RichHandler


def get_logger(name: str) -> logging.Logger:
    """
    Return a logger configured with the specified name, set to DEBUG level and using RichHandler for enhanced console output.
    
    Parameters:
        name (str): The name of the logger to retrieve.
    
    Returns:
        logging.Logger: A logger instance with RichHandler and propagation disabled.
    """
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)
    logger.handlers = [RichHandler()]
    logger.propagate = False
    return logger
