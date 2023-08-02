package output

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye-v1alpha2/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye-v1alpha2/constant"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/kube"
	"github.com/kubesphere/kubeeye-v1alpha2/pkg/template"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

type renderNode struct {
	Text     string
	Issues   bool
	Header   bool
	Children []renderNode
}

func HtmlOut(ctx context.Context, Clients *kube.KubernetesClient, Path string, TaskName string) error {

	results, err := Clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().Get(ctx, TaskName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("result not exist")
	}
	var resultCollection = make(map[string][]renderNode, 5)

	if results.Spec.OpaResult.ResourceResults != nil {
		list := getOpaList(results.Spec.OpaResult.ResourceResults)
		resultCollection[constant.Opa] = list
	}
	if results.Spec.PrometheusResult != nil {
		prometheus := getPrometheus(results.Spec.PrometheusResult)
		resultCollection[constant.Prometheus] = prometheus
	}
	if results.Spec.NodeInfoResult != nil {
		fileChange := getFileChange(results.Spec.NodeInfoResult)
		resultCollection[constant.FileChange] = fileChange
		sysctl := getSysctl(results.Spec.NodeInfoResult)
		resultCollection[constant.Sysctl] = sysctl
		systemd := getSystemd(results.Spec.NodeInfoResult)
		resultCollection[constant.Systemd] = systemd
		filter := getFileFilter(results.Spec.NodeInfoResult)
		resultCollection[constant.FileFilter] = filter
	}

	if results.Spec.ComponentResult != nil {
		component := getComponent(results.Spec.ComponentResult)
		resultCollection[constant.Component] = component
	}

	var ruleNumber [][]interface{}
	for key, val := range results.Spec.InspectRuleTotal {
		var issues = len(resultCollection[key])
		if issues > 0 {
			issues -= 1
		}
		ruleNumber = append(ruleNumber, []interface{}{key, val, issues})
	}

	data := map[string]interface{}{"title": results.CreationTimestamp.Format("2006-01-02 15:04"), "overview": ruleNumber, "details": resultCollection}
	err = renderView(data, Path)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func getOpaList(result []v1alpha2.ResourceResult) (opaList []renderNode) {
	opaList = append(opaList, renderNode{Header: true, Children: []renderNode{
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
		header := renderNode{Header: true}
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
	header := renderNode{Header: true,
		Children: []renderNode{
			{Text: "nodeName"},
			{Text: "type"},
			{Text: "name"},
			{Text: "value"}}}
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

func getFileFilter(fileResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Header: true, Children: []renderNode{
		{Text: "nodeName"},
		{Text: "FileName"},
		{Text: "Path"},
		{Text: "Issues"}},
	}
	villeinage = append(villeinage, header)

	for k, v := range fileResult {
		for _, result := range v.FileFilterResult {
			for _, issue := range result.Issues {
				content2 := []renderNode{{Text: k}, {Text: result.FileName}, {Text: result.Path}, {Text: issue}}
				villeinage = append(villeinage, renderNode{Children: content2})
			}

		}

	}

	return villeinage
}
func getComponent(component []v1alpha2.ComponentResultItem) []renderNode {
	var villeinage []renderNode
	header := renderNode{Header: true, Children: []renderNode{
		{Text: "name"},
		{Text: "namespace"},
		{Text: "endpoint"}},
	}
	villeinage = append(villeinage, header)

	for _, c := range component {
		value := []renderNode{{Text: c.Name}, {Text: c.Namespace}, {Text: c.Endpoint}}
		villeinage = append(villeinage, renderNode{Children: value})
	}

	return villeinage
}

func getSysctl(infoResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Header: true,
		Children: []renderNode{
			{Text: "nodeName"},
			{Text: "type"}, {Text: "name"},
			{Text: "value"},
		}}
	villeinage = append(villeinage, header)
	for k, v := range infoResult {

		for _, item := range v.SysctlResult {
			if item.Assert != nil && *item.Assert == false {
				val := renderNode{
					Issues: assertBoolBackBool(item.Assert),
					Children: []renderNode{
						{Text: k},
						{Text: constant.Sysctl},
						{Text: item.Name},
						{Text: *item.Value},
					}}
				villeinage = append(villeinage, val)
			}

		}

	}
	return villeinage
}

func getSystemd(infoResult map[string]v1alpha2.NodeInfoResult) []renderNode {
	var villeinage []renderNode
	header := renderNode{Header: true,
		Children: []renderNode{
			{Text: "nodeName"},
			{Text: "type"},
			{Text: "name"},
			{Text: "value"},
		},
	}
	villeinage = append(villeinage, header)
	for k, v := range infoResult {
		for _, item := range v.SystemdResult {
			if item.Assert != nil && *item.Assert == false {
				val := renderNode{
					Issues: assertBoolBackBool(item.Assert),
					Children: []renderNode{
						{Text: k},
						{Text: constant.Systemd},
						{Text: item.Name},
						{Text: *item.Value},
					}}
				villeinage = append(villeinage, val)
			}
		}

	}
	return villeinage
}

func renderView(data map[string]interface{}, p string) error {

	htmlTemplate, err := template.GetInspectResultHtmlTemplate()
	if err != nil {
		return err
	}
	name := ParseFileName(p, fmt.Sprintf("inspectionReport(%s).html", time.Now().Format("2006-01-02 15:04")))
	create, err := os.Create(name)
	if err != nil {
		return err
	}
	defer create.Close()
	err = htmlTemplate.Execute(create, data)
	if err != nil {
		return err
	}
	return nil
}

func assertBoolBackBool(b *bool) bool {
	if b == nil {
		return false
	}
	if *b {
		return false
	}
	return true
}
