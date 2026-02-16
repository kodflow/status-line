#!/bin/bash
# ============================================================================
# notification.sh - Container-friendly notification handler
# Hook: Notification (all matchers)
# Exit 0 = always (fail-open)
#
# Purpose: Audible + logged notification for external monitoring.
# ============================================================================

set +e  # Fail-open: never block

# Terminal bell
printf '\a'

# Read hook input
INPUT=""
if [ ! -t 0 ]; then
    INPUT=$(cat 2>/dev/null || true)
fi

# Append to notification log for external monitoring
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-/workspace}"
LOG_DIR="$PROJECT_DIR/.claude/logs"
mkdir -p "$LOG_DIR" 2>/dev/null || true

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
MESSAGE=""
if [ -n "$INPUT" ] && command -v jq &>/dev/null; then
    MESSAGE=$(printf '%s' "$INPUT" | jq -r '.message // ""' 2>/dev/null || true)
fi

if command -v jq &>/dev/null; then
    jq -n -c \
        --arg ts "$TIMESTAMP" \
        --arg msg "$MESSAGE" \
        '{timestamp: $ts, message: $msg}' >> "$LOG_DIR/notification.jsonl" 2>/dev/null || true
fi

exit 0
