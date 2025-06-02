#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则管理工具模块 - 为各个组件提供统一的规则处理接口
"""
from typing import Dict, List, Any, Optional, Tuple
import streamlit as st
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
    def create_rule_selection(cls, rule_type: str, key_suffix: str = "") -> List[str]:
        """
        创建规则选择界面并返回选择的规则ID列表
        
        参数:
        rule_type (str): 规则类型 node, prometheus 或 opa
        key_suffix (str): 用于区分不同调用场景的组件key后缀
        
        返回:
        List[str]: 选中的规则ID列表
        """
        # 获取显示名称和规则数据
        type_display_names = cls.get_rule_type_display_names()
        rules, rule_options, rule_display_names = cls.get_rule_selection_data(rule_type)
        
        # 显示标题
        st.subheader(type_display_names.get(rule_type, f"{rule_type}规则"))
        
        # 如果没有规则可选择
        if not rule_options:
            st.warning(f"没有可用的{rule_type}规则")
            return []
        
        # 设置"全选"复选框
        select_all_key = f"select_all_{rule_type}{key_suffix}"
        if select_all_key not in st.session_state:
            st.session_state[select_all_key] = True
        
        # 创建布局
        col1, col2 = st.columns([3, 1])
        
        with col1:
            st.info(f"选择需要执行的{type_display_names.get(rule_type, rule_type)}规则")
        
        with col2:
            select_all = st.checkbox("全选", key=select_all_key)
        
        # 规则多选框
        selected_rules = st.multiselect(
            "请选择要执行的规则",
            rule_options,
            default=rule_options if select_all else [],
            format_func=lambda x: rule_display_names.get(x, x),
            key=f"{rule_type}_rules{key_suffix}"
        )
        
        # 显示已选规则数量
        st.caption(f"已选择 {len(selected_rules)}/{len(rule_options)} 条规则")
        
        return selected_rules
    
    @classmethod
    def create_rule_selection_in_form(
        cls, rule_type: str, rule_options: List[str], rule_display_names: Dict[str, str], 
        select_all_key: str
    ) -> List[str]:
        """
        创建表单内的规则选择组件
        
        参数:
        rule_type (str): 规则类型显示名称
        rule_options (List[str]): 所有可用规则ID列表
        rule_display_names (Dict[str, str]): 规则ID到显示名称的映射
        select_all_key (str): 全选复选框的键名
        
        返回:
        List[str]: 选中的规则ID列表
        """
        if not rule_options:
            st.warning(f"没有可用的{rule_type}规则")
            return []
        
        # 设置全选复选框
        if select_all_key not in st.session_state:
            st.session_state[select_all_key] = True
        
        # 创建布局
        col1, col2 = st.columns([3, 1])
        with col2:
            select_all = st.checkbox("全选", key=select_all_key)
        
        # 创建多选框
        multiselect_key = f"{select_all_key}_multiselect"
        selected_rules = st.multiselect(
            f"请选择要执行的{rule_type}规则",
            rule_options,
            default=rule_options if select_all else [],
            format_func=lambda x: rule_display_names.get(x, x),
            key=multiselect_key
        )
        
        # 显示已选规则数量
        st.caption(f"已选择 {len(selected_rules)}/{len(rule_options)} 条规则")
        
        return selected_rules
    
    @classmethod
    def create_rule_selection_tabs(
        cls, node_check: bool, prometheus_check: bool, opa_check: bool, key_suffix: str = ""
    ) -> Tuple[List[str], List[str], List[str]]:
        """
        创建规则选择的标签页界面
        
        参数:
        node_check (bool): 是否启用节点检查
        prometheus_check (bool): 是否启用Prometheus检查
        opa_check (bool): 是否启用OPA检查
        key_suffix (str): 用于区分不同调用场景的组件key后缀
        
        返回:
        Tuple[List[str], List[str], List[str]]: 
            (selected_node_rules, selected_prometheus_rules, selected_opa_rules)
        """
        selected_node_rules = []
        selected_prometheus_rules = []
        selected_opa_rules = []
        
        # 如果没有启用任何规则类型，则直接返回空结果
        if not (node_check or prometheus_check or opa_check):
            return selected_node_rules, selected_prometheus_rules, selected_opa_rules
        
        # 定义规则类型配置
        rule_configs = [
            {
                "enabled": node_check, 
                "type": "node", 
                "tab_label": "节点巡检规则",
                "result_var": "selected_node_rules"
            },
            {
                "enabled": prometheus_check, 
                "type": "prometheus", 
                "tab_label": "Prometheus巡检规则",
                "result_var": "selected_prometheus_rules"
            },
            {
                "enabled": opa_check, 
                "type": "opa", 
                "tab_label": "OPA巡检规则",
                "result_var": "selected_opa_rules"
            }
        ]
        
        # 过滤出启用的规则配置
        enabled_configs = [cfg for cfg in rule_configs if cfg["enabled"]]
        
        # 创建标签页
        tab_labels = [cfg["tab_label"] for cfg in enabled_configs]
        rule_tabs = st.tabs(tab_labels)
        
        # 遍历所有启用的规则类型
        for i, rule_config in enumerate(enabled_configs):
            with rule_tabs[i]:
                rule_type = rule_config["type"]
                selected_rules = cls.create_rule_selection(rule_type, key_suffix)
                
                # 根据配置将结果分配给正确的变量
                if rule_config["result_var"] == "selected_node_rules":
                    selected_node_rules = selected_rules
                elif rule_config["result_var"] == "selected_prometheus_rules":
                    selected_prometheus_rules = selected_rules
                elif rule_config["result_var"] == "selected_opa_rules":
                    selected_opa_rules = selected_rules
        
        return selected_node_rules, selected_prometheus_rules, selected_opa_rules
