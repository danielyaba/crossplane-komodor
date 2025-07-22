# Komodor Crossplane Provider - Production Deployment

This directory contains production-ready configurations for deploying the Komodor Crossplane Provider.

## Prerequisites

1. **Kubernetes Cluster** with Crossplane installed
2. **Komodor API Key** with appropriate permissions
3. **kubectl** configured to access your cluster
4. **Docker** for building the provider image

## Quick Start

### 1. Build the Provider

```bash
# Build the Go binary
make build.code.platform PLATFORM=linux_arm64

# Copy binary to expected location
cp _output/bin/linux_arm64/provider bin/linux_arm64/provider

# Build Docker image
docker build --no-cache -t build-448a192b/provider-komodor-arm64:latest -f cluster/images/provider-komodor/Dockerfile .
```

### 2. Deploy to Your Cluster

```bash
# Load image to your cluster (if using kind)
kind load docker-image build-448a192b/provider-komodor-arm64:latest --name your-cluster-name

# Or push to your registry
docker tag build-448a192b/provider-komodor-arm64:latest your-registry/provider-komodor:latest
docker push your-registry/provider-komodor:latest
```

### 3. Configure API Key

```bash
# Create the secret with your Komodor API key
kubectl create secret generic komodor-api-secret \
  --from-literal=api-key="your-komodor-api-key" \
  -n crossplane-system
```

### 4. Deploy the Provider

```bash
# Apply the provider deployment
kubectl apply -f provider-deployment.yaml

# Apply the provider configuration
kubectl apply -f providerconfig.yaml

# Verify the provider is running
kubectl get pods -n crossplane-system -l app=provider-komodor
```

### 5. Create a Monitor

```bash
# Apply the example monitor
kubectl apply -f realtimemonitor.yaml

# Check the monitor status
kubectl get realtimemonitor production-app-monitor
```

## Production Considerations

### Security

- **API Key Management**: Store API keys in Kubernetes secrets, not in plain text
- **RBAC**: The provided RBAC configuration follows the principle of least privilege
- **Pod Security**: The deployment runs as non-root with read-only filesystem
- **Network Policies**: Consider implementing network policies to restrict pod communication

### Monitoring & Observability

- **Health Checks**: Liveness and readiness probes are configured
- **Metrics**: The provider exposes metrics on port 8080
- **Logging**: Debug logging is enabled for troubleshooting

### Resource Management

- **Resource Limits**: CPU and memory limits are set to prevent resource exhaustion
- **Replicas**: Configure multiple replicas for high availability in production
- **Node Affinity**: Consider using node affinity for better resource distribution

### Backup & Recovery

- **CRD Backup**: Ensure your backup solution includes Crossplane CRDs
- **Provider State**: The provider state is stored in Kubernetes resources
- **Disaster Recovery**: Test recovery procedures regularly

## Configuration Options

### Provider Configuration

The provider supports the following configuration options:

```yaml
apiVersion: komodor.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: komodor-provider-config
spec:
  credentials:
    source: Secret  # or Environment
    secretRef:
      namespace: crossplane-system
      name: komodor-api-secret
      key: api-key
```

### Monitor Configuration

```yaml
apiVersion: komodor.komodor.crossplane.io/v1alpha1
kind: RealtimeMonitor
metadata:
  name: example-monitor
spec:
  forProvider:
    name: "Monitor Name"
    description: "Monitor Description"
    query: "PromQL query string"
    threshold: 10
    operator: ">"  # >, <, >=, <=, ==, !=
    severity: "critical"  # critical, warning, info
    enabled: true
    tags:
      - "tag1"
      - "tag2"
```

## Troubleshooting

### Common Issues

1. **Provider Not Starting**
   ```bash
   kubectl logs -n crossplane-system -l app=provider-komodor
   ```

2. **API Authentication Errors**
   ```bash
   kubectl get secret komodor-api-secret -n crossplane-system -o yaml
   ```

3. **Monitor Creation Failing**
   ```bash
   kubectl describe realtimemonitor <monitor-name>
   ```

### Debug Mode

The provider runs with debug logging enabled. To view debug logs:

```bash
kubectl logs -n crossplane-system -l app=provider-komodor --tail=100 -f
```

## Support

For issues and questions:

1. Check the provider logs for error messages
2. Verify your Komodor API key has the correct permissions
3. Ensure your Kubernetes cluster has sufficient resources
4. Review the Crossplane documentation for general provider issues

## Version Information

- **Provider Version**: v0.0.0-2.g10b4b71.dirty
- **Crossplane Runtime**: v1.20.0
- **Kubernetes**: v1.24.0+
- **Go**: 1.23.8 