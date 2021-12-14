![kubeeye-logo](./docs/images/kubeeye-logo.png?raw=true)

> English | [中文](README_zh.md)

# KubeEye

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
  make install
  ```
- [可选] 安装 [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector)
注意：这一行将在你的集群上安装 npd，只有当你想要详细的报告时才需要。  
`kubeeye install -e npd`  

- KubeEye 执行
```
root@node1:# kubeeye audit
NAMESPACE     NAME              KIND          MESSAGE
default       nginx             Deployment    [nginx CPU limits should be set. nginx CPU requests should be set. nginx image tag not specified, do not use 'latest'. nginx livenessProbe should be set. nginx memory limits should be set. nginx memory requests should be set. nginx priorityClassName can be set. nginx root file system should be set read only. nginx readinessProbe should be set. nginx runAsNonRoot can be set.]
default       testcronjob       CronJob       [testcronjob CPU limits should be set. testcronjob CPU requests should be set. testcronjob allowPrivilegeEscalation should be set false. testcronjob have HighRisk capabilities. testcronjob hostIPC should not be set. testcronjob hostNetwork should not be set. testcronjob hostPID should not be set. testcronjob hostPort should not be set. testcronjob imagePullPolicy should be set 'Always'. testcronjob image tag not specified, do not use 'latest'. testcronjob have insecure capabilities. testcronjob livenessProbe should be set. testcronjob memory limits should be set. testcronjob memory requests should be set. testcronjob priorityClassName can be set. testcronjob privileged should be set false. testcronjob root file system should be set read only. testcronjob readinessProbe should be set.]
kube-system   testrole          Role          [testrole can impersonate user. testrole can delete resources. testrole can modify workloads.]
              testclusterrole   ClusterRole   [testclusterrole can impersonate user. testclusterrole can delete resource. testclusterrole can modify workloads.]

NAMESPACE     SEVERITY   PODNAME                              EVENTTIME                   REASON    MESSAGE
kube-system   Warning    vpnkit-controller.16acd7f7536c62e8   2021-10-11T15:55:08+08:00   BackOff   Back-off restarting failed container

NODENAME        SEVERITY     HEARTBEATTIME               REASON              MESSAGE
node18          Fatal        2020-11-19T10:32:03+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node19          Fatal        2020-11-19T10:31:37+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node2           Fatal        2020-11-19T10:31:14+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node3           Fatal        2020-11-27T17:36:53+08:00   KubeletNotReady     Container runtime not ready: RuntimeReady=false reason:DockerDaemonNotReady message:docker: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

NAME            SEVERITY     TIME                        MESSAGE
scheduler       Fatal        2020-11-27T17:09:59+08:00   Get http://127.0.0.1:10251/healthz: dial tcp 127.0.0.1:10251: connect: connection refused
etcd-0          Fatal        2020-11-27T17:56:37+08:00   Get https://192.168.13.8:2379/health: dial tcp 192.168.13.8:2379: connect: connection refused
```
您可以参考常见[FAQ](https://github.com/kubesphere/kubeeye/blob/main/docs/FAQ.md)内容来优化您的集群。

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
| | ETCDHealthStatus               | etcd 是否正常运行，请检查 etcd 状态 |
| | ControllerManagerHealthStatus  | kube-controller-manager 是否正常运行，请检查 kube-controller-manager 状态 |
| | SchedulerHealthStatus          | kube-scheduler 是否正常运行，请检查 kube-scheduler 状态 |           
| :white_check_mark: | NodeMemory                     | 节点内存使用是否超过阀值，请检查节点内存使用情况 |
| :white_check_mark: | DockerHealthStatus             | docker 是否正常运行，请检查 docker 状态 |             
| :white_check_mark: | NodeDisk                       | 节点磁盘使用是否超过阀值，请检查节点磁盘使用情况 |
| :white_check_mark: | KubeletHealthStatus            | kubelet 是否正常运行，请检查 kubelet 状态 |            
| :white_check_mark: | NodeCPU                        | 节点 CPU 使用是否超过阀值，请检查节点 CPU 使用情况 |
| :white_check_mark: | NodeCorruptOverlay2            | Overlay2 不可用 |            
| :white_check_mark: | NodeKernelNULLPointer          | 节点未准备就绪 |
| :white_check_mark: | NodeDeadlock                   | 死锁是指两个或两个以上的进程在争夺资源时互相等待的现象|
| :white_check_mark: | NodeOOM                        | 监控那些消耗过多内存的进程，尤其是那些消耗大量内存非常快的进程，内核会杀掉它们，防止它们耗尽内存 |
| :white_check_mark: | NodeExt4Error                  | Ext4 挂载失败 |                  
| :white_check_mark: | NodeTaskHung                   | 检查是否有持续超过120s的 D 状态进程|
| :white_check_mark: | NodeUnregisterNetDevice        | 检查节点网络|    
| :white_check_mark: | NodeCorruptDockerImage         | 检查 docker 镜像|
| :white_check_mark: | NodeAUFSUmountHung             | 检查存储 |
| :white_check_mark: | NodeDockerHung                 | Docker hang 住, 请检查 docker 的日志 |
| :white_check_mark: | PodSetImagePullBackOff          | Pod 无法正确拉出镜像，因此可以在相应节点上手动拉出镜像 |         
| :white_check_mark: | PodNoSuchFileOrDirectory        | 进入容器查看相应文件是否存在 |
| :white_check_mark: | PodIOError                      | 这通常是由于文件 IO 性能瓶颈 |
| :white_check_mark: | PodNoSuchDeviceOrAddress        | 检查网络 |
| :white_check_mark: | PodInvalidArgument              | 检查存储 |              
| :white_check_mark: | PodDeviceOrResourceBusy         | 检查对应的目录和 PID|
| :white_check_mark: | PodFileExists                   | 检查文件 |             
| :white_check_mark: | PodTooManyOpenFiles             | 程序打开的文件/套接字连接数超过系统设置值 |
| :white_check_mark: | PodNoSpaceLeftOnDevice          | 检查磁盘和索引节点的使用情况 |
| :white_check_mark: | NodeApiServerExpiredPeriod      | 将检查 ApiServer 证书的到期日期少于30天 |
|                    | NodeNotReadyAndUseOfClosedNetworkConnection        | 节点网络连接异常 |
|                    | NodeNotReady        | 无法启动 ContainerManager 无法设置属性 TasksAccounting 或未知属性 |

> 未标注的项目正在开发中

## 增加自定义命令
``` text
├── cmd
│   └── testcmd.go
├── main.go
```
testcmd.go
```go
package cmd
import (
	"fmt"
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
	Long:  `new command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("new command")
	},
}
```
main.go
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	kc "github.com/leonharetd/kubeeye_sample/cmd"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithCommand(kc.TestCmd).DO()
	cmd.Execute()
}
```
重新构建之后
```shell
>> kubeeye audit
KubeEye finds various problems on Kubernetes cluster.

Usage:
  ke [command]

Available Commands:
  audit       audit resources from the cluster
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  install     A brief description of your command
  test        test
  uninstall   A brief description of your command
```

## 添加自定义规则
### 嵌入式规则支持
>嵌入式规则，将规则打包至kubeeye中,方便使用.
- OPA 规则
- Function 规则

Function 检查规则，提供更加定制化的规则检查，如使用shell，调用第三方接口，都可以封入Function，按照约定的格式返回输出，即可在报表中统一展示。
### 非嵌入式规则
>命令和规则分离式管理，指定外部OPA规则目录，kubeeye加载目录中的规则并且和默认规则进行合并。
- OPA 规则
### OPA
- 添加OPA规则文件
> opa package Note: 包名必须是下面中的一个 

|type|package|
|---|---|
|RBAC |kubeeye_RBAC_rego|
|workloads|kubeeye_workloads_rego|
|nodes|kubeeye_nodes_rego|
|events|kubeeye_events_rego|

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
### 嵌入式OPA规则
``` text
├── main.go
└── regorules
    ├── rules
    │   ├── imageRegistryRule.rego
    │   └── testRule.rego
    └── testrule.go
```
testrule.go

specify embed directory
``` go
package regorules

import (
	"embed"
)

//go:embed rules
var EmbRegoRules embed.FS
```
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	"github.com/leonharetd/kubeeye_sample/regorules"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithRegoRule(regorules.EmbRegoRules).DO()
	cmd.Execute()
}
```
如果有多个规则文件
``` go 
cmd := cmd.NewKubeEyeCommand().WithRegoRule(RulesA).WithRegoRule(RulesB).DO()
```
编译后执行
```shell
kubeeye audit
```
### 嵌入式函数规则
github.com/leonharetd/kubeeye_sample/expirerules/expirerule.go
```go
package funcrules

import (
	"fmt"
	"strconv"
	kube "github.com/leonharetd/kubeeye/pkg/kube"
)

type ExpireTestRule struct{}

func (cer ExpireTestRule) Exec() kube.ValidateResults {
	output := kube.ValidateResults{ValidateResults: make([]kube.ResultReceiver, 0)}
	var certExpiresOutput kube.ResultReceiver
	for i := range []int{1, 2, 3, 4} {
		certExpiresOutput.Name = fmt.Sprint("test", strconv.Itoa(i))
		certExpiresOutput.Type = "testExpire"
		certExpiresOutput.Message = []string{strconv.Itoa(i), "expire"}
		output.ValidateResults = append(output.ValidateResults, certExpiresOutput)
	}
	return output
}
```
main.go
``` go
package main

import (
	"github.com/leonharetd/kubeeye/cmd"
	"github.com/leonharetd/kubeeye_sample/funcrules"
)

func main() {
	cmd := cmd.NewKubeEyeCommand().WithFuncRule(funcrules.FuncTestRule{}).DO()
	cmd.Execute()
}
```
编译后运行
```shell
kubeeye audit
```
### 添加非嵌入式自定义 OPA 检查规则
- 创建 OPA 规则存放目录
``` shell
mkdir opa
```
- 添加自定义 OPA 规则文件
> 注意：为检查工作负载设置的 OPA 规则 package 名称必须是 *kubeeye_workloads_rego*, 
> 为检查 RBAC 设置的 OPA 规则 package 名称必须是 *kubeeye_RBAC_rego*，为检查节点设置的 OPA 规则 package 名称必须是 *kubeeye_nodes_rego*

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
```shell
root:# kubeeye audit -p ./opa -f ~/.kube/config
NAMESPACE     NAME              KIND          MESSAGE
default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
```

## 文档

* [RoadMap](docs/roadmap.md)
