# Layered / N-Tier

> Architecture traditionnelle en couches horizontales

## Concept

Couches empilées : présentation → business → data. Chaque couche ne communique qu'avec celle du dessous.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Java** | Excellent (classique) |
| **C#** | Excellent (.NET) |
| **Python** | Bon |
| **PHP** | Bon |
| **Node.js** | Bon |

## Structure

```
/src
├── presentation/        # UI, API controllers
│   ├── controllers/
│   └── views/
├── business/            # Logique métier
│   ├── services/
│   └── validators/
└── data/                # Accès données
    ├── repositories/
    └── entities/
```

## Avantages

- Simple à comprendre
- Bien connu (classique)
- Séparation claire
- Facile à debugger

## Inconvénients

- Rigide (changes traversent toutes les couches)
- DB-centric (data drive tout)
- Logique métier diluée
- Testabilité moyenne

## Contraintes

- Couche N appelle couche N-1 uniquement
- Pas de saut de couche
- Pas de référence inverse

## Règles

1. Présentation → Business → Data
2. Pas de logique métier en présentation
3. Pas d'accès DB en présentation
4. Business ne connaît pas la présentation

## Quand utiliser

- Apps CRUD traditionnelles
- Équipes junior
- Migration d'ancien code
- Contraintes entreprise

## Quand éviter

- Logique métier complexe → Clean/Hexagonal
- Besoin de flexibilité → Hexagonal
- Scale → Sliceable Monolith
