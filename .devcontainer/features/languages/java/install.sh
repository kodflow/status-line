#!/bin/bash
set -e

echo "========================================="
echo "Installing Java Development Environment"
echo "========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Environment variables
export SDKMAN_DIR="${SDKMAN_DIR:-/home/vscode/.cache/sdkman}"
export MAVEN_OPTS="${MAVEN_OPTS:--Dmaven.repo.local=/home/vscode/.cache/maven}"
export GRADLE_USER_HOME="${GRADLE_USER_HOME:-/home/vscode/.cache/gradle}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    zip \
    unzip \
    git

# Install SDKMAN
echo -e "${YELLOW}Installing SDKMAN...${NC}"
curl -s "https://get.sdkman.io" | bash
source "$SDKMAN_DIR/bin/sdkman-init.sh"
echo -e "${GREEN}✓ SDKMAN installed${NC}"

# Install Java (latest LTS)
echo -e "${YELLOW}Installing Java...${NC}"
sdk install java
JAVA_VERSION=$(java -version 2>&1 | head -n 1)
echo -e "${GREEN}✓ ${JAVA_VERSION} installed${NC}"

# Install Maven
echo -e "${YELLOW}Installing Maven...${NC}"
sdk install maven
MAVEN_VERSION=$(mvn -version | head -n 1)
echo -e "${GREEN}✓ ${MAVEN_VERSION} installed${NC}"

# Install Gradle
echo -e "${YELLOW}Installing Gradle...${NC}"
sdk install gradle
GRADLE_VERSION=$(gradle -version | grep "Gradle" | head -n 1)
echo -e "${GREEN}✓ ${GRADLE_VERSION} installed${NC}"

# Create cache directories
mkdir -p /home/vscode/.cache/maven
mkdir -p /home/vscode/.cache/gradle

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Java environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - SDKMAN (SDK Manager)"
echo "  - ${JAVA_VERSION}"
echo "  - ${MAVEN_VERSION}"
echo "  - ${GRADLE_VERSION}"
echo ""
echo "Cache directories:"
echo "  - SDKMAN: $SDKMAN_DIR"
echo "  - Maven: /home/vscode/.cache/maven"
echo "  - Gradle: $GRADLE_USER_HOME"
echo ""
