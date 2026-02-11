#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 集群信息管理页面
"""

# 导入必要的库
import time
import streamlit as st
import pandas as pd
import sys
from pathlib import Path
from typing import Union,List

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

def parse_labels_string(labels_str: str) -> dict:
    """
    将标签字符串转换为字典格式
    
    Args:
        labels_str: 标签字符串，格式如 "key1=value1,key2=value2"
        
    Returns:
        标签字典
    """
    if not labels_str or not labels_str.strip():
        return {}
    
    labels = {}
    for label_pair in labels_str.split(','):
        label_pair = label_pair.strip()
        if '=' in label_pair:
            key, value = label_pair.split('=', 1)  # 使用 maxsplit=1 处理值中包含=的情况
            labels[key.strip()] = value.strip()
    return labels

def format_labels_dict(labels_dict: dict) -> str:
    """
    将标签字典转换为字符串格式
    
    Args:
        labels_dict: 标签字典
        
    Returns:
        标签字符串，格式如 "key1=value1,key2=value2"
    """
    if not labels_dict:
        return ""
    return ",".join([f"{k}={v}" for k, v in labels_dict.items()])

def ensure_node_state(node_id: Union[str, int]):
    """
    确保一个节点所需的所有 session_state key 都存在
    只初始化，不做任何业务判断
    """
    defaults = {
        f"node_ip_{node_id}": "",
        f"node_port_{node_id}": "22",
        f"node_username_{node_id}": "root",
        f"node_labels_{node_id}": "",
        f"auth_type_{node_id}": "password",
        f"select_auth_type_{node_id}": "password",
        f"node_password_{node_id}": "",
        f"node_key_{node_id}": "",
    }
    for k, v in defaults.items():
        if k not in st.session_state:
            st.session_state[k] = v

def reset_auth_related_state(node_id: Union[str, int], auth_type: str):
    """
    重置认证方式相关的会话状态，确保切换时无残留
    
    Args:
        node_id: 节点的唯一标识（数字/字符串）
        auth_type: 目标认证方式（password/key_path）
    """
    # 定义需要清空的键
    password_key = f"node_password_{node_id}"
    key_path_key = f"node_key_{node_id}"
    
    # 清空所有非目标认证方式的状态
    if auth_type == "password":
        if key_path_key in st.session_state:
            del st.session_state[key_path_key]
    elif auth_type == "key_path":
        if password_key in st.session_state:
            del st.session_state[password_key]

def render_auth_fields(
    *,
    node_id: Union[str, int],
    label_prefix: str = "认证方式"
):
    """
    渲染 SSH 认证方式选择 + 对应输入框
    仅负责 UI + session_state 同步，不负责校验、不负责提交
    """
    auth_type_key = f"auth_type_{node_id}"
    select_key = f"select_auth_type_{node_id}"
    password_key = f"node_password_{node_id}"
    key_path_key = f"node_key_{node_id}"

    # 初始化 auth_type
    ensure_node_state(node_id)  # 统一初始化，无需单独判断

    def on_auth_change():
        new_auth = st.session_state[select_key]
        st.session_state[auth_type_key] = new_auth
        reset_auth_related_state(node_id, new_auth)

    # 认证方式选择
    auth_type = st.selectbox(
        label_prefix,
        ["password", "key_path"],
        key=select_key,
        index=["password", "key_path"].index(st.session_state[auth_type_key]),
        on_change=on_auth_change
    )
    st.session_state[auth_type_key] = auth_type

    # 根据认证方式渲染输入框
    if auth_type == "password":
        st.text_input(
            "密码",
            type="password",
            key=password_key,
            value=st.session_state[password_key]
        )
    else:
        st.text_input(
            "私钥路径",
            placeholder="~/.ssh/id_rsa",
            key=key_path_key,
            value=st.session_state[key_path_key]
        )

def clear_node_state(node_ids: List[Union[str, int]]):
    """
    批量清空节点相关的会话状态（统一清理逻辑）
    
    Args:
        node_ids: 需要清空的节点ID列表
    """
    for node_id in node_ids:
        for suffix in ["ip", "port", "username", "labels", "password", "key", "auth_type", "select_auth_type"]:
            key = f"node_{suffix}_{node_id}"
            if key in st.session_state:
                del st.session_state[key]

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
                        labels_str = format_labels_dict(node.get('labels', {}))
                        node_data.append({
                            "IP": node['ip'],
                            "端口": node['port'],
                            "用户名": node['username'],
                            "认证类型": node['auth_type'],
                            "标签": labels_str if labels_str else "无"
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
    
    # 初始化节点列表
    if 'add_cluster_nodes' not in st.session_state:
        st.session_state.add_cluster_nodes = [0]  # 至少有一个节点输入框
    
    cluster_name = st.text_input("集群名称", placeholder="production")
    st.caption("请输入一个唯一的集群名称，用于标识此集群")
    
    st.subheader("节点信息")
    st.caption("添加集群节点信息，用于对节点进行巡检")
    # 节点输入区域（表单外，实现实时切换）
    nodes_container = st.container()
    with nodes_container:
        for idx, node_id in enumerate(st.session_state.add_cluster_nodes):
            # 统一初始化节点状态（无需单独调用 reset_auth_related_state）
            ensure_node_state(node_id)
            
            with st.container(border=True):
                st.markdown(f"**节点 {idx + 1}**")
                node_ip_key = f"node_ip_{node_id}"
                node_port_key = f"node_port_{node_id}"
                node_username_key = f"node_username_{node_id}"
                node_labels_key = f"node_labels_{node_id}"
                
                col1, col2 = st.columns(2)
                with col1:
                    node_ip = st.text_input(
                        "节点 IP", 
                        placeholder="192.168.1.100",
                        key=node_ip_key,
                        value=st.session_state[node_ip_key]
                    )
                    node_port = st.text_input(
                        "SSH 端口", 
                        key=node_port_key
                    )
                    node_username = st.text_input(
                        "用户名", 
                        key=node_username_key
                    )
                
                with col2:
                    render_auth_fields(node_id=node_id)
                    node_labels = st.text_input(
                        "节点标签",
                        placeholder="kubernetes.io/arch=amd64,kubernetes.io/os=linux",
                        key=node_labels_key,
                        value=st.session_state[node_labels_key],
                        help="输入节点标签，格式：key1=value1,key2=value2"
                    )
        
        # 添加节点按钮（简化参数）
        if st.button("➕ 添加节点"):
            new_id = max(st.session_state.add_cluster_nodes) + 1 if st.session_state.add_cluster_nodes else 0
            st.session_state.add_cluster_nodes.append(new_id)
            ensure_node_state(new_id)  # 初始化新节点状态
            st.rerun()

    # 表单：仅包含 Prometheus、Kubeconfig 和提交按钮
    with st.form("add_cluster_form"):
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
        # 收集所有节点信息（从会话状态读取）
        nodes_to_add = []
        for node_id in st.session_state.add_cluster_nodes:
            node_ip_key = f"node_ip_{node_id}"
            node_ip = st.session_state.get(node_ip_key, "")
            
            if node_ip:  # 只有填写了IP的节点才会被收集
                node_port = st.session_state.get(f"node_port_{node_id}", "22")
                node_username = st.session_state.get(f"node_username_{node_id}", "root")
                auth_type = st.session_state.get(f"auth_type_{node_id}", "password")
                node_labels = st.session_state.get(f"node_labels_{node_id}", "")

                node_info = {
                    "ip": node_ip,
                    "port": node_port or "22",
                    "username": node_username or "root",
                    "auth_type": auth_type,
                    "labels": parse_labels_string(node_labels)
                }
                
                # 根据认证方式添加密码或密钥路径
                if auth_type == "password":
                    node_info["password"] = st.session_state.get(f"node_password_{node_id}", "")
                else:
                    node_info["key_path"] = st.session_state.get(f"node_key_{node_id}", "")
                nodes_to_add.append(node_info)
        
        # 验证输入
        if not cluster_name:
            st.error("请输入集群名称")
        elif not nodes_to_add:
            st.error("请至少添加一个节点（需要填写节点IP）")
        else:
            # 创建新集群
            cluster_config = get_cluster(cluster_name)
            
            # 批量测试节点连接
            success_count = 0
            fail_count = 0
            failed_nodes = []
            
            progress_bar = st.progress(0)
            status_text = st.empty()
            
            # 先测试所有节点连接
            for i, node_info in enumerate(nodes_to_add):
                status_text.text(f"正在测试节点 {i+1}/{len(nodes_to_add)}: {node_info.get('ip', 'unknown')}")
                progress_bar.progress((i + 1) / len(nodes_to_add))
                
                # 测试节点连接
                success, message = test_node_connection(node_info)
                
                if success:
                    success_count += 1
                else:
                    fail_count += 1
                    failed_nodes.append({
                        "ip": node_info.get('ip', 'unknown'),
                        "error": message
                    })
            
            progress_bar.empty()
            status_text.empty()
            
            # 根据测试结果决定添加哪些节点（只添加成功的）
            nodes_to_save = []
            for node_info in nodes_to_add:
                node_ip = node_info.get('ip', '')
                # 检查这个节点是否测试成功
                is_success = not any(failed['ip'] == node_ip for failed in failed_nodes)
                if is_success:
                    nodes_to_save.append(node_info)
            
            # 将全量节点列表（只包含成功的节点）逐个调用 update_node
            for node_info in nodes_to_save:
                cluster_config.update_node(node_info)
            
            # 显示结果
            if success_count > 0:
                st.success(f"成功测试并添加 {success_count} 个节点")
            
            if fail_count > 0:
                st.warning(f"有 {fail_count} 个节点连接测试失败，未添加到配置中")
                for failed in failed_nodes:
                    st.error(f"节点 {failed['ip']}: {failed['error']}")
            
            if nodes_to_save:
                st.info(f"已添加 {len(nodes_to_save)} 个节点到集群配置")
            
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
            
            if success_count > 0:
                st.success(f"集群 {cluster_name} 已成功添加")
                clear_node_state(st.session_state.add_cluster_nodes)
                # 重置节点列表
                st.session_state.add_cluster_nodes = [0]
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
                            if st.button("删除节点", key=f"delete_node_{selected_cluster}_{i}"):
                                cluster_config.remove_node(node['ip'])
                                st.success(f"节点 {node['ip']} 已删除")
                                st.rerun()
                
                # 添加新节点
                st.write("添加新节点:")
                edit_node_id = f"edit_node_{selected_cluster}"
                ensure_node_state(edit_node_id)

                node_ip_key = f"node_ip_{edit_node_id}"
                node_port_key = f"node_port_{edit_node_id}"
                node_username_key = f"node_username_{edit_node_id}"
                auth_type_key = f"auth_type_{edit_node_id}"
                node_labels_key = f"node_labels_{edit_node_id}"

                edit_node_container = st.container(border=True)
                with edit_node_container:
                    col1, col2 = st.columns(2)
                    with col1:
                        edit_node_ip = st.text_input(
                            "节点 IP", 
                            placeholder="192.168.1.101",
                            key=node_ip_key,
                            value=st.session_state[node_ip_key]
                        )
                        edit_node_port = st.text_input(
                            "SSH 端口",
                            key=node_port_key
                        )
                        edit_node_username = st.text_input(
                            "用户名",
                            key=node_username_key
                        )
                    
                    with col2:
                        render_auth_fields(node_id=edit_node_id)
                        edit_node_labels = st.text_input(
                            "节点标签",
                            placeholder="kubernetes.io/arch=amd64,kubernetes.io/os=linux",
                            key=node_labels_key,
                            value=st.session_state[node_labels_key],
                            help="输入节点标签，格式：key1=value1,key2=value2"
                        )

                    st.divider()  # 分割线区分输入框和按钮
                    test_conn_btn = st.button(
                        "测试连接",
                        key=f"test_node_conn_{selected_cluster}",
                    )
                    add_node_btn = st.button(
                        "添加节点",
                        key=f"add_node_btn_{selected_cluster}",
                    )

                # ========== 测试连接按钮逻辑 ==========
                if test_conn_btn:
                    ip = st.session_state[node_ip_key]
                    if not ip:
                        st.error("请先输入节点 IP！")
                    else:
                        # 构建节点信息（仅用于测试）
                        node_info = {
                            "ip": ip,
                            "port": st.session_state[node_port_key] or "22",
                            "username": st.session_state[node_username_key] or "root",
                            "auth_type": st.session_state[auth_type_key],
                            "labels": parse_labels_string(st.session_state[node_labels_key])
                        }
                        
                        # 添加密码/密钥信息
                        if node_info["auth_type"] == "password":
                            node_info["password"] = st.session_state[f"node_password_{edit_node_id}"]
                        else:
                            node_info["key_path"] = st.session_state[f"node_key_{edit_node_id}"]
                        
                        # 测试节点连接（仅验证，不添加）
                        with st.spinner("正在测试节点连接..."):
                            success, message = test_node_connection(node_info)
                        if success:
                            st.success(f"✅ 节点 {ip} 连接成功！")
                        else:
                            st.error(f"❌ 节点 {ip} 连接失败：{message}")
                # ========== 添加节点按钮逻辑 ==========
                if add_node_btn:
                    ip = st.session_state[node_ip_key]
                    if not ip:
                        st.error("请先输入节点 IP！")
                    else:
                        # 构建节点信息
                        node_info = {
                            "ip": ip,
                            "port": st.session_state[node_port_key] or "22",
                            "username": st.session_state[node_username_key] or "root",
                            "auth_type": st.session_state[auth_type_key],
                            "labels": parse_labels_string(st.session_state[node_labels_key])
                        }
                        
                        # 添加密码/密钥信息
                        if node_info["auth_type"] == "password":
                            node_info["password"] = st.session_state[f"node_password_{edit_node_id}"]
                        else:
                            node_info["key_path"] = st.session_state[f"node_key_{edit_node_id}"]
                        
                        # 先测试连接，再添加
                        with st.spinner("正在测试节点连接..."):
                            success, message = test_node_connection(node_info)
                        
                        if not success:
                            st.error(f"❌ 节点连接失败，无法添加：{message}")
                        else:
                            # 添加节点配置
                            cluster_config.update_node(node_info)
                            st.success(f"✅ 节点 {ip} 已成功添加！")
                            # 清空编辑节点的输入状态
                            clear_node_state([edit_node_id])
                            time.sleep(1.5)
                            st.rerun()
            with edit_tab2:
                st.subheader("Prometheus 配置")
                
                # 获取现有配置
                prometheus_config = cluster_config.get_prometheus_config()
                
                with st.form(f"edit_prometheus_form_{selected_cluster}"):  # 加集群名避免冲突
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
                
                with st.form(f"edit_kubeconfig_form_{selected_cluster}"):  # 加集群名避免冲突
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