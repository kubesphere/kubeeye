#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Prometheus 查询工具，用于从 Prometheus 获取指标数据
"""

import requests
from typing import Dict, List, Any, Optional
import json
from datetime import datetime, timedelta
import time
import base64

class PrometheusClient:
    """Prometheus 客户端类"""
    
    def __init__(self, config: Dict):
        """
        初始化 Prometheus 客户端
        
        Args:
            config: 包含以下字段的配置字典:
                - url: Prometheus URL
                - username: 可选的用户名
                - password: 可选的密码
                - token: 可选的访问令牌
                - enabled: 是否启用
        """
        self.url = config['url'].rstrip('/')
        self.username = config.get('username', '')
        self.password = config.get('password', '')
        self.token = config.get('token', '')
        self.enabled = config.get('enabled', False)
    
    def _get_headers(self) -> Dict:
        """获取请求头"""
        headers = {'Accept': 'application/json'}
        
        if self.token:
            headers['Authorization'] = f"Bearer {self.token}"
        elif self.username and self.password:
            auth = base64.b64encode(f"{self.username}:{self.password}".encode()).decode()
            headers['Authorization'] = f"Basic {auth}"
            
        return headers
    
    def query(self, query_expr: str, time_param: Optional[str] = None) -> Dict:
        """
        执行 Prometheus 查询
        
        Args:
            query_expr: Prometheus 查询表达式
            time_param: 可选的时间参数
            
        Returns:
            查询结果字典
        """
        if not self.enabled or not self.url:
            return {'status': 'error', 'error': 'Prometheus 未配置或未启用'}
        
        params = {'query': query_expr}
        if time_param:
            params['time'] = time_param
            
        # 使用带重试的请求
        return self._request_with_retry(f"{self.url}/api/v1/query", params)
    
    def query_range(self, query_expr: str, start_time: datetime, 
                   end_time: datetime, step: str = "15s") -> Dict:
        """
        执行范围查询
        
        Args:
            query_expr: Prometheus 查询表达式
            start_time: 开始时间
            end_time: 结束时间
            step: 步长
            
        Returns:
            查询结果字典
        """
        if not self.enabled or not self.url:
            return {'status': 'error', 'error': 'Prometheus 未配置或未启用'}
        
        params = {
            'query': query_expr,
            'start': start_time.timestamp(),
            'end': end_time.timestamp(),
            'step': step
        }
        
        # 使用带重试的请求
        return self._request_with_retry(f"{self.url}/api/v1/query_range", params)
    
    def alerts(self) -> Dict:
        """获取当前触发的告警"""
        if not self.enabled or not self.url:
            return {'status': 'error', 'error': 'Prometheus 未配置或未启用'}
        
        # 使用带重试的请求
        return self._request_with_retry(f"{self.url}/api/v1/alerts")
            
    def _request_with_retry(self, url: str, params: Dict = None, max_retries: int = 3) -> Dict:
        """
        执行带重试的 Prometheus API 请求
        
        Args:
            url: API URL
            params: 请求参数
            max_retries: 最大重试次数
            
        Returns:
            响应结果
        """
        retries = 0
        last_error = None
        
        # 根据URL协议决定是否验证SSL
        verify_ssl = url.lower().startswith('https://')
        
        # 如果是HTTP请求，禁用不安全请求的警告
        if not verify_ssl:
            import urllib3
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
        
        while retries < max_retries:
            try:
                response = requests.get(
                    url,
                    headers=self._get_headers(),
                    params=params,
                    timeout=30,
                    verify=verify_ssl  # 根据协议决定是否验证SSL
                )
                
                if response.status_code == 200:
                    return response.json()
                elif response.status_code == 503 or response.status_code >= 500:
                    # 服务不可用，尝试重试
                    retries += 1
                    if retries < max_retries:
                        time.sleep(1)  # 重试前等待1秒
                        continue
                    else:
                        return {
                            'status': 'error',
                            'error': f"多次重试后仍连接失败: HTTP {response.status_code}",
                            'detail': response.text
                        }
                else:
                    return {
                        'status': 'error',
                        'error': f"查询失败: HTTP {response.status_code}",
                        'detail': response.text
                    }
            except requests.exceptions.Timeout:
                # 超时重试
                retries += 1
                if retries < max_retries:
                    time.sleep(1)
                    continue
                else:
                    return {'status': 'error', 'error': "请求超时，多次重试无效"}
            except requests.exceptions.RequestException as e:
                last_error = str(e)
                retries += 1
                if retries < max_retries:
                    time.sleep(1)
                    continue
                else:
                    return {'status': 'error', 'error': f"请求异常: {last_error}"}
                    
    def test_connection(self) -> Dict:
        """测试与 Prometheus 的连接"""
        if not self.enabled or not self.url:
            return {'status': 'error', 'error': 'Prometheus 未配置或未启用'}
        
        # 根据URL协议决定是否验证SSL
        verify_ssl = self.url.lower().startswith('https://')
        
        # 如果是HTTP请求，禁用不安全请求的警告
        if not verify_ssl:
            import urllib3
            urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
            
        try:
            response = requests.get(
                f"{self.url}/api/v1/status/config",
                headers=self._get_headers(),
                timeout=10,
                verify=verify_ssl  # 根据协议决定是否验证SSL
            )
            
            if response.status_code == 200:
                return {'status': 'success', 'message': 'Prometheus 连接成功'}
            else:
                return {
                    'status': 'error',
                    'error': f"连接失败: HTTP {response.status_code}",
                    'detail': response.text
                }
        except requests.exceptions.RequestException as e:
            return {'status': 'error', 'error': f"请求异常: {str(e)}"}
