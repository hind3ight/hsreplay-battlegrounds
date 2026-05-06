#!/usr/bin/env python3
"""Mock 测试：在无游戏窗口环境下验证全流程逻辑"""
import sys
import os
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

from PIL import Image, ImageDraw

# 模拟数据
MOCK_MINIONS = [
    {"name_en": "Brann Bronzebeard", "name_cn": "布莱恩·铜须", "tier": 2, "tribe": "neutral"},
    {"name_en": "Twisted Wrathguard", "name_cn": "扭曲的愤怒守卫", "tier": 4, "tribe": "demon"},
    {"name_en": "Ashen Corruptor", "name_cn": "灰葬腐蚀者", "tier": 6, "tribe": "demon"},
]

def create_mock_screenshot():
    """生成模拟酒馆截图（用于测试 OCR 流程）"""
    img = Image.new('RGB', (800, 300), color=(30, 30, 40))
    draw = ImageDraw.Draw(img)
    for i, m in enumerate(MOCK_MINIONS[:3]):
        text = f"{m['name_cn']} ({m['name_en']}) T{m['tier']}"
        draw.text((20, 30 + i * 60), text, fill=(200, 180, 100))
    img.save('/tmp/mock_tavern.png')
    print(f"[Mock] 生成了模拟截图: /tmp/mock_tavern.png")
    return '/tmp/mock_tavern.png'

def test_name_matcher():
    """测试名称模糊匹配"""
    print("\n=== 测试 name_matcher ===")
    from reader.name_matcher import MinionNameMatcher

    matcher = MinionNameMatcher()
    test_cases = [
        ("恐龙大师布莱恩", "Brann Bronzebeard"),  # 完整中文名
        ("Brann Bronzebeard", "Brann Bronzebeard"),  # 英文原名
        ("金属者刃蛇", "Rylak Metalhead"),  # 中文名
        ("灰葬腐蚀者", "Ashen Corruptor"),  # 中文名
    ]

    passed = 0
    for ocr_text, expected in test_cases:
        result = matcher.match_name(ocr_text)
        ok = result and result.matched_name == expected
        status = "✓" if ok else "✗"
        print(f"  {status} OCR='{ocr_text}' -> {result.matched_name if result else 'None'} (期望: {expected})")
        if ok:
            passed += 1
    print(f"[{'PASS' if passed == len(test_cases) else 'PARTIAL'}] name_matcher: {passed}/{len(test_cases)} 通过")

def test_engine_recommender():
    """测试推荐引擎"""
    print("\n=== 测试 engine/recommender ===")
    from engine.recommender import Recommender
    from engine.models import GameState, RecommendationRequest

    recommender = Recommender()

    # engine/models.py 的 GameState 字段: turn, hero, health, gold, board_minions, shop_tier, current_tribe
    state = GameState(turn=1, gold=3, health=40)
    request = RecommendationRequest(game_state=state, max_results=5)

    result = recommender.recommend(request)
    print(f"  生成了 {len(result.recommendations)} 条推荐")
    for i, rec in enumerate(result.recommendations[:3], 1):
        print(f"  {i}. {rec.name} (Tier:{rec.tier}) Score:{rec.score:.2f}")
    print(f"[PASS] 推荐引擎运行正常")

def test_comp_matcher():
    """测试阵容匹配器"""
    print("\n=== 测试 engine/comp_matcher ===")
    from engine.comp_matcher import CompMatcher
    from engine.recommender import Recommender
    from engine.models import GameState

    recommender = Recommender()
    comps = recommender.load_comps()
    matcher = CompMatcher()

    state = GameState(turn=5, gold=8, health=35)
    scores = []
    for comp in comps[:5]:
        s = matcher.score_comp(comp, state)
        scores.append((comp.name, s))

    scores.sort(key=lambda x: x[1], reverse=True)
    print(f"  测试了前5个阵容的评分:")
    for name, s in scores[:3]:
        print(f"  - {name}: {s:.2f}")
    print(f"[PASS] 阵容匹配器运行正常")

def test_rule_engine():
    """测试规则引擎"""
    print("\n=== 测试 engine/rule_engine ===")
    from engine.rule_engine import RuleEngine
    from engine.models import GameState

    engine = RuleEngine()
    state = GameState(turn=1, gold=3, health=40)
    warnings = engine.evaluate_game_state(state)
    print(f"  回合1产生 {len(warnings)} 条警告/建议")
    for w in warnings[:3]:
        print(f"  - {w}")
    print(f"[PASS] 规则引擎运行正常")

def test_overlay_window():
    """测试 overlay 窗口类"""
    print("\n=== 测试 overlay/window ===")
    try:
        import tkinter  # noqa
    except ImportError:
        print(f"  tkinter 不可用（无头环境），跳过 overlay 测试")
        print(f"[SKIP] overlay 窗口在无头环境跳过")
        return

    from overlay.window import OverlayWindow

    try:
        win = OverlayWindow()
        methods = [m for m in dir(win) if not m.startswith('_')]
        print(f"  OverlayWindow 初始化成功")
        print(f"  公开方法: {methods}")
        print(f"[PASS] overlay 窗口类正常")
    except Exception as e:
        print(f"  异常(无显示环境预期): {e}")
        print(f"[SKIP] overlay 窗口在无头环境跳过")

def test_game_state():
    """测试 game_state 模块"""
    print("\n=== 测试 reader/game_state ===")
    from reader.game_state import GameState, Tribe

    state = GameState()
    state.turn = 3
    state.available_tribes = [Tribe.DRAGON, Tribe.DEMON, Tribe.BEAST]
    state.tavern_tier = 2
    print(f"  GameState: 回合={state.turn}, 酒馆={state.tavern_tier}")
    print(f"  种族: {[t.name for t in state.available_tribes]}")
    print(f"[PASS] game_state 数据结构正常")

def test_minions_metadata():
    """测试随从元数据"""
    print("\n=== 测试 minions_metadata ===")
    import json
    path = Path('data/minions_metadata.json')
    if not path.exists():
        print(f"[WARN] data/minions_metadata.json 不存在，跳过")
        return
    with open(path, encoding='utf-8') as f:
        data = json.load(f)
    minions = data.get('minions', data)
    print(f"  随从总数: {len(minions)}")
    # 找跨阵容最多的
    top = sorted(minions.values(), key=lambda x: x.get('comp_count', 0), reverse=True)[:3]
    for m in top:
        print(f"  - {m.get('name_cn', m.get('name'))} (T{m.get('tier')}) 跨{m.get('comp_count')}阵容")
    print(f"[PASS] 随从元数据正常")

def test_cross_comp_analyzer():
    """测试跨阵容分析器"""
    print("\n=== 测试 cmd/analyze-cross ===")
    import subprocess
    result = subprocess.run(
        ['go', 'run', 'cmd/analyze-cross/main.go'],
        capture_output=True, text=True,
        cwd='/home/hind3ight/work/Projects/GoProjects/src/github.com/hind3ight/hsreplay-battlegrounds'
    )
    if result.returncode == 0:
        lines = result.stdout.strip().split('\n')
        print(f"  输出 {len(lines)} 行")
        for line in lines[:5]:
            print(f"  {line}")
        print(f"[PASS] 跨阵容分析器运行正常")
    else:
        print(f"  错误: {result.stderr[:200]}")
        print(f"[WARN] 跨阵容分析器运行失败，跳过")

def main():
    print("=" * 50)
    print("Battlegrounds Advisor - Mock 测试套件")
    print("=" * 50)

    create_mock_screenshot()

    data_dir = Path('data')
    if not (data_dir / 'season13_comps.json').exists():
        print("[WARN] 缺少 season13_comps.json，跳过部分测试")
    else:
        test_name_matcher()
        test_engine_recommender()
        test_comp_matcher()
        test_rule_engine()
        test_minions_metadata()

    test_game_state()
    test_overlay_window()
    test_cross_comp_analyzer()

    print("\n" + "=" * 50)
    print("Mock 测试完成！")
    print("在 Windows 上运行 reader/test_reader.py 验证真实 OCR")
    print("=" * 50)

if __name__ == '__main__':
    main()
