#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
节点巡检器，继承自基类实现，使用新的解析器系统和通用规则处理器
"""

import logging
import os
import re
import subprocess
from typing import Dict, List, Any, Optional, Tuple, Union

from inspectors.base_inspector import BaseInspector
from utils.inspection_result import InspectionResult
from utils.rule_loader import Rule
from utils.node_connection import NodeConnection

# 导入解析器模块（如果有的话）
try:
    from inspectors.node.parsers import get_parser, list_parsers, parse_output
except ImportError:
    get_parser = None
    list_parsers = lambda: []
    parse_output = lambda name, stdout, rule, node, extra_data=None: None

# 设置日志
logger = logging.getLogger(__name__)

class NodeInspector(BaseInspector):
    """节点巡检器，用于执行SSH命令并基于规则进行检查"""
    
    def __init__(self, config: List[Dict[str, Any]]):
        """
        初始化节点巡检器
        
        Args:
            config: 节点配置列表，包含连接信息
        """
        self.nodes = config
        super().__init__({"nodes": config})
        
        # 记录可用解析器
        available_parsers = list_parsers()
        logger.info(f"已加载的节点解析器: {available_parsers}")
    
    @property
    def inspector_type(self) -> str:
        return "node"
    
    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        """
        应用单条规则进行节点检查
        
        Args:
            rule: 要应用的规则
            context: 检查上下文
            
        Returns:
            检查结果
        """
        # 获取命令
        command = self.get_rule_config(rule, 'execution.command', '')
        if not command:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='skipped',
                description="规则未定义执行命令",
                severity='info',
                details="检查被跳过，因为规则没有定义有效的执行命令",
                solution="检查规则定义中的execution.command"
            )
        
        # 获取解析器配置
        parser_name = self.get_rule_config(rule, 'execution.parser', '')
        
        # 针对每个节点执行检查
        node_results = []
        for node in self.nodes:
            node_context = {'node': node, **context}
            try:
                # 执行命令
                output, error = self._execute_command(command, node)
                
                if error:
                    # 命令执行错误
                    node_result = self.rule_processor.format_rule_result(
                        rule=rule,
                        status='error',
                        description=f"在节点 {node['ip']} 上执行检查命令失败",
                        severity='warning',
                        details=f"错误: {error}",
                        solution="检查节点SSH连接和命令语法"
                    )
                else:
                    # 解析输出并评估结果
                    parsed_result = self._parse_output(output, rule, parser_name)
                    node_result = self._evaluate_parsed_result(rule, parsed_result, node)
                
                # 添加节点信息
                node_result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
                node_results.append(node_result)
            except Exception as e:
                logger.exception(f"在节点 {node['ip']} 上执行规则 {rule.id} 时出错: {str(e)}")
                # 添加错误结果
                error_result = self.rule_processor.format_rule_result(
                    rule=rule,
                    status='error',
                    description=f"在节点 {node['ip']} 上执行规则失败",
                    severity='warning',
                    details=f"错误: {str(e)}",
                    solution="检查日志和节点状态"
                )
                error_result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
                node_results.append(error_result)
        
        # 汇总结果
        if len(node_results) == 1:
            # 单节点结果直接返回
            return node_results[0]
        else:
            # 多节点结果需要合并
            return self._merge_node_results(rule, node_results)
        node_selector = self.get_rule_config(rule, 'scope.node_selector')
        if node_selector:
            # 这里可以实现更复杂的节点选择器逻辑
            # 当前简单返回True，表示应用到所有节点
            pass
                
        return True
    
    def _apply_rule(self, rule: Rule, context: Dict) -> List[Dict]:
        """
        应用规则到所有节点
        
        Args:
            rule: 规则对象
            context: 上下文
            
        Returns:
            所有节点的检查结果列表
        """
        results = []
        
        for node in self.nodes:
            try:
                # 为每个节点单独创建连接
                conn = NodeConnection(node)
                success, message = conn.connect()
                
                if not success:
                    results.append(
                        self.rule_processor.format_rule_result(
                            rule=rule,
                            status='failed',
                            description=f"无法连接到节点: {message}",
                            severity='critical',
                            details=f"节点 {node['ip']} 连接失败，可能是认证问题或网络不可达",
                            solution="检查节点 SSH 配置、防火墙设置和网络连接"
                        )
                    )
                    continue
                    
                # 应用规则
                node_result = self._apply_rule_to_node(rule, conn, node)
                if node_result:
                    # 添加节点标识
                    if 'name' in node_result:
                        node_result['name'] = f"{node_result['name']} - {node['ip']}"
                    results.append(node_result)
                    
                conn.close()
            except Exception as e:
                logger.exception(f"对节点 {node['ip']} 应用规则 {rule.id} 时出错")
                results.append(
                    self.rule_processor.format_rule_result(
                        rule=rule,
                        status='error',
                        description=f"执行规则时发生错误: {str(e)}",
                        severity='warning',
                        details=f"节点 {node['ip']} 执行 {rule.name} 规则失败: {str(e)}",
                        solution="检查日志和节点状态"
                    )
                )
        
        return results
    
    def _apply_rule_to_node(self, rule: Rule, conn: NodeConnection, node: Dict) -> Dict:
        """
        对单个节点应用规则
        
        Args:
            rule: 规则对象
            conn: 节点连接
            node: 节点信息
            
        Returns:
            节点的检查结果
        """
        # 处理依赖命令和准备上下文
        preprocess_data = {}
        
        # 获取主命令和解析器
        check_command = self.get_rule_config(rule, 'execution.command') or self.get_rule_config(rule, 'query')
        parser_name = self.get_rule_config(rule, 'execution.parser') or self.get_rule_config(rule, 'custom_data.parser')
        
        # 处理依赖命令
        dependencies = self.get_rule_config(rule, 'execution.dependencies') or {}
        if isinstance(dependencies, dict):
            for key, command in dependencies.items():
                if command:
                    success, stdout, _ = conn.execute_command(command)
                    if success and stdout.strip():
                        try:
                            if key == 'cpu_cores':
                                preprocess_data[key] = int(stdout.strip())
                            else:
                                preprocess_data[key] = stdout.strip()
                        except ValueError:
                            preprocess_data[key] = stdout.strip()
        
        # 获取CPU核心数命令
        cpu_cores_cmd = self.get_rule_config(rule, 'custom_data.cpu_cores_command')
        if cpu_cores_cmd:
            success, cores_stdout, _ = conn.execute_command(cpu_cores_cmd)
            if success and cores_stdout.strip():
                try:
                    preprocess_data['cpu_cores'] = int(cores_stdout.strip())
                except ValueError:
                    preprocess_data['cpu_cores'] = 1
                    
        # 处理其他预处理命令
        preprocess_commands = self.get_rule_config(rule, 'custom_data.preprocess_commands') or {}
        if isinstance(preprocess_commands, dict):
            for key, command in preprocess_commands.items():
                if command:
                    success, stdout, _ = conn.execute_command(command)
                    if success:
                        preprocess_data[key] = stdout.strip()
        
        # 处理解析器配置
        parser_config = self.get_rule_config(rule, 'execution.parser_config') or {}
        for key, value in parser_config.items():
            preprocess_data[key] = value
        
        if not check_command:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='skipped',
                description="规则未定义检查命令",
                severity='info',
                details="检查被跳过，因为规则未定义执行命令",
                solution="检查规则定义"
            )
        
        # 执行命令
        success, stdout, stderr = conn.execute_command(check_command)
        
        if not success:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='failed',
                description="执行命令失败",
                severity=rule.severity,
                details=stderr,
                solution=rule.solution if hasattr(rule, 'solution') else "检查节点状态"
            )
        
        # 将预处理数据添加到规则上下文
        rule_context = {'rule': rule, 'node': node, **preprocess_data}
        
        # 使用解析器解析输出
        if parser_name:
            # 解析输出
            try:
                parse_result = parse_output(parser_name, stdout, rule, node, preprocess_data)
                if parse_result:
                    return parse_result
                else:
                    # 解析器不存在或返回None
                    logger.warning(f"解析器 '{parser_name}' 不存在或返回空结果")
                    return self.rule_processor.format_rule_result(
                        rule=rule,
                        status='error',
                        description=f"解析器错误: '{parser_name}' 不存在或返回空结果",
                        severity='warning',
                        details=f"规则 {rule.name} 指定的解析器 '{parser_name}' 不存在或返回空结果。\n原始输出: {stdout[:200]}{'...' if len(stdout) > 200 else ''}",
                        solution="检查规则定义中的解析器名称是否正确，以及解析器是否已正确注册"
                    )
            except Exception as e:
                logger.exception(f"解析输出失败: {str(e)}")
                return self.rule_processor.format_rule_result(
                    rule=rule,
                    status='error',
                    description=f"解析输出失败: {str(e)}",
                    severity='warning',
                    details=f"规则 {rule.name} 解析输出时出错: {str(e)}\n原始输出: {stdout[:200]}{'...' if len(stdout) > 200 else ''}",
                    solution="检查解析器代码和输出格式"
                )
        
        # 基本检查逻辑 (用于没有解析器的规则)
        # 使用通用的阈值获取方法
        expected_value = self.rule_processor.get_threshold_value(rule, 'expected_value')
            
        if expected_value is not None:
            if stdout.strip() == expected_value:
                return self.rule_processor.format_rule_result(
                    rule=rule,
                    status='passed',
                    description="检查通过",
                    severity='info',
                    details=f"结果符合预期: {expected_value}",
                    solution=""
                )
            else:
                return self.rule_processor.format_rule_result(
                    rule=rule,
                    status='failed',
                    description="检查失败",
                    severity=rule.severity,
                    details=f"结果不符合预期，预期: {expected_value}，实际: {stdout.strip()}",
                    solution=rule.solution if hasattr(rule, 'solution') else "检查节点状态"
                )
        
        # 默认情况
        return self.rule_processor.format_rule_result(
            rule=rule,
            status='unknown',
            description="规则未定义解析方式",
            severity='warning',
            details=stdout,
            solution="检查规则定义"
        )
