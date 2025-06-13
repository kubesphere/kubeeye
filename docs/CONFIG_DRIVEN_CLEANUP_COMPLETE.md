# 配置驱动动态客户端清理完成

## 概述

完成了基于 OPA 规则配置的动态客户端优化，删除了不必要的预定义代码，实现了真正的配置驱动模式。

## 清理内容

### 删除的代码

1. **预定义资源映射表** (~70 行)
   ```python
   # 删除了这些预定义映射
   self.resource_mappings = {
       'pods': {'group': '', 'version': 'v1', 'plural': 'pods', 'namespaced': True},
       'deployments': {'group': 'apps', 'version': 'v1', 'plural': 'deployments', 'namespaced': True},
       # ... 大量重复定义
   }
   ```

2. **兼容性包装器** (~40 行)
   ```python
   # 删除了整个 K8sClientCompat 类
   class K8sClientCompat:
       # ... 不再需要的兼容代码
   ```

3. **资源策略配置** (~20 行)
   ```python
   # 删除了复杂的资源策略
   self.resource_strategies = {
       'default': '包含所有配置的资源',
       'controllers_only': '只检查控制器，跳过Pod',
       # ...
   }
   ```

4. **旧版本兼容方法** (~50 行)
   - `_get_cluster_resources_dynamic()` 
   - `_filter_resources_dynamic()`
   - `get_dynamic_client_info()`

### 保留的核心功能

1. **配置解析方法**
   - `build_resource_mappings_from_config()` - 从规则配置构建资源映射
   - `list_resources_from_config()` - 基于配置获取资源
   - `_kind_to_plural()` - Kind 到复数形式转换

2. **基础动态客户端功能**
   - `list_resources()` - 统一资源获取接口
   - `get_resource_definition()` - 获取资源定义
   - `_discover_resource()` - 自动资源发现

## 配置驱动工作流程

### 1. 规则配置示例

```yaml
# rules/opa/image_tag.yaml
config:
  resources:
    - kind: Pod
      apiVersion: v1
      namespaced: true
    - kind: Deployment
      apiVersion: apps/v1
      namespaced: true
    - kind: VirtualService
      apiVersion: networking.istio.io/v1beta1
      namespaced: true
  
  scope:
    namespaces:
      exclude:
        - kube-system
        - kube-public
```

### 2. 动态资源映射构建

```python
# 从配置自动构建 GVR 映射
build_resource_mappings_from_config(rule.config)

# 结果:
# 'pods': {'group': '', 'version': 'v1', 'plural': 'pods', 'namespaced': True}
# 'deployments': {'group': 'apps', 'version': 'v1', 'plural': 'deployments', 'namespaced': True}  
# 'virtualservices': {'group': 'networking.istio.io', 'version': 'v1beta1', 'plural': 'virtualservices', 'namespaced': True}
```

### 3. 基于配置的资源获取

```python
# OPA Inspector 直接使用配置
resources = self._get_cluster_resources_from_rule_config(rule)

# 自动处理:
# - 解析 apiVersion 为 group/version
# - 转换 Kind 为复数资源名
# - 应用命名空间过滤
# - 获取相应资源
```

## 优化效果

### 代码简化

| 项目 | 优化前 | 优化后 | 减少 |
|------|--------|--------|------|
| 预定义资源映射 | ~70 行 | 0 行 | -70 行 |
| 兼容性包装器 | ~40 行 | 0 行 | -40 行 |
| 资源策略配置 | ~20 行 | 0 行 | -20 行 |
| 旧版本兼容方法 | ~50 行 | 0 行 | -50 行 |
| **总计** | **~180 行** | **0 行** | **-180 行** |

### 功能提升

1. **完全配置驱动**
   - 不再需要预定义资源类型
   - 规则配置决定获取哪些资源
   - 自动支持新的资源类型和 CRD

2. **自动 GVR 解析**
   - 从 `apiVersion: apps/v1` 自动解析为 `group: apps, version: v1`
   - 从 `kind: Deployment` 自动转换为 `plural: deployments`

3. **智能命名空间过滤**
   - 基于规则配置的 `scope.namespaces` 自动过滤
   - 支持 `include` 和 `exclude` 配置

4. **CRD 自动支持**
   - 无需手动配置，自动支持所有 CRD
   - 动态发现和解析 CRD 定义

## 使用方式

### 新的 OPA 规则编写

现在只需要在 YAML 中定义需要的资源：

```yaml
config:
  resources:
    - kind: YourCustomResource
      apiVersion: your.domain/v1
      namespaced: true
  
  scope:
    namespaces:
      exclude: ["kube-system"]
```

动态客户端会自动：
1. 解析资源配置
2. 构建 GVR 映射
3. 获取相应资源
4. 应用命名空间过滤

### 开发者体验

- ✅ 无需修改代码即可支持新资源类型
- ✅ 配置即文档，资源需求一目了然
- ✅ 自动错误检测，不支持的资源会提示
- ✅ 统一的获取接口，简化开发

## 总结

通过这次清理，我们实现了：

1. **真正的配置驱动** - 完全基于规则 YAML 配置
2. **代码大幅简化** - 删除了 180+ 行不必要代码
3. **更好的可维护性** - 无硬编码，易于扩展
4. **自动适应能力** - 支持任意资源类型和 CRD

现在的动态客户端是一个真正配置驱动的解决方案，完美解决了原始问题：不再需要为每种资源类型预定义映射，而是完全基于 OPA 规则配置动态获取所需资源。
