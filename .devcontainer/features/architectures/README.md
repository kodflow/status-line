# Architectures

## Choix par défaut

| Type de projet | Architecture par défaut |
|----------------|------------------------|
| API/Backend scalable | **Sliceable Monolith** |
| Web scripting (PHP, Ruby) | **MVC** |
| CLI/Tools | **Flat** |
| Library/Package | **Package** |
| Mobile (Flutter) | **MVVM** |

## Liste des architectures

| Architecture | Fichier | Cas d'usage |
|--------------|---------|-------------|
| MVC | `mvc.md` | Web apps, PHP, Ruby, Django |
| MVP | `mvp.md` | Desktop, Android legacy |
| MVVM | `mvvm.md` | Mobile, Frontend reactif |
| Layered | `layered.md` | Apps traditionnelles |
| Clean | `clean.md` | Apps complexes, testabilité |
| Hexagonal | `hexagonal.md` | Domain-centric, ports & adapters |
| Onion | `onion.md` | Enterprise, .NET |
| DDD | `ddd.md` | Domaines complexes |
| Microservices | `microservices.md` | Grande équipe, scale indépendant |
| Sliceable Monolith | `sliceable-monolith.md` | **Recommandé** - Meilleur des deux mondes |
| Event-Driven | `event-driven.md` | Async, découplage fort |
| Serverless | `serverless.md` | Functions, pay-per-use |
| Flat | `flat.md` | Scripts, CLI, POC |
| Package | `package.md` | Libraries réutilisables |
