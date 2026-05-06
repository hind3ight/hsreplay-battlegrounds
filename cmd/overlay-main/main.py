#!/usr/bin/env python3
"""Overlay main entry point.

Hotkeys:
    Ctrl+Shift+O - Show overlay
    Ctrl+Shift+H - Hide overlay
    Ctrl+R       - Toggle visibility
    Ctrl+Q       - Quit
    Escape       - Hide overlay
"""

import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from overlay.window import OverlayWindow
from overlay.panel import RecommendationPanel, RecommendationItem
from overlay.styles import (
    BG_COLOR, TEXT_PRIMARY, FONT_FAMILY, FONT_SIZE_TITLE,
    WINDOW_PADDING, WINDOW_WIDTH
)


def main():
    """Main entry point for the overlay application."""
    overlay = OverlayWindow()
    root = overlay.create("Battlegrounds Overlay")

    # Header
    header = tk.Frame(root, bg=BG_COLOR)
    header.pack(fill=tk.X, padx=WINDOW_PADDING, pady=(WINDOW_PADDING, 8))

    title_label = tk.Label(
        header,
        text="推荐面板",
        font=(FONT_FAMILY, FONT_SIZE_TITLE, "bold"),
        fg=TEXT_PRIMARY,
        bg=BG_COLOR
    )
    title_label.pack(side=tk.LEFT)

    # Panel
    panel = RecommendationPanel(root)

    # Demo data
    demo_items = [
        RecommendationItem("Engineer", "等级 1 机械", 0.95),
        RecommendationItem("Murloc Warleader", "等级 2 鱼人", 0.88),
        RecommendationItem(" Alleycat", "等级 1 野兽", 0.82),
        RecommendationItem("Wrangler", "等级 2 野兽", 0.76),
        RecommendationItem("Spawnway N/A", "等级 3 机械", 0.71),
    ]

    def on_select(index: int):
        print(f"Selected: {demo_items[index].title}")

    panel.set_on_select(on_select)
    panel.update_recommendations(demo_items)

    def refresh_demo():
        import random
        new_items = [
            RecommendationItem(
                f"Card {i+1}",
                f"类型 {random.choice(['机械', '鱼人', '野兽', '恶魔'])}",
                random.uniform(0.5, 1.0)
            )
            for i in range(5)
        ]
        panel.update_recommendations(new_items)

    overlay.register_hotkey("<Control-f>", refresh_demo)

    print("Overlay started. Hotkeys: Ctrl+Shift+O/H, Ctrl+R, Ctrl+Q, Esc")
    overlay.mainloop()


if __name__ == "__main__":
    import tkinter as tk
    main()
