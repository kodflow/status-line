#!/bin/bash
# PreToolUse hook - Valide qu'une tâche est active
# UNIQUEMENT pour Write|Edit - Bash est géré par bash-validate.sh
# Exit 0 = autorisé, Exit 2 = bloqué
#
# Logique basée sur .state (pas .mode) :
# - planning/planned : PLAN MODE (lecture seule + exceptions docs)
# - applying : BYPASS MODE (task WIP requise)
# - applied : terminé

set -euo pipefail

# Lire l'input JSON de Claude
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // "N/A"')

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

# Si pas de session, bloquer Write/Edit
if [[ ! -f "$SESSION_FILE" ]]; then
    echo "❌ BLOQUÉ: Aucune session active."
    echo ""
    echo "→ Utilisez /plan <description> pour démarrer un workflow."
    exit 2
fi

# === Lire l'état depuis .state ===
STATE=$(jq -r '.state // "unknown"' "$SESSION_FILE")
PROJECT=$(jq -r '.project // "unknown"' "$SESSION_FILE")

# === PLAN MODE : state = planning ou planned ===
if [[ "$STATE" == "planning" || "$STATE" == "planned" ]]; then
    # Liste des chemins autorisés en PLAN MODE
    ALLOWED_PATTERNS=(
        ".claude/plans/"
        ".claude/sessions/"
        "/plans/"
        "*.md"
        "/tmp/"
        "/home/vscode/.claude/"
    )

    IS_ALLOWED=false
    for pattern in "${ALLOWED_PATTERNS[@]}"; do
        if [[ "$FILE_PATH" == *"$pattern"* ]] || [[ "$FILE_PATH" == $pattern ]]; then
            IS_ALLOWED=true
            break
        fi
    done

    if [[ "$IS_ALLOWED" == "false" ]]; then
        echo "═══════════════════════════════════════════════"
        echo "  🚫 BLOQUÉ - PLAN MODE ACTIF"
        echo "═══════════════════════════════════════════════"
        echo ""
        echo "  État   : $STATE"
        echo "  Fichier: $FILE_PATH"
        echo "  Outil  : $TOOL"
        echo ""
        echo "  En PLAN MODE, seuls ces chemins sont autorisés:"
        echo "    - .claude/plans/*"
        echo "    - .claude/sessions/*"
        echo "    - *.md (documentation)"
        echo "    - /tmp/*"
        echo ""
        echo "  Pour passer en mode exécution:"
        echo "    1. Terminez le planning (phases 1-5)"
        echo "    2. Obtenez la validation utilisateur"
        echo "    3. Créez les tasks dans Taskwarrior (phase 6)"
        echo "    4. Exécutez /apply"
        echo ""
        echo "═══════════════════════════════════════════════"
        exit 2
    fi

    # En PLAN MODE avec fichier autorisé
    echo "✓ PLAN MODE: Écriture autorisée sur $FILE_PATH"
    exit 0
fi

# === BYPASS MODE : state = applying ===
if [[ "$STATE" == "applying" ]]; then
    # Vérifier que Taskwarrior est installé
    if ! command -v task &>/dev/null; then
        echo "⚠️  Taskwarrior non installé - validation désactivée"
        exit 0
    fi

    # Vérifier qu'une task est en cours (currentTask)
    TASK_UUID=$(jq -r '.currentTask // empty' "$SESSION_FILE")

    if [[ -z "$TASK_UUID" || "$TASK_UUID" == "null" ]]; then
        echo "❌ BLOQUÉ: Aucune task active (state=applying mais currentTask=null)"
        echo ""
        echo "→ Utilisez task-start.sh <uuid> pour démarrer une task"
        exit 2
    fi

    # Vérifier que la tâche existe et est active
    TASK_STATUS=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0].status // "unknown"')

    if [[ "$TASK_STATUS" != "pending" ]]; then
        echo "❌ BLOQUÉ: Task terminée ou inexistante (status: $TASK_STATUS)"
        echo "→ Utilisez task-start.sh pour démarrer la prochaine task"
        exit 2
    fi

    # Vérifier que la tâche n'est pas bloquée par des dépendances
    BLOCKED=$(task rc.confirmation=off uuid:"$TASK_UUID" +BLOCKED count 2>/dev/null || echo "0")
    if [[ "$BLOCKED" -gt 0 ]]; then
        echo "❌ BLOQUÉ: Cette task dépend de tasks non terminées"
        exit 2
    fi

    # Log l'action à venir
    TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    task rc.confirmation=off uuid:"$TASK_UUID" annotate "pre:{\"ts\":\"$TIMESTAMP\",\"tool\":\"$TOOL\",\"file\":\"$FILE_PATH\"}" 2>/dev/null || true

    # Afficher confirmation
    TASK_DESC=$(task rc.confirmation=off uuid:"$TASK_UUID" export 2>/dev/null | jq -r '.[0].description // "Unknown"')
    echo "✓ Projet: $PROJECT"
    echo "✓ Task: $TASK_DESC"
    exit 0
fi

# === STATE = applied : terminé, autoriser ===
if [[ "$STATE" == "applied" ]]; then
    echo "✓ State=applied: Session terminée, édition autorisée"
    exit 0
fi

# === État inconnu ===
echo "⚠️  État inconnu: $STATE - autorisation par défaut"
exit 0
