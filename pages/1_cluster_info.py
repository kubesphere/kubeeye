#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 集群信息管理页面
"""

# 导入必要的库
import streamlit as st
import pandas as pd
import sys
from pathlib import Path
import json

# 设置页面配置 - 必须是第一个Streamlit命令
st.set_page_config(
    page_title="集群信息 - kubeeye",
    page_icon="🔗",
    layout="wide"
)

# 添加项目根目录到Python路径
ROOT_DIR = Path(__file__).resolve().parent.parent
if str(ROOT_DIR) not in sys.path:
    sys.path.insert(0, str(ROOT_DIR))

# 导入应用模块
from utils.common import initialize_page
from utils.cluster_config import list_clusters, get_cluster, delete_cluster
from utils.node_connection import test_node_connection
from utils.prometheus_client import PrometheusClient
from utils.k8s_client import K8sClient

# 初始化页面
initialize_page(
    title="集群信息",
    icon="🔗",
    page_title="集群信息管理", 
    page_subtitle="管理和配置您的 Kubernetes 集群连接"
)

# 选项卡
tab1, tab2, tab3 = st.tabs(["集群列表", "添加集群", "编辑集群"])

# 集群列表选项卡
with tab1:
    st.header("已配置的集群")
    
    # 刷新按钮
    if st.button("刷新列表"):
        st.rerun()
    
    # 加载集群列表
    clusters = list_clusters()
    
    if clusters:
        for cluster_name in clusters:
            with st.expander(f"集群: {cluster_name}", expanded=False):
                cluster_config = get_cluster(cluster_name)
                
                # 显示节点信息
                st.subheader("节点信息")
                nodes = cluster_config.get_nodes()
                
                if nodes:
                    node_data = []
                    for node in nodes:
                        node_data.append({
                            "IP": node['ip'],
                            "端口": node['port'],
                            "用户名": node['username'],
                            "认证类型": node['auth_type']
                        })
                    st.dataframe(pd.DataFrame(node_data))
                else:
                    st.info("未配置节点")
                
                # 显示 Prometheus 配置
                st.subheader("Prometheus 配置")
                prometheus_config = cluster_config.get_prometheus_config()
                st.json(prometheus_config)
                
                # 显示 kubeconfig 配置
                st.subheader("Kubeconfig")
                kubeconfig = cluster_config.get_kubeconfig()
                if kubeconfig:
                    st.code(kubeconfig, language="yaml")
                else:
                    st.info("未配置 kubeconfig")
                
                # 删除集群按钮
                if st.button("删除集群", key=f"delete_{cluster_name}"):
                    delete_cluster(cluster_name)
                    st.success(f"集群 {cluster_name} 已删除")
                    st.rerun()
    else:
        st.info("还没有配置任何集群，请前往「添加集群」选项卡添加集群。")

# 添加集群选项卡
with tab2:
    st.header("添加新集群")
    
    with st.form("add_cluster_form"):
        cluster_name = st.text_input("集群名称", placeholder="production")
        st.caption("请输入一个唯一的集群名称，用于标识此集群")
        
        st.subheader("节点信息")
        st.caption("添加集群节点信息，用于对节点进行巡检")
        
        col1, col2 = st.columns(2)
        with col1:
            node_ip = st.text_input("节点 IP", placeholder="192.168.1.100")
            node_port = st.text_input("SSH 端口", "22")
            node_username = st.text_input("用户名", "root")
        
        with col2:
            auth_type = st.selectbox("认证方式", ["password", "key"])
            
            if auth_type == "password":
                node_password = st.text_input("密码", type="password")
                node_key_path = ""
            else:
                node_password = ""
                node_key_path = st.text_input("密钥路径", placeholder="/home/user/.ssh/id_rsa")
        
        st.subheader("Prometheus 配置")
        st.caption("配置 Prometheus 信息，用于指标巡检")
        
        prometheus_enabled = st.checkbox("启用 Prometheus")
        
        col1, col2 = st.columns(2)
        with col1:
            prometheus_url = st.text_input("Prometheus URL", placeholder="http://prometheus.example.com:9090")
            prometheus_username = st.text_input("用户名（可选）")
        
        with col2:
            prometheus_password = st.text_input("密码（可选）", type="password")
            prometheus_token = st.text_input("Token（可选）", type="password")
        
        st.subheader("Kubeconfig")
        st.caption("粘贴 kubeconfig 内容，用于资源巡检")
        
        kubeconfig_content = st.text_area("Kubeconfig 内容", height=150)
        
        submitted = st.form_submit_button("保存集群")
        
        if submitted:
            if not cluster_name:
                st.error("请输入集群名称")
            elif not node_ip:
                st.error("请输入至少一个节点 IP")
            else:
                # 创建新集群
                cluster_config = get_cluster(cluster_name)
                
                # 添加节点信息
                node_info = {
                    "ip": node_ip,
                    "port": node_port,
                    "username": node_username,
                    "auth_type": auth_type
                }
                
                if auth_type == "password":
                    node_info["password"] = node_password
                else:
                    node_info["key_path"] = node_key_path
                
                # 测试节点连接
                with st.spinner("正在测试节点连接..."):
                    success, message = test_node_connection(node_info)
                
                if not success:
                    st.error(f"节点连接失败: {message}")
                else:
                    # 添加节点配置
                    cluster_config.update_node(node_info)
                    
                    # 更新 Prometheus 配置
                    prometheus_config = {
                        "url": prometheus_url,
                        "username": prometheus_username,
                        "password": prometheus_password,
                        "token": prometheus_token,
                        "enabled": prometheus_enabled
                    }
                    cluster_config.update_prometheus(prometheus_config)
                    
                    # 更新 kubeconfig
                    if kubeconfig_content:
                        cluster_config.update_kubeconfig(kubeconfig_content)
                    
                    st.success(f"集群 {cluster_name} 已成功添加")
                    st.info("你可以在「编辑集群」选项卡中添加更多节点")

# 编辑集群选项卡
with tab3:
    st.header("编辑集群")
    
    # 加载集群列表
    clusters = list_clusters()
    
    if not clusters:
        st.info("还没有配置任何集群，请前往「添加集群」选项卡添加集群。")
    else:
        selected_cluster = st.selectbox("选择要编辑的集群", clusters)
        
        if selected_cluster:
            cluster_config = get_cluster(selected_cluster)
            
            st.subheader(f"编辑集群: {selected_cluster}")
            
            # 编辑选项卡
            edit_tab1, edit_tab2, edit_tab3 = st.tabs(["节点管理", "Prometheus 配置", "Kubeconfig"])
            
            # 节点管理选项卡
            with edit_tab1:
                st.subheader("节点管理")
                
                # 显示现有节点
                nodes = cluster_config.get_nodes()
                
                if nodes:
                    st.write("现有节点:")
                    
                    for i, node in enumerate(nodes):
                        with st.expander(f"节点: {node['ip']}", expanded=False):
                            st.json(node)
                            
                            if st.button("删除节点", key=f"delete_node_{i}"):
                                cluster_config.remove_node(node['ip'])
                                st.success(f"节点 {node['ip']} 已删除")
                                st.rerun()
                
                # 添加新节点
                st.write("添加新节点:")
                
                with st.form("add_node_form"):
                    col1, col2 = st.columns(2)
                    with col1:
                        node_ip = st.text_input("节点 IP", placeholder="192.168.1.101", key="edit_node_ip")
                        node_port = st.text_input("SSH 端口", "22", key="edit_node_port")
                        node_username = st.text_input("用户名", "root", key="edit_node_username")
                    
                    with col2:
                        auth_type = st.selectbox("认证方式", ["password", "key"], key="edit_auth_type")
                        
                        if auth_type == "password":
                            node_password = st.text_input("密码", type="password", key="edit_node_password")
                            node_key_path = ""
                        else:
                            node_password = ""
                            node_key_path = st.text_input("密钥路径", placeholder="/home/user/.ssh/id_rsa", key="edit_node_key")
                    
                    submitted = st.form_submit_button("添加节点")
                    
                    if submitted:
                        if not node_ip:
                            st.error("请输入节点 IP")
                        else:
                            # 构建节点信息
                            node_info = {
                                "ip": node_ip,
                                "port": node_port,
                                "username": node_username,
                                "auth_type": auth_type
                            }
                            
                            if auth_type == "password":
                                node_info["password"] = node_password
                            else:
                                node_info["key_path"] = node_key_path
                            
                            # 测试节点连接
                            with st.spinner("正在测试节点连接..."):
                                success, message = test_node_connection(node_info)
                            
                            if not success:
                                st.error(f"节点连接失败: {message}")
                            else:
                                # 添加节点配置
                                cluster_config.update_node(node_info)
                                st.success(f"节点 {node_ip} 已添加")
                                st.rerun()
            
            # Prometheus 配置选项卡
            with edit_tab2:
                st.subheader("Prometheus 配置")
                
                # 获取现有配置
                prometheus_config = cluster_config.get_prometheus_config()
                
                with st.form("edit_prometheus_form"):
                    prometheus_enabled = st.checkbox("启用 Prometheus", value=prometheus_config.get('enabled', False))
                    
                    col1, col2 = st.columns(2)
                    with col1:
                        prometheus_url = st.text_input("Prometheus URL", value=prometheus_config.get('url', ''))
                        prometheus_username = st.text_input("用户名", value=prometheus_config.get('username', ''))
                    
                    with col2:
                        prometheus_password = st.text_input("密码", type="password", value=prometheus_config.get('password', ''))
                        prometheus_token = st.text_input("Token", type="password", value=prometheus_config.get('token', ''))
                    
                    # 测试连接按钮
                    test_prom_button = st.form_submit_button("测试连接")
                    
                    # 保存配置按钮
                    save_prom_button = st.form_submit_button("保存配置")
                    
                    if test_prom_button:
                        if not prometheus_url:
                            st.error("请输入 Prometheus URL")
                        else:
                            # 构建临时配置
                            test_config = {
                                "url": prometheus_url,
                                "username": prometheus_username,
                                "password": prometheus_password,
                                "token": prometheus_token,
                                "enabled": True
                            }
                            
                            # 测试连接
                            with st.spinner("正在测试连接..."):
                                client = PrometheusClient(test_config)
                                result = client.test_connection()
                            
                            if result.get('status') == 'success':
                                st.success("连接成功")
                            else:
                                st.error(f"连接失败: {result.get('error', '未知错误')}")
                    
                    if save_prom_button:
                        # 更新 Prometheus 配置
                        new_config = {
                            "url": prometheus_url,
                            "username": prometheus_username,
                            "password": prometheus_password,
                            "token": prometheus_token,
                            "enabled": prometheus_enabled
                        }
                        
                        cluster_config.update_prometheus(new_config)
                        st.success("Prometheus 配置已更新")
            
            # Kubeconfig 选项卡
            with edit_tab3:
                st.subheader("Kubeconfig 配置")
                
                # 获取现有配置
                current_kubeconfig = cluster_config.get_kubeconfig()
                
                with st.form("edit_kubeconfig_form"):
                    kubeconfig_content = st.text_area("Kubeconfig 内容", value=current_kubeconfig, height=300)
                    
                    # 测试连接按钮
                    test_kube_button = st.form_submit_button("测试连接")
                    
                    # 保存配置按钮
                    save_kube_button = st.form_submit_button("保存配置")
                    
                    if test_kube_button:
                        if not kubeconfig_content:
                            st.error("请输入 kubeconfig 内容")
                        else:
                            # 测试连接
                            with st.spinner("正在测试连接..."):
                                client = K8sClient(kubeconfig_content)
                                success, message = client.test_connection()
                            
                            if success:
                                st.success("连接成功")
                            else:
                                st.error(f"连接失败: {message}")
                    
                    if save_kube_button:
                        # 更新 kubeconfig
                        cluster_config.update_kubeconfig(kubeconfig_content)
                        st.success("Kubeconfig 配置已更新")