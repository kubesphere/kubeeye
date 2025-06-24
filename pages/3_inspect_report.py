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

# 添加自定义CSS样式
st.markdown("""
<style>
    /* 优化表格行的间距和对齐 */
    .row-container {
        padding: 8px 0;
        border-bottom: 1px solid #e0e0e0;
    }
    
    /* 小按钮样式 */
    .stButton > button {
        padding: 4px 8px;
        font-size: 12px;
        height: 28px;
        margin: 1px;
    }
    
    /* 表格标题样式 */
    .table-header {
        font-weight: bold;
        padding: 8px 0;
        border-bottom: 2px solid #ddd;
        background-color: #f8f9fa;
    }
    
    /* 状态标签样式 */
    .status-normal {
        color: #28a745;
        font-weight: bold;
    }
    
    .status-error {
        color: #dc3545;
        font-weight: bold;
    }
    
    /* 数字高亮样式 */
    .number-highlight {
        font-weight: bold;
        color: #007bff;
    }
</style>
""", unsafe_allow_html=True)

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

def safe_display_opa_violations_table(violations_data, show_expander=False, table_key=None):
    """安全地调用display_opa_violations_table函数，处理参数兼容性"""
    try:
        # 尝试使用新的参数签名
        if table_key:
            return display_opa_violations_table(violations_data, show_expander=show_expander, table_key=table_key)
        else:
            return display_opa_violations_table(violations_data, show_expander=show_expander)
    except TypeError:
        # 如果失败，使用旧的参数签名
        return display_opa_violations_table(violations_data, show_expander=show_expander)

def display_reports_overview():
    """显示报告概览页面"""
    st.markdown("查看和管理所有集群的巡检报告，快速识别问题并获取解决建议。")
    st.caption("💡 系统自动保留最近30天的报告，超出限制的旧报告会被自动清理以节省存储空间")
    
    # 加载数据
    clusters = list_clusters()
    all_results = list_results()
    
    # 检查是否有最新的巡检结果需要显示
    if hasattr(st.session_state, 'last_result_path') and st.session_state.last_result_path:
        st.success(f"🆕 检测到最新巡检结果: {os.path.basename(st.session_state.last_result_path)}")
        if st.button("🔍 查看最新结果", type="primary"):
            # 尝试从文件路径或会话状态中获取正确的result_id
            result_id = None
            
            # 方法1: 如果有last_result_id，直接使用
            if hasattr(st.session_state, 'last_result_id'):
                result_id = st.session_state.last_result_id
            
            # 方法2: 从最新的巡检结果列表中获取
            if not result_id and all_results:
                # 获取最新的结果（按时间排序后的第一个）
                latest_result = max(all_results, key=lambda x: x['timestamp'])
                result_id = latest_result['result_id']
            
            # 方法3: 尝试从文件名解析
            if not result_id and st.session_state.last_result_path:
                try:
                    # 从文件路径加载数据获取result_id
                    with open(st.session_state.last_result_path, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                    result_id = data.get('result_id')
                except Exception:
                    pass
            
            if result_id:
                st.session_state.selected_report_id = result_id
                st.session_state.view_mode = "detail"
                st.rerun()
            else:
                st.error("❌ 无法获取最新结果的ID，请从报告列表中选择")
    
    if not all_results:
        st.info("📭 暂无巡检报告。请先执行巡检任务生成报告。")
        if st.button("🚀 去执行巡检", type="primary"):
            st.switch_page("pages/2_cluster_inspect.py")
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
    
    # 计算汇总数据 - 使用简化状态系统
    total_reports = len(filtered_results)
    total_exceptions = sum(r['critical'] + r['warning'] for r in filtered_results)  # 所有异常
    total_passed = sum(r['passed'] for r in filtered_results)
    
    # 计算最新报告时间
    latest_report = max(filtered_results, key=lambda x: x['timestamp'])
    latest_time = datetime.fromisoformat(latest_report['timestamp']).strftime('%m-%d %H:%M')
    
    # 显示关键指标卡片
    col1, col2, col3, col4 = st.columns(4)
    
    with col1:
        st.metric("📄 报告总数", total_reports)
    
    with col2:
        st.metric("⚠️ 异常", total_exceptions, 
                  delta=f"-{total_exceptions}" if total_exceptions > 0 else None,
                  delta_color="inverse")
    
    with col3:
        st.metric("✅ 通过", total_passed,
                  delta=f"+{total_passed}" if total_passed > 0 else None)
    
    with col4:
        st.metric("🕒 最新报告", latest_time)

def display_reports_table(filtered_results):
    """显示报告列表 - 使用点击跳转方式"""
    st.markdown("### 📋 报告列表")
    st.markdown("💡 *点击报告ID查看详情和进行操作*")
    
    if not filtered_results:
        st.info("📭 暂无符合条件的报告")
        return
    
    # 按状态严重程度和时间排序
    sorted_results = sorted(filtered_results, 
                           key=lambda x: (-(x['critical'] + x['warning']), x['timestamp']), 
                           reverse=True)
    
    # 准备DataFrame数据
    df_data = []
    for result in sorted_results:
        timestamp = datetime.fromisoformat(result['timestamp'])
        total_exceptions = result['critical'] + result['warning']
        
        # 不再截断报告ID，让自动列宽处理
        report_id = result['result_id']
        
        df_data.append({
            "📄 报告ID": report_id,
            "🏢 集群": result['cluster_name'],
            "⏰ 巡检时间": timestamp.strftime('%m-%d %H:%M'),
            "🔄 类型": "⚡ 立即" if result['inspection_type'] == 'immediate' else "⏲️ 定时",
            "📊 状态": "🔴 异常" if total_exceptions > 0 else "🟢 正常",
            "⚠️ 异常": total_exceptions,
            "✅ 通过": result['passed'],
            "📈 总计": total_exceptions + result['passed']
        })
    
    # 创建DataFrame
    df = pd.DataFrame(df_data)
    
    # 使用事件选择模式的数据表格
    event = st.dataframe(
        df,
        use_container_width=True,
        hide_index=True,
        on_select="rerun",
        selection_mode="single-row",
        column_config={
            "📄 报告ID": st.column_config.TextColumn("📄 报告ID", help="点击行查看详情和操作"),
            "🏢 集群": st.column_config.TextColumn("🏢 集群"),
            "⏰ 巡检时间": st.column_config.TextColumn("⏰ 巡检时间"),
            "🔄 类型": st.column_config.TextColumn("🔄 类型"),
            "📊 状态": st.column_config.TextColumn("📊 状态"),
            "⚠️ 异常": st.column_config.NumberColumn("⚠️ 异常"),
            "✅ 通过": st.column_config.NumberColumn("✅ 通过"),
            "📈 总计": st.column_config.NumberColumn("📈 总计")
        }
    )
    
    # 处理行选择事件
    if len(event.selection.rows) > 0:
        selected_row = event.selection.rows[0]
        selected_result = sorted_results[selected_row]
        
        # 跳转到报告详情页面
        st.session_state.selected_report_id = selected_result['result_id']
        st.session_state.view_mode = "operations"
        st.rerun()

def delete_report(report_id):
    """删除报告文件"""
    import os
    from pathlib import Path
    
    # 查找报告文件
    results_dir = Path(__file__).parent.parent / "data" / "results"
    for file_path in results_dir.glob("*.json"):
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                data = json.load(f)
            if data.get('result_id') == report_id:
                os.remove(file_path)
                return True
        except:
            continue
    return False

# 导入导出函数
from utils.inspection_result import export_report

def display_report_operations(report_id):
    """显示报告操作页面"""
    # 返回按钮
    if st.button("⬅️ 返回报告列表"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    # 加载报告数据
    report_data = load_result(report_id)
    if not report_data:
        st.error("❌ 无法加载报告数据")
        return
    
    # 报告头部信息
    st.markdown(f"### 📄 报告操作中心")
    
    # 基本信息卡片
    col1, col2 = st.columns([3, 1])
    with col1:
        timestamp = datetime.fromisoformat(report_data['timestamp'])
        st.markdown(f"""
        **🏷️ 报告ID:** `{report_id}`  
        **🏢 集群:** {report_data['cluster_name']}  
        **⏰ 巡检时间:** {timestamp.strftime('%Y年%m月%d日 %H:%M:%S')}  
        **📋 类型:** {'⚡ 立即巡检' if report_data['inspection_type'] == 'immediate' else '⏲️ 定时巡检'}
        """)
    
    with col2:
        # 快速统计
        if 'inspection_results' in report_data:
            # 新数据结构
            all_items = []
            for inspector_type, inspector_result in report_data['inspection_results'].items():
                items = inspector_result.get('items', [])
                all_items.extend(items)
        else:
            # 旧数据结构
            all_items = report_data.get('items', [])
        
        exception_count = 0
        passed_count = 0
        
        for item in all_items:
            if isinstance(item, dict):
                status = item.get('status', 'unknown')
            elif hasattr(item, 'status'):
                status = getattr(item, 'status', 'unknown')
            else:
                status = 'unknown'
            
            if status == 'passed':
                passed_count += 1
            elif status == 'exception':
                exception_count += 1
        
        if exception_count > 0:
            st.error(f"🔴 异常: {exception_count}")
        else:
            st.success("🟢 全部通过")
        st.info(f"📊 总计: {len(all_items)}")
    
    st.divider()
    
    # 简化的操作按钮区域
    st.markdown("### 🔧 操作")
    
    # 主要操作 - 单行布局
    col1, col2, col3 = st.columns(3)
    
    with col1:
        if st.button("🔍 查看详情", type="primary", use_container_width=True):
            st.session_state.view_mode = "detail"
            st.rerun()
    
    with col2:
        if st.button("📄 导出JSON", use_container_width=True):
            export_and_download(report_id, "json", "JSON")
    
    with col3:
        if st.button("📊 导出Excel", use_container_width=True):
            export_and_download(report_id, "excel", "Excel")
    
    # 删除操作单独一行
    st.markdown("#### 🗑️ 删除操作")
    col1, col2 = st.columns([3, 1])
    
    with col1:
        st.caption("⚠️ 删除操作不可恢复，请谨慎操作")
    
    with col2:
        # 删除按钮
        confirm_key = f"confirm_delete_{report_id}"
        if st.session_state.get(confirm_key, False):
            if st.button("❌ 确认删除", type="primary", use_container_width=True):
                try:
                    delete_report(report_id)
                    st.success(f"✅ 已删除报告: {report_id}")
                    if confirm_key in st.session_state:
                        del st.session_state[confirm_key]
                    st.session_state.view_mode = "list"
                    st.rerun()
                except Exception as e:
                    st.error(f"❌ 删除失败: {str(e)}")
        else:
            if st.button("🗑️ 删除", use_container_width=True):
                st.session_state[confirm_key] = True
                st.rerun()
    
    # 删除确认提示
    if st.session_state.get(confirm_key, False):
        st.warning("⚠️ 点击确认删除按钮触发删除操作")

def display_report_detail(report_id):
    """显示报告详情"""
    # 导航返回按钮
    if st.button("⬅️ 返回操作页面"):
        st.session_state.view_mode = "operations"
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
        **📋 类型:** {'⚡ 立即巡检' if report_data['inspection_type'] == 'immediate' else '⏲️ 定时巡检'}
        """)
    
    with col2:
        # 快速统计 - 使用简化状态系统
        if 'inspection_results' in report_data:
            # 新数据结构：从inspection_results中统计
            all_items = []
            for inspector_type, inspector_result in report_data['inspection_results'].items():
                items = inspector_result.get('items', [])
                all_items.extend(items)
        else:
            # 旧数据结构：直接从items字段
            all_items = report_data.get('items', [])
        
        exception_count = 0
        passed_count = 0
        
        for item in all_items:
            if isinstance(item, dict):
                status = item.get('status', 'unknown')
            elif hasattr(item, 'status'):
                status = getattr(item, 'status', 'unknown')
            else:
                status = 'unknown'
            
            if status == 'passed':
                passed_count += 1
            elif status == 'exception':
                exception_count += 1
        
        if exception_count > 0:
            st.error(f"🔴 发现 {exception_count} 个异常")
        else:
            st.success(f"🟢 所有检查通过 ({passed_count} 项)")
    
    st.divider()
    
    # 检查项详情 - 传递处理后的所有项目
    display_inspection_items(all_items, report_id=report_id)

def display_report_preview(report_id):
    """显示报告快速预览 - 简化状态系统版本"""
    if st.button("⬅️ 返回"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    st.markdown(f"## 👁️ 报告预览 - {report_id}")
    
    # 加载数据并显示简化版本
    report_data = load_result(report_id)
    if not report_data:
        st.error("❌ 无法加载报告数据")
        return
    
    # 处理新的数据结构
    if 'inspection_results' in report_data:
        # 新数据结构：从inspection_results中获取所有项目
        all_items = []
        for inspector_type, inspector_result in report_data['inspection_results'].items():
            items = inspector_result.get('items', [])
            all_items.extend(items)
    else:
        # 旧数据结构：直接从items字段
        all_items = report_data.get('items', [])
    
    # 使用简化状态系统统计
    exception_critical = []
    exception_warning = []
    exception_other = []
    passed_items = []
    
    for item in all_items:
        if isinstance(item, dict):
            status = item.get('status', 'unknown')
            severity = item.get('severity', 'unknown')
        elif hasattr(item, 'status'):
            status = getattr(item, 'status', 'unknown')
            severity = getattr(item, 'severity', 'unknown')
        else:
            continue
            
        if status == 'passed':
            passed_items.append(item)
        elif status == 'exception':
            if severity == 'critical':
                exception_critical.append(item)
            elif severity == 'warning':
                exception_warning.append(item)
            else:
                exception_other.append(item)
    
    # 显示统计信息
    col1, col2 = st.columns(2)
    with col1:
        if exception_critical:
            st.error(f"🔴 严重异常: {len(exception_critical)}")
        if exception_warning:
            st.warning(f"🟡 一般异常: {len(exception_warning)}")
        if exception_other:
            st.info(f"ℹ️ 其他异常: {len(exception_other)}")
    
    with col2:
        st.success(f"✅ 通过: {len(passed_items)}")
        st.info(f"📊 总计: {len(all_items)}")
    
    # 只显示严重异常项
    if exception_critical:
        st.markdown("### 🚨 严重异常")
        for item in exception_critical[:5]:  # 最多显示5个
            st.error(f"**{item.get('name', '未知')}:** {item.get('description', '')}")
        
        if len(exception_critical) > 5:
            st.info(f"还有 {len(exception_critical) - 5} 个严重异常，查看完整报告了解详情")
    
    # 显示一般异常项（如果没有严重异常）
    elif exception_warning:
        st.markdown("### ⚠️ 一般异常")
        for item in exception_warning[:3]:  # 最多显示3个
            st.warning(f"**{item.get('name', '未知')}:** {item.get('description', '')}")
        
        if len(exception_warning) > 3:
            st.info(f"还有 {len(exception_warning) - 3} 个一般异常，查看完整报告了解详情")
    
    if st.button("📋 查看完整报告", type="primary"):
        st.session_state.view_mode = "detail"
        st.rerun()

def display_inspection_items(items, report_id=None):
    """显示巡检项详情 - 简化状态体系版本"""
    # 按新的状态体系分类：通过 vs 异常（按严重程度细分）
    passed_items = [item for item in items if item.get('status') == 'passed']
    exception_critical = [item for item in items 
                         if item.get('status') == 'exception' and item.get('severity') == 'critical']
    exception_warning = [item for item in items 
                        if item.get('status') == 'exception' and item.get('severity') == 'warning']
    exception_info = [item for item in items 
                     if item.get('status') == 'exception' and item.get('severity') in ['info', 'error'] or 
                     (item.get('status') == 'exception' and item.get('severity') not in ['critical', 'warning'])]
    
    # 创建标签页
    tab_names = []
    tab_data = []
    
    # 优先显示异常项
    if exception_critical:
        tab_names.append(f"🔴 严重异常 ({len(exception_critical)})")
        tab_data.append(exception_critical)
    
    if exception_warning:
        tab_names.append(f"🟡 一般异常 ({len(exception_warning)})")
        tab_data.append(exception_warning)
    
    if exception_info:
        tab_names.append(f"ℹ️ 其他异常 ({len(exception_info)})")
        tab_data.append(exception_info)
    
    # 最后显示通过项
    if passed_items:
        tab_names.append(f"✅ 通过 ({len(passed_items)})")
        tab_data.append(passed_items)
    
    if tab_names:
        tabs = st.tabs(tab_names)
        for i, (tab, data) in enumerate(zip(tabs, tab_data)):
            with tab:
                display_items_list(data, tab_names[i].startswith("✅"), report_id=report_id, tab_name=tab_names[i])

def display_items_list(items, is_passed=False, report_id=None, tab_name=None):
    """显示检查项列表 - 简化状态系统版本"""
    if not items:
        st.info("此类别下暂无项目")
        return
    
    # 对于通过的项目，默认折叠显示
    for idx, item in enumerate(items):
        title = f"{item.get('name', '未知检查项')}"
        expanded = not is_passed and item.get('severity') == 'critical'
        with st.expander(title, expanded=expanded):
            col1, col2 = st.columns([3, 1])
            with col1:
                st.markdown(f"**描述:** {item.get('description', '无')}")
                details = item.get('details', '')
                if details:
                    if 'violations' in item and isinstance(item['violations'], list):
                        st.markdown("**违规资源:**")
                        import hashlib
                        # 用 name, description, report_id, tab_name, idx 生成稳定唯一的key
                        # 加入description确保即使name相同也能区分
                        base = f"{item.get('name','')}_{item.get('description','')[:50]}_{report_id or ''}_{tab_name or ''}_{idx}"
                        item_key = hashlib.md5(base.encode()).hexdigest()[:12]
                        safe_display_opa_violations_table(item['violations'], show_expander=False, table_key=item_key)
                    else:
                        st.markdown("**详细信息:**")
                        st.text(details)
            with col2:
                status = item.get('status', 'unknown')
                severity = item.get('severity', 'info')
                if status == 'exception':
                    if severity == 'critical':
                        st.error("🔴 严重异常")
                    elif severity == 'warning':
                        st.warning("🟡 一般异常")
                    else:
                        st.info("ℹ️ 其他异常")
                elif status == 'passed':
                    st.success("✅ 通过")
                else:
                    st.info("ℹ️ 未知状态")
            solution = item.get('solution', '')
            if solution:
                st.markdown("**💡 建议解决方案:**")
                st.info(solution)

def display_export_page(report_id):
    """显示导出页面 - 优化版本"""
    if st.button("⬅️ 返回"):
        st.session_state.view_mode = "list"
        st.rerun()
    
    st.markdown(f"### 📥 导出报告 - {report_id}")
    
    # 提供快速导出选项
    st.markdown("#### 🚀 快速导出")
    col1, col2 = st.columns(2)
    
    with col1:
        if st.button("📄 导出 JSON", use_container_width=True):
            export_and_download(report_id, "json", "JSON")
    
    with col2:
        if st.button("📊 导出 Excel", use_container_width=True):
            export_and_download(report_id, "excel", "Excel")
    
    st.divider()
    
    # 自定义导出选项
    st.markdown("#### ⚙️ 自定义导出")
    col1, col2 = st.columns([2, 1])
    
    with col1:
        export_format = st.selectbox(
            "选择导出格式",
            ["JSON", "Excel"],
            help="选择您希望导出的报告格式"
        )
        
        # 导出选项
        include_passed = st.checkbox("包含通过的检查项", value=False, 
                                   help="默认只导出异常项，勾选此项将包含所有检查结果")
        
        include_details = st.checkbox("包含详细信息", value=True,
                                    help="是否包含详细的错误信息和解决方案")
    
    with col2:
        st.markdown("**预览导出内容**")
        report_data = load_result(report_id)
        if report_data:
            total_items = 0
            exception_items = 0
            
            # 统计项目数量
            if 'inspection_results' in report_data:
                for inspector_type, inspector_result in report_data['inspection_results'].items():
                    items = inspector_result.get('items', [])
                    total_items += len(items)
                    exception_items += len([item for item in items if item.get('status') != 'passed'])
            
            st.metric("总检查项", total_items)
            st.metric("异常项", exception_items)
            
            if include_passed:
                st.info(f"将导出 {total_items} 项")
            else:
                st.info(f"将导出 {exception_items} 项异常")
    
    if st.button("📥 开始自定义导出", type="primary"):
        export_and_download(report_id, export_format.lower(), export_format, 
                          include_passed, include_details)

def export_and_download(report_id, format_type, format_name, include_passed=False, include_details=True):
    """执行导出并提供下载"""
    try:
        from utils.inspection_result import export_report
        
        # 这里可以根据 include_passed 和 include_details 参数来定制导出内容
        # 暂时使用现有的导出功能
        success, file_path = export_report(report_id, format_type)
        
        if success:
            st.success(f"✅ {format_name} 导出成功!")
            
            # 提供下载按钮
            try:
                with open(file_path, "rb") as f:
                    file_data = f.read()
                
                # 根据格式设置MIME类型
                mime_types = {
                    "json": "application/json",
                    "excel": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                }
                
                st.download_button(
                    label=f"⬇️ 下载 {format_name} 文件",
                    data=file_data,
                    file_name=os.path.basename(file_path),
                    mime=mime_types.get(format_type, "application/octet-stream"),
                    type="secondary",
                    use_container_width=True
                )
                
                # 显示文件信息
                file_size = len(file_data) / 1024  # KB
                st.caption(f"文件大小: {file_size:.1f} KB | 路径: {file_path}")
                
            except Exception as e:
                st.error(f"❌ 读取文件失败: {str(e)}")
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
    
    elif st.session_state.view_mode == "operations":
        if st.session_state.selected_report_id:
            display_report_operations(st.session_state.selected_report_id)
        else:
            st.error("❌ 未选择报告")
            st.session_state.view_mode = "list"
            st.rerun()
    
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