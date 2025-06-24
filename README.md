# KubeEye - Kubernetes 集群巡检工具

## 概述

KubeEye 是一个**纯观察型** Kubernetes 集群巡检工具，专注于安全地收集集群信息和发现潜在问题。工具基于 Streamlit 开发，提供直观的 Web 界面，支持多种巡检方式。

**🔒 安全承诺**：KubeEye 采用严格的只读巡检策略，所有操作仅限于信息收集和状态查看，绝不执行任何修改、删除或危险操作，确保集群安全。

## 核心功能

- **🔒 安全第一**：强制只读模式，所有巡检命令经过严格安全检查，确保零风险
- **📊 集群信息管理**：支持配置和管理多个集群的连接信息
- **🔍 多种巡检方式**：节点状态、Prometheus 指标、OPA 规则合规性
- **📈 可视化报告**：直观展示巡检结果，包括图表和详细问题说明
- **📋 历史记录查询**：支持查看历史巡检结果和趋势分析
- **💡 修复建议**：针对发现的问题提供解决方案建议
- **🔐 敏感信息加密**：通过加密算法保护集群连接密码等敏感信息
- **📝 日志管理**：统一的日志系统，便于追踪和诊断问题
- **🛡️ 证书监控**：自动检查 kubeconfig 证书有效期，提前预警

## 🔒 安全特性

### 纯观察型设计
- **只读原则**：所有巡检操作仅限于信息获取和状态查看
- **白名单模式**：只允许明确安全的命令执行，默认拒绝所有未知命令
- **多层安全检查**：命令执行前进行严格的安全验证
- **审计日志**：完整记录所有操作和安全事件

### 安全保障措施
- **禁止修改操作**：严禁 `rm`、`chmod`、`systemctl restart` 等危险命令
- **禁止写入操作**：不允许任何文件重定向、创建文件等写入行为
- **禁止安装操作**：不允许 `apt install`、`pip install` 等软件安装
- **强制安全模式**：无法通过配置降低安全级别

## 快速开始

### 方式一：Docker 运行

#### 持久化数据
```bash
# 创建数据目录
mkdir -p /opt/kubeeye/data

# 运行容器并挂载数据目录和时间
docker run -d \
  --name kubeeye \
  -p 8501:8501 \
  -v /opt/kubeeye/data:/app/data \
  -v /etc/localtime:/etc/localtime:ro \
  kubespheredev/kubeeye:v2.0.0-alpha.1
```

访问地址：http://localhost:8501


### 方式二：Kubernetes 部署

#### 创建命名空间
```bash
kubectl create namespace kubeeye
```

#### 部署应用
```yaml
# kubeeye-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubeeye
  namespace: kubeeye-system
  labels:
    app: kubeeye
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kubeeye
  template:
    metadata:
      labels:
        app: kubeeye
    spec:
      containers:
      - name: kubeeye
        image: kubespheredev/kubeeye:v2.0.0-alpha.1
        ports:
        - containerPort: 8501
        env:
        - name: KUBEEYE_DATA_DIR
          value: "/app/data"
        volumeMounts:
        - name: data-volume
          mountPath: /app/data
        - name: localtime
          mountPath: /etc/localtime
          readOnly: true
        resources:
          limits:
            memory: "1Gi"
            cpu: "500m"
          requests:
            memory: "512Mi"
            cpu: "250m"
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: kubeeye-pvc
      - name: localtime
        hostPath:
          path: /etc/localtime
          type: File

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: kubeeye-pvc
  namespace: kubeeye-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi

---
apiVersion: v1
kind: Service
metadata:
  name: kubeeye-service
  namespace: kubeeye-system
spec:
  selector:
    app: kubeeye
  ports:
    - protocol: TCP
      port: 8501
      targetPort: 8501
  type: NodePort

```

#### 部署命令
```bash
# 应用配置
kubectl apply -f kubeeye-deployment.yaml

# 检查部署状态
kubectl get pods -n kubeeye

# 查看服务
kubectl get svc -n kubeeye
```

## 使用指南

### 1. 配置集群信息
- 访问 "集群信息" 页面
- 添加您要监控的 Kubernetes 集群
- 支持多种连接方式：kubeconfig 文件、SSH 连接

### 2. 执行立即巡检
- 前往 "集群巡检" 页面
- 选择目标集群和巡检规则
- 点击 "开始巡检" 执行检查

### 3. 配置定时巡检
- 在 "定时巡检" 页面创建定时任务
- 支持 Cron 表达式和单次定时
- 自动生成巡检报告

### 4. 查看巡检报告
- "巡检报告" 页面查看历史结果
- 支持导出为 JSON、Excel 格式
- 提供问题修复建议

### 5. 管理规则
- 在 "规则管理" 页面管理巡检规则
- 支持本地规则和 GitOps 模式
- 可以自定义规则配置

