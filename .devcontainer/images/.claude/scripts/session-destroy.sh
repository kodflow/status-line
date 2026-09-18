#!/bin/bash
# session-destroy.sh - Détruire une session proprement
# Usage: session-destroy.sh [--force]
#
# Nettoie:
# - Session JSON
# - Pointeur active-session
# - Symlink state.json
# - Branche locale (optionnel)
#
# Exit 0 = succès, Exit 1 = erreur, Exit 2 = bloqué

set -euo pipefail

FORCE="${1:-}"

# === Trouver la session active ===
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
    
    echo "$session_file"
}

# === Main ===
main() {
    local session_file
    session_file=$(find_session)
    
    if [[ ! -f "$session_file" ]]; then
        echo "❌ Aucune session active à détruire"
        exit 1
    fi
    
    # Lire infos session
    local state project branch
    state=$(jq -r '.state // "unknown"' "$session_file")
    project=$(jq -r '.project // "unknown"' "$session_file")
    branch=$(jq -r '.branch // ""' "$session_file")
    
    # Bloquer si applying (sauf --force)
    if [[ "$state" == "applying" && "$FORCE" != "--force" ]]; then
        echo "═══════════════════════════════════════════════"
        echo "  ❌ BLOQUÉ: Exécution en cours"
        echo "═══════════════════════════════════════════════"
        echo ""
        echo "  State: applying"
        echo "  Project: $project"
        echo ""
        echo "  Impossible de détruire un plan en cours."
        echo "  Terminez d'abord l'exécution ou utilisez --force."
        echo ""
        echo "═══════════════════════════════════════════════"
        exit 2
    fi
    
    # Vérifier workspace propre (sauf --force)
    if [[ "$FORCE" != "--force" ]]; then
        if ! git diff --quiet 2>/dev/null; then
            echo "⚠️  Workspace Git non propre (modifications non commitées)"
            echo "→ Commitez ou stash vos changements avant destroy"
            exit 2
        fi
    fi
    
    echo "═══════════════════════════════════════════════"
    echo "  Destruction session: $project"
    echo "═══════════════════════════════════════════════"
    echo ""
    
    # 1. Supprimer pointeur active-session
    if [[ -f "/workspace/.claude/active-session" ]]; then
        rm -f /workspace/.claude/active-session
        echo "  ✓ Pointeur active-session supprimé"
    fi
    
    # 2. Supprimer symlink state.json
    if [[ -L "/workspace/.claude/state.json" ]]; then
        rm -f /workspace/.claude/state.json
        echo "  ✓ Symlink state.json supprimé"
    fi
    
    # 3. Supprimer fichier session
    rm -f "$session_file"
    echo "  ✓ Session supprimée: $session_file"
    
    # 4. Optionnel: supprimer branche locale
    local current_branch main_branch
    current_branch=$(git branch --show-current 2>/dev/null || echo "")
    main_branch=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")
    
    if [[ -n "$branch" && "$current_branch" == "$branch" ]]; then
        git checkout "$main_branch" 2>/dev/null || true
        git branch -D "$branch" 2>/dev/null || true
        echo "  ✓ Branche locale supprimée: $branch"
    fi
    
    echo ""
    echo "  Session détruite avec succès."
    echo ""
    echo "  Note: La branche remote n'est pas supprimée."
    echo "  Pour la supprimer: git push origin --delete $branch"
    echo ""
    echo "═══════════════════════════════════════════════"
}

main "$@"
