package kubernetes

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/MIna-Maher/k8s-diff-informer/internal/queue"
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

// MetricsRecorder interface for recording informer metrics
type InformerMetricsRecorder interface {
	RecordResourceEvent(resourceType, resourceName, eventType, namespace, cluster string)
	RecordEventProcessed(resourceType, eventType, cluster string, duration time.Duration)
	SetInformerSyncStatus(resourceType, cluster string, synced bool)
	UpdateLastSyncTime(resourceType, cluster string)
	RecordDiffComputation(resourceType, cluster string, hasChanges bool, duration time.Duration)
	UpdateCacheSize(resourceType, cluster string, size int)
	RecordCacheHit(resourceType, cluster string)
	RecordCacheMiss(resourceType, cluster string)
}

// InformerManager manages multiple dynamic informers with metrics and queue
type InformerManager struct {
	dynamicClient     dynamic.Interface
	discoveryClient   *discovery.DiscoveryClient
	factory           dynamicinformer.DynamicSharedInformerFactory
	watchedResources  []string
	watchedNamespaces []string
	fieldsToRemove    []string
	taskQueue         *queue.MemoryQueue // Queue for async processing
	metricsRecorder   InformerMetricsRecorder
	clusterName       string

	// Flag to check if initial sync is complete
	isInitialSyncMu       sync.RWMutex
	isInitialSyncComplete bool

	// Event cache to store the last processed ResourceVersion of each resource
	processedEventsMu sync.RWMutex
	processedEvents   map[string]string

	// Cache size tracking
	cacheSizeMu sync.RWMutex
	cacheSizes  map[string]int
}

// NewInformerManagerWithQueue creates a new InformerManager with queue and metrics
func NewInformerManagerWithQueue(
	dynamicClient dynamic.Interface,
	discoveryClient *discovery.DiscoveryClient,
	watchedResources []string,
	watchedNamespaces []string,
	fieldsToRemove []string,
	taskQueue *queue.MemoryQueue,
	metricsRecorder InformerMetricsRecorder,
	clusterName string,
) *InformerManager {
	return &InformerManager{
		dynamicClient:         dynamicClient,
		discoveryClient:       discoveryClient,
		watchedResources:      watchedResources,
		watchedNamespaces:     watchedNamespaces,
		fieldsToRemove:        fieldsToRemove,
		taskQueue:             taskQueue,
		metricsRecorder:       metricsRecorder,
		clusterName:           clusterName,
		isInitialSyncComplete: false,
		processedEvents:       make(map[string]string),
		cacheSizes:            make(map[string]int),
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
			if im.metricsRecorder != nil {
				im.metricsRecorder.SetInformerSyncStatus(resourceName, im.clusterName, false)
			}
			return fmt.Errorf("failed to get GroupVersionResource for %s: %v", resourceName, err)
		}

		klog.Infof("Setting up informer for resource: %s.%s", resourceGV.Resource, resourceGV.Group)
		im.setupInformer(resourceGV)
	}

	// Start all informers
	im.factory.Start(stopCh)

	if err := im.waitForCacheSync(stopCh); err != nil {
		return err
	}

	// Update metrics for all resources
	if im.metricsRecorder != nil {
		for _, resourceName := range im.watchedResources {
			im.metricsRecorder.SetInformerSyncStatus(resourceName, im.clusterName, true)
			im.metricsRecorder.UpdateLastSyncTime(resourceName, im.clusterName)
		}
	}

	klog.Info("All informers are synced and ready")

	return nil
}

// HasSynced reports whether every requested informer completed its initial list.
func (im *InformerManager) HasSynced() bool {
	im.isInitialSyncMu.RLock()
	defer im.isInitialSyncMu.RUnlock()
	return im.isInitialSyncComplete
}

func (im *InformerManager) waitForCacheSync(stopCh <-chan struct{}) error {
	results := im.factory.WaitForCacheSync(stopCh)
	if len(results) == 0 {
		return fmt.Errorf("no resource informers were started")
	}
	for resource, synced := range results {
		if !synced {
			return fmt.Errorf("cache synchronization failed for %s", resource)
		}
	}
	select {
	case <-stopCh:
		return fmt.Errorf("informer startup canceled")
	default:
	}
	im.isInitialSyncMu.Lock()
	im.isInitialSyncComplete = true
	im.isInitialSyncMu.Unlock()
	return nil
}

// Shutdown waits for informers to stop after their stop channel has been closed.
func (im *InformerManager) Shutdown() {
	if im.factory != nil {
		im.factory.Shutdown()
	}
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
		start := time.Now()

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
			if im.metricsRecorder != nil {
				im.metricsRecorder.RecordCacheHit(resource.Resource, im.clusterName)
			}
			return
		}

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordCacheMiss(resource.Resource, im.clusterName)
		}

		// Check if the resource is namespaced and if we should watch it
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Update cache size
		im.updateCacheSize(resource.Resource)

		// Record metrics
		namespace := ""
		if isNamespaced {
			namespace = unstructuredObj.GetNamespace()
		}

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordResourceEvent(resource.Resource, unstructuredObj.GetName(), "add", namespace, im.clusterName)
			im.metricsRecorder.RecordEventProcessed(resource.Resource, "add", im.clusterName, time.Since(start))
		}

		// Log a brief summary of the add event to the console
		if isNamespaced {
			if slices.Contains(im.watchedNamespaces, namespace) {
				klog.Infof("ADD EVENT: Kind=%s, Name=%s, Namespace=%s",
					unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
			}
		} else {
			klog.Infof("ADD EVENT: Kind=%s, Name=%s (cluster-scoped)",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
		}

		// Check if we should process this namespace
		if isNamespaced && !im.isWatchedNamespace(namespace) {
			return
		}

		// Create notification payload
		var title, text string
		if isNamespaced {
			title = fmt.Sprintf("*Resource Added, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
				unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
			text = fmt.Sprintf("```Resource Added: %s/%s```",
				namespace, unstructuredObj.GetName())
		} else {
			title = fmt.Sprintf("*Resource Added, Kind.. `%s`, Name..:  `%s`*",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
			text = fmt.Sprintf("```Resource Added: %s```",
				unstructuredObj.GetName())
		}

		// Enqueue notification task
		task := queue.NewSlackNotificationTask(&queue.SlackNotificationPayload{
			Title:        title,
			Text:         text,
			Footer:       fmt.Sprintf("Cluster: %s", im.clusterName),
			ClusterName:  im.clusterName,
			ResourceType: resource.Resource,
			ResourceName: unstructuredObj.GetName(),
			Namespace:    namespace,
			EventType:    "add",
			Timestamp:    time.Now(),
		})

		if err := im.taskQueue.Enqueue(task); err != nil {
			klog.Errorf("Failed to enqueue notification for %s/%s: %v",
				namespace, unstructuredObj.GetName(), err)
		}
	}
}

// handleUpdateEvent returns a function that handles update events
func (im *InformerManager) handleUpdateEvent(resource schema.GroupVersionResource) func(oldObj, newObj interface{}) {
	return func(oldObj, newObj interface{}) {
		start := time.Now()

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

		// Compute differences with metrics
		diffStart := time.Now()
		diffText := diff.ComputeDiff(oldSpec, newSpec)
		diffDuration := time.Since(diffStart)
		hasChanges := diffText != ""

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordDiffComputation(resource.Resource, im.clusterName, hasChanges, diffDuration)
		}

		if diffText == "" {
			return // No significant changes
		}

		// Check if the resource is namespaced
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Update cache size
		im.updateCacheSize(resource.Resource)

		// Record metrics
		namespace := ""
		if isNamespaced {
			namespace = newResource.GetNamespace()
		}

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordResourceEvent(resource.Resource, newResource.GetName(), "update", namespace, im.clusterName)
			im.metricsRecorder.RecordEventProcessed(resource.Resource, "update", im.clusterName, time.Since(start))
		}

		// Log a brief summary of the update event to the console
		if isNamespaced {
			klog.Infof("UPDATE EVENT: Kind=%s, Name=%s, Namespace=%s",
				newResource.GetKind(), newResource.GetName(), namespace)
		} else {
			klog.Infof("UPDATE EVENT: Kind=%s, Name=%s (cluster-scoped)",
				newResource.GetKind(), newResource.GetName())
		}

		// Check if we should process this namespace
		if isNamespaced && !im.isWatchedNamespace(namespace) {
			return
		}

		// Create notification payload
		var title string
		if isNamespaced {
			title = fmt.Sprintf("*Resource Updated!, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
				newResource.GetKind(), newResource.GetName(), namespace)
		} else {
			title = fmt.Sprintf("*Resource Updated, Kind.. `%s`, Name..:  `%s`*",
				newResource.GetKind(), newResource.GetName())
		}

		// Enqueue notification task
		task := queue.NewSlackNotificationTask(&queue.SlackNotificationPayload{
			Title:        title,
			Text:         fmt.Sprintf("```%s```", diffText),
			Footer:       fmt.Sprintf("Cluster: %s", im.clusterName),
			ClusterName:  im.clusterName,
			ResourceType: resource.Resource,
			ResourceName: newResource.GetName(),
			Namespace:    namespace,
			EventType:    "update",
			Timestamp:    time.Now(),
		})

		if err := im.taskQueue.Enqueue(task); err != nil {
			klog.Errorf("Failed to enqueue notification for %s/%s: %v",
				namespace, newResource.GetName(), err)
		}
	}
}

// handleDeleteEvent returns a function that handles delete events
func (im *InformerManager) handleDeleteEvent(resource schema.GroupVersionResource) func(obj interface{}) {
	return func(obj interface{}) {
		start := time.Now()

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
			if im.metricsRecorder != nil {
				im.metricsRecorder.RecordCacheHit(resource.Resource, im.clusterName)
			}
			return
		}

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordCacheMiss(resource.Resource, im.clusterName)
		}

		// Check if the resource is namespaced and if we should watch it
		isNamespaced, err := im.isResourceNamespaced(resource.Resource)
		if err != nil {
			klog.Errorf("Error checking resource scope: %v", err)
			return
		}

		// Update cache size
		im.updateCacheSize(resource.Resource)

		// Record metrics
		namespace := ""
		if isNamespaced {
			namespace = unstructuredObj.GetNamespace()
		}

		if im.metricsRecorder != nil {
			im.metricsRecorder.RecordResourceEvent(resource.Resource, unstructuredObj.GetName(), "delete", namespace, im.clusterName)
			im.metricsRecorder.RecordEventProcessed(resource.Resource, "delete", im.clusterName, time.Since(start))
		}

		// Log a brief summary of the delete event to the console
		if isNamespaced {
			klog.Infof("DELETE EVENT: Kind=%s, Name=%s, Namespace=%s",
				unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
		} else {
			klog.Infof("DELETE EVENT: Kind=%s, Name=%s (cluster-scoped)",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
		}

		// Check if we should process this namespace
		if isNamespaced && !im.isWatchedNamespace(namespace) {
			return
		}

		// Create notification payload
		var title, text string
		if isNamespaced {
			title = fmt.Sprintf("*Resource Deleted, Kind.. `%s`, Name..:  `%s`, Namespace:..`%s`*",
				unstructuredObj.GetKind(), unstructuredObj.GetName(), namespace)
			text = fmt.Sprintf("```Resource Deleted: %s/%s```",
				namespace, unstructuredObj.GetName())
		} else {
			title = fmt.Sprintf("*Resource Deleted, Kind.. `%s`, Name..:  `%s`*",
				unstructuredObj.GetKind(), unstructuredObj.GetName())
			text = fmt.Sprintf("```Resource Deleted: %s```",
				unstructuredObj.GetName())
		}

		// Enqueue notification task
		task := queue.NewSlackNotificationTask(&queue.SlackNotificationPayload{
			Title:        title,
			Text:         text,
			Footer:       fmt.Sprintf("Cluster: %s", im.clusterName),
			ClusterName:  im.clusterName,
			ResourceType: resource.Resource,
			ResourceName: unstructuredObj.GetName(),
			Namespace:    namespace,
			EventType:    "delete",
			Timestamp:    time.Now(),
		})

		if err := im.taskQueue.Enqueue(task); err != nil {
			klog.Errorf("Failed to enqueue notification for %s/%s: %v",
				namespace, unstructuredObj.GetName(), err)
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

// updateCacheSize updates the cache size for a resource type
func (im *InformerManager) updateCacheSize(resourceType string) {
	im.cacheSizeMu.Lock()
	im.cacheSizes[resourceType]++
	currentSize := im.cacheSizes[resourceType]
	im.cacheSizeMu.Unlock()

	if im.metricsRecorder != nil {
		im.metricsRecorder.UpdateCacheSize(resourceType, im.clusterName, currentSize)
	}
}
