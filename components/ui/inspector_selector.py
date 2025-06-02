#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检器选择组件
"""
import streamlit as st

def select_inspectors(nodes, prometheus_config, kubeconfig):
    """选择要使用的巡检器类型"""
    # 使用session_state来跟踪复选框状态
    if 'run_node_check' not in st.session_state:
        st.session_state.run_node_check = True
    if 'run_prometheus_check' not in st.session_state:
        st.session_state.run_prometheus_check = bool(prometheus_config.get('enabled', False))
    if 'run_opa_check' not in st.session_state:
        st.session_state.run_opa_check = bool(kubeconfig)
    
    st.subheader("选择巡检规则")
    
    # 创建三列布局用于复选框
    check_boxes_col1, check_boxes_col2, check_boxes_col3 = st.columns(3)
    
    with check_boxes_col1:
        run_node_check = st.checkbox("节点状态巡检", value=st.session_state.run_node_check, key="node_check")
        if run_node_check and not nodes:
            st.warning("未配置节点信息，无法执行节点巡检")
            run_node_check = False
    
    with check_boxes_col2:
        run_prometheus_check = st.checkbox("Prometheus 指标巡检", value=prometheus_config.get('enabled', False))
        if run_prometheus_check and not prometheus_config.get('enabled', False):
            st.warning("未配置Prometheus信息，无法执行指标巡检")
            run_prometheus_check = False
    
    with check_boxes_col3:
        run_opa_check = st.checkbox("OPA 合规性巡检", value=bool(kubeconfig))
        if run_opa_check and not kubeconfig:
            st.warning("未配置Kubeconfig，无法执行OPA合规性巡检")
            run_opa_check = False
            
    return run_node_check, run_prometheus_check, run_opa_check
