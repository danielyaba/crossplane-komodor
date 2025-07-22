# Technical Details

## Architecture Overview

The Crossplane Komodor Provider consists of several key components:

### 1. API Client (`internal/clients/komodor/client.go`)

The Komodor API client handles all HTTP communication with the Komodor API:

- **Authentication**: Uses `X-API-KEY` header for API key authentication
- **Base URL**: `https://api.komodor.com/api/v2/realtime-monitors/config`
- **CRUD Operations**: Full Create, Read, Update, Delete support for monitors
- **Error Handling**: Proper error detection including 404 Not Found responses

### 2. Controller (`internal/controller/realtimemonitor/realtimemonitor.go`)

The controller implements Crossplane's managed resource reconciliation:

- **Observe**: Checks if external resource exists and is up-to-date
- **Create**: Creates new monitors in Komodor
- **Update**: Updates existing monitors
- **Delete**: Removes monitors from Komodor

### 3. CRD Types (`apis/komodor/v1alpha1/`)

Custom Resource Definitions with flexible schema:

- **RealtimeMonitor**: Main resource type for Komodor monitors
- **ProviderConfig**: Configuration for API authentication
- **Flexible Schema**: Uses `x-kubernetes-preserve-unknown-fields: true` for complex objects

## Key Implementation Details

### Authentication Flow

1. **ProviderConfig** references a Kubernetes secret containing the API key
2. **Controller** extracts credentials using Crossplane's credential extractor
3. **Client** sets `X-API-KEY` header on all HTTP requests
4. **Error Handling** detects authentication failures and sets appropriate conditions

### Reconciliation Logic

The controller follows Crossplane's standard reconciliation pattern:

```go
func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
    // 1. Get external name (monitor ID)
    extName := meta.GetExternalName(cr)
    if extName == "" {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    // 2. Fetch monitor from Komodor
    monitor, err := c.client.GetMonitor(ctx, extName)
    if err != nil {
        return handleGetMonitorError(cr, extName, err)
    }

    // 3. Check if monitor is deleted (CRITICAL FIX)
    if monitor.IsDeleted {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }

    // 4. Compare spec with external resource
    resourceUpToDate := isMonitorUpToDate(&cr.Spec.ForProvider, monitor, ...)

    // 5. Update status and return observation
    return managed.ExternalObservation{
        ResourceExists:   true,
        ResourceUpToDate: resourceUpToDate,
    }, nil
}
```

### Critical Fix: isDeleted Field Handling

**Problem**: Monitors with `isDeleted: true` were causing recreation loops because:

1. **Monitor exists** in Komodor API (can be fetched by ID)
2. **But `isDeleted: true`** indicates it's marked for deletion
3. **Provider thought monitor existed** and was up-to-date
4. **Komodor treated it as deleted** (not functional)
5. **Recreation loop** ensued as provider couldn't properly observe the monitor

**Solution**: Added explicit check for `isDeleted` field in Observe method:

```go
// If monitor is marked as deleted, treat it as non-existent
if monitor.IsDeleted {
    return managed.ExternalObservation{ResourceExists: false}, nil
}
```

This ensures that deleted monitors are treated as non-existent, triggering proper recreation.

### Schema Flexibility

The provider uses flexible JSON fields to handle Komodor's complex monitor schema:

```go
type RealtimeMonitorParameters struct {
    Name         string                        `json:"name"`
    Sensors      []apiextensionsv1.JSON        `json:"sensors,omitempty"`
    Sinks        apiextensionsv1.JSON          `json:"sinks,omitempty"`
    Active       bool                          `json:"active"`
    Type         string                        `json:"type"`
    Variables    apiextensionsv1.JSON          `json:"variables,omitempty"`
    SinksOptions map[string][]string           `json:"sinksOptions,omitempty"`
}
```

This allows:
- **Native YAML structures** in Kubernetes manifests
- **Automatic conversion** to/from JSON for API calls
- **Full compatibility** with Komodor API structures

### Error Handling

Comprehensive error handling with proper status conditions:

```go
func handleGetMonitorError(cr *v1alpha1.RealtimeMonitor, extName string, err error) (managed.ExternalObservation, error) {
    if komodorclient.IsNotFound(err) {
        return managed.ExternalObservation{ResourceExists: false}, nil
    }
    cr.SetConditions(xpv1.ReconcileError(errors.Wrap(err, "cannot get monitor from Komodor")))
    return managed.ExternalObservation{}, errors.Wrapf(err, "failed to get monitor %q from Komodor", extName)
}
```

## API Integration

### Komodor API Endpoints

- **List Monitors**: `GET /api/v2/realtime-monitors/config`
- **Get Monitor**: `GET /api/v2/realtime-monitors/config/{id}`
- **Create Monitor**: `POST /api/v2/realtime-monitors/config`
- **Update Monitor**: `PATCH /api/v2/realtime-monitors/config/{id}`
- **Delete Monitor**: `DELETE /api/v2/realtime-monitors/config/{id}`

### Monitor Schema

```json
{
  "id": "string",
  "name": "string",
  "active": boolean,
  "type": "string",
  "sensors": [{"cluster": "string", "labels": ["string"]}],
  "sinks": {"slack": ["string"]},
  "variables": {"duration": number, "minAvailable": "string"},
  "sinksOptions": {"notifyOn": ["string"]},
  "createdAt": "string",
  "updatedAt": "string",
  "isDeleted": boolean
}
```

## Testing Strategy

### Unit Tests

- **Mock Client**: Tests controller logic with mocked API responses
- **Error Scenarios**: Tests various error conditions and edge cases
- **Schema Validation**: Tests JSON marshaling/unmarshaling

### Integration Tests

- **Real API**: Tests against actual Komodor API
- **End-to-End**: Tests full CRUD operations
- **Error Handling**: Tests authentication and API errors

## Performance Considerations

### Memory Management

- **Preallocated slices**: Avoid memory allocations in hot paths
- **JSON pooling**: Reuse JSON encoders/decoders where possible
- **Connection reuse**: HTTP client connection pooling

### Reconciliation Efficiency

- **Up-to-date logic**: Efficient comparison of spec vs external resource
- **Conditional updates**: Only update when necessary
- **Error backoff**: Exponential backoff for transient errors

## Security Considerations

### API Key Management

- **Kubernetes Secrets**: Secure storage of API keys
- **RBAC**: Proper permissions for secret access
- **Rotation**: Support for API key rotation

### Network Security

- **TLS**: All API calls use HTTPS
- **Timeouts**: Configurable request timeouts
- **Retry Logic**: Exponential backoff for transient failures

## Troubleshooting Guide

### Common Issues

1. **Authentication Errors (403)**
   - Verify API key is correct
   - Check secret key name is `apiKey`
   - Ensure no newline characters in API key

2. **Recreation Loops**
   - Check if cluster exists in Komodor
   - Verify `isDeleted` field handling
   - Review provider logs for API errors

3. **Schema Validation Errors**
   - Ensure YAML structure matches examples
   - Check for indentation issues
   - Verify required fields are present

### Debug Commands

```bash
# Check provider status
kubectl get providers

# View provider logs
kubectl logs deployment/provider-komodor -n crossplane-system

# Check monitor status
kubectl get realtimemonitors
kubectl describe realtimemonitor <name>

# Check external resource
kubectl get realtimemonitor <name> -o yaml
```

## Future Enhancements

### Planned Features

1. **Monitor Templates**: Reusable monitor configurations
2. **Bulk Operations**: Create/update multiple monitors
3. **Advanced Filtering**: Complex sensor configurations
4. **Metrics Integration**: Prometheus metrics for provider health

### API Extensions

1. **Webhook Support**: Real-time notifications
2. **Custom Sinks**: Support for custom notification channels
3. **Advanced Variables**: Complex variable expressions
4. **Monitor Dependencies**: Monitor-to-monitor relationships 