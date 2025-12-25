# Review - Multi-Agent Code Review

$ARGUMENTS

---

## Description

Review de code multi-agents :
- **(vide)** : Review locale avec nos agents (branche/PR courante)
- **--coderabbit** : Déclenche une full review CodeRabbit sur la PR
- **--copilot** : Déclenche une full review GitHub Copilot sur la PR

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Review locale avec nos agents |
| `--coderabbit` | Full review CodeRabbit sur la PR GitHub |
| `--copilot` | Full review GitHub Copilot sur la PR GitHub |
| `--help` | Affiche l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /review - Multi-Agent Code Review
═══════════════════════════════════════════════

Usage: /review [agent]

Agents:
  (vide)          Review locale avec nos agents
  --coderabbit    Full review CodeRabbit sur la PR
  --copilot       Full review GitHub Copilot sur la PR

Exemples:
  /review                 Review locale de la branche
  /review --coderabbit    Demande review CodeRabbit
  /review --copilot       Demande review Copilot
═══════════════════════════════════════════════
```

---

## Action: (vide) - Review locale

Review de code avec nos agents sur la branche/PR courante.

### Workflow

1. **Détecter le contexte** :
   - Branche courante
   - Fichiers modifiés vs main
   - PR associée (si existe)

2. **Analyser les changements** :
   ```bash
   # Fichiers modifiés
   git diff --name-only origin/main...HEAD

   # Diff complet
   git diff origin/main...HEAD
   ```

3. **Review par catégorie** :

| Catégorie | Vérifications |
|-----------|---------------|
| **Sécurité** | Secrets, injections, XSS, auth |
| **Qualité** | Complexité, duplication, naming |
| **Performance** | N+1, memory leaks, caching |
| **Tests** | Couverture, edge cases |
| **Style** | Conventions, formatting |

4. **Générer le rapport** :

```
═══════════════════════════════════════════════
  Review locale
═══════════════════════════════════════════════

Branche : feat/add-auth
Fichiers : 5 modifiés
Base : origin/main

─────────────────────────────────────────────
  Sécurité
─────────────────────────────────────────────

⚠ src/auth.ts:42
  Token stocké en localStorage (risque XSS)
  → Préférer httpOnly cookie

─────────────────────────────────────────────
  Qualité
─────────────────────────────────────────────

✓ Pas de problème détecté

─────────────────────────────────────────────
  Performance
─────────────────────────────────────────────

⚠ src/api/users.ts:28
  Requête N+1 potentielle dans la boucle
  → Utiliser un batch query

─────────────────────────────────────────────
  Résumé
─────────────────────────────────────────────

| Catégorie   | Status |
|-------------|--------|
| Sécurité    | ⚠ 1 warning |
| Qualité     | ✓ OK |
| Performance | ⚠ 1 warning |
| Tests       | ✓ OK |
| Style       | ✓ OK |

Total : 2 warnings, 0 erreurs

═══════════════════════════════════════════════
```

---

## Action: --coderabbit

Déclenche une full review CodeRabbit sur la PR GitHub.

### Prérequis

- PR ouverte sur GitHub
- CodeRabbit configuré sur le repo (`.coderabbit.yaml`)

### Workflow

1. **Détecter la PR** :
   ```bash
   gh pr view --json number,url,title
   ```

   Si pas de PR :
   ```
   ❌ Aucune PR trouvée pour cette branche
   → Créez une PR avec /git --commit
   ```

2. **Poster le commentaire** :

   **Via MCP (prioritaire)** :
   ```
   mcp__github__add_issue_comment({
     owner: "<org>",
     repo: "<repo>",
     issue_number: <pr_number>,
     body: "@coderabbitai full review"
   })
   ```

   **Via gh CLI (fallback)** :
   ```bash
   gh pr comment <pr_number> --body "@coderabbitai full review"
   ```

3. **Confirmation** :

```
═══════════════════════════════════════════════
  ✓ CodeRabbit review demandée
═══════════════════════════════════════════════

PR : #<number> - <title>
URL : <pr_url>

→ CodeRabbit va analyser tous les fichiers
→ Les commentaires apparaîtront sur la PR
→ Utilisez /fix --pr pour corriger les retours

═══════════════════════════════════════════════
```

---

## Action: --copilot

Déclenche une full review GitHub Copilot sur la PR.

### Prérequis

- PR ouverte sur GitHub
- GitHub Copilot activé sur le repo

### Workflow

1. **Détecter la PR** :
   ```bash
   gh pr view --json number,url,title
   ```

   Si pas de PR :
   ```
   ❌ Aucune PR trouvée pour cette branche
   → Créez une PR avec /git --commit
   ```

2. **Poster le commentaire** :

   **Via MCP (prioritaire)** :
   ```
   mcp__github__add_issue_comment({
     owner: "<org>",
     repo: "<repo>",
     issue_number: <pr_number>,
     body: "@copilot review"
   })
   ```

   **Via gh CLI (fallback)** :
   ```bash
   gh pr comment <pr_number> --body "@copilot review"
   ```

3. **Confirmation** :

```
═══════════════════════════════════════════════
  ✓ Copilot review demandée
═══════════════════════════════════════════════

PR : #<number> - <title>
URL : <pr_url>

→ Copilot va analyser la PR
→ Les suggestions apparaîtront sur la PR

═══════════════════════════════════════════════
```

---

## Comparaison des agents

| Agent | Type | Quand utiliser |
|-------|------|----------------|
| Local | Instantané | Avant commit, feedback rapide |
| CodeRabbit | PR GitHub | Review détaillée, suggestions de fix |
| Copilot | PR GitHub | Review rapide, intégration GitHub |

---

## Workflow recommandé

```
1. Développer sur la branche
           ↓
2. /review              ← Review locale rapide
           ↓
3. /git --commit        ← Créer la PR
           ↓
4. /review --coderabbit ← Review détaillée
           ↓
5. /fix --pr            ← Corriger les retours
           ↓
6. /git --merge         ← Merger quand OK
```

---

## Voir aussi

- `/fix --pr` - Corriger les retours CodeRabbit un par un
- `/git --commit` - Créer une PR
- `/git --merge` - Merger la PR
