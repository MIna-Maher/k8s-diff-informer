package config

import (
	"flag"
	"os"
	"path/filepath"
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
}

// LoadConfig loads configuration from environment variables and flags
func LoadConfig() (*Config, error) {
	// Define kubeconfig flag
	var kubeConfigPath string
	if home := homedir.HomeDir(); home != "" {
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
		klog.Infof(" Watching Cluster: %s", clusterName)
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

	return &Config{
		KubeConfigPath:    kubeConfigPath,
		SlackWebhookURL:   slackWebhookURL,
		ClusterName:       clusterName,
		WatchedResources:  watchedResources,
		WatchedNamespaces: watchedNamespaces,
		FieldsToRemove:    fieldsToRemove,
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
