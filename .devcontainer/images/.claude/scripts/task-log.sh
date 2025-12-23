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

# Trouver la session active
SESSION_DIR="$HOME/.claude/sessions"
SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1)

# Si pas de session, log quand même pour Bash (autorisé sans tâche)
if [[ ! -f "$SESSION_FILE" ]]; then
    exit 0
fi

TASK_UUID=$(jq -r '.current_task_uuid // empty' "$SESSION_FILE")
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

# Mettre à jour la session
ACTIONS=$(jq -r '.actions // 0' "$SESSION_FILE")
ACTIONS=$((ACTIONS + 1))
LAST_ACTION="$TIMESTAMP"

# Mise à jour atomique du fichier session
TMP_FILE=$(mktemp)
jq ".actions = $ACTIONS | .last_action = \"$LAST_ACTION\"" "$SESSION_FILE" > "$TMP_FILE"
mv "$TMP_FILE" "$SESSION_FILE"

exit 0
