#!/bin/bash
# ============================================================================
# postStart.sh - Runs EVERY TIME the container starts
# ============================================================================
# This script runs after postCreateCommand and before postAttachCommand.
# Runs each time the container is successfully started.
# Use it for: MCP setup, services startup, recurring initialization.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

log_info "postStart: Container starting..."

# ============================================================================
# Restore Claude commands/scripts from image defaults
# ============================================================================
# Volume mounts overwrite image content, so we restore from /etc/claude-defaults/
CLAUDE_DEFAULTS="/etc/claude-defaults"

if [ -d "$CLAUDE_DEFAULTS" ]; then
    log_info "Restoring Claude configuration from image defaults..."
    
    # Ensure base directory exists
    mkdir -p "$HOME/.claude"
    
    # Restore commands (always overwrite with latest from image)
    if [ -d "$CLAUDE_DEFAULTS/commands" ]; then
        mkdir -p "$HOME/.claude/commands"
        cp -r "$CLAUDE_DEFAULTS/commands/"* "$HOME/.claude/commands/" 2>/dev/null || true
    fi
    
    # Restore scripts (always overwrite with latest from image)
    if [ -d "$CLAUDE_DEFAULTS/scripts" ]; then
        mkdir -p "$HOME/.claude/scripts"
        cp -r "$CLAUDE_DEFAULTS/scripts/"* "$HOME/.claude/scripts/" 2>/dev/null || true
        chmod -R 755 "$HOME/.claude/scripts/"
    fi
    
    # Restore settings.json only if it does not exist
    if [ -f "$CLAUDE_DEFAULTS/settings.json" ] && [ \! -f "$HOME/.claude/settings.json" ]; then
        cp "$CLAUDE_DEFAULTS/settings.json" "$HOME/.claude/settings.json"
    fi
    
    log_success "Claude configuration restored"
fi

# ============================================================================
# Ensure Claude directories exist (volume mount point)
# ============================================================================
mkdir -p "$HOME/.claude/sessions" "$HOME/.claude/plans"
log_success "Claude directories initialized"

# ============================================================================
# Restore NVM symlinks (node, npm, npx, claude)
# ============================================================================
# NVM is in package-cache volume, so we need to recreate symlinks to ~/.local/bin
NVM_DIR="${NVM_DIR:-$HOME/.cache/nvm}"
if [ -d "$NVM_DIR/versions/node" ]; then
    NODE_VERSION=$(ls "$NVM_DIR/versions/node" 2>/dev/null | head -1)
    if [ -n "$NODE_VERSION" ]; then
        log_info "Setting up Node.js symlinks for $NODE_VERSION..."
        mkdir -p "$HOME/.local/bin"
        NODE_BIN="$NVM_DIR/versions/node/$NODE_VERSION/bin"

        for cmd in node npm npx claude; do
            if [ -f "$NODE_BIN/$cmd" ]; then
                ln -sf "$NODE_BIN/$cmd" "$HOME/.local/bin/$cmd"
            fi
        done
        log_success "Node.js symlinks configured"
    fi
fi

# ============================================================================
# 1Password CLI Setup
# ============================================================================
# Fix op config directory permissions (created by Docker as root)
OP_CONFIG_DIR="/home/vscode/.config/op"
if [ -d "$OP_CONFIG_DIR" ]; then
    if [ "$(stat -c '%U' "$OP_CONFIG_DIR" 2>/dev/null)" != "vscode" ]; then
        log_info "Fixing 1Password config directory permissions..."
        sudo chown -R vscode:vscode "$OP_CONFIG_DIR" 2>/dev/null || true
    fi
    chmod 700 "$OP_CONFIG_DIR" 2>/dev/null || true
fi

# Reload .env file to get updated tokens
ENV_FILE="/workspace/.devcontainer/.env"
if [ -f "$ENV_FILE" ]; then
    log_info "Reloading environment from .env..."
    set -a
    source "$ENV_FILE"
    set +a
fi

# ============================================================================
# MCP Configuration Setup (inject secrets into template)
# ============================================================================
VAULT_ID="ypahjj334ixtiyjkytu5hij2im"
MCP_TPL="/etc/mcp/mcp.json.tpl"
MCP_OUTPUT="/workspace/.mcp.json"

# Helper function to get 1Password field (tries multiple field names)
# Usage: get_1password_field <item_name> <vault_id>
get_1password_field() {
    local item="$1"
    local vault="$2"
    local fields=("credential" "password" "identifiant" "mot de passe")

    for field in "${fields[@]}"; do
        local value
        value=$(op item get "$item" --vault "$vault" --fields "$field" --reveal 2>/dev/null || echo "")
        if [ -n "$value" ]; then
            echo "$value"
            return 0
        fi
    done
    echo ""
}

# Initialize tokens from environment variables (fallback)
CODACY_TOKEN="${CODACY_API_TOKEN:-}"
GITHUB_TOKEN="${GITHUB_API_TOKEN:-}"

# Try 1Password if OP_SERVICE_ACCOUNT_TOKEN is defined
if [ -n "$OP_SERVICE_ACCOUNT_TOKEN" ] && command -v op &> /dev/null; then
    log_info "Retrieving secrets from 1Password..."

    OP_CODACY=$(get_1password_field "mcp-codacy" "$VAULT_ID")
    OP_GITHUB=$(get_1password_field "mcp-github" "$VAULT_ID")

    [ -n "$OP_CODACY" ] && CODACY_TOKEN="$OP_CODACY"
    [ -n "$OP_GITHUB" ] && GITHUB_TOKEN="$OP_GITHUB"
fi

# Show warnings if tokens are missing
[ -z "$CODACY_TOKEN" ] && log_warning "Codacy token not available"
[ -z "$GITHUB_TOKEN" ] && log_warning "GitHub token not available"

# Generate mcp.json from template (baked in Docker image)
if [ -f "$MCP_TPL" ]; then
    log_info "Generating .mcp.json from template..."
    sed -e "s|{{CODACY_TOKEN}}|${CODACY_TOKEN}|g" \
        -e "s|{{GITHUB_TOKEN}}|${GITHUB_TOKEN}|g" \
        "$MCP_TPL" > "$MCP_OUTPUT"
    log_success "mcp.json generated successfully"

    # =========================================================================
    # Add optional MCPs based on installed features
    # =========================================================================
    # Helper function to add a conditional MCP server
    add_optional_mcp() {
        local name="$1"
        local binary="$2"
        local output="$3"

        if [ -x "$binary" ]; then
            log_info "Adding $name MCP (binary found at $binary)"
            jq --arg name "$name" --arg bin "$binary" \
               '.mcpServers[$name] = {"command": $bin, "args": [], "env": {}}' \
               "$output" > "$output.tmp" && mv "$output.tmp" "$output"
        else
            log_info "Skipping $name MCP (binary not found)"
        fi
    }

    # Rust: rust-analyzer-mcp (only if Rust feature is installed)
    add_optional_mcp "rust-analyzer" "$HOME/.cache/cargo/bin/rust-analyzer-mcp" "$MCP_OUTPUT"

    # Future conditional MCPs can be added here:
    # add_optional_mcp "gopls" "$HOME/.cache/go/bin/gopls-mcp" "$MCP_OUTPUT"
    # add_optional_mcp "pyright" "$HOME/.cache/pyenv/shims/pyright-mcp" "$MCP_OUTPUT"
else
    log_warning "MCP template not found at $MCP_TPL"
fi

# ============================================================================
# Git Credential Cleanup (remove macOS-specific helpers)
# ============================================================================
log_info "Cleaning git credential helpers..."
git config --global --unset-all credential.https://github.com.helper 2>/dev/null || true
git config --global --unset-all credential.https://gist.github.com.helper 2>/dev/null || true
log_success "Git credential helpers cleaned"

# ============================================================================
# Final message
# ============================================================================
echo ""
log_success "postStart: Container ready!"
