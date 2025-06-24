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
        logger.info(f"🔍 BaseInspector.run_inspection 开始 - 巡检器类型: {self.inspector_type}, 集群: {cluster_name}")
        logger.info(f"可用规则总数: {len(self.rules)}, 指定规则ID: {rule_ids}")
        
        result = InspectionResult(cluster_name, self.inspector_type)
        
        # 确定要运行的规则
        if rule_ids:
            active_rules = [rule for rule in self.rules if rule.id in rule_ids]
            logger.info(f"根据指定ID筛选后的规则数: {len(active_rules)}")
        else:
            active_rules = self.rules
            logger.info(f"使用所有可用规则数: {len(active_rules)}")
            
        if not active_rules:
            logger.warning(f"没有可执行的规则，巡检结束")
            return result
        
        logger.info(f"准备执行 {len(active_rules)} 条规则:")
        for rule in active_rules:
            logger.info(f"  - {rule.id}: {rule.name}")
            
        # 执行规则
        context = self._prepare_context(cluster_name)
        logger.info(f"上下文准备完成: {context}")
        
        executed_count = 0
        for rule in active_rules:
            try:
                logger.info(f"📋 开始执行规则 {rule.id}: {rule.name}")
                
                # 验证规则配置
                validation_issues = self._validate_rule_config(rule)
                if validation_issues:
                    logger.error(f"规则 {rule.id} 配置无效: {validation_issues}")
                    # 规则配置无效
                    result.add_item(self._format_invalid_result(
                        rule, 
                        "规则配置无效", 
                        f"以下配置问题阻止了规则执行: {', '.join(validation_issues)}"
                    ))
                    continue
                else:
                    logger.info(f"规则 {rule.id} 配置验证通过")
                    
                # 检查规则是否适用当前环境    
                should_apply = self._should_apply_rule(rule, context)
                logger.info(f"规则 {rule.id} 适用性检查: {should_apply}")
                
                if should_apply:
                    logger.info(f"🎯 开始应用规则 {rule.id}")
                    inspection_result = self._apply_rule(rule, context)
                    logger.info(f"规则 {rule.id} 应用完成，结果类型: {type(inspection_result)}")
                    
                    if inspection_result:
                        # 处理单个结果或结果列表
                        if isinstance(inspection_result, list):
                            logger.info(f"规则 {rule.id} 返回结果列表，长度: {len(inspection_result)}")
                            for item in inspection_result:
                                result.add_item(item)
                        else:
                            logger.info(f"规则 {rule.id} 返回单个结果")
                            result.add_item(inspection_result)
                    else:
                        logger.warning(f"规则 {rule.id} 返回空结果")
                else:
                    logger.info(f"规则 {rule.id} 不适用于当前环境")
                    # 规则不适用于当前环境
                    result.add_item(self._format_not_applicable_result(
                        rule, 
                        "规则不适用于当前环境"
                    ))
                
                executed_count += 1
                logger.info(f"✅ 规则 {rule.id} 执行完成 ({executed_count}/{len(active_rules)})")
                
            except Exception as e:
                logger.exception(f"❌ 执行规则 {rule.id} 时出错: {str(e)}")
                error_result = self._format_error_result(
                    rule,
                    f"执行规则时发生错误: {str(e)}",
                    str(e)
                )
                result.add_item(error_result)
        
        logger.info(f"🏁 BaseInspector.run_inspection 完成 - 巡检器: {self.inspector_type}, 执行规则数: {executed_count}, 结果数: {len(result.items)}")        
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
        
    def _validate_rule_config(self, rule: Rule) -> List[str]:
        """
        验证规则配置是否有效
        
        Args:
            rule: 规则对象
            
        Returns:
            配置问题列表，如果没有问题则为空列表
        """
        # 子类应该重写此方法以实现特定规则类型的验证逻辑
        return []
        
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
        
    def _format_invalid_result(self, rule: Rule, description: str, details: str) -> Dict:
        """
        格式化配置无效的规则结果（已委托给 ResultFormatter）
        
        Args:
            rule: 规则对象
            description: 简要描述
            details: 详细信息
            
        Returns:
            格式化的结果字典
        """
        return self.rule_processor.result_formatter.invalid_result(rule, description, details)
        
    def _format_not_applicable_result(self, rule: Rule, reason: str) -> Dict:
        """
        格式化不适用规则结果（已委托给 ResultFormatter）
        
        Args:
            rule: 规则对象
            reason: 不适用原因
            
        Returns:
            格式化的结果字典
        """
        return self.rule_processor.result_formatter.not_applicable_result(rule, reason)
        
    def _format_skipped_result(self, rule: Rule, reason: str) -> Dict:
        """
        格式化跳过的规则结果（已委托给 ResultFormatter）
        
        Args:
            rule: 规则对象
            reason: 跳过原因
            
        Returns:
            格式化的结果字典
        """
        return self.rule_processor.result_formatter.skipped_result(rule, reason)
        
    def _format_error_result(self, rule: Rule, description: str, error: str) -> Dict:
        """
        格式化错误的规则结果（已委托给 ResultFormatter）
        
        Args:
            rule: 规则对象
            description: 错误描述
            error: 错误详情
            
        Returns:
            格式化的结果字典
        """
        return self.rule_processor.result_formatter.error_result(rule, error, description)
