#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
结果提取器模块，用于从命令输出中提取变量
"""

import re
import logging
from typing import Dict, List, Any

# 设置日志
logger = logging.getLogger(__name__)

class ResultExtractor:
    """从命令输出中提取变量"""
    
    def extract(self, output: str, extractors: List[Dict], context: Dict = None) -> Dict[str, Any]:
        """
        根据提取器配置从输出中提取变量
        
        Args:
            output: 命令输出文本
            extractors: 提取器配置列表
            context: 已有的上下文变量
            
        Returns:
            提取的变量字典
        """
        result = context.copy() if context else {}
        
        for extractor in extractors:
            name = extractor.get("name")
            if not name:
                logger.warning("提取器缺少name字段")
                continue
                
            pattern = extractor.get("pattern")
            value_type = extractor.get("type", "str")
            
            if pattern:
                # 使用正则表达式提取
                try:
                    match = re.search(pattern, output)
                    if match:
                        # 检查是否有捕获组
                        if match.groups():
                            # 有捕获组，使用第一个捕获组
                            value = match.group(1)
                        else:
                            # 没有捕获组，使用整个匹配
                            value = match.group(0)
                        result[name] = self._convert_value(value, value_type)
                    else:
                        logger.warning(f"提取器 '{name}' 的模式 '{pattern}' 没有匹配到内容")
                        result[name] = None
                except (re.error, IndexError) as e:
                    logger.error(f"提取器 '{name}' 正则表达式错误: {str(e)}")
                    result[name] = None
        
        return result
    
    def _convert_value(self, value: str, value_type: str) -> Any:
        """
        转换值类型
        
        Args:
            value: 字符串值
            value_type: 目标类型
            
        Returns:
            转换后的值
        """
        try:
            if value_type == "int":
                return int(value)
            elif value_type == "float":
                return float(value)
            elif value_type == "bool":
                return value.lower() in ("true", "yes", "1", "on")
            else:
                return value
        except (ValueError, TypeError) as e:
            logger.error(f"类型转换失败: {str(e)}, 返回原始值")
            return value
    

