#!/bin/bash

# Multi-Platform Provider Build Script
# This script builds the provider for both linux/amd64 and linux/arm64

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_USERNAME="danielyaba"
PROVIDER_NAME="crossplane-komodor"
VERSION="v1.0.0"
PLATFORMS="linux/amd64,linux/arm64"

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

print_status "Starting multi-platform provider build..."
print_status "Docker Username: $DOCKER_USERNAME"
print_status "Provider Name: $PROVIDER_NAME"
print_status "Version: $VERSION"
print_status "Platforms: $PLATFORMS"

# Check required commands
print_status "Checking required commands..."
check_command "docker"
check_command "make"

# Set image name
IMAGE_NAME="docker.io/${DOCKER_USERNAME}/${PROVIDER_NAME}:${VERSION}"

# Step 1: Build Go binaries for all platforms
print_status "Step 1: Building Go binaries for all platforms..."
IFS=',' read -ra PLATFORM_ARRAY <<< "$PLATFORMS"
for platform in "${PLATFORM_ARRAY[@]}"; do
    print_status "Building for platform: $platform"
    # Convert platform format (linux/amd64 -> linux_amd64)
    platform_make=$(echo $platform | sed 's/\//_/g')
    make build.code.platform PLATFORM=$platform_make
done
print_success "Go binaries built for all platforms"

# Step 2: Prepare binaries for Docker build
print_status "Step 2: Preparing binaries for Docker build..."
mkdir -p bin/linux_arm64 bin/linux_amd64
cp _output/bin/linux_arm64/provider bin/linux_arm64/provider 2>/dev/null || true
cp _output/bin/linux_amd64/provider bin/linux_amd64/provider 2>/dev/null || true
print_success "Binaries prepared for Docker build"

# Step 3: Check Docker Buildx
print_status "Step 3: Checking Docker Buildx..."
if ! docker buildx version &> /dev/null; then
    print_error "Docker Buildx is not available. Please enable it:"
    print_error "  docker buildx create --use"
    exit 1
fi

# Step 4: Build and push multi-platform image
print_status "Step 4: Building and pushing multi-platform Docker image..."
print_warning "Make sure you're logged in to Docker Hub: docker login"

docker buildx build \
    --platform $PLATFORMS \
    --no-cache \
    -t $IMAGE_NAME \
    -f cluster/images/provider-komodor/Dockerfile \
    . \
    --push

print_success "Multi-platform Docker image built and pushed successfully"

# Step 5: Verify the image
print_status "Step 5: Verifying multi-platform image..."
docker buildx imagetools inspect $IMAGE_NAME

print_success "🎉 Multi-platform provider build completed successfully!"
echo ""
print_status "Summary:"
echo "  ✅ Go binaries built for: $PLATFORMS"
echo "  ✅ Multi-platform Docker image: $IMAGE_NAME"
echo "  ✅ Image pushed to Docker Hub"
echo ""
print_status "The image now supports both linux/amd64 and linux/arm64 platforms!"
echo ""
print_status "You can now install the provider:"
echo "  kubectl apply -f https://raw.githubusercontent.com/danielyaba/crossplane-komodor/main/package/crossplane.yaml" 