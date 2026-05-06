"""
酒馆战棋屏幕读取模块

提供屏幕捕获、OCR识别、种族检测和游戏状态读取功能
"""

from .screen_capture import HearthstoneCapture, WindowRegion
from .ocr_minions import OCRMinions, MinionOCRResult
from .tribe_detector import TribeDetector, Tribe, TribeMatch
from .name_matcher import MinionNameMatcher, NameMatch, match_minion_name
from .game_state import (
    BattlegroundsReader,
    GameState,
    PlayerState,
    Minion,
    GamePhase
)

__all__ = [
    # Screen capture
    "HearthstoneCapture",
    "WindowRegion",
    
    # OCR
    "OCRMinions",
    "MinionOCRResult",
    
    # Tribe detection
    "TribeDetector",
    "Tribe",
    "TribeMatch",
    
    # Name matching
    "MinionNameMatcher",
    "NameMatch",
    "match_minion_name",
    
    # Game state
    "BattlegroundsReader",
    "GameState",
    "PlayerState",
    "Minion",
    "GamePhase",
]
