<!-- updated: 2026-02-16T14:59:00Z -->
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
