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

## Features

1. Whether the core components of the cluster are healthy, including controller-manager, scheduler and etc.
2. Whether the cluster node healthy.
3. Whether the cluster pod is healthy.
> Check for more detail items [Click here](./docs/check-content_zh-CN.md)

## Results Example

```
root@node1:/home/ubuntu/go/src/kubeye# ./ke audit --kubeconfig /home/ubuntu/config
NODENAME   SEVERITY   HEARTBEATTIME               REASON              MESSAGE
node18     danger     2020-11-19T10:32:03+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node19     danger     2020-11-19T10:31:37+08:00   NodeStatusUnknown   Kubelet stopped posting node status.
node2      danger     2020-11-19T10:31:14+08:00   NodeStatusUnknown   Kubelet stopped posting node status.

NAME        SEVERITY   TIME                        MESSAGE
scheduler   danger     2020-11-27T17:09:59+08:00   Get http://127.0.0.1:10251/healthz: dial tcp 127.0.0.1:10251: connect: connection refused

NAMESPACE        NODENAME                                EVENTTIME                   REASON             MESSAGE
insights-agent   workloads-1606467120.164b519ca8c67416   2020-11-27T16:57:05+08:00   DeadlineExceeded   Job was active longer than specified deadline
kube-system      calico-node-zvl9t.164b3dc50580845d      2020-11-27T17:09:35+08:00   DNSConfigForming   Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system      kube-proxy-4bnn7.164b3dc4f4c4125d       2020-11-27T17:09:09+08:00   DNSConfigForming   Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
kube-system      nodelocaldns-2zbhh.164b3dc4f42d358b     2020-11-27T17:09:14+08:00   DNSConfigForming   Nameserver limits were exceeded, some nameservers have been omitted, the applied nameserver line is: 100.64.11.3 114.114.114.114 119.29.29.29
default          nginx-b8ffcf679-q4n9v.16491643e6b68cd7  2020-11-27T17:09:24+08:00   Failed             Error: ImagePullBackOff

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

--kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
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
