#!/bin/bash
# Don't exit on error - we want to use our retry logic
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Retry function
retry() {
    local max_attempts=$1
    local delay=$2
    shift 2
    local attempt=1
    local exit_code=0

    while [ $attempt -le $max_attempts ]; do
        if "$@"; then
            if [ $attempt -gt 1 ]; then
                log_success "Command succeeded on attempt $attempt"
            fi
            return 0
        fi

        exit_code=$?

        if [ $attempt -lt $max_attempts ]; then
            log_warning "Command failed (exit code: $exit_code), retrying in ${delay}s... (attempt $attempt/$max_attempts)"
            sleep "$delay"
        else
            log_error "Command failed after $max_attempts attempts"
        fi

        ((attempt++))
    done

    return $exit_code
}

# apt-get with retry and lock handling
apt_get_retry() {
    local max_attempts=5
    local attempt=1
    local delay=10

    while [ $attempt -le $max_attempts ]; do
        # Wait for apt locks to be released
        local lock_wait=0
        while sudo fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || \
              sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1 || \
              sudo fuser /var/cache/apt/archives/lock >/dev/null 2>&1; do
            if [ $lock_wait -eq 0 ]; then
                log_warning "Waiting for apt locks to be released..."
            fi
            sleep 2
            lock_wait=$((lock_wait + 2))

            if [ $lock_wait -ge 60 ]; then
                log_warning "Forcing apt lock release after 60s wait"
                sudo rm -f /var/lib/dpkg/lock-frontend
                sudo rm -f /var/lib/apt/lists/lock
                sudo rm -f /var/cache/apt/archives/lock
                sudo dpkg --configure -a || true
                break
            fi
        done

        # Try apt-get command
        if sudo apt-get "$@"; then
            if [ $attempt -gt 1 ]; then
                log_success "apt-get succeeded on attempt $attempt"
            fi
            return 0
        fi

        exit_code=$?

        if [ $attempt -lt $max_attempts ]; then
            log_warning "apt-get failed, running update and retrying in ${delay}s... (attempt $attempt/$max_attempts)"
            sudo apt-get update --fix-missing || true
            sudo dpkg --configure -a || true
            sleep "$delay"
        else
            log_error "apt-get failed after $max_attempts attempts"
        fi

        ((attempt++))
    done

    return $exit_code
}

# Download and pipe to shell with retry
download_and_pipe() {
    local url=$1
    shift
    local shell_cmd=("$@")

    log_info "Downloading and executing: $url"

    local temp_file
    temp_file=$(mktemp)

    if retry 3 5 curl -fsSL --connect-timeout 30 --max-time 300 -o "$temp_file" "$url"; then
        "${shell_cmd[@]}" < "$temp_file"
        local exit_code=$?
        rm -f "$temp_file"
        return $exit_code
    else
        rm -f "$temp_file"
        return 1
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Safe directory creation
mkdir_safe() {
    local dir=$1
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir" 2>/dev/null || sudo mkdir -p "$dir"

        if [ "$(whoami)" = "vscode" ]; then
            sudo chown -R vscode:vscode "$dir" 2>/dev/null || true
        fi
    fi
}

echo "========================================="
echo "Installing Node.js Development Environment"
echo "========================================="

# Environment variables
export NVM_DIR="${NVM_DIR:-/home/vscode/.cache/nvm}"
export NODE_VERSION="${NODE_VERSION:-lts/*}"
export npm_config_cache="${npm_config_cache:-/home/vscode/.cache/npm}"

# Install dependencies with retry
log_info "Installing dependencies..."
apt_get_retry update
apt_get_retry install -y curl git build-essential libssl-dev || {
    log_error "Failed to install dependencies"
    exit 1
}

# Install NVM (Node Version Manager)
log_info "Installing NVM..."
mkdir_safe "$NVM_DIR"
# Use v0.40.1 (latest stable version) instead of "latest" which doesn't exist
download_and_pipe "https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh" bash || {
    log_error "Failed to install NVM"
    exit 1
}

# Load NVM
export NVM_DIR="$NVM_DIR"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Install Node.js (latest LTS by default)
log_info "Installing Node.js ${NODE_VERSION}..."
retry 3 5 nvm install "$NODE_VERSION" || {
    log_error "Failed to install Node.js"
    exit 1
}
nvm use "$NODE_VERSION"
nvm alias default "$NODE_VERSION"

# Get installed Node and npm versions
NODE_INSTALLED=$(node --version)
NPM_INSTALLED=$(npm --version)

log_success "Node.js ${NODE_INSTALLED} installed"
log_success "npm ${NPM_INSTALLED} installed"

# Create cache directory
mkdir_safe "$npm_config_cache"

# Create global symlinks for node, npm, and npx
# This ensures they're available for subsequent devcontainer features
log_info "Creating global symlinks..."
NVM_NODE_DIR=$(nvm which current | xargs dirname)
if [ -n "$NVM_NODE_DIR" ] && [ -d "$NVM_NODE_DIR" ]; then
    # Create symlinks in /usr/local/bin (which is in default PATH)
    sudo ln -sf "$NVM_NODE_DIR/node" /usr/local/bin/node
    sudo ln -sf "$NVM_NODE_DIR/npm" /usr/local/bin/npm
    sudo ln -sf "$NVM_NODE_DIR/npx" /usr/local/bin/npx

    # Verify symlinks were created
    if [ -L /usr/local/bin/node ] && [ -L /usr/local/bin/npm ]; then
        log_success "Global symlinks created in /usr/local/bin"
    else
        log_warning "Failed to create global symlinks, but NVM is still available"
    fi
else
    log_warning "Could not determine NVM node directory, skipping symlink creation"
fi

# Add NVM to zshrc for interactive shells
ZSHRC="/home/vscode/.zshrc"
if [ -f "$ZSHRC" ]; then
    if ! grep -q "NVM_DIR" "$ZSHRC"; then
        log_info "Adding NVM to .zshrc..."
        cat >> "$ZSHRC" <<'EOF'

# NVM (Node Version Manager)
export NVM_DIR="${NVM_DIR:-/home/vscode/.cache/nvm}"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
EOF
    fi
fi

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}Node.js environment installed successfully!${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
log_success "Installation complete!"
echo ""
echo "Installed components:"
echo "  - NVM (Node Version Manager)"
echo "  - Node.js ${NODE_INSTALLED}"
echo "  - npm ${NPM_INSTALLED}"
echo ""
echo "Global availability:"
echo "  - node, npm, npx available in /usr/local/bin"
echo "  - NVM loaded in interactive shells"
echo ""
echo "Cache directory:"
echo "  - npm: $npm_config_cache"
echo ""

# Exit successfully
exit 0
