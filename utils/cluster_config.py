#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
集群配置管理模块，用于管理集群连接配置信息
"""

import os
import json
import yaml
import logging
from pathlib import Path
from typing import Dict, List, Optional, Any
from utils.crypto_utils import encrypt_password, decrypt_password

logger = logging.getLogger(__name__)

# 数据目录定义
DATA_DIR = Path(__file__).parent.parent / "data"
CLUSTERS_DIR = DATA_DIR / "clusters"
RESULTS_DIR = DATA_DIR / "results"

# 确保目录存在
os.makedirs(CLUSTERS_DIR, exist_ok=True)
os.makedirs(RESULTS_DIR, exist_ok=True)

class ClusterConfig:
    """集群配置管理类"""
    
    def __init__(self, cluster_name: str):
        self.cluster_name = cluster_name
        self.config_path = CLUSTERS_DIR / f"{cluster_name}.json"
        self.config = self._load_config()
    
    def _load_config(self) -> Dict:
        """加载集群配置"""
        if self.config_path.exists():
            with open(self.config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        return {
            'name': self.cluster_name,
            'nodes': [],
            'prometheus': {
                'url': '',
                'username': '',
                'password': '',
                'token': '',
                'enabled': False
            },
            'kubeconfig': '',
            'created_at': '',
            'updated_at': ''
        }
    
    def save_config(self) -> None:
        """保存集群配置"""
        with open(self.config_path, 'w', encoding='utf-8') as f:
            json.dump(self.config, f, ensure_ascii=False, indent=2)
    
    def update_node(self, node_info: Dict) -> None:
        """添加或更新节点信息"""
        # 创建一个节点信息的副本，以便进行加密
        node_copy = node_info.copy()
        
        # 如果是密码认证且密码未加密，则加密密码
        if (node_copy.get('auth_type') == 'password' and 
            node_copy.get('password') and 
            not node_copy.get('password_encrypted')):
            node_copy['password'] = encrypt_password(node_copy['password'])
            node_copy['password_encrypted'] = True
        
        # 更新节点信息
        for i, node in enumerate(self.config['nodes']):
            if node['ip'] == node_info['ip']:
                self.config['nodes'][i] = node_copy
                self.save_config()
                return
        
        # 如果不存在则添加新节点
        self.config['nodes'].append(node_copy)
        self.save_config()
    
    def remove_node(self, node_ip: str) -> bool:
        """删除节点信息"""
        for i, node in enumerate(self.config['nodes']):
            if node['ip'] == node_ip:
                del self.config['nodes'][i]
                self.save_config()
                return True
        return False
    
    def update_prometheus(self, prometheus_info: Dict) -> None:
        """更新 Prometheus 配置"""
        # 创建配置的副本以进行加密
        prometheus_copy = prometheus_info.copy()
        
        # 加密密码和令牌
        if prometheus_copy.get('password') and not prometheus_copy.get('password_encrypted'):
            prometheus_copy['password'] = encrypt_password(prometheus_copy['password'])
            prometheus_copy['password_encrypted'] = True
            
        if prometheus_copy.get('token') and not prometheus_copy.get('token_encrypted'):
            prometheus_copy['token'] = encrypt_password(prometheus_copy['token'])
            prometheus_copy['token_encrypted'] = True
            
        self.config['prometheus'].update(prometheus_copy)
        self.save_config()
    
    def update_kubeconfig(self, kubeconfig: str) -> None:
        """更新 Kubeconfig 配置"""
        self.config['kubeconfig'] = kubeconfig
        self.save_config()
    
    def get_nodes(self) -> List[Dict]:
        """获取集群节点列表"""
        nodes = []
        for node in self.config['nodes']:
            node_copy = node.copy()
            # 如果密码是加密的，进行解密
            if node_copy.get('auth_type') == 'password' and node_copy.get('password_encrypted'):
                try:
                    node_copy['password'] = decrypt_password(node_copy['password'])
                    node_copy['password_encrypted'] = False
                except Exception as e:
                    logger.error(f"解密节点 {node_copy.get('ip')} 的密码时出错: {str(e)}")
            nodes.append(node_copy)
        return nodes
    
    def get_prometheus_config(self) -> Dict:
        """获取 Prometheus 配置"""
        prometheus_config = self.config['prometheus'].copy()
        
        # 如果密码是加密的，进行解密
        if prometheus_config.get('password_encrypted'):
            try:
                prometheus_config['password'] = decrypt_password(prometheus_config['password'])
                prometheus_config['password_encrypted'] = False
            except Exception as e:
                logger.error(f"解密 Prometheus 密码时出错: {str(e)}")
                
        # 如果令牌是加密的，进行解密
        if prometheus_config.get('token_encrypted'):
            try:
                prometheus_config['token'] = decrypt_password(prometheus_config['token'])
                prometheus_config['token_encrypted'] = False
            except Exception as e:
                logger.error(f"解密 Prometheus 令牌时出错: {str(e)}")
                
        return prometheus_config
    
    def get_kubeconfig(self) -> str:
        """获取 Kubeconfig"""
        return self.config['kubeconfig']


def list_clusters() -> List[str]:
    """列出所有集群名称"""
    clusters = []
    for file_path in CLUSTERS_DIR.glob('*.json'):
        clusters.append(file_path.stem)
    return clusters


def get_cluster(cluster_name: str) -> ClusterConfig:
    """获取集群配置对象"""
    return ClusterConfig(cluster_name)


def delete_cluster(cluster_name: str) -> bool:
    """删除集群配置"""
    config_path = CLUSTERS_DIR / f"{cluster_name}.json"
    if config_path.exists():
        config_path.unlink()
        return True
    return False


def load_kubeconfig(kubeconfig_str: str) -> Dict:
    """加载 kubeconfig 内容为字典"""
    try:
        return yaml.safe_load(kubeconfig_str)
    except yaml.YAMLError:
        return {}
