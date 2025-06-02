# KubeEye 规则解析系统 - 设计文档

## 1. 概述

KubeEye规则解析系统是一个灵活的框架，用于解析各种节点检查命令的输出并生成标准化的检查结果。本文档描述了解析器系统的设计、实现和最佳实践。

## 2. 系统架构

解析器系统由以下主要组件组成：

1. **解析器注册表**：存储所有已注册的解析器
2. **解析器基类**：所有解析器继承的基础类
3. **解析器装饰器**：用于注册新解析器
4. **解析函数**：协调解析过程的函数
5. **结果构建器**：用于创建标准格式的结果
6. **自动发现机制**：自动发现并加载所有解析器模块

## 3. 关键组件

### 3.1 解析器基类 (BaseParser)

所有自定义解析器必须继承自`BaseParser`类：

```python
class BaseParser:
    """解析器基类，所有自定义解析器应该继承此类"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """解析命令输出并返回检查结果"""
        raise NotImplementedError("子类必须实现parse方法")
```

### 3.2 解析器装饰器

使用装饰器简化解析器注册：

```python
def register_parser(name: str):
    """用于注册解析器的装饰器"""
    def wrapper(parser_class):
        _PARSER_REGISTRY[name] = parser_class
        return parser_class
    return wrapper
```

### 3.3 解析函数

解析函数协调解析过程：

```python
def parse_output(parser_name: str, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Optional[Dict]:
    """使用指定解析器解析命令输出"""
    parser_class = get_parser(parser_name)
    if parser_class:
        return parser_class.parse(stdout, rule, node, extra_data)
    return None
```

### 3.4 自动发现机制

解析器系统包含自动发现和加载解析器的机制：

```python
def discover_parsers():
    """自动发现和加载所有解析器模块"""
    current_dir = Path(__file__).parent.absolute()
    parser_files = [f for f in os.listdir(current_dir) if f.endswith('.py') 
                    and not f.startswith('__init__') 
                    and not f.startswith('_')]
    
    for parser_file in parser_files:
        module_name = parser_file[:-3]
        module_path = f"inspectors.node.parsers.{module_name}"
        try:
            importlib.import_module(module_path)
        except ImportError as e:
            logger.warning(f"加载解析器模块 {module_path} 失败: {str(e)}")
```

### 3.5 结果构建器

结果构建器提供统一的方式来构建规则结果：

```python
class RuleResultBuilder:
    """规则结果构建器，帮助构建标准格式的规则结果"""

    @staticmethod
    def create(rule, status, description, severity=None, details=None, solution=None, node=None):
        """创建标准格式的规则结果"""
        # ...实现细节...
        return result
```

## 4. 解析器开发指南

### 4.1 创建新解析器

创建新解析器的步骤：

1. 创建一个新的Python文件，放置在`inspectors/node/parsers/`目录
2. 导入必要的模块和BaseParser基类
3. 使用`@register_parser`装饰器注册解析器
4. 实现`parse`方法

```python
from typing import Dict, Any, Optional
from . import BaseParser, register_parser
from ..rule_result_builder import RuleResultBuilder

@register_parser("my_custom_parser")
class MyCustomParser(BaseParser):
    """自定义解析器描述"""
    
    @classmethod
    def parse(cls, stdout: str, rule: Any, node: Dict, extra_data: Optional[Dict] = None) -> Dict:
        """解析命令输出"""
        try:
            # 解析逻辑
            value = float(stdout.strip())
            
            # 获取阈值参数
            extra_data = extra_data or {}
            warning_threshold = extra_data.get('warning_threshold', 70)
            
            # 根据阈值判断状态
            if value > warning_threshold:
                return RuleResultBuilder.failed(
                    rule=rule, 
                    description="检查失败",
                    details=f"值 {value} 超过阈值 {warning_threshold}",
                    node=node
                )
            else:
                return RuleResultBuilder.passed(
                    rule=rule,
                    description="检查通过",
                    details=f"值 {value} 低于阈值 {warning_threshold}",
                    node=node
                )
        except Exception as e:
            return RuleResultBuilder.error(
                rule=rule,
                description="解析错误",
                details=f"错误信息: {str(e)}",
                node=node
            )
```

### 4.2 参数优先级

解析器应按以下优先级获取参数：

1. 首先从`extra_data`（即`parser_config`中的配置）获取特定参数
2. 然后从规则的`threshold`配置中获取
3. 最后使用解析器内置的默认值

### 4.3 结果格式

使用`RuleResultBuilder`确保返回标准格式的结果：

```python
return RuleResultBuilder.create(
    rule=rule,
    status='passed',  # 'passed', 'failed', 'error', 'warning', 'unknown'
    description="简短描述",
    severity='info',  # 'info', 'warning', 'critical'
    details="详细信息",
    solution="修复建议",
    node=node
)
```

### 4.4 错误处理

所有解析器都应包含完善的异常处理：

```python
try:
    # 解析逻辑
except Exception as e:
    return RuleResultBuilder.error(
        rule=rule,
        description="解析错误",
        details=f"错误信息: {str(e)}\n原始输出: {stdout}",
        node=node
    )
```

## 5. 规则格式指南

### 5.1 标准规则格式

```yaml
---
id: rule_id
name: 规则名称
description: 规则描述
type: node
severity: warning
enabled: true

config:
  execution:
    command: "要执行的命令"
    parser: "解析器名称"
    timeout: 30
    
    # 解析器配置，会传递给解析器
    parser_config:
      warning_threshold: 80
      critical_threshold: 90
      
  # 通用阈值配置
  threshold:
    warning: 80
    critical: 90
    
  # 范围配置
  scope:
    node_selector: {}

solution: |
  修复建议

tags:
  - tag1
  - tag2
```

### 5.2 配置优先级

1. 首先使用`parser_config`中的特定配置
2. 其次使用`threshold`中的通用配置
3. 最后使用解析器内置的默认值

## 6. 测试和验证

详细的测试和验证指南可在`docs/parser_validation.md`文件中找到。

## 7. 参考资料

- [KubeEye 规则格式优化指南](/docs/rule_format_optimization.md)
- [解析器系统验证指南](/docs/parser_validation.md)
- [解析器示例](/inspectors/node/parsers/custom_examples.py)
- [规则示例](/rules/node/optimized_disk_check.yaml)
