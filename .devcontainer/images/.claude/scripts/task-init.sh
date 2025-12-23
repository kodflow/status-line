#!/bin/bash
# task-init.sh - Initialise un projet Taskwarrior (sans phases statiques)
# Usage: task-init.sh <type> <description>
#
# Ce script initialise uniquement le projet et la session.
# Les epics et tasks sont créés dynamiquement pendant le planning.

set -euo pipefail

# Vérifier que Taskwarrior est installé
if ! command -v task &>/dev/null; then
    echo "❌ Taskwarrior non installé !"
    echo ""
    echo "Installation requise pour /feature et /fix :"
    echo ""
    echo "  Ubuntu/Debian : sudo apt-get install taskwarrior"
    echo "  Alpine        : sudo apk add task"
    echo "  macOS         : brew install task"
    echo "  Arch          : sudo pacman -S task"
    echo ""
    echo "Ou exécutez: /update"
    exit 1
fi

TYPE="$1"        # feature ou fix
DESC="$2"        # Description

if [[ -z "$TYPE" || -z "$DESC" ]]; then
    echo "Usage: task-init.sh <type> <description>"
    echo "Exemple: task-init.sh feature \"authentication-system\""
    exit 1
fi

# Normaliser le nom du projet
PROJECT=$(echo "$DESC" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g')
BRANCH="${TYPE}/${PROJECT}"

# Créer le dossier sessions si nécessaire
SESSION_DIR="$HOME/.claude/sessions"
mkdir -p "$SESSION_DIR"

# Vérifier si une session existe déjà pour ce projet
if [[ -f "$SESSION_DIR/$PROJECT.json" ]]; then
    echo "⚠ Session existante trouvée pour: $PROJECT"
    echo "→ Utilisez --continue pour reprendre"
    exit 1
fi

# Configurer Taskwarrior pour usage non-interactif
echo "Configuration de Taskwarrior..."

# Désactiver les confirmations interactives (IMPORTANT pour Claude)
# Utiliser rc.confirmation=off pour éviter les prompts "Are you sure?"
task rc.confirmation=off config confirmation off >/dev/null 2>&1 || true

# Configurer les UDAs pour le système epic/task
# Note: "parent" est un mot réservé, on utilise "epic_uuid" à la place
task rc.confirmation=off config uda.epic.type numeric >/dev/null 2>&1 || true
task rc.confirmation=off config uda.epic.label Epic >/dev/null 2>&1 || true
task rc.confirmation=off config uda.epic_uuid.type string >/dev/null 2>&1 || true
task rc.confirmation=off config uda.epic_uuid.label "Epic UUID" >/dev/null 2>&1 || true

# Parallélisation
task rc.confirmation=off config uda.parallel.type string >/dev/null 2>&1 || true
task rc.confirmation=off config uda.parallel.label Parallel >/dev/null 2>&1 || true
task rc.confirmation=off config uda.parallel.values yes,no >/dev/null 2>&1 || true
task rc.confirmation=off config uda.parallel.default no >/dev/null 2>&1 || true

# Branch et PR
task rc.confirmation=off config uda.branch.type string >/dev/null 2>&1 || true
task rc.confirmation=off config uda.branch.label Branch >/dev/null 2>&1 || true
task rc.confirmation=off config uda.pr_number.type numeric >/dev/null 2>&1 || true
task rc.confirmation=off config uda.pr_number.label PR >/dev/null 2>&1 || true

echo "✓ Taskwarrior configuré"

# Créer le fichier de session (PLAN MODE par défaut)
SESSION_FILE="$SESSION_DIR/$PROJECT.json"
cat > "$SESSION_FILE" << EOF
{
    "project": "$PROJECT",
    "branch": "$BRANCH",
    "type": "$TYPE",
    "mode": "plan",
    "plan_phase": 1,
    "epics": [],
    "current_epic": null,
    "current_task": null,
    "actions": 0,
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "last_action": null
}
EOF

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✓ Projet initialisé: $PROJECT"
echo "═══════════════════════════════════════════════"
echo ""
echo "  Mode: PLAN (analyse et définition des epics/tasks)"
echo ""
echo "  Phases PLAN MODE:"
echo "    1. Analyse de la demande"
echo "    2. Recherche documentation"
echo "    3. Analyse projet existant"
echo "    4. Affûtage (boucle si nécessaire)"
echo "    5. Définition épics/tasks → VALIDATION"
echo "    6. Écriture Taskwarrior"
echo ""
echo "  Après validation → BYPASS MODE (exécution)"
echo ""
echo "  Branch: $BRANCH"
echo "  Session: $SESSION_FILE"
echo "═══════════════════════════════════════════════"
