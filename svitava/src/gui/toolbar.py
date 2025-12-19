"""Toolbar displayed on the main windo."""

import tkinter

from gui.tooltip import Tooltip


class Toolbar(tkinter.LabelFrame):
    """Toolbar displayed on the main window."""

    def __init__(self, parent: tkinter.Tk, main_window) -> None:
        """
        Create and configure the toolbar frame and its widgets for the given main window.
        
        Initializes the LabelFrame labeled "Tools", creates the project load button with its icon and tooltip, adds a spacer, and arranges these widgets using the grid geometry manager.
        
        Parameters:
            main_window: The application main window object that provides UI resources (for example, icon references) used by toolbar widgets.
        """
        super().__init__(parent, text="Tools", padx=5, pady=5)

        self.parent = parent
        self.main_window = main_window

        self.button_project_load = tkinter.Button(
            self,
            text="Load project",
            image=main_window.icons.file_open_icon,
            command=None,
        )

        Tooltip(self.button_project_load, "Load project")

        spacer1 = tkinter.Label(self, text="   ")

        self.button_project_load.grid(column=1, row=1)
        spacer1.grid(column=2, row=1)
        self.button_project_load.grid(column=3, row=1)


    @staticmethod
    def disable_button(button) -> None:
        """
        Disable a Tkinter button widget by setting its state to "disabled".
        
        Parameters:
            button (tkinter.Button): The button widget to disable.
        """
        button["state"] = "disabled"

    @staticmethod
    def enable_button(button: tkinter.Button) -> None:
        """
        Enable the given Tkinter button widget.
        
        Parameters:
            button (tkinter.Button): The button widget to enable; its state will be set to "normal".
        """
        button["state"] = "normal"