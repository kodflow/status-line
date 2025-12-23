#!/bin/bash
# PreToolUse hook - Valide qu'une tâche est active
# UNIQUEMENT pour Write|Edit - Bash est autorisé sans tâche
# Exit 0 = autorisé, Exit 2 = bloqué

set -euo pipefail

# Vérifier que Taskwarrior est installé
if ! command -v task &>/dev/null; then
    echo "⚠️  Taskwarrior non installé - validation désactivée"
    echo "→ Pour activer le suivi obligatoire: /update"
    exit 0  # Autoriser quand même (dégradé graceful)
fi

# Lire l'input JSON de Claude
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')

# Trouver la session active (cherche dans .claude/sessions/)
SESSION_DIR="$HOME/.claude/sessions"
SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1)

# Si pas de session, BLOQUER Write/Edit
if [[ ! -f "$SESSION_FILE" ]]; then
    echo "❌ BLOQUÉ: Aucune tâche Taskwarrior active."
    echo ""
    echo "→ Utilisez /feature <description> ou /fix <description>"
    echo "  pour démarrer un workflow avec suivi obligatoire."
    exit 2
fi

TASK_UUID=$(jq -r '.current_task_uuid // empty' "$SESSION_FILE")
PROJECT=$(jq -r '.project // "unknown"' "$SESSION_FILE")

if [[ -z "$TASK_UUID" ]]; then
    echo "❌ BLOQUÉ: Session corrompue - aucune tâche courante"
    exit 2
fi

# Vérifier que la tâche existe et est active
TASK_STATUS=$(task uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0].status // "unknown"')

if [[ "$TASK_STATUS" != "pending" ]]; then
    echo "❌ BLOQUÉ: Tâche terminée ou inexistante (status: $TASK_STATUS)"
    echo "→ Utilisez /feature --continue pour reprendre"
    exit 2
fi

# Vérifier que la tâche n'est pas bloquée par des dépendances
BLOCKED=$(task uuid:"$TASK_UUID" +BLOCKED count 2>/dev/null || echo "0")
if [[ "$BLOCKED" -gt 0 ]]; then
    DEPS=$(task uuid:"$TASK_UUID" depends 2>/dev/null | head -1)
    echo "❌ BLOQUÉ: Cette tâche dépend de tâches non terminées"
    echo "→ Dépendances: $DEPS"
    echo "→ Terminez d'abord les tâches précédentes"
    exit 2
fi

# Log l'action à venir (pré-événement)
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // "N/A"')

task uuid:"$TASK_UUID" annotate "pre:{\"ts\":\"$TIMESTAMP\",\"tool\":\"$TOOL\",\"file\":\"$FILE_PATH\"}" 2>/dev/null

# Afficher confirmation
TASK_DESC=$(task uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0].description // "Unknown"')
echo "✓ Projet: $PROJECT"
echo "✓ Tâche: $TASK_DESC"
exit 0
