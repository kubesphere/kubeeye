#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
新版规则加载器模块，支持断言格式
"""

import yaml
import logging
import os
from pathlib import Path
from typing import Dict, List, Any, Optional

# 设置日志
logger = logging.getLogger(__name__)

# 规则目录定义
RULES_DIR = Path(__file__).parent.parent / "rules"

class Rule:
    """规则类，表示一个巡检规则，支持断言格式"""
    
    def __init__(self, rule_data: Dict):
        """
        初始化规则对象，支持断言格式
        
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
        
        # 断言模式特定字段
        self.assertions = self.config.get('assertions', [])  # 断言配置列表
        self.extractors = self.config.get('extractors', [])  # 提取器配置列表
        
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
            'tier': self.tier,
            'config': self.config
        }
        
        return rule_dict
        
    @property
    def execution(self) -> Dict:
        """获取执行配置"""
        return self.config.get('execution', {})

def load_rules(rule_type: str = None, include_disabled: bool = False) -> List[Rule]:
    """
    加载指定类型的规则
    
    Args:
        rule_type: 规则类型，如 node、opa、prometheus，为None则加载所有规则
        include_disabled: 是否包含禁用规则
        
    Returns:
        规则列表
    """
    rules = []
    
    # 确定要搜索的目录
    search_dirs = []
    if rule_type:
        # 只搜索指定类型的规则目录
        type_dir = RULES_DIR / rule_type
        if type_dir.exists():
            search_dirs.append(type_dir)
    else:
        # 搜索所有规则目录
        for item in RULES_DIR.iterdir():
            if item.is_dir() and not item.name.startswith('_') and not item.name == 'examples':
                search_dirs.append(item)
    
    # 从每个目录加载规则
    for rules_dir in search_dirs:
        dir_rule_type = rules_dir.name  # 从目录名推断规则类型
        
        for file_path in rules_dir.glob('*.yaml'):
            try:
                # 加载YAML文件
                with open(file_path, 'r', encoding='utf-8') as f:
                    rule_data = yaml.safe_load(f)
                
                # 确保规则数据是字典
                if not isinstance(rule_data, dict):
                    logger.warning(f"规则文件 {file_path} 格式错误，应为YAML字典")
                    continue
                
                # 如果没有明确指定type，则从目录名推断
                if 'type' not in rule_data:
                    rule_data['type'] = dir_rule_type
                
                # 创建规则对象
                rule = Rule(rule_data)
                
                # 检查是否应该加入结果
                if rule.enabled or include_disabled:
                    rules.append(rule)
            
            except Exception as e:
                logger.error(f"加载规则文件 {file_path} 失败: {str(e)}")
    
    logger.info(f"已加载 {len(rules)} 条规则")
    return rules

def load_rule_from_file(file_path: str) -> Optional[Rule]:
    """
    从文件加载单个规则
    
    Args:
        file_path: 规则文件路径
        
    Returns:
        规则对象，如果加载失败则返回None
    """
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            rule_data = yaml.safe_load(f)
        
        if not isinstance(rule_data, dict):
            logger.warning(f"规则文件 {file_path} 格式错误，应为YAML字典")
            return None
        
        # 如果没有明确指定type，则从文件路径推断
        if 'type' not in rule_data:
            # 尝试从路径中提取规则类型
            path_parts = Path(file_path).parts
            for part in path_parts:
                if part in ('node', 'opa', 'prometheus'):
                    rule_data['type'] = part
                    break
        
        return Rule(rule_data)
    
    except Exception as e:
        logger.error(f"加载规则文件 {file_path} 失败: {str(e)}")
        return None

def save_rule(rule: Rule) -> bool:
    """
    保存规则到文件
    
    Args:
        rule: 要保存的规则对象
        
    Returns:
        是否保存成功
    """
    try:
        # 确定规则类型目录
        rule_type_dir = RULES_DIR / rule.type
        
        # 确保目录存在
        if not rule_type_dir.exists():
            rule_type_dir.mkdir(parents=True, exist_ok=True)
        
        # 规则文件路径
        file_path = rule_type_dir / f"{rule.id}.yaml"
        
        # 将规则转换为字典
        rule_dict = rule.to_dict()
        
        # 保存到文件
        with open(file_path, 'w', encoding='utf-8') as f:
            yaml.dump(rule_dict, f, default_flow_style=False, allow_unicode=True)
            
        logger.info(f"规则 {rule.id} 已成功保存到文件 {file_path}")
        return True
        
    except Exception as e:
        logger.error(f"保存规则 {rule.id} 失败: {str(e)}")
        return False
