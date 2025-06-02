# KubeEye 解析器系统验证指南

本文档提供了验证KubeEye解析器系统是否正确配置和运行的指南。

## 前提条件

- KubeEye 已安装并配置
- 解析器已正确导入并注册

## 验证步骤

### 1. 验证解析器注册

执行以下Python代码，检查解析器是否正确注册：

```python
from inspectors.node.parsers import list_parsers

# 列出所有已注册的解析器
parsers = list_parsers()
print(f"已注册的解析器: {parsers}")

# 确认自定义解析器是否存在
custom_parsers = [p for p in parsers if p.startswith('custom_')]
print(f"自定义解析器: {custom_parsers}")
```

期望输出应包含诸如 `custom_disk_parser`, `custom_memory_parser` 等自定义解析器。

### 2. 验证解析器参数传递

创建一个简单的测试脚本，例如 `test_parser.py`：

```python
from inspectors.node.parsers import parse_output
from utils.rule_loader import Rule

# 创建测试规则
rule = Rule({
    'id': 'test_rule',
    'name': '测试规则',
    'config': {
        'execution': {
            'parser_config': {
                'warning_threshold': 75,
                'critical_threshold': 90
            }
        },
        'threshold': {
            'warning': 80,
            'critical': 95
        }
    }
})

# 创建节点信息
node = {'ip': '192.168.1.100', 'hostname': 'test-node'}

# 创建额外数据
extra_data = {'warning_threshold': 70, 'mount_point': '/data'}

# 测试解析器
stdout = "85"  # 模拟df命令输出
result = parse_output(
    parser_name='custom_disk_parser',
    stdout=stdout,
    rule=rule,
    node=node,
    extra_data=extra_data
)

print(f"解析结果: {result}")
# 验证extra_data参数是否正确传递
assert result['details'].find('70') > 0, "没有使用extra_data中的阈值"
```

### 3. 验证高级解析器功能

创建测试脚本测试高级解析器功能：

```python
from inspectors.node.parsers import parse_output
from utils.rule_loader import Rule

# 创建测试规则
rule = Rule({
    'id': 'advanced_test',
    'name': '高级测试规则'
})

# 创建节点信息
node = {'ip': '192.168.1.100', 'hostname': 'test-node'}

# 测试不同类型的输出
test_cases = [
    {
        'name': 'CPU输出',
        'stdout': 'top - 12:34:56 up 10 days, load average: 0.52, 0.58, 0.59\nCPU usage: 25.5%',
        'config': {'mode': 'auto'}
    },
    {
        'name': 'JSON输出',
        'stdout': '{"cpu": 25.5, "memory": 40.2, "disk": 65.8}',
        'config': {'mode': 'auto'}
    },
    {
        'name': '键值对输出',
        'stdout': 'cpu: 25.5%\nmemory: 40.2%\ndisk: 65.8%',
        'config': {'mode': 'auto'}
    }
]

for case in test_cases:
    result = parse_output(
        parser_name='advanced_system_parser',
        stdout=case['stdout'],
        rule=rule,
        node=node,
        extra_data=case['config']
    )
    print(f"{case['name']} 解析结果: {result}")
```

### 4. 验证错误处理

测试解析器的错误处理功能：

```python
from inspectors.node.parsers import parse_output
from utils.rule_loader import Rule

# 创建测试规则
rule = Rule({
    'id': 'error_test',
    'name': '错误处理测试'
})

# 创建节点信息
node = {'ip': '192.168.1.100', 'hostname': 'test-node'}

# 测试无效输出
invalid_output = "not_a_number"
result = parse_output(
    parser_name='custom_disk_parser',
    stdout=invalid_output,
    rule=rule,
    node=node
)

print(f"错误处理结果: {result}")
assert result['status'] == 'error', "错误处理失败，应返回error状态"

# 测试不存在的解析器
non_existent_result = parse_output(
    parser_name='non_existent_parser',
    stdout="test",
    rule=rule,
    node=node
)

print(f"不存在解析器结果: {non_existent_result}")
assert non_existent_result is None, "对不存在的解析器应返回None"
```

## 故障排除

### 解析器未注册

如果解析器未正确注册，检查以下几点：

1. 确认解析器类使用了`@register_parser`装饰器
2. 确认解析器模块被正确导入
3. 检查解析器类是否继承自`BaseParser`
4. 确认`discover_parsers()`函数被正确调用

### 参数传递问题

如果解析器参数传递有问题：

1. 确认`parse_output`函数签名包含`extra_data`参数
2. 确认解析器类的`parse`方法也接受`extra_data`参数
3. 检查`node_inspector.py`中调用`parse_output`时是否传递了所有参数

### 结果格式问题

如果解析器返回的结果格式不符合预期：

1. 确认解析器使用了`RuleResultBuilder`来构建结果
2. 检查解析器中是否处理了所有可能的异常情况
3. 确认返回的结果包含所有必要的字段
