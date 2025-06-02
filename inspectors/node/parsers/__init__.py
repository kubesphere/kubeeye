#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
解析器模块，用于处理不同类型的命令输出
"""
import logging
import importlib
import os
import sys
import inspect
import pkgutil
from pathlib import Path
from typing import Dict, Any, Callable, Type, Optional, List, Union, TypeVar

# 设置日志
logger = logging.getLogger(__name__)

# 定义类型
T_BaseParser = TypeVar('T_BaseParser', bound='BaseParser')

# 全局解析器注册表
_PARSER_REGISTRY = {}

class BaseParser:
    """解析器基类，所有自定义解析器应该继承此类"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """
        解析命令输出并返回检查结果
        
        Args:
            stdout: 命令输出
            rule: 规则对象
            node: 节点信息
            extra_data: 额外数据，比如解析器配置
            
        Returns:
            检查结果字典
        """
        raise NotImplementedError("子类必须实现parse方法")

def register_parser(name: str):
    """
    用于注册解析器的装饰器
    
    Args:
        name: 解析器名称
    """
    def wrapper(parser_class):
        if name in _PARSER_REGISTRY:
            logger.warning(f"解析器 '{name}' 已存在，将被覆盖")
        _PARSER_REGISTRY[name] = parser_class
        logger.debug(f"已注册解析器: {name}")
        return parser_class
    return wrapper

def register_parser_class(name: str, parser_class: Type[BaseParser]) -> None:
    """
    直接注册解析器类
    
    Args:
        name: 解析器名称
        parser_class: 解析器类
    """
    if name in _PARSER_REGISTRY:
        logger.warning(f"解析器 '{name}' 已存在，将被覆盖")
    _PARSER_REGISTRY[name] = parser_class
    logger.info(f"已注册解析器: {name}")

def list_parsers() -> List[str]:
    """
    列出所有已注册的解析器名称
    
    Returns:
        解析器名称列表
    """
    return list(_PARSER_REGISTRY.keys())

def get_parser(name: str) -> Optional[Type[BaseParser]]:
    """
    获取指定名称的解析器
    
    Args:
        name: 解析器名称
        
    Returns:
        解析器类或None
    """
    return _PARSER_REGISTRY.get(name)

def parse_output(parser_name: str, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Optional[Dict]:
    """
    使用指定解析器解析命令输出
    
    Args:
        parser_name: 解析器名称
        stdout: 命令输出
        rule: 规则对象
        node: 节点信息
        extra_data: 额外数据，比如解析器配置
        
    Returns:
        解析结果或None
    """
    parser_class = get_parser(parser_name)
    if not parser_class:
        logger.warning(f"解析器 {parser_name} 不存在")
        return None
        
    try:
        # 检查解析器方法的签名
        import inspect
        sig = inspect.signature(parser_class.parse)
        parameters = list(sig.parameters.values())
        
        # 第一个参数总是 cls，我们需要确定其余参数
        required_params = []
        for param in parameters[1:]:  # 跳过cls参数
            if param.default == param.empty and param.kind not in (param.VAR_POSITIONAL, param.VAR_KEYWORD):
                # 这是一个必需参数
                required_params.append(param.name)
        
        # 准备参数 - 始终传入 stdout 和 rule，然后根据需要添加 node 和 extra_data
        # 这样可以确保 uptime_output 等需要 node 参数的解析器得到所需参数
        args = [stdout, rule]
        
        # 检查是否需要node参数
        param_names = [p.name for p in parameters[1:]]  # 跳过cls
        if 'node' in param_names:
            args.append(node)
        
        # 检查是否需要extra_data参数
        if 'extra_data' in param_names and len(param_names) > 3:  # 确保参数足够多
            args.append(extra_data or {})
            
        # 调用解析器
        logger.debug(f"调用解析器 {parser_name} 使用 {len(args)} 个参数")
        return parser_class.parse(*args)
    except Exception as e:
        logger.exception(f"执行解析器 {parser_name} 时出错: {str(e)}")
        return None

def discover_parsers():
    """
    自动发现和加载所有解析器模块
    """
    current_dir = Path(__file__).parent.absolute()
    parser_files = [f for f in os.listdir(current_dir) if f.endswith('.py') 
                    and not f.startswith('__init__') 
                    and not f.startswith('_')]
    
    for parser_file in parser_files:
        module_name = parser_file[:-3]  # 去掉.py后缀
        module_path = f"inspectors.node.parsers.{module_name}"
        try:
            logger.debug(f"尝试加载解析器模块: {module_path}")
            importlib.import_module(module_path)
            logger.info(f"成功加载解析器模块: {module_path}")
        except ImportError as e:
            logger.warning(f"加载解析器模块 {module_path} 失败: {str(e)}")

# 自动加载所有解析器
discover_parsers()
