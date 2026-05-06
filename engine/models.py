"""Data models for the recommendation engine."""

import json
from dataclasses import dataclass, field
from typing import Optional


@dataclass
class CompCard:
    """A card/minion in a composition."""
    name: str
    tier: int = 0
    attack: int = 0
    health: int = 0
    name_cn: Optional[str] = None
    tribe: Optional[str] = None
    tags: list[str] = field(default_factory=list)


@dataclass
class CompRecommendation:
    """A complete comp recommendation from hsreplay data."""
    id: int
    name: str
    tier: str  # S, A, B, C
    difficulty: str  # Easy, Medium, Hard
    description: str
    core_cards: list[CompCard]
    addon_cards: list[CompCard] = field(default_factory=list)
    how_to_play: str = ""
    when_to_commit: str = ""
    key_transition: dict[str, list[str]] = field(default_factory=dict)
    score: float = 0.0  # Computed matching score


@dataclass
class GameState:
    """Current game state for recommendation context."""
    turn: int = 1
    hero: Optional[str] = None
    health: int = 40
    gold: int = 3
    board_minions: list[CompCard] = field(default_factory=list)
    shop_tier: int = 1
    current_tribe: Optional[str] = None  # Dominant tribe on board


@dataclass
class RecommendationRequest:
    """Request for recommendations."""
    game_state: GameState
    available_minions: list[CompCard] = field(default_factory=list)
    include_tier: list[str] = field(default_factory=lambda: ["S", "A", "B"])
    max_results: int = 5


@dataclass
class RecommendationResult:
    """Result of a recommendation request."""
    recommendations: list[CompRecommendation]
    current_comp_hint: Optional[str] = None
    transition_priority: list[str] = field(default_factory=list)
    warning_messages: list[str] = field(default_factory=list)


@dataclass
class CrossCompMinion:
    """Minion data with cross-comp statistics."""
    name: str
    name_cn: Optional[str] = None
    tribe: Optional[str] = None
    tier: Optional[int] = None
    comp_count: int = 0
    comps: list[str] = field(default_factory=list)
    tribes_seen: list[str] = field(default_factory=list)
    tribe_diversity: int = 0
    tags: list[str] = field(default_factory=list)


def load_cross_comp_minions(path: str) -> dict[str, CrossCompMinion]:
    """Load cross-comp minions data from JSON file."""
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    result = {}
    for name, minion_data in data.get("minions", {}).items():
        result[name] = CrossCompMinion(
            name=minion_data.get("name", name),
            name_cn=minion_data.get("name_cn"),
            tribe=minion_data.get("tribe"),
            tier=minion_data.get("tier"),
            comp_count=minion_data.get("comp_count", 0),
            comps=minion_data.get("comps", []),
            tribes_seen=minion_data.get("tribes_seen", []),
            tribe_diversity=minion_data.get("tribe_diversity", 0),
            tags=minion_data.get("tags", []),
        )
    return result
