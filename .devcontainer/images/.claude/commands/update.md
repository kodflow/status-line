# Update - Project & DevContainer Updates

$ARGUMENTS

---

## Description

Commande unifiée pour mettre à jour le projet :
- **--context** : Génère les fichiers CLAUDE.md + vérifie les versions
- **--project** : Met à jour les dépendances (npm, go, cargo, pip...)
- **--devcontainer** : Remplace .devcontainer depuis le template Kodflow
- **(vide)** : Exécute tout (context + project + devcontainer)

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Exécute tout (--context + --project + --devcontainer) |
| `--context` | Génère CLAUDE.md hiérarchiques + vérifie versions |
| `--project` | Met à jour versions et dépendances du projet |
| `--devcontainer` | Remplace .devcontainer depuis le template Kodflow |
| `--dry-run` | Affiche les changements sans les appliquer |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /update - Mise à jour complète du projet
═══════════════════════════════════════════════

Usage: /update [options]

Options:
  (vide)            Exécute tout (context + project + devcontainer)
  --context         Génère CLAUDE.md + vérifie versions langages
  --project         Met à jour dépendances (npm, go, pip...)
  --devcontainer    Remplace .devcontainer depuis template
  --dry-run         Prévisualise sans appliquer
  --help            Affiche cette aide

Exemples:
  /update                   Met à jour tout
  /update --context         Génère le contexte Claude
  /update --project         Met à jour les dépendances
  /update --devcontainer    Met à jour le devcontainer
  /update --dry-run         Prévisualise les changements
═══════════════════════════════════════════════
```

---

## Action: --context

Génère le contexte du projet pour optimiser les interactions avec Claude.

### Étape 1 : Vérification des versions

Récupérer les dernières versions stables depuis les sources officielles.

**RÈGLE ABSOLUE** : Ne JAMAIS downgrader une version existante.

```bash
# Exemple pour Node.js
CURRENT=$(node -v 2>/dev/null | tr -d 'v')
LATEST=$(curl -sL https://nodejs.org/dist/index.json | jq -r '.[0].version' | tr -d 'v')

# Comparer et mettre à jour uniquement si plus récent
if [ "$(printf '%s\n' "$CURRENT" "$LATEST" | sort -V | tail -n1)" = "$LATEST" ]; then
    echo "Update available: $CURRENT → $LATEST"
fi
```

### Étape 2 : Génération CLAUDE.md

Créer des fichiers CLAUDE.md dans chaque dossier selon le principe de l'entonnoir :

| Profondeur | Lignes max | Contenu |
|------------|------------|---------|
| 1 (racine) | ~30 | Vue d'ensemble, structure |
| 2 | ~50 | Détails du module |
| 3+ | ~60 | Spécificités techniques |

**Structure type :**

```markdown
# <Nom du dossier>

## Purpose
<Description en 1-2 phrases>

## Structure
<Arborescence simplifiée>

## Key Files
<Fichiers importants avec description>

## Conventions
<Règles spécifiques au dossier>
```

### Étape 3 : Gitignore

Les CLAUDE.md générés (sauf racine) ne doivent PAS être commités :

```bash
# Vérifier que CLAUDE.md est dans .gitignore
if ! grep -q "CLAUDE.md" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# Generated context files" >> .gitignore
    echo "CLAUDE.md" >> .gitignore
    echo "!./CLAUDE.md" >> .gitignore  # Garder celui de la racine
fi
```

### Ressources distantes

Si des fichiers de règles ne sont pas présents localement :

```
REPO="kodflow/devcontainer-template"
BASE="https://raw.githubusercontent.com/$REPO/main/.devcontainer/features"
```

| Ressource | Local | Distant |
|-----------|-------|---------|
| RULES.md | `languages/<lang>/RULES.md` | `$BASE/languages/<lang>/RULES.md` |

**Priorité** : Local > Distant (fallback automatique)

### Output --context

```
═══════════════════════════════════════════════
  /update --context
═══════════════════════════════════════════════

Checking versions...
  ✓ Node.js: 20.10.0 (latest)
  ✓ Go: 1.21.5 (latest)

Generating CLAUDE.md files...
  ✓ /src/CLAUDE.md
  ✓ /src/api/CLAUDE.md
  ✓ /src/services/CLAUDE.md
  ✓ /tests/CLAUDE.md

✓ Context updated (4 files)
═══════════════════════════════════════════════
```

---

## Action: --project

Met à jour les versions et dépendances du projet.

### Éléments mis à jour

- **GitHub Actions** : uses: avec hash
- **Dockerfile** : ARG versions (KUBECTL, HELM, etc.)
- **package.json** : npm dependencies
- **go.mod** : Go modules
- **Cargo.toml** : Rust crates
- **requirements.txt** : Python packages
- **Gemfile** : Ruby gems

### Output --project

```
═══════════════════════════════════════════════
  /update --project
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

---

## Action: --devcontainer

Remplace .devcontainer depuis le template Kodflow.

### Workflow

1. Identifier les fichiers protégés (gitignored, .env)
2. Sauvegarder les fichiers protégés
3. Télécharger le template depuis GitHub
4. Remplacer .devcontainer/
5. Restaurer les fichiers protégés
6. Valider la configuration

### Fichiers protégés (préservés)

- `.devcontainer/.env`
- `.devcontainer/.env.local`
- `.devcontainer/hooks/shared/.env`
- `.mcp.json` (racine)
- Tous les fichiers gitignored dans .devcontainer/

### Output --devcontainer

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

---

## Action: (vide) - Tout mettre à jour

Quand `/update` est appelé sans argument, exécuter les trois actions dans l'ordre :

1. **--context** : Génère CLAUDE.md + vérifie versions
2. **--project** : Met à jour dépendances
3. **--devcontainer** : Met à jour le devcontainer

### Output (vide)

```
═══════════════════════════════════════════════
  /update - Mise à jour complète
═══════════════════════════════════════════════

[1/3] Context...
  ✓ CLAUDE.md générés (4 fichiers)
  ✓ Versions vérifiées

[2/3] Project...
  ✓ GitHub Actions à jour
  ✓ Dépendances mises à jour

[3/3] DevContainer...
  ✓ Template téléchargé
  ✓ .devcontainer/ remplacé

═══════════════════════════════════════════════
  ✓ Mise à jour complète terminée

  Prochaine étape:
    Ctrl+Shift+P → 'Rebuild Container'
═══════════════════════════════════════════════
```

---

## Voir aussi

- `/feature <description>` - Développer une nouvelle fonctionnalité
- `/fix <description>` - Corriger un bug
- `/review --pr` - Demander une review CodeRabbit
