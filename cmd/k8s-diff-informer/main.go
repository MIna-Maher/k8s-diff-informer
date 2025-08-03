package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MIna-Maher/k8s-diff-informer/internal/config"
	"github.com/MIna-Maher/k8s-diff-informer/internal/kubernetes"
	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"k8s.io/klog/v2"
)

func main() {
	// os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/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")
	// os.Setenv("WATCHED_RESOURCE_NAMES", "nodes,deployments,configmaps,namespaces,services,pipelineruns")
	// os.Setenv("WATCHED_NAMESPACES", "default,mina,kube-system,cnr-system")
	// os.Setenv("CLUSTER_NAME", "minaTestingCluster")

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		klog.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(cfg.KubeConfigPath)
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Initialize Slack client
	slackClient := slack.NewSlackClient(cfg.SlackWebhookURL, cfg.ClusterName)

	// Create and start informers
	stopCh := make(chan struct{})

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		klog.Infof("Received signal %s, shutting down...", sig)
		close(stopCh)
	}()

	// Start informers for watched resources
	informerManager := kubernetes.NewInformerManager(
		kubeClient.DynamicClient,
		kubeClient.DiscoveryClient,
		cfg.WatchedResources,
		cfg.WatchedNamespaces,
		cfg.FieldsToRemove,
		slackClient,
	)

	klog.Infof("Starting K8s-Diff-Informer...")
	klog.Infof("Watching for changes in resources: %v, in namespaces: %v", cfg.WatchedResources, cfg.WatchedNamespaces)

	if err := informerManager.Start(stopCh); err != nil {
		klog.Fatalf("Failed to start informers: %v", err)
	}

	<-stopCh
	klog.Info("Shutting down...")
}
