#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
命令安全检查器 - 防止执行高危命令
"""

import re
import logging
from typing import List, Dict, Tuple, Optional
from enum import Enum
import bashlex

logger = logging.getLogger(__name__)

class RiskLevel(Enum):
    """风险等级"""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"

class CommandSecurityChecker:
    """
    命令安全检查器 - 强制白名单模式，只允许只读操作
    
    重要安全原则：
    - 巡检工具只能观察，不能修改
    - 所有高风险命令一律禁止，无例外
    - 不提供任何降低安全级别的选项
    """
    
    def __init__(self):
        """
        初始化命令安全检查器
        
        注意：此类强制使用最严格的安全模式，不接受任何参数
        """
        # 巡检工具的安全原则：只读、只观察、不修改
        self.strict_mode = True      # 强制严格模式，不可更改
        self.whitelist_only = True   # 强制白名单模式，不可更改
        self._init_security_rules()
    
    def _init_security_rules(self):
        """初始化安全规则 - 重点关注只读命令白名单"""
        
        # 绝对禁止的命令（CRITICAL级别）- 任何修改操作
        self.critical_commands = [
            # 文件和目录删除/移动/修改
            r'\brm\s+',                                     # 任何rm命令
            r'\bmv\s+',                                     # 任何mv命令  
            r'\bcp\s+.*>\s*/',                             # 复制覆盖系统文件
            r'\bmkdir\s+',                                  # 创建目录
            r'\brmdir\s+',                                  # 删除目录
            r'\btouch\s+',                                  # 创建/修改文件时间戳
            
            # 权限和所有权修改
            r'\bchmod\s+',                                  # 任何权限修改
            r'\bchown\s+',                                  # 所有权修改
            r'\bchgrp\s+',                                  # 组修改
            
            # 系统服务管理（只禁止修改操作，允许is-active/status/show等只读）
            r'\bsystemctl\s+(start|stop|restart|reload|enable|disable|mask|unmask|kill|reset-failed)\b', # 服务管理
            r'\bservice\s+\w+\s+(start|stop|restart|reload)\b',           # service命令
            r'\binit\s+[0-6]',                             # 系统运行级别改变
            r'\bshutdown\s+',                              # 系统关机
            r'\breboot\s*',                                # 系统重启
            r'\bhalt\s*',                                  # 系统停机
            
            # 进程管理
            r'\bkill\s+(-[0-9]+|\w+)',                     # 杀进程
            r'\bkillall\s+',                               # 批量杀进程
            r'\bpkill\s+',                                 # 模式杀进程
            
            # 危险的系统命令
            r'\bdd\s+.*of=',                               # dd写入操作
            r'\bmkfs\.',                                   # 格式化文件系统
            r'\bmount\s+',                                 # 挂载操作
            r'\bumount\s+',                                # 卸载操作
            r'\bfsck\s+',                                  # 文件系统检查修复
            
            # 软件包管理
            r'\bapt\s+(install|remove|purge|upgrade)',     # Debian包管理
            r'\bapt-get\s+(install|remove|purge|upgrade)', # apt-get操作
            r'\byum\s+(install|remove|erase|update)',      # RedHat包管理
            r'\bdnf\s+(install|remove|erase|update)',      # Fedora包管理
            r'\bpip\s+install',                            # Python包安装
            r'\bnpm\s+install',                            # Node.js包安装
            
            # 网络配置修改
            r'\bifconfig\s+\w+\s+(up|down)',               # 网络接口开关
            r'\bip\s+(addr|link|route)\s+(add|del|set)',   # IP配置修改
            r'\biptables\s+(-A|-D|-I|-R|-F|-X)',          # 防火墙规则修改
            r'\bnetplan\s+apply',                          # 网络配置应用
            
            # 计划任务管理  
            r'\bcrontab\s+(-e|-r)',                        # 编辑/删除cron
            r'\bat\s+',                                    # 计划任务
            
            # 文件重定向和管道（可能修改文件）
            r'>\s*[^/]*/',                                 # 重定向到文件
            r'>>\s*[^/]*/',                                # 追加重定向到文件
            
            # 远程执行和下载
            r'(wget|curl).*\|\s*(sh|bash|python|perl)',   # 下载并执行
            r'(wget|curl).*\|.*sh',                        # 下载并执行（简化版）
            r'\bscp\s+.*:',                                # 远程复制
            r'\brsync\s+.*:',                              # 远程同步
            
            # 编译和构建
            r'\bmake\s+(install|clean)',                   # 编译安装
            r'\b\./configure\s+',                          # 配置脚本
            
            # 内核和系统参数修改
            r'\bsysctl\s+-w',                              # 修改内核参数
            r'\becho\s+.*>\s*/proc/',                      # 修改proc参数
            r'\bmodprobe\s+',                              # 加载内核模块
            r'\brmmod\s+',                                 # 卸载内核模块
        ]
        
        # 安全的只读命令白名单（明确允许的命令）
        self.safe_readonly_patterns = [
            # 系统信息查看
            r'^\s*cat\s+(/proc/|/sys/|/etc/hostname|/etc/os-release)',  # 查看系统文件
            r'^\s*less\s+(/var/log/|/proc/|/sys/)',        # 查看日志和系统信息
            r'^\s*more\s+(/var/log/|/proc/|/sys/)',        # 查看文件内容
            r'^\s*head\s+(-\d+\s+)?(/var/log/|/proc/|/sys/|/etc/)', # 查看文件开头
            r'^\s*tail\s+(-\d+\s+)?(/var/log/|/proc/|/sys/)',        # 查看文件结尾
            r'^\s*(grep|awk|sed)\s+.*(/var/log/|/proc/|/sys/)',      # 文本处理只读
            
            # 系统状态查看
            r'^\s*(uname|hostname|whoami|id|date|uptime)\s*',         # 基本系统信息
            r'^\s*(w|who|last|lastlog)\s*',                           # 用户信息
            r'^\s*(ps|top|htop|pstree|pgrep)\s+',                     # 进程信息
            r'^\s*(free|vmstat|iostat|sar)\s+',                       # 系统资源
            r'^\s*(df|du|lsblk|lsof|fuser)\s+',                       # 磁盘和文件信息
            
            # 网络状态查看  
            r'^\s*(netstat|ss)\s+',                                   # 网络连接状态
            r'^\s*lsof\s+-i',                                         # 网络文件打开
            r'^\s*iptables\s+-L',                                     # 查看防火墙规则
            r'^\s*ip\s+(addr|link|route)\s*(show|list)?',             # IP配置查看
            r'^\s*ifconfig\s*$',                                      # 网络接口查看（无参数）
            
            # 文件系统查看
            r'^\s*ls\s+',                                             # 列出文件
            r'^\s*find\s+.*-type\s+f.*-name',                        # 查找文件
            r'^\s*locate\s+',                                         # 定位文件
            r'^\s*which\s+',                                          # 查找命令路径
            r'^\s*whereis\s+',                                        # 查找命令相关文件
            
            # Kubernetes相关只读命令
            r'^\s*kubectl\s+(get|describe|logs|explain|api-resources|api-versions|version|cluster-info)\s+', # K8s查看命令
            r'^\s*docker\s+(ps|images|version|info|logs)\s+',         # Docker查看命令
            
            # 日志查看
            r'^\s*journalctl\s+(-u\s+\w+\s+)?(-f\s+)?(-n\s+\d+\s+)?(-S\s+.*)?$', # systemd日志查看
            r'^\s*dmesg\s*$',                                         # 内核消息
            
            # 其他只读命令
            r'^\s*history\s*$',                                       # 命令历史
            r'^\s*env\s*$',                                           # 环境变量
            r'^\s*printenv\s*',                                       # 打印环境变量
            r'^\s*echo\s+\$\w+',                                      # 打印变量值
            r'^\s*printf\s+',                                         # 格式化输出
        ]
        self.safe_readonly_patterns.extend([
            r'^\s*(awk|wc|tail|head|xargs|grep|sed)\b.*',
            r'^\s*systemctl\s+(is-active|status|show)\b.*',
            r'^\s*service\s+\w+\s+(status)\b.*',
            r'^\s*echo\b.*',  # 允许 echo 任意内容
        ])

    # 提取 command node 的完整命令字符串
    def _extract_command_string(self, node):
        """
        从 bashlex Command 节点还原命令字符串
        """
        parts = []
        for part in node.parts:
            if hasattr(part, "word"):
                parts.append(part.word)
        return " ".join(parts).strip()

    # 遍历 AST
    def _walk(self, node):
        yield node
        for attr in ("parts", "list", "commands"):
            child = getattr(node, attr, None)
            if child:
                for item in child:
                    yield from self._walk(item)

        # 处理 command substitution
        if hasattr(node, "command"):
            yield from self._walk(node.command)

    def _normalize_command(self, node) -> str:
        """
        统一提取 + 去 sudo + 去重定向
        """

        cmd_str = self._extract_command_string(node)
        if not cmd_str:
            return ""

        # 去 sudo
        if cmd_str.startswith("sudo "):
            cmd_str = cmd_str[5:].lstrip()

        # 去重定向
        cmd_str = re.sub(r'\s*[0-9]*>>?\s*[^ ]+', '', cmd_str)
        cmd_str = re.sub(r'\s*[0-9]*>&\s*[0-9]+', '', cmd_str)
        cmd_str = re.sub(r'\s*<<?\s*[^ ]+', '', cmd_str)
        return cmd_str.strip()


    def check_command_security(self, command: str) -> Tuple[bool, RiskLevel, str]:
        """
        检查命令安全性 - 采用白名单优先模式
        
        Args:
            command: 要检查的命令
            
        Returns:
            (是否安全, 风险等级, 风险描述)
        """
        command = command.strip()
        
        logger.info(f"🔍 安全检查开始 - 命令: {command[:100]}{'...' if len(command) > 100 else ''}")
        
        if not command:
            logger.info("空命令，判定为安全")
            return True, RiskLevel.LOW, "空命令"
        
        # 第一步：检查是否为明确的只读安全命令（白名单）
        if self._is_safe_readonly_command(command):
            logger.info("✅ 命令通过白名单检查")
            return True, RiskLevel.LOW, "安全的只读命令"
        
        # 第二步：检查是否包含绝对禁止的操作（黑名单）
        if self._contains_critical_operations(command):
            risk_level, risk_desc = self._analyze_command_risk(command)
            logger.error(f"❌ 检测到禁止的修改操作: {command[:100]}... 风险: {risk_desc}")
            return False, risk_level, risk_desc
        
        # 第三步：如果启用仅白名单模式，则拒绝所有未明确允许的命令
        if self.whitelist_only:
            logger.warning(f"⚠️ 仅白名单模式：命令未在安全白名单中: {command[:100]}...")
            return False, RiskLevel.HIGH, "命令未在安全白名单中，巡检工具只允许执行只读查看命令"
        
        # 第四步：传统风险分析（用于兼容模式）
        risk_level, risk_desc = self._analyze_command_risk(command)
        
        if risk_level == RiskLevel.CRITICAL:
            logger.error(f"检测到关键风险命令: {command[:100]}... 风险: {risk_desc}")
            return False, risk_level, risk_desc
        
        if self.strict_mode and risk_level == RiskLevel.HIGH:
            logger.warning(f"严格模式下拒绝高风险命令: {command[:100]}... 风险: {risk_desc}")
            return False, risk_level, risk_desc
        
        if risk_level in [RiskLevel.MEDIUM, RiskLevel.HIGH]:
            logger.warning(f"检测到风险命令: {command[:100]}... 风险级别: {risk_level.value}, 描述: {risk_desc}")
            return not self.strict_mode, risk_level, risk_desc
        
        return True, RiskLevel.LOW, "命令通过安全检查"
    
    def _contains_critical_operations(self, command: str) -> bool:
        """
        基于 AST 检查是否包含绝对禁止命令
        """
        logger.info(f"🔍 检查是否包含绝对禁止命令 - 原始命令: {command}")
        try:
            trees = bashlex.parse(command)
        except bashlex.errors.ParsingError as e:
            logger.warning(f"❌ 语法解析失败: {e}")
            return True
        for tree in trees:
            for node in self._walk(tree):
                if node.kind != "command":
                    continue
                cmd_str = self._normalize_command(node)
                if not cmd_str:
                    continue
                logger.info(f"检查子命令: {cmd_str}")
                for pattern in self.critical_commands:
                    if re.search(pattern, cmd_str, re.IGNORECASE):
                        logger.error(f"❌ 匹配到禁止的修改操作: {cmd_str} 模式: {pattern}")
                        return True
                logger.info(f"子命令通过检查: {cmd_str}")

        return False
    
    def _is_safe_readonly_command(self, command: str) -> bool:
        """检查命令是否为安全的只读命令，支持sudo前缀、管道/逻辑组合、去除重定向"""
        logger.info(f"🔍 白名单检查 - 原始命令: {command}")
        try:
            trees = bashlex.parse(command)
        except bashlex.errors.ParsingError as e:
            logger.warning(f"❌ 语法解析失败: {e}")
            return False

        # 遍历所有 AST 节点
        for tree in trees:
            for node in self._walk(tree):
                # 只检查真正的 command 节点
                if node.kind != "command":
                    continue
                cmd_str = self._normalize_command(node)
                if not cmd_str:
                    continue
                logger.info(f"检查子命令: {cmd_str}")
                matched = False
                for pattern in self.safe_readonly_patterns:
                    if re.match(pattern, cmd_str, re.IGNORECASE):
                        logger.info(f"✅ 匹配白名单: {pattern}")
                        matched = True
                        break
                if not matched:
                    logger.warning(f"❌ 未匹配白名单: {cmd_str}")
                    return False
        logger.info("✅ 所有命令节点通过检查")
        return True
    
    def _analyze_command_risk(self, command: str) -> Tuple[RiskLevel, str]:
        """
        分析命令风险 - 识别命令中的风险模式和等级
        
        Args:
            command: 要分析的命令
            
        Returns:
            (风险等级, 风险描述)
        """
        command = command.strip()
        
        if not command:
            return RiskLevel.LOW, "空命令"
        try:
            trees = bashlex.parse(command)
        except bashlex.errors.ParsingError as e:
            logger.warning(f"❌ 命令语法解析失败: {e}")
            return RiskLevel.HIGH, f"命令语法异常: {e}"

        max_level = RiskLevel.LOW
        max_desc = "低风险命令"

        for tree in trees:
            for node in self._walk(tree):

                if node.kind != "command":
                    continue

                cmd_str = self._normalize_command(node)
                if not cmd_str:
                    continue

                level, desc = self._analyze_single_command_risk(cmd_str)

                if level.value > max_level.value:
                    max_level = level
                    max_desc = desc

        return max_level, max_desc
    
    def _analyze_single_command_risk(self, command: str) -> Tuple[RiskLevel, str]:
        """
        分析单个命令风险 - 识别单个命令中的风险模式和等级
        
        Args:
            command: 要分析的命令
            
        Returns:
            (风险等级, 风险描述)
        """
        command = command.strip()
        
        if not command:
            return RiskLevel.LOW, "空命令"
        
        # 检查是否为绝对禁止的命令
        for pattern in self.critical_commands:
            if re.search(pattern, command, re.IGNORECASE):
                return RiskLevel.CRITICAL, "包含绝对禁止的修改操作"
        
        # 检查是否为高风险命令
        high_risk_patterns = [
            r'\b(dd|mkfs|mount|umount|chmod|chown|chgrp|systemctl|service|kill|killall|pkill|reboot|shutdown|halt|apt-get|yum|dnf|pip|npm)\b',
            r'\b(find|locate|grep|awk|sed|xargs|wc|sort|uniq|tee|cut|tr|head|tail)\s+.*[|&]', # 管道/逻辑组合
        ]
        for pattern in high_risk_patterns:
            if re.search(pattern, command, re.IGNORECASE):
                return RiskLevel.HIGH, "包含高风险命令或操作"
        
        # 检查是否为中风险命令
        medium_risk_patterns = [
            r'\b(less|more|cat|echo|printf|env|printenv|history|journalctl|dmesg)\b', # 仅限部分参数
            r'\b(systemctl|service)\s+\w+\s+(status|show|is-active)\b',           # 只读状态查看
        ]
        for pattern in medium_risk_patterns:
            if re.search(pattern, command, re.IGNORECASE):
                return RiskLevel.MEDIUM, "包含中风险命令或操作"
        
        return RiskLevel.LOW, "低风险命令"
