# Status Line

CLI Go pour afficher une status line personnalisée dans Claude Code.

## Architecture

```
cmd/statusline/     # Point d'entrée CLI
internal/
├── adapter/        # Adaptateurs (Claude JSON input)
├── terminal/       # Rendu terminal (couleurs, largeur)
├── usage/          # Calcul usage API Anthropic
├── updater/        # Auto-update binaire
├── mcp/            # Détection serveurs MCP
└── system/         # Info système (OS, Docker)
```

## Développement

```bash
make build          # Compile le binaire
make test           # Lance les tests
make lint           # Vérifie le code
```

## Segments

| Segment | Description |
|---------|-------------|
| OS | Icône système (Linux/macOS/Windows/Docker) |
| Model | Pill colorée (Haiku/Sonnet/Opus) |
| Progress | Barre de contexte avec curseur burn-rate |
| Path | Répertoire courant |
| Git | Branche + fichiers modifiés/non-trackés |
| Changes | Lignes ajoutées/supprimées |

## Convention Go

- Tests: `*_test.go` dans le même package
- Pas de `/src`, utiliser `cmd/` et `internal/`
- Go 1.23+ (voir .devcontainer/features/languages/go/RULES.md)
