#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
节点连接管理模块，用于通过 SSH 连接集群节点并执行命令
"""

import paramiko
import socket
import logging
from typing import Dict, List, Tuple, Optional
import os
from pathlib import Path

class NodeConnection:
    """节点 SSH 连接类"""
    
    def __init__(self, node_info: Dict):
        """
        初始化节点连接
        
        Args:
            node_info: 节点信息字典，包含：
                - ip: 节点 IP
                - port: SSH 端口
                - username: SSH 用户名
                - auth_type: 认证类型 ('password' 或 'key')
                - password: 密码（如果 auth_type 是 'password'）
                - key_path: 密钥路径（如果 auth_type 是 'key'）
        """
        self.node_info = node_info
        self.client = paramiko.SSHClient()
        self.client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        self.connected = False
    
    def __enter__(self):
        """上下文管理器入口，连接到节点并返回自身"""
        self.connect()
        return self
        
    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器退出，关闭连接"""
        self.close()
        return False  # 让异常正常传播
    
    def connect(self) -> Tuple[bool, str]:
        """
        连接到节点
        
        Returns:
            成功连接时返回 (True, "")，失败时返回 (False, error_message)
        """
        try:
            # 记录连接信息
            logging.info(f"正在连接到节点: {self.node_info['ip']}:{self.node_info['port']} 用户名: {self.node_info['username']}")
            
            # 确保我们不会要求交互式输入密码
            # 设置look_for_keys=False可以避免Paramiko尝试使用SSH代理或寻找密钥文件
            # 设置allow_agent=False可以避免使用SSH代理
            if self.node_info['auth_type'] == 'password':
                self.client.connect(
                    hostname=self.node_info['ip'],
                    port=int(self.node_info['port']),
                    username=self.node_info['username'],
                    password=self.node_info['password'],
                    timeout=10,
                    look_for_keys=False,
                    allow_agent=False
                )
            else:  # key-based auth
                key_path = self.node_info['key_path']
                if not os.path.isfile(key_path):
                    return False, f"密钥文件 {key_path} 不存在"
                
                key = paramiko.RSAKey.from_private_key_file(key_path)
                self.client.connect(
                    hostname=self.node_info['ip'],
                    port=int(self.node_info['port']),
                    username=self.node_info['username'],
                    pkey=key,
                    timeout=10,
                    look_for_keys=False,
                    allow_agent=False
                )
            
            self.connected = True
            return True, ""
        except socket.timeout:
            return False, "连接超时"
        except paramiko.AuthenticationException:
            return False, "认证失败"
        except paramiko.SSHException as e:
            return False, f"SSH 错误: {str(e)}"
        except Exception as e:
            return False, f"连接错误: {str(e)}"
    
    def execute_command(self, command: str) -> Tuple[bool, str, str]:
        """
        在节点上执行命令
        
        Args:
            command: 要执行的命令
            
        Returns:
            元组 (success, stdout, stderr)
        """
        if not self.connected:
            success, message = self.connect()
            if not success:
                return False, "", message
        
        try:
            # 执行命令并等待完成
            stdin, stdout, stderr = self.client.exec_command(command, timeout=60)
            
            # 关闭标准输入
            stdin.close()
            
            # 读取标准输出和标准错误
            stdout_data = stdout.read().decode('utf-8')
            stderr_data = stderr.read().decode('utf-8')
            
            # 等待命令完成并获取退出状态码
            exit_status = stdout.channel.recv_exit_status()
            
            # 确保所有通道关闭
            stdout.close()
            stderr.close()
            
            success = exit_status == 0
            return success, stdout_data, stderr_data
        except Exception as e:
            return False, "", str(e)
    
    def close(self) -> None:
        """关闭连接"""
        if self.connected:
            self.client.close()
            self.connected = False


def test_node_connection(node_info: Dict) -> Tuple[bool, str]:
    """
    测试节点连接
    
    Args:
        node_info: 节点配置信息
        
    Returns:
        (success, message) 元组
    """
    conn = NodeConnection(node_info)
    success, message = conn.connect()
    if success:
        conn.close()
    return success, message
