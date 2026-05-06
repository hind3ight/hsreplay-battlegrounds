"""
游戏状态整合模块 - 整合屏幕捕获、OCR 和种族识别
"""
import numpy as np
from dataclasses import dataclass, field
from typing import List, Optional, Dict
from enum import Enum

from .screen_capture import HearthstoneCapture, WindowRegion
from .ocr_minions import OCRMinions, MinionOCRResult
from .tribe_detector import TribeDetector, Tribe
from .name_matcher import MinionNameMatcher, NameMatch


class GamePhase(Enum):
    """游戏阶段"""
    UNKNOWN = "unknown"
    LOBBY = "lobby"
    SHOP = "shop"
    COMBAT = "combat"
    END = "end"


@dataclass
class Minion:
    """随从数据类"""
    name: str                    # 英文名
    name_cn: str = ""            # 中文名
    tribe: Tribe = Tribe.UNKNOWN # 种族
    tier: int = 0                # 随从等级 (1-6)
    attack: int = 0              # 攻击力
    health: int = 0              # 生命值
    attack_ocr: int = 0          # OCR 识别的攻击力
    health_ocr: int = 0          # OCR 识别的生命值
    name_confidence: float = 0.0  # 名称匹配置信度
    position: int = 0            # 位置 (0-6)
    
    def to_dict(self) -> Dict:
        return {
            "name": self.name,
            "name_cn": self.name_cn,
            "tribe": self.tribe.value if isinstance(self.tribe, Tribe) else str(self.tribe),
            "tier": self.tier,
            "attack": self.attack or self.attack_ocr,
            "health": self.health or self.health_ocr,
            "position": self.position
        }


@dataclass
class PlayerState:
    """玩家状态"""
    health: int = 40                    # 生命值
    gold: int = 0                        # 金币
    tier: int = 1                        # 酒馆等级 (1-6)
    tavern_level: int = 1                # 酒馆等级 (同 tier)
    board_minions: List[Minion] = field(default_factory=list)  # 场上随从
    bench_minions: List[Minion] = field(default_factory=list) # 备选随从
    hand_minions: List[Minion] = field(default_factory=list)  # 手牌随从
    
    def total_minions(self) -> int:
        return len(self.board_minions) + len(self.bench_minions) + len(self.hand_minions)
    
    def to_dict(self) -> Dict:
        return {
            "health": self.health,
            "gold": self.gold,
            "tier": self.tier,
            "board_minions": [m.to_dict() for m in self.board_minions],
            "bench_minions": [m.to_dict() for m in self.bench_minions],
            "hand_minions": [m.to_dict() for m in self.hand_minions]
        }


@dataclass
class GameState:
    """游戏状态"""
    phase: GamePhase = GamePhase.UNKNOWN    # 游戏阶段
    turn: int = 0                            # 回合数
    player: PlayerState = field(default_factory=PlayerState)
    enemy: PlayerState = field(default_factory=PlayerState)
    timestamp: float = 0.0                   # 时间戳
    
    def to_dict(self) -> Dict:
        return {
            "phase": self.phase.value,
            "turn": self.turn,
            "player": self.player.to_dict(),
            "enemy": self.enemy.to_dict(),
            "timestamp": self.timestamp
        }


class BattlegroundsReader:
    """酒馆战棋读取器主类"""
    
    def __init__(self, 
                 assets_dir: str = "assets",
                 data_dir: str = "data"):
        """
        初始化读取器
        
        Args:
            assets_dir: 资源目录 (包含种族图标等)
            data_dir: 数据目录 (包含元数据 JSON)
        """
        self.assets_dir = assets_dir
        self.data_dir = data_dir
        
        # 初始化各组件
        self.capture = HearthstoneCapture()
        self.ocr = OCRMinions()
        self.tribe_detector = TribeDetector(icons_dir=f"{assets_dir}/tribe_icons")
        self.name_matcher = MinionNameMatcher(
            metadata_path=f"{data_dir}/minions_metadata.json",
            names_path=f"{data_dir}/minion_names.json"
        )
        
        # 当前游戏状态
        self.game_state = GameState()
    
    def capture_and_process(self) -> GameState:
        """
        捕获屏幕并处理整个游戏状态
        
        Returns:
            游戏状态对象
        """
        import time
        self.game_state.timestamp = time.time()
        
        # 1. 捕获随从栏
        minion_bar_img = self.capture.capture_minion_bar()
        
        # 2. OCR 识别随从名称和属性
        ocr_results = self.ocr.recognize_minion_bar(minion_bar_img)
        
        # 3. 种族检测
        preprocessed = self.ocr.preprocess_image(minion_bar_img)
        detected_tribes = self.tribe_detector.detect_tribes_in_bar(preprocessed)
        
        # 4. 名称匹配
        minion_names = [r.name for r in ocr_results]
        name_matches = self.name_matcher.match_batch(minion_names)
        
        # 5. 整合结果
        board_minions = []
        for i, (ocr_res, tribe, name_match) in enumerate(zip(ocr_results, detected_tribes, name_matches)):
            minion = Minion(
                name=name_match.matched_name if name_match else ocr_res.name,
                name_cn=name_match.name_cn if name_match else "",
                tribe=tribe,
                tier=name_match.tier if name_match else 0,
                attack=ocr_res.attack,
                health=ocr_res.health,
                attack_ocr=ocr_res.attack,
                health_ocr=ocr_res.health,
                name_confidence=name_match.score if name_match else ocr_res.confidence,
                position=i
            )
            board_minions.append(minion)
        
        self.game_state.player.board_minions = board_minions
        
        return self.game_state
    
    def read_player_health(self) -> int:
        """读取玩家生命值 (需要根据实际游戏界面调整区域)"""
        # TODO: 实现玩家生命值读取
        return self.game_state.player.health
    
    def read_player_gold(self) -> int:
        """读取玩家金币 (需要根据实际游戏界面调整区域)"""
        # TODO: 实现玩家金币读取
        return self.game_state.player.gold
    
    def read_tavern_tier(self) -> int:
        """读取酒馆等级"""
        # TODO: 实现酒馆等级读取
        return self.game_state.player.tier
    
    def set_capture_region(self, left: int, top: int, width: int, height: int):
        """设置随从栏捕获区域"""
        self.capture.set_minion_bar_region(left, top, width, height)
    
    def close(self):
        """关闭所有资源"""
        self.capture.close()
        self.ocr.close()
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


def create_test_state() -> GameState:
    """创建测试用游戏状态"""
    state = GameState()
    state.phase = GamePhase.SHOP
    state.turn = 5
    
    state.player.health = 32
    state.player.gold = 6
    state.player.tier = 3
    
    # 测试随从
    minions = [
        Minion(name="Brann Bronzebeard", name_cn="恐龙大师布莱恩", 
               tribe=Tribe.BEAST, tier=2, attack=2, health=4, position=0),
        Minion(name="Goldrinn", name_cn="戈德林", 
               tribe=Tribe.BEAST, tier=5, attack=8, health=8, position=1),
    ]
    
    state.player.board_minions = minions
    
    return state


if __name__ == "__main__":
    # 测试代码
    print("Testing BattlegroundsReader...")
    
    # 创建测试状态
    test_state = create_test_state()
    print(f"Test state: {test_state.phase.value}, Turn {test_state.turn}")
    print(f"Player health: {test_state.player.health}, Gold: {test_state.player.gold}")
    
    for minion in test_state.player.board_minions:
        print(f"  - {minion.name_cn} [{minion.tribe.value}] {minion.attack}/{minion.health}")
