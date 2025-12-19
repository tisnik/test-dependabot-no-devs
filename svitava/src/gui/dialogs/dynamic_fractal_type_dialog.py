import tkinter

class DynamicFractalTypeDialog(tkinter.Toplevel):

    def __init__(self, parent):
        """
        Initialize the DynamicFractalTypeDialog as a Toplevel window attached to the given parent.
        
        Parameters:
            parent: The parent Tk or widget to which this dialog window will be attached.
        """
        super().__init__(parent)