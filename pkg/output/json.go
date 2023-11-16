package output

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
	"os"
	"strconv"
	"strings"
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

	if results.Spec.ServiceConnectResult != nil {
		result[constant.ServiceConnect] = results.Spec.ServiceConnectResult
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

func ParseCustomizedStruct(data *kubeeyev1alpha2.InspectResult) map[string]interface{} {
	status := ParseApiStatus(data.Spec.PrometheusResult)
	resources := ParseResources(data.Spec.PrometheusResult)
	top := ParseResourcesTop(data.Spec.PrometheusResult)
	metric := ParseOtherMetric(data.Spec.PrometheusResult)

	return map[string]interface{}{"name": data.Name, "cluster": data.Spec.InspectCluster.Name, "component_status": data.Spec.ComponentResult, "api_status": status, "resources_usage": resources, "resources_usage_top": top, "metric": metric}
}

func ParseApiStatus(result []kubeeyev1alpha2.PrometheusResult) map[string]string {
	apiStatus := make(map[string]string, 2)
	for _, pro := range result {
		if strings.ToUpper(pro.Name) == strings.ToUpper("apiserver_request_latencies") {
			apiStatus["apiserver_request_latencies"] = pro.ParseString()["value"]
		}
		if strings.ToUpper(pro.Name) == strings.ToUpper("apiserver_request_rate") {
			apiStatus["apiserver_request_rate"] = pro.ParseString()["value"]
		}

	}
	return apiStatus
}
func ParseResources(result []kubeeyev1alpha2.PrometheusResult) map[string]map[string]float64 {
	metricData := make(map[string]float64)
	metrics := []string{"cluster_cpu_usage", "cluster_cpu_total", "cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_disk_size_usage", "cluster_disk_size_capacity", "cluster_pod_running_count", "cluster_pod_quota"}
	for _, pro := range result {
		if slices.Contains(metrics, strings.ToLower(pro.Name)) {
			float, err := strconv.ParseFloat(pro.ParseString()["value"], 64)
			if err != nil {
				float = 0
			}
			metricData[pro.Name] = float
		}
	}

	resourcesData := make(map[string]map[string]float64)

	resourcesData["cpu"] = map[string]float64{"total": metricData["cluster_cpu_total"], "usage": metricData["cluster_cpu_usage"], "percent": metricData["cluster_cpu_usage"] / metricData["cluster_cpu_total"]}
	resourcesData["memory"] = map[string]float64{"total": metricData["cluster_memory_total"], "usage": metricData["cluster_memory_usage_wo_cache"], "percent": metricData["cluster_memory_usage_wo_cache"] / metricData["cluster_memory_total"]}
	resourcesData["disk"] = map[string]float64{"total": metricData["cluster_disk_size_capacity"], "usage": metricData["cluster_disk_size_usage"], "percent": metricData["cluster_disk_size_usage"] / metricData["cluster_disk_size_capacity"]}
	resourcesData["pod"] = map[string]float64{"total": metricData["cluster_pod_quota"], "usage": metricData["cluster_pod_running_count"], "percent": metricData["cluster_pod_running_count"] / metricData["cluster_pod_quota"]}
	return resourcesData
}

func ParseResourcesTop(result []kubeeyev1alpha2.PrometheusResult) map[string]map[string]string {
	resourcesTopMetrics := []string{"node_cpu_utilisation", "node_memory_utilisation", "node_disk_size_utilisation"}
	metricsTop := make(map[string]map[string]string)
	for _, r := range result {
		if slices.Contains(resourcesTopMetrics, strings.ToLower(r.Name)) {
			p := r.ParseString()
			m := metricsTop[r.Name]
			if m == nil {
				metricsTop[r.Name] = map[string]string{"cluster": p["cluster"], "node": p["node"], "value": p["value"]}
			} else {
				mF, err := strconv.ParseFloat(m["value"], 64)
				if err != nil {
					mF = 0
				}
				pF, err := strconv.ParseFloat(p["value"], 64)
				if err != nil {
					pF = 0
				}
				if mF < pF {
					metricsTop[r.Name] = map[string]string{"cluster": p["cluster"], "node": p["node"], "value": p["value"]}
				}
			}
		}
	}
	return metricsTop
}
func ParseOtherMetric(result []kubeeyev1alpha2.PrometheusResult) map[string]map[string]string {
	otherMetricData := make(map[string]map[string]string)
	for _, r := range result {
		if strings.ToLower(r.Name) == strings.ToLower("node_load_15") {
			p := r.ParseString()
			fmt.Println(p)
		}
		if strings.ToLower(r.Name) == strings.ToLower("filesystem_readonly") {
			p := r.ParseString()
			otherMetricData[r.Name] = map[string]string{"cluster": p["cluster"], "node": p["instance"], "mountpoint": p["mountpoint"], "value": p["value"]}
		}
		if strings.ToLower(r.Name) == strings.ToLower("filesystem_avail") {
			p := r.ParseString()
			otherMetricData[r.Name] = map[string]string{"cluster": p["cluster"], "node": p["instance"], "mountpoint": p["mountpoint"], "value": p["value"]}
		}
		if strings.ToLower(r.Name) == strings.ToLower("harbor_status") {
			p := r.ParseString()
			fmt.Println("harborStatus", p)
		}
		if strings.ToLower(r.Name) == strings.ToLower("harbor_copy") {
			p := r.ParseString()
			fmt.Println("harborCopy", p)
		}
		if strings.ToLower(r.Name) == strings.ToLower("etcd_back_up") {
			p := r.ParseString()
			fmt.Println("etcdBackup", p)
		}
	}
	return otherMetricData
}
