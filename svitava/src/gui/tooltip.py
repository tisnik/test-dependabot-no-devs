"""Create a tooltip for a given widget."""

# taken from:
# https://stackoverflow.com/questions/3221956/how-do-i-display-tooltips-in-tkinter#36221216

import tkinter


class Tooltip:
    """Create a tooltip for a given widget."""

    def __init__(self, widget: tkinter.Button, text: str = "widget info") -> None:
        """
        Create a Tooltip attached to the given tkinter widget.
        
        Initializes tooltip configuration (delay and wrap length), stores the widget and text, and binds the widget's <Enter>, <Leave>, and <ButtonPress> events to manage showing and hiding the tooltip.
        
        Parameters:
            widget (tkinter.Button): Target widget to attach the tooltip to.
            text (str): Text to display inside the tooltip. Defaults to "widget info".
        """
        self.waittime = 500  # miliseconds
        self.wraplength = 180  # pixels
        self.widget = widget
        self.text = text
        self.widget.bind("<Enter>", self.enter)
        self.widget.bind("<Leave>", self.leave)
        self.widget.bind("<ButtonPress>", self.leave)
"""Create a tooltip for a given widget."""

# taken from:
# https://stackoverflow.com/questions/3221956/how-do-i-display-tooltips-in-tkinter#36221216

import tkinter
from typing import Optional
        self.tw = None

    def enter(self, event: tkinter.Event | None = None) -> None:
        """
        Schedule the tooltip to appear after the configured delay when the pointer enters the widget.
        
        Parameters:
            event (tkinter.Event | None): Optional event from the widget's Enter binding (ignored).
        """
        self.schedule()

    def leave(self, event: tkinter.Event | None = None) -> None:
        """
        Cancel any scheduled tooltip display and hide the tooltip when the pointer leaves the widget.
        
        Parameters:
            event (tkinter.Event | None): The leave event from the widget, or None if invoked directly.
        """
        self.unschedule()
        self.hidetip()

    def schedule(self) -> None:
        """
        Schedule showing the tooltip after the configured delay.
        
        Cancels any previously scheduled show and sets a new timer to call `showtip` after `waittime` milliseconds.
        """
        self.unschedule()
        self.id = self.widget.after(self.waittime, self.showtip)

    def unschedule(self) -> None:
        """
        Cancel any pending scheduled tooltip display and clear the stored schedule id.
        
        If a callback was scheduled with the widget's `after`, cancels it via `widget.after_cancel`; does nothing when no schedule exists.
        """
        id = self.id
        self.id = None
        if id:
            self.widget.after_cancel(id)

    def showtip(self, event=None) -> None:
        """
        Display the tooltip near the associated widget.
        
        Creates a borderless toplevel window positioned adjacent to the widget and shows the configured tooltip text.
        
        Parameters:
        	event (tkinter.Event | None): Optional event that triggered showing the tooltip; ignored by this method.
        """
        x = y = 0
        x, y, cx, cy = self.widget.bbox("insert")
        x += self.widget.winfo_rootx() + 25
        y += self.widget.winfo_rooty() + 20
        # creates a toplevel window
        self.tw = tkinter.Toplevel(self.widget)
        # Leaves only the label and removes the app window
        self.tw.wm_overrideredirect(True)
        self.tw.wm_geometry("+%d+%d" % (x, y))
        label = tkinter.Label(
            self.tw,
            text=self.text,
            justify="left",
            background="#ffffff",
            relief="solid",
            borderwidth=1,
            wraplength=self.wraplength,
        )
        label.pack(ipadx=1)

    def hidetip(self) -> None:
        """
        Hide the tooltip window if present and clear the internal reference.
        
        This destroys the tooltip's toplevel window (if one exists) and sets the
        internal tooltip attribute to None so the tooltip is no longer shown.
        """
        tw = self.tw
        self.tw = None
        if tw:
            tw.destroy()