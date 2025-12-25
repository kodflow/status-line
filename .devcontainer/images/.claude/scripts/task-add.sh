#!/bin/bash
# task-add.sh - Ajouter une task à un epic et mettre à jour la session
# Usage: task-add.sh <project> <epic_num> <epic_uuid> <task_name> [parallel:yes|no] [ctx:JSON]
# Exemple: task-add.sh "feat-login" 1 "uuid-xxx" "Créer AuthService" "no" '{"files":["src/auth.ts"]}'

set -e

# Vérifier Taskwarrior
if ! command -v task &>/dev/null; then
    echo "❌ Taskwarrior non installé"
    exit 1
fi

PROJECT="$1"
EPIC_NUM="$2"
EPIC_UUID="$3"
TASK_NAME="$4"
PARALLEL="${5:-no}"
CTX_JSON="${6:-}"

if [[ -z "$PROJECT" || -z "$EPIC_NUM" || -z "$EPIC_UUID" || -z "$TASK_NAME" ]]; then
    echo "Usage: task-add.sh <project> <epic_num> <epic_uuid> <task_name> [parallel] [ctx_json]"
    exit 1
fi

# Créer la task dans Taskwarrior (rc.confirmation=off pour éviter les prompts)
# Note: "parent" est un mot réservé dans Taskwarrior, on utilise "epic_uuid" à la place
OUTPUT=$(task rc.confirmation=off add project:"$PROJECT" "$TASK_NAME" +task epic:"$EPIC_NUM" epic_uuid:"$EPIC_UUID" parallel:"$PARALLEL" 2>&1)
TASK_ID=$(echo "$OUTPUT" | grep -oP 'Created task \K\d+' || echo "")

if [[ -z "$TASK_ID" ]]; then
    echo "❌ Erreur création task"
    echo "$OUTPUT" >&2
    exit 1
fi

# Récupérer l'UUID correctement (utiliser _get pour avoir JUSTE l'UUID)
TASK_UUID=$(task rc.verbose=nothing _get "$TASK_ID".uuid 2>/dev/null)

if [[ -z "$TASK_UUID" ]]; then
    echo "❌ Impossible de récupérer l'UUID de la task"
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

TASK_NUM=1
if [[ -f "$SESSION_FILE" ]]; then
    # Compter les tasks existantes dans cet epic
    EXISTING_TASKS=$(jq --arg epic "$EPIC_NUM" '
        (.epics[] | select(.id == ($epic | tonumber))).tasks // [] | length
    ' "$SESSION_FILE" 2>/dev/null || echo "0")
    TASK_NUM=$((EXISTING_TASKS + 1))
fi
EXTERNAL_ID="T${EPIC_NUM}.${TASK_NUM}"

# Ajouter l'external_id comme annotation
task rc.confirmation=off uuid:"$TASK_UUID" annotate "external_id:$EXTERNAL_ID" >/dev/null 2>&1 || true

# Ajouter le contexte JSON si fourni
if [[ -n "$CTX_JSON" ]]; then
    task rc.confirmation=off uuid:"$TASK_UUID" annotate "ctx:$CTX_JSON" >/dev/null 2>&1 || true
fi

# Mettre à jour la session JSON avec la task créée
if [[ -f "$SESSION_FILE" ]]; then
    TMP_FILE=$(mktemp)
    jq --arg epic_num "$EPIC_NUM" --arg name "$TASK_NAME" --arg uuid "$TASK_UUID" --arg parallel "$PARALLEL" --arg ext_id "$EXTERNAL_ID" '
        (.epics[] | select(.id == ($epic_num | tonumber))).tasks += [{
            "name": $name,
            "uuid": $uuid,
            "externalId": $ext_id,
            "status": "TODO",
            "parallel": $parallel
        }]
    ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"
fi

# Retourner l'UUID
echo "$TASK_UUID"
