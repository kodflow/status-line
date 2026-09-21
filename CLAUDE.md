<!-- updated: 2026-09-21T12:00:00Z -->
# Status Line

CLI Go pour afficher une status line Powerline personnalisée dans Claude Code.

## Architecture (Clean Architecture / Hexagonal)

```
cmd/statusline/              # Point d'entrée CLI (stdin JSON → stdout ANSI)
internal/
├── application/             # Service orchestration (StatusLineService)
├── domain/
│   ├── model/               # Entités (Input, Limit, LimitSet, Progress, Git, MCP...)
│   └── port/                # Interfaces (InputProvider, Renderer, GitRepository...)
├── adapter/                 # Adaptateurs externes
│   ├── git/                 # Git status + diff stats
│   ├── mcp/                 # Détection serveurs MCP (config files)
│   ├── system/              # Info système (OS, Docker)
│   ├── terminal/            # Info terminal (largeur, couleurs)
│   ├── updater/             # Auto-update binaire (GitHub releases)
│   └── usage/               # Usage API Anthropic (OAuth, burn-rate)
└── presentation/
    └── renderer/            # Rendu Powerline ANSI (segments, couleurs, icônes)
```

## Développement

```bash
make build          # Compile le binaire → bin/status-line
make test           # Lance les tests (go test ./...)
make lint           # Vérifie le code (ktn-linter)
make demo           # Démo avec données exemple
```

## Affichage

`STATUSLINE_LINE_GAP` = lignes vides entre les deux rangées, 0-3 (défaut `0`)
`STATUS_LINE_NO_SELF_UPDATE` = `1` désactive l'auto-update (images managées)

L'auto-update vérifie le `.sha256` publié avec l'asset avant de remplacer le
binaire : une somme absente, malformée ou différente annule la mise à jour.
`STATUSLINE_GLYPHS` = `nerd` (défaut) | `text` (repli ASCII, sans Nerd Font)
`STATUSLINE_COLORS` = `truecolor` | `256` force la profondeur de couleur ;
absent, elle suit `COLORTERM` (`truecolor`/`24bit` → 24 bits, sinon 256)
`STATUSLINE_HIDE` = pastilles à masquer, séparées par des virgules :
`context`, `session`, `weekly`, `model`, `credits`, `health`

Les glyphes viennent tous des plages Nerd Font, comme le reste de la ligne : un
symbole Unicode générique retombe sur une autre police et devient illisible.

**Ligne 1 — tout ce qui concerne la session :**

| Segment | Description |
|---------|-------------|
| OS | Icône système + étincelles 󰙴 = état de Claude (vert/orange/rouge) |
| Model | Pill colorée (Haiku/Sonnet/Opus/Fable) + jauge d'effort + fast mode |
| Path | Répertoire où la session travaille réellement (voir ci-dessous) |
| Git | Branche + modifiés/non-trackés + nombre de worktrees liés (hors prunable) |
| Changes | Lignes ajoutées/supprimées |

Les quotas du compte (session 5h, hebdo, quota scopé au modèle courant) sont
rendus dans le segment du modèle, séparés par un `\ue0b1`. Un quota scopé à une
famille de modèles ne s'affiche que si ce modèle est en cours d'utilisation.

| Segment | Description |
|---------|-------------|

| Pastille | Description |
|----------|-------------|
| ctx | Fenêtre de contexte, en pourcentage |
| session | Quota 5h : repère de brûlure régulière, atterrissage, reset |
| weekly | Quota 7j global — absent sur les forfaits qui n'en ont pas |
| *modèle* | Quota 7j scopé par famille de modèle (`limits[]`) |
| coût / credits | Coût cumulé de la session, solde de crédits |

**Ligne ambiante:** pills MCP, notification de mise à jour.

**Ligne 2 :** la liste de tâches de la session mène la ligne, puis les pills
MCP. Aucune troncature : un titre long passe à la ligne plutôt que d'être coupé.

## Effort

Claude Code n'envoie que le nom du niveau (`effort.level`), jamais l'échelle.
L'échelle `low → medium → high → xhigh → max` est figée dans
`model.effortScale` (testée) ; chaque niveau atteint est un disque dans
l'encre de la pastille, les autres un disque dans sa teinte pâle
(`GetModelTrack`). Un niveau hors échelle s'écrit en clair (`· ultra`), jamais
sur une jauge fausse.

## Tâches

`adapter/tasks` lit `<config>/tasks/<liste>/<n>.json` (un fichier par tâche,
écrit par TaskCreate/TaskUpdate). La liste vaut `CLAUDE_CODE_TASK_LIST_ID`
sinon `session-` + 8 premiers caractères de `session_id`. Barre segmentée
(fait / en cours / à faire), `fait/total`, titre de la tâche en cours ; une
liste terminée n'est plus dessinée. Depuis Claude Code 2.1.278 les outils de
tâches ne sont offerts d'office qu'aux modèles antérieurs à Opus/Sonnet 5 :
`CLAUDE_CODE_ENABLE_TODO_TOOLS=1` les réactive.

## Répertoire actif

`workspace.current_dir` ne bouge pas quand l'agent travaille par `cd X && …`
ou par chemins absolus. `adapter/activity` lit les 256 derniers Ko de
`transcript_path` et prend le dernier emplacement nommé par un appel d'outil
(`file_path`/`path`, `cd X` en tête, `git -C X`), remonté à sa racine git.
`~/.claude/projects` (journaux, mémoire) est ignoré. Git s'exécute dans ce
répertoire ; MCP garde le répertoire de session.

## État de Claude

`adapter/health` lit `status.claude.com/api/v2/summary.json`, hors composant
« Government » et hors groupes. 1 composant dégradé/partiel = orange ; ≥ 2, ou
un `major_outage` = rouge ; la maintenance ne compte pas. Cache 2 min
rafraîchi par `--refresh-health` détaché ; au-delà de 15 min, rien n'est
dessiné. Le rendu ne touche jamais le réseau.

## Palette

- Fonds pâles (xterm-256), encre sombre de la même teinte : chaque couple
  encre/fond tient **≥ 4.5:1** (`contrast_internal_test.go`, qui lit aussi
  les échappements 24 bits `38;2;r;g;b`).
- Pastilles modèle : fonds poudrés (Haiku 224, Sonnet 189, Opus 223, Fable
  194, inconnu 252). Encres en **24 bits** vers ~5.4:1 — le cube 256 n'a rien
  dans ces teintes entre une encre profonde (7:1+) et une sous 4.5:1. Chaque
  encre a un repli 256 (`inkXxx256`) : la couleur du cube la plus proche en
  CIELAB qui tient 4.5:1 (Opus saute l'olive 58, jugé sale sur la pêche).
  Les deux jeux sont testés ; `pickInk` choisit au démarrage.
- Le curseur de rythme `●` et les barres prennent l'encre de leur pastille :
  pas de couleur d'accent fixe (l'ancien orange 166 tombait à 2:1 sur Sonnet).
- Contexte : fond 152, encre 24. Deux segments adjacents ne partagent jamais
  un fond proche (`TestAdjacentSegmentsDifferInGround`).

## Quotas : d'où viennent les chiffres

stdin (`rate_limits`, Claude Code >= 2.1.140) est prioritaire — gratuit et
synchrone. L'API OAuth n'apporte que ce que stdin ignore : quotas scopés par
modèle, crédits, et les buckets non envoyés. Côté API, `limits[]` est la source
agnostique au forfait ; `five_hour`/`seven_day` ne sont qu'un repli et valent
`null` sur certains forfaits.

**Un quota absent des deux sources reste absent** : l'afficher à 0 % mentirait
sur le compte. Réponse API mise en cache 60 s, rafraîchie par un processus
détaché — l'affichage n'attend jamais le réseau.

## Convention Go

- Tests: `*_test.go` dans le même package (`_internal_test` et `_external_test`)
- Structure: `cmd/` et `internal/` (pas de `/src`)
- Go 1.25.5 (go.mod), toolchain Go 1.26

## Documentation

| Fichier | Contenu |
|---------|---------|
| [`docs/vision.md`](docs/vision.md) | Vision projet, problème résolu, principes, non-goals |
| [`docs/architecture.md`](docs/architecture.md) | Diagramme C4, composants, data flow, contraintes |
| [`docs/workflows.md`](docs/workflows.md) | Setup, dev loop, tests, déploiement, CI/CD |
| [`AGENTS.md`](AGENTS.md) | Mapping stack → agents spécialistes (`/review`, `/lint`) |

## Reviews IA (label-triggered)

CodeRabbit et Qodo Merge ne tournent **pas** automatiquement sur chaque PR
(trop bruyant alongside l'un de l'autre). Ils sont déclenchés par label :

| Label | Outil | Quand l'utiliser |
|-------|-------|------------------|
| `coderabbit` | CodeRabbit | Review approfondie (89 règles ast-grep + path_instructions) |
| `qodo` | Qodo Merge | Second avis indépendant (security review + tests review) |

Les chemins `.devcontainer/**`, `.github/**`, `vendor/**` sont exclus des
deux outils — c'est du contenu template-managed, pas du code projet.

## Branch protection (main)

| Règle | Valeur |
|-------|--------|
| Required status checks | `test`, `build (linux\|darwin\|windows, amd64\|arm64)` (7 checks) |
| Required approving reviews | 0 (solo maintainer) |
| Dismiss stale reviews on push | ✓ (CodeRabbit CHANGES_REQUESTED s'auto-dismiss au push suivant) |
| Force push | ✗ |
| Branch deletion | ✗ |
