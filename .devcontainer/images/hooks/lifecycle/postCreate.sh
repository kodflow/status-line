#!/bin/bash
# shellcheck disable=SC1090,SC1091
# ============================================================================
# postCreate.sh - Runs ONCE after container is assigned to user
# ============================================================================
# This script runs once after the dev container is assigned to a user.
# Use it for: User-specific setup, environment variables, shell config.
# Has access to user-specific secrets and permissions.
#
# Uses run_step pattern: each step runs in an isolated subshell so that
# failures (e.g. unconfigured git email, missing GPG keys) never kill
# the entire script. The container always starts successfully.
# ============================================================================

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

echo ""
echo -e "${CYAN}=========================================${NC}"
echo -e "${CYAN}   DevContainer Setup${NC}"
echo -e "${CYAN}=========================================${NC}"
echo ""

init_steps

# ============================================================================
# Step functions
# ============================================================================

# Prevents "dubious ownership" errors when container user differs from
# directory owner (common in Docker where /workspace may be owned by root)
step_git_safe_directory() {
    if ! git config --global --get-all safe.directory 2>/dev/null | grep -q "^/workspace$"; then
        git config --global --add safe.directory /workspace
        log_success "Git safe.directory configured for /workspace"
    else
        log_info "Git safe.directory already configured"
    fi
}

# Disable SSL verification (for corporate proxies/self-signed certs)
step_git_ssl_config() {
    git config --global http.sslVerify false
    log_success "Git SSL verification disabled"
}

# GPG commit signing configuration
step_gpg_signing() {
    if [ ! -d "/home/vscode/.gnupg" ] || [ -z "$(gpg --list-secret-keys --keyid-format LONG 2>/dev/null)" ]; then
        log_info "No GPG keys available - commit signing disabled"
        return 0
    fi

    # Get GIT_EMAIL from .env or git config
    local git_email=""
    if [ -f "/workspace/.env" ]; then
        git_email=$(grep -E "^GIT_EMAIL=" /workspace/.env 2>/dev/null | cut -d'=' -f2 | tr -d '"' || true)
    fi
    if [ -z "$git_email" ]; then
        git_email=$(git config --global user.email 2>/dev/null || true)
    fi

    local gpg_key=""
    if [ -n "$git_email" ]; then
        # Priority: Find GPG key matching the configured GIT_EMAIL
        gpg_key=$(gpg --list-secret-keys --keyid-format LONG 2>/dev/null | \
            grep -B1 "$git_email" 2>/dev/null | \
            grep -E "^sec" 2>/dev/null | head -1 | awk '{print $2}' | cut -d'/' -f2 || true)
    fi

    if [ -n "$gpg_key" ]; then
        git config --global user.signingkey "$gpg_key"
        git config --global commit.gpgsign true
        git config --global tag.forceSignAnnotated true
        git config --global gpg.program gpg
        log_success "Git GPG signing configured with key: $gpg_key (matching $git_email)"
    else
        # No matching key found - GPG signing will be configured via /git skill
        log_warning "No GPG key found for email '$git_email' - configure via /git skill"
    fi
}

# Create environment initialization script (~/.devcontainer-env.sh)
step_create_env_script() {
    log_info "Setting up environment variables and aliases..."

    cat > /home/vscode/.devcontainer-env.sh << 'ENVEOF'
# DevContainer Environment Initialization (v2 - two-phase)
# This file is sourced by ~/.zshrc and ~/.bashrc
#
# Architecture: Two-phase loading for fast shell startup
#   Phase 1 (always): PATH exports, env vars — fast, no subprocesses
#   Phase 2 (real terminal only): completions, version manager init, aliases
#
# Why: VS Code's ptyHost spawns a shell to resolve env vars with a 10s timeout.
# Heavy init (eval, source <(...), nvm.sh) easily exceeds this on ARM64.
# Phase 1 gives VS Code the PATH/env it needs; Phase 2 only runs in terminals.

# ============================================================================
# Phase 1: Fast PATH and Environment Variables (no subprocesses)
# ============================================================================

# NVM (Node.js Version Manager)
export NVM_DIR="/usr/local/share/nvm"
export NVM_SYMLINK_CURRENT=true
# Add NVM current bin to PATH directly (no need to source heavy nvm.sh)
[ -d "$NVM_DIR/current/bin" ] && export PATH="$NVM_DIR/current/bin:$PATH"

# pyenv (Python Version Manager)
export PYENV_ROOT="/home/vscode/.cache/pyenv"
if [ -d "$PYENV_ROOT" ]; then
    export PATH="$PYENV_ROOT/shims:$PYENV_ROOT/bin:$PATH"
fi

# rbenv (Ruby Version Manager)
export RBENV_ROOT="/home/vscode/.cache/rbenv"
if [ -d "$RBENV_ROOT" ]; then
    export PATH="$RBENV_ROOT/shims:$RBENV_ROOT/bin:$PATH"
fi

# SDKMAN (Java/JVM SDK Manager)
export SDKMAN_DIR="/home/vscode/.cache/sdkman"
if [ -d "$SDKMAN_DIR/candidates" ]; then
    for _sdk_bin in "$SDKMAN_DIR"/candidates/*/current/bin; do
        [ -d "$_sdk_bin" ] && PATH="$_sdk_bin:$PATH"
    done
    unset _sdk_bin
fi

# Rust/Cargo
export CARGO_HOME="/home/vscode/.cache/cargo"
export RUSTUP_HOME="/home/vscode/.cache/rustup"
[ -d "$CARGO_HOME/bin" ] && export PATH="$CARGO_HOME/bin:$PATH"

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

# Scala (SBT)
export SBT_HOME="/home/vscode/.cache/sbt"
[ -d "$SBT_HOME/bin" ] && export PATH="$SBT_HOME/bin:$PATH"

# .NET (C#, VB.NET)
export DOTNET_ROOT="/usr/share/dotnet"
[ -d "$DOTNET_ROOT" ] && export PATH="$DOTNET_ROOT:$HOME/.dotnet/tools:$PATH"

# R
export R_HOME="/usr/lib/R"

# ============================================================================
# Phase 2: Interactive Terminal Features (completions, version managers, aliases)
# ============================================================================
# Skip when stdout is not a real terminal (e.g., VS Code env resolution).
# This is the key optimization: VS Code only needs PATH/env from Phase 1.
if [ ! -t 1 ]; then
    return 0 2>/dev/null || true
fi

# NVM full initialization (provides 'nvm' command and bash completion)
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

# pyenv full initialization (provides 'pyenv' shim rehash and virtualenv hooks)
if [ -d "$PYENV_ROOT" ]; then
    eval "$(pyenv init -)" 2>/dev/null || true
    eval "$(pyenv virtualenv-init -)" 2>/dev/null || true
fi

# rbenv full initialization (provides 'rbenv' shim rehash)
if [ -d "$RBENV_ROOT" ]; then
    eval "$(rbenv init -)" 2>/dev/null || true
fi

# SDKMAN full initialization (provides 'sdk' command)
[[ -s "$SDKMAN_DIR/bin/sdkman-init.sh" ]] && source "$SDKMAN_DIR/bin/sdkman-init.sh"

# Aliases
# super-claude: runs claude with MCP config if available, otherwise without
super-claude() {
    local mcp_config="/workspace/mcp.json"

    # Check if jq is available for JSON validation
    if ! command -v jq &>/dev/null; then
        echo "Warning: jq not found, skipping MCP config validation" >&2
        # Still use mcp config if it looks like JSON (skip leading whitespace/newlines)
        if [ -s "$mcp_config" ] && LC_ALL=C tr -d ' \t\r\n' < "$mcp_config" 2>/dev/null | head -c 1 | grep -q '{'; then
            claude --dangerously-skip-permissions --mcp-config "$mcp_config" "$@"
        else
            claude --dangerously-skip-permissions "$@"
        fi
        return
    fi

    if [ -f "$mcp_config" ] && jq empty "$mcp_config" 2>/dev/null; then
        claude --dangerously-skip-permissions --mcp-config "$mcp_config" "$@"
    else
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

    log_success "Environment script created at ~/.devcontainer-env.sh"
}

# Mark container as initialized
step_mark_initialized() {
    touch /home/vscode/.devcontainer-initialized
    log_success "DevContainer marked as initialized"
}

# ============================================================================
# Execution (always runs git steps, skips env if already initialized)
# ============================================================================

# Git steps run every time (safe directory, SSL, GPG)
run_step "Git safe directory"    step_git_safe_directory
run_step "Git SSL configuration" step_git_ssl_config
run_step "GPG signing"           step_gpg_signing

# Note: Tools (status-line, ktn-linter) are now baked into the Docker image
# No longer need to update on each rebuild

# Check if already initialized (but only if env file also exists)
# If ~/.devcontainer-env.sh is missing, we must recreate it even if marker exists
if [ -f /home/vscode/.devcontainer-initialized ] && [ -f /home/vscode/.devcontainer-env.sh ]; then
    log_success "DevContainer already initialized"
    echo ""
    exit 0
fi

run_step "Environment script"    step_create_env_script
run_step "Mark initialized"      step_mark_initialized

print_step_summary "postCreate"

exit 0
