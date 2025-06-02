#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
统一的巡检控制器，负责协调不同类型的巡检器执行
"""

import logging
import time
from datetime import datetime
from typing import Dict, List, Any, Optional

# 导入巡检器
from inspectors.base_inspector import BaseInspector
from inspectors.node.node_inspector import NodeInspector
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from inspectors.opa.opa_inspector import OpaInspector

from utils.inspection_result import InspectionResult

# 设置日志
logger = logging.getLogger(__name__)

class InspectionController:
    """
    巡检控制器，协调多种巡检器执行，管理巡检结果
    """
    
    def __init__(self, config: Dict[str, Any]):
        """
        初始化巡检控制器
        
        Args:
            config: 巡检控制器配置
        """
        self.config = config
        self.inspectors = {}
        self._initialize_inspectors()
    
    def _initialize_inspectors(self):
        """初始化所有巡检器"""
        # 初始化节点巡检器
        if self.config.get('enable_node_inspection', True) and 'nodes' in self.config:
            try:
                self.inspectors['node'] = NodeInspector(self.config['nodes'])
                logger.info("节点巡检器已初始化")
            except Exception as e:
                logger.error(f"初始化节点巡检器失败: {str(e)}")
        
        # 初始化Prometheus巡检器
        if self.config.get('enable_prometheus_inspection', True) and 'prometheus' in self.config:
            try:
                self.inspectors['prometheus'] = PrometheusInspector(self.config['prometheus'])
                logger.info("Prometheus巡检器已初始化")
            except Exception as e:
                logger.error(f"初始化Prometheus巡检器失败: {str(e)}")
        
        # 初始化OPA巡检器
        if self.config.get('enable_opa_inspection', True) and 'opa' in self.config:
            try:
                self.inspectors['opa'] = OpaInspector(self.config['opa'])
                logger.info("OPA巡检器已初始化")
            except Exception as e:
                logger.error(f"初始化OPA巡检器失败: {str(e)}")
    
    def get_available_inspectors(self) -> List[str]:
        """
        获取可用的巡检器类型
        
        Returns:
            巡检器类型列表
        """
        return list(self.inspectors.keys())
    
    def get_inspector(self, inspector_type: str) -> Optional[BaseInspector]:
        """
        获取指定类型的巡检器
        
        Args:
            inspector_type: 巡检器类型
            
        Returns:
            巡检器实例，如果不存在则返回None
        """
        return self.inspectors.get(inspector_type)
    
    def run_inspection(self, cluster_name: str, inspector_types: List[str] = None, rule_ids: Dict[str, List[str]] = None) -> Dict[str, InspectionResult]:
        """
        运行指定类型的巡检
        
        Args:
            cluster_name: 集群名称
            inspector_types: 要运行的巡检器类型列表，None表示所有可用巡检器
            rule_ids: 各巡检器要运行的规则ID映射，格式为 {inspector_type: [rule_id1, rule_id2, ...]}
            
        Returns:
            巡检结果映射，格式为 {inspector_type: InspectionResult}
        """
        results = {}
        start_time = time.time()
        
        # 确定要运行的巡检器类型
        types_to_run = inspector_types if inspector_types else list(self.inspectors.keys())
        logger.info(f"将对集群 {cluster_name} 执行以下类型的巡检: {', '.join(types_to_run)}")
        
        # 运行每种类型的巡检
        for inspector_type in types_to_run:
            inspector = self.inspectors.get(inspector_type)
            if not inspector:
                logger.warning(f"巡检器类型 '{inspector_type}' 不可用，已跳过")
                continue
            
            try:
                # 获取此类型巡检器的规则ID列表
                type_rule_ids = None
                if rule_ids and inspector_type in rule_ids:
                    type_rule_ids = rule_ids[inspector_type]
                
                # 执行巡检
                logger.info(f"开始执行 {inspector_type} 巡检...")
                result = inspector.run_inspection(cluster_name, type_rule_ids)
                results[inspector_type] = result
                logger.info(f"{inspector_type} 巡检完成，共 {len(result.items)} 条结果")
                
            except Exception as e:
                logger.exception(f"执行 {inspector_type} 巡检时出错: {str(e)}")
        
        end_time = time.time()
        duration = end_time - start_time
        logger.info(f"巡检完成，耗时 {duration:.2f} 秒")
        
        return results
    
    def run_rule(self, cluster_name: str, rule_id: str) -> Optional[Dict]:
        """
        运行单个规则
        
        Args:
            cluster_name: 集群名称
            rule_id: 规则ID
            
        Returns:
            规则执行结果，如果规则不存在则返回None
        """
        # 查找包含此规则的巡检器
        for inspector_type, inspector in self.inspectors.items():
            if inspector.get_rule_by_id(rule_id):
                # 找到包含规则的巡检器，执行此规则
                result = inspector.run_inspection(cluster_name, [rule_id])
                if result.items:
                    return result.items[0]
                return None
        
        logger.warning(f"未找到规则 {rule_id}")
        return None
    
    def get_all_rules(self) -> Dict[str, List[Dict]]:
        """
        获取所有规则信息
        
        Returns:
            按巡检器类型分组的规则信息
        """
        all_rules = {}
        
        for inspector_type, inspector in self.inspectors.items():
            rules_info = []
            for rule in inspector.rules:
                info = {
                    'id': rule.id,
                    'name': rule.name,
                    'description': rule.description if hasattr(rule, 'description') else '',
                    'severity': rule.severity if hasattr(rule, 'severity') else 'unknown',
                    'enabled': rule.enabled if hasattr(rule, 'enabled') else True,
                    'tags': rule.tags if hasattr(rule, 'tags') else []
                }
                rules_info.append(info)
            
            all_rules[inspector_type] = rules_info
        
        return all_rules
    
    def save_inspection_result(self, results: Dict[str, InspectionResult], cluster_name: str) -> str:
        """
        保存巡检结果
        
        Args:
            results: 巡检结果映射
            cluster_name: 集群名称
            
        Returns:
            保存路径
        """
        import os
        import json
        
        # 创建结果目录
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        base_dir = os.path.join(self.config.get('results_dir', 'data/results'), cluster_name)
        result_dir = os.path.join(base_dir, timestamp)
        os.makedirs(result_dir, exist_ok=True)
        
        # 保存总结果
        summary = {
            'cluster_name': cluster_name,
            'timestamp': timestamp,
            'inspectors': list(results.keys()),
            'total_items': sum(len(result.items) for result in results.values()),
            'status': {
                'passed': sum(sum(1 for item in result.items if item.get('status') == 'passed') for result in results.values()),
                'failed': sum(sum(1 for item in result.items if item.get('status') == 'failed') for result in results.values()),
                'warning': sum(sum(1 for item in result.items if item.get('status') == 'warning') for result in results.values()),
                'error': sum(sum(1 for item in result.items if item.get('status') == 'error') for result in results.values()),
                'skipped': sum(sum(1 for item in result.items if item.get('status') == 'skipped') for result in results.values()),
                'unknown': sum(sum(1 for item in result.items if item.get('status') not in ['passed', 'failed', 'warning', 'error', 'skipped']) for result in results.values())
            },
            'inspector_results': {}
        }
        
        # 保存每个巡检器的结果
        for inspector_type, result in results.items():
            # 保存结果到单独文件
            type_file = os.path.join(result_dir, f"{inspector_type}.json")
            with open(type_file, 'w') as f:
                json.dump(result.to_dict(), f, ensure_ascii=False, indent=2)
            
            # 添加到总结果
            summary['inspector_results'][inspector_type] = {
                'total_items': len(result.items),
                'status': {
                    'passed': sum(1 for item in result.items if item.get('status') == 'passed'),
                    'failed': sum(1 for item in result.items if item.get('status') == 'failed'),
                    'warning': sum(1 for item in result.items if item.get('status') == 'warning'),
                    'error': sum(1 for item in result.items if item.get('status') == 'error'),
                    'skipped': sum(1 for item in result.items if item.get('status') == 'skipped'),
                    'unknown': sum(1 for item in result.items if item.get('status') not in ['passed', 'failed', 'warning', 'error', 'skipped'])
                }
            }
        
        # 保存总结果
        summary_file = os.path.join(result_dir, "summary.json")
        with open(summary_file, 'w') as f:
            json.dump(summary, f, ensure_ascii=False, indent=2)
        
        # 保存最新结果的链接
        latest_link = os.path.join(base_dir, "latest")
        if os.path.exists(latest_link):
            if os.path.islink(latest_link):
                os.unlink(latest_link)
            else:
                os.remove(latest_link)
        
        # 在Unix系统上创建符号链接
        if hasattr(os, 'symlink'):
            os.symlink(timestamp, latest_link)
        
        logger.info(f"巡检结果已保存至: {result_dir}")
        return result_dir
    
    def generate_report(self, results: Dict[str, InspectionResult], cluster_name: str, report_format: str = 'html') -> str:
        """
        生成巡检报告
        
        Args:
            results: 巡检结果映射
            cluster_name: 集群名称
            report_format: 报告格式，支持 'html', 'pdf', 'markdown'
            
        Returns:
            报告文件路径
        """
        # TODO: 实现报告生成逻辑
        # 这里可以根据不同格式生成报告
        pass
