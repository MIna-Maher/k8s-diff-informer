package config

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func cleanEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"SLACK_WEBHOOK_URL", "CLUSTER_NAME", "WATCHED_RESOURCE_NAMES", "WATCHED_NAMESPACES", "RESOURCES_IGNORED_FIELDS", "QUEUE_ENABLED", "QUEUE_WORKERS", "QUEUE_SIZE", "METRICS_PORT"} {
		t.Setenv(key, "")
	}
	t.Setenv("SLACK_WEBHOOK_URL", "https://example.invalid/webhook")
}

func TestConfigurationOverrides(t *testing.T) {
	cleanEnvironment(t)
	t.Setenv("QUEUE_WORKERS", "3")
	t.Setenv("QUEUE_SIZE", "42")
	t.Setenv("METRICS_PORT", "9090")
	t.Setenv("WATCHED_RESOURCE_NAMES", "deployments, , services")
	cfg, err := loadConfig([]string{"--kubeconfig", "/tmp/custom-kubeconfig"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KubeConfigPath != "/tmp/custom-kubeconfig" || cfg.QueueWorkers != 3 || cfg.QueueSize != 42 || cfg.HTTPPort != 9090 || len(cfg.WatchedResources) != 2 {
		t.Fatalf("overrides not applied: %+v", cfg)
	}
	cfg, err = loadConfig([]string{"--kubeconfig="})
	if err != nil || cfg.KubeConfigPath != "" {
		t.Fatalf("explicit in-cluster config: %v", err)
	}
}

func TestInvalidConfiguration(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"SLACK_WEBHOOK_URL", ""}, {"SLACK_WEBHOOK_URL", "not-a-url"},
		{"SLACK_WEBHOOK_URL", "https://user:secret@example.invalid/webhook"},
		{"QUEUE_ENABLED", "false"}, {"QUEUE_ENABLED", "maybe"},
		{"QUEUE_WORKERS", "0"}, {"QUEUE_WORKERS", "-1"}, {"QUEUE_WORKERS", "many"},
		{"QUEUE_SIZE", "0"}, {"QUEUE_SIZE", "-5"}, {"QUEUE_SIZE", "large"},
		{"METRICS_PORT", "0"}, {"METRICS_PORT", "65536"}, {"METRICS_PORT", "http"},
		{"WATCHED_RESOURCE_NAMES", " , "}, {"WATCHED_NAMESPACES", " , "},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			cleanEnvironment(t)
			t.Setenv(tc.key, tc.value)
			_, err := loadConfig(nil)
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("expected %s validation error, got %v", tc.key, err)
			}
			if strings.Contains(err.Error(), "user:secret") {
				t.Fatal("error leaked credentials")
			}
		})
	}
}

func TestHelpAndBadFlags(t *testing.T) {
	cleanEnvironment(t)
	t.Setenv("SLACK_WEBHOOK_URL", "")
	if _, err := loadConfig([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help: %v", err)
	}
	for _, args := range [][]string{{"--unknown"}, {"unexpected"}} {
		if _, err := loadConfig(args); err == nil {
			t.Fatalf("accepted invalid arguments: %v", args)
		}
	}
}
