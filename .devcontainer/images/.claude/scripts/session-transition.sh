#!/bin/bash
# session-transition.sh - Transitions atomiques de phases
# Usage: session-transition.sh --complete-phase N | --to-phase N | --reset
#
# Garantit:
# - Session active valide (schemaVersion >= 3)
# - Transitions autorisées uniquement (pas de saut, pas de régression)
# - Cohérence completedPhases vs currentPhase
# - Workspace Git propre avant finalize
# - Log actions[] avec événements horodatés
#
# Exit 0 = succès, Exit 1 = erreur, Exit 2 = transition invalide

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

# === Valider la session ===
validate_session() {
    local session_file="$1"
    
    if [[ ! -f "$session_file" ]]; then
        echo "❌ Aucune session active trouvée"
        exit 1
    fi
    
    local schema_version
    schema_version=$(jq -r '.schemaVersion // 0' "$session_file")
    
    if [[ "$schema_version" -lt 3 ]]; then
        echo "❌ Session incompatible: schemaVersion=$schema_version (requis >= 3)"
        exit 1
    fi
}

# === Vérifier la cohérence branche/session ===
check_branch_coherence() {
    local session_file="$1"
    local session_branch
    session_branch=$(jq -r '.branch // ""' "$session_file")
    
    local current_branch
    current_branch=$(git branch --show-current 2>/dev/null || echo "")
    
    if [[ -n "$session_branch" && -n "$current_branch" && "$session_branch" != "$current_branch" ]]; then
        echo "⚠️  Incohérence: branche=$current_branch mais session.branch=$session_branch"
        echo "→ Changez de branche ou détruisez la session"
        exit 2
    fi
}

# === Vérifier workspace Git propre ===
check_workspace_clean() {
    # Vérifier si le workspace est propre
    if ! git diff --quiet 2>/dev/null; then
        echo "⚠️  Workspace Git non propre (modifications non commitées)"
        echo "→ Commitez vos changements avant de finaliser"
        exit 2
    fi
    
    if ! git diff --cached --quiet 2>/dev/null; then
        echo "⚠️  Modifications stagées non commitées"
        echo "→ Commitez ou unstage avant de finaliser"
        exit 2
    fi
}

# === Compléter une phase ===
complete_phase() {
    local session_file="$1"
    local phase="$2"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    local current_phase state
    current_phase=$(jq -r '.currentPhase // 0' "$session_file")
    state=$(jq -r '.state // "unknown"' "$session_file")
    
    if [[ "$state" != "planning" ]]; then
        echo "❌ Impossible de compléter phase en state=$state (requis: planning)"
        exit 2
    fi
    
    if [[ "$phase" != "$current_phase" ]]; then
        echo "❌ Phase $phase n'est pas la phase courante ($current_phase)"
        exit 2
    fi
    
    if [[ "$phase" -gt 1 ]]; then
        local prev_completed
        prev_completed=$(jq --arg p "$((phase - 1))" \
            '.completedPhases | map(select(.phase == ($p | tonumber))) | length' \
            "$session_file")
        
        if [[ "$prev_completed" -eq 0 ]]; then
            echo "❌ Phase $((phase - 1)) n'est pas complétée"
            exit 2
        fi
    fi
    
    local already_completed
    already_completed=$(jq --arg p "$phase" \
        '.completedPhases | map(select(.phase == ($p | tonumber))) | length' \
        "$session_file")
    
    if [[ "$already_completed" -gt 0 ]]; then
        echo "⚠️  Phase $phase déjà complétée"
        exit 0
    fi
    
    if [[ "$phase" -eq 5 ]]; then
        local epics_count
        epics_count=$(jq '.epics | length' "$session_file")
        if [[ "$epics_count" -eq 0 ]]; then
            echo "❌ Phase 5 requiert au moins 1 epic défini"
            exit 2
        fi
    fi
    
    local next_phase=$((phase + 1))
    [[ "$phase" -eq 6 ]] && next_phase=6
    
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg phase "$phase" --arg ts "$timestamp" --arg next "$next_phase" '
        .completedPhases += [{
            "phase": ($phase | tonumber),
            "completedAt": $ts,
            "status": "completed"
        }] |
        .currentPhase = ($next | tonumber) |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "phase_complete",
            "phase": ($phase | tonumber)
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "✓ Phase $phase complétée → Phase $next_phase"
}

# === Aller à une phase spécifique (reset partiel) ===
to_phase() {
    local session_file="$1"
    local target_phase="$2"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    local current_phase state
    current_phase=$(jq -r '.currentPhase // 0' "$session_file")
    state=$(jq -r '.state // "unknown"' "$session_file")
    
    if [[ "$state" != "planning" ]]; then
        echo "❌ --to-phase uniquement en state=planning"
        exit 2
    fi
    
    if [[ "$target_phase" -gt "$current_phase" ]]; then
        echo "❌ Utilisez --complete-phase pour avancer"
        exit 2
    fi
    
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg target "$target_phase" --arg ts "$timestamp" '
        .completedPhases = [.completedPhases[] | select(.phase < ($target | tonumber))] |
        .currentPhase = ($target | tonumber) |
        .validated = false |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "phase_reset",
            "to": ($target | tonumber)
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "✓ Reset à phase $target_phase"
}

# === Reset complet (refus utilisateur) ===
reset_all() {
    local session_file="$1"
    local reason="${2:-user_rejection}"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    local state
    state=$(jq -r '.state // "unknown"' "$session_file")
    
    if [[ "$state" == "applying" ]]; then
        echo "❌ Impossible de reset en state=applying"
        exit 2
    fi
    
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg ts "$timestamp" --arg reason "$reason" '
        .rejectionHistory = ((.rejectionHistory // []) + [{
            "at": $ts,
            "reason": $reason,
            "previousPhases": .completedPhases,
            "previousEpics": (.epics | length)
        }]) |
        .completedPhases = [] |
        .currentPhase = 1 |
        .validated = false |
        .epics = [] |
        .state = "planning" |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "full_reset",
            "reason": $reason
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "✓ Reset complet → Phase 1"
}

# === Transition vers state=planned ===
finalize_planning() {
    local session_file="$1"
    local timestamp
    timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    
    # Vérifications strictes
    local validated completed_phases epics_count tasks_count
    validated=$(jq -r '.validated // false' "$session_file")
    completed_phases=$(jq '.completedPhases | length' "$session_file")
    epics_count=$(jq '.epics | length' "$session_file")
    tasks_count=$(jq '[.epics[].tasks | length] | add // 0' "$session_file")
    
    if [[ "$validated" != "true" ]]; then
        echo "❌ Validation utilisateur requise (validated=false)"
        echo "→ session-validate.sh --approve"
        exit 2
    fi
    
    if [[ "$completed_phases" -lt 5 ]]; then
        echo "❌ Phases 1-5 doivent être complétées ($completed_phases/5)"
        exit 2
    fi
    
    if [[ "$epics_count" -eq 0 ]]; then
        echo "❌ Au moins 1 epic requis"
        exit 2
    fi
    
    if [[ "$tasks_count" -eq 0 ]]; then
        echo "❌ Au moins 1 task requise"
        exit 2
    fi
    
    local tmp_file
    tmp_file=$(mktemp)
    
    jq --arg ts "$timestamp" --argjson epics "$epics_count" --argjson tasks "$tasks_count" '
        .completedPhases += [{
            "phase": 6,
            "completedAt": $ts,
            "status": "completed"
        }] |
        .state = "planned" |
        .currentPhase = 6 |
        .actions = ((.actions // []) + [{
            "at": $ts,
            "type": "planning_finalized",
            "epics": $epics,
            "tasks": $tasks
        }])
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
    
    echo "✓ Planning finalisé"
    echo "  State: planned | Epics: $epics_count | Tasks: $tasks_count"
    echo "  → Prêt pour /apply"
}

# === Main ===
main() {
    if [[ $# -eq 0 ]]; then
        echo "Usage: session-transition.sh [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --complete-phase N    Compléter la phase N"
        echo "  --to-phase N          Reset à la phase N"
        echo "  --reset [reason]      Reset complet phase 1"
        echo "  --finalize            Finaliser → state=planned"
        echo "  --status              Afficher état"
        exit 1
    fi
    
    local session_file
    session_file=$(find_session)
    validate_session "$session_file"
    check_branch_coherence "$session_file"
    
    case "$1" in
        --complete-phase)
            [[ -z "${2:-}" ]] && { echo "❌ Phase requise"; exit 1; }
            complete_phase "$session_file" "$2"
            ;;
        --to-phase)
            [[ -z "${2:-}" ]] && { echo "❌ Phase requise"; exit 1; }
            to_phase "$session_file" "$2"
            ;;
        --reset)
            reset_all "$session_file" "${2:-user_rejection}"
            ;;
        --finalize)
            check_workspace_clean
            finalize_planning "$session_file"
            ;;
        --status)
            echo "Session: $session_file"
            jq '{state, currentPhase, validated, completedPhases: (.completedPhases | length), epics: (.epics | length), actions: (.actions | length)}' "$session_file"
            ;;
        *)
            echo "❌ Option inconnue: $1"
            exit 1
            ;;
    esac
}

main "$@"
