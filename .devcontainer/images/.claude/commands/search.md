# Search - Documentation Research

$ARGUMENTS

---

## Description

Recherche d'informations sur les documentations officielles UNIQUEMENT :

- Croise minimum 2 sources pour valider chaque information
- Questionne l'utilisateur pour affiner la recherche
- Génère un fichier `.context.md` (non commité) utilisable par `/plan` et `/apply`

**Principe** : Fiabilité > Quantité. Mieux vaut peu d'infos confirmées que beaucoup d'infos douteuses.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<query>` | Nouvelle recherche sur le sujet |
| `--append` | Ajoute au contexte existant au lieu de le remplacer |
| `--status` | Affiche le contexte actuel |
| `--clear` | Supprime le fichier .context.md |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /search - Documentation Research
═══════════════════════════════════════════════

Usage: /search <query> [options]

Options:
  <query>           Sujet de recherche
  --append          Ajoute au contexte existant
  --status          Affiche le contexte actuel
  --clear           Supprime .context.md
  --help            Affiche cette aide

Comportement:
  - Sources officielles uniquement
  - Croisement obligatoire (min 2 sources)
  - Questions pour affiner la recherche

Exemples:
  /search OAuth2 avec JWT
  /search Go generics --append
  /search --status

Workflow:
  /search <query> → itérer → /plan → /apply
═══════════════════════════════════════════════
```

---

## Sources officielles (Whitelist)

**RÈGLE ABSOLUE** : Utiliser UNIQUEMENT les domaines suivants pour WebSearch.

### Langages

| Langage | Domaines autorisés |
|---------|-------------------|
| Node.js | nodejs.org, developer.mozilla.org |
| Python | docs.python.org, python.org |
| Go | go.dev, golang.org, pkg.go.dev |
| Rust | rust-lang.org, doc.rust-lang.org |
| Java | docs.oracle.com, openjdk.org |
| C/C++ | cppreference.com, isocpp.org |
| PHP | php.net |
| Ruby | ruby-lang.org, ruby-doc.org |

### Cloud & Infrastructure

| Service | Domaines autorisés |
|---------|-------------------|
| AWS | docs.aws.amazon.com |
| GCP | cloud.google.com |
| Azure | learn.microsoft.com, docs.microsoft.com |
| Docker | docs.docker.com |
| Kubernetes | kubernetes.io |
| Terraform | developer.hashicorp.com |

### Bases de données

| DB | Domaines autorisés |
|----|-------------------|
| PostgreSQL | postgresql.org |
| MySQL | dev.mysql.com |
| MongoDB | mongodb.com/docs |
| Redis | redis.io |

### Frameworks

| Framework | Domaines autorisés |
|-----------|-------------------|
| React | react.dev, reactjs.org |
| Vue | vuejs.org |
| Angular | angular.io |
| Next.js | nextjs.org |
| Django | docs.djangoproject.com |
| Flask | flask.palletsprojects.com |
| Spring | spring.io |
| FastAPI | fastapi.tiangolo.com |

### Généralistes fiables

| Type | Domaines autorisés |
|------|-------------------|
| Web APIs | developer.mozilla.org |
| Standards | w3.org, whatwg.org |
| Security | owasp.org |
| RFCs | rfc-editor.org, tools.ietf.org |

### Blacklist implicite

- ❌ Blogs personnels
- ❌ Medium, Dev.to (sauf domaines officiels)
- ❌ Stack Overflow (OK pour identifier problèmes, PAS pour solutions)
- ❌ Tutoriels tiers
- ❌ Sites de cours (Udemy, Coursera...)
- ❌ ChatGPT/AI-generated content

---

## Workflow de recherche (5 phases)

### Phase 1 : Analyse de la query

1. Identifier les technologies mentionnées
2. Détecter les concepts clés
3. Lister les sources officielles à cibler

**Output Phase 1 :**
```
═══════════════════════════════════════════════
  /search <query>
═══════════════════════════════════════════════

  Technologies détectées :
    • <tech1> → <domaine officiel>
    • <tech2> → <domaine officiel>

  Concepts clés :
    • <concept1>
    • <concept2>

  Recherche en cours...

═══════════════════════════════════════════════
```

---

### Phase 2 : Recherche documentations officielles

Utiliser WebSearch avec `allowed_domains` :

```
WebSearch({
  query: "<query optimisée>",
  allowed_domains: ["<domain1>", "<domain2>", ...]
})
```

Puis WebFetch pour extraire le contenu pertinent :

```
WebFetch({
  url: "<url doc officielle>",
  prompt: "Extraire les informations sur <sujet>"
})
```

**IMPORTANT** : Ne jamais utiliser de source non-officielle, même si elle semble pertinente.

---

### Phase 3 : Croisement des sources

**Règle** : Chaque affirmation doit être confirmée par minimum 2 sources officielles.

| Situation | Action |
|-----------|--------|
| 2+ sources confirment | ✓ Inclure avec confidence: HIGH |
| 1 source officielle | ⚠ Inclure avec confidence: MEDIUM |
| Sources contradictoires | 🔄 Approfondir ou signaler |
| 0 source officielle | ❌ Ne pas inclure |

**Détection des contradictions :**

- Comparer les versions (doc ancienne vs récente)
- Vérifier les dates de mise à jour
- Signaler les incohérences à l'utilisateur

---

### Phase 4 : Questions de clarification

**OBLIGATOIRE** : Utiliser AskUserQuestion pour affiner la recherche.

Questions typiques :

- Version spécifique à cibler ?
- Cas d'usage précis ?
- Contraintes techniques ?
- Priorités (performance vs simplicité) ?
- Environnement cible (dev/prod) ?

```
AskUserQuestion: {
  questions: [
    {
      question: "Quelle version de <tech> ciblez-vous ?",
      header: "Version",
      options: [
        { label: "Dernière stable", description: "Recommandée" },
        { label: "LTS", description: "Support long terme" },
        { label: "Spécifique", description: "Je précise" }
      ]
    }
  ]
}
```

**Itération** : Si les réponses révèlent de nouveaux besoins → retour Phase 2.

---

### Phase 5 : Génération context.md

Créer `/workspace/.context.md` avec le format suivant :

```markdown
# Context: <sujet>

Generated: <ISO8601>
Query: <recherche initiale>
Iterations: <nombre>

## Summary

<Résumé en 2-3 phrases des informations clés>

## Key Information

### <Sous-thème 1>

<Information validée>

**Sources:**
- [<Titre doc>](<url>) - "<extrait pertinent>"
- [<Titre doc 2>](<url>) - "<confirmation>"

**Confidence:** HIGH

### <Sous-thème 2>

<Information avec une seule source>

**Sources:**
- [<Titre doc>](<url>) - "<extrait>"

**Confidence:** MEDIUM

**Note:** Information non confirmée par une seconde source.

## Clarifications

<Questions posées et réponses utilisateur>

| Question | Réponse |
|----------|---------|
| Version ciblée ? | 3.x LTS |
| Environnement ? | Production |

## Recommendations

<Suggestions basées sur la recherche croisée>

1. <Recommandation 1>
2. <Recommandation 2>

## Warnings

<Points d'attention identifiés>

- ⚠ <Warning 1>
- ⚠ <Warning 2>

## Sources Summary

| Source | Domain | Confidence | Sections |
|--------|--------|------------|----------|
| <titre> | <domain> | HIGH | §1, §2 |
| <titre> | <domain> | MEDIUM | §2 |

---

_Ce fichier est généré automatiquement par `/search`. Ne pas commiter._
```

---

## --append

Quand `--append` est passé :

1. **Lire** le fichier `.context.md` existant
2. **Ajouter** les nouvelles informations (pas de duplicata)
3. **Mettre à jour** le timestamp et le compteur d'itérations
4. **Fusionner** les sources

**Output --append :**
```
═══════════════════════════════════════════════
  /search --append <query>
═══════════════════════════════════════════════

  Context existant : .context.md
  Sujet actuel     : <sujet existant>
  Iterations       : 2 → 3

  Ajout de nouvelles informations...

─────────────────────────────────────────────

  + 2 nouvelles sections
  ~ 1 section enrichie
  = 3 sections inchangées

  ✓ Context mis à jour

═══════════════════════════════════════════════
```

---

## --status

Afficher un résumé du contexte actuel :

```
═══════════════════════════════════════════════
  Context actuel
═══════════════════════════════════════════════

  Fichier     : .context.md
  Sujet       : <sujet>
  Généré      : <date>
  Iterations  : <n>

─────────────────────────────────────────────
  Sections
─────────────────────────────────────────────

  1. <Section 1> [HIGH]
  2. <Section 2> [MEDIUM]
  3. <Section 3> [HIGH]

─────────────────────────────────────────────
  Sources
─────────────────────────────────────────────

  • nodejs.org (3 références)
  • developer.mozilla.org (2 références)

─────────────────────────────────────────────
  Statistiques
─────────────────────────────────────────────

  Sections      : 3
  Sources       : 5
  Confidence    : 80% HIGH, 20% MEDIUM

═══════════════════════════════════════════════
```

---

## --clear

Supprimer le fichier `.context.md` :

```bash
rm -f /workspace/.context.md
```

**Output --clear :**
```
═══════════════════════════════════════════════
  ✓ Context supprimé
═══════════════════════════════════════════════

  Fichier supprimé : .context.md

═══════════════════════════════════════════════
```

---

## Intégration avec autres commandes

| Commande | Utilisation du context |
|----------|------------------------|
| `/plan` | Lit `.context.md` en Phase 2 pour les informations techniques |
| `/apply` | Référence les URLs pour installer les dépendances |
| `/fix` | Utilise le context pour rechercher des solutions |

**Détection automatique :**

- Si `.context.md` existe, les commandes l'utilisent automatiquement
- Affichage d'un message : "Context chargé : `<sujet>`"

---

## GARDE-FOUS (ABSOLUS)

| Action | Status |
|--------|--------|
| Utiliser source non-officielle | ❌ **INTERDIT** |
| Inclure info non-vérifiée (0 source) | ❌ **INTERDIT** |
| Skip Phase 4 (questions) | ❌ **INTERDIT** |
| Générer context sans croisement | ❌ **INTERDIT** |

---

## Voir aussi

- `/plan` - Planifier une implémentation
- `/apply` - Exécuter le plan
- `/update --context` - Mettre à jour le contexte projet
