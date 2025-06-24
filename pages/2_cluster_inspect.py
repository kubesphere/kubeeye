#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 集群巡检执行页面
"""

# 导入必要的库
import streamlit as st
import sys
from pathlib import Path

# 设置页面配置 - 必须是第一个Streamlit命令
st.set_page_config(
    page_title="集群巡检 - kubeeye",
    page_icon="🔍",
    layout="wide"
)

# 添加项目根目录到Python路径
ROOT_DIR = Path(__file__).resolve().parent.parent
if str(ROOT_DIR) not in sys.path:
    sys.path.insert(0, str(ROOT_DIR))

# 导入工具模块
from utils.common import initialize_page
# 导入组件模块
from components.immediate_scan import render_immediate_scan_tab
from components.scheduled_scan import render_scheduled_scan_tab
from components.rule_management import render_rule_management_tab

# 初始化页面
initialize_page(
    title="集群巡检",
    icon="🔍",
    page_title="集群巡检中心", 
    page_subtitle="执行立即或定时巡检，管理巡检规则"
)

# 创建三个选项卡
tab1, tab2, tab3 = st.tabs(["立即巡检", "定时巡检", "规则管理"])

# 渲染立即巡检选项卡
with tab1:
    render_immediate_scan_tab()

# 渲染定时巡检选项卡
with tab2:
    render_scheduled_scan_tab()

# 渲染规则管理选项卡
with tab3:
    render_rule_management_tab()