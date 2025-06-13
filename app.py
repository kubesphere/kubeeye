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
from utils.inspection_result import list_results
from utils.rule_loader import load_rules
from utils.version import VERSION, APP_NAME, APP_DESCRIPTION, RELEASE_DATE

# 页面配置已经在上面设置完成，现在初始化其他页面组件
initialize_page(
    title="首页",
    icon="🏠",
    page_title="KubeEye",
    page_subtitle="Kubernetes 集群巡检工具"
)

# 加载集群列表
clusters = list_clusters()

# 加载规则数据 - 所有规则现在都使用统一的断言格式
node_rules = load_rules('node')
prometheus_rules = load_rules('prometheus')
opa_rules = load_rules('opa')

# 计算总规则数
total_rules = len(node_rules) + len(prometheus_rules) + len(opa_rules)

# 断言系统已经是默认系统
using_assertion_system = True

# 设置强调样式，增强视觉效果，更加和谐的色彩
st.markdown("""
<style>
.highlight {
    padding: 1.2rem;
    border-radius: 0.5rem;
    background-color: #f8f9fa;
    border-left: 3px solid #00a971;
    margin-bottom: 1.5rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.03);
}
.stat-card {
    background-color: #ffffff;
    border-radius: 0.5rem;
    padding: 1.2rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.04);
    text-align: center;
    transition: all 0.25s ease;
    border: 1px solid rgba(0,0,0,0.03);
}
.stat-card:hover {
    transform: translateY(-3px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}
.stat-number {
    font-size: 2.2rem;
    font-weight: 600;
    color: #00a971;
    margin-bottom: 0.3rem;
}
.stat-label {
    color: #555;
    font-size: 0.95rem;
}
/* 改进文本可读性 */
p, li {
    color: #444;
    line-height: 1.6;
}
h2, h3 {
    margin-top: 1.5rem;
    color: #333;
}
small {
    color: #777;
}
</style>
""", unsafe_allow_html=True)

# 首页功能区域
st.markdown(f"""
## 欢迎使用 {APP_NAME} {VERSION}

### 系统概述

**{APP_NAME}** 是一款专为 Kubernetes 集群设计的综合巡检工具，旨在帮助运维人员和 SRE 团队快速发现、诊断和解决集群中的各类问题。通过自动化的检查流程，本工具可以显著提高集群的可靠性、安全性和性能。

<small>版本发布日期: {RELEASE_DATE}</small>

### 核心功能

本工具提供三大类巡检功能，全面覆盖集群各层面的健康状况：

1. **节点状态检查**
   - 监控 CPU、内存、磁盘使用率，及时预警资源不足
   - 检查关键系统服务（kubelet、docker、containerd）的运行状态
   - 分析系统负载，识别性能瓶颈
   - 验证节点网络连通性和内核参数配置

2. **Prometheus 指标检查**
   - 通过 Prometheus 指标分析集群关键性能数据
   - 监控容器重启次数、Pod 状态异常等关键指标
   - 追踪资源使用趋势，提前预警潜在问题
   - 支持自定义指标查询和阈值设置

3. **OPA 合规性检查**
   - 基于 OPA（Open Policy Agent）验证资源配置合规性
   - 检查安全风险，如特权容器、不安全挂载等
   - 验证资源配置最佳实践（副本数、资源限制等）
   - 确保集群配置符合组织安全策略和行业标准

### 灵活的规则配置

- **基于 YAML 的规则定义**：简单易读，方便版本控制和分享
- **当前已配置 {total_rules} 条规则**：覆盖常见的集群问题和最佳实践
- **规则管理界面**：通过 UI 直接编辑和创建规则，无需编写代码
- **分级严重性**：问题按严重程度分类（关键、警告、信息），便于优先处理
- **解决方案建议**：智能提供针对性的问题解决建议

### 使用流程

1. **配置集群信息**：添加您需要巡检的 Kubernetes 集群信息
2. **选择巡检规则**：根据需求选择要执行的巡检规则
3. **执行巡检**：系统自动执行所选规则并收集结果
4. **查看分析报告**：获取直观的巡检报告，包含问题详情和解决建议
5. **导出分享**：支持导出报告为多种格式，方便团队协作和问题追踪

### 导航指南

请使用左侧导航栏访问各功能页面：

- **集群信息**：管理集群连接配置，包括节点 SSH、Prometheus 和 Kubeconfig 设置
- **集群巡检**：执行巡检任务并管理巡检规则
- **巡检报告**：查看详细的巡检结果，支持历史记录查询和数据可视化

开始使用本工具，确保您的 Kubernetes 集群始终处于最佳状态！
""")

# 显示统计信息卡片
st.markdown('<div class="highlight">', unsafe_allow_html=True)
cols = st.columns(4)

# 集群数统计卡片
with cols[0]:
    st.markdown(f"""
    <div class="stat-card">
        <div class="stat-number">{len(clusters)}</div>
        <div class="stat-label">已配置集群</div>
    </div>
    """, unsafe_allow_html=True)

# 规则数统计卡片
with cols[1]:
    st.markdown(f"""
    <div class="stat-card">
        <div class="stat-number">{total_rules}</div>
        <div class="stat-label">可用规则</div>
    </div>
    """, unsafe_allow_html=True)

# 节点规则统计卡片
with cols[2]:
    st.markdown(f"""
    <div class="stat-card">
        <div class="stat-number">{len(node_rules)}</div>
        <div class="stat-label">节点检查规则</div>
    </div>
    """, unsafe_allow_html=True)

# OPA规则统计卡片
with cols[3]:
    st.markdown(f"""
    <div class="stat-card">
        <div class="stat-number">{len(opa_rules)}</div>
        <div class="stat-label">OPA检查规则</div>
    </div>
    """, unsafe_allow_html=True)

st.markdown('</div>', unsafe_allow_html=True)

# 分栏显示主要内容
col1, col2 = st.columns(2)

# 显示集群信息
with col1:
    st.subheader("已配置的集群")
    
    if clusters:
        cluster_data = []
        
        for cluster_name in clusters:
            cluster_config = get_cluster(cluster_name)
            nodes_count = len(cluster_config.get_nodes())
            prometheus_enabled = "✅" if cluster_config.get_prometheus_config().get('enabled', False) else "❌"
            kubeconfig = "✅" if cluster_config.get_kubeconfig() else "❌"
            
            cluster_data.append({
                "集群名称": cluster_name,
                "节点数量": nodes_count,
                "Prometheus": prometheus_enabled,
                "Kubeconfig": kubeconfig
            })
        
        st.dataframe(pd.DataFrame(cluster_data), use_container_width=True)
    else:
        st.info("还没有配置任何集群，请前往「集群信息」页面添加集群。")
        
    st.markdown("---")
    
    # 快速链接
    st.subheader("快速操作")
    
    col1_1, col1_2, col1_3 = st.columns(3)
    
    with col1_1:
        if st.button("➕ 添加新集群", use_container_width=True):
            st.switch_page("pages/1_cluster_info.py")
    
    with col1_2:
        if st.button("🔍 执行巡检", use_container_width=True):
            st.switch_page("pages/2_cluster_scan.py")
    
    with col1_3:
        if st.button("📊 查看报告", use_container_width=True):
            st.switch_page("pages/3_scan_report.py")
            
    st.info("💡 **小贴士**: 定期执行巡检可以帮助您提前发现潜在问题，建议每周至少进行一次全面巡检。")

# 显示最近的巡检结果
with col2:
    st.subheader("最近巡检结果")
    
    # 获取最近的10条巡检结果
    results = list_results()
    
    if results:
        # 取前10个结果
        recent_results = results[:10]
        
        result_data = []
        for result in recent_results:
            timestamp = datetime.fromisoformat(result['timestamp']).strftime("%Y-%m-%d %H:%M")
            
            result_data.append({
                "集群名称": result['cluster_name'],
                "巡检类型": result['inspection_type'],
                "时间": timestamp,
                "关键问题": result['critical'],
                "警告": result['warning'],
                "通过": result['passed']
            })
        
        st.dataframe(pd.DataFrame(result_data), use_container_width=True)
    else:
        st.info("还没有执行过任何巡检，请前往「集群巡检」页面执行巡检。")

# 页面底部
st.markdown("---")

# 创建两列布局
footer_col1, footer_col2 = st.columns(2)

with footer_col1:
    st.markdown("""
    ### 巡检最佳实践
    - **定期执行**: 建议每周进行一次完整巡检，确保集群稳定
    - **变更后检查**: 集群重大变更后应立即执行相关巡检
    - **问题追踪**: 对发现的问题建立跟踪机制，确保及时修复
    - **规则迭代**: 根据运维经验不断优化和丰富巡检规则
    """)

with footer_col2:
    st.markdown("""
    ### 项目说明
    kubeeye 旨在提供全面的 Kubernetes 集群巡检解决方案，帮助管理员维护健康、安全的集群环境。通过自动化检查和详细报告，大幅降低运维成本，提升集群可靠性。
    
    持续开发中，欢迎提供功能建议和使用反馈，共同改进本工具。
    """)
    
    # 显示断言系统信息
    if 'using_assertion_system' in locals() and using_assertion_system:
        st.info("✨ **新功能**: 基于断言的规则系统已启用！了解更多请查看 [断言系统文档](/docs/assertion_system.md) 和 [迁移指南](/docs/rule_migration_guide.md)。")
