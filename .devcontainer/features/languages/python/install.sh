#!/bin/bash
set -e

echo "========================================="
echo "Installing Python Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export PYENV_ROOT="${PYENV_ROOT:-$HOME/.cache/pyenv}"
export PYTHON_VERSION="${PYTHON_VERSION:-3.12}"
export PIP_CACHE_DIR="${PIP_CACHE_DIR:-$HOME/.cache/pip}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    build-essential \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    curl \
    git \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    libxml2-dev \
    libxmlsec1-dev \
    libffi-dev \
    liblzma-dev

# Install pyenv (Python Version Manager)
echo -e "${YELLOW}Installing pyenv...${NC}"
curl https://pyenv.run | bash

# Setup pyenv
export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init -)"
eval "$(pyenv virtualenv-init -)"

# Install Python (latest stable of specified major.minor)
echo -e "${YELLOW}Installing Python ${PYTHON_VERSION}...${NC}"
LATEST_PYTHON=$(pyenv install --list | grep -E "^\s*${PYTHON_VERSION}\.[0-9]+$" | tail -1 | xargs)
pyenv install "$LATEST_PYTHON"
pyenv global "$LATEST_PYTHON"

PYTHON_INSTALLED=$(python --version)
echo -e "${GREEN}✓ ${PYTHON_INSTALLED} installed${NC}"

# Upgrade pip
echo -e "${YELLOW}Upgrading pip...${NC}"
python -m pip install --break-system-packages --ignore-installed --upgrade pip
PIP_VERSION=$(pip --version)
echo -e "${GREEN}✓ ${PIP_VERSION}${NC}"

# Create cache directory
mkdir -p "$PIP_CACHE_DIR"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Python environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - pyenv (Python Version Manager)"
echo "  - ${PYTHON_INSTALLED}"
echo "  - pip (upgraded)"
echo ""
echo "Cache directory:"
echo "  - pip: $PIP_CACHE_DIR"
echo ""
