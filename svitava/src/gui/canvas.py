"""Canvas to display the vector drawing."""

import tkinter


class Canvas(tkinter.Canvas):
    """Canvas to display fractals."""

    def __init__(self, parent, width, height, main_window):
        """
        Create a Canvas configured for fractal display within the given Tkinter parent.
        
        Parameters:
        	parent: The Tkinter parent widget that will contain this canvas.
        	width (int): Width of the canvas in pixels.
        	height (int): Height of the canvas in pixels.
        	main_window: Reference to the main application window or controller using this canvas.
        """
        super().__init__(parent, width=width, height=height,
                         background="white")
        # self.draw_grid(width, height, Canvas.GRID_SIZE)
        # self.draw_boundary(width, height)
        self.width = width
        self.height = height
        self.main_window = main_window