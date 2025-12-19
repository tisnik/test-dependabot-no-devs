"""Canvas to display the vector drawing."""

import tkinter


class Canvas(tkinter.Canvas):
    """Canvas to display fractals."""

    def __init__(self, parent, width, height, main_window):
        """Initialize canvas."""
        super().__init__(parent, width=width, height=height,
                         background="white")
        # self.draw_grid(width, height, Canvas.GRID_SIZE)
        # self.draw_boundary(width, height)
        self.width = width
        self.height = height
        self.main_window = main_window
