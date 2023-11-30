<div align=center><img src="docs/images/KubeEye-O.svg?raw=true"></div>

<p align=center>
<a href="https://github.com/kubesphere/kubeeye/actions?query=event%3Apush+branch%3Amain+workflow%3ACI+"><img src="https://github.com/kubesphere/kubeeye/workflows/CI/badge.svg?branch=main&event=push"></a>
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<a href="https://github.com/kubesphere/kubeeye#contributors-"><img src="https://img.shields.io/badge/all_contributors-10-orange.svg?style=flat-square"></a>
<!-- ALL-CONTRIBUTORS-BADGE:END -->
</p>

> English | [中文](README_zh.md)

KubeEye 是为 Kubernetes 设计的云原生集群巡检工具，用于根据自定义规则发现 Kubernetes 集群中存在的问题及风险。

## 快速开始

### 安装
可从 [Releases](https://github.com/kubesphere/kubeeye/releases) 中下载安装包（其中包含helm chart，demo rules 以及 供离线环境部署的镜像）

```shell
VERSION=v1.0.0

wget https://github.com/kubesphere/kubeeye/releases/download/${VERSION}/kubeeye-offline-${VERSION}.tar.gz

tar -zxvf kubeeye-offline-${VERSION}.tar.gz

cd kubeeye-offline-${VERSION}

# 如需离线安装，请自行将images中的镜像导入本地仓库，并修改 chart/kubeeye/values.yaml 中的镜像地址。

helm upgrade --install kubeeye chart/kubeeye -n kubeeye-system --create-namespace

```

### 使用

#### 导入规则
   
> 安装包中的rule目录下提供了demo规则，可根据需求自定义规则。

> 注意 prometheus 规则需提前为规则设置prometheus的endpoint。

```shell
kubectl apply -f rule
```

#### 创建巡检计划

    按需配置巡检计划
```shell
cat > plan.yaml << EOF
apiVersion: kubeeye.kubesphere.io/v1alpha2
kind: InspectPlan
metadata:
  name: inspectplan
spec:
  # 需要执行检查的计划时间，仅支持cron表达式，例："*/30 * * * ?"表示每30分钟执行一次巡检。
  # 如果仅需单次巡检，则将该参数移除。
  schedule: "*/30 * * * ?"
  # 巡检结果最大保留数量，不填写则是保留全部
  maxTasks: 10 
  # 是否暂停巡检计划, 仅作用于周期巡检，true 或 flase （默认false）
  suspend: false
  # 巡检超时时间, 默认 10m
  timeout: 10m
  # 巡检规则列表，用于关联对应的巡检规则，填写 inspectRule 名称
  # 可通过 kubectl get inspectrule 查看集群中巡检规则
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
  # 多集群巡检（目前仅支持 KubeSphere 多集群巡检）
  # clusterName: 
  # - name: host
EOF


kubectl apply -f plan.yaml
```

#### 巡检报告获取
##### 查询巡检结果
```shell
# 查看巡检结果名称，用于后续巡检报告下载
kubectl get inspectresult
```
##### 获取巡检报告
###### 命令行方式下载
```shell
## 获取 kubeeye-apiserver svc地址和端口
kubectl get svc -n kubeeye-system kubeeye-apiserver -o custom-columns=CLUSTER-IP:.spec.clusterIP,PORT:.spec.ports[*].port

## 下载巡检报告, 注意替换 <> 为环境中查询到的实际信息
curl http://<svc-ip>:9090/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/<result name>\?type\=html -o inspectReport.html

## 下载后可使用浏览器打开html文件查看
```
###### 浏览器查看
```shell
## 为 kubeeye-apiserver 创建 nodePort 类型svc
kubectl -n kubeeye-system expose deploy kubeeye-apiserver --port=9090 --type=NodePort --name=ke-apiserver-node-port

## 浏览器输入巡检报告url查看, 注意替换 <> 为环境中查询到的实际信息
http://<node address>:<node port>/kapis/kubeeye.kubesphere.io/v1alpha2/inspectresults/<result name>?type=html
```
## 支持规则清单
* OPA 规则
* PromQL 规则
* 文件变更规则
* 内核参数配置规则
* Systemd 服务状态规则
* 节点基本信息规则
* 文件内容检查规则
* 服务连通性检查规则