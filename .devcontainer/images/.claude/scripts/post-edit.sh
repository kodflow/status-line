#!/bin/bash
# ============================================================================
# post-edit.sh - Format only (fast, runs after every Write/Edit)
# Hook: PostToolUse (Write|Edit)
# Exit 0 = always (fail-open)
#
# Purpose: Auto-format edited files for immediate feedback.
# Lint, typecheck, test, and security scans are batched in on-stop-quality.sh
# (Stop hook) to avoid redundant CPU work when Claude edits multiple files.
# ============================================================================

set +e  # Fail-open: hooks should never block unexpectedly

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Read file_path from stdin JSON (preferred) or fallback to argument
INPUT="$(cat 2>/dev/null || true)"
FILE=""
if [ -n "$INPUT" ] && command -v jq &>/dev/null; then
    FILE=$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // ""' 2>/dev/null || true)
fi
FILE="${FILE:-${1:-}}"

if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

# Skip format for documentation and config files
if [[ "$FILE" == *".claude/contexts/"* ]] || \
   [[ "$FILE" == *".claude/plans/"* ]] || \
   [[ "$FILE" == *".claude/sessions/"* ]] || \
   [[ "$FILE" == */plans/* ]] || \
   [[ "$FILE" == *.md ]] || \
   [[ "$FILE" == /tmp/* ]] || \
   [[ "$FILE" == /home/vscode/.claude/* ]]; then
    exit 0
fi

# === Format only (fast: goimports ~100ms, ruff ~50ms, prettier ~200ms) ===
FMT_OUT=$("$SCRIPT_DIR/format.sh" "$FILE" 2>&1) || true

# === ktn-linter: focused scan after edit ===
# Calls ktn-linter HTTP endpoint to lint the modified file.
# Can block via permissionDecision:"deny" if critical violation found.
# Phase scope = phases 1..7 (excludes 'tests' and 'health' by default — same
# semantics as the legacy YAML default, just made explicit at request time).
KTN_PORT="${KTN_LINTER_PORT:-7717}"
if command -v curl &>/dev/null && [[ "$FILE" != *.md ]] && [[ "$FILE" != *.json ]] && \
   [[ "$FILE" != *.yaml ]] && [[ "$FILE" != *.yml ]] && [[ "$FILE" != *.toml ]] && \
   [[ "$FILE" != /tmp/* ]] && [[ "$FILE" != *".claude/"* ]]; then
    # Override via KTN_POST_PHASES env var, comma-separated.
    KTN_PHASES_CSV="${KTN_POST_PHASES:-structural,signatures,logic,performance,modern,style,comment}"
    KTN_PHASES_CSV="${KTN_PHASES_CSV// /}"
    KTN_BODY="${INPUT:-}"
    [ -z "$KTN_BODY" ] && KTN_BODY="{}"
    if command -v jq &>/dev/null; then
        KTN_TRY=$(printf '%s' "$KTN_BODY" | jq -c --arg p "$KTN_PHASES_CSV" \
            '. + {phases: ($p | split(","))}' 2>/dev/null) \
            && KTN_BODY="$KTN_TRY"
    fi
    KTN_RESP=$(curl -sf --max-time 14 \
        -H "Content-Type: application/json" \
        -d "$KTN_BODY" \
        "http://localhost:${KTN_PORT}/hooks/post-tool-use" 2>/dev/null) || true
    if [ -n "$KTN_RESP" ] && [ "$KTN_RESP" != "{}" ] && [ "$KTN_RESP" != "null" ] && command -v jq &>/dev/null; then
        if printf '%s' "$KTN_RESP" | jq -e '.hookSpecificOutput' &>/dev/null; then
            printf '%s' "$KTN_RESP"
        else
            jq -n -c --arg ctx "$(printf '%s' "$KTN_RESP" | head -c 1000)" \
                '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":$ctx}}' \
                2>/dev/null || true
        fi
    fi
fi

# Track edited file for on-stop-quality.sh batch processing (session-scoped)
SESSION_ID="${CLAUDE_SESSION_ID:-default}"
TRACKER="/tmp/.claude-edited-files-${SESSION_ID}"
if command -v flock &>/dev/null; then
    flock -w 2 "$TRACKER.lock" bash -c "echo '$FILE' >> '$TRACKER'" 2>/dev/null || true
else
    echo "$FILE" >> "$TRACKER" 2>/dev/null || true
fi

# Output additionalContext only if formatter reported issues
if [ -n "$FMT_OUT" ] && command -v jq &>/dev/null; then
    CONTEXT="Format issues in $FILE:\n${FMT_OUT:0:300}"
    jq -n -c \
        --arg ctx "$CONTEXT" \
        '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":$ctx}}' \
        2>/dev/null || true
fi

exit 0
