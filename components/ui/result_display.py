#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检结果展示组件
"""
import streamlit as st
import pandas as pd
import re
from typing import Dict, List, Any, Optional

def display_status(status):
    """显示状态的彩色标记"""
    status_colors = {
        'passed': '🟢',
        'failed': '🔴', 
        'warning': '🟡',
        'error': '🔴',
        'skipped': '⚪'
    }
    return status_colors.get(status, '❓')

def format_status_badge(status):
    """格式化状态标记"""
    status_map = {
        'passed': '✅ 通过',
        'failed': '❌ 失败',
        'warning': '⚠️ 警告',
        'error': '🔥 错误',
        'skipped': '⏭️ 跳过'
    }
    return status_map.get(status, f'❓ {status}')

def parse_opa_violations_to_table(violations_text):
    """将OPA违规文本解析为表格数据"""
    if not violations_text or violations_text == "无违规资源":
        return []
    
    violations = []
    
    # 解析不同格式的违规信息
    lines = violations_text.split('\n')
    current_violation = {}
    
    for line in lines:
        line = line.strip()
        if not line:
            continue
            
        # 尝试匹配不同的格式
        if line.startswith('Name:') or line.startswith('name:'):
            if current_violation:
                violations.append(current_violation)
            current_violation = {'name': line.split(':', 1)[1].strip()}
        elif line.startswith('Kind:') or line.startswith('kind:'):
            current_violation['kind'] = line.split(':', 1)[1].strip()
        elif line.startswith('Namespace:') or line.startswith('namespace:'):
            current_violation['namespace'] = line.split(':', 1)[1].strip()
        elif line.startswith('Message:') or line.startswith('message:'):
            current_violation['message'] = line.split(':', 1)[1].strip()
    
    if current_violation:
        violations.append(current_violation)
    
    return violations

def get_items_safely(result):
    """安全地获取result.items，处理items是方法或属性的情况"""
    if not result:
        return []
    
    if hasattr(result, 'items'):
        items_attr = getattr(result, 'items')
        # 检查items是属性还是方法
        if callable(items_attr):
            try:
                items = items_attr()  # 如果是方法，调用它
            except:
                items = []
        else:
            items = items_attr if isinstance(items_attr, list) else []
    elif isinstance(result, dict) and 'items' in result:
        items = result['items'] if isinstance(result['items'], list) else []
    else:
        items = []
    
    return items

def count_status(items):
    """统计各状态的数量"""
    status_counts = {
        'passed': 0,
        'failed': 0,
        'warning': 0,
        'error': 0,
        'skipped': 0,
    }
    
    # 安全地统计状态，处理不同类型的item
    for item in items:
        if isinstance(item, dict):
            status = item.get('status', 'unknown')
            if status in status_counts:
                status_counts[status] += 1
    
    return status_counts

def display_opa_violations_table(violations_data: List[Dict], show_expander: bool = True, table_key: str = None):
    """以表格形式显示OPA违规资源 - 优化版本"""
    if not violations_data:
        st.info("无违规资源")
        return
    
    # 创建DataFrame
    import pandas as pd
    df = pd.DataFrame(violations_data)
    
    # 重命名列
    column_mapping = {
        'kind': '资源类型',
        'name': '资源名称', 
        'namespace': '命名空间',
        'message': '违规详情'
    }
    df = df.rename(columns=column_mapping)
    
    # 显示表格标题
    st.markdown("**违规资源列表:**")
    
    # 检查违规数量，决定显示方式
    if len(violations_data) <= 10:
        # 少量违规，使用优化的表格显示
        st.dataframe(
            df,
            use_container_width=True,
            hide_index=True,
            column_config={
                "资源类型": st.column_config.TextColumn(
                    "资源类型", 
                    help="Kubernetes资源类型"
                ),
                "资源名称": st.column_config.TextColumn(
                    "资源名称", 
                    help="资源实例名称"
                ),
                "命名空间": st.column_config.TextColumn(
                    "命名空间", 
                    help="Kubernetes命名空间"
                ),
                "违规详情": st.column_config.TextColumn(
                    "违规详情", 
                    help="具体的违规信息和描述"
                )
            },
            # 设置表格高度以避免过度压缩
            height=min(400, len(violations_data) * 50 + 100)
        )
    else:
        # 大量违规，使用分页或展开式显示
        st.info(f"发现 {len(violations_data)} 个违规资源，采用分页显示")
        
        # 分页显示
        page_size = 10
        total_pages = (len(violations_data) + page_size - 1) // page_size
        
        # 页面选择器
        if total_pages > 1:
            # 使用violations_data的id和时间戳生成唯一的selectbox key
            import time
            selectbox_key = f"violations_page_selector_{id(violations_data)}_{int(time.time() * 1000) % 10000}"
            page = st.selectbox(
                "选择页面", 
                range(1, total_pages + 1),
                format_func=lambda x: f"第 {x} 页 (共 {total_pages} 页)",
                key=selectbox_key
            ) - 1
        else:
            page = 0
        
        # 显示当前页的数据
        start_idx = page * page_size
        end_idx = min(start_idx + page_size, len(violations_data))
        page_df = df.iloc[start_idx:end_idx]
        
        st.dataframe(
            page_df,
            use_container_width=True,
            hide_index=True,
            column_config={
                "资源类型": st.column_config.TextColumn(
                    "资源类型"
                ),
                "资源名称": st.column_config.TextColumn(
                    "资源名称"
                ),
                "命名空间": st.column_config.TextColumn(
                    "命名空间"
                ),
                "违规详情": st.column_config.TextColumn(
                    "违规详情"
                )
            },
            height=400
        )
        
        # 显示页面信息
        st.caption(f"显示第 {start_idx + 1}-{end_idx} 项，共 {len(violations_data)} 项违规")
    
    # 根据参数决定是否显示详细视图选项
    if show_expander and len(violations_data) > 0:
        with st.expander("📋 查看详细列表", expanded=False):
            for i, violation in enumerate(violations_data, 1):
                st.markdown(f"**{i}. {violation.get('资源类型', violation.get('kind', 'Unknown'))}/{violation.get('资源名称', violation.get('name', 'unnamed'))}**")
                
                col1, col2 = st.columns([1, 3])
                with col1:
                    st.markdown(f"**命名空间:** {violation.get('命名空间', violation.get('namespace', '-'))}")
                with col2:
                    message = violation.get('违规详情', violation.get('message', '无详细信息'))
                    st.markdown(f"**违规详情:** {message}")
                
                if i < len(violations_data):
                    st.divider()

def display_summary_metrics(all_results):
    """显示巡检结果摘要信息"""
    if not all_results:
        return
        
    # 计算总体情况
    total_items = sum(len(get_items_safely(result)) for result in all_results.values())
    
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
        result_counts = count_status(get_items_safely(result))
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

def display_inspection_results(inspector_type: str, result, show_summary: bool = True):
    """
    显示巡检结果 - 简化状态系统版本
    
    Args:
        inspector_type: 巡检器类型
        result: 巡检结果
        show_summary: 是否显示摘要信息
    """
    if not result:
        st.info(f"📝 {inspector_type} 巡检结果为空")
        return
        
    items = get_items_safely(result)
    if not items:
        st.info(f"📝 {inspector_type} 无检查项")
        return
    
    # 按状态分组
    passed_items = [item for item in items if item.get('status') == 'passed']
    failed_items = [item for item in items if item.get('status') == 'failed']
    warning_items = [item for item in items if item.get('status') == 'warning']
    error_items = [item for item in items if item.get('status') == 'error']
    
    # 显示摘要
    if show_summary:
        col1, col2, col3, col4 = st.columns(4)
        with col1:
            st.metric("✅ 通过", len(passed_items))
        with col2:
            st.metric("❌ 失败", len(failed_items))
        with col3:
            st.metric("⚠️ 警告", len(warning_items))
        with col4:
            st.metric("🔥 错误", len(error_items))
    
    # 显示失败项
    if failed_items:
        st.markdown("#### ❌ 合规性问题")
        for item in failed_items:
            with st.expander(f"🔴 {item.get('name', '未命名检查')} - {item.get('description', '')}", expanded=True):
                # 添加基本信息展示
                col1, col2 = st.columns([2, 1])
                with col1:
                    st.markdown(f"**检查项:** {item.get('name', '未知检查')}")
                    st.markdown(f"**描述:** {item.get('description', '无描述')}")
                with col2:
                    severity = item.get('severity', 'unknown')
                    if severity == 'critical':
                        st.error(f"🔴 严重级别: {severity}")
                    elif severity == 'warning':
                        st.warning(f"🟡 警告级别: {severity}")
                    else:
                        st.info(f"ℹ️ 级别: {severity}")
                
                st.divider()
                
                # 显示违规详情
                details_content = item.get('details', '')
                if details_content and details_content != "无违规资源":
                    # 首先尝试使用原始violations数据
                    violations_data = []
                    if 'violations' in item and isinstance(item['violations'], list):
                        # 直接使用原始violations数据，包含完整的命名空间信息
                        violations_data = item['violations']
                    else:
                        # 备选方案：解析文本格式
                        violations_data = parse_opa_violations_to_table(details_content)
                    
                    if violations_data:
                        # 在专门的容器中显示表格，禁用expander避免嵌套
                        violations_container = st.container()
                        with violations_container:
                            display_opa_violations_table(violations_data, show_expander=False)
                    else:
                        st.text_area("详细信息", details_content, height=150)
                
                # 显示解决方案
                if item.get('solution'):
                    st.divider()
                    st.markdown("**💡 解决方案:**")
                    st.info(item['solution'])
    
    # 显示警告项
    if warning_items:
        st.markdown("#### ⚠️ 合规性警告")
        for item in warning_items:
            with st.expander(f"🟡 {item.get('name', '未命名检查')} - {item.get('description', '')}"):
                # 添加基本信息展示
                col1, col2 = st.columns([2, 1])
                with col1:
                    st.markdown(f"**检查项:** {item.get('name', '未知检查')}")
                    st.markdown(f"**描述:** {item.get('description', '无描述')}")
                with col2:
                    severity = item.get('severity', 'warning')
                    st.warning(f"⚠️ 警告级别: {severity}")
                
                st.divider()
                
                # 显示违规详情
                details_content = item.get('details', '')
                if details_content and details_content != "无违规资源":
                    # 首先尝试使用原始violations数据
                    violations_data = []
                    if 'violations' in item and isinstance(item['violations'], list):
                        # 直接使用原始violations数据，包含完整的命名空间信息
                        violations_data = item['violations']
                    else:
                        # 备选方案：解析文本格式
                        violations_data = parse_opa_violations_to_table(details_content)
                    
                    if violations_data:
                        # 在专门的容器中显示表格，禁用expander避免嵌套
                        violations_container = st.container()
                        with violations_container:
                            display_opa_violations_table(violations_data, show_expander=False)
                    else:
                        st.text_area("详细信息", details_content, height=150)
                
                # 显示解决方案
                if item.get('solution'):
                    st.divider()
                    st.markdown("**💡 建议方案:**")
                    st.info(item['solution'])
    
    # 显示错误项
    if error_items:
        st.markdown("#### 🔥 系统错误")
        for item in error_items:
            with st.expander(f"🔥 {item.get('name', '未命名检查')} - {item.get('description', '')}"):
                st.error(f"错误信息: {item.get('details', '无详细错误信息')}")
                if item.get('solution'):
                    st.markdown("**💡 解决方案:**")
                    st.info(item['solution'])

def display_result_summary(results):
    """显示结果摘要"""
    if not results:
        st.info("没有巡检结果")
        return
    
    # 计算总体统计
    total_passed = 0
    total_failed = 0
    total_warning = 0
    total_error = 0
    
    for result in results.values():
        items = get_items_safely(result)
        status_counts = count_status(items)
        total_passed += status_counts['passed']
        total_failed += status_counts['failed'] 
        total_warning += status_counts['warning']
        total_error += status_counts['error']
    
    # 显示摘要卡片
    col1, col2, col3, col4 = st.columns(4)
    
    with col1:
        st.metric(
            label="✅ 通过",
            value=total_passed,
            delta=None
        )
    
    with col2:
        st.metric(
            label="❌ 失败", 
            value=total_failed,
            delta=None
        )
    
    with col3:
        st.metric(
            label="⚠️ 警告",
            value=total_warning, 
            delta=None
        )
    
    with col4:
        st.metric(
            label="🔥 错误",
            value=total_error,
            delta=None
        )

def display_opa_results(result):
    """显示OPA巡检结果"""
    return display_inspection_results("OPA", result)

def display_node_results(result):
    """显示节点巡检结果"""
    return display_inspection_results("节点检查", result)

def display_prometheus_results(result):
    """显示Prometheus巡检结果"""
    return display_inspection_results("Prometheus", result)
