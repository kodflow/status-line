#!/bin/bash
# Combined post-edit hook: format + imports + lint + WIP check
# Usage: post-edit.sh <file_path>
# En PLAN MODE: skip format/lint pour les fichiers autorisés (plans, docs)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FILE="$1"

if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

# === Trouver la session active (déterministe) ===
SESSION_FILE=""

# Priorité 1: Pointeur explicite
if [[ -f "/workspace/.claude/active-session" ]]; then
    SESSION_FILE=$(cat /workspace/.claude/active-session 2>/dev/null || true)
fi

# Priorité 2: Symlink state.json
if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    if [[ -f "/workspace/.claude/state.json" ]]; then
        SESSION_FILE=$(readlink -f /workspace/.claude/state.json 2>/dev/null || echo "/workspace/.claude/state.json")
    fi
fi

# Priorité 3: Dernière session (fallback)
if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    SESSION_DIR="$HOME/.claude/sessions"
    SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1 || true)
fi

# Lire l'état (schéma v3: .state)
STATE="unknown"
CURRENT_TASK=""
if [[ -f "$SESSION_FILE" ]]; then
    STATE=$(jq -r '.state // "unknown"' "$SESSION_FILE" 2>/dev/null || echo "unknown")
    CURRENT_TASK=$(jq -r '.currentTask // ""' "$SESSION_FILE" 2>/dev/null || echo "")
fi

# === PLAN MODE: Skip format/lint pour fichiers autorisés ===
if [[ "$STATE" == "planning" || "$STATE" == "planned" ]]; then
    # Fichiers autorisés en PLAN MODE (documentation, plans, sessions)
    if [[ "$FILE" == *".claude/plans/"* ]] || \
       [[ "$FILE" == *".claude/sessions/"* ]] || \
       [[ "$FILE" == */plans/* ]] || \
       [[ "$FILE" == *.md ]] || \
       [[ "$FILE" == /tmp/* ]] || \
       [[ "$FILE" == /home/vscode/.claude/* ]]; then
        # Skip silently - ces fichiers n'ont pas besoin de format/lint
        exit 0
    fi
fi

# === WIP Task Check ===
# Vérifie qu'une task est en WIP si on est en BYPASS mode (applying)
if [[ "$STATE" == "applying" && -z "$CURRENT_TASK" ]]; then
    echo "⚠️  POST-EDIT WARNING: Edit effectué sans task WIP active"
    echo "   Fichier: $FILE"
    echo "   State: applying"
    echo ""
    echo "   Rappel: En mode applying, démarrez une task avant d'éditer:"
    echo "   /home/vscode/.claude/scripts/task-start.sh <uuid>"
    echo ""
    # Log pour audit
    logger -t "claude-wip-check" "Edit without WIP task: $FILE" 2>/dev/null || true
fi

# === Format/Lint pipeline (seulement pour fichiers code) ===

# 1. Format
"$SCRIPT_DIR/format.sh" "$FILE"

# 2. Sort imports
"$SCRIPT_DIR/imports.sh" "$FILE"

# 3. Lint (with auto-fix)
"$SCRIPT_DIR/lint.sh" "$FILE"

exit 0
