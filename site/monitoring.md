---
layout: default
title: Monitoring
---

# Monitoring

The application exposes Prometheus metrics, liveness, and readiness on the configured HTTP port (default <code>8080</code>).

| Path | Purpose |
| --- | --- |
| <code>/metrics</code> | Prometheus metrics |
| <code>/health</code> and <code>/healthz</code> | Process liveness |
| <code>/ready</code> and <code>/readyz</code> | HTTP 503 until every informer cache has synchronized |

Readiness does not test Slack delivery or ongoing Kubernetes API connectivity. The HTTP endpoints remain active when the Helm metrics Service is disabled.

## Prometheus Operator

The chart creates a metrics Service by default. ServiceMonitor and PrometheusRule objects are disabled by default because they require Prometheus Operator CRDs. Enable them in clusters where the CRDs are installed:

~~~sh
helm upgrade --install diff-monitor ./deployment/helm \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.prometheusRule.enabled=true
~~~

Set labels to match your Prometheus instance using <code>metrics.serviceMonitor.labels</code> and <code>metrics.prometheusRule.labels</code>.

Useful metrics include resource events, diff calculations, informer synchronization, Slack delivery errors, and queue drops. See the complete [monitoring guide](https://github.com/MIna-Maher/k8s-diff-informer/blob/main/docs/MONITORING.md).
