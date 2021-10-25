package expend

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kubesphere/kubeeye/pkg/kube"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/restmapper"
)

func CreateResource(path string, ctx context.Context, resource []byte) (err error) {
	kubernetesClient := kube.KubernetesAPI(path)
	clientset := kubernetesClient.ClientSet
	dynamicClient := kubernetesClient.DynamicClient
	namespace := metav1.NamespaceDefault

	// decode resource for convert the resource to unstructur.
	newreader := bytes.NewReader(resource)
	decode := yaml.NewYAMLOrJSONDecoder(newreader, 4096)

	// get resource kind and group
	disc := clientset.Discovery()
	restMapperRes, _ := restmapper.GetAPIGroupResources(disc)
	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)
	ext := runtime.RawExtension{}
	if err := decode.Decode(&ext); err != nil {
		err = fmt.Errorf("Decode error: %s\n", err.Error())
		return err
	}
	// get resource.Object
	obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
	if err != nil {
		err = fmt.Errorf("failed to get resource object: %s\n", err.Error())
		return err
	}
	// identifies a preferred resource mapping
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		err = fmt.Errorf("failed to get resource mapping: %s\n", err.Error())
		return err
	}

	// convert the resource.Object into unstructured
	var unstruct unstructured.Unstructured
	unstruct.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		err = fmt.Errorf("failed to converts an object into unstructured representation: %s\n", err.Error())
		return err
	}
	// get namespace from resource.Object
	namespace = unstruct.GetNamespace()
	// create resource by dynamic client
	result, err := dynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(ctx, &unstruct, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			err = fmt.Errorf("Create resource failed, resource is already exists: %s \n", err.Error())
			return err
		}
	} else if errors.IsInvalid(err) {
		err = fmt.Errorf("Create resource failed, resource is invalid: %s \n", err.Error())
		return err
	}
	fmt.Printf("%s \t %s \t %s \t created\n", result.GetNamespace(), result.GetKind(), result.GetName())

	return nil
}

func RemoveResource(path string, ctx context.Context, resource []byte) (err error) {
	kubernetesClient := kube.KubernetesAPI(path)
	clientset := kubernetesClient.ClientSet
	dynamicClient := kubernetesClient.DynamicClient
	namespace := metav1.NamespaceDefault

	// decode resource for convert the resource to unstructur.
	newreader := bytes.NewReader(resource)
	decode := yaml.NewYAMLOrJSONDecoder(newreader, 4096)

	// get resource kind and group
	disc := clientset.Discovery()
	restMapperRes, _ := restmapper.GetAPIGroupResources(disc)
	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes)
	ext := runtime.RawExtension{}
	if err := decode.Decode(&ext); err != nil {
		err = fmt.Errorf("Decode error: %s\n", err.Error())
		return err
	}
	// get resource.Object
	obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
	if err != nil {
		err = fmt.Errorf("failed to get resource object: %s\n", err.Error())
		return err
	}
	// identifies a preferred resource mapping
	mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		err = fmt.Errorf("failed to get resource mapping: %s\n", err.Error())
		return err
	}
	// convert the resource.Object into unstructured
	var unstruct unstructured.Unstructured
	unstruct.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		err = fmt.Errorf("failed to converts an object into unstructured representation: %s\n", err.Error())
		return err
	}
	// get resource kind name and namespace
	kind := unstruct.GetKind()
	name := unstruct.GetName()
	namespace = unstruct.GetNamespace()

	// delete resource by dynamic client
	err = dynamicClient.Resource(mapping.Resource).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		err = fmt.Errorf("failed to remove resource %s in %s: %s\n", name, namespace, err.Error())
		return err
	}
	fmt.Printf("%s \t %s \t %s \t deleted\n", namespace, kind, name)
	return nil
}
