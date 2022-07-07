/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeeye

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/pkg/kubeeye"
	"github.com/kubesphere/kubeeye/pkg/suggests"
	"k8s.io/klog/v2"
)

func PluginsResultsReceiver() {
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/healthz", http.HandlerFunc(health))
	ServeMux.Handle("/plugins", http.HandlerFunc(PluginsResult))
	ServeMux.Handle("/suggestlist", http.HandlerFunc(suggestList))
	ServeMux.Handle("/suggest", http.HandlerFunc(suggest))
	log.Fatalln(http.ListenAndServe(":8888", ServeMux))
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed")
		klog.Info("Method Not Allowed")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PluginsResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed")
		klog.Info("Method Not Allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Parse form failed : %s \n", err)
		klog.Info("Parse form failed")
		return
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "get plugin result failed : %s \n", err)
		klog.Info("get plugin result failed")
		return
	}

	result := kubeeyev1alpha1.PluginsResult{}
	if len(r.Form) > 0 {
		pluginName := r.Form.Get("name")
		if pluginName != "" {
			result.Name = pluginName
		} else {
			result.Name = "UnknownPlugin"
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "start update %s result to clusterInsight \n", result.Name)
	klog.Infof("start update %s result to clusterInsight", result.Name)

	go UpdatePluginsResults(resp, result)
}

func UpdatePluginsResults(resp []byte, result kubeeyev1alpha1.PluginsResult) {
	ctx := context.TODO()

	clientSet, err := kube.GetClientSetInCluster()
	if err != nil {
		klog.Error("update plugins results failed: get client set failed", err)
		return
	}

	clusterInsight, err := kubeeye.GetClusterInsights(ctx, clientSet)
	if err != nil {
		klog.Error("update plugins results failed: get clusterInsight failed", err)
		return
	}

	if err := kubeeye.UpdateClusterInsights(ctx, clientSet, clusterInsight, resp, result); err != nil {
		klog.Error("update plugins results failed: update clusterInsight failed", err)
		return
	}
}

func suggest(w http.ResponseWriter, r *http.Request)  {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "suggest parse form failed : %s \n", err)
		return
	}
	name,ok  := r.Form["name"]
	if !ok {
		fmt.Fprintf(w, "without name \n")
		return
	}
	modifySuggest := new(suggests.ModifySuggest)
	modifySuggest.GetModifySuggests(name[0])
	if modifySuggest.Name == ""{
		fmt.Fprintf(w, "modifySuggest without name \n")
		return
	}
	jsonResults, _ := json.Marshal(&modifySuggest)
	w.Write(jsonResults)
}

func suggestList(w http.ResponseWriter, r *http.Request)  {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Get SuggestList Method Not Allowed")
		klog.Info("Get SuggestList Method Not Allowed")
		return
	}
	modifySuggest := new(suggests.ModifySuggest)
	modifySuggestList := modifySuggest.GetAllModifySuggests()
	jsonResults, _ := json.Marshal(&modifySuggestList)
	w.Write(jsonResults)
	
}