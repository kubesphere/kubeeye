#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
进度显示组件
"""
import streamlit as st
import time

class InspectionProgress:
    """巡检进度显示类"""
    
    def __init__(self):
        self.progress_text = st.empty()
        self.progress_bar = st.progress(0)
        self.status_text = st.empty()
        self.total_steps = 0
        self.current_step = 0
        
    def initialize(self, total_steps=3):
        """初始化进度组件"""
        self.total_steps = total_steps
        self.current_step = 0
        self.progress_bar.progress(0)
        self.status_text.text("准备开始巡检...")
        
    def update(self, step_message, step_complete=False):
        """更新进度"""
        self.progress_text.text(step_message)
        self.status_text.text(step_message)
        
        if step_complete:
            self.current_step += 1
            progress = min(1.0, self.current_step / self.total_steps)
            self.progress_bar.progress(progress)
            
    def complete(self, delay=0.5):
        """完成进度显示"""
        self.progress_bar.progress(1.0)
        self.status_text.text("巡检完成！")
        if delay > 0:
            time.sleep(delay)  # 给用户一个视觉反馈
            
    def error(self, message):
        """显示错误信息"""
        self.status_text.text(f"错误: {message}")
        
    def warning(self, message):
        """显示警告信息"""
        self.status_text.text(f"警告: {message}")
