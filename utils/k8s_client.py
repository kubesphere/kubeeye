#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 客户端工具，用于与 Kubernetes 集群交互
"""

import yaml
import json
import base64
import datetime
from typing import Dict, List, Any, Optional, Tuple
from kubernetes import client
from kubernetes.client.rest import ApiException
import logging

from .k8s_base_client import K8sBaseClient

logger = logging.getLogger(__name__)

class K8sClient(K8sBaseClient):
    """Kubernetes 客户端类"""
    
    def __init__(self, kubeconfig_content: str = None):
        """
        初始化 Kubernetes 客户端
        
        Args:
            kubeconfig_content: kubeconfig 文件内容
        """
        super().__init__(kubeconfig_content)
        self.init_client()
    
    def init_client(self) -> bool:
        """
        初始化 Kubernetes 客户端配置
        
        Returns:
            成功返回 True，失败返回 False
        """
        try:
            # 使用基类的通用初始化逻辑
            if not self.init_client_base():
                return False
                
            # 初始化具体的API客户端
            self.core_v1 = client.CoreV1Api()
            self.apps_v1 = client.AppsV1Api()
            self.batch_v1 = client.BatchV1Api()
            self.networking_v1 = client.NetworkingV1Api()
            self.storage_v1 = client.StorageV1Api()
            self.rbac_v1 = client.RbacAuthorizationV1Api()
            self.custom_objects = client.CustomObjectsApi()
            
            logger.info("Kubernetes客户端API初始化成功")
            return True
        except Exception as e:
            logger.error(f"初始化 Kubernetes 客户端失败: {e}")
            self.initialized = False
            return False
    
    def _configure_no_verify_ssl(self):
        """已移至基类K8sBaseClient"""
        return super()._configure_no_verify_ssl()
    
    def __del__(self):
        """已移至基类K8sBaseClient"""
        super().__del__()
    
    def test_connection(self) -> Tuple[bool, str]:
        """已移至基类K8sBaseClient"""
        return super().test_connection()
    
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
    
    def list_resources(self, resource_type: str):
        """
        获取所有命名空间下指定类型的资源列表 - 增强版本，支持更多资源类型
        
        Args:
            resource_type: 资源类型，如'pods', 'deployments', 'services'等
            
        Returns:
            资源对象列表
        """
        try:
            if resource_type == 'pods':
                items = self.core_v1.list_pod_for_all_namespaces().items
            elif resource_type == 'deployments':
                items = self.apps_v1.list_deployment_for_all_namespaces().items
            elif resource_type == 'services':
                items = self.core_v1.list_service_for_all_namespaces().items
            elif resource_type == 'statefulsets':
                items = self.apps_v1.list_stateful_set_for_all_namespaces().items
            elif resource_type == 'daemonsets':
                items = self.apps_v1.list_daemon_set_for_all_namespaces().items
            elif resource_type == 'replicasets':
                items = self.apps_v1.list_replica_set_for_all_namespaces().items
            elif resource_type == 'configmaps':
                items = self.core_v1.list_config_map_for_all_namespaces().items
            elif resource_type == 'secrets':
                items = self.core_v1.list_secret_for_all_namespaces().items
            elif resource_type == 'persistentvolumeclaims':
                items = self.core_v1.list_persistent_volume_claim_for_all_namespaces().items
            elif resource_type == 'serviceaccounts':
                items = self.core_v1.list_service_account_for_all_namespaces().items
            elif resource_type == 'networkpolicies':
                items = self.networking_v1.list_network_policy_for_all_namespaces().items
            elif resource_type == 'roles':
                items = self.rbac_v1.list_role_for_all_namespaces().items
            elif resource_type == 'rolebindings':
                items = self.rbac_v1.list_role_binding_for_all_namespaces().items
            elif resource_type == 'ingresses':
                items = self.networking_v1.list_ingress_for_all_namespaces().items
            else:
                # 尝试作为CRD资源处理
                items = self._list_custom_resources(resource_type)
                if items is None:
                    logger.warning(f"不支持的资源类型: {resource_type}")
                    return []
                
            return [self._convert_k8s_object_to_dict(item) for item in items]
        except Exception as e:
            logger.error(f"获取资源 {resource_type} 失败: {e}")
            return []
    
    def list_cluster_resources(self, resource_type: str):
        """
        获取集群级别的资源列表 - 增强版本，支持更多资源类型
        
        Args:
            resource_type: 资源类型，如'nodes', 'persistentvolumes'等
            
        Returns:
            资源对象列表
        """
        try:
            if resource_type == 'nodes':
                items = self.core_v1.list_node().items
            elif resource_type == 'persistentvolumes':
                items = self.core_v1.list_persistent_volume().items
            elif resource_type == 'clusterroles':
                items = self.rbac_v1.list_cluster_role().items
            elif resource_type == 'clusterrolebindings':
                items = self.rbac_v1.list_cluster_role_binding().items
            elif resource_type == 'storageclasses':
                items = self.storage_v1.list_storage_class().items
            else:
                # 尝试作为CRD集群资源处理
                items = self._list_custom_cluster_resources(resource_type)
                if items is None:
                    logger.warning(f"不支持的集群资源类型: {resource_type}")
                    return []
                
            return [self._convert_k8s_object_to_dict(item) for item in items]
        except Exception as e:
            logger.error(f"获取集群资源 {resource_type} 失败: {e}")
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
    
    def get_custom_resources(self, group: str, version: str, plural: str, 
                            namespace: str = None) -> List[Dict]:
        """
        获取自定义资源(CRD)列表
        
        Args:
            group: API组，如 'networking.istio.io'
            version: API版本，如 'v1beta1'
            plural: 资源复数名称，如 'virtualservices'
            namespace: 命名空间，如果为None则获取集群级别资源
            
        Returns:
            CRD资源对象列表
        """
        try:
            if namespace:
                # 获取命名空间级别的CRD资源
                response = self.custom_objects.list_namespaced_custom_object(
                    group=group,
                    version=version,
                    namespace=namespace,
                    plural=plural
                )
            else:
                # 获取所有命名空间的CRD资源
                response = self.custom_objects.list_cluster_custom_object(
                    group=group,
                    version=version,
                    plural=plural
                )
            
            items = response.get('items', [])
            return [self._convert_dict_to_k8s_format(item) for item in items]
            
        except Exception as e:
            logger.debug(f"获取CRD资源 {group}/{version}/{plural} 失败: {str(e)}")
            return []

    def _list_custom_resources(self, resource_type: str) -> List:
        """
        尝试列出自定义资源
        
        Args:
            resource_type: 资源类型
            
        Returns:
            资源列表或None（如果不支持）
        """
        # 常见CRD资源映射
        crd_mappings = {
            'virtualservices': ('networking.istio.io', 'v1beta1'),
            'destinationrules': ('networking.istio.io', 'v1beta1'),
            'gateways': ('networking.istio.io', 'v1beta1'),
            'serviceentries': ('networking.istio.io', 'v1beta1'),
            'certificates': ('cert-manager.io', 'v1'),
            'certificaterequests': ('cert-manager.io', 'v1'),
            'issuers': ('cert-manager.io', 'v1'),
            'prometheuses': ('monitoring.coreos.com', 'v1'),
            'servicemonitors': ('monitoring.coreos.com', 'v1'),
            'alertmanagers': ('monitoring.coreos.com', 'v1'),
            'prometheusrules': ('monitoring.coreos.com', 'v1'),
        }
        
        if resource_type in crd_mappings:
            group, version = crd_mappings[resource_type]
            try:
                response = self.custom_objects.list_cluster_custom_object(
                    group=group,
                    version=version,
                    plural=resource_type
                )
                return [self._convert_dict_to_k8s_format(item) for item in response.get('items', [])]
            except Exception as e:
                logger.debug(f"获取CRD资源 {resource_type} 失败: {str(e)}")
                return []
        
        return None

    def _list_custom_cluster_resources(self, resource_type: str) -> List:
        """
        尝试列出集群级别的自定义资源
        
        Args:
            resource_type: 资源类型
            
        Returns:
            资源列表或None（如果不支持）
        """
        # 集群级别CRD资源映射
        cluster_crd_mappings = {
            'clusterissuers': ('cert-manager.io', 'v1'),
            'clusterpolicies': ('kyverno.io', 'v1'),
            'customresourcedefinitions': ('apiextensions.k8s.io', 'v1'),
        }
        
        if resource_type in cluster_crd_mappings:
            group, version = cluster_crd_mappings[resource_type]
            try:
                response = self.custom_objects.list_cluster_custom_object(
                    group=group,
                    version=version,
                    plural=resource_type
                )
                return [self._convert_dict_to_k8s_format(item) for item in response.get('items', [])]
            except Exception as e:
                logger.debug(f"获取集群级CRD资源 {resource_type} 失败: {str(e)}")
                return []
        
        return None

    def _convert_dict_to_k8s_format(self, item: Dict) -> Dict:
        """
        将字典格式的CRD资源转换为标准K8s格式
        
        Args:
            item: CRD资源字典
            
        Returns:
            标准格式的资源字典
        """
        # CRD资源已经是字典格式，只需要确保格式一致
        return item

    def _convert_datetime_in_dict(self, data: Dict) -> None:
        """
        递归转换字典中的datetime对象为字符串
        
        Args:
            data: 要转换的字典
        """
        for key, value in data.items():
            if isinstance(value, datetime.datetime):
                data[key] = value.isoformat()
            elif isinstance(value, datetime.date):
                data[key] = value.isoformat()
            elif isinstance(value, dict):
                self._convert_datetime_in_dict(value)
            elif isinstance(value, list):
                for item in value:
                    if isinstance(item, dict):
                        self._convert_datetime_in_dict(item)

    def list_all_custom_resource_definitions(self) -> List[Dict]:
        """
        获取集群中所有的CRD定义
        
        Returns:
            CRD定义列表
        """
        try:
            # 获取集群中所有的CRD
            from kubernetes.client import ApiextensionsV1Api
            extensions_v1 = ApiextensionsV1Api()
            crds = extensions_v1.list_custom_resource_definition()
            
            crd_list = []
            for crd in crds.items:
                crd_info = {
                    'name': crd.metadata.name,
                    'group': crd.spec.group,
                    'versions': [v.name for v in crd.spec.versions],
                    'scope': crd.spec.scope,  # Cluster 或 Namespaced
                    'kind': crd.spec.names.kind,
                    'plural': crd.spec.names.plural,
                }
                crd_list.append(crd_info)
            
            return crd_list
        except Exception as e:
            logger.debug(f"获取CRD列表失败: {str(e)}")
            return []

    def get_custom_resource_by_crd(self, crd_info: Dict, namespace: str = None) -> List[Dict]:
        """
        根据CRD定义获取自定义资源
        
        Args:
            crd_info: CRD定义信息
            namespace: 命名空间（如果是命名空间级别的资源）
            
        Returns:
            自定义资源列表
        """
        try:
            group = crd_info['group']
            # 使用最新版本
            version = crd_info['versions'][0] if crd_info['versions'] else 'v1'
            plural = crd_info['plural']
            scope = crd_info.get('scope', 'Namespaced')
            
            if scope == 'Namespaced' and namespace:
                response = self.custom_objects.list_namespaced_custom_object(
                    group=group,
                    version=version,
                    namespace=namespace,
                    plural=plural
                )
            elif scope == 'Cluster':
                response = self.custom_objects.list_cluster_custom_object(
                    group=group,
                    version=version,
                    plural=plural
                )
            else:
                # 获取所有命名空间的资源
                response = self.custom_objects.list_cluster_custom_object(
                    group=group,
                    version=version,
                    plural=plural
                )
            
            items = response.get('items', [])
            return [self._convert_dict_to_k8s_format(item) for item in items]
            
        except Exception as e:
            logger.debug(f"获取CRD资源失败: {str(e)}")
            return []

    def discover_and_list_all_crd_resources(self) -> Dict[str, List[Dict]]:
        """
        发现并列出所有CRD资源
        
        Returns:
            按类型分组的CRD资源字典
        """
        crd_resources = {}
        
        # 获取所有CRD定义
        crds = self.list_all_custom_resource_definitions()
        
        for crd in crds:
            try:
                resource_key = f"{crd['plural']}.{crd['group']}"
                crd_resources[resource_key] = self.get_custom_resource_by_crd(crd)
                
                # 如果是命名空间级别的资源，也获取各个命名空间的资源
                if crd.get('scope') == 'Namespaced':
                    namespaces = self.list_namespaces()
                    all_resources = []
                    for ns in namespaces:
                        ns_name = ns.get('metadata', {}).get('name', '')
                        if ns_name:
                            ns_resources = self.get_custom_resource_by_crd(crd, ns_name)
                            all_resources.extend(ns_resources)
                    crd_resources[resource_key] = all_resources
                    
            except Exception as e:
                logger.debug(f"获取CRD资源 {crd['plural']} 失败: {str(e)}")
                continue
        
        return crd_resources
