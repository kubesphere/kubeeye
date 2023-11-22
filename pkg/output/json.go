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
	status := ParseApiStatus(data)
	resources := ParseResources(data)
	top := ParseResourcesTop(data)
	metric := ParseOtherMetric(data)
	count := OverviewCount(metric, data.Spec.ComponentResult)

	return map[string]interface{}{"name": data.Name, "cluster": data.Spec.InspectCluster.Name, "overview_count": count, "component_status": data.Spec.ComponentResult, "api_status": status, "resources_usage": resources, "resources_usage_top": top, "metric": metric}
}

func ParseApiStatus(result *kubeeyev1alpha2.InspectResult) map[string]string {
	apiStatus := make(map[string]string, 2)
	for _, pro := range result.Spec.PrometheusResult {
		if strings.ToUpper(pro.Name) == strings.ToUpper("apiserver_request_latencies") {
			apiStatus["apiserver_request_latencies"] = pro.ParseString()["value"]
		}
		if strings.ToUpper(pro.Name) == strings.ToUpper("apiserver_request_rate") {
			apiStatus["apiserver_request_rate"] = pro.ParseString()["value"]
		}

	}
	return apiStatus
}
func ParseResources(result *kubeeyev1alpha2.InspectResult) map[string]map[string]float64 {
	metricData := make(map[string]float64)
	metrics := []string{"cluster_cpu_usage", "cluster_cpu_total", "cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_disk_size_usage", "cluster_disk_size_capacity", "cluster_pod_running_count", "cluster_pod_quota"}
	for _, pro := range result.Spec.PrometheusResult {
		if slices.Contains(metrics, strings.ToLower(pro.Name)) {
			float, err := strconv.ParseFloat(pro.ParseString()["value"], 64)
			if err != nil {
				float = 0
			}
			metricData[pro.Name] = float
		}
	}

	resourcesData := make(map[string]map[string]float64)
	if len(metricData) > 0 {
		resourcesData["cpu"] = map[string]float64{"total": metricData["cluster_cpu_total"], "usage": metricData["cluster_cpu_usage"], "percent": metricComputed(metricData["cluster_cpu_usage"], metricData["cluster_cpu_total"])}
		resourcesData["memory"] = map[string]float64{"total": metricData["cluster_memory_total"], "usage": metricData["cluster_memory_usage_wo_cache"], "percent": metricComputed(metricData["cluster_memory_usage_wo_cache"], metricData["cluster_memory_total"])}
		resourcesData["disk"] = map[string]float64{"total": metricData["cluster_disk_size_capacity"], "usage": metricData["cluster_disk_size_usage"], "percent": metricComputed(metricData["cluster_disk_size_usage"], metricData["cluster_disk_size_capacity"])}
		resourcesData["pod"] = map[string]float64{"total": metricData["cluster_pod_quota"], "usage": metricData["cluster_pod_running_count"], "percent": metricComputed(metricData["cluster_pod_running_count"], metricData["cluster_pod_quota"])}
	}
	return resourcesData
}

func ParseResourcesTop(result *kubeeyev1alpha2.InspectResult) map[string]map[string]string {
	resourcesTopMetrics := []string{"node_cpu_utilisation", "node_memory_utilisation", "node_disk_size_utilisation"}
	metricsTop := make(map[string]map[string]string)
	for _, r := range result.Spec.PrometheusResult {
		if slices.Contains(resourcesTopMetrics, strings.ToLower(r.Name)) {
			p := r.ParseString()
			m := metricsTop[r.Name]
			if m == nil {
				metricsTop[r.Name] = map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["node"], "value": p["value"]}
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
					metricsTop[r.Name] = map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["node"], "value": p["value"]}
				}
			}
		}
	}
	return metricsTop
}
func ParseOtherMetric(result *kubeeyev1alpha2.InspectResult) map[string][]map[string]string {
	otherMetricData := make(map[string][]map[string]string)
	for _, r := range result.Spec.PrometheusResult {
		if strings.ToLower(r.Name) == strings.ToLower("node_load_15") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["node"], "value": p["value"]})
		}
		if strings.ToLower(r.Name) == strings.ToLower("filesystem_readonly") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["instance"], "mountpoint": p["mountpoint"], "value": p["value"]})
		}
		if strings.ToLower(r.Name) == strings.ToLower("filesystem_avail") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["instance"], "mountpoint": p["mountpoint"], "value": p["value"]})
		}
		if strings.ToLower(r.Name) == strings.ToLower("harbor_health") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "name": p["name"], "value": p["value"]})
		}
		if strings.ToLower(r.Name) == strings.ToLower("harbor_ref_work_replication") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "value": p["value"]})
		}
		if strings.ToLower(r.Name) == strings.ToLower("node_loadapp_etcd_backup_status") {
			p := r.ParseString()
			otherMetricData[r.Name] = append(otherMetricData[r.Name], map[string]string{"cluster": result.Spec.InspectCluster.Name, "node": p["instance"], "value": p["value"]})
		}
	}
	return otherMetricData
}

func metricComputed(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b

}

func OverviewCount(metric map[string][]map[string]string, com []kubeeyev1alpha2.ComponentResultItem) map[string]int {
	count := map[string]int{"node_load_15": 0, "filesystem_readonly": 0, "filesystem_avail": 0, "harbor_health": 0, "harbor_ref_work_replication": 0, "node_loadapp_etcd_backup_status": 0}

	for key := range count {
		count[key] = len(metric[key])
	}
	componentCount := 0
	for _, c := range com {
		if c.Assert {
			componentCount += 1
		}
	}
	count["component"] = componentCount
	return count
}
