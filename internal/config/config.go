package config

import (
	"flag"
	"fmt"
	"net/url"
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
	HTTPPort     int
}

// LoadConfig loads configuration from environment variables and flags
func LoadConfig() (*Config, error) {
	return loadConfig(os.Args[1:])
}

func loadConfig(args []string) (*Config, error) {
	var kubeConfigPath string
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err != nil {
		if home := homedir.HomeDir(); home != "" {
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}
	flags := flag.NewFlagSet("k8s-diff-informer", flag.ContinueOnError)
	flags.StringVar(&kubeConfigPath, "kubeconfig", kubeConfigPath, "path to kubeconfig (empty uses in-cluster credentials)")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, fmt.Errorf("unexpected positional arguments")
	}

	slackWebhookURL := strings.TrimSpace(os.Getenv("SLACK_WEBHOOK_URL"))
	webhook, err := url.Parse(slackWebhookURL)
	if err != nil || webhook.Hostname() == "" || (webhook.Scheme != "https" && webhook.Scheme != "http") || webhook.User != nil {
		return nil, fmt.Errorf("SLACK_WEBHOOK_URL must be a non-empty HTTP or HTTPS URL without user information")
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

	// Synchronous delivery is not implemented. Reject invalid settings before startup.
	queueEnabled := true
	if value := os.Getenv("QUEUE_ENABLED"); value != "" {
		var err error
		queueEnabled, err = strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("QUEUE_ENABLED must be a boolean")
		}
	}
	if !queueEnabled {
		return nil, fmt.Errorf("QUEUE_ENABLED must be true; synchronous mode is not supported")
	}
	queueWorkers, err := positiveEnvInt("QUEUE_WORKERS", 10)
	if err != nil {
		return nil, err
	}
	queueSize, err := positiveEnvInt("QUEUE_SIZE", 1000)
	if err != nil {
		return nil, err
	}
	httpPort, err := positiveEnvInt("METRICS_PORT", 8080)
	if err != nil {
		return nil, err
	}
	if httpPort > 65535 {
		return nil, fmt.Errorf("METRICS_PORT must be between 1 and 65535")
	}
	if len(watchedResources) == 0 || len(watchedNamespaces) == 0 {
		return nil, fmt.Errorf("WATCHED_RESOURCE_NAMES and WATCHED_NAMESPACES must contain at least one name")
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
		HTTPPort:          httpPort,
	}, nil
}

// splitAndTrim splits comma-separated values and discards blank entries.
func splitAndTrim(s string) []string {
	var parts []string
	for _, value := range strings.Split(s, ",") {
		if value = strings.TrimSpace(value); value != "" {
			parts = append(parts, value)
		}
	}
	return parts
}

func positiveEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return number, nil
}
