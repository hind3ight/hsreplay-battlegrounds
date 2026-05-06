#!/usr/bin/env python3
"""
reader 模块测试脚本

测试屏幕捕获、OCR、种族识别和名称匹配功能
"""
import sys
import time
import numpy as np

# 测试各个模块
def test_screen_capture():
    """测试屏幕捕获"""
    print("\n=== Testing Screen Capture ===")
    try:
        from screen_capture import HearthstoneCapture, WindowRegion
        
        with HearthstoneCapture() as capture:
            print(f"Monitor info: {capture.get_monitor_info()}")
            
            # 尝试捕获随从栏区域
            img = capture.capture_minion_bar()
            print(f"Captured image shape: {img.shape}")
            
            # 测试设置自定义区域
            capture.set_minion_bar_region(100, 700, 1720, 280)
            img2 = capture.capture_minion_bar()
            print(f"Custom region shape: {img2.shape}")
        
        print("Screen capture test PASSED")
        return True
    except Exception as e:
        print(f"Screen capture test FAILED: {e}")
        return False


def test_ocr():
    """测试 OCR 识别"""
    print("\n=== Testing OCR ===")
    try:
        from ocr_minions import OCRMinions
        
        with OCRMinions() as ocr:
            # 测试图像预处理
            test_img = np.zeros((200, 200, 4), dtype=np.uint8)
            processed = ocr.preprocess_image(test_img)
            print(f"Preprocessed image shape: {processed.shape}")
            
            # 注意: 实际 OCR 需要真实游戏截图
            print("OCR module loaded successfully")
        
        print("OCR test PASSED")
        return True
    except Exception as e:
        print(f"OCR test FAILED: {e}")
        return False


def test_tribe_detector():
    """测试种族检测"""
    print("\n=== Testing Tribe Detector ===")
    try:
        from tribe_detector import TribeDetector, Tribe
        
        detector = TribeDetector()
        print(f"Loaded {len(detector)} tribe icons")
        
        # 测试颜色检测
        test_img = np.zeros((100, 100, 3), dtype=np.uint8)
        test_img[:, :] = [0, 0, 255]  # 蓝色
        
        tribe = detector.detect_tribe_by_color(test_img)
        print(f"Color-based detection (blue): {tribe}")
        
        # 测试模板匹配 (无图标时会返回空)
        matches = detector.match_tribe_template_matching(test_img, threshold=0.5)
        print(f"Template matching results: {len(matches)} matches")
        
        print("Tribe detector test PASSED")
        return True
    except Exception as e:
        print(f"Tribe detector test FAILED: {e}")
        return False


def test_name_matcher():
    """测试名称匹配"""
    print("\n=== Testing Name Matcher ===")
    try:
        from name_matcher import MinionNameMatcher
        
        matcher = MinionNameMatcher(
            metadata_path="../data/minions_metadata.json",
            names_path="../data/minion_names.json"
        )
        
        # 测试匹配
        test_names = [
            "Brann Bronzebeard",
            "Goldrinn, the Great Wolf", 
            "Air Revenant",
            "Some Random Text"
        ]
        
        print(f"Loaded {len(matcher.all_names)} minion names")
        
        for name in test_names:
            result = matcher.match_name(name)
            if result:
                print(f"  '{name}' -> {result.matched_name} [{result.tribe}] score={result.score:.1f}")
            else:
                print(f"  '{name}' -> No match")
        
        print("Name matcher test PASSED")
        return True
    except Exception as e:
        print(f"Name matcher test FAILED: {e}")
        import traceback
        traceback.print_exc()
        return False


def test_game_state():
    """测试游戏状态整合"""
    print("\n=== Testing Game State ===")
    try:
        from game_state import (
            BattlegroundsReader, 
            GameState, 
            PlayerState, 
            Minion, 
            GamePhase
        )
        from tribe_detector import Tribe
        
        # 测试创建游戏状态
        state = GameState()
        state.phase = GamePhase.SHOP
        state.turn = 5
        state.player.health = 32
        state.player.gold = 6
        
        # 添加测试随从
        minion = Minion(
            name="Brann Bronzebeard",
            name_cn="恐龙大师布莱恩",
            tribe=Tribe.BEAST,
            tier=2,
            attack=2,
            health=4,
            position=0
        )
        state.player.board_minions.append(minion)
        
        # 测试序列化
        state_dict = state.to_dict()
        print(f"Game state phase: {state_dict['phase']}")
        print(f"Player minions: {len(state_dict['player']['board_minions'])}")
        
        # 测试创建测试状态
        test_state = GameState()
        print(f"Empty state phase: {test_state.phase.value}")
        
        print("Game state test PASSED")
        return True
    except Exception as e:
        print(f"Game state test FAILED: {e}")
        import traceback
        traceback.print_exc()
        return False


def test_integration():
    """集成测试"""
    print("\n=== Testing Integration ===")
    try:
        from game_state import BattlegroundsReader
        
        reader = BattlegroundsReader(
            assets_dir="../assets",
            data_dir="../data"
        )
        
        print(f"Reader initialized with {len(reader.name_matcher.all_names)} minion names")
        
        # 创建测试状态
        state = reader.game_state
        print(f"Initial state phase: {state.phase.value}")
        
        reader.close()
        
        print("Integration test PASSED")
        return True
    except Exception as e:
        print(f"Integration test FAILED: {e}")
        import traceback
        traceback.print_exc()
        return False


def main():
    """运行所有测试"""
    print("=" * 50)
    print("Battlegrounds Reader Module Tests")
    print("=" * 50)
    
    results = []
    
    # 运行各项测试
    results.append(("Screen Capture", test_screen_capture()))
    results.append(("OCR", test_ocr()))
    results.append(("Tribe Detector", test_tribe_detector()))
    results.append(("Name Matcher", test_name_matcher()))
    results.append(("Game State", test_game_state()))
    results.append(("Integration", test_integration()))
    
    # 汇总结果
    print("\n" + "=" * 50)
    print("Test Results Summary")
    print("=" * 50)
    
    passed = 0
    failed = 0
    
    for name, result in results:
        status = "PASSED" if result else "FAILED"
        print(f"  {name}: {status}")
        if result:
            passed += 1
        else:
            failed += 1
    
    print(f"\nTotal: {passed} passed, {failed} failed")
    
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
