# Package / Library

> **DEFAULT** pour bibliothèques réutilisables

## Concept

Code destiné à être importé par d'autres projets.

## Langages recommandés

| Langage | Registry | Adaptation |
|---------|----------|-----------|
| **Go** | pkg.go.dev | Excellent |
| **Rust** | crates.io | Excellent |
| **Node.js** | npm | Excellent |
| **Python** | PyPI | Très bon |
| **Java** | Maven Central | Bon |
| **Ruby** | RubyGems | Bon |

## Structure

```
/src
├── lib/                 # Code public (exported)
│   ├── client.go
│   ├── types.go
│   └── errors.go
├── internal/            # Code privé (non exported)
│   └── helpers.go
├── examples/            # Exemples d'utilisation
│   └── basic/
├── README.md
├── LICENSE
├── CHANGELOG.md
└── go.mod / package.json / Cargo.toml
```

## Avantages

- Réutilisable
- Versionné (semver)
- Documenté
- Testé
- Maintenable

## Inconvénients

- API publique = engagement
- Breaking changes sensibles
- Documentation obligatoire
- Backward compatibility

## Contraintes

- Semantic versioning strict
- Public API stable
- Documentation exhaustive
- Tests >90% coverage
- Pas de dépendances lourdes

## Règles

1. API publique minimale
2. Internal pour l'implémentation
3. Examples obligatoires
4. CHANGELOG maintenu
5. Breaking = major version

## Conventions

```
v1.0.0 → v1.1.0  # Nouvelle feature (backward compatible)
v1.1.0 → v1.1.1  # Bug fix
v1.1.1 → v2.0.0  # Breaking change
```

## Quand utiliser

- Code partagé entre projets
- SDK/Client pour API
- Utilitaires communs
- Open source

## Quand éviter

- Application finale → autre architecture
- Code non réutilisable
- Prototype
