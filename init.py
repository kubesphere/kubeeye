#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
KubeEye 初始化脚本
用于Docker容器启动时的初始化工作
"""

import os
import sys
from pathlib import Path

def ensure_data_directories():
    """确保数据目录存在"""
    data_dir = Path(os.environ.get('KUBEEYE_DATA_DIR', '/app/data'))
    
    # 创建必要的目录
    directories = [
        data_dir / 'clusters',
        data_dir / 'results',
        data_dir / 'logs',
        data_dir / 'schedules',
        data_dir / 'git_rules'
    ]
    
    for directory in directories:
        directory.mkdir(parents=True, exist_ok=True)
        print(f"✅ 确保目录存在: {directory}")

def validate_environment():
    """验证环境配置"""
    required_env_vars = [
        'PYTHONPATH',
        'KUBEEYE_DATA_DIR'
    ]
    
    missing_vars = []
    for var in required_env_vars:
        if not os.environ.get(var):
            missing_vars.append(var)
    
    if missing_vars:
        print(f"⚠️ 缺少环境变量: {', '.join(missing_vars)}")
        return False
    
    print("✅ 环境变量验证通过")
    return True

def main():
    """主初始化函数"""
    print("🚀 KubeEye 初始化开始...")
    
    try:
        # 验证环境
        if not validate_environment():
            sys.exit(1)
        
        # 确保数据目录
        ensure_data_directories()
        
        print("✅ KubeEye 初始化完成!")
        
    except Exception as e:
        print(f"❌ 初始化失败: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
