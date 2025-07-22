# 🚀 Quick Start Guide

## ✅ **Recommended Installation Method**

Since the Crossplane CLI can be tricky to install, we recommend using the **Simple Installation Method** that doesn't require the Crossplane CLI.

### **Step 1: Run the Simple Installation Script**

```bash
./scripts/install-provider-simple.sh
```

This script will:
- ✅ Apply all CRDs
- ✅ Set up RBAC permissions
- ✅ Deploy the provider controller
- ✅ Create provider configuration
- ✅ Verify the installation

### **Step 2: Update API Key**

Edit the provider configuration to add your Komodor API key:

```bash
kubectl edit providerconfig komodor-provider-config
```

Replace the empty API key with your base64-encoded Komodor API key.

### **Step 3: Create a Monitor**

```bash
kubectl apply -f examples/production/realtimemonitor.yaml
```

### **Step 4: Verify Everything Works**

```bash
# Check provider status
kubectl get pods -n crossplane-system | grep crossplane-komodor

# Check monitor status
kubectl get realtimemonitors

# Check monitor details
kubectl describe realtimemonitor my-app-monitor
```

## 🔧 **Alternative: Manual Installation**

If you prefer to install manually:

```bash
# 1. Apply CRDs
kubectl apply -f package/crds/

# 2. Apply RBAC
kubectl apply -f examples/production/rbac.yaml

# 3. Deploy provider controller
kubectl apply -f examples/production/provider-deployment.yaml

# 4. Create provider configuration
kubectl apply -f examples/production/providerconfig.yaml

# 5. Create a monitor
kubectl apply -f examples/production/realtimemonitor.yaml
```

## 🎯 **What Gets Installed**

- **CRDs**: Custom Resource Definitions for RealtimeMonitor
- **RBAC**: ClusterRole and ClusterRoleBinding for permissions
- **Provider Controller**: Deployment running the provider logic
- **Provider Configuration**: Configuration for API credentials

## 🔍 **Troubleshooting**

If you encounter issues:

1. **Check provider logs**:
   ```bash
   kubectl logs -n crossplane-system deployment/crossplane-komodor
   ```

2. **Check monitor status**:
   ```bash
   kubectl describe realtimemonitor my-app-monitor
   ```

3. **See full troubleshooting guide**: `TROUBLESHOOTING.md`

## 📋 **Next Steps**

After successful installation:

1. **Create more monitors** by applying different YAML files
2. **Monitor the logs** to see reconciliation in action
3. **Explore the Komodor API** to understand available options
4. **Customize monitor configurations** for your specific needs

## 🎉 **Success Indicators**

You'll know everything is working when:

- ✅ Provider pod is running: `kubectl get pods -n crossplane-system | grep crossplane-komodor`
- ✅ Monitor is created: `kubectl get realtimemonitors`
- ✅ Monitor shows "Ready" status: `kubectl describe realtimemonitor my-app-monitor`
- ✅ Monitor appears in Komodor dashboard 