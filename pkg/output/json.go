package output

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubeeye/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"time"
)

func JsonOut(ctx context.Context, clients *kube.KubernetesClient, outPath string, TaskName string, TaskNameSpace string) error {
	results, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults(TaskNameSpace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{constant.LabelName: TaskName})),
	})
	if err != nil || len(results.Items) == 0 {
		return errors.Errorf("result not exist")
	}
	var result = make(map[string]interface{}, 3)
	for _, item := range results.Items {
		if item.Spec.OpaResult.ResourceResults != nil {
			result[constant.Opa] = item.Spec.OpaResult.ResourceResults
		}
		if item.Spec.PrometheusResult != nil {
			result[constant.Prometheus] = item.Spec.PrometheusResult
		}
		if item.Spec.NodeInfoResult != nil {
			result["nodeInfo"] = item.Spec.NodeInfoResult
		}
		if item.Spec.FilterResult != nil {
			result[constant.FileFilter] = item.Spec.FilterResult
		}
		if item.Spec.ComponentResult != nil {
			result[constant.Component] = item.Spec.ComponentResult
		}
	}
	marshal, err := json.Marshal(result)
	if err != nil {
		return err
	}

	name := ParseFileName(outPath, fmt.Sprintf("巡检报告(%s).json", time.Now().Format("2006-01-02")))

	jsonFile, err := os.Create(name)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	jsonFile.Write(marshal)
	return nil
}
