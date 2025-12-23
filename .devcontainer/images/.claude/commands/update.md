# Update - Project & DevContainer Updates

$ARGUMENTS

---

## Description

Met à jour les versions et dépendances du projet, ou remplace le devcontainer depuis le template.

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Analyse et met à jour toutes les versions/dépendances |
| `--devcontainer` | Remplace .devcontainer depuis le template Kodflow |
| `--dry-run` | Affiche les changements sans les appliquer |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /update - Mise à jour du projet
═══════════════════════════════════════════════

Usage: /update [options]

Options:
  (vide)            Met à jour versions et dépendances
  --devcontainer    Remplace .devcontainer depuis template
  --dry-run         Prévisualise sans appliquer
  --help            Affiche cette aide

Éléments mis à jour (mode par défaut):
  - GitHub Actions (uses: avec hash)
  - Dockerfile (ARG versions)
  - package.json, go.mod, Cargo.toml, etc.

Mode --devcontainer:
  - Télécharge kodflow/devcontainer-template
  - Remplace tous les fichiers .devcontainer
  - Préserve les fichiers .gitignore (.env, etc.)

Exemples:
  /update                   Met à jour le projet
  /update --dry-run         Prévisualise les mises à jour
  /update --devcontainer    Met à jour le devcontainer
═══════════════════════════════════════════════
```

---

## Action: Mode par défaut (versions/dépendances)

Exécuter ce script bash pour mettre à jour les versions :

```bash
#!/bin/bash
set -e

DRY_RUN="${DRY_RUN:-false}"
UPDATES_COUNT=0

echo "═══════════════════════════════════════════════"
echo "  /update - Mise à jour des versions"
echo "═══════════════════════════════════════════════"
echo ""

# =============================================================================
# Helper: Get latest GitHub Action version with hash
# =============================================================================
get_action_latest() {
    local OWNER="$1"
    local REPO="$2"
    local CURRENT_HASH="$3"
    
    # Get latest release tag
    local TAG=$(curl -sL "https://api.github.com/repos/$OWNER/$REPO/releases/latest" 2>/dev/null | jq -r '.tag_name // empty')
    
    if [ -z "$TAG" ]; then
        # Fallback: get latest tag
        TAG=$(curl -sL "https://api.github.com/repos/$OWNER/$REPO/tags" 2>/dev/null | jq -r '.[0].name // empty')
    fi
    
    if [ -z "$TAG" ]; then
        return 1
    fi
    
    # Get commit SHA for the tag (handle annotated vs lightweight)
    local REF_DATA=$(curl -sL "https://api.github.com/repos/$OWNER/$REPO/git/ref/tags/$TAG" 2>/dev/null)
    local OBJ_TYPE=$(echo "$REF_DATA" | jq -r '.object.type // empty')
    local SHA=""
    
    if [ "$OBJ_TYPE" = "tag" ]; then
        # Annotated tag - need to dereference
        local TAG_URL=$(echo "$REF_DATA" | jq -r '.object.url')
        SHA=$(curl -sL "$TAG_URL" 2>/dev/null | jq -r '.object.sha // empty')
    else
        # Lightweight tag
        SHA=$(echo "$REF_DATA" | jq -r '.object.sha // empty')
    fi
    
    if [ -z "$SHA" ]; then
        return 1
    fi
    
    # Check if update needed
    if [ "${SHA:0:40}" != "${CURRENT_HASH:0:40}" ]; then
        echo "$SHA $TAG"
        return 0
    fi
    
    return 1
}

# =============================================================================
# Update GitHub Actions
# =============================================================================
update_github_actions() {
    echo "GitHub Actions:"
    
    local WORKFLOW_FILES=$(find .github/workflows -name "*.yml" -o -name "*.yaml" 2>/dev/null)
    
    if [ -z "$WORKFLOW_FILES" ]; then
        echo "  (aucun workflow trouvé)"
        echo ""
        return
    fi
    
    local FOUND_UPDATES=false
    
    for WF in $WORKFLOW_FILES; do
        # Extract all uses: lines with hash format
        while IFS= read -r LINE; do
            # Parse: uses: owner/repo@hash # vX
            if [[ "$LINE" =~ uses:[[:space:]]*([^/]+)/([^@]+)@([a-f0-9]+) ]]; then
                local OWNER="${BASH_REMATCH[1]}"
                local REPO="${BASH_REMATCH[2]}"
                local CURRENT_HASH="${BASH_REMATCH[3]}"
                
                # Get current version comment
                local CURRENT_VER=""
                if [[ "$LINE" =~ \#[[:space:]]*(v[0-9.]+) ]]; then
                    CURRENT_VER="${BASH_REMATCH[1]}"
                fi
                
                # Get latest
                local LATEST=$(get_action_latest "$OWNER" "$REPO" "$CURRENT_HASH")
                
                if [ -n "$LATEST" ]; then
                    local NEW_HASH=$(echo "$LATEST" | cut -d' ' -f1)
                    local NEW_TAG=$(echo "$LATEST" | cut -d' ' -f2)
                    
                    FOUND_UPDATES=true
                    echo "  $WF:"
                    echo "    $OWNER/$REPO: $CURRENT_VER → $NEW_TAG"
                    
                    if [ "$DRY_RUN" != "true" ]; then
                        # Replace in file
                        sed -i "s|$OWNER/$REPO@$CURRENT_HASH|$OWNER/$REPO@$NEW_HASH # $NEW_TAG|g" "$WF"
                        ((UPDATES_COUNT++))
                    fi
                fi
            fi
        done < <(grep -E "uses:[[:space:]]*[^/]+/[^@]+@[a-f0-9]+" "$WF" 2>/dev/null || true)
    done
    
    if [ "$FOUND_UPDATES" = false ]; then
        echo "  (toutes les actions sont à jour)"
    fi
    echo ""
}

# =============================================================================
# Update Dockerfile ARG versions
# =============================================================================
update_dockerfile_versions() {
    echo "Dockerfile ARG versions:"
    
    local DOCKERFILES=$(find . -name "Dockerfile" -not -path "./.git/*" 2>/dev/null)
    
    if [ -z "$DOCKERFILES" ]; then
        echo "  (aucun Dockerfile trouvé)"
        echo ""
        return
    fi
    
    local FOUND_UPDATES=false
    
    for DF in $DOCKERFILES; do
        # Check for KUBECTL_VERSION
        if grep -q "ARG KUBECTL_VERSION=" "$DF" 2>/dev/null; then
            local CURRENT=$(grep "ARG KUBECTL_VERSION=" "$DF" | sed 's/.*=//')
            local LATEST=$(curl -sL "https://dl.k8s.io/release/stable.txt" 2>/dev/null | tr -d 'v')
            
            if [ -n "$LATEST" ] && [ "$CURRENT" != "$LATEST" ]; then
                FOUND_UPDATES=true
                echo "  $DF:"
                echo "    KUBECTL_VERSION: $CURRENT → $LATEST"
                
                if [ "$DRY_RUN" != "true" ]; then
                    sed -i "s/ARG KUBECTL_VERSION=.*/ARG KUBECTL_VERSION=$LATEST/" "$DF"
                    ((UPDATES_COUNT++))
                fi
            fi
        fi
        
        # Check for HELM_VERSION
        if grep -q "ARG HELM_VERSION=" "$DF" 2>/dev/null; then
            local CURRENT=$(grep "ARG HELM_VERSION=" "$DF" | sed 's/.*=//')
            local LATEST=$(curl -sL "https://api.github.com/repos/helm/helm/releases/latest" 2>/dev/null | jq -r '.tag_name' | tr -d 'v')
            
            if [ -n "$LATEST" ] && [ "$CURRENT" != "$LATEST" ]; then
                FOUND_UPDATES=true
                echo "  $DF:"
                echo "    HELM_VERSION: $CURRENT → $LATEST"
                
                if [ "$DRY_RUN" != "true" ]; then
                    sed -i "s/ARG HELM_VERSION=.*/ARG HELM_VERSION=$LATEST/" "$DF"
                    ((UPDATES_COUNT++))
                fi
            fi
        fi
    done
    
    if [ "$FOUND_UPDATES" = false ]; then
        echo "  (toutes les versions sont à jour)"
    fi
    echo ""
}

# =============================================================================
# Update package.json dependencies
# =============================================================================
update_package_json() {
    echo "Node.js (package.json):"
    
    if [ ! -f "package.json" ]; then
        echo "  (aucun package.json trouvé)"
        echo ""
        return
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        local OUTDATED=$(npm outdated --json 2>/dev/null || echo "{}")
        if [ "$OUTDATED" != "{}" ] && [ -n "$OUTDATED" ]; then
            echo "$OUTDATED" | jq -r 'to_entries[] | "    \(.key): \(.value.current) → \(.value.latest)"' 2>/dev/null || echo "  (à jour)"
        else
            echo "  (toutes les dépendances sont à jour)"
        fi
    else
        echo "  Mise à jour des dépendances..."
        npm update --save 2>/dev/null && echo "  ✓ Dépendances mises à jour" || echo "  ⚠ Erreur npm update"
        ((UPDATES_COUNT++))
    fi
    echo ""
}

# =============================================================================
# Update go.mod dependencies
# =============================================================================
update_go_mod() {
    echo "Go (go.mod):"
    
    if [ ! -f "go.mod" ]; then
        echo "  (aucun go.mod trouvé)"
        echo ""
        return
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        go list -m -u all 2>/dev/null | grep '\[' | head -10 | while read -r LINE; do
            echo "    $LINE"
        done || echo "  (à jour)"
    else
        echo "  Mise à jour des dépendances..."
        go get -u ./... 2>/dev/null && go mod tidy 2>/dev/null && echo "  ✓ Dépendances mises à jour" || echo "  ⚠ Erreur go get"
        ((UPDATES_COUNT++))
    fi
    echo ""
}

# =============================================================================
# Update Cargo.toml dependencies
# =============================================================================
update_cargo_toml() {
    echo "Rust (Cargo.toml):"
    
    if [ ! -f "Cargo.toml" ]; then
        echo "  (aucun Cargo.toml trouvé)"
        echo ""
        return
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        cargo outdated 2>/dev/null | head -15 || echo "  (à jour ou cargo-outdated non installé)"
    else
        echo "  Mise à jour des dépendances..."
        cargo update 2>/dev/null && echo "  ✓ Dépendances mises à jour" || echo "  ⚠ Erreur cargo update"
        ((UPDATES_COUNT++))
    fi
    echo ""
}

# =============================================================================
# Update requirements.txt
# =============================================================================
update_requirements_txt() {
    echo "Python (requirements.txt):"
    
    if [ ! -f "requirements.txt" ]; then
        echo "  (aucun requirements.txt trouvé)"
        echo ""
        return
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        pip list --outdated 2>/dev/null | head -10 || echo "  (à jour)"
    else
        echo "  Mise à jour des dépendances..."
        pip install --upgrade -r requirements.txt 2>/dev/null && echo "  ✓ Dépendances mises à jour" || echo "  ⚠ Erreur pip"
        ((UPDATES_COUNT++))
    fi
    echo ""
}

# =============================================================================
# Update Gemfile
# =============================================================================
update_gemfile() {
    echo "Ruby (Gemfile):"
    
    if [ ! -f "Gemfile" ]; then
        echo "  (aucun Gemfile trouvé)"
        echo ""
        return
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        bundle outdated 2>/dev/null | head -10 || echo "  (à jour)"
    else
        echo "  Mise à jour des dépendances..."
        bundle update 2>/dev/null && echo "  ✓ Dépendances mises à jour" || echo "  ⚠ Erreur bundle"
        ((UPDATES_COUNT++))
    fi
    echo ""
}

# =============================================================================
# Main
# =============================================================================
update_github_actions
update_dockerfile_versions
update_package_json
update_go_mod
update_cargo_toml
update_requirements_txt
update_gemfile

echo "═══════════════════════════════════════════════"
if [ "$DRY_RUN" = "true" ]; then
    echo "  Mode dry-run: aucun changement appliqué"
else
    echo "  ✓ $UPDATES_COUNT mises à jour appliquées"
fi
echo "═══════════════════════════════════════════════"
```

---

## Action: --devcontainer

Exécuter ce script bash pour mettre à jour le devcontainer :

```bash
#!/bin/bash
set -e

DRY_RUN="${DRY_RUN:-false}"
REPO="kodflow/devcontainer-template"
BRANCH="main"
BACKUP_DIR="/tmp/devcontainer-backup-$$"
TEMP_DIR="/tmp/devcontainer-download-$$"

echo "═══════════════════════════════════════════════"
echo "  /update --devcontainer"
echo "═══════════════════════════════════════════════"
echo ""

# =============================================================================
# Check if .devcontainer exists
# =============================================================================
if [ ! -d ".devcontainer" ]; then
    echo "Erreur: .devcontainer/ n'existe pas"
    echo "Utilisez /install pour créer un nouveau devcontainer"
    exit 1
fi

# =============================================================================
# Identify protected files (gitignored + hardcoded)
# =============================================================================
echo "Identification des fichiers protégés..."

PROTECTED_FILES=""

# Git-ignored files in .devcontainer
if command -v git &>/dev/null && [ -d ".git" ]; then
    PROTECTED_FILES=$(git ls-files --ignored --exclude-standard .devcontainer/ 2>/dev/null || true)
fi

# Add hardcoded protected patterns
HARDCODED_PROTECTED=".devcontainer/.env .devcontainer/.env.local .devcontainer/hooks/shared/.env"
for F in $HARDCODED_PROTECTED; do
    if [ -f "$F" ]; then
        PROTECTED_FILES="$PROTECTED_FILES $F"
    fi
done

# Also protect .mcp.json at root
if [ -f ".mcp.json" ]; then
    PROTECTED_FILES="$PROTECTED_FILES .mcp.json"
fi

echo ""
echo "Fichiers protégés (préservés):"
if [ -n "$PROTECTED_FILES" ]; then
    for F in $PROTECTED_FILES; do
        echo "  ✓ $F"
    done
else
    echo "  (aucun)"
fi
echo ""

# =============================================================================
# Dry run: show what would happen
# =============================================================================
if [ "$DRY_RUN" = "true" ]; then
    echo "Mode dry-run: prévisualisation"
    echo ""
    echo "Actions prévues:"
    echo "  1. Télécharger $REPO (branche $BRANCH)"
    echo "  2. Sauvegarder les fichiers protégés"
    echo "  3. Remplacer .devcontainer/"
    echo "  4. Restaurer les fichiers protégés"
    echo "  5. Valider la configuration"
    echo ""
    echo "═══════════════════════════════════════════════"
    echo "  Dry-run terminé. Aucun changement."
    echo "═══════════════════════════════════════════════"
    exit 0
fi

# =============================================================================
# Backup protected files
# =============================================================================
echo "Sauvegarde des fichiers protégés..."
mkdir -p "$BACKUP_DIR"

for F in $PROTECTED_FILES; do
    if [ -f "$F" ]; then
        mkdir -p "$BACKUP_DIR/$(dirname "$F")"
        cp "$F" "$BACKUP_DIR/$F"
        echo "  ✓ $F sauvegardé"
    fi
done
echo ""

# =============================================================================
# Download template
# =============================================================================
echo "Téléchargement de $REPO..."
mkdir -p "$TEMP_DIR"

curl -sL "https://github.com/$REPO/archive/refs/heads/$BRANCH.tar.gz" | tar xz -C "$TEMP_DIR"

if [ ! -d "$TEMP_DIR/devcontainer-template-$BRANCH/.devcontainer" ]; then
    echo "Erreur: Template non trouvé dans l'archive"
    rm -rf "$TEMP_DIR" "$BACKUP_DIR"
    exit 1
fi

echo "  ✓ Template téléchargé"
echo ""

# =============================================================================
# Replace .devcontainer
# =============================================================================
echo "Remplacement de .devcontainer/..."

# Remove old .devcontainer
rm -rf .devcontainer

# Move new .devcontainer
mv "$TEMP_DIR/devcontainer-template-$BRANCH/.devcontainer" .devcontainer

echo "  ✓ .devcontainer/ remplacé"
echo ""

# =============================================================================
# Restore protected files
# =============================================================================
echo "Restauration des fichiers protégés..."

for F in $PROTECTED_FILES; do
    if [ -f "$BACKUP_DIR/$F" ]; then
        mkdir -p "$(dirname "$F")"
        cp "$BACKUP_DIR/$F" "$F"
        echo "  ✓ $F restauré"
    fi
done
echo ""

# =============================================================================
# Cleanup
# =============================================================================
rm -rf "$TEMP_DIR" "$BACKUP_DIR"

# =============================================================================
# Validate
# =============================================================================
echo "Validation de la configuration..."

if command -v docker &>/dev/null; then
    if docker compose -f .devcontainer/docker-compose.yml config --quiet 2>/dev/null; then
        echo "  ✓ docker-compose.yml valide"
    else
        echo "  ⚠ Attention: docker-compose.yml peut nécessiter des ajustements"
    fi
else
    echo "  (docker non disponible, validation ignorée)"
fi
echo ""

# =============================================================================
# Summary
# =============================================================================
echo "═══════════════════════════════════════════════"
echo "  ✓ DevContainer mis à jour"
echo ""
echo "  Prochaine étape:"
echo "    Ctrl+Shift+P → 'Rebuild Container'"
echo "═══════════════════════════════════════════════"
```

---

## Output

### Mode par défaut
```
═══════════════════════════════════════════════
  /update - Mise à jour des versions
═══════════════════════════════════════════════

GitHub Actions:
  .github/workflows/docker-images.yml:
    actions/checkout: v4 → v4.2.2
    docker/build-push-action: v5 → v5.2.0

Dockerfile ARG versions:
  .devcontainer/images/Dockerfile:
    KUBECTL_VERSION: 1.32.0 → 1.33.0
    HELM_VERSION: 3.16.3 → 3.17.0

Node.js (package.json):
  (aucun package.json trouvé)

Go (go.mod):
  (aucun go.mod trouvé)

═══════════════════════════════════════════════
  ✓ 4 mises à jour appliquées
═══════════════════════════════════════════════
```

### Mode --devcontainer
```
═══════════════════════════════════════════════
  /update --devcontainer
═══════════════════════════════════════════════

Identification des fichiers protégés...

Fichiers protégés (préservés):
  ✓ .devcontainer/.env
  ✓ .devcontainer/hooks/shared/.env

Sauvegarde des fichiers protégés...
  ✓ .devcontainer/.env sauvegardé

Téléchargement de kodflow/devcontainer-template...
  ✓ Template téléchargé

Remplacement de .devcontainer/...
  ✓ .devcontainer/ remplacé

Restauration des fichiers protégés...
  ✓ .devcontainer/.env restauré

Validation de la configuration...
  ✓ docker-compose.yml valide

═══════════════════════════════════════════════
  ✓ DevContainer mis à jour

  Prochaine étape:
    Ctrl+Shift+P → 'Rebuild Container'
═══════════════════════════════════════════════
```
