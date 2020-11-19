# kubeye

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
```

3、Install Node-problem-Detector in the inspection cluster

> Note: The NPD module does not need to be installed When more detailed node information does not need to be probed.

* Create a ConfigMap file for Node-Problem-Detector, which contains fault patrol rules and can be added by the user  [npd-config.yaml](./docs/npd-config.yaml).  
`kubectl apply -f npd-config.yaml`

* Create the DaemonSet file for Node-Problem-Detector  [npd.yaml](./docs/npd.yaml).  
`kubectl apply -f npd.yaml`

## Results

1. Basic information of cluster, including Kubernetes version, number of nodes, pod number, etc.
2. Kubernetes Best Practices configuration.
3. Nodes information when running.
4. Runtime cluster failure and other information.

## Example display
```
root@node1:/home/ubuntu/go/src/kubeye# ./ke audit --kubeconfig /home/ubuntu/config
allNodeStatusResults:
- heartbeatTime: "2020-11-10T11:00:19+08:00"
  message: kubelet is posting ready status
  name: node1
  reason: KubeletReady
  status: "True"
- heartbeatTime: "2020-10-21T17:34:49+08:00"
  message: Kubelet stopped posting node status.
  name: node2
  reason: NodeStatusUnknown
  status: Unknown
- heartbeatTime: "2020-10-21T17:35:21+08:00"
  message: Kubelet stopped posting node status.
  name: node3
  reason: NodeStatusUnknown
  status: Unknown
basicClusterInformation:
  k8sVersion: "1.16"
  namespaceNum: 6
  nodeNum: 3
  podNum: 28
basicComponentStatus:
  controller-manager: ok
  etcd-0: '{"health":"true"}'
  scheduler: ok
clusterCheckResults:
- eventTime: "2020-11-10T10:57:23+08:00"
  message: 'Error: ImagePullBackOff'
  name: nginx-6c74496488-s45tg.163ff88f7263ccc7
  namespace: test
  reason: Failed
clusterConfigurationResults:
- containerResults:
  - results:
      cpuLimitsMissing:
        category: Resources
        id: cpuLimitsMissing
        message: CPU limits should be set
        severity: warning
  createdTime: "2020-11-10T11:00:21+08:00"
  kind: Deployment
  name: coredns
  namespace: kube-system
```