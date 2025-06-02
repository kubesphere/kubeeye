#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
KubeEye 项目初始化脚本

此脚本用于:
1. 设置必要的数据目录结构
2. 初始化示例集群配置
3. 安装必要的依赖
4. 验证环境设置

在首次运行 KubeEye 应用前执行此脚本
"""

import os
import sys
from pathlib import Path
import argparse
import logging
import json
import shutil
from datetime import datetime

# 确保能够导入项目模块
PROJECT_ROOT = Path(__file__).resolve().parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

# 导入项目模块
from utils.version import VERSION, APP_NAME
from utils.logging_config import setup_logger

# 设置日志
logger = setup_logger("init")

def init_data_directories():
    """初始化数据目录"""
    logger.info("开始初始化数据目录")
    
    # 创建数据目录
    data_dirs = [
        "data/clusters",
        "data/results",
        "data/logs",
        "data/cache"
    ]
    
    for dir_path in data_dirs:
        path = PROJECT_ROOT / dir_path
        path.mkdir(parents=True, exist_ok=True)
        logger.info(f"已创建目录: {path}")
    
    logger.info("数据目录初始化完成")
    return True

def init_demo_cluster():
    """初始化示例集群配置"""
    logger.info("开始初始化示例集群")
    
    demo_config = {
        "name": "demo-cluster",
        "nodes": [
            {
                "ip": "192.168.1.100",
                "port": "22",
                "username": "root",
                "auth_type": "password",
                "password": "demo-password",
                "password_encrypted": False
            }
        ],
        "prometheus": {
            "url": "http://192.168.1.100:9090",
            "username": "",
            "password": "",
            "token": "",
            "enabled": False
        },
        "kubeconfig": "",
        "created_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "updated_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    }
    
    # 保存示例集群配置
    demo_file = PROJECT_ROOT / "data/clusters/demo-cluster.json"
    
    # 如果已存在，不覆盖
    if demo_file.exists():
        logger.info(f"示例集群配置已存在: {demo_file}")
        return True
        
    with open(demo_file, 'w', encoding='utf-8') as f:
        json.dump(demo_config, f, ensure_ascii=False, indent=2)
        
    logger.info(f"已创建示例集群配置: {demo_file}")
    return True

def install_dependencies():
    """检查并安装项目依赖"""
    try:
        import pkg_resources
        
        # 检查 requirements.txt
        req_file = PROJECT_ROOT / "requirements.txt"
        if not req_file.exists():
            logger.warning("未找到 requirements.txt 文件")
            return False
            
        # 读取依赖列表
        with open(req_file, 'r', encoding='utf-8') as f:
            required = [line.strip() for line in f if line.strip() and not line.startswith('#')]
        
        # 检查是否已安装
        installed = {pkg.key for pkg in pkg_resources.working_set}
        missing = [pkg for pkg in required if pkg.split('>=')[0].lower() not in installed]
        
        if missing:
            logger.info(f"检测到缺少的依赖: {', '.join(missing)}")
            if sys.platform != 'win32':
                # 在非 Windows 系统上使用 sudo pip 安装
                cmd = f"{sys.executable} -m pip install {' '.join(missing)}"
            else:
                # 在 Windows 上直接使用 pip 安装
                cmd = f"{sys.executable} -m pip install {' '.join(missing)}"
                
            logger.info(f"执行命令: {cmd}")
            os.system(cmd)
            logger.info("依赖安装完成")
        else:
            logger.info("所有依赖已安装")
            
        return True
    except Exception as e:
        logger.error(f"安装依赖时出错: {str(e)}")
        return False

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description=f"{APP_NAME} {VERSION} 初始化工具")
    parser.add_argument('-d', '--deps', action='store_true', help="安装项目依赖")
    parser.add_argument('-a', '--all', action='store_true', help="执行所有初始化操作")
    args = parser.parse_args()
    
    print(f"\n{APP_NAME} {VERSION} 初始化工具\n")
    
    success = True
    
    # 初始化目录结构
    success &= init_data_directories()
    
    # 初始化示例集群
    success &= init_demo_cluster()
    
    # 安装依赖
    if args.deps or args.all:
        success &= install_dependencies()
    
    if success:
        print(f"\n{APP_NAME} 初始化成功！\n")
        print("运行方式:")
        print("  启动应用: streamlit run 首页.py")
    else:
        print(f"\n{APP_NAME} 初始化过程中出现错误，请查看日志。\n")
    
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())
