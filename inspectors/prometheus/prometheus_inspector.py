#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Prometheus 指标巡检器，继承自基类实现，使用规则处理器
"""
import logging
import re
from typing import Dict, List, Any, Optional, Tuple

from inspectors.base_inspector import BaseInspector
from utils.prometheus_client import PrometheusClient
from utils.rule_loader import Rule

# 设置日志
logger = logging.getLogger(__name__)

class PrometheusInspector(BaseInspector):
    """Prometheus 指标巡检器"""
    
    def __init__(self, config: Dict):
        """
        初始化Prometheus巡检器
        
        Args:
            config: 配置字典，必须包含 prometheus_config 键
        """
        super().__init__(config)
        
        # 获取Prometheus客户端配置
        prometheus_config = config.get("prometheus_config", {})
        
        # 创建Prometheus客户端
        prometheus_config.setdefault("url", "http://localhost:9090")
        self.client = PrometheusClient(prometheus_config)
    
    @property
    def inspector_type(self) -> str:
        """获取巡检器类型"""
        return "prometheus"
    
    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        """
        应用规则检查 Prometheus 指标
        
        Args:
            rule: 要应用的规则
            context: 上下文
            
        Returns:
            检查结果
        """
        # 获取查询
        query = self._get_query(rule)
        if not query:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='skipped',
                description="缺少查询语句",
                severity='info',
                details="规则未定义 Prometheus 查询语句",
                solution="编辑规则，添加有效的 PromQL 查询"
            )
        
        # 执行查询
        success, result = self._execute_query(query)
        if not success:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='error',
                description="执行查询失败",
                severity='warning',
                details=f"执行 PromQL 查询时出错: {result}",
                solution="检查 Prometheus 服务器状态和查询语法"
            )
        
        # 解析结果
        metrics = self._parse_result(result)
        if not metrics:
            return self.rule_processor.format_rule_result(
                rule=rule,
                status='passed',  # 假设无数据意味着没有问题
                description="无指标数据",
                severity='info',
                details=f"查询 '{query}' 未返回任何结果",
                solution=""
            )
        
        # 评估结果
        return self._evaluate_metrics(rule, metrics, query)
    
    def _get_query(self, rule: Rule) -> str:
        """
        获取规则的Prometheus查询语句
        
        Args:
            rule: 规则对象
            
        Returns:
            查询语句字符串
        """
        # 只从新格式中获取查询
        return self.get_rule_config(rule, 'query.promql') or ''
    
    def _execute_query(self, query: str) -> Tuple[bool, Any]:
        """
        执行Prometheus查询
        
        Args:
            query: PromQL查询语句
            
        Returns:
            (成功标志, 结果数据)
        """
        try:
            result = self.client.query(query)
            return True, result
        except Exception as e:
            logger.exception(f"执行Prometheus查询时出错: {str(e)}")
            return False, str(e)
    
    def _parse_result(self, result: Dict) -> Dict:
        """
        解析Prometheus查询结果
        
        Args:
            result: 查询结果
            
        Returns:
            解析后的指标数据
        """
        metrics = {}
        
        try:
            # 处理瞬时查询结果 (instant query)
            if 'data' in result and 'result' in result['data']:
                for item in result['data']['result']:
                    metric_key = self._get_metric_key(item.get('metric', {}))
                    
                    # 获取值
                    if 'value' in item:
                        try:
                            # value is [timestamp, value_string]
                            timestamp, value_str = item['value']
                            value = float(value_str)
                            metrics[metric_key] = {'value': value, 'labels': item.get('metric', {})}
                        except (ValueError, TypeError, IndexError) as e:
                            logger.error(f"解析指标值时出错: {str(e)}")
                    
                    # 处理区间查询结果
                    elif 'values' in item:
                        try:
                            # 取最新的值
                            if item['values']:
                                latest = item['values'][-1]
                                timestamp, value_str = latest
                                value = float(value_str)
                                metrics[metric_key] = {'value': value, 'labels': item.get('metric', {})}
                        except (ValueError, TypeError, IndexError) as e:
                            logger.error(f"解析区间指标值时出错: {str(e)}")
        except Exception as e:
            logger.exception(f"解析Prometheus结果时出错: {str(e)}")
        
        return metrics
    
    def _get_metric_key(self, labels: Dict) -> str:
        """
        从标签中获取指标键名
        
        Args:
            labels: 指标标签
            
        Returns:
            指标键名
        """
        if '__name__' in labels:
            base_name = labels['__name__']
        else:
            base_name = "unknown_metric"
            
        if labels:
            label_strs = [f'{k}="{v}"' for k, v in labels.items() if k != '__name__']
            if label_strs:
                return f"{base_name}{{{','.join(label_strs)}}}"
        
        return base_name
    
    def _evaluate_metrics(self, rule: Rule, metrics: Dict, query: str) -> Dict:
        """
        评估指标
        
        Args:
            rule: 规则对象
            metrics: 指标数据
            query: 执行的查询
            
        Returns:
            检查结果
        """
        # 获取阈值
        thresholds = self._get_thresholds(rule)
        
        # 获取比较运算符
        comparator = self.get_rule_config(rule, 'query.comparator', '>')
        
        # 记录所有违规指标
        violations = []
        warning_count = 0
        critical_count = 0
        
        # 评估每个指标
        for key, metric_data in metrics.items():
            value = metric_data.get('value')
            labels = metric_data.get('labels', {})
            
            # 跳过没有值的指标
            if value is None:
                continue
                
            # 检查是否超过警告阈值
            is_warning = False
            is_critical = False
            
            if 'warning' in thresholds and thresholds['warning'] is not None:
                # 使用比较操作符进行比较
                if self._compare_value(value, thresholds['warning'], comparator):
                    is_warning = True
                    warning_count += 1
            
            # 检查是否超过严重阈值
            if 'critical' in thresholds and thresholds['critical'] is not None:
                if self._compare_value(value, thresholds['critical'], comparator):
                    is_critical = True
                    critical_count += 1
            
            # 如果有违规，添加到列表
            if is_warning or is_critical:
                violation = {
                    'metric': key,
                    'value': value,
                    'labels': labels,
                    'severity': 'critical' if is_critical else 'warning'
                }
                violations.append(violation)
        
        # 确定整体状态和严重性
        if critical_count > 0:
            status = 'fail'
            severity = 'critical'
            description = f"{rule.name} - {critical_count} 个指标超过严重阈值"
        elif warning_count > 0:
            status = 'warn'
            severity = 'warning'
            description = f"{rule.name} - {warning_count} 个指标超过警告阈值"
        else:
            status = 'pass'
            severity = 'info'
            description = f"{rule.name} - 指标正常"
        
        # 生成详细信息
        if violations:
            details = f"查询: {query}\n\n"
            details += f"警告阈值: {thresholds.get('warning')}, 严重阈值: {thresholds.get('critical')}, 比较运算符: {comparator}\n\n"
            
            # 添加违规详情
            details += "违规指标:\n"
            for idx, v in enumerate(violations[:10]):  # 仅显示前10个
                details += f"{idx+1}. {v['metric']} = {v['value']} ({v['severity']})\n"
                
            if len(violations) > 10:
                details += f"\n...共 {len(violations)} 个违规指标"
        else:
            details = f"查询: {query}\n\n所有指标均符合阈值要求"
        
        # 返回结果
        return self.rule_processor.format_rule_result(
            rule=rule,
            status=status,
            description=description,
            severity=severity,
            details=details,
            solution=rule.solution if hasattr(rule, 'solution') else "根据指标情况优化系统配置"
        )
    
    def _get_thresholds(self, rule: Rule) -> Dict:
        """
        获取规则的阈值配置
        
        Args:
            rule: 规则对象
            
        Returns:
            阈值配置字典
        """
        # 创建一个包含warning和critical阈值的字典
        thresholds = {}
        
        # 直接从rule.thresholds获取
        rule_thresholds = getattr(rule, 'thresholds', {}) or {}
        
        # 提取warning阈值
        warning = rule_thresholds.get('warning')
        if warning is not None:
            thresholds['warning'] = warning
            
        # 提取critical阈值
        critical = rule_thresholds.get('critical')
        if critical is not None:
            thresholds['critical'] = critical
            
        return thresholds
    
    def _compare_value(self, value: float, threshold: float, comparator: str) -> bool:
        """
        比较值和阈值
        
        Args:
            value: 要比较的值
            threshold: 阈值
            comparator: 比较运算符
            
        Returns:
            比较结果
        """
        try:
            if comparator == '>':
                return value > threshold
            elif comparator == '>=':
                return value >= threshold
            elif comparator == '<':
                return value < threshold
            elif comparator == '<=':
                return value <= threshold
            elif comparator == '==':
                return value == threshold
            else:
                # 默认使用大于
                return value > threshold
        except (TypeError, ValueError) as e:
            logger.error(f"比较值和阈值时出错: {str(e)}")
            return False
