#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
调度管理器 - 用于管理定时巡检任务
"""

import json
import os
import time
from datetime import datetime as dt
from pathlib import Path
import threading
import schedule
from croniter import croniter
import logging

# 导入巡检模块
from utils.cluster_config import get_cluster
from utils.inspection_result import InspectionResult
from inspectors.node.node_inspector import NodeInspector
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from inspectors.opa.opa_inspector import OpaInspector

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger("ScheduleManager")

# 定义调度任务存储目录
DATA_DIR = Path(__file__).resolve().parent.parent / "data"
SCHEDULE_DIR = DATA_DIR / "schedules"
SCHEDULE_FILE = SCHEDULE_DIR / "schedules.json"

# 确保目录存在
SCHEDULE_DIR.mkdir(parents=True, exist_ok=True)

# 全局调度线程
_scheduler_thread = None
_stop_event = threading.Event()

class ScheduleTask:
    """调度任务类"""
    
    def __init__(self, task_id=None, cluster=None, name=None, description=None, 
                 cron_expr=None, enabled=True, rules=None, task_type="cron",
                 last_run=None, last_status=None, created_at=None, run_datetime=None):
        self.task_id = task_id or f"task_{int(time.time())}"
        self.cluster = cluster
        self.name = name or f"巡检任务 {self.task_id}"
        self.description = description or ""
        self.cron_expr = cron_expr
        self.enabled = enabled
        self.rules = rules or {}
        self.task_type = task_type  # cron, monthly, weekly, daily, hourly, once
        self.last_run = last_run
        self.last_status = last_status
        self.created_at = created_at or dt.now().isoformat()
        self.run_datetime = run_datetime  # 新增字段，单次定时任务用
        
    def to_dict(self):
        """将任务转换为字典"""
        return {
            "task_id": self.task_id,
            "cluster": self.cluster,
            "name": self.name,
            "description": self.description,
            "cron_expr": self.cron_expr,
            "enabled": self.enabled,
            "rules": self.rules,
            "task_type": self.task_type,
            "last_run": self.last_run,
            "last_status": self.last_status,
            "created_at": self.created_at,
            "run_datetime": self.run_datetime  # 新增字段
        }
    
    @classmethod
    def from_dict(cls, data):
        """从字典创建任务"""
        return cls(
            task_id=data.get("task_id"),
            cluster=data.get("cluster"),
            name=data.get("name"),
            description=data.get("description"),
            cron_expr=data.get("cron_expr"),
            enabled=data.get("enabled", True),
            rules=data.get("rules", {}),
            task_type=data.get("task_type", "cron"),
            last_run=data.get("last_run"),
            last_status=data.get("last_status"),
            created_at=data.get("created_at"),
            run_datetime=data.get("run_datetime")  # 新增字段
        )
    
    def is_valid_cron(self):
        """验证cron表达式是否有效"""
        try:
            if self.cron_expr:
                croniter(self.cron_expr)
                return True
        except Exception:
            return False
        return False
    
    def get_next_run(self):
        """获取下次运行时间"""
        for job in schedule.jobs:
            if self.task_id in job.tags:
                return job.next_run
        return None
    
    def get_pretty_schedule(self):
        """获取友好的调度描述"""
        if self.task_type == "cron":
            return f"自定义: {self.cron_expr}"
        elif self.task_type == "once":
            return f"单次定时: {self.run_datetime}"
        elif self.task_type == "hourly":
            return "每小时"
        elif self.task_type == "daily":
            return "每天"
        elif self.task_type == "weekly":
            return "每周"
        elif self.task_type == "monthly":
            return "每月"
        return "未知调度类型"

def load_schedules():
    """加载所有调度任务"""
    if not SCHEDULE_FILE.exists():
        save_schedules([])
        return []
    
    try:
        with open(SCHEDULE_FILE, 'r') as f:
            data = json.load(f)
            return [ScheduleTask.from_dict(task) for task in data]
    except Exception as e:
        logger.error(f"加载调度任务失败: {e}")
        return []

def save_schedules(tasks):
    """保存所有调度任务"""
    try:
        with open(SCHEDULE_FILE, 'w') as f:
            json.dump([task.to_dict() for task in tasks], f, indent=2, ensure_ascii=False)
        return True
    except Exception as e:
        logger.error(f"保存调度任务失败: {e}")
        return False

def add_schedule(task):
    """添加调度任务"""
    tasks = load_schedules()
    # 检查是否有相同ID的任务
    for i, t in enumerate(tasks):
        if t.task_id == task.task_id:
            tasks[i] = task
            return save_schedules(tasks)
    
    tasks.append(task)
    return save_schedules(tasks)

def delete_schedule(task_id):
    """删除调度任务"""
    tasks = load_schedules()
    tasks = [t for t in tasks if t.task_id != task_id]
    return save_schedules(tasks)

def get_schedule(task_id):
    """获取特定调度任务"""
    tasks = load_schedules()
    for task in tasks:
        if task.task_id == task_id:
            return task
    return None

def update_task_status(task_id, last_run=None, last_status=None):
    """更新任务状态"""
    task = get_schedule(task_id)
    if task:
        task.last_run = last_run or dt.now().isoformat()
        task.last_status = last_status
        return add_schedule(task)
    return False

def run_inspection_bg(task):
    """在后台执行巡检任务"""
    try:
        logger.info(f"开始执行巡检任务: {task.name} ({task.task_id})")
        
        # 更新任务状态为运行中
        update_task_status(task.task_id, last_status="running")
        
        # 导入执行组件
        from components.ui.task_execution import execute_inspection_task
        
        # 执行任务但不显示进度
        success, message, _ = execute_inspection_task(task, show_progress=False)
        
        # 更新任务状态
        if success:
            update_task_status(task.task_id, last_status="success")
            logger.info(f"任务执行完成: {task.name} ({task.task_id})")
        else:
            update_task_status(task.task_id, last_status="failed")
            logger.error(f"任务执行失败: {task.name} ({task.task_id}) - {message}")
        
        return success
    except Exception as e:
        error_msg = f"执行巡检任务失败: {str(e)}"
        logger.error(error_msg, exc_info=True)  # 添加完整的异常堆栈
        update_task_status(task.task_id, last_status="failed")
        return False

def run_inspection(task_id, return_results=False):
    """执行巡检任务
    
    Args:
        task_id: 任务ID
        return_results: 是否返回巡检结果而不是在后台运行
        
    Returns:
        如果return_results为False: (成功标志, 消息)
        如果return_results为True: (成功标志, 消消息, 结果字典)
    """
    task = get_schedule(task_id)
    if not task:
        logger.error(f"找不到任务: {task_id}")
        return (False, "找不到指定的任务", None) if return_results else (False, "找不到指定的任务")
    
    try:
        if return_results:
            # 导入执行组件
            from components.ui.task_execution import execute_inspection_task
            # 在当前线程中执行并返回结果
            success, message, results = execute_inspection_task(task, show_progress=True)
            return success, message, results
        else:
            # 在后台线程中执行
            threading.Thread(target=lambda: run_inspection_bg(task)).start()
            return True, "任务已启动"
    except Exception as e:
        logger.error(f"启动任务失败: {e}")
        return (False, str(e), None) if return_results else (False, str(e))

def _scheduler_loop():
    """调度器循环，在后台线程中运行"""
    while not _stop_event.is_set():
        schedule.run_pending()
        
        # 检查单次定时任务 - 这里只检查通过_scheduler_loop管理的单次任务
        # 实际上，单次任务应该通过schedule库来管理，而不是在这里手动检查
        # 但为了兼容性，保留这个检查，但添加更严格的条件
        tasks = load_schedules()
        now = dt.now()
        for task in tasks:
            if (task.enabled and task.task_type == "once" and task.run_datetime and 
                not task.last_run):  # 只有从未执行过的任务才检查
                try:
                    run_dt = dt.strptime(task.run_datetime, "%Y-%m-%d %H:%M")
                except Exception:
                    continue
                
                # 检查是否到了执行时间，且任务从未执行过
                if now >= run_dt:
                    logger.info(f"单次定时任务到点执行: {task.name} ({task.task_id})")
                    success = run_inspection_bg(task)
                    
                    # 执行后禁用该任务，避免重复执行
                    task.enabled = False
                    add_schedule(task)
                    
                    # 从schedule中也移除该任务
                    schedule.clear(task.task_id)
        
        time.sleep(30)  # 每30秒检查一次

def schedule_tasks():
    """设置所有启用的调度任务"""
    
    # 清除所有现有的任务
    schedule.clear()
    
    # 加载所有任务
    tasks = load_schedules()
    
    # 注册任务
    for task in tasks:
        if not task.enabled:
            continue
            
        if task.task_type == "cron" and task.is_valid_cron():
            # 对于cron任务，我们创建一个包装函数来处理重复调度
            def create_cron_job(task_obj):
                def cron_job():
                    # 执行任务
                    run_inspection_bg(task_obj)
                    # 任务执行后，重新计算下一次运行时间并重新调度
                    reschedule_cron_task(task_obj)
                return cron_job
            
            # 计算首次运行时间
            from croniter import croniter
            base = dt.now()
            cron = croniter(task.cron_expr, base)
            next_run = cron.get_next(dt)
            delta_seconds = (next_run - base).total_seconds()
            
            # 如果下次运行时间很近（小于1分钟），则延迟到下下次
            if delta_seconds < 60:
                next_run = cron.get_next(dt)
                delta_seconds = (next_run - base).total_seconds()
            
            # 调度首次执行
            schedule.every(int(delta_seconds)).seconds.do(
                create_cron_job(task)
            ).tag(task.task_id)
            
            logger.info(f"已调度Cron任务: {task.name} ({task.task_id})，将在 {next_run.strftime('%Y-%m-%d %H:%M:%S')} 首次执行")
            
        elif task.task_type == "hourly":
            schedule.every().hour.do(
                lambda t=task: run_inspection_bg(t)
            ).tag(task.task_id)
            
        elif task.task_type == "daily":
            schedule.every().day.at("00:00").do(
                lambda t=task: run_inspection_bg(t)
            ).tag(task.task_id)
            
        elif task.task_type == "weekly":
            schedule.every().monday.at("00:00").do(
                lambda t=task: run_inspection_bg(t)
            ).tag(task.task_id)
            
        elif task.task_type == "monthly":
            schedule.every().day.at("00:00").do(
                lambda t=task: run_inspection_bg(t) if dt.now().day == 1 else None
            ).tag(task.task_id)
        elif task.task_type == "once" and task.run_datetime:
            # 计算当前时间到指定运行时间的秒数
            run_time = dt.fromisoformat(task.run_datetime)
            now = dt.now()
            if run_time > now:
                delta_seconds = (run_time - now).total_seconds()
                schedule.every(int(delta_seconds)).seconds.do(
                    lambda t=task: run_inspection_bg(t)
                ).tag(task.task_id)
                logger.info(f"已调度一次性任务: {task.name} ({task.task_id})，将在 {run_time} 执行")
    
    logger.info(f"已调度 {len([t for t in tasks if t.enabled])} 个巡检任务")

def reschedule_cron_task(task):
    """重新调度cron任务的下一次执行"""
    try:
        # 先取消当前任务
        schedule.clear(task.task_id)
        
        # 计算下一次执行时间
        from croniter import croniter
        base = dt.now()
        cron = croniter(task.cron_expr, base)
        next_run = cron.get_next(dt)
        delta_seconds = (next_run - base).total_seconds()
        
        # 创建新的调度
        def create_cron_job(task_obj):
            def cron_job():
                run_inspection_bg(task_obj)
                reschedule_cron_task(task_obj)
            return cron_job
        
        schedule.every(int(delta_seconds)).seconds.do(
            create_cron_job(task)
        ).tag(task.task_id)
        
        logger.info(f"重新调度Cron任务: {task.name} ({task.task_id})，下次执行时间: {next_run.strftime('%Y-%m-%d %H:%M:%S')}")
        
    except Exception as e:
        logger.error(f"重新调度任务失败: {task.name} ({task.task_id}), 错误: {e}")

def start_scheduler():
    """启动调度器"""
    global _scheduler_thread, _stop_event
    
    if _scheduler_thread and _scheduler_thread.is_alive():
        logger.info("调度器已在运行")
        return
    
    # 设置调度任务
    schedule_tasks()
    
    # 重置停止事件
    _stop_event.clear()
    
    # 创建并启动线程
    _scheduler_thread = threading.Thread(target=_scheduler_loop)
    _scheduler_thread.daemon = True
    _scheduler_thread.start()
    
    logger.info("调度器已启动")
    return True

def stop_scheduler():
    """停止调度器"""
    global _scheduler_thread, _stop_event
    
    if not _scheduler_thread or not _scheduler_thread.is_alive():
        logger.info("调度器未运行")
        return
    
    _stop_event.set()
    _scheduler_thread.join(timeout=5)
    _scheduler_thread = None
    
    logger.info("调度器已停止")
    return True

def restart_scheduler():
    """重启调度器"""
    stop_scheduler()
    return start_scheduler()

# 在模块加载时自动启动调度器
try:
    start_scheduler()
except Exception as e:
    logger.error(f"启动调度器失败: {e}")