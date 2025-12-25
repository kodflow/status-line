# Kodflow DevContainer - Workflow Obligatoire

## MODES DE TRAVAIL

Tu dois TOUJOURS travailler dans l'un de ces deux modes :

### PLAN MODE (Analyse - tu réfléchis)

**Quand:** Au démarrage de `/feature` ou `/fix`

**Phases obligatoires:**
1. **Analyse demande** - Comprendre ce que l'utilisateur veut
2. **Recherche documentation** - WebSearch, docs projet
3. **Analyse projet** - Glob/Grep/Read pour comprendre l'existant
4. **Affûtage** - Croiser les infos (retour phase 2 si manque info)
5. **Définition épics/tasks** - Présenter le plan → **VALIDATION USER**
6. **Écriture Taskwarrior** - Créer épics et tasks

**INTERDIT en PLAN MODE:**
- ❌ Write/Edit sur fichiers code
- ❌ Bash modifiant l'état du projet
- ✅ Write/Edit sur fichiers `/plans/` uniquement

**Output attendu phase 5:**
```
Epic 1: <nom>
  ├─ Task 1.1: <description> [parallel:no]
  ├─ Task 1.2: <description> [parallel:yes]
  └─ Task 1.3: <description> [parallel:yes]
Epic 2: <nom>
  ├─ Task 2.1: <description> [parallel:no]
  └─ Task 2.2: <description> [parallel:no]
```

Puis `AskUserQuestion: "Valider ce plan ?"`

---

### BYPASS MODE (Exécution - tu agis)

**Quand:** Après validation du plan et écriture dans Taskwarrior

**Workflow par task:**
```bash
# 1. Démarrer la task (TODO → WIP)
/home/vscode/.claude/scripts/task-start.sh <uuid>

# 2. Exécuter la task avec le contexte JSON
# (lire ctx.files, ctx.action, ctx.description)

# 3. Terminer la task (WIP → DONE)
/home/vscode/.claude/scripts/task-done.sh <uuid>

# 4. Passer à la task suivante
```

**OBLIGATOIRE en BYPASS MODE:**
- ✅ Une task DOIT être WIP avant Write/Edit
- ✅ Suivre l'ordre des tasks (sauf parallel:yes)
- ❌ Write/Edit sans task WIP = BLOQUÉ

**Exécution parallèle (automatique):**
Si plusieurs tasks consécutives ont `parallel:yes`:
- Démarrer TOUTES en WIP simultanément
- Lancer multiples Tool calls en parallèle
- Attendre que TOUTES soient DONE avant continuer

---

## COMMANDES TASKWARRIOR

| Script | Usage |
|--------|-------|
| `task-init.sh <type> <desc>` | Initialiser projet |
| `task-epic.sh <project> <num> <name>` | Créer un epic |
| `task-add.sh <project> <epic> <uuid> <name> [parallel] [ctx]` | Ajouter task |
| `task-start.sh <uuid>` | TODO → WIP |
| `task-done.sh <uuid>` | WIP → DONE |

Chemin: `/home/vscode/.claude/scripts/`

---

## STRUCTURE TASKWARRIOR

```
project:"feat-xxx"              # Conteneur global
├─ Epic 1 (+epic)               # Phase
│  ├─ Task 1.1 (+task)          # Action atomique
│  ├─ Task 1.2 [parallel:yes]
│  └─ Task 1.3 [parallel:yes]
└─ Epic 2
   └─ Task 2.1
```

---

## FORMAT CONTEXTE JSON (ctx) - v2

Chaque task a un contexte JSON annoté (schéma v2) :

```json
{
  "schemaVersion": 2,
  "files": ["src/auth.ts", "src/types.ts"],
  "action": "create|modify|delete|refactor|test|document",
  "locks": ["src/auth.ts", "src/types/*.ts"],
  "deps": ["bcrypt", "jsonwebtoken"],
  "description": "Description détaillée de la task",
  "tests": ["src/__tests__/auth.test.ts"],
  "acceptance_criteria": [
    "tests passent",
    "lint ok",
    "no breaking change"
  ],
  "commands": ["go test ./...", "npm run lint"],
  "risk": "low|medium|high",
  "rollback": "git revert HEAD"
}
```

**Champs obligatoires:** `schemaVersion`, `files`, `action`

**Nouveaux champs v2:**

- `locks`: Chemins verrouillés (empêche parallel tasks sur mêmes fichiers)
- `acceptance_criteria`: Critères mesurables de succès
- `commands`: Commandes de validation
- `risk`: Niveau de risque (low/medium/high)
- `rollback`: Instructions de rollback

**External ID:** Chaque task a un ID lisible (ex: T1.2) stocké dans les annotations.

---

## SESSION JSON - Schéma v2

### Convention de nommage

| Élément | Convention |
|---------|------------|
| Dossier | `~/.claude/sessions/` |
| Fichier | `<project>.json` où `project` = slug de la branche sans préfixe |
| Exemple | Branche `feat/auth-system` → `auth-system.json` |

### Structure session (schemaVersion: 2)

```json
{
  "schemaVersion": 2,
  "state": "planning|planned|applying|applied",
  "type": "feature|fix",
  "project": "<slug>",
  "branch": "feat/<slug>|fix/<slug>",
  "currentTask": null,
  "currentEpic": null,
  "lockedPaths": [],
  "epics": [...],
  "createdAt": "<ISO8601>"
}
```

### Machine d'états

```
planning ──→ planned ──→ applying ──→ applied
    │            │
    └────────────┘ (itératif via /plan)
```

| State | Description | Transitions autorisées |
|-------|-------------|------------------------|
| `planning` | Analyse en cours | → `planned` |
| `planned` | Plan validé, prêt pour /apply | → `applying`, → `planning` |
| `applying` | Exécution en cours | → `applied` |
| `applied` | Terminé (PR créée) | FIN |

### Invariants obligatoires

- `schemaVersion` = 2 (obligatoire)
- `state` ∈ {planning, planned, applying, applied}
- `type` ∈ {feature, fix}
- `branch` cohérente avec `type` (préfixe `feat/` ou `fix/`)
- `project` non vide, slug-safe (a-z, 0-9, -)
- `epics[].id` monotone croissant
- `tasks[].externalId` unique, format `T<epic>.<num>`
- `tasks[].parallel` obligatoire (yes|no)

---

## HOOKS ACTIFS

| Hook | Déclencheur | Action |
|------|-------------|--------|
| `task-validate.sh` | PreToolUse (Write/Edit) | Bloque si mode/task invalide |
| `task-log.sh` | PostToolUse | Log l'action dans Taskwarrior |
| `pre-validate.sh` | PreToolUse | Protège fichiers critiques |
| `post-edit.sh` | PostToolUse | Format + Lint auto |

---

## GARDE-FOUS ABSOLUS

| Action | Status |
|--------|--------|
| Merge automatique | ❌ **INTERDIT** |
| Push sur main/master | ❌ **INTERDIT** |
| Skip PLAN MODE | ❌ **INTERDIT** |
| Write/Edit sans task WIP | ❌ **BLOQUÉ** |
| Force push sans --force-with-lease | ❌ **INTERDIT** |

---

## RÉSUMÉ

```
/feature ou /fix
       │
       ▼
┌─────────────────┐
│   PLAN MODE     │ ← Analyse, pas d'édition code
│                 │
│ 1. Analyse      │
│ 2. Recherche    │
│ 3. Existant     │
│ 4. Affûtage     │
│ 5. Épics/Tasks  │ → Validation utilisateur
│ 6. Taskwarrior  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  BYPASS MODE    │ ← Exécution, task WIP obligatoire
│                 │
│ Pour chaque task│
│  → start        │
│  → execute      │
│  → done         │
└────────┬────────┘
         │
         ▼
      PR créée
```

---

## SERVEURS MCP (Model Context Protocol)

### Vérification obligatoire au démarrage

**TOUJOURS** vérifier `/workspace/.mcp.json` au début de chaque session :

```bash
# Lire la config MCP
cat /workspace/.mcp.json 2>/dev/null | jq -r '.mcpServers | keys[]'
```

### Priorité d'utilisation

| Action | Priorité 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| GitHub PR | `mcp__github__create_pull_request` | `gh pr create` |
| GitHub Issues | `mcp__github__create_issue` | `gh issue create` |
| Merge PR | `mcp__github__merge_pull_request` | `gh pr merge` |

**RÈGLE ABSOLUE :** Si un outil MCP est disponible, l'utiliser EN PRIORITÉ.

### Ne JAMAIS demander ce qui est déjà configuré

Si `.mcp.json` contient un serveur (ex: `github`), **NE PAS** :
- ❌ Demander un token GitHub à l'utilisateur
- ❌ Suggérer `gh auth login`
- ❌ Utiliser le CLI en fallback si MCP disponible

**TOUJOURS** :
- ✅ Utiliser directement les outils `mcp__<server>__*`
- ✅ En cas d'échec MCP, informer l'utilisateur du problème
- ✅ Extraire le token de `.mcp.json` si CLI fallback nécessaire

### Diagnostic des erreurs MCP

Si les outils MCP ne sont pas disponibles alors que `.mcp.json` existe :

1. **Vérifier le démarrage** : `super-claude` affiche les serveurs actifs
2. **Vérifier Node.js** : Les serveurs MCP utilisent `npx`, Node.js doit être installé
3. **Logs d'erreur** : Les échecs de démarrage sont affichés au lancement

```bash
# Tester un serveur manuellement
GITHUB_PERSONAL_ACCESS_TOKEN="..." npx -y @modelcontextprotocol/server-github
```
