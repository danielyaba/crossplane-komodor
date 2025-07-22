# 🚀 Installation Methods for Komodor Provider

## ✅ **Method 1: Direct Package Installation (Recommended)**

Since we now have a proper Crossplane package (`.xpkg` file), you can install it directly using the Crossplane CLI:

### **Step 1: Install Crossplane CLI (if not already installed)**
```bash
# Install Crossplane CLI
curl -sL "https://cli.crossplane.io/install.sh" | sh

# Or using Homebrew on macOS
brew install crossplane/tap/crossplane
```

### **Step 2: Install the Provider Package**
```bash
# Install the provider package
kubectl crossplane install provider package/crossplane-komodor-47101aff7daa.xpkg
```

### **Step 3: Create Provider Configuration**
```bash
# Create the provider configuration
kubectl apply -f examples/production/providerconfig.yaml
```

### **Step 4: Create RBAC Resources**
```bash
# Apply RBAC resources
kubectl apply -f examples/production/rbac.yaml
```

### **Step 5: Create a Monitor**
```bash
# Create a sample monitor
kubectl apply -f examples/production/realtimemonitor.yaml
```

## 🔧 **Method 2: Manual Installation (Alternative)**

If you prefer to install components manually:

### **Step 1: Apply CRDs**
```bash
kubectl apply -f package/crds/
```

### **Step 2: Apply RBAC**
```bash
kubectl apply -f examples/production/rbac.yaml
```

### **Step 3: Deploy Provider Controller**
```bash
# Create deployment using the multi-platform image
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: crossplane-komodor
  namespace: crossplane-system
  labels:
    app: crossplane-komodor
spec:
  replicas: 1
  selector:
    matchLabels:
      app: crossplane-komodor
  template:
    metadata:
      labels:
        app: crossplane-komodor
    spec:
      serviceAccountName: crossplane
      containers:
      - name: provider
        image: docker.io/danielyaba/crossplane-komodor:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 65532
EOF
```

### **Step 4: Create Provider Configuration**
```bash
kubectl apply -f examples/production/providerconfig.yaml
```

### **Step 5: Create a Monitor**
```bash
kubectl apply -f examples/production/realtimemonitor.yaml
```

## 🎯 **Method 3: One-Command Installation Script**

Create a simple installation script:

```bash
#!/bin/bash
# install-provider.sh

set -e

echo "🚀 Installing Komodor Provider..."

# Install the package
echo "📦 Installing provider package..."
kubectl crossplane install provider package/crossplane-komodor-47101aff7daa.xpkg

# Apply RBAC
echo "🔐 Applying RBAC resources..."
kubectl apply -f examples/production/rbac.yaml

# Create provider configuration
echo "⚙️ Creating provider configuration..."
kubectl apply -f examples/production/providerconfig.yaml

echo "✅ Installation completed!"
echo ""
echo "Next steps:"
echo "1. Update the API key in the provider configuration"
echo "2. Create a monitor: kubectl apply -f examples/production/realtimemonitor.yaml"
echo "3. Check status: kubectl get realtimemonitors"
```

## 🔍 **Verification Commands**

After installation, verify everything is working:

```bash
# Check provider status
kubectl get providers

# Check CRDs
kubectl get crd | grep komodor

# Check RBAC
kubectl get clusterrole crossplane-komodor
kubectl get clusterrolebinding crossplane-komodor

# Check provider pods
kubectl get pods -n crossplane-system | grep crossplane-komodor

# Check monitor resources
kubectl get realtimemonitors
```

## 📋 **Package Details**

- **Package File**: `package/crossplane-komodor-47101aff7daa.xpkg`
- **Docker Image**: `docker.io/danielyaba/crossplane-komodor:v1.0.0`
- **Platforms**: `linux/amd64`, `linux/arm64`
- **Crossplane Version**: `>=v1.20.0`

## 🎉 **Recommended Approach**

**Use Method 1** (Direct Package Installation) as it's the most standard and reliable approach for Crossplane providers. 