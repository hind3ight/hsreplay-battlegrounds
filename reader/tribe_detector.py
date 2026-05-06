"""
种族检测模块 - 使用 OpenCV 图标匹配识别随从种族
"""
import numpy as np
import cv2
from dataclasses import dataclass
from typing import List, Optional, Dict
from enum import Enum


class Tribe(Enum):
    """随从种族枚举"""
    BEAST = "Beast"
    DRAGON = "Dragon"
    ELEMENTAL = "Elemental"
    MECH = "Mech"
    MURLOC = "Murloc"
    DEMON = "Demon"
    QUILBOAR = "Quilboar"
    PIRATE = "Pirate"
    UNDEAD = "Undead"
    Naga = "Naga"
    TREANT = "Treant"
    PANDAS = "Pand"
    GOBLIN = "Goblin"
    TOTEM = "Totem"  # 图腾
    UNKNOWN = "Unknown"


# 种族图标路径 (相对于 assets/tribe_icons/)
TRIBE_ICON_NAMES = {
    Tribe.BEAST: "beast.png",
    Tribe.DRAGON: "dragon.png",
    Tribe.ELEMENTAL: "elemental.png",
    Tribe.MECH: "mech.png",
    Tribe.MURLOC: "murloc.png",
    Tribe.DEMON: "demon.png",
    Tribe.QUILBOAR: "quilboar.png",
    Tribe.PIRATE: "pirate.png",
    Tribe.UNDEAD: "undead.png",
    Tribe.Naga: "naga.png",
    Tribe.TREANT: "treant.png",
    Tribe.PANDAS: "pandas.png",
    Tribe.GOBLIN: "goblin.png",
    Tribe.TOTEM: "totem.png",
}


@dataclass
class TribeMatch:
    """种族匹配结果"""
    tribe: Tribe
    confidence: float
    location: tuple  # (x, y)


class TribeDetector:
    """种族检测器类"""
    
    def __init__(self, icons_dir: str = "assets/tribe_icons"):
        """
        初始化种族检测器
        
        Args:
            icons_dir: 种族图标目录
        """
        self.icons_dir = icons_dir
        self.tribe_templates: Dict[Tribe, np.ndarray] = {}
        self.tribe_features: Dict[Tribe, dict] = {}  # 预计算的特征
        
        # 加载所有种族图标
        self._load_tribe_icons()
    
    def _load_tribe_icons(self):
        """加载所有种族图标"""
        for tribe, icon_name in TRIBE_ICON_NAMES.items():
            icon_path = f"{self.icons_dir}/{icon_name}"
            try:
                template = cv2.imread(icon_path, cv2.IMREAD_COLOR)
                if template is not None:
                    self.tribe_templates[tribe] = template
            except Exception:
                pass
    
    def preprocess_image(self, image: np.ndarray) -> np.ndarray:
        """
        预处理图像
        
        Args:
            image: 输入图像
            
        Returns:
            预处理后的图像
        """
        # 转换为灰度图
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        # 图像增强
        clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
        enhanced = clahe.apply(gray)
        
        return enhanced
    
    def match_tribe_template_matching(self, image: np.ndarray, threshold: float = 0.7) -> List[TribeMatch]:
        """
        使用模板匹配检测种族
        
        Args:
            image: 输入图像 (RGB 或 BGR)
            threshold: 匹配阈值 (0-1)
            
        Returns:
            匹配的种族列表
        """
        matches = []
        
        # 预处理图像
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        for tribe, template in self.tribe_templates.items():
            if template is None:
                continue
            
            # 模板也要转换为灰度
            if len(template.shape) == 3:
                t_gray = cv2.cvtColor(template, cv2.COLOR_BGR2GRAY)
            else:
                t_gray = template
            
            # 模板匹配
            try:
                result = cv2.matchTemplate(gray, t_gray, cv2.TM_CCOEFF_NORMED)
                min_val, max_val, min_loc, max_loc = cv2.minMaxLoc(result)
                
                if max_val >= threshold:
                    matches.append(TribeMatch(
                        tribe=tribe,
                        confidence=max_val,
                        location=max_loc
                    ))
            except cv2.error:
                continue
        
        # 按置信度排序
        matches.sort(key=lambda x: x.confidence, reverse=True)
        
        return matches
    
    def detect_tribe_in_minion(self, minion_image: np.ndarray) -> Optional[Tribe]:
        """
        检测单个随从的种族
        
        Args:
            minion_image: 随从图像
            
        Returns:
            种族或 None
        """
        matches = self.match_tribe_template_matching(minion_image, threshold=0.6)
        
        if matches:
            return matches[0].tribe
        
        return None
    
    def detect_tribes_in_bar(self, bar_image: np.ndarray, minion_count: int = 7) -> List[Tribe]:
        """
        检测随从栏中所有随从的种族
        
        Args:
            bar_image: 随从栏图像
            minion_count: 随从数量
            
        Returns:
            每个随从的种族列表
        """
        img_height, img_width = bar_image.shape[:2]
        minion_width = img_width // minion_count
        
        tribes = []
        
        for i in range(minion_count):
            left = i * minion_width
            right = (i + 1) * minion_width
            
            # 提取单个随从区域
            margin = minion_width // 10
            minion_img = bar_image[:, left + margin:right - margin]
            
            tribe = self.detect_tribe_in_minion(minion_img)
            tribes.append(tribe if tribe else Tribe.UNKNOWN)
        
        return tribes
    
    def detect_tribe_by_color(self, image: np.ndarray) -> Optional[Tribe]:
        """
        使用颜色特征检测种族 (辅助方法)
        
        不同种族有代表性的颜色:
        - Dragon: 金色/红色
        - Murloc: 蓝色/绿色
        - Demon: 红色/紫色
        - Beast: 棕色/橙色
        - Mech: 橙色/黄色
        - Elemental: 蓝绿色
        - Undead: 紫色/灰色
        
        Args:
            image: 输入图像
            
        Returns:
            可能的种族
        """
        if len(image.shape) == 2:
            hsv = cv2.cvtColor(cv2.cvtColor(image, cv2.COLOR_GRAY2BGR), cv2.COLOR_BGR2HSV)
        else:
            hsv = cv2.cvtColor(image, cv2.COLOR_BGR2HSV)
        
        # 计算平均颜色
        mean_hue = np.mean(hsv[:, :, 0])
        mean_sat = np.mean(hsv[:, :, 1])
        mean_val = np.mean(hsv[:, :, 2])
        
        # 基于色调判断
        if 0 <= mean_hue <= 20 or mean_hue >= 340:  # 红色
            return Tribe.DEMON
        elif 20 <= mean_hue <= 40:  # 橙色/黄色
            return Tribe.MECH if mean_sat > 100 else Tribe.BEAST
        elif 40 <= mean_hue <= 70:  # 绿色
            return Tribe.MURLOC if mean_sat > 80 else Tribe.TREANT
        elif 70 <= mean_hue <= 130:  # 青色/绿色
            return Tribe.ELEMENTAL
        elif 130 <= mean_hue <= 180:  # 蓝色
            return Tribe.MURLOC if mean_sat > 100 else Tribe.DRAGON
        elif 260 <= mean_hue <= 340:  # 紫色
            return Tribe.UNDEAD
        
        return None
    
    def add_tribe_icon(self, tribe: Tribe, icon_path: str):
        """
        添加自定义种族图标
        
        Args:
            tribe: 种族
            icon_path: 图标路径
        """
        template = cv2.imread(icon_path, cv2.IMREAD_COLOR)
        if template is not None:
            self.tribe_templates[tribe] = template
    
    def __len__(self):
        return len(self.tribe_templates)


if __name__ == "__main__":
    # 测试代码
    detector = TribeDetector()
    print(f"Loaded {len(detector)} tribe icons")
    
    # 测试颜色检测
    import mss
    sct = mss.mss()
    img = np.array(sct.grab({"left": 0, "top": 0, "width": 100, "height": 100}))
    
    tribe = detector.detect_tribe_by_color(img)
    print(f"Detected tribe by color: {tribe}")
    
    sct.close()
