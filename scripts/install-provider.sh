#!/bin/bash
# Komodor Provider Installation Script

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

# Function to check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 is required but not installed."
        return 1
    fi
    return 0
}

print_status "🚀 Installing Komodor Provider for Crossplane..."

# Check required commands
print_status "Checking required commands..."
if ! check_command "kubectl"; then
    print_error "kubectl is required but not installed."
    exit 1
fi

if ! check_command "kubectl crossplane"; then
    print_warning "Crossplane CLI not found. Installing..."
    curl -sL "https://cli.crossplane.io/install.sh" | sh
    # Add the binary to PATH for current session
    export PATH=$PATH:$HOME/.local/bin
    # Check if installation was successful
    if ! check_command "kubectl crossplane"; then
        print_error "Failed to install Crossplane CLI. Please install manually:"
        print_error "  curl -sL https://cli.crossplane.io/install.sh | sh"
        print_error "  export PATH=\$PATH:\$HOME/.local/bin"
        exit 1
    fi
    print_success "Crossplane CLI installed"
fi

# Check if package file exists
PACKAGE_FILE="package/crossplane-komodor-47101aff7daa.xpkg"
if [ ! -f "$PACKAGE_FILE" ]; then
    print_error "Package file not found: $PACKAGE_FILE"
    print_error "Please run the build script first: ./scripts/build-multi-platform.sh"
    exit 1
fi

# Step 1: Install the provider package
print_status "Step 1: Installing provider package..."
kubectl crossplane install provider $PACKAGE_FILE
print_success "Provider package installed"

# Step 2: Apply RBAC resources
print_status "Step 2: Applying RBAC resources..."
if [ -f "examples/production/rbac.yaml" ]; then
    kubectl apply -f examples/production/rbac.yaml
    print_success "RBAC resources applied"
else
    print_warning "RBAC file not found, skipping RBAC setup"
fi

# Step 3: Create provider configuration
print_status "Step 3: Creating provider configuration..."
if [ -f "examples/production/providerconfig.yaml" ]; then
    kubectl apply -f examples/production/providerconfig.yaml
    print_success "Provider configuration created"
else
    print_warning "Provider config file not found, skipping provider config setup"
fi

# Step 4: Wait for provider to be ready
print_status "Step 4: Waiting for provider to be ready..."
kubectl wait --for=condition=Available provider/crossplane-komodor --timeout=5m
print_success "Provider is ready"

# Step 5: Verification
print_status "Step 5: Verifying installation..."

echo ""
print_status "Checking provider status:"
kubectl get providers

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