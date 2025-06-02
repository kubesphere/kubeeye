#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
OPA 规则巡检器，继承自基类实现，使用规则处理器
"""

import logging
import os
import tempfile
import json
import subprocess
import datetime
from typing import Dict, List, Any, Optional, Tuple

from inspectors.base_inspector import BaseInspector
from utils.k8s_client import K8sClient
from utils.rule_loader import Rule

# 设置日志
logger = logging.getLogger(__name__)

# 自定义JSON编码器，处理datetime类型
class DateTimeEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime.datetime):
            return obj.isoformat()
        elif isinstance(obj, datetime.date):
            return obj.isoformat()
        elif isinstance(obj, datetime.timedelta):
            return str(obj)
        return super().default(obj)

class OpaInspector(BaseInspector):
    """
    OPA规则巡检器，使用OPA规则检查Kubernetes资源
    """
    
    def __init__(self, opa_config: Dict[str, Any]):
        """
        初始化OPA巡检器
        
        Args:
            opa_config: OPA配置
        """
        self.k8s_client = K8sClient(opa_config.get('kubeconfig'))
        self.opa_path = opa_config.get('opa_path', 'opa')  # OPA可执行文件路径
        super().__init__(opa_config)
    
    @property
    def inspector_type(self) -> str:
        return "opa"
    
    def _prepare_context(self, cluster_name: str) -> Dict:
        """
        准备OPA巡检上下文，预加载资源
        
        Args:
            cluster_name: 集群名称
            
        Returns:
            准备好的上下文
        """
        context = super()._prepare_context(cluster_name)
        # 预加载资源数据，避免多次获取
        context['resources'] = self._get_cluster_resources()
        return context
    
    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        """
        应用OPA规则进行检查
        
        Args:
            rule: 规则对象
            context: 上下文
            
        Returns:
            检查结果
        """
        # 获取规则内容
        rule_content = self._get_rule_content(rule)
        if not rule_content:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='skipped',
                description="规则未定义OPA规则内容",
                severity='info',
                details="检查被跳过，因为规则没有定义有效的OPA规则内容",
                solution="检查规则定义"
            )
        
        # 从上下文获取已加载的资源
        resources = context.get('resources', {})
        if not resources:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='error',
                description="无法获取集群资源",
                severity='warning',
                details="未能加载集群资源数据",
                solution="检查Kubernetes API连接"
            )
        
        # 评估OPA规则
        success, result = self._evaluate_rule(rule_content, resources)
        
        if not success:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='error',
                description="执行OPA规则失败",
                severity='warning',
                details=f"错误: {result}",
                solution="检查OPA规则语法或OPA运行环境"
            )
        
        # 解析违规结果
        violations = self._parse_violations(result)
        
        if violations:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='failed',
                description=f"发现 {len(violations)} 个资源不符合规则要求",
                severity=rule.severity if hasattr(rule, 'severity') else 'warning',
                details=self._format_violations(violations),
                solution=rule.solution if hasattr(rule, 'solution') else "查看违规详情并修正资源配置"
            )
        else:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='passed',
                description="所有资源符合规则要求",
                severity='info',
                details="所有Kubernetes资源均通过OPA规则验证",
                solution=""
            )
    
    def _get_rule_content(self, rule: Rule) -> str:
        """
        获取规则的OPA Rego内容
        
        Args:
            rule: 规则对象
            
        Returns:
            Rego规则内容
        """
        # 只从新格式配置路径获取规则内容
        content = self.get_rule_config(rule, 'rego.inline') or ''
        
        # 如果有外部文件引用
        file_path = self.get_rule_config(rule, 'rego.file')
        if not content and file_path:
            try:
                with open(file_path, 'r') as f:
                    content = f.read()
            except Exception as e:
                logger.error(f"读取规则文件 {file_path} 失败: {e}")
        
        return content
    
    def _get_cluster_resources(self) -> Dict:
        """
        获取集群资源
        
        Returns:
            包含集群各类资源的字典
        """
        resources = {}
        
        # 要获取的资源类型列表
        resource_types = [
            ('pods', 'list_pods_all_namespaces'),
            ('deployments', 'list_deployments_all_namespaces'),
            ('statefulsets', 'list_statefulsets_all_namespaces'),
            ('daemonsets', 'list_daemonsets_all_namespaces'),
            ('services', 'list_services_all_namespaces'),
            ('ingresses', 'list_ingresses_all_namespaces'),
            ('serviceaccounts', 'list_serviceaccounts_all_namespaces'),
            ('configmaps', 'list_configmaps_all_namespaces'),
            ('secrets', 'list_secrets_all_namespaces'),
            ('persistentvolumeclaims', 'list_persistent_volume_claims_all_namespaces'),
            ('namespaces', 'list_namespaces')
        ]
        
        try:
            for resource_name, method_name in resource_types:
                try:
                    # 动态调用方法
                    method = getattr(self.k8s_client, method_name)
                    resources[resource_name] = method()
                except Exception as e:
                    logger.warning(f"获取资源 {resource_name} 失败: {e}")
                    resources[resource_name] = []
        except Exception as e:
            logger.exception(f"获取集群资源时出错: {str(e)}")
        
        return resources
    
    def _evaluate_rule(self, rule_content: str, resources: Dict) -> Tuple[bool, Any]:
        """
        评估OPA规则
        
        Args:
            rule_content: OPA规则内容
            resources: 集群资源数据
            
        Returns:
            (success, result) 元组
        """
        try:
            # 创建临时文件存储规则和输入数据
            with tempfile.NamedTemporaryFile(suffix='.rego', delete=False) as rule_file, \
                 tempfile.NamedTemporaryFile(suffix='.json', delete=False) as input_file:
                
                # 写入规则文件
                rule_file.write(rule_content.encode('utf-8'))
                rule_file.flush()
                
                # 写入输入文件 - 使用自定义编码器处理datetime类型
                input_file.write(json.dumps(resources, cls=DateTimeEncoder).encode('utf-8'))
                input_file.flush()
                
                try:
                    # 构建OPA命令
                    cmd = [
                        self.opa_path, 'eval',
                        '--data', rule_file.name,
                        '--input', input_file.name,
                        '--format', 'json',
                        'data.kubeeye.violations'
                    ]
                    
                    # 执行OPA评估
                    process = subprocess.run(
                        cmd,
                        stdout=subprocess.PIPE,
                        stderr=subprocess.PIPE,
                        text=True
                    )
                finally:
                    # 确保临时文件被删除
                    try:
                        os.unlink(rule_file.name)
                    except:
                        pass
                    
                    try:
                        os.unlink(input_file.name)
                    except:
                        pass
                
                if process.returncode != 0:
                    return False, process.stderr
                
                # 解析OPA输出
                try:
                    result = json.loads(process.stdout)
                    return True, result
                except json.JSONDecodeError:
                    return False, f"无法解析OPA输出: {process.stdout}"
                
        except Exception as e:
            logger.exception(f"执行OPA规则时出错: {str(e)}")
            return False, str(e)
    
    def _parse_violations(self, result: Dict) -> List[Dict]:
        """
        解析OPA规则违规结果
        
        Args:
            result: OPA评估结果
            
        Returns:
            违规列表
        """
        violations = []
        
        try:
            # 标准格式:
            # { "result": [ { "violations": [ {...violation data...}, ... ] } ] }
            if 'result' in result and isinstance(result['result'], list):
                for item in result['result']:
                    if isinstance(item, dict) and 'violations' in item and isinstance(item['violations'], list):
                        violations.extend(item['violations'])
            
            # 替代格式:
            # { "result": [ { "expressions": [ { "value": [ {...violation data...}, ... ] } ] } ] }
            elif 'result' in result and isinstance(result['result'], list):
                for item in result['result']:
                    if isinstance(item, dict) and 'expressions' in item and isinstance(item['expressions'], list):
                        for expr in item['expressions']:
                            if isinstance(expr, dict) and 'value' in expr and isinstance(expr['value'], list):
                                violations.extend(expr['value'])
        except Exception as e:
            logger.warning(f"解析OPA违规结果时出错: {str(e)}")
        
        return violations
    
    def _format_violations(self, violations: List[Dict]) -> str:
        """
        格式化违规信息
        
        Args:
            violations: 违规列表
            
        Returns:
            格式化的违规信息
        """
        lines = []
        for i, violation in enumerate(violations, 1):
            lines.append(f"违规 #{i}:")
            
            # 格式化资源引用
            if 'resource_kind' in violation and 'resource_name' in violation:
                kind = violation.get('resource_kind', 'Unknown')
                name = violation.get('resource_name', 'Unknown')
                namespace = violation.get('resource_namespace', 'default')
                lines.append(f"  资源: {kind}/{name} (命名空间: {namespace})")
            else:
                name = violation.get('name', 'Unknown')
                namespace = violation.get('namespace', 'default')
                lines.append(f"  资源: {name} (命名空间: {namespace})")
            
            # 违规信息
            if 'message' in violation:
                lines.append(f"  信息: {violation['message']}")
            
            # 严重性
            if 'severity' in violation:
                lines.append(f"  严重性: {violation['severity']}")
            
            # 添加空行
            lines.append("")
        
        return "\n".join(lines)
