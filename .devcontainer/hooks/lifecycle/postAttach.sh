#!/bin/bash
# ============================================================================
# postAttach.sh - Runs when IDE connects to the container
# ============================================================================
# This script runs when a tool (like VS Code) connects to the dev container.
# It's the only command that consistently allows user interaction.
# Use it for: Welcome messages, status display, interactive prompts.
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

echo ""
echo -e "${CYAN}=========================================${NC}"
echo -e "${CYAN}   Kodflow DevContainer Ready${NC}"
echo -e "${CYAN}=========================================${NC}"
echo ""

# Display useful information
log_success "IDE connected to DevContainer"
echo ""
echo -e "Tip: Use ${GREEN}super-claude${NC} for Claude CLI with MCP config"
echo ""
