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
        tkinter.Toplevel.__init__(self, parent)
        top_part = tkinter.LabelFrame(self, text="Fractal type", padx=5, pady=5)
        top_part.grid(row=1, column=1, sticky="NWSE")

        cplx_button, cplx_icons = self.fractal_button(top_part, "In complex plane", "mandelbrot", 0, 1, on_cplx_clicked)
        ifs_button, ifs_icons = self.fractal_button(top_part, "Dynamic system", "dynamic", 0, 2, on_dynamic_clicked)
        ifs_button, ifs_icons = self.fractal_button(top_part, "IFS", "ifs", 1, 1, on_ifs_clicked)
        ifs_button, ifs_icons = self.fractal_button(top_part, "L-system system", "lsystem", 1, 2, on_l_system_clicked)

        # rest
        cancelButton = tkinter.Button(self, text="Cancel", command=self.cancel)
        cancelButton.grid(row=2, column=1, sticky="NWSE")

        # close the dialog on 'x' click
        self.protocol("WM_DELETE_WINDOW", self.destroy)

        # how the buttons should behave
        self.bind("<Escape>", lambda event: self.destroy())

        self.grab_set()

    def fractal_button(self, placement, text, icon_name, row, column, command):
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
        self.destroy()

    def show(self):
        self.wm_deiconify()
        self.wait_window()
        # return self.rooms, self.id.get()


def select_fractal_type_dialog():
    FractalTypeDialog(None)

def on_cplx_clicked():
    ComplexFractalTypeDialog(None)

def on_dynamic_clicked():
    DynamicFractalTypeDialog(None)

def on_ifs_clicked():
    IFSFractalTypeDialog(None)

def on_l_system_clicked():
    LSystemFractalTypeDialog(None)
