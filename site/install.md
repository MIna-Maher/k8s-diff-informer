---
layout: default
title: Installation
---

# Install with Helm

Requires Helm 3, <code>kubectl</code>, access to a Kubernetes cluster, permission to create a ClusterRole and ClusterRoleBinding, a Slack webhook, and an image the cluster can pull.

## Container image availability

The chart defaults to <code>ghcr.io/mina-maher/k8s-diff-informer:1.0.0</code>. That stable image has not been published yet. The GitHub Container Registry <code>beta</code> package is private. Build and push an image from this source to a registry your cluster can access:

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
  --set image.repository=YOUR_REGISTRY/k8s-diff-informer \
  --set image.tag=YOUR_TAG \
  --wait --timeout 5m
~~~

Set exactly one of <code>slack.existingSecret</code> and <code>slack.webhookUrl</code>. Existing Secrets must be in the release namespace and contain a non-empty HTTP(S) URL under <code>slack.existingSecretKey</code> (default: <code>webhook-url</code>). For private registries, pass credentials with <code>imagePullSecrets</code>.

Read the complete [Helm chart guide](https://github.com/MIna-Maher/k8s-diff-informer/blob/main/deployment/helm/README.md) for configuration, RBAC extensions, security, and troubleshooting.

## Local execution

~~~sh
export SLACK_WEBHOOK_URL='https://YOUR_WEBHOOK_HOST/YOUR_WEBHOOK_PATH'
go run ./cmd/k8s-diff-informer --kubeconfig "$HOME/.kube/config"
~~~

The application uses <code>~/.kube/config</code> outside a cluster unless <code>--kubeconfig</code> selects another file. In a pod it uses in-cluster credentials. See the repository's [configuration table](https://github.com/MIna-Maher/k8s-diff-informer#configuration).
