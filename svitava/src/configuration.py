"""Class representing configuration of Chainring."""

import configparser


class Configuration:
    """Class representing configuration of Chainring."""

    CONFIG_FILE_NAME = "config.ini"

    def __init__(self, path: str = ".") -> None:
        """Initialize the class."""
        self.config = configparser.ConfigParser()
        self.config.read(path + "/" + Configuration.CONFIG_FILE_NAME)

    @property
    def window_width(self) -> int:
        """Property holding window width."""
        return self.config.getint("ui", "window_width")

    @property
    def window_height(self) -> int:
        """Property holding window height."""
        return self.config.getint("ui", "window_height")

    def write(self) -> None:
        """Write the configuration back to disk under different name."""
        with open("config2.ini", "w") as fout:
            self.config.write(fout)

    def check_configuration_option(self, section, option) -> None:
        """Check one configuration option."""
        if not self.config.has_option(section, option):
            msg = f"V konfiguračním souboru 'config.ini' chybí volba '{option}' v sekci '{section}'"
            raise Exception(msg)

    def check_configuration(self) -> None:
        """Check the loaded configuration."""
        self.check_configuration_option("ui", "window_width")
        self.check_configuration_option("ui", "window_height")
