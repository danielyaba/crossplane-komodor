# 🚀 Komodor Provider Installation Guide

## ✅ **Complete Installation with RBAC**

When you install the Komodor provider, **everything gets installed automatically** including RBAC resources!

### **Single Command Installation**

```bash
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
```

## 📋 **What Gets Installed Automatically**

### **1. Provider Controller**
- **Docker Image**: `danielyaba/crossplane-komodor:latest`
- **Deployment**: Runs in `crossplane-system` namespace
- **Controller**: Manages Komodor Real Time Monitors

### **2. Custom Resource Definitions (CRDs)**
- `realtimemonitors.komodor.komodor.crossplane.io` - Komodor monitors
- `providerconfigs.komodor.crossplane.io` - Provider configuration
- `providerconfigusages.komodor.crossplane.io` - Provider usage tracking
- `storeconfigs.komodor.crossplane.io` - Secret store configuration

### **3. RBAC Resources** ✅ **NEW!**
- **ClusterRole**: `crossplane-komodor` with permissions for:
  - Managing `RealtimeMonitor` resources
  - Managing `ProviderConfig` resources
  - Accessing secrets and events
- **ClusterRoleBinding**: Binds the ClusterRole to the Crossplane ServiceAccount in `crossplane-system` namespace

### **4. Package Configuration**
- Provider metadata and settings
- Package pull policy and revision management

## 🔧 **Complete Setup Workflow**

### **Step 1: Install Provider (Everything Included)**
```bash
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
```

### **Step 2: Create Provider Configuration**
```bash
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/examples/production/providerconfig.yaml
```

### **Step 3: Create a Monitor**
```bash
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/examples/production/realtimemonitor.yaml
```

## 🔍 **Verify Installation**

### **Check Provider Status**
```bash
# Check provider installation
kubectl get providers

# Check provider pods
kubectl get pods -n crossplane-system | grep crossplane-komodor
```

### **Check CRDs**
```bash
# Verify CRDs are installed
kubectl get crd | grep komodor
```

### **Check RBAC Resources**
```bash
# Verify RBAC resources
kubectl get clusterrole crossplane-komodor
kubectl get clusterrolebinding crossplane-komodor
kubectl get serviceaccount crossplane -n crossplane-system
```

### **Check Monitor Resources**
```bash
# List monitor resources
kubectl get realtimemonitors

# Check monitor details
kubectl describe realtimemonitor production-app-monitor
```

## 🎯 **What Users Get**

### **Before (Manual RBAC Setup Required)**
```bash
# Install provider
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml

# Manually install RBAC (separate step)
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/examples/provider/rbac.yaml
```

### **After (Everything Automatic)** ✅
```bash
# Install everything with one command
kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml
```

## 📊 **Package Contents**

```
package/
├── crossplane.yaml                    # Provider metadata
├── crds/
│   ├── komodor.crossplane.io_providerconfigs.yaml
│   ├── komodor.crossplane.io_providerconfigusages.yaml
│   ├── komodor.crossplane.io_storeconfigs.yaml
│   ├── komodor.komodor.crossplane.io_realtimemonitors.yaml
│   └── rbac.yaml                      # ✅ RBAC resources included
```

## 🚀 **Benefits of Complete Package**

1. **One-Command Installation**: Everything installed automatically
2. **No Manual RBAC Setup**: Permissions configured automatically
3. **Consistent Setup**: Same RBAC across all installations
4. **User-Friendly**: Minimal setup required
5. **Production Ready**: Proper security permissions included

## 🔒 **Security Permissions**

The included RBAC provides:
- **Read/Write access** to Komodor monitor resources
- **Read/Write access** to provider configuration
- **Access to secrets** for API key management
- **Event creation** for monitoring and debugging

---

**🎉 Result**: Users can now install your provider with a single command and get everything they need, including proper RBAC permissions! 