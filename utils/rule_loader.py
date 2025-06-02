
#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则加载器模块，仅支持新版规则格式
"""

import yaml
import logging
from pathlib import Path
from typing import Dict, List, Any, Optional

# 设置日志
logger = logging.getLogger(__name__)

# 规则目录定义
RULES_DIR = Path(__file__).parent.parent / "rules"

class Rule:
    """规则类，表示一个巡检规则，仅支持新格式"""
    
    def __init__(self, rule_data: Dict):
        """
        初始化规则对象，仅支持新的规则格式
        
        Args:
            rule_data: 规则数据字典
        """
        # 基本元数据
        self.id = rule_data.get('id', '')
        self.name = rule_data.get('name', '')
        self.description = rule_data.get('description', '')
        self.type = rule_data.get('type', '')  # 规则类型：node, prometheus, opa
        self.category = rule_data.get('category', '')  # 规则类别
        self.severity = rule_data.get('severity', 'warning')  # 严重程度
        self.enabled = rule_data.get('enabled', True)  # 是否启用
        self.solution = rule_data.get('solution', '')  # 解决方案
        self.tags = rule_data.get('tags', [])  # 标签
        self.tier = rule_data.get('tier', 'basic')  # 规则层级(basic/standard/extended)
        
        # 核心配置
        self.config = rule_data.get('config', {})  # 统一的配置对象
        self.thresholds = self.config.get('thresholds', rule_data.get('thresholds', {}))  # 阈值配置
    
    def to_dict(self) -> Dict:
        """将规则转换为字典"""
        # 基础字段
        rule_dict = {
            'id': self.id,
            'name': self.name,
            'description': self.description,
            'type': self.type,
            'category': self.category,
            'severity': self.severity,
            'enabled': self.enabled,
            'solution': self.solution,
            'tags': self.tags,
            'tier': self.tier
        }
        
        # 添加配置
        if self.config:
            rule_dict['config'] = self.config
            
        return rule_dict


def load_rules(rule_type: Optional[str] = None, category: Optional[str] = None) -> List[Rule]:
    """
    加载规则
    
    Args:
        rule_type: 规则类型，如果为 None 则加载所有类型
        category: 规则类别，如果为 None 则加载所有类别
        
    Returns:
        规则对象列表
    """
    rules = []
    
    # 确定要搜索的目录
    if rule_type:
        dirs_to_search = [RULES_DIR / rule_type]
    else:
        dirs_to_search = [
            RULES_DIR / 'node',
            RULES_DIR / 'prometheus',
            RULES_DIR / 'opa'
        ]
    
    # 处理每个目录
    for rules_dir in dirs_to_search:
        if not rules_dir.exists() or not rules_dir.is_dir():
            logger.warning(f"规则目录不存在或不是目录: {rules_dir}")
            continue
        
        # 处理每个YAML文件
        for rule_file in rules_dir.glob('*.yaml'):
            try:
                # 加载YAML文件
                with open(rule_file, 'r', encoding='utf-8') as f:
                    rule_data = yaml.safe_load(f)
                    
                # 创建规则对象
                if rule_data:
                    rule = Rule(rule_data)
                    
                    # 应用过滤器
                    if category and rule.category != category:
                        continue
                        
                    rules.append(rule)
                    
            except Exception as e:
                logger.exception(f"加载规则文件时出错: {rule_file}")
    
    logger.info(f"加载了 {len(rules)} 条规则")
    return rules


def get_rule_by_id(rule_id: str, rule_type: Optional[str] = None) -> Optional[Rule]:
    """
    根据ID获取规则
    
    Args:
        rule_id: 规则ID
        rule_type: 规则类型，如果提供可以加速搜索
        
    Returns:
        规则对象或None
    """
    rules = load_rules(rule_type=rule_type)
    
    for rule in rules:
        if rule.id == rule_id:
            return rule
            
    return None

def save_rule(rule: Rule) -> bool:
    """
    保存规则到YAML文件
    
    Args:
        rule: 规则对象
        
    Returns:
        是否保存成功
    """
    if not rule.type or not rule.id:
        logger.error("规则保存失败: 规则类型或ID为空")
        return False
    
    # 确定保存路径
    rule_dir = RULES_DIR / rule.type
    if not rule_dir.exists():
        try:
            rule_dir.mkdir(parents=True, exist_ok=True)
        except Exception as e:
            logger.error(f"创建规则目录失败: {e}")
            return False
    
    # 文件名: <id>.yaml
    rule_file = rule_dir / f"{rule.id}.yaml"
    
    try:
        # 将规则对象转换为字典并保存为YAML
        with open(rule_file, 'w', encoding='utf-8') as f:
            yaml.dump(rule.to_dict(), f, default_flow_style=False, sort_keys=False, allow_unicode=True)
        
        logger.info(f"规则已保存: {rule_file}")
        return True
    except Exception as e:
        logger.error(f"保存规则失败: {e}")
        return False

# 导出符号
__all__ = ["Rule", "load_rules", "get_rule_by_id", "save_rule"]
