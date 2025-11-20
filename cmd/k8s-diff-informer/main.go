package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MIna-Maher/k8s-diff-informer/internal/config"
	httpserver "github.com/MIna-Maher/k8s-diff-informer/internal/http"
	"github.com/MIna-Maher/k8s-diff-informer/internal/kubernetes"
	"github.com/MIna-Maher/k8s-diff-informer/internal/metrics"
	"github.com/MIna-Maher/k8s-diff-informer/internal/queue"
	"github.com/MIna-Maher/k8s-diff-informer/internal/slack"
	"k8s.io/klog/v2"
)

func main() {
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/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")
	os.Setenv("WATCHED_RESOURCE_NAMES", "nodes,deployments,configmaps,namespaces,services")
	os.Setenv("WATCHED_NAMESPACES", "default,mina,kube-system,cnr-system")
	os.Setenv("CLUSTER_NAME", "my-test-cluster")
	os.Setenv("QUEUE_ENABLED", "true")
	os.Setenv("QUEUE_WORKERS", "20")
	os.Setenv("QUEUE_SIZE", "1000")

	//initialize metrics
	appMetrics := metrics.NewMetrics()
	klog.Info("Prometheus metrics initialized")

	// Start uptime recorder
	appMetrics.StartUptimeRecorder()

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		klog.Fatalf("Failed to load configuration: %v", err)
	}

	//Set watched resources and namespaces metrics
	appMetrics.SetWatchedResourcesCount(len(cfg.WatchedResources))
	appMetrics.SetWatchedNamespacesCount(len(cfg.WatchedNamespaces))

	// Create Kubernetes client
	kubeClient, err := kubernetes.NewClient(cfg.KubeConfigPath)
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Initialize Slack client
	slackClient := slack.NewSlackClientWithMetrics(cfg.SlackWebhookURL, cfg.ClusterName, appMetrics)

	// Initialize queue if enabled
	var taskQueue *queue.MemoryQueue
	if cfg.QueueEnabled {
		klog.Info("Initializing task queue...")
		taskQueue = queue.NewMemoryQueue(cfg.QueueWorkers, cfg.QueueSize, appMetrics)

		// Create task handler
		taskHandler := queue.NewHandler(slackClient, appMetrics)

		// Start queue workers
		taskQueue.Start(taskHandler.Handle)
		klog.Infof("Task queue started with %d workers", cfg.QueueWorkers)
	} else {
		klog.Warning("Queue is disabled - running in synchronous mode")
	}

	// Create and start informers
	stopCh := make(chan struct{})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	//A background goroutine waits for OS signals on sigCh
	go func() {
		sig := <-sigCh
		klog.Infof("Received signal %s, shutting down...", sig)

		// Stop the queue first to drain pending tasks
		if taskQueue != nil {
			klog.Info("Stopping task queue...")
			taskQueue.Stop()
			klog.Info("Task queue stopped...")
		}

		// Then close the stop channel for informers
		close(stopCh)
	}()

	// Start HTTP server for metrics and health checks
	httpServer := httpserver.NewServer(8080)
	//background goroutine runs the HTTP metrics server
	go func() {
		if err := httpServer.Start(ctx); err != nil {
			klog.Errorf("HTTP server error: %v", err)
		}
	}()

	// Create informer manager based on queue availability
	var informerManager *kubernetes.InformerManager

	if cfg.QueueEnabled && taskQueue != nil {
		// Use queue-based informer manager
		informerManager = kubernetes.NewInformerManagerWithQueue(
			kubeClient.DynamicClient,
			kubeClient.DiscoveryClient,
			cfg.WatchedResources,
			cfg.WatchedNamespaces,
			cfg.FieldsToRemove,
			taskQueue,
			appMetrics,
			cfg.ClusterName,
		)
		klog.Info("Using queue-based informer manager (asynchronous mode)")
	} else {
		// Fallback to direct Slack client (synchronous mode)
		//Todo: implement synchronous informer manager
		// Note: You'll need to implement NewInformerManagerWithSlackClient
		// or modify the existing code to support both modes
		klog.Fatal("Synchronous mode not implemented in this version. Please enable queue.")
	}

	klog.Infof("Starting K8s-Diff-Informer...")
	klog.Infof("Watching for changes in resources: %v, in namespaces: %v", cfg.WatchedResources, cfg.WatchedNamespaces)
	klog.Infof("Metrics server running on :8080/metrics")
	klog.Infof("Queue mode: %v", cfg.QueueEnabled)

	if err := informerManager.Start(stopCh); err != nil {
		klog.Fatalf("Failed to start informers: %v", err)
	}

	<-stopCh // Main goroutine waits here.

	klog.Info("Shutting down gracefully...")

	// Print queue statistics if available
	if taskQueue != nil {
		stats := taskQueue.Stats()
		klog.Infof("Queue statistics: %+v", stats)
	}

	klog.Info("Shutdown complete..., Hope see you later!")
}
