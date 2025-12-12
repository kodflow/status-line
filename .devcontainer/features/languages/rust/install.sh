#!/bin/bash
set -e

echo "========================================="
echo "Installing Rust Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export CARGO_HOME="${CARGO_HOME:-$HOME/.cache/cargo}"
export RUSTUP_HOME="${RUSTUP_HOME:-$HOME/.cache/rustup}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    build-essential \
    gcc \
    make \
    cmake \
    pkg-config \
    libssl-dev

# Install rustup (Rust toolchain installer)
echo -e "${YELLOW}Installing rustup...${NC}"
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path

# Setup Rust environment
export PATH="$CARGO_HOME/bin:$PATH"

# Source cargo env
source "$CARGO_HOME/env"

RUST_VERSION=$(rustc --version)
CARGO_VERSION=$(cargo --version)
echo -e "${GREEN}✓ ${RUST_VERSION} installed${NC}"
echo -e "${GREEN}✓ ${CARGO_VERSION} installed${NC}"

# Install stable toolchain
echo -e "${YELLOW}Installing stable toolchain...${NC}"
rustup toolchain install stable
rustup default stable
echo -e "${GREEN}✓ Stable toolchain installed${NC}"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Rust environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - rustup (Rust toolchain manager)"
echo "  - ${RUST_VERSION}"
echo "  - ${CARGO_VERSION}"
echo ""
echo "Cache directories:"
echo "  - CARGO_HOME: $CARGO_HOME"
echo "  - RUSTUP_HOME: $RUSTUP_HOME"
echo ""
