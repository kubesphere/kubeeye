package output

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/clients/clientset/versioned/scheme"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/template"
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

func HtmlOut(ctx context.Context, Clients *kube.KubernetesClient, Path string, TaskName string, TaskNameSpace string) error {

	listOptions := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{constant.LabelName: TaskName})),
	}
	var results v1alpha2.InspectResultList
	err := Clients.VersionClientSet.KubeeyeV1alpha2().RESTClient().Get().Resource("inspectresults").VersionedParams(&listOptions, scheme.ParameterCodec).Do(ctx).Into(&results)
	if err != nil || len(results.Items) == 0 {
		return errors.Errorf("result not exist")
	}
	var resultCollection = make(map[string][]renderNode, 5)

	for _, item := range results.Items {
		if item.Spec.OpaResult.ResourceResults != nil {
			list := getOpaList(item.Spec.OpaResult.ResourceResults)
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
		if item.Spec.FilterResult != nil {
			filter := getFileFilter(item.Spec.FilterResult)
			resultCollection[constant.FileFilter] = filter
		}
		if item.Spec.ComponentResult != nil {
			component := getComponent(item.Spec.ComponentResult)
			resultCollection[constant.Component] = component
		}
	}
	var task v1alpha2.InspectTask
	err = Clients.VersionClientSet.KubeeyeV1alpha2().RESTClient().Get().Resource("inspecttasks").Name(TaskName).Do(ctx).Into(&task)
	if err != nil {
		klog.Errorf("Failed to get  inspect task. err:%s", err)
		return err
	}
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

	data := map[string]interface{}{"title": task.CreationTimestamp.Format("2006-01-02 15:04"), "overview": ruleNumber, "details": resultCollection}
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

func getFileFilter(fileResult map[string][]v1alpha2.FileChangeResultItem) []renderNode {
	var villeinage []renderNode
	header := renderNode{Header: true, Children: []renderNode{
		{Text: "nodeName"},
		{Text: "FileName"},
		{Text: "Path"},
		{Text: "Issues"}},
	}
	villeinage = append(villeinage, header)

	for k, v := range fileResult {
		for _, result := range v {
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
	name := ParseFileName(p, fmt.Sprintf("巡检报告(%s).html", time.Now().Format("2006-01-02 15:04")))
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

func assertBoolBackString(bool string) string {
	if bool == "" {
		return bool
	}
	if bool == "false" {
		return "true"
	}
	return "false"

}
