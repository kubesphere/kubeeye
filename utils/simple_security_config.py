#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
KubeEye 简化安全配置 - 只读取基本配置参数
"""

import os
import yaml
import logging
from typing import Dict, List, Optional
from pathlib import Path

logger = logging.getLogger(__name__)

class SimpleSecurityConfig:
    """简化的安全配置类 - 只处理可调整参数"""
    
    def __init__(self, config_file: str = "config/security.yaml"):
        """
        初始化简化安全配置
        
        Args:
            config_file: 配置文件路径
        """
        self.config_file = config_file
        self.config = self._load_config()
    
    def _load_config(self) -> Dict:
        """加载配置文件"""
        try:
            config_path = Path(self.config_file)
            if config_path.exists():
                with open(config_path, 'r', encoding='utf-8') as f:
                    config = yaml.safe_load(f) or {}
            else:
                logger.warning(f"配置文件不存在: {config_path}，使用默认配置")
                config = {}
            
            # 设置默认值
            return {
                'max_command_length': config.get('max_command_length', 1000),
                'command_timeout': config.get('command_timeout', 30),
                'audit_log_path': config.get('audit_log_path', 'data/logs/security_audit.log'),
                'audit_retention_days': config.get('audit_retention_days', 90),
                'enable_detailed_logging': config.get('enable_detailed_logging', True),
                'allowed_ports': config.get('allowed_ports', [22, 80, 443, 6443, 8080, 9090, 10250]),
                'blocked_ips': config.get('blocked_ips', []),
                'require_key_auth': config.get('require_key_auth', False)
            }
        except Exception as e:
            logger.error(f"加载安全配置失败: {e}")
            return self._get_default_config()
    
    def _get_default_config(self) -> Dict:
        """获取默认配置"""
        return {
            'max_command_length': 1000,
            'command_timeout': 30,
            'audit_log_path': 'data/logs/security_audit.log',
            'audit_retention_days': 90,
            'enable_detailed_logging': True,
            'allowed_ports': [22, 80, 443, 6443, 8080, 9090, 10250],
            'blocked_ips': [],
            'require_key_auth': False
        }
    
    def get(self, key: str, default=None):
        """获取配置值"""
        return self.config.get(key, default)
    
    @property
    def max_command_length(self) -> int:
        """最大命令长度"""
        return self.config['max_command_length']
    
    @property
    def command_timeout(self) -> int:
        """命令超时时间"""
        return self.config['command_timeout']
    
    @property
    def audit_log_path(self) -> str:
        """审计日志路径"""
        return self.config['audit_log_path']
    
    @property
    def audit_retention_days(self) -> int:
        """审计日志保留天数"""
        return self.config['audit_retention_days']
    
    @property
    def enable_detailed_logging(self) -> bool:
        """是否启用详细日志"""
        return self.config['enable_detailed_logging']
    
    @property
    def allowed_ports(self) -> List[int]:
        """允许的端口列表"""
        return self.config['allowed_ports']
    
    @property
    def blocked_ips(self) -> List[str]:
        """阻止的IP列表"""
        return self.config['blocked_ips']
    
    @property
    def require_key_auth(self) -> bool:
        """是否要求密钥认证"""
        return self.config['require_key_auth']

# 全局配置实例
_global_security_config = None

def get_security_config() -> SimpleSecurityConfig:
    """获取全局安全配置实例"""
    global _global_security_config
    if _global_security_config is None:
        _global_security_config = SimpleSecurityConfig()
    return _global_security_config
