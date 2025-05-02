package kubernetes

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// Client holds the Kubernetes clients
type Client struct {
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	ClusterName     string
	Config          *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient(kubeConfigPath string) (*Client, error) {
	var config *rest.Config
	var err error
	var clusterName string

	// Try to use provided kubeconfig; otherwise use in-cluster config
	if kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
		}

		// Extract cluster name from current context
		rawConfig, err := clientcmd.LoadFromFile(kubeConfigPath)
		if err != nil {
			klog.Warningf("Error loading raw Kubernetes configuration: %v", err)
			clusterName = "unknown-cluster"
		} else {
			currentContext := rawConfig.CurrentContext
			context, exists := rawConfig.Contexts[currentContext]
			if !exists {
				clusterName = "unknown-cluster"
			} else {
				clusterName = context.Cluster
			}
		}

		klog.Info("Using out-of-cluster config")
	} else {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build in-cluster config: %v", err)
		}
		clusterName = "in-cluster"
		klog.Info("Using in-cluster config")
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %v", err)
	}

	return &Client{
		DynamicClient:   dynamicClient,
		DiscoveryClient: discoveryClient,
		ClusterName:     clusterName,
		Config:          config,
	}, nil
}
