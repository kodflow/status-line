# Feature - Développement de fonctionnalités

$ARGUMENTS

---

## Description

Workflow complet pour développer une nouvelle fonctionnalité avec **suivi Taskwarrior obligatoire** :
1. **PLAN MODE** → Analyse, recherche, définition des epics/tasks
2. **Validation utilisateur** → Approbation du plan avant exécution
3. **BYPASS MODE** → Exécution des tasks une par une
4. **CI validation** → Vérifier que la pipeline passe
5. **PR sans merge** → Créer la PR, merge manuel requis

**Chaque action Write/Edit est tracée** et bloquée si aucune task Taskwarrior n'est en WIP.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<description>` | Nouvelle feature avec ce nom |
| `--continue` | Reprendre la feature en cours (via session) |
| `--status` | Afficher le statut de la branche courante |
| `--help` | Affiche l'aide de la commande |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /feature - Developpement de fonctionnalites
═══════════════════════════════════════════════

Usage: /feature <description> [options]

Options:
  <description>     Nouvelle feature avec ce nom
  --continue        Reprendre la feature en cours
  --status          Afficher le statut de la branche
  --help            Affiche cette aide

Exemples:
  /feature add-auth         Cree feat/add-auth
  /feature --continue       Reprend la derniere session
  /feature --status         Affiche l'etat de la PR
═══════════════════════════════════════════════
```

---

## Workflow complet

### Étape 0 : Initialisation (OBLIGATOIRE)

**AVANT toute action**, initialiser le projet et la branche :

```bash
# Exécuter le script d'initialisation
/home/vscode/.claude/scripts/task-init.sh "feature" "<description>"

# Déterminer la branche principale
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")

# Sync avec remote et créer la branche
git fetch origin
BRANCH="feat/$(echo "$DESCRIPTION" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g')"
git checkout -b "$BRANCH" "origin/$MAIN_BRANCH"
```

**Output attendu :**
```
═══════════════════════════════════════════════
  ✓ Projet initialisé: <project-name>
═══════════════════════════════════════════════

  Mode: PLAN (analyse et définition des epics/tasks)

  Phases PLAN MODE:
    1. Analyse de la demande
    2. Recherche documentation
    3. Analyse projet existant
    4. Affûtage (boucle si nécessaire)
    5. Définition épics/tasks → VALIDATION
    6. Écriture Taskwarrior

  Après validation → BYPASS MODE (exécution)

  Branch: feat/<project-name>
  Session: $HOME/.claude/sessions/<project-name>.json
═══════════════════════════════════════════════
```

---

### Étape 1 : PLAN MODE (6 phases obligatoires)

**Mode actif** : PLAN (pas d'édition de code autorisée)

#### Phase 1 : Analyse de la demande
- Comprendre ce que l'utilisateur veut
- Identifier les contraintes et exigences

#### Phase 2 : Recherche documentation
- WebSearch pour APIs/libs externes
- Lire docs existantes du projet

#### Phase 3 : Analyse projet existant
- Glob/Grep pour trouver code existant
- Read fichiers pertinents
- Comprendre patterns/architecture

#### Phase 4 : Affûtage
- Croiser infos (demande + docs + existant)
- Si manque info → retour Phase 2
- Identifier tous les fichiers à modifier

#### Phase 5 : Définition épics/tasks → VALIDATION USER

**Output attendu :**
```
Epic 1: <nom>
  ├─ Task 1.1: <description> [parallel:no]
  ├─ Task 1.2: <description> [parallel:yes]
  └─ Task 1.3: <description> [parallel:yes]
Epic 2: <nom>
  ├─ Task 2.1: <description> [parallel:no]
  └─ Task 2.2: <description> [parallel:no]
```

**Chaque task doit avoir un contexte JSON :**
```json
{
  "files": ["src/auth.ts", "src/types.ts"],
  "action": "create|modify|delete|refactor",
  "deps": ["bcrypt", "jsonwebtoken"],
  "description": "Description détaillée de la task",
  "tests": ["src/__tests__/auth.test.ts"]
}
```

Puis `AskUserQuestion: "Valider ce plan ?"`

#### Phase 6 : Écriture Taskwarrior

Après validation utilisateur :

```bash
SESSION_FILE=$(ls -t $HOME/.claude/sessions/*.json | head -1)
PROJECT=$(jq -r '.project' "$SESSION_FILE")

# Créer les epics
EPIC1_UUID=$(/home/vscode/.claude/scripts/task-epic.sh "$PROJECT" 1 "Setup infrastructure")
EPIC2_UUID=$(/home/vscode/.claude/scripts/task-epic.sh "$PROJECT" 2 "Implementation")

# Créer les tasks dans chaque epic
/home/vscode/.claude/scripts/task-add.sh "$PROJECT" 1 "$EPIC1_UUID" "Créer dossier structure" "no" '{"files":["src/"],"action":"create"}'
/home/vscode/.claude/scripts/task-add.sh "$PROJECT" 2 "$EPIC2_UUID" "Implémenter AuthService" "no" '{"files":["src/services/auth.ts"],"action":"create"}'

# Le mode passe automatiquement en bypass quand on démarre une task
```

---

### Étape 2 : BYPASS MODE (exécution)

**Mode actif** : BYPASS (édition autorisée SI task WIP)

#### Workflow par task

```bash
# 1. Démarrer la task (TODO → WIP)
/home/vscode/.claude/scripts/task-start.sh <uuid>

# 2. Exécuter la task
# - Lire le contexte JSON (files, action, description)
# - Effectuer les modifications requises
# - Chaque Write/Edit est automatiquement tracé

# 3. Commit conventionnel
git add <files>
git commit -m "feat(<scope>): <description>"
git push -u origin "$BRANCH"

# 4. Terminer la task (WIP → DONE)
/home/vscode/.claude/scripts/task-done.sh <uuid>

# 5. Passer à la task suivante
```

#### Exécution parallèle (automatique)

Si plusieurs tasks consécutives ont `parallel:yes` :
- Démarrer TOUTES en WIP simultanément
- Lancer multiples Tool calls en parallèle
- Attendre que TOUTES soient DONE avant continuer

**Exemple :**
```
Task 2.1 [parallel:no]  → exécuter seul
Task 2.2 [parallel:yes] ┐
Task 2.3 [parallel:yes] ├→ exécuter en parallèle
Task 2.4 [parallel:yes] ┘
Task 2.5 [parallel:no]  → attendre 2.2-2.4, puis exécuter
```

---

### Étape 3 : Sync avec main (si nécessaire)

Si des commits ont été ajoutés sur main pendant le développement :

```bash
git fetch origin "$MAIN_BRANCH"
git rebase "origin/$MAIN_BRANCH"

# Si conflits :
# 1. Résoudre les conflits
# 2. git add <resolved-files>
# 3. git rebase --continue

git push --force-with-lease
```

---

### Étape 4 : Vérification CI

#### Détection du provider Git

```bash
REMOTE=$(git remote get-url origin 2>/dev/null)

case "$REMOTE" in
    *github.com*)    PROVIDER="github" ;;
    *gitlab.com*)    PROVIDER="gitlab" ;;
    *bitbucket.org*) PROVIDER="bitbucket" ;;
    *)               PROVIDER="unknown" ;;
esac
```

#### Vérification du statut (ordre de priorité)

**1. MCP connecté :**
```
mcp__github__get_pull_request_status (si GitHub)
mcp__gitlab__get_merge_request (si GitLab)
```

**2. CLI disponible :**
```bash
gh pr checks "$BRANCH"        # GitHub
glab mr view "$BRANCH"        # GitLab
```

#### En cas d'échec CI

1. Analyser les logs d'erreur
2. Identifier la cause (tests, lint, build, etc.)
3. Corriger le problème
4. Commit + push
5. Réessayer (max 3 tentatives)

---

### Étape 5 : Création PR

**Via MCP (priorité) :**
```
mcp__github__create_pull_request
mcp__gitlab__create_merge_request
```

**Via CLI (fallback) :**
```bash
gh pr create --title "feat: $DESCRIPTION" --body "..."
glab mr create --title "feat: $DESCRIPTION" --description "..."
```

**Format du body :**
```markdown
## Summary
- <Point 1>
- <Point 2>

## Changes
- `path/to/file1.ts` : description
- `path/to/file2.ts` : description

## Test plan
- [ ] Test 1
- [ ] Test 2
```

---

## GARDE-FOUS (ABSOLUS)

### INTERDICTIONS

| Action | Status |
|--------|--------|
| Merge automatique | **INTERDIT** |
| Push sur main/master | **INTERDIT** |
| Skip PLAN MODE | **INTERDIT** |
| Write/Edit sans task WIP | **BLOQUÉ** |
| Force push sans --force-with-lease | **INTERDIT** |

### Message de fin

```
═══════════════════════════════════════════════
  ✓ Feature prête !

  Branche : feat/<description>
  PR : https://github.com/<owner>/<repo>/pull/<number>
  CI : ✓ Passed

  ⚠️  MERGE MANUEL REQUIS
  → Le merge automatique est désactivé
  → Revue de code recommandée avant merge
═══════════════════════════════════════════════
```

---

## --continue

Reprendre une feature en cours via la session Taskwarrior :

```bash
SESSION_DIR="$HOME/.claude/sessions"

# Trouver la session la plus récente (ou spécifier un projet)
if [[ -n "$1" ]]; then
    SESSION_FILE="$SESSION_DIR/$1.json"
else
    SESSION_FILE=$(ls -t "$SESSION_DIR"/*.json 2>/dev/null | head -1)
fi

if [[ ! -f "$SESSION_FILE" ]]; then
    echo "❌ Aucune session trouvée"
    echo "→ Utilisez /feature <description> pour démarrer"
    exit 1
fi

PROJECT=$(jq -r '.project' "$SESSION_FILE")
BRANCH=$(jq -r '.branch' "$SESSION_FILE")
MODE=$(jq -r '.mode' "$SESSION_FILE")
CURRENT_EPIC=$(jq -r '.current_epic // "N/A"' "$SESSION_FILE")
CURRENT_TASK=$(jq -r '.current_task // "N/A"' "$SESSION_FILE")

echo "═══════════════════════════════════════════════"
echo "  Reprise: $PROJECT"
echo "═══════════════════════════════════════════════"
echo ""
echo "  Mode: $MODE"
echo "  Epic courant: $CURRENT_EPIC"
echo "  Task courante: $CURRENT_TASK"
echo ""

# Afficher les epics et leur statut
echo "  Epics:"
jq -r '.epics[] | "    \(.id). \(.name) [\(.status)]"' "$SESSION_FILE" 2>/dev/null

echo ""
echo "═══════════════════════════════════════════════"

# Vérifier la branche git
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null)
if [[ "$CURRENT_BRANCH" != "$BRANCH" ]]; then
    echo ""
    echo "⚠ Branche actuelle: $CURRENT_BRANCH"
    echo "→ Branche attendue: $BRANCH"
    echo "→ Exécuter: git checkout $BRANCH"
fi
```

---

## --status

Afficher le statut de la feature :

```
## Statut : feat/<description>

| Élément | Status |
|---------|--------|
| Branche | feat/<description> |
| Mode | PLAN / BYPASS |
| Epics | 2/3 terminés |
| Tasks | 5/8 terminées |
| PR | #42 (open) |
| CI | ✓ Passed |
| Merge | En attente (manuel) |
```

---

## Outputs

### Initialisation
```
═══════════════════════════════════════════════
  /feature add-user-authentication
═══════════════════════════════════════════════

✓ Branche créée : feat/add-user-authentication
✓ Base : origin/main (abc1234)
✓ Mode : PLAN

→ Commencez par analyser la demande...
```

### Après validation du plan
```
═══════════════════════════════════════════════
  ✓ Plan validé - Passage en BYPASS MODE
═══════════════════════════════════════════════

  Epics créés: 3
  Tasks créées: 8

  Prochaine task:
    Epic 1: Setup infrastructure
    Task 1.1: Créer structure dossiers

  → Démarrer avec: task-start.sh <uuid>
═══════════════════════════════════════════════
```

### Après CI success
```
═══════════════════════════════════════════════
  ✓ Feature prête !

  Branche : feat/add-user-authentication
  Commits : 3
  PR : https://github.com/owner/repo/pull/42
  CI : ✓ Passed (2m 34s)

  ⚠️  MERGE MANUEL REQUIS
═══════════════════════════════════════════════
```
