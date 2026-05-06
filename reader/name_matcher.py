"""
随从名称匹配模块 - 使用 rapidfuzz 进行模糊匹配
"""
import json
import os
from typing import List, Optional, Dict, Tuple
from dataclasses import dataclass
from rapidfuzz import fuzz, process


@dataclass
class NameMatch:
    """名称匹配结果"""
    matched_name: str       # 匹配到的随从名称
    name_cn: str          # 中文名
    tribe: str            # 种族
    tier: int             # 随从等级
    score: float          # 匹配分数
    attack: int           # 攻击力
    health: int           # 生命值


class MinionNameMatcher:
    """随从名称匹配器"""
    
    def __init__(self, metadata_path: str = "data/minions_metadata.json", 
                 names_path: str = "data/minion_names.json"):
        """
        初始化匹配器
        
        Args:
            metadata_path: 随从元数据 JSON 路径
            names_path: 随从名称映射 JSON 路径
        """
        self.metadata_path = metadata_path
        self.names_path = names_path
        
        # 加载数据
        self.minions_data: Dict = {}
        self.name_to_cn: Dict[str, str] = {}
        self.all_names: List[str] = []
        
        self._load_data()
    
    def _load_data(self):
        """加载随从数据"""
        # 加载元数据
        if os.path.exists(self.metadata_path):
            with open(self.metadata_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
                self.minions_data = data.get('minions', {})
        
        # 加载名称映射
        if os.path.exists(self.names_path):
            with open(self.names_path, 'r', encoding='utf-8') as f:
                self.name_to_cn = json.load(f)
        
        # 合并所有名称用于匹配
        # minion_names.json: key=EN -> value=CN，需要构建 CN->EN 反向映射
        cn_to_en: Dict[str, str] = {v: k for k, v in self.name_to_cn.items()}
        self.all_names = list(self.minions_data.keys()) + list(self.name_to_cn.keys()) + list(cn_to_en.keys())
        self._cn_to_en = cn_to_en
    
    def match_name(self, ocr_text: str, threshold: int = 60) -> Optional[NameMatch]:
        """
        模糊匹配随从名称
        
        Args:
            ocr_text: OCR 识别的文本
            threshold: 匹配阈值 (0-100)
            
        Returns:
            匹配结果或 None
        """
        if not ocr_text or len(ocr_text.strip()) == 0:
            return None
        
        # 清理文本
        cleaned_text = self._clean_text(ocr_text)
        
        if not cleaned_text:
            return None
        
        # 尝试精确匹配
        if cleaned_text in self.minions_data:
            return self._create_match(cleaned_text, 100)
        
        # 尝试名称映射表精确匹配
        if cleaned_text in self.name_to_cn:
            # 反向查找
            cn_name = self.name_to_cn[cleaned_text]
            for en_name, cn in self.name_to_cn.items():
                if cn == cn_name:
                    if en_name in self.minions_data:
                        return self._create_match(en_name, 100)
        
        # 模糊匹配
        result = process.extractOne(
            cleaned_text, 
            self.all_names,
            scorer=fuzz.WRatio
        )
        
        if result and result[1] >= threshold:
            matched_name = result[0]
            score = result[1]
            
            # 如果匹配到的是英文名
            if matched_name in self.minions_data:
                return self._create_match(matched_name, score)
            
            # 如果匹配到的是中文名, 查找对应的英文名
            if matched_name in self._cn_to_en:
                en_name = self._cn_to_en[matched_name]
                if en_name in self.minions_data:
                    return self._create_match(en_name, score)
        
        return None
    
    def _clean_text(self, text: str) -> str:
        """
        清理 OCR 文本
        
        Args:
            text: 原始 OCR 文本
            
        Returns:
            清理后的文本
        """
        # 移除多余空白
        text = ' '.join(text.split())
        
        # 移除特殊字符 (保留字母、数字、空格、中文)
        import re
        text = re.sub(r'[^\w\s\u4e00-\u9fff]', '', text)
        
        return text.strip()
    
    def _create_match(self, name: str, score: float) -> NameMatch:
        """创建匹配结果"""
        # 如果匹配到的是中文名，需要先转成英文名
        en_name = self._cn_to_en.get(name, name)
        metadata = self.minions_data.get(en_name, {})

        return NameMatch(
            matched_name=en_name,
            name_cn=self.name_to_cn.get(en_name, metadata.get('name_cn', name)),
            tribe=metadata.get('tribe', 'Unknown'),
            tier=metadata.get('tier', 0),
            score=score,
            attack=metadata.get('attack', 0),
            health=metadata.get('health', 0)
        )
    
    def match_batch(self, ocr_texts: List[str], threshold: int = 60) -> List[Optional[NameMatch]]:
        """
        批量匹配随从名称
        
        Args:
            ocr_texts: OCR 识别文本列表
            threshold: 匹配阈值
            
        Returns:
            匹配结果列表
        """
        return [self.match_name(text, threshold) for text in ocr_texts]
    
    def get_minion_info(self, name: str) -> Optional[Dict]:
        """
        获取随从详细信息
        
        Args:
            name: 随从名称
            
        Returns:
            随从信息字典
        """
        return self.minions_data.get(name)
    
    def get_all_minions_by_tribe(self, tribe: str) -> List[str]:
        """
        获取指定种族的随从列表
        
        Args:
            tribe: 种族名称
            
        Returns:
            随从名称列表
        """
        result = []
        for name, data in self.minions_data.items():
            if data.get('tribe', '').lower() == tribe.lower():
                result.append(name)
        return result
    
    def reload_data(self):
        """重新加载数据"""
        self.minions_data = {}
        self.name_to_cn = {}
        self.all_names = []
        self._load_data()


def load_minion_names_mapping(names_path: str = "data/minion_names.json") -> Dict[str, str]:
    """加载随从名称映射 (便捷函数)"""
    if os.path.exists(names_path):
        with open(names_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    return {}


def match_minion_name(ocr_text: str, 
                      metadata_path: str = "data/minions_metadata.json",
                      names_path: str = "data/minion_names.json",
                      threshold: int = 60) -> Optional[NameMatch]:
    """便捷函数: 单次匹配随从名称"""
    matcher = MinionNameMatcher(metadata_path, names_path)
    return matcher.match_name(ocr_text, threshold)


if __name__ == "__main__":
    # 测试代码
    matcher = MinionNameMatcher()
    
    # 测试匹配
    test_names = ["Brann Bronzebeard", "Goldrinn the Great Wolf", "一些错误的名称"]
    
    for name in test_names:
        result = matcher.match_name(name)
        if result:
            print(f"'{name}' -> {result.matched_name} [{result.tribe}] score={result.score}")
        else:
            print(f"'{name}' -> No match")
