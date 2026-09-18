#!/bin/bash
# PreToolUse hook - Valide les phases obligatoires en PLAN MODE
# Empêche de sauter des phases du workflow de planification
# Exit 0 = autorisé, Exit 2 = bloqué

set -euo pipefail

# Lire l'input JSON de Claude
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')

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

# Si pas de session, autoriser (pas en mode workflow)
if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    exit 0
fi

# Lire l'état de la session
STATE=$(jq -r '.state // "unknown"' "$SESSION_FILE")
CURRENT_PHASE=$(jq -r '.currentPhase // 0' "$SESSION_FILE")
SCHEMA_VERSION=$(jq -r '.schemaVersion // 2' "$SESSION_FILE")

# Si pas en mode planning, autoriser
if [[ "$STATE" != "planning" ]]; then
    exit 0
fi

# === VALIDATION DES PHASES ===


# Vérifier les sauts de phase (si schéma v3+)
if [[ "$SCHEMA_VERSION" -ge 3 ]] && [[ "$CURRENT_PHASE" -gt 0 ]]; then
    COMPLETED_PHASES=$(jq -r '.completedPhases | map(.phase) | sort | .[]' "$SESSION_FILE" 2>/dev/null || echo "")
    
    # Vérifier que toutes les phases précédentes sont complétées
    for ((i=1; i<CURRENT_PHASE; i++)); do
        if ! echo "$COMPLETED_PHASES" | grep -q "^$i$"; then
            echo "═══════════════════════════════════════════════"
            echo "  🚫 BLOQUÉ - SAUT DE PHASE DÉTECTÉ"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  Phase courante: $CURRENT_PHASE"
            echo "  Phase manquante: $i"
            echo ""
            echo "  Les phases doivent être complétées dans l'ordre:"
            echo "    1 → 2 → 3 → 4 → 5 → validation → 6"
            echo ""
            echo "  Retournez à la phase $i pour continuer."
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
    done
fi

# Tout OK
exit 0
