#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
节点巡检器，支持并发执行优化
"""

import logging
import os
import re
import subprocess
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import Dict, List, Any, Optional, Tuple, Union

from inspectors.base_inspector import BaseInspector
from utils.inspection_result import InspectionResult
from utils.rule_loader import Rule
from utils.node_connection import NodeConnection

# 设置日志
logger = logging.getLogger(__name__)

class NodeInspector(BaseInspector):
    """节点巡检器，支持并发执行优化"""
    
    def __init__(self, config: List[Dict[str, Any]], enable_concurrent: bool = True, 
                 max_workers: int = 5, timeout: int = 30):
        """
        初始化节点巡检器
        
        Args:
            config: 节点配置列表，包含连接信息
            enable_concurrent: 是否启用并发执行，默认True
            max_workers: 最大并发线程数，默认5个
            timeout: 单个节点命令执行超时时间（秒），默认30秒
        """
        self.nodes = config
        self.enable_concurrent = enable_concurrent
        self.max_workers = min(max_workers, len(config)) if enable_concurrent else 1
        self.timeout = timeout
        
        # 统计信息
        self.stats = {
            'total_rules': 0,
            'total_node_executions': 0,
            'successful_executions': 0,
            'failed_executions': 0,
            'total_time': 0
        }
        
        super().__init__({"nodes": config})
        
        logger.info(f"节点巡检器初始化完成 - 并发模式: {'开启' if enable_concurrent else '关闭'}, "
                   f"最大并发数: {self.max_workers}, 超时: {timeout}秒")
    
    @property
    def inspector_type(self) -> str:
        return "node"
    
    def _validate_rule_config(self, rule: Rule) -> List[str]:
        """
        验证规则配置是否有效
        
        Args:
            rule: 规则对象
            
        Returns:
            配置问题列表，如果没有问题则为空列表
        """
        issues = []
        
        # 检查必要的命令配置
        command = self.get_rule_config(rule, 'execution.command', '')
        if not command:
            issues.append("缺少必要的执行命令(execution.command)")
        
        # 检查必要的断言配置
        assertions = self.get_rule_config(rule, 'assertions', [])
        if not assertions:
            issues.append("缺少必要的断言配置(assertions)")
            
        return issues
            
    def _apply_rule(self, rule: Rule, context: Dict) -> List[Dict]:
        """
        应用单条规则进行节点检查（支持并发）
        
        Args:
            rule: 要应用的规则
            context: 检查上下文
            
        Returns:
            检查结果列表
        """
        rule_start_time = time.time()
        
        # 获取命令和断言配置
        command = self.get_rule_config(rule, 'execution.command', '')
        assertions = self.get_rule_config(rule, 'assertions', [])
        
        # 获取节点选择器并过滤节点
        node_selector = self.get_rule_config(rule, 'scope.node_selector', {})
        target_nodes = self._filter_nodes_by_selector(self.nodes, node_selector)
        
        if not target_nodes:
            logger.warning(f"规则 {rule.id} 没有匹配的节点")
            return []
        
        # 选择执行模式：如果启用并发且节点数>1，使用并发；否则串行
        if self.enable_concurrent and len(target_nodes) > 1:
            logger.info(f"规则 {rule.id}: 将在 {len(target_nodes)} 个节点上并发执行（最大并发数: {self.max_workers}）")
            node_results = self._execute_rule_concurrently(rule, command, assertions, target_nodes)
        else:
            logger.info(f"规则 {rule.id}: 将在 {len(target_nodes)} 个节点上串行执行")
            node_results = self._execute_rule_sequentially(rule, command, assertions, target_nodes)
        
        rule_duration = time.time() - rule_start_time
        logger.info(f"规则 {rule.id} 执行完成，耗时 {rule_duration:.2f}秒")
        
        # 更新统计信息
        self.stats['total_rules'] += 1
        self.stats['total_node_executions'] += len(target_nodes)
        self.stats['total_time'] += rule_duration
        
        return node_results
    
    def _execute_rule_concurrently(self, rule: Rule, command: str, 
                                 assertions: List[Dict], target_nodes: List[Dict]) -> List[Dict]:
        """
        在多个节点上并发执行规则
        
        Args:
            rule: 规则对象
            command: 要执行的命令
            assertions: 断言列表
            target_nodes: 目标节点列表
            
        Returns:
            所有节点的检查结果列表
        """
        node_results = []
        
        # 使用线程池并发执行
        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            # 提交所有节点任务
            future_to_node = {}
            for node in target_nodes:
                future = executor.submit(
                    self._execute_rule_on_single_node,
                    rule, command, assertions, node
                )
                future_to_node[future] = node
            
            # 收集结果
            for future in as_completed(future_to_node, timeout=self.timeout * len(target_nodes)):
                node = future_to_node[future]
                node_name = node.get('name', node['ip'])
                
                try:
                    result = future.result(timeout=self.timeout)
                    node_results.append(result)
                    self.stats['successful_executions'] += 1
                    logger.debug(f"节点 {node_name} 执行完成")
                    
                except Exception as e:
                    logger.error(f"节点 {node_name} 执行失败: {str(e)}")
                    # 创建错误结果
                    error_result = self._format_error_result(rule, 
                        f"节点执行失败", str(e))
                    error_result['name'] = f"{rule.name} - {node_name}"
                    error_result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
                    node_results.append(error_result)
                    self.stats['failed_executions'] += 1
        
        return node_results
    
    def _execute_rule_sequentially(self, rule: Rule, command: str, 
                                 assertions: List[Dict], target_nodes: List[Dict]) -> List[Dict]:
        """
        在多个节点上串行执行规则（原始方式）
        
        Args:
            rule: 规则对象
            command: 要执行的命令
            assertions: 断言列表
            target_nodes: 目标节点列表
            
        Returns:
            所有节点的检查结果列表
        """
        node_results = []
        
        for node in target_nodes:
            try:
                result = self._execute_rule_on_single_node(rule, command, assertions, node)
                node_results.append(result)
                self.stats['successful_executions'] += 1
                
            except Exception as e:
                node_name = node.get('name', node['ip'])
                logger.error(f"节点 {node_name} 执行失败: {str(e)}")
                
                error_result = self._format_error_result(rule, 
                    f"节点执行失败", str(e))
                error_result['name'] = f"{rule.name} - {node_name}"
                error_result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
                node_results.append(error_result)
                self.stats['failed_executions'] += 1
        
        return node_results
    
    def _execute_rule_on_single_node(self, rule: Rule, command: str, 
                                   assertions: List[Dict], node: Dict) -> Dict:
        """
        在单个节点上执行规则
        
        Args:
            rule: 规则对象
            command: 要执行的命令
            assertions: 断言列表
            node: 节点信息
            
        Returns:
            该节点的检查结果
        """
        node_name = node.get('name', node['ip'])
        
        # 执行命令
        output, error = self._execute_command(command, node)
        
        if error:
            # 命令执行错误
            node_result = self._format_error_result(rule, 
                f"命令执行失败", error)
            node_result['name'] = f"{rule.name} - {node_name}"
            node_result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
        else:
            # 准备变量字典 - 简化版本
            variables = {
                'output': output.strip(),  # 直接使用命令输出
                'node_ip': node['ip'],
                'node_name': node.get('name', node['ip'])
            }
            
            # 评估断言
            node_result = self._evaluate_assertions(rule, assertions, variables, node)
        
        return node_result
        
        return node_results
        
    def _evaluate_assertions(self, rule: Rule, assertions: List[Dict], 
                            variables: Dict[str, Any], node: Dict) -> Dict:
        """
        评估断言
        
        Args:
            rule: 规则对象
            assertions: 断言列表
            variables: 变量字典
            node: 节点信息
            
        Returns:
            评估结果
        """
        # 评估所有断言
        assertion_result = self.rule_processor.evaluate_assertions(assertions, variables)
        
        # 根据断言评估结果格式化检查结果
        if assertion_result['passed']:
            status = "passed"
            severity = "info"
            # 对于通过的检查，在描述中显示具体的检查结果
            first_assertion = assertions[0] if assertions else {}
            first_assertion_desc = first_assertion.get('description', '')
            if first_assertion_desc:
                # 渲染模板以显示具体值
                from utils.assertion_evaluator import AssertionEvaluator
                evaluator = AssertionEvaluator()
                rendered_desc = evaluator._render_template(first_assertion_desc, variables)
                description = f"{rule.name}: {rendered_desc}"
            else:
                description = f"{rule.name}: 当前值为 {variables.get('output', 'N/A')}"
            details = "检查通过，系统状态正常"
            solution = ""
        else:
            status = "failed"
            severity = assertion_result['severity']
            # 移除"断言失败"前缀，直接使用描述
            description = assertion_result['description'].replace("断言失败: ", "")
            
            # 构建详细信息
            failed_assertions = assertion_result['failed_assertions']
            details = "检查失败详情:\n" + "\n".join(
                [f"- {fa['name']}: {fa['description']}" for fa in failed_assertions]
            )
            solution = rule.solution
        
        # 格式化结果
        result = self.rule_processor.format_rule_result(
            rule=rule,
            status=status,
            description=description,
            severity=severity,
            details=details,
            solution=solution
        )
        
        # 修改名称以包含节点信息，方便UI显示
        node_name = node.get('name', node['ip'])
        result['name'] = f"{rule.name} - {node_name}"
        
        # 添加节点信息和变量信息
        result['node'] = {'ip': node['ip'], 'name': node.get('name', node['ip'])}
        result['variables'] = variables
        result['assertions'] = {
            'total': len(assertions),
            'failed': len(assertion_result.get('failed_assertions', [])),
            'failures': assertion_result.get('failed_assertions')
        }
        
        return result

    def _execute_command(self, command: str, node: Dict) -> Tuple[str, str]:
        """
        在节点上执行命令
        
        Args:
            command: 要执行的命令
            node: 节点信息
            
        Returns:
            命令输出和错误信息的元组
        """
        try:
            # 获取节点基本信息
            ip = node.get('ip', 'unknown')
            port = node.get('port', 22)
            username = node.get('username', 'unknown')
            node_name = node.get('name', ip)
            
            logger.info(f"正在连接节点 {node_name} ({ip}:{port}) 用户: {username}")

            # 验证节点配置完整性
            auth_type = node.get('auth_type', 'password')
            if auth_type == 'password' and not node.get('password'):
                error_msg = f"节点 {node_name} 配置错误: 使用密码认证但未提供密码"
                logger.error(error_msg)
                return "", error_msg
            elif auth_type == 'key' and not node.get('key_path'):
                error_msg = f"节点 {node_name} 配置错误: 使用密钥认证但未提供密钥路径"
                logger.error(error_msg)
                return "", error_msg
            
            # 简化命令显示（如果命令太长）
            display_command = command[:100] + "..." if len(command) > 100 else command
            logger.info(f"在节点 {node_name} 上执行命令: {display_command}")
                
            # 使用SSH执行命令
            with NodeConnection(node) as conn:
                if not conn.connected:
                    error_msg = f"无法连接到节点 {node_name} ({ip}): SSH连接失败"
                    logger.error(error_msg)
                    return "", error_msg
                    
                success, stdout, stderr = conn.execute_command(command)
                if success:
                    logger.info(f"命令在节点 {node_name} 上执行成功，输出长度: {len(stdout)}")
                    return stdout, ""
                else:
                    error_msg = f"命令执行失败: {stderr}"
                    logger.error(f"节点 {node_name}: {error_msg}")
                    return "", error_msg
                    
        except ConnectionError as e:
            error_msg = f"网络连接错误: {str(e)}"
            logger.error(f"连接节点 {node.get('name', node.get('ip'))} 失败: {error_msg}")
            return "", error_msg
        except TimeoutError as e:
            error_msg = f"连接超时: {str(e)}"
            logger.error(f"连接节点 {node.get('name', node.get('ip'))} 超时: {error_msg}")
            return "", error_msg
        except Exception as e:
            error_msg = f"执行命令时发生意外错误: {str(e)}"
            logger.error(f"节点 {node.get('name', node.get('ip'))}: {error_msg}", exc_info=True)
            return "", error_msg
            
    def _filter_nodes_by_selector(self, nodes: List[Dict], node_selector: Dict) -> List[Dict]:
        """
        根据节点选择器过滤节点
        
        Args:
            nodes: 节点列表
            node_selector: 节点选择器配置
            
        Returns:
            过滤后的节点列表
        """
        if not node_selector:
            return nodes
            
        filtered_nodes = []
        for node in nodes:
            # 检查节点标签是否匹配选择器
            node_labels = node.get('labels', {})
            match = True
            
            for label_key, label_value in node_selector.items():
                if label_key not in node_labels or str(node_labels[label_key]) != str(label_value):
                    match = False
                    break
                    
            if match:
                filtered_nodes.append(node)
                
        return filtered_nodes

    def _should_apply_rule(self, rule: Rule, context: Dict) -> bool:
        """
        判断规则是否应该应用于当前上下文
        
        Args:
            rule: 规则
            context: 上下文
            
        Returns:
            是否应用规则
        """
        # 检查节点选择器
        node_selector = self.get_rule_config(rule, 'scope.node_selector', {})
        if not node_selector:
            # 没有节点选择器，适用于所有节点
            return True
            
        # 因为我们是在 _apply_rule 中遍历节点，所以这里只需要返回 True
        # 实际的节点筛选会在 _apply_rule 中根据标签进行
        return True
    
    def _format_error_result(self, rule: Rule, description: str, error_msg: str) -> Dict:
        """格式化错误结果"""
        return self.rule_processor.format_rule_result(
            rule=rule,
            status="error",
            description=description,
            severity="error",
            details=error_msg,
            solution="请检查节点配置和网络连接"
        )
    
    def get_execution_stats(self) -> Dict:
        """获取执行统计信息"""
        return {
            **self.stats,
            'average_time_per_rule': self.stats['total_time'] / max(self.stats['total_rules'], 1),
            'success_rate': self.stats['successful_executions'] / max(self.stats['total_node_executions'], 1) * 100,
            'concurrent_mode': self.enable_concurrent,
            'max_workers': self.max_workers,
            'timeout': self.timeout
        }
    
    def print_execution_summary(self):
        """打印执行摘要"""
        stats = self.get_execution_stats()
        
        print(f"\n=== 节点巡检执行摘要 ===")
        print(f"执行模式: {'并发' if stats['concurrent_mode'] else '串行'}")
        print(f"总规则数: {stats['total_rules']}")
        print(f"总节点执行数: {stats['total_node_executions']}")
        print(f"成功执行数: {stats['successful_executions']}")
        print(f"失败执行数: {stats['failed_executions']}")
        print(f"成功率: {stats['success_rate']:.1f}%")
        print(f"总耗时: {stats['total_time']:.2f}秒")
        print(f"平均每规则耗时: {stats['average_time_per_rule']:.2f}秒")
        if stats['concurrent_mode']:
            print(f"最大并发数: {stats['max_workers']}")
        print(f"超时设置: {stats['timeout']}秒")
        print(f"========================\n")
    
    @classmethod
    def create_optimized(cls, config: List[Dict[str, Any]]) -> 'NodeInspector':
        """
        创建优化配置的节点巡检器
        
        Args:
            config: 节点配置列表
            
        Returns:
            优化配置的节点巡检器实例
        """
        node_count = len(config)
        
        # 根据节点数量自适应配置
        if node_count <= 3:
            max_workers = node_count
            timeout = 30
        elif node_count <= 10:
            max_workers = min(5, node_count)
            timeout = 25
        elif node_count <= 20:
            max_workers = min(8, node_count)
            timeout = 20
        else:
            max_workers = min(10, node_count)
            timeout = 15
        
        return cls(
            config=config,
            enable_concurrent=node_count > 1,  # 单节点时不启用并发
            max_workers=max_workers,
            timeout=timeout
        )
