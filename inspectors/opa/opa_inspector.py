#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
OPA 规则巡检器 - 简化版本
专注核心功能，去除冗余代码和过度日志
"""

import logging
import os
import tempfile
import json
import subprocess
import datetime
from typing import Dict, List, Any, Optional

from inspectors.base_inspector import BaseInspector
from utils.k8s_dynamic_client import K8sDynamicClient
from utils.rule_loader import Rule

logger = logging.getLogger(__name__)


class DateTimeEncoder(json.JSONEncoder):
    """自定义JSON编码器，处理datetime类型"""
    def default(self, obj):
        if isinstance(obj, (datetime.datetime, datetime.date)):
            return obj.isoformat()
        elif isinstance(obj, datetime.timedelta):
            return str(obj)
        return super().default(obj)

class OpaInspector(BaseInspector):
    """OPA规则巡检器 - 简化版本"""
    
    def __init__(self, opa_config: Dict[str, Any]):
        self.k8s_client = K8sDynamicClient(opa_config.get('kubeconfig'))
        self.opa_path = opa_config.get('opa_path', '/usr/local/bin/opa')
        super().__init__(opa_config)
    
    @property
    def inspector_type(self) -> str:
        return "opa"
    
    def validate_rule(self, rule: Rule) -> List[str]:
        """验证规则配置"""
        issues = []
        
        # 检查Rego规则
        if not (self.get_rule_config(rule, 'rego.inline') or 
                self.get_rule_config(rule, 'rego.file')):
            issues.append("缺少Rego规则配置")
        
        # 检查资源配置
        if not self.get_rule_config(rule, 'resources', []):
            issues.append("缺少资源配置")
            
        # 检查断言配置
        if not self.get_rule_config(rule, 'assertions', []):
            issues.append("缺少断言配置")
            
        return issues

    def _prepare_context(self, cluster_name: str) -> Dict:
        """准备巡检上下文"""
        return super()._prepare_context(cluster_name)

    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        """执行OPA规则检查"""
        logger.info(f"执行规则: {rule.id}")
        
        try:
            # 获取Rego规则内容
            rego_content = self._get_rego_content(rule)
            if not rego_content:
                return self._error_result(rule, "无法获取Rego规则内容")
            
            # 获取集群资源
            resources = self._get_cluster_resources(rule)
            if not resources:
                return self._pass_result(rule, "无匹配资源")
            
            # 执行OPA评估
            violations = self._evaluate_opa(rego_content, resources)
            
            # 评估断言
            return self._evaluate_assertions(rule, violations, len(resources))
            
        except Exception as e:
            logger.error(f"规则 {rule.id} 执行失败: {e}")
            return self._error_result(rule, f"执行失败: {str(e)}")
    
    def _get_rego_content(self, rule: Rule) -> Optional[str]:
        """获取Rego规则内容"""
        # 尝试从inline配置获取
        rego_content = self.get_rule_config(rule, 'rego.inline')
        if rego_content:
            return rego_content
            
        # 尝试从文件获取
        rego_file = self.get_rule_config(rule, 'rego.file')
        if rego_file and os.path.exists(rego_file):
            try:
                with open(rego_file, 'r', encoding='utf-8') as f:
                    return f.read()
            except Exception as e:
                logger.error(f"读取Rego文件失败: {e}")
        
        return None
    
    def _get_cluster_resources(self, rule: Rule) -> List[Dict]:
        """获取集群资源"""
        try:
            # 优先使用优化版本
            if hasattr(self.k8s_client, 'list_resources_from_config_optimized'):
                resources_dict = self.k8s_client.list_resources_from_config_optimized(rule.config)
            else:
                resources_dict = self.k8s_client.list_resources_from_config(rule.config)
            
            # 展平资源列表
            all_resources = []
            for resource_list in resources_dict.values():
                all_resources.extend(resource_list)
            
            logger.info(f"获取到 {len(all_resources)} 个资源")
            return all_resources
            
        except Exception as e:
            logger.error(f"获取集群资源失败: {e}")
            return []
    
    def _evaluate_opa(self, rego_content: str, resources: List[Dict]) -> List[Dict]:
        """执行OPA评估"""
        if not resources:
            return []
            
        rego_path = None
        input_path = None
        
        try:
            # 创建临时文件
            with tempfile.NamedTemporaryFile(suffix='.rego', delete=False) as f:
                f.write(rego_content.encode('utf-8'))
                rego_path = f.name
                
            input_data = {"resources": resources}
            with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
                json.dump(input_data, f, cls=DateTimeEncoder)
                input_path = f.name
            
            # 执行OPA命令
            cmd = [
                self.opa_path, "eval",
                f"--data={rego_path}",
                f"--input={input_path}",
                "data.kubernetes.violations"
            ]
            
            result = subprocess.run(
                cmd, 
                stdout=subprocess.PIPE, 
                stderr=subprocess.PIPE,
                universal_newlines=True,
                check=True
            )
            
            # 解析结果
            if result.stdout:
                output = json.loads(result.stdout.strip())
                
                # 从 result[0].expressions[0].value 获取实际违规数据
                try:
                    if 'result' in output and len(output['result']) > 0:
                        result_item = output['result'][0]
                        if 'expressions' in result_item and len(result_item['expressions']) > 0:
                            violations = result_item['expressions'][0].get('value', [])
                            logger.info(f"发现 {len(violations)} 个违规")
                            return violations if isinstance(violations, list) else []
                except (KeyError, IndexError, TypeError) as e:
                    logger.error(f"解析OPA结果时出错: {e}")
                    # 尝试旧的解析方式作为备选
                    violations = output.get('result', [])
                    return violations if isinstance(violations, list) else []
            
            return []
            
        except subprocess.CalledProcessError as e:
            logger.error(f"OPA执行失败: {e.stderr}")
            raise Exception(f"OPA执行失败: {e.stderr}")
        except Exception as e:
            logger.error(f"OPA评估出错: {e}")
            raise
        finally:
            # 清理临时文件
            for path in [rego_path, input_path]:
                if path and os.path.exists(path):
                    try:
                        os.unlink(path)
                    except Exception:
                        pass
    
    def _evaluate_assertions(self, rule: Rule, violations: List[Dict], resource_count: int) -> Dict:
        """评估断言并返回结果（使用统一的 AssertionManager）"""
        try:
            # 确保violations是列表
            if not isinstance(violations, list):
                logger.warning(f"违规结果不是列表类型: {type(violations)}")
                violations = [] if violations is None else [violations]
            
            assertion_vars = {
                'violation_count': len(violations),
                'violations': violations,
                'resource_count': resource_count
            }
            
            assertions = self.get_rule_config(rule, 'assertions', [])
            assertion_result = self.rule_processor.assertion_manager.evaluate_assertions(
                assertions, assertion_vars, mode="simple"
            )
            
            if assertion_result['passed']:
                description = assertion_result.get('pass_description', f"{rule.name}: 检查通过")
                return self._pass_result(rule, description, f"检查了 {resource_count} 个资源")
            else:
                description = assertion_result.get('fail_description', f"{rule.name}: 检查失败")
                details = self._format_violations(violations)
                return self._fail_result(
                    rule, 
                    description, 
                    details, 
                    assertion_result.get('severity', 'warning'),
                    violations
                )
        except Exception as e:
            logger.error(f"评估断言时出错: {e}")
            return self._error_result(rule, f"断言评估失败: {str(e)}")
    
    def _format_violations(self, violations: List[Dict]) -> str:
        """格式化违规信息"""
        if not violations:
            return "无违规资源"
        
        details = []
        for i, violation in enumerate(violations):
            try:
                if isinstance(violation, dict):
                    kind = violation.get('kind', 'Unknown')
                    name = violation.get('name', 'unnamed')
                    namespace = violation.get('namespace')
                    message = violation.get('message', '未知违规')
                    
                    if namespace and namespace not in ['-', '', 'null', None]:
                        detail = f"- {kind}/{name} (命名空间: {namespace}): {message}"
                    else:
                        detail = f"- {kind}/{name}: {message}"
                elif isinstance(violation, str):
                    detail = f"- {violation}"
                else:
                    # 处理其他类型的违规数据
                    detail = f"- {str(violation)}"
                
                details.append(detail)
            except Exception as e:
                logger.error(f"格式化第{i}个违规时出错: {e}, violation type: {type(violation)}")
                details.append(f"- 格式化错误的违规项: {str(violation)[:100]}")
        
        return "\n".join(details)
    
    def _pass_result(self, rule: Rule, description: str, details: str = "") -> Dict:
        """生成通过结果（已委托给 ResultFormatter）"""
        return self.rule_processor.result_formatter.pass_result(rule, description, details)
    
    def _fail_result(self, rule: Rule, description: str, details: str, 
                     severity: str = "warning", violations: List[Dict] = None) -> Dict:
        """生成失败结果（已委托给 ResultFormatter）"""
        return self.rule_processor.result_formatter.fail_result(
            rule, description, details, severity, violations
        )
    
    def _error_result(self, rule: Rule, error_msg: str) -> Dict:
        """生成错误结果（已委托给 ResultFormatter）"""
        return self.rule_processor.result_formatter.error_result(rule, error_msg)
