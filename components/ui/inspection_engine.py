#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
统一巡检执行引擎 - 消除重复代码，提供统一的巡检执行接口
"""
import streamlit as st
import logging
import json
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime

from utils.cluster_config import get_cluster
from utils.inspection_result import InspectionResult
from inspectors.node.node_inspector import NodeInspector
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from inspectors.opa.opa_inspector import OpaInspector
from inspectors.controller import InspectionController

logger = logging.getLogger(__name__)
from components.ui.progress import InspectionProgress


class InspectionEngine:
    """统一的巡检执行引擎"""
    
    def __init__(self):
        self.progress = None
        
    def execute_inspection(
        self, 
        cluster_name: str,
        selected_rules: Dict[str, List[str]] = None,
        inspection_type: str = "immediate",
        show_progress: bool = True,
        show_ui_feedback: bool = True
    ) -> Tuple[bool, str, Optional[Dict[str, Any]]]:
        """
        执行统一的巡检任务
        
        Args:
            cluster_name: 集群名称
            selected_rules: 选中的规则 {"node": [...], "prometheus": [...], "opa": [...]}
            inspection_type: 巡检类型 ("immediate" 或 "scheduled")
            show_progress: 是否显示进度条
            show_ui_feedback: 是否显示UI反馈
            
        Returns:
            (成功标志, 消息, 结果字典)
        """
        if show_progress:
            # 计算总规则数量
            total_rules = 0
            if "node" in selected_rules:
                total_rules += len(selected_rules["node"])
            if "prometheus" in selected_rules:
                total_rules += len(selected_rules["prometheus"])
            if "opa" in selected_rules:
                total_rules += len(selected_rules["opa"])
            
            self.progress = InspectionProgress()
            self.progress.initialize(total_rules, by_rules=True)
        
        try:
            # 获取集群配置
            if show_progress:
                self.progress.update("正在获取集群配置...")
            
            cluster_config = get_cluster(cluster_name)
            if not cluster_config:
                error_msg = f"无法找到集群配置: {cluster_name}"
                if show_progress:
                    self.progress.error(error_msg)
                return False, error_msg, None
            
            # 获取集群资源配置
            nodes = cluster_config.get_nodes() if hasattr(cluster_config, 'get_nodes') else []
            prometheus_config = cluster_config.get_prometheus_config() if hasattr(cluster_config, 'get_prometheus_config') else {}
            kubeconfig = cluster_config.get_kubeconfig() if hasattr(cluster_config, 'get_kubeconfig') else ""
            
            # 确定可用的巡检类型
            run_node_check = bool(nodes) and (selected_rules and selected_rules.get("node"))
            run_prometheus_check = bool(prometheus_config and prometheus_config.get('enabled', False)) and (selected_rules and selected_rules.get("prometheus"))
            run_opa_check = bool(kubeconfig) and (selected_rules and selected_rules.get("opa"))
            
            # 调试日志：打印选择的规则
            logger.info(f"🔍 巡检类型判断 - 节点数量: {len(nodes)}, Prometheus启用: {prometheus_config.get('enabled', False) if prometheus_config else False}, kubeconfig: {'有' if kubeconfig else '无'}")
            logger.info(f"🔍 选择的规则: {selected_rules}")
            logger.info(f"🔍 巡检类型决策 - 节点: {run_node_check}, Prometheus: {run_prometheus_check}, OPA: {run_opa_check}")
            
            if not (run_node_check or run_prometheus_check or run_opa_check):
                error_msg = "没有可用的巡检类型，请检查集群配置和规则选择"
                if show_progress:
                    self.progress.error(error_msg)
                return False, error_msg, None
            
            # 执行巡检
            all_results = {}
            
            # 节点巡检
            if run_node_check:
                success, result = self._execute_node_inspection(
                    cluster_name, nodes, selected_rules["node"], show_progress
                )
                if success:
                    all_results['node'] = result
                else:
                    return False, f"节点巡检失败: {result}", None
            else:
                if show_progress:
                    if self.progress.by_rules and "node" in selected_rules:
                        # 按规则模式：跳过的节点规则也要计入进度
                        for rule_id in selected_rules["node"]:
                            self.progress.update(f"跳过节点规则: {rule_id}(节点巡检未启用)", step_complete=True)
                    else:
                        self.progress.update("跳过节点巡检", step_complete=True)
            
            # Prometheus巡检
            if run_prometheus_check:
                success, result = self._execute_prometheus_inspection(
                    cluster_name, prometheus_config, selected_rules["prometheus"], show_progress
                )
                if success:
                    all_results['prometheus'] = result
                else:
                    return False, f"Prometheus巡检失败: {result}", None
            else:
                if show_progress:
                    if self.progress.by_rules and "prometheus" in selected_rules:
                        # 按规则模式：跳过的Prometheus规则也要计入进度
                        for rule_id in selected_rules["prometheus"]:
                            self.progress.update(f"跳过Prometheus规则: {rule_id}(Prometheus巡检未启用)", step_complete=True)
                    else:
                        self.progress.update("跳过Prometheus指标巡检", step_complete=True)
            
            # OPA巡检
            if run_opa_check:
                success, result = self._execute_opa_inspection(
                    cluster_name, kubeconfig, selected_rules["opa"], show_progress
                )
                if success:
                    all_results['opa'] = result
                else:
                    return False, f"OPA巡检失败: {result}", None
            else:
                if show_progress:
                    if self.progress.by_rules and "opa" in selected_rules:
                        # 按规则模式：跳过的OPA规则也要计入进度
                        for rule_id in selected_rules["opa"]:
                            self.progress.update(f"跳过OPA规则: {rule_id}(OPA巡检未启用)", step_complete=True)
                    else:
                        self.progress.update("跳过OPA合规性巡检", step_complete=True)
            
            if show_progress:
                self.progress.complete()
            
            # 保存结果
            if all_results:
                result_path = self._save_inspection_results(
                    all_results, cluster_name, cluster_config, inspection_type
                )
                
                # 显示UI反馈
                if show_ui_feedback:
                    self._show_inspection_completion_ui(all_results, result_path, cluster_name)
                
                return True, f"巡检完成，结果已保存到: {result_path}", all_results
            else:
                error_msg = "没有生成任何巡检结果"
                if show_progress:
                    self.progress.warning(error_msg)
                return False, error_msg, None
                
        except Exception as e:
            error_msg = f"执行巡检任务失败: {str(e)}"
            if show_progress:
                self.progress.error(error_msg)
            return False, error_msg, None
    
    def _execute_node_inspection(self, cluster_name: str, nodes: List[Dict], 
                               selected_rules: List[str], show_progress: bool) -> Tuple[bool, Any]:
        """执行节点巡检"""
        try:
            node_inspector = NodeInspector(nodes)
            
            if show_progress and self.progress.by_rules:
                # 按规则逐个执行和报告进度
                combined_result = InspectionResult(cluster_name, "node")
                for rule_id in selected_rules:
                    self.progress.update(f"正在执行节点规则: {rule_id}", rule_name=rule_id)
                    rule_result = node_inspector.run_inspection(cluster_name, [rule_id])
                    if rule_result and hasattr(rule_result, 'items'):
                        for item in rule_result.items:
                            combined_result.add_item(item)
                    self.progress.update(f"节点规则 {rule_id} 执行完成", step_complete=True)
                return True, combined_result
            else:
                # 传统方式：按类别执行
                if show_progress:
                    self.progress.update("正在执行节点巡检...")
                
                result = node_inspector.run_inspection(cluster_name, selected_rules)
                
                if show_progress:
                    self.progress.update("节点巡检完成", step_complete=True)
                return True, result
            
        except Exception as e:
            if show_progress:
                self.progress.error(f"节点巡检失败: {e}")
            return False, str(e)
    
    def _execute_prometheus_inspection(self, cluster_name: str, prometheus_config: Dict,
                                     selected_rules: List[str], show_progress: bool) -> Tuple[bool, Any]:
        """执行Prometheus巡检"""
        try:
            if prometheus_config and prometheus_config.get("enabled", False):
                prometheus_inspector = PrometheusInspector(prometheus_config)
                
                if show_progress and self.progress.by_rules:
                    # 按规则逐个执行和报告进度
                    combined_result = InspectionResult(cluster_name, "prometheus")
                    for rule_id in selected_rules:
                        self.progress.update(f"正在执行Prometheus规则: {rule_id}", rule_name=rule_id)
                        rule_result = prometheus_inspector.run_inspection(cluster_name, [rule_id])
                        if rule_result and hasattr(rule_result, 'items'):
                            for item in rule_result.items:
                                combined_result.add_item(item)
                        self.progress.update(f"Prometheus规则 {rule_id} 执行完成", step_complete=True)
                    return True, combined_result
                else:
                    # 传统方式：按类别执行
                    if show_progress:
                        self.progress.update("正在执行Prometheus指标巡检...")
                    
                    result = prometheus_inspector.run_inspection(cluster_name, selected_rules)
                    
                    if show_progress:
                        self.progress.update("Prometheus指标巡检完成", step_complete=True)
                    return True, result
            else:
                if show_progress:
                    if self.progress.by_rules:
                        # 跳过的规则也要计入进度
                        for rule_id in selected_rules:
                            self.progress.update(f"跳过Prometheus规则: {rule_id}(未配置)", step_complete=True)
                    else:
                        self.progress.update("跳过Prometheus指标巡检(未配置)", step_complete=True)
                return True, None
                
        except Exception as e:
            if show_progress:
                self.progress.error(f"Prometheus指标巡检失败: {e}")
            return False, str(e)
    
    def _execute_opa_inspection(self, cluster_name: str, kubeconfig: str,
                              selected_rules: List[str], show_progress: bool) -> Tuple[bool, Any]:
        """执行OPA巡检"""
        try:
            if kubeconfig:
                opa_config = {
                    'kubeconfig': kubeconfig,
                    'opa_path': 'opa'
                }
                opa_inspector = OpaInspector(opa_config)
                
                if show_progress and self.progress.by_rules:
                    # 按规则逐个执行和报告进度
                    combined_result = InspectionResult(cluster_name, "opa")
                    for rule_id in selected_rules:
                        self.progress.update(f"正在执行OPA规则: {rule_id}", rule_name=rule_id)
                        rule_result = opa_inspector.run_inspection(cluster_name, [rule_id])
                        if rule_result and hasattr(rule_result, 'items'):
                            for item in rule_result.items:
                                combined_result.add_item(item)
                        self.progress.update(f"OPA规则 {rule_id} 执行完成", step_complete=True)
                    return True, combined_result
                else:
                    # 传统方式：按类别执行
                    if show_progress:
                        self.progress.update("正在执行OPA合规性巡检...")
                    
                    result = opa_inspector.run_inspection(cluster_name, selected_rules)
                    
                    if show_progress:
                        self.progress.update("OPA合规性巡检完成", step_complete=True)
                    return True, result
            else:
                if show_progress:
                    if self.progress.by_rules:
                        # 跳过的规则也要计入进度
                        for rule_id in selected_rules:
                            self.progress.update(f"跳过OPA规则: {rule_id}(未配置)", step_complete=True)
                    else:
                        self.progress.update("跳过OPA合规性巡检(未配置)", step_complete=True)
                return True, None
                
        except Exception as e:
            if show_progress:
                self.progress.error(f"OPA合规性巡检失败: {e}")
            return False, str(e)
    
    def _save_inspection_results(self, all_results: Dict, cluster_name: str, 
                               cluster_config: Any, inspection_type: str) -> str:
        """保存巡检结果"""
        try:
            if hasattr(cluster_config, 'get_dict'):
                config_dict = cluster_config.get_dict()
            else:
                config_dict = {
                    'nodes': cluster_config.get_nodes() if hasattr(cluster_config, 'get_nodes') else [],
                    'prometheus': cluster_config.get_prometheus_config() if hasattr(cluster_config, 'get_prometheus_config') else {},
                    'opa': {'kubeconfig': cluster_config.get_kubeconfig() if hasattr(cluster_config, 'get_kubeconfig') else ''}
                }
        except Exception:
            config_dict = {'nodes': [], 'prometheus': {}, 'opa': {'kubeconfig': ''}}
        
        controller = InspectionController(config_dict)
        return controller.save_inspection_result(all_results, cluster_name, inspection_type)
    
    def _show_inspection_completion_ui(self, all_results: Dict, result_path: str, cluster_name: str):
        """显示巡检完成UI反馈"""
        st.success("✅ 巡检任务完成！")
        
        # 计算统计信息 - 使用简化的状态体系
        total_items = sum(len(result.items) if hasattr(result, 'items') else 0 for result in all_results.values() if result)
        passed_count = 0
        exception_count = 0
        
        for result in all_results.values():
            if result and hasattr(result, 'items'):
                passed_count += len([item for item in result.items if item.get('status') == 'passed'])
                exception_count += len([item for item in result.items if item.get('status') != 'passed'])
        
        # 创建任务摘要卡片 - 简化为通过/异常两类
        col1, col2, col3 = st.columns(3)
        with col1:
            st.metric("总检查项", total_items)
        with col2:
            st.metric("✅ 通过", passed_count)
        with col3:
            st.metric("⚠️ 异常", exception_count)

        
        # 设置会话状态，用于报告页面
        st.session_state.last_result_path = result_path
        st.session_state.last_cluster_name = cluster_name
        
        # 尝试从保存的结果中获取result_id
        try:
            with open(result_path, 'r', encoding='utf-8') as f:
                result_data = json.load(f)
            st.session_state.last_result_id = result_data.get('result_id')
        except Exception:
            # 如果无法读取，生成一个默认的result_id
            from datetime import datetime
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            st.session_state.last_result_id = f"immediate_{timestamp}"
        
        # 操作按钮和提示信息
        col1, col2 = st.columns(2)
        
        with col1:
            # 设置报告页面参数，方便用户查看
            result_id = st.session_state.get("last_result_id")
            if result_id:
                st.session_state.selected_report_id = result_id
                st.session_state.view_mode = "detail"
            
            # 显示友好的查看提示
            st.success("✅ 巡检完成！报告已准备就绪")
            st.info("📊 请点击左侧导航栏中的「**巡检报告**」页面查看详细结果")
                    
        with col2:
            if st.button("🔄 重新巡检", use_container_width=True):
                st.rerun()


# 创建全局引擎实例
inspection_engine = InspectionEngine()


def execute_inspection_unified(cluster_name: str, selected_rules: Dict[str, List[str]] = None,
                             inspection_type: str = "immediate", show_progress: bool = True,
                             show_ui_feedback: bool = True) -> Tuple[bool, str, Optional[Dict[str, Any]]]:
    """
    统一的巡检执行接口 - 供所有组件使用
    
    Args:
        cluster_name: 集群名称
        selected_rules: 选中的规则
        inspection_type: 巡检类型
        show_progress: 是否显示进度条
        show_ui_feedback: 是否显示UI反馈
        
    Returns:
        (成功标志, 消息, 结果字典)
    """
    return inspection_engine.execute_inspection(
        cluster_name, selected_rules, inspection_type, show_progress, show_ui_feedback
    )


def execute_inspection_task(task, show_progress=True):
    """
    兼容旧接口 - 为定时任务提供向后兼容性
    
    Args:
        task: 定时任务对象
        show_progress: 是否显示进度条
        
    Returns:
        (成功标志, 消息, 结果字典)
    """
    # 构建规则字典
    selected_rules = {}
    if hasattr(task, 'rules') and task.rules:
        for rule_type in ['node', 'prometheus', 'opa']:
            if rule_type in task.rules and task.rules[rule_type].get("enabled", False):
                selected_rules[rule_type] = task.rules[rule_type].get("rules", [])
    
    return execute_inspection_unified(
        cluster_name=task.cluster,
        selected_rules=selected_rules,
        inspection_type="scheduled",
        show_progress=show_progress,
        show_ui_feedback=False  # 定时任务不显示UI反馈
    )
