#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Prometheus 巡检器 - 简化版本，支持"一个规则=一个query=一个巡检项"
"""

import logging
import datetime
from typing import Dict, List, Any, Optional, Tuple

from inspectors.base_inspector import BaseInspector
from utils.prometheus_client import PrometheusClient
from utils.rule_loader import Rule

# 设置日志
logger = logging.getLogger(__name__)

class PrometheusInspector(BaseInspector):
    """
    Prometheus规则巡检器 - 简化版本
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        初始化Prometheus巡检器
        
        Args:
            config: Prometheus配置字典，应包含以下字段：
                - url: Prometheus服务器URL
                - username: 可选的用户名
                - password: 可选的密码
                - token: 可选的访问令牌
                - enabled: 是否启用
        """
        # 确保配置是有效的
        if not isinstance(config, dict):
            raise TypeError("配置必须是一个字典")
        
        # 如果传入的配置缺少必要的字段，则添加默认值
        if 'url' not in config:
            raise ValueError("Prometheus配置缺少url字段")
        
        # 创建PrometheusClient实例
        self.prometheus_client = PrometheusClient(config)
        
        # 调用父类初始化
        super().__init__(config)
    
    @property
    def inspector_type(self) -> str:
        return "prometheus"
    
    def _prepare_context(self, cluster_name: str) -> Dict:
        """
        准备Prometheus巡检上下文
        
        Args:
            cluster_name: 集群名称
            
        Returns:
            准备好的上下文
        """
        context = super()._prepare_context(cluster_name)
        return context
    
    def _validate_rule_config(self, rule: Rule) -> List[str]:
        """
        验证规则配置是否有效
        
        Args:
            rule: 规则对象
            
        Returns:
            配置问题列表，如果没有问题则为空列表
        """
        issues = []
        
        # 检查必要的查询配置
        query = self.get_rule_config(rule, 'query', '')
        if not query:
            issues.append("缺少必要的Prometheus查询(query)")
        
        # 检查必要的断言配置
        assertions = self.get_rule_config(rule, 'assertions', [])
        if not assertions:
            issues.append("缺少必要的断言配置(assertions)")
            
        return issues

    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        """
        应用Prometheus规则进行检查 - 简化版本
        
        Args:
            rule: 规则对象
            context: 上下文
            
        Returns:
            检查结果
        """
        # 获取查询和断言配置
        query = self.get_rule_config(rule, 'query', '')
        assertions = self.get_rule_config(rule, 'assertions', [])
            
        # 执行查询
        try:
            # 执行即时查询（简化版本，不再支持复杂的时间范围查询）
            result = self.prometheus_client.query(query)
                
            # 处理结果
            metrics = self._process_query_result(result)
            if not metrics:
                return self.rule_processor.format_rule_result(
                    rule=rule,
                    status="passed",
                    description=f"{rule.name}: 无数据",
                    severity="info",
                    details="Prometheus查询没有返回匹配的指标数据",
                    solution=""
                )
                
            # 提取数据进行断言评估
            variables = self._extract_metrics_variables(metrics)
            
            # 生成包含上下文信息的名称后缀
            name_suffix = self._generate_context_suffix(metrics, variables)
            
            # 评估断言
            assertion_result = self.rule_processor.evaluate_assertions(assertions, variables)
            
            # 根据断言结果返回检查结果
            if assertion_result['passed']:
                # 对于通过的检查，在描述中显示具体的监控数据
                first_assertion = assertions[0] if assertions else {}
                first_assertion_desc = first_assertion.get('description', '')
                if first_assertion_desc:
                    # 渲染模板以显示具体值
                    from utils.assertion_evaluator import AssertionEvaluator
                    evaluator = AssertionEvaluator()
                    rendered_desc = evaluator._render_template(first_assertion_desc, variables)
                    description = f"{rule.name}: {rendered_desc}"
                else:
                    # 显示关键指标值
                    if 'max_value' in variables:
                        description = f"{rule.name}: 最大值 {variables['max_value']:.2f}"
                    elif 'value' in variables:
                        description = f"{rule.name}: 当前值 {variables['value']:.2f}"
                    else:
                        description = f"{rule.name}: 检查通过"
                
                result = self.rule_processor.format_rule_result(
                    rule=rule,
                    status="passed",
                    description=description,
                    severity="info",
                    details="监控指标正常",
                    solution=""
                )
            else:
                # 移除"断言失败"前缀，直接使用描述
                clean_description = assertion_result['description'].replace("断言失败: ", "")
                
                result = self.rule_processor.format_rule_result(
                    rule=rule,
                    status="failed",
                    description=clean_description,
                    severity=assertion_result['severity'],
                    details=f"监控告警触发\n查询: {query}\n结果值: {self._format_simple_metrics(metrics)}",
                    solution=rule.solution
                )
            
            # 添加上下文信息到名称
            if name_suffix:
                result['name'] = f"{rule.name} - {name_suffix}"
            
            return result
                
        except Exception as e:
            logger.exception(f"执行Prometheus查询时出错: {str(e)}")
            return self._format_error_result(rule,
                "执行Prometheus查询失败", str(e))
    
    def _process_query_result(self, result: Dict) -> List[Dict]:
        """
        处理Prometheus查询结果 - 简化版本，只支持vector类型
        
        Args:
            result: Prometheus查询结果
            
        Returns:
            处理后的指标列表
        """
        metrics = []
        
        # 检查结果格式
        if not result or not isinstance(result, dict):
            return metrics
            
        data = result.get('data', {})
        result_type = data.get('resultType')
        result_data = data.get('result', [])
        
        if not result_data:
            return metrics
            
        # 只处理即时查询结果（vector类型）
        if result_type == 'vector':
            for item in result_data:
                metric = {
                    'metric': item.get('metric', {}),
                    'value': float(item.get('value', [0, '0'])[1]) if item.get('value') else 0
                }
                metrics.append(metric)
        else:
            # 不再支持matrix类型的复杂时间序列数据
            logger.warning(f"不支持的查询结果类型: {result_type}，请使用即时查询")
        
        return metrics
    
    def _extract_metrics_variables(self, metrics: List[Dict]) -> Dict[str, Any]:
        """
        从指标中提取变量用于断言评估
        
        Args:
            metrics: 指标列表
            
        Returns:
            变量字典
        """
        variables = {
            # 存储所有值的列表，方便计算平均值、最大值等
            'values': [m.get('value', 0) for m in metrics],
        }
        
        # 如果只有一个指标，直接使用其值
        if len(metrics) == 1:
            variables['value'] = metrics[0].get('value', 0)
            
            # 添加标签作为变量
            metric_labels = metrics[0].get('metric', {})
            for label, label_value in metric_labels.items():
                variables[f"label_{label}"] = label_value
        
        # 添加聚合值
        if variables['values']:
            variables['max_value'] = max(variables['values'])
            variables['min_value'] = min(variables['values'])
            variables['avg_value'] = sum(variables['values']) / len(variables['values'])
            
        return variables
    
    def _generate_context_suffix(self, metrics: List[Dict], variables: Dict[str, Any]) -> str:
        """
        生成包含上下文信息的名称后缀
        
        Args:
            metrics: 指标列表
            variables: 变量字典
            
        Returns:
            上下文后缀字符串
        """
        if not metrics:
            return ""
        
        # 如果只有一个指标，尝试提取有意义的标签
        if len(metrics) == 1:
            metric_labels = metrics[0].get('metric', {})
            
            # 优先显示节点相关信息
            if 'instance' in metric_labels:
                instance = metric_labels['instance']
                # 清理instance格式 (通常是IP:PORT或hostname:PORT)
                if ':' in instance:
                    instance = instance.split(':')[0]
                return f"节点 {instance}"
            elif 'node' in metric_labels:
                return f"节点 {metric_labels['node']}"
            elif 'job' in metric_labels:
                return f"作业 {metric_labels['job']}"
            elif '__name__' in metric_labels:
                return f"指标 {metric_labels['__name__']}"
        
        # 如果有多个指标，显示指标数量
        elif len(metrics) > 1:
            # 尝试找到共同的标签
            first_metric_labels = metrics[0].get('metric', {})
            if 'job' in first_metric_labels:
                job_name = first_metric_labels['job']
                return f"{len(metrics)}个{job_name}实例"
            else:
                return f"{len(metrics)}个实例"
        
        return ""
    
    def _format_simple_metrics(self, metrics: List[Dict]) -> str:
        """
        简化的指标格式化方法
        
        Args:
            metrics: 指标列表
            
        Returns:
            格式化的指标字符串
        """
        if not metrics:
            return "无数据"
        
        if len(metrics) == 1:
            metric = metrics[0]
            value = metric.get('value', 'N/A')
            return f"当前值: {value}"
        else:
            values = [m.get('value', 0) for m in metrics]
            max_val = max(values)
            avg_val = sum(values) / len(values)
            return f"最大值: {max_val:.2f}, 平均值: {avg_val:.2f}, 共{len(metrics)}个实例"
    
    def get_rule_config(self, rule: Rule, key: str, default: Any = None) -> Any:
        """
        从规则配置中获取特定键的值
        
        Args:
            rule: 规则对象
            key: 配置键
            default: 默认值，如果键不存在则返回此值
            
        Returns:
            配置值或默认值
        """
        if rule.config and key in rule.config:
            return rule.config[key]
        return default
        
    def _format_skipped_result(self, rule: Rule, reason: str) -> Dict:
        """
        格式化跳过的规则结果
        
        Args:
            rule: 规则对象
            reason: 跳过原因
            
        Returns:
            结果字典
        """
        return self.rule_processor.format_rule_result(
            rule=rule,
            status="skipped",
            description=f"{rule.name} 已跳过: {reason}",
            severity="info",
            details=reason,
            solution=""
        )
        
    def _format_error_result(self, rule: Rule, error_type: str, error_msg: str) -> Dict:
        """
        格式化错误结果
        
        Args:
            rule: 规则对象
            error_type: 错误类型
            error_msg: 错误消息
            
        Returns:
            结果字典
        """
        return self.rule_processor.format_rule_result(
            rule=rule,
            status="error",
            description=f"{rule.name} 出错: {error_type}",
            severity="critical",
            details=f"错误类型: {error_type}\n错误信息: {error_msg}",
            solution="请检查Prometheus连接配置和查询语法"
        )
