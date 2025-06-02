#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
日志配置模块，用于统一管理项目日志
"""

import os
import logging
import logging.handlers
from pathlib import Path
from typing import Optional

# 数据目录定义
DATA_DIR = Path(__file__).parent.parent / "data"
LOGS_DIR = DATA_DIR / "logs"

# 确保日志目录存在
os.makedirs(LOGS_DIR, exist_ok=True)

# 默认日志文件路径
DEFAULT_LOG_FILE = LOGS_DIR / "kubeeye.log"

# 日志级别映射
LOG_LEVELS = {
    "debug": logging.DEBUG,
    "info": logging.INFO,
    "warning": logging.WARNING,
    "error": logging.ERROR,
    "critical": logging.CRITICAL
}

# 日志格式
DEFAULT_LOG_FORMAT = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"

def setup_logger(name: str, level: str = "info", 
                log_file: Optional[Path] = None, 
                log_format: str = DEFAULT_LOG_FORMAT) -> logging.Logger:
    """
    设置日志器
    
    Args:
        name: 日志器名称
        level: 日志级别
        log_file: 日志文件路径，如果为 None 则使用默认路径
        log_format: 日志格式
        
    Returns:
        配置好的日志器
    """
    # 获取日志级别
    log_level = LOG_LEVELS.get(level.lower(), logging.INFO)
    
    # 创建日志器
    logger = logging.getLogger(name)
    logger.setLevel(log_level)
    
    # 清除已有的处理器
    if logger.hasHandlers():
        logger.handlers.clear()
    
    # 创建控制台处理器
    console_handler = logging.StreamHandler()
    console_handler.setLevel(log_level)
    
    # 创建文件处理器
    if log_file is None:
        log_file = DEFAULT_LOG_FILE
        
    file_handler = logging.handlers.RotatingFileHandler(
        log_file, maxBytes=10*1024*1024, backupCount=5, encoding='utf-8'
    )
    file_handler.setLevel(log_level)
    
    # 创建格式器
    formatter = logging.Formatter(log_format)
    console_handler.setFormatter(formatter)
    file_handler.setFormatter(formatter)
    
    # 添加处理器
    logger.addHandler(console_handler)
    logger.addHandler(file_handler)
    
    return logger

# 创建默认的项目日志器
project_logger = setup_logger("kubeeye")

def get_logger(module_name: str) -> logging.Logger:
    """
    获取指定模块的日志器
    
    Args:
        module_name: 模块名称
        
    Returns:
        配置好的模块日志器
    """
    return logging.getLogger(f"kubeeye.{module_name}")

# 设置第三方库的日志级别
logging.getLogger("paramiko").setLevel(logging.WARNING)
logging.getLogger("kubernetes").setLevel(logging.WARNING)
logging.getLogger("urllib3").setLevel(logging.WARNING)
logging.getLogger("streamlit").setLevel(logging.WARNING)
