# k8s-diff-informer Helm chart

Deploy one Kubernetes resource watcher with Slack notifications.

## Prerequisites

- Helm 3 and access to a Kubernetes cluster.
- Permission to create the chart's ClusterRole and ClusterRoleBinding.
- A Slack incoming webhook stored in a Secret in the installation namespace.
- Access to the selected container image.

The chart targets image `ghcr.io/mina-maher/k8s-diff-informer:1.0.0`.
The public GHCR package includes the stable `1.0.0` image, which the chart uses
by default. It can be pulled without registry credentials. To use a custom
build or another registry, push the image and override `image.repository` and
`image.tag` below.

## Install from the repository

Run these commands from the repository root. Set `SLACK_WEBHOOK_URL` in your
shell to your own webhook before creating the Secret.

```sh
kubectl create namespace monitoring
kubectl create secret generic informer-slack --namespace monitoring \
  --from-literal=webhook-url="$SLACK_WEBHOOK_URL"

helm upgrade --install diff-monitor ./deployment/helm \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --set config.clusterName=my-cluster \
  --wait --timeout 5m
```

If using a locally built image published to your own registry, append
`--set image.repository=YOUR_REGISTRY/k8s-diff-informer --set image.tag=YOUR_TAG`.
For a private registry, configure `imagePullSecrets` explicitly; the default is
an empty list. No private registry credentials or Prometheus Operator CRDs are
required by the default chart configuration.

Set exactly one of `slack.existingSecret` or `slack.webhookUrl`. Supplying neither
or both fails chart rendering. For an existing Secret, ensure the Secret exists
in the same namespace and contains a non-empty webhook under
`slack.existingSecretKey` (default `webhook-url`). Helm does not fetch or validate
external Secret contents during rendering; Kubernetes reports missing Secrets
or keys at pod startup, and the application validates the URL on startup.

Alternatively, set `slack.webhookUrl` through a private values file. Helm creates
a Secret using the same configurable key. Inline webhook values are retained in
Helm release data; do not commit that values file.

## Configuration

| Value | Default | Behavior |
| --- | --- | --- |
| `replicaCount` | `1` | Must be exactly one. |
| `image.repository` | `ghcr.io/mina-maher/k8s-diff-informer` | Container repository. |
| `image.tag` | `1.0.0` | Stable application version; override for a development build. |
| `image.pullPolicy` | `IfNotPresent` | Pull policy for versioned images. |
| `imagePullSecrets` | `[]` | Optional credentials for private registries. |
| `slack.webhookUrl` | `""` | HTTP(S) webhook; mutually exclusive with `existingSecret`. |
| `slack.existingSecret` | `""` | Secret in the release namespace. |
| `slack.existingSecretKey` | `webhook-url` | Key used by both existing and chart-created Secrets. |
| `config.clusterName` | `kubernetes-cluster` | Cluster label in notifications. |
| `config.watchedResources` | `[deployments, services, configmaps, clusterroles, namespaces]` | Resource names, with corresponding RBAC rules. |
| `config.watchedNamespaces` | `[default, kube-system]` | Notification filter for namespaced resources. |
| `config.fieldsToRemove` | See `values.yaml` | Fields omitted before computing diffs. |
| `queue.enabled` | `true` | Required; synchronous mode is unsupported. |
| `queue.workers` | `10` | Positive integer. |
| `queue.size` | `1000` | Positive integer; queue capacity. |
| `metrics.enabled` | `true` | Create a Service and scrape annotations. |
| `metrics.port` | `8080` | Application HTTP port, including health probes. |
| `metrics.service.port` | `8080` | Service port, forwarded to the application HTTP port. |
| `metrics.serviceMonitor.enabled` | `false` | Opt in when Prometheus Operator is installed. |
| `metrics.prometheusRule.enabled` | `false` | Opt in when Prometheus Operator is installed. |
| `rbac.rules` | Explicit rules in `values.yaml` | Read-only resource permissions. |

## Single replica and upgrades

Each process watches resources independently and uses an in-memory queue.
Multiple replicas produce duplicate notifications. The chart rejects replica
counts other than one and uses `strategy: Recreate` to avoid overlapping old and
new pods during upgrades. This introduces downtime during an upgrade. Do not add
an HPA or run overlapping releases that watch the same resources and namespaces.

Pending notifications can be lost on restart; the queue is not durable. Initial
resource listings are not sent as new-resource notifications. For higher traffic,
tune `queue.workers`, `queue.size`, and resource limits instead of replica count.

## Resource permissions

The default ClusterRole grants `get`, `list`, and `watch` for only the default
watched resource types, plus API discovery. It does not grant access to Secrets.
The informer lists and watches across namespaces, then filters notifications;
`config.watchedNamespaces` is not a Kubernetes authorization boundary.

When adding a resource, update both `config.watchedResources` and `rbac.rules`.
Helm replaces arrays, so retain rules for resources you still watch. For example,
to watch only Deployments and Pods:

```yaml
config:
  watchedResources: [deployments, pods]
  watchedNamespaces: [default]
rbac:
  rules:
    - apiGroups: [apps]
      resources: [deployments]
      verbs: [get, list, watch]
    - apiGroups: [""]
      resources: [pods]
      verbs: [get, list, watch]
```

For custom resources, use their API group and plural resource name. Missing
permissions prevent initial synchronization, so the pod stays unready. Resource
diffs can contain sensitive configuration; choose watched resources and ignored
fields accordingly.

## Health and monitoring

`/health` and `/healthz` report process health. `/ready` and `/readyz` return HTTP
503 until all watched informer caches finish their initial synchronization, and
HTTP 200 afterward. Shutdown makes readiness false. Readiness does not guarantee
Slack delivery or continuously verify Kubernetes API connectivity after the
initial sync. HTTP bind failures stop the application with an error.

`metrics.port` controls HTTP endpoints even when `metrics.enabled=false`; disabling
the metrics Service does not disable the application's health or metrics handlers.

For Prometheus Operator installations, enable integrations explicitly:

```sh
helm upgrade --install diff-monitor ./deployment/helm \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.prometheusRule.enabled=true
```

Match ServiceMonitor labels to your Prometheus installation. Enabling these
resources requires their CRDs. See [monitoring documentation](../../docs/MONITORING.md).

## Validate and troubleshoot

```sh
python3 -B -m unittest discover -s test/helm
helm lint ./deployment/helm --set slack.existingSecret=informer-slack
helm template diff-monitor ./deployment/helm --set slack.existingSecret=informer-slack
kubectl get pods --namespace monitoring -l app.kubernetes.io/instance=diff-monitor
kubectl logs --namespace monitoring deployment/diff-monitor-k8s-diff-informer
helm test diff-monitor --namespace monitoring
```

The Helm test probes readiness through the metrics Service and fails on a
non-success response. It is omitted when `metrics.enabled=false`; in that case,
use pod readiness and `kubectl rollout status` to check startup.

For `ImagePullBackOff`, check image availability and registry credentials. For
`CreateContainerConfigError`, check the Secret name and key. For a pod that stays
unready, inspect logs for discovery, network, or RBAC errors. If NetworkPolicy is
enabled, allow DNS and the actual Kubernetes API and webhook endpoints; default
egress ports cover DNS (53), HTTPS (443), and the common API server port (6443).

## Local execution

```sh
export SLACK_WEBHOOK_URL="https://YOUR_WEBHOOK_HOST/YOUR_WEBHOOK_PATH"
export WATCHED_RESOURCE_NAMES="deployments,services"
export WATCHED_NAMESPACES="default"
go run ./cmd/k8s-diff-informer --kubeconfig "$HOME/.kube/config"
```

`--kubeconfig` works outside and inside a cluster. With no flag, the application
uses in-cluster credentials when a service account directory exists, otherwise
`~/.kube/config`. An explicitly empty path selects in-cluster credentials.
`--help` works without Slack configuration or cluster access.

`SLACK_WEBHOOK_URL` must be a valid HTTP(S) URL. `QUEUE_ENABLED` must be true;
`QUEUE_WORKERS` and `QUEUE_SIZE` must be positive integers. `METRICS_PORT` defaults
to 8080 and must be between 1 and 65535. Invalid settings fail before workers start.
