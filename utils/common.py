#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
公共组件和辅助函数 - 减少页面间的重复代码
"""

import streamlit as st
import sys
from pathlib import Path

# 导入项目模块
from utils.navbar import set_app_styles, show_app_logo, create_sidebar_header, create_page_header

# 全局变量，确保只初始化一次
_background_services_initialized = False

def _initialize_background_services():
    """
    初始化后台服务（定时任务、数据清理等）
    使用全局变量确保只初始化一次
    """
    global _background_services_initialized
    
    if _background_services_initialized:
        return
    
    try:
        # 导入并启动数据清理模块
        import utils.data_cleanup
        
        # 导入并启动定时任务调度器
        import utils.schedule_manager
        
        _background_services_initialized = True
        
    except ImportError as e:
        # 如果模块不存在，记录错误但不影响页面加载
        pass
    except Exception as e:
        # 其他错误也不影响页面加载
        pass

def initialize_page(title, icon="🔍", sidebar_name="", page_title="", page_subtitle="", page_icon=""):
    """
    初始化页面设置，包括样式和导航栏
    注意: 此函数假设 st.set_page_config() 已经在调用此函数之前被调用
    
    Args:
        title: 页面标题
        icon: 页面图标
        sidebar_name: 侧边栏名称
        page_title: 页面标题（如果与title不同）
        page_subtitle: 页面副标题
        page_icon: 页面图标（如果与icon不同）
    
    Returns:
        None
    """
    # 初始化后台服务（只在第一次调用时执行）
    _initialize_background_services()
    
    # 不再调用 st.set_page_config() - 必须在使用此函数之前调用
    
    # 确保项目根目录在Python路径中
    ROOT_DIR = Path(__file__).resolve().parent.parent
    if str(ROOT_DIR) not in sys.path:
        sys.path.insert(0, str(ROOT_DIR))
    
    # 设置应用样式
    set_app_styles()
    
    # 显示应用Logo
    show_app_logo()
    
    # 创建侧边栏和页面标题
    create_sidebar_header(sidebar_name or title)
    create_page_header(page_title or title, page_subtitle, icon=page_icon or icon)

def create_status_badge(status, text=None):
    """
    创建状态徽章
    
    Args:
        status: 状态类型 ('success', 'warning', 'error', 'info')
        text: 显示文本，如果为None则使用状态本身
    
    Returns:
        str: HTML 徽章代码
    """
    if text is None:
        text = status.title()
        
    colors = {
        'success': ('#E7F9ED', '#1E8E3E'),  # 浅绿色背景，深绿色文本
        'warning': ('#FEF7E0', '#E67700'),  # 浅黄色背景，橙色文本
        'error': ('#FFE5E5', '#D93025'),    # 浅红色背景，红色文本
        'info': ('#E8F0FE', '#1A73E8')      # 浅蓝色背景，蓝色文本
    }
    
    bg_color, text_color = colors.get(status.lower(), colors['info'])
    
    return f"""
    <span style="
        background-color: {bg_color}; 
        color: {text_color}; 
        padding: 4px 8px; 
        border-radius: 4px; 
        font-size: 0.8rem; 
        font-weight: 500;
    ">{text}</span>
    """
