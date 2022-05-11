package plugins

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/go-logr/logr"
    kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
    "github.com/kubesphere/kubeeye/pkg/kube"
    "github.com/kubesphere/kubeeye/plugins/kubebench/pkg"
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
    result ,err =OnetypeOfResult(resp, result)
    if err != nil {
        return result, err
    }
    result.Ready = true
    return result, nil
}

func OnetypeOfResult(resp *http.Response, result kubeeyev1alpha1.PluginsResult) (kubeeyev1alpha1.PluginsResult, error) {
    var pluginResult pkg.KubeBenchResults
    var fmtResults []kubeeyev1alpha1.AuditResults
    var auditResult kubeeyev1alpha1.AuditResults
    var resultReceiver kubeeyev1alpha1.ResultInfos
    var resultItems kubeeyev1alpha1.ResultItems
    if result.Name == "kubebench" {
        if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
            return result, err
        }
    
        for _, control := range pluginResult.Controls {
            auditResult.NameSpace = ""
            for _, group := range control.Groups {
                for _, check := range group.Checks {
                    resourceInfos := kubeeyev1alpha1.ResourceInfos{}
                    if check.State != "PASS" {
                        resultReceiver.ResourceType = group.Text
                        resourceInfos.Name = check.Text
                        resultItems.Message= check.Remediation
                        resultItems.Reason = check.Reason
                        resultItems.Level = "warring"
                    } else {
                        continue
                    }
                    resourceInfos.ResultItems = append(resourceInfos.ResultItems, resultItems)
                    resultReceiver.ResourceInfos = resourceInfos
                    auditResult.ResultInfos = append(auditResult.ResultInfos, resultReceiver)
                }
            }
        }
    
        fmtResults = append(fmtResults, auditResult)
        result.Results.KubeBenchResults = fmtResults
    } else if result.Name == "kubehunter" {
        var pluginResult kubeeyev1alpha1.KubeHunterResults
        if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
            return result, err
        }
        result.Results.KubeHunterResults = append(result.Results.KubeHunterResults, pluginResult)
    // } else if result.Name == "kubescape" {
    //     var pluginResult []reporthandling.FrameworkReport
    //     if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
    //         return result, err
    //     }
    //     for _, report := range pluginResult {
    //         for _, controlReport := range report.ControlReports {
    //             for _, ruleReport := range controlReport.RuleReports {
    //                 if ruleReport.ResourceUniqueCounter.FailedResources != 0 && ruleReport.ResourceUniqueCounter.WarningResources != 0 {
    //                     for _, respons := range ruleReport.RuleResponses {
    //                         resourceInfos := kubeeyev1alpha1.ResourceInfos{}
    //                         resources := respons.AlertObject.K8SApiObjects
    //                         resource := &unstructured.Unstructured{Object: resources[0]}
    //
    //                         auditResult.NameSpace = resource.GetNamespace()
    //                         resultReceiver.ResourceType = resource.GetKind()
    //                         resourceInfos.Name = resource.GetName()
    //                         resultItems.Level = "warring"
    //                         resultItems.Message = controlReport.Description
    //                         resultItems.Reason = controlReport.Remediation
    //
    //                         resourceInfos.ResultItems = append(resourceInfos.ResultItems , resultItems)
    //                         resultReceiver.ResourceInfos = resourceInfos
    //                     }
    //                     auditResult.ResultInfos = append(auditResult.ResultInfos, resultReceiver)
    //                 } else {
    //                     continue
    //                 }
    //             }
    //         }
    //         fmtResults = append(fmtResults, auditResult)
    //     }
    //     result.Results.KubescapeResults = fmtResults
    } else {
        var pluginResult interface{}
        if err := json.NewDecoder(resp.Body).Decode(&pluginResult); err != nil {
            return result, err
        }
        r, err := json.Marshal(&pluginResult)
        if err != nil {
            return result,err
        }
        result.Results.StringResults = string(r)
    }
    return result, nil
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