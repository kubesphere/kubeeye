#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Kubernetes 动态客户端工具 - 类似Go的Dynamic Client
使用kubernetes.dynamic模块和资源映射表来简化资源操作
"""

import yaml
import json
import base64
import datetime
from typing import Dict, List, Any, Optional, Tuple
from kubernetes import client
from kubernetes.client.rest import ApiException
from kubernetes.dynamic import DynamicClient
from kubernetes.dynamic.exceptions import ResourceNotFoundError
import logging
import tempfile
import os
import subprocess

from .k8s_base_client import K8sBaseClient

logger = logging.getLogger(__name__)

class K8sDynamicClient(K8sBaseClient):
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
        super().__init__(kubeconfig_content)
        self.dynamic_client = None
        self.api_client = None
        self.resource_cache = {}  # 资源定义缓存
        self.resource_mappings = {}  # 动态资源映射表，从规则配置中构建
        
        self.init_client()
    
    def init_client(self) -> bool:
        """
        初始化 Kubernetes 动态客户端，支持自定义 kubeconfig_content
        
        Returns:
            成功返回 True，失败返回 False
        """
        try:
            # 使用基类的通用初始化逻辑
            if not self.init_client_base():
                self.initialized = False
                return False

            # 优先用自定义 kubeconfig_content 初始化
            if self.kubeconfig_content:
                import tempfile
                from kubernetes import config
                tmp = tempfile.NamedTemporaryFile(delete=False, suffix='.yaml', mode='w', encoding='utf-8')
                tmp.write(self.kubeconfig_content)
                tmp.close()
                kubeconfig_path = tmp.name
                try:
                    config.load_kube_config(config_file=kubeconfig_path)
                finally:
                    os.unlink(kubeconfig_path)
            else:
                from kubernetes import config
                config.load_kube_config()

            self.api_client = client.ApiClient()
            self.dynamic_client = DynamicClient(self.api_client)

            # 认证校验（可选，失败则初始化失败）
            if self.kubeconfig_content:
                tmp = tempfile.NamedTemporaryFile(delete=False, suffix='.yaml', mode='w', encoding='utf-8')
                tmp.write(self.kubeconfig_content)
                tmp.close()
                kubeconfig_path = tmp.name
                try:
                    if not K8sDynamicClient.validate_kubeconfig_auth(kubeconfig_path):
                        logger.error("kubeconfig 认证校验失败，初始化终止")
                        self.initialized = False
                        return False
                finally:
                    os.unlink(kubeconfig_path)
            self.initialized = True
            logger.info("Kubernetes动态客户端初始化成功")
            return True
            
        except Exception as e:
            logger.error(f"初始化 Kubernetes 动态客户端失败: {e}")
            self.initialized = False
            return False
    
    def __del__(self):
        """已移至基类K8sBaseClient"""
        super().__del__()
    
    def get_resource_definition(self, resource_type: str) -> Optional[Any]:
        """
        获取资源定义 - 类似Go Dynamic Client的Resource()方法
        
        Args:
            resource_type: 资源类型，如'pods', 'deployments'等
            
        Returns:
            资源定义对象或None
        """
        if not self.initialized:
            logger.warning(f"动态客户端未初始化，无法获取资源定义: {resource_type}")
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
                    kind=mapping['kind']
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
            logger.error(f"获取资源定义失败 {resource_type}: {str(e)}")
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
            logger.error(f"客户端未初始化，无法列举资源: {resource_type}")
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
                result = resource.get(namespace=namespace)
            elif is_namespaced:
                result = resource.get()
            else:
                result = resource.get()
            
            # 转换为字典格式
            items = result.items if hasattr(result, 'items') else [result]
            converted_items = []
            for item in items:
                try:
                    converted = self._convert_dynamic_object_to_dict(item)
                    converted_items.append(converted)
                except Exception as convert_e:
                    logger.error(f"转换资源对象失败: {str(convert_e)}")
            
            logger.info(f"成功列举资源 {resource_type}: {len(converted_items)} 个")
            return converted_items
            
        except ResourceNotFoundError as e:
            logger.warning(f"资源类型不存在: {resource_type}, 错误: {str(e)}")
            return []
        except Exception as e:
            logger.error(f"❌ [list_resources] 列出资源失败 {resource_type}: {str(e)}")
            logger.error(f"❌ [list_resources] 异常类型: {type(e).__name__}")
            logger.error(f"❌ [list_resources] 详细错误信息: {repr(e)}")
            logger.error(f"❌ [list_resources] 当前映射表键: {list(self.resource_mappings.keys())}")
            logger.error(f"❌ [list_resources] 当前缓存键: {list(self.resource_cache.keys())}")
            return []
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
        # 支持两种配置结构：
        # 1. rule_config 直接包含 resources: rule_config.get('resources', [])
        # 2. rule_config 包含 config 字段: rule_config.get('config', {}).get('resources', [])
        resources_config = rule_config.get('resources', [])
        
        if not resources_config:
            # 尝试从嵌套的config字段获取
            config_data = rule_config.get('config', {})
            resources_config = config_data.get('resources', [])
        
        # 清空现有映射（避免累积）
        self.resource_mappings.clear()
        
        for resource in resources_config:
            
            kind = resource.get('kind', '')
            api_version = resource.get('apiVersion', '')
            namespaced = resource.get('namespaced', True)
            
            if not kind or not api_version:
                logger.warning(f"资源信息不完整，跳过: Kind={kind}, API版本={api_version}")
                continue
            
            # 解析 apiVersion 获取 group 和 version
            if '/' in api_version:
                group, version = api_version.split('/', 1)
            else:
                group = ''
                version = api_version
            
            # 生成复数形式的资源名称
            plural = self._kind_to_plural(kind)
            
            # 添加到资源映射表
            mapping_key = plural.lower()
            mapping_value = {
                'group': group,
                'version': version, 
                'plural': plural.lower(),
                'namespaced': namespaced,
                'kind': kind
            }
            
            self.resource_mappings[mapping_key] = mapping_value
            
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
        logger.debug(f"🧠 [_analyze_resource_strategy] 开始分析资源策略")
        logger.debug(f"🧠 [_analyze_resource_strategy] 输入配置数量: {len(resource_config)}")
        logger.debug(f"🧠 [_analyze_resource_strategy] 输入资源类型: {[r.get('kind', 'unknown') for r in resource_config]}")
        
        configured_kinds = {res['kind'] for res in resource_config}
        logger.debug(f"🧠 [_analyze_resource_strategy] 配置的资源类型集合: {configured_kinds}")
        
        # 定义控制器类型
        workload_controllers = {'Deployment', 'StatefulSet', 'DaemonSet'}
        job_controllers = {'Job', 'CronJob'}
        all_controllers = workload_controllers | job_controllers
        
        logger.debug(f"🧠 [_analyze_resource_strategy] 工作负载控制器定义: {workload_controllers}")
        logger.debug(f"🧠 [_analyze_resource_strategy] 作业控制器定义: {job_controllers}")
        logger.debug(f"🧠 [_analyze_resource_strategy] 所有控制器定义: {all_controllers}")
        
        # 分析配置
        has_pods = 'Pod' in configured_kinds
        has_controllers = bool(configured_kinds & all_controllers)
        configured_controllers = configured_kinds & all_controllers
        
        logger.debug(f"🧠 [_analyze_resource_strategy] 配置分析结果:")
        logger.debug(f"    - 包含Pod: {has_pods}")
        logger.debug(f"    - 包含控制器: {has_controllers}")
        logger.debug(f"    - 配置的控制器: {configured_controllers}")
        logger.debug(f"    - 控制器数量: {len(configured_controllers)}")
        
        strategy = {
            'mode': 'default',
            'get_pods_directly': has_pods,
            'get_controllers': list(configured_controllers),
            'filter_controlled_pods': False,
            'optimization_applied': False,
            'reasoning': 'Default strategy - get all configured resources'
        }
        
        logger.debug(f"🧠 [_analyze_resource_strategy] 初始策略: {json.dumps(strategy, indent=2)}")
        
        # 如果同时配置了控制器和Pod，应用优化策略
        if has_pods and has_controllers:
            logger.debug(f"🧠 [_analyze_resource_strategy] 同时包含Pod和控制器，进入优化策略分析")
            logger.debug(f"🧠 [_analyze_resource_strategy] 控制器数量: {len(configured_controllers)}")
            
            if len(configured_controllers) >= 2:
                # 多个控制器 + Pod：推荐仅控制器策略
                logger.debug("🧠 [_analyze_resource_strategy] 触发仅控制器策略（多个控制器+Pod）")
                strategy.update({
                    'mode': 'controller_only',
                    'get_pods_directly': False,
                    'filter_controlled_pods': False,
                    'optimization_applied': True,
                    'reasoning': '多个控制器+Pod配置，采用仅控制器策略避免重复'
                })
                logger.debug("🧠 [_analyze_resource_strategy] 应用仅控制器策略，禁用直接获取Pod")
            else:
                # 单个控制器 + Pod：推荐智能混合策略
                logger.debug("🧠 [_analyze_resource_strategy] 触发智能混合策略（单个控制器+Pod）")
                strategy.update({
                    'mode': 'hybrid_smart',
                    'get_pods_directly': True,
                    'filter_controlled_pods': True,
                    'optimization_applied': True,
                    'reasoning': '控制器+Pod配置，采用智能混合策略过滤受控Pod'
                })
                logger.debug("🧠 [_analyze_resource_strategy] 应用智能混合策略，启用Pod过滤")
        else:
            logger.debug("🧠 [_analyze_resource_strategy] 使用默认策略（无Pod+控制器组合，或只有单一类型）")
            logger.debug(f"🧠 [_analyze_resource_strategy] 原因: has_pods={has_pods}, has_controllers={has_controllers}")
        
        logger.debug(f"✅ [_analyze_resource_strategy] 最终策略分析完成:")
        logger.debug(f"    - 模式: {strategy['mode']}")
        logger.debug(f"    - 直接获取Pod: {strategy['get_pods_directly']}")
        logger.debug(f"    - 获取控制器: {strategy['get_controllers']}")
        logger.debug(f"    - 过滤受控Pod: {strategy['filter_controlled_pods']}")
        logger.debug(f"    - 优化应用: {strategy['optimization_applied']}")
        logger.debug(f"    - 原因: {strategy['reasoning']}")
        
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
        logger.info("🚀 [list_resources_from_config_optimized] ==== 开始优化资源获取 ====")
        logger.debug(f"🚀 [list_resources_from_config_optimized] 输入配置结构: {list(rule_config.keys())}")
        logger.debug(f"🚀 [list_resources_from_config_optimized] 输入配置类型: {type(rule_config)}")
        
        # 构建资源映射
        logger.debug("🚀 [list_resources_from_config_optimized] 步骤1: 构建资源映射")
        logger.debug(f"🚀 [list_resources_from_config_optimized] 构建前映射表键: {list(self.resource_mappings.keys())}")
        self.build_resource_mappings_from_config(rule_config)
        logger.debug(f"🚀 [list_resources_from_config_optimized] 构建后映射表键: {list(self.resource_mappings.keys())}")
        
        # 分析资源获取策略
        logger.debug("🚀 [list_resources_from_config_optimized] 步骤2: 分析资源获取策略")
        resource_config = rule_config.get('config', {}).get('resources', [])
        logger.debug(f"🚀 [list_resources_from_config_optimized] 从config.resources提取的资源: {[r.get('kind', 'unknown') for r in resource_config]}")
        
        if not resource_config:
            # 尝试从根级别获取
            resource_config = rule_config.get('resources', [])
            logger.debug(f"🚀 [list_resources_from_config_optimized] 从根级别resources提取的资源: {[r.get('kind', 'unknown') for r in resource_config]}")
        
        strategy = self._analyze_resource_strategy(resource_config)
        
        logger.info(f"🚀 [list_resources_from_config_optimized] 选定资源获取策略: {strategy['mode']} - {strategy['reasoning']}")
        logger.debug(f"🚀 [list_resources_from_config_optimized] 完整策略信息: {json.dumps(strategy, indent=2)}")
        
        # 获取命名空间配置
        logger.debug("🚀 [list_resources_from_config_optimized] 步骤3: 解析命名空间配置")
        scope_config = rule_config.get('scope', {})
        logger.debug(f"🚀 [list_resources_from_config_optimized] scope配置: {scope_config}")
        namespace_config = scope_config.get('namespaces', {})
        logger.debug(f"🚀 [list_resources_from_config_optimized] namespace配置: {namespace_config}")
        include_namespaces = namespace_config.get('include', [])
        exclude_namespaces = namespace_config.get('exclude', [])
        
        logger.debug(f"🚀 [list_resources_from_config_optimized] 命名空间过滤:")
        logger.debug(f"    - 包含命名空间: {include_namespaces}")
        logger.debug(f"    - 排除命名空间: {exclude_namespaces}")
        
        resources = {}
        
        # 根据策略获取资源
        logger.debug(f"🚀 [list_resources_from_config_optimized] 步骤4: 根据策略 '{strategy['mode']}' 获取资源")
        
        if strategy['mode'] == 'controller_only':
            logger.info("🎯 [list_resources_from_config_optimized] 执行仅控制器策略")
            controllers_to_get = strategy['get_controllers']
            logger.debug(f"🎯 [list_resources_from_config_optimized] 需要获取的控制器: {controllers_to_get}")
            
            # 仅获取控制器资源
            for i, resource_type in enumerate(controllers_to_get):
                resource_type_lower = resource_type.lower() + 's'
                logger.debug(f"🎯 [list_resources_from_config_optimized] 处理控制器 {i+1}/{len(controllers_to_get)}: {resource_type} -> {resource_type_lower}")
                
                if resource_type_lower in self.resource_mappings:
                    logger.debug(f"✅ [list_resources_from_config_optimized] 在资源映射中找到: {resource_type_lower}")
                    logger.debug(f"🚀 [list_resources_from_config_optimized] 调用 _get_namespaced_resources...")
                    controller_resources = self._get_namespaced_resources(
                        resource_type_lower, include_namespaces, exclude_namespaces
                    )
                    resources[resource_type_lower] = controller_resources
                    logger.info(f"✅ [list_resources_from_config_optimized] 获取 {resource_type_lower}: {len(controller_resources)} 个")
                else:
                    logger.warning(f"⚠️ [list_resources_from_config_optimized] 在资源映射中未找到: {resource_type_lower}")
                    logger.debug(f"🎯 [list_resources_from_config_optimized] 可用的映射: {list(self.resource_mappings.keys())}")
        
        elif strategy['mode'] == 'hybrid_smart':
            logger.info("🎯 [list_resources_from_config_optimized] 执行智能混合策略")
            # 智能混合模式：获取控制器 + 过滤独立Pod
            controllers = {}
            controllers_to_get = strategy['get_controllers']
            logger.debug(f"🎯 [list_resources_from_config_optimized] 需要获取的控制器: {controllers_to_get}")
            
            # 获取控制器
            logger.debug("🎯 [list_resources_from_config_optimized] 阶段1: 获取控制器资源")
            for i, resource_type in enumerate(controllers_to_get):
                resource_type_lower = resource_type.lower() + 's'
                logger.debug(f"🎯 [list_resources_from_config_optimized] 处理控制器 {i+1}/{len(controllers_to_get)}: {resource_type} -> {resource_type_lower}")
                
                if resource_type_lower in self.resource_mappings:
                    logger.debug(f"✅ [list_resources_from_config_optimized] 在资源映射中找到控制器: {resource_type_lower}")
                    controller_resources = self._get_namespaced_resources(
                        resource_type_lower, include_namespaces, exclude_namespaces
                    )
                    resources[resource_type_lower] = controller_resources
                    controllers[resource_type_lower] = controller_resources
                    logger.info(f"✅ [list_resources_from_config_optimized] 获取控制器 {resource_type_lower}: {len(controller_resources)} 个")
                else:
                    logger.warning(f"⚠️ [list_resources_from_config_optimized] 控制器资源映射缺失: {resource_type_lower}")
                    logger.debug(f"🎯 [list_resources_from_config_optimized] 可用的映射: {list(self.resource_mappings.keys())}")
            
            # 获取Pod并过滤
            logger.debug("🎯 [list_resources_from_config_optimized] 阶段2: 获取和过滤Pod资源")
            logger.debug(f"🎯 [list_resources_from_config_optimized] 是否需要直接获取Pod: {strategy['get_pods_directly']}")
            logger.debug(f"🎯 [list_resources_from_config_optimized] pods是否在映射中: {'pods' in self.resource_mappings}")
            
            if strategy['get_pods_directly'] and 'pods' in self.resource_mappings:
                logger.debug("🎯 [list_resources_from_config_optimized] 获取所有Pod...")
                all_pods = self._get_namespaced_resources(
                    'pods', include_namespaces, exclude_namespaces
                )
                logger.debug(f"🎯 [list_resources_from_config_optimized] 获取到 {len(all_pods)} 个Pod")
                
                if strategy['filter_controlled_pods']:
                    logger.debug("🎯 [list_resources_from_config_optimized] 过滤受控制器管理的Pod...")
                    logger.debug(f"🎯 [list_resources_from_config_optimized] 可用控制器数据: {list(controllers.keys())}")
                    # 过滤掉受控制器管理的Pod
                    independent_pods = self._filter_controlled_pods(all_pods, controllers)
                    resources['pods'] = independent_pods
                    logger.info(f"✅ [list_resources_from_config_optimized] 过滤后独立Pod: {len(independent_pods)} 个")
                else:
                    resources['pods'] = all_pods
                    logger.info(f"✅ [list_resources_from_config_optimized] 未过滤Pod: {len(all_pods)} 个")
            else:
                logger.debug("🎯 [list_resources_from_config_optimized] 跳过Pod获取（策略不需要或映射缺失）")
        
        else:
            logger.info("🎯 [list_resources_from_config_optimized] 执行默认策略")
            # 默认策略：获取所有配置的资源（不重复构建资源映射）
            logger.debug("🎯 [list_resources_from_config_optimized] 使用已构建的资源映射获取所有配置的资源")
            logger.debug(f"🎯 [list_resources_from_config_optimized] 已构建的映射表: {list(self.resource_mappings.keys())}")
            
            # 获取所有配置的资源类型
            total_types = len(self.resource_mappings)
            logger.debug(f"🎯 [list_resources_from_config_optimized] 需要处理 {total_types} 种资源类型")
            
            for i, resource_type in enumerate(self.resource_mappings.keys()):
                logger.debug(f"🎯 [list_resources_from_config_optimized] 处理资源类型 {i+1}/{total_types}: {resource_type}")
                
                try:
                    mapping = self.resource_mappings[resource_type]
                    is_namespaced = mapping.get('namespaced', True)
                    logger.debug(f"🎯 [list_resources_from_config_optimized] 资源 {resource_type} 映射信息: {json.dumps(mapping, indent=2)}")
                    logger.debug(f"🎯 [list_resources_from_config_optimized] 是否命名空间级别: {is_namespaced}")
                    
                    if is_namespaced:
                        logger.debug(f"🎯 [list_resources_from_config_optimized] 处理命名空间级别资源: {resource_type}")
                        # 处理命名空间级别的资源
                        if include_namespaces:
                            logger.debug(f"🎯 [list_resources_from_config_optimized] 获取指定命名空间 {include_namespaces} 的资源")
                            # 只获取指定命名空间的资源
                            all_resources = []
                            for j, namespace in enumerate(include_namespaces):
                                if namespace not in exclude_namespaces:
                                    logger.debug(f"🎯 [list_resources_from_config_optimized] 处理命名空间 {j+1}/{len(include_namespaces)}: {namespace}")
                                    ns_resources = self.list_resources(resource_type, namespace=namespace)
                                    all_resources.extend(ns_resources)
                                    logger.debug(f"🎯 [list_resources_from_config_optimized] 命名空间 {namespace} 获取: {len(ns_resources)} 个")
                                else:
                                    logger.debug(f"🎯 [list_resources_from_config_optimized] 跳过排除的命名空间: {namespace}")
                            resources[resource_type] = all_resources
                            logger.debug(f"🎯 [list_resources_from_config_optimized] 指定命名空间总计: {len(all_resources)} 个")
                        else:
                            logger.debug("🎯 [list_resources_from_config_optimized] 获取所有命名空间的资源")
                            # 获取所有命名空间的资源，然后过滤
                            all_resources = self.list_resources(resource_type)
                            logger.debug(f"🎯 [list_resources_from_config_optimized] 所有命名空间获取: {len(all_resources)} 个")
                            if exclude_namespaces:
                                logger.debug(f"🎯 [list_resources_from_config_optimized] 过滤排除的命名空间: {exclude_namespaces}")
                                filtered_resources = [
                                    res for res in all_resources 
                                    if res.get('metadata', {}).get('namespace') not in exclude_namespaces
                                ]
                                resources[resource_type] = filtered_resources
                                logger.debug(f"🎯 [list_resources_from_config_optimized] 过滤后: {len(filtered_resources)} 个（原: {len(all_resources)} 个）")
                            else:
                                resources[resource_type] = all_resources
                                logger.debug(f"🎯 [list_resources_from_config_optimized] 无需过滤: {len(all_resources)} 个")
                    else:
                        logger.debug(f"🎯 [list_resources_from_config_optimized] 处理集群级别资源: {resource_type}")
                        # 集群级别资源，直接获取
                        cluster_resources = self.list_resources(resource_type)
                        resources[resource_type] = cluster_resources
                        logger.debug(f"🎯 [list_resources_from_config_optimized] 集群级别获取: {len(cluster_resources)} 个")
                        
                    logger.info(f"✅ [list_resources_from_config_optimized] 获取 {resource_type}: {len(resources[resource_type])} 个资源")
                    
                except Exception as e:
                    logger.error(f"❌ [list_resources_from_config_optimized] 获取资源 {resource_type} 失败: {str(e)}")
                    logger.error(f"❌ [list_resources_from_config_optimized] 异常类型: {type(e).__name__}")
                    logger.error(f"❌ [list_resources_from_config_optimized] 详细错误: {repr(e)}")
                    resources[resource_type] = []
                    
                except Exception as e:
                    logger.error(f"获取资源类型 {resource_type} 失败: {str(e)}")
                    logger.debug(f"详细错误: {repr(e)}")
                    resources[resource_type] = []
        
        # 记录结果
        total_resources = sum(len(res_list) for res_list in resources.values())
        logger.info(f"优化后获取 {len(resources)} 种资源类型，共 {total_resources} 个资源")
        logger.debug(f"资源分布: {[(k, len(v)) for k, v in resources.items()]}")
        
        if strategy['optimization_applied']:
            logger.info("✅ 已应用资源获取优化，避免重复获取")
        
        logger.info("==== 优化资源获取完成 ====")
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
        logger.debug(f"🏢 [_get_namespaced_resources] 开始获取命名空间资源")
        logger.debug(f"🏢 [_get_namespaced_resources] 输入参数:")
        logger.debug(f"    - resource_type: {resource_type}")
        logger.debug(f"    - include_namespaces: {include_namespaces}")
        logger.debug(f"    - exclude_namespaces: {exclude_namespaces}")
        
        try:
            # 检查资源映射
            if resource_type not in self.resource_mappings:
                logger.error(f"❌ [_get_namespaced_resources] 资源类型 {resource_type} 不在映射表中")
                logger.debug(f"🏢 [_get_namespaced_resources] 当前映射表键: {list(self.resource_mappings.keys())}")
                return []
            
            mapping = self.resource_mappings[resource_type]
            is_namespaced = mapping.get('namespaced', True)
            
            logger.debug(f"🏢 [_get_namespaced_resources] 资源映射信息:")
            logger.debug(f"    - 映射: {json.dumps(mapping, indent=2)}")
            logger.debug(f"    - 是否命名空间级别: {is_namespaced}")
            
            if not is_namespaced:
                # 集群级别资源
                logger.debug(f"🏢 [_get_namespaced_resources] 资源为集群级别，直接获取")
                logger.debug(f"🚀 [_get_namespaced_resources] 调用 list_resources({resource_type})...")
                cluster_resources = self.list_resources(resource_type)
                logger.debug(f"✅ [_get_namespaced_resources] 集群级别资源获取完成: {len(cluster_resources)} 个")
                return cluster_resources
            
            # 命名空间级别资源处理
            if include_namespaces:
                # 只获取指定命名空间的资源
                logger.debug(f"🏢 [_get_namespaced_resources] 策略: 获取指定命名空间的资源")
                logger.debug(f"🏢 [_get_namespaced_resources] 需要处理的命名空间: {include_namespaces}")
                all_resources = []
                
                for i, namespace in enumerate(include_namespaces):
                    logger.debug(f"🏢 [_get_namespaced_resources] 处理命名空间 {i+1}/{len(include_namespaces)}: {namespace}")
                    
                    if namespace not in exclude_namespaces:
                        logger.debug(f"🏢 [_get_namespaced_resources] 命名空间 {namespace} 未被排除，开始获取")
                        logger.debug(f"🚀 [_get_namespaced_resources] 调用 list_resources({resource_type}, {namespace})...")
                        namespace_resources = self.list_resources(resource_type, namespace)
                        all_resources.extend(namespace_resources)
                        logger.debug(f"✅ [_get_namespaced_resources] 命名空间 {namespace} 获取: {len(namespace_resources)} 个")
                    else:
                        logger.debug(f"⚠️ [_get_namespaced_resources] 跳过排除的命名空间: {namespace}")
                
                logger.debug(f"✅ [_get_namespaced_resources] 指定命名空间总资源数量: {len(all_resources)}")
                return all_resources
            else:
                # 获取所有命名空间的资源，然后过滤
                logger.debug(f"🏢 [_get_namespaced_resources] 策略: 获取所有命名空间的资源")
                logger.debug(f"🚀 [_get_namespaced_resources] 调用 list_resources({resource_type})...")
                all_resources = self.list_resources(resource_type)
                logger.debug(f"✅ [_get_namespaced_resources] 所有命名空间资源获取完成: {len(all_resources)} 个")
                
                if exclude_namespaces:
                    logger.debug(f"🏢 [_get_namespaced_resources] 需要过滤排除的命名空间: {exclude_namespaces}")
                    logger.debug(f"🏢 [_get_namespaced_resources] 开始过滤...")
                    
                    # 统计过滤前各命名空间的资源
                    ns_count_before = {}
                    for res in all_resources:
                        ns = res.get('metadata', {}).get('namespace', '<cluster-scoped>')
                        ns_count_before[ns] = ns_count_before.get(ns, 0) + 1
                    logger.debug(f"🏢 [_get_namespaced_resources] 过滤前各命名空间资源数: {ns_count_before}")
                    
                    filtered_resources = [
                        resource for resource in all_resources
                        if resource.get('metadata', {}).get('namespace') not in exclude_namespaces
                    ]
                    
                    # 统计过滤后各命名空间的资源
                    ns_count_after = {}
                    for res in filtered_resources:
                        ns = res.get('metadata', {}).get('namespace', '<cluster-scoped>')
                        ns_count_after[ns] = ns_count_after.get(ns, 0) + 1
                    logger.debug(f"🏢 [_get_namespaced_resources] 过滤后各命名空间资源数: {ns_count_after}")
                    
                    logger.debug(f"✅ [_get_namespaced_resources] 过滤完成: {len(filtered_resources)} 个（原: {len(all_resources)} 个）")
                    return filtered_resources
                else:
                    logger.debug(f"🏢 [_get_namespaced_resources] 无需过滤，返回所有资源: {len(all_resources)} 个")
                    return all_resources
                
        except Exception as e:
            logger.error(f"❌ [_get_namespaced_resources] 获取资源 {resource_type} 失败: {str(e)}")
            logger.error(f"❌ [_get_namespaced_resources] 异常类型: {type(e).__name__}")
            logger.error(f"❌ [_get_namespaced_resources] 详细错误: {repr(e)}")
            logger.error(f"❌ [_get_namespaced_resources] 当前资源映射: {list(self.resource_mappings.keys())}")
            return []
    
    # 新增：从集群json文件提取kubeconfig并写入临时文件
    def get_kubeconfig_file_from_cluster_json(cluster_json_path: str) -> str:
        """
        从 data/clusters/<集群名>.json 读取 kubeconfig 字符串，写入临时文件并返回路径。
        Args:
            cluster_json_path: 集群json文件路径
        Returns:
            kubeconfig 临时文件路径
        Raises:
            Exception: kubeconfig 字段不存在或写入失败
        """
        with open(cluster_json_path, 'r', encoding='utf-8') as f:
            cluster_data = json.load(f)
        kubeconfig_str = cluster_data.get('kubeconfig')
        if not kubeconfig_str:
            raise Exception(f"kubeconfig 字段不存在于 {cluster_json_path}")
        # 写入临时文件
        tmp = tempfile.NamedTemporaryFile(delete=False, suffix='.yaml', mode='w', encoding='utf-8')
        tmp.write(kubeconfig_str)
        tmp.close()
        return tmp.name
    
    @staticmethod
    def validate_kubeconfig_auth(kubeconfig_path: str) -> bool:
        """
        用纯 Python kubernetes client 校验 kubeconfig 是否能正常认证和访问 API。
        Args:
            kubeconfig_path: kubeconfig 文件路径
        Returns:
            True 表示认证通过，False 表示失败
        """
        try:
            from kubernetes import config, client
            from kubernetes.client.rest import ApiException
            
            # 临时保存当前配置
            original_config = None
            try:
                original_config = config.KUBE_CONFIG_DEFAULT_LOCATION
            except:
                pass
            
            try:
                # 加载指定的 kubeconfig
                config.load_kube_config(config_file=kubeconfig_path)
                
                # 创建 API 客户端并测试连接
                v1 = client.CoreV1Api()
                
                # 尝试列出命名空间，这是一个基本的权限测试
                # 设置较短的超时时间避免长时间等待
                api_client = v1.api_client
                api_client.rest_timeout = 10  # 10秒超时
                
                namespaces = v1.list_namespace(_request_timeout=5)
                
                logger.info(f"kubeconfig 认证校验通过: {kubeconfig_path}")
                logger.debug(f"成功连接到集群，发现 {len(namespaces.items)} 个命名空间")
                return True
                
            except ApiException as e:
                if e.status == 401:
                    logger.error(f"kubeconfig 认证失败 - 未授权: {kubeconfig_path}")
                elif e.status == 403:
                    logger.error(f"kubeconfig 认证失败 - 权限不足: {kubeconfig_path}")
                else:
                    logger.error(f"kubeconfig API 调用失败 (状态码 {e.status}): {e.reason}")
                return False
                
            except Exception as e:
                logger.error(f"kubeconfig 连接测试失败: {str(e)}")
                return False
                
            finally:
                # 恢复原始配置（如果存在）
                if original_config:
                    try:
                        config.load_kube_config(config_file=original_config)
                    except:
                        pass
                        
        except ImportError:
            logger.error("kubernetes client 库未安装，无法验证 kubeconfig")
            return False
        except Exception as e:
            logger.error(f"kubeconfig 认证校验异常: {str(e)}")
            return False

# 用法示例：
# kubeconfig_path = get_kubeconfig_file_from_cluster_json('data/clusters/demo.json')
# k8s_client = config.new_client_from_config(config_file=kubeconfig_path)
# dyn_client = DynamicClient(k8s_client)

