"""Transparent always-on-top overlay window."""

import tkinter as tk
import sys
from typing import Optional, Callable
from .styles import (
    WINDOW_ALPHA, WINDOW_WIDTH, WINDOW_HEIGHT,
    WINDOW_PADDING, BG_COLOR, FONT_FAMILY, FONT_SIZE_TITLE,
    TEXT_PRIMARY, BORDER_COLOR
)


class OverlayWindow:
    """Transparent always-on-top overlay window with hotkey support."""

    def __init__(self):
        self.root: Optional[tk.Tk] = None
        self._hotkey_callbacks: dict = {}
        self._is_visible = True

    def create(self, title: str = "Overlay") -> tk.Tk:
        """Create the overlay window."""
        self.root = tk.Tk()
        self.root.title(title)
        self._configure_window()
        self._setup_hotkeys()
        return self.root

    def _configure_window(self):
        """Configure window properties."""
        if not self.root:
            return

        screen_w = self.root.winfo_screenwidth()
        screen_h = self.root.winfo_screenheight()

        x = screen_w - WINDOW_WIDTH - 20
        y = screen_h - WINDOW_HEIGHT - 60

        self.root.geometry(f"{WINDOW_WIDTH}x{WINDOW_HEIGHT}+{x}+{y}")
        self.root.attributes("-alpha", WINDOW_ALPHA)
        self.root.attributes("-topmost", True)
        self.root.configure(bg=BG_COLOR)

        self.root.overrideredirect(True)

    def _setup_hotkeys(self):
        """Setup global hotkey bindings."""
        if not self.root:
            return

        self.root.bind("<Escape>", lambda e: self.hide())
        self.root.bind("<Control-q>", lambda e: self.destroy())
        self.root.bind("<Control-r>", lambda e: self.toggle())

        # Ctrl+Shift+O to show
        self.root.bind("<Control-Shift-O>", lambda e: self.show())
        self.root.bind("<Control-Shift-H>", lambda e: self.hide())

    def register_hotkey(self, key: str, callback: Callable):
        """Register a custom hotkey callback."""
        if self.root:
            self._hotkey_callbacks[key] = callback
            self.root.bind(key, lambda e: callback())

    def show(self):
        """Show the overlay window."""
        if self.root:
            self.root.deiconify()
            self._is_visible = True

    def hide(self):
        """Hide the overlay window."""
        if self.root:
            self.root.withdraw()
            self._is_visible = False

    def toggle(self):
        """Toggle overlay visibility."""
        if self._is_visible:
            self.hide()
        else:
            self.show()

    def destroy(self):
        """Destroy the overlay window and exit."""
        if self.root:
            self.root.quit()
            self.root.destroy()
            self.root = None

    def mainloop(self):
        """Start the Tkinter main loop."""
        if self.root:
            self.root.mainloop()

    @property
    def is_visible(self) -> bool:
        return self._is_visible
