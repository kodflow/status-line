#!/bin/bash
# task-epic.sh - Créer un epic dans Taskwarrior et mettre à jour la session
# Usage: task-epic.sh <project> <epic_number> <epic_name>
# Exemple: task-epic.sh "feat-login" 1 "Setup infrastructure"

set -e

# Vérifier Taskwarrior
if ! command -v task &>/dev/null; then
    echo "❌ Taskwarrior non installé"
    exit 1
fi

PROJECT="$1"
EPIC_NUM="$2"
EPIC_NAME="$3"

if [[ -z "$PROJECT" || -z "$EPIC_NUM" || -z "$EPIC_NAME" ]]; then
    echo "Usage: task-epic.sh <project> <epic_number> <epic_name>"
    echo "Exemple: task-epic.sh \"feat-login\" 1 \"Setup infrastructure\""
    exit 1
fi

# Créer l'epic dans Taskwarrior (rc.confirmation=off pour éviter les prompts)
OUTPUT=$(task rc.confirmation=off add project:"$PROJECT" "Epic $EPIC_NUM: $EPIC_NAME" +epic +planning epic:"$EPIC_NUM" 2>&1)
TASK_ID=$(echo "$OUTPUT" | grep -oP 'Created task \K\d+' || echo "")

if [[ -z "$TASK_ID" ]]; then
    echo "❌ Erreur création epic"
    echo "$OUTPUT" >&2
    exit 1
fi

# Récupérer l'UUID correctement (utiliser _get pour avoir JUSTE l'UUID)
EPIC_UUID=$(task rc.verbose=nothing _get "$TASK_ID".uuid 2>/dev/null)

if [[ -z "$EPIC_UUID" ]]; then
    echo "❌ Impossible de récupérer l'UUID de l'epic"
    exit 1
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

if [[ -f "$SESSION_FILE" ]]; then
    TMP_FILE=$(mktemp)
    jq --arg num "$EPIC_NUM" --arg name "$EPIC_NAME" --arg uuid "$EPIC_UUID" '
        .epics += [{
            "id": ($num | tonumber),
            "name": $name,
            "uuid": $uuid,
            "status": "TODO",
            "tasks": []
        }]
    ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"
fi

# Retourner l'UUID (pour utilisation par task-add.sh)
echo "$EPIC_UUID"
