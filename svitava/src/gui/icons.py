"""All icons used on the GUI."""

import tkinter

import icons.application_exit
import icons.help_faq
import icons.help_about
import icons.fractal_new
import icons.file_open
import icons.file_save
import icons.file_save_as
import icons.filter_new
import icons.fill_color
import icons.edit
import icons.draw_arrow_forward
import icons.pattern


class Icons:
    """All icons used on the GUI."""

    def __init__(self):
        """
        Create PhotoImage objects for all GUI icons and assign them to instance attributes.
        
        Creates the following attributes, each initialized from the corresponding icons.* module data:
        - exit_icon (from icons.application_exit.icon)
        - help_faq_icon (from icons.help_faq.icon)
        - help_about_icon (from icons.help_about.icon)
        - fractal_new_icon (from icons.fractal_new.icon)
        - filter_new_icon (from icons.filter_new.icon)
        - fill_color_icon (from icons.fill_color.icon)
        - pattern_new_icon (from icons.pattern.icon)
        - draw_arrow_forward_icon (from icons.draw_arrow_forward.icon)
        - file_open_icon (from icons.file_open.icon)
        - file_save_icon (from icons.file_save.icon)
        - file_save_as_icon (from icons.file_save_as.icon)
        - edit_icon (from icons.edit.icon)
        """
        self.exit_icon = tkinter.PhotoImage(data=icons.application_exit.icon)
        self.help_faq_icon = tkinter.PhotoImage(data=icons.help_faq.icon)
        self.help_about_icon = tkinter.PhotoImage(data=icons.help_about.icon)

        self.fractal_new_icon = tkinter.PhotoImage(data=icons.fractal_new.icon)
        self.filter_new_icon = tkinter.PhotoImage(data=icons.filter_new.icon)
        self.fill_color_icon = tkinter.PhotoImage(data=icons.fill_color.icon)
        self.pattern_new_icon = tkinter.PhotoImage(data=icons.pattern.icon)

        self.draw_arrow_forward_icon = tkinter.PhotoImage(data=icons.draw_arrow_forward.icon)

        self.file_open_icon = tkinter.PhotoImage(data=icons.file_open.icon)
        self.file_save_icon = tkinter.PhotoImage(data=icons.file_save.icon)
        self.file_save_as_icon = tkinter.PhotoImage(data=icons.file_save_as.icon)

        self.edit_icon = tkinter.PhotoImage(data=icons.edit.icon)