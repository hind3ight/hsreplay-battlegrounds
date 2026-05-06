"""
OCR随从识别模块 - 使用 RapidOCR 识别随从名称和属性
"""
import numpy as np
from dataclasses import dataclass
from typing import List, Optional, Tuple
from rapidocr_onnxruntime import RapidOCR


@dataclass
class MinionOCRResult:
    """单次 OCR 识别结果"""
    name: str           # 随从名称
    attack: int         # 攻击力
    health: int        # 生命值
    name_cn: Optional[str] = None  # 中文名 (如果识别到)
    confidence: float = 0.0         # 识别置信度


class OCRMinions:
    """随从 OCR 识别类"""
    
    def __init__(self):
        """初始化 RapidOCR"""
        self.ocr = RapidOCR()
    
    def preprocess_image(self, image: np.ndarray) -> np.ndarray:
        """
        图像预处理
        
        Args:
            image: 输入图像 (BGRA 或 BGR 格式)
            
        Returns:
            预处理后的图像
        """
        # MSS 捕获的是 BGRA 格式, 转换为 BGR
        if image.shape[2] == 4:
            image = image[:, :, :3]
        
        # 可选: 图像增强
        # 可以添加对比度调整、去噪等处理
        
        return image
    
    def detect_minion_stats(self, image: np.ndarray) -> Tuple[int, int]:
        """
        识别随从属性 (攻击力和生命值)
        
        Args:
            image: 随从属性区域图像
            
        Returns:
            (攻击力, 生命值)
        """
        result, elapse = self.ocr(image)
        
        if not result:
            return 0, 0
        
        attack, health = 0, 0
        
        for item in result:
            text = item[1]
            confidence = item[2]
            
            # 解析攻击力和生命值
            # 通常格式为 "X/Y" 或单独的数值
            if "/" in text:
                parts = text.split("/")
                try:
                    attack = int(parts[0])
                    health = int(parts[1])
                except ValueError:
                    pass
            elif text.isdigit():
                # 需要根据上下文判断是攻击还是生命
                value = int(text)
                if attack == 0:
                    attack = value
                else:
                    health = value
        
        return attack, health
    
    def recognize_minion(self, image: np.ndarray) -> Optional[MinionOCRResult]:
        """
        识别单个随从
        
        Args:
            image: 随从区域图像
            
        Returns:
            MinionOCRResult 或 None
        """
        result, elapse = self.ocr(image)
        
        if not result or len(result) == 0:
            return None
        
        # 合并所有识别结果
        all_text = " ".join([item[1] for item in result])
        avg_confidence = sum([item[2] for item in result]) / len(result)
        
        # 尝试提取名称和属性
        # 炉石随从名称通常在左边, 属性在右边
        attack, health = 0, 0
        name = ""
        
        for item in result:
            text = item[1]
            # 检查是否包含数字
            if any(c.isdigit() for c in text):
                if "/" in text:
                    parts = text.split("/")
                    try:
                        attack = int(parts[0].replace(",", ""))
                        health = int(parts[1].replace(",", ""))
                    except ValueError:
                        pass
                elif text.isdigit():
                    value = int(text)
                    if attack == 0:
                        attack = value
                    elif health == 0:
                        health = value
            else:
                # 非数字文本可能是名称
                if name:
                    name += " " + text
                else:
                    name = text
        
        return MinionOCRResult(
            name=name.strip(),
            attack=attack,
            health=health,
            confidence=avg_confidence
        )
    
    def recognize_minion_bar(self, image: np.ndarray, minion_count: int = 7) -> List[MinionOCRResult]:
        """
        识别随从栏中的所有随从
        
        Args:
            image: 随从栏完整图像
            minion_count: 随从数量, 默认7个
        
        Returns:
            随从识别结果列表
        """
        # 图像宽度
        img_height, img_width = image.shape[:2]
        
        # 计算每个随从的宽度
        minion_width = img_width // minion_count
        
        results = []
        
        for i in range(minion_count):
            left = i * minion_width
            right = (i + 1) * minion_width
            
            # 提取单个随从区域 (左右各留一点边距)
            margin = minion_width // 10
            minion_img = image[:, left + margin:right - margin]
            
            result = self.recognize_minion(minion_img)
            if result:
                results.append(result)
            else:
                # 占位
                results.append(MinionOCRResult(
                    name="", attack=0, health=0, confidence=0.0
                ))
        
        return results
    
    def close(self):
        """关闭 OCR 引擎"""
        # RapidOCR 不需要显式关闭
        pass
    
    def __enter__(self):
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


if __name__ == "__main__":
    # 测试代码
    import mss
    
    with OCRMinions() as ocr:
        sct = mss.mss()
        img = np.array(sct.grab({"left": 200, "top": 800, "width": 1520, "height": 200}))
        
        # 预处理
        img = ocr.preprocess_image(img)
        
        # 识别随从栏
        results = ocr.recognize_minion_bar(img)
        
        for i, r in enumerate(results):
            print(f"Minion {i+1}: {r.name} [{r.attack}/{r.health}] conf={r.confidence:.2f}")
        
        sct.close()
