#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
巡检结果管理模块，用于保存和加载巡检结果
"""

import json
import yaml
import os
import csv
import openpyxl
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Any, Optional, Union, Tuple

# 数据目录定义
DATA_DIR = Path(__file__).parent.parent / "data"
RESULTS_DIR = DATA_DIR / "results"

# 确保目录存在
os.makedirs(RESULTS_DIR, exist_ok=True)

class InspectionResult:
    """巡检结果类"""
    
    def __init__(self, cluster_name: str, inspection_type: str):
        """
        初始化巡检结果
        
        Args:
            cluster_name: 集群名称
            inspection_type: 巡检类型
        """
        self.cluster_name = cluster_name
        self.inspection_type = inspection_type
        self.timestamp = datetime.now()
        self.result_id = f"{cluster_name}_{inspection_type}_{self.timestamp.strftime('%Y%m%d%H%M%S')}"
        self.items = []
        
    def add_item(self, item: Dict) -> None:
        """
        添加巡检项
        
        Args:
            item: 巡检项字典，需包含：
                - name: 巡检项名称
                - status: 'passed' 或 'exception' (简化后的状态体系)
                - description: 描述
                - severity: 严重程度 ('critical', 'warning', 'info') - 仅用于异常项的细分级别
                - details: 详细内容
                - solution: 解决方案 (可选)
        """
        if 'solution' not in item:
            item['solution'] = ''
            
        # 状态标准化：统一将 failed、warning、error 转换为 exception
        if item.get('status') in ['failed', 'warning', 'error']:
            item['status'] = 'exception'
            
        self.items.append(item)
    
    def get_items(self) -> List[Dict]:
        """获取所有巡检项"""
        return self.items
    
    def get_summary(self) -> Dict:
        """获取巡检摘要"""
        passed = 0
        exception_critical = 0
        exception_warning = 0
        exception_info = 0
        
        for item in self.items:
            # 安全地获取status和severity，处理不同类型的item
            if isinstance(item, dict):
                status = item.get('status', 'unknown')
                severity = item.get('severity', 'unknown')
            elif hasattr(item, 'status'):
                status = getattr(item, 'status', 'unknown')
                severity = getattr(item, 'severity', 'unknown')
            else:
                status = 'unknown'
                severity = 'unknown'
            
            # 简化的状态体系：只有 passed 和 exception
            if status == 'passed':
                passed += 1
            else:
                # 所有非通过的状态都视为异常，按严重程度细分
                if severity == 'critical':
                    exception_critical += 1
                elif severity == 'warning':
                    exception_warning += 1
                else:
                    exception_info += 1
                
        total_exceptions = exception_critical + exception_warning + exception_info
                
        return {
            'cluster_name': self.cluster_name,
            'inspection_type': self.inspection_type,
            'timestamp': self.timestamp,
            'result_id': self.result_id,
            'total': len(self.items),
            'passed': passed,
            'total_exceptions': total_exceptions,
            'exception_critical': exception_critical,
            'exception_warning': exception_warning,
            'exception_info': exception_info,
            # 为兼容性保留旧字段
            'critical': exception_critical,
            'warning': exception_warning,
            'info': exception_info
        }
    
    def save(self) -> str:
        """
        保存巡检结果
        
        Returns:
            结果文件路径
        """
        # 创建集群结果目录
        cluster_dir = RESULTS_DIR / self.cluster_name
        os.makedirs(cluster_dir, exist_ok=True)
        
        result_data = {
            'cluster_name': self.cluster_name,
            'inspection_type': self.inspection_type,
            'timestamp': self.timestamp.isoformat(),
            'result_id': self.result_id,
            'items': self.items
        }
        
        result_file = cluster_dir / f"{self.result_id}.json"
        with open(result_file, 'w', encoding='utf-8') as f:
            json.dump(result_data, f, ensure_ascii=False, indent=2)
            
        return str(result_file)


def load_result(result_id: str) -> Optional[Dict]:
    """
    加载巡检结果
    
    Args:
        result_id: 巡检结果 ID
        
    Returns:
        巡检结果字典，如果不存在则返回 None
    """
    # 搜索results目录下所有json文件，找到匹配的result_id
    for file_path in RESULTS_DIR.glob('*.json'):
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                result_data = json.load(f)
                
            # 检查result_id是否匹配
            if result_data.get('result_id') == result_id:
                return result_data
        except Exception:
            continue
            
    # 如果没有找到，也尝试从文件名匹配
    result_file = RESULTS_DIR / f"inspection_result_{result_id.replace('_', '_')}.json"
    if result_file.exists():
        with open(result_file, 'r', encoding='utf-8') as f:
            return json.load(f)
    
    return None


def export_report(result_id: str, format_type: str = "json") -> Tuple[bool, str]:
    """
    导出巡检报告为不同格式
    
    Args:
        result_id: 巡检结果ID
        format_type: 导出格式，支持 "json", "csv", "excel"
        
    Returns:
        (成功, 文件路径) 元组，成功为 True 时返回导出文件路径
    """
    # 加载巡检结果
    result_data = load_result(result_id)
    if not result_data:
        return False, "找不到指定巡检结果"
    
    # 从 result_id 解析出集群名
    parts = result_id.split('_')
    if len(parts) < 3:
        return False, "无效的结果ID格式"
        
    cluster_name = parts[0]
    export_dir = RESULTS_DIR / cluster_name / "exports"
    os.makedirs(export_dir, exist_ok=True)
    
    # 根据格式类型导出
    if format_type == "json":
        # 直接使用原始结果文件
        source_path = RESULTS_DIR / cluster_name / f"{result_id}.json"
        export_path = export_dir / f"{result_id}.json"
        
        if source_path.exists():
            import shutil
            shutil.copy(source_path, export_path)
            return True, str(export_path)
        else:
            return False, "找不到源文件"
            
    elif format_type == "csv":
        export_path = export_dir / f"{result_id}.csv"
        
        try:
            with open(export_path, 'w', newline='', encoding='utf-8') as csvfile:
                fieldnames = ['name', 'status', 'severity', 'description', 'details', 'solution']
                writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
                
                writer.writeheader()
                for item in result_data.get('items', []):
                    writer.writerow({
                        'name': item.get('name', ''),
                        'status': item.get('status', ''),
                        'severity': item.get('severity', ''),
                        'description': item.get('description', ''),
                        'details': item.get('details', '').replace('\n', ' '),
                        'solution': item.get('solution', '').replace('\n', ' ')
                    })
                    
            return True, str(export_path)
        except Exception as e:
            return False, f"导出CSV失败: {str(e)}"
            
    elif format_type == "excel":
        export_path = export_dir / f"{result_id}.xlsx"
        
        try:
            wb = openpyxl.Workbook()
            ws = wb.active
            ws.title = "巡检结果"
            
            # 添加标题信息
            ws['A1'] = "集群巡检报告"
            ws['A2'] = f"集群名称: {result_data.get('cluster_name', '')}"
            ws['A3'] = f"巡检时间: {result_data.get('timestamp', '')}"
            ws['A4'] = f"报告ID: {result_data.get('result_id', '')}"
            
            # 添加表头
            headers = ['名称', '状态', '严重程度', '描述', '详细信息', '解决方案']
            for col, header in enumerate(headers, start=1):
                ws.cell(row=6, column=col, value=header)
            
            # 添加数据
            for row_idx, item in enumerate(result_data.get('items', []), start=7):
                ws.cell(row=row_idx, column=1, value=item.get('name', ''))
                ws.cell(row=row_idx, column=2, value=item.get('status', ''))
                ws.cell(row=row_idx, column=3, value=item.get('severity', ''))
                ws.cell(row=row_idx, column=4, value=item.get('description', ''))
                ws.cell(row=row_idx, column=5, value=item.get('details', ''))
                ws.cell(row=row_idx, column=6, value=item.get('solution', ''))
            
            # 保存工作簿
            wb.save(export_path)
            return True, str(export_path)
        except Exception as e:
            return False, f"导出Excel失败: {str(e)}"
    else:
        return False, f"不支持的导出格式: {format_type}"


def list_results(cluster_name: Optional[str] = None) -> List[Dict]:
    """
    列出巡检结果
    
    Args:
        cluster_name: 可选的集群名称过滤
        
    Returns:
        巡检结果摘要列表
    """
    results = []
    
    # 直接搜索results目录下的所有json文件
    for file_path in RESULTS_DIR.glob('*.json'):
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                result_data = json.load(f)
                
            # 如果指定了集群名称，进行过滤
            if cluster_name and result_data.get('cluster_name') != cluster_name:
                continue
                    
            # 处理新的数据结构
            if 'summary' in result_data:
                # 新格式：有summary字段 - 使用简化状态系统
                summary_data = result_data['summary']
                passed = summary_data.get('passed', 0)
                # 在简化状态系统中，所有异常都在 error 字段中
                total_exceptions = summary_data.get('error', 0) + summary_data.get('failed', 0) + summary_data.get('warning', 0)
                critical = total_exceptions  # 所有异常都显示为需要关注的问题
                warning = 0  # 简化状态系统中不再区分警告
                info = 0
                total = summary_data.get('total_items', 0)
            elif 'inspection_results' in result_data:
                # 新格式：但没有summary，需要计算
                critical = 0
                warning = 0
                info = 0
                passed = 0
                total = 0
                
                for inspector_type, inspector_result in result_data.get('inspection_results', {}).items():
                    items = inspector_result.get('items', [])
                    total += len(items)
                    
                    for item in items:
                        # 安全地获取status，处理不同类型的item
                        if isinstance(item, dict):
                            status = item.get('status', 'unknown')
                        elif hasattr(item, 'status'):
                            status = getattr(item, 'status', 'unknown')
                        else:
                            status = 'unknown'
                        
                        if status == 'passed':
                            passed += 1
                        elif status == 'exception':  # 使用新的简化状态
                            critical += 1
                        else:
                            info += 1
            else:
                # 旧格式：直接items字段
                critical = 0
                warning = 0
                info = 0
                passed = 0
                
                for item in result_data.get('items', []):
                    # 安全地获取status和severity，处理不同类型的item
                    if isinstance(item, dict):
                        status = item.get('status', 'unknown')
                        severity = item.get('severity', 'unknown')
                    elif hasattr(item, 'status'):
                        status = getattr(item, 'status', 'unknown')
                        severity = getattr(item, 'severity', 'unknown')
                    else:
                        status = 'unknown'
                        severity = 'unknown'
                    
                    if status == 'passed':
                        passed += 1
                    elif severity == 'critical':
                        critical += 1
                    elif severity == 'warning':
                        warning += 1
                    else:
                        info += 1
                
                total = len(result_data.get('items', []))
            
            summary = {
                'cluster_name': result_data.get('cluster_name', ''),
                'inspection_type': result_data.get('inspection_type', 'unknown'),
                'timestamp': result_data.get('timestamp', ''),
                'result_id': result_data.get('result_id', ''),
                'total': total,
                'passed': passed,
                'critical': critical,
                'warning': warning,
                'info': info
            }
            
            results.append(summary)
        except Exception as e:
            # 跳过无法解析的文件
            continue
    
    # 按时间戳排序，最新的在前
    results.sort(key=lambda x: x['timestamp'], reverse=True)
    return results


def get_latest_result_by_cluster(cluster_name: str) -> Optional[Dict[str, Any]]:
    """
    获取指定集群的最新巡检结果
    
    Args:
        cluster_name (str): 集群名称
        
    Returns:
        Optional[Dict[str, Any]]: 最新的巡检结果，如果没有则返回 None
    """
    results = list_results()
    
    for result in results:
        if result.get('cluster_name') == cluster_name:
            return result
    
    return None
