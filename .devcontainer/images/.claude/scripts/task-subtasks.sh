#!/bin/bash
# Crée des sous-tâches dans Taskwarrior depuis le plan validé
# Usage: task-subtasks.sh <project> <plan_file>

set -euo pipefail

PROJECT="$1"
PLAN_FILE="$2"

SESSION_DIR="$HOME/.claude/sessions"
SESSION_FILE="$SESSION_DIR/$PROJECT.json"

if [[ ! -f "$SESSION_FILE" ]]; then
    echo "❌ Session non trouvée: $PROJECT"
    exit 1
fi

if [[ ! -f "$PLAN_FILE" ]]; then
    echo "❌ Plan non trouvé: $PLAN_FILE"
    exit 1
fi

# Lire la phase 2 (Implementation) UUID et ID
PHASE2_UUID=$(jq -r '.phases["2"].uuid' "$SESSION_FILE")
PHASE2_ID=$(jq -r '.phases["2"].id' "$SESSION_FILE")
BRANCH=$(jq -r '.branch' "$SESSION_FILE")

echo "Création des sous-tâches depuis le plan..."

SUBTASK_IDS=()
STEP_NUM=0

# Extraire les étapes du plan (lignes commençant par un numéro suivi d'un point)
# Format attendu: "1. Description de la tâche"
while IFS= read -r LINE; do
    STEP_NUM=$((STEP_NUM + 1))
    DESC=$(echo "$LINE" | sed 's/^[0-9]*\.\s*//')

    # Créer la sous-tâche
    SUBTASK_OUTPUT=$(task add "$DESC" \
        project:"$PROJECT" +claude +subtask phase:2 \
        model:sonnet parallel:no branch:"$BRANCH" 2>&1)

    SUBTASK_ID=$(echo "$SUBTASK_OUTPUT" | grep -oP 'Created task \K\d+' || echo "")

    if [[ -n "$SUBTASK_ID" ]]; then
        SUBTASK_UUID=$(task "$SUBTASK_ID" uuid 2>/dev/null || echo "")
        echo "  ✓ Sous-tâche $STEP_NUM: $DESC"
        SUBTASK_IDS+=("$SUBTASK_ID")
    else
        echo "  ⚠ Échec création: $DESC"
    fi
done < <(grep -E "^[0-9]+\." "$PLAN_FILE" 2>/dev/null || true)

# Mettre à jour la session avec les sous-tâches
if [[ ${#SUBTASK_IDS[@]} -gt 0 ]]; then
    SUBTASKS_JSON=$(printf '%s\n' "${SUBTASK_IDS[@]}" | jq -R 'tonumber' | jq -s '.')
    TMP_FILE=$(mktemp)
    jq ".phases[\"2\"].subtasks = $SUBTASKS_JSON" "$SESSION_FILE" > "$TMP_FILE"
    mv "$TMP_FILE" "$SESSION_FILE"

    echo ""
    echo "✓ ${#SUBTASK_IDS[@]} sous-tâches créées pour Phase 2"
else
    echo ""
    echo "⚠ Aucune sous-tâche créée (vérifiez le format du plan)"
    echo "  Format attendu: lignes commençant par '1. ', '2. ', etc."
fi
