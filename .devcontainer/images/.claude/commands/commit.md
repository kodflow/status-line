# Commit - Git Workflow Automation

$ARGUMENTS

---

## Priorite des outils

**IMPORTANT** : Toujours privilegier les outils MCP GitHub quand disponibles.

| Action        | Priorite 1 (MCP)                   | Fallback (CLI)    |
| ------------- | ---------------------------------- | ----------------- |
| Creer branche | `mcp__github__create_branch`       | `git checkout -b` |
| Creer PR      | `mcp__github__create_pull_request` | `gh pr create`    |
| Lister PRs    | `mcp__github__list_pull_requests`  | `gh pr list`      |
| Voir PR       | `mcp__github__get_pull_request`    | `gh pr view`      |
| Merger PR     | `mcp__github__merge_pull_request`  | `gh pr merge`     |
| Creer issue   | `mcp__github__create_issue`        | `gh issue create` |

**Git local** (toujours CLI - pas de MCP) :

- `git status`, `git add`, `git commit`, `git push`
- `git branch`, `git checkout`, `git diff`

**Extraction owner/repo pour MCP** :

```bash
git remote get-url origin | sed -E 's#.*[:/]([^/]+)/([^/.]+)(\.git)?$#\1 \2#'
```

Retourne : `owner repo`

---

## Detection du contexte

1. **Verifier les changements** :

   ```bash
   git status --porcelain
   git diff --stat
   git diff --cached --stat
   ```

2. **Verifier la branche courante** :

   ```bash
   git branch --show-current
   ```

---

## Arguments

| Pattern           | Action                                         |
| ----------------- | ---------------------------------------------- |
| (vide)            | Workflow complet : branch, commit, push, PR    |
| `--branch <nom>`  | Force le nom de branche                        |
| `--no-pr`         | Skip la creation de PR                         |
| `--amend`         | Amend le dernier commit (meme branche)         |
| `--rename <nom>`  | Renomme la branche avant push/PR               |
| `--help`          | Affiche l'aide de la commande                  |

---

## --help

Quand `--help` est passe, afficher :

```
═══════════════════════════════════════════════
  /commit - Git Workflow Automation
═══════════════════════════════════════════════

Usage: /commit [options]

Options:
  (vide)            Workflow complet : branch, commit, push, PR
  --branch <nom>    Force le nom de branche
  --no-pr           Skip la creation de PR
  --amend           Amend le dernier commit
  --rename <nom>    Renomme la branche avant push/PR
  --help            Affiche cette aide

Exemples:
  /commit                   Commit + PR automatique
  /commit --no-pr           Commit sans creer de PR
  /commit --amend           Amend le dernier commit
  /commit --branch feat/x   Force le nom de branche
═══════════════════════════════════════════════
```

---

## Workflow principal

### 1. Verifier les changements

```bash
git status --porcelain
```

- **Si aucun changement** : Erreur "Aucun changement a commiter"
- **Si changements** : Continuer

### 2. Gestion de la branche (AUTONOME - JAMAIS DEMANDER)

**Verifier la branche courante** :

```bash
git branch --show-current
```

**Regles de decision AUTOMATIQUE** :

1. **Si `main` ou `master`** : CREER nouvelle branche automatiquement

2. **Si autre branche** : Analyser la coherence :
   - Extraire le type et scope du nom de branche
   - Analyser les fichiers modifies
   - **Si coherent** : Utiliser la branche existante
   - **Si NON coherent** : CREER nouvelle branche depuis main

**Detection de coherence** :

| Branche         | Fichiers modifies    | Decision          |
| --------------- | -------------------- | ----------------- |
| `feat/auth`     | `src/auth/*`         | Coherent          |
| `feat/auth`     | `docs/readme.md`     | Nouvelle branche  |
| `fix/api-error` | `src/api/*`          | Coherent          |
| `fix/api-error` | `src/ui/button.ts`   | Nouvelle branche  |
| `chore/deps`    | `package.json`       | Coherent          |
| `docs/readme`   | `*.md`               | Coherent          |

**IMPORTANT** : Ne JAMAIS demander a l'utilisateur quelle branche utiliser.

**Creation de branche automatique** :

```bash
git checkout main && git pull origin main
git checkout -b <type>/<description>
```

### 3. Evolution du contexte de branche

**Si la branche n'a PAS encore de PR** et que le contexte a evolue :

- On peut renommer la branche pour refleter les nouveaux changements
- Utiliser `--rename` ou detecter automatiquement si le nom ne correspond plus

**Renommage automatique** (avant premier push seulement) :

```bash
git branch -m <ancien-nom> <nouveau-nom>
```

**Si deja pushe mais pas de PR** :

```bash
git branch -m <ancien-nom> <nouveau-nom>
git push origin --delete <ancien-nom>
git push -u origin <nouveau-nom>
```

### 4. Stage et Commit

**Stage tous les fichiers pertinents** :

```bash
git add -A
```

**Generer le message de commit** (Conventional Commits) :

Format :

```text
<type>(<scope>): <description>

[body optionnel - details des changements]
```

| Type       | Usage                                   |
| ---------- | --------------------------------------- |
| `feat`     | Nouvelle fonctionnalite                 |
| `fix`      | Correction de bug                       |
| `refactor` | Refactoring sans changement fonctionnel |
| `docs`     | Documentation uniquement                |
| `test`     | Ajout/modification de tests             |
| `chore`    | Maintenance, config, dependances        |
| `style`    | Formatting, whitespace                  |
| `perf`     | Amelioration de performance             |
| `ci`       | CI/CD configuration                     |

**INTERDIT** :

- Jamais de mention d'IA dans le message
- Jamais de "Generated by", "Co-authored-by", "AI", "Claude", "GPT"
- Jamais d'emoji dans les messages de commit

### 5. Push

```bash
git push -u origin <branch>
```

### 6. Creation de PR

**Methode 1 - MCP (PRIORITAIRE)** :

Utiliser `mcp__github__create_pull_request` avec :

- `owner` : proprietaire du repo (extrait de git remote)
- `repo` : nom du repo
- `title` : titre au format conventional commits
- `head` : branche source (branche courante)
- `base` : branche cible (main ou master)
- `body` : description de la PR

**Methode 2 - CLI (FALLBACK)** :

Si MCP non disponible, recuperer le token et utiliser gh :

```bash
source /workspace/.devcontainer/.env
GH_TOKEN=$(op item get "mcp-github" --vault "vault-id" --fields credential --reveal)
GH_TOKEN=$GH_TOKEN gh pr create --title "<titre>" --body "<body>"
```

**Format du body** :

```markdown
## Summary

<Description concise - 2-3 bullet points>

## Changes

<Liste des fichiers/composants modifies>

## Test plan

<Comment tester>
```

**INTERDIT dans la PR** :

- Jamais de mention d'IA
- Jamais de "Generated by", "AI-assisted"
- Jamais de Co-authored-by avec @anthropic, @openai

---

## Detection automatique du type

| Fichiers modifies                  | Type suggere     |
| ---------------------------------- | ---------------- |
| `src/` nouveaux fichiers           | `feat`           |
| `src/` corrections                 | `fix`            |
| `src/` restructuration             | `refactor`       |
| `tests/`, `*_test.*`, `*.spec.*`   | `test`           |
| `*.md`, `docs/`                    | `docs`           |
| `Dockerfile`, `.devcontainer/`, CI | `ci` ou `chore`  |
| `package.json`, `go.mod`, deps     | `chore`          |
| `.gitignore`, config files         | `chore`          |

---

## Detection automatique du scope

- Analyser le dossier principal modifie
- Exemples : `api`, `auth`, `db`, `ui`, `cli`, `config`
- Si multiple scopes : scope le plus significatif ou omis

---

## Output

### Succes

```text
## Commit & PR crees

| Etape   | Status                              |
|---------|-------------------------------------|
| Branche | `feat/add-user-auth`                |
| Commit  | `feat(auth): add user auth`         |
| Push    | origin/feat/add-user-auth           |
| PR      | #42 - feat(auth): add user auth     |

URL: https://github.com/<owner>/<repo>/pull/42
```

### Erreur - Pas de changements

```text
## Erreur

Aucun changement detecte a commiter.
```

---

## Cas speciaux

### --amend

1. Verifier qu'on n'est PAS sur main/master
2. Verifier que le dernier commit n'est pas pushe
3. `git commit --amend --no-edit`

### --no-pr

Skip l'etape 6, s'arrete apres le push.

### --branch (nom)

Force le nom de branche.

### --rename (nom)

Renomme la branche courante avant push/PR.
