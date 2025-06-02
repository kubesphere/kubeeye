# KubeEye 规则格式优化实施计划

## 一、优化目标

1. **标准化规则格式**：为 Node、OPA 和 Prometheus 三种规则类型提供统一的格式规范
2. **分层规则管理**：将规则分为基础(basic)、标准(standard)和扩展(extended)三个层级
3. **统一参数配置**：将配置参数放在结构化的位置，使用一致的命名
4. **完善规则元数据**：为每条规则添加完整的描述信息和解决方案
5. **结构化标签分类**：使用标准化的标签体系便于规则过滤和管理

## 二、规则格式统一标准

### 1. 通用规则结构

```yaml
---
id: type_rule_id           # 唯一标识符，以规则类型为前缀
name: 规则名称             # 简短易懂的名称
description: 规则详细描述   # 完整描述规则的目的和检查内容
type: node|opa|prometheus  # 规则类型
category: category_name    # 规则分类
severity: info|warning|critical  # 严重程度
enabled: true|false        # 是否默认启用
tier: basic|standard|extended  # 规则层级

# 规则配置
config:
  # 类型特定配置...

# 问题解决建议
solution: |
  详细的问题解决步骤...

# 标签，用于过滤和分组
tags:
  - tag1
  - tag2
```

### 2. 类型特定配置

#### Node 规则配置

```yaml
config:
  execution:
    command: "..."        # 要执行的命令
    parser: "..."         # 解析器名称
    timeout: 30           # 执行超时时间（秒）
    dependencies: {...}   # 依赖命令
    parser_config: {...}  # 解析器配置
  
  thresholds:             # 阈值配置
    warning: {...}        # 警告阈值
    critical: {...}       # 严重阈值
  
  scope:
    node_selector: {...}  # 节点选择器
```

#### OPA 规则配置

```yaml
config:
  resources:              # 资源类型配置
    - kind: ...
      apiVersion: ...
      namespaced: true|false
  
  scope:
    namespaces:
      include: [...]      # 包含的命名空间
      exclude: [...]      # 排除的命名空间
  
  rego:                   # Rego 规则
    inline: |
      package kubeeye
      # rego 规则...
    
  thresholds:             # 阈值配置
    value: ...
    comparator: "..."
  
  exemptions:             # 豁免配置
    - name: "..."
      namespace: "..."
      reason: "..."
```

#### Prometheus 规则配置

```yaml
config:
  query:                  # 查询配置
    promql: "..."         # PromQL 查询
    duration: "5m"        # 查询时间范围
    interval: "1m"        # 查询间隔
  
  thresholds:             # 阈值配置
    warning: ...          # 警告阈值
    critical: ...         # 严重阈值
    comparator: "..."     # 比较操作符
  
  alert:                  # 告警配置
    summary: "..."        # 告警摘要
    description: "..."    # 告警描述
```

## 三、规则命名规范

### 1. 规则 ID 命名

- **Node 规则**：以 `node_` 为前缀
- **OPA 规则**：以 `opa_` 为前缀
- **Prometheus 规则**：以 `prometheus_` 为前缀

### 2. 规则分类标准

- **Node 规则分类**：system, performance, security, availability, storage, network
- **OPA 规则分类**：security, compliance, availability, best-practice, resource
- **Prometheus 规则分类**：performance, availability, storage, network, resource, stability

### 3. 标签使用规范

每条规则应至少包含以下标签：
- 规则类型标签（node, opa, prometheus）
- 规则分类标签（与category一致）
- 资源类型标签（cpu, memory, disk, network等）

## 四、优化实施步骤

### 第一阶段：规则模板创建（已完成）

1. 为每种规则类型创建标准模板 ✓
   - 已创建 Node、OPA、Prometheus 三种规则的标准模板

### 第二阶段：现有规则优化

2. 优化 Node 规则
   - ✓ 已优化 `cpu_load.yaml`
   - ✓ 已优化 `custom_parser_example.yaml`
   - 待优化 `memory_usage.yaml`
   - 待优化 `disk_usage_check.yaml`
   - 待优化 `docker_status.yaml`
   - 待优化 `kubelet_status.yaml`
   - 待优化 `advanced_system_check.yaml`

3. 优化 OPA 规则
   - ✓ 已优化 `deployment_replicas.yaml`
   - 待优化 `pod_security_enhanced.yaml`
   - 待优化 `image_tag.yaml`
   - 待优化 `resource_limits_enhanced.yaml`
   - 待优化 `privileged_containers.yaml`

4. 优化 Prometheus 规则
   - ✓ 已优化 `node_cpu_usage.yaml`
   - 待优化 `container_restarts.yaml`
   - 待优化 `node_memory_usage.yaml`
   - 待优化 `node_filesystem_usage.yaml`

### 第三阶段：规则分层组织

5. 将规则按照层级进行分类
   - 基础层（basic）：核心规则，低误报率，适用于所有环境
   - 标准层（standard）：全面规则，中等误报率，适用于生产环境
   - 扩展层（extended）：高级规则，可能有较高误报率，适用于特定场景

### 第四阶段：规则测试验证

6. 为每个优化后的规则进行测试，确保功能正常

### 第五阶段：文档更新

7. 更新规则相关文档
   - 更新规则格式文档
   - 创建规则使用指南
   - 更新自定义规则开发文档

## 五、优化实施进度跟踪

| 规则类型 | 总数量 | 已优化 | 进度 |
|---------|-------|------|-----|
| Node    | 7     | 2    | 28% |
| OPA     | 5     | 1    | 20% |
| Prometheus | 4  | 1    | 25% |
| 总计     | 16    | 4    | 25% |

## 六、后续工作

1. **规则合并与删除**：根据规则集优化方案，整合重复规则
2. **新规则开发**：开发更多有价值的规则，特别是针对安全性和性能方面
3. **规则文档自动生成**：开发工具从规则文件自动生成文档
4. **规则定制化工具**：开发用户友好的规则定制化界面

## 七、总结

规则格式优化是提高 KubeEye 规则系统质量的关键步骤。通过本计划的实施，我们将实现规则格式的标准化、层级化和结构化，使 KubeEye 能够更好地适应不同的使用场景和需求。优化后的规则系统将更加易于维护和扩展，也更加用户友好。
