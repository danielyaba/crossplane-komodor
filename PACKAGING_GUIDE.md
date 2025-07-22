# 📦 Komodor Provider Packaging Guide

## 🎯 **Single Provider Package Overview**

Yes! You can absolutely create a **single provider package** that includes your Docker image and push it to Docker Hub. This creates a self-contained package that users can install with a single command.

## 🚀 **How It Works**

### **1. Package Structure**
```
package/
├── crossplane.yaml          # Provider metadata and Docker image reference
├── crds/                    # Custom Resource Definitions
│   ├── komodor.crossplane.io_providerconfigs.yaml
│   ├── komodor.crossplane.io_providerconfigusages.yaml
│   ├── komodor.crossplane.io_storeconfigs.yaml
│   └── komodor.komodor.crossplane.io_realtimemonitors.yaml
└── README.md               # Package documentation
```

### **2. Package Configuration**
The `package/crossplane.yaml` file defines:
- **Provider metadata** (name, maintainer, description)
- **Docker image reference** (your Docker Hub image)
- **Package behavior** (pull policy, revision management)

## 📋 **Step-by-Step Packaging Process**

### **Step 1: Update Package Configuration**

1. **Update `package/crossplane.yaml`**:
   ```yaml
   apiVersion: pkg.crossplane.io/v1
   kind: Provider
   metadata:
     name: provider-komodor
     annotations:
       meta.crossplane.io/maintainer: "Your Name <your.email@example.com>"
       meta.crossplane.io/source: "github.com/yourusername/crossplane-komodor"
       meta.crossplane.io/license: "Apache-2.0"
       meta.crossplane.io/description: |
         A Crossplane provider for managing Komodor Real Time Monitors.
   spec:
     package: "yourusername/provider-komodor:latest"
     packagePullPolicy: IfNotPresent
     revisionActivationPolicy: Automatic
     revisionHistoryLimit: 1
   ```

2. **Replace placeholders**:
   - `yourusername` → Your Docker Hub username
   - `your.email@example.com` → Your email
   - `github.com/yourusername/crossplane-komodor` → Your GitHub repo

### **Step 2: Build and Push Docker Image**

```bash
# Build the Go binary
make build.code.platform PLATFORM=linux_arm64

# Copy binary to expected location
cp _output/bin/linux_arm64/provider bin/linux_arm64/provider

# Build Docker image with your Docker Hub username
docker build --no-cache -t yourusername/provider-komodor:latest -f cluster/images/provider-komodor/Dockerfile .

# Push to Docker Hub
docker push yourusername/provider-komodor:latest
```

### **Step 3: Build the Package (Optional)**

If you want to create a `.xpkg` file for distribution:

```bash
# Install Crossplane CLI
curl -fsSLo /tmp/crank --create-dirs "https://releases.crossplane.io/stable/latest/bin/darwin_arm64/crank?source=build"
chmod +x /tmp/crank

# Build package
/tmp/crank xpkg build --package-root package --output provider-komodor.xpkg
```

## 🎯 **User Installation Experience**

### **Single Command Installation**

Users can install your provider with just one command:

```bash
# Install the provider
kubectl apply -f https://raw.githubusercontent.com/yourusername/crossplane-komodor/main/package/crossplane.yaml
```

### **What Happens During Installation**

1. **Crossplane downloads** your Docker image from Docker Hub
2. **CRDs are installed** automatically
3. **Provider controller starts** running
4. **Users can create** `RealtimeMonitor` resources

### **User Workflow**

```bash
# 1. Install provider
kubectl apply -f https://raw.githubusercontent.com/yourusername/crossplane-komodor/main/package/crossplane.yaml

# 2. Create provider configuration
kubectl apply -f examples/production/providerconfig.yaml

# 3. Create a monitor
kubectl apply -f examples/production/realtimemonitor.yaml
```

## 🔧 **Advanced Packaging Options**

### **Multi-Platform Support**

Build for multiple architectures:

```bash
# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 -t yourusername/provider-komodor:latest --push -f cluster/images/provider-komodor/Dockerfile .
```

### **Versioned Releases**

Create versioned packages:

```bash
# Build versioned image
docker build -t yourusername/provider-komodor:v1.0.0 -f cluster/images/provider-komodor/Dockerfile .
docker push yourusername/provider-komodor:v1.0.0

# Update package to use specific version
# In package/crossplane.yaml:
spec:
  package: "yourusername/provider-komodor:v1.0.0"
```

### **Package with Examples**

Include example manifests in your package:

```bash
# Copy examples to package
cp -r examples/production package/examples/

# Update package structure
package/
├── crossplane.yaml
├── crds/
├── examples/
│   ├── providerconfig.yaml
│   ├── realtimemonitor.yaml
│   └── README.md
└── README.md
```

## 📊 **Package Distribution Options**

### **Option 1: GitHub + Docker Hub (Recommended)**
- **Package YAML**: Hosted on GitHub
- **Docker Image**: Pushed to Docker Hub
- **Installation**: `kubectl apply -f https://raw.githubusercontent.com/...`

### **Option 2: Crossplane Registry**
- **Package**: Published to Crossplane registry
- **Installation**: `kubectl crossplane install provider yourusername/provider-komodor`

### **Option 3: OCI Registry**
- **Package**: Pushed to OCI registry (GitHub Container Registry, etc.)
- **Installation**: `kubectl crossplane install provider oci://ghcr.io/yourusername/provider-komodor`

## ✅ **Benefits of Single Package Approach**

1. **Simple Installation**: One command to install everything
2. **Self-Contained**: All CRDs and controller in one package
3. **Version Management**: Easy to update and rollback
4. **Distribution**: Works with any OCI-compatible registry
5. **User Experience**: Minimal setup required

## 🚀 **Ready to Package!**

Your provider is ready to be packaged and distributed as a single Docker image. Users will be able to install it with a single command and start managing Komodor monitors immediately!

---

**Next Steps**:
1. Update the package configuration with your details
2. Build and push your Docker image
3. Share the installation command with users
4. Consider publishing to Crossplane registry for wider distribution 