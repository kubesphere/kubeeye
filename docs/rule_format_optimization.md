# KubeEye 规则格式和解析器系统优化指南

## 规则格式规范

KubeEye的规则系统支持多种类型的检查，包括Node节点检查、Prometheus监控数据检查以及OPA策略检查。本文档主要关注Node规则格式的优化。

### 标准Node规则格式

```yaml
---
id: unique_rule_id           # 必填：规则唯一标识符
name: 规则名称               # 必填：规则显示名称
description: 规则描述        # 可选：规则的详细描述
category: system             # 可选：规则分类，如system, security, performance等
type: node                  # 必填：规则类型，node表示节点规则
severity: warning           # 必填：规则严重程度，可选值为info, warning, critical
enabled: true               # 可选：是否启用此规则

config:
  execution:
    command: "执行命令"      # 必填：要在节点上执行的Shell命令
    parser: "解析器名称"      # 可选：用于解析命令输出的解析器名称
    timeout: 30             # 可选：命令执行超时时间（秒）
    
    # 解析器配置，会作为extra_data参数传递给解析器
    parser_config:
      key1: value1
      key2: value2
      
    # 依赖命令，用于在主命令前获取额外信息
    dependencies:
      cpu_cores: "nproc"    # 在preprocess_data中会有key为cpu_cores的结果
      
  # 阈值配置，可由解析器使用
  threshold:
    warning: 80             # 警告阈值
    critical: 90            # 严重阈值
    
  # 执行范围配置
  scope:
    node_selector:          # 节点选择器，用于选择应用规则的节点
      key1: value1
      
solution: |                  # 可选：修复建议
  当规则检查失败时的修复建议文本

tags:                        # 可选：规则标签，用于分组和过滤
  - tag1
  - tag2
```

### 解析器系统

Node规则使用解析器（Parser）来解释命令输出并生成检查结果。解析器系统的基本架构如下：

1. **解析器注册**：使用`@register_parser`装饰器注册解析器类
2. **解析方法**：每个解析器类必须实现`parse`方法
3. **参数传递**：解析器接收命令输出、规则对象、节点信息和额外数据
4. **结果格式**：解析器返回标准化的检查结果字典

### 标准解析器结果格式

```python
{
    'name': "规则名称 - 节点标识",        # 结果名称
    'status': "passed/failed/error",    # 状态：通过、失败或错误
    'description': "简洁描述",           # 简短描述
    'severity': "info/warning/critical", # 严重程度
    'details': "详细信息",               # 详细信息
    'solution': "修复建议"               # 修复建议
}
```

## 优化历史

### 1. 参数不匹配问题修复

原问题：NodeInspector调用parse_output函数时传递了5个参数，但函数定义只接收4个参数，导致extra_data参数无法传递给解析器。

修复措施：
- 更新`parse_output`函数签名，添加`extra_data`参数
- 确保调用具体解析器时传递所有参数
- 更新BaseParser基类，使所有解析器具有一致的参数定义

### 2. 解析器自动加载

增加了解析器自动发现和加载功能，使系统能自动加载所有符合命名规范的解析器模块，无需手动导入。

### 3. 规则参数统一

优化了规则参数的获取方式，建立了清晰的优先级：
1. 首先从`parser_config`获取特定解析器的配置
2. 其次从`threshold`配置获取通用阈值
3. 最后使用解析器内置的默认值

### 4. 高级解析器示例

添加了更高级的解析器示例，展示了如何处理复杂的命令输出，包括：
- 自动检测输出类型
- 分类处理不同信息
- 支持JSON和键值对格式解析
- 更灵活的配置参数

## 最佳实践

1. **规则设计**：
   - 每个规则专注于一个明确的检查目标
   - 提供清晰的名称、描述和修复建议
   - 使用适当的严重程度级别

2. **解析器实现**：
   - 继承BaseParser基类
   - 使用@register_parser装饰器注册
   - 严格遵循parse方法签名
   - 始终包含异常处理
   - 返回标准格式的结果

3. **参数处理**：
   - 优先使用`parser_config`中的特定配置
   - 在缺少特定配置时使用`threshold`中的通用阈值
   - 提供合理的默认值作为最后的备选

## 未来改进方向

1. **参数验证**：为解析器添加参数验证机制
2. **结果格式验证**：确保解析器返回的结果符合标准格式
3. **更多内置解析器**：支持更多常见系统检查场景
4. **文档自动生成**：从解析器注释自动生成文档
5. **跨规则依赖**：支持规则之间的依赖关系和结果共享
