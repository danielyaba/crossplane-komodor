# Crossplane Komodor Provider

A Crossplane provider for managing Komodor Real Time Monitors.

## Overview

This provider allows you to manage Komodor Real Time Monitors as Kubernetes resources using Crossplane. It provides a declarative way to create, update, and delete monitors in Komodor through Kubernetes manifests.

## Features

- **Real Time Monitor Management**: Create, update, and delete Komodor Real Time Monitors
- **Declarative Configuration**: Use Kubernetes manifests to manage monitors
- **Crossplane Integration**: Leverage Crossplane's powerful resource management capabilities
- **Flexible Schema**: Support for complex monitor configurations with sensors, sinks, and variables

## Prerequisites

- Kubernetes cluster with Crossplane installed
- Komodor account with API access
- Valid Komodor API key
- **Important**: The cluster referenced in monitor sensors must exist in your Komodor setup

## Usage

### 1. Install the Provider

```bash
# Apply CRDs
kubectl apply -f package/crds/

# Apply the provider
kubectl apply -f package/crossplane.yaml
```

### 2. Configure Authentication

Create a secret with your Komodor API key:

```bash
kubectl create secret generic komodor-api-key \
  --from-literal=apiKey=your-actual-api-key \
  -n crossplane-system
```

Create a ProviderConfig:

```yaml
apiVersion: komodor.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: komodor-provider
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: komodor-api-key
      key: apiKey
```

### 3. Create a Monitor

**Important**: Replace `YOUR_ACTUAL_CLUSTER_NAME` with a cluster that exists in your Komodor setup.

```yaml
apiVersion: komodor.komodor.crossplane.io/v1alpha1
kind: RealtimeMonitor
metadata:
  name: my-app-monitor
spec:
  providerConfigRef:
    name: komodor-provider
  forProvider:
    name: "my-app"
    active: true
    type: "availability"
    sensors:
      - cluster: "YOUR_ACTUAL_CLUSTER_NAME"  # Must exist in Komodor
        exclude: {}
        labels:
          - "app:my-app"
    sinks:
      slack:
        - "my-app-alerts"
    sinksOptions:
      notifyOn:
        - "Creating/Initializing"
        - "Scheduling"
        - "Container Creation"
        - "NonZeroExitCode"
        - "Unhealthy - failed probes"
        - "OOMKilled"
        - "Image"
        - "BackOff"
    variables:
      categories:
        - "Creating/Initializing"
        - "Scheduling"
        - "Container Creation"
        - "NonZeroExitCode"
        - "Unhealthy - failed probes"
        - "OOMKilled"
        - "Image"
        - "BackOff"
      duration: 300
      minAvailable: "85%"
```

### 4. Check Status

```bash
kubectl get realtimemonitors
kubectl describe realtimemonitor my-app-monitor
```

## Troubleshooting

### Monitor Not Visible in Komodor UI

If monitors are created but not visible in the Komodor UI:

1. **Check Cluster Existence**: Ensure the cluster referenced in the monitor's sensors exists in your Komodor setup
2. **Verify Cluster Name**: Use the exact cluster name as it appears in Komodor
3. **Check Monitor Status**: Use `kubectl describe realtimemonitor` to check for errors

### Recreation Loop

If monitors are being recreated repeatedly:

1. **Verify API Key**: Ensure the API key is correct and has proper permissions
2. **Check Cluster References**: Ensure all referenced clusters exist in Komodor
3. **Review Provider Logs**: Check provider logs for API errors

## API Reference

See the [Technical Details](TECHNICAL_DETAILS.md) for comprehensive API documentation and implementation details.
