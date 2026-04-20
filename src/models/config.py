"""Model with service configuration."""

from typing import Optional

from pydantic import BaseModel, model_validator, FilePath, AnyHttpUrl, PositiveInt
from typing_extensions import Self

import constants

from utils import checks


class TLSConfiguration(BaseModel):
    """TLS configuration."""

    tls_certificate_path: Optional[FilePath] = None
    tls_key_path: Optional[FilePath] = None
    tls_key_password: Optional[FilePath] = None

    @model_validator(mode="after")
    def check_tls_configuration(self) -> Self:
        """
        Performs post-validation for TLS configuration.
        
        Returns:
            Self: The validated TLSConfiguration instance.
        """
        return self


class ServiceConfiguration(BaseModel):
    """Service configuration."""

    host: str = "localhost"
    port: int = 8080
    auth_enabled: bool = False
    workers: int = 1
    color_log: bool = True
    access_log: bool = True
    tls_config: TLSConfiguration = TLSConfiguration()

    @model_validator(mode="after")
    def check_service_configuration(self) -> Self:
        """
        Validates the service configuration for port range and minimum worker count.
        
        Raises:
            ValueError: If the port is not between 1 and 65535, or if the number of workers is less than 1.
        
        Returns:
            Self: The validated ServiceConfiguration instance.
        """
        if self.port <= 0:
            raise ValueError("Port value should not be negative")
        if self.port > 65535:
            raise ValueError("Port value should be less than 65536")
        if self.workers < 1:
            raise ValueError("Workers must be set to at least 1")
        return self


class ModelContextProtocolServer(BaseModel):
    """model context protocol server configuration."""

    name: str
    provider_id: str = "model-context-protocol"
    url: str


class LLamaStackConfiguration(BaseModel):
    """Llama stack configuration."""

    url: Optional[str] = None
    api_key: Optional[str] = None
    use_as_library_client: Optional[bool] = None
    library_client_config_path: Optional[str] = None

    @model_validator(mode="after")
    def check_llama_stack_model(self) -> Self:
        """
        Validates the LLama stack configuration, ensuring required fields are set based on the selected mode.
        
        Raises:
            ValueError: If neither a URL nor library client mode is specified, if library client mode is enabled without a configuration file path, or if the configuration is otherwise incomplete.
            
        Returns:
            Self: The validated LLama stack configuration instance.
        """
        if self.url is None:
            if self.use_as_library_client is None:
                raise ValueError(
                    "LLama stack URL is not specified and library client mode is not specified"
                )
            if self.use_as_library_client is False:
                raise ValueError(
                    "LLama stack URL is not specified and library client mode is not enabled"
                )
        if self.use_as_library_client is None:
            self.use_as_library_client = False
        if self.use_as_library_client:
            if self.library_client_config_path is None:
                # pylint: disable=line-too-long
                raise ValueError(
                    "LLama stack library client mode is enabled but a configuration file path is not specified"  # noqa: C0301
                )
        return self


class DataCollectorConfiguration(BaseModel):
    """Data collector configuration for sending data to ingress server."""

    enabled: bool = False
    ingress_server_url: Optional[str] = None
    ingress_server_auth_token: Optional[str] = None
    ingress_content_service_name: Optional[str] = None
    collection_interval: PositiveInt = constants.DATA_COLLECTOR_COLLECTION_INTERVAL
    cleanup_after_send: bool = True  # Remove local files after successful send
    connection_timeout: PositiveInt = constants.DATA_COLLECTOR_CONNECTION_TIMEOUT

    @model_validator(mode="after")
    def check_data_collector_configuration(self) -> Self:
        """
        Validates that required fields are set when the data collector is enabled.
        
        Raises:
            ValueError: If `ingress_server_url` or `ingress_content_service_name` is missing when data collection is enabled.
        
        Returns:
            Self: The validated DataCollectorConfiguration instance.
        """
        if self.enabled and self.ingress_server_url is None:
            raise ValueError(
                "ingress_server_url is required when data collector is enabled"
            )
        if self.enabled and self.ingress_content_service_name is None:
            raise ValueError(
                "ingress_content_service_name is required when data collector is enabled"
            )
        return self


class UserDataCollection(BaseModel):
    """User data collection configuration."""

    feedback_disabled: bool = True
    feedback_storage: Optional[str] = None
    transcripts_disabled: bool = True
    transcripts_storage: Optional[str] = None
    data_collector: DataCollectorConfiguration = DataCollectorConfiguration()

    @model_validator(mode="after")
    def check_storage_location_is_set_when_needed(self) -> Self:
        """
        Validates that storage locations are set when feedback or transcripts collection is enabled.
        
        Raises:
            ValueError: If feedback is enabled but `feedback_storage` is not set, or if transcripts collection is enabled but `transcripts_storage` is not set.
        
        Returns:
            Self: The validated instance.
        """
        if not self.feedback_disabled and self.feedback_storage is None:
            raise ValueError("feedback_storage is required when feedback is enabled")
        if not self.transcripts_disabled and self.transcripts_storage is None:
            raise ValueError(
                "transcripts_storage is required when transcripts is enabled"
            )
        return self


class AuthenticationConfiguration(BaseModel):
    """Authentication configuration."""

    module: str = constants.DEFAULT_AUTHENTICATION_MODULE
    skip_tls_verification: bool = False
    k8s_cluster_api: Optional[AnyHttpUrl] = None
    k8s_ca_cert_path: Optional[FilePath] = None

    @model_validator(mode="after")
    def check_authentication_model(self) -> Self:
        """
        Validates that the authentication module is supported.
        
        Raises:
            ValueError: If the specified authentication module is not among the supported modules.
        
        Returns:
            Self: The validated AuthenticationConfiguration instance.
        """
        if self.module not in constants.SUPPORTED_AUTHENTICATION_MODULES:
            supported_modules = ", ".join(constants.SUPPORTED_AUTHENTICATION_MODULES)
            raise ValueError(
                f"Unsupported authentication module '{self.module}'. "
                f"Supported modules: {supported_modules}"
            )
        return self


class Customization(BaseModel):
    """Service customization."""

    disable_query_system_prompt: bool = False
    system_prompt_path: Optional[FilePath] = None
    system_prompt: Optional[str] = None

    @model_validator(mode="after")
    def check_customization_model(self) -> Self:
        """
        Validates and loads the system prompt from a file if a system prompt path is specified.
        
        If `system_prompt_path` is set, verifies the file exists and loads its content into the `system_prompt` attribute.
        
        Returns:
            Self: The updated instance with the system prompt loaded if applicable.
        """
        if self.system_prompt_path is not None:
            checks.file_check(self.system_prompt_path, "system prompt")
            self.system_prompt = checks.get_attribute_from_file(
                dict(self), "system_prompt_path"
            )
        return self


class Configuration(BaseModel):
    """Global service configuration."""

    name: str
    service: ServiceConfiguration
    llama_stack: LLamaStackConfiguration
    user_data_collection: UserDataCollection
    mcp_servers: list[ModelContextProtocolServer] = []
    authentication: Optional[AuthenticationConfiguration] = (
        AuthenticationConfiguration()
    )
    customization: Optional[Customization] = None

    def dump(self, filename: str = "configuration.json") -> None:
        """
        Write the current configuration to a JSON file.
        
        Parameters:
            filename (str): The path to the output file. Defaults to "configuration.json".
        """
        with open(filename, "w", encoding="utf-8") as fout:
            fout.write(self.model_dump_json(indent=4))
