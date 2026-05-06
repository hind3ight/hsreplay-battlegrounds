"""Unit tests for the recommendation engine."""

import json
import os
import sys
import tempfile
import unittest
from pathlib import Path

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent))

from engine.models import (
    CompCard,
    CompRecommendation,
    GameState,
    RecommendationRequest,
    RecommendationResult,
)
from engine.comp_matcher import CompMatcher
from engine.rule_engine import RuleEngine, create_default_rules_yaml


class TestCompMatcher(unittest.TestCase):
    """Tests for CompMatcher class."""

    def setUp(self):
        """Set up test fixtures."""
        self.matcher = CompMatcher()
        
        self.sample_comp = CompRecommendation(
            id=1,
            name="Test Comp",
            tier="S",
            difficulty="Medium",
            description="A test composition",
            core_cards=[
                CompCard(name="Card A", tier=2, attack=4, health=4, tribe="Dragon"),
                CompCard(name="Card B", tier=4, attack=6, health=6, tribe="Dragon"),
                CompCard(name="Card C", tier=6, attack=8, health=8, tribe="Dragon"),
            ],
        )

    def test_tier_weights(self):
        """Test that tier weights are applied correctly."""
        score = self.matcher.score_comp(
            comp=self.sample_comp,
            game_state=GameState(turn=8),
        )
        # Should have base tier score contribution
        self.assertGreater(score, 0)

    def test_board_overlap(self):
        """Test board overlap scoring."""
        game_state = GameState(
            turn=8,
            board_minions=[
                CompCard(name="Card A"),
                CompCard(name="Card B"),
            ],
        )
        
        score = self.matcher.score_comp(
            comp=self.sample_comp,
            game_state=game_state,
        )
        
        # Should score higher with 2 overlapping cards
        self.assertGreater(score, 0.3)

    def test_tribe_affinity(self):
        """Test tribe affinity scoring."""
        game_state = GameState(
            turn=8,
            current_tribe="Dragon",
            board_minions=[
                CompCard(name="Card A", tribe="Dragon"),
            ],
        )
        
        score = self.matcher.score_comp(
            comp=self.sample_comp,
            game_state=game_state,
        )
        
        # Should score well with matching tribe
        self.assertGreater(score, 0)

    def test_find_best_comp_for_board(self):
        """Test finding best comp given board minions."""
        comps = [
            self.sample_comp,
            CompRecommendation(
                id=2,
                name="Other Comp",
                tier="A",
                difficulty="Easy",
                description="Another comp",
                core_cards=[
                    CompCard(name="Card X"),
                    CompCard(name="Card Y"),
                ],
            ),
        ]
        
        board = [
            CompCard(name="Card A"),
            CompCard(name="Card B"),
        ]
        
        best = self.matcher.find_best_comp_for_board(comps, board)
        
        self.assertIsNotNone(best)
        self.assertEqual(best.id, 1)


class TestRuleEngine(unittest.TestCase):
    """Tests for RuleEngine class."""

    def setUp(self):
        """Set up test fixtures."""
        # Create a temporary rules file
        self.temp_rules = tempfile.NamedTemporaryFile(
            mode="w",
            suffix=".yaml",
            delete=False,
        )
        create_default_rules_yaml(self.temp_rules.name)
        self.temp_rules.close()
        
        self.engine = RuleEngine(self.temp_rules.name)

    def tearDown(self):
        """Clean up temp files."""
        os.unlink(self.temp_rules.name)

    def test_load_rules(self):
        """Test that rules are loaded from file."""
        self.assertIn("rules", self.engine.rules)
        self.assertIn("turn_warnings", self.engine.rules)

    def test_low_health_warning(self):
        """Test low health warning trigger."""
        game_state = GameState(turn=4, health=10)
        warnings = self.engine.evaluate_game_state(game_state)
        
        self.assertTrue(
            any("health" in w.lower() for w in warnings),
            f"Expected health warning, got: {warnings}"
        )

    def test_no_warning_for_healthy_early(self):
        """Test no warning when healthy early game."""
        game_state = GameState(turn=3, health=35)
        warnings = self.engine.evaluate_game_state(game_state)
        
        # Should not have low health warning
        health_warnings = [w for w in warnings if "health" in w.lower()]
        self.assertEqual(len(health_warnings), 0)

    def test_turn_warning(self):
        """Test turn-specific warnings."""
        game_state = GameState(turn=6)
        warnings = self.engine.evaluate_game_state(game_state)
        
        self.assertTrue(len(warnings) > 0)

    def test_reload_rules(self):
        """Test rules reload capability."""
        original_rules = self.engine.rules.copy()
        self.engine.reload_rules()
        
        self.assertEqual(self.engine.rules, original_rules)


class TestModels(unittest.TestCase):
    """Tests for data models."""

    def test_comp_card_creation(self):
        """Test CompCard creation with defaults."""
        card = CompCard(name="Test Card")
        
        self.assertEqual(card.name, "Test Card")
        self.assertEqual(card.tier, 0)
        self.assertEqual(card.attack, 0)
        self.assertEqual(card.health, 0)
        self.assertEqual(card.tags, [])

    def test_game_state_creation(self):
        """Test GameState creation with defaults."""
        state = GameState()
        
        self.assertEqual(state.turn, 1)
        self.assertEqual(state.health, 40)
        self.assertEqual(state.gold, 3)
        self.assertEqual(state.board_minions, [])
        self.assertIsNone(state.current_tribe)

    def test_recommendation_request(self):
        """Test RecommendationRequest with custom values."""
        state = GameState(turn=8, current_tribe="Dragon")
        request = RecommendationRequest(
            game_state=state,
            include_tier=["S", "A"],
            max_results=3,
        )
        
        self.assertEqual(request.game_state.turn, 8)
        self.assertEqual(request.include_tier, ["S", "A"])
        self.assertEqual(request.max_results, 3)


class TestRecommender(unittest.TestCase):
    """Tests for the main Recommender class."""

    def setUp(self):
        """Set up test fixtures with temp data files."""
        # Create temp comps file
        self.temp_comps = tempfile.NamedTemporaryFile(
            mode="w",
            suffix=".json",
            delete=False,
        )
        self.temp_comps.write(json.dumps({
            "comps": [
                {
                    "id": 1,
                    "name": "Dragons - Shiny Ring",
                    "tier": "S",
                    "difficulty": "Medium",
                    "description": "Dragon scaling",
                    "core_cards": [
                        {"name": "Dragon A", "tier": 2, "attack": 4, "health": 4},
                        {"name": "DK", "tier": 2, "attack": 7, "health": 5},
                    ],
                    "addon_cards": [
                        {"name": "4星龙"}
                    ],
                    "when_to_commit": "turn 6-8",
                    "key_transition": {
                        "1_star": ["Dragon A"],
                        "2_star": ["DK"],
                    },
                },
                {
                    "id": 2,
                    "name": "Murlocs - APM",
                    "tier": "A",
                    "difficulty": "Hard",
                    "description": "Murloc spam",
                    "core_cards": [
                        {"name": "Murloc A", "tier": 1},
                    ],
                },
            ]
        }))
        self.temp_comps.close()
        
        # Create temp rules file
        self.temp_rules = tempfile.NamedTemporaryFile(
            mode="w",
            suffix=".yaml",
            delete=False,
        )
        create_default_rules_yaml(self.temp_rules.name)
        self.temp_rules.close()
        
        from engine.recommender import Recommender
        self.recommender = Recommender(
            comps_path=self.temp_comps.name,
            rules_path=self.temp_rules.name,
        )

    def tearDown(self):
        """Clean up temp files."""
        os.unlink(self.temp_comps.name)
        os.unlink(self.temp_rules.name)

    def test_load_comps(self):
        """Test loading compositions from JSON."""
        comps = self.recommender.load_comps()
        
        self.assertEqual(len(comps), 2)
        self.assertEqual(comps[0].name, "Dragons - Shiny Ring")
        self.assertEqual(comps[0].tier, "S")
        self.assertEqual(len(comps[0].core_cards), 2)

    def test_recommend(self):
        """Test getting recommendations."""
        game_state = GameState(
            turn=6,
            current_tribe="Dragon",
            board_minions=[CompCard(name="DK")],
        )
        
        request = RecommendationRequest(
            game_state=game_state,
            include_tier=["S", "A"],
            max_results=5,
        )
        
        result = self.recommender.recommend(request)
        
        self.assertIsInstance(result, RecommendationResult)
        self.assertGreater(len(result.recommendations), 0)
        # Top comp should be Dragon since board has DK
        self.assertIn("Dragon", result.recommendations[0].name)

    def test_transition_priority(self):
        """Test transition priority generation."""
        game_state = GameState(turn=2)
        
        request = RecommendationRequest(
            game_state=game_state,
            include_tier=["S"],
            max_results=1,
        )
        
        result = self.recommender.recommend(request)
        
        # Should have transition priority based on turn
        self.assertIsNotNone(result.transition_priority)


if __name__ == "__main__":
    unittest.main()
