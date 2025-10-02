# Prometheus Integration for K8s Diff Informer

This document describes the Prometheus monitoring integration for the K8s Diff Informer project.

## Overview

The Prometheus integration provides comprehensive monitoring and alerting capabilities for the K8s Diff Informer, including:

- **Application metrics**: Resource events, processing performance, and health status
- **Business metrics**: Slack notifications, diff computations, and cache performance
- **Infrastructure metrics**: CPU, memory usage, and uptime
- **Alerting rules**: Pre-configured alerts for common issues
- **Grafana dashboard**: Visualization of all metrics

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   K8s Diff      │    │   Prometheus    │    │     Grafana     │
│   Informer      │───▶│    Server       │───▶│   Dashboard     │
│   (Port 8080)   │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Metrics        │    │ ServiceMonitor  │    │   Alertmanager  │
│  Endpoint       │    │ PrometheusRule  │    │                 │
│  /metrics       │    │                 │    │   (Alerts)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Available Metrics

### Resource Event Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_resource_events_total` | Counter | Total number of Kubernetes resource events processed | `resource_type`, `event_type`, `namespace`, `cluster` |
| `k8s_diff_informer_events_processed_total` | Counter | Total number of events processed by informers | `resource_type`, `event_type`, `cluster` |
| `k8s_diff_informer_processing_duration_seconds` | Histogram | Duration of event processing | |

### Slack Notification Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_slack_notifications_total` | Counter | Total number of Slack notifications sent | `status`, `cluster` |
| `k8s_diff_informer_slack_notification_duration_seconds` | Histogram | Duration of Slack notification requests | |
| `k8s_diff_informer_slack_notification_errors_total` | Counter | Total number of Slack notification errors | `error_type`, `cluster` |

### Informer Status Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_sync_status` | Gauge | Status of informer sync (1 = synced, 0 = not synced) | `resource_type`, `cluster` |
| `k8s_diff_informer_last_sync_timestamp` | Gauge | Timestamp of last successful informer sync | `resource_type`, `cluster` |

### Application Health Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_uptime_seconds` | Gauge | Application uptime in seconds | |
| `k8s_diff_informer_watched_resources_count` | Gauge | Number of resources being watched | |
| `k8s_diff_informer_watched_namespaces_count` | Gauge | Number of namespaces being watched | |

### Resource Diff Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_diff_computations_total` | Counter | Total number of resource diff computations | `resource_type`, `has_changes`, `cluster` |
| `k8s_diff_informer_diff_computation_duration_seconds` | Histogram | Duration of resource diff computations | |

### Cache Metrics

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|---------|
| `k8s_diff_informer_cache_size` | Gauge | Size of resource cache | `resource_type`, `cluster` |
| `k8s_diff_informer_cache_hits_total` | Counter | Total number of cache hits | `resource_type`, `cluster` |
| `k8s_diff_informer_cache_misses_total` | Counter | Total number of cache misses | `resource_type`, `cluster` |

## Installation

### 1. Update Dependencies

Add the Prometheus client library to your `go.mod`:

```bash
go mod tidy
```

### 2. Deploy with Prometheus Support

Deploy the application with metrics enabled:

```bash
helm install k8s-diff-informer ./deployment/helm \
  --set metrics.enabled=true \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.prometheusRule.enabled=true \
  --set slack.webhookUrl="YOUR_SLACK_WEBHOOK_URL"
```

### 3. Verify Metrics Endpoint

Check that metrics are being exposed:

```bash
# Port forward to the metrics port
kubectl port-forward deployment/k8s-diff-informer 8080:8080

# Check metrics endpoint
curl http://localhost:8080/metrics
```

## Configuration

### Helm Values

Enable metrics in your `values.yaml`:

```yaml
metrics:
  enabled: true
  port: 8080
  
  service:
    type: ClusterIP
    port: 8080
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "8080"
      prometheus.io/path: "/metrics"

  serviceMonitor:
    enabled: true
    interval: 30s
    path: /metrics
    
  prometheusRule:
    enabled: true
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `METRICS_PORT` | Port for metrics server | `8080` |

## Alerting Rules

The following alerting rules are included:

### Critical Alerts

- **K8sDiffInformerDown**: Application is down for more than 5 minutes
- **K8sDiffInformerHighMemoryUsage**: Memory usage over 80%
- **K8sDiffInformerHighCPUUsage**: CPU usage over 80%

### Warning Alerts

- **K8sDiffInformerSlackErrorsHigh**: High Slack notification error rate
- **K8sDiffInformerNotSynced**: Informer not synced for 10 minutes
- **K8sDiffInformerSlowProcessing**: 95th percentile processing time > 1 second
- **K8sDiffInformerLargeCacheSize**: Cache size > 10,000 entries

### Info Alerts

- **K8sDiffInformerNoEventsProcessed**: No events processed for 30 minutes

## Grafana Dashboard

Import the provided Grafana dashboard to visualize metrics:

1. Copy the dashboard JSON from the artifacts
2. In Grafana, go to **Dashboards** → **Import**
3. Paste the JSON content
4. Configure the Prometheus datasource
5. Save the dashboard

### Dashboard Panels

- **Resource Events Rate**: Rate of resource events by type
- **Informer Sync Status**: Health status of all informers
- **Slack Notifications Rate**: Success/failure rate of notifications
- **Event Processing Duration**: P50 and P95 processing times
- **Application Metrics**: Uptime, watched resources, cache sizes

## Monitoring Best Practices

### 1. Set Up Proper Resource Limits

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

### 2. Configure Alerting Thresholds

Adjust alert thresholds based on your environment:

```yaml
metrics:
  prometheusRule:
    enabled: true
    rules:
      - alert: CustomHighEventRate
        expr: rate(k8s_diff_informer_resource_events_total[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High event processing rate"
```

### 3. Monitor Key SLIs

- **Availability**: `up{job="k8s-diff-informer"}`
- **Latency**: `histogram_quantile(0.95, rate(k8s_diff_informer_processing_duration_seconds_bucket[5m]))`
- **Error Rate**: `rate(k8s_diff_informer_slack_notification_errors_total[5m])`
- **Throughput**: `rate(k8s_diff_informer_resource_events_total[5m])`

## Troubleshooting

### Metrics Not Appearing

1. Check if metrics are enabled:
```bash
kubectl get service k8s-diff-informer-metrics
```

2. Verify ServiceMonitor is created:
```bash
kubectl get servicemonitor k8s-diff-informer
```

3. Check Prometheus targets:
```bash
# Access Prometheus UI and check Status → Targets
```

### High Memory Usage

Monitor cache sizes and consider:

```bash
# Check cache metrics
curl http://localhost:8080/metrics | grep cache_size

# Reduce watched resources if needed
helm upgrade k8s-diff-informer ./deployment/helm \
  --set config.watchedResources="{deployments,services}"
```

### Missing Alerts

1. Verify PrometheusRule is created:
```bash
kubectl get prometheusrule k8s-diff-informer
```

2. Check Prometheus rules are loaded:
```bash
# In Prometheus UI: Status → Rules
```

## Integration with Monitoring Stack

### Prometheus Operator

If using Prometheus Operator:

```yaml
metrics:
  serviceMonitor:
    enabled: true
    labels:
      prometheus: kube-prometheus
      
  prometheusRule:
    enabled: true
    labels:
      prometheus: kube-prometheus
```

### Grafana

For automatic dashboard provisioning:

```yaml
# In Grafana ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard-k8s-diff-informer
data:
  dashboard.json: |
    # Paste the dashboard JSON here
```

### Alertmanager

Configure notification channels:

```yaml
# alertmanager.yml
route:
  group_by: ['alertname']
  routes:
  - match:
      service: k8s-diff-informer
    receiver: 'slack-notifications'

receivers:
- name: 'slack-notifications'
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#alerts'
    title: 'K8s Diff Informer Alert'
```

## Performance Considerations

- **Metrics Collection**: Default scrape interval is 30s
- **Retention**: Consider metrics retention policy
- **Cardinality**: Monitor label cardinality to avoid high cardinality issues
- **Resource Usage**: Metrics collection adds ~10-20% overhead

## Security

- Metrics endpoint doesn't require authentication by default
- Consider using network policies to restrict access
- Use TLS if metrics contain sensitive information

## Example Queries

### Resource Event Rates by Namespace
```promql
sum(rate(k8s_diff_informer_resource_events_total[5m])) by (namespace)
```

### Slack Notification Success Rate
```promql
rate(k8s_diff_informer_slack_notifications_total{status="success"}[5m]) / 
rate(k8s_diff_informer_slack_notifications_total[5m]) * 100
```

### Top Resource Types by Processing Time
```promql
topk(5, histogram_quantile(0.95, 
  rate(k8s_diff_informer_processing_duration_seconds_bucket[5m])
))
```

### Cache Hit Rate
```promql
rate(k8s_diff_informer_cache_hits_total[5m]) / 
(rate(k8s_diff_informer_cache_hits_total[5m]) + 
 rate(k8s_diff_informer_cache_misses_total[5m])) * 100
```

## Support

For issues with metrics:

1. Check application logs for metric-related errors
2. Verify Prometheus is scraping the endpoint
3. Ensure ServiceMonitor labels match your Prometheus configuration
4. Review Helm chart values for metrics configuration