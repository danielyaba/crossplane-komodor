# Crossplane Komodor Provider

A Crossplane provider for managing Komodor Real Time Monitors via the Komodor API.

## 🚀 Status

**Ready for Testing** - The provider is fully implemented and ready for deployment and testing.

## ✨ Features

- **Full CRUD Operations**: Create, Read, Update, Delete Real Time Monitors
- **Real-time Status**: Monitor reconciliation status and external resource state
- **Flexible Configuration**: Support for complex monitor configurations with sensors, sinks, and variables
- **Secure Authentication**: API key authentication via Kubernetes secrets
- **Robust Error Handling**: Comprehensive error handling with proper status conditions
- **Schema Compatibility**: Native YAML support with automatic JSON conversion

## 🔐 Authentication

The provider uses API key authentication with the `X-API-KEY` header:

```bash
X-API-KEY: your-api-key-here
```

### Quick Setup

1. **Create API Key Secret**:
```bash
kubectl create secret generic komodor-api-key \
  --from-literal=apiKey=your-actual-api-key \
  -n crossplane-system
```

2. **Apply ProviderConfig**:
```yaml
apiVersion: komodor.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: komodor-provider
spec:
  credentials:
    source: Secret
    secretRef:
      name: komodor-api-key
      key: apiKey
```

## 📋 Resource Schema

### RealtimeMonitor

The `RealtimeMonitor` resource supports the full Komodor Real Time Monitor structure:

#### Required Fields
- `name`: Monitor name (string)
- `sensors`: Array of sensor configurations (flexible JSON structure)
- `sinks`: Sink configurations (flexible JSON structure)
- `active`: Whether monitor is active (boolean, defaults to true)
- `type`: Monitor type (string, e.g., "availability")

#### Optional Fields
- `variables`: Monitor variables (flexible JSON structure)
- `sinksOptions`: Sink notification options (map[string][]string)

## 📖 Examples

### Real-world Monitor Example

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
          - "app:my-app"
    sinks:
      slack:
        - "my-app-alert"
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

See [examples/sample/realtimemonitor.yaml](examples/sample/realtimemonitor.yaml) for more examples.

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- kind (for local testing)

### Building

```bash
# Build the provider
make build

# Run tests
make test

# Run linting
make lint
```

### Local Testing

```bash
# Create kind cluster
kind create cluster --name crossplane-test

# Install Crossplane
helm install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace

# Load provider image
kind load docker-image build-448a192b/provider-komodor-arm64:latest

# Apply provider
kubectl apply -f package/crossplane.yaml
```

## 📚 Documentation

- [Technical Details](memory-bank/docs/TECHNICAL_DETAILS.md) - Implementation details and architecture
- [Examples Guide](memory-bank/examples/README.md) - Comprehensive examples and usage
- [Provider Checklist](PROVIDER_CHECKLIST.md) - Development checklist and guidelines

## 🔧 Architecture

### Components

1. **API Client** (`internal/clients/komodor/client.go`)
   - HTTP client with X-API-KEY authentication
   - CRUD operations for monitors
   - Error handling and status checking

2. **Controller** (`internal/controller/realtimemonitor/`)
   - Managed resource reconciliation
   - Status observation and updates
   - Error handling with conditions

3. **CRD Types** (`apis/komodor/v1alpha1/`)
   - Resource definitions with validation
   - Flexible schema for complex configurations

### Key Design Decisions

- **Interface-based design** for testability
- **Helper functions** for JSON marshaling/unmarshaling
- **Comprehensive error handling** with context
- **Up-to-date logic** for efficient reconciliation
- **Memory preallocation** for performance

## 🐛 Troubleshooting

### Common Issues

1. **403 Authentication Error**
   - Verify API key is correct
   - Ensure secret key is named `apiKey`
   - Check provider logs for authentication details

2. **Schema Validation Errors**
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
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make reviewable`
6. Submit a pull request

See [CONTRIBUTING.md](memory-bank/docs/CONTRIBUTING.md) for detailed guidelines.

## 📄 License

Apache 2.0 - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Crossplane](https://crossplane.io/) for the provider framework
- [Komodor](https://komodor.com/) for the Real Time Monitors API
- The Crossplane community for guidance and best practices
