#!/bin/bash
set -e

echo "========================================="
echo "Installing Ruby Development Environment"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Environment variables
export RBENV_ROOT="${RBENV_ROOT:-$HOME/.cache/rbenv}"
export RUBY_VERSION="${RUBY_VERSION:-3.3}"
export GEM_HOME="${GEM_HOME:-$HOME/.cache/gems}"
export BUNDLE_PATH="${BUNDLE_PATH:-$HOME/.cache/bundle}"

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"
sudo apt-get update && sudo apt-get install -y \
    curl \
    git \
    build-essential \
    libssl-dev \
    libreadline-dev \
    zlib1g-dev \
    autoconf \
    bison \
    libyaml-dev \
    libncurses5-dev \
    libffi-dev \
    libgdbm-dev

# Install rbenv (Ruby Version Manager)
echo -e "${YELLOW}Installing rbenv...${NC}"
git clone https://github.com/rbenv/rbenv.git "$RBENV_ROOT"
git clone https://github.com/rbenv/ruby-build.git "$RBENV_ROOT/plugins/ruby-build"

# Setup rbenv
export PATH="$RBENV_ROOT/bin:$PATH"
eval "$(rbenv init -)"

# Install Ruby (latest stable of specified major.minor)
echo -e "${YELLOW}Installing Ruby ${RUBY_VERSION}...${NC}"
LATEST_RUBY=$(rbenv install --list | grep -E "^\s*${RUBY_VERSION}\.[0-9]+$" | tail -1 | xargs)
rbenv install "$LATEST_RUBY"
rbenv global "$LATEST_RUBY"

RUBY_INSTALLED=$(ruby --version)
echo -e "${GREEN}✓ ${RUBY_INSTALLED} installed${NC}"

# Update RubyGems
echo -e "${YELLOW}Updating RubyGems...${NC}"
gem update --system
GEM_VERSION=$(gem --version)
echo -e "${GREEN}✓ RubyGems ${GEM_VERSION} installed${NC}"

# Install Bundler
echo -e "${YELLOW}Installing Bundler...${NC}"
gem install bundler
rbenv rehash
BUNDLER_VERSION=$(bundler --version)
echo -e "${GREEN}✓ ${BUNDLER_VERSION} installed${NC}"

# Create cache directories
mkdir -p "$GEM_HOME"
mkdir -p "$BUNDLE_PATH"

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Ruby environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Installed components:"
echo "  - rbenv (Ruby Version Manager)"
echo "  - ${RUBY_INSTALLED}"
echo "  - RubyGems ${GEM_VERSION}"
echo "  - ${BUNDLER_VERSION}"
echo ""
echo "Cache directories:"
echo "  - rbenv: $RBENV_ROOT"
echo "  - gems: $GEM_HOME"
echo "  - bundler: $BUNDLE_PATH"
echo ""
