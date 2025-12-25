# Plan - Infrastructure as Code Planning

$ARGUMENTS

---

## Description

Commande de planification façon Terraform. Crée un état déterministe dans Taskwarrior :
- Analyse complète (6 phases **OBLIGATOIRES**)
- Génération des epics/tasks
- État reproductible et versionné

**Comportement intelligent :**
- **Sur `main`** → Création automatique de branche (pas de question)
- **`/plan` répété** → Met à jour le plan existant (affinage itératif)
- **Refus validation** → Reset complet Phase 1 (ré-analyse intégrale)

**Workflow** : `/plan <desc>` → (affiner) → `/plan` → validation → `/apply`

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<description>` | Nouveau plan OU mise à jour du plan existant |
| `--fix` | Mode bugfix (branche fix/ au lieu de feat/) |
| `--status` | Afficher l'état du plan actuel |
| `--destroy` | Abandonner le plan et nettoyer |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /plan - Infrastructure as Code Planning
═══════════════════════════════════════════════

Usage: /plan <description> [options]

Options:
  <description>     Nouveau plan ou mise à jour
  --fix             Mode bugfix (branche fix/)
  --status          Afficher l'état du plan
  --destroy         Abandonner et nettoyer
  --help            Affiche cette aide

Comportement:
  Sur main          → Crée branche automatiquement
  /plan répété      → Met à jour le plan existant
  Refus validation  → Reset Phase 1

Exemples:
  /plan add-auth            Nouveau plan feature
  /plan                     Affiner le plan en cours
  /plan login-bug --fix     Nouveau plan bugfix
  /plan --status            Voir l'état

Workflow:
  /plan <desc> → /plan (affiner) → /apply
═══════════════════════════════════════════════
```

---

## Concept : État déterministe

Comme Terraform, `/plan` produit un état reproductible :

```
Session JSON = État du plan (comme terraform.tfstate)
Taskwarrior  = Ressources déclarées (comme les resources TF)
/apply       = Application de l'état (comme terraform apply)
```

### Fichier de session (schéma v3)

```json
{
  "schemaVersion": 3,
  "state": "planning",
  "type": "feature|fix",
  "project": "<project-name>",
  "branch": "feat/<name>|fix/<name>",
  "currentPhase": 1,
  "completedPhases": [],
  "validated": false,
  "validatedAt": null,
  "validationToken": null,
  "validationHistory": [],
  "rejectionHistory": [],
  "actions": [],
  "epics": [],
  "createdAt": "2024-01-01T00:00:00Z"
}
```

**Champs v3 :**

- `currentPhase` : Phase en cours (1-6)
- `completedPhases` : Historique des phases complétées avec timestamps
- `validated` : Validation utilisateur obtenue
- `validationToken` : Hash unique de validation (traçabilité)
- `validationHistory` : Historique de toutes les validations/rejets
- `rejectionHistory` : Historique des refus avec feedback
- `actions` : Tableau d'événements horodatés (audit trail)

### États possibles

| State | Description | Transition |
|-------|-------------|------------|
| `planning` | En cours d'analyse | → `planned` |
| `planned` | Prêt pour /apply | → `applying` |
| `applying` | Exécution en cours | → `applied` |
| `applied` | Terminé (PR créée) | FIN |

---

## Scripts de transition (OBLIGATOIRES)

**IMPORTANT** : Les transitions de phase sont gérées par des scripts atomiques.
Ne JAMAIS utiliser de commandes jq directes pour modifier la session.

### session-transition.sh

```bash
# Compléter une phase
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 1

# Voir l'état
/home/vscode/.claude/scripts/session-transition.sh --status

# Reset à une phase antérieure (si manque info)
/home/vscode/.claude/scripts/session-transition.sh --to-phase 2

# Reset complet (refus utilisateur)
/home/vscode/.claude/scripts/session-transition.sh --reset "feedback utilisateur"

# Finaliser planning → state=planned
/home/vscode/.claude/scripts/session-transition.sh --finalize
```

### session-validate.sh

```bash
# Approuver le plan (génère token de validation)
/home/vscode/.claude/scripts/session-validate.sh --approve

# Rejeter le plan (reset complet phase 1)
/home/vscode/.claude/scripts/session-validate.sh --reject "raison du refus"

# Voir état validation
/home/vscode/.claude/scripts/session-validate.sh --status
```

---

## Workflow complet

### Étape 0 : Initialisation

```bash
CURRENT_BRANCH=$(git branch --show-current)
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")
SESSION_FILE=$(ls -t $HOME/.claude/sessions/*.json 2>/dev/null | head -1)

# CAS 1 : Sur main → Créer nouvelle branche (AUTOMATIQUE, LOCAL)
if [[ "$CURRENT_BRANCH" == "$MAIN_BRANCH" || "$CURRENT_BRANCH" == "master" ]]; then
    TYPE="${HAS_FIX:+fix}" || "feature"
    PREFIX="${TYPE:0:4}"
    BRANCH="$PREFIX/$(echo "$DESCRIPTION" | tr ' ' '-' | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]//g')"

    git checkout -b "$BRANCH" "$MAIN_BRANCH"
    /home/vscode/.claude/scripts/task-init.sh "$TYPE" "<description>"
fi

# CAS 2 : Session existante → Mise à jour du plan
if [[ -f "$SESSION_FILE" ]]; then
    STATE=$(jq -r '.state' "$SESSION_FILE")
    if [[ "$STATE" == "planning" || "$STATE" == "planned" ]]; then
        echo "Plan existant détecté, mise à jour..."
    fi
fi

# CAS 3 : Sur branche sans session → Créer session
if [[ ! -f "$SESSION_FILE" && "$CURRENT_BRANCH" != "$MAIN_BRANCH" ]]; then
    TYPE=$([[ "$CURRENT_BRANCH" == fix/* ]] && echo "fix" || echo "feature")
    /home/vscode/.claude/scripts/task-init.sh "$TYPE" "${CURRENT_BRANCH#*/}"
fi
```

---

### Étape 1 : PLAN MODE (6 phases OBLIGATOIRES)

**INTERDIT en PLAN MODE :**
- ❌ Write/Edit sur fichiers code
- ❌ Bash modifiant l'état du projet (voir bash-validate.sh)
- ❌ Sauter des phases (hook phase-validate.sh)
- ❌ Écrire dans Taskwarrior sans validation
- ❌ Commandes jq directes sur la session
- ✅ Write/Edit sur `/plans/` uniquement
- ✅ Scripts session-*.sh

#### Phase 1 : Analyse de la demande

**Objectif :** Comprendre ce que l'utilisateur veut

**Actions :**
- Lire et reformuler la demande
- Identifier contraintes et exigences
- Pour un fix : identifier les étapes de reproduction

**Fin de phase :**
```bash
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 1
```

#### Phase 2 : Recherche documentation

**Objectif :** Collecter les informations externes nécessaires

**Actions :**
- WebSearch pour APIs/libs externes
- Lire docs existantes du projet
- Pour un fix : rechercher bugs similaires

**Fin de phase :**
```bash
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 2
```

#### Phase 3 : Analyse projet existant

**Objectif :** Comprendre le code existant

**Actions :**
- Glob/Grep pour trouver code existant
- Read fichiers pertinents
- Comprendre patterns/architecture
- Pour un fix : reproduire le bug

**Fin de phase :**
```bash
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 3
```

#### Phase 4 : Affûtage

**Objectif :** Synthétiser et valider la compréhension

**Actions :**
- Croiser infos (demande + docs + existant)
- Si manque info → **retour Phase 2** :
  ```bash
  /home/vscode/.claude/scripts/session-transition.sh --to-phase 2
  ```
- Identifier tous les fichiers à modifier
- Pour un fix : identifier la cause racine

**Fin de phase :**
```bash
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 4
```

#### Phase 5 : Définition épics/tasks → VALIDATION OBLIGATOIRE

**Output attendu :**
```
═══════════════════════════════════════════════
  Plan généré
═══════════════════════════════════════════════

Epic 1: <nom>
  ├─ T1.1: <description> [parallel:no]
  │        files: [src/api.ts]
  │        action: create
  ├─ T1.2: <description> [parallel:yes]
  └─ T1.3: <description> [parallel:yes]

Epic 2: <nom>
  ├─ T2.1: <description> [parallel:no]
  └─ T2.2: <description> [parallel:no]

─────────────────────────────────────────────
  Résumé
─────────────────────────────────────────────

  Epics  : 2
  Tasks  : 5
  Files  : 8 fichiers modifiés

═══════════════════════════════════════════════
```

**VALIDATION UTILISATEUR (OBLIGATOIRE) :**

```
AskUserQuestion: "Valider ce plan ?"
```

**Si OUI (approuvé) :**
```bash
/home/vscode/.claude/scripts/session-validate.sh --approve
/home/vscode/.claude/scripts/session-transition.sh --complete-phase 5
```

**Si NON (rejeté) → RESET COMPLET Phase 1 :**
```bash
/home/vscode/.claude/scripts/session-validate.sh --reject "raison du refus"
```

```
═══════════════════════════════════════════════
  ⚠ Plan refusé - Reset Phase 1
═══════════════════════════════════════════════

  Feedback utilisateur stocké.
  Ré-analyse complète en cours...

  → Retour Phase 1 avec nouveau contexte

═══════════════════════════════════════════════
```

#### Phase 6 : Écriture Taskwarrior

**PRÉ-CONDITION :** `validated = true` (vérifié par session-transition.sh --finalize)

Après validation utilisateur :

```bash
SESSION_FILE=$(cat /workspace/.claude/active-session)
PROJECT=$(jq -r '.project' "$SESSION_FILE")

# Créer les epics
/home/vscode/.claude/scripts/task-epic.sh "$PROJECT" 1 "Epic 1 name"
/home/vscode/.claude/scripts/task-epic.sh "$PROJECT" 2 "Epic 2 name"

# Créer les tasks
/home/vscode/.claude/scripts/task-add.sh "$PROJECT" 1 "<uuid>" "Task name" "no" '{"files":["..."],"action":"..."}'

# Finaliser planning → state=planned
/home/vscode/.claude/scripts/session-transition.sh --finalize
```

**Output final :**
```
═══════════════════════════════════════════════
  ✓ Plan enregistré
═══════════════════════════════════════════════

  State   : planned
  Phases  : 6/6 ✓
  Epics   : 2
  Tasks   : 5

  Le plan est prêt. Pour l'exécuter :

    /apply

  Pour voir le plan :

    /plan --status

═══════════════════════════════════════════════
```

---

## --fix

Identique au workflow standard mais avec :
- Branche `fix/<name>` au lieu de `feat/<name>`
- Commit prefix `fix(scope):` au lieu de `feat(scope):`
- PR body format "Bug / Root cause / Fix" au lieu de "Summary / Changes"

```bash
/home/vscode/.claude/scripts/task-init.sh "fix" "<description>"
```

---

## --status

Afficher l'état complet du plan :

```bash
/home/vscode/.claude/scripts/session-transition.sh --status
/home/vscode/.claude/scripts/session-validate.sh --status
```

```
═══════════════════════════════════════════════
  État du plan
═══════════════════════════════════════════════

  Project : <name>
  Type    : feature|fix
  State   : planning|planned|applying|applied
  Branch  : <branch>
  Phase   : 3/6

─────────────────────────────────────────────
  Phases
─────────────────────────────────────────────

  [✓] 1. Analyse demande
  [✓] 2. Recherche docs
  [→] 3. Analyse projet    ← EN COURS
  [ ] 4. Affûtage
  [ ] 5. Définition + Validation
  [ ] 6. Écriture Taskwarrior

─────────────────────────────────────────────
  Validation
─────────────────────────────────────────────

  Validated : false
  Token     : none

═══════════════════════════════════════════════
```

---

## --destroy (safe, local only)

Abandonner le plan et nettoyer **localement** :

**Pré-conditions :**

- ❌ INTERDIT si `state=applying` (exécution en cours)
- ✅ Confirmation utilisateur obligatoire

```bash
SESSION_FILE=$(cat /workspace/.claude/active-session)
PROJECT=$(jq -r '.project' "$SESSION_FILE")
BRANCH=$(jq -r '.branch' "$SESSION_FILE")
STATE=$(jq -r '.state' "$SESSION_FILE")
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")

# Bloquer si applying
if [[ "$STATE" == "applying" ]]; then
    echo "❌ Impossible de détruire un plan en cours d'exécution"
    exit 1
fi

# Demander confirmation
AskUserQuestion: "Abandonner le plan et supprimer la branche locale $BRANCH ?"
```

**Actions (après confirmation) :**
```bash
# Nettoyer LOCALEMENT
git checkout "$MAIN_BRANCH"
git branch -D "$BRANCH"
rm "$SESSION_FILE"
rm -f /workspace/.claude/active-session
rm -f /workspace/.claude/state.json

# Archiver tasks Taskwarrior
task project:"$PROJECT" rc.confirmation=off modify status:deleted
```

**Output :**
```
═══════════════════════════════════════════════
  ✓ Plan abandonné (local)
═══════════════════════════════════════════════

  Branche locale supprimée : <branch>
  Session supprimée        : <file>
  Tasks archivées          : <count>

  Note: La branche remote n'est pas supprimée.

═══════════════════════════════════════════════
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Hook/Script | Status |
|--------|-------------|--------|
| Write/Edit code en PLAN MODE | `task-validate.sh` | ❌ **BLOQUÉ** |
| Bash écriture en PLAN MODE | `bash-validate.sh` | ❌ **BLOQUÉ** |
| Sauter des phases (1→4) | `session-transition.sh` | ❌ **BLOQUÉ** |
| Taskwarrior sans validation | `phase-validate.sh` | ❌ **BLOQUÉ** |
| Skip validation utilisateur | `session-validate.sh` | ❌ **INTERDIT** |
| Passer à /apply sans "planned" | `task-validate.sh` | ❌ **BLOQUÉ** |
| Commandes jq directes session | `bash-validate.sh` | ❌ **BLOQUÉ** |

---

## Voir aussi

- `/apply` - Exécuter le plan
- `/review` - Demander une code review
- `/git --commit` - Commit manuel
- `.devcontainer/docs/workflow-plan-apply.md` - Diagrammes Mermaid
