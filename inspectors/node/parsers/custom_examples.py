#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
自定义解析器示例，演示如何扩展解析器系统
"""
from typing import Dict, Any, Optional

from . import BaseParser, register_parser

@register_parser("custom_memory_parser")
class CustomMemoryParser(BaseParser):
    """自定义内存解析器示例"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """
        自定义解析内存使用情况
        这是一个示例解析器，演示如何扩展解析器系统
        """
        try:
            # 示例逻辑：假设输出是简单的内存使用百分比
            memory_usage = float(stdout.strip())
            
            # 获取阈值（优先从extra_data获取，其次使用默认值）
            extra_data = extra_data or {}
            warning_threshold = extra_data.get('warning_threshold')
            critical_threshold = extra_data.get('critical_threshold')
            description_prefix = extra_data.get('description_prefix', '自定义内存检查')
            
            # 使用默认阈值（如果未设置）
            if warning_threshold is None:
                warning_threshold = 70
            if critical_threshold is None:
                critical_threshold = 85
                
            # 从规则的配置中获取阈值 - 同时支持新的 thresholds 和旧的 threshold 格式
            if hasattr(rule, 'config') and 'thresholds' in rule.config:
                if warning_threshold is None and 'warning' in rule.config['thresholds']:
                    warning_threshold = rule.config['thresholds']['warning'].get('value', 70)
                if critical_threshold is None and 'critical' in rule.config['thresholds']:
                    critical_threshold = rule.config['thresholds']['critical'].get('value', 85)
            # 兼容旧格式的 thresholds 属性
            elif hasattr(rule, 'thresholds'):
                if warning_threshold is None and 'warning' in rule.thresholds:
                    warning_threshold = rule.thresholds['warning']
                if critical_threshold is None and 'critical' in rule.thresholds:
                    critical_threshold = rule.thresholds['critical']
            # 兼容旧格式的 threshold 属性
            elif hasattr(rule, 'threshold'):
                if warning_threshold is None and 'warning' in rule.threshold:
                    warning_threshold = rule.threshold['warning']
                if critical_threshold is None and 'critical' in rule.threshold:
                    critical_threshold = rule.threshold['critical']
            
            if memory_usage >= critical_threshold:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': f"{description_prefix}: 使用率严重偏高",
                    'severity': 'critical',
                    'details': f"内存使用率为 {memory_usage:.1f}%，超过严重阈值 {critical_threshold}%",
                    'solution': rule.solution or "检查内存泄漏或增加内存"
                }
            elif memory_usage >= warning_threshold:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': f"{description_prefix}: 使用率偏高",
                    'severity': 'warning',
                    'details': f"内存使用率为 {memory_usage:.1f}%，超过警告阈值 {warning_threshold}%",
                    'solution': rule.solution or "监控内存使用情况"
                }
            else:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'passed',
                    'description': f"{description_prefix}: 使用率正常",
                    'severity': 'info',
                    'details': f"内存使用率为 {memory_usage:.1f}%，低于阈值",
                    'solution': ""
                }
        except Exception as e:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'unknown',
                'description': f"{extra_data.get('description_prefix', '自定义内存检查')}: 解析失败",
                'severity': 'warning',
                'details': f"解析内存使用率数据时出错: {str(e)}",
                'solution': "检查命令输出格式"
            }
