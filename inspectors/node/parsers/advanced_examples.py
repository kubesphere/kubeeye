#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
高级解析器示例，用于展示更复杂的解析情况和参数处理
"""
from typing import Dict, Any, Optional
import re
import json

from . import BaseParser, register_parser, logger
from ..rule_result_builder import RuleResultBuilder

@register_parser("advanced_system_parser")
class AdvancedSystemParser(BaseParser):
    """高级系统解析器示例"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """
        解析多种系统信息的高级解析器
        
        Args:
            stdout: 命令输出
            rule: 规则对象
            node: 节点信息
            extra_data: 额外数据，比如解析器配置
            
        Returns:
            解析结果
        """
        try:
            # 获取解析器配置
            extra_data = extra_data or {}
            parser_mode = extra_data.get('mode', 'auto')
            
            # 自动检测模式
            if parser_mode == 'auto':
                # 尝试判断输出类型
                if re.search(r'cpu|processor|MHz', stdout, re.I):
                    parser_mode = 'cpu'
                elif re.search(r'mem|memory|swap', stdout, re.I):
                    parser_mode = 'memory'
                elif re.search(r'disk|mount|filesystem|/dev/', stdout, re.I):
                    parser_mode = 'disk'
                elif re.search(r'load average|uptime', stdout, re.I):
                    parser_mode = 'load'
                else:
                    parser_mode = 'generic'
            
            # 根据模式选择不同的解析策略
            if parser_mode == 'cpu':
                return cls._parse_cpu_info(stdout, rule, node, extra_data)
            elif parser_mode == 'memory':
                return cls._parse_memory_info(stdout, rule, node, extra_data)
            elif parser_mode == 'disk':
                return cls._parse_disk_info(stdout, rule, node, extra_data)
            elif parser_mode == 'load':
                return cls._parse_load_info(stdout, rule, node, extra_data)
            else:
                return cls._parse_generic_info(stdout, rule, node, extra_data)
                
        except Exception as e:
            logger.exception(f"解析过程发生错误: {str(e)}")
            return RuleResultBuilder.error(
                rule=rule,
                description="高级系统解析器出错",
                details=f"解析错误: {str(e)}\n原始输出: {stdout}",
                node=node
            )
    
    @classmethod
    def _parse_cpu_info(cls, stdout: str, rule: Any, node: Dict, extra_data: Dict) -> Dict:
        """解析CPU信息"""
        # 简单的CPU使用率判断示例
        cpu_usage_match = re.search(r'(\d+\.?\d*)%', stdout)
        if cpu_usage_match:
            usage = float(cpu_usage_match.group(1))
            warning = extra_data.get('cpu_warning', 70)
            critical = extra_data.get('cpu_critical', 90)
            
            if usage >= critical:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': "CPU使用率严重偏高",
                    'severity': 'critical',
                    'details': f"CPU使用率: {usage:.1f}% (超过严重阈值 {critical}%)",
                    'solution': rule.solution if hasattr(rule, 'solution') else "检查高CPU使用进程"
                }
            elif usage >= warning:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': "CPU使用率偏高",
                    'severity': 'warning',
                    'details': f"CPU使用率: {usage:.1f}% (超过警告阈值 {warning}%)",
                    'solution': rule.solution if hasattr(rule, 'solution') else "监控CPU使用情况"
                }
            else:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'passed',
                    'description': "CPU使用率正常",
                    'severity': 'info',
                    'details': f"CPU使用率: {usage:.1f}%",
                    'solution': ""
                }
        
        # 如果没有找到百分比，返回解析失败
        return {
            'name': f"{rule.name} - {node['ip']}",
            'status': 'unknown',
            'description': "无法解析CPU使用率",
            'severity': 'warning',
            'details': f"无法从输出中提取CPU使用率: {stdout}",
            'solution': "检查命令输出格式"
        }
    
    @classmethod
    def _parse_memory_info(cls, stdout: str, rule: Any, node: Dict, extra_data: Dict) -> Dict:
        """解析内存信息"""
        # 实现内存信息解析逻辑
        # 这里省略具体实现...
        return {
            'name': f"{rule.name} - {node['ip']}",
            'status': 'passed',
            'description': "内存使用情况解析示例",
            'severity': 'info',
            'details': f"这是内存解析器的占位结果，实际实现请根据输出格式定制",
            'solution': ""
        }
    
    @classmethod
    def _parse_disk_info(cls, stdout: str, rule: Any, node: Dict, extra_data: Dict) -> Dict:
        """解析磁盘信息"""
        # 实现磁盘信息解析逻辑
        # 这里省略具体实现...
        return {
            'name': f"{rule.name} - {node['ip']}",
            'status': 'passed',
            'description': "磁盘使用情况解析示例",
            'severity': 'info',
            'details': f"这是磁盘解析器的占位结果，实际实现请根据输出格式定制",
            'solution': ""
        }
    
    @classmethod
    def _parse_load_info(cls, stdout: str, rule: Any, node: Dict, extra_data: Dict) -> Dict:
        """解析负载信息"""
        # 实现负载信息解析逻辑
        # 这里省略具体实现...
        return {
            'name': f"{rule.name} - {node['ip']}",
            'status': 'passed',
            'description': "系统负载解析示例",
            'severity': 'info',
            'details': f"这是负载解析器的占位结果，实际实现请根据输出格式定制",
            'solution': ""
        }
    
    @classmethod
    def _parse_generic_info(cls, stdout: str, rule: Any, node: Dict, extra_data: Dict) -> Dict:
        """解析通用信息"""
        # 尝试处理JSON格式
        try:
            data = json.loads(stdout)
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': "解析JSON格式数据",
                'severity': 'info',
                'details': f"成功解析JSON数据，包含 {len(data)} 个条目",
                'solution': ""
            }
        except json.JSONDecodeError:
            pass
        
        # 如果不是JSON，尝试匹配键值对
        kv_pairs = {}
        for line in stdout.splitlines():
            if ':' in line:
                key, value = line.split(':', 1)
                kv_pairs[key.strip()] = value.strip()
        
        if kv_pairs:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': "解析键值对数据",
                'severity': 'info',
                'details': f"提取了 {len(kv_pairs)} 个键值对",
                'solution': ""
            }
        
        # 如果都不是，返回原始文本
        return {
            'name': f"{rule.name} - {node['ip']}",
            'status': 'passed',
            'description': "通用文本解析",
            'severity': 'info',
            'details': f"原始输出包含 {len(stdout)} 字符",
            'solution': ""
        }
