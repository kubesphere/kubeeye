<div align=center><img src="docs/images/KubeEye-O.svg?raw=true"></div>

<p align=center>
<a href="https://github.com/kubesphere/kubeeye/actions?query=event%3Apush+branch%3Amain+workflow%3ACI+"><img src="https://github.com/kubesphere/kubeeye/workflows/CI/badge.svg?branch=main&event=push"></a>
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="https://github.com/kubesphere/kubeeye#contributors-"><img src="https://img.shields.io/badge/all_contributors-10-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

> English | [中文](README_zh.md)

KubeEye 是为 Kubernetes 设计的巡检工具，用于发现 Kubernetes 资源（使用 [OPA](https://github.com/open-policy-agent/opa) ）、集群组件、集群节点（使用[Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector)）等配置是否符合最佳实践，对于不符合最佳实践的，将给出修改建议。

KubeEye 支持自定义巡检规则、插件安装，通过 [KubeEye Operator](#kubeeye-operator) 能够使用 web 页面的图形化展示来查看巡检结果以及给出修复建议。

## 架构图

KubeEye 通过 Kubernetes API 获取资源详情，通过巡检规则和插件检查获取到的资源配置，并生成诊断结果，详见架构图。

![kubeeye-architecture](./docs/images/kubeeye-architecture.svg?raw=true)

## 安装并使用 KubeEye

1. 机器上安装 KubeEye。
  - 方法 1：从 [Releases](https://github.com/kubesphere/kubeeye/releases) 中下载预构建的可执行文件。

  - 方法 2：从源代码构建。
  > 提示：构建完成后将会在 `/usr/local/bin/` 目录下生成 KubeEye 文件。

  ```
  git clone https://github.com/kubesphere/kubeeye.git
  cd kubeeye 
  make installke
  ```
2. [可选] 安装 [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector)。

  > 提示：如果您需要详细的节点报告，可以运行该命令。运行后，将在你的集群上安装 NPD。

   ```shell
   kubeeye install npd
   ```

3. 使用 KubeEye 进行巡检。

```shell
kubeeye audit
KIND          NAMESPACE        NAME                                                           REASON                                        LEVEL    MESSAGE
Node                           docker-desktop                                                 kubelet has no sufficient memory available   warning    KubeletHasNoSufficientMemory
Node                           docker-desktop                                                 kubelet has no sufficient PID available      warning    KubeletHasNoSufficientPID
Node                           docker-desktop                                                 kubelet has disk pressure                    warning    KubeletHasDiskPressure
Deployment    default          testkubeeye                                                                                                                  NoCPULimits
Deployment    default          testkubeeye                                                                                                                  NoReadinessProbe
Deployment    default          testkubeeye                                                                                                                  NotRunAsNonRoot
Deployment    kube-system      coredns                                                                                                               NoCPULimits
Deployment    kube-system      coredns                                                                                                               ImagePullPolicyNotAlways
Deployment    kube-system      coredns                                                                                                               NotRunAsNonRoot
Deployment    kubeeye-system   kubeeye-controller-manager                                                                                            ImagePullPolicyNotAlways
Deployment    kubeeye-system   kubeeye-controller-manager                                                                                            NotRunAsNonRoot
DaemonSet     kube-system      kube-proxy                                                                                                            NoCPULimits
DaemonSet     kube-system      kube-proxy                                                                                                            NotRunAsNonRoot
Event         kube-system      coredns-558bd4d5db-c26j8.16d5fa3ddf56675f                      Unhealthy                                    warning   Readiness probe failed: Get "http://10.1.0.87:8181/ready": dial tcp 10.1.0.87:8181: connect: connection refused
Event         kube-system      coredns-558bd4d5db-c26j8.16d5fa3fbdc834c9                      Unhealthy                                    warning   Readiness probe failed: HTTP probe failed with statuscode: 503
Event         kube-system      vpnkit-controller.16d5ac2b2b4fa1eb                             BackOff                                      warning   Back-off restarting failed container
Event         kube-system      vpnkit-controller.16d5fa44d0502641                             BackOff                                      warning   Back-off restarting failed container
Event         kubeeye-system   kubeeye-controller-manager-7f79c4ccc8-f2njw.16d5fa3f5fc3229c   Failed                                       warning   Failed to pull image "controller:latest": rpc error: code = Unknown desc = Error response from daemon: pull access denied for controller, repository does not exist or may require 'docker login': denied: requested access to the resource is denied
Event         kubeeye-system   kubeeye-controller-manager-7f79c4ccc8-f2njw.16d5fa3f61b28527   Failed                                       warning   Error: ImagePullBackOff
Role          kubeeye-system   kubeeye-leader-election-role                                                                                          CanDeleteResources
ClusterRole                    kubeeye-manager-role                                                                                                  CanDeleteResources
ClusterRole                    kubeeye-manager-role                                                                                                  CanModifyWorkloads
ClusterRole                    vpnkit-controller                                                                                                     CanImpersonateUser
ClusterRole                    vpnkit-controller                                                                                           CanDeleteResources
```

## KubeEye 能为您做什么

- KubeEye 根据 Kubernetes 最佳实践来检查集群资源，确保集群保持最佳配置，稳定运行。
- KubeEye 可以帮助您发现集群控制平面问题，包括 kube-apiserver、kube-controller-manager、etcd 等。
- KubeEye 可以帮助您检测各种集群节点问题，包括内存、CPU、磁盘压力、意外的内核错误日志等。

## 检查项

|是/否|检查项 |描述|级别|
|---|---|---|---|
| :white_check_mark: | PrivilegeEscalationAllowed     | 允许特权升级 | 紧急 |
| :white_check_mark: | CanImpersonateUser             | role/clusterrole 有伪装成其他用户权限 | 警告 |
| :white_check_mark: | CanDeleteResources             | role/clusterrole 有删除 kubernetes 资源权限 | 警告 |
| :white_check_mark: | CanModifyWorkloads             | role/clusterrole 有修改 kubernetes 资源权限 | 警告 |
| :white_check_mark: | NoCPULimits                    | 资源没有设置 CPU 使用限制 | 紧急 |
| :white_check_mark: | NoCPURequests                  | 资源没有设置预留 CPU | 紧急 |
| :white_check_mark: | HighRiskCapabilities           | 开启了高危功能，例如 ALL/SYS_ADMIN/NET_ADMIN | 紧急 |
| :white_check_mark: | HostIPCAllowed                 | 开启了主机 IPC | 紧急 |
| :white_check_mark: | HostNetworkAllowed             | 开启了主机网络 | 紧急 |
| :white_check_mark: | HostPIDAllowed                 | 开启了主机PID | 紧急 |
| :white_check_mark: | HostPortAllowed                | 开启了主机端口 | 紧急 |
| :white_check_mark: | ImagePullPolicyNotAlways       | 镜像拉取策略不是 always | 警告 |
| :white_check_mark: | ImageTagIsLatest               | 镜像标签是 latest | 警告 |
| :white_check_mark: | ImageTagMiss                   | 镜像没有标签 | 紧急 |
| :white_check_mark: | InsecureCapabilities           | 开启了不安全的功能，例如 KILL/SYS_CHROOT/CHOWN | 警告 |
| :white_check_mark: | NoLivenessProbe                | 没有设置存活状态检查 | 警告 |
| :white_check_mark: | NoMemoryLimits                 | 资源没有设置内存使用限制 | 紧急 |
| :white_check_mark: | NoMemoryRequests               | 资源没有设置预留内存 | 紧急 |
| :white_check_mark: | NoPriorityClassName            | 没有设置资源调度优先级 | 通知 |
| :white_check_mark: | PrivilegedAllowed              | 以特权模式运行资源 | 紧急 |
| :white_check_mark: | NoReadinessProbe               | 没有设置就绪状态检查 | 警告 |
| :white_check_mark: | NotReadOnlyRootFilesystem      | 没有设置根文件系统为只读 | 警告 |
| :white_check_mark: | NotRunAsNonRoot                | 没有设置禁止以 root 用户启动进程 | 警告 |
| :white_check_mark: | CertificateExpiredPeriod       | 将检查 ApiServer 证书的到期日期少于30天 | 紧急 |
| :white_check_mark: | EventAudit                     | 事件检查 | 警告 |
| :white_check_mark: | NodeStatus                     | 节点状态检查 | 警告 |
| :white_check_mark: | DockerStatus                   | docker 状态检查 | 警告 |          
| :white_check_mark: | KubeletStatus                  | kubelet 状态检查 | 警告 |

## 添加自定义检查规则

### 添加自定义 OPA 检查规则

1. 创建 OPA 规则存放目录。

   ```shell
   mkdir opa
   ```

2. 添加自定义 OPA 规则文件。

   > 注意：
   - 为检查工作负载设置的 OPA 规则， package 名称必须是 *kubeeye_workloads_rego*。
   - 为检查 RBAC 设置的 OPA 规则， package 名称必须是 *kubeeye_RBAC_rego*。
   - 为检查节点设置的 OPA 规则， package 名称必须是 *kubeeye_nodes_rego*。

3. 为检查镜像仓库地址规则，保存以下规则到规则文件 *imageRegistryRule.rego*。

  ```rego
  package kubeeye_workloads_rego

  deny[msg] {
      resource := input
      type := resource.Object.kind
      resourcename := resource.Object.metadata.name
      resourcenamespace := resource.Object.metadata.namespace
      workloadsType := {"Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"}
      workloadsType[type]

      not workloadsImageRegistryRule(resource)

      msg := {
          "Name": sprintf("%v", [resourcename]),
          "Namespace": sprintf("%v", [resourcenamespace]),
          "Type": sprintf("%v", [type]),
          "Message": "ImageRegistryNotmyregistry"
      }
  }

  workloadsImageRegistryRule(resource) {
      regex.match("^myregistry.public.kubesphere/basic/.+", resource.Object.spec.template.spec.containers[_].image)
  }
  ```

4. 使用新增规则运行 KubeEye.

  > 提示：KubeEye 将读取指定目录下所有 *.rego* 结尾的文件。

```shell
kubeeye audit -p ./opa -f ~/.kube/config
NAMESPACE     NAME              KIND          MESSAGE
default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
```
### 添加自定义 NPD 检查规则

1. 执行以下命令修改 ConfigMap：

   ```shell
   kubectl edit ConfigMap node-problem-detector-config -n kube-system 
   ```

2. 执行以下命令重启 NPD：
   ```shell
   kubectl rollout restart DaemonSet node-problem-detector -n kube-system
   ```

## KubeEye Operator

### 什么是 KubeEye Operator

KubeEye Operator 是为 Kubernetes 设计的巡检平台。通过 Operator 管理 KubeEye，能够在 Kubernetes 集群中定期执行 KubeEye 巡检，并生成巡检报告。

### KubeEye Operator 能为您做什么

- 提供 web 管理页面。通过 CR 记录 KubeEye 巡检结果，让您可视化查看和对比集群巡检结果。
- 支持安装更多插件。
- 提供更加详细的修改建议。

### 部署 KubeEye Operator

```shell
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye.yaml
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye_insights.yaml
```
### 查看 KubeEye Operator 巡检结果

```shell
kubectl get clusterinsight -o yaml
```

```shell
apiVersion: v1
items:
- apiVersion: kubeeye.kubesphere.io/v1alpha1
  kind: ClusterInsight
  metadata:
    name: clusterinsight-sample
    namespace: default
  spec:
    auditPeriod: 24h
  status:
    auditResults:
      auditResults:
      - resourcesType: Node
        resultInfos:
        - namespace: ""
          resourceInfos:
          - items:
            - level: warning
              message: KubeletHasNoSufficientMemory
              reason: kubelet has no sufficient memory available
            - level: warning
              message: KubeletHasNoSufficientPID
              reason: kubelet has no sufficient PID available
            - level: warning
              message: KubeletHasDiskPressure
              reason: kubelet has disk pressure
            name: kubeeyeNode
```

## 相关文档
* [RoadMap](docs/roadmap.md)
* [FAQ](docs/FAQ.md)

