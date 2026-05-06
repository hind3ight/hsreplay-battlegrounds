"""Composition matcher - scores how well a comp matches the current game state."""

from typing import Optional

from engine.models import (
    CompCard,
    CompRecommendation,
    GameState,
)


class CompMatcher:
    """Matches and scores compositions against current game state."""

    # Tier weights for scoring
    TIER_WEIGHTS = {"S": 1.0, "A": 0.8, "B": 0.6, "C": 0.4}

    # Tribe match bonus
    TRIBE_MATCH_BONUS = 0.15

    # Transition timing weight
    TRANSITION_BONUS = 0.1

    def score_comp(
        self,
        comp: CompRecommendation,
        game_state: GameState,
        available_minions: list[CompCard] = None,
    ) -> float:
        """
        Calculate a match score for a comp given current game state.
        
        Score is 0.0 to 1.0, higher is better.
        
        Args:
            comp: The composition to score.
            game_state: Current game state.
            available_minions: Minions currently available in shop.
            
        Returns:
            Match score between 0.0 and 1.0.
        """
        available_minions = available_minions or []
        score = 0.0

        # Base score from tier
        score += self.TIER_WEIGHTS.get(comp.tier, 0.5) * 0.3

        # Tribe affinity bonus
        tribe_score = self._score_tribe_affinity(comp, game_state)
        score += tribe_score * 0.25

        # Board overlap score - how many core cards we already have
        board_overlap = self._score_board_overlap(comp, game_state)
        score += board_overlap * 0.25

        # Turn appropriateness - are we at the right turn for this comp?
        turn_score = self._score_turn_appropriateness(comp, game_state)
        score += turn_score * 0.15

        # Availability of key cards in shop
        availability_score = self._score_availability(comp, available_minions)
        score += availability_score * 0.05

        return min(score, 1.0)

    def _score_tribe_affinity(
        self,
        comp: CompRecommendation,
        game_state: GameState,
    ) -> float:
        """Score based on tribe match between board and comp."""
        if not game_state.current_tribe:
            return 0.5  # Neutral if no tribe identified

        # Count cards in comp matching current tribe
        tribe_card_count = sum(
            1 for card in comp.core_cards if card.tribe == game_state.current_tribe
        )
        tribe_ratio = tribe_card_count / max(len(comp.core_cards), 1)

        return tribe_ratio

    def _score_board_overlap(
        self,
        comp: CompRecommendation,
        game_state: GameState,
    ) -> float:
        """Score based on how many core cards we already have on board."""
        if not game_state.board_minions:
            return 0.3  # Neutral - no board to compare

        board_names = {m.name for m in game_state.board_minions}
        core_names = {c.name for c in comp.core_cards}

        overlap = board_names & core_names
        overlap_ratio = len(overlap) / max(len(core_names), 1)

        # Bonus for having at least 2 core cards
        if len(overlap) >= 2:
            overlap_ratio += 0.1

        return min(overlap_ratio, 1.0)

    def _score_turn_appropriateness(
        self,
        comp: CompRecommendation,
        game_state: GameState,
    ) -> float:
        """
        Score based on whether current turn matches when to commit.
        
        Early turns = prefer easy/fast comps
        Late turns = prefer scaling comps
        """
        turn = game_state.turn
        when_to_commit = comp.when_to_commit.lower()

        # Very early game - be more flexible
        if turn <= 4:
            return 0.6

        # Check for timing keywords in when_to_commit
        if "early" in when_to_commit or "turn 1" in when_to_commit or "turn 2" in when_to_commit:
            if turn <= 6:
                return 0.9
            else:
                return 0.4

        if "mid" in when_to_commit or "turn 8" in when_to_commit or "turn 10" in when_to_commit:
            if 6 <= turn <= 12:
                return 0.8
            else:
                return 0.5

        if "late" in when_to_commit or "end" in when_to_commit:
            if turn >= 10:
                return 0.9
            else:
                return 0.4

        # Default - mid-game preference
        if 6 <= turn <= 10:
            return 0.7
        return 0.5

    def _score_availability(
        self,
        comp: CompRecommendation,
        available_minions: list[CompCard],
    ) -> float:
        """Score based on availability of key cards in shop."""
        if not available_minions:
            return 0.5  # Neutral if no shop data

        available_names = {m.name for m in available_minions}

        # Check how many core cards are available
        core_names = {c.name for c in comp.core_cards}
        available_core = available_names & core_names

        if not core_names:
            return 0.5

        return len(available_core) / len(core_names)

    def find_best_comp_for_board(
        self,
        comps: list[CompRecommendation],
        board_minions: list[CompCard],
    ) -> Optional[CompRecommendation]:
        """
        Find the best matching comp given a set of board minions.
        
        Args:
            comps: List of all available comps.
            board_minions: Current minions on board.
            
        Returns:
            Best matching comp or None.
        """
        if not board_minions or not comps:
            return None

        board_names = {m.name for m in board_minions}
        best_comp = None
        best_score = 0

        for comp in comps:
            core_names = {c.name for c in comp.core_cards}
            overlap = board_names & core_names

            # Require at least 2 overlapping cards
            if len(overlap) >= 2:
                score = len(overlap) / len(core_names)
                if score > best_score:
                    best_score = score
                    best_comp = comp

        return best_comp
