# k8s-diff-informer Helm Chart

A Helm chart for deploying k8s-diff-informer, a Kubernetes resource change monitoring application with Slack notifications.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- A Slack webhook URL for notifications

## Installation

### 1. Add the Helm repository (if published)
```bash
# If you publish to a Helm repository
helm repo add k8s-diff-informer https://your-repo-url.com
helm repo update
```

### 2. Install from local directory
```bash
# Clone the repository and navigate to the helm chart directory
git clone https://github.com/MIna-Maher/k8s-diff-informer.git
cd k8s-diff-informer/helm

# Install the chart
helm install my-k8s-diff-informer . \
  --set slack.webhookUrl="YOUR_SLACK_WEBHOOK_URL" \
  --set config.clusterName="my-cluster"
```

### 3. Install with custom values
```bash
# Create a custom values file
cat > my-values.yaml << EOF
config:
  clusterName: "production-cluster"
  watchedResources:
    - "pods"
    - "deployments"
    - "services"
    - "configmaps"
    - "secrets"
  watchedNamespaces:
    - "default"
    - "kube-system"
    - "production"

slack:
  webhookUrl: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi

replicaCount: 2

podDisruptionBudget:
  enabled: true
  minAvailable: 1
EOF

helm install my-k8s-diff-informer . -f my-values.yaml
```

### 4. Install with existing secret
```bash
# Create the secret first
kubectl create secret generic my-slack-secret \
  --from-literal=webhook-url="YOUR_SLACK_WEBHOOK_URL"

# Install referencing the existing secret
helm install my-k8s-diff-informer . \
  --set slack.existingSecret="my-slack-secret" \
  --set slack.existingSecretKey="webhook-url"
```

## Configuration

### Core Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `k8s-diff-informer` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Slack Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `slack.webhookUrl` | Slack webhook URL (creates secret) | `""` |
| `slack.existingSecret` | Name of existing secret containing webhook URL | `""` |
| `slack.existingSecretKey` | Key in existing secret | `webhook-url` |

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.clusterName` | Name of the Kubernetes cluster | `kubernetes-cluster` |
| `config.watchedResources` | List of resources to monitor | `["pods", "deployments", "services", "configmaps"]` |
| `config.watchedNamespaces` | List of namespaces to monitor | `["default", "kube-system"]` |
| `config.fieldsToRemove` | Fields to ignore in diffs | See values.yaml |

### Security Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run as non-root user | `true` |
| `podSecurityContext.runAsUser` | User ID to run as | `65532` |
| `podSecurityContext.runAsGroup` | Group ID to run as | `65532` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.capabilities.drop` | Dropped capabilities | `["ALL"]` |

### Resource Management

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |

### Autoscaling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `3` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization | `80` |

### High Availability

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podDisruptionBudget.enabled` | Enable PDB | `false` |
| `podDisruptionBudget.minAvailable` | Minimum available pods | `1` |

### Network Security

| Parameter | Description | Default |
|-----------|-------------|---------|
| `networkPolicy.enabled` | Enable network policy | `false` |
| `networkPolicy.policyTypes` | Policy types | `["Egress"]` |

## Security Best Practices Implemented

### 1. **Non-Root Execution**
- Runs as user ID 65532 (non-root)
- Explicitly sets `runAsNonRoot: true`
- Drops all capabilities

### 2. **Read-Only Root Filesystem**
- Container filesystem is read-only
- Temporary directory mounted for any needed writes

### 3. **Minimal RBAC Permissions**
- ClusterRole with only necessary permissions
- Read-only access to monitored resources
- No write permissions granted

### 4. **Secret Management**
- Slack webhook stored in Kubernetes Secret
- Base64 encoded and not exposed in values
- Support for external secret management

### 5. **Resource Limits**
- CPU and memory limits defined
- Prevents resource exhaustion
- Configurable based on cluster size

### 6. **Network Security**
- Optional NetworkPolicy for egress control
- Restricts outbound traffic to necessary endpoints

### 7. **Security Contexts**
- Pod and container security contexts configured
- Seccomp profile set to RuntimeDefault
- No privilege escalation allowed

## Usage Examples

### Basic Monitoring Setup
```bash
helm install diff-monitor . \
  --set slack.webhookUrl="YOUR_WEBHOOK_URL" \
  --set config.clusterName="production" \
  --set config.watchedNamespaces="{default,production,staging}"
```

### High Availability Setup
```bash
helm install diff-monitor . \
  --set replicaCount=3 \
  --set podDisruptionBudget.enabled=true \
  --set autoscaling.enabled=true \
  --set slack.webhookUrl="YOUR_WEBHOOK_URL"
```

### Security-Hardened Setup
```bash
helm install diff-monitor . \
  --set networkPolicy.enabled=true \
  --set podSecurityContext.seccompProfile.type="RuntimeDefault" \
  --set slack.webhookUrl="YOUR_WEBHOOK_URL"
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade my-k8s-diff-informer . \
  --set config.watchedResources="{pods,deployments,services,secrets}"

# Upgrade with new image version
helm upgrade my-k8s-diff-informer . \
  --set image.tag="v1.1.0"
```

## Uninstalling

```bash
helm uninstall my-k8s-diff-informer
```

## Testing

Run the included Helm tests:

```bash
helm test my-k8s-diff-informer
```

## Troubleshooting

### Check deployment status
```bash
kubectl get deployment my-k8s-diff-informer
kubectl describe deployment my-k8s-diff-informer
```

### Check pod logs
```bash
kubectl logs -f deployment/my-k8s-diff-informer
```

### Verify RBAC permissions
```bash
kubectl auth can-i list pods --as=system:serviceaccount:default:my-k8s-diff-informer
```

### Check secret configuration
```bash
kubectl get secret my-k8s-diff-informer-slack -o yaml
```

### Validate configuration
```bash
kubectl get configmap my-k8s-diff-informer -o yaml
```

## Building and Pushing the Docker Image

Before deploying, ensure your image is built and available:

```bash
# Build the image
docker build -t k8s-diff-informer:latest .

# Tag for your registry
docker tag k8s-diff-informer:latest your-registry.com/k8s-diff-informer:latest

# Push to registry
docker push your-registry.com/k8s-diff-informer:latest

# Update Helm values
helm install my-k8s-diff-informer . \
  --set image.repository="your-registry.com/k8s-diff-informer" \
  --set image.tag="latest"
```

## Development and Customization

### Adding New Resources to Monitor
1. Update `config.watchedResources` in values.yaml
2. Ensure RBAC permissions include the new resource types
3. Update the ClusterRole in `templates/clusterrole.yaml`

Example:
```yaml
# In values.yaml
config:
  watchedResources:
    - "pods"
    - "deployments" 
    - "services"
    - "ingresses"  # New resource
    - "persistentvolumes"  # New resource
```

```yaml
# In templates/clusterrole.yaml - add new rules
- apiGroups: ["networking.k8s.io"]
  resources:
    - ingresses
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
    - persistentvolumes
  verbs: ["get", "list", "watch"]
```

### Custom Security Policies

For environments requiring additional security:

```yaml
# values-security-hardened.yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532
  runAsGroup: 65532
  fsGroup: 65532
  seccompProfile:
    type: RuntimeDefault
  supplementalGroups: []

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65532
  runAsGroup: 65532
  capabilities:
    drop:
    - ALL
  seccompProfile:
    type: RuntimeDefault

networkPolicy:
  enabled: true
  policyTypes:
    - Egress
  egress:
    # Kubernetes API server
    - to: []
      ports:
        - protocol: TCP
          port: 6443
    # Slack webhook (HTTPS)
    - to: []
      ports:
        - protocol: TCP
          port: 443
    # DNS
    - to: []
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
```

### Monitoring Integration

For environments with Prometheus monitoring:

```yaml
# values-monitoring.yaml
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    namespace: "monitoring"
    labels:
      prometheus: "kube-prometheus"
    interval: 30s
    path: /metrics

# Add metrics port to the container (requires code changes)
service:
  enabled: true
  type: ClusterIP
  port: 8080
  targetPort: 8080
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
```

## Advanced Configuration Examples

### Multi-Environment Setup

For managing multiple environments with different configurations:

```bash
# Development environment
cat > values-dev.yaml << EOF
config:
  clusterName: "development"
  watchedNamespaces: ["default", "dev"]
  watchedResources: ["pods", "deployments"]

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 64Mi

replicaCount: 1
EOF

# Production environment  
cat > values-prod.yaml << EOF
config:
  clusterName: "production"
  watchedNamespaces: ["default", "production", "kube-system"]
  watchedResources: ["pods", "deployments", "services", "configmaps", "secrets"]

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi

replicaCount: 3
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5

podDisruptionBudget:
  enabled: true
  minAvailable: 2

networkPolicy:
  enabled: true
EOF

# Deploy to different environments
helm install diff-monitor-dev . -f values-dev.yaml
helm install diff-monitor-prod . -f values-prod.yaml
```

### GitOps Integration

For ArgoCD or Flux integration:

```yaml
# argocd-application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: k8s-diff-informer
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/MIna-Maher/k8s-diff-informer.git
    targetRevision: HEAD
    path: helm
    helm:
      valueFiles:
        - values-production.yaml
      parameters:
        - name: config.clusterName
          value: "production-cluster"
        - name: slack.existingSecret
          value: "slack-webhook-secret"
  destination:
    server: https://kubernetes.default.svc
    namespace: monitoring
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

## Best Practices

### 1. **Secret Management**
- Use external secret management (Vault, AWS Secrets Manager, etc.)
- Rotate secrets regularly
- Use separate secrets per environment

```yaml
# Using External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "k8s-diff-informer"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: slack-webhook
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: k8s-diff-informer-slack
    creationPolicy: Owner
  data:
  - secretKey: webhook-url
    remoteRef:
      key: slack/k8s-diff-informer
      property: webhook-url
```

### 2. **Resource Management**
- Set appropriate resource limits based on cluster size
- Monitor resource usage and adjust accordingly
- Use HPA for variable workloads

### 3. **High Availability**
- Use multiple replicas in production
- Configure Pod Disruption Budgets
- Spread pods across nodes with anti-affinity

```yaml
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - k8s-diff-informer
        topologyKey: kubernetes.io/hostname
```

### 4. **Monitoring and Alerting**
- Monitor application logs
- Set up alerts for application failures
- Monitor resource usage

## Common Issues and Solutions

### Issue: RBAC Permission Denied
```bash
# Check service account permissions
kubectl auth can-i list pods \
  --as=system:serviceaccount:default:k8s-diff-informer

# Fix: Update ClusterRole with required permissions
```

### Issue: Slack Notifications Not Working
```bash
# Verify secret content
kubectl get secret k8s-diff-informer-slack -o jsonpath='{.data.webhook-url}' | base64 -d

# Check application logs
kubectl logs -f deployment/k8s-diff-informer
```

### Issue: High Memory Usage
```bash
# Monitor resource usage
kubectl top pods -l app.kubernetes.io/name=k8s-diff-informer

# Increase memory limits or reduce watched resources
helm upgrade k8s-diff-informer . \
  --set resources.limits.memory=1Gi
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with different Kubernetes versions
5. Update documentation
6. Submit a pull request

## Support

- **Issues**: Create an issue on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Security**: Report security issues privately

## Changelog

### v0.1.0
- Initial Helm chart release
- Security best practices implementation
- ConfigMap and Secret management
- RBAC configuration
- Optional NetworkPolicy support
- HPA and PDB support

## License

This Helm chart is licensed under the MIT License.