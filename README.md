# KubeEye
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-8-orange.svg?style=flat-square)](#contributors-)
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
   make installke
   ```
   
- [Optional] Install [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector)
> Note: This will install npd on your cluster, only required if you want detailed report.

```shell
kubeeye install npd
```
- Run KubeEye
> Note: The results of kubeeye sort by resource kind.

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

## What KubeEye can do

- KubeEye validates your workloads yaml specs against industry best practice, helps you make your cluster stable.
- KubeEye can find problems of your cluster control plane, including kube-apiserver/kube-controller-manager/etcd, etc.
- KubeEye helps you detect all kinds of node problems, including memory/cpu/disk pressure, unexpected kernel error logs, etc.

## Checklist

|YES/NO|CHECK ITEM |Description|
|---|---|---|
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
| :white_check_mark: | CertificateExpiredPeriod       | Certificate expiration date less than 30 days |
| :white_check_mark: | EventAudit                     | Event audit  |
| :white_check_mark: | NodeStatus                     | node status audit |
| :white_check_mark: | DockerStatus                   | docker status audit |             
| :white_check_mark: | KubeletStatus                  | kubelet status audit |

## Add your own audit rules
### Add custom OPA rules
- create a directory for OPA rules
```shell
mkdir opa
```
- Add custom OPA rules files
> Note: the OPA rule for workloads, package name must be *kubeeye_workloads_rego*
> for RBAC, package name must be *kubeeye_RBAC_rego*
> for nodes, package name must be *kubeeye_nodes_rego*

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
root:# kubeeye audit -p ./opa
NAMESPACE     NAME              KIND          MESSAGE
default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
```

### Add custom NPD rules
- edit configmap
```shell
kubectl edit ConfigMap node-problem-detector-config -n kube-system 
```
- restart NPD deployment
```shell
kubectl rollout restart DaemonSet node-problem-detector -n kube-system
```

## deploy in Kubernetes
### deploy Kubeeye
```shell
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye.yaml
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye_insights.yaml
```
### get the audit results
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
            - level: waring
              message: KubeletHasNoSufficientMemory
              reason: kubelet has no sufficient memory available
            - level: waring
              message: KubeletHasNoSufficientPID
              reason: kubelet has no sufficient PID available
            - level: waring
              message: KubeletHasDiskPressure
              reason: kubelet has disk pressure
            name: kubeeyeNode
```

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/ruiyaoOps"><img src="https://avatars.githubusercontent.com/u/35256376?v=4?s=100" width="100px;" alt=""/><br /><sub><b>ruiyaoOps</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=ruiyaoOps" title="Code">💻</a> <a href="https://github.com/kubesphere/kubeeye/commits?author=ruiyaoOps" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/Forest-L"><img src="https://avatars.githubusercontent.com/u/50984129?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Forest</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=Forest-L" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/zryfish"><img src="https://avatars.githubusercontent.com/u/3326354?v=4?s=100" width="100px;" alt=""/><br /><sub><b>zryfish</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=zryfish" title="Documentation">📖</a></td>
    <td align="center"><a href="https://www.chenshaowen.com/"><img src="https://avatars.githubusercontent.com/u/43693241?v=4?s=100" width="100px;" alt=""/><br /><sub><b>shaowenchen</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=shaowenchen" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/pixiake"><img src="https://avatars.githubusercontent.com/u/22290449?v=4?s=100" width="100px;" alt=""/><br /><sub><b>pixiake</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=pixiake" title="Documentation">📖</a></td>
    <td align="center"><a href="https://kubesphere.io"><img src="https://avatars.githubusercontent.com/u/40452856?v=4?s=100" width="100px;" alt=""/><br /><sub><b>pengfei</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=FeynmanZhou" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/RealHarshThakur"><img src="https://avatars.githubusercontent.com/u/38140305?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Harsh Thakur</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=RealHarshThakur" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/leonharetd"><img src="https://avatars.githubusercontent.com/u/10416045?v=4?s=100" width="100px;" alt=""/><br /><sub><b>leonharetd</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=leonharetd" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## Documents

* [RoadMap](docs/roadmap.md)
* [FAQ](docs/FAQ.md)
