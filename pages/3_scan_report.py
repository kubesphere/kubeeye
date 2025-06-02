#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 集群巡检报告页面
"""

# 导入必要的库
import streamlit as st
import pandas as pd
from datetime import datetime
import json
import os
import re
import sys
from pathlib import Path
import plotly.express as px
import plotly.graph_objects as go

# 设置页面配置 - 必须是第一个Streamlit命令
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

# 初始化页面
initialize_page(
    title="巡检报告", 
    icon="📊",
    page_title="巡检报告中心",
    page_subtitle="查看和分析巡检结果"
)
st.markdown("""
<style>
    .main-header {color:#2E86C1; font-size:24px; font-weight:bold; margin-bottom:20px;}
    .sub-header {color:#3498DB; font-size:18px; margin-top:20px; margin-bottom:10px;}
    .status-warning {padding:5px; border-radius:5px; background-color:#FFF0D9; color:#D4A76A; font-weight:bold;}
    .status-passed {padding:5px; border-radius:5px; background-color:#E5F5EE; color:#5AAF8C; font-weight:bold;}
    .report-card {border:1px solid #ddd; border-radius:8px; padding:15px; margin-bottom:15px;}
    .report-summary {background-color:#F8F9F9; padding:10px; border-radius:5px;}
</style>
""", unsafe_allow_html=True)

st.markdown("<p>在这里查看和分析集群巡检结果，识别潜在问题并获取解决建议。</p>", unsafe_allow_html=True)

# 加载集群列表和巡检结果
clusters = list_clusters()
all_results = list_results()

# 创建美观的过滤区域
st.markdown("<div class='main-header'>报告过滤</div>", unsafe_allow_html=True)
filter_col1, filter_col2 = st.columns(2)

with filter_col1:
    selected_cluster = st.selectbox(
        "选择集群", 
        ["全部"] + clusters,
        help="选择要查看的特定集群，或查看所有集群结果"
    )

with filter_col2:
    inspection_types = ["全部", "unified", "node", "prometheus", "opa"]
    selected_type = st.selectbox(
        "巡检类型",
        inspection_types,
        index=0,  # 默认选择"全部"
        help="选择要查看的巡检类型",
        format_func=lambda x: "综合报告" if x == "unified" else ("节点检查" if x == "node" else 
                            ("Prometheus检查" if x == "prometheus" else 
                            ("合规性检查" if x == "opa" else x)))
    )

# 筛选结果
filtered_results = all_results
if selected_cluster != "全部":
    filtered_results = [r for r in filtered_results if r["cluster_name"] == selected_cluster]
if selected_type != "全部":
    filtered_results = [r for r in filtered_results if r["inspection_type"] == selected_type]

# 显示巡检摘要卡片
st.markdown("<div class='main-header'>巡检摘要</div>", unsafe_allow_html=True)

if not filtered_results:
    st.info("⚠️ 没有找到符合条件的巡检结果")
else:
    # 汇总数据
    total_critical = sum(r['critical'] for r in filtered_results)
    total_warning = sum(r['warning'] for r in filtered_results)
    total_passed = sum(r['passed'] for r in filtered_results)
    total_reports = len(filtered_results)
    
    summary_col1, summary_col2, summary_col3, summary_col4 = st.columns(4)
    
    with summary_col1:
        st.metric(
            label="巡检报告总数", 
            value=total_reports,
            delta=None,
            delta_color="off"
        )
        
    with summary_col2:
        st.metric(
            label="关键问题", 
            value=total_critical,
            delta=None,
            delta_color="inverse"
        )
        
    with summary_col3:
        st.metric(
            label="警告", 
            value=total_warning,
            delta=None,
            delta_color="inverse"
        )
        
    with summary_col4:
        st.metric(
            label="通过检查", 
            value=total_passed,
            delta=None,
            delta_color="normal"
        )
    
    # 准备数据
    result_data = []
    for result in filtered_results:
        timestamp = datetime.fromisoformat(result['timestamp']).strftime("%Y-%m-%d %H:%M")
        
        # 创建状态文本
        status_text = "正常"
        status_color = "normal"
        if result['critical'] > 0:
            status_text = "严重问题"
            status_color = "critical"
        elif result['warning'] > 0:
            status_text = "需要注意"
            status_color = "warning"
            
        inspection_type_display = "综合报告" if result['inspection_type'] == "unified" else (
            "节点检查" if result['inspection_type'] == "node" else (
            "Prometheus检查" if result['inspection_type'] == "prometheus" else (
            "合规性检查" if result['inspection_type'] == "opa" else result['inspection_type'])))
        
        result_data.append({
            "巡检ID": result['result_id'],
            "集群名称": result['cluster_name'],
            "巡检类型": inspection_type_display,
            "时间": timestamp,
            "状态": status_text,
            "关键问题": result['critical'],
            "警告": result['warning'],
            "通过": result['passed']
        })
    
    # 使用更美观的数据表格
    st.markdown("<div class='sub-header'>巡检结果列表</div>", unsafe_allow_html=True)
    result_df = pd.DataFrame(result_data)
    
    # 添加交互式表格和查看详情功能
    selected_indices = st.dataframe(
        result_df,
        use_container_width=True,
        column_config={
            "巡检ID": st.column_config.TextColumn("巡检ID", width="medium"),
            "集群名称": st.column_config.TextColumn("集群名称", width="small"),
            "巡检类型": st.column_config.TextColumn("巡检类型", width="small"),
            "时间": st.column_config.TextColumn("时间", width="medium"),
            "状态": st.column_config.TextColumn("状态", width="small"),
            "关键问题": st.column_config.NumberColumn(
                "关键问题", 
                width="small", 
                format="%d",
                help="严重问题数量"
            ),
            "警告": st.column_config.NumberColumn(
                "警告", 
                width="small", 
                format="%d",
                help="警告数量"
            ),
            "通过": st.column_config.NumberColumn(
                "通过", 
                width="small", 
                format="%d",
                help="通过检查项数量"
            )
        },
        height=250,
        hide_index=True
    )
    
    # 显示统计图
    st.markdown("<div class='main-header'>巡检结果趋势</div>", unsafe_allow_html=True)
    
    col1, col2 = st.columns(2)
    
    if filtered_results:
        # 按集群统计问题数
        with col1:
            cluster_stats = {}
            for result in filtered_results:
                cluster_name = result['cluster_name']
                if cluster_name not in cluster_stats:
                    cluster_stats[cluster_name] = {
                        "critical": 0,
                        "warning": 0,
                        "passed": 0
                    }
                cluster_stats[cluster_name]["critical"] += result['critical']
                cluster_stats[cluster_name]["warning"] += result['warning']
                cluster_stats[cluster_name]["passed"] += result['passed']
            
            # 准备绘图数据
            chart_data = []
            for cluster, stats in cluster_stats.items():
                chart_data.append({"集群": cluster, "类型": "关键问题", "数量": stats["critical"]})
                chart_data.append({"集群": cluster, "类型": "警告", "数量": stats["warning"]})
                chart_data.append({"集群": cluster, "类型": "通过", "数量": stats["passed"]})
            
            chart_df = pd.DataFrame(chart_data)
            
            fig = px.bar(
                chart_df, 
                x="集群", 
                y="数量", 
                color="类型",
                color_discrete_map={"关键问题": "#FF7979", "警告": "#FFCA7A", "通过": "#7DD39F"},
                title="集群问题统计",
                template="plotly_white"
            )
            fig.update_layout(
                legend=dict(
                    orientation="h",
                    yanchor="bottom",
                    y=1.02,
                    xanchor="right",
                    x=1
                ),
                yaxis_title="检查项数量",
                xaxis_title="集群名称",
                barmode='stack'
            )
            st.plotly_chart(fig, use_container_width=True)
        
        # 按巡检类型统计问题数
        with col2:
            type_stats = {}
            for result in filtered_results:
                inspection_type = result['inspection_type']
                if inspection_type not in type_stats:
                    type_stats[inspection_type] = {
                        "critical": 0,
                        "warning": 0,
                        "passed": 0
                    }
                type_stats[inspection_type]["critical"] += result['critical']
                type_stats[inspection_type]["warning"] += result['warning']
                type_stats[inspection_type]["passed"] += result['passed']
            
            # 准备绘图数据
            chart_data = []
            for insp_type, stats in type_stats.items():
                chart_data.append({"类型": insp_type, "状态": "关键问题", "数量": stats["critical"]})
                chart_data.append({"类型": insp_type, "状态": "警告", "数量": stats["warning"]})
                chart_data.append({"类型": insp_type, "状态": "通过", "数量": stats["passed"]})
            
            chart_df = pd.DataFrame(chart_data)
            
            fig = px.bar(
                chart_df, 
                x="类型", 
                y="数量", 
                color="状态",
                color_discrete_map={"关键问题": "#FF7979", "警告": "#FFCA7A", "通过": "#7DD39F"},
                title="巡检类型统计",
                template="plotly_white"
            )
            fig.update_layout(
                legend=dict(
                    orientation="h",
                    yanchor="bottom",
                    y=1.02,
                    xanchor="right",
                    x=1
                ),
                yaxis_title="检查项数量",
                xaxis_title="巡检类型",
                barmode='stack'
            )
            st.plotly_chart(fig, use_container_width=True)
    
    # 查看详细报告
    st.markdown("<div class='main-header'>报告详情</div>", unsafe_allow_html=True)
    
    # 选择要查看的报告
    if result_data:
        selected_report_id = st.selectbox(
            "选择巡检报告查看详情", 
            options=[r["巡检ID"] for r in result_data],
            format_func=lambda x: f"{x} ({next(r['集群名称'] for r in result_data if r['巡检ID'] == x)} - {next(r['时间'] for r in result_data if r['巡检ID'] == x)})"
        )
        
        if selected_report_id:
            # 加载详细报告
            detailed_result = load_result(selected_report_id)
            
            if detailed_result:
                # 使用卡片式布局显示报告头部信息
                inspection_type_display = "综合报告" if detailed_result['inspection_type'] == "unified" else (
                    "节点检查" if detailed_result['inspection_type'] == "node" else (
                    "Prometheus检查" if detailed_result['inspection_type'] == "prometheus" else (
                    "合规性检查" if detailed_result['inspection_type'] == "opa" else detailed_result['inspection_type'])))
                
                st.markdown(f"""
                <div class="report-card">
                    <h3>📄 {detailed_result['cluster_name']} 集群巡检报告</h3>
                    <div class="report-summary">
                        <p><strong>报告ID:</strong> {selected_report_id}</p>
                        <p><strong>集群名称:</strong> {detailed_result['cluster_name']}</p>
                        <p><strong>巡检类型:</strong> {inspection_type_display}</p>
                        <p><strong>巡检时间:</strong> {datetime.fromisoformat(detailed_result['timestamp']).strftime('%Y-%m-%d %H:%M:%S')}</p>
                    </div>
                </div>
                """, unsafe_allow_html=True)
                
                # 统计问题数量
                critical_count = sum(1 for item in detailed_result['items'] if item['status'] == 'failed' and item['severity'] == 'critical')
                warning_count = sum(1 for item in detailed_result['items'] if item['status'] == 'failed' and item['severity'] == 'warning')
                passed_count = sum(1 for item in detailed_result['items'] if item['status'] == 'passed')
                
                # 创建仪表板布局
                dashboard_col1, dashboard_col2 = st.columns([3, 2])
                
                with dashboard_col1:
                    # 创建更美观的饼图
                    fig = go.Figure(data=[go.Pie(
                        labels=['关键问题', '警告', '通过'],
                        values=[critical_count, warning_count, passed_count],
                        hole=.4,
                        marker_colors=['#FF7979', '#FFCA7A', '#7DD39F'],
                        textinfo='label+percent',
                        pull=[0.1 if critical_count > 0 else 0, 0.05 if warning_count > 0 else 0, 0]
                    )])
                    fig.update_layout(
                        title="检查结果分布",
                        title_font_size=18,
                        legend=dict(
                            orientation="h",
                            yanchor="bottom",
                            y=-0.2,
                            xanchor="center",
                            x=0.5
                        )
                    )
                    st.plotly_chart(fig, use_container_width=True)
                
                with dashboard_col2:
                    # 创建更美观的指标统计卡片
                    st.markdown("<div style='text-align:center; font-size:18px; font-weight:bold; margin-bottom:10px;'>巡检结果统计</div>", unsafe_allow_html=True)
                    
                    # 总检查项数
                    total_items = critical_count + warning_count + passed_count
                    st.markdown(f"""
                    <div style='background-color:#F8F9F9; padding:10px; border-radius:5px; margin-bottom:10px; text-align:center;'>
                        <div style='font-size:16px;'>总检查项</div>
                        <div style='font-size:24px; font-weight:bold;'>{total_items}</div>
                    </div>
                    """, unsafe_allow_html=True)
                    
                    # 关键问题
                    st.markdown(f"""
                    <div style='background-color:#FFE5E5; padding:10px; border-radius:5px; margin-bottom:10px; text-align:center;'>
                        <div style='font-size:16px;'>关键问题</div>
                        <div style='font-size:24px; font-weight:bold; color:#D46A6A;'>{critical_count}</div>
                        <div style='font-size:12px;'>{round(critical_count/total_items*100 if total_items > 0 else 0, 1)}% 的检查项</div>
                    </div>
                    """, unsafe_allow_html=True)
                    
                    # 警告
                    st.markdown(f"""
                    <div style='background-color:#FFF0D9; padding:10px; border-radius:5px; margin-bottom:10px; text-align:center;'>
                        <div style='font-size:16px;'>警告</div>
                        <div style='font-size:24px; font-weight:bold; color:#D4A76A;'>{warning_count}</div>
                        <div style='font-size:12px;'>{round(warning_count/total_items*100 if total_items > 0 else 0, 1)}% 的检查项</div>
                    </div>
                    """, unsafe_allow_html=True)
                    
                    # 通过
                    st.markdown(f"""
                    <div style='background-color:#E5F5EE; padding:10px; border-radius:5px; text-align:center;'>
                        <div style='font-size:16px;'>通过检查</div>
                        <div style='font-size:24px; font-weight:bold; color:#5AAF8C;'>{passed_count}</div>
                        <div style='font-size:12px;'>{round(passed_count/total_items*100 if total_items > 0 else 0, 1)}% 的检查项</div>
                    </div>
                    """, unsafe_allow_html=True)
                
                # 添加报告导出功能
                st.markdown("<div class='sub-header'>导出报告</div>", unsafe_allow_html=True)
                
                export_col1, export_col2 = st.columns([4, 1])
                
                with export_col1:
                    export_format = st.selectbox(
                        "选择导出格式",
                        ["JSON", "CSV", "Excel"],
                        help="选择您希望导出的报告格式",
                        index=0
                    )
                
                with export_col2:
                    if st.button("📥 导出", use_container_width=True):
                        format_map = {"JSON": "json", "CSV": "csv", "Excel": "excel"}
                        format_type = format_map[export_format]
                        
                        # 尝试先安装依赖
                        if format_type == "excel" and export_format == "Excel":
                            st.info("准备导出 Excel 格式...")
                            import importlib
                            if importlib.util.find_spec("openpyxl") is None:
                                import subprocess
                                import sys
                                subprocess.check_call([sys.executable, "-m", "pip", "install", "openpyxl"])
                                st.success("已安装 openpyxl 依赖")
                        
                        from utils.inspection_result import export_report
                        success, file_path = export_report(selected_report_id, format_type)
                        
                        if success:
                            import base64
                            with open(file_path, "rb") as f:
                                file_data = f.read()
                                b64 = base64.b64encode(file_data).decode()
                                
                                file_name = file_path.split('/')[-1]
                                mime_types = {
                                    "json": "application/json",
                                    "csv": "text/csv",
                                    "excel": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                                }
                                mime_type = mime_types.get(format_type, "application/octet-stream")
                                
                                href = f'<a href="data:{mime_type};base64,{b64}" download="{file_name}">点击下载 {export_format} 报告</a>'
                                st.markdown(href, unsafe_allow_html=True)
                        else:
                            st.error(f"导出失败: {file_path}")
                
                # 显示详细巡检项
                st.markdown("<div class='main-header'>详细检查项</div>", unsafe_allow_html=True)
                
                # 筛选显示选项和搜索框
                filter_col1, filter_col2 = st.columns([3, 2])
                
                with filter_col1:
                    show_option = st.radio(
                        "筛选显示", 
                        ["全部检查项", "仅严重问题", "仅警告", "仅通过项"],
                        horizontal=True,
                        help="选择要显示的检查项类型"
                    )
                
                with filter_col2:
                    search_term = st.text_input("搜索检查项", placeholder="输入关键字搜索...", help="搜索检查项名称和描述")
                
                filtered_items = detailed_result['items']
                
                # 应用筛选条件
                if show_option == "仅严重问题":
                    filtered_items = [item for item in filtered_items if item['status'] == 'failed' and item['severity'] == 'critical']
                elif show_option == "仅警告":
                    filtered_items = [item for item in filtered_items if item['status'] == 'failed' and item['severity'] == 'warning']
                elif show_option == "仅通过项":
                    filtered_items = [item for item in filtered_items if item['status'] == 'passed']
                
                # 应用搜索
                if search_term:
                    filtered_items = [item for item in filtered_items if 
                                     search_term.lower() in item['name'].lower() or 
                                     search_term.lower() in item['description'].lower()]
                
                # 显示检查项计数
                st.markdown(f"<p>显示 {len(filtered_items)} 个检查项</p>", unsafe_allow_html=True)
                
                # 使用Tab分类显示不同严重程度的问题
                if len(filtered_items) > 0:
                    # 先按严重程度分组
                    critical_items = [item for item in filtered_items if item['status'] == 'failed' and item['severity'] == 'critical']
                    warning_items = [item for item in filtered_items if item['status'] == 'failed' and item['severity'] == 'warning']
                    passed_items = [item for item in filtered_items if item['status'] == 'passed']
                    
                    # 创建选项卡
                    tab_titles = [
                        f"严重问题 ({len(critical_items)})" if critical_items else "严重问题 (0)",
                        f"警告 ({len(warning_items)})" if warning_items else "警告 (0)",
                        f"通过项 ({len(passed_items)})" if passed_items else "通过项 (0)"
                    ]
                    
                    # 如果筛选后没有显示某一类项目，则不创建对应的空选项卡
                    tabs = []
                    if show_option in ["全部检查项", "仅严重问题"] and critical_items:
                        tabs.append("严重问题")
                    if show_option in ["全部检查项", "仅警告"] and warning_items:
                        tabs.append("警告")
                    if show_option in ["全部检查项", "仅通过项"] and passed_items:
                        tabs.append("通过项")
                    
                    if tabs:  # 确保至少有一个选项卡
                        selected_tab = st.radio("检查项分类", tabs, horizontal=True)
                        
                        # 根据选中的选项卡显示对应的项目
                        display_items = []
                        if selected_tab == "严重问题":
                            display_items = critical_items
                        elif selected_tab == "警告":
                            display_items = warning_items
                        else:
                            display_items = passed_items
                        
                        # 定义每种类型的样式
                        style_map = {
                            "critical": {
                                "icon": "❌",
                                "bg_color": "#FFE5E5",
                                "border_color": "#FF7979",
                                "text_color": "#D46A6A"
                            },
                            "warning": {
                                "icon": "⚠️",
                                "bg_color": "#FFF0D9",
                                "border_color": "#FFCA7A",
                                "text_color": "#D4A76A"
                            },
                            "info": {
                                "icon": "✅",
                                "bg_color": "#E5F5EE",
                                "border_color": "#7DD39F",
                                "text_color": "#5AAF8C"
                            }
                        }
                        
                        # 显示巡检项
                        for i, item in enumerate(display_items):
                            # 确定项目的样式
                            style = style_map["info"]  # 默认为通过样式
                            if item['status'] == 'failed':
                                if item['severity'] == 'critical':
                                    style = style_map["critical"]
                                else:
                                    style = style_map["warning"]
                            
                            # 获取检查项图标
                            status_icons = {
                                "critical": "❌",
                                "warning": "⚠️",
                                "info": "✅"
                            }
                            severity_text = {
                                "critical": "严重",
                                "warning": "警告",
                                "info": "正常"
                            }
                            
                            status_icon = status_icons.get(item['severity'], "❔")
                            node_info = ""
                            # 检查名称中是否包含节点IP
                            if " - " in item['name']:
                                parts = item['name'].split(" - ")
                                check_name = parts[0]
                                node_ip = parts[1]
                                # 使用st原生方式展示标题和节点信息，避免HTML注入问题
                                expander_title = f"{status_icon} {check_name}"
                            else:
                                check_name = item['name']
                                node_ip = None
                                expander_title = f"{status_icon} {check_name}"
                            
                            # 创建美观的展开面板
                            with st.expander(expander_title, expanded=(item['status'] == 'failed' and item['severity'] == 'critical')):
                                # 如果有节点信息，以标签形式显示
                                if node_ip:
                                    st.markdown(f"""
                                    <div style='margin-bottom:10px;'>
                                        <span style='background-color:#EBF0F5; padding:2px 6px; border-radius:4px; font-size:12px;'>节点: {node_ip}</span>
                                    </div>
                                    """, unsafe_allow_html=True)
                                st.markdown(f"""
                                <div style='padding:12px 15px; background-color:{style['bg_color']}; border-left:4px solid {style['border_color']}; border-radius:4px; margin-bottom:15px; box-shadow:0 1px 2px rgba(0,0,0,0.05);'>
                                    <div style='color:{style['text_color']}; font-weight:bold; margin-bottom:8px; font-size:15px;'>{item['description']}</div>
                                    <div style='font-size:12px; display:inline-block; padding:2px 8px; background-color:rgba(0,0,0,0.05); border-radius:10px;'>
                                        严重程度: <span style='font-weight:500;'>{severity_text.get(item['severity'], item['severity'])}</span>
                                    </div>
                                </div>
                                """, unsafe_allow_html=True)
                                
                                # 详细信息区域
                                st.markdown("<div style='font-weight:600; color:#2C3E50; margin-top:15px; margin-bottom:8px; font-size:16px;'>📋 详细信息</div>", unsafe_allow_html=True)
                                
                                # 处理详细信息的格式化显示 - 使用简洁清晰的方式展示，不添加进度条等复杂元素
                                details_content = item['details']
                                
                                # 根据内容类型选择不同的展示方式
                                if '\n' in details_content and ':' in details_content:
                                    # 键值对类型数据，使用简单的表格格式显示
                                    lines = details_content.strip().split('\n')
                                    
                                    # 检测是否是服务状态数据，以避免特殊处理
                                    is_service_status = any(x in details_content for x in ["服务状态", "service", "active", "inactive", "failed"])
                                    
                                    if is_service_status:
                                        # 服务状态数据直接以原始文本展示
                                        st.markdown(f"""
                                        <div style='background-color:#F8F9F9; padding:15px; border-radius:6px; margin-bottom:15px; border:1px solid #E9EEF2;'>
                                            <div style='color:#2C3E50; font-family:system-ui; white-space:pre-wrap;'>{details_content}</div>
                                        </div>
                                        """, unsafe_allow_html=True)
                                    else:
                                        # 其他键值对数据使用表格展示
                                        table_rows = []
                                        for line in lines:
                                            if ':' in line:
                                                key, value = line.split(':', 1)
                                                table_rows.append(f"""
                                                <tr>
                                                    <td style='padding:6px 10px; font-weight:500; text-align:right; color:#34495E; border-bottom:1px solid #EDF2F7;'>{key.strip()}</td>
                                                    <td style='padding:6px 10px; color:#2C3E50; border-bottom:1px solid #EDF2F7;'>{value.strip()}</td>
                                                </tr>
                                                """)
                                            else:
                                                table_rows.append(f"""
                                                <tr>
                                                    <td colspan='2' style='padding:6px 10px; color:#2C3E50; border-bottom:1px solid #EDF2F7;'>{line.strip()}</td>
                                                </tr>
                                                """)
                                        
                                        table_html = f"""
                                        <table style='width:100%; border-collapse:collapse;'>
                                            {''.join(table_rows)}
                                        </table>
                                        """
                                        
                                        st.markdown(f"""
                                        <div style='background-color:#F8F9F9; padding:15px; border-radius:6px; margin-bottom:15px; border:1px solid #E9EEF2; overflow-x:auto;'>
                                        {table_html}
                                        </div>
                                        """, unsafe_allow_html=True)
                                    
                                # 用于百分比类数据（但不添加进度条）
                                elif "%" in details_content and ("使用率" in details_content or "负载" in details_content):
                                    # 提取百分比值用于添加颜色标识，但不绘制进度条
                                    percent_match = re.search(r'(\d+\.?\d*)%', details_content)
                                    text_color = "#2C3E50"  # 默认文本颜色
                                    
                                    if percent_match:
                                        try:
                                            percent = float(percent_match.group(1))
                                            if percent >= 90:
                                                text_color = "#D46A6A"  # 高危值用红色
                                            elif percent >= 80:
                                                text_color = "#D4A76A"  # 警告值用黄色
                                        except ValueError:
                                            pass
                                            
                                    st.markdown(f"""
                                    <div style='background-color:#F8F9F9; padding:15px; border-radius:6px; margin-bottom:15px; font-family:system-ui; line-height:1.5; border:1px solid #E9EEF2;'>
                                        <div style='color:{text_color}; font-weight:500;'>{details_content}</div>
                                    </div>
                                    """, unsafe_allow_html=True)
                                        
                                # 服务状态数据
                                elif any(x in details_content for x in ["服务状态", "active", "inactive", "failed"]):
                                    # 简化显示逻辑，不做额外处理，直接显示原始文本
                                    # 使用样式较少的纯文本展示方式
                                    st.markdown(f"""
                                    <div style='background-color:#F8F9F9; padding:15px; border-radius:6px; margin-bottom:15px; border:1px solid #E9EEF2;'>
                                        <div style='color:#2C3E50; font-family:system-ui;'>{details_content}</div>
                                    </div>
                                    """, unsafe_allow_html=True)
                                else:
                                    # 普通文本显示 - 保持简单
                                    st.markdown(f"""
                                    <div style='background-color:#F8F9F9; padding:15px; border-radius:6px; margin-bottom:15px; font-family:system-ui; line-height:1.5; border:1px solid #E9EEF2; color:#2C3E50;'>
                                    {details_content}
                                    </div>
                                    """, unsafe_allow_html=True)
                                
                                # 解决方案区域
                                if item['solution']:
                                    st.markdown("<div style='font-weight:600; color:#2C3E50; margin-top:15px; margin-bottom:8px; font-size:16px;'>💡 建议解决方案</div>", unsafe_allow_html=True)
                                    
                                    # 尝试检测解决方案内容类型
                                    solution_content = item['solution']
                                    
                                    # 检查是否是多步骤解决方案
                                    if re.search(r'^\d+[\.\)]|\- ', solution_content, re.MULTILINE):
                                        # 处理有序列表格式 (1. 2. 等)
                                        steps = re.split(r'\n(?=\d+[\.\)]|\- )', solution_content)
                                        formatted_solution = ""
                                        for i, step in enumerate(steps):
                                            step = step.strip()
                                            if step:
                                                formatted_solution += f"""
                                                <div style='display:flex; margin-bottom:8px;'>
                                                    <div style='min-width:24px; height:24px; border-radius:50%; background-color:#3498DB; color:white; display:flex; align-items:center; justify-content:center; margin-right:10px; font-weight:bold;'>{i+1}</div>
                                                    <div style='flex:1; padding-top:2px;'>{step}</div>
                                                </div>
                                                """
                                                
                                        st.markdown(f"""
                                        <div style='background-color:#EBF5FB; padding:15px; border-radius:6px; margin-bottom:15px; line-height:1.6; color:#2C3E50;'>
                                            {formatted_solution}
                                        </div>
                                        """, unsafe_allow_html=True)
                                    
                                    # 检查是否包含多个命令（通常是以命令符号开头的行）
                                    elif re.search(r'(^|\n)(systemctl|kubectl|docker|grep|find|ps|mkdir|rm|cd|cp|mv|apt|yum|dnf|vim|curl|wget) ', solution_content):
                                        # 处理包含命令的解决方案
                                        commands = re.findall(r'(?:^|\n)(systemctl|kubectl|docker|grep|find|ps|mkdir|rm|cd|cp|mv|apt|yum|dnf|vim|curl|wget)([^\n]+)', solution_content)
                                        other_text = re.sub(r'(?:^|\n)(systemctl|kubectl|docker|grep|find|ps|mkdir|rm|cd|cp|mv|apt|yum|dnf|vim|curl|wget)([^\n]+)', '', solution_content).strip()
                                        
                                        formatted_solution = ""
                                        # 先添加普通文本描述（如果有）
                                        if other_text:
                                            formatted_solution += f"<p>{other_text}</p>"
                                        
                                        # 添加命令部分
                                        formatted_solution += "<div style='margin-top:10px;'>"
                                        for cmd, args in commands:
                                            formatted_solution += f"""
                                            <div style='background-color:#2C3E50; color:white; padding:8px 12px; border-radius:4px; font-family:monospace; margin-bottom:8px; overflow-x:auto;'>
                                                <span style='color:#F39C12;'>{cmd}</span><span style='color:#ECF0F1;'>{args}</span>
                                            </div>
                                            """
                                        formatted_solution += "</div>"
                                        
                                        st.markdown(f"""
                                        <div style='background-color:#EBF5FB; padding:15px; border-radius:6px; margin-bottom:15px; line-height:1.6; color:#2C3E50;'>
                                            {formatted_solution}
                                        </div>
                                        """, unsafe_allow_html=True)
                                    
                                    else:
                                        # 简单的单步骤解决方案
                                        st.markdown(f"""
                                        <div style='background-color:#EBF5FB; padding:15px; border-radius:6px; border-left:4px solid #3498DB; margin-bottom:15px; line-height:1.6; color:#2C3E50;'>
                                        <p>{solution_content}</p>
                                        </div>
                                        """, unsafe_allow_html=True)
                    else:
                        st.info("没有符合条件的检查项")
            else:
                st.error("找不到巡检报告详情")