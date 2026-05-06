"""Recommendation panel with Canvas rendering."""

import tkinter as tk
from typing import List, Dict, Callable, Optional
from .styles import (
    BG_COLOR, TEXT_PRIMARY, TEXT_SECONDARY, ITEM_HEIGHT,
    ITEM_PADDING, CORNER_RADIUS, FONT_FAMILY, FONT_SIZE_NORMAL,
    FONT_SIZE_SMALL, WINDOW_PADDING, ACCENT_HIGHLIGHT
)


class RecommendationItem:
    """Single recommendation item."""

    def __init__(self, title: str, subtitle: str = "", score: float = 0.0):
        self.title = title
        self.subtitle = subtitle
        self.score = score


class RecommendationPanel:
    """Canvas-based recommendation list panel."""

    def __init__(self, parent: tk.Widget):
        self.canvas = tk.Canvas(
            parent,
            bg=BG_COLOR,
            highlightthickness=0,
            bd=0
        )
        self.scrollbar = tk.Scrollbar(parent, orient=tk.VERTICAL, command=self.canvas.yview)
        self.frame = tk.Frame(self.canvas, bg=BG_COLOR)
        self.canvas.configure(yscrollcommand=self.scrollbar.set)

        self.scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        self.canvas.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)

        self.canvas.create_window((0, 0), window=self.frame, anchor="nw")
        self.frame.bind("<Configure>", self._on_frame_configure)

        self._item_widgets: List[tk.Frame] = []
        self._on_select_callback: Optional[Callable[[int], None]] = None

    def _on_frame_configure(self, event=None):
        self.canvas.configure(scrollregion=self.canvas.bbox("all"))

    def set_on_select(self, callback: Callable[[int], None]):
        """Set callback for item selection."""
        self._on_select_callback = callback

    def update_recommendations(self, items: List[RecommendationItem]):
        """Update the recommendation list."""
        for widget in self._item_widgets:
            widget.destroy()
        self._item_widgets.clear()

        for idx, item in enumerate(items):
            frame = tk.Frame(self.frame, bg=BG_COLOR, cursor="hand2")
            frame.pack(fill=tk.X, pady=ITEM_PADDING // 2)

            label_title = tk.Label(
                frame,
                text=item.title,
                font=(FONT_FAMILY, FONT_SIZE_NORMAL, "bold"),
                fg=TEXT_PRIMARY,
                bg=BG_COLOR,
                anchor="w"
            )
            label_title.pack(fill=tk.X, padx=WINDOW_PADDING)

            if item.subtitle:
                label_sub = tk.Label(
                    frame,
                    text=item.subtitle,
                    font=(FONT_FAMILY, FONT_SIZE_SMALL),
                    fg=TEXT_SECONDARY,
                    bg=BG_COLOR,
                    anchor="w"
                )
                label_sub.pack(fill=tk.X, padx=WINDOW_PADDING)

            frame.bind("<Button-1>", lambda e, i=idx: self._handle_click(i))
            label_title.bind("<Button-1>", lambda e, i=idx: self._handle_click(i))
            if item.subtitle:
                label_sub.bind("<Button-1>", lambda e, i=idx: self._handle_click(i))

            self._item_widgets.append(frame)

    def _handle_click(self, index: int):
        if self._on_select_callback:
            self._on_select_callback(index)
