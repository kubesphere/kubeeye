# KubeEye 巡检规则格式优化指南

## 规则格式优化目标

本次规则格式优化旨在解决以下问题：

1. 消除 `query` 和 `custom_data.check_command` 等字段的重复性
2. 统一不同类型规则的格式结构，提高一致性
3. 提供更清晰、结构化的配置方式
4. 保持向后兼容性，确保现有规则仍然可用
5. 实现规则分层，便于管理不同重要性和场景的规则
6. 标准化规则元数据，便于分类和过滤

## 新旧格式对照

### 节点巡检规则

**旧格式：**
```yaml
id: node_cpu_load
name: CPU 负载检查
description: 检查节点 CPU 负载是否过高
type: node
category: system
severity: warning
enabled: true
query: uptime
threshold:
  warning:
    load_per_core: 1.0
  critical:
    load_per_core: 2.0
solution: 检查高 CPU 使用的进程，优化应用或增加计算资源
tags:
  - cpu
  - performance
  - system
custom_data:
  check_command: uptime
  parser: uptime_output
  cpu_cores_command: nproc
```

**新格式：**
```yaml
id: node_cpu_load
name: CPU 负载检查
description: 检查节点 CPU 负载是否过高
type: node
category: system
severity: warning
enabled: true

# 规则配置
config:
  # 执行配置
  execution:
    command: "uptime"  # 要执行的命令
    parser: "uptime_output"  # 结果解析器
    dependencies:  # 依赖命令
      cpu_cores: "nproc"  # 获取CPU核心数的命令

  # 阈值配置
  threshold:
    warning:
      load_per_core: 1.0
    critical:
      load_per_core: 2.0

  # 应用范围
  scope:
    node_selector:  # 节点选择器，为空表示适用于所有节点
      # kubernetes.io/role: worker  # 可以添加标签选择器

solution: |
  CPU负载过高，请检查：
  1. 使用 'top' 命令查看高占用进程
  2. 优化应用代码或配置
  3. 考虑扩容计算资源

tags:
  - cpu
  - performance
  - system
```

### Prometheus巡检规则

**旧格式：**
```yaml
id: prometheus_node_cpu_usage
name: 节点 CPU 使用率检查
description: 通过 Prometheus 监控节点 CPU 使用率
type: prometheus
category: performance
severity: warning
enabled: true
query: 'avg by (instance) (rate(node_cpu_seconds_total{mode!="idle"}[5m])) * 100'
threshold:
  warning: 80
  critical: 90
solution: 检查高 CPU 使用的进程，优化应用或增加计算资源
tags:
  - cpu
  - performance
  - prometheus
```

**新格式：**
```yaml
id: prometheus_node_cpu_usage
name: 节点CPU使用率检查
description: 通过Prometheus监控节点CPU使用率
type: prometheus
category: performance
severity: warning
enabled: true

# 规则配置
config:
  # 查询配置
  query:
    promql: 'avg by (instance) (rate(node_cpu_seconds_total{mode!="idle"}[5m])) * 100'
    duration: "5m"  # 查询时间范围
    interval: "1m"  # 查询间隔

  # 阈值配置
  threshold:
    warning: 80    # CPU使用率超过80%触发警告
    critical: 90   # CPU使用率超过90%触发严重警告
    comparator: ">"  # 比较操作符：> >= < <= == !=
    
  # 告警配置
  alert:
    summary: "节点 {{ $labels.instance }} CPU使用率高"
    description: "节点 {{ $labels.instance }} CPU使用率为 {{ $value }}%，超过阈值"

solution: |
  CPU使用率过高，请检查：
  1. 查看占用CPU较高的Pod和进程
  2. 分析应用性能瓶颈
  3. 考虑优化应用或增加资源配额

tags:
  - cpu
  - performance
  - prometheus
```

### OPA规则巡检

**旧格式：**
```yaml
id: deployment_replicas
name: Deployment 副本数检查
description: 检查 Deployment 资源是否配置足够的副本数以确保高可用
type: opa
category: availability
severity: warning
enabled: true
query: spec.replicas
threshold:
  value: 2
  comparator: gt
solution: 为关键服务的 Deployment 配置至少 2 个副本，以确保高可用
tags:
  - deployment
  - availability
  - best-practice
resources:
  - kind: Deployment
    apiVersion: apps/v1
    namespaced: true
namespaces:
  - default
  - exclude:
      - kube-system
      - kube-public
      - kube-node-lease
rego_content: |
  package kubeeye
  
  deny[msg] {
    input.kind == "Deployment"
    replicas := input.spec.replicas
    replicas < 2
    
    msg := {
      "Name": input.metadata.name,
      "Namespace": input.metadata.namespace,
      "Message": "Deployment应配置至少2个副本以确保高可用",
      "Level": "warning"
    }
  }
```

**新格式：**
```yaml
id: deployment_replicas
name: Deployment副本数检查
description: 检查Deployment资源是否配置足够的副本数以确保高可用
type: opa
category: availability
severity: warning
enabled: true

# 规则配置
config:
  # 资源配置
  resources:
    - kind: Deployment
      apiVersion: apps/v1
      namespaced: true

  # 范围配置
  scope:
    namespaces:
      include:
        - default
      exclude:
        - kube-system
        - kube-public
        - kube-node-lease

  # Rego规则配置
  rego:
    # 内联Rego规则定义
    inline: |
      package kubeeye
      
      deny[result] {
        input.kind == "Deployment"
        replicas := input.spec.replicas
        replicas < 2
        
        result := {
          "name": input.metadata.name,
          "namespace": input.metadata.namespace,
          "message": "Deployment应配置至少2个副本以确保高可用",
          "severity": "warning"
        }
      }

  # 阈值配置
  threshold:
    value: 2
    comparator: ">="  # 副本数应大于等于2

  # 豁免列表
  exemptions:
    - name: "test-deployment"
      namespace: "test"
      reason: "测试环境允许单副本部署"

solution: |
  生产环境中的Deployment应配置至少2个副本以确保高可用性，建议：
  1. 修改Deployment的replicas字段为2或更高
  2. 使用HorizontalPodAutoscaler进行自动扩缩容
  3. 确保有足够的节点资源支持多副本运行

tags:
  - deployment
  - availability
  - best-practice
```

## 规则迁移步骤

要将现有规则迁移到新格式，请按照以下步骤操作：

1. 根据规则类型选择相应的新格式模板
2. 将规则基本信息（id、name、description等）复制到新格式
3. 根据规则类型，将相关配置迁移到新的 `config` 结构中
4. 更新解决方案文本，使用更详细的描述
5. 保留和更新标签

## 向后兼容性

为了确保平滑迁移，我们的规则加载器（`rule_loader.py`）支持同时加载新旧格式的规则。你可以逐步将规则迁移到新格式，而不必一次性更新所有规则。

在规则加载时，新格式将被解析并映射到系统内部使用的旧结构，确保所有现有代码仍然可以正常工作。

## 迁移建议

1. 先迁移频繁使用的重要规则
2. 在测试环境中验证新规则
3. 分批次迁移规则，每次确保系统正常运行
4. 最终将所有规则更新到新格式

## 新规则格式优势

1. **结构清晰**: 通过分层结构组织规则配置
2. **消除重复**: 合并类似功能的字段，减少冗余
3. **统一接口**: 三种规则类型使用相似的结构
4. **扩展性强**: 便于添加新的配置选项
5. **自解释性**: 字段命名和结构更加符合直觉

## 新规则格式标准化

为进一步提高 KubeEye 规则的质量和一致性，我们对规则格式进行了标准化升级。主要改进包括：

1. **统一的基础结构**：所有规则类型共享相同的基本结构
2. **分层规则系统**：将规则分为基础、标准和扩展三个层级
3. **统一的阈值配置**：统一使用 `thresholds` 字段进行阈值配置
4. **完整的元数据**：规范化规则的描述、分类和解决方案
5. **合理的命名规则**：规范化规则 ID 和属性命名

### 通用规则结构

所有类型的规则都遵循以下基本结构：

```yaml
---
id: rule_id                    # 唯一标识符，使用下划线分隔单词
name: 规则名称                   # 简短易懂的名称
description: 规则详细描述        # 完整描述规则的目的和检查内容
type: node|opa|prometheus      # 规则类型
category: system|security|...  # 规则分类
severity: info|warning|critical # 严重程度
enabled: true|false            # 是否默认启用
tier: basic|standard|extended  # 规则层级

# 规则配置
config:
  # 类型特定配置...

# 问题解决建议
solution: |
  详细的问题解决步骤...

# 标签
tags:
  - tag1
  - tag2
```

### 规则层级说明

- **基础层 (Basic)**：最关键的环境检查，适用于所有集群，默认启用，阈值宽松
- **标准层 (Standard)**：更全面的检查，适用于生产环境，默认启用，阈值适中
- **扩展层 (Extended)**：高级检查，对特定环境的深度优化，默认禁用，阈值严格

### 类型特定配置

#### Node 规则配置

```yaml
config:
  execution:
    command: "执行命令"         # 要执行的 Shell 命令
    parser: "parser_name"      # 解析器名称
    timeout: 30                # 超时时间（秒）
    dependencies:              # 依赖命令，可选
      key: "command"
    parser_config:             # 解析器特定配置
      key1: value1
  
  thresholds:                  # 阈值配置（统一格式）
    warning:                   # 警告级别阈值
      param1: value1
    critical:                  # 严重级别阈值
      param1: value2
  
  scope:                       # 应用范围
    node_selector:            # 节点选择器
      key: value
```

#### OPA 规则配置

```yaml
config:
  resources:                   # 资源类型配置
    - kind: Kind              # Kubernetes 资源类型
      apiVersion: version      # API 版本
      namespaced: true|false   # 是否命名空间作用域
  
  scope:                       # 范围配置
    namespaces:                # 命名空间范围
      include:                 # 包含的命名空间
        - namespace1
      exclude:                 # 排除的命名空间
        - namespace2
  
  rego:                        # Rego 规则
    inline: |                  # 内联 Rego 代码
      package kubeeye
      # rego 规则内容...
    # 或者使用外部文件
    # file: "path/to/file.rego"
  
  thresholds:                  # 阈值配置
    value: 2                  # 阈值值
    comparator: ">="          # 比较操作符
  
  exemptions:                  # 豁免配置
    - name: "resource_name"    # 资源名称
      namespace: "namespace"   # 命名空间
      reason: "豁免原因"       # 豁免原因说明
```

#### Prometheus 规则配置

```yaml
config:
  query:                       # 查询配置
    promql: "prometheus查询表达式"  # PromQL 查询表达式
    duration: "5m"             # 查询时间范围
    interval: "1m"             # 查询间隔
  
  thresholds:                  # 阈值配置
    warning: 80                # 警告阈值
    critical: 90               # 严重阈值
    comparator: ">"            # 比较操作符
  
  alert:                       # 告警配置
    summary: "告警摘要模板"      # 告警摘要
    description: "告警描述模板"  # 告警描述
```

## 规则命名约定

### 规则 ID 命名规则

- Node 规则：以 `node_` 开头，如 `node_cpu_load`
- OPA 规则：以 `opa_` 或 `k8s_` 开头，如 `opa_pod_security`
- Prometheus 规则：以 `prometheus_` 开头，如 `prometheus_node_cpu_usage`

### 规则分类建议

- Node 规则分类：`system`, `performance`, `security`, `availability`, `storage`, `network`
- OPA 规则分类：`security`, `compliance`, `availability`, `best-practice`, `resource`
- Prometheus 规则分类：`performance`, `availability`, `storage`, `network`, `resource`, `stability`

## 自动迁移工具

我们提供了规则自动迁移和验证工具，帮助将旧格式规则转换为新格式：

```bash
# 迁移所有规则
python3 tools/rule_migration.py --input-dir ./rules --output-dir ./rules/new

# 验证迁移后的规则
python3 tools/rule_validation.py --rules-dir ./rules/new
```
