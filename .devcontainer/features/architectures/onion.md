# Onion Architecture

> Jeffrey Palermo - Domain au centre, couches concentriques

## Concept

Similaire à Clean/Hexagonal mais avec couches nommées différemment.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **C#** | Excellent (.NET) |
| **Java** | Très bon |
| **TypeScript** | Bon |

## Structure

```
/src
├── core/                # Centre - Domain Model
│   └── entities/
├── domain/              # Domain Services
│   └── services/
├── application/         # Use Cases
│   ├── interfaces/
│   └── services/
└── infrastructure/      # External
    ├── persistence/
    └── external/
```

## Avantages

- Domain isolé
- Testabilité
- Dépendances vers le centre

## Inconvénients

- Confusion avec Clean/Hexagonal
- Verbeux
- Moins documenté

## Quand utiliser

- Équipes .NET
- Standards entreprise

## Quand éviter

- Autres langages → préférer Clean ou Hexagonal
- Petits projets
