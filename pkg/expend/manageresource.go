package expend

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/kube"
	kubeeyev1alpha1 "github.com/kubesphere/kubeeye/plugins/plugin-manage/api/v1alpha1"
	pluginPkg "github.com/kubesphere/kubeeye/plugins/plugin-manage/pkg"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"strings"
)

func ResourceCreater(clients *kube.KubernetesClient, ctx context.Context, resource []byte) (err error) {
	dynamicClient := clients.DynamicClient
	clientset := clients.ClientSet

	// Parse Resources,get the unstructured resource
	mapping, unstructuredResource, err := ParseResources(clientset, resource)
	if err != nil {
		return err
	}

	// create resource
	if err := CreateResource(ctx, dynamicClient, mapping, unstructuredResource); err != nil {
		return err
	}

	return nil
}

func CreateResource(ctx context.Context, dynamicClient dynamic.Interface, mapping *meta.RESTMapping, unstructuredResource *unstructured.Unstructured, ) error {
	// get namespace from resource.Object
	namespace := unstructuredResource.GetNamespace()
	result, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(ctx, unstructuredResource, metav1.CreateOptions{})
	if err != nil {
		if kubeErr.IsAlreadyExists(err) {
			return nil
		} else if kubeErr.IsInvalid(err) {
			return errors.Wrap(err, "Create resource failed, resource is invalid")
		}
	}

	fmt.Printf("%s\t%s\t created\n", result.GetKind(), result.GetName())
	return nil
}

func ResourceRemover(clients *kube.KubernetesClient, ctx context.Context, resource []byte) (err error) {
	clientset := clients.ClientSet
	dynamicClient := clients.DynamicClient

	mapping, unstructuredResource, err := ParseResources(clientset, resource)
	if err != nil {
		return err
	}

	if err := RemoveResource(ctx, dynamicClient, mapping, unstructuredResource); err != nil {
		return err
	}

	return nil
}

func RemoveResource(ctx context.Context, dynamicClient dynamic.Interface, mapping *meta.RESTMapping, unstructuredResource *unstructured.Unstructured) error {
	name := unstructuredResource.GetName()
	namespace := unstructuredResource.GetNamespace()

	// delete resource by dynamic client
	if err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if kubeErr.IsNotFound(err) {
			return nil
		} else {
			return errors.Wrap(err, "failed to remove resource")
		}
	}
	fmt.Printf("%s\t%s\t deleted\n", unstructuredResource.GetKind(), name)
	return nil
}

// ParseResources by parsing the resource, put the result into unstructuredResource and return it.
func ParseResources(clientset kubernetes.Interface, resource []byte) (*meta.RESTMapping, *unstructured.Unstructured, error) {
	var unstructuredResource unstructured.Unstructured
	r := dedent.Dedent(string(resource))
	// decode resource for convert the resource to unstructur.
	newreader := strings.NewReader(r)
	decode := yaml.NewYAMLOrJSONDecoder(newreader, 4096)
	// get resource kind and group
	disc := clientset.Discovery()
	restMapperRes, _ := restmapper.GetAPIGroupResources(disc)
	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)
	ext := runtime.RawExtension{}
	if err := decode.Decode(&ext); err != nil {
		return nil, &unstructuredResource, errors.Wrap(err, "decode error")
	}
	// get resource.Object
	obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
	if err != nil {
		return nil, &unstructuredResource, errors.Wrap(err, "failed to get resource object")
	}
	// identifies a preferred resource mapping
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, &unstructuredResource, errors.Wrap(err, "failed to get resource mapping")
	}

	// convert the resource.Object into unstructured

	unstructuredResource.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)

	if err != nil {
		return nil, &unstructuredResource, errors.Wrap(err, "failed to converts an object into unstructured representation")
	}
	return mapping, &unstructuredResource, nil
}

//ListCRDResources, get plugin list
func ListCRDResources(ctx context.Context, client dynamic.Interface, namespace string) ([]string, error) {
	var gvr = schema.GroupVersionResource{
		Group:    pluginPkg.Group,
		Version:  pluginPkg.Version,
		Resource: pluginPkg.Resource,
	}
	list, err := client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	data, err := list.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var pluginList kubeeyev1alpha1.PluginSubscriptionList
	if err := json.Unmarshal(data, &pluginList); err != nil {
		return nil, err
	}
	plugins := make([]string, 0)
	for _, t := range pluginList.Items {
		if t.Status.Enabled && t.Status.Install == pluginPkg.PluginIntalled {
			plugins = append(plugins, t.Name)
		}
	}

	return plugins, nil
}
