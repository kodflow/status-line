# DDD - Domain-Driven Design

> Eric Evans - Modélisation du domaine métier au centre

## Variantes

| Variante | Focus |
|----------|-------|
| **Tactical** | Patterns (Entity, VO, Aggregate) |
| **Strategic** | Bounded Contexts, Ubiquitous Language |
| **+ CQRS** | Sépare lecture/écriture |
| **+ Event Sourcing** | État via événements |

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Java** | Excellent |
| **C#** | Excellent |
| **Scala** | Excellent |
| **TypeScript** | Très bon |
| **Go** | Bon |
| **Python** | Bon |

## Structure (Strategic + Tactical)

```
/src
├── contexts/                    # Bounded Contexts
│   └── <context>/
│       ├── domain/
│       │   ├── aggregates/      # Racines d'agrégat
│       │   ├── entities/        # Entités
│       │   ├── valueobjects/    # Value Objects
│       │   ├── events/          # Domain Events
│       │   ├── repositories/    # Interfaces
│       │   └── services/        # Domain Services
│       ├── application/
│       │   ├── commands/
│       │   ├── queries/
│       │   └── handlers/
│       └── infrastructure/
│           └── persistence/
└── shared/
    └── kernel/                  # Shared Kernel
```

## Avantages

- Alignement métier/code
- Langage ubiquitaire
- Boundaries claires
- Évolutif
- Complexité maîtrisée

## Inconvénients

- Complexe à apprendre
- Over-engineering si mal utilisé
- Experts domaine requis
- Verbeux

## Contraintes

- Aggregate = unité de cohérence
- Entity = identité propre
- Value Object = immuable, sans identité
- Repository = persistence d'Aggregates uniquement
- Domain Service = logique sans entité

## Règles

1. Un Aggregate = une transaction
2. Référencer Aggregates par ID uniquement
3. Invariants dans Aggregate Root
4. Events pour communication inter-contextes
5. Anti-corruption layer entre contextes

## Quand utiliser

- Domaine métier complexe
- Logique métier riche
- Équipe avec experts domaine
- Long terme (>3 ans)

## Quand éviter

- CRUD simple → MVC
- Pas d'expert domaine
- POC/MVP → Flat ou MVC
- Équipe junior
