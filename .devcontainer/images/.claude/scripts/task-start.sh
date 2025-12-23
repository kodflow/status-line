#!/bin/bash
# task-start.sh - Démarrer une task (TODO → WIP)
# Usage: task-start.sh <uuid>
# Met à jour la session JSON et démarre la task dans Taskwarrior

set -e

# Vérifier Taskwarrior
if ! command -v task &>/dev/null; then
    echo "❌ Taskwarrior non installé"
    exit 1
fi

TASK_UUID="$1"

if [[ -z "$TASK_UUID" ]]; then
    echo "Usage: task-start.sh <uuid>"
    exit 1
fi

# Vérifier que la task existe
if ! task rc.confirmation=off uuid:"$TASK_UUID" info >/dev/null 2>&1; then
    echo "❌ Task non trouvée: $TASK_UUID"
    exit 1
fi

# Récupérer les infos de la task
TASK_DATA=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0]')
TASK_DESC=$(echo "$TASK_DATA" | jq -r '.description // "Unknown"')
EPIC_NUM=$(echo "$TASK_DATA" | jq -r '.epic // 1')

# Démarrer la task dans Taskwarrior
task rc.confirmation=off uuid:"$TASK_UUID" start >/dev/null 2>&1 || true

# Mettre à jour la session si elle existe
SESSION_DIR="$HOME/.claude/sessions"
SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1)

if [[ -f "$SESSION_FILE" ]]; then
    # Mettre à jour la session avec l'UUID et le status
    TMP_FILE=$(mktemp)
    jq --arg uuid "$TASK_UUID" --arg epic "$EPIC_NUM" '
        .mode = "bypass" |
        .current_task_uuid = $uuid |
        .current_epic = ($epic | tonumber) |
        (.epics[].tasks[] | select(.uuid == $uuid)).status = "WIP"
    ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"

    # Mettre à jour l'epic en WIP si pas déjà
    TMP_FILE=$(mktemp)
    jq --arg epic "$EPIC_NUM" '
        (.epics[] | select(.id == ($epic | tonumber) and .status == "TODO")).status = "WIP"
    ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"
fi

# Afficher info
echo "▶ Task démarrée: $TASK_DESC"
echo "  UUID: $TASK_UUID"
