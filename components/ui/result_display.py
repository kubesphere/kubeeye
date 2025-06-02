#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检结果展示组件
"""
import streamlit as st
import pandas as pd
from typing import Dict, List, Any, Optional

def display_status(status):
    """显示状态的彩色标记"""
    if status == 'passed':
        return "✅ 通过"
    elif status == 'failed':
        return "❌ 失败"
    elif status == 'warning':
        return "⚠️ 警告"
    elif status == 'error':
        return "🔴 错误"
    elif status == 'skipped':
        return "⏭️ 跳过"
    else:
        return "❓ 未知"

def count_status(items):
    """统计各状态的数量"""
    status_counts = {
        'passed': 0,
        'failed': 0,
        'warning': 0,
        'error': 0,
        'skipped': 0,
    }
    
    # 使用列表推导式而不是循环，提高性能
    status_counts['passed'] = sum(1 for item in items if item.get('status') == 'passed')
    status_counts['failed'] = sum(1 for item in items if item.get('status') == 'failed')
    status_counts['warning'] = sum(1 for item in items if item.get('status') == 'warning')
    status_counts['error'] = sum(1 for item in items if item.get('status') == 'error')
    status_counts['skipped'] = sum(1 for item in items if item.get('status') == 'skipped')
    
    return status_counts

def display_inspection_results(inspector_type, result):
    """显示巡检结果详情"""
    # 统计各状态的数量
    status_counts = count_status(result.items)
    
    # 显示统计信息
    st.write(f"总共检查了 **{len(result.items)}** 个项目")
    st.write(f"通过: **{status_counts['passed']}**, 失败: **{status_counts['failed']}**, 警告: **{status_counts['warning']}**, 错误: **{status_counts['error']}**, 跳过: **{status_counts['skipped']}**")
    
    # 创建表格数据 - 使用列表推导式代替循环，提高性能
    table_data = [{
        '规则名称': item.get('name', 'Unknown'),
        '状态': display_status(item.get('status', 'unknown')),
        '严重程度': item.get('severity', 'info'),
        '描述': item.get('description', ''),
    } for item in result.items]
    
    # 显示表格
    if table_data:
        st.write("#### 检查项目详情")
        st.dataframe(table_data, use_container_width=True)
    
        # 只过滤一次问题项，而不是每次循环都检查
        problems = [item for item in result.items if item.get('status') not in ['passed', 'skipped']]
        if problems:
            st.write("#### 问题详情")
            
            for i, item in enumerate(problems):
                with st.expander(f"{i+1}. {item.get('name', 'Unknown')} ({item.get('status', 'unknown')})"):
                    st.markdown(f"**描述:** {item.get('description', 'N/A')}")
                    st.markdown(f"**详情:** {item.get('details', 'N/A')}")
                    if item.get('solution'):
                        st.markdown("**解决方案:**")
                        st.markdown(f"{item.get('solution')}")

def display_summary_metrics(all_results):
    """显示巡检结果摘要信息"""
    if not all_results:
        return
        
    # 计算总体情况
    total_items = sum(len(result.items) for result in all_results.values())
    
    # 合并所有结果的状态计数
    status_counts = {
        'passed': 0,
        'failed': 0,
        'warning': 0,
        'error': 0,
        'skipped': 0,
    }
    
    # 使用字典推导式合并计数，更加高效
    for result in all_results.values():
        result_counts = count_status(result.items)
        for status, count in result_counts.items():
            status_counts[status] += count
    
    # 显示摘要信息
    st.write(f"总共检查了 **{total_items}** 个项目")
    
    # 创建状态计数的列布局
    status_col1, status_col2, status_col3, status_col4, status_col5 = st.columns(5)
    
    with status_col1:
        st.metric("通过", status_counts['passed'], delta=None)
    
    with status_col2:
        st.metric("失败", status_counts['failed'], delta=None)
        
    with status_col3:
        st.metric("警告", status_counts['warning'], delta=None)
        
    with status_col4:
        st.metric("错误", status_counts['error'], delta=None)
        
    with status_col5:
        st.metric("跳过", status_counts['skipped'], delta=None)
