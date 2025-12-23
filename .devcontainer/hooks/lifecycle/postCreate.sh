#!/bin/bash
# ============================================================================
# postCreate.sh - Runs ONCE after container is assigned to user
# ============================================================================
# This script runs once after the dev container is assigned to a user.
# Use it for: User-specific setup, environment variables, shell config.
# Has access to user-specific secrets and permissions.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

echo ""
echo -e "${CYAN}=========================================${NC}"
echo -e "${CYAN}   Kodflow DevContainer Setup${NC}"
echo -e "${CYAN}=========================================${NC}"
echo ""

# Note: Tools (status-line, ktn-linter) are now baked into the Docker image
# No longer need to update on each rebuild

# Check if already initialized
if [ -f /home/vscode/.kodflow-initialized ]; then
    log_success "Kodflow already initialized"
    echo ""
    exit 0
fi

log_info "Setting up environment variables and aliases..."

# Create environment initialization script
cat > /home/vscode/.kodflow-env.sh << 'ENVEOF'
# Kodflow Environment Initialization
# This file is sourced by ~/.zshrc and ~/.bashrc

# NVM (Node.js Version Manager)
export NVM_DIR="/home/vscode/.cache/nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

# pyenv (Python Version Manager)
export PYENV_ROOT="/home/vscode/.cache/pyenv"
if [ -d "$PYENV_ROOT" ]; then
    export PATH="$PYENV_ROOT/bin:$PATH"
    eval "$(pyenv init -)" 2>/dev/null || true
    eval "$(pyenv virtualenv-init -)" 2>/dev/null || true
fi

# rbenv (Ruby Version Manager)
export RBENV_ROOT="/home/vscode/.cache/rbenv"
if [ -d "$RBENV_ROOT" ]; then
    export PATH="$RBENV_ROOT/bin:$PATH"
    eval "$(rbenv init -)" 2>/dev/null || true
fi

# SDKMAN (Java/JVM SDK Manager)
export SDKMAN_DIR="/home/vscode/.cache/sdkman"
[[ -s "$SDKMAN_DIR/bin/sdkman-init.sh" ]] && source "$SDKMAN_DIR/bin/sdkman-init.sh"

# Rust/Cargo
export CARGO_HOME="/home/vscode/.cache/cargo"
export RUSTUP_HOME="/home/vscode/.cache/rustup"
[ -f "$CARGO_HOME/env" ] && source "$CARGO_HOME/env"

# Go
export GOPATH="/home/vscode/.cache/go"
if [ -d "/usr/local/go" ]; then
    export GOROOT="/usr/local/go"
    export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"
fi

# Flutter/Dart
export FLUTTER_ROOT="/home/vscode/.cache/flutter"
export PUB_CACHE="/home/vscode/.cache/pub-cache"
if [ -d "$FLUTTER_ROOT" ]; then
    export PATH="$FLUTTER_ROOT/bin:$PUB_CACHE/bin:$PATH"
fi

# Composer (PHP)
export COMPOSER_HOME="/home/vscode/.cache/composer"
export PATH="$COMPOSER_HOME/vendor/bin:$PATH"

# Mix (Elixir)
export MIX_HOME="/home/vscode/.cache/mix"
export PATH="$MIX_HOME/escripts:$PATH"

# npm global packages
export PATH="/home/vscode/.local/share/npm-global/bin:$PATH"

# pnpm
export PNPM_HOME="/home/vscode/.cache/pnpm"
export PATH="$PNPM_HOME:$PATH"

# Local bin
export PATH="/home/vscode/.local/bin:$PATH"

# vcpkg
export VCPKG_ROOT="/home/vscode/.cache/vcpkg"
export PATH="$VCPKG_ROOT:$PATH"

# Carbon
export CARBON_PATH="/home/vscode/.cache/carbon"
export PATH="$CARBON_PATH/bin:$PATH"

# Bazel
export BAZEL_USER_ROOT="/home/vscode/.cache/bazel"

# Aliases
# super-claude: runs claude with MCP config if available, otherwise without
super-claude() {
    local mcp_config="/workspace/.mcp.json"
    if [ -f "$mcp_config" ] && jq empty "$mcp_config" 2>/dev/null; then
        claude --dangerously-skip-permissions --mcp-config "$mcp_config" "$@"
    else
        echo "⚠️  MCP config not found or invalid, starting without MCP..."
        claude --dangerously-skip-permissions "$@"
    fi
}

# Kubernetes auto-completion (if kubectl is installed)
if command -v kubectl &> /dev/null; then
    source <(kubectl completion zsh) 2>/dev/null || true
fi

# Helm auto-completion (if helm is installed)
if command -v helm &> /dev/null; then
    source <(helm completion zsh) 2>/dev/null || true
fi

# Terraform auto-completion (if terraform is installed)
if command -v terraform &> /dev/null; then
    complete -o nospace -C "$(which terraform)" terraform 2>/dev/null || true
fi

# Vault auto-completion (if vault is installed)
if command -v vault &> /dev/null; then
    complete -o nospace -C "$(which vault)" vault 2>/dev/null || true
fi

# Consul auto-completion (if consul is installed)
if command -v consul &> /dev/null; then
    complete -o nospace -C "$(which consul)" consul 2>/dev/null || true
fi

# Nomad auto-completion (if nomad is installed)
if command -v nomad &> /dev/null; then
    complete -o nospace -C "$(which nomad)" nomad 2>/dev/null || true
fi

# Packer auto-completion (if packer is installed)
if command -v packer &> /dev/null; then
    complete -o nospace -C "$(which packer)" packer 2>/dev/null || true
fi

# Docker auto-completion (if docker is installed)
if command -v docker &> /dev/null; then
    source <(docker completion zsh) 2>/dev/null || true
fi

# AWS CLI auto-completion (if aws is installed)
if command -v aws_completer &> /dev/null; then
    complete -C aws_completer aws 2>/dev/null || true
fi

# Google Cloud SDK auto-completion (if gcloud is installed)
if [ -f "/usr/share/google-cloud-sdk/completion.zsh.inc" ]; then
    source "/usr/share/google-cloud-sdk/completion.zsh.inc" 2>/dev/null || true
fi

# Go auto-completion
if command -v go &> /dev/null; then
    source <(go env GOROOT)/misc/zsh/go 2>/dev/null || true
fi

# Cargo/Rust auto-completion (if rustup is installed)
if command -v rustup &> /dev/null; then
    source <(rustup completions zsh) 2>/dev/null || true
    source <(rustup completions zsh cargo) 2>/dev/null || true
fi

# npm auto-completion (if npm is installed)
if command -v npm &> /dev/null; then
    source <(npm completion) 2>/dev/null || true
fi

# pnpm auto-completion (if pnpm is installed)
if command -v pnpm &> /dev/null; then
    source <(pnpm completion zsh) 2>/dev/null || true
fi

# gh (GitHub CLI) auto-completion (if gh is installed)
if command -v gh &> /dev/null; then
    source <(gh completion -s zsh) 2>/dev/null || true
fi
ENVEOF

log_success "Environment script created at ~/.kodflow-env.sh"

# Mark as initialized
touch /home/vscode/.kodflow-initialized

echo ""
echo -e "${CYAN}=========================================${NC}"
echo -e "${CYAN}   postCreate Complete${NC}"
echo -e "${CYAN}=========================================${NC}"
echo ""

exit 0
