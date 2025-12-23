# Sliceable Monolith

> **DEFAULT RECOMMANDÉ** pour projets backend/API scalables

## Concept

Monolithe modulaire où chaque domaine peut être extrait et déployé indépendamment.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Go** | Excellent - modules natifs |
| **Java** | Excellent - Spring Modulith |
| **Node.js/TS** | Très bon - workspaces |
| **Rust** | Très bon - workspace Cargo |
| **Python** | Bon - packages |
| **Scala** | Bon - sbt multi-project |
| **Elixir** | Bon - umbrella apps |

## Structure

```
/src
├── shared/                     # Shared Kernel
│   ├── kernel/                 # Types, interfaces
│   └── infra/                  # DB, messaging, config
│
├── domains/                    # Bounded Contexts
│   └── <domain>/
│       ├── api/                # HTTP/gRPC handlers
│       ├── application/        # Use cases
│       ├── domain/             # Business logic
│       ├── infrastructure/     # Implementations
│       ├── Dockerfile          # Standalone deploy
│       └── main.go             # Isolated entry
│
├── cmd/
│   ├── monolith/               # All domains
│   └── <domain>/               # Single domain
│
└── deployments/
    ├── docker-compose.yml
    └── k8s/
```

## Avantages

- Dev simple (monorepo, un build)
- Scale granulaire (extrait ce qui en a besoin)
- Pas de duplication (shared kernel)
- Migration progressive (pas de big bang)
- Tests intégrés faciles
- Refactoring sûr (tout dans un repo)

## Inconvénients

- Discipline requise (boundaries strictes)
- Complexité initiale plus élevée
- Nécessite conventions d'équipe
- Shared kernel = couplage potentiel

## Contraintes

- Chaque domain DOIT être autonome
- Communication inter-domain via events/interfaces
- Pas d'import direct entre domains
- Shared kernel minimal et stable
- Chaque domain a son propre Dockerfile

## Règles

1. Un domain ne peut PAS importer un autre domain directement
2. Communication via shared kernel (events, interfaces)
3. Chaque domain expose une API publique claire
4. Infrastructure par domain (pas de DB partagée)
5. Tests par domain + tests d'intégration globaux

## Commandes

```bash
make build                    # Monolithe complet
make build-domain D=billing   # Domain seul
make extract D=billing        # Prépare extraction
make test                     # Tests tous domains
make test-domain D=billing    # Tests un domain
```

## Quand utiliser

- Projet qui va scaler mais incertain où
- Équipe moyenne (3-15 devs)
- Besoin de flexibilité déploiement
- Domaines métier clairs
- Budget K8s/infra variable

## Quand éviter

- Petit projet/POC (trop de structure)
- Équipe énorme (>30) avec ownership clair → microservices
- Script/CLI simple → Flat
- Web traditionnel PHP/Ruby → MVC
