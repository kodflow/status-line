#!/bin/bash
# task-init.sh - Initialise un projet Taskwarrior avec session v3
# Usage: task-init.sh <type> <description>
#
# Ce script initialise le projet et la session avec le schéma v3.
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
task rc.confirmation=off config confirmation off >/dev/null 2>&1 || true

# Configurer les UDAs pour le système epic/task
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

# Créer le fichier session v3 (schéma complet)
SESSION_FILE="$SESSION_DIR/$PROJECT.json"
TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)

cat > "$SESSION_FILE" << ENDJSON
{
    "schemaVersion": 3,
    "state": "planning",
    "type": "$TYPE",
    "project": "$PROJECT",
    "branch": "$BRANCH",
    "currentPhase": 1,
    "completedPhases": [],
    "validated": false,
    "rejectionHistory": [],
    "currentTask": null,
    "currentEpic": null,
    "lockedPaths": [],
    "epics": [],
    "actions": 0,
    "lastAction": null,
    "createdAt": "$TIMESTAMP"
}
ENDJSON

# Créer symlink pour session active (déterministe)
mkdir -p /workspace/.claude
ln -sf "$SESSION_FILE" /workspace/.claude/state.json
echo "$SESSION_FILE" > /workspace/.claude/active-session

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✓ Projet initialisé: $PROJECT"
echo "═══════════════════════════════════════════════"
echo ""
echo "  State   : planning"
echo "  Type    : $TYPE"
echo "  Branch  : $BRANCH"
echo "  Phase   : 1/6"
echo "  Session : $SESSION_FILE"
echo ""
echo "  Phases OBLIGATOIRES:"
echo "    [→] 1. Analyse de la demande"
echo "    [ ] 2. Recherche documentation"
echo "    [ ] 3. Analyse projet existant"
echo "    [ ] 4. Affûtage"
echo "    [ ] 5. Définition épics/tasks → VALIDATION"
echo "    [ ] 6. Écriture Taskwarrior → state=planned"
echo ""
echo "  INTERDIT en PLAN MODE:"
echo "    ❌ Write/Edit sur fichiers code"
echo "    ❌ Bash avec redirections (>, >>, heredoc)"
echo "    ❌ Sauter des phases"
echo ""
echo "  Workflow: /plan → /apply"
echo "═══════════════════════════════════════════════"
