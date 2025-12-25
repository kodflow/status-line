#!/bin/bash
# session-validate.sh - Validation utilisateur stricte
# Usage: session-validate.sh --approve | --reject [feedback]
#
# Garantit:
# - Validation uniquement en phase 5
# - Enregistrement du token de validation
# - Historique des validations
#
# Exit 0 = succès, Exit 1 = erreur, Exit 2 = validation invalide

set -euo pipefail

# === Trouver la session active (déterministe) ===
find_session() {
    local session_file=""
    
    if [[ -f "/workspace/.claude/active-session" ]]; then
        session_file=$(cat /workspace/.claude/active-session 2>/dev/null || true)
    fi
    
    if [[ -z "$session_file" || ! -f "$session_file" ]]; then
        if [[ -f "/workspace/.claude/state.json" ]]; then
            session_file=$(readlink -f /workspace/.claude/state.json 2>/dev/null || echo "/workspace/.claude/state.json")
        fi
    fi
    
    if [[ -z "$session_file" || ! -f "$session_file" ]]; then
        local session_dir="$HOME/.claude/sessions"
        session_file=$(ls -t "$session_dir"/*.json 2>/dev/null | head -1 || true)
    fi
    
    echo "$session_file"
}

# === Valider pré-conditions ===
check_preconditions() {
    local session_file="$1"
    
    if [[ ! -f "$session_file" ]]; then
        echo "❌ Aucune session active"
        exit 1
    fi
    
    local schema_version
    schema_version=$(jq -r '.schemaVersion // 0' "$session_file")
    if [[ "$schema_version" -lt 3 ]]; then
        echo "❌ schemaVersion incompatible ($schema_version < 3)"
        exit 1
    fi
    
    local state
    state=$(jq -r '.state // "unknown"' "$session_file")
    if [[ "$state" != "planning" ]]; then
        echo "❌ Validation uniquement en state=planning (actuel: $state)"
        exit 2
    fi
    
    local current_phase
    current_phase=$(jq -r '.currentPhase // 0' "$session_file")
    if [[ "$current_phase" -ne 5 ]]; then
        echo "❌ Validation uniquement en phase 5 (actuel: phase $current_phase)"
        exit 2
    fi
    
    # Vérifier que phases 1-4 sont complétées
    local completed_count
    completed_count=$(jq '.completedPhases | length' "$session_file")
    if [[ "$completed_count" -lt 4 ]]; then
        echo "❌ Phases 1-4 doivent être complétées ($completed_count/4)"
        exit 2
    fi
    
    # Vérifier qu'il y a des epics définis
    local epics_count
    epics_count=$(jq '.epics | length' "$session_file")
    if [[ "$epics_count" -eq 0 ]]; then
        echo "❌ Au moins 1 epic doit être défini avant validation"
        exit 2
    fi
}

# === Approuver le plan ===
approve() {
    local session_file="$1"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    # Générer un token de validation (hash simple pour traçabilité)
    local validation_token
    validation_token=$(echo "$timestamp-$session_file-$$" | sha256sum | cut -c1-16)
    
    # Récupérer stats pour le log
    local epics_count tasks_count
    epics_count=$(jq '.epics | length' "$session_file")
    tasks_count=$(jq '[.epics[].tasks | length] | add // 0' "$session_file")
    
    # Appliquer validation
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg ts "$timestamp" --arg token "$validation_token" --argjson epics "$epics_count" --argjson tasks "$tasks_count" '
        .validated = true |
        .validatedAt = $ts |
        .validationToken = $token |
        .validationHistory = ((.validationHistory // []) + [{
            "at": $ts,
            "result": "approved",
            "token": $token,
            "epics": $epics,
            "tasks": $tasks
        }]) |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "user_validate",
            "result": "approved",
            "token": $token
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "═══════════════════════════════════════════════"
    echo "  ✓ Plan validé par l'utilisateur"
    echo "═══════════════════════════════════════════════"
    echo ""
    echo "  Token     : $validation_token"
    echo "  Timestamp : $timestamp"
    echo "  Epics     : $epics_count"
    echo "  Tasks     : $tasks_count"
    echo ""
    echo "  Prochaines étapes:"
    echo "    1. session-transition.sh --complete-phase 5"
    echo "    2. Écrire epics/tasks dans Taskwarrior"
    echo "    3. session-transition.sh --finalize"
    echo "    4. /apply"
    echo ""
    echo "═══════════════════════════════════════════════"
}

# === Rejeter le plan ===
reject() {
    local session_file="$1"
    local feedback="${2:-No feedback provided}"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    # Sauvegarder état avant reset
    local epics_count completed_phases
    epics_count=$(jq '.epics | length' "$session_file")
    completed_phases=$(jq '.completedPhases | length' "$session_file")
    
    # Appliquer rejet + reset complet
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg ts "$timestamp" --arg fb "$feedback" --argjson epics "$epics_count" --argjson phases "$completed_phases" '
        .validationHistory = ((.validationHistory // []) + [{
            "at": $ts,
            "result": "rejected",
            "feedback": $fb,
            "previousEpics": $epics,
            "previousPhases": $phases
        }]) |
        .rejectionHistory = ((.rejectionHistory // []) + [{
            "at": $ts,
            "feedback": $fb,
            "previousPhases": .completedPhases,
            "previousEpics": $epics
        }]) |
        .completedPhases = [] |
        .currentPhase = 1 |
        .validated = false |
        .validatedAt = null |
        .validationToken = null |
        .epics = [] |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "user_validate",
            "result": "rejected",
            "feedback": $fb
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "═══════════════════════════════════════════════"
    echo "  ⚠ Plan rejeté - Reset Phase 1"
    echo "═══════════════════════════════════════════════"
    echo ""
    echo "  Feedback : $feedback"
    echo "  Timestamp: $timestamp"
    echo ""
    echo "  État reset:"
    echo "    - completedPhases: []"
    echo "    - currentPhase: 1"
    echo "    - epics: [] (vidé)"
    echo "    - Historique sauvegardé"
    echo ""
    echo "  → Ré-analyse complète requise"
    echo ""
    echo "═══════════════════════════════════════════════"
}

# === Afficher statut validation ===
status() {
    local session_file="$1"
    
    echo "═══════════════════════════════════════════════"
    echo "  État de validation"
    echo "═══════════════════════════════════════════════"
    echo ""
    
    jq -r '
        "  Phase courante : \(.currentPhase // 0)",
        "  Phases complétées : \(.completedPhases | length)",
        "  Validated : \(.validated // false)",
        "  Token : \(.validationToken // "none")",
        "  ValidatedAt : \(.validatedAt // "never")",
        "",
        "  Historique validations : \((.validationHistory // []) | length)",
        "  Historique rejets : \((.rejectionHistory // []) | length)"
    ' "$session_file"
    
    echo ""
    echo "═══════════════════════════════════════════════"
}

# === Main ===
main() {
    if [[ $# -eq 0 ]]; then
        echo "Usage: session-validate.sh [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --approve           Valider le plan (requiert phase 5)"
        echo "  --reject [feedback] Rejeter + reset phase 1"
        echo "  --status            Afficher état validation"
        exit 1
    fi
    
    local session_file
    session_file=$(find_session)
    
    case "$1" in
        --approve)
            check_preconditions "$session_file"
            approve "$session_file"
            ;;
        --reject)
            check_preconditions "$session_file"
            reject "$session_file" "${2:-}"
            ;;
        --status)
            if [[ ! -f "$session_file" ]]; then
                echo "❌ Aucune session active"
                exit 1
            fi
            status "$session_file"
            ;;
        *)
            echo "❌ Option inconnue: $1"
            exit 1
            ;;
    esac
}

main "$@"
