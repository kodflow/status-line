# Merge - Auto-merge avec CI validation

$ARGUMENTS

---

## Description

Commande intelligente pour merger une PR avec :
1. **Sync automatique** avec main (rebase)
2. **Attente CI** si en cours
3. **Auto-fix** si CI échoue (max 3 tentatives)
4. **Commentaire PR** détaillé si abandon
5. **Cleanup** automatique après merge

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) | Merge la PR de la branche courante |
| `--pr <number>` | Merge une PR spécifique |
| `--strategy <type>` | Méthode: merge/squash/rebase (défaut: squash) |
| `--no-delete` | Ne pas supprimer la branche après merge |
| `--dry-run` | Vérifier sans merger |
| `--help` | Afficher l'aide |

---

## --help

Quand `--help` est passé, afficher :

```
═══════════════════════════════════════════════
  /merge - Auto-merge avec CI validation
═══════════════════════════════════════════════

Usage: /merge [options]

Options:
  (vide)              Merge la PR de la branche courante
  --pr <number>       Merge une PR specifique
  --strategy <type>   Methode: merge/squash/rebase (defaut: squash)
  --no-delete         Garder la branche apres merge
  --dry-run           Verifier sans merger
  --help              Affiche cette aide

Workflow:
  1. Rebase sur main (sync)
  2. Attente CI si en cours
  3. Auto-fix si CI echoue (max 3x)
  4. Merge squash
  5. Cleanup branche

Exemples:
  /merge                    Merge la PR courante
  /merge --pr 42            Merge la PR #42
  /merge --strategy rebase  Force rebase merge
  /merge --dry-run          Test sans merger
═══════════════════════════════════════════════
```

---

## Priorité des outils

**IMPORTANT** : Toujours privilégier les outils MCP GitHub.

| Action | Priorité 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Lister PRs | `mcp__github__list_pull_requests` | `gh pr list` |
| Voir PR | `mcp__github__get_pull_request` | `gh pr view` |
| Status CI | `mcp__github__get_pull_request_status` | `gh pr checks` |
| Merge PR | `mcp__github__merge_pull_request` | `gh pr merge` |
| Commenter | `mcp__github__add_issue_comment` | `gh pr comment` |

**Extraction owner/repo** :
```bash
REMOTE=$(git remote get-url origin)
OWNER=$(echo "$REMOTE" | sed -E 's#.*[:/]([^/]+)/.*#\1#')
REPO=$(echo "$REMOTE" | sed -E 's#.*/([^/.]+)(\.git)?$#\1#')
```

---

## Workflow complet

### Étape 1 : Détection du contexte

```bash
# Provider Git
REMOTE=$(git remote get-url origin)
case "$REMOTE" in
    *github.com*)    PROVIDER="github" ;;
    *gitlab.com*)    PROVIDER="gitlab" ;;
    *bitbucket.org*) PROVIDER="bitbucket" ;;
    *)               PROVIDER="unknown" ;;
esac

# Owner/Repo
OWNER=$(echo "$REMOTE" | sed -E 's#.*[:/]([^/]+)/.*#\1#')
REPO=$(echo "$REMOTE" | sed -E 's#.*/([^/.]+)(\.git)?$#\1#')

# Branche courante
BRANCH=$(git branch --show-current)

# Main branch
MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's@^refs/remotes/origin/@@' || echo "main")
```

**Output** :
```
═══════════════════════════════════════════════
  /merge
═══════════════════════════════════════════════

  Provider : GitHub
  Repo     : owner/repo
  Branch   : feat/add-auth
  Main     : main

  → Recherche de la PR...
```

---

### Étape 2 : Trouver la PR

**MCP (prioritaire)** :
```
mcp__github__list_pull_requests:
  owner: OWNER
  repo: REPO
  head: "owner:BRANCH"
  state: "open"
```

**CLI (fallback)** :
```bash
gh pr list --head "$BRANCH" --state open --json number,title,url
```

**Si aucune PR trouvée** :
```
═══════════════════════════════════════════════
  ❌ Aucune PR trouvée
═══════════════════════════════════════════════

  Branche : feat/add-auth
  
  → Créez d'abord une PR avec /commit
  → Ou spécifiez --pr <number>

═══════════════════════════════════════════════
```

---

### Étape 3 : Validation des garde-fous

**Vérifications OBLIGATOIRES** :

```bash
# 1. Pas sur main/master
if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
    echo "❌ INTERDIT: Merge depuis main/master"
    exit 1
fi

# 2. PR existe et est ouverte
if [[ -z "$PR_NUMBER" ]]; then
    echo "❌ Aucune PR ouverte pour cette branche"
    exit 1
fi

# 3. Pas de conflits
MERGEABLE=$(gh pr view "$PR_NUMBER" --json mergeable -q '.mergeable')
if [[ "$MERGEABLE" == "CONFLICTING" ]]; then
    echo "❌ Conflits détectés - résolvez-les d'abord"
    exit 1
fi
```

---

### Étape 4 : Sync avec main (REBASE)

**Avant tout**, synchroniser la branche avec main :

```bash
echo "→ Synchronisation avec $MAIN_BRANCH..."

git fetch origin "$MAIN_BRANCH"

# Vérifier si rebase nécessaire
BEHIND=$(git rev-list --count HEAD.."origin/$MAIN_BRANCH")
if [[ "$BEHIND" -gt 0 ]]; then
    echo "  ⚠ Branche en retard de $BEHIND commits"
    echo "  → Rebase en cours..."
    
    git rebase "origin/$MAIN_BRANCH"
    
    # En cas de conflits
    if [[ $? -ne 0 ]]; then
        echo "❌ Conflits lors du rebase"
        echo "→ Résolvez les conflits puis relancez /merge"
        exit 1
    fi
    
    git push --force-with-lease
    echo "  ✓ Rebase terminé et pushé"
else
    echo "  ✓ Branche à jour avec $MAIN_BRANCH"
fi
```

---

### Étape 5 : Boucle CI avec auto-fix

**Configuration** :
```
MAX_FIX_ATTEMPTS = 3
CI_POLL_INTERVAL = 30 secondes
MAX_CI_WAIT = 20 polls (10 minutes)
```

**Boucle principale** :

```
fix_attempts = 0
ci_polls = 0

WHILE true:

    # Récupérer statut CI
    status = get_ci_status(PR_NUMBER)
    
    SWITCH status:
    
        CASE "success":
            → Sortir de la boucle
            → Procéder au merge
            
        CASE "pending":
            ci_polls++
            IF ci_polls > MAX_CI_WAIT:
                → Timeout, abandon
            ELSE:
                → Afficher "⏳ CI en cours..."
                → Attendre 30 secondes
                → Continue
                
        CASE "failure":
            fix_attempts++
            IF fix_attempts > MAX_FIX_ATTEMPTS:
                → Poster commentaire détaillé sur PR
                → Abandon
            ELSE:
                → Analyser l'erreur CI
                → Appliquer fix automatique
                → Commit + Push
                → Continue
```

**Récupération statut CI** :

```
# MCP (prioritaire)
mcp__github__get_pull_request_status:
  owner: OWNER
  repo: REPO
  pull_number: PR_NUMBER

# Retourne: state ("success", "pending", "failure")
# Et détails des checks individuels
```

**CLI (fallback)** :
```bash
gh pr checks "$PR_NUMBER" --json name,state,conclusion
```

---

### Étape 6 : Auto-fix (si CI échoue)

**Workflow de fix automatique** :

1. **Récupérer les logs d'erreur** :
```bash
# Identifier le job qui a échoué
gh run list --branch "$BRANCH" --limit 1 --json databaseId,conclusion
RUN_ID=$(...)

# Récupérer les logs
gh run view "$RUN_ID" --log-failed
```

2. **Analyser l'erreur** :
   - Parser les logs pour identifier le fichier/ligne
   - Comprendre le type d'erreur (test, lint, build, type)

3. **Appliquer le fix** :
   - Modifier le fichier concerné
   - Suivre les patterns de correction connus

4. **Commit et push** :
```bash
git add -A
git commit -m "fix(<scope>): <description courte>"
git push
```

**Output pendant le fix** :
```
═══════════════════════════════════════════════
  🔧 Auto-fix (tentative 1/3)
═══════════════════════════════════════════════

  Job échoué : test
  Erreur    : FAIL src/auth/login.test.ts:42
              Expected: true
              Received: false

  Analyse   : Assertion incorrecte après refactor

  Fix       : Mise à jour de l'assertion

  Commit    : fix(test): update login assertion
  Push      : ✓

  → Attente nouveau CI...

═══════════════════════════════════════════════
```

---

### Étape 7 : Commentaire PR (si abandon)

**Si 3 tentatives échouent**, poster un commentaire détaillé :

```
mcp__github__add_issue_comment:
  owner: OWNER
  repo: REPO
  issue_number: PR_NUMBER
  body: |
    ## ❌ Auto-merge abandonné après 3 tentatives
    
    ### Dernière erreur CI
    - **Job** : {job_name}
    - **Erreur** : `{error_message}`
    
    ### Tentatives de correction
    1. `{commit_1}` - CI still failing
    2. `{commit_2}` - CI still failing
    3. `{commit_3}` - CI still failing
    
    ### Analyse
    {detailed_analysis}
    
    ### Action requise
    - [ ] Examiner les logs CI
    - [ ] Corriger manuellement
    - [ ] Relancer `/merge` après correction
```

**Output** :
```
═══════════════════════════════════════════════
  ❌ Merge abandonné après 3 tentatives
═══════════════════════════════════════════════

  PR #42: feat/add-auth

  Dernier échec:
    Job   : test
    Error : Cannot resolve module 'xyz'

  ✓ Commentaire posté sur la PR
    → Détails des 3 tentatives
    → Analyse de l'erreur
    → Actions recommandées

  → Intervention manuelle requise

═══════════════════════════════════════════════
```

---

### Étape 8 : Merge (SQUASH)

**Une fois CI passé** :

```
# MCP (prioritaire)
mcp__github__merge_pull_request:
  owner: OWNER
  repo: REPO
  pull_number: PR_NUMBER
  merge_method: "squash"
  commit_title: "<PR title> (#PR_NUMBER)"
```

**CLI (fallback)** :
```bash
gh pr merge "$PR_NUMBER" --squash --delete-branch
```

---

### Étape 9 : Cleanup

**Après merge réussi** :

```bash
# Supprimer la branche remote (si pas --no-delete)
git push origin --delete "$BRANCH"

# Supprimer la branche locale
git branch -D "$BRANCH"

# Retour sur main
git checkout "$MAIN_BRANCH"
git pull origin "$MAIN_BRANCH"
```

---

## Outputs finaux

### Succès complet
```
═══════════════════════════════════════════════
  ✓ PR #42 merged successfully
═══════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  Rebase  : ✓ Synced (was 3 commits behind)
  CI      : ✓ Passed (2m 34s)
  Commits : 5 commits → 1 squashed

  Cleanup:
    ✓ Remote branch deleted
    ✓ Local branch deleted
    ✓ Switched to main
    ✓ Pulled latest (now at abc1234)

═══════════════════════════════════════════════
```

### Succès avec auto-fix
```
═══════════════════════════════════════════════
  ✓ PR #42 merged (after 1 auto-fix)
═══════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  CI      : ✓ Passed (after fix)
  
  Auto-fixes applied:
    1. fix(test): update login assertion

  Cleanup:
    ✓ Branch cleaned up
    ✓ On main (abc1234)

═══════════════════════════════════════════════
```

### Dry-run
```
═══════════════════════════════════════════════
  🔍 Dry-run: PR #42 ready to merge
═══════════════════════════════════════════════

  Branch  : feat/add-auth
  PR      : #42 - Add user authentication
  CI      : ✓ All checks passed
  
  Would execute:
    1. Merge with squash strategy
    2. Delete branch feat/add-auth
    3. Checkout main

  Run without --dry-run to proceed.

═══════════════════════════════════════════════
```

---

## GARDE-FOUS (ABSOLUS)

| Action | Status |
|--------|--------|
| Merge depuis main/master | ❌ **INTERDIT** |
| Merge sans PR | ❌ **INTERDIT** |
| Force merge si CI échoue x3 | ❌ **INTERDIT** |
| Push sans --force-with-lease | ❌ **INTERDIT** |
| Mentions IA dans commits | ❌ **INTERDIT** |
| Merge avec conflits | ❌ **INTERDIT** |

---

## Cas spéciaux

### --pr (numéro spécifique)

```bash
/merge --pr 42
```
Merge la PR #42 au lieu de chercher la PR de la branche courante.

### --strategy

```bash
/merge --strategy rebase
```
Force une stratégie de merge différente :
- `squash` (défaut) : Combine tous les commits
- `merge` : Crée un merge commit
- `rebase` : Applique les commits sur main

### --no-delete

```bash
/merge --no-delete
```
Garde la branche après le merge (utile pour référence).

### --dry-run

```bash
/merge --dry-run
```
Vérifie tout sans rien merger :
- Valide les garde-fous
- Vérifie le statut CI
- Affiche ce qui serait fait
