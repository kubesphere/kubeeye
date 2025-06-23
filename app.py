#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
KubeEye - Kubernetes 集群巡检工具主页

此文件是应用程序的入口点，初始化界面并显示主页内容。
"""

# 标准库导入
import os
import sys
import json
from datetime import datetime
from pathlib import Path

# 第三方库导入
import streamlit as st
import pandas as pd

# 设置页面配置 - 必须是第一个Streamlit命令
st.set_page_config(
    page_title="KubeEye - Kubernetes 集群巡检工具",
    page_icon="🔍",
    layout="wide"
)

# 项目模块导入
from utils.common import initialize_page
from utils.cluster_config import list_clusters, get_cluster
from utils.inspection_result import list_results, get_latest_result_by_cluster
from utils.rule_loader import load_rules
from utils.version import VERSION, APP_NAME, APP_DESCRIPTION, RELEASE_DATE
from utils.cert_checker import get_cluster_cert_status

# 页面配置已经在上面设置完成，现在初始化其他页面组件
initialize_page(
    title="巡检总览",
    icon="📊",
    page_title="KubeEye 集群巡检总览",
    page_subtitle="实时监控您的 Kubernetes 集群健康状况"
)

# 加载集群列表
clusters = list_clusters()

# 加载规则数据 - 所有规则现在都使用统一的断言格式
node_rules = load_rules('node')
prometheus_rules = load_rules('prometheus')
opa_rules = load_rules('opa')

# 计算总规则数
total_rules = len(node_rules) + len(prometheus_rules) + len(opa_rules)

# 设置简洁样式
st.markdown("""
<style>
.status-healthy { color: #28a745; font-weight: bold; }
.status-warning { color: #ffc107; font-weight: bold; }
.status-critical { color: #dc3545; font-weight: bold; }
.status-unknown { color: #6c757d; font-weight: bold; }
</style>
""", unsafe_allow_html=True)

# 获取数据
clusters = list_clusters()
all_results = list_results()

# 计算统计数据
total_clusters = len(clusters)

# 最近24小时的巡检结果
recent_scans = 0
recent_issues = 0

if all_results:
    # 计算最近24小时的数据
    now = datetime.now()
    for result in all_results:
        result_time = datetime.fromisoformat(result['timestamp'])
        if (now - result_time).total_seconds() < 24 * 3600:  # 24小时内
            recent_scans += 1
            recent_issues += result.get('critical', 0) + result.get('warning', 0)

# 获取每个集群的最新状态
cluster_statuses = {}
for cluster_name in clusters:
    latest_result = None
    for result in all_results:
        if result['cluster_name'] == cluster_name:
            latest_result = result
            break
    
    if latest_result:
        critical = latest_result.get('critical', 0)
        warning = latest_result.get('warning', 0)
        
        if critical > 0:
            status = 'critical'
        elif warning > 0:
            status = 'warning'
        else:
            status = 'healthy'
    else:
        status = 'unknown'
    
    cluster_statuses[cluster_name] = {
        'status': status,
        'latest_result': latest_result
    }

# 顶部概览统计
st.markdown("### 📊 概览")
cols = st.columns(5)

with cols[0]:
    st.metric(
        label="总集群数",
        value=total_clusters,
        delta=None
    )

with cols[1]:
    st.metric(
        label="24小时内巡检次数",
        value=recent_scans,
        delta=None
    )

with cols[2]:
    st.metric(
        label="发现的问题数",
        value=recent_issues,
        delta=None  
    )

with cols[3]:
    latest_scan_time = "从未执行"
    if all_results:
        latest_time = datetime.fromisoformat(all_results[0]['timestamp'])
        latest_scan_time = latest_time.strftime("%m-%d %H:%M")
    
    st.metric(
        label="最近巡检时间",
        value=latest_scan_time,
        delta=None
    )

with cols[4]:
    st.metric(
        label="巡检规则总数",
        value=total_rules,
        delta=None
    )

# 安全状态检查
try:
    from utils.command_security import CommandSecurityChecker
    security_checker = CommandSecurityChecker()
    
    # 测试危险命令和安全命令
    dangerous_safe, _, _ = security_checker.check_command_security("rm -rf /")
    safe_safe, _, _ = security_checker.check_command_security("ps aux")
    
    # 强制安全模式状态显示
    if not dangerous_safe and safe_safe:
        security_status = "� 强制安全模式已启用"
        security_color = "green"
    else:
        security_status = "🔴 安全检查器异常"
        security_color = "red"
        
    st.markdown(f"""
    <div style="text-align: center; margin: 15px 0; padding: 10px; 
                background-color: {'#d4edda' if security_color == 'green' else '#f8d7da'}; 
                border: 1px solid {'#c3e6cb' if security_color == 'green' else '#f5c6cb'}; 
                border-radius: 5px;">
        <span style="color: {security_color}; font-weight: bold; font-size: 1.1em;">
            {security_status}
        </span>
        <br>
        <small style="color: #666;">只读巡检 • 禁止修改操作 • 安全第一</small>
    </div>
    """, unsafe_allow_html=True)
except Exception as e:
    st.markdown(f'''
    <div style="text-align: center; color: orange; padding: 10px; 
                background-color: #fff3cd; border: 1px solid #ffeaa7; border-radius: 5px;">
        ⚠️ 安全检查器状态未知: {str(e)}
    </div>
    ''', unsafe_allow_html=True)

st.markdown("---")

# 主要内容区域 - 集群状态表格
st.markdown("### 🏗️ 集群状态详情")

if not clusters:
    st.warning("📝 还没有配置任何集群，请先前往「集群信息」页面添加集群配置。")
    if st.button("➕ 立即添加集群", type="primary"):
        st.switch_page("pages/1_cluster_info.py")
else:
    # 构建集群状态表格数据
    cluster_data = []
    
    for cluster_name in clusters:
        status_info = cluster_statuses[cluster_name]
        status = status_info['status']
        latest_result = status_info['latest_result']
        
        # 获取集群配置
        cluster_config = get_cluster(cluster_name)
        nodes_count = len(cluster_config.get_nodes())
        
        # 检查证书状态
        kubeconfig = cluster_config.get_kubeconfig()
        cert_status_info = {'status': 'unknown', 'days_remaining': None}
        if kubeconfig:
            cert_status_info = get_cluster_cert_status(cluster_name, kubeconfig)
        
        # 状态显示
        status_icons = {
            'healthy': '✅ 健康',
            'warning': '⚠️ 警告',
            'critical': '❌ 异常',
            'unknown': '❓ 未知'
        }
        
        cert_status_text = {
            'valid': '✅ 正常',
            'warning': '⚠️ 即将过期',
            'critical': '🔴 临近过期',
            'expired': '❌ 已过期',
            'unknown': '❓ 未知'
        }
        
        cert_status = cert_status_info['status']
        days_remaining = cert_status_info.get('days_remaining')
        
        cert_display = cert_status_text.get(cert_status, '❓ 未知')
        if days_remaining is not None and days_remaining >= 0:
            cert_display += f" ({days_remaining}天)"
        elif days_remaining is not None and days_remaining < 0:
            cert_display += f" (过期{abs(days_remaining)}天)"
        
        # 最近巡检时间
        last_scan = "从未巡检"
        if latest_result:
            scan_time = datetime.fromisoformat(latest_result['timestamp'])
            last_scan = scan_time.strftime("%m-%d %H:%M")
        
        # 巡检结果统计
        critical_count = latest_result.get('critical', 0) if latest_result else 0
        warning_count = latest_result.get('warning', 0) if latest_result else 0
        passed_count = latest_result.get('passed', 0) if latest_result else 0
        
        cluster_data.append({
            "集群名称": f"**{cluster_name}**",
            "状态": status_icons[status],
            "节点数": nodes_count,
            "kubeconfig 有效期": cert_display,
            "最近巡检": last_scan,
            "关键问题": critical_count,
            "警告": warning_count,
            "通过": passed_count
        })
    
    # 显示集群状态表格
    cluster_df = pd.DataFrame(cluster_data)
    
    def color_status(val):
        if '✅' in str(val):
            return 'color: #28a745; font-weight: bold'
        elif '⚠️' in str(val):
            return 'color: #ffc107; font-weight: bold'
        elif '❌' in str(val) or '🔴' in str(val):
            return 'color: #dc3545; font-weight: bold'
        else:
            return 'color: #6c757d'
    
    def color_numbers(val):
        if val > 0:
            return 'color: #dc3545; font-weight: bold'
        return ''
    
    styled_df = cluster_df.style.applymap(color_status, subset=['状态', 'kubeconfig 有效期']) \
                               .applymap(color_numbers, subset=['关键问题', '警告'])
    
    st.dataframe(styled_df, use_container_width=True, hide_index=True)
    
    # 快速操作按钮
    st.markdown("#### 🚀 快速操作")
    cols = st.columns(3)
    
    with cols[0]:
        if st.button("🔍 执行巡检", use_container_width=True, type="primary"):
            st.switch_page("pages/2_cluster_scan.py")
    
    with cols[1]:
        if st.button("📊 查看报告", use_container_width=True):
            st.switch_page("pages/3_scan_report.py")
    
    with cols[2]:
        if st.button("⚙️ 管理集群", use_container_width=True):
            st.switch_page("pages/1_cluster_info.py")

# 最近巡检记录表格
st.markdown("### 📈 最近巡检记录")

if all_results:
    # 构建巡检记录表格数据
    recent_results = all_results[:10]  # 最近10条记录
    
    scan_records = []
    for result in recent_results:
        timestamp = datetime.fromisoformat(result['timestamp'])
        time_str = timestamp.strftime("%Y-%m-%d %H:%M")
        
        # 计算状态
        critical = result.get('critical', 0)
        warning = result.get('warning', 0)
        passed = result.get('passed', 0)
        
        if critical > 0:
            status = '❌ 异常'
        elif warning > 0:
            status = '⚠️ 警告'
        else:
            status = '✅ 正常'
        
        scan_records.append({
            "时间": time_str,
            "集群": f"**{result['cluster_name']}**",
            "类型": result['inspection_type'],
            "状态": status,
            "关键问题": critical,
            "警告": warning,
            "通过": passed
        })
    
    scan_df = pd.DataFrame(scan_records)
    
    def color_scan_status(val):
        if '✅' in str(val):
            return 'color: #28a745; font-weight: bold'
        elif '⚠️' in str(val):
            return 'color: #ffc107; font-weight: bold'
        elif '❌' in str(val):
            return 'color: #dc3545; font-weight: bold'
        return ''
    
    def color_scan_numbers(val):
        if val > 0:
            return 'color: #dc3545; font-weight: bold'
        return ''
    
    styled_scan_df = scan_df.style.applymap(color_scan_status, subset=['状态']) \
                                  .applymap(color_scan_numbers, subset=['关键问题', '警告'])
    
    st.dataframe(styled_scan_df, use_container_width=True, hide_index=True)
else:
    st.info("📋 还没有巡检记录，执行首次巡检后这里将显示历史记录。")

# 页面底部信息
st.markdown("---")
st.markdown(f"""
<div style="text-align: center; color: #666; font-size: 0.85rem; padding: 1rem;">
    KubeEye {VERSION} | 
    <a href="#" onclick="window.location.reload()">刷新页面</a> | 
    最后更新: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
</div>
""", unsafe_allow_html=True)
