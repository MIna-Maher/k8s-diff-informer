package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, flag.ErrHelp) {
		klog.Error(err)
		os.Exit(1)
	}
}

func run(parent context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	kubeClient, err := kubernetes.NewClient(cfg.KubeConfigPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	appMetrics := metrics.NewMetrics()
	appMetrics.StartUptimeRecorder()
	appMetrics.SetWatchedResourcesCount(len(cfg.WatchedResources))
	appMetrics.SetWatchedNamespacesCount(len(cfg.WatchedNamespaces))
	slackClient := slack.NewSlackClientWithMetrics(cfg.SlackWebhookURL, cfg.ClusterName, appMetrics)
	taskQueue := queue.NewMemoryQueue(cfg.QueueWorkers, cfg.QueueSize, appMetrics)
	taskQueue.Start(queue.NewHandler(slackClient, appMetrics).Handle)
	informerManager := kubernetes.NewInformerManagerWithQueue(
		kubeClient.DynamicClient, kubeClient.DiscoveryClient, cfg.WatchedResources,
		cfg.WatchedNamespaces, cfg.FieldsToRemove, taskQueue, appMetrics, cfg.ClusterName,
	)
	httpServer := httpserver.NewServer(cfg.HTTPPort, func() bool {
		return ctx.Err() == nil && informerManager.HasSynced()
	})
	serverDone := make(chan error, 1)
	go func() {
		err := httpServer.Start(ctx)
		serverDone <- err
		if err != nil {
			cancel()
		}
	}()
	defer func() {
		cancel()
		informerManager.Shutdown()
		taskQueue.Stop()
	}()

	klog.Infof("Watching resources %v in namespaces %v", cfg.WatchedResources, cfg.WatchedNamespaces)
	startupErr := informerManager.Start(ctx.Done())
	if startupErr == nil {
		<-ctx.Done()
	}
	cancel()
	serverErr := <-serverDone
	if serverErr != nil {
		return fmt.Errorf("HTTP server: %w", serverErr)
	}
	if startupErr != nil && parent.Err() == nil {
		return startupErr
	}
	return nil
}
