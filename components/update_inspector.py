#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
系统信息组件 - 显示当前系统版本和配置信息
"""
import streamlit as st
from utils.version import get_version

def render_system_info_tab():
    """渲染系统信息选项卡"""
    st.subheader("系统信息")
    
    version = get_version()
    st.write(f"当前版本: {version}")
    
    # 显示当前系统信息
    st.success("系统正常运行中")
    
    # 添加系统信息
    from datetime import datetime
    st.write(f"当前时间: {datetime.now().strftime('%Y-%m-%d %H:%M')}")
    st.write(f"最后更新: 2025-06-02")
