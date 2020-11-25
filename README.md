# Kubeye

Kubeye is a tool for inspecting Kubernetes clusters. It runs a variety of checks to ensure that Kubernetes pods are configured using best practices, helping you avoid problems in the future. 
Quickly get cluster core component status and cluster size information and abnormal Pods information and tons of node problems. Developed by the GO language. Support for user-defined best practice configuration rules and the addition of cluster fault scouts, which can refer to the [Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector) project。

## Usage

1、Get the Installer Excutable File
* Binary downloads of the kubeye.
```shell script
wget https://installertest.pek3b.qingstor.com/ke
chmod +x ke
```
* Build Binary from Source Code
```shell script
git clone https://github.com/kubesphere/kubeye.git
cd kubeye 
make
```
2、Perform operation
```shell script
./ke audit --kubeconfig ***

--kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
> Note: If it is an external cluster, the server needs an external network address in the config file.
```

3、Install Node-problem-Detector in the inspection cluster

> Note: The NPD module does not need to be installed When more detailed node information does not need to be probed.

```shell script
./ke add npd --kubeconfig ***

--kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
> Note: If it is an external cluster, the server needs an external network address in the config file.
```

* Continue with step 2.

## Results

1. Whether the core components of the cluster are healthy, including controller-manager, scheduler and etc.
2. Whether the cluster node healthy.
3. Whether the cluster pod is healthy.
> Check for more detail items [Click here](./docs/check-content_zh-CN.md)

## Results Example

```
root@node1:/home/ubuntu/go/src/kubeye# ./ke audit --kubeconfig /home/ubuntu/config
HEARTBEATTIME                   SEVERITY                                 NODENAME   REASON              MESSAGE
2020-11-19 10:32:03 +0800 CST   danger                                   node18     NodeStatusUnknown   Kubelet stopped posting node status.
2020-11-19 10:31:37 +0800 CST   danger                                   node19     NodeStatusUnknown   Kubelet stopped posting node status.
2020-11-19 10:31:14 +0800 CST   danger                                   node2      NodeStatusUnknown   Kubelet stopped posting node status.
2020-11-19 10:31:58 +0800 CST   danger                                   node3      NodeStatusUnknown   Kubelet stopped posting node status.

NAME                            SEVERITY                                 MESSAGE
scheduler                       danger                                   Get http://127.0.0.1:10251/healthz: dial tcp 127.0.0.1:10251: connect: connection refused

EVENTTIME                       NODENAME                                 NAMESPACE     REASON       MESSAGE
2020-11-20 18:52:13 +0800 CST   nginx-b8ffcf679-q4n9v.16491643e6b68cd7   default       Failed       Error: ImagePullBackOff

TIME                            NAME                                     NAMESPACE     KIND         MESSAGE
2020-11-20T18:54:44+08:00       calico-node                              kube-system   DaemonSet    [{map[cpuLimitsMissing:{cpuLimitsMissing CPU limits should be set false    warning  Resources} runningAsPrivileged:{runningAsPrivileged Should not be running as privileged false    warning  Security}]}]
2020-11-20T18:54:44+08:00       kube-proxy                               kube-system   DaemonSet    [{map[runningAsPrivileged:{runningAsPrivileged Should not be running as privileged false    warning  Security}]}]
2020-11-20T18:54:44+08:00       coredns                                  kube-system   Deployment   [{map[cpuLimitsMissing:{cpuLimitsMissing CPU limits should be set false    warning  Resources}]}]
2020-11-20T18:54:44+08:00       nodelocaldns                             kube-system   DaemonSet    [{map[cpuLimitsMissing:{cpuLimitsMissing CPU limits should be set false    warning  Resources} hostPortSet:{hostPortSet Host port should not be configured false    warning  Networking} runningAsPrivileged:{runningAsPrivileged Should not be running as privileged false    warning  Security}]}]
2020-11-20T18:54:44+08:00       nginx                                    default       Deployment   [{map[cpuLimitsMissing:{cpuLimitsMissing CPU limits should be set false    warning  Resources} livenessProbeMissing:{livenessProbeMissing Liveness probe should be configured false    warning  Health Checks} tagNotSpecified:{tagNotSpecified Image tag should be specified false    danger   Images  }]}]
2020-11-20T18:54:44+08:00       calico-kube-controllers                  kube-system   Deployment   [{map[cpuLimitsMissing:{cpuLimitsMissing CPU limits should be set false    warning  Resources} livenessProbeMissing:{livenessProbeMissing Liveness probe should be configured false    warning  Health Checks}]}
```

## Custom check
