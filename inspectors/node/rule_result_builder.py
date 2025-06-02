#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则结果构建器，提供统一的方式来构建规则结果
"""
from typing import Dict, Any, Optional, Union


class RuleResultBuilder:
    """规则结果构建器，帮助构建标准格式的规则结果"""

    @staticmethod
    def create(rule: Any, 
               status: str, 
               description: str,
               severity: Optional[str] = None,
               details: Optional[str] = None,
               solution: Optional[str] = None,
               node: Optional[Dict] = None) -> Dict:
        """
        创建标准格式的规则结果
        
        Args:
            rule: 规则对象
            status: 状态，可选值为 'passed', 'failed', 'warning', 'error', 'unknown'
            description: 简短描述
            severity: 严重程度，可选值为 'info', 'warning', 'critical'
            details: 详细信息
            solution: 修复建议
            node: 节点信息，用于生成结果名称
            
        Returns:
            标准格式的规则结果字典
        """
        # 确定严重程度
        if severity is None:
            if status == 'passed':
                severity = 'info'
            elif status == 'failed':
                severity = getattr(rule, 'severity', 'warning')
            else:
                severity = 'warning'
        
        # 构建结果名称
        name = getattr(rule, 'name', str(rule))
        if node and 'ip' in node:
            name = f"{name} - {node['ip']}"
        
        # 获取解决方案
        if solution is None:
            solution = getattr(rule, 'solution', '') if hasattr(rule, 'solution') else ''
        
        return {
            'name': name,
            'status': status,
            'description': description,
            'severity': severity,
            'details': details or '',
            'solution': solution
        }
    
    @classmethod
    def passed(cls, rule: Any, description: str, details: Optional[str] = None, node: Optional[Dict] = None) -> Dict:
        """创建通过状态的规则结果"""
        return cls.create(
            rule=rule,
            status='passed',
            description=description,
            severity='info',
            details=details,
            node=node
        )
    
    @classmethod
    def failed(cls, rule: Any, description: str, details: Optional[str] = None, node: Optional[Dict] = None) -> Dict:
        """创建失败状态的规则结果"""
        return cls.create(
            rule=rule,
            status='failed',
            description=description,
            severity=getattr(rule, 'severity', 'warning'),
            details=details,
            node=node
        )
    
    @classmethod
    def error(cls, rule: Any, description: str, details: Optional[str] = None, node: Optional[Dict] = None) -> Dict:
        """创建错误状态的规则结果"""
        return cls.create(
            rule=rule,
            status='error',
            description=description,
            severity='warning',
            details=details,
            node=node
        )
    
    @classmethod
    def warning(cls, rule: Any, description: str, details: Optional[str] = None, node: Optional[Dict] = None) -> Dict:
        """创建警告状态的规则结果"""
        return cls.create(
            rule=rule,
            status='warning',
            description=description,
            severity='warning',
            details=details,
            node=node
        )
    
    @classmethod
    def unknown(cls, rule: Any, description: str, details: Optional[str] = None, node: Optional[Dict] = None) -> Dict:
        """创建未知状态的规则结果"""
        return cls.create(
            rule=rule,
            status='unknown',
            description=description,
            severity='warning',
            details=details,
            node=node
        )
