# MVP - Model View Presenter

> Variante de MVC avec Presenter au lieu de Controller

## Concept

Le Presenter contient toute la logique de présentation, la View est passive.

## Langages recommandés

| Langage | Framework | Adaptation |
|---------|-----------|-----------|
| **Kotlin** | Android (legacy) | Bon |
| **Java** | Android (legacy) | Bon |
| **C#** | WinForms | Bon |

## Structure

```
/src
├── models/              # Données
├── views/               # UI passive
│   └── interfaces/      # View contracts
└── presenters/          # Logique présentation
```

## Avantages

- View testable (mock presenter)
- Séparation claire
- View passive = simple

## Inconvénients

- Presenter peut grossir
- Boilerplate interfaces
- Moins populaire aujourd'hui

## Quand utiliser

- Android legacy
- Migration depuis MVC

## Quand éviter

- Nouveau projet mobile → MVVM
- Web → MVC
- Backend → Clean/Hexagonal
