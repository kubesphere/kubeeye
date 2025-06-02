#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
集群信息展示组件
"""
import streamlit as st
from utils.cluster_config import get_cluster

def display_cluster_info(cluster_name):
    """显示集群基本信息"""
    cluster_config = get_cluster(cluster_name)
    if not cluster_config:
        st.error(f"无法加载集群配置: {cluster_name}")
        return None, None, None
        
    with st.expander("集群基本信息", expanded=True):
        col1, col2 = st.columns(2)
        
        with col1:
            st.markdown(f"**集群名称:** {cluster_name}")
            # 安全处理描述字段，可能不存在
            description = getattr(cluster_config, 'description', '') or cluster_config.config.get('description', '无描述')
            st.markdown(f"**集群描述:** {description}")
            nodes = cluster_config.get_nodes()
            st.markdown(f"**节点数量:** {len(nodes)}")
        
        with col2:
            # Prometheus信息
            prometheus_config = cluster_config.get_prometheus_config() or {}
            if prometheus_config and prometheus_config.get('enabled', False):
                st.markdown(f"**Prometheus地址:** {prometheus_config.get('url', 'N/A')}")
            else:
                st.markdown("**Prometheus:** 未配置")
            
            # Kubeconfig信息
            kubeconfig = cluster_config.get_kubeconfig()
            if kubeconfig:
                st.markdown("**Kubeconfig:** 已配置")
            else:
                st.markdown("**Kubeconfig:** 未配置")
    
    # 返回配置信息，避免重复获取
    return cluster_config, cluster_config.get_nodes(), cluster_config.get_prometheus_config() or {}
