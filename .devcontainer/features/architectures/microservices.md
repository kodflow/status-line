# Microservices

> Services indépendants, déployables séparément

## Concept

Chaque service = un repo, une équipe, un déploiement.

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Go** | Excellent |
| **Java** | Excellent |
| **Node.js** | Très bon |
| **Python** | Bon |
| **Rust** | Bon |

## Structure (multi-repo)

```
# Repo: user-service
/src
├── api/
├── domain/
├── infrastructure/
├── Dockerfile
└── k8s/

# Repo: order-service
/src
├── api/
├── domain/
├── infrastructure/
├── Dockerfile
└── k8s/

# Repo: shared-libs (optionnel)
/packages
├── auth-client/
└── common-types/
```

## Avantages

- Scale indépendant
- Déploiement indépendant
- Équipes autonomes
- Technologie par service
- Isolation des pannes

## Inconvénients

- Complexité opérationnelle
- Latence réseau
- Debugging distribué
- Transactions distribuées
- Duplication possible

## Contraintes

- Un service = une responsabilité
- Communication async privilégiée
- Pas de DB partagée
- API contracts stricts
- Observabilité obligatoire

## Règles

1. Un service ≤ 1 équipe (2 pizza rule)
2. API first (OpenAPI, gRPC)
3. Backward compatible toujours
4. Circuit breakers obligatoires
5. Tracing distribué requis

## Quand utiliser

- Grande organisation (>30 devs)
- Scale énorme requis
- Équipes indépendantes
- Polyglotte assumé

## Quand éviter

- Petite équipe (<10) → Sliceable Monolith
- MVP/POC → Monolith
- Budget ops limité
- Pas de DevOps mature
