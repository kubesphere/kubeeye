# KubeEye

<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-8-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

![kubeeye-logo](./docs/images/kubeeye-logo.png?raw=true)

> English | [中文](README_zh.md)

KubeEye 旨在发现 Kubernetes 上的各种问题，比如应用配置错误（使用 [OPA](https://github.com/open-policy-agent/opa) ）、集群组件不健康和节点问题（使用[Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector)）。除了预定义的规则，它还支持自定义规则。

## 架构图

KubeEye 通过调用Kubernetes API，通过匹配资源中的关键字和容器语法的规则匹配来获取集群诊断数据，详见架构图。

![kubeeye-architecture](./docs/images/kubeeye-architecture.svg?raw=true)

## 怎么使用

- 机器上安装 KubeEye
  - 从 [Releases](https://github.com/kubesphere/kubeeye/releases) 中下载预构建的可执行文件。
    
  - 或者你也可以从源代码构建
  > 提示：构建完成后将会在 /usr/local/bin/ 目录下生成 kubeeye 文件。
  
  ```
  git clone https://github.com/kubesphere/kubeeye.git
  cd kubeeye 
  make installke
  ```
- [可选] 安装 [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector)
注意：这将在你的集群上安装 npd，只有当你想要详细的节点报告时才需要。  
```shell
kubeeye install npd
```

- KubeEye 执行
```shell
kubeeye audit
KIND          NAMESPACE        NAME                                                           REASON                                        LEVEL    MESSAGE
Node                           docker-desktop                                                 kubelet has no sufficient memory available   waring    KubeletHasNoSufficientMemory
Node                           docker-desktop                                                 kubelet has no sufficient PID available      waring    KubeletHasNoSufficientPID
Node                           docker-desktop                                                 kubelet has disk pressure                    waring    KubeletHasDiskPressure
Deployment    default          testkubeeye                                                                                                                  NoCPULimits
Deployment    default          testkubeeye                                                                                                                  NoReadinessProbe
Deployment    default          testkubeeye                                                                                                                  NotRunAsNonRoot
Deployment    kube-system      coredns                                                                                                               NoCPULimits
Deployment    kube-system      coredns                                                                                                               ImagePullPolicyNotAlways
Deployment    kube-system      coredns                                                                                                               NotRunAsNonRoot
Deployment    kubeeye-system   kubeeye-controller-manager                                                                                            ImagePullPolicyNotAlways
Deployment    kubeeye-system   kubeeye-controller-manager                                                                                            NotRunAsNonRoot
DaemonSet     kube-system      kube-proxy                                                                                                            NoCPULimits
DaemonSet     k          ube-system      kube-proxy                                                                                                            NotRunAsNonRoot
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

## KubeEye 能做什么

- KubeEye 根据行业最佳实践审查你的工作负载 yaml 规范，帮助你使你的集群稳定。
- KubeEye 可以发现你的集群控制平面的问题，包括 kube-apiserver/kube-controller-manager/etcd 等。
- KubeEye 可以帮助你检测各种节点问题，包括内存/CPU/磁盘压力，意外的内核错误日志等。

## 检查项

|是/否|检查项 |描述|
|---|---|---|
| :white_check_mark: | PrivilegeEscalationAllowed     | 允许特权升级 |
| :white_check_mark: | CanImpersonateUser             | role/clusterrole 有伪装成其他用户权限 |
| :white_check_mark: | CanDeleteResources             | role/clusterrole 有删除 kubernetes 资源权限 |
| :white_check_mark: | CanModifyWorkloads             | role/clusterrole 有修改 kubernetes 资源权限 |
| :white_check_mark: | NoCPULimits                    | 资源没有设置 CPU 使用限制 |
| :white_check_mark: | NoCPURequests                  | 资源没有设置预留 CPU |
| :white_check_mark: | HighRiskCapabilities           | 开启了高危功能，例如 ALL/SYS_ADMIN/NET_ADMIN |
| :white_check_mark: | HostIPCAllowed                 | 开启了主机 IPC |
| :white_check_mark: | HostNetworkAllowed             | 开启了主机网络 |
| :white_check_mark: | HostPIDAllowed                 | 开启了主机PID |
| :white_check_mark: | HostPortAllowed                | 开启了主机端口 |
| :white_check_mark: | ImagePullPolicyNotAlways       | 镜像拉取策略不是 always |
| :white_check_mark: | ImageTagIsLatest               | 镜像标签是 latest |
| :white_check_mark: | ImageTagMiss                   | 镜像没有标签 |
| :white_check_mark: | InsecureCapabilities           | 开启了不安全的功能，例如 KILL/SYS_CHROOT/CHOWN |
| :white_check_mark: | NoLivenessProbe                | 没有设置存活状态检查 |
| :white_check_mark: | NoMemoryLimits                 | 资源没有设置内存使用限制 |
| :white_check_mark: | NoMemoryRequests               | 资源没有设置预留内存 |
| :white_check_mark: | NoPriorityClassName            | 没有设置资源调度优先级 |
| :white_check_mark: | PrivilegedAllowed              | 以特权模式运行资源 |
| :white_check_mark: | NoReadinessProbe               | 没有设置就绪状态检查 |
| :white_check_mark: | NotReadOnlyRootFilesystem      | 没有设置 root 文件系统为只读 |
| :white_check_mark: | NotRunAsNonRoot                |  没有设置禁止以 root 用户启动进程 | 
| :white_check_mark: | CertificateExpiredPeriod       | 将检查 ApiServer 证书的到期日期少于30天 |
| :white_check_mark: | EventAudit                     | 事件检查 |
| :white_check_mark: | NodeStatus                     | 节点状态检查 |
| :white_check_mark: | DockerStatus                   | docker 状态检查 |             
| :white_check_mark: | KubeletStatus                  | kubelet 状态检查 |

## 添加自定义检查规则

### 添加自定义 OPA 检查规则
- 创建 OPA 规则存放目录
```shell
mkdir opa
```
- 添加自定义 OPA 规则文件
> 注意：为检查工作负载设置的 OPA 规则， package 名称必须是 *kubeeye_workloads_rego*
> 为检查 RBAC 设置的 OPA 规则， package 名称必须是 *kubeeye_RBAC_rego*
> 为检查节点设置的 OPA 规则， package 名称必须是 *kubeeye_nodes_rego*

- 以下为检查镜像仓库地址规则，保存以下规则到规则文件 *imageRegistryRule.rego*
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

- 使用额外的规则运行 kubeeye
> 提示：kubeeye 将读取指定目录下所有 *.rego* 结尾的文件

```shell
kubeeye audit -p ./opa -f ~/.kube/config
NAMESPACE     NAME              KIND          MESSAGE
default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
```
### 添加自定义 NPD 检查规则
- 修改 configmap
```shell
kubectl edit ConfigMap node-problem-detector-config -n kube-system 
```
- 重启 NPD
```shell
kubectl rollout restart DaemonSet node-problem-detector -n kube-system
```

## 文档
* [RoadMap](docs/roadmap.md)
* [FAQ](docs/FAQ.md)

