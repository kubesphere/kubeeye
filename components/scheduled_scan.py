#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
定时巡检组件
"""
import streamlit as st
import pandas as pd
from utils.cluster_config import list_clusters, get_cluster
from utils.schedule_manager import (
    ScheduleTask, load_schedules, add_schedule, delete_schedule,
    run_inspection, restart_scheduler, start_scheduler, stop_scheduler
)

from .common import create_rule_selection_tabs, create_rule_selection_in_form
# 导入UI组件
from components.ui import display_cluster_info, InspectionProgress, display_inspection_results, display_summary_metrics

def render_scheduled_scan_tab():
    """渲染定时巡检标签页内容"""
    st.subheader("定时巡检")
    
    # 检查调度器状态
    if "scheduler_status" not in st.session_state:
        st.session_state.scheduler_status = True  # 假设默认是启动的
    
    # 控制调度器的按钮
    scheduler_col1, scheduler_col2, scheduler_col3 = st.columns([1, 1, 2])
    with scheduler_col1:
        if st.session_state.scheduler_status:
            if st.button("停止调度器", key="stop_scheduler"):
                if stop_scheduler():
                    st.session_state.scheduler_status = False
                    st.success("调度器已停止")
                    st.rerun()
                else:
                    st.error("停止调度器失败")
        else:
            if st.button("启动调度器", key="start_scheduler"):
                if start_scheduler():
                    st.session_state.scheduler_status = True
                    st.success("调度器已启动")
                    st.rerun()
                else:
                    st.error("启动调度器失败")
    
    with scheduler_col2:
        if st.button("重新加载任务", key="restart_scheduler"):
            if restart_scheduler():
                st.success("任务已重新加载")
                st.rerun()
            else:
                st.error("重新加载任务失败")
    
    with scheduler_col3:
        scheduler_status_text = "运行中" if st.session_state.scheduler_status else "已停止"
        scheduler_status_color = "#00a971" if st.session_state.scheduler_status else "#ff4b4b"
        st.markdown(f"""
        <div style="padding: 8px; border-radius: 4px; background-color: #f8f9fa; text-align: left;">
            <span>调度器状态: </span>
            <span style="color: {scheduler_status_color}; font-weight: bold;">
                {scheduler_status_text}
            </span>
        </div>
        """, unsafe_allow_html=True)
    
    # 加载已有的调度任务
    tasks = load_schedules()
    
    # 显示任务列表和创建新任务的选项卡
    task_tab1, task_tab2 = st.tabs(["任务列表", "创建新任务"])
    
    with task_tab1:
        render_task_list_tab(tasks)
    
    # 创建新任务标签页
    with task_tab2:
        render_create_task_tab()

def render_task_list_tab(tasks):
    """渲染任务列表标签页"""
    if not tasks:
        st.info("还没有创建任何定时巡检任务，请点击「创建新任务」标签页创建。")
    else:
        # 创建任务数据表
        task_data = []
        
        for task in tasks:
            # 准备状态指示器
            if task.last_status == "success":
                status_icon = "✅"
                status_text = "成功"
            elif task.last_status == "running":
                status_icon = "⏳"
                status_text = "运行中"
            elif task.last_status == "failed":
                status_icon = "❌"
                status_text = "失败"
            else:
                status_icon = "⏸️"
                status_text = "未运行"
            
            # 准备启用状态
            enabled_text = "启用" if task.enabled else "禁用"
            
            # 获取下次运行时间
            next_run = task.get_next_run()
            next_run_text = next_run.strftime("%Y-%m-%d %H:%M") if next_run else "未设置"
            
            # 准备调度类型
            schedule_type = task.get_pretty_schedule()
            
            # 添加到数据表
            task_data.append({
                "任务名称": task.name,
                "集群": task.cluster,
                "调度": schedule_type,
                "状态": f"{status_icon} {status_text}",
                "启用": enabled_text,
                "上次运行": task.last_run.split("T")[0] if task.last_run else "从未运行",
                "下次运行": next_run_text,
                "ID": task.task_id
            })
        
        # 显示数据表
        if task_data:
            task_df = pd.DataFrame(task_data)
            st.dataframe(
                task_df,
                use_container_width=True,
                hide_index=True
            )
            
            # 任务操作区域
            st.subheader("任务操作")
            
            # 选择要操作的任务
            selected_task_id = st.selectbox(
                "选择任务", 
                [t["ID"] for t in task_data],
                format_func=lambda x: next((t["任务名称"] for t in task_data if t["ID"] == x), x)
            )
            
            if selected_task_id:
                selected_task = next((t for t in tasks if t.task_id == selected_task_id), None)
                
                if selected_task:
                    # 显示任务详情
                    with st.expander("任务详情", expanded=True):
                        col1, col2 = st.columns(2)
                        
                        with col1:
                            st.markdown(f"**任务名称:** {selected_task.name}")
                            st.markdown(f"**集群:** {selected_task.cluster}")
                            st.markdown(f"**描述:** {selected_task.description}")
                        
                        with col2:
                            st.markdown(f"**调度类型:** {selected_task.get_pretty_schedule()}")
                            st.markdown(f"**状态:** {selected_task.last_status}")
                            st.markdown(f"**创建时间:** {selected_task.created_at}")
                        
                        # 显示规则配置
                        st.markdown("#### 规则配置")
                        rules = selected_task.rules
                        
                        if rules:
                            for rule_type, rule_config in rules.items():
                                if rule_config.get('enabled', False):
                                    st.write(f"**{rule_type.capitalize()}** 规则: {len(rule_config.get('rules', []))} 条")
                    
                    # 任务操作按钮组
                    col1, col2, col3 = st.columns(3)
                    
                    with col1:
                        if st.button("立即执行", key=f"run_{selected_task_id}"):
                            # 使用UI组件执行任务并显示进度
                            from components.ui import execute_inspection_task
                            
                            # 执行任务
                            success, message, results = run_inspection(selected_task_id, return_results=True)
                            
                            if success:
                                st.success(f"任务 {selected_task.name} 已成功执行")
                                
                                # 显示巡检结果
                                if results:
                                    st.subheader("巡检结果概览")
                                    
                                    # 显示摘要信息
                                    display_summary_metrics(results)
                                    
                                    # 创建详细结果的标签页
                                    if len(results) > 1:
                                        result_tabs = st.tabs([f"{k.capitalize()}巡检结果" for k in results.keys()])
                                        
                                        # 填充每个标签页的内容
                                        for i, (key, result) in enumerate(results.items()):
                                            with result_tabs[i]:
                                                display_inspection_results(key, result)
                                    else:
                                        # 如果只有一个巡检结果，直接显示
                                        key, result = next(iter(results.items()))
                                        display_inspection_results(key, result)
                                st.rerun()
                            else:
                                st.error(f"执行任务失败: {message}")
                    
                    with col2:
                        # 启用/禁用按钮
                        if selected_task.enabled:
                            if st.button("禁用任务", key=f"disable_{selected_task_id}"):
                                selected_task.enabled = False
                                if add_schedule(selected_task, update=True):
                                    st.success(f"任务 {selected_task.name} 已禁用")
                                    restart_scheduler()
                                    st.rerun()
                                else:
                                    st.error("禁用任务失败")
                        else:
                            if st.button("启用任务", key=f"enable_{selected_task_id}"):
                                selected_task.enabled = True
                                if add_schedule(selected_task, update=True):
                                    st.success(f"任务 {selected_task.name} 已启用")
                                    restart_scheduler()
                                    st.rerun()
                                else:
                                    st.error("启用任务失败")
                    
                    with col3:
                        # 删除按钮
                        if st.button("删除任务", key=f"delete_{selected_task_id}", type="primary"):
                            if delete_schedule(selected_task_id):
                                st.success(f"任务 {selected_task.name} 已删除")
                                restart_scheduler()
                                st.rerun()
                            else:
                                st.error("删除任务失败")

def render_create_task_tab():
    """渲染创建新任务标签页"""
    import time
    import yaml
    
    st.subheader("创建新的定时巡检任务")
    
    # 获取集群列表
    clusters = list_clusters()
    
    if not clusters:
        st.warning("还没有配置任何集群。请前往「集群信息」页面添加集群。")
        if st.button("转到集群信息页面"):
            st.switch_page("pages/1_cluster_info.py")
        return
    
    # --- 调度类型选择放到表单外部 ---
    schedule_types = ["单次定时", "周期定时（Cron表达式）"]
    if "schedule_type" not in st.session_state:
        st.session_state["schedule_type"] = schedule_types[0]
    st.session_state["schedule_type"] = st.selectbox(
        "调度类型",
        schedule_types,
        index=schedule_types.index(st.session_state["schedule_type"]),
        key="schedule_type_selectbox"
    )
    task_type = st.session_state["schedule_type"]
    # --- 表单内部 ---
    with st.form(key="new_task_form"):
        # 基本信息
        task_name = st.text_input("任务名称", placeholder="例如：每日巡检任务")
        task_description = st.text_area("任务描述", placeholder="描述任务的目的和范围")
        selected_cluster = st.selectbox("选择集群", clusters)
        
        # 获取选定集群的配置信息并显示
        if selected_cluster:
            # 使用UI组件显示集群信息
            cluster_config, nodes, prometheus_config = display_cluster_info(selected_cluster)
        
        # 调度设置
        cron_expr = ""
        run_date = None
        run_time = None
        
        if task_type == "周期定时（Cron表达式）":
            # 创建说明和输入框
            st.caption("Cron表达式格式：分 时 日 月 周")
            
            # 创建Cron字段配置
            cron_fields = [
                {"name": "分", "key": "cron_min", "default": "0", "help": "0-59，*为每分钟"},
                {"name": "时", "key": "cron_hour", "default": "8", "help": "0-23，*为每小时"},
                {"name": "日", "key": "cron_dom", "default": "*", "help": "1-31，*为每天"},
                {"name": "月", "key": "cron_month", "default": "*", "help": "1-12，*为每月"},
                {"name": "周", "key": "cron_dow", "default": "*", "help": "0=周日，1=周一...6=周六，*为每周"}
            ]
            
            # 创建输入栏
            cols = st.columns(len(cron_fields))
            cron_values = {}
            
            for i, field in enumerate(cron_fields):
                with cols[i]:
                    value = st.text_input(
                        field["name"],
                        value=st.session_state.get(field["key"], field["default"]),
                        key=field["key"],
                        help=field["help"]
                    )
                    cron_values[field["key"]] = value
            
            # 组装Cron表达式
            cron_expr = f"{cron_values['cron_min']} {cron_values['cron_hour']} {cron_values['cron_dom']} {cron_values['cron_month']} {cron_values['cron_dow']}"
            
            # 显示常见示例
            with st.expander("常见Cron表达式示例"):
                st.markdown("""
                - `0 8 * * *` - 每天早上8点
                - `0 0 * * 0` - 每周日午夜
                - `0 18 * * 1-5` - 每个工作日下午6点
                - `0 0 1 * *` - 每月1日午夜
                - `*/15 * * * *` - 每15分钟
                """)
                
        elif task_type == "单次定时":
            col1, col2 = st.columns(2)
            with col1:
                run_date = st.date_input("执行日期", key="once_date")
            with col2:
                run_time = st.time_input("执行时间", key="once_time")
        # 集群配置
        cluster_config = get_cluster(selected_cluster)
        prometheus_config = cluster_config.get_prometheus_config() if cluster_config else None
        kubeconfig = cluster_config.get_kubeconfig() if cluster_config else None
        
        check_boxes_col1, check_boxes_col2, check_boxes_col3 = st.columns(3)
        with check_boxes_col1:
            node_check = st.checkbox("启用节点状态巡检", value=True)
        with check_boxes_col2:
            prometheus_check = st.checkbox("启用 Prometheus 指标巡检", value=prometheus_config.get('enabled', False) if prometheus_config else False, disabled=not prometheus_config or not prometheus_config.get('enabled', False))
        with check_boxes_col3:
            opa_check = st.checkbox("启用 OPA 合规性巡检", value=bool(kubeconfig), disabled=not kubeconfig)
        selected_node_rules = []
        selected_prometheus_rules = []
        selected_opa_rules = []
        if node_check or prometheus_check or opa_check:
            from utils.rule_manager import RuleManager
            
            # 使用RuleManager统一处理规则选择
            selected_node_rules, selected_prometheus_rules, selected_opa_rules = RuleManager.create_rule_selection_tabs(
                node_check, prometheus_check, opa_check, "_schedule"
            )
        st.divider()
        st.divider()
        # 任务启用状态
        task_enabled = st.checkbox("立即启用任务", value=True)
        
        # 提交按钮
        submit_button = st.form_submit_button("创建任务")
        if submit_button:
            # 验证输入
            if not task_name:
                st.error("请输入任务名称")
            elif task_type == "周期定时（Cron表达式）" and not cron_expr:
                st.error("请输入有效的Cron表达式")
            elif task_type == "单次定时" and (not run_date or not run_time):
                st.error("请选择执行日期和时间")
            elif not (node_check or prometheus_check or opa_check):
                st.error("请至少选择一种巡检类型")
            else:
                # 创建任务对象
                task_id = f"task_{int(time.time())}"
                # 确定任务类型和CRON表达式
                task_type_map = {
                    "单次定时": "once",
                    "周期定时（Cron表达式）": "cron"
                }
                actual_task_type = task_type_map.get(task_type, "cron")
                # 生成cron表达式和run_datetime
                run_datetime = None
                if task_type == "单次定时":
                    run_datetime = f"{run_date} {run_time.strftime('%H:%M')}"
                    actual_cron_expr = ""
                elif task_type == "周期定时（Cron表达式）":
                    actual_cron_expr = cron_expr
                else:
                    actual_cron_expr = ""
                rules_config = {}
                if node_check:
                    rules_config["node"] = {
                        "enabled": True,
                        "rules": selected_node_rules
                    }
                if prometheus_check:
                    rules_config["prometheus"] = {
                        "enabled": True,
                        "rules": selected_prometheus_rules
                    }
                if opa_check:
                    rules_config["opa"] = {
                        "enabled": True,
                        "rules": selected_opa_rules
                    }
                new_task = ScheduleTask(
                    task_id=task_id,
                    cluster=selected_cluster,
                    name=task_name,
                    description=task_description,
                    cron_expr=actual_cron_expr,
                    enabled=task_enabled,
                    rules=rules_config,
                    task_type=actual_task_type,
                    run_datetime=run_datetime
                )
                if add_schedule(new_task):
                    st.success("定时巡检任务创建成功")
                    restart_scheduler()
                    st.rerun()
                else:
                    st.error("创建任务失败")
        
        # 在表单提交后展示cron表达式和含义
        if submit_button and task_type == "周期定时（Cron表达式）":
            st.info(f"当前Cron表达式： `{cron_expr}`")
            cron_desc = get_cron_description(cron_values['cron_min'], cron_values['cron_hour'], cron_values['cron_dom'], cron_values['cron_month'], cron_values['cron_dow'])
            st.caption(f"含义预览：{cron_desc}")

def get_cron_description(minute, hour, dom, month, dow):
    """
    将cron表达式转为人类可读的中文描述。
    支持基本的cron表达式格式。
    
    Args:
        minute: 分钟字段 (0-59)
        hour: 小时字段 (0-23)
        dom: 日期字段 (1-31)
        month: 月份字段 (1-12)
        dow: 星期字段 (0-6，0=周日)
        
    Returns:
        str: 人类可读的cron表达式描述
    """
    # 定义时间部分的映射关系
    mappings = {
        'minute': {'field': minute, 'wild': '每分钟', 'format': '{}分'},
        'hour': {'field': hour, 'wild': '每小时', 'format': '{}时'},
        'dom': {'field': dom, 'wild': '每天', 'format': '{}日'},
        'month': {'field': month, 'wild': '每月', 'format': '{}月'},
        'dow': {'field': dow, 'wild': '每周', 'format': '周{}', 
                'names': {'0': '日', '1': '一', '2': '二', '3': '三', '4': '四', '5': '五', '6': '六'}}
    }
    
    parts = []
    for part_name, config in mappings.items():
        value = config['field']
        if value == '*':
            parts.append(config['wild'])
        elif part_name == 'dow' and value in config['names']:
            parts.append(config['format'].format(config['names'][value]))
        else:
            parts.append(config['format'].format(value))
    
    return "，".join(parts)
