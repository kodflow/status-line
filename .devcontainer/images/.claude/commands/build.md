# Build - Context Generator

$ARGUMENTS

---

## Description

Génère le contexte du projet pour optimiser les interactions avec Claude :
- Fichiers CLAUDE.md hiérarchiques dans chaque dossier
- Versions des dépendances à jour depuis les sources officielles

---

## Arguments

| Pattern | Action |
|---------|--------|
| `--context` | Génère CLAUDE.md + update versions |
| `--help` | Affiche l'aide de la commande |

---

## --help

Quand `--help` est passe, afficher :

```
═══════════════════════════════════════════════
  /build - Context Generator
═══════════════════════════════════════════════

Usage: /build [options]

Options:
  --context       Genere CLAUDE.md + update versions
  --help          Affiche cette aide

Exemples:
  /build --context        Genere le contexte complet
═══════════════════════════════════════════════
```

---

## --context

### Étape 1 : Update des versions

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

Les CLAUDE.md générés ne doivent PAS être commités :

```bash
# Vérifier que CLAUDE.md est dans .gitignore
if ! grep -q "CLAUDE.md" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# Generated context files" >> .gitignore
    echo "CLAUDE.md" >> .gitignore
    echo "!./CLAUDE.md" >> .gitignore  # Garder celui de la racine
fi
```

---

## Ressources distantes

Si des fichiers de règles ne sont pas présents localement :

```
REPO="kodflow/devcontainer-template"
BASE="https://raw.githubusercontent.com/$REPO/main/.devcontainer/features"
```

| Ressource | Local | Distant |
|-----------|-------|---------|
| RULES.md | `languages/<lang>/RULES.md` | `$BASE/languages/<lang>/RULES.md` |

**Priorité** : Local > Distant (fallback automatique)

---

## Output

```
═══════════════════════════════════════════════
  /build --context
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

## Voir aussi

- `/feature <description>` - Développer une nouvelle fonctionnalité
- `/fix <description>` - Corriger un bug
