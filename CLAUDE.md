<!-- updated: 2026-09-22T18:00:00Z -->
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
│   ├── mcp/                 # Serveurs MCP (config, --mcp-config de l'hôte, plugins)
│   ├── mcpcalls/            # Serveurs MCP appelés en ce moment (transcripts)
│   ├── sessionstate/        # Session occupée + pid hôte (<config>/sessions/<pid>.json)
│   ├── system/              # Info système (OS, Docker)
│   ├── terminal/            # Largeur du terminal (COLUMNS)
│   ├── transcript/          # Lecture de la fin d'un transcript JSONL
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
go run ./demo/widths            # ligne 1 d'une session chargée à 200…60 colonnes
go run ./demo/mcpline1 [-busy]  # maquettes : pastille MCP courte en ligne 1 (non branchée)
```

`demo/` = outils de dev (`demo/sample` = état réaliste partagé) ; ils
utilisent le vrai renderer mais rien de ce qu'ils composent n'est dans le produit.

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
| OS | Icône système + étincelles 󰙴 = état de Claude (vert/orange/rouge), puis `󰚩 N` sous-agents hors épic affiché |
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

**Largeur (ligne 1 seulement)** : `adapter/terminal` lit `COLUMNS` (posé
par l'hôte pour la commande de status line) ; absent, non entier, ≤ 0 ou
> 10000 → 120. Pas de `/dev/tty` ni d'exec. Budget = `COLUMNS − 4`
(`lineMargin` : l'hôte rembourre la ligne, et une ligne pile au bord passe à
la ligne au moindre désaccord sur un glyphe). Largeur visible
(`VisibleWidth`, `width.go`) : échappements CSI/OSC ignorés, glyphes Nerd
Font (PUA) = 1 cellule, blocs larges CJK/emoji = 2 (petite table), marques
combinantes = 0. `fit.go` : niveaux cumulatifs, dans l'ordre voulu par
l'utilisateur — 1 barre du contexte (reste icône + %), 2 barre du quota
scopé (libellé + % + compte à rebours), 3 barre hebdo, 4 barre session
(% seul), 5 chemin à 20, 6 branche à 20 runes (début gardé + `…`),
7 comptes à rebours — puis, pour tenir en 80 colonnes : 8 chemin à son
dernier élément et branche à 12, 9 noms de quota à l'initiale (`W`, `O`),
10 sans les changements (+/−), 11 sans le chemin (la branche reste ; hors
dépôt, jamais). Le premier niveau qui tient gagne, trouvé par bissection
(largeurs décroissantes le long des niveaux, testé) : 1 rendu si la ligne
tient, 5 au plus sinon (~20 µs vs ~80 µs sur i5-3210M). Aucun niveau ne
tient → le dernier. Largeur 0 (tests) = pas de contrainte. Une ligne
complète fait ~240 cellules, ~120 après l'étape 7, ~75 après la 11.

**Ligne 2 :** une pastille par épic ouvert mène la ligne, puis **une**
pastille MCP, puis la mise à jour. Aucune troncature : un titre long passe à la
ligne plutôt que d'être coupé.

## Serveurs MCP

`adapter/mcp` lit, par ordre de priorité (le premier qui nomme un serveur
gagne, dédoublonnage par nom) :

1. **managed** — `/etc/claude-code/managed-mcp.json` (macOS : `/Library/Application Support/ClaudeCode`)
2. **ligne de commande** — `--mcp-config` de l'hôte : `sessionstate` retrouve
   le pid dans `<config>/sessions/<pid>.json` (un seul scan, partagé avec
   `Working`), puis `/proc/<pid>/cmdline`. Répétable, variadique jusqu'au
   flag suivant, `--mcp-config=v` accepté ; valeur = JSON si elle commence par
   `{`, sinon chemin (relatif au `cwd` de l'hôte). `--strict-mcp-config` = seuls
   managed + ligne de commande comptent. Hors Linux (pas de `/proc`) : ignoré.
3. **local** — `projects[<dir>].mcpServers` du fichier global
4. **projet** — `<dir>/.mcp.json`, repli `mcp.json`
5. **user** — `mcpServers` du fichier global
6. **plugins** — `<config>/plugins/installed_plugins.json` ; activé si
   `enabledPlugins["<plugin>@<marketplace>"]` vaut `true` (`<config>/settings.json`,
   puis `<dir>/.claude/settings.json`, puis `settings.local.json`) ;
   serveurs dans `<installPath>/.mcp.json` (`mcpServers` ou map nue), marqués
   `Plugin`.

Fichier global = `$CLAUDE_CONFIG_DIR/.claude.json`, sinon `~/.claude.json`
puis `~/.claude/.claude.json`. `<config>` = `$CLAUDE_CONFIG_DIR` ou `~/.claude`.
Désactivé = `"disabled": true`, ou listé dans `projects[<dir>].disabledMcpServers`
(nom nu ou `plugin:<plugin>:<serveur>`) ; `disabledMcpjsonServers` vise
`.mcp.json`. Tout fichier illisible ou malformé est ignoré ; pas d'exec.

**Pastille (courte)** : à gauche le libellé `MCP` gras, blanc 255 sur
sarcelle foncée 23 (capuchon gauche en 23) ; une flèche `\ue0b0` (texte : `>`)
passe au corps sur sarcelle claire 116, encre 23 : les serveurs actifs
**comptés par portée** dans l'ordre de priorité, en minuscules —
`managed`, `cli`, `local`, `project`, `user`, `plugin` (`model.MCPSources`) —
portée vide omise, séparées par un point médian : `cli 5 · user 1 · plugin 1`.
Les désactivés, toutes portées confondues, ferment la liste en `· off N`,
barré, encre 239 (240 ne tient que 4.31:1 sur 116), si N > 0. L'adaptateur
pose `MCPServer.Source` (`WithSource`) sur chaque source avant la fusion : la
portée qui gagne le nom est celle comptée. Capuchon droit en 116. Aucun
serveur, aucune pastille. `demo/mcpline1` montre 5 emplacements possibles en
ligne 1 (dans l'OS, entre modèle et contexte, après git, dans le contexte,
en fin) : à 80 colonnes, aucun ne tient (~40 cellules de trop).

**Appels en cours** (`adapter/mcpcalls`, sans hook) : 128 Ko de fin de
`transcript_path` et des transcripts des sous-agents en cours
(`agents.json`, `<dir transcript>/<session>/subagents/agent-<id>.jsonl`).
Appel en vol = `tool_use` `mcp__<clé>__<outil>` sans `tool_result` plus loin ;
il reste allumé 2 s après le `timestamp` du résultat (la ligne se redessine
chaque seconde, un appel court ne se verrait jamais). Un appel sans résultat
depuis 30 min est tenu pour perdu. Clé → serveur (`model.WithBusy`) :
nom normalisé (`[^A-Za-z0-9_-]` → `_`), `plugin_<plugin>_<serveur>` → serveur ;
une clé inconnue est ajoutée, allumée, sans portée. Un appel allume
l'**entrée de la portée** qui possède le serveur (`cli 5` devient une pastille
dans la pastille aux couleurs du libellé, 255 gras sur 23, 7.5:1) ; un serveur
désactivé appelé allume `off N` ; un serveur sans portée s'affiche seul,
`+nom`, allumé, avant `off N`. Lecture de queue partagée avec `activity` : `adapter/transcript`.

## Effort

Claude Code n'envoie que le nom du niveau (`effort.level`), jamais l'échelle.
L'échelle `low → medium → high → xhigh → max` est figée dans
`model.effortScale` (testée) ; chaque niveau atteint est un disque dans
l'encre de la pastille, les autres un disque dans sa teinte pâle
(`GetModelTrack`). Un niveau hors échelle s'écrit en clair (`· ultra`), jamais
sur une jauge fausse.

## Tâches, épics et sous-agents

`adapter/tasks` lit deux sources, par session :

1. **MCP tasks de `kodflow-hooks`** (prioritaire dès que le fichier existe) :
   `<config>/kodflow/sessions/<session_id>/tasks.json`. v2 = liste `epics`
   (`id`, `agent`, `title` ≤ 20, `touched`) + `active` (agent → épic) ; un
   fichier v1 (`epics` en dict agent → épic courant) est converti à la
   lecture (`touched` = maintenant, `active[agent]` = son épic). Seules les
   tâches de `main` comptent ; `epic: 0` = hors épic ; une tâche d'un épic
   absent du fichier (épic fini, v1) est ignorée. Un fichier illisible
   n'affiche rien — il ne rebascule pas sur la source native.
2. **Outils natifs** (repli, fichier MCP absent) : `<config>/tasks/<liste>/<n>.json`,
   liste = `CLAUDE_CODE_TASK_LIST_ID` sinon `session-` + 8 premiers caractères,
   rendue comme la pastille `Tâches`.

**Épic ouvert** = au moins une tâche non terminée, ou l'épic actif encore
vide (il disparaît dès que le focus passe ailleurs). Un épic fini n'est plus
dessiné ; une tâche ajoutée à un épic fini le rouvre. `active.main` à 0,
absent ou pointant nulle part = aucun épic actif : la pastille `Tâches`
(tâches `epic: 0`) devient l'active. Ordre : l'actif, puis par `touched`
décroissant (`Tâches` : dernier `created`/`updated` de ses tâches).

**Pastilles (ligne 2)** : fond mauve 182, capuchons arrondis, encre prune 53.

- *Repliée* : `titre fait/total` (+ `󰚩 N`).
- *Dépliée* — seulement l'épic actif, seulement **pendant que la session
  travaille** : `titre fait/total cases titre-de-tâche` (+ `󰚩 N`). Cases
  triées : fait ■ encre 53, en cours ■ ambre `#82480b` qui **pulse** en
  `#9e5204` gras les secondes paires (horloge murale, `clockNow`), le reste
  (à faire, en attente) □ sur la piste pâle `#b48cb4`. Titre = la tâche en
  cours, sinon la première en attente de l'utilisateur, sinon la prochaine.
  Replis 256 : 94, 94 gras (le cube n'a pas d'ambre plus clair qui tienne
  3:1 sur le mauve — 130 tombe à 2.47:1), 139.
- Le pulse suppose `statusLine.refreshInterval: 1` — le minimum de l'hôte ;
  le clignotement ANSI (SGR 5) est filtré et ne sert à rien.

**Session occupée** : `adapter/sessionstate` lit `<config>/sessions/*.json`
(écrits par l'hôte) ; le premier dont `sessionId` = le `session_id` de stdin
décide : `status == "busy"`. Fichiers illisibles ignorés, pas d'exec.

**Sous-agents** : `agents.json` (hooks SubagentStart/Stop), `epic` = l'épic
actif de `main` au démarrage (absent = 0). En cours = pas de `stop` et
démarré il y a < 12 h. Un sous-agent dont l'épic est affiché va dans sa
pastille ; les autres (épic 0 ou épic non affiché) vont en **ligne 1**, dans
le segment OS après le glyphe de santé : `󰚩 N`.

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
