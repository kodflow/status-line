#!/bin/bash
# ============================================================================
# postStart.sh - Runs EVERY TIME the container starts
# ============================================================================
# This script runs after postCreateCommand and before postAttachCommand.
# Runs each time the container is successfully started.
# Use it for: Claude update, status-line build, services startup.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

log_info "postStart: Container starting..."

# ============================================================================
# Claude CLI Update
# ============================================================================
log_info "Updating Claude CLI..."
claude update -y 2>/dev/null && log_success "Claude CLI updated" || log_warning "Claude CLI update failed (may already be latest)"

# ============================================================================
# Status-line Build (if binary doesn't exist)
# ============================================================================
STATUSLINE_BINARY="/workspace/bin/status-line"
if [ ! -f "$STATUSLINE_BINARY" ]; then
    if [ -f /workspace/Makefile ] && grep -q "status-line" /workspace/Makefile; then
        log_info "Building status-line (binary not found)..."
        make -C /workspace build 2>/dev/null && log_success "status-line built" || log_warning "status-line build failed"
    fi
fi

# ============================================================================
# Claude CLI Configuration with Status-line
# ============================================================================
log_info "Configuring Claude CLI..."
mkdir -p /home/vscode/.claude
cat > /home/vscode/.claude/settings.json <<'EOF'
{
  "enableAllProjectMcpServers": true,
  "alwaysThinkingEnabled": true,
  "statusLine": {
    "type": "command",
    "command": "/workspace/bin/status-line",
    "padding": 0
  }
}
EOF
log_success "Claude CLI configured with status-line"

# ============================================================================
# Final message
# ============================================================================
echo ""
log_success "postStart: Container ready!"
