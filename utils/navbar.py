#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
导航栏组件 - 用于在所有页面创建一致的导航栏
"""

import streamlit as st
from pathlib import Path
from utils.image_utils import get_image_html
from utils.version import VERSION

def set_app_styles():
    """
    设置应用基本样式
    
    Returns:
        None
    """
    # 简化的基本样式
    st.markdown("""
    <style>
    /* 隐藏默认页面标题和页脚 */
    .main .block-container h1:first-child { display: none; }
    footer { visibility: hidden; }
    #MainMenu { visibility: hidden; }
    
    /* 改善整体界面和间距 */
    .main .block-container { padding-top: 1.5rem; }
    
    /* 改进侧边栏样式 */
    [data-testid="stSidebar"] {
        background-color: #f0f2f5;
        border-right: 1px solid rgba(0,0,0,0.05);
        position: relative;
        min-height: 100vh;
    }
    
    /* 标题样式 */
    h1, h2, h3 {
        color: #333;
        font-weight: 600;
    }
    
    /* 底部版权信息样式 - 侧边栏内底部贴边 */
    .sidebar-footer {
        position: fixed;
        bottom: 0;
        left: 0;
        width: 21rem; /* 固定宽度与侧边栏宽度一致 */
        max-width: 100%;
        padding: 10px 5px;
        background-color: #f0f2f5;
        border-top: 1px solid rgba(0,0,0,0.05);
        text-align: center;
        font-size: 0.85rem;
        color: #555;
        z-index: 99;
        display: flex;
        flex-direction: column;
        line-height: 1.5;
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
        {"title": "巡检报告", "path": "pages/3_scan_report.py", "label": "巡检报告", "icon": "📊"}
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
        
        # 为菜单留出底部空间，以免被固定版权信息覆盖
        st.markdown('<div style="margin-bottom: 50px;"></div>', unsafe_allow_html=True)
        
        # 固定在底部的版权信息
        st.markdown(f"""
        <div class="sidebar-footer">
            <div style="margin-bottom: 8px;">© 2025 KubeEye | v{VERSION}</div>
            <div>
                <a href="https://kubesphere.io" style="color: #00a971; text-decoration: none;">KubeSphere</a>
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
