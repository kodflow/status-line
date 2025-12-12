#!/bin/bash
# ============================================================================
# onCreate.sh - Runs INSIDE container immediately after first start
# ============================================================================
# This script runs inside the container after it starts for the first time.
# Use it for: Initial container setup that doesn't need user-specific config.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

log_info "onCreate: Container created, performing initial setup..."

# Ensure cache directories exist with proper permissions
CACHE_DIRS=(
    "/home/vscode/.cache"
    "/home/vscode/.config"
    "/home/vscode/.local/bin"
    "/home/vscode/.local/share"
)

for dir in "${CACHE_DIRS[@]}"; do
    mkdir_safe "$dir"
done

log_success "onCreate: Initial container setup complete"
