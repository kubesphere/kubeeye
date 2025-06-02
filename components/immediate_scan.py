#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
立即巡检组件
"""
import time
import streamlit as st
from utils.cluster_config import get_cluster
from utils.inspection_result import InspectionResult
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from inspectors.opa.opa_inspector import OpaInspector

# 导入节点巡检器
from inspectors.node.node_inspector import NodeInspector

from .common import create_rule_selection_tabs
from components.ui import display_cluster_info, select_inspectors, display_inspection_results, display_summary_metrics, InspectionProgress

def render_immediate_scan_tab():
    """渲染立即巡检标签页内容"""
    from utils.cluster_config import list_clusters
    
    # 加载集群列表
    clusters = list_clusters()
    
    if not clusters:
        st.warning("还没有配置任何集群。请前往「集群信息」页面添加集群。")
        if st.button("转到集群信息页面"):
            st.switch_page("pages/1_cluster_info.py")
    else:
        # 选择集群
        selected_cluster = st.selectbox("选择要巡检的集群", clusters)
        
        if selected_cluster:
            # 显示集群信息并获取配置
            cluster_config, nodes, prometheus_config = display_cluster_info(selected_cluster)
            if not cluster_config:
                return
                
            # 获取kubeconfig
            kubeconfig = cluster_config.get_kubeconfig()
            
            # 选择巡检器
            run_node_check, run_prometheus_check, run_opa_check = select_inspectors(nodes, prometheus_config, kubeconfig)
                
            # 使用RuleManager创建规则选择区域
            from utils.rule_manager import RuleManager
            selected_node_rules, selected_prometheus_rules, selected_opa_rules = RuleManager.create_rule_selection_tabs(
                run_node_check, run_prometheus_check, run_opa_check, key_suffix="_run"
            )
            
            # 运行巡检按钮
            run_inspection = st.button("开始巡检", type="primary")
            
            if run_inspection:
                # 检查是否选择了至少一个巡检类型
                if not (run_node_check or run_prometheus_check or run_opa_check):
                    st.error("请选择至少一种巡检类型")
                else:
                    # 创建进度条和结果变量
                    progress = InspectionProgress()
                    progress.initialize(3)  # 三个巡检步骤
                    all_results = {}
                    
                    # 配置节点巡检器
                    node_inspector_class = NodeInspector
                    
                    # 定义巡检器配置
                    inspector_configs = [
                        {
                            "enabled": run_node_check and nodes,
                            "type": "node",
                            "display_name": "节点巡检",
                            "inspector_class": node_inspector_class,
                            "config": nodes,
                            "rules": selected_node_rules,
                        },
                        {
                            "enabled": run_prometheus_check and prometheus_config.get('enabled', False),
                            "type": "prometheus",
                            "display_name": "Prometheus指标巡检",
                            "inspector_class": PrometheusInspector,
                            "config": prometheus_config,
                            "rules": selected_prometheus_rules,
                        },
                        {
                            "enabled": run_opa_check and kubeconfig,
                            "type": "opa",
                            "display_name": "OPA合规性巡检",
                            "inspector_class": OpaInspector,
                            "config": {'kubeconfig': kubeconfig},
                            "rules": selected_opa_rules,
                        }
                    ]
                    
                    # 执行每种巡检
                    for inspector_config in inspector_configs:
                        if not inspector_config["enabled"]:
                            progress.update(f"跳过{inspector_config['display_name']}", step_complete=True)
                            continue
                            
                        progress.update(f"正在执行{inspector_config['display_name']}...")
                        
                        try:
                            # 初始化巡检器
                            inspector = inspector_config["inspector_class"](inspector_config["config"])
                            
                            # 运行巡检
                            result = inspector.run_inspection(selected_cluster, inspector_config["rules"])
                            all_results[inspector_config["type"]] = result
                            
                            # 更新进度
                            progress.update(f"{inspector_config['display_name']}完成", step_complete=True)
                            
                        except Exception as e:
                            st.error(f"{inspector_config['display_name']}失败: {e}")
                            progress.update(f"{inspector_config['display_name']}失败", step_complete=True)
                    
                    progress.complete()
                    
                    # 合并所有巡检结果为一个统一报告
                    if all_results:
                        # 创建结果展示
                        st.subheader("巡检结果概览")
                        
                        # 显示摘要信息
                        display_summary_metrics(all_results)
                        
                        # 创建详细结果的标签页
                        if len(all_results) > 1:
                            result_tabs = st.tabs([f"{k.capitalize()}巡检结果" for k in all_results.keys()])
                            
                            # 填充每个标签页的内容
                            for i, (key, result) in enumerate(all_results.items()):
                                with result_tabs[i]:
                                    display_inspection_results(key, result)
                        else:
                            # 如果只有一个巡检结果，直接显示
                            key, result = next(iter(all_results.items()))
                            display_inspection_results(key, result)
                        
                        # 保存结果按钮
                        if st.button("保存巡检结果"):
                            # 导入控制器并保存结果
                            from inspectors.controller import InspectionController
                            controller = InspectionController(cluster_config.get_dict())
                            result_path = controller.save_inspection_result(all_results, selected_cluster)
                            st.success(f"巡检结果已保存至: {result_path}")
                            
                            # 添加查看详细报告的按钮
                            if st.button("查看详细报告"):
                                st.session_state.last_result_path = result_path
                                st.session_state.last_cluster_name = selected_cluster
                                st.switch_page("pages/3_scan_report.py")
                    else:
                        st.warning("未产生任何巡检结果")
                        
# 这部分代码已移至components/ui/result_display.py
