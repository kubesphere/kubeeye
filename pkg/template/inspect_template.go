package template

import texttemplate "text/template"
import hemltemplate "text/template"

func GetInspectRuleTemplate() (*texttemplate.Template, error) {

	return texttemplate.New("examples-inspect-rule").Parse(`
apiVersion: kubeeye.kubesphere.io/v1alpha2
kind: InspectRule
metadata:
  labels:
    app.kubernetes.io/name: inspectrule
    app.kubernetes.io/instance: inspectrule-sample
    app.kubernetes.io/part-of: kubeeye
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: kubeeye
    kubeeye.kubesphere.io/rule-tag: kubeeye_workloads_rego
  name: inspect-rules-sample
  namespace: kubeeye-system
spec:
  sysctl:
    - name: vm.max_map_count
      rule: vm.max_map_count==262144      
      nodeName: master
      nodeSelector:
		kubernetes.io/name: ""
  systemd:
    - name: docker
      rule: docker = "active"
      nodeName: master
      nodeSelector:
		kubernetes.io/name: ""
  fileChange:
    - name: 
      path: 
      nodeName: master
      nodeSelector:
		kubernetes.io/name: ""
  prometheusEndpoint: 
  prometheus:
    - name: 
      rule: 
  opas:
    - module: kubeeye_workloads_rego
      name: 
      rule: |-
`)

}

func GetInspectPlanTemplate() (*texttemplate.Template, error) {
	return texttemplate.New("examples-inspect-plan").Parse(`
	apiVersion: kubeeye.kubesphere.io/v1alpha2
	kind: InspectPlan
	metadata:
	  name: inspect-plan-sample
	  namespace: kubeeye-system
	spec:
	  schedule: "*/30 * * *  ?"
	  maxTasks: 10
	  tag: kubeeye_workloads_rego
	`)
}

func GetInspectResultHtmlTemplate() (*hemltemplate.Template, error) {
	return hemltemplate.New("result").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
</head>
<body>

<div class="header">巡检报告（{{- .title -}}）</div>

<div class="content">
    <div style="font-size: 30px;min-width: 800px;">总览</div>
    <table border="1" cellpadding="0" cellspacing="0" class="overview">
        <thead>
        <tr>
            <td>规则名称</td>
            <td>检查规则数量</td>
            <td>发现问题数量</td>
        </tr>
        </thead>
        <tbody>

        {{range .overview}}
        <tr>{{range $i,$v:= .}}
            {{if eq $i 0}}
            <td><a href="#{{.}}">{{.}}</a></td>
            {{else}}
            <td> {{.}}</td>
            {{end}}
            {{end}}
        </tr>
        {{end}}
        </tbody>
    </table>
</div>

{{range $k,$v:= .details}}

<div class="content">
    <div style="font-size: 30px;width: 100%"><a id="{{$k}}">{{$k}}</a></div>
    <table border="1" cellpadding="0" cellspacing="0" class="overview">
        {{range $i,$v1:=$v }}

        {{if $v1.Issues}}
        <tr class="issues">
            {{else}}
        <tr>
            {{end}}
            {{range $v1.Children}}

            {{if $v1.Header}}
            <th>{{.Text}}</th>
            {{else}}
            <td>{{.Text}}</td>
            {{end}}
            {{end}}
        </tr>
        {{end}}
    </table>
</div>

{{end}}

</body>

<style>
    .header {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 50px;
        width: 100%;
        box-sizing: border-box;
        font-size: 50px;
        font-weight: bold;
    }

    .content {
        width: 100%;
        display: flex;
        flex-direction: column;
        font-size: 18px;
        margin-top: 30px;

    }

    .overview {
        min-width: 800px;
        text-align: center;
    }

    .issues {
        background-color: #ffc3481a;
    }
</style>

</html>
`)
}