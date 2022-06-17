package kubeeye

import (
	v1alpha12 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
)

func MergePluginsResults(pluginsResults []v1alpha12.PluginsResult, newResult v1alpha12.PluginsResult) []v1alpha12.PluginsResult {
	var newPluginResults []v1alpha12.PluginsResult
	existPluginsMap := make(map[string]bool)
	for _, result := range pluginsResults {
		existPluginsMap[result.Name] = true
	}

	if existPluginsMap[newResult.Name] {
		for _, result := range pluginsResults {
			if result.Name == newResult.Name {
				result = newResult
			}
			newPluginResults = append(newPluginResults, result)
		}
	} else {
		newPluginResults = append(pluginsResults, newResult)
	}

	return newPluginResults
}
