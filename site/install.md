---
layout: default
title: Installation
---

# Install with Helm

Requires Helm 3, <code>kubectl</code>, access to a Kubernetes cluster, permission to create a ClusterRole and ClusterRoleBinding, a Slack webhook, and an image the cluster can pull.

## Container image availability

The chart defaults to the published <code>ghcr.io/mina-maher/k8s-diff-informer:1.1.0</code> image. The GHCR package is public, so no registry credentials are required. To install a custom build, push it to a registry your cluster can access:

~~~sh
docker build -t YOUR_REGISTRY/k8s-diff-informer:YOUR_TAG .
docker push YOUR_REGISTRY/k8s-diff-informer:YOUR_TAG
~~~

## Create the namespace and webhook Secret

Use a protected local file that contains only the Slack webhook URL:

~~~sh
kubectl create namespace monitoring
kubectl create secret generic informer-slack -n monitoring \
  --from-file=webhook-url=/secure/path/slack-webhook
~~~

## Install the chart

~~~sh
helm upgrade --install diff-monitor ./deployment/helm \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --set config.clusterName=my-cluster \
  --wait --timeout 5m
~~~

The chart uses the published stable image by default. To install a custom build
or use another registry, set <code>image.repository</code> and
<code>image.tag</code> explicitly.

Stable releases created after OCI chart publishing is merged will also publish
the chart to GHCR. Install a published version by replacing <code>1.1.0</code>
with the version you need:

~~~sh
helm upgrade --install diff-monitor \
  oci://ghcr.io/mina-maher/charts/k8s-diff-informer \
  --version 1.1.0 \
  --namespace monitoring \
  --set slack.existingSecret=informer-slack \
  --wait --timeout 5m
~~~

The existing <code>1.0.0</code> chart remains available as a <code>.tgz</code>
asset on its [GitHub Release](https://github.com/MIna-Maher/k8s-diff-informer/releases/tag/v1.0.0).
The OCI chart package must be public for anonymous pulls.

Set exactly one of <code>slack.existingSecret</code> and <code>slack.webhookUrl</code>. Existing Secrets must be in the release namespace and contain a non-empty HTTP(S) URL under <code>slack.existingSecretKey</code> (default: <code>webhook-url</code>). For private registries, pass credentials with <code>imagePullSecrets</code>.

Read the complete [Helm chart guide](https://github.com/MIna-Maher/k8s-diff-informer/blob/main/deployment/helm/README.md) for configuration, RBAC extensions, security, and troubleshooting.

## Redact sensitive values

Secret resource `data` and `stringData` values are always redacted in update
diffs. Common credential-like field names are also masked. For application
specific paths, add `config.fieldsToRedact` to a Helm values file. For example,
to mask ConfigMap values and literal container environment values:

~~~yaml
config:
  fieldsToRedact:
    - data.*
    - spec.template.spec.containers.*.env.*.value
~~~

The diff reports that a redacted path changed without exposing its old or new
value. Automatic name checks are best-effort and cannot identify every secret
inside arbitrary configuration strings; configure paths for the resources you
watch. A `*` matches one map key or list index. Escape a literal dot in a key
with `\.` (for example, `data.app\.yaml`).

## Local execution

~~~sh
export SLACK_WEBHOOK_URL='https://YOUR_WEBHOOK_HOST/YOUR_WEBHOOK_PATH'
go run ./cmd/k8s-diff-informer --kubeconfig "$HOME/.kube/config"
~~~

The application uses <code>~/.kube/config</code> outside a cluster unless <code>--kubeconfig</code> selects another file. In a pod it uses in-cluster credentials. See the repository's [configuration table](https://github.com/MIna-Maher/k8s-diff-informer#configuration).
