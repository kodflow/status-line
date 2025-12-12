#!/bin/bash
set -e

echo "========================================="
echo "Installing C/C++ Development Environment"
echo "========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Install C/C++ toolchain
echo -e "${YELLOW}Installing C/C++ toolchain...${NC}"
sudo apt-get update && sudo apt-get install -y \
    build-essential \
    gcc \
    g++ \
    gdb \
    clang \
    make \
    cmake \
    git \
    curl

GCC_VERSION=$(gcc --version | head -n 1)
CLANG_VERSION=$(clang --version | head -n 1)
CMAKE_VERSION=$(cmake --version | head -n 1)
MAKE_VERSION=$(make --version | head -n 1)

echo -e "${GREEN}✓ ${GCC_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${CLANG_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${CMAKE_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${MAKE_VERSION} installed${NC}"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}C/C++ environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - ${GCC_VERSION}"
echo "  - ${CLANG_VERSION}"
echo "  - ${CMAKE_VERSION}"
echo "  - ${MAKE_VERSION}"
echo "  - gdb (debugger)"
echo ""
