#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
页面公共组件
"""
import streamlit as st
from utils.rule_loader import load_rules
from utils.rule_manager import RuleManager

def create_rule_selection_tabs(node_check, prometheus_check, opa_check, key_suffix=""):
    """
    创建规则选择的标签页界面
    
    参数:
    node_check (bool): 是否启用节点检查
    prometheus_check (bool): 是否启用Prometheus检查
    opa_check (bool): 是否启用OPA检查
    key_suffix (str): 用于区分不同调用场景的组件key后缀
    
    返回:
    tuple: (selected_node_rules, selected_prometheus_rules, selected_opa_rules)
    """
    return RuleManager.create_rule_selection_tabs(node_check, prometheus_check, opa_check, key_suffix)


def create_rule_selection(rule_type, key_suffix=""):
    """
    通用规则选择界面生成函数
    
    参数:
    rule_type (str): 规则类型 'node', 'prometheus' 或 'opa'
    key_suffix (str): 用于区分不同调用场景的组件key后缀
    
    返回:
    list: 选中的规则ID列表
    """
    return RuleManager.create_rule_selection(rule_type, key_suffix)


def create_node_rules_selection(key_suffix=""):
    """创建节点规则选择界面"""
    return create_rule_selection('node', key_suffix)


def create_prometheus_rules_selection(key_suffix=""):
    """创建Prometheus规则选择界面"""
    return create_rule_selection('prometheus', key_suffix)


def create_opa_rules_selection(key_suffix=""):
    """创建OPA规则选择界面"""
    return create_rule_selection('opa', key_suffix)


def create_rule_selection_in_form(rule_type, all_rules, rule_display_names, select_all_key):
    """
    创建表单内的规则选择组件
    
    参数:
    rule_type (str): 规则类型显示名称
    all_rules (list): 所有可用规则ID列表
    rule_display_names (dict): 规则ID到显示名称的映射
    select_all_key (str): 全选复选框的键名
    
    返回:
    list: 选中的规则ID列表
    """
    return RuleManager.create_rule_selection_in_form(rule_type, all_rules, rule_display_names, select_all_key)
