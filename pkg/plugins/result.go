package plugins

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/go-logr/logr"
    kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
    "github.com/kubesphere/kubeeye/pkg/kube"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/util/yaml"
)

func PluginsAudit(logs logr.Logger, pluginName string)  {
    pluginsResult, err := GetPluginsResult(logs, pluginName)
    if err != nil {
        logs.Error(err, fmt.Sprintf("failed to get the result of the plugin %s", pluginName))
    }
    kube.PluginResultChan <- pluginsResult
}

func GetPluginsResult(logs logr.Logger,pluginName string) (kubeeyev1alpha1.PluginsResult, error) {
    result := kubeeyev1alpha1.PluginsResult{}
    result.Name = pluginName
    // Check if service is ready
    logs.Info(fmt.Sprintf("check the health of the plugin %s", pluginName))
    _, err := http.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/healthz",pluginName))
    if err != nil {
        return result, err
    }

    tr := &http.Transport{
        IdleConnTimeout: 30 * time.Second,  // the maximum amount of time an idle connection will remain idle before closing itself.
        DisableCompression: true,       // prevents the Transport from requesting compression with an "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
        WriteBufferSize: 10 << 10 ,     // specifies the size of the write buffer to 10KB used when writing to the transport.
    }
    client := &http.Client{Transport: tr}
    // get the result by plugin service
    logs.Info(fmt.Sprintf("get the result of the plugin %s", pluginName))
    resp, err := client.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/plugins",pluginName))
    if err != nil {
        return result, err
    }

    // todo We need to save the result as the result's own structure instead of the string
    logs.Info(fmt.Sprintf("decode the result of the plugin %s", pluginName))
    result.Result ,err = DecodeResult(resp)
    if err != nil {
        return result, err
    }
    result.Ready = true
    return result, nil
}

func DecodeResult(resp *http.Response) (runtime.RawExtension, error) {
    ext := runtime.RawExtension{}
    var pluginResult interface{}
    if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
        return ext, err
    }
    r, err := json.Marshal(&pluginResult)
    if err != nil {
        return ext, err
    }

    d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(r), 4096)
    if err := d.Decode(&ext); err != nil {
        return ext, err
    }
    return ext, nil
}

func MergePluginsResults(pluginsResults []kubeeyev1alpha1.PluginsResult, newResult kubeeyev1alpha1.PluginsResult) []kubeeyev1alpha1.PluginsResult  {
    var tmpResults []kubeeyev1alpha1.PluginsResult
    existPluginsMap := make(map[string]bool)
    for _, result := range pluginsResults {
        existPluginsMap[result.Name] = true
    }

    if existPluginsMap[newResult.Name] {
        for _, result := range pluginsResults {
            if result.Name == newResult.Name {
                result = newResult
            }
            tmpResults = append(tmpResults, result)
        }
    } else {
        tmpResults = append(pluginsResults, newResult)
    }

    return tmpResults
}