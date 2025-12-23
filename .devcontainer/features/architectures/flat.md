# Flat / Scripts

> **DEFAULT** pour CLI tools, scripts, POC

## Concept

Structure minimale, fichiers au même niveau.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Go** | Excellent (single binary) |
| **Python** | Excellent (scripts) |
| **Rust** | Excellent (CLI) |
| **Bash** | Bon (scripts système) |
| **Node.js** | Bon |

## Structure

```
/src
├── main.go              # Entry point
├── config.go            # Configuration
├── commands.go          # CLI commands
├── utils.go             # Helpers
└── types.go             # Types/structs
```

Ou avec sous-dossiers légers :

```
/src
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   └── utils/
└── pkg/                 # Si réutilisable
```

## Avantages

- Simple
- Rapide à démarrer
- Facile à comprendre
- Peu de boilerplate
- Un fichier = visible

## Inconvénients

- Ne scale pas
- Pas de séparation
- Refactoring difficile
- Tests limités

## Contraintes

- < 10 fichiers idéalement
- < 1000 lignes par fichier
- Pas de logique métier complexe

## Règles

1. Un fichier par responsabilité
2. Nommage explicite
3. Pas de dépendances complexes
4. Migrate vers Clean si ça grossit

## Quand utiliser

- CLI tools
- Scripts système
- POC/Prototypes
- Outils internes
- Automation

## Quand éviter

- Logique métier → Clean/Hexagonal
- Web app → MVC
- Scale prévu → Sliceable Monolith
- >2000 lignes → Refactor
