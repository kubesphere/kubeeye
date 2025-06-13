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
        self.by_rules = False  # 是否按规则数量显示进度
        
    def initialize(self, total_steps=3, by_rules=False):
        """初始化进度组件"""
        self.total_steps = total_steps
        self.current_step = 0
        self.by_rules = by_rules
        self.progress_bar.progress(0)
        
        if by_rules:
            self.status_text.text(f"准备开始巡检...（共{total_steps}个规则）")
        else:
            self.status_text.text("准备开始巡检...")
        
    def update(self, step_message, step_complete=False, rule_name=None):
        """更新进度"""
        self.progress_text.text(step_message)
        
        if self.by_rules and rule_name:
            # 按规则显示进度
            status_msg = f"正在执行规则: {rule_name} ({self.current_step + 1}/{self.total_steps})"
            self.status_text.text(status_msg)
        else:
            self.status_text.text(step_message)
        
        if step_complete:
            self.current_step += 1
            progress = min(1.0, self.current_step / self.total_steps)
            self.progress_bar.progress(progress)
            
            if self.by_rules:
                # 更新规则完成状态
                status_msg = f"已完成 {self.current_step}/{self.total_steps} 个规则"
                self.status_text.text(status_msg)
            
    def complete(self, delay=0.5):
        """完成进度显示"""
        self.progress_bar.progress(1.0)
        if self.by_rules:
            self.status_text.text(f"巡检完成！共执行了 {self.total_steps} 个规则")
        else:
            self.status_text.text("巡检完成！")
        if delay > 0:
            time.sleep(delay)  # 给用户一个视觉反馈
            
    def error(self, message):
        """显示错误信息"""
        self.status_text.text(f"错误: {message}")
        
    def warning(self, message):
        """显示警告信息"""
        self.status_text.text(f"警告: {message}")
