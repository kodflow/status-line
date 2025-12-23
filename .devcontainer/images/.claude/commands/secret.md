# Secret - 1Password Secret Management

$ARGUMENTS

---

## Prerequis

### Verification du token OP

**OBLIGATOIRE** : Avant toute action, verifier si `OP_SERVICE_ACCOUNT_TOKEN` est
disponible.

```bash
# Verifier si le token existe dans l'environnement
echo "${OP_SERVICE_ACCOUNT_TOKEN:+set}"
```

**Si le token n'est PAS defini** :

1. Chercher dans le projet :

   ```bash
   cat /workspace/.devcontainer/.env 2>/dev/null | grep OP_SERVICE_ACCOUNT_TOKEN
   ```

2. **Si trouve** : Informer l'utilisateur de sourcer le fichier :

   ```text
   ## Token trouve

   Le token existe dans `.devcontainer/.env` mais n'est pas charge.
   Rechargez le container ou sourcez le fichier :

   source /workspace/.devcontainer/.env
   ```

3. **Si NON trouve** : **STOP IMMEDIAT** avec ce message :

   ```text
   ## Token 1Password manquant

   `OP_SERVICE_ACCOUNT_TOKEN` n'est pas configure.

   ### Configuration requise

   1. Creez un Service Account sur 1Password :
      https://my.1password.com/developer-tools/infrastructure-secrets/

   2. Ajoutez le token dans `.devcontainer/.env` :
      OP_SERVICE_ACCOUNT_TOKEN="ops_votre_token_ici"

   3. Rechargez le container DevContainer.

   **Action requise** : Configurez le token puis relancez `/secret`.
   ```

   **NE PAS CONTINUER** - Attendre que l'utilisateur configure le token.

---

## Configuration

```bash
VAULT_ID="ypahjj334ixtiyjkytu5hij2im"
```

---

## Arguments

| Pattern                 | Action                             |
| ----------------------- | ---------------------------------- |
| (vide) ou `list`        | Liste tous les secrets disponibles |
| `get <name>`            | Recupere la valeur d'un secret     |
| `add <name> [value]`    | Ajoute un nouveau secret           |
| `update <name> [value]` | Met a jour un secret existant      |
| `remove <name>`         | Supprime un secret                 |
| `--help`                | Affiche l'aide de la commande      |

---

## --help

Quand `--help` est passe, afficher :

```
═══════════════════════════════════════════════
  /secret - 1Password Secret Management
═══════════════════════════════════════════════

Usage: /secret [action] [name] [value]

Actions:
  (vide) ou list        Liste les secrets disponibles
  get <name>            Recupere un secret
  add <name> [value]    Ajoute un secret
  update <name> [value] Met a jour un secret
  remove <name>         Supprime un secret
  --help                Affiche cette aide

Prerequis:
  OP_SERVICE_ACCOUNT_TOKEN doit etre configure

Exemples:
  /secret                   Liste tous les secrets
  /secret get mcp-github    Recupere le token GitHub
  /secret add my-key        Ajoute un nouveau secret
  /secret remove old-key    Supprime un secret
═══════════════════════════════════════════════
```

---

## Actions

### /secret list (ou sans argument)

Liste tous les items du vault accessibles.

```bash
op item list --vault "$VAULT_ID" --format json | \
  jq -r '.[] | "- \(.title) (\(.category))"'
```

**Output** :

```text
## Secrets disponibles

| Nom        | Categorie      | Derniere modification |
|------------|----------------|----------------------|
| mcp-github | API Credential | 2024-01-15           |
| mcp-codacy | API Credential | 2024-01-15           |
| ...        | ...            | ...                  |
```

---

### /secret get (name)

Recupere et affiche la valeur d'un secret.

```bash
op item get "<name>" --vault "$VAULT_ID" --fields credential --reveal
```

**Output succes** :

```text
## Secret: <name>

**Valeur** : `<valeur_masquee_partiellement>`

Pour copier la valeur complete :
op item get "<name>" --vault "$VAULT_ID" --fields credential --reveal
```

**Output erreur** :

```text
## Erreur

Secret `<name>` non trouve dans le vault.

Secrets disponibles :
- mcp-github
- mcp-codacy
```

---

### /secret add (name) [value]

Cree un nouveau secret dans le vault.

**Si value fournie** :

```bash
op item create \
  --category "API Credential" \
  --title "<name>" \
  --vault "$VAULT_ID" \
  "credential=<value>"
```

**Si value NON fournie** : Demander a l'utilisateur :

```text
## Ajout de secret: <name>

Quelle est la valeur du secret `<name>` ?

(La valeur sera stockee de maniere securisee dans 1Password)
```

**Output succes** :

```text
## Secret cree

| Attribut  | Valeur         |
|-----------|----------------|
| Nom       | `<name>`       |
| Vault     | `$VAULT_ID`    |
| Categorie | API Credential |

Le secret est maintenant disponible via :
op item get "<name>" --vault "$VAULT_ID" --fields credential --reveal
```

---

### /secret update (name) [value]

Met a jour un secret existant.

**Verifier que le secret existe** :

```bash
op item get "<name>" --vault "$VAULT_ID" --format json
```

**Si existe et value fournie** :

```bash
op item edit "<name>" --vault "$VAULT_ID" "credential=<value>"
```

**Si existe et value NON fournie** : Demander :

```text
## Mise a jour: <name>

Quelle est la nouvelle valeur pour `<name>` ?
```

**Si n'existe PAS** :

```text
## Erreur

Secret `<name>` non trouve. Utilisez `/secret add <name>` pour le creer.
```

**Output succes** :

```text
## Secret mis a jour

`<name>` a ete mis a jour avec succes.
```

---

### /secret remove (name)

Supprime un secret du vault.

**ATTENTION** : Demander confirmation avant suppression.

```text
## Confirmation requise

Voulez-vous vraiment supprimer le secret `<name>` ?

Cette action est **irreversible**.

Repondez "oui" pour confirmer.
```

**Si confirme** :

```bash
op item delete "<name>" --vault "$VAULT_ID"
```

**Output succes** :

```text
## Secret supprime

`<name>` a ete supprime du vault.
```

---

## Exemples d'utilisation

```bash
# Lister tous les secrets
/secret
/secret list

# Recuperer un secret
/secret get mcp-github

# Ajouter un nouveau secret
/secret add my-api-key
/secret add my-api-key sk-1234567890

# Mettre a jour un secret
/secret update mcp-github
/secret update mcp-github ghp_newtoken123

# Supprimer un secret
/secret remove old-secret
```

---

## Securite

- Les valeurs ne sont JAMAIS loggees en clair dans l'historique
- Utilisez `--reveal` uniquement quand necessaire
- Les tokens 1Password Service Account ont des permissions limitees
- Le vault ID est specifique a ce projet

---

## Troubleshooting

### "not logged in"

Le token n'est pas charge. Verifiez `OP_SERVICE_ACCOUNT_TOKEN`.

### "vault not found"

Le vault ID est incorrect ou le Service Account n'y a pas acces.

### "item not found"

Le secret n'existe pas. Utilisez `/secret list` pour voir les disponibles.
