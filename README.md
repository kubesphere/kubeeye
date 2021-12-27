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
| :white_check_mark: | InsecureCapabilities           | Have insecure options in capabilities such as KILL/SYS_CHROOT/CHOWN |
| :white_check_mark: | PodNoSuchFileOrDirectory        | Go into the container to see if the corresponding file exists|
| :white_check_mark: | PodIOError                      | This is usually due to file IO performance bottlenecks|
| :white_check_mark: | PodNoSuchDeviceOrAddress        | Check corresponding net|                 
| :white_check_mark: | PodNoSpaceLeftOnDevice          | Check for disk and inode usage|
| :white_check_mark: | NodeApiServerExpiredPeriod      | ApiServer certificate expiration date less than 30 days will be checked|
|                    | NodeNotReadyAndUseOfClosedNetworkConnection        | http2-max-streams-per-connection |
|                    | NodeNotReady        | Failed to start ContainerManager Cannot set property TasksAccounting, or unknown property |
> [more](./docs/check-content_zh-EN.md)


## Add your own custom rules(./docs/custom-rules.md)
### embed rules
>Embedded rules, package the rules into kubeeye for easy use
- [OPA rules](./docs/custom-rules.md)
- [Function rules](./docs/custom-rules.md)
  
Function check rules provide more customized rule checks. For example, by using a shell and calling a third-party interface, you can enclose the function and return the output according to the agreed format, which can be displayed uniformly in the report.

### non-self-embeding rules
>Separate management of commands and rules, specify the external OPA rule directory, kubeeye load the rules in the directory and merge them with the default rules.
- [OPA  rules](./docs/custom-rules.md)

### OPA
- Add custom OPA rules files
> opa package Note: package name must be select from tabel 

|type|package|
|---|---|
|RBAC |kubeeye_RBAC_rego|
|workloads|kubeeye_workloads_rego|
|nodes|kubeeye_nodes_rego|
|events|kubeeye_events_rego|
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

### non-self-embeding custom OPA rules
- create a directory for OPA rules
``` shell
mkdir opa
```
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
