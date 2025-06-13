#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检器选择组件
"""
import streamlit as st

def select_inspectors(nodes, prometheus_config, kubeconfig):
    """确定可用的巡检器类型"""
    # 根据配置确定可用的巡检器
    run_node_check = bool(nodes)
    run_prometheus_check = bool(prometheus_config and prometheus_config.get('enabled', False))
    run_opa_check = bool(kubeconfig)
    
    # 设置session state以便在其他地方使用
    st.session_state.run_node_check = run_node_check
    st.session_state.run_prometheus_check = run_prometheus_check
    st.session_state.run_opa_check = run_opa_check
            
    return run_node_check, run_prometheus_check, run_opa_check
