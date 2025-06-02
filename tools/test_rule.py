#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
规则测试工具

此工具用于测试单个规则的执行结果，方便调试和验证规则的有效性。
可以模拟在节点上执行规则命令，并使用指定的解析器解析结果。

用法：
    python3 test_rule.py --rule-file <规则文件路径> [--mock-output <模拟输出>]
"""

import os
import sys
import yaml
import argparse
import logging
import json
from pathlib import Path
from typing import Dict, List, Any, Optional

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def load_rule(rule_file: str) -> Dict:
    """加载规则文件"""
    try:
        with open(rule_file, 'r', encoding='utf-8') as f:
            rule = yaml.safe_load(f)
        return rule
    except Exception as e:
        logger.error(f"加载规则文件失败: {e}")
        return None

def get_parser_module_path(rule_type: str) -> str:
    """获取解析器模块路径"""
    if rule_type == "node":
        return "inspectors.node.parsers"
    elif rule_type == "opa":
        return "inspectors.opa.parsers"
    elif rule_type == "prometheus":
        return "inspectors.prometheus.parsers"
    else:
        return None

def import_parser(parser_name: str, rule_type: str) -> Any:
    """导入解析器模块"""
    parser_module_path = get_parser_module_path(rule_type)
    if not parser_module_path:
        logger.error(f"不支持的规则类型: {rule_type}")
        return None
    
    try:
        # 尝试直接导入解析器
        module_path = f"{parser_module_path}"
        exec(f"import {module_path}")
        module = eval(module_path)
        
        # 检查解析器是否存在
        if hasattr(module, 'parse_output') and callable(module.parse_output):
            logger.info(f"找到解析器函数: {parser_name}")
            return module.parse_output
        else:
            logger.error(f"解析器模块中没有找到parse_output函数")
            return None
    except ImportError as e:
        logger.error(f"导入解析器模块失败: {e}")
        return None

def simulate_command(command: str, mock_output: Optional[str] = None) -> str:
    """模拟执行命令，返回输出"""
    if mock_output:
        logger.info(f"使用模拟输出: {mock_output}")
        return mock_output
    
    # 实际执行命令（小心使用）
    logger.info(f"执行命令: {command}")
    try:
        import subprocess
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        if result.returncode == 0:
            logger.info("命令执行成功")
            return result.stdout
        else:
            logger.error(f"命令执行失败: {result.stderr}")
            return result.stderr
    except Exception as e:
        logger.error(f"执行命令时出错: {e}")
        return f"ERROR: {str(e)}"

def test_rule(rule_file: str, mock_output: Optional[str] = None) -> Dict:
    """测试规则"""
    # 加载规则
    rule = load_rule(rule_file)
    if not rule:
        return {"status": "error", "message": "加载规则失败"}
    
    # 检查规则类型
    rule_type = rule.get("type")
    if not rule_type:
        return {"status": "error", "message": "规则缺少type字段"}
    
    # 提取执行配置
    if "config" not in rule or "execution" not in rule["config"]:
        return {"status": "error", "message": "规则缺少execution配置"}
    
    execution_config = rule["config"]["execution"]
    command = execution_config.get("command")
    parser_name = execution_config.get("parser")
    
    if not command:
        return {"status": "error", "message": "规则缺少command字段"}
    
    if not parser_name:
        return {"status": "error", "message": "规则缺少parser字段"}
    
    # 模拟执行命令
    command_output = simulate_command(command, mock_output)
    
    # 导入解析器
    parser_func = import_parser(parser_name, rule_type)
    if not parser_func:
        return {
            "status": "error", 
            "message": f"无法导入解析器: {parser_name}",
            "command_output": command_output
        }
    
    # 解析结果
    try:
        # 准备解析器参数
        parse_params = {
            "parser_name": parser_name,
            "stdout": command_output,
            "rule": rule,
            "node": {"ip": "127.0.0.1", "name": "test-node"},
            "extra_data": execution_config.get("parser_config", {})
        }
        
        # 调用解析器函数
        parse_result = parser_func(**parse_params)
        
        return {
            "status": "success",
            "command": command,
            "command_output": command_output,
            "parser": parser_name,
            "rule_id": rule.get("id"),
            "rule_name": rule.get("name"),
            "parse_result": parse_result
        }
    except Exception as e:
        logger.exception(f"解析结果时出错: {e}")
        return {
            "status": "error",
            "message": f"解析结果时出错: {str(e)}",
            "command_output": command_output
        }

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description="规则测试工具")
    parser.add_argument("--rule-file", required=True, help="规则文件路径")
    parser.add_argument("--mock-output", help="模拟命令输出")
    parser.add_argument("--json", action="store_true", help="以JSON格式输出结果")
    args = parser.parse_args()
    
    # 测试规则
    result = test_rule(args.rule_file, args.mock_output)
    
    # 输出结果
    if args.json:
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print("\n规则测试结果:")
        print(f"状态: {result.get('status')}")
        
        if result.get('status') == 'error':
            print(f"错误: {result.get('message')}")
        else:
            print(f"规则ID: {result.get('rule_id')}")
            print(f"规则名称: {result.get('rule_name')}")
            print(f"命令: {result.get('command')}")
            print("\n命令输出:")
            print(result.get('command_output'))
            
            print("\n解析结果:")
            parse_result = result.get('parse_result')
            if parse_result:
                for key, value in parse_result.items():
                    print(f"{key}: {value}")
            else:
                print("解析结果为空")
    
    # 设置退出代码
    sys.exit(0 if result.get('status') == 'success' else 1)

if __name__ == "__main__":
    main()
