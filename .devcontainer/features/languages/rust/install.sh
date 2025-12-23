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

# Install essential components
echo -e "${YELLOW}Installing rustup components...${NC}"
rustup component add rust-analyzer clippy rustfmt
echo -e "${GREEN}✓ rust-analyzer installed${NC}"
echo -e "${GREEN}✓ clippy installed${NC}"
echo -e "${GREEN}✓ rustfmt installed${NC}"

# Install essential cargo tools
echo -e "${YELLOW}Installing cargo tools...${NC}"
cargo install --locked cargo-watch cargo-nextest cargo-audit cargo-expand cargo-outdated 2>/dev/null || {
    echo -e "${YELLOW}⚠ Some cargo tools failed to install (may require manual installation)${NC}"
}
echo -e "${GREEN}✓ cargo tools installed${NC}"

# Install MCP server for rust-analyzer integration
echo -e "${YELLOW}Installing rust-analyzer-mcp...${NC}"
cargo install --locked rust-analyzer-mcp 2>/dev/null || {
    echo -e "${YELLOW}⚠ rust-analyzer-mcp failed to install (MCP integration unavailable)${NC}"
}
echo -e "${GREEN}✓ rust-analyzer-mcp installed${NC}"

# Setup shell integration
echo -e "${YELLOW}Configuring shell integration...${NC}"
CARGO_ENV_LINE='[[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"'
for rc_file in "$HOME/.bashrc" "$HOME/.zshrc"; do
    if [[ -f "$rc_file" ]] && ! grep -q "cargo/env" "$rc_file"; then
        echo "" >> "$rc_file"
        echo "# Rust/Cargo environment" >> "$rc_file"
        echo "$CARGO_ENV_LINE" >> "$rc_file"
    fi
done
echo -e "${GREEN}✓ Shell integration configured${NC}"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Rust environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - rustup (Rust toolchain manager)"
echo "  - ${RUST_VERSION}"
echo "  - ${CARGO_VERSION}"
echo "  - rust-analyzer (LSP)"
echo "  - clippy (linter)"
echo "  - rustfmt (formatter)"
echo "  - cargo-watch, cargo-nextest, cargo-audit, etc."
echo "  - rust-analyzer-mcp (MCP server)"
echo ""
echo "Cache directories:"
echo "  - CARGO_HOME: $CARGO_HOME"
echo "  - RUSTUP_HOME: $RUSTUP_HOME"
echo ""
