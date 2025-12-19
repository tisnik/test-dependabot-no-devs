"""Main window shown on screen."""

import tkinter

"""Main window shown on screen."""

import tkinter
from tkinter import messagebox

from gui.canvas import Canvas
from gui.menubar import Menubar
from gui.icons import Icons
from gui.status_bar import StatusBar
from gui.toolbar import Toolbar


class MainWindow:
    """Main window shown on screen."""

    def __init__(self):
        """
        Create and configure the main application window and its primary UI components.
        
        Initializes the Tk root window, sets its title, and constructs application-wide UI elements:
        an Icons registry, Toolbar, StatusBar, Canvas, and Menubar. Arranges these widgets
        in the window grid so the canvas expands with the window. Exposes the following
        attributes on the instance: `root`, `icons`, `toolbar`, `statusbar`, `canvas`, and `menubar`.
        """
        self.root = tkinter.Tk()
        self.root.title("Svitava GUI")

        self.icons = Icons()

        self.toolbar = Toolbar(self.root, self)
        self.statusbar = StatusBar(self.root)
        self.canvas = Canvas(self.root, 800, 600, self)

        self.configure_grid()
        self.toolbar.grid(column=1, row=1, columnspan=2, sticky="WE")
        self.canvas.grid(column=1, row=2, sticky="NWSE")
        self.statusbar.grid(column=1, row=3, columnspan=2, sticky="WE")

        self.menubar = Menubar(self.root, self)
        self.root.config(menu=self.menubar)

    def show(self):
        """Display the main window on screen."""
        self.root.mainloop()

    def quit(self):
        """
        Prompt the user to confirm quitting and exit the application on confirmation.
        
        Shows a Yes/No confirmation dialog; if the user selects Yes, stops the application's main event loop.
        """
        answer = messagebox.askyesno("Do you want to quit the program?",
                                     "Do you want to quit the program?")
        if answer:
            self.root.quit()

    def configure_grid(self):
        """
        Configure the root window grid to allow the central canvas to expand with the window.
        
        Sets the row index 2 and column index 2 weights on the root grid so widgets placed there grow when the main window is resized.
        """
        tkinter.Grid.rowconfigure(self.root, 2, weight=1)
        tkinter.Grid.columnconfigure(self.root, 2, weight=1)