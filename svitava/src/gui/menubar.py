"""Menu bar displayed on the main window."""

import tkinter

from gui.dialogs.about_dialog import about
from gui.dialogs.help_dialog import help
from gui.dialogs.fractal_type_dialog import select_fractal_type_dialog


class Menubar(tkinter.Menu):
    """Menu bar displayed on the main window."""

    def __init__(self, parent, main_window):
        """Initialize the menu bar."""
        super().__init__(tearoff=0)

        self.parent = parent
        self.main_window = main_window

        self.file_menu = tkinter.Menu(self, tearoff=0)
        self.file_menu.add_command(label="Load project", image=main_window.icons.file_open_icon,
                                    compound="left", underline=0, accelerator="Ctrl+L")
        self.file_menu.add_command(label="Save project", image=main_window.icons.file_save_icon,
                                    compound="left", underline=0, accelerator="Ctrl+S")
        self.file_menu.add_separator()
        self.file_menu.add_command(label="Quit", image=main_window.icons.exit_icon,
                                  compound="left", underline=0, accelerator="Ctrl+Q",
                                  command=parent.quit)

        self.renderer_menu = tkinter.Menu(self, tearoff=0)
        self.renderer_menu.add_command(label="New fractal", image=main_window.icons.fractal_new_icon,
                                     compound="left", underline=0, accelerator="Ctrl+N",
                                     command=self.command_fractal_type_dialog)
        self.renderer_menu.add_separator()
        self.renderer_menu.add_command(label="New pattern", image=main_window.icons.pattern_new_icon,
                                     compound="left", underline=0, accelerator="Ctrl+P",
                                     command=self.command_fractal_type_dialog)

        self.compositor_menu = tkinter.Menu(self, tearoff=0)
        self.compositor_menu.add_command(label="New filter", image=main_window.icons.filter_new_icon,
                                     compound="left", underline=4, accelerator="Ctrl+F",
                                     command=None)
        self.compositor_menu.add_command(label="New color mapper", image=main_window.icons.fill_color_icon,
                                     compound="left", underline=4, accelerator="Ctrl+C",
                                     command=None)
        self.compositor_menu.add_separator()

        self.palette_menu = tkinter.Menu(self, tearoff=0)
        self.palette_menu.add_command(label="Load palette", image=main_window.icons.file_open_icon,
                                    compound="left", underline=0)
        self.palette_menu.add_command(label="Save palette", image=main_window.icons.file_save_icon,
                                    compound="left", underline=0)
        self.palette_menu.add_command(label="Palette editor", image=main_window.icons.edit_icon,
                                    compound="left", underline=0)

        self.help_menu = tkinter.Menu(self, tearoff=0)
        self.help_menu.add_command(label="Help",
                                  image=main_window.icons.help_faq_icon,
                                  compound="left", underline=0, accelerator="F1",
                                  command=help)
        self.help_menu.add_separator()
        self.help_menu.add_command(label="About",
                                  image=main_window.icons.help_about_icon, accelerator="F11",
                                  compound="left", underline=0, command=about)

        self.add_cascade(label="File", menu=self.file_menu, underline=0, font = ("", 12)) ## TODO
        self.add_cascade(label="Renderer", menu=self.renderer_menu, underline=0)
        self.add_cascade(label="Compositor", menu=self.compositor_menu, underline=0)
        self.add_cascade(label="Palette", menu=self.palette_menu, underline=0)
        self.add_cascade(label="Help", menu=self.help_menu, underline=0)

        self.parent.bind('<F1>', lambda event: help())
        self.parent.bind('<F11>', lambda event: about())
        self.parent.bind('<Control-n>', lambda event: self.command_fractal_type_dialog())

    def command_fractal_type_dialog(self):
        select_fractal_type_dialog()
