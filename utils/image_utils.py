#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
图像处理工具，用于在Streamlit页面中展示图像
已弃用：建议直接使用Streamlit原生的st.image组件
"""

import streamlit as st
from pathlib import Path
import warnings

def get_image_html(file_path, width=None, height=None, alt="Image"):
    """
    已弃用的函数，仅为保持向后兼容性
    
    Args:
        file_path: 图片文件路径
        width: 图片宽度 (可选，已忽略)
        height: 图片高度 (可选，已忽略)
        alt: 替代文本 (可选，已忽略)
        
    Returns:
        str: 图片文件路径
        
    Warning:
        此函数已弃用，请使用 st.image() 替代
    """
    warnings.warn(
        "get_image_html() 已弃用，请直接使用 st.image() 显示图像",
        DeprecationWarning,
        stacklevel=2
    )
    # 只返回文件路径，调用处应改为使用st.image
    return str(file_path)
