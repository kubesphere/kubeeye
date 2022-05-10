package plugins

import (
    "math/rand"
    "time"

    kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
)

// NotReadyPluginsList is used to get the list of not-ready plugins.
func NotReadyPluginsList(pluginsResults []kubeeyev1alpha1.PluginsResult, pluginsList []string, ) []string {
    pluginsMap := make(map[string]bool)
    var notReadyPluginsList []string
    for _, result := range pluginsResults {
        pluginsMap[result.Name] = result.Ready
    }

    for _, pluginsname := range pluginsList {
        if !pluginsMap[pluginsname] {
            notReadyPluginsList = append(notReadyPluginsList, pluginsname)
        }
    }
    return notReadyPluginsList
}

func RandomPluginName(pluginsList []string) (randomPluginName string) {
    rand.Seed(time.Now().Unix())
    l := len(pluginsList)
    randomIndex := rand.Intn(l)
    randomPluginName = pluginsList[randomIndex]
    return randomPluginName
}