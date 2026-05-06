"""YAML-based rule engine for game state evaluation and warnings."""

import re
from pathlib import Path
from typing import Optional

import yaml

from engine.models import CompRecommendation, GameState


class RuleEngine:
    """Evaluates game state against YAML-defined rules."""

    def __init__(self, rules_path: Optional[str] = None):
        """
        Initialize rule engine with rules file.
        
        Args:
            rules_path: Path to YAML rules file.
        """
        self.rules_path = rules_path or self._default_rules_path()
        self.rules = self._load_rules()

    def _default_rules_path(self) -> str:
        """Get default rules path."""
        base = Path(__file__).parent.parent
        return str(base / "rules" / "default_rules.yaml")

    def _load_rules(self) -> dict:
        """Load rules from YAML file."""
        try:
            with open(self.rules_path, "r", encoding="utf-8") as f:
                return yaml.safe_load(f) or {}
        except FileNotFoundError:
            return {"rules": [], "warnings": []}

    def reload_rules(self) -> None:
        """Reload rules from disk."""
        self.rules = self._load_rules()

    def evaluate_game_state(
        self,
        game_state: GameState,
        recommended_comp: Optional[CompRecommendation] = None,
    ) -> list[str]:
        """
        Evaluate game state and return warning messages.
        
        Args:
            game_state: Current game state.
            recommended_comp: The top recommended comp.
            
        Returns:
            List of warning message strings.
        """
        warnings = []

        # Evaluate general rules
        for rule in self.rules.get("rules", []):
            warning = self._evaluate_rule(rule, game_state, recommended_comp)
            if warning:
                warnings.append(warning)

        # Evaluate tier-based warnings
        turn_warnings = self._get_turn_warnings(game_state)
        warnings.extend(turn_warnings)

        # Evaluate comp-specific warnings
        if recommended_comp:
            comp_warnings = self._get_comp_warnings(game_state, recommended_comp)
            warnings.extend(comp_warnings)

        return warnings

    def _evaluate_rule(
        self,
        rule: dict,
        game_state: GameState,
        recommended_comp: Optional[CompRecommendation],
    ) -> Optional[str]:
        """Evaluate a single rule against game state."""
        condition = rule.get("condition", {})
        message = rule.get("message", "")

        # Check turn condition
        if "turn" in condition:
            turn_range = condition["turn"]
            if isinstance(turn_range, int):
                if game_state.turn != turn_range:
                    return None
            elif isinstance(turn_range, dict):
                min_turn = turn_range.get("min", 1)
                max_turn = turn_range.get("max", 999)
                if not (min_turn <= game_state.turn <= max_turn):
                    return None

        # Check health condition
        if "health_below" in condition:
            if game_state.health >= condition["health_below"]:
                return None

        # Check gold condition
        if "gold_below" in condition:
            if game_state.gold >= condition["gold_below"]:
                return None

        # Check tribe condition
        if "tribe" in condition:
            if game_state.current_tribe != condition["tribe"]:
                return None

        # Check board_has condition
        if "board_has" in condition:
            board_names = {m.name for m in game_state.board_minions}
            if condition["board_has"] not in board_names:
                return None

        return message

    def _get_turn_warnings(self, game_state: GameState) -> list[str]:
        """Get turn-specific warnings."""
        warnings = []
        turn_config = self.rules.get("turn_warnings", {})

        for turn_key, warning_list in turn_config.items():
            if turn_key == str(game_state.turn):
                warnings.extend(warning_list)
            elif "-" in str(turn_key):
                # Handle ranges like "8-10"
                match = re.match(r"(\d+)-(\d+)", str(turn_key))
                if match:
                    min_t, max_t = int(match.group(1)), int(match.group(2))
                    if min_t <= game_state.turn <= max_t:
                        warnings.extend(warning_list)

        return warnings

    def _get_comp_warnings(
        self,
        game_state: GameState,
        comp: CompRecommendation,
    ) -> list[str]:
        """Get comp-specific warnings based on tags or type."""
        warnings = []
        comp_rules = self.rules.get("comp_rules", {})

        # Check for high-tier comps
        if comp.tier == "S":
            s_tier_warnings = comp_rules.get("s_tier_warnings", [])
            warnings.extend(s_tier_warnings)

        # Check for specific comp patterns
        comp_name_lower = comp.name.lower()
        for pattern, warning_list in comp_rules.get("by_name_pattern", {}).items():
            if pattern.lower() in comp_name_lower:
                warnings.extend(warning_list)

        return warnings


def create_default_rules_yaml(path: str) -> None:
    """
    Create a default rules YAML file.
    
    Args:
        path: Where to write the file.
    """
    default_rules = {
        "rules": [
            {
                "name": "low_health_early",
                "condition": {"turn": {"min": 1, "max": 6}, "health_below": 15},
                "message": "Low health early game - consider aggressive rerolls or pivoting",
            },
            {
                "name": "no_tribe_by_turn_6",
                "condition": {"turn": 6, "tribe": None},
                "message": "No clear tribe by turn 6 - consider committing to a comp soon",
            },
            {
                "name": "poor_economy_early",
                "condition": {"turn": {"min": 3, "max": 5}, "gold_below": 7},
                "message": "Low gold on turns 3-5 - avoid excessive rerolls",
            },
        ],
        "turn_warnings": {
            "6": ["Turn 6: Consider if you should level or stay at current tier"],
            "8": ["Turn 8: High value units available - evaluate your board strength"],
            "10": ["Turn 10: Final push - ensure your key units are tripled"],
            "8-10": ["Mid-game transition zone - pivot now if needed"],
        },
        "comp_rules": {
            "s_tier_warnings": [
                "S-tier comps are harder to execute - ensure you have key pieces before committing",
            ],
            "by_name_pattern": {
                "murloc": ["Murlocs require multiple triples for full power"],
                "demon": ["Demons need proper economy setup - ensure you have Brann"],
                "dragon": ["Dragons scale with gold - prioritize Shiny Ring targets"],
            },
        },
    }

    with open(path, "w", encoding="utf-8") as f:
        yaml.dump(default_rules, f, default_flow_style=False, sort_keys=False)
