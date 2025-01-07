package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	kubeeyev1alpha2 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha2"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/utils"
	"github.com/prometheus/client_golang/api"
	apiprometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"sort"
	"strings"
	"time"
)

type prometheusInspect struct {
}

func init() {
	RuleOperatorMap[constant.Prometheus] = &prometheusInspect{}
}

func (o *prometheusInspect) RunInspect(ctx context.Context, rules []kubeeyev1alpha2.JobRule, clients *kube.KubernetesClient, currentJobName string, informers informers.SharedInformerFactory, ownerRef ...metav1.OwnerReference) ([]byte, error) {

	_, exist, phase := utils.ArrayFinds(rules, func(m kubeeyev1alpha2.JobRule) bool {
		return m.JobName == currentJobName
	})
	if exist {
		var promRules []kubeeyev1alpha2.PrometheusRule
		err := json.Unmarshal(phase.RunRule, &promRules)
		if err != nil {
			klog.Error(err, " Failed to marshal kubeeye result")
			return nil, err
		}

		var promRuleResult []kubeeyev1alpha2.PrometheusResult
		for _, promRule := range promRules {
			if promRule.Prometheus == nil {
				klog.Error("prometheus config is nil")
				return nil, err
			}
			queryApi, err := NewPrometheusAPI(promRule.Prometheus)
			if err != nil {
				klog.Errorf("failed to create prometheus api for %s", promRule.Prometheus.Endpoint)
				return nil, err
			}
			query, _, err := queryApi.Query(ctx, promRule.Rule, time.Now())
			if err != nil {
				klog.Errorf("failed to query rule:%s", promRule.Rule)
				return nil, err
			}
			marshal, err := json.Marshal(query)
			if err != nil {
				klog.Error("marshal modal Samples failed", err)
				continue
			}
			var queryResults model.Samples
			err = json.Unmarshal(marshal, &queryResults)
			if err != nil {
				klog.Error("unmarshal modal Samples failed", err)
				continue
			}
			for _, result := range queryResults {
				promRuleResult = append(promRuleResult, kubeeyev1alpha2.PrometheusResult{
					Result:         toString(result),
					Rule:           promRule.Rule,
					RawDataEnabled: promRule.RawDataEnabled,
					BaseResult: kubeeyev1alpha2.BaseResult{
						Name:   promRule.Name,
						Assert: true,
						Level:  promRule.Level,
					},
				})
			}
		}

		marshal, err := json.Marshal(promRuleResult)
		if err != nil {
			return nil, err
		}
		return marshal, nil
	}
	return nil, nil
}

func (o *prometheusInspect) GetResult(runNodeName string, resultCm *corev1.ConfigMap, resultCr *kubeeyev1alpha2.InspectResult) (*kubeeyev1alpha2.InspectResult, error) {
	var prometheus []kubeeyev1alpha2.PrometheusResult
	json.Unmarshal(resultCm.BinaryData[constant.Data], &prometheus)
	if prometheus == nil {
		return resultCr, nil
	}

	resultCr.Spec.PrometheusResult = prometheus

	return resultCr, nil
}

func toString(val *model.Sample) string {
	if val == nil {
		return "{}"
	}

	metricName, hasName := val.Metric[model.MetricNameLabel]
	numLabels := len(val.Metric) - 1
	if !hasName {
		numLabels = len(val.Metric)
	}
	labelStrings := make([]string, 0, numLabels)
	for label, value := range val.Metric {
		if label != model.MetricNameLabel {
			labelStrings = append(labelStrings, fmt.Sprintf("%q=%q", label, value))
		}
	}
	labelStrings = append(labelStrings, fmt.Sprintf("%q=%q", "value", val.Value))
	labelStrings = append(labelStrings, fmt.Sprintf("%q=%q", "timestamp", val.Timestamp))
	labelStrings = append(labelStrings, fmt.Sprintf("%q=%q", "metricName", metricName))

	sort.Strings(labelStrings)
	return fmt.Sprintf("{%s}", strings.Join(labelStrings, ", "))
}

// NewPrometheusAPI creates a new Prometheus API client.
func NewPrometheusAPI(prometheus *kubeeyev1alpha2.PrometheusConfig) (apiprometheusv1.API, error) {
	if prometheus.Endpoint == "" {
		return nil, fmt.Errorf("prometheus endpoint is empty")
	}

	httpConfig := config.HTTPClientConfig{}
	basicToken := prometheus.GetBasicToken()
	if basicToken != "" {
		httpConfig.HTTPHeaders = &config.Headers{
			Headers: map[string]config.Header{
				"Authorization": config.Header{
					Values: []string{"Basic " + basicToken},
				},
			},
		}
	}

	bearerToken := prometheus.GetBearerToken()
	if bearerToken != "" {
		httpConfig.BearerToken = config.Secret(bearerToken)
	}

	if prometheus.InsecureSkipVerify != nil {
		httpConfig.TLSConfig = config.TLSConfig{
			InsecureSkipVerify: *prometheus.InsecureSkipVerify,
		}
	}

	tr, err := config.NewRoundTripperFromConfig(httpConfig, "prometheus")
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(api.Config{
		Address:      prometheus.Endpoint,
		RoundTripper: tr,
	})
	if err != nil {
		return nil, err
	}
	return apiprometheusv1.NewAPI(client), nil
}
