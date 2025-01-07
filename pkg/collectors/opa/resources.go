package opa

import (
	"context"
	"fmt"
	"github.com/kubesphere/kubeeye/pkg/constant"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
	statsApi "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
	"strings"
)

type ResourceCollector struct {
	client    kubernetes.Interface
	dynamic   dynamic.Interface
	discovery discovery.DiscoveryInterface
}

// NewResourceCollector creates a new ResourceCollector
func NewResourceCollector(config *rest.Config) (*ResourceCollector, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	return &ResourceCollector{
		client:    clientset,
		dynamic:   dynamicClient,
		discovery: clientset.Discovery(),
	}, nil
}

// ResourcesManager manages resources
type ResourcesManager struct {
	Resources    map[string][]unstructured.Unstructured
	StatsSummary map[string][]statsApi.Summary
}

// NewResourcesManager creates a new ResourcesManager
func NewResourcesManager() *ResourcesManager {
	return &ResourcesManager{
		Resources:    make(map[string][]unstructured.Unstructured),
		StatsSummary: make(map[string][]statsApi.Summary),
	}
}

// AddResource adds a resource to the manager
func (rm *ResourcesManager) AddResource(resource string, collector *ResourceCollector) error {
	// parse resource
	parts := strings.SplitN(resource, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid resource format")
	}

	kind := parts[0]
	version := parts[1]

	//var resources []unstructured.Unstructured
	//var err error

	if kind == constant.NodeStatsSummary {
		// collect node stats summary
		statsSummaryResults, err := collector.CollectNodeStatsSummary()
		if err != nil {
			return fmt.Errorf("failed to collect node stats summary: %v", err)
		}
		rm.StatsSummary[resource] = statsSummaryResults

		klog.Infof("resource: %s, count: %d", resource, len(statsSummaryResults))
	} else {
		// collect resources
		resources, err := collector.CollectResources(kind, version)
		if err != nil {
			return fmt.Errorf("failed to collect resources: %v", err)
		}

		// add resources to manager
		rm.Resources[resource] = resources

		klog.Infof("resource: %s, count: %d", resource, len(resources))
	}

	return nil
}

// CollectResources collects resources of a given kind and version
func (rc *ResourceCollector) CollectResources(kind, version string) ([]unstructured.Unstructured, error) {
	// create REST mapper
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(rc.discovery))

	// parse group from version
	var group string
	if strings.Contains(version, "/") {
		parts := strings.Split(version, "/")
		group = parts[0]
		version = parts[1]
	}

	// get REST mapping
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: group, Kind: kind}, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %v", err)
	}

	// list resources
	var items []unstructured.Unstructured
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// get namespaces
		namespaces, err := rc.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list namespaces: %v", err)
		}

		// list resources in each namespace
		for _, ns := range namespaces.Items {
			list, err := rc.dynamic.Resource(mapping.Resource).Namespace(ns.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				// skip if resource is not found in namespace
				klog.Errorf("failed to list resources %v in namespace %s: %v", mapping.Resource, ns.Name, err)
				continue
			}
			items = append(items, list.Items...)
		}
	} else {
		// list cluster-scoped resources
		list, err := rc.dynamic.Resource(mapping.Resource).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list cluster-scoped resources: %v", err)
		}
		items = list.Items
	}

	return items, nil
}

// collect multiple resources
func (rc *ResourceCollector) CollectMultipleResources(resources []struct {
	Kind    string
	Version string
}) (map[string][]unstructured.Unstructured, error) {
	result := make(map[string][]unstructured.Unstructured)

	for _, res := range resources {
		key := fmt.Sprintf("%s.%s", res.Kind, res.Version)
		items, err := rc.CollectResources(res.Kind, res.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to collect %s: %v", key, err)
		}
		result[key] = items
	}

	return result, nil
}

// collect resources with filter
type ResourceFilter struct {
	NamespaceSelector []string // filter by namespace
	LabelSelector     string   // filter by label
	FieldSelector     string   // filter by field
	Names             []string // filter by resource name
	ExcludeNames      []string // exclude resource name
}

func (rc *ResourceCollector) CollectResourcesWithFilter(kind, version string, filter ResourceFilter) ([]unstructured.Unstructured, error) {
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(rc.discovery))

	var group string
	if strings.Contains(version, "/") {
		parts := strings.Split(version, "/")
		group = parts[0]
		version = parts[1]
	}

	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: group, Kind: kind}, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %v", err)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: filter.LabelSelector,
		FieldSelector: filter.FieldSelector,
	}

	var items []unstructured.Unstructured

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		var namespaces []string

		if len(filter.NamespaceSelector) > 0 {
			namespaces = filter.NamespaceSelector
		} else {
			// get all namespaces
			nsList, err := rc.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list namespaces: %v", err)
			}
			for _, ns := range nsList.Items {
				namespaces = append(namespaces, ns.Name)
			}
		}

		// list resources in each namespace
		for _, ns := range namespaces {
			list, err := rc.dynamic.Resource(mapping.Resource).Namespace(ns).List(context.TODO(), listOptions)
			if err != nil {
				klog.Errorf("failed to list resources %v in namespace %s: %v", mapping.Resource, ns, err)
				continue
			}
			items = append(items, list.Items...)
		}
	} else {
		list, err := rc.dynamic.Resource(mapping.Resource).List(context.TODO(), listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list cluster-scoped resources: %v", err)
		}
		items = list.Items
	}

	// filter by resource name
	if len(filter.Names) > 0 || len(filter.ExcludeNames) > 0 {
		filteredItems := make([]unstructured.Unstructured, 0)
		for _, item := range items {
			name := item.GetName()

			// check if the resource name is in the exclude list
			excluded := false
			for _, excludeName := range filter.ExcludeNames {
				if name == excludeName {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			// check if the resource name is in the include list
			if len(filter.Names) > 0 {
				included := false
				for _, includeName := range filter.Names {
					if name == includeName {
						included = true
						break
					}
				}
				if !included {
					continue
				}
			}

			filteredItems = append(filteredItems, item)
		}
		items = filteredItems
	}

	return items, nil
}
