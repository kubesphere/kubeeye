#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 集群巡检报告页面 - 重新设计版本
提供更好的用户体验和更清晰的报告管理界面
"""

import streamlit as st
import pandas as pd
from datetime import datetime, timedelta
import json
import os
import sys
from pathlib import Path

# 设置页面配置
st.set_page_config(
    page_title="巡检报告 - kubeeye",
    page_icon="📊",
    layout="wide"
)

# 添加项目根目录到Python路径
ROOT_DIR = Path(__file__).resolve().parent.parent
if str(ROOT_DIR) not in sys.path:
    sys.path.insert(0, str(ROOT_DIR))

# 导入工具模块
from utils.common import initialize_page
from utils.cluster_config import list_clusters
from utils.inspection_result import list_results, load_result
from components.ui.result_display import display_inspection_results
from components.ui.result_display import parse_opa_violations_to_table, display_opa_violations_table

def display_reports_overview():
    """显示报告概览页面"""
    st.markdown("## 📊 巡检报告中心")
    st.markdown("查看和管理所有集群的巡检报告，快速识别问题并获取解决建议。")
    
    # 加载数据
    clusters = list_clusters()
    all_results = list_results()
    
    # 检查是否有最新的巡检结果需要显示
    if hasattr(st.session_state, 'last_result_path') and st.session_state.last_result_path:
        st.success(f"🆕 检测到最新巡检结果: {os.path.basename(st.session_state.last_result_path)}")
        if st.button("🔍 查看最新结果", type="primary"):
            st.session_state.selected_report_id = st.session_state.get('last_cluster_name', '未知集群')
            st.session_state.report_detail_path = st.session_state.last_result_path
            st.session_state.view_mode = "detail"
            st.rerun()
    
    if not all_results:
        st.info("📭 暂无巡检报告。请先执行巡检任务生成报告。")
        if st.button("🚀 去执行巡检", type="primary"):
            st.switch_page("pages/2_cluster_scan.py")
        return
    
    # 过滤控制区域
    st.markdown("### 🔍 筛选报告")
    col1, col2, col3 = st.columns(3)
    
    with col1:
        selected_cluster = st.selectbox(
            "选择集群", 
            ["全部"] + clusters,
            help="筛选特定集群的报告"
        )
    
    with col2:
        inspection_types = ["全部", "immediate", "scheduled"]
        selected_type = st.selectbox(
            "巡检类型",
            inspection_types,
            format_func=lambda x: "立即巡检" if x == "immediate" else ("定时巡检" if x == "scheduled" else x)
        )
    
    with col3:
        date_filter = st.selectbox(
            "时间范围",
            ["全部", "今天", "最近7天", "最近30天"]
        )
    
    # 应用筛选
    filtered_results = all_results
    if selected_cluster != "全部":
        filtered_results = [r for r in filtered_results if r["cluster_name"] == selected_cluster]
    if selected_type != "全部":
        filtered_results = [r for r in filtered_results if r["inspection_type"] == selected_type]
    
    # 日期筛选逻辑
    if date_filter != "全部":
        now = datetime.now()
        if date_filter == "今天":
            filtered_results = [r for r in filtered_results if 
                               datetime.fromisoformat(r['timestamp']).date() == now.date()]
        elif date_filter == "最近7天":
            week_ago = now - timedelta(days=7)
            filtered_results = [r for r in filtered_results if 
                               datetime.fromisoformat(r['timestamp']) >= week_ago]
        elif date_filter == "最近30天":
            month_ago = now - timedelta(days=30)
            filtered_results = [r for r in filtered_results if 
                               datetime.fromisoformat(r['timestamp']) >= month_ago]
    
    # 显示统计概览
    if filtered_results:
        display_statistics_overview(filtered_results)
        display_reports_table(filtered_results)
    else:
        st.info("🔍 没有找到符合条件的巡检报告")

def display_statistics_overview(filtered_results):
    """显示统计概览"""
    st.markdown("### 📈 统计概览")
    
    # 计算汇总数据
    total_reports = len(filtered_results)
    total_critical = sum(r['critical'] for r in filtered_results)
    total_warning = sum(r['warning'] for r in filtered_results)
    total_passed = sum(r['passed'] for r in filtered_results)
    
    # 计算最新报告时间
    latest_report = max(filtered_results, key=lambda x: x['timestamp'])
    latest_time = datetime.fromisoformat(latest_report['timestamp']).strftime('%m-%d %H:%M')
    
    # 显示关键指标卡片
    col1, col2, col3, col4, col5 = st.columns(5)
    
    with col1:
        st.metric("📄 报告总数", total_reports)
    
    with col2:
        st.metric("❌ 关键问题", total_critical, 
                  delta=f"-{total_critical}" if total_critical > 0 else None,
                  delta_color="inverse")
    
    with col3:
        st.metric("⚠️ 警告", total_warning,
                  delta=f"-{total_warning}" if total_warning > 0 else None, 
                  delta_color="inverse")
    
    with col4:
        st.metric("✅ 通过", total_passed,
                  delta=f"+{total_passed}" if total_passed > 0 else None)
    
    with col5:
        st.metric("🕒 最新报告", latest_time)

def display_reports_table(filtered_results):
    """显示报告表格"""
    st.markdown("### 📋 报告列表")
    
    # 准备表格数据
    table_data = []
    for result in filtered_results:
        timestamp = datetime.fromisoformat(result['timestamp'])
        
        # 状态计算
        if result['critical'] > 0:
            status = "🔴 严重"
            status_score = 3
        elif result['warning'] > 0:
            status = "🟡 警告" 
            status_score = 2
        else:
            status = "🟢 正常"
            status_score = 1
        
        table_data.append({
            "集群": result['cluster_name'],
            "类型": "立即" if result['inspection_type'] == "immediate" else "定时",
            "时间": timestamp.strftime('%m-%d %H:%M'),
            "状态": status,
            "状态分数": status_score,  # 用于排序
            "问题": result['critical'] + result['warning'],
            "通过": result['passed'],
            "报告ID": result['result_id']
        })
    
    # 按状态严重程度和时间排序
    table_data.sort(key=lambda x: (-x['状态分数'], x['时间']), reverse=True)
    
    # 显示表格
    df = pd.DataFrame(table_data)
    df = df.drop('状态分数', axis=1)  # 移除排序用的列
    
    # 使用dataframe显示，支持选择
    event = st.dataframe(
        df,
        use_container_width=True,
        hide_index=True,
        on_select="rerun",
        selection_mode="single-row",
        column_config={
            "集群": st.column_config.TextColumn("集群", width="small"),
            "类型": st.column_config.TextColumn("类型", width="small"), 
            "时间": st.column_config.TextColumn("时间", width="small"),
            "状态": st.column_config.TextColumn("状态", width="small"),
            "问题": st.column_config.NumberColumn("问题数", width="small"),
            "通过": st.column_config.NumberColumn("通过数", width="small"),
            "报告ID": st.column_config.TextColumn("报告ID", width="medium")
        }
    )
    
    # 处理行选择
    if event.selection.rows:
        selected_idx = event.selection.rows[0]
        selected_report_id = table_data[selected_idx]['报告ID']
        
        # 操作按钮
        col1, col2, col3, col4 = st.columns(4)
        with col1:
            if st.button("🔍 查看详情", type="primary"):
                st.session_state.selected_report_id = selected_report_id
                st.session_state.view_mode = "detail"
                st.rerun()
        
        with col2:
            if st.button("📊 快速预览"):
                st.session_state.selected_report_id = selected_report_id
                st.session_state.view_mode = "preview"
                st.rerun()
        
        with col3:
            if st.button("📥 导出报告"):
                st.session_state.selected_report_id = selected_report_id
                st.session_state.view_mode = "export"
                st.rerun()

def display_report_detail(report_id):
    """显示报告详情"""
    # 导航返回按钮
    if st.button("⬅️ 返回报告列表"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    # 加载报告数据
    report_data = load_result(report_id)
    if not report_data:
        st.error("❌ 无法加载报告数据")
        return
    
    # 报告头部信息
    st.markdown(f"## 📄 巡检报告详情")
    
    # 基本信息卡片
    col1, col2 = st.columns([2, 1])
    with col1:
        st.markdown(f"""
        **🏷️ 报告ID:** `{report_id}`  
        **🖥️ 集群:** {report_data['cluster_name']}  
        **⏰ 时间:** {datetime.fromisoformat(report_data['timestamp']).strftime('%Y年%m月%d日 %H:%M:%S')}  
        **📋 类型:** {'立即巡检' if report_data['inspection_type'] == 'immediate' else '定时巡检'}
        """)
    
    with col2:
        # 快速统计 - 需要处理新的数据结构
        if 'inspection_results' in report_data:
            # 新数据结构：从inspection_results中统计
            all_items = []
            for inspector_type, inspector_result in report_data['inspection_results'].items():
                items = inspector_result.get('items', [])
                all_items.extend(items)
        else:
            # 旧数据结构：直接从items字段
            all_items = report_data.get('items', [])
        
        critical_count = 0
        warning_count = 0
        passed_count = 0
        
        for item in all_items:
            if isinstance(item, dict):
                status = item.get('status', 'unknown')
                severity = item.get('severity', 'unknown')
            elif hasattr(item, 'status'):
                status = getattr(item, 'status', 'unknown')
                severity = getattr(item, 'severity', 'unknown')
            else:
                status = 'unknown'
                severity = 'unknown'
            
            if status == 'passed':
                passed_count += 1
            elif status == 'failed' and severity == 'critical':
                critical_count += 1
            elif status == 'failed' or severity == 'warning':
                warning_count += 1
        
        if critical_count > 0:
            st.error(f"🔴 发现 {critical_count} 个严重问题")
        elif warning_count > 0:
            st.warning(f"🟡 发现 {warning_count} 个警告")
        else:
            st.success(f"🟢 所有检查通过 ({passed_count} 项)")
    
    st.divider()
    
    # 检查项详情 - 传递处理后的所有项目
    display_inspection_items(all_items)

def display_report_preview(report_id):
    """显示报告快速预览"""
    if st.button("⬅️ 返回"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    st.markdown(f"## 👁️ 报告预览 - {report_id}")
    
    # 加载数据并显示简化版本
    report_data = load_result(report_id)
    if not report_data:
        st.error("❌ 无法加载报告数据")
        return
    
    # 只显示关键统计和严重问题 - 处理新的数据结构
    if 'inspection_results' in report_data:
        # 新数据结构：从inspection_results中获取所有项目
        all_items = []
        for inspector_type, inspector_result in report_data['inspection_results'].items():
            items = inspector_result.get('items', [])
            all_items.extend(items)
    else:
        # 旧数据结构：直接从items字段
        all_items = report_data.get('items', [])
    
    critical_items = []
    warning_items = []
    
    for item in all_items:
        if isinstance(item, dict):
            status = item.get('status', 'unknown')
            severity = item.get('severity', 'unknown')
        elif hasattr(item, 'status'):
            status = getattr(item, 'status', 'unknown')
            severity = getattr(item, 'severity', 'unknown')
        else:
            continue
            
        if status == 'failed' and severity == 'critical':
            critical_items.append(item)
        elif status == 'failed' and severity == 'warning':
            warning_items.append(item)
    
    col1, col2 = st.columns(2)
    with col1:
        st.metric("🔴 严重问题", len(critical_items))
        st.metric("🟡 警告", len(warning_items))
    
    with col2:
        passed_count = sum(1 for item in all_items 
                          if (isinstance(item, dict) and item.get('status') == 'passed') or 
                             (hasattr(item, 'status') and getattr(item, 'status') == 'passed'))
        st.metric("✅ 通过", passed_count)
        st.metric("📊 总计", len(all_items))
    
    # 只显示严重问题
    if critical_items:
        st.markdown("### 🚨 严重问题")
        for item in critical_items[:5]:  # 最多显示5个
            st.error(f"**{item.get('name', '未知')}:** {item.get('description', '')}")
        
        if len(critical_items) > 5:
            st.info(f"还有 {len(critical_items) - 5} 个严重问题，查看完整报告了解详情")
    
    if st.button("📋 查看完整报告", type="primary"):
        st.session_state.view_mode = "detail"
        st.rerun()

def display_inspection_items(items):
    """显示巡检项详情 - 简化版本"""
    # 按严重程度分类
    critical_items = [item for item in items 
                     if item.get('status') == 'failed' and item.get('severity') == 'critical']
    warning_items = [item for item in items 
                    if item.get('status') == 'failed' and item.get('severity') == 'warning']
    passed_items = [item for item in items if item.get('status') == 'passed']
    other_items = [item for item in items if item not in critical_items + warning_items + passed_items]
    
    # 创建标签页
    tab_names = []
    tab_data = []
    
    if critical_items:
        tab_names.append(f"🔴 严重问题 ({len(critical_items)})")
        tab_data.append(critical_items)
    
    if warning_items:
        tab_names.append(f"🟡 警告 ({len(warning_items)})")
        tab_data.append(warning_items)
    
    if other_items:
        tab_names.append(f"⚠️ 其他 ({len(other_items)})")
        tab_data.append(other_items)
    
    if passed_items:
        tab_names.append(f"✅ 通过 ({len(passed_items)})")
        tab_data.append(passed_items)
    
    if tab_names:
        tabs = st.tabs(tab_names)
        for i, (tab, data) in enumerate(zip(tabs, tab_data)):
            with tab:
                display_items_list(data, tab_names[i].startswith("✅"))

def display_items_list(items, is_passed=False):
    """显示检查项列表"""
    if not items:
        st.info("此类别下暂无项目")
        return
    
    # 对于通过的项目，默认折叠显示
    for item in items:
        title = f"{item.get('name', '未知检查项')}"
        
        # 根据类型选择展开状态
        expanded = not is_passed and item.get('severity') == 'critical'
        
        with st.expander(title, expanded=expanded):
            # 基本信息
            col1, col2 = st.columns([3, 1])
            with col1:
                st.markdown(f"**描述:** {item.get('description', '无')}")
                
                # 显示详细信息
                details = item.get('details', '')
                if details:
                    # 特殊处理OPA违规
                    if 'violations' in item and isinstance(item['violations'], list):
                        st.markdown("**违规资源:**")
                        display_opa_violations_table(item['violations'])
                    else:
                        st.markdown("**详细信息:**")
                        st.text(details)
            
            with col2:
                # 状态标签
                status = item.get('status', 'unknown')
                severity = item.get('severity', 'info')
                
                if status == 'failed':
                    if severity == 'critical':
                        st.error("🔴 严重")
                    else:
                        st.warning("🟡 警告")
                elif status == 'passed':
                    st.success("✅ 通过")
                else:
                    st.info("ℹ️ 其他")
            
            # 解决方案
            solution = item.get('solution', '')
            if solution:
                st.markdown("**💡 建议解决方案:**")
                st.info(solution)

def display_export_page(report_id):
    """显示导出页面"""
    if st.button("⬅️ 返回"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    st.markdown(f"### 📥 导出报告 - {report_id}")
    
    col1, col2 = st.columns([2, 1])
    with col1:
        export_format = st.selectbox(
            "选择导出格式",
            ["JSON", "CSV", "Excel"],
            help="选择您希望导出的报告格式"
        )
    
    with col2:
        if st.button("📥 开始导出", type="primary"):
            try:
                from utils.inspection_result import export_report
                format_map = {"JSON": "json", "CSV": "csv", "Excel": "excel"}
                success, file_path = export_report(report_id, format_map[export_format])
                
                if success:
                    st.success(f"✅ 导出成功: {file_path}")
                    
                    # 提供下载按钮
                    with open(file_path, "rb") as f:
                        st.download_button(
                            label=f"下载 {export_format} 文件",
                            data=f.read(),
                            file_name=os.path.basename(file_path),
                            mime="application/octet-stream"
                        )
                else:
                    st.error(f"❌ 导出失败: {file_path}")
            except Exception as e:
                st.error(f"❌ 导出时发生错误: {str(e)}")

def main():
    """主函数"""
    # 初始化页面
    initialize_page(
        title="巡检报告", 
        icon="📊",
        page_title="巡检报告",
        page_subtitle="集群健康状况一览"
    )
    
    # 初始化会话状态
    if 'view_mode' not in st.session_state:
        st.session_state.view_mode = "list"
    
    if 'selected_report_id' not in st.session_state:
        st.session_state.selected_report_id = None
    
    # 根据视图模式显示不同内容
    if st.session_state.view_mode == "list":
        display_reports_overview()
    
    elif st.session_state.view_mode == "detail":
        if st.session_state.selected_report_id:
            display_report_detail(st.session_state.selected_report_id)
        else:
            st.error("❌ 未选择报告")
            st.session_state.view_mode = "list"
            st.rerun()
    
    elif st.session_state.view_mode == "preview":
        if st.session_state.selected_report_id:
            display_report_preview(st.session_state.selected_report_id)
        else:
            st.error("❌ 未选择报告")
            st.session_state.view_mode = "list"
            st.rerun()
    
    elif st.session_state.view_mode == "export":
        if st.session_state.selected_report_id:
            display_export_page(st.session_state.selected_report_id)
        else:
            st.session_state.view_mode = "list"
            st.rerun()

if __name__ == "__main__":
    main()