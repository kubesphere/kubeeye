# KubeEye 重构与优化指南

本文档提供了 KubeEye 集群巡检工具的重构和优化指南，以帮助开发人员了解新的代码组织结构和架构。

## 1. 重构目标

1. **减少代码冗余**：删除重复代码，采用更加统一的接口和实现
2. **提高可维护性**：清晰的代码组织和职责划分
3. **增强可扩展性**：易于添加新的巡检器和规则
4. **保持向后兼容**：支持已有的规则格式和配置

## 2. 新的架构设计

### 2.1 核心组件

新的 KubeEye 架构包含以下核心组件：

1. **巡检控制器**（`InspectionController`）：协调多种巡检器执行，管理巡检结果
2. **巡检器基类**（`BaseInspector`）：定义通用的巡检器接口和基础实现
3. **专用巡检器**：继承基类的特定巡检实现
   - `NodeInspector`：节点巡检器
   - `PrometheusInspector`：Prometheus指标巡检器
   - `OpaInspector`：OPA规则巡检器
4. **解析器系统**：处理命令输出的可扩展模块
5. **规则加载器**：统一的规则加载和解析机制
6. **UI组件**：可重用的UI组件模块，用于构建Web界面
   - `common.py`：共享UI元素和功能
   - `immediate_scan.py`：立即巡检功能
   - `scheduled_scan.py`：定时巡检功能
   - `rule_management.py`：规则管理功能

### 2.2 文件组织结构

```
kubeeye/
├── components/             # UI组件，可被多个页面重用
│   ├── __init__.py
│   ├── common.py           # 共享UI组件
│   ├── immediate_scan.py   # 立即巡检组件
│   ├── rule_management.py  # 规则管理组件
│   └── scheduled_scan.py   # 定时巡检组件
├── inspectors/
│   ├── __init__.py
│   ├── base_inspector.py         # 巡检器基类
│   ├── controller.py             # 统一巡检控制器
│   ├── node/                     # 节点巡检器
│   │   ├── __init__.py
│   │   ├── node_inspector_unified.py
│   │   └── parsers/              # 节点命令输出解析器
│   ├── prometheus/               # Prometheus巡检器
│   │   ├── __init__.py
│   │   └── prometheus_inspector_unified.py
│   └── opa/                      # OPA规则巡检器
│       ├── __init__.py
│       └── opa_inspector_unified.py
├── rules/                        # 规则定义
│   ├── node/
│   ├── prometheus/
│   └── opa/
├── utils/                        # 工具函数
│   ├── __init__.py
│   ├── inspection_result.py
│   ├── k8s_client.py
│   ├── node_connection.py
│   ├── prometheus_client.py
│   ├── rule_loader.py
│   └── ...
├── kubeeye.py                    # 主程序
└── config.yaml.example           # 配置示例
```

## 3. 主要改进

### 3.1 统一的巡检器接口

所有巡检器现在继承自 `BaseInspector` 基类，提供统一的接口：

```python
class BaseInspector(ABC):
    @property
    @abstractmethod
    def inspector_type(self) -> str:
        pass
        
    @abstractmethod
    def _load_rules(self):
        pass
        
    @abstractmethod
    def _apply_rule(self, rule: Rule, context: Dict) -> Dict:
        pass
    
    def run_inspection(self, cluster_name: str, rule_ids: List[str] = None) -> InspectionResult:
        # 默认实现...
```

### 3.2 优化的节点巡检器

新的节点巡检器 `NodeInspector` 采用了更清晰的结构，关键改进：

1. **规则应用逻辑集中**：统一处理规则应用
2. **灵活的解析器机制**：支持通过注册机制动态添加新解析器
3. **更好的错误处理**：详细的错误日志和异常捕获

### 3.3 Prometheus 和 OPA 巡检器

所有巡检器现在使用相同的模式，使开发人员更容易理解和维护代码。

### 3.4 中央巡检控制器

`InspectionController` 负责：
1. 初始化和管理所有巡检器
2. 协调巡检流程
3. 汇总和保存巡检结果
4. 生成巡检报告

## 4. 配置方式

配置文件现在使用 YAML 格式，更加清晰和易于编辑：

```yaml
default_cluster: demo
results_dir: "data/results"

enable_node_inspection: true
enable_prometheus_inspection: true
enable_opa_inspection: true

nodes:
  - name: node1
    ip: 192.168.1.101
    # 其他节点配置...

prometheus:
  url: "http://prometheus.example.com:9090"
  # 其他Prometheus配置...

opa:
  kubeconfig: "~/.kube/config"
  # 其他OPA配置...
```

## 5. 使用方式

### 5.1 命令行使用

```bash
# 运行完整巡检
python kubeeye.py --cluster demo

# 运行特定巡检器
python kubeeye.py --cluster demo --types node,prometheus

# 运行特定规则
python kubeeye.py --cluster demo --rules node_cpu_load,node_memory_usage

# 列出所有规则
python kubeeye.py --list-rules

# 查看帮助
python kubeeye.py --help
```

### 5.2 作为库使用

```python
from inspectors.controller import InspectionController

# 初始化控制器
config = {
    "nodes": [{"name": "node1", "ip": "192.168.1.101", ...}],
    "prometheus": {"url": "http://prometheus:9090"},
    # ...
}
controller = InspectionController(config)

# 运行巡检
results = controller.run_inspection("my-cluster")

# 保存结果
result_dir = controller.save_inspection_result(results, "my-cluster")
```

## 6. 向后兼容性

重构后的代码保持对原有规则格式和配置的向后兼容性：

1. 同时支持新旧规则格式
2. 保留原有规则解析逻辑
3. 兼容原有的命令行参数

## 7. 未来工作

1. **完全移除旧代码**：逐步删除冗余的实现和兼容性代码
2. **增加单元测试覆盖**：为新的组件编写完整的单元测试
3. **完善文档**：更新开发者文档和用户手册
4. **增强Web UI**：更新Streamlit界面以利用新的架构

## 8. 贡献指南

1. 请遵循新的架构设计添加功能
2. 新的巡检器应继承 `BaseInspector` 基类
3. 新的规则解析器应使用解析器注册系统
4. 提交前添加适当的单元测试
