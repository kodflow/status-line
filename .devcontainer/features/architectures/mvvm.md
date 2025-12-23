# MVVM - Model View ViewModel

> **DEFAULT** pour projets mobile (Flutter, SwiftUI, Compose)

## Concept

Séparation avec ViewModel comme intermédiaire réactif entre View et Model.

## Langages recommandés

| Langage | Framework | Adaptation |
|---------|-----------|-----------|
| **Dart** | Flutter | Excellent |
| **Swift** | SwiftUI | Excellent |
| **Kotlin** | Compose | Excellent |
| **TypeScript** | Vue.js, Angular | Très bon |
| **C#** | WPF, MAUI | Excellent |

## Structure

```
/src
├── models/              # Données, entités
│   └── user.dart
├── views/               # UI widgets/components
│   └── user_view.dart
├── viewmodels/          # État + logique présentation
│   └── user_viewmodel.dart
├── services/            # API, persistence
│   └── user_service.dart
└── di/                  # Injection de dépendances
```

## Avantages

- Binding bidirectionnel
- Testabilité (ViewModel isolé)
- Séparation UI/Logique claire
- Réactivité native
- Code UI simplifié

## Inconvénients

- Verbeux (beaucoup de fichiers)
- Learning curve binding
- Over-engineering possible
- State management complexe

## Contraintes

- ViewModel ne connaît PAS View
- View observe ViewModel
- Model est passif (données)
- Pas de logique dans View

## Règles

1. Un ViewModel par View (ou feature)
2. ViewModel expose des observables
3. View ne fait QUE afficher et capturer events
4. Model = pure data
5. Services pour side effects (API, DB)

## Quand utiliser

- Apps mobiles natives/cross-platform
- Desktop apps (WPF, MAUI)
- Frontend SPA réactif
- UI complexe avec état

## Quand éviter

- Backend/API → pas de View
- Sites statiques
- CLI apps
- Scripts
