#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes证书检查模块，用于检查集群证书的过期状态
"""

import yaml
import base64
import logging
import tempfile
from datetime import datetime, timezone
from typing import Dict, Optional, Tuple
from pathlib import Path
from cryptography import x509
from cryptography.hazmat.backends import default_backend

logger = logging.getLogger(__name__)

def get_cluster_cert_status(cluster_name: str, kubeconfig_content: str) -> Dict[str, any]:
    """
    检查集群证书状态
    
    Args:
        cluster_name: 集群名称
        kubeconfig_content: kubeconfig文件内容
        
    Returns:
        Dict: 包含证书状态信息的字典
        {
            'status': 'valid|warning|critical|expired|unknown',
            'days_remaining': int,  # 剩余天数，负数表示已过期
            'cert_info': {
                'issuer': str,
                'subject': str,
                'not_before': str,
                'not_after': str
            },
            'error': str  # 如果有错误
        }
    """
    try:
        if not kubeconfig_content:
            return {
                'status': 'unknown',
                'days_remaining': None,
                'error': 'kubeconfig内容为空'
            }
            
        # 解析kubeconfig
        kubeconfig = yaml.safe_load(kubeconfig_content)
        if not kubeconfig:
            return {
                'status': 'unknown',
                'days_remaining': None,
                'error': '无法解析kubeconfig内容'
            }
            
        # 检查客户端证书
        cert_info = _extract_client_cert_info(kubeconfig)
        if not cert_info:
            # 如果没有客户端证书，检查集群CA证书
            cert_info = _extract_cluster_ca_info(kubeconfig)
            
        if not cert_info:
            return {
                'status': 'unknown',
                'days_remaining': None,
                'error': '未找到有效的证书信息'
            }
            
        # 计算证书状态
        return _calculate_cert_status(cert_info)
        
    except Exception as e:
        logger.error(f"检查集群 {cluster_name} 证书状态时发生错误: {str(e)}")
        return {
            'status': 'unknown',
            'days_remaining': None,
            'error': f'检查证书时发生错误: {str(e)}'
        }


def _extract_client_cert_info(kubeconfig: Dict) -> Optional[Dict]:
    """从kubeconfig中提取客户端证书信息"""
    try:
        users = kubeconfig.get('users', [])
        for user in users:
            user_info = user.get('user', {})
            
            # 检查client-certificate-data
            cert_data = user_info.get('client-certificate-data')
            if cert_data:
                cert_bytes = base64.b64decode(cert_data)
                return _parse_certificate(cert_bytes)
                
            # 检查client-certificate文件路径
            cert_path = user_info.get('client-certificate')
            if cert_path and Path(cert_path).exists():
                with open(cert_path, 'rb') as f:
                    cert_bytes = f.read()
                return _parse_certificate(cert_bytes)
                
        return None
        
    except Exception as e:
        logger.error(f"提取客户端证书信息失败: {str(e)}")
        return None


def _extract_cluster_ca_info(kubeconfig: Dict) -> Optional[Dict]:
    """从kubeconfig中提取集群CA证书信息"""
    try:
        clusters = kubeconfig.get('clusters', [])
        for cluster in clusters:
            cluster_info = cluster.get('cluster', {})
            
            # 检查certificate-authority-data
            ca_data = cluster_info.get('certificate-authority-data')
            if ca_data:
                cert_bytes = base64.b64decode(ca_data)
                return _parse_certificate(cert_bytes)
                
            # 检查certificate-authority文件路径
            ca_path = cluster_info.get('certificate-authority')
            if ca_path and Path(ca_path).exists():
                with open(ca_path, 'rb') as f:
                    cert_bytes = f.read()
                return _parse_certificate(cert_bytes)
                
        return None
        
    except Exception as e:
        logger.error(f"提取集群CA证书信息失败: {str(e)}")
        return None


def _parse_certificate(cert_bytes: bytes) -> Optional[Dict]:
    """解析证书并提取信息"""
    try:
        # 尝试解析PEM格式证书
        cert = x509.load_pem_x509_certificate(cert_bytes, default_backend())
        
        return {
            'issuer': cert.issuer.rfc4514_string(),
            'subject': cert.subject.rfc4514_string(),
            'not_before': cert.not_valid_before_utc,
            'not_after': cert.not_valid_after_utc,
            'serial_number': str(cert.serial_number)
        }
        
    except Exception as e:
        logger.error(f"解析证书失败: {str(e)}")
        return None


def _calculate_cert_status(cert_info: Dict) -> Dict[str, any]:
    """根据证书信息计算状态"""
    try:
        not_after = cert_info['not_after']
        not_before = cert_info['not_before']
        now = datetime.now(timezone.utc)
        
        # 确保时间对象有时区信息
        if not_after.tzinfo is None:
            not_after = not_after.replace(tzinfo=timezone.utc)
        if not_before.tzinfo is None:
            not_before = not_before.replace(tzinfo=timezone.utc)
            
        # 计算剩余天数
        time_remaining = not_after - now
        days_remaining = time_remaining.days
        
        # 确定状态
        if now > not_after:
            status = 'expired'
        elif days_remaining <= 7:  # 7天内过期
            status = 'critical'
        elif days_remaining <= 30:  # 30天内过期
            status = 'warning'
        else:
            status = 'valid'
            
        return {
            'status': status,
            'days_remaining': days_remaining,
            'cert_info': {
                'issuer': cert_info['issuer'],
                'subject': cert_info['subject'],
                'not_before': not_before.strftime('%Y-%m-%d %H:%M:%S UTC'),
                'not_after': not_after.strftime('%Y-%m-%d %H:%M:%S UTC'),
                'serial_number': cert_info.get('serial_number', '')
            }
        }
        
    except Exception as e:
        logger.error(f"计算证书状态失败: {str(e)}")
        return {
            'status': 'unknown',
            'days_remaining': None,
            'error': f'计算证书状态失败: {str(e)}'
        }


def check_certificate_file(cert_path: str) -> Dict[str, any]:
    """
    检查单个证书文件的状态
    
    Args:
        cert_path: 证书文件路径
        
    Returns:
        Dict: 证书状态信息
    """
    try:
        if not Path(cert_path).exists():
            return {
                'status': 'unknown',
                'days_remaining': None,
                'error': f'证书文件不存在: {cert_path}'
            }
            
        with open(cert_path, 'rb') as f:
            cert_bytes = f.read()
            
        cert_info = _parse_certificate(cert_bytes)
        if not cert_info:
            return {
                'status': 'unknown',
                'days_remaining': None,
                'error': f'无法解析证书文件: {cert_path}'
            }
            
        return _calculate_cert_status(cert_info)
        
    except Exception as e:
        logger.error(f"检查证书文件 {cert_path} 失败: {str(e)}")
        return {
            'status': 'unknown',
            'days_remaining': None,
            'error': f'检查证书文件失败: {str(e)}'
        }


def get_certificate_details(kubeconfig_content: str) -> Dict[str, any]:
    """
    获取kubeconfig中所有证书的详细信息
    
    Args:
        kubeconfig_content: kubeconfig文件内容
        
    Returns:
        Dict: 包含所有证书详细信息的字典
    """
    try:
        kubeconfig = yaml.safe_load(kubeconfig_content)
        if not kubeconfig:
            return {'error': '无法解析kubeconfig内容'}
            
        details = {
            'client_certificates': [],
            'cluster_ca_certificates': [],
            'summary': {
                'total_certs': 0,
                'expired_certs': 0,
                'expiring_soon_certs': 0,
                'earliest_expiry': None
            }
        }
        
        # 检查客户端证书
        users = kubeconfig.get('users', [])
        for user in users:
            user_name = user.get('name', 'unknown')
            user_info = user.get('user', {})
            
            cert_data = user_info.get('client-certificate-data')
            cert_path = user_info.get('client-certificate')
            
            if cert_data:
                cert_bytes = base64.b64decode(cert_data)
                cert_info = _parse_certificate(cert_bytes)
                if cert_info:
                    status = _calculate_cert_status(cert_info)
                    details['client_certificates'].append({
                        'user': user_name,
                        'type': 'client-certificate-data',
                        **status
                    })
            elif cert_path:
                status = check_certificate_file(cert_path)
                details['client_certificates'].append({
                    'user': user_name,
                    'type': 'client-certificate-file',
                    'path': cert_path,
                    **status
                })
        
        # 检查集群CA证书
        clusters = kubeconfig.get('clusters', [])
        for cluster in clusters:
            cluster_name = cluster.get('name', 'unknown')
            cluster_info = cluster.get('cluster', {})
            
            ca_data = cluster_info.get('certificate-authority-data')
            ca_path = cluster_info.get('certificate-authority')
            
            if ca_data:
                cert_bytes = base64.b64decode(ca_data)
                cert_info = _parse_certificate(cert_bytes)
                if cert_info:
                    status = _calculate_cert_status(cert_info)
                    details['cluster_ca_certificates'].append({
                        'cluster': cluster_name,
                        'type': 'certificate-authority-data',
                        **status
                    })
            elif ca_path:
                status = check_certificate_file(ca_path)
                details['cluster_ca_certificates'].append({
                    'cluster': cluster_name,
                    'type': 'certificate-authority-file',
                    'path': ca_path,
                    **status
                })
        
        # 计算摘要信息
        all_certs = details['client_certificates'] + details['cluster_ca_certificates']
        details['summary']['total_certs'] = len(all_certs)
        
        earliest_expiry = None
        for cert in all_certs:
            if cert.get('status') == 'expired':
                details['summary']['expired_certs'] += 1
            elif cert.get('status') in ['critical', 'warning']:
                details['summary']['expiring_soon_certs'] += 1
            
            days_remaining = cert.get('days_remaining')
            if days_remaining is not None:
                if earliest_expiry is None or days_remaining < earliest_expiry:
                    earliest_expiry = days_remaining
        
        details['summary']['earliest_expiry'] = earliest_expiry
        
        return details
        
    except Exception as e:
        logger.error(f"获取证书详细信息失败: {str(e)}")
        return {'error': f'获取证书详细信息失败: {str(e)}'}
