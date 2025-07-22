# 🔧 Troubleshooting Guide

## Common Issues and Solutions

### ❌ **Error: "cannot initialize parser backend: couldn't find "package.yaml" file after checking 1428 files in the archive (annotated layer: false): EOF"**

**Problem**: Crossplane is looking for a `package.yaml` file in the Docker image, but our current setup doesn't include the proper Crossplane package structure.

**Error Message**:
```
cannot initialize parser backend: couldn't find "package.yaml" file after checking 1428 files in the archive (annotated layer: false): EOF
```

**Solution**: We now have a proper Crossplane package (`.xpkg` file) that should be used instead of the Docker image directly.

**Fixed Installation**:
```bash
# Use the proper Crossplane package
kubectl crossplane install provider package/crossplane-komodor-47101aff7daa.xpkg
```

**Or use the installation script**:
```bash
./scripts/install-provider.sh
```

### ❌ **Error: "DENIED: requested access to the resource is denied"**

**Problem**: Crossplane is trying to pull from GitHub Container Registry (ghcr.io) instead of Docker Hub.

**Error Message**:
```
Warning  UnpackPackage  2s (x5 over 15s)  packages/provider.pkg.crossplane.io  cannot unpack package: failed to fetch package digest from remote: failed to fetch package descriptor with a GET request after a previous HEAD request failure: GET https://ghcr.io/token?scope=repository%!A(MISSING)danielyaba%!F(MISSING)crossplane-komodor%!A(MISSING)pull&service=ghcr.io: DENIED: requested access to the resource is denied
```

**Solution**: The package configuration now explicitly uses Docker Hub with `docker.io/` prefix.

**Fixed Configuration**:
```yaml
spec:
  package: "docker.io/danielyaba/crossplane-komodor:v1.0.0"
```

### ❌ **Error: "ImagePullBackOff" or "ErrImagePull"**

**Problem**: Docker image doesn't exist or can't be pulled.

**Solutions**:
1. **Check if image exists**:
   ```bash
   docker pull docker.io/danielyaba/crossplane-komodor:v1.0.0
   ```

2. **Build and push the image**:
   ```bash
   # Use the packaging script
   ./scripts/package-provider.sh -u danielyaba -v v1.0.0
   ```

3. **Check Docker Hub login**:
   ```bash
   docker login
   ```

### ❌ **Error: "no matches for kind 'Provider'"**

**Problem**: Crossplane is not installed or the package schema is incorrect.

**Solutions**:
1. **Install Crossplane first**:
   ```bash
   kubectl create namespace crossplane-system
   helm repo add crossplane-stable https://charts.crossplane.io/stable
   helm repo update
   helm install crossplane crossplane-stable/crossplane --namespace crossplane-system
   ```

2. **Wait for Crossplane to be ready**:
   ```bash
   kubectl wait --for=condition=Available deployment/crossplane --timeout=5m -n crossplane-system
   ```

### ❌ **Error: "failed to create provider"**

**Problem**: RBAC or namespace issues.

**Solutions**:
1. **Check if crossplane-system namespace exists**:
   ```bash
   kubectl get namespace crossplane-system
   ```

2. **Create namespace if missing**:
   ```bash
   kubectl create namespace crossplane-system
   ```

3. **Check RBAC resources**:
   ```bash
   kubectl get clusterrole crossplane-komodor
   kubectl get clusterrolebinding crossplane-komodor
   ```

### ❌ **Error: "ProviderConfig not found"**

**Problem**: Provider configuration is missing or incorrect.

**Solutions**:
1. **Create provider configuration**:
   ```bash
   kubectl apply -f examples/production/providerconfig.yaml
   ```

2. **Check provider configuration**:
   ```bash
   kubectl get providerconfig
   kubectl describe providerconfig
   ```

### ❌ **Error: "API key not found"**

**Problem**: Komodor API key is not configured.

**Solutions**:
1. **Update providerconfig.yaml with your API key**:
   ```yaml
   apiVersion: komodor.crossplane.io/v1alpha1
   kind: ProviderConfig
   metadata:
     name: komodor-provider-config
   spec:
     credentials:
       source: Secret
       secretRef:
         name: komodor-api-key
         namespace: crossplane-system
         key: api-key
   ```

2. **Create the secret**:
   ```bash
   kubectl create secret generic komodor-api-key \
     --from-literal=api-key="YOUR_BASE64_ENCODED_API_KEY" \
     -n crossplane-system
   ```

## 🔍 **Diagnostic Commands**

### **Check Provider Status**
```bash
# Check provider installation
kubectl get providers

# Check provider details
kubectl describe provider crossplane-komodor

# Check provider logs
kubectl logs -n crossplane-system -l app=crossplane-komodor
```

### **Check CRDs**
```bash
# List all Komodor CRDs
kubectl get crd | grep komodor

# Check specific CRD
kubectl describe crd realtimemonitors.komodor.komodor.crossplane.io
```

### **Check RBAC**
```bash
# Check ClusterRole
kubectl get clusterrole crossplane-komodor

# Check ClusterRoleBinding
kubectl get clusterrolebinding crossplane-komodor

# Check ServiceAccount
kubectl get serviceaccount crossplane -n crossplane-system
```

### **Check Resources**
```bash
# Check RealtimeMonitor resources
kubectl get realtimemonitors

# Check ProviderConfig resources
kubectl get providerconfig

# Check events
kubectl get events --sort-by='.lastTimestamp'
```

## 🚀 **Complete Reset and Reinstall**

If you need to completely reset and reinstall:

```bash
# 1. Delete all Komodor resources
kubectl delete realtimemonitors --all
kubectl delete providerconfig --all

# 2. Delete the provider
kubectl delete provider crossplane-komodor

# 3. Delete CRDs
kubectl delete crd realtimemonitors.komodor.komodor.crossplane.io
kubectl delete crd providerconfigs.komodor.crossplane.io
kubectl delete crd providerconfigusages.komodor.crossplane.io
kubectl delete crd storeconfigs.komodor.crossplane.io

# 4. Delete RBAC
kubectl delete clusterrole crossplane-komodor
kubectl delete clusterrolebinding crossplane-komodor

# 5. Reinstall everything
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
```

## 📞 **Getting Help**

If you're still experiencing issues:

1. **Check the logs**:
   ```bash
   kubectl logs -n crossplane-system -l app=crossplane-komodor --tail=100
   ```

2. **Check Crossplane logs**:
   ```bash
   kubectl logs -n crossplane-system deployment/crossplane --tail=100
   ```

3. **Open an issue** on GitHub with:
   - Error messages
   - Kubernetes version
   - Crossplane version
   - Steps to reproduce 