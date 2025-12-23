#!/bin/bash
set -e

echo "========================================="
echo "Installing Dart/Flutter Development Environment"
echo "========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Environment variables
export FLUTTER_ROOT="${FLUTTER_ROOT:-/home/vscode/.cache/flutter}"
export PUB_CACHE="${PUB_CACHE:-/home/vscode/.cache/pub-cache}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    git \
    unzip \
    xz-utils \
    zip

# Install Flutter (includes Dart)
echo -e "${YELLOW}Installing Flutter...${NC}"

# Clone with retry
MAX_RETRIES=3
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if git clone https://github.com/flutter/flutter.git -b stable "$FLUTTER_ROOT"; then
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
        echo -e "${YELLOW}Git clone failed, retrying (attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)...${NC}"
        rm -rf "$FLUTTER_ROOT"
        sleep 5
    else
        echo -e "${RED}Failed to clone Flutter repository after $MAX_RETRIES attempts${NC}"
        exit 1
    fi
done

# Setup Flutter
export PATH="$FLUTTER_ROOT/bin:$PATH"

# Run flutter doctor to download dependencies
flutter doctor

FLUTTER_VERSION=$(flutter --version | head -n 1)
DART_VERSION=$(dart --version 2>&1)
echo -e "${GREEN}✓ ${FLUTTER_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${DART_VERSION} installed${NC}"

# Create cache directories
mkdir -p "$PUB_CACHE"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Dart/Flutter environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - ${FLUTTER_VERSION}"
echo "  - ${DART_VERSION}"
echo "  - Pub (package manager)"
echo ""
echo "Cache directories:"
echo "  - Flutter: $FLUTTER_ROOT"
echo "  - Pub cache: $PUB_CACHE"
echo ""
