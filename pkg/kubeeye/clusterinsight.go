package kubeeye

import (
	"bytes"
	"context"

	v1alpha12 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

func GetClusterInsights(ctx context.Context, clientSet *versioned.Clientset) (clusterInsight *v1alpha12.ClusterInsight, err error) {
	listOptions := metav1.ListOptions{}
	clusterInsightList, err := clientSet.KubeeyeV1alpha1().ClusterInsights().List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	if len(clusterInsightList.Items) > 0 {
		clusterInsight = &clusterInsightList.Items[0]
		return clusterInsight, nil
	}
	return nil, errors.Wrap(err, "ClusterInsight not ready")
}

func UpdateClusterInsights(ctx context.Context, clientSet *versioned.Clientset, clusterInsight *v1alpha12.ClusterInsight, resp []byte, result v1alpha12.PluginsResult) error {
	updateOptions := metav1.UpdateOptions{}
	ext := runtime.RawExtension{}

	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(resp), 4096)
	if err := d.Decode(&ext); err != nil {
		return err
	}
	result.Result = ext
	result.Ready = true

	pluginsResult := MergePluginsResults(clusterInsight.Status.PluginsResults, result)
	clusterInsight.Status.PluginsResults = pluginsResult

	_, err := clientSet.KubeeyeV1alpha1().ClusterInsights().UpdateStatus(ctx, clusterInsight, updateOptions)
	if err != nil {
		return err
	}
	klog.Infof("update plugin %s result successful", result.Name)

	return nil
}

func ClearClusterInsightStatus(ClusterInsight *v1alpha12.ClusterInsight) *v1alpha12.ClusterInsight {
	ClusterInsight.Status = v1alpha12.ClusterInsightStatus{}
	return ClusterInsight
}
