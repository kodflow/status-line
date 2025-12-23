# Clean Architecture

> Uncle Bob - Indépendance des frameworks et testabilité maximale

## Concept

Couches concentriques avec dépendances vers l'intérieur uniquement.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Go** | Excellent |
| **Java** | Excellent |
| **TypeScript** | Très bon |
| **Kotlin** | Très bon |
| **C#** | Excellent |
| **Python** | Bon |

## Structure

```
/src
├── entities/            # Business objects (centre)
│   └── user.go
├── usecases/            # Application logic
│   ├── create_user.go
│   └── interfaces.go    # Port definitions
├── adapters/            # Interface adapters
│   ├── controllers/     # HTTP handlers
│   ├── presenters/      # Output formatting
│   └── gateways/        # Repository impl
└── frameworks/          # External (DB, Web)
    ├── database/
    └── web/
```

## Avantages

- Indépendant du framework
- Testabilité excellente
- Logique métier protégée
- Interchangeabilité (DB, UI, etc.)
- Maintenable long terme

## Inconvénients

- Verbeux
- Over-engineering pour petits projets
- Learning curve
- Beaucoup d'interfaces

## Contraintes

- Dépendances → vers le centre uniquement
- Entities ne connaissent RIEN d'externe
- Use cases définissent les interfaces (ports)
- Adapters implémentent les interfaces

## Règles

1. Dependency Rule : dépendances vers l'intérieur
2. Entities = logique métier universelle
3. Use Cases = logique application
4. Adapters = traduction in/out
5. Frameworks = détails (interchangeables)

## Quand utiliser

- Applications complexes
- Long terme (>2 ans)
- Changements fréquents de frameworks
- Tests critiques
- Équipe expérimentée

## Quand éviter

- POC/MVP rapide
- CRUD simple → MVC
- Scripts → Flat
- Équipe junior sans mentorat
