# Workflow /plan & /apply - Visualisation

Ce document présente les diagrammes de flux pour les commandes `/plan` et `/apply` du DevContainer Kodflow.

## Vue d'ensemble du workflow

```mermaid
flowchart TB
    subgraph User["Utilisateur"]
        A["/plan description"]
        Z["Review PR"]
    end

    subgraph PlanMode["PLAN MODE - Lecture seufaudrle"]
        B{Sur main/master?}
        C[Créer branche feat/ ou fix/]
        D[Session existante?]
        E[Charger session]
        F[Créer nouvelle session]

        subgraph Phases["6 Phases d'analyse - OBLIGATOIRES"]
            P1["Phase 1: Analyse demande"]
            P2["Phase 2: Recherche docs"]
            P3["Phase 3: Analyse projet"]
            P4["Phase 4: Affûtage"]
            P5["Phase 5: Définition épics/tasks"]
            P6["Phase 6: Écriture Taskwarrior"]
        end

        V{Validation utilisateur?}
    end

    subgraph ApplyMode["/apply - BYPASS MODE"]
        AP1["Vérifier état = planned"]
        AP2["State → applying"]

        subgraph TaskLoop["Boucle par Task"]
            T1["task-start.sh → WIP"]
            T2["Exécuter action"]
            T3["task-done.sh → DONE"]
        end

        AP3["Commit conventionnel"]
        AP4["Push branche"]
        AP5["Créer PR"]
        AP6["State → applied"]
    end

    A --> B
    B -->|OUI| C
    B -->|NON| D
    C --> F
    D -->|OUI| E
    D -->|NON| F
    E --> P1
    F --> P1

    P1 -->|"currentPhase=1 ✓"| P2
    P2 -->|"currentPhase=2 ✓"| P3
    P3 -->|"currentPhase=3 ✓"| P4
    P4 -->|Manque info| P2
    P4 -->|"currentPhase=4 ✓"| P5
    P5 --> V

    V -->|"NON - Refus utilisateur"| P1
    V -->|"OUI - Validation"| P6

    P6 -->|"currentPhase=6 ✓"| AP1

    AP1 --> AP2
    AP2 --> TaskLoop
    TaskLoop --> AP3
    AP3 --> AP4
    AP4 --> AP5
    AP5 --> AP6
    AP6 --> Z

    style PlanMode fill:#e1f5fe,stroke:#01579b
    style ApplyMode fill:#e8f5e9,stroke:#1b5e20
    style Phases fill:#fff3e0,stroke:#e65100
    style TaskLoop fill:#f3e5f5,stroke:#4a148c
    style V fill:#fff9c4,stroke:#f9a825
```

---

## Machine d'états avec phases

```mermaid
stateDiagram-v2
    [*] --> planning: /plan description

    state planning {
        [*] --> phase1
        phase1: Phase 1 - Analyse demande
        phase2: Phase 2 - Recherche docs
        phase3: Phase 3 - Analyse projet
        phase4: Phase 4 - Affûtage
        phase5: Phase 5 - Définition épics/tasks
        validation: Validation utilisateur

        phase1 --> phase2
        phase2 --> phase3
        phase3 --> phase4
        phase4 --> phase2: Manque info
        phase4 --> phase5
        phase5 --> validation
        validation --> phase1: REFUS → Reset complet
        validation --> [*]: OK
    }

    planning --> planned: Validation OK + Phase 6

    planned --> applying: /apply
    planned --> planning: /plan mise à jour → Reset phases

    applying --> applied: Toutes tasks DONE
    applying --> applying: Task en cours

    applied --> [*]: PR créée

    note right of planning
        phases[] track progression
        currentPhase = 1-6
        Hook bloque si saut
    end note
```

---

## Tracking des phases (Session JSON v3)

```mermaid
classDiagram
    class Session {
        +int schemaVersion = 3
        +string state
        +string type
        +string project
        +string branch
        +int currentPhase
        +array completedPhases
        +datetime phaseStartedAt
        +string currentTask
        +array lockedPaths
        +array epics
        +datetime createdAt
    }

    class PhaseTracking {
        +int phase
        +datetime startedAt
        +datetime completedAt
        +string status
        +object artifacts
    }

    Session "1" --> "*" PhaseTracking: completedPhases

    note for Session "currentPhase: 1-6\ncompletedPhases: historique"
    note for PhaseTracking "artifacts: résultats de la phase\n(ex: fichiers analysés, docs trouvés)"
```

**Nouveau schéma session v3 :**

```json
{
  "schemaVersion": 3,
  "state": "planning",
  "type": "feature",
  "project": "my-feature",
  "branch": "feat/my-feature",
  "currentPhase": 2,
  "completedPhases": [
    {
      "phase": 1,
      "startedAt": "2024-01-01T10:00:00Z",
      "completedAt": "2024-01-01T10:05:00Z",
      "status": "completed",
      "artifacts": {
        "requirements": ["auth system", "JWT tokens"],
        "constraints": ["no breaking changes"]
      }
    }
  ],
  "phaseStartedAt": "2024-01-01T10:05:00Z",
  "epics": [],
  "createdAt": "2024-01-01T10:00:00Z"
}
```

---

## Hook phase-validate.sh

```mermaid
flowchart TB
    Start["PreToolUse appelé"]

    Check1{Session existe?}
    Check2{state = planning?}
    Check3{currentPhase défini?}

    subgraph PhaseCheck["Validation progression phases"]
        PC1{Phase 6 sans<br/>phases 1-5 complétées?}
        PC2{Taskwarrior appelé<br/>sans validation?}
        PC3{Write épics sans<br/>phase 5 complétée?}
    end

    Block["BLOQUÉ<br/>Phases obligatoires"]
    Allow["AUTORISÉ"]

    Start --> Check1
    Check1 -->|NON| Allow
    Check1 -->|OUI| Check2
    Check2 -->|NON| Allow
    Check2 -->|OUI| Check3
    Check3 -->|NON| Block
    Check3 -->|OUI| PhaseCheck

    PC1 -->|OUI| Block
    PC1 -->|NON| PC2
    PC2 -->|OUI| Block
    PC2 -->|NON| PC3
    PC3 -->|OUI| Block
    PC3 -->|NON| Allow

    style Block fill:#ffcdd2,stroke:#b71c1c
    style Allow fill:#c8e6c9,stroke:#1b5e20
    style PhaseCheck fill:#fff3e0,stroke:#e65100
```

---

## Contraintes et Garde-fous

```mermaid
flowchart TB
    subgraph Forbidden["INTERDIT - Garde-fous absolus"]
        F1["Merge automatique"]
        F2["Push sur main/master"]
        F3["Skip PLAN MODE"]
        F4["Force push sans --force-with-lease"]
        F5["Mentions IA dans commits"]
    end

    subgraph BlockedPhases["BLOQUÉ - Phases obligatoires"]
        PH1["Sauter Phase 1 → Phase 4"]
        PH2["Phase 6 sans validation"]
        PH3["Taskwarrior sans phases 1-5"]
        PH4["Épics sans Phase 5 complète"]
    end

    subgraph BlockedPlan["BLOQUÉ en PLAN MODE"]
        BP1["Write/Edit sur code source"]
        BP2["Bash modifiant état"]
    end

    subgraph BlockedApply["BLOQUÉ en /apply"]
        BA1["Write/Edit sans task WIP"]
        BA2["/apply sans état planned"]
    end

    subgraph Allowed["AUTORISÉ"]
        A1["Write/Edit sur /plans/"]
        A2["Read, Glob, Grep"]
        A3["WebSearch, WebFetch"]
    end

    style Forbidden fill:#ffebee,stroke:#c62828
    style BlockedPhases fill:#fce4ec,stroke:#ad1457
    style BlockedPlan fill:#fff3e0,stroke:#e65100
    style BlockedApply fill:#f3e5f5,stroke:#7b1fa2
    style Allowed fill:#e8f5e9,stroke:#2e7d32
```

---

## Flow refus utilisateur

```mermaid
flowchart TB
    subgraph Phase5["Phase 5: Définition épics/tasks"]
        P5A["Présenter plan"]
        P5B["AskUserQuestion:<br/>Valider ce plan?"]
    end

    Decision{Réponse utilisateur}

    subgraph Reset["Reset complet"]
        R1["completedPhases = []"]
        R2["currentPhase = 1"]
        R3["epics = []"]
        R4["Stocker feedback utilisateur"]
    end

    subgraph Phase1Bis["Phase 1: Ré-analyse complète"]
        P1A["Relire demande originale"]
        P1B["Intégrer feedback refus"]
        P1C["Nouvelle analyse"]
    end

    subgraph Phase6["Phase 6: Écriture Taskwarrior"]
        P6A["Créer épics"]
        P6B["Créer tasks"]
        P6C["state → planned"]
    end

    P5A --> P5B
    P5B --> Decision

    Decision -->|"NON + feedback"| Reset
    Decision -->|"OUI"| Phase6

    Reset --> R1
    R1 --> R2
    R2 --> R3
    R3 --> R4
    R4 --> Phase1Bis

    P6A --> P6B
    P6B --> P6C

    style Reset fill:#ffcdd2,stroke:#b71c1c
    style Phase1Bis fill:#e3f2fd,stroke:#1565c0
    style Phase6 fill:#c8e6c9,stroke:#1b5e20
    style Decision fill:#fff9c4,stroke:#f9a825
```

---

## Flow des Hooks

```mermaid
flowchart LR
    subgraph PreToolUse["PreToolUse Hooks"]
        H1["pre-validate.sh"]
        H2["task-validate.sh"]
        H3["bash-validate.sh"]
        H4["phase-validate.sh"]
    end

    subgraph Action["Action Claude"]
        A1["Write"]
        A2["Edit"]
        A3["Bash"]
    end

    subgraph PostToolUse["PostToolUse Hooks"]
        P1["post-edit.sh"]
        P2["task-log.sh"]
    end

    subgraph PostEditChain["Chaîne post-edit.sh"]
        PE1["format.sh"]
        PE2["imports.sh"]
        PE3["lint.sh"]
        PE4["security.sh"]
        PE5["test.sh"]
    end

    H1 -->|Protège .claude/, .devcontainer/| A1
    H2 -->|Vérifie mode/task| A1
    H4 -->|Vérifie phases obligatoires| A1
    H1 --> A2
    H2 --> A2
    H4 --> A2
    H3 -->|Bloque commandes dangereuses| A3

    A1 --> P1
    A2 --> P1
    A1 --> P2
    A2 --> P2

    P1 --> PE1
    PE1 --> PE2
    PE2 --> PE3
    PE3 --> PE4
    PE4 --> PE5

    style PreToolUse fill:#ffecb3,stroke:#ff6f00
    style PostToolUse fill:#c8e6c9,stroke:#388e3c
    style PostEditChain fill:#e1bee7,stroke:#7b1fa2
    style H4 fill:#f48fb1,stroke:#c2185b
```

---

## Détail task-validate.sh

```mermaid
flowchart TB
    Start["PreToolUse: Write/Edit"]

    Check1{Session existe?}
    Check2{State = applying?}
    Check3{Task WIP existe?}
    Check4{Fichier dans ctx.files?}

    Block["BLOQUÉ"]
    Allow["AUTORISÉ"]

    Start --> Check1
    Check1 -->|NON| Block
    Check1 -->|OUI| Check2
    Check2 -->|NON| Block
    Check2 -->|OUI| Check3
    Check3 -->|NON| Block
    Check3 -->|OUI| Check4
    Check4 -->|NON| Block
    Check4 -->|OUI| Allow

    style Block fill:#ffcdd2,stroke:#b71c1c
    style Allow fill:#c8e6c9,stroke:#1b5e20
```

---

## Flow /plan --destroy

```mermaid
flowchart TB
    Start["/plan --destroy"]

    Check1{Session existe?}
    Check2{State = applying?}
    Confirm{Confirmation utilisateur?}

    Actions["Actions locales"]
    A1["git checkout main"]
    A2["git branch -D branch"]
    A3["rm session.json"]
    A4["task modify status:deleted"]

    End["Plan abandonné"]
    Error1["Aucun plan à détruire"]
    Error2["Impossible pendant applying"]

    Start --> Check1
    Check1 -->|NON| Error1
    Check1 -->|OUI| Check2
    Check2 -->|OUI| Error2
    Check2 -->|NON| Confirm
    Confirm -->|NON| End
    Confirm -->|OUI| Actions
    Actions --> A1
    A1 --> A2
    A2 --> A3
    A3 --> A4
    A4 --> End

    style Error1 fill:#ffcdd2,stroke:#b71c1c
    style Error2 fill:#ffcdd2,stroke:#b71c1c
    style End fill:#c8e6c9,stroke:#1b5e20
```

---

## Exécution parallèle des tasks

```mermaid
flowchart LR
    subgraph Sequential["Séquentiel"]
        S1["T2.1 parallel:no"]
    end

    subgraph Parallel["Parallèle"]
        P1["T2.2 parallel:yes"]
        P2["T2.3 parallel:yes"]
        P3["T2.4 parallel:yes"]
    end

    subgraph Wait["Synchronisation"]
        W1["Attendre fin"]
    end

    subgraph Next["Séquentiel"]
        N1["T2.5 parallel:no"]
    end

    S1 --> P1
    S1 --> P2
    S1 --> P3
    P1 --> W1
    P2 --> W1
    P3 --> W1
    W1 --> N1

    style Sequential fill:#e3f2fd,stroke:#1565c0
    style Parallel fill:#f3e5f5,stroke:#7b1fa2
    style Wait fill:#fff3e0,stroke:#ef6c00
    style Next fill:#e3f2fd,stroke:#1565c0
```

---

## Légende des couleurs

| Couleur | Signification |
|---------|---------------|
| Bleu clair | PLAN MODE - Lecture seule |
| Vert clair | BYPASS MODE / Autorisé |
| Orange | Phases / Hooks PreToolUse |
| Violet | Parallélisme / PostToolUse |
| Rose | Phases obligatoires |
| Jaune | Point de décision critique |
| Rouge clair | Bloqué / Interdit |

---

## Résumé des changements v3

| Avant | Après |
|-------|-------|
| Refus → retour Phase 5 | Refus → **Reset Phase 1** |
| Pas de tracking phases | `currentPhase` + `completedPhases` |
| Saut de phases possible | **Hook phase-validate.sh** bloque |
| Taskwarrior sans validation | **BLOQUÉ** par hook |

---

## Extension VSCode recommandée

Pour visualiser les diagrammes Mermaid directement dans VSCode, installez :

### Markdown Preview Mermaid Support

- ID: `bierner.markdown-mermaid`
- [Marketplace](https://marketplace.visualstudio.com/items?itemName=bierner.markdown-mermaid)

```bash
code --install-extension bierner.markdown-mermaid
```

Ou via la palette de commandes : `Ctrl+Shift+X` → chercher "mermaid"
