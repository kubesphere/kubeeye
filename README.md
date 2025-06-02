# KubeEye - Kubernetes 集群巡检工具

## 概述

KubeEye 是一个用于 Kubernetes 集群巡检的工具，可以帮助运维人员快速发现集群中的潜在问题和安全风险。工具基于 Streamlit 开发，提供了直观的 Web 界面，支持多种巡检方式：

- 节点状态巡检（CPU、内存、磁盘、关键服务等）
- Prometheus 指标巡检（资源监控、Pod 异常等）
- OPA 规则合规性检查（安全配置、最佳实践等）

## 功能特点

- **集群信息管理**：支持配置和管理多个集群的连接信息
- **多种巡检方式**：节点状态、Prometheus 指标、OPA 规则合规性
- **可视化报告**：直观展示巡检结果，包括图表和详细问题说明
- **历史记录查询**：支持查看历史巡检结果和趋势分析
- **修复建议**：针对发现的问题提供解决方案建议
- **敏感信息加密**：通过加密算法保护集群连接密码等敏感信息
- **日志管理**：统一的日志系统，便于追踪和诊断问题

## 安装方法

1. 克隆代码库：
   ```
   git clone https://github.com/pixiake/kubeeye.git
   cd kubeeye
   ```

2. 使用初始化脚本：
   ```
   python init.py --all
   ```
   这将自动完成以下操作：
   - 创建必要的数据目录
   - 安装所需依赖
   - 初始化示例集群配置

   或者手动安装：
   ```
   pip install -r requirements.txt
   ```

3. 运行应用：
   ```
   streamlit run app.py
   ```

## 使用说明

1. **添加集群**：
   - 在「集群信息」页面添加集群名称、节点信息
   - 配置 Prometheus 连接信息（可选）
   - 配置 Kubeconfig（可选）

2. **执行巡检**：
   - 在「集群巡检」页面选择集群和要执行的巡检规则
   - 点击「开始巡检」按钮执行巡检

3. **查看报告**：
   - 在「巡检报告」页面查看巡检结果统计和详细问题列表
   - 可以按集群、巡检类型等条件筛选查看历史报告

4. **管理敏感信息**：
   - 敏感信息（如节点密码、Prometheus 认证信息）会自动加密存储
   - 可以使用 `tools/encrypt_config.py` 工具加密现有配置

## 数据存储

- 集群配置信息保存在 `data/clusters/` 目录下
- 巡检结果保存在 `data/results/` 目录下

## OPA 规则

OPA 规则文件存储在 `inspectors/opa/rules/` 目录下，按资源类型分类：

- `pod/`：Pod 相关规则
- `deployment/`：Deployment 相关规则
- `service/`：Service 相关规则
- `configmap/`：ConfigMap 相关规则
- `security/`：安全相关规则

## 依赖项

- Python 3.8+
- Streamlit
- Pandas
- Plotly
- PyYAML
- Kubernetes Python Client
- Paramiko
- Requests

## 贡献指南

欢迎提交 Issues 和 Pull Requests 来完善此工具。
