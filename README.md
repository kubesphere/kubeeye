# kubeye

Kubeye is a tool for inspecting Kubernetes clusters. It runs a variety of checks to ensure that Kubernetes pods are configured using best practices, helping you avoid problems in the future. 
Quickly get cluster core component status and cluster size information and abnormal Pods information and tons of node problems. Developed by the GO language. Support for user-defined best practice configuration rules and the addition of cluster fault scouts, which can refer to the [Node-Problem-Detector](https://github.com/kubernetes/node-problem-detector) project。

## Usage

1、Install Node-problem-Detector in the inspection cluster
* Create a ConfigMap file for Node-Problem-Detector, which contains fault patrol rules and can be added by the user[npd-config.yaml](./docs/npd-config.yaml).  
`kubectl apply -f npd-config.yaml`

* Create the DaemonSet file for Node-Problem-Detector[npd.yaml](./docs/npd.yaml).  
`kubectl apply -f npd.yaml`

2、Get the Installer Excutable File
```shell script
wget https://installertest.pek3b.qingstor.com/ke
chmod +x ke
```

3、Perform operation
```shell script
./ke audit --kubeconfig ***

--kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
```

## Results

1. Basic information of cluster, including Kubernetes version, number of nodes, pod number, etc.
2. Kubernetes Best Practices configuration.
3. Nodes information when running.
4. Runtime cluster failure and other information.

## Example display
```
root@node1:/home/ubuntu/go/src/kubeye# ./ke audit --kubeconfig /home/ubuntu/config
AuditAddress: https://192.168.0.3:6443
AuditTime: "2020-11-06T17:39:15+08:00"
ClusterInfo:
  K8sVersion: "1.16"
  NamespaceNum: 6
  NodeNum: 3
  PodNum: 28
ComponentStatus:
  controller-manager: ok
  etcd-0: '{"health":"true"}'
  scheduler: ok
GoodPractice:
- ContainerResults:
  - Results:
      cpuLimitsMissing:
        Category: Resources
        ID: cpuLimitsMissing
        Message: CPU limits should be set
        Severity: warning
        Success: false
  CreatedTime: "2020-11-06T17:39:15+08:00"
  Kind: Deployment
  Name: openebs-ndm-operator
  Namespace: openebs
NodeStatus:
- HeartbeatTime: "2020-11-06T17:38:16+08:00"
  Message: kubelet is posting ready status
  Name: node1
  Reason: KubeletReady
  Status: "True"
- HeartbeatTime: "2020-10-21T17:34:49+08:00"
  Message: Kubelet stopped posting node status.
  Name: node2
  Reason: NodeStatusUnknown
  Status: Unknown
ProblemDetector:
- EventTime: "2020-11-06T17:37:08+08:00"
  Message: 'Error: ImagePullBackOff'
  Name: nginx-6c74496488-s45tg.163ff88f7263ccc7
  Namespace: test
  Reason: Failed
```