#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
统一断言管理器模块
提供统一的断言评估和模板渲染功能，消除 AssertionEvaluator 和 RuleProcessor 之间的重复
"""

import logging
from typing import Dict, List, Any, Union
import jinja2
from simpleeval import simple_eval

# 设置日志
logger = logging.getLogger(__name__)

class AssertionManager:
    """
    统一的断言管理器
    合并了 AssertionEvaluator 和 RuleProcessor 的断言评估功能
    """
    
    def __init__(self):
        """初始化断言管理器"""
        # 创建 Jinja2 环境用于模板渲染
        self.env = jinja2.Environment(
            undefined=jinja2.DebugUndefined,  # 使用调试模式，未定义变量显示为调试信息
            autoescape=False
        )
    
    def evaluate_condition(self, condition: str, context: Dict[str, Any]) -> bool:
        """
        评估单个条件表达式
        
        Args:
            condition: 条件表达式字符串
            context: 上下文变量字典
            
        Returns:
            条件是否满足
        """
        try:
            # 提供默认值给未定义的变量
            safe_context = self._create_safe_context(context)
            
            # 使用 simpleeval 评估表达式
            result = simple_eval(condition, names=safe_context)
            return bool(result)
            
        except Exception as e:
            logger.error(f"评估条件表达式时出错: {str(e)}")
            logger.debug(f"条件: {condition}, 上下文: {context}")
            return False
    
    def render_template(self, template: str, context: Dict[str, Any]) -> str:
        """
        使用 Jinja2 渲染模板
        
        Args:
            template: 模板字符串
            context: 上下文变量
            
        Returns:
            渲染后的字符串
        """
        try:
            # 提供默认值给未定义的变量
            safe_context = self._create_safe_context(context)
            
            # 使用安全的上下文渲染模板
            jinja_template = self.env.from_string(template)
            return jinja_template.render(**safe_context)
        except Exception as e:
            logger.error(f"渲染模板时出错: {str(e)}")
            return f"ERROR: {str(e)}"
    
    def evaluate_assertions(self, assertions: List[Dict], context: Dict[str, Any], 
                          mode: str = "detailed") -> Dict[str, Any]:
        """
        评估断言列表，支持两种模式
        
        Args:
            assertions: 断言配置列表
            context: 上下文变量字典
            mode: 评估模式
                - "simple": 简化模式，只处理第一个断言（兼容旧 AssertionEvaluator）
                - "detailed": 详细模式，处理所有断言（兼容 RuleProcessor）
            
        Returns:
            评估结果字典
        """
        if not assertions:
            return {
                'passed': True,
                'failed_assertions': [],
                'severity': 'info',
                'description': '没有断言需要评估',
                'pass_description': '无断言需要评估',
                'fail_description': ''
            }
        
        if mode == "simple":
            return self._evaluate_simple_mode(assertions, context)
        else:
            return self._evaluate_detailed_mode(assertions, context)
    
    def _evaluate_simple_mode(self, assertions: List[Dict], context: Dict[str, Any]) -> Dict[str, Any]:
        """
        简化模式：只处理第一个断言（兼容原 AssertionEvaluator）
        """
        assertion = assertions[0]
        condition = assertion.get('condition', 'True')
        severity = assertion.get('severity', 'warning')
        description = assertion.get('description', '断言检查')
        
        try:
            # 评估条件
            passed = self.evaluate_condition(condition, context)
            
            if passed:
                # 生成通过描述
                if '{{' in description:
                    pass_description = self.render_template(description, context)
                else:
                    pass_description = description
                
                return {
                    'passed': True,
                    'pass_description': pass_description,
                    'fail_description': '',
                    'severity': 'info'
                }
            else:
                # 生成失败描述
                if '{{' in description:
                    fail_description = self.render_template(description, context)
                else:
                    fail_description = description
                
                return {
                    'passed': False,
                    'pass_description': '',
                    'fail_description': fail_description,
                    'severity': severity
                }
                
        except Exception as e:
            logger.error(f"评估断言时出错: {str(e)}")
            return {
                'passed': False,
                'pass_description': '',
                'fail_description': f'断言评估错误: {str(e)}',
                'severity': 'error'
            }
    
    def _evaluate_detailed_mode(self, assertions: List[Dict], context: Dict[str, Any]) -> Dict[str, Any]:
        """
        详细模式：处理所有断言（兼容原 RuleProcessor）
        """
        failed_assertions = []
        
        for assertion in assertions:
            name = assertion.get('name', '未命名断言')
            condition = assertion.get('condition', '')
            severity = assertion.get('severity', 'warning')
            description = assertion.get('description', '断言失败')
            
            if not condition:
                logger.warning(f"断言 '{name}' 没有定义条件")
                continue
                
            try:
                # 评估条件
                passed = self.evaluate_condition(condition, context)
                if not passed:
                    # 断言失败
                    failed_assertion = {
                        'name': name,
                        'condition': condition,
                        'severity': severity,
                        'description': self.render_template(description, context)
                    }
                    failed_assertions.append(failed_assertion)
            except Exception as e:
                logger.exception(f"评估断言 '{name}' 时出错: {str(e)}")
                failed_assertion = {
                    'name': name,
                    'condition': condition,
                    'severity': 'error',
                    'description': f"评估断言时出错: {str(e)}"
                }
                failed_assertions.append(failed_assertion)
        
        # 确定整体评估结果
        passed = len(failed_assertions) == 0
        
        # 获取最高严重级别
        severities = [fa['severity'] for fa in failed_assertions]
        highest_severity = self._get_highest_severity(severities) if severities else 'info'
        
        # 构建描述信息
        if failed_assertions:
            descriptions = [fa['description'] for fa in failed_assertions]
            result_description = "; ".join(descriptions)
        else:
            result_description = "所有断言都通过了"
            
        return {
            'passed': passed,
            'failed_assertions': failed_assertions,
            'severity': highest_severity,
            'description': result_description
        }
    
    def _create_safe_context(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        创建安全的上下文，为未定义的变量提供默认值
        
        Args:
            context: 原始上下文变量字典
            
        Returns:
            安全的上下文字典
        """
        safe_context = {}
        for key, value in context.items():
            if value is None:
                safe_context[key] = 0
            elif isinstance(value, str) and value.strip() == '':
                safe_context[key] = 0
            else:
                safe_context[key] = value
        return safe_context
    
    def _get_highest_severity(self, severities: List[str]) -> str:
        """
        获取最高严重级别
        
        Args:
            severities: 严重级别列表
            
        Returns:
            最高严重级别
        """
        severity_order = {
            'critical': 4,
            'error': 3,
            'warning': 2,
            'info': 1
        }
        
        max_order = 0
        highest_severity = 'info'
        
        for severity in severities:
            order = severity_order.get(severity.lower(), 1)
            if order > max_order:
                max_order = order
                highest_severity = severity
                
        return highest_severity

# 为了向后兼容，保留 AssertionEvaluator 类作为 AssertionManager 的别名
AssertionEvaluator = AssertionManager
