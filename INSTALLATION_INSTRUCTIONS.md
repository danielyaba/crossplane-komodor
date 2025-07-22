# 🚀 Komodor Provider Installation

## Quick Install

Install the Komodor provider with a single command:

```bash
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
```

## Complete Setup

1. **Install the provider**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
   ```

2. **Create provider configuration**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/examples/production/providerconfig.yaml
   ```

3. **Create a monitor**:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/examples/production/realtimemonitor.yaml
   ```

## Verify Installation

```bash
# Check provider status
kubectl get providers

# Check CRDs
kubectl get crd | grep komodor

# Check monitor resources
kubectl get realtimemonitors
```

## Image Details

- **Image**: docker.io/danielyaba/crossplane-komodor:v1.0.0
- **Platform**: linux_arm64,linux_amd64
- **Version**: v1.0.0
