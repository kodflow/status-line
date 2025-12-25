#!/bin/bash
# pre-validate.sh - Valide les modifications de fichiers protégés
# Usage: pre-validate.sh <file_path>
# Exit 0 = allow, Exit 2 = block

set -euo pipefail

FILE="$1"
if [ -z "$FILE" ]; then
    exit 0
fi

# Chemins des fichiers de configuration
PROTECTED_PATHS_FILE="/workspace/.claude/protected-paths.yml"
PROTECTED_PATHS_DEFAULT="$HOME/.claude/protected-paths.yml"

# Utiliser yq si disponible, sinon fallback sur patterns hardcodés
USE_YQ=false
if command -v yq &>/dev/null; then
    if [[ -f "$PROTECTED_PATHS_FILE" ]]; then
        USE_YQ=true
        CONFIG_FILE="$PROTECTED_PATHS_FILE"
    elif [[ -f "$PROTECTED_PATHS_DEFAULT" ]]; then
        USE_YQ=true
        CONFIG_FILE="$PROTECTED_PATHS_DEFAULT"
    fi
fi

# Patterns protégés par défaut (fallback)
PROTECTED_PATTERNS=(
    "node_modules/"
    ".git/"
    "vendor/"
    "dist/"
    "build/"
    ".env"
    "*.lock"
    "package-lock.json"
    "yarn.lock"
    "pnpm-lock.yaml"
    "Cargo.lock"
    "poetry.lock"
    "go.sum"
    ".claude/scripts/"
    ".claude/commands/"
    ".claude/settings.json"
    ".devcontainer/"
)

# Exceptions (toujours autorisées)
EXCEPTIONS=(
    "*.md"
    "README*"
    "CHANGELOG*"
    ".claude/plans/"
    ".claude/sessions/"
)

# Fonction pour vérifier si le fichier match une exception
is_exception() {
    local file="$1"
    for pattern in "${EXCEPTIONS[@]}"; do
        # Utiliser le pattern matching bash
        if [[ "$file" == *"$pattern"* ]] || [[ "$file" == $pattern ]]; then
            return 0
        fi
    done
    return 1
}

# Vérifier les exceptions d'abord
if is_exception "$FILE"; then
    exit 0
fi

# === Break Glass Check ===
BREAK_GLASS_VAR="${ALLOW_PROTECTED_EDIT:-0}"
if [[ "$BREAK_GLASS_VAR" == "1" ]]; then
    echo "⚠️  BREAK-GLASS activé pour: $FILE"
    echo "   Variable: ALLOW_PROTECTED_EDIT=1"
    # Log l'utilisation du break-glass
    logger -t "claude-protected" "BREAK-GLASS used for: $FILE by $(whoami)" 2>/dev/null || true
    exit 0
fi

# === Vérification avec yq si disponible ===
if [[ "$USE_YQ" == "true" ]]; then
    # Lire les patterns protégés depuis le fichier YAML
    YAML_PATTERNS=$(yq -r '.protected[]' "$CONFIG_FILE" 2>/dev/null || echo "")

    for pattern in $YAML_PATTERNS; do
        [[ -z "$pattern" ]] && continue

        # Vérifier si le fichier match le pattern
        if [[ "$FILE" == *"$pattern"* ]] || [[ "$FILE" == $pattern ]]; then
            echo "═══════════════════════════════════════════════"
            echo "  🚫 FICHIER PROTÉGÉ"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  Fichier: $FILE"
            echo "  Pattern: $pattern"
            echo ""
            echo "  Ce fichier est protégé contre les modifications"
            echo "  accidentelles par .claude/protected-paths.yml"
            echo ""
            echo "  Pour modifier ce fichier:"
            echo "    1. Utilisez une task avec risk=high"
            echo "    2. Activez: export ALLOW_PROTECTED_EDIT=1"
            echo "    3. Fournissez une justification"
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
    done
else
    # Fallback: utiliser les patterns hardcodés
    for pattern in "${PROTECTED_PATTERNS[@]}"; do
        if [[ "$FILE" == *"$pattern"* ]]; then
            echo "🚫 Fichier protégé: $FILE"
            echo "   Pattern: $pattern"
            echo "   Pour forcer: export ALLOW_PROTECTED_EDIT=1"
            exit 2
        fi
    done
fi

# Vérification spéciale pour les commits sur main/master
if [[ "$FILE" == *"git commit"* ]] || [[ "$FILE" == *"git push"* ]]; then
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    if [[ "$BRANCH" == "main" ]] || [[ "$BRANCH" == "master" ]]; then
        echo "⚠️  Attention: opération sur branche $BRANCH"
    fi
fi

exit 0
