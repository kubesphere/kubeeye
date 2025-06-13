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
from components.ui import display_cluster_info, select_inspectors, InspectionProgress

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
            
            # 确定可用的巡检类型
            run_node_check, run_prometheus_check, run_opa_check = select_inspectors(nodes, prometheus_config, kubeconfig)
                
            # 使用新版RuleManager创建规则选择区域 - 直接展示所有可用规则类型的标签页
            from utils.rule_manager import RuleManager
            selected_node_rules, selected_prometheus_rules, selected_opa_rules = RuleManager.create_rule_selection_tabs(
                run_node_check, run_prometheus_check, run_opa_check, key_suffix="_run"
            )
            
            # 运行巡检按钮
            run_inspection = st.button("开始巡检", type="primary")
            
            if run_inspection:
                # 检查是否至少有一种巡检类型可用
                if not (run_node_check or run_prometheus_check or run_opa_check):
                    st.error("没有可用的巡检类型，请检查集群配置")
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
                    
                    # 保存巡检结果并提供查看选项
                    if all_results:
                        # 自动保存结果
                        from inspectors.controller import InspectionController
                        
                        # 安全地获取集群配置字典
                        try:
                            if hasattr(cluster_config, 'get_dict'):
                                config_dict = cluster_config.get_dict()
                            else:
                                # 如果没有get_dict方法，手动构建配置字典
                                config_dict = {
                                    'nodes': cluster_config.get_nodes() if hasattr(cluster_config, 'get_nodes') else [],
                                    'prometheus': cluster_config.get_prometheus_config() if hasattr(cluster_config, 'get_prometheus_config') else {},
                                    'opa': {'kubeconfig': cluster_config.get_kubeconfig() if hasattr(cluster_config, 'get_kubeconfig') else ''}
                                }
                        except Exception as e:
                            st.error(f"获取集群配置失败: {e}")
                            config_dict = {'nodes': [], 'prometheus': {}, 'opa': {'kubeconfig': ''}}
                        
                        controller = InspectionController(config_dict)
                        result_path = controller.save_inspection_result(all_results, selected_cluster, "immediate")
                        
                        # 显示巡检完成状态
                        st.success("✅ 巡检任务完成！")
                        
                        # 显示任务统计信息
                        total_items = sum(len(result.items) for result in all_results.values())
                        passed_count = sum(len([item for item in result.items if item.get('status') == 'passed']) for result in all_results.values())
                        failed_count = sum(len([item for item in result.items if item.get('status') == 'failed']) for result in all_results.values())
                        warning_count = sum(len([item for item in result.items if item.get('status') == 'warning']) for result in all_results.values())
                        
                        # 创建任务摘要卡片
                        col1, col2, col3, col4 = st.columns(4)
                        with col1:
                            st.metric("总检查项", total_items)
                        with col2:
                            st.metric("✅ 通过", passed_count)
                        with col3:
                            st.metric("⚠️ 警告", warning_count)
                        with col4:
                            st.metric("❌ 失败", failed_count)
                        
                        # 结果文件信息
                        st.info(f"📄 巡检结果已保存: `{result_path}`")
                        
                        # 设置会话状态，用于报告页面
                        st.session_state.last_result_path = result_path
                        st.session_state.last_cluster_name = selected_cluster
                        
                        # 操作按钮
                        col1, col2 = st.columns(2)
                        with col1:
                            if st.button("📊 查看详细报告", type="primary", use_container_width=True):
                                st.switch_page("pages/3_scan_report.py")
                        with col2:
                            if st.button("🔄 重新巡检", use_container_width=True):
                                st.rerun()
                    else:
                        st.warning("未产生任何巡检结果")
                        
# 这部分代码已移至components/ui/result_display.py
