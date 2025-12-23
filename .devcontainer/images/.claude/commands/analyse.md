# Analyse - Critical Code & Information Analyzer

$ARGUMENTS

---

## Description

Analyse critique et objective de code, d'informations techniques, ou de concepts pour determiner leur validite, pertinence et veracite. Cette commande adopte une posture sceptique et scientifique pour distinguer le contenu pertinent du "bullshit" technique.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `<prompt>` | Analyse le contenu fourni |
| `--url <url>` | Analyse le contenu d'une URL |
| `--file <path>` | Analyse un fichier local |
| `--deep` | Analyse approfondie avec recherche web |
| `--help` | Affiche l'aide de la commande |

---

## --help

Quand `--help` est passe, afficher :

```
=============================================
  /analyse - Critical Code & Info Analyzer
=============================================

Usage: /analyse <prompt|code|info> [options]

Options:
  <prompt>          Analyse le contenu fourni directement
  --url <url>       Analyse le contenu d'une URL
  --file <path>     Analyse un fichier local
  --deep            Mode approfondi avec verification web
  --help            Affiche cette aide

Exemples:
  /analyse "ce code utilise la recursion de maniere optimale"
  /analyse --url https://example.com/article
  /analyse --file src/utils.ts
  /analyse --deep "les microservices sont toujours mieux"
=============================================
```

---

## Principes d'analyse

### Posture critique obligatoire

L'analyse DOIT etre :
- **Sceptique** : Ne jamais accepter une affirmation sans preuve
- **Objective** : Pas de biais de confirmation
- **Factuelle** : Basee sur des faits verifiables
- **Nuancee** : Reconnaitre les zones grises

### Categories d'evaluation

| Categorie | Description |
|-----------|-------------|
| BULLSHIT | Faux, trompeur, ou sans fondement technique |
| PARTIEL | Contient du vrai mais incomplet/biaise |
| CONTEXTUEL | Vrai dans certains contextes seulement |
| VALIDE | Techniquement correct et pertinent |

### Red flags a detecter

- **Affirmations absolues** : "toujours", "jamais", "le meilleur"
- **Appel a l'autorite** : "les experts disent", "tout le monde sait"
- **Buzzwords sans substance** : termes a la mode sans definition claire
- **Absence de nuance** : solutions miracles, comparaisons simplistes
- **Cargo cult** : patterns copies sans comprendre le contexte
- **Premature optimization** : complexite injustifiee
- **Not Invented Here** : rejet de solutions eprouvees sans raison
- **Resume fallacy** : "X utilise Y donc Y est bon"

---

## Workflow d'analyse

### Etape 1 : Reception du contenu

```bash
# Si URL fournie
if [[ -n "$URL" ]]; then
    # Utiliser WebFetch pour recuperer le contenu
    CONTENT=$(WebFetch "$URL")
fi

# Si fichier fourni
if [[ -n "$FILE" ]]; then
    # Utiliser Read pour lire le fichier
    CONTENT=$(Read "$FILE")
fi

# Sinon utiliser le prompt directement
```

### Etape 2 : Classification du contenu

Identifier le type de contenu :
- **Code source** : Analyse syntaxique et semantique
- **Affirmation technique** : Verification factuelle
- **Architecture/Design** : Evaluation des patterns
- **Performance claim** : Demande de benchmarks
- **Best practice** : Verification du contexte d'application

### Etape 3 : Analyse critique

Pour chaque element du contenu :

1. **Identification des claims**
   - Extraire chaque affirmation implicite ou explicite
   - Noter les assumptions sous-jacentes

2. **Verification**
   - Si `--deep` : WebSearch pour verifier les faits
   - Comparer avec les connaissances etablies
   - Chercher des contre-exemples

3. **Evaluation**
   - Attribuer une categorie (BULLSHIT/PARTIEL/CONTEXTUEL/VALIDE)
   - Justifier avec des arguments techniques

### Etape 4 : Synthese

Produire un rapport structure avec :
- Verdict global
- Points valides identifies
- Bullshit detecte avec explication
- Recommandations

---

## Format de sortie

### En-tete

```
=============================================
  Analyse Critique
=============================================

Type: [Code|Affirmation|Architecture|...]
Mode: [Standard|Approfondi]
Source: [Direct|URL|Fichier]
```

### Corps de l'analyse

Pour chaque claim analyse :

```
---------------------------------------------
Claim #N: "<citation exacte>"
---------------------------------------------

Verdict: [BULLSHIT|PARTIEL|CONTEXTUEL|VALIDE]

Analyse:
  - [Point d'analyse 1]
  - [Point d'analyse 2]
  - [...]

Evidence:
  - [Fait ou reference supportant l'analyse]
  - [...]

Contre-exemple (si applicable):
  - [Situation ou le claim est faux]

Nuance (si CONTEXTUEL):
  - Vrai quand: [contexte]
  - Faux quand: [contexte]
```

### Synthese finale

```
=============================================
  Synthese
=============================================

Verdict Global: [BULLSHIT|PARTIEL|CONTEXTUEL|VALIDE]

Score de fiabilite: X/10

Resume:
  [2-3 phrases resumant l'analyse]

Points Valides:
  - [Liste des elements corrects]

Bullshit Detecte:
  - [Liste des elements faux/trompeurs]

Recommandations:
  - [Actions suggerees]

=============================================
```

---

## Mode --deep

Quand `--deep` est active :

1. **Recherche web systematique**
   - WebSearch pour chaque claim majeur
   - Verification des sources citees
   - Recherche de contre-arguments

2. **Analyse comparative**
   - Comparer avec documentation officielle
   - Chercher des benchmarks independants
   - Verifier les dates (info obsolete?)

3. **Cross-reference**
   - Comparer plusieurs sources
   - Identifier les consensus et controverses

---

## Exemples d'analyse

### Exemple 1 : Affirmation sur les performances

**Input:**
```
/analyse "MongoDB est plus rapide que PostgreSQL"
```

**Output:**
```
=============================================
  Analyse Critique
=============================================

Type: Affirmation technique
Mode: Standard

---------------------------------------------
Claim #1: "MongoDB est plus rapide que PostgreSQL"
---------------------------------------------

Verdict: BULLSHIT

Analyse:
  - Comparaison sans contexte specifique
  - "Plus rapide" est non-quantifie
  - Ignore les types de workload
  - Pas de mention des conditions de test

Evidence:
  - Les performances dependent du use case
  - PostgreSQL excelle en requetes complexes/ACID
  - MongoDB excelle en lectures simples/documents
  - Benchmarks varies selon configuration

Contre-exemple:
  - Requetes JOIN complexes: PostgreSQL >> MongoDB
  - Transactions ACID: PostgreSQL >> MongoDB

=============================================
  Synthese
=============================================

Verdict Global: BULLSHIT

Score de fiabilite: 2/10

Resume:
  Affirmation trop simpliste qui ignore le contexte.
  Ni MongoDB ni PostgreSQL n'est "plus rapide" dans
  l'absolu - cela depend entierement du use case.

Bullshit Detecte:
  - Generalisation abusive
  - Absence de metriques
  - Comparaison de pommes et d'oranges

Recommandations:
  - Definir le workload specifique
  - Executer des benchmarks dans votre contexte
  - Considerer les besoins fonctionnels d'abord

=============================================
```

### Exemple 2 : Code suspect

**Input:**
```
/analyse --file src/auth.ts
```

**Output:**
```
=============================================
  Analyse Critique
=============================================

Type: Code source
Mode: Standard
Fichier: src/auth.ts

---------------------------------------------
Claim #1: Utilisation de MD5 pour les mots de passe
---------------------------------------------

Verdict: BULLSHIT

Analyse:
  - MD5 est cryptographiquement casse depuis 2004
  - Vulnerable aux rainbow tables
  - Pas de salting apparent

Evidence:
  - NIST deprecie MD5 pour usage cryptographique
  - bcrypt/argon2 sont les standards actuels

Recommandations:
  - Migrer vers bcrypt ou argon2id
  - Implementer le salting

---------------------------------------------
Claim #2: JWT stocke en localStorage
---------------------------------------------

Verdict: PARTIEL

Analyse:
  - Fonctionnel mais risque XSS
  - httpOnly cookie serait plus securise
  - Acceptable si CSP strict en place

=============================================
  Synthese
=============================================

Verdict Global: PARTIEL

Score de fiabilite: 4/10

Resume:
  Code fonctionnel mais avec des failles de securite
  significatives. Le hashing MD5 est un probleme critique.

Bullshit Detecte:
  - MD5 pour passwords = faille critique

Points a Ameliorer:
  - Storage JWT pourrait etre plus securise

=============================================
```

---

## Integration avec autres commandes

| Workflow | Usage |
|----------|-------|
| Review code | `/analyse --file` puis `/review` |
| Verification article | `/analyse --url --deep` |
| Validation architecture | `/analyse "<description>" --deep` |

---

## Notes importantes

- Cette commande est concue pour etre **constructive**, pas destructive
- L'objectif est d'aider a ameliorer, pas de denigrer
- Les verdicts doivent toujours etre **justifies**
- En cas de doute, utiliser `--deep` pour verification approfondie
- Reconnaitre quand l'information est **insuffisante** pour conclure
