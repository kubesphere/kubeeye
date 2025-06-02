#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
自定义解析器，包括磁盘使用率等常用解析器
"""
from typing import Dict, Any, Optional

from . import BaseParser, register_parser
from ..rule_result_builder import RuleResultBuilder

@register_parser("disk_usage_parser")
@register_parser("custom_disk_parser")
class CustomDiskParser(BaseParser):
    """自定义磁盘使用率解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """
        解析磁盘使用率
        
        Args:
            stdout: df命令输出，预期是一个数字（已经过滤掉%符号）
            rule: 规则对象
            node: 节点信息
            extra_data: 额外数据，比如解析器配置
            
        Returns:
            解析结果
        """
        try:
            # 解析磁盘使用率
            disk_usage = float(stdout.strip())
            
            # 获取阈值和配置
            extra_data = extra_data or {}
            warning_threshold = extra_data.get('warning_threshold', 80)
            critical_threshold = extra_data.get('critical_threshold', 90)
            mount_point = extra_data.get('mount_point', '/')
            
            # 从规则配置中获取阈值 - 同时支持新的 thresholds 和旧的 threshold 格式
            if hasattr(rule, 'config') and 'thresholds' in rule.config:
                if not warning_threshold:
                    warning_threshold = rule.config['thresholds'].get('warning', {}).get('value', 80)
                if not critical_threshold:
                    critical_threshold = rule.config['thresholds'].get('critical', {}).get('value', 90)
            # 也可以从规则的threshold获取 (旧格式兼容)
            elif hasattr(rule, 'threshold'):
                if not warning_threshold:
                    warning_threshold = rule.threshold.get('warning', 80)
                if not critical_threshold:
                    critical_threshold = rule.threshold.get('critical', 90)
            
            # 构建结果描述
            if disk_usage >= critical_threshold:
                return RuleResultBuilder.create(
                    rule=rule,
                    status='failed',
                    description="磁盘使用率严重偏高",
                    severity='critical',
                    details=f"挂载点 {mount_point} 使用率为 {disk_usage:.1f}%，超过严重阈值 {critical_threshold}%",
                    solution=getattr(rule, 'solution', "清理磁盘或扩容存储"),
                    node=node
                )
            elif disk_usage >= warning_threshold:
                return RuleResultBuilder.create(
                    rule=rule,
                    status='failed',
                    description="磁盘使用率偏高",
                    severity='warning',
                    details=f"挂载点 {mount_point} 使用率为 {disk_usage:.1f}%，超过警告阈值 {warning_threshold}%",
                    solution=getattr(rule, 'solution', "监控磁盘使用情况"),
                    node=node
                )
            else:
                return RuleResultBuilder.passed(
                    rule=rule,
                    description="磁盘使用率正常",
                    details=f"挂载点 {mount_point} 使用率为 {disk_usage:.1f}%，低于警告阈值 {warning_threshold}%",
                    node=node
                )
        except Exception as e:
            return RuleResultBuilder.error(
                rule=rule,
                description="解析磁盘使用率失败",
                details=f"解析错误: {str(e)}\n原始输出: {stdout}",
                node=node
            )
