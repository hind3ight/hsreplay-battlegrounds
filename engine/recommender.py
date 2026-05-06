"""Core recommendation engine for battlegrounds comps."""

import json
from pathlib import Path
from typing import Optional

from engine.models import (
    CompCard,
    CompRecommendation,
    GameState,
    RecommendationRequest,
    RecommendationResult,
)
from engine.comp_matcher import CompMatcher
from engine.rule_engine import RuleEngine


class Recommender:
    """Main recommender that coordinates comp matching and rule evaluation."""

    def __init__(
        self,
        comps_path: Optional[str] = None,
        rules_path: Optional[str] = None,
    ):
        """
        Initialize recommender with comps and rules data.
        
        Args:
            comps_path: Path to season comps JSON. If None, uses default.
            rules_path: Path to rules YAML. If None, uses default.
        """
        self.comps_path = comps_path or self._default_comps_path()
        self.rules_path = rules_path or self._default_rules_path()
        
        self.comp_matcher = CompMatcher()
        self.rule_engine = RuleEngine(self.rules_path)
        self._comps_cache: Optional[list[CompRecommendation]] = None

    def _default_comps_path(self) -> str:
        """Get default path to comps data."""
        base = Path(__file__).parent.parent
        return str(base / "data" / "season13_comps.json")

    def _default_rules_path(self) -> str:
        """Get default path to rules file."""
        base = Path(__file__).parent.parent
        return str(base / "rules" / "default_rules.yaml")

    def load_comps(self, force_reload: bool = False) -> list[CompRecommendation]:
        """Load and parse comps from JSON file."""
        if self._comps_cache and not force_reload:
            return self._comps_cache

        comps = []
        with open(self.comps_path, "r", encoding="utf-8") as f:
            data = json.load(f)

        for comp_data in data.get("comps", []):
            core_cards = []
            for card_data in comp_data.get("core_cards", []):
                card = CompCard(
                    name=card_data["name"],
                    tier=card_data.get("tier", 0),
                    attack=card_data.get("attack", 0),
                    health=card_data.get("health", 0),
                )
                core_cards.append(card)

            addon_cards = []
            for card_data in comp_data.get("addon_cards", []):
                card = CompCard(name=card_data["name"])
                addon_cards.append(card)

            comp = CompRecommendation(
                id=comp_data["id"],
                name=comp_data["name"],
                tier=comp_data.get("tier", "B"),
                difficulty=comp_data.get("difficulty", "Medium"),
                description=comp_data.get("description", ""),
                core_cards=core_cards,
                addon_cards=addon_cards,
                how_to_play=comp_data.get("how_to_play", ""),
                when_to_commit=comp_data.get("when_to_commit", ""),
                key_transition=comp_data.get("key_transition", {}),
            )
            comps.append(comp)

        self._comps_cache = comps
        return comps

    def recommend(self, request: RecommendationRequest) -> RecommendationResult:
        """
        Generate recommendations based on game state and available minions.
        
        Args:
            request: The recommendation request with game state and constraints.
            
        Returns:
            RecommendationResult with ranked comps and transition guidance.
        """
        # Load comps if not cached
        all_comps = self.load_comps()

        # Filter by allowed tiers
        filtered_comps = [
            c for c in all_comps if c.tier in request.include_tier
        ]

        # Score each comp against current game state
        scored_comps = []
        for comp in filtered_comps:
            score = self.comp_matcher.score_comp(
                comp=comp,
                game_state=request.game_state,
                available_minions=request.available_minions,
            )
            comp.score = score
            scored_comps.append(comp)

        # Sort by score descending
        scored_comps.sort(key=lambda c: c.score, reverse=True)

        # Get top N results
        top_comps = scored_comps[: request.max_results]

        # Build transition priority based on current turn
        transition_priority = self._get_transition_priority(
            top_comps, request.game_state.turn
        )

        # Get rule-based warnings
        warnings = self.rule_engine.evaluate_game_state(
            game_state=request.game_state,
            recommended_comp=top_comps[0] if top_comps else None,
        )

        return RecommendationResult(
            recommendations=top_comps,
            current_comp_hint=self._guess_current_comp_hint(top_comps, request.game_state),
            transition_priority=transition_priority,
            warning_messages=warnings,
        )

    def _get_transition_priority(
        self, comps: list[CompRecommendation], turn: int
    ) -> list[str]:
        """Get ordered list of key transitions to prioritize."""
        if not comps:
            return []

        # Find turn-appropriate transitions from top comp
        turn_key = self._turn_to_key(turn)
        top_comp = comps[0]

        transitions = []
        if turn_key in top_comp.key_transition:
            transitions.extend(top_comp.key_transition[turn_key])

        # Add next turn's transitions as "work toward"
        next_key = self._turn_to_key(turn + 1)
        if next_key in top_comp.key_transition:
            for card_name in top_comp.key_transition[next_key]:
                if card_name not in transitions:
                    transitions.append(f"[next] {card_name}")

        return transitions

    def _turn_to_key(self, turn: int) -> str:
        """Convert turn number to transition key."""
        if turn <= 3:
            return "1_star"
        elif turn <= 6:
            return "2_star"
        elif turn <= 9:
            return "3_star"
        elif turn <= 12:
            return "4_star"
        else:
            return "5_star"

    def _guess_current_comp_hint(
        self,
        comps: list[CompRecommendation],
        game_state: GameState,
    ) -> Optional[str]:
        """Guess the current comp based on board minions."""
        if not comps or not game_state.board_minions:
            return None

        # Simple heuristic: check which comp's cards are on board
        board_names = {m.name for m in game_state.board_minions}
        
        for comp in comps:
            core_names = {c.name for c in comp.core_cards}
            overlap = board_names & core_names
            if len(overlap) >= 2:
                return f"You appear to be building {comp.name} ({len(overlap)} core cards found)"

        return None
