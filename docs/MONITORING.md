# Monitoring

k8s-diff-informer serves Prometheus metrics and health endpoints from its HTTP server. The default port is <code>8080</code>; configure it with <code>METRICS_PORT</code> or the Helm value <code>metrics.port</code>.

## Endpoints

| Path | Purpose |
| --- | --- |
| <code>/metrics</code> | Prometheus metrics |
| <code>/health</code>, <code>/healthz</code> | Process liveness |
| <code>/ready</code>, <code>/readyz</code> | Readiness; HTTP 503 until all informer caches synchronize |

Readiness confirms initial informer synchronization. It does not test Slack delivery or guarantee ongoing API connectivity. The metrics and health server remains active when <code>metrics.enabled=false</code>; that value disables the Helm Service and scrape annotations.

## Prometheus Operator

The chart creates a metrics Service by default. ServiceMonitor and PrometheusRule resources are optional because they require Prometheus Operator CRDs. Enable them only when those CRDs are installed:

~~~sh
helm upgrade --install diff-monitor ./deployment/helm \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.prometheusRule.enabled=true
~~~

The default values use a ServiceMonitor label of <code>release: prometheus-operator</code>; override <code>metrics.serviceMonitor.labels</code> and <code>metrics.prometheusRule.labels</code> to match your Prometheus configuration. The Prometheus server must also discover the release namespace.

## Selected metrics

The application exports resource event, diff, informer, cache, Slack delivery, and queue metrics. Names are defined in <code>internal/metrics/mterics.go</code>. Common examples include:

- <code>k8s_diff_informer_resource_events_total</code>
- <code>k8s_diff_informer_diff_computations_total</code>
- <code>k8s_diff_informer_sync_status</code>
- <code>k8s_diff_informer_slack_notifications_total</code>
- <code>k8s_diff_informer_slack_notification_errors_total</code>
- <code>k8s_diff_informer_queue_size</code>
- <code>k8s_diff_informer_queue_tasks_dropped_total</code>

Check the live <code>/metrics</code> endpoint for the current set and labels.

## Troubleshooting

~~~sh
kubectl port-forward -n monitoring service/diff-monitor-k8s-diff-informer-metrics 8080:8080
curl -fsS http://localhost:8080/health
curl -i http://localhost:8080/ready
curl -fsS http://localhost:8080/metrics
~~~

A pod that remains unready may be waiting for API discovery or list/watch permissions. Review pod logs and the release's ClusterRole. A missing Prometheus target may indicate a ServiceMonitor label or namespace selector mismatch, or a ServiceMonitor created before the Operator CRDs existed.
