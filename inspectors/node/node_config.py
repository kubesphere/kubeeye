#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
节点巡检并发配置管理
"""

import os
from typing import Dict, Any
from dataclasses import dataclass

@dataclass
class NodeInspectorConfig:
    """节点巡检器配置"""
    
    # 并发控制
    max_workers: int = 5           # 最大并发线程数
    timeout: int = 30              # 单个节点命令执行超时时间（秒）
    
    # 连接配置
    connection_timeout: int = 10   # SSH连接超时时间（秒）
    retry_attempts: int = 2        # 连接失败重试次数
    retry_delay: int = 1           # 重试间隔（秒）
    
    # 性能优化
    enable_connection_pool: bool = True   # 启用连接池
    pool_size: int = 10            # 连接池大小
    keep_alive: bool = True        # 保持连接活跃
    
    # 日志配置
    verbose_logging: bool = False  # 详细日志
    log_command_output: bool = False  # 记录命令输出
    
    @classmethod
    def from_env(cls) -> 'NodeInspectorConfig':
        """从环境变量创建配置"""
        return cls(
            max_workers=int(os.getenv('NODE_INSPECTOR_MAX_WORKERS', '5')),
            timeout=int(os.getenv('NODE_INSPECTOR_TIMEOUT', '30')),
            connection_timeout=int(os.getenv('NODE_INSPECTOR_CONNECTION_TIMEOUT', '10')),
            retry_attempts=int(os.getenv('NODE_INSPECTOR_RETRY_ATTEMPTS', '2')),
            retry_delay=int(os.getenv('NODE_INSPECTOR_RETRY_DELAY', '1')),
            enable_connection_pool=os.getenv('NODE_INSPECTOR_CONNECTION_POOL', 'true').lower() == 'true',
            pool_size=int(os.getenv('NODE_INSPECTOR_POOL_SIZE', '10')),
            keep_alive=os.getenv('NODE_INSPECTOR_KEEP_ALIVE', 'true').lower() == 'true',
            verbose_logging=os.getenv('NODE_INSPECTOR_VERBOSE', 'false').lower() == 'true',
            log_command_output=os.getenv('NODE_INSPECTOR_LOG_OUTPUT', 'false').lower() == 'true'
        )
    
    @classmethod
    def adaptive(cls, node_count: int) -> 'NodeInspectorConfig':
        """根据节点数量自适应配置"""
        if node_count <= 3:
            max_workers = node_count
            timeout = 30
        elif node_count <= 10:
            max_workers = min(5, node_count)
            timeout = 25
        elif node_count <= 20:
            max_workers = min(8, node_count)
            timeout = 20
        else:
            max_workers = min(10, node_count)
            timeout = 15
        
        return cls(
            max_workers=max_workers,
            timeout=timeout,
            connection_timeout=min(10, timeout // 3),
            retry_attempts=2 if node_count <= 10 else 1,
            verbose_logging=node_count <= 5  # 节点少时开启详细日志
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """转换为字典"""
        return {
            'max_workers': self.max_workers,
            'timeout': self.timeout,
            'connection_timeout': self.connection_timeout,
            'retry_attempts': self.retry_attempts,
            'retry_delay': self.retry_delay,
            'enable_connection_pool': self.enable_connection_pool,
            'pool_size': self.pool_size,
            'keep_alive': self.keep_alive,
            'verbose_logging': self.verbose_logging,
            'log_command_output': self.log_command_output
        }
    
    def validate(self) -> List[str]:
        """验证配置有效性"""
        issues = []
        
        if self.max_workers < 1:
            issues.append("max_workers必须大于0")
        if self.max_workers > 20:
            issues.append("max_workers不建议超过20，可能导致资源过载")
        
        if self.timeout < 5:
            issues.append("timeout不建议小于5秒")
        if self.timeout > 300:
            issues.append("timeout不建议超过5分钟")
        
        if self.connection_timeout < 1:
            issues.append("connection_timeout必须大于0")
        
        if self.retry_attempts < 0:
            issues.append("retry_attempts不能小于0")
        if self.retry_attempts > 5:
            issues.append("retry_attempts不建议超过5次")
        
        return issues
