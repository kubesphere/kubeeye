#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 动态客户端工具 - 类似Go的Dynamic Client
使用kubernetes.dynamic模块和资源映射表来简化资源操作
"""

import yaml
import tempfile
import os
import json
import base64
import datetime
from typing import Dict, List, Any, Optional, Tuple
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from kubernetes.dynamic import DynamicClient
from kubernetes.dynamic.exceptions import ResourceNotFoundError
import logging

logger = logging.getLogger(__name__)

class K8sDynamicClient:
    """
    Kubernetes 动态客户端类 - 模仿Go的Dynamic Client
    支持动态资源发现和操作，减少重复代码
    """
    
    def __init__(self, kubeconfig_content: str = None):
        """
        初始化 Kubernetes 动态客户端
        
        Args:
            kubeconfig_content: kubeconfig 文件内容
        """
        self.kubeconfig_content = kubeconfig_content
        self.temp_config = None
        self.initialized = False
        self.dynamic_client = None
        self.api_client = None
        self.resource_cache = {}  # 资源定义缓存
        self.resource_mappings = {}  # 动态资源映射表，从规则配置中构建
        
        self.init_client()
    
    def init_client(self) -> bool:
        """
        初始化 Kubernetes 动态客户端
        
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
            
            # 创建API客户端和动态客户端
            self.api_client = client.ApiClient()
            self.dynamic_client = DynamicClient(self.api_client)
            
            # 禁用SSL证书验证（如果需要）
            configuration = client.Configuration.get_default_copy()
            configuration.verify_ssl = False
            client.Configuration.set_default(configuration)
            
            self.initialized = True
            logger.info("Kubernetes动态客户端初始化成功")
            return True
            
        except Exception as e:
            logger.error(f"初始化 Kubernetes 动态客户端失败: {e}")
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
    
    def get_resource_definition(self, resource_type: str) -> Optional[Any]:
        """
        获取资源定义 - 类似Go Dynamic Client的Resource()方法
        
        Args:
            resource_type: 资源类型，如'pods', 'deployments'等
            
        Returns:
            资源定义对象或None
        """
        if not self.initialized:
            return None
        
        # 检查缓存
        if resource_type in self.resource_cache:
            return self.resource_cache[resource_type]
        
        try:
            # 从映射表获取资源信息
            if resource_type in self.resource_mappings:
                mapping = self.resource_mappings[resource_type]
                api_version = f"{mapping['group']}/{mapping['version']}" if mapping['group'] else mapping['version']
                
                # 使用动态客户端获取资源定义
                resource = self.dynamic_client.resources.get(
                    api_version=api_version,
                    kind=mapping['plural'].capitalize().rstrip('s')  # 转换为单数形式的Kind
                )
                
                self.resource_cache[resource_type] = resource
                return resource
            
            else:
                # 尝试自动发现资源
                resource = self._discover_resource(resource_type)
                if resource:
                    self.resource_cache[resource_type] = resource
                return resource
                
        except Exception as e:
            logger.debug(f"获取资源定义失败 {resource_type}: {str(e)}")
            return None
    
    def _discover_resource(self, resource_type: str) -> Optional[Any]:
        """
        自动发现资源定义
        """
        try:
            # 尝试在所有API组中查找资源
            for resource in self.dynamic_client.resources.search(name=resource_type):
                return resource
            return None
        except Exception as e:
            logger.debug(f"自动发现资源失败 {resource_type}: {str(e)}")
            return None
    
    def list_resources(self, resource_type: str, namespace: str = None) -> List[Dict]:
        """
        统一的资源列表方法 - 类似Go Dynamic Client的List()方法
        
        Args:
            resource_type: 资源类型
            namespace: 命名空间（可选）
            
        Returns:
            资源对象列表
        """
        if not self.initialized:
            logger.error("客户端未初始化")
            return []
        
        try:
            # 获取资源定义
            resource = self.get_resource_definition(resource_type)
            if not resource:
                logger.warning(f"无法获取资源定义: {resource_type}")
                return []
            
            # 检查资源是否是命名空间级别的
            mapping = self.resource_mappings.get(resource_type, {})
            is_namespaced = mapping.get('namespaced', True)
            
            # 列出资源
            if is_namespaced and namespace:
                # 列出指定命名空间的资源
                result = resource.get(namespace=namespace)
            elif is_namespaced:
                # 列出所有命名空间的资源
                result = resource.get()
            else:
                # 列出集群级别资源
                result = resource.get()
            
            # 转换为字典格式
            items = result.items if hasattr(result, 'items') else [result]
            return [self._convert_dynamic_object_to_dict(item) for item in items]
            
        except ResourceNotFoundError:
            logger.warning(f"资源类型不存在: {resource_type}")
            return []
        except Exception as e:
            logger.error(f"列出资源失败 {resource_type}: {str(e)}")
            return []
    
    def get_resource(self, resource_type: str, name: str, namespace: str = None) -> Optional[Dict]:
        """
        获取单个资源 - 类似Go Dynamic Client的Get()方法
        
        Args:
            resource_type: 资源类型
            name: 资源名称
            namespace: 命名空间（可选）
            
        Returns:
            资源对象或None
        """
        if not self.initialized:
            return None
        
        try:
            resource = self.get_resource_definition(resource_type)
            if not resource:
                return None
            
            # 获取资源
            if namespace:
                result = resource.get(name=name, namespace=namespace)
            else:
                result = resource.get(name=name)
            
            return self._convert_dynamic_object_to_dict(result)
            
        except Exception as e:
            logger.debug(f"获取资源失败 {resource_type}/{name}: {str(e)}")
            return None
    
    def list_all_resources(self, resource_types: List[str]) -> Dict[str, List[Dict]]:
        """
        批量获取多种资源类型 - 优化版本
        
        Args:
            resource_types: 资源类型列表
            
        Returns:
            按类型分组的资源字典
        """
        results = {}
        
        for resource_type in resource_types:
            try:
                resources = self.list_resources(resource_type)
                results[resource_type] = resources
                logger.debug(f"获取 {resource_type}: {len(resources)} 个资源")
            except Exception as e:
                logger.warning(f"获取资源类型 {resource_type} 失败: {str(e)}")
                results[resource_type] = []
        
        return results
    
    def discover_all_crds(self) -> List[Dict]:
        """
        发现集群中的所有CRD - 动态发现
        
        Returns:
            CRD定义列表
        """
        try:
            # 获取CRD资源定义
            crd_resource = self.get_resource_definition('customresourcedefinitions')
            if not crd_resource:
                logger.warning("无法获取CRD资源定义")
                return []
            
            crds = crd_resource.get()
            crd_list = []
            
            for crd in crds.items:
                crd_info = {
                    'name': crd.metadata.name,
                    'group': crd.spec.group,
                    'versions': [v.name for v in crd.spec.versions],
                    'scope': crd.spec.scope,
                    'kind': crd.spec.names.kind,
                    'plural': crd.spec.names.plural,
                }
                crd_list.append(crd_info)
                
                # 动态添加到资源映射表
                latest_version = crd.spec.versions[0].name if crd.spec.versions else 'v1'
                self.resource_mappings[crd.spec.names.plural] = {
                    'group': crd.spec.group,
                    'version': latest_version,
                    'plural': crd.spec.names.plural,
                    'namespaced': crd.spec.scope == 'Namespaced'
                }
            
            logger.info(f"发现 {len(crd_list)} 个CRD，已添加到资源映射表")
            return crd_list
            
        except Exception as e:
            logger.error(f"发现CRD失败: {str(e)}")
            return []
    
    def list_all_crd_resources(self) -> Dict[str, List[Dict]]:
        """
        获取所有CRD资源实例
        
        Returns:
            按CRD类型分组的资源字典
        """
        crd_resources = {}
        
        # 首先发现所有CRD
        crds = self.discover_all_crds()
        
        for crd in crds:
            try:
                plural = crd['plural']
                resources = self.list_resources(plural)
                crd_resources[plural] = resources
                logger.debug(f"获取CRD {plural}: {len(resources)} 个实例")
            except Exception as e:
                logger.warning(f"获取CRD资源 {crd['plural']} 失败: {str(e)}")
                crd_resources[crd['plural']] = []
        
        return crd_resources
    
    def _convert_dynamic_object_to_dict(self, obj) -> Dict:
        """
        将动态对象转换为字典格式
        
        Args:
            obj: 动态对象
            
        Returns:
            字典格式的对象
        """
        try:
            if hasattr(obj, 'to_dict'):
                result = obj.to_dict()
            elif hasattr(obj, '__dict__'):
                result = dict(obj.__dict__)
            else:
                result = dict(obj)
            
            # 处理datetime对象
            self._convert_datetime_in_dict(result)
            return result
            
        except Exception as e:
            logger.warning(f"转换动态对象失败: {str(e)}")
            return {}
    
    def _convert_datetime_in_dict(self, data: Any) -> None:
        """
        递归转换字典中的datetime对象为字符串
        """
        if isinstance(data, dict):
            for key, value in data.items():
                if isinstance(value, datetime.datetime):
                    data[key] = value.isoformat()
                elif isinstance(value, datetime.date):
                    data[key] = value.isoformat()
                elif isinstance(value, (dict, list)):
                    self._convert_datetime_in_dict(value)
        elif isinstance(data, list):
            for item in data:
                if isinstance(item, (dict, list)):
                    self._convert_datetime_in_dict(item)
    
    def get_supported_resource_types(self) -> List[str]:
        """
        获取支持的资源类型列表
        
        Returns:
            支持的资源类型列表
        """
        return list(self.resource_mappings.keys())
    
    def is_namespaced_resource(self, resource_type: str) -> bool:
        """
        检查资源是否是命名空间级别的
        
        Args:
            resource_type: 资源类型
            
        Returns:
            是否为命名空间级别资源
        """
        mapping = self.resource_mappings.get(resource_type, {})
        return mapping.get('namespaced', True)
    
    def test_connection(self) -> Tuple[bool, str]:
        """
        测试与集群的连接
        
        Returns:
            (成功, 消息) 元组
        """
        if not self.initialized:
            return False, "动态客户端未初始化"
        
        try:
            # 尝试列出命名空间来测试连接
            namespaces = self.list_resources('namespaces')
            return True, f"连接成功，发现 {len(namespaces)} 个命名空间"
        except Exception as e:
            return False, f"连接测试失败: {str(e)}"
    
    def build_resource_mappings_from_config(self, rule_config: Dict) -> None:
        """
        从规则配置中构建资源映射表
        
        Args:
            rule_config: 规则配置字典，包含 resources 和 scope 信息
        """
        # rule_config 是 rule.config，所以 resources 直接在根级别
        resources_config = rule_config.get('resources', [])
        
        for resource in resources_config:
            kind = resource.get('kind', '')
            api_version = resource.get('apiVersion', '')
            namespaced = resource.get('namespaced', True)
            
            # 解析 apiVersion 获取 group 和 version
            if '/' in api_version:
                group, version = api_version.split('/', 1)
            else:
                group = ''
                version = api_version
            
            # 生成复数形式的资源名称（简单转换）
            plural = self._kind_to_plural(kind)
            
            # 添加到资源映射表
            self.resource_mappings[plural.lower()] = {
                'group': group,
                'version': version, 
                'plural': plural.lower(),
                'namespaced': namespaced,
                'kind': kind
            }
            
        logger.info(f"从规则配置构建了 {len(self.resource_mappings)} 个资源映射")
    
    def _kind_to_plural(self, kind: str) -> str:
        """
        将 Kind 转换为复数形式的资源名称
        
        Args:
            kind: Kubernetes 资源的 Kind
            
        Returns:
            复数形式的资源名称
        """
        # 常见的单数转复数规则
        kind_lower = kind.lower()
        
        # 特殊情况
        special_cases = {
            'networkpolicy': 'networkpolicies',
            'ingress': 'ingresses',
            'horizontalpodautoscaler': 'horizontalpodautoscalers',
            'poddisruptionbudget': 'poddisruptionbudgets',
            'priorityclass': 'priorityclasses',
            'storageclass': 'storageclasses',
            'volumeattachment': 'volumeattachments',
            'customresourcedefinition': 'customresourcedefinitions',
        }
        
        if kind_lower in special_cases:
            return special_cases[kind_lower]
        
        # 一般规则
        if kind_lower.endswith('y'):
            return kind_lower[:-1] + 'ies'
        elif kind_lower.endswith(('s', 'sh', 'ch', 'x', 'z')):
            return kind_lower + 'es'
        else:
            return kind_lower + 's'
    
    def list_resources_from_config(self, rule_config: Dict) -> Dict[str, List[Dict]]:
        """
        根据规则配置获取所需的资源
        
        Args:
            rule_config: 规则配置字典
            
        Returns:
            按资源类型分组的资源字典
        """
        # 构建资源映射
        self.build_resource_mappings_from_config(rule_config)
        
        # 获取命名空间配置
        scope_config = rule_config.get('scope', {})
        namespace_config = scope_config.get('namespaces', {})
        include_namespaces = namespace_config.get('include', [])
        exclude_namespaces = namespace_config.get('exclude', [])
        
        resources = {}
        
        # 获取所有配置的资源类型
        for resource_type in self.resource_mappings.keys():
            try:
                mapping = self.resource_mappings[resource_type]
                is_namespaced = mapping.get('namespaced', True)
                
                if is_namespaced:
                    # 处理命名空间级别的资源
                    if include_namespaces:
                        # 只获取指定命名空间的资源
                        all_resources = []
                        for namespace in include_namespaces:
                            if namespace not in exclude_namespaces:
                                ns_resources = self.list_resources(resource_type, namespace=namespace)
                                all_resources.extend(ns_resources)
                        resources[resource_type] = all_resources
                    else:
                        # 获取所有命名空间的资源，然后过滤
                        all_resources = self.list_resources(resource_type)
                        if exclude_namespaces:
                            filtered_resources = [
                                res for res in all_resources 
                                if res.get('metadata', {}).get('namespace') not in exclude_namespaces
                            ]
                            resources[resource_type] = filtered_resources
                        else:
                            resources[resource_type] = all_resources
                else:
                    # 集群级别资源，直接获取
                    resources[resource_type] = self.list_resources(resource_type)
                    
                logger.debug(f"获取 {resource_type}: {len(resources[resource_type])} 个资源")
                
            except Exception as e:
                logger.warning(f"获取资源类型 {resource_type} 失败: {str(e)}")
                resources[resource_type] = []
        
        return resources
    
    def _analyze_resource_strategy(self, resource_config: List[Dict]) -> Dict[str, Any]:
        """
        分析资源配置，选择最优的获取策略
        
        Args:
            resource_config: 资源配置列表
            
        Returns:
            策略分析结果
        """
        configured_kinds = {res['kind'] for res in resource_config}
        
        # 定义控制器类型
        workload_controllers = {'Deployment', 'StatefulSet', 'DaemonSet'}
        job_controllers = {'Job', 'CronJob'}
        all_controllers = workload_controllers | job_controllers
        
        # 分析配置
        has_pods = 'Pod' in configured_kinds
        has_controllers = bool(configured_kinds & all_controllers)
        configured_controllers = configured_kinds & all_controllers
        
        strategy = {
            'mode': 'default',
            'get_pods_directly': has_pods,
            'get_controllers': list(configured_controllers),
            'filter_controlled_pods': False,
            'optimization_applied': False,
            'reasoning': 'Default strategy - get all configured resources'
        }
        
        # 如果同时配置了控制器和Pod，应用优化策略
        if has_pods and has_controllers:
            if len(configured_controllers) >= 2:
                # 多个控制器 + Pod：推荐仅控制器策略
                strategy.update({
                    'mode': 'controller_only',
                    'get_pods_directly': False,
                    'filter_controlled_pods': False,
                    'optimization_applied': True,
                    'reasoning': '多个控制器+Pod配置，采用仅控制器策略避免重复'
                })
            else:
                # 单个控制器 + Pod：推荐智能混合策略
                strategy.update({
                    'mode': 'hybrid_smart',
                    'get_pods_directly': True,
                    'filter_controlled_pods': True,
                    'optimization_applied': True,
                    'reasoning': '控制器+Pod配置，采用智能混合策略过滤受控Pod'
                })
        
        return strategy
    
    def _filter_controlled_pods(self, pods: List[Dict], controllers: Dict[str, List[Dict]]) -> List[Dict]:
        """
        过滤掉被控制器管理的Pod，只保留独立Pod
        
        Args:
            pods: Pod资源列表
            controllers: 控制器资源字典
            
        Returns:
            过滤后的独立Pod列表
        """
        if not pods:
            return []
        
        # 收集所有控制器的UID
        controller_uids = set()
        
        for controller_type, controller_list in controllers.items():
            for controller in controller_list:
                uid = controller.get('metadata', {}).get('uid')
                if uid:
                    controller_uids.add(uid)
        
        # 过滤Pod
        independent_pods = []
        controlled_pods = []
        
        for pod in pods:
            owner_refs = pod.get('metadata', {}).get('ownerReferences', [])
            is_controlled = False
            
            for owner_ref in owner_refs:
                if owner_ref.get('uid') in controller_uids:
                    is_controlled = True
                    controlled_pods.append(pod)
                    break
            
            if not is_controlled:
                independent_pods.append(pod)
        
        logger.info(f"Pod过滤结果: 总数={len(pods)}, 受控={len(controlled_pods)}, 独立={len(independent_pods)}")
        return independent_pods

    def list_resources_from_config_optimized(self, rule_config: Dict) -> Dict[str, List[Dict]]:
        """
        根据规则配置获取资源 - 优化版本，避免重复获取
        
        Args:
            rule_config: 规则配置字典
            
        Returns:
            按资源类型分组的资源字典
        """
        # 构建资源映射
        self.build_resource_mappings_from_config(rule_config)
        
        # 分析资源获取策略
        resource_config = rule_config.get('config', {}).get('resources', [])
        strategy = self._analyze_resource_strategy(resource_config)
        
        logger.info(f"资源获取策略: {strategy['mode']} - {strategy['reasoning']}")
        
        # 获取命名空间配置
        scope_config = rule_config.get('scope', {})
        namespace_config = scope_config.get('namespaces', {})
        include_namespaces = namespace_config.get('include', [])
        exclude_namespaces = namespace_config.get('exclude', [])
        
        resources = {}
        
        # 根据策略获取资源
        if strategy['mode'] == 'controller_only':
            # 仅获取控制器资源
            for resource_type in strategy['get_controllers']:
                resource_type_lower = resource_type.lower() + 's'
                if resource_type_lower in self.resource_mappings:
                    resources[resource_type_lower] = self._get_namespaced_resources(
                        resource_type_lower, include_namespaces, exclude_namespaces
                    )
        
        elif strategy['mode'] == 'hybrid_smart':
            # 智能混合模式：获取控制器 + 过滤独立Pod
            controllers = {}
            
            # 获取控制器
            for resource_type in strategy['get_controllers']:
                resource_type_lower = resource_type.lower() + 's'
                if resource_type_lower in self.resource_mappings:
                    controller_resources = self._get_namespaced_resources(
                        resource_type_lower, include_namespaces, exclude_namespaces
                    )
                    resources[resource_type_lower] = controller_resources
                    controllers[resource_type_lower] = controller_resources
            
            # 获取Pod并过滤
            if strategy['get_pods_directly'] and 'pods' in self.resource_mappings:
                all_pods = self._get_namespaced_resources(
                    'pods', include_namespaces, exclude_namespaces
                )
                
                if strategy['filter_controlled_pods']:
                    # 过滤掉受控制器管理的Pod
                    independent_pods = self._filter_controlled_pods(all_pods, controllers)
                    resources['pods'] = independent_pods
                else:
                    resources['pods'] = all_pods
        
        else:
            # 默认策略：获取所有配置的资源（原有逻辑）
            resources = self.list_resources_from_config(rule_config)
        
        # 记录结果
        total_resources = sum(len(res_list) for res_list in resources.values())
        logger.info(f"优化后获取 {len(resources)} 种资源类型，共 {total_resources} 个资源")
        
        if strategy['optimization_applied']:
            logger.info("✅ 已应用资源获取优化，避免重复获取")
        
        return resources
    
    def _get_namespaced_resources(self, resource_type: str, include_namespaces: List[str], 
                                 exclude_namespaces: List[str]) -> List[Dict]:
        """
        获取命名空间级别的资源（辅助方法）
        
        Args:
            resource_type: 资源类型
            include_namespaces: 包含的命名空间
            exclude_namespaces: 排除的命名空间
            
        Returns:
            资源列表
        """
        try:
            mapping = self.resource_mappings[resource_type]
            is_namespaced = mapping.get('namespaced', True)
            
            if not is_namespaced:
                # 集群级别资源
                return self.list_resources(resource_type)
            
            if include_namespaces:
                # 只获取指定命名空间的资源
                all_resources = []
                for namespace in include_namespaces:
                    if namespace not in exclude_namespaces:
                        namespace_resources = self.list_resources(resource_type, namespace)
                        all_resources.extend(namespace_resources)
                return all_resources
            else:
                # 获取所有命名空间的资源，然后过滤
                all_resources = self.list_resources(resource_type)
                if exclude_namespaces:
                    filtered_resources = [
                        resource for resource in all_resources
                        if resource.get('metadata', {}).get('namespace') not in exclude_namespaces
                    ]
                    return filtered_resources
                return all_resources
                
        except Exception as e:
            logger.error(f"获取资源 {resource_type} 失败: {str(e)}")
            return []

