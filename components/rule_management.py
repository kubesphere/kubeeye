#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
现代化规则管理组件 - 支持GitOps模式
"""
import streamlit as st
import yaml
import json
import pandas as pd
import git
import os
import requests
import shutil
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional, Tuple
from utils.rule_loader import load_rules, save_rule, Rule, RULES_DIR
from utils.rule_manager import RuleManager

# GitOps配置
GITOPS_CONFIG_FILE = Path(__file__).parent.parent / "data" / "gitops_config.json"
DEFAULT_RULE_REPOS = [
    {
        "name": "KubeEye官方规则库",
        "url": "https://github.com/kubesphere/kubeeye",
        "branch": "rules",
        "description": "KubeEye官方维护的规则库，包含经过验证的Kubernetes检查规则和最佳实践"
    }
]

class GitOpsRuleManager:
    """GitOps规则管理器"""
    
    def __init__(self):
        self.config_file = GITOPS_CONFIG_FILE
        self.local_rules_dir = RULES_DIR
        self.git_rules_dir = Path(__file__).parent.parent / "data" / "git_rules"
        self.git_rules_dir.mkdir(parents=True, exist_ok=True)
        
    def load_config(self) -> Dict:
        """加载GitOps配置"""
        if self.config_file.exists():
            with open(self.config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        return {
            "mode": "local",  # local 或 gitops
            "current_repository": None,  # 当前启用的单个仓库
            "auto_sync": False,
            "sync_interval": 3600  # 秒
        }
    
    def save_config(self, config: Dict):
        """保存GitOps配置"""
        self.config_file.parent.mkdir(parents=True, exist_ok=True)
        with open(self.config_file, 'w', encoding='utf-8') as f:
            json.dump(config, f, indent=2, ensure_ascii=False)
    
    def clone_or_update_repo(self, repo_url: str, repo_name: str, branch: str = "main") -> Tuple[bool, str]:
        """克隆或更新Git仓库"""
        repo_path = self.git_rules_dir / repo_name
        
        try:
            if repo_path.exists():
                # 更新现有仓库
                repo = git.Repo(repo_path)
                origin = repo.remotes.origin
                origin.pull(branch)
                message = f"仓库 {repo_name} 更新成功"
            else:
                # 克隆新仓库
                git.Repo.clone_from(repo_url, repo_path, branch=branch)
                message = f"仓库 {repo_name} 克隆成功"
            
            return True, message
        except Exception as e:
            return False, f"操作失败: {str(e)}"
    
    def get_repo_rules(self, repo_name: str) -> List[Rule]:
        """获取Git仓库中的规则"""
        repo_path = self.git_rules_dir / repo_name
        rules = []
        
        if not repo_path.exists():
            return rules
        
        # 遍历仓库中的规则文件
        for rule_type in ["node", "prometheus", "opa"]:
            type_dir = repo_path / rule_type
            if type_dir.exists():
                for yaml_file in type_dir.glob("*.yaml"):
                    try:
                        with open(yaml_file, 'r', encoding='utf-8') as f:
                            rule_data = yaml.safe_load(f)
                        
                        if isinstance(rule_data, dict):
                            # 标记为Git规则
                            rule_data['source'] = 'git'
                            rule_data['repository'] = repo_name
                            rule_data['file_path'] = str(yaml_file.relative_to(repo_path))
                            rules.append(Rule(rule_data))
                    except Exception as e:
                        st.warning(f"加载规则文件 {yaml_file} 失败: {e}")
        
        return rules
    
    def sync_git_rule_to_local(self, rule: Rule, target_type: str) -> bool:
        """将Git规则同步到本地"""
        try:
            # 创建本地规则文件
            local_file = self.local_rules_dir / target_type / f"{rule.id}.yaml"
            local_file.parent.mkdir(parents=True, exist_ok=True)
            
            # 移除Git特有字段
            rule_data = rule.to_dict()
            rule_data.pop('source', None)
            rule_data.pop('repository', None) 
            rule_data.pop('file_path', None)
            
            with open(local_file, 'w', encoding='utf-8') as f:
                yaml.dump(rule_data, f, default_flow_style=False, allow_unicode=True)
            
            return True
        except Exception as e:
            st.error(f"同步规则失败: {e}")
            return False

def render_rule_management_tab():
    """渲染规则管理标签页"""
    st.markdown("### 🛠️ 规则管理中心")
    
    gitops_manager = GitOpsRuleManager()
    config = gitops_manager.load_config()
    
    # 模式选择器
    render_mode_selector(gitops_manager, config)
    
    # 根据模式显示不同的界面
    if config["mode"] == "local":
        render_local_mode(gitops_manager)
    else:
        render_gitops_mode(gitops_manager, config)

def render_mode_selector(gitops_manager: GitOpsRuleManager, config: Dict):
    """渲染模式选择器"""
    st.markdown("#### 🎯 规则管理模式")
    
    col1, col2, col3 = st.columns([2, 2, 2])
    
    with col1:
        current_mode = config.get("mode", "local")
        mode_options = {
            "local": "📁 本地模式",
            "gitops": "🔄 GitOps模式"
        }
        
        selected_mode = st.selectbox(
            "选择管理模式",
            options=list(mode_options.keys()),
            format_func=lambda x: mode_options[x],
            index=list(mode_options.keys()).index(current_mode)
        )
        
        if selected_mode != current_mode:
            config["mode"] = selected_mode
            gitops_manager.save_config(config)
            st.success(f"已切换到{mode_options[selected_mode]}")
            st.rerun()
    
    with col2:
        # 显示当前统计
        local_rules_count = sum(len(load_rules(rt)) for rt in ["node", "prometheus", "opa"])
        st.metric("本地规则", local_rules_count)
    
    with col3:
        # 快速操作
        if st.button("🔄 刷新", help="刷新规则列表"):
            st.rerun()

def render_local_mode(gitops_manager: GitOpsRuleManager):
    """渲染本地模式界面"""
    st.markdown("#### 📁 本地规则")
    
    # 直接显示规则列表
    render_local_rule_list()

def render_gitops_mode(gitops_manager: GitOpsRuleManager, config: Dict):
    """渲染GitOps模式界面"""
    st.markdown("#### 🔄 GitOps规则管理")
    
    # 选项卡
    tab_manage, tab_browse = st.tabs(["📚 仓库管理", "🔍 规则浏览"])
    
    with tab_manage:
        render_repository_management(gitops_manager, config)
    
    with tab_browse:
        render_git_rule_browser(gitops_manager, config)

def render_local_rule_list():
    """渲染本地规则列表"""
    # 规则类型选择
    col1, col2 = st.columns([3, 1])
    
    with col1:
        selected_types = st.multiselect(
            "选择规则类型",
            ["node", "prometheus", "opa"],
            default=["node", "prometheus", "opa"],
            format_func=lambda x: {"node": "🖥️ 节点规则", "prometheus": "📊 监控规则", "opa": "🔒 安全规则"}[x]
        )
    
    with col2:
        show_disabled = st.checkbox("显示已禁用", value=False)
    
    # 加载并显示规则
    all_rules = []
    for rule_type in selected_types:
        rules = load_rules(rule_type, include_disabled=show_disabled)
        all_rules.extend(rules)
    
    if not all_rules:
        st.info("没有找到规则，请检查规则目录或创建新规则。")
        return
    
    # 创建规则表格
    rule_data = []
    severity_map = {
        "info": "ℹ️ 信息", 
        "low": "🟢 低", 
        "warning": "⚠️ 警告", 
        "medium": "🟡 中等",
        "high": "🟠 高",
        "critical": "🚨 严重"
    }
    
    for rule in all_rules:
        rule_data.append({
            "ID": rule.id,
            "名称": rule.name,
            "类型": {"node": "🖥️ 节点", "prometheus": "📊 监控", "opa": "🔒 安全"}[rule.type],
            "状态": "✅ 启用" if rule.enabled else "❌ 禁用",
            "严重性": severity_map.get(rule.severity, f"❓ {rule.severity}"),
            "描述": rule.description[:50] + "..." if len(rule.description) > 50 else rule.description
        })
    
    if rule_data:
        df = pd.DataFrame(rule_data)
        st.dataframe(df, use_container_width=True)

def render_repository_management(gitops_manager: GitOpsRuleManager, config: Dict):
    """渲染仓库管理"""
    current_repo = config.get("current_repository", None)
    
    if current_repo:
        # 当前仓库信息
        st.markdown("### 📋 当前仓库")
        
        col1, col2 = st.columns([3, 1])
        
        with col1:
            st.markdown(f"**名称:** {current_repo['name']}")
            st.markdown(f"**地址:** `{current_repo['url']}`")
            st.markdown(f"**分支:** `{current_repo['branch']}`")
            if current_repo.get('description'):
                st.markdown(f"**描述:** {current_repo['description']}")
        
        with col2:
            if st.button("🔄 同步仓库", type="primary"):
                sync_repository(gitops_manager, current_repo)
            
            if st.button("❌ 切换仓库"):
                config["current_repository"] = None
                gitops_manager.save_config(config)
                st.success("已清除当前仓库，请选择新仓库")
                st.rerun()
    
    else:
        # 仓库选择
        st.markdown("### 🌟 选择Git规则仓库")
        
        # 官方推荐
        official_repo = DEFAULT_RULE_REPOS[0]
        
        st.markdown("#### 🏆 官方推荐")
        col1, col2 = st.columns([3, 1])
        
        with col1:
            st.markdown(f"**{official_repo['name']}**")
            st.markdown(f"地址: `{official_repo['url']}`")
            st.markdown(f"描述: {official_repo['description']}")
        
        with col2:
            if st.button("🚀 启用官方仓库", type="primary"):
                config["current_repository"] = official_repo.copy()
                gitops_manager.save_config(config)
                st.success("✅ 已启用官方仓库")
                st.rerun()

def render_git_rule_browser(gitops_manager: GitOpsRuleManager, config: Dict):
    """渲染Git规则浏览器"""
    current_repo = config.get("current_repository", None)
    
    if not current_repo:
        st.warning("💡 请先在'仓库管理'中启用一个Git仓库")
        return
    
    # 检查仓库状态
    repo_path = gitops_manager.git_rules_dir / current_repo["name"]
    if not repo_path.exists():
        st.warning(f"⚠️ 仓库 **{current_repo['name']}** 尚未同步")
        if st.button("🔄 立即同步", type="primary"):
            sync_repository(gitops_manager, current_repo)
        return
    
    # 加载规则
    git_rules = gitops_manager.get_repo_rules(current_repo["name"])
    
    if not git_rules:
        st.info("📭 该仓库中暂无规则文件")
        return
    
    st.success(f"📚 仓库: **{current_repo['name']}** | 共 **{len(git_rules)}** 个规则")
    
    # 简单的规则列表
    for i, rule in enumerate(git_rules):
        with st.expander(f"📋 {rule.name} ({rule.type})", expanded=False):
            col1, col2 = st.columns([3, 1])
            
            with col1:
                st.markdown(f"**描述:** {rule.description}")
                st.markdown(f"**类型:** {rule.type} | **严重性:** {rule.severity}")
            
            with col2:
                if st.button(f"📥 导入", key=f"import_{rule.id}_{i}", type="primary"):
                    if gitops_manager.sync_git_rule_to_local(rule, rule.type):
                        st.success("✅ 已导入到本地")
                        st.rerun()
                    else:
                        st.error("❌ 导入失败")

def sync_repository(gitops_manager: GitOpsRuleManager, repo: Dict):
    """同步单个仓库"""
    with st.spinner(f"正在同步仓库 {repo['name']}..."):
        success, message = gitops_manager.clone_or_update_repo(
            repo["url"], 
            repo["name"], 
            repo.get("branch", "main")
        )
        
        if success:
            st.success(message)
            rules_count = len(gitops_manager.get_repo_rules(repo["name"]))
            st.info(f"发现 {rules_count} 个规则")
        else:
            st.error(message)
        
        st.rerun()
