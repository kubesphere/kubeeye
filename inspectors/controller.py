#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检控制器，负责调度和协调各种类型的巡检器
"""

import logging
from typing import Dict, List, Any, Optional

from inspectors.node.node_inspector import NodeInspector
from inspectors.opa.opa_inspector import OpaInspector
from inspectors.prometheus.prometheus_inspector import PrometheusInspector
from utils.inspection_result import InspectionResult

# 设置日志
logger = logging.getLogger(__name__)

class InspectionController:
    """巡检控制器，负责协调巡检过程"""
    
    def __init__(self, config: Dict[str, Any]):
        """
        初始化巡检控制器
        
        Args:
            config: 控制器配置
        """
        self.config = config
        self.inspectors = {}
        self._initialize_inspectors()
        
    def _initialize_inspectors(self):
        """初始化所有巡检器"""
        # 初始化节点巡检器
        if 'nodes' in self.config and self.config['nodes']:
            self.inspectors['node'] = NodeInspector(self.config['nodes'])
            logger.info("已初始化节点巡检器")
            
        # 初始化OPA巡检器
        if 'opa' in self.config:
            self.inspectors['opa'] = OpaInspector(self.config['opa'])
            logger.info("已初始化OPA巡检器")
            
        # 初始化Prometheus巡检器
        if 'prometheus' in self.config:
            self.inspectors['prometheus'] = PrometheusInspector(self.config['prometheus'])
            logger.info("已初始化Prometheus巡检器")
            
        logger.info(f"已初始化 {len(self.inspectors)} 个巡检器")
        
    def get_available_inspectors(self) -> List[str]:
        """
        获取可用巡检器类型
        
        Returns:
            巡检器类型列表
        """
        return list(self.inspectors.keys())
        
    def run_inspection(self, cluster_name: str, inspector_types: List[str] = None, 
                      rule_ids: Dict[str, List[str]] = None) -> Dict[str, InspectionResult]:
        """
        运行巡检
        
        Args:
            cluster_name: 集群名称
            inspector_types: 要运行的巡检器类型列表，如果为None则运行所有巡检器
            rule_ids: 每种巡检器要运行的规则ID字典，格式为 {inspector_type: [rule_id1, rule_id2]}
            
        Returns:
            巡检结果字典，格式为 {inspector_type: inspection_result}
        """
        results = {}
        
        # 确定要运行的巡检器
        if inspector_types:
            active_inspectors = {k: v for k, v in self.inspectors.items() if k in inspector_types}
        else:
            active_inspectors = self.inspectors
            
        if not active_inspectors:
            logger.warning("没有可用的巡检器")
            return results
            
        # 执行巡检
        for inspector_type, inspector in active_inspectors.items():
            try:
                # 获取此巡检器要运行的规则ID
                inspector_rule_ids = None
                if rule_ids and inspector_type in rule_ids:
                    inspector_rule_ids = rule_ids[inspector_type]
                    
                logger.info(f"运行 {inspector_type} 巡检...")
                result = inspector.run_inspection(cluster_name, inspector_rule_ids)
                results[inspector_type] = result
                logger.info(f"{inspector_type} 巡检完成，发现 {len(result.items)} 个结果")
                
            except Exception as e:
                logger.exception(f"运行 {inspector_type} 巡检时出错: {str(e)}")
                
        return results
        
    def save_inspection_result(self, all_results: Dict[str, InspectionResult], 
                             cluster_name: str, inspection_type: str = "immediate") -> str:
        """
        保存巡检结果到文件
        
        Args:
            all_results: 巡检结果字典
            cluster_name: 集群名称
            inspection_type: 巡检类型 ('immediate' 或 'scheduled')
            
        Returns:
            保存的文件路径
        """
        import os
        import json
        from datetime import datetime
        from pathlib import Path
        
        # 确保results目录存在
        results_dir = Path("data/results")
        results_dir.mkdir(parents=True, exist_ok=True)
        
        # 生成结果ID和文件名
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        result_id = f"{inspection_type}_{timestamp}"
        filename = f"inspection_result_{cluster_name}_{timestamp}.json"
        result_path = results_dir / filename
        
        # 计算统计信息
        total_items = 0
        total_passed = 0
        total_failed = 0
        total_warning = 0
        total_error = 0
        
        # 序列化巡检结果
        serialized_results = {}
        for inspector_type, result in all_results.items():
            if hasattr(result, 'items'):
                items = result.items
            else:
                items = []
                
            # 计算统计 - 安全地访问item属性
            total_items += len(items)
            for item in items:
                # 安全地获取状态
                if hasattr(item, '__dict__'):
                    status = getattr(item, 'status', 'unknown')
                elif isinstance(item, dict):
                    status = item.get('status', 'unknown')
                elif isinstance(item, (tuple, list)) and len(item) > 1:
                    status = str(item[1])
                else:
                    status = 'unknown'
                
                # 统计各状态 - 使用简化的状态体系
                if status == 'passed':
                    total_passed += 1
                else:
                    # 所有非通过状态都视为异常
                    # 根据旧状态映射进行兼容性处理
                    if status in ['failed', 'error']:
                        total_failed += 1
                    elif status == 'warning':
                        total_warning += 1
                    else:
                        # 未知状态按错误处理
                        total_error += 1
            
            # 序列化items - 确保所有items都是字典格式
            serialized_items = []
            for item in items:
                if hasattr(item, '__dict__'):
                    # 如果是对象，转换为字典
                    serialized_items.append(item.__dict__)
                elif isinstance(item, dict):
                    # 如果已经是字典，直接使用
                    serialized_items.append(item)
                elif isinstance(item, (tuple, list)) and len(item) >= 2:
                    # 如果是元组或列表，尝试转换为基本字典格式
                    serialized_items.append({
                        'name': str(item[0]) if len(item) > 0 else 'Unknown',
                        'status': str(item[1]) if len(item) > 1 else 'unknown',
                        'description': str(item[2]) if len(item) > 2 else '',
                        'severity': 'info',
                        'details': str(item),
                        'solution': ''
                    })
                else:
                    # 其他情况，创建基本字典
                    serialized_items.append({
                        'name': str(item),
                        'status': 'unknown',
                        'description': f'Converted from {type(item).__name__}',
                        'severity': 'info',
                        'details': str(item),
                        'solution': ''
                    })
            
            serialized_results[inspector_type] = {
                "inspector_type": inspector_type,
                "items": serialized_items
            }
        
        # 构建完整的结果结构
        result_data = {
            "result_id": result_id,
            "cluster_name": cluster_name,
            "timestamp": datetime.now().isoformat(),
            "inspection_type": inspection_type,  # immediate 或 scheduled
            "execution_info": {
                "triggered_by": "user" if inspection_type == "immediate" else "scheduler",
                "inspectors_used": list(all_results.keys()),
                "execution_duration": "N/A"  # 可以后续添加计时功能
            },
            "inspection_results": serialized_results,
            "summary": {
                "total_items": total_items,
                "passed": total_passed,
                "failed": total_failed,
                "warning": total_warning,
                "error": total_error
            },
            # 兼容旧版本的字段
            "critical": total_failed,  # 将failed映射为critical以兼容现有代码
            "warning": total_warning,
            "passed": total_passed
        }
        
        # 保存到文件
        with open(result_path, 'w', encoding='utf-8') as f:
            json.dump(result_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"巡检结果已保存到: {result_path}")
        return str(result_path)
