#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
立即巡检组件 - 重构版本，使用统一巡检引擎
"""
import streamlit as st
from utils.cluster_config import list_clusters
from components.ui import display_cluster_info, select_inspectors
from components.ui.inspection_engine import execute_inspection_unified

def render_immediate_scan_tab():
    """渲染立即巡检标签页内容 - 使用统一巡检引擎"""
    # 加载集群列表
    clusters = list_clusters()
    
    if not clusters:
        st.warning("还没有配置任何集群。请前往「集群信息」页面添加集群。")
        if st.button("转到集群信息页面", key="goto_cluster_info_btn1"):
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
                
            # 使用新版RuleManager创建规则选择区域
            from utils.rule_manager import RuleManager
            selected_node_rules, selected_prometheus_rules, selected_opa_rules = RuleManager.create_rule_selection_tabs(
                run_node_check, run_prometheus_check, run_opa_check, key_suffix="_run"
            )
            
            # 运行巡检按钮
            run_inspection = st.button("开始巡检", type="primary")
            
            if run_inspection:
                # 构建选中的规则字典
                selected_rules = {}
                if run_node_check and selected_node_rules:
                    selected_rules["node"] = selected_node_rules
                if run_prometheus_check and selected_prometheus_rules:
                    selected_rules["prometheus"] = selected_prometheus_rules
                if run_opa_check and selected_opa_rules:
                    selected_rules["opa"] = selected_opa_rules
                
                # 使用统一巡检引擎执行
                success, message, results = execute_inspection_unified(
                    cluster_name=selected_cluster,
                    selected_rules=selected_rules,
                    inspection_type="immediate",
                    show_progress=True,
                    show_ui_feedback=True
                )
                
                if not success:
                    st.error(f"巡检失败: {message}")

