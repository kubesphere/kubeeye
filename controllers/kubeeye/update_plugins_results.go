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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	// "github.com/docker/docker/api/server/httputils"
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/apis/kubeeye/v1alpha1"
	kubeeyeclientset "github.com/kubesphere/kubeeye/pkg/clients/clientset/versioned"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func PluginsResultsReceiver() {
	ServeMux := http.NewServeMux()
	ServeMux.Handle("/healthz", http.HandlerFunc(health))
	ServeMux.Handle("/plugins", http.HandlerFunc(PluginsResult))
	log.Fatalln(http.ListenAndServe(":8888", ServeMux))
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed")
		log.Println("Method Not Allowed")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func PluginsResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method Not Allowed")
		log.Println("Method Not Allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Parse form failed : %s \n", err)
		log.Println("Parse form failed")
		return
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "get plugin result failed : %s \n", err)
		log.Println("get plugin result failed")
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
	log.Println(fmt.Sprintf("start update %s result to clusterInsight \n", result.Name))

	go UpdatePluginsResults(resp, result)
}

func UpdatePluginsResults(resp []byte, result kubeeyev1alpha1.PluginsResult) {
	ctx := context.TODO()

	clientSet, err := kube.GetClientSetInCluster()
	if err != nil {
		log.Println("update plugins results failed: get client set failed")
		return
	}

	clusterInsight, err := GetClusterInsights(ctx, clientSet)
	if err != nil {
		log.Println(fmt.Sprintf("update plugins results failed: get clusterInsight failed \n %s", err))
		return
	}

	if err := UpdateClusterInsights(ctx, clientSet, clusterInsight, resp, result); err != nil {
		log.Println(fmt.Sprintf("update plugins results failed: update clusterInsight failed \n %s", err))
		return
	}
}

func GetClusterInsights(ctx context.Context, clientSet *kubeeyeclientset.Clientset) (clusterInsight *kubeeyev1alpha1.ClusterInsight, err error) {
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

func UpdateClusterInsights(ctx context.Context, clientSet *kubeeyeclientset.Clientset, clusterInsight *kubeeyev1alpha1.ClusterInsight, resp []byte, result kubeeyev1alpha1.PluginsResult) error {
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

	return nil
}

func MergePluginsResults(pluginsResults []kubeeyev1alpha1.PluginsResult, newResult kubeeyev1alpha1.PluginsResult) []kubeeyev1alpha1.PluginsResult {
	var newPluginResults []kubeeyev1alpha1.PluginsResult
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
