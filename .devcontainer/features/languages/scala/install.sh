#!/bin/bash
set -e

echo "========================================="
echo "Installing Scala Development Environment"
echo "========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Environment variables
export SDKMAN_DIR="${SDKMAN_DIR:-/home/vscode/.cache/sdkman}"
export COURSIER_CACHE="${COURSIER_CACHE:-/home/vscode/.cache/coursier}"

# Check if SDKMAN is installed (requires Java feature)
if [ ! -d "$SDKMAN_DIR" ]; then
    echo -e "${RED}Error: SDKMAN not found. Please install the Java feature first.${NC}"
    exit 1
fi

# Source SDKMAN
source "$SDKMAN_DIR/bin/sdkman-init.sh"

# Install Scala via SDKMAN
echo -e "${YELLOW}Installing Scala...${NC}"
sdk install scala
SCALA_VERSION=$(scala -version 2>&1 | head -n 1)
echo -e "${GREEN}✓ ${SCALA_VERSION} installed${NC}"

# Install sbt via SDKMAN
echo -e "${YELLOW}Installing sbt...${NC}"
sdk install sbt
SBT_VERSION=$(sbt --version 2>&1 | grep "sbt script" | head -n 1 || echo "sbt installed")
echo -e "${GREEN}✓ ${SBT_VERSION}${NC}"

# Install Coursier (Scala artifact fetcher)
echo -e "${YELLOW}Installing Coursier...${NC}"
curl -fL "https://github.com/coursier/launchers/raw/master/cs-x86_64-pc-linux.gz" | gzip -d > /tmp/cs
chmod +x /tmp/cs
sudo mv /tmp/cs /usr/local/bin/cs
echo -e "${GREEN}✓ Coursier installed${NC}"

# Install common Scala tools via Coursier
echo -e "${YELLOW}Installing Scala CLI...${NC}"
cs install scala-cli
echo -e "${GREEN}✓ Scala CLI installed${NC}"

# Install Metals (Scala LSP)
echo -e "${YELLOW}Installing Metals (LSP)...${NC}"
cs install metals
echo -e "${GREEN}✓ Metals installed${NC}"

# Install scalafmt
echo -e "${YELLOW}Installing scalafmt...${NC}"
cs install scalafmt
echo -e "${GREEN}✓ scalafmt installed${NC}"

# Create cache directories
mkdir -p "$COURSIER_CACHE"
mkdir -p /home/vscode/.cache/sbt

# Add Coursier bin to PATH
COURSIER_BIN="$HOME/.local/share/coursier/bin"
if ! grep -q "COURSIER" /home/vscode/.zshrc 2>/dev/null; then
    echo "" >> /home/vscode/.zshrc
    echo "# Coursier" >> /home/vscode/.zshrc
    echo "export PATH=\"\$PATH:$COURSIER_BIN\"" >> /home/vscode/.zshrc
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Scala environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - ${SCALA_VERSION}"
echo "  - sbt (Scala Build Tool)"
echo "  - Coursier (artifact fetcher)"
echo "  - Scala CLI"
echo "  - Metals (LSP for IDE support)"
echo "  - scalafmt (formatter)"
echo ""
echo "Cache directories:"
echo "  - Coursier: $COURSIER_CACHE"
echo "  - sbt: /home/vscode/.cache/sbt"
echo ""
echo "Bazel integration:"
echo "  - Use rules_scala: https://github.com/bazelbuild/rules_scala"
echo ""
