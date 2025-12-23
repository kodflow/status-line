# Serverless / FaaS

> Functions as a Service - Pay per execution

## Concept

Code exécuté à la demande, sans gestion de serveurs.

## Langages recommandés

| Langage | Platform | Adaptation |
|---------|----------|-----------|
| **Node.js** | AWS Lambda, Vercel | Excellent |
| **Python** | AWS Lambda, GCP | Excellent |
| **Go** | AWS Lambda, GCP | Très bon |
| **Rust** | AWS Lambda | Bon |
| **Java** | AWS Lambda | Moyen (cold start) |

## Structure

```
/src
├── functions/
│   ├── create-user/
│   │   ├── handler.ts
│   │   └── schema.json
│   ├── process-order/
│   │   ├── handler.ts
│   │   └── schema.json
│   └── send-email/
│       └── handler.ts
├── shared/
│   ├── utils/
│   └── types/
└── serverless.yml       # ou terraform/
```

## Avantages

- Zero ops (managed)
- Pay per use
- Scale automatique
- Déploiement simple
- Focus sur le code

## Inconvénients

- Cold starts
- Vendor lock-in
- Stateless obligatoire
- Coût imprévisible à scale
- Debug difficile
- Limites d'exécution

## Contraintes

- Stateless (pas d'état local)
- Timeout (15min max AWS)
- Memory limits
- Idempotency requise

## Règles

1. Une fonction = une responsabilité
2. Stateless toujours
3. Idempotent toujours
4. External state (DynamoDB, S3)
5. Fast cold start (bundle petit)

## Quand utiliser

- Trafic variable/imprévisible
- Tâches event-driven
- APIs légères
- Budget variable
- POC rapide

## Quand éviter

- Trafic constant élevé → EC2/K8s moins cher
- Besoins temps réel stricts
- Long-running tasks
- Complex state management
