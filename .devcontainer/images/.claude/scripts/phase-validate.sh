#!/bin/bash
# PreToolUse hook - Valide les phases obligatoires en PLAN MODE
# Empêche de sauter des phases ou d'écrire dans Taskwarrior sans validation
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

# Si pas en mode planning, autoriser (vérification task-validate.sh s'en charge)
if [[ "$STATE" != "planning" ]]; then
    exit 0
fi

# === VALIDATION DES PHASES ===

# Vérifier si on essaie d'appeler les scripts Taskwarrior
if [[ "$TOOL" == "Bash" ]]; then
    COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""')
    
    # Bloquer task-epic.sh et task-add.sh sans phases 1-5 complétées
    if [[ "$COMMAND" == *"task-epic.sh"* ]] || [[ "$COMMAND" == *"task-add.sh"* ]]; then
        # Vérifier que les phases 1-5 sont complétées
        COMPLETED_COUNT=$(jq -r '.completedPhases | length // 0' "$SESSION_FILE")
        
        if [[ "$COMPLETED_COUNT" -lt 5 ]]; then
            echo "═══════════════════════════════════════════════"
            echo "  🚫 BLOQUÉ - PHASES OBLIGATOIRES"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  Action: Écriture Taskwarrior"
            echo "  Commande: $COMMAND"
            echo ""
            echo "  Phases complétées: $COMPLETED_COUNT/5"
            echo ""
            echo "  Les phases suivantes doivent être complétées:"
            echo "    1. Analyse de la demande"
            echo "    2. Recherche documentation"
            echo "    3. Analyse projet existant"
            echo "    4. Affûtage"
            echo "    5. Définition épics/tasks + VALIDATION"
            echo ""
            echo "  La phase 6 (écriture Taskwarrior) nécessite"
            echo "  la validation utilisateur en phase 5."
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
        
        # Vérifier que la validation utilisateur a eu lieu
        VALIDATED=$(jq -r '.validated // false' "$SESSION_FILE")
        if [[ "$VALIDATED" != "true" ]]; then
            echo "═══════════════════════════════════════════════"
            echo "  🚫 BLOQUÉ - VALIDATION REQUISE"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  L'écriture dans Taskwarrior nécessite"
            echo "  la validation utilisateur."
            echo ""
            echo "  Utilisez AskUserQuestion pour valider"
            echo "  le plan avant de créer les épics/tasks."
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
    fi
fi

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
