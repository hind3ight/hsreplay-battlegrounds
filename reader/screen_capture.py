"""
屏幕捕获模块 - 使用 MSS 库捕获炉石传说窗口
"""
import mss
import numpy as np
from dataclasses import dataclass
from typing import Optional, Tuple, List


@dataclass
class WindowRegion:
    """窗口区域定义"""
    left: int
    top: int
    width: int
    height: int


class HearthstoneCapture:
    """炉石传说窗口捕获类"""
    
    # 炉石传说窗口默认区域 (可根据实际情况调整)
    BOARD_REGION = WindowRegion(left=0, top=0, width=1920, height=1080)
    
    # 随从栏区域 (酒馆战棋模式下)
    MINION_BAR_REGION = WindowRegion(left=200, top=800, width=1520, height=200)
    
    def __init__(self, monitor_index: int = 1):
        """
        初始化屏幕捕获
        
        Args:
            monitor_index: 监视器索引, 默认1 (主显示器)
        """
        self.monitor_index = monitor_index
        self.sct = mss.mss()
    
    def get_monitor_info(self) -> dict:
        """获取监视器信息"""
        return self.sct.monitors[self.monitor_index]
    
    def capture_region(self, region: WindowRegion) -> np.ndarray:
        """
        捕获指定区域
        
        Args:
            region: 窗口区域
            
        Returns:
            numpy数组格式的图像 (BGRA格式)
        """
        monitor = {
            "left": region.left,
            "top": region.top,
            "width": region.width,
            "height": region.height
        }
        screenshot = self.sct.grab(monitor)
        return np.array(screenshot)
    
    def capture_full_screen(self) -> np.ndarray:
        """捕获整个屏幕"""
        return self.capture_region(WindowRegion(
            left=0, top=0, 
            width=self.get_monitor_info()["width"],
            height=self.get_monitor_info()["height"]
        ))
    
    def capture_board(self) -> np.ndarray:
        """捕获战棋棋盘区域"""
        return self.capture_region(self.BOARD_REGION)
    
    def capture_minion_bar(self) -> np.ndarray:
        """捕获随从栏区域"""
        return self.capture_region(self.MINION_BAR_REGION)
    
    def set_capture_region(self, left: int, top: int, width: int, height: int):
        """设置捕获区域"""
        self.BOARD_REGION = WindowRegion(left, top, width, height)
    
    def set_minion_bar_region(self, left: int, top: int, width: int, height: int):
        """设置随从栏区域"""
        self.MINION_BAR_REGION = WindowRegion(left, top, width, height)
    
    def close(self):
        """关闭 MSS 实例"""
        self.sct.close()
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


def find_hearthstone_window() -> Optional[WindowRegion]:
    """
    查找炉石传说窗口 (需要 win32gui, 仅 Windows)
    
    Returns:
        窗口区域, 未找到返回 None
    """
    try:
        import win32gui
        import win32con
        
        def callback(hwnd, windows):
            if win32gui.IsWindowVisible(hwnd):
                title = win32gui.GetWindowText(hwnd)
                if "Hearthstone" in title:
                    rect = win32gui.GetWindowRect(hwnd)
                    windows.append(WindowRegion(
                        left=rect[0],
                        top=rect[1],
                        width=rect[2] - rect[0],
                        height=rect[3] - rect[1]
                    ))
            return True
        
        windows = []
        win32gui.EnumWindows(callback, windows)
        
        if windows:
            return windows[0]
    except ImportError:
        pass
    
    return None


if __name__ == "__main__":
    # 测试代码
    with HearthstoneCapture() as capture:
        print(f"Monitor info: {capture.get_monitor_info()}")
        img = capture.capture_minion_bar()
        print(f"Captured image shape: {img.shape}")
