#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则处理器模块，提供基于断言的规则处理功能
"""

import logging
from typing import Dict, List, Any, Optional

from utils.rule_loader import Rule
from utils.assertion_manager import AssertionManager
from utils.result_formatter import ResultFormatter
from utils.result_extractor import ResultExtractor

# 设置日志
logger = logging.getLogger(__name__)

class RuleProcessor:
    """
    规则处理器，提供通用的规则处理功能
    """
    
    def __init__(self):
        """初始化规则处理器"""
        self.assertion_manager = AssertionManager()
        self.result_formatter = ResultFormatter()
        self.result_extractor = ResultExtractor()
        
        # 为了向后兼容，保留旧的属性名
        self.assertion_evaluator = self.assertion_manager
    
    @staticmethod
    def get_rule_config(rule: Rule, path: str, default_value: Any = None) -> Any:
        """
        安全地从规则配置中获取值，支持点表示法路径
        
        Args:
            rule: 规则对象
            path: 配置路径，使用点表示法，例如 "execution.command"
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
    
    def format_rule_result(self, rule: Rule, status: str, description: str, severity: str, 
                          details: str, solution: Optional[str] = None, 
                          violations: Optional[List[Dict]] = None, **kwargs) -> Dict:
        """
        格式化规则检查结果（已委托给 ResultFormatter）
        
        Args:
            rule: 规则对象
            status: 状态（passed, failed, error, warning, skipped, unknown）
            description: 结果简短描述
            severity: 严重性级别
            details: 详细信息
            solution: 可选的解决方案
            violations: 可选的违规列表
            **kwargs: 其他额外字段
            
        Returns:
            格式化的结果字典
        """
        return self.result_formatter.format_result(
            rule=rule,
            status=status,
            description=description,
            severity=severity,
            details=details,
            solution=solution,
            violations=violations,
            **kwargs
        )
    
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
            severities: 严重性级别列表
            
        Returns:
            最高级别的严重性
        """
        if not severities:
            return 'unknown'
            
        highest = 'unknown'
        highest_order = 0
        
        for severity in severities:
            order = RuleProcessor.get_severity_order(severity)
            if order > highest_order:
                highest = severity
                highest_order = order
                
        return highest
    
    def evaluate_assertions(self, assertions: List[Dict], context: Dict[str, Any]) -> Dict:
        """
        评估一组断言（已委托给 AssertionManager）
        
        Args:
            assertions: 断言列表
            context: 上下文变量字典
            
        Returns:
            评估结果字典，包含是否通过、失败的断言等
        """
        return self.assertion_manager.evaluate_assertions(assertions, context, mode="detailed")

    def extract_variables(self, output: str, extractors: List[Dict], context: Dict = None) -> Dict[str, Any]:
        """
        从输出中提取变量
        
        Args:
            output: 命令输出
            extractors: 提取器配置列表
            context: 上下文变量
            
        Returns:
            提取的变量字典
        """
        return self.result_extractor.extract(output, extractors, context)
