# Event-Driven Architecture

> Communication par événements asynchrones

## Concept

Composants découplés communiquant via événements (pub/sub).

## Langages recommandés

| Langage | Adaptation |
|---------|-----------|
| **Java** | Excellent (Kafka, RabbitMQ) |
| **Go** | Excellent (NATS, Kafka) |
| **Node.js** | Très bon |
| **Scala** | Excellent (Akka) |
| **Elixir** | Excellent (natif) |
| **Python** | Bon |

## Structure

```
/src
├── events/              # Définitions événements
│   ├── user_created.go
│   └── order_placed.go
├── producers/           # Émetteurs
│   └── user_service/
├── consumers/           # Récepteurs
│   └── notification_service/
├── handlers/            # Event handlers
│   └── on_user_created.go
└── infrastructure/
    └── messaging/       # Kafka, RabbitMQ, NATS
```

## Avantages

- Découplage fort
- Scalabilité (async)
- Résilience
- Extensibilité (add consumers)
- Audit trail naturel

## Inconvénients

- Complexité debugging
- Eventual consistency
- Ordering challenges
- Idempotency requise
- Infrastructure complexe

## Contraintes

- Events = immutables
- Consumers = idempotents
- At-least-once delivery assumé
- Schema evolution gérée

## Règles

1. Event = fait passé (past tense)
2. Un event = une responsabilité
3. Consumer indépendant du producer
4. Retry + dead letter queue
5. Event versioning obligatoire

## Quand utiliser

- Workflows async
- Intégration systèmes
- Audit/compliance
- Scale horizontal
- Réactivité (notifications)

## Quand éviter

- Besoin de réponse synchrone
- Transactions ACID strictes
- Équipe junior
- Simple CRUD
