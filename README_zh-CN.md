# KubeEye

KubeEye旨在发现Kubernetes上的各种问题，比如应用配置错误（使用[Polaris](https://github.com/FairwindsOps/polaris)）、集群组件不健康和节点问题（使用[Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector)）。除了预定义的规则，它还支持自定义规则。

## 架构图

KubeEye通过调用Kubernetes API，通过常规匹配日志中的关键错误信息和容器语法的规则匹配来获取集群诊断数据，详见架构。
![Image](./docs/KubeEye-Architecture.jpg?raw=true)

## 怎么使用

- 机器上安装KubeEye
  - 从 [Releases](https://github.com/kubesphere/kubeeye/releases) 中下载预构建的可执行文件。
  - 或者你也可以从源代码构建
  ```
  git clone https://github.com/kubesphere/kubeeye.git
  cd kubeeye 
  make install
  ```
- [可选] 安装 Node-problem-Detector  
注意：这一行将在你的集群上安装npd，只有当你想要详细的报告时才需要。  
`ke install npd`  

- KubeEye执行
```
root@node1:# ke diag
NODENAME        SEVERITY     HEARTBEATTIME               REASON              MESSAGE
node18          Fatal        2020-11-19T10:32:03+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node19          Fatal        2020-11-19T10:31:37+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node2           Fatal        2020-11-19T10:31:14+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node3           Fatal        2020-11-27T17:36:53+08:00   KubeletNotReady     Container runtime not ready: RuntimeReady=false reason:DockerDaemonNotReady message:docker: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

NAME            SEVERITY     TIME                        MESSAGE
scheduler       Fatal        2020-11-27T17:09:59+08:00   Get http://127.0.0.1:10251/healthz: dial tcp 127.0.0.1:10251: connect: connection refused
etcd-0          Fatal        2020-11-27T17:56:37+08:00   Get https://192.168.13.8:2379/health: dial tcp 192.168.13.8:2379: connect: connection refused

NAMESPACE       SEVERITY     PODNAME                                          EVENTTIME                   REASON                MESSAGE
default         Warning      node3.164b53d23ea79fc7                           2020-11-27T17:37:34+08:00   ContainerGCFailed     rpc error: code = Unknown desc = Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
default         Warning      node3.164b553ca5740aae                           2020-11-27T18:03:31+08:00   FreeDiskSpaceFailed   failed to garbage collect required amount of images. Wanted to free 5399374233 bytes, but freed 416077545 bytes
default         Warning      nginx-b8ffcf679-q4n9v.16491643e6b68cd7           2020-11-27T17:09:24+08:00   Failed                Error: ImagePullBackOff
default         Warning      node3.164b5861e041a60e                           2020-11-27T19:01:09+08:00   SystemOOM             System OOM encountered, victim process: stress, pid: 16713
default         Warning      node3.164b58660f8d4590                           2020-11-27T19:01:27+08:00   OOMKilling            Out of memory: Kill process 16711 (stress) score 205 or sacrifice child Killed process 16711 (stress), UID 0, total-vm:826516kB, anon-rss:819296kB, file-rss:0kB, shmem-rss:0kB
insights-agent  Warning      workloads-1606467120.164b519ca8c67416            2020-11-27T16:57:05+08:00   DeadlineExceeded      Job was active longer than specified deadline
kube-system     Warning      calico-node-zvl9t.164b3dc50580845d               2020-11-27T17:09:35+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system     Warning      kube-proxy-4bnn7.164b3dc4f4c4125d                2020-11-27T17:09:09+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system     Warning      nodelocaldns-2zbhh.164b3dc4f42d358b              2020-11-27T17:09:14+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29


NAMESPACE       SEVERITY     NAME                      KIND         TIME                        MESSAGE
kube-system     Warning      node-problem-detector     DaemonSet    2020-11-27T17:09:59+08:00   [livenessProbeMissing runAsPrivileged]
kube-system     Warning      calico-node               DaemonSet    2020-11-27T17:09:59+08:00   [runAsPrivileged cpuLimitsMissing]
kube-system     Warning      nodelocaldns              DaemonSet    2020-11-27T17:09:59+08:00   [cpuLimitsMissing runAsPrivileged]
default         Warning      nginx                     Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing livenessProbeMissing tagNotSpecified]
insights-agent  Warning      workloads                 CronJob      2020-11-27T17:09:59+08:00   [livenessProbeMissing]
insights-agent  Warning      cronjob-executor          Job          2020-11-27T17:09:59+08:00   [livenessProbeMissing]
kube-system     Warning      calico-kube-controllers   Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing livenessProbeMissing]
kube-system     Warning      coredns                   Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing]   
```
您可以参考常见[FAQ](https://github.com/kubesphere/kubeeye/blob/main/docs/FAQ.md)内容来优化您的集群。

## KubeEye能做什么

- KubeEye可以发现你的集群控制平面的问题，包括kube-apiserver/kube-controller-manager/etcd等。
- KubeEye可以帮助你检测各种节点问题，包括内存/CPU/磁盘压力，意外的内核错误日志等。
- KubeEye根据行业最佳实践验证你的工作负载yaml规范，帮助你使你的集群稳定。

## 检查项

|是/否|检查项 |描述|
|---|---|---|
| :white_check_mark: | ETCDHealthStatus | 如果 etcd 启动并正常运行 |
| :white_check_mark: | ControllerManagerHealthStatus | 如果kubernetes kube-controller-manager正常启动并运行 |
| :white_check_mark: | SchedulerHealthStatus | 如果kubernetes kube-schedule正常启动并运行 |       
| :white_check_mark: | NodeMemory | 如果节点内存使用量超过阈值 |
| :white_check_mark: | DockerHealthStatus | 如果docker正常运行|                                            
| :white_check_mark: | NodeDisk | 如果节点磁盘使用量超过阈值 |
| :white_check_mark: | KubeletHealthStatus | 如果kubelet激活状态且正常运行 |
| :white_check_mark: | NodeCPU | 如果节点CPu使用量超过阈值                                                                        |
| :white_check_mark: | NodeCorruptOverlay2 | Overlay2 不可用|                                                                                 
| :white_check_mark: | NodeKernelNULLPointer | node 显示NotReady|
| :white_check_mark: | NodeDeadlock | 死锁是指两个或两个以上的进程在争夺资源时互相等待的现象。|                                                                               
| :white_check_mark: | NodeOOM | 监控那些消耗过多内存的进程，尤其是那些消耗大量内存非常快的进程，内核会杀掉它们，防止它们耗尽内存|
| :white_check_mark: | NodeExt4Error | Ext4 挂载失败|
| :white_check_mark: | NodeTaskHung | 检查D状态下是否有超过120s的进程|
| :white_check_mark: | NodeUnregisterNetDevice | 检查对应网络|
| :white_check_mark: | NodeCorruptDockerImage          | 检查docker镜像|
| :white_check_mark: | NodeAUFSUmountHung            |  检查存储|
| :white_check_mark: | NodeDockerHung                  | Docker hang住, 检查docker的日志|
| :white_check_mark: | PodSetLivenessProbe |如果为pod中的每一个容器设置了livenessProbe|
| :white_check_mark: | PodSetTagNotSpecified | 镜像地址没有声明标签或标签是最新|
| :white_check_mark: | PodSetRunAsPrivileged | 以特权模式运行Pod意味着Pod可以访问主机的资源和内核功能|
| :white_check_mark: | PodSetImagePullBackOff          | Pod无法正确拉出镜像，因此可以在相应节点上手动拉出镜像|
| :white_check_mark: | PodSetImageRegistry             | 检查镜像形式是否在相应仓库|
| :white_check_mark: | PodSetCpuLimitsMissing          |  未声明CPU资源限制|
| :white_check_mark: | PodNoSuchFileOrDirectory        | 进入容器查看相应文件是否存在|
| :white_check_mark: | PodIOError                      | 这通常是由于文件IO性能瓶颈|
| :white_check_mark: | PodNoSuchDeviceOrAddress        | 检查对应网络                                                                        |
| :white_check_mark: | PodInvalidArgument              | 检查对应存储|                                                                             
| :white_check_mark: | PodDeviceOrResourceBusy         | 检查对应的目录和PID|
| :white_check_mark: | PodFileExists                   | 检查现有文件|
| :white_check_mark: | PodTooManyOpenFiles             | 程序打开的文件/套接字连接数超过系统设置值|
| :white_check_mark: | PodNoSpaceLeftOnDevice          | 检查磁盘和索引节点的使用情况|
| :white_check_mark: | NodeApiServerExpiredPeriod      | 将检查ApiServer证书的到期日期少于30天|
| :white_check_mark: | PodSetCpuRequestsMissing        | 未声明CPU资源请求值|
| :white_check_mark: | PodSetHostIPCSet                | 设置主机IP|
| :white_check_mark: | PodSetHostNetworkSet            | 设置主机网络|
| :white_check_mark: | PodHostPIDSet                   | 设置主机PID|
| :white_check_mark: | PodMemoryRequestsMiss           | 没有声明内存资源请求值|
| :white_check_mark: | PodSetHostPort                  | 设置主机端口|
| :white_check_mark: | PodSetMemoryLimitsMissing       | 没有声明内存资源限制值|
| :white_check_mark: | PodNotReadOnlyRootFiles         | 文件系统未设置为只读|
| :white_check_mark: | PodSetPullPolicyNotAlways       | 镜像拉策略并非总是如此|
| :white_check_mark: | PodSetRunAsRootAllowed          | 以root用户执行|
| :white_check_mark: | PodDangerousCapabilities        | 您在ALL / SYS_ADMIN / NET_ADMIN等功能中有危险的选择|
| :white_check_mark:|PodlivenessProbeMissing|未声明ReadinessProbe|
| :white_check_mark: | privilegeEscalationAllowed        | 允许特权升级|
|                    | NodeNotReadyAndUseOfClosedNetworkConnection        | http                                                                        2-max-streams-per-connection |
|                    | NodeNotReady        | 无法启动ContainerManager无法设置属性TasksAccounting或未知属性 |

未标注的项目正在开发中

## 添加自定义检查规则

### 添加npd自定义检查规则

- 安装 NPD 指令 `ke install npd`
- 由kubectl编辑 configmap kube-system/node-problem-detector-config,
``` 
kubectl edit cm -n kube-system node-problem-detector-config
```
- 在configMap的规则下添加异常日志信息，规则遵循正则表达式。

### 自定义最佳实践规则

- 准备一个规则yaml，例如，下面的规则将验证你的pod规范，以确保镜像只来自授权的注册处。
```
checks:
  imageFromUnauthorizedRegistry: warning

customChecks:
  imageFromUnauthorizedRegistry:
    promptMessage: When the corresponding rule does not match. Show that image from an unauthorized registry.
    category: Images
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      properties:
        image:
          type: string
          not:
            pattern: ^quay.io
```
- 将上述规则保存为yaml，例如，rule.yaml。
- 用 rule.yaml 运行 KubeEye。
```
root:# ke diag -f rule.yaml --kubeconfig ~/.kube/config
NAMESPACE     SEVERITY    NAME                      KIND         TIME                        MESSAGE
default       Warning     nginx                     Deployment   2020-11-27T17:18:31+08:00   [imageFromUnauthorizedRegistry]
kube-system   Warning     node-problem-detector     DaemonSet    2020-11-27T17:18:31+08:00   [livenessProbeMissing runAsPrivileged]
kube-system   Warning     calico-node               DaemonSet    2020-11-27T17:18:31+08:00   [cpuLimitsMissing runAsPrivileged]
kube-system   Warning     calico-kube-controllers   Deployment   2020-11-27T17:18:31+08:00   [cpuLimitsMissing livenessProbeMissing]
kube-system   Warning     nodelocaldns              DaemonSet    2020-11-27T17:18:31+08:00   [runAsPrivileged cpuLimitsMissing]
default       Warning     nginx                     Deployment   2020-11-27T17:18:31+08:00   [livenessProbeMissing cpuLimitsMissing]
kube-system   Warning     coredns                   Deployment   2020-11-27T17:18:31+08:00   [cpuLimitsMissing]
```
