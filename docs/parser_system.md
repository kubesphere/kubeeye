# KubeEye 解析器系统使用文档

KubeEye 解析器系统允许您灵活地处理节点巡检命令的输出，无需修改核心代码即可添加新的解析逻辑。本文档将介绍如何使用、配置和开发自定义解析器。

## 目录

1. [解析器系统概述](#解析器系统概述)
2. [内置解析器](#内置解析器)
3. [在规则中使用解析器](#在规则中使用解析器)
4. [创建自定义解析器](#创建自定义解析器)
5. [解析器加载路径配置](#解析器加载路径配置)
6. [提示与技巧](#提示与技巧)

## 解析器系统概述

解析器系统是 KubeEye 节点巡检功能的重要组成部分，它允许您:

- 以标准化方式处理各种命令输出
- 定义特定命令输出的解析逻辑
- 根据解析结果生成规范的检查结果
- 在不修改核心代码的情况下添加新的解析器

## 内置解析器

KubeEye 提供了以下内置解析器:

| 解析器名称 | 描述 | 对应命令 | 参数 |
|------------|------|----------|------|
| `uptime_output` | 解析 uptime 命令输出，检查系统负载 | `uptime` | `load_per_core` |
| `free_output` | 解析 free 命令输出，检查内存使用情况 | `free -m` | `warning_threshold`, `critical_threshold` |
| `df_output` | 解析 df 命令输出，检查磁盘使用情况 | `df -h` | `warning_threshold`, `critical_threshold` |
| `service_status` | 检查服务状态 | `systemctl status X` | - |
| `container_runtime_status` | 检查容器运行时状态 | 各种容器运行时检查命令 | - |

## 在规则中使用解析器

### 新格式规则示例

```yaml
id: "node_memory_usage"
name: "内存使用率检查"
category: "node"
description: "检查节点内存使用率是否过高"
remediation: "清理不必要的进程或增加内存"
labels:
  severity: warning
  tag: resource
  type: memory
enabled: true

# 新格式配置
config:
  execution:
    # 命令将被远程执行
    command: "free -m"
    # 使用内置解析器
    parser: "free_output"
    # 解析器的配置参数
    parser_config:
      warning_threshold: 70
      critical_threshold: 85
```

### 旧格式规则示例（向后兼容）

```yaml
id: "node_memory_usage"
name: "内存使用率检查"
category: "node"
description: "检查节点内存使用率是否过高"
query: "free -m"
threshold:
  warning:
    usage_percent: 70
  critical:
    usage_percent: 85
remediation: "清理不必要的进程或增加内存"
custom_data:
  parser: "free_output"
labels:
  severity: warning
  tag: resource
  type: memory
enabled: true
```

## 创建自定义解析器

### 步骤 1: 创建解析器文件

创建一个 Python 文件，例如 `my_custom_parsers.py`，并放置在以下路径之一:

- `/etc/kubeeye/parsers/`
- `~/.kubeeye/parsers/`
- `./parsers/` (相对于当前工作目录)

### 步骤 2: 定义解析器类

```python
#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
自定义解析器示例
"""
from inspectors.node.parsers import BaseParser, register_parser

@register_parser("my_custom_parser")
class MyCustomParser(BaseParser):
    """我的自定义解析器"""
    
    @classmethod
    def parse(cls, stdout, rule, node):
        """
        解析命令输出
        
        Args:
            stdout: 命令输出字符串
            rule: 规则对象
            node: 节点信息字典
            
        Returns:
            字典形式的检查结果
        """
        try:
            # 假设输出是一个简单的数值
            value = float(stdout.strip())
            
            # 获取配置的阈值，如果没有则使用默认值
            warning_threshold = getattr(rule, '_warning_threshold', 70)
            critical_threshold = getattr(rule, '_critical_threshold', 90)
            
            if value >= critical_threshold:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': "值超过严重阈值",
                    'severity': 'critical',
                    'details': f"当前值: {value}, 阈值: {critical_threshold}",
                    'solution': rule.solution or "请检查系统状态"
                }
            elif value >= warning_threshold:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'failed',
                    'description': "值超过警告阈值",
                    'severity': 'warning',
                    'details': f"当前值: {value}, 阈值: {warning_threshold}",
                    'solution': rule.solution or "请留意系统状态"
                }
            else:
                return {
                    'name': f"{rule.name} - {node['ip']}",
                    'status': 'passed',
                    'description': "值在正常范围内",
                    'severity': 'info',
                    'details': f"当前值: {value}",
                    'solution': ""
                }
        except Exception as e:
            return {
                'name': f"{rule.name} - {node['ip']}",
                'status': 'failed',
                'description': "解析输出时出错",
                'severity': 'warning',
                'details': f"错误: {str(e)}, 输出: {stdout}",
                'solution': "检查命令输出格式"
            }
```

### 步骤 3: 在规则中使用自定义解析器

```yaml
id: "my_custom_check"
name: "自定义检查"
category: "node"
description: "使用自定义解析器进行检查"
remediation: "根据结果采取适当的措施"
labels:
  severity: info
  tag: custom
  type: example
enabled: true

config:
  execution:
    command: "your_command_here"
    parser: "my_custom_parser"
    parser_config:
      warning_threshold: 70
      critical_threshold: 90
```

## 解析器加载路径配置

KubeEye 会从以下路径加载自定义解析器:

1. 环境变量定义的配置目录: `$KUBEEYE_CONFIG_DIR/parsers`（默认为 `/etc/kubeeye/parsers`）
2. 用户主目录: `~/.kubeeye/parsers`
3. 当前工作目录: `./parsers`

您也可以在创建 `NodeInspector` 实例时指定自定义路径:

```python
# 指定自定义解析器路径
custom_parser_paths = ["/path/to/my/parsers", "/another/path"]
inspector = NodeInspector(nodes, custom_parser_paths=custom_parser_paths)
```

## 提示与技巧

1. **使用预处理数据**: 
   解析器可以访问规则中定义的预处理数据或通过 `parser_config` 传递的配置。

2. **解析器命名约定**:
   建议使用描述性名称，如 `command_name_output` 或 `check_purpose_parser`。

3. **错误处理**:
   在解析器中始终包含异常处理代码，确保即使命令输出不符合预期也能返回有意义的结果。

4. **测试自定义解析器**:
   通过单元测试验证您的自定义解析器，确保在各种输入条件下都能正确工作。

5. **扩展现有解析器**:
   如果您只需对现有解析器进行小改动，可以继承现有解析器类并重写特定方法。
