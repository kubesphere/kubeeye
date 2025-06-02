#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则管理组件
"""
import streamlit as st
import yaml
from utils.rule_loader import load_rules, save_rule, Rule

def create_rule_form(rule=None, rule_type="node", is_new=False):
    """
    创建规则编辑表单
    
    参数:
        rule: 现有规则对象，如果是新建则为None
        rule_type: 规则类型 (node, prometheus, opa)
        is_new: 是否为新建规则
    
    返回:
        tuple: (submit_success, rule_data) 表单是否提交成功以及规则数据
    """
    form_key = "new_rule_form" if is_new else "rule_editor"
    rule_type_mapping = {
        "节点规则": "node",
        "Prometheus规则": "prometheus",
        "OPA规则": "opa"
    }
    
    actual_rule_type = rule_type_mapping.get(rule_type, rule_type)
    
    # 默认值或从现有规则获取值
    default_values = {
        "id": rule.id if rule else "",
        "name": rule.name if rule else "",
        "description": rule.description if rule else "",
        "category": rule.category if rule else "",
        "severity": rule.severity if rule else "warning",
        "enabled": rule.enabled if rule else True,
        "solution": rule.solution if rule else "",
        "tags": ", ".join(rule.tags) if rule and rule.tags else "",
        "tier": rule.tier if rule else "basic"
    }
    
    # 查询配置 - 从config或直接属性获取
    if rule:
        query = ""
        if hasattr(rule, 'config') and rule.config and 'query' in rule.config and 'promql' in rule.config['query']:
            query = rule.config['query']['promql']
        default_values["query"] = query
    else:
        default_values["query"] = ""
    
    # prometheus规则特有的阈值设置
    thresholds = {}
    if rule and hasattr(rule, 'thresholds'):
        thresholds = rule.thresholds
    
    default_values.update({
        "warning_threshold": thresholds.get("warning", 80) if thresholds else 80,
        "critical_threshold": thresholds.get("critical", 90) if thresholds else 90,
        "threshold_value": str(thresholds.get("value", "")) if thresholds else ""
    })
    
    # OPA规则特有设置 - 使用新格式从config字段获取
    resources = []
    namespaces = []
    rego_content = ""
    
    if rule and hasattr(rule, 'config'):
        # 从config中获取资源
        if rule.config and 'resources' in rule.config:
            resources = rule.config['resources']
        
        # 从config中获取命名空间
        if rule.config and 'scope' in rule.config and 'namespaces' in rule.config['scope']:
            namespaces = rule.config['scope']['namespaces']
        
        # 从config中获取rego内容
        if rule.config and 'rego' in rule.config and 'inline' in rule.config['rego']:
            rego_content = rule.config['rego']['inline']
    
    default_values.update({
        "resources": yaml.dump(resources, default_flow_style=False) if resources else "[]",
        "namespaces": yaml.dump(namespaces, default_flow_style=False) if namespaces else "[]",
        "rego_content": rego_content if rego_content else ""
    })
    
    with st.form(key=form_key):
        # 规则基本信息
        col1, col2 = st.columns(2)
        with col1:
            rule_id = st.text_input("规则ID", value=default_values["id"])
            rule_name = st.text_input("规则名称", value=default_values["name"])
            rule_description = st.text_area("规则描述", value=default_values["description"])
            
        with col2:
            rule_category = st.text_input("规则类别", value=default_values["category"])
            rule_severity = st.selectbox("严重程度", ["info", "warning", "critical"], 
                                        index=["info", "warning", "critical"].index(default_values["severity"]))
            rule_enabled = st.checkbox("启用规则", value=default_values["enabled"])
        
        # 规则查询和阈值
        rule_query = st.text_area("查询语句", value=default_values["query"])
        
        # 根据规则类型显示不同的阈值配置
        threshold_col1, threshold_col2 = st.columns(2)
        
        if actual_rule_type == "prometheus":
            with threshold_col1:
                warning_threshold = st.number_input("警告阈值", value=default_values["warning_threshold"])
            
            with threshold_col2:
                critical_threshold = st.number_input("严重阈值", value=default_values["critical_threshold"])
                
            thresholds = {"warning": warning_threshold, "critical": critical_threshold}
        else:
            # 其他规则类型可以根据需要扩展
            threshold_value = st.text_input("阈值值", value=default_values["threshold_value"])
            thresholds = {"value": threshold_value}
        
        # OPA规则特有字段
        resources = []
        namespaces = []
        rego_content = ""
        
        if actual_rule_type == "opa":
            st.subheader("OPA规则特有配置")
            
            # 资源配置
            resources_yaml = st.text_area(
                "资源配置 (YAML格式)",
                value=default_values["resources"],
                help="要检查的Kubernetes资源类型列表"
            )
            
            # 命名空间配置
            namespaces_yaml = st.text_area(
                "命名空间配置 (YAML格式)",
                value=default_values["namespaces"],
                help="要检查的命名空间列表，如果为空则检查所有命名空间"
            )
            
            # Rego内容
            rego_content = st.text_area(
                "Rego规则内容",
                value=default_values["rego_content"],
                height=300,
                help="OPA Rego规则内容，用于检查资源合规性"
            )
            
            try:
                resources = yaml.safe_load(resources_yaml)
                namespaces = yaml.safe_load(namespaces_yaml)
            except Exception as e:
                st.error(f"YAML解析错误: {e}")
        
        # 解决方案和标签
        rule_solution = st.text_area("解决方案", value=default_values["solution"])
        rule_tags = st.text_input("标签 (用逗号分隔)", value=default_values["tags"])
        
        # 配置数据 (替代原有的自定义数据)
        config_yaml = st.text_area("配置 (YAML格式)", value=yaml.dump(rule.config) if rule and rule.config else "{}")
        
        # 提交按钮文本
        submit_text = "保存新规则" if is_new else "保存规则"
        submit_button = st.form_submit_button(submit_text)
        
        if submit_button:
            try:
                # 验证必填字段
                if not rule_id or not rule_name:
                    st.error("规则ID和规则名称为必填项")
                    return False, None
                
                # 解析标签和配置数据
                tags = [tag.strip() for tag in rule_tags.split(",") if tag.strip()]
                
                # 解析配置YAML
                try:
                    config_data = yaml.safe_load(config_yaml) if config_yaml else {}
                except yaml.YAMLError as e:
                    st.error(f"配置YAML解析错误: {e}")
                    return False, None
                
                # 准备规则数据 - 使用新的规则格式
                rule_data = {
                    "id": rule_id,
                    "name": rule_name,
                    "description": rule_description,
                    "type": actual_rule_type,
                    "category": rule_category,
                    "severity": rule_severity,
                    "enabled": rule_enabled,
                    "solution": rule_solution,
                    "tags": tags,
                    "tier": "basic",  # 默认为basic层级
                    "config": config_data, # 使用解析的配置数据
                    "thresholds": thresholds
                }
                
                # 根据规则类型，添加特定配置
                if actual_rule_type == "prometheus":
                    rule_data["config"]["query"] = {"promql": rule_query}
                
                elif actual_rule_type == "node":
                    if rule_query:
                        rule_data["config"]["execution"] = {"command": rule_query}
                    
                # 添加OPA规则特有字段
                elif actual_rule_type == "opa":
                    if resources:
                        rule_data["config"]["resources"] = resources
                    if namespaces:
                        rule_data["config"]["scope"] = {"namespaces": namespaces}
                    if rego_content:
                        rule_data["config"]["rego"] = {"inline": rego_content}
                
                return True, rule_data
                
            except Exception as e:
                st.error(f"处理规则数据出错: {e}")
                return False, None
        
        return False, None

def render_rule_management_tab():
    """渲染规则管理标签页内容"""
    st.subheader("规则管理")
    
    # 选择规则类型
    rule_type = st.selectbox("选择规则类型", ["节点规则", "Prometheus规则", "OPA规则"])
    
    # 规则类型映射
    rule_type_mapping = {
        "节点规则": "node",
        "Prometheus规则": "prometheus",
        "OPA规则": "opa"
    }
    
    # 加载规则
    rules = load_rules(rule_type_mapping[rule_type])
    
    # 创建规则列表
    if rules:
        # 选择规则进行查看或编辑
        rule_options = [f"{rule.id}: {rule.name}" for rule in rules]
        selected_rule_option = st.selectbox("选择规则", rule_options)
        
        if selected_rule_option:
            selected_rule_id = selected_rule_option.split(":")[0].strip()
            selected_rule = next((rule for rule in rules if rule.id == selected_rule_id), None)
            
            if selected_rule:
                # 使用表单编辑规则
                with st.expander("查看/编辑规则", expanded=True):
                    success, rule_data = create_rule_form(selected_rule, rule_type_mapping[rule_type], is_new=False)
                    
                    if success:
                        # 创建新的规则对象
                        updated_rule = Rule(rule_data)
                    
                        # 保存规则
                        if save_rule(updated_rule):
                            st.success(f"规则 {rule_data['id']} 已成功保存")
                            st.rerun()  # 刷新页面显示更新后的规则
                        else:
                            st.error("保存规则失败")
                
                # 显示规则YAML预览
                with st.expander("规则YAML预览"):
                    rule_yaml = yaml.dump(selected_rule.to_dict(), default_flow_style=False)
                    st.code(rule_yaml, language="yaml")
    else:
        st.info(f"没有找到 {rule_type} 类型的规则。")
    
    # 创建新规则的按钮
    st.divider()
    if "creating_new_rule" not in st.session_state:
        st.session_state.creating_new_rule = False
        
    if st.button("创建新规则"):
        st.session_state.creating_new_rule = True
    
    # 创建新规则的表单
    if st.session_state.get("creating_new_rule", False):
        st.subheader("创建新规则")
        
        success, rule_data = create_rule_form(None, rule_type_mapping[rule_type], is_new=True)
        
        if success:
            # 创建新的规则对象
            new_rule = Rule(rule_data)
            
            # 保存规则
            if save_rule(new_rule):
                st.success(f"规则 {rule_data['id']} 已成功创建")
                st.session_state.creating_new_rule = False
                st.rerun()  # 刷新页面显示新规则
            else:
                st.error("创建规则失败")
        
        if st.button("取消"):
            st.session_state.creating_new_rule = False
            st.rerun()
