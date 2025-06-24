#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
数据清理管理器 - 自动清理过期数据防止存储撑爆
"""

import os
import json
import logging
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Tuple, List
import threading
import time

# 获取项目根目录
ROOT_DIR = Path(__file__).parent.parent
DATA_DIR = ROOT_DIR / "data"

logger = logging.getLogger(__name__)

class DataCleanupManager:
    """数据清理管理器"""
    
    def __init__(self):
        self.data_dir = DATA_DIR
        self.cleanup_config = {
            'results': {
                'path': self.data_dir / "results",
                'pattern': "*.json",
                'retention_days': 30,
                'max_files': 1000,
                'enabled': True
            },
            'logs': {
                'path': self.data_dir / "logs",
                'pattern': "*.log",
                'retention_days': 7,
                'max_size_mb': 100,
                'enabled': True
            },
            'schedules': {
                'path': self.data_dir / "schedules",
                'pattern': "*.json",
                'retention_days': 90,
                'max_files': 100,
                'enabled': True
            }
        }
        
        # 定期清理的时间间隔（秒）
        self.cleanup_interval = 24 * 3600  # 每24小时清理一次
        self._cleanup_thread = None
        self._stop_cleanup = False
    
    def cleanup_by_age(self, dir_path: Path, pattern: str, retention_days: int) -> Tuple[int, int]:
        """按文件年龄清理"""
        if not dir_path.exists():
            return 0, 0
        
        cutoff_date = datetime.now() - timedelta(days=retention_days)
        deleted_count = 0
        total_size = 0
        
        for file_path in dir_path.glob(pattern):
            try:
                # 获取文件修改时间
                file_mtime = datetime.fromtimestamp(file_path.stat().st_mtime)
                
                if file_mtime < cutoff_date:
                    file_size = file_path.stat().st_size
                    file_path.unlink()
                    deleted_count += 1
                    total_size += file_size
                    logger.info(f"清理过期文件: {file_path}")
                    
            except Exception as e:
                logger.warning(f"清理文件失败 {file_path}: {e}")
        
        return deleted_count, total_size
    
    def cleanup_by_count(self, dir_path: Path, pattern: str, max_files: int) -> Tuple[int, int]:
        """按文件数量清理（保留最新的文件）"""
        if not dir_path.exists():
            return 0, 0
        
        files = list(dir_path.glob(pattern))
        if len(files) <= max_files:
            return 0, 0
        
        # 按修改时间排序，删除最旧的文件
        files.sort(key=lambda x: x.stat().st_mtime, reverse=True)
        files_to_delete = files[max_files:]
        
        deleted_count = 0
        total_size = 0
        
        for file_path in files_to_delete:
            try:
                file_size = file_path.stat().st_size
                file_path.unlink()
                deleted_count += 1
                total_size += file_size
                logger.info(f"清理多余文件: {file_path}")
            except Exception as e:
                logger.warning(f"清理文件失败 {file_path}: {e}")
        
        return deleted_count, total_size
    
    def cleanup_by_size(self, dir_path: Path, pattern: str, max_size_mb: int) -> Tuple[int, int]:
        """按目录大小清理"""
        if not dir_path.exists():
            return 0, 0
        
        files = list(dir_path.glob(pattern))
        if not files:
            return 0, 0
        
        # 计算总大小
        total_size = sum(f.stat().st_size for f in files)
        max_size_bytes = max_size_mb * 1024 * 1024
        
        if total_size <= max_size_bytes:
            return 0, 0
        
        # 按修改时间排序，删除最旧的文件直到满足大小要求
        files.sort(key=lambda x: x.stat().st_mtime)
        
        deleted_count = 0
        deleted_size = 0
        
        for file_path in files:
            if total_size - deleted_size <= max_size_bytes:
                break
                
            try:
                file_size = file_path.stat().st_size
                file_path.unlink()
                deleted_count += 1
                deleted_size += file_size
                logger.info(f"清理文件(大小限制): {file_path}")
            except Exception as e:
                logger.warning(f"清理文件失败 {file_path}: {e}")
        
        return deleted_count, deleted_size
    
    def cleanup_inspection_results(self) -> Dict:
        """清理巡检结果"""
        config = self.cleanup_config['results']
        if not config['enabled']:
            return {'skipped': True}
        
        results = {}
        
        # 按年龄清理
        deleted_count, deleted_size = self.cleanup_by_age(
            config['path'], 
            config['pattern'], 
            config['retention_days']
        )
        results['by_age'] = {'count': deleted_count, 'size': deleted_size}
        
        # 按数量清理
        deleted_count, deleted_size = self.cleanup_by_count(
            config['path'], 
            config['pattern'], 
            config['max_files']
        )
        results['by_count'] = {'count': deleted_count, 'size': deleted_size}
        
        return results
    
    def cleanup_logs(self) -> Dict:
        """清理日志文件"""
        config = self.cleanup_config['logs']
        if not config['enabled']:
            return {'skipped': True}
        
        results = {}
        
        # 按年龄清理
        deleted_count, deleted_size = self.cleanup_by_age(
            config['path'], 
            config['pattern'], 
            config['retention_days']
        )
        results['by_age'] = {'count': deleted_count, 'size': deleted_size}
        
        # 按大小清理
        deleted_count, deleted_size = self.cleanup_by_size(
            config['path'], 
            config['pattern'], 
            config['max_size_mb']
        )
        results['by_size'] = {'count': deleted_count, 'size': deleted_size}
        
        return results
    
    def cleanup_schedules(self) -> Dict:
        """清理计划任务文件"""
        config = self.cleanup_config['schedules']
        if not config['enabled']:
            return {'skipped': True}
        
        results = {}
        
        # 按年龄清理
        deleted_count, deleted_size = self.cleanup_by_age(
            config['path'], 
            config['pattern'], 
            config['retention_days']
        )
        results['by_age'] = {'count': deleted_count, 'size': deleted_size}
        
        return results
    
    def cleanup_all(self) -> Dict:
        """执行完整的清理"""
        logger.info("开始数据清理任务")
        
        results = {
            'timestamp': datetime.now().isoformat(),
            'results': {},
            'logs': {},
            'schedules': {}
        }
        
        try:
            results['results'] = self.cleanup_inspection_results()
            results['logs'] = self.cleanup_logs()
            results['schedules'] = self.cleanup_schedules()
            
            logger.info("数据清理任务完成")
            
        except Exception as e:
            logger.error(f"数据清理任务失败: {e}")
            results['error'] = str(e)
        
        return results
    
    def get_storage_stats(self) -> Dict:
        """获取存储统计信息"""
        stats = {}
        
        for name, config in self.cleanup_config.items():
            path = config['path']
            if not path.exists():
                stats[name] = {'files': 0, 'size': 0}
                continue
            
            files = list(path.glob(config['pattern']))
            total_size = sum(f.stat().st_size for f in files if f.is_file())
            
            stats[name] = {
                'files': len(files),
                'size': total_size,
                'size_mb': round(total_size / (1024 * 1024), 2)
            }
        
        return stats
    
    def start_auto_cleanup(self):
        """启动自动清理"""
        if self._cleanup_thread and self._cleanup_thread.is_alive():
            return
        
        self._stop_cleanup = False
        self._cleanup_thread = threading.Thread(target=self._auto_cleanup_worker, daemon=True)
        self._cleanup_thread.start()
        logger.info("自动数据清理已启动")
    
    def stop_auto_cleanup(self):
        """停止自动清理"""
        self._stop_cleanup = True
        if self._cleanup_thread and self._cleanup_thread.is_alive():
            self._cleanup_thread.join(timeout=5)
        logger.info("自动数据清理已停止")
    
    def _auto_cleanup_worker(self):
        """自动清理工作线程"""
        while not self._stop_cleanup:
            try:
                self.cleanup_all()
                # 等待下次清理
                for _ in range(self.cleanup_interval):
                    if self._stop_cleanup:
                        break
                    time.sleep(1)
            except Exception as e:
                logger.error(f"自动清理异常: {e}")
                time.sleep(60)  # 出错后等待1分钟再重试

# 全局清理管理器实例
_cleanup_manager = None

def get_cleanup_manager() -> DataCleanupManager:
    """获取清理管理器实例"""
    global _cleanup_manager
    if _cleanup_manager is None:
        _cleanup_manager = DataCleanupManager()
    return _cleanup_manager

def cleanup_old_data(category: str = 'all') -> Dict:
    """清理过期数据的便捷函数"""
    manager = get_cleanup_manager()
    
    if category == 'all':
        return manager.cleanup_all()
    elif category == 'results':
        return manager.cleanup_inspection_results()
    elif category == 'logs':
        return manager.cleanup_logs()
    elif category == 'schedules':
        return manager.cleanup_schedules()
    else:
        raise ValueError(f"不支持的清理类别: {category}")

def get_storage_usage() -> Dict:
    """获取存储使用情况"""
    manager = get_cleanup_manager()
    return manager.get_storage_stats()

# 在模块导入时启动自动清理
def _init_auto_cleanup():
    """初始化自动清理"""
    try:
        manager = get_cleanup_manager()
        manager.start_auto_cleanup()
    except Exception as e:
        logger.warning(f"启动自动清理失败: {e}")

# 注册程序退出时的清理
import atexit
atexit.register(lambda: get_cleanup_manager().stop_auto_cleanup())

# 启动自动清理（可以通过环境变量禁用）
if os.getenv('KUBEEYE_DISABLE_AUTO_CLEANUP', 'false').lower() != 'true':
    _init_auto_cleanup()
