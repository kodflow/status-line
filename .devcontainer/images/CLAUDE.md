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

## FORMAT CONTEXTE JSON (ctx)

Chaque task a un contexte JSON annoté :

```json
{
  "files": ["src/auth.ts", "src/types.ts"],
  "action": "create|modify|delete|refactor",
  "deps": ["bcrypt", "jsonwebtoken"],
  "description": "Description détaillée de la task",
  "tests": ["src/__tests__/auth.test.ts"]
}
```

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
