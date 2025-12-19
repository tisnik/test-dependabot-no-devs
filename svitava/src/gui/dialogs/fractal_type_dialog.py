"""Implementation of 'Select fractal type' dialog."""

#
#  (C) Copyright 2019, 2020  Pavel Tisnovsky
#
#  All rights reserved. This program and the accompanying materials
#  are made available under the terms of the Eclipse Public License v1.0
#  which accompanies this distribution, and is available at
#  http://www.eclipse.org/legal/epl-v10.html
#
#  Contributors:
#      Pavel Tisnovsky
#

import tkinter


from gui.dialogs.complex_fractal_type_dialog import ComplexFractalTypeDialog
from gui.dialogs.dynamic_fractal_type_dialog import DynamicFractalTypeDialog
from gui.dialogs.ifs_fractal_type_dialog import IFSFractalTypeDialog
from gui.dialogs.l_system_fractal_type_dialog import LSystemFractalTypeDialog


class FractalTypeDialog(tkinter.Toplevel):

    def __init__(self, parent):
        """
        Initialize the fractal type selection dialog as a modal top-level window.
        
        Creates a labeled frame titled "Fractal type", adds four selectable fractal-type buttons (complex plane, dynamic system, IFS, L-system) each with color and grayscale icons and hover behavior, and a Cancel button. Registers window close handling (window manager close and Escape key) and grabs focus to make the dialog modal.
        
        Parameters:
            parent: The parent widget for this dialog (typically a tkinter root or window).
        """
        tkinter.Toplevel.__init__(self, parent)
        top_part = tkinter.LabelFrame(self, text="Fractal type", padx=5, pady=5)
        top_part.grid(row=1, column=1, sticky="NWSE")

        _cplx_button, _cplx_icons = self.fractal_button(top_part, "In complex plane", "mandelbrot", 0, 1, on_cplx_clicked)
        _dynamic_button, _dynamic_icons = self.fractal_button(top_part, "Dynamic system", "dynamic", 0, 2, on_dynamic_clicked)
        _ifs_button, _ifs_icons = self.fractal_button(top_part, "IFS", "ifs", 1, 1, on_ifs_clicked)
        _lsystem_button, _lsystem_icons = self.fractal_button(top_part, "L-system system", "lsystem", 1, 2, on_l_system_clicked)

        # rest
        cancelButton = tkinter.Button(self, text="Cancel", command=self.cancel)
        cancelButton.grid(row=2, column=1, sticky="NWSE")

        # close the dialog on 'x' click
        self.protocol("WM_DELETE_WINDOW", self.destroy)

        # how the buttons should behave
        self.bind("<Escape>", lambda event: self.destroy())

        self.grab_set()

    def fractal_button(self, placement, text, icon_name, row, column, command):
        """
        Create and place a framed button for selecting a fractal type with hoverable icons.
        
        Parameters:
            placement: The parent tkinter widget or container where the button will be placed.
            text (str): Label displayed below the icon on the button.
            icon_name (str): Base name of the icon files located in the images/ directory (without suffix).
            row (int): Grid row index where the button will be placed.
            column (int): Grid column index where the button will be placed.
            command (callable): Callback invoked when the button is clicked.
        
        Returns:
            tuple: (button, icons) where `button` is the created tkinter.Button and `icons` is a two-tuple
            (color_photoimage, bw_photoimage) of tkinter.PhotoImage objects used for hover and default states.
        """
        icons = (
            tkinter.PhotoImage(file="images/" + icon_name + ".png"),
            tkinter.PhotoImage(file="images/" + icon_name + "_bw.png")
        )

        button = tkinter.Button(placement, command=command)
        button.config(text=text, image=icons[1], compound=tkinter.TOP)
        button.grid(row=row, column=column)

        button.bind("<Enter>", lambda e:button.config(image=icons[0]))
        button.bind("<Leave>", lambda e:button.config(image=icons[1]))

        return button, icons

    def cancel(self):
        """
        Close and destroy the dialog window.
        """
        self.destroy()

    def show(self):
        """
        Show the dialog and block execution until the window is closed.
        
        This deiconifies the toplevel window and waits for it to be destroyed, returning control after the dialog is closed.
        """
        self.wm_deiconify()
        self.wait_window()
        # return self.rooms, self.id.get()


def select_fractal_type_dialog():
    """
    Open the "Select fractal type" dialog.
    
    Creates a FractalTypeDialog with no parent, opening the modal dialog that lets the user choose a fractal type.
    """
    FractalTypeDialog(None)

def on_cplx_clicked():
    """
    Open the dialog for selecting or configuring a complex-plane fractal type.
    """
    ComplexFractalTypeDialog(None)

def on_dynamic_clicked():
    """
    Open the "Dynamic system" fractal type dialog.
    
    Instantiates a DynamicFractalTypeDialog with no parent.
    """
    DynamicFractalTypeDialog(None)

def on_ifs_clicked():
    """
    Open the IFS fractal type selection dialog.
    
    Creates an IFSFractalTypeDialog with no parent window, triggering the dialog UI for configuring an IFS fractal.
    """
    IFSFractalTypeDialog(None)

def on_l_system_clicked():
    """
    Open the L-system fractal type dialog.
    
    Instantiates the LSystemFractalTypeDialog with no parent window.
    """
    LSystemFractalTypeDialog(None)