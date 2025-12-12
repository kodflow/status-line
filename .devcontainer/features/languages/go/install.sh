#!/bin/bash
set -e

echo "========================================="
echo "Installing Go Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export GO_VERSION="${GO_VERSION:-latest}"
export GOROOT="${GOROOT:-/usr/local/go}"
export GOPATH="${GOPATH:-/home/vscode/.cache/go}"
export GOCACHE="${GOCACHE:-/home/vscode/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-/home/vscode/.cache/go/pkg/mod}"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    git \
    make \
    gcc \
    build-essential

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        GO_ARCH="amd64"
        ;;
    aarch64|arm64)
        GO_ARCH="arm64"
        ;;
    armv7l)
        GO_ARCH="armv6l"
        ;;
    *)
        echo -e "${RED}✗ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Get latest Go version if requested
if [ "$GO_VERSION" = "latest" ]; then
    echo -e "${YELLOW}Fetching latest Go version...${NC}"
    GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n 1 | sed 's/go//')
fi

# Download and install Go
echo -e "${YELLOW}Installing Go ${GO_VERSION} for ${GO_ARCH}...${NC}"
GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
GO_URL="https://dl.google.com/go/${GO_TARBALL}"

# Download Go tarball
curl -fsSL --retry 3 --retry-delay 5 -o "/tmp/${GO_TARBALL}" "$GO_URL"

# Remove any existing Go installation
if [ -d "$GOROOT" ]; then
    echo -e "${YELLOW}Removing existing Go installation...${NC}"
    sudo rm -rf "$GOROOT"
fi

# Extract to /usr/local
sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"

# Clean up
rm "/tmp/${GO_TARBALL}"

GO_INSTALLED=$(go version)
echo -e "${GREEN}✓ ${GO_INSTALLED} installed${NC}"

# Create necessary directories
mkdir -p "$GOPATH/bin"
mkdir -p "$GOPATH/pkg"
mkdir -p "$GOPATH/src"
mkdir -p "$GOCACHE"
mkdir -p "$GOMODCACHE"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Go environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - ${GO_INSTALLED}"
echo "  - Go Modules (package manager)"
echo ""
echo "Cache directories:"
echo "  - GOPATH: $GOPATH"
echo "  - GOCACHE: $GOCACHE"
echo "  - GOMODCACHE: $GOMODCACHE"
echo ""
