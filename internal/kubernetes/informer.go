package kubernetes

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"github.com/MIna-Maher/k8s-diff-informer/pkg/diff"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// InformerManager manages multiple dynamic informers
type InformerManager struct {
	dynamicClient     dynamic.Interface
	discoveryClient   *discovery.DiscoveryClient
	factory           dynamicinformer.DynamicSharedInformerFactory
	watchedResources  []string
	watchedNamespaces []string
	fieldsToRemove    []string
	slackClient       *slack.Client

	// Flag to check if initial sync is complete
	isInitialSyncMu       sync.RWMutex
	isInitialSyncComplete bool

	// Event cache to store the last processed ResourceVersion of each resource
	processedEventsMu sync.RWMutex
	processedEvents   map[string]string
}

// NewInformerManager creates a new InformerManager
func NewInformerManager(
	dynamicClient dynamic.Interface,
	discoveryClient *discovery.DiscoveryClient,
	watchedResources []string,
	watchedNamespaces []string,
	fieldsToRemove []string,
	slackClient *slack.Client,
) *InformerManager {
	return &InformerManager{
		dynamicClient:         dynamicClient,
		discoveryClient:       discoveryClient,
		watchedResources:      watchedResources,
		watchedNamespaces:     watchedNamespaces,
		fieldsToRemove:        fieldsToRemove,
		slackClient:           slackClient,
		isInitialSyncComplete: false,
		processedEvents:       make(map[string]string),
	}
}

// Start starts all informers
func (im *InformerManager) Start(stopCh <-chan struct{}) error {
	// Create dynamic informer factory
	im.factory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		im.dynamicClient,
		time.Duration(0), // No resync
		v1.NamespaceAll,
		nil,
	)

	// Setup informers for all watched resources
	for _, resourceName := range im.watchedResources {
		resourceGV, err := im.getResourceGVR(resourceName)
		if err != nil {
			return fmt.Errorf("failed to get GroupVersionResource for %s: %v", resourceName, err)
		}

		klog.Infof("Setting up informer for resource: %s.%s", resourceGV.Resource, resourceGV.Group)
		im.setupInformer(resourceGV)
	}

	// Start all informers
	im.factory.Start(stopCh)

	// Wait for the initial sync to complete
	im.factory.WaitForCacheSync(stopCh)

	im.isInitialSyncMu.Lock()
	im.isInitialSyncComplete = true
	im.isInitialSyncMu.Unlock()

	klog.Info("All informers are synced and ready")

	return nil
}

// setupInformer configures an informer for a specific resource
func (im *InformerManager) setupInformer(resource schema.GroupVersionResource) {
	informer := im.factory.ForResource(resource).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.handleAddEvent(resource),
		UpdateFunc: im.handleUpdateEvent(resource),
		DeleteFunc: im.handleDeleteEvent(resource),
	})
}

// handleAddEvent returns a function that handles add events
func (im *InformerManager) handleAddEvent(resource schema.GroupVersionResource) func(obj interface{}) {
	return func(obj interface{}) {
		im.isInitialSyncMu.RLock()
		isComplete := im.isInitialSyncComplete
		im.isInitialSyncMu.RUnlock()

		if !isComplete {
			// Skip events during initial sync
			return
		}

		// Cast the object to an unstructured resource
		unstructuredObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			klog.Errorf("Failed to cast object to *unstructured.Unstructured")
			return
		}

		// Check if this is a new event
		if !im.isNewEvent(unstructuredObj) {
			return
		}

		// Check if the resource is namespaced and if we should watch it
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Log a brief summary of the add event to the console
		if isNamespaced {
			namespace := unstructuredObj.GetNamespace()
			if slices.Contains(im.watchedNamespaces, namespace) {
				klog.Infof("ADD EVENT: Kind=%s, Name=%s, Namespace=%s",
					unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
			}
		} else {
			klog.Infof("ADD EVENT: Kind=%s, Name=%s (cluster-scoped)",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
		}

		var message *slack.Message

		if isNamespaced {
			namespace := unstructuredObj.GetNamespace()
			if !im.isWatchedNamespace(namespace) {
				return
			}

			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Added, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
					unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace),
				Text: fmt.Sprintf("```Resource Added: %s/%s```",
					namespace, unstructuredObj.GetName()),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		} else {
			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Added, Kind.. `%s`, Name..:  `%s`*",
					unstructuredObj.GetKind(), unstructuredObj.GetName()),
				Text: fmt.Sprintf("```Resource Added: %s```",
					unstructuredObj.GetName()),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		}

		if err := im.slackClient.SendMessage(message); err != nil {
			klog.Errorf("Failed to send message to Slack: %v", err)
		}
	}
}

// handleUpdateEvent returns a function that handles update events
func (im *InformerManager) handleUpdateEvent(resource schema.GroupVersionResource) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		im.isInitialSyncMu.RLock()
		isComplete := im.isInitialSyncComplete
		im.isInitialSyncMu.RUnlock()

		if !isComplete {
			return
		}

		// Cast the objects to unstructured resources
		oldResource, okOld := oldObj.(*unstructured.Unstructured)
		newResource, okNew := newObj.(*unstructured.Unstructured)
		if !okOld || !okNew {
			klog.Error("Failed to cast objects to *unstructured.Unstructured")
			return
		}

		// Extract specs and remove ignored fields
		oldSpec, _, _ := unstructured.NestedMap(oldResource.Object)
		newSpec, _, _ := unstructured.NestedMap(newResource.Object)

		diff.RemoveFields(oldSpec, im.fieldsToRemove)
		diff.RemoveFields(newSpec, im.fieldsToRemove)

		// Compute differences
		diffText := diff.ComputeDiff(oldSpec, newSpec)
		if diffText == "" {
			return // No significant changes
		}

		// Check if the resource is namespaced
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Log a brief summary of the update event to the console
		if isNamespaced {
			namespace := newResource.GetNamespace()
			klog.Infof("UPDATE EVENT: Kind=%s, Name=%s, Namespace=%s",
				newResource.GetKind(), newResource.GetName(), namespace)
		} else {
			klog.Infof("UPDATE EVENT: Kind=%s, Name=%s (cluster-scoped)",
				newResource.GetKind(), newResource.GetName())
		}

		var message *slack.Message

		if isNamespaced {
			namespace := oldResource.GetNamespace()
			if !im.isWatchedNamespace(namespace) {
				return
			}

			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Updated!, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
					newResource.GetKind(), newResource.GetName(), namespace),
				Text:   fmt.Sprintf("```%s```", diffText),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		} else {
			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Updated, Kind.. `%s`, Name..:  `%s`*",
					newResource.GetKind(), newResource.GetName()),
				Text:   fmt.Sprintf("```%s```", diffText),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		}

		if err := im.slackClient.SendMessage(message); err != nil {
			klog.Errorf("Failed to send message to Slack: %v", err)
		}
	}
}

// handleDeleteEvent returns a function that handles delete events
func (im *InformerManager) handleDeleteEvent(resource schema.GroupVersionResource) func(obj interface{}) {
	return func(obj interface{}) {
		im.isInitialSyncMu.RLock()
		isComplete := im.isInitialSyncComplete
		im.isInitialSyncMu.RUnlock()

		if !isComplete {
			return
		}

		// Cast the object to an unstructured resource
		unstructuredObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			klog.Error("Failed to cast object to *unstructured.Unstructured")
			return
		}

		// Check if this is a new event
		if !im.isNewEvent(unstructuredObj) {
			return
		}

		// Check if the resource is namespaced and if we should watch it
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Log a brief summary of the delete event to the console
		if isNamespaced {
			namespace := unstructuredObj.GetNamespace()
			klog.Infof("DELETE EVENT: Kind=%s, Name=%s, Namespace=%s",
				unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
		} else {
			klog.Infof("DELETE EVENT: Kind=%s, Name=%s (cluster-scoped)",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
		}

		var message *slack.Message

		if isNamespaced {
			namespace := unstructuredObj.GetNamespace()
			if !im.isWatchedNamespace(namespace) {
				return
			}

			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Deleted, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
					unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace),
				Text: fmt.Sprintf("```Resource Deleted: %s/%s```",
					namespace, unstructuredObj.GetName()),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		} else {
			message = &slack.Message{
				Title: fmt.Sprintf("*Resource Deleted, Kind.. `%s`, Name..:  `%s`*",
					unstructuredObj.GetKind(), unstructuredObj.GetName()),
				Text: fmt.Sprintf("```Resource Deleted: %s```",
					unstructuredObj.GetName()),
				Footer: fmt.Sprintf("Cluster: %s", im.slackClient.ClusterName),
			}
		}

		if err := im.slackClient.SendMessage(message); err != nil {
			klog.Errorf("Failed to send message to Slack: %v", err)
		}
	}
}

// isNewEvent checks if an event is new by comparing the ResourceVersion
func (im *InformerManager) isNewEvent(resource *unstructured.Unstructured) bool {
	// Construct a unique identifier for the resource
	eventID := fmt.Sprintf("%s/%s", resource.GetNamespace(), resource.GetName())

	// Get current ResourceVersion
	currentResourceVersion := resource.GetResourceVersion()
	if currentResourceVersion == "" {
		klog.V(4).Infof("ResourceVersion not found for resource %s", eventID)
		return false
	}

	im.processedEventsMu.Lock()
	defer im.processedEventsMu.Unlock()

	// Check if event ID exists in cache and if ResourceVersion has changed
	lastSeenVersion, exists := im.processedEvents[eventID]
	if !exists || lastSeenVersion != currentResourceVersion {
		// Update cache with new ResourceVersion
		im.processedEvents[eventID] = currentResourceVersion
		return true
	}

	return false
}

// isWatchedNamespace checks if a namespace is in the list of watched namespaces
func (im *InformerManager) isWatchedNamespace(namespace string) bool {
	for _, ns := range im.watchedNamespaces {
		if ns == namespace {
			return true
		}
	}
	return false
}
