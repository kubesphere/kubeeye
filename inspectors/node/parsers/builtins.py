#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
内置解析器模块，提供常用命令输出解析
"""
import re
from typing import Dict, Any

from . import BaseParser, register_parser

@register_parser("uptime_output")
class UptimeOutputParser(BaseParser):
    """uptime命令输出解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析 uptime 命令输出"""
        # 获取阈值 - 同时支持新的 thresholds 和旧的 threshold 格式
        if hasattr(rule, 'config') and 'thresholds' in rule.config:
            warning_threshold = rule.config['thresholds'].get('warning', {}).get('load_per_core', 1.0) 
            critical_threshold = rule.config['thresholds'].get('critical', {}).get('load_per_core', 2.0)
        elif hasattr(rule, 'threshold'):
            warning_threshold = rule.threshold.get('warning', {}).get('load_per_core', 1.0)
            critical_threshold = rule.threshold.get('critical', {}).get('load_per_core', 2.0)
        else:
            warning_threshold = 1.0  # 默认值
            critical_threshold = 2.0  # 默认值
        
        # 匹配负载值
        load_match = re.search(r'load average: ([\d.]+), ([\d.]+), ([\d.]+)', stdout)
        if not load_match:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"无法解析 CPU 负载信息",
                'severity': 'warning',
                'details': stdout,
                'solution': "检查 uptime 命令输出"
            }
            
        load_1min = float(load_match.group(1))
        load_5min = float(load_match.group(2))
        load_15min = float(load_match.group(3))
        
        # 通过额外参数获取CPU核心数
        if hasattr(rule, '_cpu_cores') and rule._cpu_cores:
            cpu_cores = rule._cpu_cores
        else:
            # 默认值
            cpu_cores = 4
                
        # 计算每核心负载
        load_per_core_1min = load_1min / cpu_cores
        load_per_core_5min = load_5min / cpu_cores
        
        if load_per_core_5min >= critical_threshold:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"CPU 负载严重偏高",
                'severity': 'critical',
                'details': f"5 分钟负载: {load_5min} (每核心: {load_per_core_5min:.2f}), CPU核心数: {cpu_cores}",
                'solution': rule.solution or "检查是否有异常进程或考虑增加资源"
            }
        elif load_per_core_5min >= warning_threshold:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"CPU 负载偏高",
                'severity': 'warning',
                'details': f"5 分钟负载: {load_5min} (每核心: {load_per_core_5min:.2f}), CPU核心数: {cpu_cores}",
                'solution': rule.solution or "监控 CPU 使用情况"
            }
        else:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"CPU 负载正常",
                'severity': 'info',
                'details': f"1/5/15 分钟负载: {load_1min}/{load_5min}/{load_15min}, CPU核心数: {cpu_cores}",
                'solution': ""
            }

@register_parser("free_output")
class FreeOutputParser(BaseParser):
    """free命令输出解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析 free 命令输出"""
        # 获取阈值 - 同时支持新的 thresholds 和旧的 threshold 格式
        if hasattr(rule, 'config') and 'thresholds' in rule.config:
            warning_threshold = rule.config['thresholds'].get('warning', {}).get('value', 80) 
            critical_threshold = rule.config['thresholds'].get('critical', {}).get('value', 90)
        elif hasattr(rule, 'threshold'):
            warning_threshold = rule.threshold.get('warning', 80)
            critical_threshold = rule.threshold.get('critical', 90)
        else:
            warning_threshold = 80  # 默认值
            critical_threshold = 90  # 默认值
        
        lines = stdout.strip().split('\n')
        if len(lines) < 2:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"内存信息格式异常",
                'severity': 'warning',
                'details': stdout,
                'solution': "检查 free 命令输出"
            }
            
        # 查找 "Mem:" 行
        mem_line = None
        for line in lines:
            if line.startswith('Mem:'):
                mem_line = line
                break
            
        if mem_line is None:
            mem_info = lines[1].split()  # 传统格式，第二行是内存信息
        else:
            mem_info = mem_line.split()
            
        if len(mem_info) < 3:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"内存信息格式异常",
                'severity': 'warning',
                'details': stdout,
                'solution': "检查 free 命令输出"
            }
            
        try:
            total = float(mem_info[1])
            used = float(mem_info[2])
        except (ValueError, IndexError):
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"内存信息解析异常",
                'severity': 'warning',
                'details': f"无法从输出中提取内存使用数据: {stdout}",
                'solution': "检查 free 命令输出格式"
            }
        
        if total == 0:
            usage_percent = 0
        else:
            usage_percent = (used / total) * 100
            
        if usage_percent >= critical_threshold:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"内存使用率严重偏高",
                'severity': 'critical',
                'details': f"内存使用率为 {usage_percent:.1f}%",
                'solution': rule.solution or "检查内存泄漏问题或增加内存"
            }
        elif usage_percent >= warning_threshold:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"内存使用率偏高",
                'severity': 'warning',
                'details': f"内存使用率为 {usage_percent:.1f}%",
                'solution': rule.solution or "监控内存使用情况"
            }
        else:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"内存使用率正常",
                'severity': 'info',
                'details': f"内存使用率为 {usage_percent:.1f}%",
                'solution': ""
            }

@register_parser("df_output")
class DfOutputParser(BaseParser):
    """df命令输出解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析 df 命令输出"""
        critical_partitions = []
        warning_partitions = []
        
        # 获取阈值 - 同时支持新的 thresholds 和旧的 threshold 格式
        if hasattr(rule, 'config') and 'thresholds' in rule.config:
            warning_threshold = rule.config['thresholds'].get('warning', {}).get('value', 80) 
            critical_threshold = rule.config['thresholds'].get('critical', {}).get('value', 90)
        elif hasattr(rule, 'threshold'):
            warning_threshold = rule.threshold.get('warning', 80)
            critical_threshold = rule.threshold.get('critical', 90)
        else:
            warning_threshold = 80  # 默认值
            critical_threshold = 90  # 默认值
        
        # 获取排除的文件系统类型和挂载点
        exclude_fs = rule.custom_data.get('exclude_filesystems', []) if hasattr(rule, 'custom_data') else []
        exclude_mounts = rule.custom_data.get('exclude_mounts', []) if hasattr(rule, 'custom_data') else []
        
        # 从新格式规则获取排除项
        if hasattr(rule, 'config') and rule.config:
            if 'filters' in rule.config:
                exclude_fs = rule.config['filters'].get('exclude_filesystems', exclude_fs)
                exclude_mounts = rule.config['filters'].get('exclude_mounts', exclude_mounts)
        
        for line in stdout.strip().split('\n')[1:]:  # 跳过标题行
            parts = line.split()
            if len(parts) >= 5:
                device = parts[0]
                mount_point = parts[5] if len(parts) >= 6 else parts[4]
                usage_str = parts[4]
                
                # 判断是否应该排除此文件系统
                if any(fs in device for fs in exclude_fs) or any(mount in mount_point for mount in exclude_mounts):
                    continue
                
                if '%' in usage_str:
                    try:
                        usage = int(usage_str.rstrip('%'))
                        
                        if usage >= critical_threshold:
                            critical_partitions.append(f"{mount_point}: {usage}%")
                        elif usage >= warning_threshold:
                            warning_partitions.append(f"{mount_point}: {usage}%")
                    except ValueError:
                        continue  # 跳过无法解析的使用率
        
        if critical_partitions:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"磁盘使用率严重偏高",
                'severity': 'critical',
                'details': f"以下分区使用率超过 {critical_threshold}%: {', '.join(critical_partitions)}",
                'solution': rule.solution or "清理磁盘空间或扩容"
            }
        elif warning_partitions:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"磁盘使用率偏高",
                'severity': 'warning',
                'details': f"以下分区使用率超过 {warning_threshold}%: {', '.join(warning_partitions)}",
                'solution': rule.solution or "监控磁盘空间并考虑清理"
            }
        else:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"磁盘使用率正常",
                'severity': 'info',
                'details': f"所有磁盘分区使用率正常",
                'solution': ""
            }

@register_parser("service_status")
class ServiceStatusParser(BaseParser):
    """服务状态解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析服务状态输出"""
        # 获取期望值 - 同时支持新的 thresholds 和旧的 threshold 格式
        if hasattr(rule, 'config') and 'thresholds' in rule.config:
            expected_value = rule.config['thresholds'].get('expected_value', 'active')
        elif hasattr(rule, 'threshold'):
            expected_value = rule.threshold.get('expected_value', 'active')
        else:
            expected_value = 'active'  # 默认值
        status = stdout.strip()
        
        if status != expected_value:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"服务异常",
                'severity': rule.severity,
                'details': f"服务状态: {status}，期望状态: {expected_value}",
                'solution': rule.solution or f"检查并重启服务"
            }
        else:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"服务正常",
                'severity': 'info',
                'details': f"服务状态: {status}",
                'solution': ""
            }

@register_parser("container_runtime_status")
class ContainerRuntimeStatusParser(BaseParser):
    """容器运行时状态解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析容器运行时状态输出"""
        status = stdout.strip()
        
        if status == "not found":
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"未找到容器运行时服务",
                'severity': rule.severity,
                'details': "节点上没有检测到 docker 或 containerd 服务",
                'solution': rule.solution or "安装或配置容器运行时"
            }
        elif status != "active":
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"容器运行时服务异常",
                'severity': rule.severity,
                'details': f"容器运行时服务状态: {status}",
                'solution': rule.solution or "检查并重启容器运行时服务"
            }
        else:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"容器运行时服务正常",
                'severity': 'info',
                'details': f"容器运行时服务状态: {status}",
                'solution': ""
            }

@register_parser("simple_ping_parser")
class SimplePingParser(BaseParser):
    """Ping命令输出解析器"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict) -> Dict:
        """解析ping命令输出"""
        # ping成功会包含"1 received"字符串
        success = "1 received" in stdout
        
        if success:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'passed',
                'description': f"节点网络连通性正常",
                'severity': 'info',
                'details': f"Ping命令成功，节点可以正常响应",
                'solution': ""
            }
        else:
            # 获取期望值 - 同时支持新的 thresholds 和旧的 threshold 格式
            if hasattr(rule, 'config') and 'thresholds' in rule.config:
                expected = rule.config['thresholds'].get('expected_value', 0)
            elif hasattr(rule, 'threshold'):
                expected = rule.threshold.get('expected_value', 0)
            else:
                expected = 0
                
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': f"节点网络连通性异常",
                'severity': rule.severity,
                'details': f"Ping命令失败，节点可能存在网络问题。\n原始输出: {stdout[:200]}",
                'solution': rule.solution if hasattr(rule, 'solution') else "检查节点网络配置和连接状态"
            }
