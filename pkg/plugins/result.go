package plugins

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
    "github.com/kubesphere/kubeeye/pkg/kube"
)

func PluginsResults(pluginsList []string)  {

    pluginsResults := []kubeeyev1alpha1.PluginsResult{}
    pluginsResult := PluginsResult(context.TODO(), pluginsList)
    for Result := range pluginsResult {
        pluginsResults = append(pluginsResults, Result)
    }
    kube.PluginsResultsChan <- pluginsResults
}


func PluginsResult(ctx context.Context, pluginsList []string) <- chan kubeeyev1alpha1.PluginsResult {
    pluginsResultChan := make(chan kubeeyev1alpha1.PluginsResult)

    if len(pluginsList) != 0 {

        var wg sync.WaitGroup
        wg.Add(len(pluginsList))

        for _, pluginName := range pluginsList {
            go func(pluginName string) {
                defer wg.Done()

                pluginsResultChan <- GetPluginsResult(pluginName)
            }(pluginName)
        }

        go func() {
            defer close(pluginsResultChan)
            wg.Wait()
        }()
    }
    return pluginsResultChan
}

func GetPluginsResult(pluginName string) kubeeyev1alpha1.PluginsResult {
    result := kubeeyev1alpha1.PluginsResult{}
    result.Name = pluginName
    // Check if hunter service is ready
    _, err := http.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/healthz",pluginName))
    if err != nil {
        return result
    }

    tr := &http.Transport{
        IdleConnTimeout: 30 * time.Second,  //the maximum amount of time an idle connection will remain idle before closing itself.
        DisableCompression: true,       //prevents the Transport from requesting compression with an "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
        WriteBufferSize: 10 << 10 ,     //specifies the size of the write buffer to 10KB used when writing to the transport.
    }
    client := &http.Client{Transport: tr}
    resp, err := client.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/plugins",pluginName))
    if err != nil {
        return result
    }

   result.Results ,err =DecodePluginResult(resp)
   if err != nil {
       return result
   }
    return result
}

func DecodePluginResult( resp *http.Response ) (result string, err error) {
    var pluginResult interface{}
    if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
        return "", err
    }
    r, err := json.Marshal(&pluginResult)
    if err != nil {
        return "",err
    }
    result = string(r)
    return result, nil
}
