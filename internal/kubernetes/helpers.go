package kubernetes

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// getResourceGVR returns the GroupVersionResource for a given resource name
func (im *InformerManager) getResourceGVR(resourceName string) (schema.GroupVersionResource, error) {
	apiResourceLists, err := im.discoveryClient.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to fetch API resources: %v", err)
	}

	for _, apiResourceList := range apiResourceLists {
		groupVersion := apiResourceList.GroupVersion
		gv, err := schema.ParseGroupVersion(groupVersion)
		if err != nil {
			klog.V(4).Infof("Error parsing GroupVersion %s: %v", groupVersion, err)
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Name == resourceName {
				klog.V(4).Infof("Found resource %s in group %s, version %s",
					resourceName, gv.Group, gv.Version)

				return schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: apiResource.Name,
				}, nil
			}
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("resource %s not found", resourceName)
}

// isResourceNamespaced determines if a resource is namespaced or cluster-scoped
func (im *InformerManager) isResourceNamespaced(resourceName string) (bool, error) {
	apiResourceLists, err := im.discoveryClient.ServerPreferredResources()
	if err != nil {
		return false, fmt.Errorf("failed to fetch API resources: %v", err)
	}

	for _, apiResourceList := range apiResourceLists {
		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Name == resourceName {
				return apiResource.Namespaced, nil
			}
		}
	}

	return false, fmt.Errorf("resource %s not found", resourceName)
}
