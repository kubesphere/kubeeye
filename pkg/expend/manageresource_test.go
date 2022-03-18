package expend

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/kubesphere/kubeeye/testdata"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
)

func TestCreateResource(t *testing.T) {
	tcs := []struct {
		resource    string
		name        string
		obj         *unstructured.Unstructured
		path        string
	}{
		{
			resource: "rtest",
			name:     "normal_create",
			path:     "/apis/gtest/vtest/namespaces/tns/rtest",
			obj:      getObject("gtest/vTest", "rTest", "normal_create", "tns"),
		},
		{
			resource:  "rtest",
			name:      "namespaced_create",
			path:      "/apis/gtest/vtest/namespaces/tns/rtest",
			obj:       getObject("gtest/vTest", "rTest", "namespaced_create", "tns"),
		},
		{
			resource:    "rtest",
			name:        "normal_subresource_create",
			path:        "/apis/gtest/vtest/rtest",
			obj:         getObject("vTest", "srTest", "normal_subresource_create", ""),
		},
		{
			resource:    "rtest/",
			name:        "namespaced_subresource_create",
			path:        "/apis/gtest/vtest/namespaces/tns/rtest",
			obj:         getObject("vTest", "srTest", "namespaced_subresource_create", "tns"),
		},
	}

	for _, tc := range tcs {
		resource := schema.GroupVersionResource{Group: "gtest", Version: "vtest", Resource: tc.resource}
		cl, srv, err := getClientServer(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Create(%q) got HTTP method %s. wanted POST", tc.name, r.Method)
			}

			if r.URL.Path != tc.path {
				t.Errorf("Create(%q) got path %s. wanted %s", tc.name, r.URL.Path, tc.path)
			}

			content := r.Header.Get("Content-Type")
			if content != runtime.ContentTypeJSON {
				t.Errorf("Create(%q) got Content-Type %s. wanted %s", tc.name, content, runtime.ContentTypeJSON)
			}

			w.Header().Set("Content-Type", runtime.ContentTypeJSON)
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Errorf("Create(%q) unexpected error reading body: %v", tc.name, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Write(data)
		})
		if err != nil {
			t.Errorf("unexpected error when creating client: %v", err)
			continue
		}
		defer srv.Close()

		var mapping meta.RESTMapping
		mapping.Resource = resource

		err = CreateResource(context.TODO(), cl, &mapping, tc.obj)
		if err != nil {
			t.Fatalf("failed to Create Resource")
		}
	}
}

func TestRemoveResource(t *testing.T) {
	background := metav1.DeletePropagationBackground
	uid := types.UID("uid")

	statusOK := &metav1.Status{
		TypeMeta: metav1.TypeMeta{Kind: "Status"},
		Status:   metav1.StatusSuccess,
	}
	tcs := []struct {
		name          string
		path          string
		deleteOptions metav1.DeleteOptions
		obj         *unstructured.Unstructured
	}{
		{
			name: "normal_delete",
			path: "/apis/gtest/vtest/namespaces/nstest/rtest/normal_delete",
			obj:      getObject("gtest/vTest", "rTest", "normal_delete","nstest"),
		},
		{
			name:      "namespaced_delete",
			path:      "/apis/gtest/vtest/namespaces/nstest/rtest/namespaced_delete",
			obj:       getObject("gtest/vTest", "rTest", "namespaced_delete", "nstest"),
		},
		{
			name:          "namespaced_delete_with_options",
			path:          "/apis/gtest/vtest/namespaces/nstest/rtest/namespaced_delete_with_options",
			deleteOptions: metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}, PropagationPolicy: &background},
			obj:         getObject("gtest/vTest", "srTest", "namespaced_delete_with_options", "nstest"),
		},
	}

	for _, tc := range tcs {
		resource := schema.GroupVersionResource{Group: "gtest", Version: "vtest", Resource: "rtest"}
		cl, srv, err := getClientServer(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("Delete(%q) got HTTP method %s. wanted DELETE", tc.name, r.Method)
			}

			if r.URL.Path != tc.path {
				t.Errorf("Delete(%q) got path %s. wanted %s", tc.name, r.URL.Path, tc.path)
			}

			content := r.Header.Get("Content-Type")
			if content != runtime.ContentTypeJSON {
				t.Errorf("Delete(%q) got Content-Type %s. wanted %s", tc.name, content, runtime.ContentTypeJSON)
			}

			w.Header().Set("Content-Type", runtime.ContentTypeJSON)
			unstructured.UnstructuredJSONScheme.Encode(statusOK, w)
		})
		if err != nil {
			t.Errorf("unexpected error when creating client: %v", err)
			continue
		}
		defer srv.Close()

		var mapping meta.RESTMapping
		mapping.Resource = resource
		if err := RemoveResource(context.TODO(), cl, &mapping, tc.obj);err != nil {
			t.Fatalf("failed to Delete Resource")
		}
	}

}

func TestParseResources(t *testing.T) {
	tcs := []struct{
		name string
		resource string
		expected *unstructured.Unstructured
	}{
		{
			name: "Deployment",
			resource: testdata.DeploymentSource,
			expected: testdata.UnstructuredResource,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			kubeConfig, err := kube.GetKubeConfig("")
			if err != nil {
				t.Fatalf("failed to load config file")
			}
			var kc kube.KubernetesClient
			clients, err := kc.K8SClients(kubeConfig)
			if err != nil {
				t.Fatalf("failed to get kubernetes clients")
			}
			clientSet := clients.ClientSet
			_, unstructuredResource, err := ParseResources(clientSet,[]byte(tc.resource))
			if !reflect.DeepEqual(unstructuredResource, tc.expected) {
				t.Errorf("unexpected result:\nexpected = %s\ngot = %s", spew.Sdump(tc.expected), spew.Sdump(unstructuredResource))
			}
		})
	}
}



func getObject(version, kind, name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": version,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name,
				"namespace": namespace,
			},
		},
	}
}

func getClientServer(h func(http.ResponseWriter, *http.Request)) (dynamic.Interface, *httptest.Server, error) {
	srv := httptest.NewServer(http.HandlerFunc(h))
	cl, err := dynamic.NewForConfig(&restclient.Config{
		Host: srv.URL,
	})
	if err != nil {
		srv.Close()
		return nil, nil, err
	}
	return cl, srv, nil
}