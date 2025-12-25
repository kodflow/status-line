#!/bin/bash
# PostToolUse hook - Log l'action complétée
# Fonctionne pour Write, Edit, et Bash

set -e

# Sortie gracieuse si jq non disponible
if ! command -v jq &>/dev/null; then
    exit 0
fi

# Sortie gracieuse si task non disponible
if ! command -v task &>/dev/null; then
    exit 0
fi

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')
EXIT_CODE=$(echo "$INPUT" | jq -r '.tool_response.exit_code // 0')

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

# Si pas de session, log quand même pour Bash (autorisé sans tâche)
if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    exit 0
fi

# Schéma v3: currentTask (avec fallback sur current_task_uuid pour compatibilité)
TASK_UUID=$(jq -r '.currentTask // .current_task_uuid // empty' "$SESSION_FILE")
[[ -z "$TASK_UUID" ]] && exit 0

TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Construire l'événement selon l'outil
case "$TOOL" in
    Write|Edit)
        FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // "N/A"')
        if [[ "$EXIT_CODE" == "0" && -f "$FILE_PATH" ]]; then
            LINES=$(wc -l < "$FILE_PATH" 2>/dev/null || echo "0")
            EXT="${FILE_PATH##*.}"
            EVENT="{\"type\":\"done\",\"ts\":\"$TIMESTAMP\",\"tool\":\"$TOOL\",\"file\":\"$FILE_PATH\",\"ext\":\"$EXT\",\"lines\":$LINES}"
        else
            EVENT="{\"type\":\"error\",\"ts\":\"$TIMESTAMP\",\"tool\":\"$TOOL\",\"file\":\"$FILE_PATH\",\"exit\":$EXIT_CODE}"
        fi
        ;;
    Bash)
        CMD=$(echo "$INPUT" | jq -r '.tool_input.command // "unknown"' | head -c 100)
        if [[ "$EXIT_CODE" == "0" ]]; then
            EVENT="{\"type\":\"done\",\"ts\":\"$TIMESTAMP\",\"tool\":\"Bash\",\"cmd\":\"$CMD\"}"
        else
            STDERR=$(echo "$INPUT" | jq -r '.tool_response.stderr // ""' | head -c 200 | tr '\n' ' ')
            EVENT="{\"type\":\"error\",\"ts\":\"$TIMESTAMP\",\"tool\":\"Bash\",\"cmd\":\"$CMD\",\"exit\":$EXIT_CODE,\"err\":\"$STDERR\"}"
        fi
        ;;
    *)
        EVENT="{\"type\":\"done\",\"ts\":\"$TIMESTAMP\",\"tool\":\"$TOOL\"}"
        ;;
esac

# Logger l'événement dans Taskwarrior
task uuid:"$TASK_UUID" annotate "post:$EVENT" 2>/dev/null || true

# Mettre à jour la session (avec fallback pour .actions)
TMP_FILE=$(mktemp)
jq --arg ts "$TIMESTAMP" '
    .actions = ((.actions // 0) + 1) |
    .lastAction = $ts
' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"

exit 0
