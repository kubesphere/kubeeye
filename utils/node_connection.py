#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
节点连接管理模块，用于通过 SSH 连接集群节点并执行命令
"""

import paramiko
import socket
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
    
    def connect(self) -> Tuple[bool, str]:
        """
        连接到节点
        
        Returns:
            成功连接时返回 (True, "")，失败时返回 (False, error_message)
        """
        try:
            if self.node_info['auth_type'] == 'password':
                self.client.connect(
                    hostname=self.node_info['ip'],
                    port=int(self.node_info['port']),
                    username=self.node_info['username'],
                    password=self.node_info['password'],
                    timeout=10
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
                    timeout=10
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
            stdin, stdout, stderr = self.client.exec_command(command, timeout=60)
            return True, stdout.read().decode('utf-8'), stderr.read().decode('utf-8')
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
