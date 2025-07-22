# 🔧 RBAC Service Account Update

## ✅ **Updated to Use Crossplane Service Account**

The RBAC configuration has been updated to use the **standard Crossplane service account** named `crossplane` in the `crossplane-system` namespace.

## 📋 **What Changed**

### **Service Account Reference**
```yaml
# Before
subjects:
- kind: ServiceAccount
  name: default
  namespace: crossplane-system

# After ✅
subjects:
- kind: ServiceAccount
  name: crossplane
  namespace: crossplane-system
```

## 🎯 **Why This Matters**

### **Crossplane Standard**
- **Standard Practice**: Uses the official Crossplane service account
- **Consistent Integration**: Works seamlessly with Crossplane ecosystem
- **Proper Permissions**: Leverages existing Crossplane RBAC structure
- **No Conflicts**: Avoids potential conflicts with other providers

### **Benefits**
1. **Follows Crossplane Conventions**: Uses the standard `crossplane` service account
2. **Better Integration**: Integrates properly with Crossplane infrastructure
3. **Consistent Permissions**: Inherits proper Crossplane permissions
4. **Simpler Setup**: No custom service account management needed
5. **Production Ready**: Aligns with Crossplane best practices

## 🔍 **Verification**

Users can verify the RBAC setup with:

```bash
# Check RBAC resources
kubectl get clusterrole crossplane-komodor
kubectl get clusterrolebinding crossplane-komodor
kubectl get serviceaccount crossplane -n crossplane-system

# Check provider pods are using Crossplane service account
kubectl get pods -n crossplane-system | grep crossplane-komodor
```

## 📊 **Complete RBAC Structure**

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: crossplane-komodor
  labels:
    app: crossplane-komodor
rules:
- apiGroups: ["komodor.komodor.crossplane.io"]
  resources: ["realtimemonitors"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["komodor.komodor.crossplane.io"]
  resources: ["realtimemonitors/status"]
  verbs: ["get", "update", "patch"]
- apiGroups: ["komodor.crossplane.io"]
  resources: ["providerconfigs", "providerconfigusages"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["events", "secrets"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crossplane-komodor
  labels:
    app: crossplane-komodor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: crossplane-komodor
subjects:
- kind: ServiceAccount
  name: crossplane                    # ✅ Standard Crossplane service account
  namespace: crossplane-system
```

## 🚀 **Result**

Your provider now follows **Crossplane best practices** by using the standard `crossplane` service account. This ensures:

- **Proper Integration**: Works seamlessly with Crossplane
- **Standard Permissions**: Uses established Crossplane RBAC patterns
- **Production Ready**: Follows Crossplane conventions
- **User-Friendly**: No additional service account setup required

---

**Status**: ✅ **Updated to Crossplane Standards** 