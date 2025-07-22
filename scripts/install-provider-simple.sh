#!/bin/bash
# Simple Komodor Provider Installation Script (No Crossplane CLI required)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_status "🚀 Installing Komodor Provider (Simple Method)..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is required but not installed."
    exit 1
fi

# Step 1: Apply CRDs
print_status "Step 1: Applying CRDs..."
kubectl apply -f package/crds/
print_success "CRDs applied"

# Step 2: Apply RBAC resources
print_status "Step 2: Applying RBAC resources..."
if [ -f "examples/production/rbac.yaml" ]; then
    kubectl apply -f examples/production/rbac.yaml
    print_success "RBAC resources applied"
else
    print_warning "RBAC file not found, skipping RBAC setup"
fi

# Step 3: Deploy Provider Controller
print_status "Step 3: Deploying provider controller..."
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
print_success "Provider controller deployed"

# Step 4: Create provider configuration
print_status "Step 4: Creating provider configuration..."
if [ -f "examples/production/providerconfig.yaml" ]; then
    kubectl apply -f examples/production/providerconfig.yaml
    print_success "Provider configuration created"
else
    print_warning "Provider config file not found, skipping provider config setup"
fi

# Step 5: Wait for deployment to be ready
print_status "Step 5: Waiting for deployment to be ready..."
kubectl wait --for=condition=Available deployment/crossplane-komodor -n crossplane-system --timeout=5m
print_success "Provider deployment is ready"

# Step 6: Verification
print_status "Step 6: Verifying installation..."

echo ""
print_status "Checking CRDs:"
kubectl get crd | grep komodor

echo ""
print_status "Checking RBAC:"
kubectl get clusterrole crossplane-komodor 2>/dev/null || echo "ClusterRole not found"
kubectl get clusterrolebinding crossplane-komodor 2>/dev/null || echo "ClusterRoleBinding not found"

echo ""
print_status "Checking provider pods:"
kubectl get pods -n crossplane-system | grep crossplane-komodor

print_success "🎉 Installation completed successfully!"
echo ""
print_status "Next steps:"
echo "  1. Update the API key in the provider configuration:"
echo "     kubectl edit providerconfig komodor-provider-config"
echo ""
echo "  2. Create a sample monitor:"
echo "     kubectl apply -f examples/production/realtimemonitor.yaml"
echo ""
echo "  3. Check monitor status:"
echo "     kubectl get realtimemonitors"
echo ""
print_status "For troubleshooting, see: TROUBLESHOOTING.md" 