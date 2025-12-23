# Hexagonal / Ports & Adapters

> Alistair Cockburn - Domain au centre, isolé du monde extérieur

## Concept

Le domain est au centre, communique via des ports (interfaces) implémentés par des adapters.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Go** | Excellent |
| **Java** | Excellent |
| **TypeScript** | Très bon |
| **Rust** | Très bon |
| **Python** | Bon |
| **Scala** | Très bon |

## Structure

```
/src
├── domain/              # Cœur métier (aucune dépendance)
│   ├── model/           # Entities, Value Objects
│   ├── services/        # Domain services
│   └── events/          # Domain events
├── ports/               # Interfaces (contrats)
│   ├── inbound/         # Driven by (API, CLI)
│   │   └── user_service.go
│   └── outbound/        # Driving (DB, external)
│       └── user_repository.go
└── adapters/            # Implémentations
    ├── inbound/         # HTTP, gRPC, CLI
    │   └── http/
    └── outbound/        # Postgres, Redis, APIs
        └── postgres/
```

## Avantages

- Domain complètement isolé
- Testable sans infrastructure
- Adaptable (change de DB = change d'adapter)
- Ports explicites
- Symétrie in/out

## Inconvénients

- Beaucoup d'indirection
- Verbeux
- Complexe pour petits projets
- Discipline requise

## Contraintes

- Domain = ZERO dépendance externe
- Ports = interfaces dans domain
- Adapters = implémentent les ports
- Injection de dépendances obligatoire

## Règles

1. Domain ne connaît que lui-même
2. Ports inbound = ce que l'app offre
3. Ports outbound = ce dont l'app a besoin
4. Un adapter par technologie
5. Tests domain sans mocks d'infra

## Quand utiliser

- Apps avec logique métier riche
- Besoin de flexibilité technique
- Tests critiques
- Équipe moyenne à grande

## Quand éviter

- CRUD simple → MVC
- Scripts → Flat
- Prototypage rapide
