"""Class representing configuration of Chainring."""

import configparser


class Configuration:
    """Class representing configuration of Chainring."""

    CONFIG_FILE_NAME = "config.ini"

    def __init__(self, path: str = ".") -> None:
        """
        Initialize the Configuration instance and load the configuration file.
        
        Parameters:
            path (str): Directory where Configuration.CONFIG_FILE_NAME is located; defaults to the current directory. The parsed configuration is stored in the instance attribute `config`.
        """
        self.config = configparser.ConfigParser()
        self.config.read(path + "/" + Configuration.CONFIG_FILE_NAME)

    @property
    def window_width(self) -> int:
        """
        Provide window width from the "ui" section of the loaded configuration.
        
        Returns:
            int: The window width value from the configuration.
        """
        return self.config.getint("ui", "window_width")

    @property
    def window_height(self) -> int:
        """
        Get the configured UI window height.
        
        Returns:
            int: Window height read from the "ui" section's "window_height" option.
        """
        return self.config.getint("ui", "window_height")

    def write(self) -> None:
        """
        Write the current in-memory configuration to a file named "config2.ini".
        
        Overwrites any existing file with that name.
        """
        with open("config2.ini", "w") as fout:
            self.config.write(fout)

    def check_configuration_option(self, section, option) -> None:
        """
        Verify that an option exists in the given configuration section.
        
        Parameters:
            section (str): Name of the configuration section to check.
            option (str): Name of the option within the section.
        
        Raises:
            Exception: If the option is missing; the exception message indicates the missing option and section in 'config.ini' (Czech).
        """
        if not self.config.has_option(section, option):
            msg = f"V konfiguračním souboru 'config.ini' chybí volba '{option}' v sekci '{section}'"
            raise Exception(msg)

    def check_configuration(self) -> None:
        """
        Validate that the loaded configuration contains the required UI size options.
        
        Raises:
            Exception: if the 'window_width' or 'window_height' option is missing in the 'ui' section.
        """
        self.check_configuration_option("ui", "window_width")
        self.check_configuration_option("ui", "window_height")