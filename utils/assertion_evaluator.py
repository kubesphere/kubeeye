#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
断言评估器模块，使用 simpleeval 评估条件表达式
"""

import logging
from typing import Dict, Any
import jinja2
from simpleeval import simple_eval

# 设置日志
logger = logging.getLogger(__name__)

class AssertionEvaluator:
    """断言评估器，用于评估条件表达式"""
    
    def __init__(self):
        """初始化断言评估器"""
        # 创建 Jinja2 环境用于模板渲染
        self.env = jinja2.Environment(
            undefined=jinja2.DebugUndefined,  # 使用调试模式，未定义变量显示为调试信息
            autoescape=False
        )
    
    def evaluate(self, condition: str, context: Dict[str, Any]) -> bool:
        """
        评估条件表达式
        
        Args:
            condition: 条件表达式字符串
            context: 上下文变量字典
            
        Returns:
            条件是否满足
        """
        try:
            # 提供默认值给未定义的变量
            safe_context = {}
            for key, value in context.items():
                if value is None:
                    safe_context[key] = 0
                elif isinstance(value, str) and value.strip() == '':
                    safe_context[key] = 0
                else:
                    safe_context[key] = value
            
            # 使用 simpleeval 评估表达式
            result = simple_eval(condition, names=safe_context)
            return bool(result)
            
        except Exception as e:
            logger.error(f"评估条件表达式时出错: {str(e)}")
            logger.debug(f"条件: {condition}, 上下文: {context}")
            return False
    
    def _render_template(self, template: str, context: Dict[str, Any]) -> str:
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
            safe_context = {}
            for key, value in context.items():
                if value is None:
                    safe_context[key] = 0
                elif isinstance(value, str) and value.strip() == '':
                    safe_context[key] = 0
                else:
                    safe_context[key] = value
            
            # 使用安全的上下文渲染模板
            jinja_template = self.env.from_string(template)
            return jinja_template.render(**safe_context)
        except Exception as e:
            logger.error(f"渲染模板时出错: {str(e)}")
            return f"ERROR: {str(e)}"
    
    def evaluate_assertions(self, assertions: list, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        评估断言列表
        
        Args:
            assertions: 断言配置列表
            context: 上下文变量字典
            
        Returns:
            评估结果字典，包含passed、fail_description、pass_description、severity等
        """
        if not assertions:
            return {
                'passed': True,
                'pass_description': '无断言需要评估',
                'fail_description': '',
                'severity': 'info'
            }
        
        # 简化版本：只处理第一个断言（一个规则一个检查）
        assertion = assertions[0]
        condition = assertion.get('condition', 'True')
        severity = assertion.get('severity', 'warning')
        description = assertion.get('description', '断言检查')
        
        try:
            # 评估条件
            passed = self.evaluate(condition, context)
            
            if passed:
                # 生成通过描述
                if '{{' in description:
                    pass_description = self._render_template(description, context)
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
                    fail_description = self._render_template(description, context)
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
