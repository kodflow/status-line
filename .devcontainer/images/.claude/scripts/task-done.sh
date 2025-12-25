#!/bin/bash
# task-done.sh - Terminer une task (WIP → DONE)
# Usage: task-done.sh <uuid>
# Met à jour la session JSON et marque la task comme terminée

set -e

# Vérifier Taskwarrior
if ! command -v task &>/dev/null; then
    echo "❌ Taskwarrior non installé"
    exit 1
fi

TASK_UUID="$1"

if [[ -z "$TASK_UUID" ]]; then
    echo "Usage: task-done.sh <uuid>"
    exit 1
fi

# Vérifier que la task existe
if ! task rc.confirmation=off uuid:"$TASK_UUID" info >/dev/null 2>&1; then
    echo "❌ Task non trouvée: $TASK_UUID"
    exit 1
fi

# Récupérer les infos de la task AVANT de la terminer
TASK_DATA=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0]')
PROJECT=$(echo "$TASK_DATA" | jq -r '.project // ""')
EPIC_NUM=$(echo "$TASK_DATA" | jq -r '.epic // ""')
TASK_DESC=$(echo "$TASK_DATA" | jq -r '.description // "Unknown"')

# Marquer comme terminée dans Taskwarrior
task rc.confirmation=off uuid:"$TASK_UUID" done >/dev/null 2>&1 || true

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
    # Récupérer les locks de la task terminée pour les retirer
    TASK_LOCKS=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | \
        jq -r '.[0].annotations[]?.description // empty' | \
        grep '^ctx:' | sed 's/^ctx://' | jq -r '.locks // [] | .[]' 2>/dev/null || echo "")

    # Construire le tableau JSON des locks à retirer
    LOCKS_TO_REMOVE="[]"
    if [[ -n "$TASK_LOCKS" ]]; then
        LOCKS_TO_REMOVE=$(echo "$TASK_LOCKS" | jq -R -s 'split("\n") | map(select(length > 0))')
    fi

    # Mettre à jour state.json: status DONE, reset currentTask, retirer locks
    # Note: (.actions // 0) pour éviter erreur si champ absent
    TMP_FILE=$(mktemp)
    jq --arg uuid "$TASK_UUID" --argjson locksToRemove "$LOCKS_TO_REMOVE" '
        .currentTask = null |
        .lockedPaths = (.lockedPaths - $locksToRemove) |
        .lastAction = (now | todate) |
        .actions = ((.actions // 0) + 1) |
        (.epics[].tasks[] | select(.uuid == $uuid)).status = "DONE"
    ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"
fi

# === Auto-close Epic dans Taskwarrior ===
# Si la task a un epic, vérifier si toutes les tasks de l'epic sont terminées
EPIC_CLOSED=false
if [[ -n "$PROJECT" && -n "$EPIC_NUM" ]]; then
    # Compter les tasks non terminées pour cet epic (excluant l'epic lui-même)
    REMAINING=$(task rc.confirmation=off project:"$PROJECT" epic:"$EPIC_NUM" +task status:pending count 2>/dev/null || echo "0")

    if [[ "$REMAINING" == "0" ]]; then
        # Trouver et fermer l'epic parent dans Taskwarrior
        EPIC_UUID=$(task rc.confirmation=off project:"$PROJECT" epic:"$EPIC_NUM" +epic status:pending _uuids 2>/dev/null | head -1)
        if [[ -n "$EPIC_UUID" ]]; then
            task rc.confirmation=off uuid:"$EPIC_UUID" done >/dev/null 2>&1 || true
            EPIC_CLOSED=true

            # Mettre à jour le status de l'epic dans la session
            if [[ -f "$SESSION_FILE" ]]; then
                TMP_FILE=$(mktemp)
                jq --arg epic "$EPIC_NUM" '
                    (.epics[] | select(.id == ($epic | tonumber))).status = "DONE"
                ' "$SESSION_FILE" > "$TMP_FILE" 2>/dev/null && mv "$TMP_FILE" "$SESSION_FILE"
            fi
        fi
    fi
fi

# Afficher résultat
echo "✓ Task terminée: $TASK_DESC"
if [[ "$EPIC_CLOSED" == "true" ]]; then
    echo "✓ Epic $EPIC_NUM auto-fermé (toutes les tasks terminées)"
fi
