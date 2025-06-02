#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试KubeEye解析器系统的工具脚本
"""
import sys
import os
import json
import argparse
from pathlib import Path
import traceback
import yaml

# 添加项目根目录到路径，以便导入模块
sys.path.append(str(Path(__file__).parent.parent))

# 导入必要的模块
from inspectors.node.parsers import list_parsers, parse_output
from utils.rule_loader import Rule

# 设置参数解析
parser = argparse.ArgumentParser(description='测试KubeEye解析器系统')
parser.add_argument('--list', action='store_true', help='列出所有已注册的解析器')
parser.add_argument('--test-parser', type=str, help='测试指定的解析器')
parser.add_argument('--test-rule', type=str, help='使用规则文件测试解析器')
parser.add_argument('--test-output', type=str, help='用于测试的命令输出，如果不指定则使用示例输出')
parser.add_argument('--verbose', '-v', action='store_true', help='显示详细输出')

args = parser.parse_args()

# 模拟节点信息
node = {'ip': '192.168.1.100', 'hostname': 'test-node'}

# 列出所有解析器
if args.list:
    parsers = list_parsers()
    print(f"\n已注册解析器 ({len(parsers)}):")
    print("="*50)
    for i, parser_name in enumerate(sorted(parsers), 1):
        print(f"{i}. {parser_name}")
    sys.exit(0)

# 测试指定解析器
if args.test_parser:
    parser_name = args.test_parser
    
    # 准备测试数据
    sample_outputs = {
        "custom_disk_parser": "85",
        "custom_memory_parser": "75",
        "advanced_system_parser": "top - 12:34:56 up 10 days, load average: 0.52, 0.58, 0.59\nCPU usage: 25.5%"
    }
    
    # 使用指定的输出或默认样本
    stdout = args.test_output or sample_outputs.get(parser_name, "test_output")
    
    # 创建测试规则
    rule = Rule({
        'id': 'test_rule',
        'name': '测试规则',
        'config': {
            'execution': {
                'parser_config': {
                    'warning_threshold': 75,
                    'critical_threshold': 90,
                    'mount_point': '/data'
                }
            },
            'threshold': {
                'warning': 80,
                'critical': 95
            }
        },
        'solution': "这是测试规则的解决方案"
    })
    
    print(f"\n测试解析器: {parser_name}")
    print("="*50)
    print(f"输入: {stdout}")
    print("-"*50)
    
    try:
        # 执行解析
        result = parse_output(
            parser_name=parser_name,
            stdout=stdout,
            rule=rule,
            node=node,
            extra_data=rule.config.get('execution', {}).get('parser_config', {})
        )
        
        if result:
            print("解析结果:")
            if args.verbose:
                print(json.dumps(result, indent=2, ensure_ascii=False))
            else:
                print(f"状态: {result.get('status')}")
                print(f"描述: {result.get('description')}")
                print(f"严重程度: {result.get('severity')}")
                if 'details' in result:
                    print(f"详细信息: {result.get('details')}")
        else:
            print(f"错误: 解析器 '{parser_name}' 不存在或返回空结果")
    except Exception as e:
        print(f"错误: {str(e)}")
        if args.verbose:
            traceback.print_exc()
    
    sys.exit(0)

# 使用规则文件测试
if args.test_rule:
    rule_path = args.test_rule
    
    if not os.path.exists(rule_path):
        print(f"错误: 规则文件 '{rule_path}' 不存在")
        sys.exit(1)
    
    print(f"\n测试规则: {rule_path}")
    print("="*50)
    
    try:
        # 加载规则文件
        with open(rule_path, 'r') as f:
            rule_data = yaml.safe_load(f)
        
        rule = Rule(rule_data)
        
        # 获取解析器名称
        parser_name = rule.config.get('execution', {}).get('parser')
        if not parser_name:
            print("错误: 规则未指定解析器")
            sys.exit(1)
            
        # 获取命令输出
        command = rule.config.get('execution', {}).get('command')
        print(f"规则命令: {command}")
        
        # 使用指定的输出或创建模拟输出
        stdout = args.test_output
        if not stdout:
            if parser_name == 'custom_disk_parser':
                stdout = "85"
            elif parser_name == 'custom_memory_parser':
                stdout = "75" 
            else:
                stdout = "模拟输出"
        
        print(f"模拟输出: {stdout}")
        print("-"*50)
        
        # 解析配置
        parser_config = rule.config.get('execution', {}).get('parser_config', {})
        
        # 执行解析
        result = parse_output(
            parser_name=parser_name,
            stdout=stdout,
            rule=rule,
            node=node,
            extra_data=parser_config
        )
        
        if result:
            print("解析结果:")
            if args.verbose:
                print(json.dumps(result, indent=2, ensure_ascii=False))
            else:
                print(f"状态: {result.get('status')}")
                print(f"描述: {result.get('description')}")
                print(f"严重程度: {result.get('severity')}")
                if 'details' in result:
                    print(f"详细信息: {result.get('details')}")
        else:
            print(f"错误: 解析器 '{parser_name}' 不存在或返回空结果")
            
    except Exception as e:
        print(f"错误: {str(e)}")
        if args.verbose:
            traceback.print_exc()
    
    sys.exit(0)

# 如果没有指定操作，显示帮助信息
parser.print_help()
