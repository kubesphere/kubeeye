package inspect

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strings"
)

type renderNode struct {
	Text     string
	Issues   *bool
	Children []renderNode
}

func HtmlOutput(clients *kube.KubernetesClient, outPath *string, taskName string, namespace string) error {

	results, _ := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{constant.LabelName: taskName})),
	})
	var resultCollection = make(map[string][]renderNode, 5)

	for _, item := range results.Items {
		if item.Spec.OpaResult.ResourceResults != nil {
			list := GetOpaList(item.Spec.OpaResult.ResourceResults)
			resultCollection[constant.Opa] = list
		}
		if item.Spec.PrometheusResult != nil {
			prometheus := getPrometheus(item.Spec.PrometheusResult)
			resultCollection[constant.Prometheus] = prometheus
		}
		if item.Spec.NodeInfoResult != nil {
			fileChange := getFileChange(item.Spec.NodeInfoResult)
			resultCollection[constant.FileChange] = fileChange
			sysctl := getSysctl(item.Spec.NodeInfoResult)
			resultCollection[constant.Sysctl] = sysctl
			systemd := getSystemd(item.Spec.NodeInfoResult)
			resultCollection[constant.Systemd] = systemd
		}
	}
	task, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectTasks(namespace).Get(context.TODO(), taskName, metav1.GetOptions{})
	year, month, day := task.CreationTimestamp.Date()
	var ruleNumber [][]interface{}
	if err == nil {
		for key, val := range task.Spec.InspectRuleTotal {
			var issues = len(resultCollection[key])
			if issues > 0 {
				issues -= 1
			}
			ruleNumber = append(ruleNumber, []interface{}{key, val, issues})
		}
	}
	data := map[string]interface{}{"title": fmt.Sprintf("%d-%d-%d", year, month, day), "overview": ruleNumber, "details": resultCollection}
	err = renderView(data)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func GetOpaList(result []v1alpha2.ResourceResult) (opaList []renderNode) {
	opaList = append(opaList, renderNode{Children: []renderNode{
		{Text: "NameSpace"}, {Text: "Kind"}, {Text: "Name"}, {Text: "Level"}, {Text: "Message"}, {Text: "Reason"},
	}})
	for _, resourceResult := range result {

		for _, item := range resourceResult.ResultItems {
			items := []renderNode{
				{Text: resourceResult.NameSpace},
				{Text: resourceResult.ResourceType},
				{Text: resourceResult.Name},
				{Text: item.Level},
				{Text: item.Message},
				{Text: item.Reason},
			}
			opaList = append(opaList, renderNode{Children: items})
		}
	}

	return opaList
}

func getPrometheus(pro [][]map[string]string) []renderNode {
	var prometheus []renderNode
	for _, p := range pro {
		header := renderNode{}
		for _, val := range p {
			if len(header.Children) == 0 {
				for k := range val {
					header.Children = append(header.Children, renderNode{Text: k})
				}
				prometheus = append(prometheus, header)
			}
			value := renderNode{}
			for i := range header.Children {
				value.Children = append(value.Children, renderNode{Text: val[header.Children[i].Text]})
			}
			prometheus = append(prometheus, value)
		}
	}
	return prometheus
}

func getFileChange(infoResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Children: []renderNode{{Text: "nodeName"}, {Text: "type"}, {Text: "name"}, {Text: "value"}}}
	villeinage = append(villeinage, header)
	for k, v := range infoResult {
		for _, item := range v.FileChangeResult {

			if item.Issues != nil && len(item.Issues) > 0 {
				val := renderNode{
					Children: []renderNode{
						{Text: k},
						{Text: constant.FileChange},
						{Text: item.FileName},
						{Text: strings.Join(item.Issues, ",")},
					},
				}
				villeinage = append(villeinage, val)
			}

		}

	}
	return villeinage
}

func getSysctl(infoResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Children: []renderNode{{Text: "nodeName"}, {Text: "type"}, {Text: "name"}, {Text: "value"}, {Text: "assert"}}}
	villeinage = append(villeinage, header)
	for k, v := range infoResult {
		for _, item := range v.SysctlResult {
			val := renderNode{
				Issues: item.Assert,
				Children: []renderNode{
					{Text: k},
					{Text: constant.Sysctl},
					{Text: item.Name},
					{Text: *item.Value},
					{Text: utils.FormatBool(item.Assert)},
				}}
			villeinage = append(villeinage, val)

		}

	}
	return villeinage
}

func getSystemd(infoResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Children: []renderNode{{Text: "nodeName"}, {Text: "type"}, {Text: "name"}, {Text: "value"}, {Text: "assert"}}}
	villeinage = append(villeinage, header)
	for k, v := range infoResult {
		for _, item := range v.SystemdResult {
			val := renderNode{
				Issues: item.Assert,
				Children: []renderNode{
					{Text: k},
					{Text: constant.Systemd},
					{Text: item.Name},
					{Text: *item.Value},
					{Text: utils.FormatBool(item.Assert)},
				}}
			villeinage = append(villeinage, val)
		}

	}
	return villeinage
}

func renderView(data map[string]interface{}) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := path.Join(dir, "pkg", "template", "result.html")
	files, err := template.ParseFiles(filePath)
	if err != nil {
		return err
	}
	create, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer create.Close()
	err = files.Execute(create, data)
	if err != nil {
		return err
	}
	return nil
}
