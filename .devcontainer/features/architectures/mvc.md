# MVC - Model View Controller

> **DEFAULT** pour projets web scripting (PHP, Ruby, Python/Django)

## Concept

Séparation en 3 couches : données, présentation, logique de contrôle.

## Langages recommandés

| Langage | Framework | Adaptation |
|---------|-----------|-----------|
| **PHP** | Laravel, Symfony | Excellent |
| **Ruby** | Rails | Excellent |
| **Python** | Django | Excellent |
| **Node.js** | Express | Bon |
| **Java** | Spring MVC | Bon |

## Structure

```
/src
├── models/              # Données, ORM, validations
│   ├── User.php
│   └── Post.php
├── views/               # Templates, UI
│   ├── layouts/
│   └── pages/
├── controllers/         # Logique requête/réponse
│   ├── UserController.php
│   └── PostController.php
├── routes/              # Définition URLs
└── config/
```

## Avantages

- Simple à comprendre
- Bien documenté (30+ ans)
- Frameworks matures
- Conventions établies
- Onboarding rapide

## Inconvénients

- Controllers obèses possible
- Couplage View-Model
- Difficulté à tester isolément
- Scale horizontal limité
- Logique métier dispersée

## Contraintes

- Controller = orchestration uniquement
- Pas de logique métier dans views
- Models = données + validations
- Un controller par ressource

## Règles

1. Controllers fins (orchestration)
2. Logique métier dans Models ou Services
3. Views sans logique (juste affichage)
4. Routes explicites et RESTful
5. Pas d'accès DB dans Controllers

## Quand utiliser

- Sites web traditionnels
- CRUD apps
- CMS, blogs, e-commerce simple
- Équipe junior
- Prototypage rapide

## Quand éviter

- Logique métier complexe → Clean/Hexagonal
- Besoin de scale → Sliceable Monolith
- API pure (sans views) → Layered
- Mobile app backend → Clean
