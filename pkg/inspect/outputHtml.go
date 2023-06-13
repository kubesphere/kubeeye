package inspect

import (
	"context"
	"github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
	"strings"
)

func HtmlOutput(clients *kube.KubernetesClient, outPath *string, taskName string, namespace string) error {

	results, _ := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{constant.LabelName: taskName})),
	})
	var resultCollection = make(map[string][][]string, 5)

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
	data := map[string]interface{}{"overview": ruleNumber, "details": resultCollection}
	err = renderView(data)
	if err != nil {
		return err
	}
	return nil
}

func GetOpaList(result []v1alpha2.ResourceResult) [][]string {
	OpsList := [][]string{
		{"NameSpace", "Kind", "Name", "Level", "Message", "Reason"},
	}
	for _, resourceResult := range result {

		for _, item := range resourceResult.ResultItems {
			items := []string{resourceResult.NameSpace, resourceResult.ResourceType, resourceResult.Name}
			items = append(items, item.Level, item.Message, item.Reason)
			OpsList = append(OpsList, items)
		}
	}

	return OpsList
}

func getPrometheus(pro [][]map[string]string) [][]string {
	var prometheus [][]string
	for _, p := range pro {
		var header []string
		for _, val := range p {
			if len(header) == 0 {
				for k := range val {
					header = append(header, k)
				}
				prometheus = append(prometheus, header)
			}
			var value []string
			for i := range header {
				value = append(value, val[header[i]])
			}
			prometheus = append(prometheus, value)
		}
	}
	return prometheus
}

func getFileChange(infoResult map[string]v1alpha2.NodeInfoResult) [][]string {
	villeinage := [][]string{{"nodeName", "type", "name", "value"}}

	for k, v := range infoResult {
		for _, item := range v.FileChangeResult {

			if item.Issues != nil && len(item.Issues) > 0 {

				villeinage = append(villeinage, []string{k, constant.FileChange, item.FileName, strings.Join(item.Issues, ",")})
			}

		}

	}
	return villeinage
}

func getSysctl(infoResult map[string]v1alpha2.NodeInfoResult) [][]string {
	villeinage := [][]string{{"nodeName", "type", "name", "value", "assert"}}

	for k, v := range infoResult {
		for _, item := range v.SysctlResult {

			villeinage = append(villeinage, []string{k, constant.Sysctl, item.Name, *item.Value, utils.FormatBool(item.Assert)})

		}

	}
	return villeinage
}

func getSystemd(infoResult map[string]v1alpha2.NodeInfoResult) [][]string {
	villeinage := [][]string{{"nodeName", "type", "name", "value", "assert"}}

	for k, v := range infoResult {
		for _, item := range v.SystemdResult {

			villeinage = append(villeinage, []string{k, constant.Systemd, item.Name, *item.Value, utils.FormatBool(item.Assert)})

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
