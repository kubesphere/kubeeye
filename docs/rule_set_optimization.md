# KubeEye 规则合理化优化方案

## 概述

本文档提供了对 KubeEye 规则集进行合理化优化的详细方案，包括规则保留、合并和删除的建议，以及规则分层策略。目标是提供一个精简、有效且结构清晰的规则集，同时保持对关键检查点的全面覆盖。

## 规则集优化原则

1. **消除重复**：移除功能重叠的规则，保留最全面、最有效的实现
2. **保持核心功能**：确保核心检查点不会在优化中丢失
3. **逻辑分组**：按照功能和目的对规则进行合理分组
4. **遵循分层**：将规则按照重要性和适用范围分为基础、标准和扩展三层

## Node 规则优化

### 规则保留、合并和删除建议

| 规则文件 | 操作 | 层级 | 理由 |
|---------|------|------|------|
| `cpu_load.yaml` | **保留并优化** | 基础 | 核心系统监控指标，使用新格式重命名为 `node_cpu_load.yaml` |
| `memory_usage.yaml` | **保留并优化** | 基础 | 核心系统监控指标，使用新格式重命名为 `node_memory_usage.yaml` |
| `disk_usage_check.yaml` | **保留并优化** | 基础 | 核心系统监控指标，重命名为 `node_disk_usage.yaml` |
| `kubelet_status.yaml` | **保留并优化** | 基础 | Kubernetes 核心组件检查 |
| `docker_status.yaml` | **保留并优化** | 基础 | 容器运行时检查，重命名为 `node_container_runtime.yaml` |
| `advanced_system_check.yaml` | **保留并移至标准层** | 标准 | 高级系统检查，与基础检查互补 |
| `optimized_disk_check.yaml` | **移除** | - | 与 `disk_usage_check.yaml` 功能重复 |
| `custom_parser_example.yaml` | **移至示例目录** | - | 非实际检查规则，作为自定义规则示例 |
| `simple_check.yaml` | **移至示例目录** | - | 非实际检查规则，作为简单规则示例 |
| `custom_disk_check.yaml` | **移至示例目录** | - | 与其他磁盘检查规则重复，作为自定义解析器示例 |

### 新增规则建议

| 规则名称 | 层级 | 描述 |
|---------|------|------|
| `node_network_connectivity.yaml` | 标准 | 检查节点网络连通性和性能 |
| `node_system_logs.yaml` | 标准 | 检查关键系统日志中的错误 |
| `node_kernel_parameters.yaml` | 标准 | 检查 Kubernetes 推荐的内核参数设置 |
| `node_resource_pressure.yaml` | 基础 | 检测资源压力导致的驱逐风险 |
| `node_time_sync.yaml` | 标准 | 检查节点时间同步状态 |

## OPA 规则优化

### 规则保留、合并和删除建议

| 规则文件 | 操作 | 层级 | 理由 |
|---------|------|------|------|
| `deployment_replicas.yaml` | **保留并优化** | 基础 | 高可用性基本检查，使用新格式重命名为 `opa_deployment_replicas.yaml` |
| `pod_security_enhanced.yaml` | **保留并优化** | 标准 | 安全基线检查，重命名为 `opa_pod_security.yaml` |
| `resource_limits_enhanced.yaml` | **保留并优化** | 基础 | 资源管理基本检查，重命名为 `opa_resource_limits.yaml` |
| `privileged_containers.yaml` | **合并至 pod_security** | - | 与 `pod_security_enhanced.yaml` 功能重叠，合并为一个完整的安全检查 |
| `image_tag.yaml` | **保留并优化** | 基础 | 镜像标签检查，重命名为 `opa_image_tag.yaml` |
| `resource_requests.yaml` | **合并至 resource_limits** | - | 与 `resource_limits_enhanced.yaml` 功能重叠，合并为一个完整的资源检查 |

### 新增规则建议

| 规则名称 | 层级 | 描述 |
|---------|------|------|
| `opa_network_policy.yaml` | 标准 | 检查是否定义了适当的网络策略 |
| `opa_storage_security.yaml` | 标准 | 检查 PV/PVC 配置安全性 |
| `opa_service_exposure.yaml` | 基础 | 检查服务暴露方式的安全性 |
| `opa_pod_disruption_budget.yaml` | 标准 | 检查是否配置了 PodDisruptionBudget |
| `opa_namespace_isolation.yaml` | 扩展 | 检查命名空间隔离策略 |

## Prometheus 规则优化

### 规则保留、合并和删除建议

| 规则文件 | 操作 | 层级 | 理由 |
|---------|------|------|------|
| `container_restarts.yaml` | **保留并优化** | 基础 | 容器稳定性关键指标，重命名为 `prometheus_container_restarts.yaml` |
| `node_cpu_usage.yaml` | **保留并优化** | 标准 | 性能监控指标，使用新格式 |
| `node_memory_usage.yaml` | **保留并优化** | 标准 | 性能监控指标，使用新格式 |
| `node_filesystem_usage.yaml` | **保留并优化** | 基础 | 存储监控指标，使用新格式 |

### 新增规则建议

| 规则名称 | 层级 | 描述 |
|---------|------|------|
| `prometheus_node_load_balance.yaml` | 标准 | 检查集群内节点资源使用的平衡性 |
| `prometheus_pod_resource_usage.yaml` | 标准 | 检查 Pod 资源使用率和限制的比例 |
| `prometheus_network_latency.yaml` | 标准 | 检查集群网络性能指标 |
| `prometheus_pod_pending.yaml` | 基础 | 检测长时间处于 Pending 状态的 Pod |
| `prometheus_etcd_health.yaml` | 基础 | 监控 etcd 集群健康状态 |

## 规则分层实施策略

### 基础层 (Basic)

基础层规则具有以下特点：
- 适用于所有 Kubernetes 集群
- 检查 Kubernetes 集群的基本健康状态
- 默认启用，阈值相对宽松
- 检查失败通常表示明显的问题

基础层规则建议名单：
- `node_cpu_load.yaml`
- `node_memory_usage.yaml`
- `node_disk_usage.yaml`
- `node_kubelet_status.yaml`
- `node_container_runtime.yaml`
- `node_resource_pressure.yaml`
- `opa_deployment_replicas.yaml`
- `opa_resource_limits.yaml`
- `opa_image_tag.yaml`
- `opa_service_exposure.yaml`
- `prometheus_container_restarts.yaml`
- `prometheus_node_filesystem_usage.yaml`
- `prometheus_pod_pending.yaml`
- `prometheus_etcd_health.yaml`

### 标准层 (Standard)

标准层规则具有以下特点：
- 适用于生产环境 Kubernetes 集群
- 检查更全面的状态和最佳实践遵从度
- 默认启用，阈值适中
- 检查失败通常表示需要注意的问题

标准层规则建议名单：
- `node_advanced_system_check.yaml`
- `node_network_connectivity.yaml`
- `node_system_logs.yaml`
- `node_kernel_parameters.yaml`
- `node_time_sync.yaml`
- `opa_pod_security.yaml`
- `opa_network_policy.yaml`
- `opa_storage_security.yaml`
- `opa_pod_disruption_budget.yaml`
- `prometheus_node_cpu_usage.yaml`
- `prometheus_node_memory_usage.yaml`
- `prometheus_node_load_balance.yaml`
- `prometheus_pod_resource_usage.yaml`
- `prometheus_network_latency.yaml`

### 扩展层 (Extended)

扩展层规则具有以下特点：
- 适用于特定场景或高要求环境
- 检查高级特性或特殊配置
- 默认禁用，阈值严格
- 检查失败通常表示优化机会

扩展层规则建议名单：
- `node_specialized_kernel_optimization.yaml`（待开发）
- `node_security_hardening.yaml`（待开发）
- `opa_namespace_isolation.yaml`
- `opa_compliance_check.yaml`（待开发）
- `prometheus_advanced_metrics.yaml`（待开发）

## 规则目录结构优化

为了更好地组织规则文件，建议采用以下目录结构：

```
rules/
├── node/           # Node规则
│   ├── basic/      # 基础层Node规则
│   ├── standard/   # 标准层Node规则
│   └── extended/   # 扩展层Node规则
├── opa/            # OPA规则
│   ├── basic/
│   ├── standard/
│   └── extended/
├── prometheus/     # Prometheus规则
│   ├── basic/
│   ├── standard/
│   └── extended/
└── examples/       # 示例规则
    ├── node/
    ├── opa/
    └── prometheus/
```

## 实施路径

1. **创建规则模板**：为每种规则类型创建标准模板（已完成）
2. **开发迁移工具**：实现规则格式迁移脚本（已完成）
3. **规则优化**：按照本文档中的建议，保留、合并或删除规则
4. **规则分层**：将规则按照建议分配到不同层级
5. **目录结构调整**：创建新的目录结构，并移动规则文件
6. **规则验证**：使用验证工具确保所有规则符合新格式
7. **示例规则分离**：将示例性质的规则移至 examples 目录
8. **新规则开发**：按照建议开发新的规则

## 总结

通过这一规则合理化优化方案，KubeEye 将拥有一个更加精简、结构清晰且功能全面的规则集。优化后的规则集将更容易管理和扩展，同时为用户提供更好的使用体验。规则的分层组织将使 KubeEye 能够适应从基本测试环境到高要求生产环境的不同场景。
