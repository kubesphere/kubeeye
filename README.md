# KubeEye
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-6-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

![kubeeye-logo](./docs/images/kubeeye-logo.png?raw=true)

> English | [中文](README_zh.md)

KubeEye aims to find various problems on Kubernetes, such as application misconfiguration(using [OPA](https://github.com/open-policy-agent/opa)), cluster components unhealthy and node problems(using [Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector)). Besides predefined rules, it also supports custom defined rules.

## Architecture
KubeEye gets cluster diagnostic data by calling the Kubernetes API, by regular matching of key error messages in resources and by rule matching of container syntax. See Architecture for details.

![kubeeye-architecture](./docs/images/kubeeye-architecture.svg?raw=true)

## How to use
-  Install KubeEye on your machine
    - Download pre built executables from [Releases](https://github.com/kubesphere/kubeeye/releases).
    
    - Or you can build from source code
> Note: make install will create kubeeye in /usr/local/bin/ on your machine.

    ```shell
    git clone https://github.com/kubesphere/kubeeye.git
    cd kubeeye 
    make install
    ```
   
- [Optional] Install [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector)
> Note: This line will install npd on your cluster, only required if you want detailed report.

```shell
kubeeye install -e npd
```
- Run KubeEye
> Note: The results of kubeeye sort by resource kind.

```shell
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
> You can refer to the [FAQ](./docs/FAQ.md) content to optimize your cluster.

## What KubeEye can do

- KubeEye validates your workloads yaml specs against industry best practice, helps you make your cluster stable.
- KubeEye can find problems of your cluster control plane, including kube-apiserver/kube-controller-manager/etcd, etc.
- KubeEye helps you detect all kinds of node problems, including memory/cpu/disk pressure, unexpected kernel error logs, etc.

## Checklist
|YES/NO|CHECK ITEM |Description|
|---|---|---|
| :white_check_mark: | NodeDockerHung                 | Docker hung, you can check docker log|
| :white_check_mark: | PrivilegeEscalationAllowed     | Privilege escalation is allowed |
| :white_check_mark: | CanImpersonateUser             | The role/clusterrole can impersonate other user |
| :white_check_mark: | CanDeleteResources             | The role/clusterrole can delete kubernetes resources |
| :white_check_mark: | CanModifyWorkloads             | The role/clusterrole can modify kubernetes workloads |
| :white_check_mark: | NoCPULimits                    | The resource does not set limits of CPU in containers.resources |
| :white_check_mark: | NoCPURequests                  | The resource does not set requests of CPU in containers.resources |
| :white_check_mark: | HighRiskCapabilities           | Have high-Risk options in capabilities such as ALL/SYS_ADMIN/NET_ADMIN |
| :white_check_mark: | HostIPCAllowed                 | HostIPC Set to true |
| :white_check_mark: | HostNetworkAllowed             | HostNetwork Set to true |
| :white_check_mark: | HostPIDAllowed                 | HostPID Set to true |
| :white_check_mark: | HostPortAllowed                | HostPort Set to true |
| :white_check_mark: | ImagePullPolicyNotAlways       | Image pull policy not always |
| :white_check_mark: | ImageTagIsLatest               | The image tag is latest |
| :white_check_mark: | ImageTagMiss                   | The image tag do not declare |
| :white_check_mark: | InsecureCapabilities           | Have insecure options in capabilities such as KILL/SYS_CHROOT/CHOWN |
| :white_check_mark: | NoLivenessProbe                | The resource does not set livenessProbe |
| :white_check_mark: | NoMemoryLimits                 | The resource does not set limits of memory in containers.resources |
| :white_check_mark: | NoMemoryRequests               | The resource does not set requests of memory in containers.resources |
| :white_check_mark: | NoPriorityClassName            | The resource does not set priorityClassName |
| :white_check_mark: | PrivilegedAllowed              | Running a pod in a privileged mode means that the pod can access the host’s resources and kernel capabilities |
| :white_check_mark: | NoReadinessProbe               | The resource does not set readinessProbe |
| :white_check_mark: | NotReadOnlyRootFilesystem      | The resource does not set readOnlyRootFilesystem to true |
| :white_check_mark: | NotRunAsNonRoot                | The resource does not set runAsNonRoot to true, maybe executed run as a root account |
| :white_check_mark: | ETCDHealthStatus               | if etcd is up and running normally, please check etcd status |
| :white_check_mark: | ControllerManagerHealthStatus  | if kubernetes kube-controller-manager is up and running normally, please check kube-controller-manager status |
| :white_check_mark: | SchedulerHealthStatus          | if kubernetes kube-scheduler is up and running normally, please check kube-scheduler status |           
| :white_check_mark: | NodeMemory                     | if node memory usage is above threshold, please check node memory usage |
| :white_check_mark: | DockerHealthStatus             | if docker is up and running, please check docker status |             
| :white_check_mark: | NodeDisk                       | if node disk usage is above given threshold, please check node disk usage |
| :white_check_mark: | KubeletHealthStatus            | if kubelet is active and running normally |            
| :white_check_mark: | NodeCPU                        | if node cpu usage is above the given threshold |
| :white_check_mark: | NodeCorruptOverlay2            | Overlay2 is not available|            
| :white_check_mark: | NodeKernelNULLPointer          | the node displays NotReady|
| :white_check_mark: | NodeDeadlock                   | A deadlock is a phenomenon in which two or more processes are waiting for each other as they compete for resources|                  
| :white_check_mark: | NodeOOM                        | Monitor processes that consume too much memory, especially those that consume a lot of memory very quickly, and the kernel kill them to prevent them from running out of memory|
| :white_check_mark: | NodeExt4Error                  | Ext4 mount error|                  
| :white_check_mark: | NodeTaskHung                   | Check to see if there is a process in state D for more than 120s|
| :white_check_mark: | NodeUnregisterNetDevice        | Check corresponding net|    
| :white_check_mark: | NodeCorruptDockerImage         | Check docker image|
| :white_check_mark: | NodeAUFSUmountHung             | Check storage|
| :white_check_mark: | PodSetImagePullBackOff          | Pod can't pull the image properly, so it can be pulled manually on the corresponding node|         
| :white_check_mark: | PodNoSuchFileOrDirectory        | Go into the container to see if the corresponding file exists|
| :white_check_mark: | PodIOError                      | This is usually due to file IO performance bottlenecks|
| :white_check_mark: | PodNoSuchDeviceOrAddress        | Check corresponding net|
| :white_check_mark: | PodInvalidArgument              | Check the storage|              
| :white_check_mark: | PodDeviceOrResourceBusy         | Check corresponding dirctory and PID|
| :white_check_mark: | PodFileExists                   | Check for existing files|             
| :white_check_mark: | PodTooManyOpenFiles             | The number of file /socket connections opened by the program exceeds the system set value|
| :white_check_mark: | PodNoSpaceLeftOnDevice          | Check for disk and inode usage|
| :white_check_mark: | NodeApiServerExpiredPeriod      | ApiServer certificate expiration date less than 30 days will be checked|
|                    | NodeNotReadyAndUseOfClosedNetworkConnection        | http2-max-streams-per-connection |
|                    | NodeNotReady        | Failed to start ContainerManager Cannot set property TasksAccounting, or unknown property |
> unmarked items are under heavy development


## Add your own audit rules
### Add custom OPA rules
- create a directory for OPA rules
``` shell
mkdir opa
```
- Add custom OPA rules files
> Note: the OPA rule for workloads package name must be *kubeeye_workloads_rego*,
> for RBAC package name must be *kubeeye_RBAC_rego*, for nodes package name must be *kubeeye_nodes_rego*.

- Save the following rule to rule file such as *imageRegistryRule.rego* for audit the image registry address complies with rules.
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

- Run KubeEye with custom rules
> Note: Specify the path then Kubeeye will read all files in the directory that end with *.rego*.

```shell
root:# kubeeye audit -p ./opa -f ~/.kube/config
NAMESPACE     NAME              KIND          MESSAGE
default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
```

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/ruiyaoOps"><img src="https://avatars.githubusercontent.com/u/35256376?v=4?s=100" width="100px;" alt=""/><br /><sub><b>ruiyaoOps</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=ruiyaoOps" title="Code">💻</a> <a href="https://github.com/kubesphere/kubeeye/commits?author=ruiyaoOps" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/Forest-L"><img src="https://avatars.githubusercontent.com/u/50984129?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Forest</b></sub></a><br /> <a href="https://github.com/kubesphere/kubeeye/commits?author=Forest-L" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/zryfish"><img src="https://avatars.githubusercontent.com/u/3326354?v=4?s=100" width="100px;" alt=""/><br /><sub><b>zryfish</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=zryfish" title="Documentation">📖</a></td>
    <td align="center"><a href="https://www.chenshaowen.com/"><img src="https://avatars.githubusercontent.com/u/43693241?v=4?s=100" width="100px;" alt=""/><br /><sub><b>shaowenchen</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=shaowenchen" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/pixiake"><img src="https://avatars.githubusercontent.com/u/22290449?v=4?s=100" width="100px;" alt=""/><br /><sub><b>pixiake</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=pixiake" title="Documentation">📖</a></td>
    <td align="center"><a href="https://kubesphere.io"><img src="https://avatars.githubusercontent.com/u/40452856?v=4?s=100" width="100px;" alt=""/><br /><sub><b>pengfei</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=FeynmanZhou" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/RealHarshThakur"><img src="https://avatars.githubusercontent.com/u/38140305?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Harsh Thakur</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=RealHarshThakur" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## Documents

* [RoadMap](docs/roadmap.md)
