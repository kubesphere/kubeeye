<div align=center><img src="docs/images/KubeEye-O.svg?raw=true"></div>

<p align=center>
<a href="https://github.com/kubesphere/kubeeye/actions?query=event%3Apush+branch%3Amain+workflow%3ACI+"><img src="https://github.com/kubesphere/kubeeye/workflows/CI/badge.svg?branch=main&event=push"></a>
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="https://github.com/kubesphere/kubeeye#contributors-"><img src="https://img.shields.io/badge/all_contributors-10-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

> English | [中文](README_zh.md)

KubeEye is a cloud-native cluster inspection tool specifically designed for Kubernetes, capable of identifying issues and risks within the Kubernetes cluster based on custom rules.

## QuickStart

### Installation
Download the installation package from [Releases](https://github.com/kubesphere/kubeeye/releases), which includes Helm chart, demo rules, and images for offline installation.

```shell
VERSION=v1.0.0

wget https://github.com/kubesphere/kubeeye/releases/download/${VERSION}/kubeeye-offline-${VERSION}.tar.gz

tar -zxvf kubeeye-offline-${VERSION}.tar.gz

cd kubeeye-offline-${VERSION}

# offline installation, please import the images in the 'images' folder into the local container repository yourself and modify the images repo in `chart/kubeeye/values.yaml`.

helm upgrade --install kubeeye chart/kubeeye -n kubeeye-system --create-namespace

```

### Usage

#### Import Inspect Rules
   
> The `rule` directory in the installation package provides demo rules, which can be customized according to specific needs.

> Notice： Prometheus rules need to have the endpoint of Prometheus set in advance.

```shell
kubectl apply -f rule
```

#### Create Inspect Plan

Configure inspection plans on demand.
```shell
cat > plan.yaml << EOF
apiVersion: kubeeye.kubesphere.io/v1alpha2
kind: InspectPlan
metadata:
  name: inspectplan
spec:
  # The planned time for executing inspections only supports cron expressions. For example, '*/30 * * * ?' means that the inspection will be performed every 30 minutes.'
  # If only a single inspection is required, then remove this parameter.
  schedule: "*/30 * * * ?"
  # The maximum number of retained inspection results, if not filled in, will retain all.
  maxTasks: 10 
  # Should the inspection plan be paused, applicable only to periodic inspections, true or false (default is false).
  suspend: false
  # Inspection timeout, default 10 minutes.
  timeout: 10m
  # Inspection rule list, used to associate corresponding inspection rules, please fill in the inspectRule name.
  # Execute `kubectl get inspectrule` to view the inspection rules in the cluster.
  ruleNames:
    - name: inspect-rule-filter-file
    - name: inspect-rule-node-info
    - name: inspect-rule-node
    - name: inspect-rule-sbnormalpodstatus 
    - name: inspect-rule-deployment
    - name: inspect-rule-sysctl
    - name: inspect-rule-prometheus
    - name: inspect-rule-filechange
    - name: inspect-rule-systemd
  # nodeName: master
  # nodeSelector:
  #   node-role.kubernetes.io/master: ""        
  # Multi-cluster inspection (currently only supports multi-cluster inspection in KubeSphere)
  # clusterName: 
  # - name: host
EOF


kubectl apply -f plan.yaml
```

#### Obtaining Inspection Reports
##### Check Inspection Results
```shell
# View the name of the inspection result for inspection report download.
kubectl get inspectresult
```
###### Command
```shell
## Get the address and port of kubeeye-apiserver service.
kubectl get svc -n kubeeye-system kubeeye-apiserver -o custom-columns=CLUSTER-IP:.spec.clusterIP,PORT:.spec.ports[*].port

## Download the inspection report, and please replace <> with the actual information obtained from the environment.
curl http://<svc-ip>:9090/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/<result name>\?type\=html -o inspectReport.html

## After downloading, you can use a browser to open the HTML file for viewing.
```
###### Web Console
```shell
## Create a nodePort type svc for kubeeye-apiserver.
kubectl -n kubeeye-system expose deploy kubeeye-apiserver --port=9090 --type=NodePort --name=ke-apiserver-node-port

## Enter the inspection report URL in the browser to view, and remember to replace <> with the actual information obtained from the environment.
http://<node address>:<node port>/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/<result name>?type=html
```
## Supported Rules List
* OPA 
* PromQL 
* File Change
* Kernel Parameter Configuration
* Systemd Service Status
* Node Basic Info
* File Content Inspection
* Service Connectivity