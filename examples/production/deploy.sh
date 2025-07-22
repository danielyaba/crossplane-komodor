#!/bin/bash

# Komodor Crossplane Provider - Production Deployment Script
# This script automates the deployment of the Komodor provider to production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROVIDER_IMAGE="build-448a192b/provider-komodor-arm64:latest"
NAMESPACE="crossplane-system"
SECRET_NAME="komodor-api-secret"

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
        print_error "$1 is not installed. Please install it first."
        exit 1
    fi
}

# Function to check if kubectl is configured
check_kubectl() {
    if ! kubectl cluster-info &> /dev/null; then
        print_error "kubectl is not configured or cluster is not accessible."
        exit 1
    fi
}

# Function to check if namespace exists
check_namespace() {
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        print_error "Namespace $NAMESPACE does not exist. Please install Crossplane first."
        exit 1
    fi
}

# Function to build the provider
build_provider() {
    print_status "Building the provider..."
    
    # Check if we're in the right directory
    if [ ! -f "Makefile" ]; then
        print_error "Makefile not found. Please run this script from the project root."
        exit 1
    fi
    
    # Build the Go binary
    print_status "Building Go binary..."
    make build.code.platform PLATFORM=linux_arm64
    
    # Copy binary to expected location
    print_status "Copying binary to expected location..."
    mkdir -p bin/linux_arm64
    cp _output/bin/linux_arm64/provider bin/linux_arm64/provider
    
    # Build Docker image
    print_status "Building Docker image..."
    docker build --no-cache -t $PROVIDER_IMAGE -f cluster/images/provider-komodor/Dockerfile .
    
    print_success "Provider built successfully!"
}

# Function to deploy to cluster
deploy_to_cluster() {
    print_status "Deploying to cluster..."
    
    # Check if using kind
    if kubectl config current-context | grep -q "kind"; then
        print_status "Detected kind cluster, loading image..."
        kind load docker-image $PROVIDER_IMAGE --name $(kubectl config current-context | sed 's/kind-//')
    else
        print_warning "Not using kind. Please ensure the image is available in your cluster's registry."
        print_status "You may need to push the image to your registry:"
        echo "  docker tag $PROVIDER_IMAGE your-registry/provider-komodor:latest"
        echo "  docker push your-registry/provider-komodor:latest"
        echo "  Then update the image in provider-deployment.yaml"
    fi
    
    print_success "Deployment preparation completed!"
}

# Function to create API key secret
create_api_secret() {
    print_status "Setting up API key secret..."
    
    if [ -z "$KOMODOR_API_KEY" ]; then
        print_error "KOMODOR_API_KEY environment variable is not set."
        print_status "Please set it with: export KOMODOR_API_KEY='your-api-key'"
        exit 1
    fi
    
    # Create the secret
    kubectl create secret generic $SECRET_NAME \
        --from-literal=api-key="$KOMODOR_API_KEY" \
        -n $NAMESPACE \
        --dry-run=client -o yaml | kubectl apply -f -
    
    print_success "API key secret created!"
}

# Function to deploy the provider
deploy_provider() {
    print_status "Deploying the provider..."
    
    # Apply the provider deployment
    kubectl apply -f examples/production/provider-deployment.yaml
    
    # Apply the provider configuration
    kubectl apply -f examples/production/providerconfig.yaml
    
    # Wait for the provider to be ready
    print_status "Waiting for provider to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/provider-komodor -n $NAMESPACE
    
    print_success "Provider deployed successfully!"
}

# Function to verify deployment
verify_deployment() {
    print_status "Verifying deployment..."
    
    # Check if pod is running
    if kubectl get pods -n $NAMESPACE -l app=provider-komodor --no-headers | grep -q "Running"; then
        print_success "Provider pod is running!"
    else
        print_error "Provider pod is not running. Check logs with:"
        echo "  kubectl logs -n $NAMESPACE -l app=provider-komodor"
        exit 1
    fi
    
    # Check if CRDs are installed
    if kubectl get crd realtimemonitors.komodor.komodor.crossplane.io &> /dev/null; then
        print_success "CRDs are installed!"
    else
        print_error "CRDs are not installed. Check if Crossplane is properly configured."
        exit 1
    fi
}

# Function to create example monitor
create_example_monitor() {
    print_status "Creating example monitor..."
    
    kubectl apply -f examples/production/realtimemonitor.yaml
    
    print_success "Example monitor created!"
    print_status "You can check its status with:"
    echo "  kubectl get realtimemonitor production-app-monitor"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --build-only        Only build the provider, don't deploy"
    echo "  --deploy-only       Only deploy, don't build (assumes image exists)"
    echo "  --skip-example      Skip creating the example monitor"
    echo "  --help              Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  KOMODOR_API_KEY     Your Komodor API key (required for deployment)"
    echo ""
    echo "Examples:"
    echo "  $0                    # Full build and deploy"
    echo "  $0 --build-only       # Only build"
    echo "  $0 --deploy-only      # Only deploy"
}

# Main script
main() {
    print_status "Starting Komodor Crossplane Provider deployment..."
    
    # Parse command line arguments
    BUILD_ONLY=false
    DEPLOY_ONLY=false
    SKIP_EXAMPLE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --build-only)
                BUILD_ONLY=true
                shift
                ;;
            --deploy-only)
                DEPLOY_ONLY=true
                shift
                ;;
            --skip-example)
                SKIP_EXAMPLE=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Check prerequisites
    check_command kubectl
    check_command docker
    check_command make
    
    if [ "$DEPLOY_ONLY" = false ]; then
        check_kubectl
        check_namespace
    fi
    
    # Build phase
    if [ "$DEPLOY_ONLY" = false ]; then
        build_provider
        deploy_to_cluster
    fi
    
    # Deploy phase
    if [ "$BUILD_ONLY" = false ]; then
        create_api_secret
        deploy_provider
        verify_deployment
        
        if [ "$SKIP_EXAMPLE" = false ]; then
            create_example_monitor
        fi
    fi
    
    print_success "Deployment completed successfully!"
    print_status "Next steps:"
    echo "  1. Check provider logs: kubectl logs -n $NAMESPACE -l app=provider-komodor"
    echo "  2. Create your own monitors using the RealtimeMonitor CRD"
    echo "  3. Monitor the provider metrics on port 8080"
}

# Run main function
main "$@" 