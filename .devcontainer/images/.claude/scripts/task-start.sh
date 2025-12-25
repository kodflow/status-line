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

# === Vérification des locks AVANT démarrage ===
if [[ -f "$SESSION_FILE" ]]; then
    # Récupérer les locks de la task depuis les annotations ctx
    TASK_LOCKS=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | \
        jq -r '.[0].annotations[]?.description // empty' | \
        grep '^ctx:' | sed 's/^ctx://' | jq -r '.locks // [] | .[]' 2>/dev/null || echo "")

    # Récupérer les locks actuels dans state.json
    CURRENT_LOCKS=$(jq -r '.lockedPaths // [] | .[]' "$SESSION_FILE" 2>/dev/null || echo "")

    # Vérifier les conflits de locks
    if [[ -n "$TASK_LOCKS" && -n "$CURRENT_LOCKS" ]]; then
        CONFLICT=""
        while IFS= read -r new_lock; do
            [[ -z "$new_lock" ]] && continue
            while IFS= read -r existing_lock; do
                [[ -z "$existing_lock" ]] && continue
                # Vérifier si les locks se chevauchent (match exact ou pattern)
                if [[ "$new_lock" == "$existing_lock" ]] || \
                   [[ "$new_lock" == *"$existing_lock"* ]] || \
                   [[ "$existing_lock" == *"$new_lock"* ]]; then
                    CONFLICT="$new_lock (conflit avec: $existing_lock)"
                    break 2
                fi
            done <<< "$CURRENT_LOCKS"
        done <<< "$TASK_LOCKS"

        if [[ -n "$CONFLICT" ]]; then
            echo "═══════════════════════════════════════════════"
            echo "  ❌ BLOQUÉ: Conflit de locks"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  Task: $TASK_DESC"
            echo "  Conflit: $CONFLICT"
            echo ""
            echo "  Une autre task WIP verrouille ces fichiers."
            echo "  Terminez d'abord la task en cours avant de"
            echo "  démarrer celle-ci."
            echo ""
            echo "  Locks actuels:"
            while IFS= read -r lock_line; do echo "    $lock_line"; done <<< "$CURRENT_LOCKS"
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 1
        fi
    fi
fi

# Démarrer la task dans Taskwarrior
task rc.confirmation=off uuid:"$TASK_UUID" start >/dev/null 2>&1 || true

if [[ -f "$SESSION_FILE" ]]; then
    # Récupérer les locks de la task (déjà fait ci-dessus, mais on refait pour être sûr)
    TASK_LOCKS=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | \
        jq -r '.[0].annotations[]?.description // empty' | \
        grep '^ctx:' | sed 's/^ctx://' | jq -r '.locks // [] | .[]' 2>/dev/null || echo "")

    # Construire le tableau JSON des locks
    LOCKS_JSON="[]"
    if [[ -n "$TASK_LOCKS" ]]; then
        LOCKS_JSON=$(echo "$TASK_LOCKS" | jq -R -s 'split("\n") | map(select(length > 0))')
    fi

    # Mettre à jour state.json (schéma v3: utilise .state)
    TMP_FILE=$(mktemp)
    jq --arg uuid "$TASK_UUID" --arg epic "$EPIC_NUM" --argjson locks "$LOCKS_JSON" '
        # Transition: planned -> applying (si pas déjà)
        (if .state == "planned" then .state = "applying" else . end) |
        .currentTask = $uuid |
        .currentEpic = ($epic | tonumber) |
        .lockedPaths = ((.lockedPaths // []) + $locks | unique) |
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
