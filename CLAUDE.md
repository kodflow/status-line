<!-- updated: 2026-05-01T16:30:00Z -->
# Status Line

CLI Go pour afficher une status line Powerline personnalisée dans Claude Code.

## Architecture (Clean Architecture / Hexagonal)

```
cmd/statusline/              # Point d'entrée CLI (stdin JSON → stdout ANSI)
internal/
├── application/             # Service orchestration (StatusLineService)
├── domain/
│   ├── model/               # Entités (Input, Progress, Usage, Git, MCP, Taskwarrior...)
│   └── port/                # Interfaces (InputProvider, Renderer, GitRepository...)
├── adapter/                 # Adaptateurs externes
│   ├── git/                 # Git status + diff stats
│   ├── mcp/                 # Détection serveurs MCP (config files)
│   ├── system/              # Info système (OS, Docker)
│   ├── taskwarrior/         # Intégration Taskwarrior (épics/tâches)
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

## Segments (2 lignes)

**Ligne 1:**

| Segment | Description |
|---------|-------------|
| OS | Icône système (Linux/macOS/Windows/Docker) |
| Model | Pill colorée (Haiku/Sonnet/Opus) + barre progression session |
| Weekly | Barre burn-rate hebdomadaire (auto-hide si API indisponible) |
| Path | Répertoire courant (relatif au projet) |
| Git | Branche + fichiers modifiés/non-trackés |
| Changes | Lignes ajoutées/supprimées |

**Ligne 2:**

| Segment | Description |
|---------|-------------|
| Taskwarrior | Épic/tâche en cours avec barre de progression |
| MCP | Pills des serveurs MCP actifs |
| Update | Notification de mise à jour disponible |

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
