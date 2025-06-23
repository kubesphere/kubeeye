#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
导航栏组件 - 用于在所有页面创建一致的导航栏
"""

import streamlit as st
from pathlib import Path
from utils.version import VERSION

def set_app_styles():
    """
    设置应用基本样式
    
    Returns:
        None
    """
    st.markdown("""
    <style>
    /* 隐藏默认页面标题和页脚 */
    .main .block-container h1:first-child { display: none; }
    footer { visibility: hidden; }
    #MainMenu { visibility: hidden; }
    
    /* 改善整体界面和间距 */
    .main .block-container { padding-top: 1.5rem; }
    
    /* 侧边栏基础样式 */
    [data-testid="stSidebar"] {
        background-color: #f0f2f5;
        border-right: 1px solid rgba(0,0,0,0.05);
        resize: none !important;
        min-width: 244px !important;
        max-width: 244px !important;
        width: 244px !important;
    }
    
    /* 禁用侧边栏的拖拽调整功能 */
    [data-testid="stSidebar"] .css-1d391kg,
    [data-testid="stSidebar"] .css-1y4p8pa,
    [data-testid="stSidebar"] .css-1cypcdb {
        resize: none !important;
        min-width: 244px !important;
        max-width: 244px !important;
        width: 244px !important;
    }
    
    /* 隐藏侧边栏的拖拽手柄 */
    [data-testid="stSidebar"] .css-1d391kg::after,
    [data-testid="stSidebar"] .css-1y4p8pa::after,
    [data-testid="stSidebar"] .css-1cypcdb::after {
        display: none !important;
    }
    
    /* 禁用侧边栏右边缘的鼠标调整 */
    [data-testid="stSidebar"]:hover {
        cursor: default !important;
    }
    
    [data-testid="stSidebar"] * {
        resize: none !important;
    }
    
    /* 标题样式 */
    h1, h2, h3 {
        color: #333;
        font-weight: 600;
    }
    
    /* 改善按钮样式 */
    [data-testid="stSidebar"] .stButton > button {
        width: 100% !important;
        margin-bottom: 0.5rem !important;
        border-radius: 6px !important;
        transition: all 0.2s ease !important;
    }
    
    /* 改善链接样式 */
    [data-testid="stSidebar"] a {
        transition: color 0.2s ease !important;
    }
    
    [data-testid="stSidebar"] a:hover {
        color: #007acc !important;
    }
    </style>
    """, unsafe_allow_html=True)

def show_app_logo():
    """
    显示应用Logo - 此函数应该在 st.set_page_config 之后调用
    
    Returns:
        None
    """
    try:
        # 获取项目根目录中的logo路径
        root_dir = Path(__file__).resolve().parents[1]
        logo_path = str(root_dir / "static/kubeeye-logo.svg")
        icon_path = str(root_dir / "static/kubeeye.ico")
        
        # 使用 Streamlit 标准方法显示 logo
        st.logo(
            image=logo_path,
            size="large", 
            icon_image=icon_path,
            link="https://github.com/kubesphere/kubeeye"
        )
        
        st.markdown('<div style="height: 10px"></div>', unsafe_allow_html=True)
    except Exception as e:
        st.warning(f"Logo加载失败: {str(e)}")
        st.title("KubeEye - Kubernetes 集群巡检工具")

def create_sidebar_header(active_page="首页"):
    """
    创建侧边栏顶部的导航菜单和底部版权信息
    
    Args:
        active_page: 当前活动页面名称
        
    Returns:
        None
    """    
    # 定义导航菜单项
    menu_items = [
        {"title": "首页", "path": "app.py", "label": "首页", "icon": "🏠"},
        {"title": "集群信息", "path": "pages/1_cluster_info.py", "label": "集群信息", "icon": "🔗"},
        {"title": "集群巡检", "path": "pages/2_cluster_scan.py", "label": "集群巡检", "icon": "🔍"},
        {"title": "巡检报告", "path": "pages/3_scan_report.py", "label": "巡检报告", "icon": "📊"},
    ]
    
    with st.sidebar:
        # 导航菜单标题
        st.markdown("###")
        
        # 导航菜单按钮
        for item in menu_items:
            is_active = active_page == item["label"]
            button_type = "primary" if is_active else "secondary"
            
            if st.button(f"{item['icon']} {item['title']}", 
                        type=button_type,
                        use_container_width=True,
                        key=f"nav_{item['label']}"):
                try:
                    st.switch_page(item["path"])
                except Exception as e:
                    st.error(f"页面跳转失败: {str(e)}")
                    st.info(f"尝试跳转到: {item['path']}")
        
        # 简化方案：固定宽度244px，禁用拖拽调整
        st.markdown("""
        <style>
        .sidebar-footer {
            position: fixed !important;
            bottom: 0 !important;
            left: 0 !important;
            width: 244px !important;
            background-color: rgba(240, 242, 245, 0.95) !important;
            border-top: 1px solid rgba(0, 0, 0, 0.15) !important;
            padding: 0.8rem 1rem !important;
            text-align: center !important;
            font-size: 0.7rem !important;
            color: #666 !important;
            line-height: 1.4 !important;
            z-index: 9999 !important;
            transition: all 0.3s ease !important;
        }
        
        .sidebar-footer a {
            color: #00a971 !important;
            text-decoration: none !important;
            font-weight: 500 !important;
        }
        
        .sidebar-footer a:hover {
            color: #007f5f !important;
        }
        
        /* 当侧边栏收起时隐藏版权信息 */
        [data-testid="stSidebar"][aria-expanded="false"] ~ * .sidebar-footer,
        [data-testid="stSidebar"].st-emotion-cache-1d391kg ~ * .sidebar-footer {
            transform: translateX(-100%) !important;
            opacity: 0 !important;
        }
        
        /* 响应式处理 */
        @media (max-width: 768px) {
            .sidebar-footer {
                display: none !important;
            }
        }
        </style>
        """, unsafe_allow_html=True)
        
        # 在侧边栏中添加版权信息（使用固定定位）
        st.markdown(f"""
        <div class="sidebar-footer">
            <div style="margin-bottom: 4px; font-weight: 500;">© 2025 KubeEye v{VERSION}</div>
            <div>
                <a href="https://kubesphere.io" target="_blank">KubeSphere</a>
            </div>
        </div>
        """, unsafe_allow_html=True)
        

def create_page_header(title, subtitle="", icon=""):
    """
    创建页面标题区域
    
    Args:
        title: 页面标题
        subtitle: 页面副标题
        icon: 图标 (可选)
        
    Returns:
        None
    """
    # 使用更加优雅的标题样式
    st.markdown(f"""
    <div style="margin-bottom: 1rem;">
        <h2 style="color: #333; font-weight: 600; margin-bottom: 0.25rem;">
            {icon} {title}
        </h2>
        <p style="color: #666; font-size: 1rem; margin-top: 0;">
            {subtitle}
        </p>
    </div>
    """, unsafe_allow_html=True)
    
    # 添加一个细一点的分隔线
    st.markdown('<hr style="height: 1px; border: none; background: #eaeaea; margin: 1rem 0;" />', unsafe_allow_html=True)
