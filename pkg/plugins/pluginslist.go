package plugins

import (
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	kubeeyepluginsv1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeyeplugins/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/conf"
)

// NotReadyPluginsList is used to get the list of not-ready plugins.
func NotReadyPluginsList(pluginsResults []kubeeyev1alpha1.PluginsResult, pluginsList *kubeeyepluginsv1alpha1.PluginSubscriptionList) []string {
	pluginsMap := make(map[string]bool)
	var notReadyPluginsList []string
	for _, result := range pluginsResults {
		pluginsMap[result.Name] = result.Ready
	}

	for _, plugins := range pluginsList.Items {
		pluginsname := plugins.Name
		if !pluginsMap[pluginsname] && plugins.Status.State == conf.PluginInstalled {
			notReadyPluginsList = append(notReadyPluginsList, pluginsname)
		}
	}
	return notReadyPluginsList
}
