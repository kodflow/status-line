#!/bin/bash
# task-check-locks.sh - Vérifier les locks actuels et potentiels conflits
# Usage: task-check-locks.sh [uuid]
#   Sans argument: affiche les locks actuels
#   Avec uuid: vérifie si la task peut démarrer (pas de conflit)

set -e

SESSION_DIR="$HOME/.claude/sessions"
SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1)

if [[ ! -f "$SESSION_FILE" ]]; then
    echo "Aucune session active"
    exit 0
fi

TASK_UUID="$1"

# Afficher les locks actuels
echo "═══════════════════════════════════════════════"
echo "  Locks actuels"
echo "═══════════════════════════════════════════════"
echo ""

CURRENT_LOCKS=$(jq -r '.lockedPaths // [] | .[]' "$SESSION_FILE" 2>/dev/null)
CURRENT_TASK=$(jq -r '.currentTask // "none"' "$SESSION_FILE" 2>/dev/null)

if [[ -z "$CURRENT_LOCKS" ]]; then
    echo "  Aucun lock actif"
else
    echo "  Task WIP: $CURRENT_TASK"
    echo ""
    echo "  Chemins verrouillés:"
    while IFS= read -r lock_line; do echo "    - $lock_line"; done <<< "$CURRENT_LOCKS"
fi

echo ""

# Si un UUID est fourni, vérifier les conflits
if [[ -n "$TASK_UUID" ]]; then
    echo "═══════════════════════════════════════════════"
    echo "  Vérification pour: $TASK_UUID"
    echo "═══════════════════════════════════════════════"
    echo ""

    # Récupérer les locks de la task
    TASK_LOCKS=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | \
        jq -r '.[0].annotations[]?.description // empty' | \
        grep '^ctx:' | sed 's/^ctx://' | jq -r '.locks // [] | .[]' 2>/dev/null || echo "")

    if [[ -z "$TASK_LOCKS" ]]; then
        echo "  Cette task n'a pas de locks définis"
        echo "  ✓ Peut démarrer sans conflit"
    else
        echo "  Locks demandés:"
        while IFS= read -r lock_line; do echo "    - $lock_line"; done <<< "$TASK_LOCKS"
        echo ""

        # Vérifier les conflits
        CONFLICT=""
        if [[ -n "$CURRENT_LOCKS" ]]; then
            while IFS= read -r new_lock; do
                [[ -z "$new_lock" ]] && continue
                while IFS= read -r existing_lock; do
                    [[ -z "$existing_lock" ]] && continue
                    if [[ "$new_lock" == "$existing_lock" ]] || \
                       [[ "$new_lock" == *"$existing_lock"* ]] || \
                       [[ "$existing_lock" == *"$new_lock"* ]]; then
                        CONFLICT="$new_lock ↔ $existing_lock"
                        break 2
                    fi
                done <<< "$CURRENT_LOCKS"
            done <<< "$TASK_LOCKS"
        fi

        if [[ -n "$CONFLICT" ]]; then
            echo "  ❌ CONFLIT DÉTECTÉ: $CONFLICT"
            echo ""
            echo "  Cette task ne peut pas démarrer tant que"
            echo "  la task $CURRENT_TASK n'est pas terminée."
            exit 1
        else
            echo "  ✓ Pas de conflit - peut démarrer"
        fi
    fi
fi

echo ""
echo "═══════════════════════════════════════════════"
