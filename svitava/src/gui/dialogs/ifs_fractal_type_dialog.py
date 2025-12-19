import tkinter

class IFSFractalTypeDialog(tkinter.Toplevel):

    def __init__(self, parent):
        """
        Create the IFSFractalTypeDialog as a tkinter Toplevel window attached to the given parent.
        
        Parameters:
            parent (tkinter.Widget): The parent/master widget or window to which this dialog will be attached.
        """
        super().__init__(parent)