package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

// Config holds application configuration
type Config struct {
	KubeConfigPath    string
	SlackWebhookURL   string
	ClusterName       string
	WatchedResources  []string
	WatchedNamespaces []string
	FieldsToRemove    []string

	// Queue configuration
	QueueEnabled bool
	QueueWorkers int
	QueueSize    int
}

// LoadConfig loads configuration from environment variables and flags
func LoadConfig() (*Config, error) {
	// Define kubeconfig flag
	var kubeConfigPath string

	// Check if we're running in a pod (in-cluster)
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		// We're in-cluster, don't set kubeconfig path
		kubeConfigPath = ""
		klog.Info("Detected in-cluster environment, using in-cluster config")
	} else if home := homedir.HomeDir(); home != "" {
		// We're out-of-cluster, use default kubeconfig path
		flag.StringVar(&kubeConfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&kubeConfigPath, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Get Slack webhook URL
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		klog.Exit("SLACK_WEBHOOK_URL environment variable not set..")
	}

	// Get cluster name
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "kubernetes-cluster"
		klog.Warningf("CLUSTER_NAME environment variable not set, using default: %s", clusterName)
	} else {
		klog.Infof("Watching Cluster: %s", clusterName)
	}

	// Get watched resources
	watchedResourcesEnv := os.Getenv("WATCHED_RESOURCE_NAMES")
	var watchedResources []string
	if watchedResourcesEnv == "" {
		watchedResources = []string{"pods"}
		klog.Warningf("WATCHED_RESOURCE_NAMES environment variable not set, using default: %v", watchedResources)
	} else {
		watchedResources = splitAndTrim(watchedResourcesEnv)
		klog.Infof("Watching resources: %v", watchedResources)
	}

	// Get watched namespaces
	watchedNamespacesEnv := os.Getenv("WATCHED_NAMESPACES")
	var watchedNamespaces []string
	if watchedNamespacesEnv == "" {
		watchedNamespaces = []string{"default"}
		klog.Warningf("WATCHED_NAMESPACES environment variable not set, using default: %v", watchedNamespaces)
	} else {
		watchedNamespaces = splitAndTrim(watchedNamespacesEnv)
		klog.Infof("Watching namespaces: %v", watchedNamespaces)
	}

	// Get fields to remove
	fieldsToRemoveEnv := os.Getenv("RESOURCES_IGNORED_FIELDS")
	var fieldsToRemove []string
	if fieldsToRemoveEnv == "" {
		fieldsToRemove = []string{
			"", "data.status", "status",
			"metadata.creationTimestamp", "metadata.uid", "metadata.resourceVersion",
			"metadata.selfLink", "metadata.managedFields", "metadata.generation",
			"metadata.finalizers", "metadata.annotations", "metadata.ownerReferences",
		}
		klog.Warningf("RESOURCES_IGNORED_FIELDS environment variable not set, using defaults %v ", fieldsToRemove)
	} else {
		fieldsToRemove = splitAndTrim(fieldsToRemoveEnv)
		klog.Infof("Ignoring fields: %v", fieldsToRemove)
	}

	// Queue configuration
	queueEnabled := getEnvAsBool("QUEUE_ENABLED", true) // Enabled by default
	queueWorkers := getEnvAsInt("QUEUE_WORKERS", 10)
	queueSize := getEnvAsInt("QUEUE_SIZE", 1000)

	if queueEnabled {
		klog.Infof("Queue enabled with %d workers and buffer size %d", queueWorkers, queueSize)
	} else {
		klog.Warning("Queue disabled - notifications will be sent synchronously")
	}

	return &Config{
		KubeConfigPath:    kubeConfigPath,
		SlackWebhookURL:   slackWebhookURL,
		ClusterName:       clusterName,
		WatchedResources:  watchedResources,
		WatchedNamespaces: watchedNamespaces,
		FieldsToRemove:    fieldsToRemove,
		QueueEnabled:      queueEnabled,
		QueueWorkers:      queueWorkers,
		QueueSize:         queueSize,
	}, nil
}

// splitAndTrim splits a string by comma and trims spaces from each element
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		klog.Warningf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as a boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		klog.Warningf("Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
