import tkinter

class ComplexFractalTypeDialog(tkinter.Toplevel):

    def __init__(self, parent):
        """
        Create a modal top-level dialog for selecting a fractal in the complex plane.
        
        Initializes the Toplevel window as a child of `parent`, builds the dialog's labeled frame and a Cancel button, binds the window close button and Escape key to close the dialog, and grabs focus to make the dialog modal.
        
        Parameters:
            parent (tkinter.Widget): Parent widget or window that owns this dialog.
        """
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
        """
        Close the dialog window.
        
        Destroys the Toplevel window, closing the dialog and releasing its associated resources.
        """
        self.destroy()

    def show(self):
        """
        Display the dialog (restore if minimized) and block until the window is closed.
        
        This brings the dialog to the foreground if it was minimized and then waits for the dialog window to be destroyed, preventing code execution from continuing until the user closes the dialog.
        """
        self.wm_deiconify()
        self.wait_window()