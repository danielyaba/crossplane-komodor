#!/bin/bash

# Komodor Provider Packaging Script
# This script automates the process of building and packaging the provider

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_USERNAME=""
PROVIDER_NAME="provider-komodor"
VERSION="latest"
PLATFORM="linux_arm64,linux_amd64"
MULTI_PLATFORM=true

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
        print_error "$1 is required but not installed. Please install it first."
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -u, --username USERNAME    Docker Hub username (required)"
    echo "  -v, --version VERSION      Version tag (default: latest)"
    echo "  -p, --platform PLATFORM    Target platform (default: linux/arm64,linux/amd64)"
    echo "  -h, --help                 Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -u myusername"
    echo "  $0 -u myusername -v v1.0.0"
    echo "  $0 -u myusername -v v1.0.0 -p linux/amd64"
    echo "  $0 -u myusername -v v1.0.0 -p linux/arm64,linux/amd64"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--username)
            DOCKER_USERNAME="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -h|--help)
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

# Validate required arguments
if [ -z "$DOCKER_USERNAME" ]; then
    print_error "Docker Hub username is required. Use -u or --username option."
    show_usage
    exit 1
fi

# Check required commands
print_status "Checking required commands..."
check_command "docker"
check_command "make"
check_command "kubectl"

# Set image name
IMAGE_NAME="docker.io/${DOCKER_USERNAME}/${PROVIDER_NAME}:${VERSION}"

print_status "Starting provider packaging process..."
print_status "Docker Username: $DOCKER_USERNAME"
print_status "Provider Name: $PROVIDER_NAME"
print_status "Version: $VERSION"
print_status "Platform: $PLATFORM"
print_status "Image Name: $IMAGE_NAME"

# Step 1: Build multi-platform Go binaries
print_status "Step 1: Building multi-platform Go binaries..."
IFS=',' read -ra PLATFORMS <<< "$PLATFORM"
for platform in "${PLATFORMS[@]}"; do
    print_status "Building for platform: $platform"
    make build.code.platform PLATFORM=$platform
done
print_success "Multi-platform Go binaries built successfully"

# Step 2: Prepare binaries for Docker build
print_status "Step 2: Preparing binaries for Docker build..."
mkdir -p bin/linux_arm64 bin/linux_amd64
cp _output/bin/linux_arm64/provider bin/linux_arm64/provider 2>/dev/null || true
cp _output/bin/linux_amd64/provider bin/linux_amd64/provider 2>/dev/null || true
print_success "Binaries prepared for Docker build"

# Step 3: Build multi-platform Docker image
print_status "Step 3: Building multi-platform Docker image..."
print_warning "Make sure you have Docker Buildx enabled: docker buildx create --use"

# Convert platform format for Docker buildx (linux_arm64 -> linux/arm64)
DOCKER_PLATFORMS=$(echo $PLATFORM | sed 's/linux_arm64/linux\/arm64/g' | sed 's/linux_amd64/linux\/amd64/g')
docker buildx build --platform $DOCKER_PLATFORMS --no-cache -t $IMAGE_NAME -f cluster/images/provider-komodor/Dockerfile . --push
print_success "Multi-platform Docker image built and pushed successfully"

# Step 5: Update package configuration
print_status "Step 5: Updating package configuration..."
PACKAGE_FILE="package/crossplane.yaml"

# Create backup
cp $PACKAGE_FILE ${PACKAGE_FILE}.backup

# Update the package configuration
cat > $PACKAGE_FILE << EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: crossplane-komodor
  annotations:
    meta.crossplane.io/maintainer: "$DOCKER_USERNAME <$DOCKER_USERNAME@example.com>"
    meta.crossplane.io/source: "github.com/$DOCKER_USERNAME/crossplane-komodor"
    meta.crossplane.io/license: "Apache-2.0"
    meta.crossplane.io/description: |
      A Crossplane provider for managing Komodor Real Time Monitors.
      This provider allows you to create, update, and delete monitors
      in Komodor using Kubernetes Custom Resources.
spec:
  package: "$IMAGE_NAME"
  packagePullPolicy: IfNotPresent
  revisionActivationPolicy: Automatic
  revisionHistoryLimit: 1
EOF

print_success "Package configuration updated"

# Step 6: Generate installation instructions
print_status "Step 6: Generating installation instructions..."

cat > INSTALLATION_INSTRUCTIONS.md << EOF
# 🚀 Komodor Provider Installation

## Quick Install

Install the Komodor provider with a single command:

\`\`\`bash
kubectl apply -f https://raw.githubusercontent.com/$DOCKER_USERNAME/crossplane-komodor/main/package/crossplane.yaml
\`\`\`

## Complete Setup

1. **Install the provider**:
   \`\`\`bash
   kubectl apply -f https://raw.githubusercontent.com/$DOCKER_USERNAME/crossplane-komodor/main/package/crossplane.yaml
   \`\`\`

2. **Create provider configuration**:
   \`\`\`bash
   kubectl apply -f https://raw.githubusercontent.com/$DOCKER_USERNAME/crossplane-komodor/main/examples/production/providerconfig.yaml
   \`\`\`

3. **Create a monitor**:
   \`\`\`bash
   kubectl apply -f https://raw.githubusercontent.com/$DOCKER_USERNAME/crossplane-komodor/main/examples/production/realtimemonitor.yaml
   \`\`\`

## Verify Installation

\`\`\`bash
# Check provider status
kubectl get providers

# Check CRDs
kubectl get crd | grep komodor

# Check monitor resources
kubectl get realtimemonitors
\`\`\`

## Image Details

- **Image**: $IMAGE_NAME
- **Platform**: $PLATFORM
- **Version**: $VERSION
EOF

print_success "Installation instructions generated"

# Summary
print_success "🎉 Provider packaging completed successfully!"
echo ""
print_status "Summary:"
echo "  ✅ Go binary built for $PLATFORM"
echo "  ✅ Docker image built: $IMAGE_NAME"
echo "  ✅ Image pushed to Docker Hub"
echo "  ✅ Package configuration updated"
echo "  ✅ Installation instructions generated"
echo ""
print_status "Next steps:"
echo "  1. Commit and push your changes to GitHub"
echo "  2. Share the installation command with users:"
echo "     kubectl apply -f https://raw.githubusercontent.com/$DOCKER_USERNAME/crossplane-komodor/main/package/crossplane.yaml"
echo "  3. Consider publishing to Crossplane registry for wider distribution"
echo ""
print_status "Files created/modified:"
echo "  - package/crossplane.yaml (updated)"
echo "  - INSTALLATION_INSTRUCTIONS.md (new)"
echo "  - package/crossplane.yaml.backup (backup)" 