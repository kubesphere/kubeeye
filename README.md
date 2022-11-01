<div align=center><img src="docs/images/KubeEye-O.svg?raw=true"></div>

<p align=center>
<a href="https://github.com/kubesphere/kubeeye/actions?query=event%3Apush+branch%3Amain+workflow%3ACI+"><img src="https://github.com/kubesphere/kubeeye/workflows/CI/badge.svg?branch=main&event=push"></a>
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="https://github.com/kubesphere/kubeeye#contributors-"><img src="https://img.shields.io/badge/all_contributors-10-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

> English | [中文](README_zh.md)

KubeEye is an inspection tool for Kubernetes. It discovers whether Kubernetes resources (by using [OPA](https://github.com/open-policy-agent/opa) ), cluster components, cluster nodes (by using [Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector)), and other configurations comply with best practices and makes modification suggestions accordingly.

KubeEye supports custom inspection rules and plugin installation. With [KubeEye Operator](#kubeeye-operator), you can intuitively view the inspection results and modification suggestions on the web console.

## Architecture
KubeEye obtains cluster resource details by using Kubernetes APIs, inspects resource configurations by using inspection rules and plugins, and generates inspection results. The architecture of KubeEye is as follows:

![kubeeye-architecture](./docs/images/kubeeye-architecture.svg?raw=true)

## Install and use KubeEye

1. Install KubeEye on your machine.

   - Method 1: Download the pre-built executable file from [Releases](https://github.com/kubesphere/kubeeye/releases).

   - Method 2: Build from the source code.
   > Note: KubeEye files will be generated in `/usr/local/bin/` on your machine.

   ```shell
   git clone https://github.com/kubesphere/kubeeye.git
   cd kubeeye
   make installke
   ```

2. (Optional) Install [Node-problem-Detector](https://github.com/kubernetes/node-problem-detector).

   > Note: If you need detailed reports, run the following command, and then NPD will be installed on your cluster.

   ```shell
   kubeeye install npd
   ```
3. Run KubeEye to inspect clusters.

> Note: The results of KubeEye are sorted by resource kind.

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

## How KubeEye can help you

- It inspects cluster resources according to Kubernetes best practices to ensure that clusters run stably.
- It detects the control plane problems of the cluster, including kube-apiserver, kube-controller-manager, and etcd.
- It detects node problems, including memory, CPU, disk pressure, and unexpected kernel error logs.

## Checklist

|Yes/No |Check Item |Description |Severity |
|---|---|---|---|
| :white_check_mark: | PrivilegeEscalationAllowed     | Privilege escalation is allowed. | danger |
| :white_check_mark: | CanImpersonateUser             | The Role/ClusterRole can impersonate users. | warning |
| :white_check_mark: | CanModifyResources             | The Role/ClusterRole can delete Kubernetes resources. | warning |
| :white_check_mark: | CanModifyWorkloads             | The Role/ClusterRole can modify Kubernetes resources. | warning |
| :white_check_mark: | NoCPULimits                    | No CPU limits are set. | danger |
| :white_check_mark: | NoCPURequests                  | No CPU resources are reserved. | danger |
| :white_check_mark: | HighRiskCapabilities           | High-risk features, such as ALL, SYS_ADMIN, and NET_ADMIN, are enabled. | danger |
| :white_check_mark: | HostIPCAllowed                 | HostIPC is set to `true`. | danger |
| :white_check_mark: | HostNetworkAllowed             | HostNetwork is set to `true`. | danger |
| :white_check_mark: | HostPIDAllowed                 | HostPID is set to `true`. | danger |
| :white_check_mark: | HostPortAllowed                | HostPort is set to `true`. | danger |
| :white_check_mark: | ImagePullPolicyNotAlways       | The image pull policy is not set to `always`. | warning |
| :white_check_mark: | ImageTagIsLatest               | The image tag is `latest`. | warning |
| :white_check_mark: | ImageTagMiss                   | The image tag is missing. | danger |
| :white_check_mark: | InsecureCapabilities           | Insecure options are missing, such as KILL, SYS_CHROOT, and CHOWN. | danger |
| :white_check_mark: | NoLivenessProbe                | Liveless probe is not set. | warning |
| :white_check_mark: | NoMemoryLimits                 | No memory limits are set. | danger |
| :white_check_mark: | NoMemoryRequests               | No memory resources are reserved. | danger |
| :white_check_mark: | NoPriorityClassName            | Resource scheduling priority is not set. | ignore |
| :white_check_mark: | PrivilegedAllowed              | Pods are running in the privileged mode. | danger |
| :white_check_mark: | NoReadinessProbe               | Readiness probe is not set. | warning |
| :white_check_mark: | NotReadOnlyRootFilesystem      | readOnlyRootFilesystem is not set to `true`. | warning |
| :white_check_mark: | NotRunAsNonRoot                | runAsNonRoot is not set to `true`. | warning |
| :white_check_mark: | CertificateExpiredPeriod       | The certificate expiry date of the API Server is less than 30 days. | danger |
| :white_check_mark: | EventAudit                     | Events need to be audited. | warning |
| :white_check_mark: | NodeStatus                     | Node status needs to be checked. | warning |
| :white_check_mark: | DockerStatus                   | Docker status needs to be checked. | warning |         
| :white_check_mark: | KubeletStatus                  | kubelet status needs to be checked. | warning |

## Add your own inspection rules
### Add custom OPA rules

1. Create a directory for storing OPA rules.

   ```shell
   mkdir opa
   ```
2. Add custom OPA rule files.

   > Note:
   - OPA rule for checking workloads: The package name must be *kubeeye_workloads_rego*.
   - OPA rule for checking RBAC settings: The package name must be *kubeeye_RBAC_rego*.
   - OPA rule for checking node settings: The package name must be *kubeeye_nodes_rego*.

3. To check whether the image registry address complies with rules, save the following rules to *imageRegistryRule.rego* 

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

4. Run KubeEye with custom rules.

  > Note: Kubeeye will read all files ending with *.rego* in the directory.

  ```shell
  root:# kubeeye audit -p ./opa
  NAMESPACE     NAME              KIND          MESSAGE
  default       nginx1            Deployment    [ImageRegistryNotmyregistry NotReadOnlyRootFilesystem NotRunAsNonRoot]
  default       nginx11           Deployment    [ImageRegistryNotmyregistry PrivilegeEscalationAllowed HighRiskCapabilities HostIPCAllowed HostPortAllowed ImagePullPolicyNotAlways ImageTagIsLatest InsecureCapabilities NoPriorityClassName PrivilegedAllowed NotReadOnlyRootFilesystem NotRunAsNonRoot]
  default       nginx111          Deployment    [ImageRegistryNotmyregistry NoCPULimits NoCPURequests ImageTagMiss NoLivenessProbe NoMemoryLimits NoMemoryRequests NoPriorityClassName NotReadOnlyRootFilesystem NoReadinessProbe NotRunAsNonRoot]
  ```

### Add custom NPD rules

1. Run the following command to change the ConfigMap:

   ```shell
   kubectl edit ConfigMap node-problem-detector-config -n kube-system 
   ```
2. Run the following command to restart NPD:

   ```shell
   kubectl rollout restart DaemonSet node-problem-detector -n kube-system
   ```

## KubeEye Operator
### What is KubeEye Operator

KubeEye Operator is an inspection platform for Kubernetes. It manages KubeEye to regularly inspect clusters and generate inspection results.

### How KubeEye Operator can help you

- It records inspection results by using CR and provide a web page for you to intuitively view and compare cluster inspection results.
- It provides more plugins.
- It provides more detailed modification suggestions.

### Deploy KubeEye Operator

```shell
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye.yaml
kubectl apply -f https://raw.githubusercontent.com/kubesphere/kubeeye/main/deploy/kubeeye_insights.yaml
```
### Obtain the inspection results

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
    <td align="center"><a href="https://github.com/panzhen6668"><img src="https://avatars.githubusercontent.com/u/55566964?v=4?s=100" width="100px;" alt=""/><br /><sub><b>panzhen6668</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=panzhen6668" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/zheng1"><img src="https://avatars.githubusercontent.com/u/4156721?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Zhengyi Lai</b></sub></a><br /><a href="https://github.com/kubesphere/kubeeye/commits?author=zheng1" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specifications. Contributions of any kind are welcome!

## Related Documents

* [RoadMap](docs/roadmap.md)
* [FAQ](docs/FAQ.md)
