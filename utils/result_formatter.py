#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
统一结果格式化器模块
提供统一的结果格式化功能，消除各检查器中重复的 _pass_result、_fail_result、_error_result 方法
"""

import logging
from typing import Dict, List, Optional, Any
from utils.rule_loader import Rule

# 设置日志
logger = logging.getLogger(__name__)

class ResultFormatter:
    """
    统一的结果格式化器
    合并了各检查器中重复的结果格式化方法
    """
    
    @staticmethod
    def format_result(rule: Rule, status: str, description: str, severity: str, 
                     details: str, solution: Optional[str] = None, 
                     violations: Optional[List[Dict]] = None, 
                     **kwargs) -> Dict:
        """
        格式化检查结果（统一版本）
        
        Args:
            rule: 规则对象
            status: 状态（passed, failed, error, warning, skipped等）
            description: 结果简短描述
            severity: 严重性级别
            details: 详细信息
            solution: 可选的解决方案
            violations: 可选的违规列表
            **kwargs: 其他额外字段（如node、variables、assertions等）
            
        Returns:
            格式化的结果字典
        """
        if solution is None:
            solution = rule.solution if hasattr(rule, 'solution') else ""
            
        result = {
            'name': rule.name,
            'status': status,
            'description': description,
            'severity': severity,
            'details': details,
            'solution': solution,
            'rule_id': rule.id
        }
        
        # 如果有violations，添加到结果中
        if violations is not None:
            result['violations'] = violations
        
        # 添加其他额外字段
        result.update(kwargs)
            
        return result
    
    @staticmethod
    def pass_result(rule: Rule, description: str, details: str = "", **kwargs) -> Dict:
        """
        生成通过结果
        
        Args:
            rule: 规则对象
            description: 描述信息
            details: 详细信息
            **kwargs: 其他额外字段
            
        Returns:
            格式化的通过结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="passed",
            description=description,
            severity="info",
            details=details,
            solution="",
            **kwargs
        )
    
    @staticmethod
    def fail_result(rule: Rule, description: str, details: str, 
                   severity: str = "warning", violations: List[Dict] = None, 
                   **kwargs) -> Dict:
        """
        生成失败结果
        
        Args:
            rule: 规则对象
            description: 描述信息
            details: 详细信息
            severity: 严重级别
            violations: 违规列表
            **kwargs: 其他额外字段
            
        Returns:
            格式化的失败结果
        """
        solution = rule.solution if hasattr(rule, 'solution') else ""
        return ResultFormatter.format_result(
            rule=rule,
            status="failed",
            description=description,
            severity=severity,
            details=details,
            solution=solution,
            violations=violations or [],
            **kwargs
        )
    
    @staticmethod
    def error_result(rule: Rule, error_msg: str, description: str = None, **kwargs) -> Dict:
        """
        生成错误结果
        
        Args:
            rule: 规则对象
            error_msg: 错误消息
            description: 自定义描述，默认为"规则执行失败"
            **kwargs: 其他额外字段
            
        Returns:
            格式化的错误结果
        """
        if description is None:
            description = "规则执行失败"
            
        return ResultFormatter.format_result(
            rule=rule,
            status="error",
            description=description,
            severity="error",
            details=error_msg,
            solution="",
            **kwargs
        )
    
    @staticmethod
    def warning_result(rule: Rule, description: str, details: str, **kwargs) -> Dict:
        """
        生成警告结果
        
        Args:
            rule: 规则对象
            description: 描述信息
            details: 详细信息
            **kwargs: 其他额外字段
            
        Returns:
            格式化的警告结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="warning",
            description=description,
            severity="warning",
            details=details,
            solution=rule.solution if hasattr(rule, 'solution') else "",
            **kwargs
        )
    
    @staticmethod
    def skipped_result(rule: Rule, reason: str, **kwargs) -> Dict:
        """
        生成跳过结果
        
        Args:
            rule: 规则对象
            reason: 跳过原因
            **kwargs: 其他额外字段
            
        Returns:
            格式化的跳过结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="skipped",
            description=f"{rule.name} 已跳过: {reason}",
            severity="info",
            details=reason,
            solution="",
            **kwargs
        )
    
    @staticmethod
    def not_applicable_result(rule: Rule, reason: str, **kwargs) -> Dict:
        """
        生成不适用结果
        
        Args:
            rule: 规则对象
            reason: 不适用原因
            **kwargs: 其他额外字段
            
        Returns:
            格式化的不适用结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="not_applicable",
            description=f"规则不适用: {reason}",
            severity="info",
            details=f"规则 {rule.name} 不适用于当前环境: {reason}",
            solution="",
            **kwargs
        )
    
    @staticmethod
    def invalid_result(rule: Rule, description: str, details: str, **kwargs) -> Dict:
        """
        生成配置无效结果
        
        Args:
            rule: 规则对象
            description: 简要描述
            details: 详细信息
            **kwargs: 其他额外字段
            
        Returns:
            格式化的配置无效结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="invalid",
            description=description,
            severity="warning",
            details=details,
            solution="请检查规则配置并修正问题",
            **kwargs
        )
    
    @staticmethod
    def critical_result(rule: Rule, description: str, details: str, **kwargs) -> Dict:
        """
        生成严重问题结果
        
        Args:
            rule: 规则对象
            description: 描述信息
            details: 详细信息
            **kwargs: 其他额外字段
            
        Returns:
            格式化的严重问题结果
        """
        return ResultFormatter.format_result(
            rule=rule,
            status="failed",
            description=description,
            severity="critical",
            details=details,
            solution=rule.solution if hasattr(rule, 'solution') else "",
            **kwargs
        )
