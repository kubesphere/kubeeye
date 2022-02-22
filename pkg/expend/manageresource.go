package expend

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

func CreateResource(path string, ctx context.Context, resource []byte) (err error) {
	kubeConfig, err := kube.GetKubeConfig(path)
	if err != nil {
		return errors.Wrap(err, "failed to load config file")
	}

	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		return err
	}

	dynamicClient := clients.DynamicClient
	clientset := clients.ClientSet

	// Parse Resources,get the unstructured resource
	mapping, unstructuredResource, err := ParseResources(clientset, resource)
	if err != nil {
		return err
	}

	// get namespace from resource.Object
	namespace := unstructuredResource.GetNamespace()
	// create unstructured resource by dynamic client
	result, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(ctx, &unstructuredResource, metav1.CreateOptions{})
	if err != nil {
		if kubeErr.IsAlreadyExists(err) {
			return errors.Wrap(err, "Create resource failed, resource is already exists")
		}
	} else if kubeErr.IsInvalid(err) {
		return errors.Wrap(err, "Create resource failed, resource is invalid")
	}
	fmt.Printf("%s \t %s \t %s \t created\n", result.GetNamespace(), result.GetKind(), result.GetName())

	return nil
}

func RemoveResource(path string, ctx context.Context, resource []byte) (err error) {
	kubeConfig, err := kube.GetKubeConfig(path)
	if err != nil {
		return errors.Wrap(err, "failed to load config file")
	}

	var kc kube.KubernetesClient
	clients, err := kc.K8SClients(kubeConfig)
	if err != nil {
		return err
	}

	clientset := clients.ClientSet
	dynamicClient := clients.DynamicClient

	mapping, unstructuredResource, err := ParseResources(clientset, resource)
	if err != nil {
		return err
	}

	name := unstructuredResource.GetName()
	namespace := unstructuredResource.GetNamespace()

	// delete resource by dynamic client
	err = dynamicClient.Resource(mapping.Resource).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to remove resource")
	}
	fmt.Printf("%s \t %s \t %s \t deleted\n", namespace, unstructuredResource.GetKind(), name)
	return nil
}

// ParseResources by parsing the resource, put the result into unstructuredResource and return it.
func ParseResources(clientset kubernetes.Interface, resource []byte) (mapping *meta.RESTMapping, unstructuredResource unstructured.Unstructured, err error) {

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
		return nil, unstructuredResource, errors.Wrap(err, "decode error")
	}
	// get resource.Object
	obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
	if err != nil {
		return nil, unstructuredResource, errors.Wrap(err, "failed to get resource object")
	}
	// identifies a preferred resource mapping
	mapping, err = restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, unstructuredResource, errors.Wrap(err, "failed to get resource mapping")
	}

	// convert the resource.Object into unstructured

	unstructuredResource.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, unstructuredResource, errors.Wrap(err, "failed to converts an object into unstructured representation")
	}
	return mapping, unstructuredResource, nil
}
