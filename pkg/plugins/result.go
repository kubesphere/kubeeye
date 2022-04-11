package plugins

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
    "github.com/pkg/errors"
)

func GetPluginsResult(pluginName string) (result v1alpha1.PluginsResult,err error) {
    // Check if hunter service is ready
    _, err = http.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/healthz",pluginName))
    if err != nil {
        return result, errors.Wrap(err, fmt.Sprintf("Unable to access %s service", pluginName))
    }

    tr := &http.Transport{
        IdleConnTimeout: 30 * time.Second,  //the maximum amount of time an idle connection will remain idle before closing itself.
        DisableCompression: true,       //prevents the Transport from requesting compression with an "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
        WriteBufferSize: 10 << 10 ,     //specifies the size of the write buffer to 10KB used when writing to the transport.
    }
    client := &http.Client{Transport: tr}
    resp, err := client.Get(fmt.Sprintf("http://%s.kubeeye-system.svc/plugins",pluginName))
    if err != nil {
        return result, errors.Wrap(err, fmt.Sprintf("Unable to get result of %s service", pluginName))
    }

    result.Name = pluginName
    //switch {
    //case pluginName == "kubebench":
    //    var pluginResult KubeBenchResults
    //    result.Results ,err =DecodePluginResult(pluginResult, resp)
    //    if err != nil {
    //        return result, errors.Wrap(err, fmt.Sprintf("Unable to decode result of %s service", pluginName))
    //    }
    //case pluginName == "kubehunter":
    //    var pluginResult *KubeHunterResults
    //    result.Results ,err =DecodePluginResult(pluginResult, resp)
    //    if err != nil {
    //        return result, errors.Wrap(err, fmt.Sprintf("Unable to decode result of %s service", pluginName))
    //    }
    //case pluginName == "kubescape":
    //    var pluginResult []reporthandling.FrameworkReport
    //    result.Results ,err =DecodePluginResult(pluginResult, resp)
    //    if err != nil {
    //        return result, errors.Wrap(err, fmt.Sprintf("Unable to decode result of %s service", pluginName))
    //    }
    //}

   result.Results ,err =DecodePluginResult(resp)
   if err != nil {
       return result, errors.Wrap(err, fmt.Sprintf("Unable to decode result of %s service", pluginName))
   }
    return result, nil
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
