#!/usr/bin/env python3
"""CLI entry point for HSReplay Battlegrounds Advisor."""

import argparse
import json
import sys
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from engine.models import CompCard, GameState, RecommendationRequest
from engine.recommender import Recommender


def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(
        description="HSReplay Battlegrounds Advisor - Get comp recommendations",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --turn 8 --tribe dragon
  %(prog)s --turn 6 --hero "Ragnaros" --health 25 --gold 10
  %(prog)s --turn 10 --board "Brann Bronzebeard,Twisted Wrathguard,Ashen Corruptor"
  %(prog)s --list-comps
        """,
    )

    parser.add_argument(
        "--turn", "-t",
        type=int,
        default=1,
        help="Current turn number (default: 1)",
    )
    parser.add_argument(
        "--hero", "-H",
        type=str,
        default=None,
        help="Hero name",
    )
    parser.add_argument(
        "--health", "-hp",
        type=int,
        default=40,
        help="Current health (default: 40)",
    )
    parser.add_argument(
        "--gold", "-g",
        type=int,
        default=3,
        help="Current gold (default: 3)",
    )
    parser.add_argument(
        "--tribe", "-T",
        type=str,
        default=None,
        help="Current dominant tribe (e.g., dragon, murloc, demon)",
    )
    parser.add_argument(
        "--board", "-b",
        type=str,
        default=None,
        help="Comma-separated list of minion names on your board",
    )
    parser.add_argument(
        "--tiers",
        type=str,
        default="S,A,B",
        help="Comma-separated list of tiers to include (default: S,A,B)",
    )
    parser.add_argument(
        "--max-results", "-n",
        type=int,
        default=5,
        help="Maximum number of recommendations (default: 5)",
    )
    parser.add_argument(
        "--list-comps",
        action="store_true",
        help="List all available compositions and exit",
    )
    parser.add_argument(
        "--json-output",
        action="store_true",
        help="Output in JSON format",
    )
    parser.add_argument(
        "--comps-path",
        type=str,
        default=None,
        help="Path to comps JSON file",
    )
    parser.add_argument(
        "--rules-path",
        type=str,
        default=None,
        help="Path to rules YAML file",
    )

    return parser.parse_args()


def parse_board(board_str: str) -> list[CompCard]:
    """Parse board string into list of CompCards."""
    if not board_str:
        return []
    
    minions = []
    for name in board_str.split(","):
        name = name.strip()
        if name:
            minions.append(CompCard(name=name))
    return minions


def list_comps(recommender: Recommender):
    """List all available compositions."""
    comps = recommender.load_comps()
    
    print(f"\nAvailable Compositions ({len(comps)} total):")
    print("=" * 60)
    
    for comp in comps:
        core_names = ", ".join(c.name for c in comp.core_cards[:3])
        if len(comp.core_cards) > 3:
            core_names += "..."
        
        print(f"\n[{comp.tier}] {comp.name}")
        print(f"  Difficulty: {comp.difficulty}")
        print(f"  Description: {comp.description}")
        print(f"  Core Cards: {core_names}")


def print_recommendations(result, json_output: bool = False):
    """Print recommendation results."""
    if json_output:
        # JSON output
        output = {
            "recommendations": [],
            "current_comp_hint": result.current_comp_hint,
            "transition_priority": result.transition_priority,
            "warnings": result.warning_messages,
        }
        
        for rec in result.recommendations:
            output["recommendations"].append({
                "id": rec.id,
                "name": rec.name,
                "tier": rec.tier,
                "difficulty": rec.difficulty,
                "description": rec.description,
                "how_to_play": rec.how_to_play,
                "when_to_commit": rec.when_to_commit,
                "score": round(rec.score, 3),
                "core_cards": [c.name for c in rec.core_cards],
                "addon_cards": [c.name for c in rec.addon_cards],
            })
        
        print(json.dumps(output, indent=2, ensure_ascii=False))
        return
    
    # Human-readable output
    print("\n" + "=" * 70)
    print("HSREPLAY BATTLEGROUNDS ADVISOR - RECOMMENDATIONS")
    print("=" * 70)
    
    # Warnings
    if result.warning_messages:
        print("\n[!] WARNINGS:")
        for warning in result.warning_messages:
            print(f"    - {warning}")
    
    # Current comp hint
    if result.current_comp_hint:
        print(f"\n[*] {result.current_comp_hint}")
    
    # Transition priority
    if result.transition_priority:
        print("\n[*] TRANSITION PRIORITY:")
        for i, card in enumerate(result.transition_priority, 1):
            print(f"    {i}. {card}")
    
    # Recommendations
    print("\n[*] TOP RECOMMENDATIONS:")
    print("-" * 70)
    
    for i, rec in enumerate(result.recommendations, 1):
        print(f"\n{i}. [{rec.tier}] {rec.name} (Score: {rec.score:.2f})")
        print(f"   Difficulty: {rec.difficulty}")
        print(f"   Description: {rec.description}")
        
        print(f"   Core Cards:")
        for card in rec.core_cards:
            tier_str = f"T{card.tier}" if card.tier > 0 else ""
            print(f"     - {card.name} {tier_str}")
        
        if rec.addon_cards:
            print(f"   Addon Cards:")
            for card in rec.addon_cards[:5]:
                print(f"     - {card.name}")
            if len(rec.addon_cards) > 5:
                print(f"     ... and {len(rec.addon_cards) - 5} more")
        
        if rec.when_to_commit:
            print(f"   When to Commit: {rec.when_to_commit}")
        
        if rec.how_to_play:
            how_to = rec.how_to_play[:150] + "..." if len(rec.how_to_play) > 150 else rec.how_to_play
            print(f"   How to Play: {how_to}")
    
    print("\n" + "=" * 70)


def main():
    """Main entry point."""
    args = parse_args()
    
    # Initialize recommender
    recommender = Recommender(
        comps_path=args.comps_path,
        rules_path=args.rules_path,
    )
    
    # List comps mode
    if args.list_comps:
        list_comps(recommender)
        return
    
    # Build game state
    board_minions = parse_board(args.board)
    
    game_state = GameState(
        turn=args.turn,
        hero=args.hero,
        health=args.health,
        gold=args.gold,
        board_minions=board_minions,
        current_tribe=args.tribe,
    )
    
    # Build request
    include_tiers = [t.strip().upper() for t in args.tiers.split(",")]
    
    request = RecommendationRequest(
        game_state=game_state,
        include_tier=include_tiers,
        max_results=args.max_results,
    )
    
    # Get recommendations
    try:
        result = recommender.recommend(request)
        print_recommendations(result, args.json_output)
    except FileNotFoundError as e:
        print(f"Error: Could not find data file - {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
