#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则处理器模块，提供通用的规则处理功能

此模块旨在消除 NodeInspector、PrometheusInspector 和 OpaInspector 之间的重复代码。
它提供了共享的规则加载、解析、阈值处理和结果格式化功能。
"""

import logging
from typing import Dict, List, Any, Optional, Callable
from utils.rule_loader import Rule

# 设置日志
logger = logging.getLogger(__name__)

class RuleProcessor:
    """
    规则处理器，提供通用的规则处理功能
    """
    
    @staticmethod
    def get_rule_config(rule: Rule, path: str, default_value: Any = None) -> Any:
        """
        安全地从规则配置中获取值，支持点表示法路径
        
        Args:
            rule: 规则对象
            path: 配置路径，使用点表示法，例如 "execution.query"
            default_value: 如果路径不存在，返回的默认值
            
        Returns:
            路径指向的值，如果路径不存在则返回默认值
        """
        parts = path.split('.')
        current = getattr(rule, 'config', {})
        
        # 处理非config情况，直接从rule获取顶层属性
        if not current and hasattr(rule, parts[0]):
            if len(parts) == 1:
                return getattr(rule, parts[0])
            else:
                # 如果属性值是字典，继续处理子路径
                current = getattr(rule, parts[0])
                if not isinstance(current, dict):
                    return default_value
                parts = parts[1:]
        
        for part in parts:
            if isinstance(current, dict) and part in current:
                current = current[part]
            else:
                return default_value
        
        return current
    
    @staticmethod
    def combine_threshold_result(value: Any, threshold: Dict, comparison_fn: Optional[Callable] = None) -> Dict:
        """
        比较值与阈值，并返回结果
        
        Args:
            value: 待比较的值
            threshold: 阈值配置字典，例如 {"warning": 80, "critical": 90}
            comparison_fn: 可选的比较函数，默认为大于等于
            
        Returns:
            比较结果字典，包含是否违规、严重性级别等
        """
        if comparison_fn is None:
            comparison_fn = lambda v, t: float(v) >= float(t)
            
        severity = None
        violation = False
        threshold_level = None
        threshold_value = None
            
        # 按严重性顺序检查阈值（从低到高）
        for level in ['info', 'warning', 'critical']:
            if level in threshold and threshold[level] is not None:
                if comparison_fn(value, threshold[level]):
                    severity = level
                    violation = True
                    threshold_level = level
                    threshold_value = threshold[level]
        
        return {
            'violation': violation,
            'severity': severity,
            'threshold_level': threshold_level,
            'threshold_value': threshold_value,
            'value': value
        }
    
    @staticmethod
    def format_rule_result(rule: Rule, status: str, description: str, severity: str, 
                           details: str, solution: Optional[str] = None) -> Dict:
        """
        格式化规则检查结果
        
        Args:
            rule: 规则对象
            status: 状态（passed, failed, error, warning, skipped, unknown）
            description: 结果简短描述
            severity: 严重性级别
            details: 详细信息
            solution: 可选的解决方案
            
        Returns:
            格式化的结果字典
        """
        if solution is None:
            solution = rule.solution if hasattr(rule, 'solution') else ""
            
        return {
            'name': rule.name,
            'status': status,
            'description': description,
            'severity': severity,
            'details': details,
            'solution': solution,
            'rule_id': rule.id
        }
    
    @staticmethod
    def get_severity_order(severity: str) -> int:
        """
        获取严重性级别的顺序值，用于排序
        
        Args:
            severity: 严重性级别名称
            
        Returns:
            严重性级别的整数顺序值
        """
        severity_order = {
            'critical': 4,
            'high': 3,
            'warning': 2,
            'info': 1,
            'unknown': 0
        }
        
        return severity_order.get(severity.lower(), 0)
    
    @staticmethod
    def get_highest_severity(severities: List[str]) -> str:
        """
        获取最高级别的严重性
        
        Args:
            severities: 严重性列表
            
        Returns:
            最高级别的严重性
        """
        highest = 'info'
        highest_value = 0
        
        for sev in severities:
            sev = sev.lower()
            value = RuleProcessor.get_severity_order(sev)
            if value > highest_value:
                highest = sev
                highest_value = value
        
        return highest
        
    @staticmethod
    def get_comparison_function(operator: str) -> Callable:
        """
        根据操作符获取比较函数
        
        Args:
            operator: 比较操作符（>=, >, ==, <, <=）
            
        Returns:
            比较函数
        """
        if operator == '>':
            return lambda v, t: float(v) > float(t)
        elif operator == '>=':
            return lambda v, t: float(v) >= float(t)
        elif operator == '==':
            return lambda v, t: float(v) == float(t)
        elif operator == '<':
            return lambda v, t: float(v) < float(t)
        elif operator == '<=':
            return lambda v, t: float(v) <= float(t)
        else:
            # 默认为大于等于
            return lambda v, t: float(v) >= float(t)
    
    @staticmethod
    def get_threshold_value(rule: Rule, key: str, default_value: Any = None) -> Any:
        """
        获取规则的阈值值
        
        Args:
            rule: 规则对象
            key: 阈值键，例如 'warning', 'critical', 'expected_value'
            default_value: 如果不存在，返回的默认值
            
        Returns:
            阈值值，如果不存在则返回默认值
        """
        # 获取阈值配置
        thresholds = getattr(rule, 'thresholds', {}) or {}
        
        # 直接检查key
        if key in thresholds:
            return thresholds.get(key)
        
        # 检查第一级嵌套值
        for k, v in thresholds.items():
            if isinstance(v, dict) and key in v:
                return v.get(key)
        
        # 如果找不到，返回默认值
        return default_value
