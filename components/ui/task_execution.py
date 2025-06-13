#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
任务执行组件 - 重构版本，使用统一巡检引擎
"""
from components.ui.inspection_engine import execute_inspection_task

# 直接导出execute_inspection_task函数，保持向后兼容性
__all__ = ['execute_inspection_task']
