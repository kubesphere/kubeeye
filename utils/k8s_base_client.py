#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 基础客户端 - 提供共同的初始化和配置逻辑
减少 k8s_client.py 和 k8s_dynamic_client.py 之间的代码重复
"""

import tempfile
import os
import logging
from typing import Optional, Tuple
from kubernetes import client, config
from kubernetes.client.rest import ApiException

logger = logging.getLogger(__name__)

class K8sBaseClient:
    """Kubernetes 基础客户端，提供通用的初始化和配置功能"""
    
    def __init__(self, kubeconfig_content: str = None):
        """
        初始化基础客户端
        
        Args:
            kubeconfig_content: kubeconfig 文件内容
        """
        self.kubeconfig_content = kubeconfig_content
        self.temp_config = None
        self.initialized = False
        
    def init_client_base(self) -> bool:
        """
        通用的客户端初始化逻辑
        
        Returns:
            成功返回 True，失败返回 False
        """
        try:
            if self.kubeconfig_content:
                # 创建临时文件存储 kubeconfig
                self.temp_config = tempfile.NamedTemporaryFile(delete=False)
                self.temp_config.write(self.kubeconfig_content.encode())
                self.temp_config.flush()
                config.load_kube_config(self.temp_config.name)
            else:
                # 尝试默认方式加载配置
                config.load_kube_config()
                
            # 配置SSL证书验证
            client.Configuration.set_default(self._configure_no_verify_ssl())
            
            self.initialized = True
            logger.info("Kubernetes基础客户端初始化成功")
            return True
            
        except Exception as e:
            logger.error(f"初始化 Kubernetes 基础客户端失败: {e}")
            self.initialized = False
            return False
    
    def _configure_no_verify_ssl(self):
        """
        配置Kubernetes客户端跳过SSL证书验证，用于自签名证书环境
        
        Returns:
            配置好的客户端配置
        """
        # 获取当前客户端配置
        configuration = client.Configuration.get_default_copy()
        
        # 禁用SSL证书验证
        configuration.verify_ssl = False
        configuration.ssl_ca_cert = None
        
        # 设置警告消息
        logger.warning("已禁用SSL证书验证，这可能存在安全风险")
        
        return configuration
    
    def test_connection(self) -> Tuple[bool, str]:
        """
        测试连接到Kubernetes集群
        
        Returns:
            (是否成功, 消消息)
        """
        try:
            if not self.initialized:
                return False, "客户端未初始化"
                
            # 尝试获取集群版本信息
            version_api = client.VersionApi()
            version = version_api.get_code().git_version
            return True, f"连接成功，集群版本: {version}"
            
        except ApiException as e:
            logger.error(f"连接测试失败: {e}")
            return False, f"API错误: {e.reason}"
        except Exception as e:
            logger.error(f"连接测试失败: {e}")
            return False, f"连接失败: {str(e)}"
    
    def __del__(self):
        """析构函数，删除临时文件"""
        if self.temp_config:
            try:
                self.temp_config.close()
                os.unlink(self.temp_config.name)
            except:
                pass
