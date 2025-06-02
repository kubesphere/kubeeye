#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 客户端工具，用于与 Kubernetes 集群交互
"""

import yaml
import tempfile
import os
import json
import base64
from typing import Dict, List, Any, Optional, Tuple
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import logging

logger = logging.getLogger(__name__)

class K8sClient:
    """Kubernetes 客户端类"""
    
    def __init__(self, kubeconfig_content: str = None):
        """
        初始化 Kubernetes 客户端
        
        Args:
            kubeconfig_content: kubeconfig 文件内容
        """
        self.kubeconfig_content = kubeconfig_content
        self.temp_config = None
        self.initialized = False
        self.init_client()
    
    def init_client(self) -> bool:
        """
        初始化 Kubernetes 客户端配置
        
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
                
            self.core_v1 = client.CoreV1Api()
            self.apps_v1 = client.AppsV1Api()
            self.batch_v1 = client.BatchV1Api()
            self.networking_v1 = client.NetworkingV1Api()
            self.storage_v1 = client.StorageV1Api()
            self.rbac_v1 = client.RbacAuthorizationV1Api()
            self.custom_objects = client.CustomObjectsApi()
            
            self.initialized = True
            return True
        except Exception as e:
            logger.error(f"初始化 Kubernetes 客户端失败: {e}")
            self.initialized = False
            return False
    
    def __del__(self):
        """析构函数，删除临时文件"""
        if self.temp_config:
            try:
                self.temp_config.close()
                os.unlink(self.temp_config.name)
            except:
                pass
    
    def test_connection(self) -> Tuple[bool, str]:
        """
        测试与集群的连接
        
        Returns:
            (成功, 消息) 元组
        """
        if not self.initialized:
            return False, "客户端未初始化"
        
        try:
            version = self.core_v1.get_api_resources()
            return True, "连接成功"
        except ApiException as e:
            return False, f"API 错误: {e.reason}"
        except Exception as e:
            return False, f"连接错误: {str(e)}"
    
    def get_nodes(self) -> Dict:
        """
        获取集群节点列表
        
        Returns:
            节点信息字典
        """
        if not self.initialized:
            return {'status': 'error', 'error': '客户端未初始化'}
        
        try:
            nodes = self.core_v1.list_node()
            result = []
            
            for node in nodes.items:
                node_info = {
                    'name': node.metadata.name,
                    'status': self._get_node_status(node),
                    'roles': self._get_node_roles(node),
                    'kubelet_version': node.status.node_info.kubelet_version,
                    'os_image': node.status.node_info.os_image,
                    'kernel_version': node.status.node_info.kernel_version,
                    'cpu': node.status.capacity.get('cpu', 'N/A'),
                    'memory': node.status.capacity.get('memory', 'N/A'),
                    'pods': node.status.capacity.get('pods', 'N/A'),
                    'conditions': self._get_node_conditions(node),
                    'taints': self._get_node_taints(node),
                    'creation_timestamp': node.metadata.creation_timestamp,
                    'labels': node.metadata.labels,
                }
                result.append(node_info)
                
            return {'status': 'success', 'nodes': result}
        except ApiException as e:
            return {'status': 'error', 'error': f"API 错误: {e.reason}"}
        except Exception as e:
            return {'status': 'error', 'error': f"获取节点错误: {str(e)}"}
    
    def _get_node_status(self, node) -> str:
        """获取节点状态"""
        for condition in node.status.conditions:
            if condition.type == 'Ready':
                return 'Ready' if condition.status == 'True' else 'NotReady'
        return 'Unknown'
    
    def _get_node_roles(self, node) -> List[str]:
        """获取节点角色"""
        roles = []
        labels = node.metadata.labels or {}
        
        for label in labels:
            if label.startswith('node-role.kubernetes.io/'):
                role = label.split('/')[-1]
                roles.append(role)
        
        if not roles:
            roles.append('worker')
            
        return roles
    
    def _get_node_conditions(self, node) -> List[Dict]:
        """获取节点状态条件"""
        conditions = []
        
        for condition in node.status.conditions:
            conditions.append({
                'type': condition.type,
                'status': condition.status,
                'reason': condition.reason,
                'message': condition.message,
                'last_transition_time': condition.last_transition_time
            })
            
        return conditions
    
    def _get_node_taints(self, node) -> List[Dict]:
        """获取节点污点"""
        taints = []
        
        if node.spec.taints:
            for taint in node.spec.taints:
                taints.append({
                    'key': taint.key,
                    'value': taint.value,
                    'effect': taint.effect
                })
                
        return taints
    
    def get_pods(self, namespace: str = None) -> Dict:
        """
        获取 Pod 列表
        
        Args:
            namespace: 命名空间
            
        Returns:
            Pod 列表字典
        """
        if not self.initialized:
            return {'status': 'error', 'error': '客户端未初始化'}
        
        try:
            if namespace:
                pods = self.core_v1.list_namespaced_pod(namespace)
            else:
                pods = self.core_v1.list_pod_for_all_namespaces()
                
            result = []
            
            for pod in pods.items:
                pod_info = {
                    'name': pod.metadata.name,
                    'namespace': pod.metadata.namespace,
                    'status': pod.status.phase,
                    'node': pod.spec.node_name,
                    'ip': pod.status.pod_ip,
                    'creation_timestamp': pod.metadata.creation_timestamp,
                    'containers': [c.name for c in pod.spec.containers],
                    'restart_count': sum(c.restart_count for c in pod.status.container_statuses if c.restart_count) if pod.status.container_statuses else 0,
                }
                result.append(pod_info)
                
            return {'status': 'success', 'pods': result}
        except ApiException as e:
            return {'status': 'error', 'error': f"API 错误: {e.reason}"}
        except Exception as e:
            return {'status': 'error', 'error': f"获取 Pod 错误: {str(e)}"}
    
    def list_pods_all_namespaces(self):
        """获取所有命名空间下的Pod列表"""
        try:
            pods = self.core_v1.list_pod_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(pod) for pod in pods.items]
        except Exception as e:
            logger.error(f"获取所有Pod失败: {e}")
            return []
    
    def list_deployments_all_namespaces(self):
        """获取所有命名空间下的Deployment列表"""
        try:
            deployments = self.apps_v1.list_deployment_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(d) for d in deployments.items]
        except Exception as e:
            logger.error(f"获取所有Deployment失败: {e}")
            return []
    
    def list_statefulsets_all_namespaces(self):
        """获取所有命名空间下的StatefulSet列表"""
        try:
            statefulsets = self.apps_v1.list_stateful_set_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(s) for s in statefulsets.items]
        except Exception as e:
            logger.error(f"获取所有StatefulSet失败: {e}")
            return []
    
    def list_daemonsets_all_namespaces(self):
        """获取所有命名空间下的DaemonSet列表"""
        try:
            daemonsets = self.apps_v1.list_daemon_set_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(d) for d in daemonsets.items]
        except Exception as e:
            logger.error(f"获取所有DaemonSet失败: {e}")
            return []
    
    def list_services_all_namespaces(self):
        """获取所有命名空间下的Service列表"""
        try:
            services = self.core_v1.list_service_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(s) for s in services.items]
        except Exception as e:
            logger.error(f"获取所有Service失败: {e}")
            return []
    
    def list_ingresses_all_namespaces(self):
        """获取所有命名空间下的Ingress列表"""
        try:
            ingresses = self.networking_v1.list_ingress_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(i) for i in ingresses.items]
        except Exception as e:
            logger.error(f"获取所有Ingress失败: {e}")
            return []
    
    def list_serviceaccounts_all_namespaces(self):
        """获取所有命名空间下的ServiceAccount列表"""
        try:
            serviceaccounts = self.core_v1.list_service_account_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(sa) for sa in serviceaccounts.items]
        except Exception as e:
            logger.error(f"获取所有ServiceAccount失败: {e}")
            return []
    
    def list_configmaps_all_namespaces(self):
        """获取所有命名空间下的ConfigMap列表"""
        try:
            configmaps = self.core_v1.list_config_map_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(cm) for cm in configmaps.items]
        except Exception as e:
            logger.error(f"获取所有ConfigMap失败: {e}")
            return []
    
    def list_secrets_all_namespaces(self):
        """获取所有命名空间下的Secret列表"""
        try:
            secrets = self.core_v1.list_secret_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(s) for s in secrets.items]
        except Exception as e:
            logger.error(f"获取所有Secret失败: {e}")
            return []
    
    def list_persistent_volume_claims_all_namespaces(self):
        """获取所有命名空间下的PVC列表"""
        try:
            pvcs = self.core_v1.list_persistent_volume_claim_for_all_namespaces()
            return [self._convert_k8s_object_to_dict(pvc) for pvc in pvcs.items]
        except Exception as e:
            logger.error(f"获取所有PVC失败: {e}")
            return []
    
    def list_namespaces(self):
        """获取所有命名空间列表"""
        try:
            namespaces = self.core_v1.list_namespace()
            return [self._convert_k8s_object_to_dict(ns) for ns in namespaces.items]
        except Exception as e:
            logger.error(f"获取所有命名空间失败: {e}")
            return []
    
    def _convert_k8s_object_to_dict(self, obj):
        """将Kubernetes对象转换为字典格式"""
        try:
            # 使用to_dict()方法，如果对象支持
            if hasattr(obj, 'to_dict'):
                return obj.to_dict()
            else:
                # 使用kubernetes客户端的序列化方法，确保处理datetime对象
                json_str = client.ApiClient().sanitize_for_serialization(obj)
                
                # 确保所有datetime被转换为字符串
                if isinstance(json_str, dict):
                    self._convert_datetime_in_dict(json_str)
                    
                return json_str
        except Exception as e:
            logger.warning(f"转换Kubernetes对象失败: {e}")
            # 返回最小对象
            if hasattr(obj, 'metadata') and hasattr(obj.metadata, 'name'):
                return {
                    'metadata': {
                        'name': obj.metadata.name,
                        'namespace': obj.metadata.namespace if hasattr(obj.metadata, 'namespace') else None
                    }
                }
            return {}
