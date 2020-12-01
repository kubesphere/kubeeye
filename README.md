# Kubeye

Kubeye is a tool for inspecting Kubernetes clusters. It runs a variety of checks to ensure that Kubernetes pods are configured using best practices, helping you avoid problems in the future. 
Quickly get cluster core component status and cluster size information and abnormal Pods information and tons of node problems. Developed by the GO language. Support for user-defined best practice configuration rules and the addition of cluster fault scouts, which can refer to the [Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector) project。

## Usage

1、Get the Installer Excutable File
* Binary downloads of the kubeye can be found on the [Releases page](https://github.com/kubesphere/kubeye/releases). Unpack the binary and you are good to go!

* Build Binary from Source Code
```shell script
git clone https://github.com/kubesphere/kubeye.git
cd kubeye 
make ke-linux
```
2、Perform operation
```shell script
./ke audit --kubeconfig /home/ubuntu/.kube/config
```

3、(Optional) Install Node-problem-Detector in the inspection cluster

> Note: The NPD module does not need to be installed When more detailed node information does not need to be probed.

```shell script
./ke install npd --kubeconfig /home/ubuntu/.kube/config
```

## What kubeye can do

1. Core component detection in the cluster, including controller-manager, scheduler and ETCD exception detection.
2. Node detection in the cluster, including Kubelet abnormalities, insufficient machine MEMORY/CPU/IO resources, docker service exceptions.
3. Pod detection int the cluster, including pod best practices, pod exceptions information.

## Features
| YES/NO |          CHECK ITEM             | YES/NO |            CHECK ITEM           | 
| ------ | --------------------------------| ------ | --------------------------------|
| :white_check_mark: | ETCDHealthStatus                | :white_check_mark: | Controller-ManagerHealthStatus  | 
| :white_check_mark: | ScheduleHealthStatus            | :white_check_mark: | TheNodeMemoryIsFull             |
| :white_check_mark: | DockerHealthStatus              | :white_check_mark: | NodeDiskIsFull                  | 
| :white_check_mark: | KubeletHealthStatus             | :white_check_mark: | NodeCPUIsFull                   | 
| :white_check_mark: | NodeCorruptOverlay2             | :white_check_mark: | NodeKernelNULLPointer           |
| :white_check_mark: | NodeDeadlock                    | :white_check_mark: | NodeOOM                         | 
| :white_check_mark: | NodeExt4Error                   | :white_check_mark: | NodeTaskHung                    | 
| :white_check_mark: | NodeUnregisterNetDevice         | :white_check_mark: | NodeCorruptDockerImage          |
| :white_check_mark: | NodeAUFSUmountHung              | :white_check_mark: | NodeDockerHung                  | 
| :white_check_mark: | PodSetLiveNessProbe             | :white_check_mark: | PodSetTagNotSpecified           | 
| :white_check_mark: | PodSetRunAsPrivileged           | :white_check_mark: | PodSetImagePullBackOff          |           
| :white_check_mark: | PodSetImageRegistry             | :white_check_mark: | PodSetCpuLimitsMissing          |             
| :white_check_mark: | PodNoSuchFileOrDirectory        | :white_check_mark: | PodIOError                      | 
| :white_check_mark: | PodNoSuchDeviceOrAddress        | :white_check_mark: | PodInvalidArgument              |               
| :white_check_mark: | PodDeviceOrResourceBusy         | :white_check_mark: | PodFileExists                   |              
| :white_check_mark: | PodTooManyOpenFiles             | :white_check_mark: | PodNoSpaceLeftOnDevice          |
|                    | NodeTokenExpired                |                    | NodeApiServerExpired            |
|                    | NodeKubeletExpired              |                    | PodSetCpuRequestsMissing        | 
|                    | PodSetHostIPCSet                |                    | PodSetHostNetworkSet            | 
|                    | PodHostPIDSet                   |                    | PodMemoryRequestsMiss           | 
|                    | PodSetHostPort                  |                    | PodSetMemoryLimitsMissing       |
|                    | PodNotReadOnlyRootFiles         |                    | PodSetPullPolicyNotAlways       | 
|                    | PodSetRunAsRootAllowed          |                    | PodDangerousCapabilities        |

## Results Example

```
root@node1:/home/ubuntu/go/src/kubeye# ./ke audit --kubeconfig /home/ubuntu/config
NODENAME   SEVERITY   HEARTBEATTIME               REASON              MESSAGE
node18     danger     2020-11-19T10:32:03+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node19     danger     2020-11-19T10:31:37+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node2      danger     2020-11-19T10:31:14+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node3      danger     2020-11-27T17:36:53+08:00   KubeletNotReady     Container runtime not ready: RuntimeReady=false reason:DockerDaemonNotReady message:docker: failed to get docker version: Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?

NAME        SEVERITY   TIME                        MESSAGE
scheduler   danger     2020-11-27T17:09:59+08:00   Get http://127.0.0.1:10251/healthz: dial tcp 127.0.0.1:10251: connect: connection refused
etcd-0      danger     2020-11-27T17:56:37+08:00   Get https://192.168.13.8:2379/health: dial tcp 192.168.13.8:2379: connect: connection refused

NAMESPACE        NODENAME                                EVENTTIME                   REASON                MESSAGE
default          node3.164b53d23ea79fc7                  2020-11-27T17:37:34+08:00   ContainerGCFailed     rpc error: code = Unknown desc = Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
default          node3.164b553ca5740aae                  2020-11-27T18:03:31+08:00   FreeDiskSpaceFailed   failed to garbage collect required amount of images. Wanted to free 5399374233 bytes, but freed 416077545 bytes
default          nginx-b8ffcf679-q4n9v.16491643e6b68cd7  2020-11-27T17:09:24+08:00   Failed                Error: ImagePullBackOff
default          node3.164b5861e041a60e                  2020-11-27T19:01:09+08:00   SystemOOM             System OOM encountered, victim process: stress, pid: 16713
default          node3.164b58660f8d4590                  2020-11-27T19:01:27+08:00   OOMKilling            Out of memory: Kill process 16711 (stress) score 205 or sacrifice child Killed process 16711 (stress), UID 0, total-vm:826516kB, anon-rss:819296kB, file-rss:0kB, shmem-rss:0kB
insights-agent   workloads-1606467120.164b519ca8c67416   2020-11-27T16:57:05+08:00   DeadlineExceeded      Job was active longer than specified deadline
kube-system      calico-node-zvl9t.164b3dc50580845d      2020-11-27T17:09:35+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system      kube-proxy-4bnn7.164b3dc4f4c4125d       2020-11-27T17:09:09+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system      nodelocaldns-2zbhh.164b3dc4f42d358b     2020-11-27T17:09:14+08:00   DNSConfigForming      Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29


NAMESPACE        NAME                      KIND         TIME                        MESSAGE
kube-system      node-problem-detector     DaemonSet    2020-11-27T17:09:59+08:00   [livenessProbeMissing runAsPrivileged]
kube-system      calico-node               DaemonSet    2020-11-27T17:09:59+08:00   [runAsPrivileged cpuLimitsMissing]
kube-system      nodelocaldns              DaemonSet    2020-11-27T17:09:59+08:00   [cpuLimitsMissing runAsPrivileged]
default          nginx                     Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing livenessProbeMissing tagNotSpecified]
insights-agent   workloads                 CronJob      2020-11-27T17:09:59+08:00   [livenessProbeMissing]
insights-agent   cronjob-executor          Job          2020-11-27T17:09:59+08:00   [livenessProbeMissing]
kube-system      calico-kube-controllers   Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing livenessProbeMissing]
kube-system      coredns                   Deployment   2020-11-27T17:09:59+08:00   [cpuLimitsMissing]   
```

## Custom check

* Add custom npd rule methods
```
1. Deploy npd, ./ke add npd --kubeconfig /home/ubuntu/.kube/config
2. Ddit node-problem-detector-config configMap, such as: kubectl edit cm -n kube-system node-problem-detector-config
3. Add exception log information under the rule of configMap, rules follow regular expressions.
```
* Add custom best practice configuration
```
1. Use the -f parameter and file name config.yaml.
./ke audit -f /home/ubuntu/go/src/kubeye/examples/tmp/config.yaml --kubeconfig /home/ubuntu/.kube/config
2. config.yaml example, follow the JSON syntax.
ubuntu@node1:~/go/src/kubeye/examples/tmp$ cat config.yaml
checks:
  imageRegistry: warning

customChecks:
  imageRegistry:
    successMessage: Image comes from allowed registries
    failureMessage: Image should not be from disallowed registry
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

ubuntu@node1:~/go/src/kubeye/examples/tmp$./ke audit -f /home/ubuntu/go/src/kubeye/examples/tmp/config.yaml
NAMESPACE     NAME                      KIND         TIME                        MESSAGE
default       nginx                     Deployment   2020-11-27T17:18:31+08:00   [imageRegistry]
kube-system   node-problem-detector     DaemonSet    2020-11-27T17:18:31+08:00   [livenessProbeMissing runAsPrivileged]
kube-system   calico-node               DaemonSet    2020-11-27T17:18:31+08:00   [cpuLimitsMissing runAsPrivileged]
kube-system   calico-kube-controllers   Deployment   2020-11-27T17:18:31+08:00   [cpuLimitsMissing livenessProbeMissing]
kube-system   nodelocaldns              DaemonSet    2020-11-27T17:18:31+08:00   [runAsPrivileged cpuLimitsMissing]
default       nginx                     Deployment   2020-11-27T17:18:31+08:00   [livenessProbeMissing cpuLimitsMissing]
kube-system   coredns                   Deployment   2020-11-27T17:18:31+08:00   [cpuLimitsMissing]
```
