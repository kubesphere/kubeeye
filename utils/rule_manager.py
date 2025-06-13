#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则管理工具模块 - 为各个组件提供统一的规则处理接口
更新版本，支持断言系统和表格形式显示规则
"""
from typing import Dict, List, Any, Optional, Tuple
import streamlit as st
import pandas as pd
import yaml
from utils.rule_loader import load_rules, Rule

class RuleManager:
    """规则管理类，提供统一的规则操作接口"""
    
    @staticmethod
    def get_rule_type_display_names() -> Dict[str, str]:
        """获取规则类型的显示名称映射"""
        return {
            'node': '节点状态巡检规则',
            'prometheus': 'Prometheus 指标规则',
            'opa': 'OPA 合规性巡检规则'
        }

    @staticmethod
    def get_enabled_rules(rule_type: str) -> List[Rule]:
        """获取指定类型的已启用规则"""
        return [rule for rule in load_rules(rule_type) if rule.enabled]

    @staticmethod
    def get_rule_display_names(rules: List[Rule]) -> Dict[str, str]:
        """获取规则ID到显示名称的映射"""
        return {rule.id: rule.name for rule in rules}
    
    @staticmethod
    def get_rule_options(rules: List[Rule]) -> List[str]:
        """获取规则ID列表"""
        return [rule.id for rule in rules]
    
    @classmethod
    def get_rule_selection_data(cls, rule_type: str) -> Tuple[List[Rule], List[str], Dict[str, str]]:
        """获取规则选择所需的数据"""
        rules = cls.get_enabled_rules(rule_type)
        options = cls.get_rule_options(rules)
        display_names = cls.get_rule_display_names(rules)
        return rules, options, display_names

    @classmethod
    def rule_to_dataframe(cls, rules: List[Rule]) -> pd.DataFrame:
        """将规则转换为DataFrame，用于表格显示"""
        if not rules:
            return pd.DataFrame()
            
        data = []
        for rule in rules:
            data.append({
                "ID": rule.id,
                "名称": rule.name,
                "类别": rule.category,
                "严重性": rule.severity,
                "描述": rule.description[:50] + "..." if len(rule.description) > 50 else rule.description,
                "断言数量": len(rule.assertions) if hasattr(rule, 'assertions') else 0
            })
        return pd.DataFrame(data)
    
    @classmethod
    def create_rule_selection(cls, rule_type: str, key_suffix: str = "") -> List[str]:
        """
        简化版：只用data_editor表格选择规则，不做任何表格外联动
        通过使用data_editor的编辑结果更新session_state
        """
        rules, options, display_names = cls.get_rule_selection_data(rule_type)
        if not rules:
            st.info(f"没有找到已启用的 {cls.get_rule_type_display_names()[rule_type]}，请在规则管理中添加启用规则。")
            return []
        
        form_key = f"rule_selection_{rule_type}{key_suffix}"
        table_key = f"rule_table_{rule_type}{key_suffix}"
        
        # 初始化session state (如果还没有)
        if form_key not in st.session_state:
            st.session_state[form_key] = options.copy()  # 默认全选
            
        # 构造表格数据 (基于当前session_state)
        data = []
        for rule in rules:
            selected = rule.id in st.session_state[form_key]
            data.append({
                "选择": selected,
                "ID": rule.id,
                "名称": rule.name,
                "类别": rule.category,
                "严重性": rule.severity,
                "描述": rule.description[:50] + "..." if len(rule.description) > 50 else rule.description,
                "断言数量": len(rule.assertions) if hasattr(rule, 'assertions') else 0
            })
        rules_df = pd.DataFrame(data)
        
        # 使用data_editor（但不直接使用其返回值更新session_state）
        edited_df = st.data_editor(
            rules_df,
            use_container_width=True,
            hide_index=True,
            column_config={
                "选择": st.column_config.CheckboxColumn("选择", help="选择要运行的规则", width="small"),
                "ID": st.column_config.TextColumn("规则ID", width="medium"),
                "名称": st.column_config.TextColumn("规则名称", width="medium"),
                "类别": st.column_config.TextColumn("类别", width="small"),
                "严重性": st.column_config.TextColumn("严重性", width="small"),
                "描述": st.column_config.TextColumn("描述", width="large"),
                "断言数量": st.column_config.NumberColumn("断言数量", width="small"),
            },
            disabled=["ID", "名称", "类别", "严重性", "描述", "断言数量"],
            key=table_key,
            on_change=None  # 不使用on_change，避免重新加载
        )
        
        # 从编辑后的表格中获取选择的规则IDs
        selected_rules = [row["ID"] for _, row in edited_df.iterrows() if row["选择"]]
        
        # 更新session_state (但不导致页面重新加载)
        st.session_state[form_key] = selected_rules
        
        st.caption(f"已选择: {len(selected_rules)}/{len(options)} 规则")
        return selected_rules

    @classmethod
    def create_rule_selection_in_form(cls, rule_type: str, key_suffix: str = "") -> List[str]:
        """
        表单内简化版：只用data_editor表格选择规则
        针对表单内使用做了特别优化，避免过多刷新
        """
        rules, options, display_names = cls.get_rule_selection_data(rule_type)
        if not rules:
            st.info(f"没有找到已启用的 {cls.get_rule_type_display_names()[rule_type]}，请在规则管理中添加启用规则。")
            return []
        
        # 表单专用的状态键
        form_key = f"rule_selection_form_{rule_type}{key_suffix}"
        table_key = f"rule_table_form_{rule_type}{key_suffix}"
        
        # 初始化session state
        if form_key not in st.session_state:
            st.session_state[form_key] = options.copy()  # 默认全选
        
        # 构造表格数据
        data = []
        for rule in rules:
            selected = rule.id in st.session_state[form_key]
            data.append({
                "选择": selected,
                "ID": rule.id,
                "名称": rule.name,
                "类别": rule.category,
                "严重性": rule.severity,
                "描述": rule.description[:50] + "..." if len(rule.description) > 50 else rule.description,
                "断言数量": len(rule.assertions) if hasattr(rule, 'assertions') else 0
            })
        rules_df = pd.DataFrame(data)
        
        # 在表单中使用data_editor
        edited_df = st.data_editor(
            rules_df,
            use_container_width=True,
            hide_index=True,
            column_config={
                "选择": st.column_config.CheckboxColumn("选择", help="选择要运行的规则", width="small"),
                "ID": st.column_config.TextColumn("规则ID", width="medium"),
                "名称": st.column_config.TextColumn("规则名称", width="medium"),
                "类别": st.column_config.TextColumn("类别", width="small"),
                "严重性": st.column_config.TextColumn("严重性", width="small"),
                "描述": st.column_config.TextColumn("描述", width="large"),
                "断言数量": st.column_config.NumberColumn("断言数量", width="small"),
            },
            disabled=["ID", "名称", "类别", "严重性", "描述", "断言数量"],
            key=table_key,
        )
        
        # 提取选择的规则
        selected_rules = [row["ID"] for _, row in edited_df.iterrows() if row["选择"]]
        
        # 更新session state
        if st.session_state[form_key] != selected_rules:
            st.session_state[form_key] = selected_rules
            
        st.caption(f"已选择: {len(selected_rules)}/{len(options)} 规则")
        return selected_rules
    
    @classmethod
    def create_rule_selection_tabs(
        cls, node_check: bool, prometheus_check: bool, opa_check: bool, key_suffix: str = "", 
        in_form: bool = False
    ) -> Tuple[List[str], List[str], List[str]]:
        """
        创建规则选择的标签页界面，只显示可用规则类型的标签页
        用户在各标签页中自行选择要使用的规则
        
        参数:
        node_check (bool): 节点巡检是否可用
        prometheus_check (bool): Prometheus巡检是否可用
        opa_check (bool): OPA巡检是否可用
        key_suffix (str): 用于区分不同调用场景的组件key后缀
        in_form (bool): 是否在表单内使用
        
        返回:
        Tuple[List[str], List[str], List[str]]: 
            (selected_node_rules, selected_prometheus_rules, selected_opa_rules)
        """
        selected_node_rules = []
        selected_prometheus_rules = []
        selected_opa_rules = []
        
        # 定义规则类型配置
        rule_configs = [
            {
                "available": node_check,
                "type": "node", 
                "tab_label": "节点巡检规则",
                "result_var": "selected_node_rules"
            },
            {
                "available": prometheus_check, 
                "type": "prometheus", 
                "tab_label": "Prometheus巡检规则",
                "result_var": "selected_prometheus_rules"
            },
            {
                "available": opa_check, 
                "type": "opa", 
                "tab_label": "OPA巡检规则",
                "result_var": "selected_opa_rules"
            }
        ]
        
        # 过滤出可用的规则配置
        available_configs = [cfg for cfg in rule_configs if cfg["available"]]
        
        # 如果没有可用规则类型，提示用户
        if not available_configs:
            st.warning("没有可用的巡检类型，请检查集群配置。")
            return selected_node_rules, selected_prometheus_rules, selected_opa_rules
        
        # 预先加载规则数据，避免切换标签页时重新加载
        for cfg in available_configs:
            key_suffix_type = f"{key_suffix}_{cfg['type']}"
            selection_key = f"rule_selection_{cfg['type']}{key_suffix_type}"
            if selection_key not in st.session_state:
                cls.get_rule_selection_data(cfg["type"])  # 预加载规则数据
        
        # 创建标签页标题列表
        tab_labels = [cfg["tab_label"] for cfg in available_configs]
        
        # 选项卡状态键
        tab_key = f"rule_tabs{key_suffix}"
        
        # 初始化选项卡状态，以确保有效索引
        if tab_key not in st.session_state or st.session_state[tab_key] >= len(tab_labels):
            st.session_state[tab_key] = 0
            
        # 创建标签页
        rule_tabs = st.tabs(tab_labels)
        
        # 在每个选项卡中创建规则选择
        for i, cfg in enumerate(available_configs):
            with rule_tabs[i]:
                if in_form:
                    # 在表单中使用表单兼容的规则选择
                    selected_rules = cls.create_rule_selection_in_form(
                        cfg["type"], 
                        key_suffix=f"{key_suffix}_{cfg['type']}"
                    )
                else:
                    # 非表单环境使用标准规则选择
                    selected_rules = cls.create_rule_selection(
                        cfg["type"], 
                        key_suffix=f"{key_suffix}_{cfg['type']}"
                    )
                
                # 根据配置设置结果变量
                if cfg["result_var"] == "selected_node_rules":
                    selected_node_rules = selected_rules
                elif cfg["result_var"] == "selected_prometheus_rules":
                    selected_prometheus_rules = selected_rules
                elif cfg["result_var"] == "selected_opa_rules":
                    selected_opa_rules = selected_rules
        
        return selected_node_rules, selected_prometheus_rules, selected_opa_rules
