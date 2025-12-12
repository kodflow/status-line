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
# MCP Configuration Setup
# ============================================================================
VAULT_ID="ypahjj334ixtiyjkytu5hij2im"
MCP_TPL="/workspace/.devcontainer/hooks/shared/mcp.json.tpl"
MCP_OUTPUT="/home/vscode/.devcontainer/mcp.json"

# Initialize tokens
CODACY_TOKEN=""
GITHUB_TOKEN=""

# Try 1Password if OP_SERVICE_ACCOUNT_TOKEN is defined
if [ -n "$OP_SERVICE_ACCOUNT_TOKEN" ] && command -v op &> /dev/null; then
    log_info "Retrieving secrets from 1Password..."

    CODACY_TOKEN=$(op item get "mcp-codacy" --vault "$VAULT_ID" --fields credential --reveal 2>/dev/null || echo "")
    GITHUB_TOKEN=$(op item get "mcp-github" --vault "$VAULT_ID" --fields credential --reveal 2>/dev/null || echo "")
fi

# Use environment variables as fallback
if [ -z "$CODACY_TOKEN" ] && [ -n "$CODACY_API_TOKEN" ]; then
    log_info "Using Codacy token from CODACY_API_TOKEN"
    CODACY_TOKEN="$CODACY_API_TOKEN"
fi

if [ -z "$GITHUB_TOKEN" ] && [ -n "$GITHUB_API_TOKEN" ]; then
    log_info "Using GitHub token from GITHUB_API_TOKEN"
    GITHUB_TOKEN="$GITHUB_API_TOKEN"
fi

# Show warnings if tokens are missing
[ -z "$CODACY_TOKEN" ] && log_warning "Codacy token not available"
[ -z "$GITHUB_TOKEN" ] && log_warning "GitHub token not available"

# Generate mcp.json from template
if [ -f "$MCP_TPL" ]; then
    log_info "Generating mcp.json..."
    mkdir -p "$(dirname "$MCP_OUTPUT")"
    sed "s|{{ with secret \"secret/mcp/codacy\" }}{{ .Data.data.token }}{{ end }}|${CODACY_TOKEN}|g" "$MCP_TPL" | \
        sed "s|{{ with secret \"secret/mcp/github\" }}{{ .Data.data.token }}{{ end }}|${GITHUB_TOKEN}|g" \
        > "$MCP_OUTPUT"
    log_success "mcp.json generated successfully"
fi

# ============================================================================
# Claude CLI Configuration
# ============================================================================
log_info "Configuring Claude CLI..."
mkdir -p /home/vscode/.claude
cat > /home/vscode/.claude/settings.json <<'EOF'
{
  "enableAllProjectMcpServers": true,
  "alwaysThinkingEnabled": true
}
EOF
log_success "Claude CLI configured"

# ============================================================================
# Final message
# ============================================================================
echo ""
log_success "postStart: Container ready!"
