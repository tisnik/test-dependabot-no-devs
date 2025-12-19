import tkinter

class ComplexFractalTypeDialog(tkinter.Toplevel):

    def __init__(self, parent):
        tkinter.Toplevel.__init__(self, parent)
        top_part = tkinter.LabelFrame(self, text="Fractal in complex plane", padx=5, pady=5)
        top_part.grid(row=1, column=1, sticky="NWSE")

        # rest
        cancelButton = tkinter.Button(self, text="Cancel", command=self.cancel)
        cancelButton.grid(row=2, column=1, sticky="NWSE")

        # close the dialog on 'x' click
        self.protocol("WM_DELETE_WINDOW", self.destroy)

        # how the buttons should behave
        self.bind("<Escape>", lambda event: self.destroy())

        self.grab_set()

    def cancel(self):
        self.destroy()

    def show(self):
        self.wm_deiconify()
        self.wait_window()
