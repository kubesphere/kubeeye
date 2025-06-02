#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
任务执行组件
"""
import streamlit as st
from datetime import datetime
from utils.cluster_config import get_cluster
from inspectors.node.node_inspector import NodeInspector
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from inspectors.opa.opa_inspector import OpaInspector
from utils.inspection_result import InspectionResult

def execute_inspection_task(task, show_progress=True):
    """
    执行巡检任务并返回结果
    
    Args:
        task: 任务对象
        show_progress: 是否显示进度条
    
    Returns:
        tuple: (成功与否, 消息, 结果字典)
    """
    if show_progress:
        from components.ui.progress import InspectionProgress
        progress = InspectionProgress()
        progress.initialize(3)  # 三个巡检步骤
    
    try:
        if show_progress:
            progress.update("正在获取集群配置...")
        
        # 获取集群配置
        cluster_config = get_cluster(task.cluster)
        if not cluster_config:
            if show_progress:
                progress.error(f"无法找到集群配置: {task.cluster}")
            return False, f"无法找到集群配置: {task.cluster}", None
        
        # 初始化结果字典
        all_results = {}
        
        # 运行节点巡检
        if "node" in task.rules and task.rules["node"].get("enabled", False):
            if show_progress:
                progress.update("正在执行节点巡检...")
            
            try:
                node_inspector = NodeInspector(cluster_config.get_nodes())
                node_result = node_inspector.run_inspection(
                    task.cluster, 
                    task.rules["node"].get("rules", [])
                )
                all_results['node'] = node_result
                
                if show_progress:
                    progress.update("节点巡检完成", step_complete=True)
                
            except Exception as e:
                if show_progress:
                    progress.error(f"节点巡检失败: {e}")
                return False, f"节点巡检失败: {e}", None
        else:
            if show_progress:
                progress.update("跳过节点巡检", step_complete=True)
        
        # 运行 Prometheus 巡检
        if "prometheus" in task.rules and task.rules["prometheus"].get("enabled", False):
            if show_progress:
                progress.update("正在执行Prometheus指标巡检...")
            
            prom_config = cluster_config.get_prometheus_config()
            if prom_config and prom_config.get("enabled", False):
                try:
                    prometheus_inspector = PrometheusInspector(prom_config)
                    prometheus_result = prometheus_inspector.run_inspection(
                        task.cluster, 
                        task.rules["prometheus"].get("rules", [])
                    )
                    all_results['prometheus'] = prometheus_result
                    
                    if show_progress:
                        progress.update("Prometheus指标巡检完成", step_complete=True)
                    
                except Exception as e:
                    if show_progress:
                        progress.error(f"Prometheus指标巡检失败: {e}")
                    return False, f"Prometheus指标巡检失败: {e}", None
            else:
                if show_progress:
                    progress.update("跳过Prometheus指标巡检(未配置)", step_complete=True)
        else:
            if show_progress:
                progress.update("跳过Prometheus指标巡检", step_complete=True)
        
        # 运行 OPA 巡检
        if "opa" in task.rules and task.rules["opa"].get("enabled", False):
            if show_progress:
                progress.update("正在执行OPA合规性巡检...")
            
            kubeconfig = cluster_config.get_kubeconfig()
            if kubeconfig:
                try:
                    opa_inspector = OpaInspector(kubeconfig)
                    opa_result = opa_inspector.run_inspection(
                        task.cluster, 
                        task.rules["opa"].get("rules", [])
                    )
                    all_results['opa'] = opa_result
                    
                    if show_progress:
                        progress.update("OPA合规性巡检完成", step_complete=True)
                    
                except Exception as e:
                    if show_progress:
                        progress.error(f"OPA合规性巡检失败: {e}")
                    return False, f"OPA合规性巡检失败: {e}", None
            else:
                if show_progress:
                    progress.update("跳过OPA合规性巡检(未配置)", step_complete=True)
        else:
            if show_progress:
                progress.update("跳过OPA合规性巡检", step_complete=True)
        
        if show_progress:
            progress.complete()
        
        # 合并所有巡检结果
        if all_results:
            timestamp = datetime.now().strftime('%Y%m%d%H%M%S')
            unified_result_id = f"{task.cluster}_定时巡检_{task.task_id}_{timestamp}"
            
            # 准备合并后的结果数据
            unified_items = []
            for result in all_results.values():
                unified_items.extend(result.get_items())
            
            # 创建综合结果对象
            unified_result = InspectionResult(task.cluster, "unified")
            unified_result.result_id = unified_result_id
            unified_result.items = unified_items
            
            # 保存统一结果
            unified_result.save()
            
            return True, "巡检完成", all_results
        else:
            if show_progress:
                progress.warning("没有生成任何巡检结果")
            return False, "没有生成任何巡检结果", None
            
    except Exception as e:
        if show_progress:
            progress.error(f"执行巡检任务失败: {e}")
        return False, f"执行巡检任务失败: {e}", None
