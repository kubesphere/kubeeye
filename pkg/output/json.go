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

func JsonOut(ctx context.Context, clients *kube.KubernetesClient, outPath string, TaskName string) error {
	results, err := clients.VersionClientSet.KubeeyeV1alpha2().InspectResults().Get(ctx, TaskName, metav1.GetOptions{})
	if err != nil {
		return errors.Errorf("result not exist")
	}
	var result = make(map[string]interface{}, 3)

	if results.Spec.OpaResult.ResourceResults != nil {
		result[constant.Opa] = results.Spec.OpaResult.ResourceResults
	}
	if results.Spec.PrometheusResult != nil {
		result[constant.Prometheus] = results.Spec.PrometheusResult
	}
	if results.Spec.NodeInfoResult != nil {
		result["nodeInfo"] = results.Spec.NodeInfoResult
	}
	if results.Spec.ComponentResult != nil {
		result[constant.Component] = results.Spec.ComponentResult
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
	_, err = jsonFile.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
