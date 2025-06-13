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
        # 安全地获取status
        if isinstance(item, dict):
            status = item.get('status', 'unknown')
        elif hasattr(item, 'status'):
            status = getattr(item, 'status', 'unknown')
        else:
            status = 'unknown'
        
        if status in status_counts:
            status_counts[status] += 1
    
    return status_counts

def parse_opa_violations_to_table(details_content: str) -> List[Dict]:
    """解析OPA违规信息为表格数据"""
    if not details_content or details_content == "无违规资源":
        return []
    
    violations = []
    if details_content.startswith("- "):
        for line in details_content.strip().split('\n'):
            if line.strip():
                violation_text = line.strip("- ")
                
                # 解析格式：kind/name (命名空间: namespace): message
                # 或者：kind/name: message
                parts = violation_text.split(': ', 1)
                if len(parts) >= 2:
                    resource_part = parts[0]
                    message = parts[1]
                    
                    # 解析资源信息
                    if ' (命名空间: ' in resource_part:
                        resource_name, namespace_part = resource_part.split(' (命名空间: ', 1)
                        namespace = namespace_part.rstrip(')')
                    else:
                        resource_name = resource_part
                        namespace = "-"
                    
                    # 解析kind和name
                    if '/' in resource_name:
                        kind, name = resource_name.split('/', 1)
                    else:
                        kind = "Unknown"
                        name = resource_name
                    
                    violations.append({
                        'kind': kind,
                        'name': name,
                        'namespace': namespace,
                        'message': message
                    })
    
    return violations

def display_opa_violations_table(violations_data: List[Dict]):
    """以表格形式显示OPA违规资源"""
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
    
    # 显示表格
    st.markdown("**违规资源列表:**")
    st.dataframe(
        df,
        use_container_width=True,
        hide_index=True,
        column_config={
            "资源类型": st.column_config.TextColumn(width="small"),
            "资源名称": st.column_config.TextColumn(width="medium"),
            "命名空间": st.column_config.TextColumn(width="small"),
            "违规详情": st.column_config.TextColumn(width="large")
        }
    )

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
    统一的巡检结果展示函数，支持不同类型的巡检器
    
    Args:
        inspector_type: 巡检器类型 (node, prometheus, opa)
        result: 巡检结果对象
        show_summary: 是否显示摘要信息
    """
    items = get_items_safely(result)
    if not items:
        st.warning(f"没有 {inspector_type} 巡检结果")
        return
    
    # 显示摘要信息
    if show_summary:
        display_result_summary(inspector_type, result)
    
    # 根据巡检类型选择不同的展示方式
    if inspector_type == "opa":
        display_opa_results(result)
    elif inspector_type == "node":
        display_node_results(result)
    elif inspector_type == "prometheus":
        display_prometheus_results(result)
    else:
        # 默认展示方式
        display_generic_results(result)

def display_result_summary(inspector_type: str, result):
    """显示巡检结果摘要"""
    type_names = {
        "node": "节点巡检",
        "prometheus": "Prometheus指标",
        "opa": "OPA合规性"
    }
    
    st.subheader(f"📊 {type_names.get(inspector_type, inspector_type)} 结果摘要")
    
    items = get_items_safely(result)
    status_counts = count_status(items)
    total = sum(status_counts.values())
    
    if total == 0:
        st.info("没有检查项")
        return
    
    # 显示指标卡片
    col1, col2, col3, col4, col5 = st.columns(5)
    
    with col1:
        st.metric("总计", total)
    
    with col2:
        passed_pct = round(status_counts['passed'] / total * 100, 1) if total > 0 else 0
        st.metric("通过", f"{status_counts['passed']} ({passed_pct}%)")
    
    with col3:
        failed_pct = round(status_counts['failed'] / total * 100, 1) if total > 0 else 0
        st.metric("失败", f"{status_counts['failed']} ({failed_pct}%)")
    
    with col4:
        warning_pct = round(status_counts['warning'] / total * 100, 1) if total > 0 else 0
        st.metric("警告", f"{status_counts['warning']} ({warning_pct}%)")
    
    with col5:
        error_pct = round(status_counts['error'] / total * 100, 1) if total > 0 else 0
        st.metric("错误", f"{status_counts['error']} ({error_pct}%)")

def display_opa_results(result):
    """显示OPA合规性巡检结果"""
    st.subheader("🔒 合规性检查详情")
    
    # 按状态分组显示
    items = get_items_safely(result)
    failed_items = [item for item in items if item.get('status') == 'failed']
    warning_items = [item for item in items if item.get('status') == 'warning']
    passed_items = [item for item in items if item.get('status') == 'passed']
    
    # 显示失败项
    if failed_items:
        st.markdown("#### ❌ 合规性问题")
        for item in failed_items:
            with st.expander(f"🔴 {item.get('name', '未命名检查')} - {item.get('description', '')}"):
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
                        display_opa_violations_table(violations_data)
                    else:
                        st.text(details_content)
                
                # 显示解决方案
                if item.get('solution'):
                    st.markdown("**解决方案:**")
                    st.info(item['solution'])
    
    # 显示警告项
    if warning_items:
        st.markdown("#### ⚠️ 合规性警告")
        for item in warning_items:
            with st.expander(f"🟡 {item.get('name', '未命名检查')} - {item.get('description', '')}"):
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
                        display_opa_violations_table(violations_data)
                    else:
                        st.text(details_content)
    
    # 显示通过项（可折叠）
    if passed_items:
        with st.expander(f"✅ 已通过的检查项 ({len(passed_items)}个)"):
            for item in passed_items:
                st.markdown(f"- ✅ {item.get('name', '未命名检查')}: {item.get('description', '')}")

def display_node_results(result):
    """显示节点巡检结果"""
    st.subheader("🖥️ 节点检查详情")
    
    # 按节点分组显示结果
    items = get_items_safely(result)
    node_groups = {}
    for item in items:
        node_name = extract_node_name(item.get('name', ''))
        if node_name not in node_groups:
            node_groups[node_name] = []
        node_groups[node_name].append(item)
    
    for node_name, items in node_groups.items():
        st.markdown(f"#### 🖥️ 节点: {node_name}")
        
        # 按状态分类
        failed_items = [item for item in items if item.get('status') == 'failed']
        warning_items = [item for item in items if item.get('status') == 'warning']
        passed_items = [item for item in items if item.get('status') == 'passed']
        
        # 显示节点状态概览
        col1, col2, col3 = st.columns(3)
        with col1:
            st.metric("失败", len(failed_items))
        with col2:
            st.metric("警告", len(warning_items))
        with col3:
            st.metric("通过", len(passed_items))
        
        # 显示详细信息
        if failed_items:
            st.markdown("**❌ 失败项:**")
            for item in failed_items:
                with st.expander(f"🔴 {item.get('description', item.get('name', '未命名检查'))}"):
                    display_generic_item_details(item)
        
        if warning_items:
            st.markdown("**⚠️ 警告项:**")
            for item in warning_items:
                with st.expander(f"🟡 {item.get('description', item.get('name', '未命名检查'))}"):
                    display_generic_item_details(item)
        
        # 通过的检查项可折叠显示
        if passed_items:
            with st.expander(f"✅ 通过的检查项 ({len(passed_items)}个)"):
                for item in passed_items:
                    st.markdown(f"- ✅ {item.get('description', item.get('name', '未命名检查'))}")
        
        st.divider()

def display_prometheus_results(result):
    """显示Prometheus指标巡检结果"""
    st.subheader("📈 Prometheus指标检查详情")
    
    # 按类别分组显示
    items = get_items_safely(result)
    categories = {}
    for item in items:
        category = item.get('category', '其他')
        if category not in categories:
            categories[category] = []
        categories[category].append(item)
    
    for category, items in categories.items():
        st.markdown(f"#### 📊 {category}")
        
        # 按状态分类
        failed_items = [item for item in items if item.get('status') == 'failed']
        warning_items = [item for item in items if item.get('status') == 'warning']
        passed_items = [item for item in items if item.get('status') == 'passed']
        
        # 显示指标
        if failed_items or warning_items:
            for item in failed_items + warning_items:
                status_icon = "🔴" if item.get('status') == 'failed' else "🟡"
                severity = item.get('severity', 'warning')
                
                with st.expander(f"{status_icon} {item.get('description', item.get('name', '未命名检查'))} [{severity}]"):
                    display_prometheus_item_details(item)
        
        # 通过的检查项
        if passed_items:
            with st.expander(f"✅ 正常指标 ({len(passed_items)}个)"):
                for item in passed_items:
                    st.markdown(f"- ✅ {item.get('description', item.get('name', '未命名检查'))}")
        
        st.divider()

def display_generic_results(result):
    """通用的巡检结果展示"""
    st.subheader("📋 巡检结果详情")
    
    items = get_items_safely(result)
    for item in items:
        status = item.get('status', 'unknown')
        status_icon = {
            'passed': '✅',
            'failed': '❌', 
            'warning': '⚠️',
            'error': '🔴',
            'skipped': '⏭️'
        }.get(status, '❓')
        
        with st.expander(f"{status_icon} {item.get('name', '未命名检查')} - {item.get('description', '')}"):
            display_generic_item_details(item)

def display_generic_item_details(item):
    """显示通用的检查项详情"""
    # 基本信息
    if item.get('description'):
        st.markdown(f"**描述:** {item['description']}")
    
    if item.get('severity'):
        st.markdown(f"**严重程度:** {item['severity']}")
    
    # 详细信息
    if item.get('details'):
        st.markdown("**详细信息:**")
        st.code(item['details'], language='text')
    
    # 解决方案
    if item.get('solution'):
        st.markdown("**解决方案:**")
        st.info(item['solution'])

def display_prometheus_item_details(item):
    """显示Prometheus检查项的详细信息"""
    # 基本信息
    if item.get('description'):
        st.markdown(f"**检查项:** {item['description']}")
    
    if item.get('severity'):
        severity_colors = {
            'critical': '🔴',
            'warning': '🟡', 
            'info': '🔵'
        }
        severity_icon = severity_colors.get(item['severity'], '⚪')
        st.markdown(f"**严重程度:** {severity_icon} {item['severity']}")
    
    # 指标详情
    details = item.get('details', '')
    if details:
        st.markdown("**指标详情:**")
        
        # 检查是否包含百分比信息，格式化显示
        if "%" in details and ("使用率" in details or "负载" in details):
            lines = details.strip().split('\n')
            for line in lines:
                if line.strip():
                    if ":" in line:
                        key, value = line.split(':', 1)
                        # 如果值包含百分比，用进度条显示
                        if "%" in value:
                            try:
                                pct_match = re.search(r'(\d+\.?\d*)%', value)
                                if pct_match:
                                    percentage = float(pct_match.group(1))
                                    st.markdown(f"**{key.strip()}:**")
                                    st.progress(min(percentage / 100, 1.0))
                                    st.caption(f"{percentage}%")
                                else:
                                    st.markdown(f"**{key.strip()}:** {value.strip()}")
                            except:
                                st.markdown(f"**{key.strip()}:** {value.strip()}")
                        else:
                            st.markdown(f"**{key.strip()}:** {value.strip()}")
                    else:
                        st.text(line)
        else:
            st.code(details, language='text')
    
    # 解决方案
    if item.get('solution'):
        st.markdown("**建议措施:**")
        st.info(item['solution'])

def extract_node_name(item_name: str) -> str:
    """从检查项名称中提取节点名称"""
    # 节点巡检项通常格式为 "检查项名称 - 节点IP/名称"
    if " - " in item_name:
        return item_name.split(" - ")[-1]
    return "未知节点"
