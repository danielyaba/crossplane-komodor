# Komodor Provider Examples

This directory contains example manifests for the Crossplane Komodor Provider.

## Example Structure

### Real-world Monitor Example

The `realtimemonitor.yaml` contains a real-world example based on an actual Komodor Real Time Monitor:

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
      - cluster: "my-cluster"
        exclude: {}
        labels:
          - "team:Mars"
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

### Minimal Monitor Example

A simplified example with only required fields:

```yaml
apiVersion: komodor.komodor.crossplane.io/v1alpha1
kind: RealtimeMonitor
metadata:
  name: minimal-monitor
spec:
  providerConfigRef:
    name: komodor-provider
  forProvider:
    name: Minimal Monitor
    active: true
    type: availability
    sensors:
      - cluster: "my-cluster"
        labels:
          - "app:myapp"
    sinks:
      slack:
        - "alerts"
```

## Schema Compatibility

### Flexible JSON Fields

The provider uses `x-kubernetes-preserve-unknown-fields: true` for complex objects, allowing:

- **Native YAML structures** in manifests (no need for JSON strings)
- **Automatic conversion** to `apiextensionsv1.JSON` internally
- **Full compatibility** with Komodor API structures

### Supported Field Types

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Monitor name |
| `sensors` | array | ✅ | Sensor configurations (flexible JSON) |
| `sinks` | object | ✅ | Sink configurations (flexible JSON) |
| `active` | boolean | ✅ | Whether monitor is active (defaults to true) |
| `type` | string | ✅ | Monitor type (e.g., "availability") |
| `variables` | object | ❌ | Monitor variables (flexible JSON) |
| `sinksOptions` | map[string][]string | ❌ | Sink notification options |

### Sensor Configuration Examples

```yaml
# Kubernetes cluster sensor
sensors:
  - cluster: "my-cluster"
    labels:
      - "app:myapp"
      - "env:prod"

# With exclude filters
sensors:
  - cluster: "my-cluster"
    exclude:
      namespaces:
        - "kube-system"
        - "monitoring"
    labels:
      - "team:backend"
```

### Sink Configuration Examples

```yaml
# Slack sink
sinks:
  slack:
    - "alerts"
    - "oncall"

# Multiple sink types
sinks:
  slack:
    - "alerts"
  webhook:
    - "https://hooks.slack.com/services/xxx/yyy/zzz"
```

### Variables Configuration Examples

```yaml
# Availability monitor variables
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

# Custom monitor variables
variables:
  threshold: 70
  timeout: 60
  retries: 3
```

## Usage Instructions

### Prerequisites

1. **Provider installed** and running
2. **ProviderConfig** configured with API key
3. **API key secret** created in crossplane-system namespace

### Apply Examples

```bash
# Apply real-world example
kubectl apply -f examples/sample/realtimemonitor.yaml

# Check status
kubectl get realtimemonitors
kubectl describe realtimemonitor my-app-monitor

# Check provider logs
kubectl logs deployment/provider-komodor -n crossplane-system
```

### Validation

The examples are validated against the CRD schema and should work without modification. Key validation points:

- ✅ **YAML syntax** is correct
- ✅ **Required fields** are present
- ✅ **Field types** match schema expectations
- ✅ **JSON structures** are properly formatted
- ✅ **ProviderConfig reference** is valid

## Troubleshooting Examples

### Common Issues

1. **ProviderConfig not found**
   ```bash
   # Check if ProviderConfig exists
   kubectl get providerconfigs
   
   # Create if missing
   kubectl apply -f examples/provider/config.yaml
   ```

2. **API key authentication failed**
   ```bash
   # Check secret exists
   kubectl get secret komodor-api-key -n crossplane-system
   
   # Verify secret key name
   kubectl get secret komodor-api-key -n crossplane-system -o jsonpath='{.data.apiKey}' | base64 --decode
   ```

3. **Schema validation errors**
   ```bash
   # Validate YAML syntax
   kubectl apply --dry-run=client -f examples/sample/realtimemonitor.yaml
   ```

### Debug Commands

```bash
# Check monitor status
kubectl get realtimemonitors -o wide

# View detailed status
kubectl describe realtimemonitor <name>

# Check reconciliation events
kubectl get events --field-selector involvedObject.name=<monitor-name>

# View provider logs
kubectl logs deployment/provider-komodor -n crossplane-system -f
```

## Customization

### Modifying Examples

1. **Change monitor name** and metadata
2. **Update cluster names** in sensors
3. **Modify Slack channels** in sinks
4. **Adjust variables** for your use case
5. **Add/remove categories** as needed

### Best Practices

- **Use descriptive names** for monitors
- **Include relevant labels** for filtering
- **Set appropriate thresholds** in variables
- **Test with minimal configuration** first
- **Monitor provider logs** during creation

## Related Files

- `examples/provider/config.yaml` - ProviderConfig example
- `examples/sample/realtimemonitor.yaml` - Monitor examples
- `package/crds/` - CRD definitions
- `internal/controller/` - Controller implementation 