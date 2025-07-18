"""Data archival service for packaging and sending feedback and transcripts."""

import tarfile
import tempfile
import time
from datetime import datetime, UTC
from pathlib import Path
from typing import List

import requests
import constants
from configuration import configuration
from log import get_logger

logger = get_logger(__name__)


class DataCollectorService:  # pylint: disable=too-few-public-methods
    """Service for collecting and sending user data to ingress server.

    This service handles the periodic collection and transmission of user data
    including feedback and transcripts to the configured ingress server.
    """

    def run(self) -> None:
        """
        Starts the data collection service loop, periodically collecting and sending user data files according to configuration.
        
        The loop continues until interrupted, performing data collection at configured intervals. Handles graceful shutdown on keyboard interrupt and retries after a delay on network or OS errors.
        """
        collector_config = (
            configuration.user_data_collection_configuration.data_collector
        )

        logger.info("Starting data collection service")

        while True:
            try:
                self._perform_collection()
                logger.info(
                    "Next collection scheduled in %s seconds",
                    collector_config.collection_interval,
                )
                if collector_config.collection_interval is not None:
                    time.sleep(collector_config.collection_interval)
            except KeyboardInterrupt:
                logger.info("Data collection service stopped by user")
                break
            except (OSError, requests.RequestException) as e:
                logger.error("Error during collection process: %s", e, exc_info=True)
                time.sleep(
                    constants.DATA_COLLECTOR_RETRY_INTERVAL
                )  # Wait 5 minutes before retrying on error

    def _perform_collection(self) -> None:
        """
        Performs a single cycle of data collection, packaging, and transmission.
        
        Collects feedback and transcript files if available, creates tarball archives for each data type, and sends them to the configured ingress server. Logs the outcome and raises exceptions on failure to create or send archives.
        """
        logger.info("Starting data collection process")

        # Collect files to archive
        feedback_files = self._collect_feedback_files()
        transcript_files = self._collect_transcript_files()

        if not feedback_files and not transcript_files:
            logger.info("No files to collect")
            return

        logger.info(
            "Found %s feedback files and %s transcript files to collect",
            len(feedback_files),
            len(transcript_files),
        )

        # Create and send archives
        collections_sent = 0
        try:
            if feedback_files:
                udc_config = configuration.user_data_collection_configuration
                if udc_config.feedback_storage:
                    feedback_base = Path(udc_config.feedback_storage)
                    collections_sent += self._create_and_send_tarball(
                        feedback_files, "feedback", feedback_base
                    )
            if transcript_files:
                udc_config = configuration.user_data_collection_configuration
                if udc_config.transcripts_storage:
                    transcript_base = Path(udc_config.transcripts_storage)
                    collections_sent += self._create_and_send_tarball(
                        transcript_files, "transcripts", transcript_base
                    )

            logger.info(
                "Successfully sent %s collections to ingress server", collections_sent
            )
        except (OSError, requests.RequestException, tarfile.TarError) as e:
            logger.error("Failed to create or send collections: %s", e, exc_info=True)
            raise

    def _collect_feedback_files(self) -> List[Path]:
        """
        Returns a list of feedback JSON files from the configured feedback storage directory if feedback collection is enabled and the directory exists; otherwise, returns an empty list.
        """
        udc_config = configuration.user_data_collection_configuration

        if udc_config.feedback_disabled or not udc_config.feedback_storage:
            return []

        feedback_dir = Path(udc_config.feedback_storage)
        if not feedback_dir.exists():
            return []

        return list(feedback_dir.glob("*.json"))

    def _collect_transcript_files(self) -> List[Path]:
        """
        Returns a list of transcript JSON files found recursively under the configured transcripts storage directory, or an empty list if transcript collection is disabled or the directory does not exist.
        """
        udc_config = configuration.user_data_collection_configuration

        if udc_config.transcripts_disabled or not udc_config.transcripts_storage:
            return []

        transcripts_dir = Path(udc_config.transcripts_storage)
        if not transcripts_dir.exists():
            return []

        # Recursively find all JSON files in the transcript directory structure
        return list(transcripts_dir.rglob("*.json"))

    def _create_and_send_tarball(
        self, files: List[Path], data_type: str, base_directory: Path
    ) -> int:
        """
        Creates a tarball archive from the provided files, sends it to the ingress server, and performs cleanup if configured.
        
        Parameters:
            files (List[Path]): List of file paths to include in the tarball.
            data_type (str): Type of data being archived (e.g., 'feedback', 'transcripts').
            base_directory (Path): Base directory for relative paths inside the tarball.
        
        Returns:
            int: 1 if the tarball was created and sent, 0 if no files were provided.
        """
        if not files:
            return 0

        collector_config = (
            configuration.user_data_collection_configuration.data_collector
        )

        # Create one tarball with all files
        tarball_path = self._create_tarball(files, data_type, base_directory)
        try:
            self._send_tarball(tarball_path)
            if collector_config.cleanup_after_send:
                self._cleanup_files(files)
                self._cleanup_empty_directories()
            return 1
        finally:
            self._cleanup_tarball(tarball_path)

    def _create_tarball(
        self, files: List[Path], data_type: str, base_directory: Path
    ) -> Path:
        """
        Create a gzip-compressed tarball archive containing the specified files, preserving their relative paths to the given base directory.
        
        Parameters:
            files (List[Path]): List of file paths to include in the tarball.
            data_type (str): Label used in the tarball filename to indicate the type of data.
            base_directory (Path): Base directory for calculating relative paths inside the archive.
        
        Returns:
            Path: The path to the created tarball file in the system temporary directory.
        """
        timestamp = datetime.now(UTC).strftime("%Y%m%d_%H%M%S")
        tarball_name = f"{data_type}_{timestamp}.tar.gz"

        # Create tarball in a temporary directory
        temp_dir = Path(tempfile.gettempdir())
        tarball_path = temp_dir / tarball_name

        logger.info("Creating tarball %s with %s files", tarball_path, len(files))

        with tarfile.open(tarball_path, "w:gz") as tar:
            for file_path in files:
                try:
                    # Add file with relative path to maintain directory structure
                    arcname = str(file_path.relative_to(base_directory))
                    tar.add(file_path, arcname=arcname)
                except (OSError, ValueError) as e:
                    logger.warning("Failed to add %s to tarball: %s", file_path, e)

        logger.info(
            "Created tarball %s (%s bytes)", tarball_path, tarball_path.stat().st_size
        )
        return tarball_path

    def _send_tarball(self, tarball_path: Path) -> None:
        """
        Send a tarball archive to the configured ingress server via HTTP POST.
        
        Raises:
            ValueError: If the ingress server URL is not configured.
            requests.HTTPError: If the server responds with an error status code.
        """
        collector_config = (
            configuration.user_data_collection_configuration.data_collector
        )

        if collector_config.ingress_server_url is None:
            raise ValueError("Ingress server URL is not configured")

        # pylint: disable=line-too-long
        headers = {
            "Content-Type": f"application/vnd.redhat.{collector_config.ingress_content_service_name}.periodic+tar",
        }

        if collector_config.ingress_server_auth_token:
            headers["Authorization"] = (
                f"Bearer {collector_config.ingress_server_auth_token}"
            )

        with open(tarball_path, "rb") as f:
            data = f.read()

        logger.info(
            "Sending tarball %s to %s",
            tarball_path.name,
            collector_config.ingress_server_url,
        )

        response = requests.post(
            collector_config.ingress_server_url,
            data=data,
            headers=headers,
            timeout=collector_config.connection_timeout,
        )

        if response.status_code >= 400:
            raise requests.HTTPError(
                f"Failed to send tarball to ingress server. "
                f"Status: {response.status_code}, Response: {response.text}"
            )

        logger.info("Successfully sent tarball %s to ingress server", tarball_path.name)

    def _cleanup_files(self, files: List[Path]) -> None:
        """
        Deletes the specified files from the filesystem after they have been successfully transmitted.
        
        Parameters:
        	files (List[Path]): List of file paths to remove.
        """
        for file_path in files:
            try:
                file_path.unlink()
                logger.debug("Removed file %s", file_path)
            except OSError as e:
                logger.warning("Failed to remove file %s: %s", file_path, e)

    def _cleanup_empty_directories(self) -> None:
        """
        Removes empty conversation and user directories from the transcripts storage path if transcript collection is enabled.
        
        Skips cleanup if transcripts are disabled or the storage path does not exist. Ignores errors encountered during directory removal.
        """
        udc_config = configuration.user_data_collection_configuration

        if udc_config.transcripts_disabled or not udc_config.transcripts_storage:
            return

        transcripts_dir = Path(udc_config.transcripts_storage)
        if not transcripts_dir.exists():
            return

        # Remove empty directories (conversation and user directories)
        for user_dir in transcripts_dir.iterdir():
            if user_dir.is_dir():
                for conv_dir in user_dir.iterdir():
                    if conv_dir.is_dir() and not any(conv_dir.iterdir()):
                        try:
                            conv_dir.rmdir()
                            logger.debug("Removed empty directory %s", conv_dir)
                        except OSError:
                            pass

                # Remove user directory if empty
                if not any(user_dir.iterdir()):
                    try:
                        user_dir.rmdir()
                        logger.debug("Removed empty directory %s", user_dir)
                    except OSError:
                        pass

    def _cleanup_tarball(self, tarball_path: Path) -> None:
        """
        Deletes the specified temporary tarball file after transmission.
        
        Parameters:
        	tarball_path (Path): Path to the tarball file to be removed.
        """
        try:
            tarball_path.unlink()
            logger.debug("Removed temporary tarball %s", tarball_path)
        except OSError as e:
            logger.warning("Failed to remove temporary tarball %s: %s", tarball_path, e)
