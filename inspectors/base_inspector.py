#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检器基类，定义了所有巡检器的通用接口和基础功能
"""

from abc import ABC, abstractmethod
import logging
from typing import Dict, List, Any, Optional, Union

from utils.inspection_result import InspectionResult
from utils.rule_loader import Rule, load_rules
from inspectors.rule_processor import RuleProcessor

# 设置日志
logger = logging.getLogger(__name__)

class BaseInspector(ABC):
    """巡检器基类，所有类型的巡检器都应继承此类"""
    
    def __init__(self, config: Dict[str, Any]):
        """
        初始化巡检器
        
        Args:
            config: 巡检器配置
        """
        self.config = config
        self.rules = []
        self.rule_processor = RuleProcessor()
        self._load_rules()
        
    @property
    @abstractmethod
    def inspector_type(self) -> str:
        """返回巡检器类型，如 'node', 'opa', 'prometheus'"""
        pass
        
    def _load_rules(self):
        """加载适用于此巡检器的规则"""
        yaml_rules = load_rules(rule_type=self.inspector_type)
        self.rules = [rule for rule in yaml_rules if rule.enabled]
        logger.info(f"加载了 {len(self.rules)} 条{self.inspector_type}巡检规则")
        
    @abstractmethod
    def _apply_rule(self, rule: Rule, context: Dict) -> Union[Dict, List[Dict], None]:
        """
        应用单条规则进行检查
        
        Args:
            rule: 要应用的规则
            context: 检查上下文
            
        Returns:
            检查结果，可以是单个结果字典，结果列表，或者None（表示规则不适用）
        """
        pass
        
    def get_rule_by_id(self, rule_id: str) -> Optional[Rule]:
        """
        根据ID获取规则
        
        Args:
            rule_id: 规则ID
            
        Returns:
            规则对象，如果未找到返回None
        """
        for rule in self.rules:
            if rule.id == rule_id:
                return rule
        return None
        
    def run_inspection(self, cluster_name: str, rule_ids: List[str] = None) -> InspectionResult:
        """
        运行巡检
        
        Args:
            cluster_name: 集群名称
            rule_ids: 要运行的规则ID列表，如果为None则运行所有规则
            
        Returns:
            巡检结果对象
        """
        result = InspectionResult(cluster_name, self.inspector_type)
        
        # 确定要运行的规则
        if rule_ids:
            active_rules = [rule for rule in self.rules if rule.id in rule_ids]
        else:
            active_rules = self.rules
            
        if not active_rules:
            return result
            
        # 执行规则
        context = self._prepare_context(cluster_name)
        
        for rule in active_rules:
            try:
                if self._should_apply_rule(rule, context):
                    inspection_result = self._apply_rule(rule, context)
                    
                    if inspection_result:
                        # 处理单个结果或结果列表
                        if isinstance(inspection_result, list):
                            for item in inspection_result:
                                result.add_item(item)
                        else:
                            result.add_item(inspection_result)
            except Exception as e:
                logger.exception(f"执行规则 {rule.id} 时出错: {str(e)}")
                error_result = self.rule_processor.format_rule_result(
                    rule=rule,
                    status='error',
                    description=f"执行规则时发生错误: {str(e)}",
                    severity='warning',
                    details=f"执行规则 {rule.name} 失败: {str(e)}",
                    solution="检查日志和系统状态"
                )
                result.add_item(error_result)
                
        return result
        
    def _prepare_context(self, cluster_name: str) -> Dict:
        """
        准备检查上下文
        
        Args:
            cluster_name: 集群名称
            
        Returns:
            检查上下文
        """
        return {'cluster_name': cluster_name}
        
    def _should_apply_rule(self, rule: Rule, context: Dict) -> bool:
        """
        判断规则是否应该应用于当前上下文
        
        Args:
            rule: 规则
            context: 上下文
            
        Returns:
            是否应用规则
        """
        # 默认实现总是返回True
        # 子类可以覆盖此方法以实现更复杂的规则过滤
        return True
        
    def get_rule_config(self, rule: Rule, path: str, default_value: Any = None) -> Any:
        """
        从规则中获取配置值，支持嵌套路径
        
        Args:
            rule: 规则对象
            path: 配置路径，使用点表示法，例如 "execution.command"
            default_value: 默认值，当路径不存在时返回
            
        Returns:
            配置值或默认值
        """
        return self.rule_processor.get_rule_config(rule, path, default_value)
