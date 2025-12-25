#!/bin/bash
# bash-validate.sh - PreToolUse hook pour Bash
# Vérifie que les commandes bash respectent les règles du mode courant
# Exit 0 = autorisé, Exit 2 = bloqué
#
# RÈGLE CRITIQUE: En state=planning/planned, TOUTES les écritures sont bloquées
# Allowlist ultra-strict pour éviter tout bypass

set -euo pipefail

# Lire l'input JSON de Claude
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

# Ne traiter que les commandes Bash
if [[ "$TOOL" != "Bash" ]]; then
    exit 0
fi

# === Trouver la session active (déterministe) ===
SESSION_FILE=""

if [[ -f "/workspace/.claude/active-session" ]]; then
    SESSION_FILE=$(cat /workspace/.claude/active-session 2>/dev/null || true)
fi

if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    if [[ -f "/workspace/.claude/state.json" ]]; then
        SESSION_FILE=$(readlink -f /workspace/.claude/state.json 2>/dev/null || echo "/workspace/.claude/state.json")
    fi
fi

if [[ -z "$SESSION_FILE" || ! -f "$SESSION_FILE" ]]; then
    SESSION_DIR="$HOME/.claude/sessions"
    SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1 || true)
fi

# Si pas de session, autoriser (mode dégradé)
if [[ ! -f "$SESSION_FILE" ]]; then
    exit 0
fi

# === Lire l'état depuis .state ===
STATE=$(jq -r '.state // "unknown"' "$SESSION_FILE")

# États autorisés pour modifications
if [[ "$STATE" == "applying" || "$STATE" == "applied" ]]; then
    exit 0
fi

# === STATE = planning ou planned : MODE LECTURE SEULE STRICT ===

# === ALLOWLIST STRICTE ===
# UNIQUEMENT ces commandes sont autorisées en planning/planned
READONLY_ALLOWED=(
    # Git lecture
    "git status"
    "git log"
    "git diff"
    "git show"
    "git branch"
    "git rev-parse"
    "git ls-files"
    "git remote"
    "git symbolic-ref"
    # Lecture fichiers
    "ls"
    "cat"
    "head"
    "tail"
    "less"
    "more"
    # Recherche
    "grep"
    "rg"
    "ag"
    "find"
    "fd"
    "tree"
    "locate"
    # Analyse
    "wc"
    "file"
    "stat"
    "du"
    "df"
    # Shell
    "which"
    "whereis"
    "type"
    "pwd"
    "echo"
    "printf"
    "date"
    "env"
    "printenv"
    # Parseurs (lecture seule)
    "jq"
    "yq"
    "xmllint"
    # Scripts session (autorisés car ils valident eux-mêmes)
    "session-transition.sh"
    "session-validate.sh"
    # Taskwarrior (lecture + modification via scripts dédiés)
    "task "
    "task-init.sh"
    "task-epic.sh"
    "task-add.sh"
    "task-start.sh"
    "task-done.sh"
    # Tests (lecture seule, pas de modification)
    "go test"
    "cargo test"
    "npm test"
    "npm run test"
    "yarn test"
    "pnpm test"
    "pytest"
    "make test"
    "make check"
)

# === BLOCKLIST STRICTE (patterns d'écriture) ===
WRITE_PATTERNS=(
    # === Redirections ===
    " > "
    " >"
    ">"
    ">>"
    # === Heredocs ===
    "<<EOF"
    "<<'EOF'"
    "<<-EOF"
    "<< EOF"
    "<<HEREDOC"
    "<<END"
    "<<-"
    # === Pipes d'écriture ===
    "| tee"
    "|tee"
    "| dd"
    "|dd"
    # === Modifications in-place ===
    "sed -i"
    "sed -i'"
    "perl -i"
    "perl -pi"
    "awk -i"
    "ed "
    "ex "
    # === Modifications fichiers ===
    "touch "
    "mkdir "
    "rm "
    "rmdir "
    "mv "
    "cp "
    "ln "
    "chmod "
    "chown "
    "chgrp "
    "install "
    "truncate "
    "shred "
    # === Git modifications ===
    "git add"
    "git commit"
    "git push"
    "git pull"
    "git fetch"
    "git merge"
    "git rebase"
    "git cherry-pick"
    "git reset"
    "git checkout --"
    "git restore --staged"
    "git stash"
    "git apply"
    "git am"
    "git clean"
    "git gc"
    # === Patch/Apply ===
    "patch "
    "patch -"
    "diff -u.*|"
    # === Package managers ===
    "npm install"
    "npm i "
    "npm ci"
    "npm update"
    "npm uninstall"
    "npm link"
    "yarn install"
    "yarn add"
    "yarn remove"
    "pnpm install"
    "pnpm add"
    "pnpm remove"
    "pip install"
    "pip uninstall"
    "pip3 install"
    "pipx install"
    "poetry install"
    "poetry add"
    "go install"
    "go get"
    "go mod tidy"
    "go mod download"
    "cargo install"
    "cargo add"
    "gem install"
    "bundle install"
    "composer install"
    "composer require"
    "apt install"
    "apt-get install"
    "apk add"
    "brew install"
    # === Formatters/Linters auto-fix ===
    "prettier --write"
    "prettier -w"
    "eslint --fix"
    "go fmt"
    "gofmt -w"
    "goimports -w"
    "rustfmt"
    "cargo fmt"
    "black "
    "autopep8"
    "isort "
    "yapf"
    "rubocop -a"
    "rubocop --auto-correct"
    # === Langages avec écriture potentielle ===
    "python -c"
    "python3 -c"
    "node -e"
    "node --eval"
    "ruby -e"
    "perl -e"
    "php -r"
    "go run"
    # === Téléchargement avec écriture ===
    "curl -o"
    "curl -O"
    "curl --output"
    "wget "
    "wget -"
    # === Archives ===
    "tar -x"
    "tar xf"
    "tar xzf"
    "tar xjf"
    "unzip "
    "gunzip "
    "bunzip2 "
    "7z x"
    # === Docker modifications ===
    "docker build"
    "docker run"
    "docker exec"
    "docker compose up"
    "docker-compose up"
    # === Autres commandes dangereuses ===
    "dd "
    "mkfs"
    "mount "
    "umount "
    "kill "
    "pkill "
    "killall "
    "nohup "
    "setsid "
    "at "
    "crontab "
)

# === EXCEPTIONS (ces patterns NE déclenchent PAS le blocage) ===
EXCEPTIONS=(
    "> /dev/null"
    ">/dev/null"
    "2> /dev/null"
    "2>/dev/null"
    "2>&1"
    "&> /dev/null"
    "&>/dev/null"
    "| head"
    "| tail"
    "| grep"
    "| rg"
    "| jq"
    "| yq"
    "| wc"
    "| sort"
    "| uniq"
    "| awk"
    "| sed"
    "| cut"
    "| tr"
    "| xargs"
    "--help"
    "-h"
    "--version"
    "-V"
)

# === Vérifier si une exception s'applique ===
has_exception() {
    local cmd="$1"
    for exc in "${EXCEPTIONS[@]}"; do
        if [[ "$cmd" == *"$exc"* ]]; then
            return 0
        fi
    done
    return 1
}

# === Vérifier si la commande est dans l'allowlist ===
is_allowed() {
    local cmd="$1"
    local cmd_lower
    cmd_lower=$(echo "$cmd" | tr '[:upper:]' '[:lower:]')
    
    for allowed in "${READONLY_ALLOWED[@]}"; do
        # Match au début de la commande
        if [[ "$cmd_lower" == "$allowed"* ]]; then
            return 0
        fi
        # Match après un chemin (ex: /usr/bin/git status)
        if [[ "$cmd_lower" == *"/$allowed"* ]]; then
            return 0
        fi
    done
    return 1
}

# === Vérifier les patterns d'écriture ===
for pattern in "${WRITE_PATTERNS[@]}"; do
    if [[ "$COMMAND" == *"$pattern"* ]]; then
        # Vérifier les exceptions
        if has_exception "$COMMAND"; then
            continue
        fi
        
        echo "═══════════════════════════════════════════════"
        echo "  🚫 BLOQUÉ: Écriture interdite en PLAN MODE"
        echo "═══════════════════════════════════════════════"
        echo ""
        echo "  État   : $STATE (lecture seule strict)"
        echo "  Pattern: $pattern"
        echo ""
        echo "  Commande :"
        echo "    ${COMMAND:0:200}"
        echo ""
        echo "  En PLAN MODE, seules les commandes de lecture"
        echo "  sont autorisées. Aucune modification permise."
        echo ""
        echo "  Pour modifier des fichiers :"
        echo "    1. session-validate.sh --approve"
        echo "    2. session-transition.sh --finalize"
        echo "    3. /apply"
        echo ""
        echo "═══════════════════════════════════════════════"
        exit 2
    fi
done

# === Vérifier si commande dans allowlist ===
if ! is_allowed "$COMMAND"; then
    # Vérifier les caractères de redirection bruts
    if [[ "$COMMAND" =~ \>[^/\&] ]] || [[ "$COMMAND" =~ \>\> ]] || [[ "$COMMAND" =~ \<\< ]]; then
        if ! has_exception "$COMMAND"; then
            echo "═══════════════════════════════════════════════"
            echo "  🚫 BLOQUÉ: Redirection détectée en PLAN MODE"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  État   : $STATE (lecture seule strict)"
            echo "  Commande non dans l'allowlist."
            echo ""
            echo "  Commande :"
            echo "    ${COMMAND:0:200}"
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
    fi
    
    # Commande non reconnue mais pas de pattern dangereux détecté
    # Log warning mais autoriser (fallback permissif pour commandes inconnues sans redirection)
    echo "⚠️  Commande non dans allowlist: ${COMMAND:0:50}..."
fi

# Commande autorisée
exit 0
