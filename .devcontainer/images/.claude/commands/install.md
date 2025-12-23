# Install - Development Tools Installer

$ARGUMENTS

---

## Description

Installe les outils de développement utilisés par les hooks Claude Code.
Les hooks fonctionnent même sans ces outils (silencieusement ignorés), mais avec eux tu bénéficies de :

- **Format automatique** à chaque édition
- **Tri des imports** automatique
- **Linting** avec auto-fix
- **Détection de secrets** avant commit
- **Type checking** en temps réel

---

## Arguments

| Pattern | Action |
|---------|--------|
| (vide) ou `all` | Installe TOUS les outils (complet) |
| `<lang>` | Installe les outils pour un langage spécifique |
| `security` | Installe uniquement les outils de sécurité |
| `list` | Liste les outils par catégorie |
| `--help` | Affiche l'aide de la commande |

**Langages supportés** : `js`, `ts`, `python`, `go`, `rust`, `shell`, `java`, `php`, `ruby`, `c`, `lua`, `sql`, `terraform`, `docker`, `elixir`, `dart`, `kotlin`, `swift`, `zig`, `nim`, `toml`, `protobuf`

---

## --help

Quand `--help` est passe, afficher :

```
═══════════════════════════════════════════════
  /install - Development Tools Installer
═══════════════════════════════════════════════

Usage: /install [lang|option]

Options:
  (vide) ou all     Installe tous les outils
  <lang>            Installe pour un langage (js, python, go...)
  security          Outils de securite uniquement
  list              Liste les outils disponibles
  --help            Affiche cette aide

Langages: js, ts, python, go, rust, shell, java, php,
          ruby, c, lua, sql, terraform, docker, elixir,
          dart, kotlin, swift, zig, nim, toml, protobuf

Exemples:
  /install              Installe tout
  /install python       Outils Python (ruff, black, mypy...)
  /install security     Outils securite (trivy, gitleaks...)
  /install list         Liste tous les outils
═══════════════════════════════════════════════
```

---

## Actions

### /install list

Affiche les outils organisés par catégorie :

```text
## Outils par catégorie

### 🔒 Sécurité (security)
- detect-secrets : Détection de secrets dans le code
- trivy : Scanner de vulnérabilités
- gitleaks : Détection de fuites de credentials

### 📝 JavaScript/TypeScript (js, ts)
- prettier : Formatage (JS/TS/JSON/YAML/MD/HTML/CSS)
- eslint : Linting avec auto-fix
- tsc : Type checking TypeScript

### 🐍 Python (python)
- ruff : Formatage + Linting ultra-rapide
- black : Formatage (alternatif à ruff)
- isort : Tri des imports
- mypy : Type checking
- pyright : Type checking (alternatif)
- pytest : Tests

### 🐹 Go (go)
- goimports : Formatage + tri imports
- golangci-lint : Linting complet
- staticcheck : Analyse statique

### 🦀 Rust (rust)
- rustfmt : Formatage
- clippy : Linting

### 🐚 Shell (shell)
- shfmt : Formatage
- shellcheck : Linting

### 🐳 Docker (docker)
- hadolint : Linting Dockerfile

### ☕ Java (java)
- google-java-format : Formatage
- checkstyle : Linting

### 🔷 C/C++ (c)
- clang-format : Formatage
- clang-tidy : Linting
- cppcheck : Analyse statique

### 🌍 Terraform (terraform)
- tflint : Linting
- terraform : CLI (fmt/validate)

### 💎 Ruby (ruby)
- rubocop : Formatage + Linting

### 🐘 PHP (php)
- php-cs-fixer : Formatage
- phpstan : Analyse statique

### 📄 Autres
- yamlfmt / yamllint : YAML
- markdownlint : Markdown
- jsonlint : JSON
- stylelint : CSS/SCSS
- taplo : TOML
- buf : Protobuf
- sqlfluff : SQL
```

---

### /install (ou /install all)

Installe TOUS les outils essentiels. Exécuter dans l'ordre :

**1. Sécurité (prioritaire)** :
```bash
pip install --user detect-secrets gitleaks-py
# trivy via script officiel
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /home/vscode/.local/bin
```

**2. JavaScript/TypeScript** :
```bash
npm install -g prettier eslint typescript
```

**3. Python** :
```bash
pip install --user ruff black isort mypy pyright pytest
```

**4. Go** :
```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

**5. Rust** (si cargo disponible) :
```bash
rustup component add rustfmt clippy
```

**6. Shell** :
```bash
go install mvdan.cc/sh/v3/cmd/shfmt@latest
# shellcheck via apt ou brew
```

**7. Autres** :
```bash
npm install -g markdownlint-cli jsonlint stylelint yaml-lint
pip install --user yamllint sqlfluff
go install github.com/tamasfe/taplo-cli/cmd/taplo@latest
```

**Output** :
```text
## Installation complète

Installation des outils pour les hooks Claude Code...

✅ Sécurité : detect-secrets, trivy, gitleaks
✅ JavaScript/TypeScript : prettier, eslint, tsc
✅ Python : ruff, black, isort, mypy, pytest
✅ Go : goimports, golangci-lint, staticcheck
✅ Rust : rustfmt, clippy
✅ Shell : shfmt, shellcheck
✅ Autres : yamllint, markdownlint, jsonlint

## Vérification

Les hooks sont maintenant actifs. Test avec :
claude --print-hooks
```

---

### /install security

Installe uniquement les outils de sécurité :

```bash
# detect-secrets - Détection de patterns secrets
pip install --user detect-secrets

# trivy - Scanner complet (secrets, vulns, misconfig)
curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /home/vscode/.local/bin

# gitleaks - Détection de credentials dans git
go install github.com/gitleaks/gitleaks/v8@latest
```

**Output** :
```text
## Outils de sécurité installés

✅ detect-secrets : Patterns de secrets (API keys, passwords)
✅ trivy : Vulnérabilités + secrets + misconfig
✅ gitleaks : Fuites de credentials dans l'historique git

Les hooks de sécurité sont maintenant actifs sur chaque édition.
```

---

### /install js (ou ts)

```bash
npm install -g prettier eslint typescript @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

---

### /install python

```bash
pip install --user ruff black isort mypy pyright pytest autopep8
```

---

### /install go

```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
```

---

### /install rust

```bash
rustup component add rustfmt clippy
```

---

### /install shell

```bash
go install mvdan.cc/sh/v3/cmd/shfmt@latest
# shellcheck - selon le système
apt-get install -y shellcheck 2>/dev/null || brew install shellcheck 2>/dev/null || true
```

---

### /install docker

```bash
# hadolint
wget -qO /home/vscode/.local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
chmod +x /home/vscode/.local/bin/hadolint
```

---

### /install terraform

```bash
# tflint
curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
```

---

### /install java

```bash
# google-java-format (nécessite Java)
wget -qO /home/vscode/.local/bin/google-java-format.jar https://github.com/google/google-java-format/releases/latest/download/google-java-format-all-deps.jar
echo '#!/bin/bash\njava -jar /home/vscode/.local/bin/google-java-format.jar "$@"' > /home/vscode/.local/bin/google-java-format
chmod +x /home/vscode/.local/bin/google-java-format
```

---

### /install c

```bash
apt-get install -y clang-format clang-tidy cppcheck
```

---

### /install ruby

```bash
gem install rubocop
```

---

### /install php

```bash
composer global require friendsofphp/php-cs-fixer phpstan/phpstan
```

---

### /install lua

```bash
luarocks install --local luacheck
cargo install stylua
```

---

### /install sql

```bash
pip install --user sqlfluff
# pg_format via apt si PostgreSQL
apt-get install -y pgformatter 2>/dev/null || true
```

---

### /install toml

```bash
cargo install taplo-cli
```

---

### /install protobuf

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

---

### /install elixir

```bash
mix local.hex --force
mix archive.install hex credo --force
```

---

### /install dart

```bash
# dart est inclus avec Flutter
# Sinon : apt-get install dart
dart pub global activate dart_style
```

---

### /install kotlin

```bash
# ktlint
curl -sSLO https://github.com/pinterest/ktlint/releases/latest/download/ktlint
chmod +x ktlint
mv ktlint /home/vscode/.local/bin/
```

---

## Vérification post-installation

Après installation, vérifier que les outils sont disponibles :

```bash
# Vérifier un outil spécifique
which prettier ruff goimports

# Tester les hooks Claude
claude --print-hooks
```

---

## Notes importantes

1. **Tous les outils sont OPTIONNELS** - Les hooks ignorent silencieusement les outils manquants
2. **Priorité recommandée** : security → langage principal → autres
3. **PATH** : Les outils sont installés dans `~/.local/bin` (déjà dans PATH via postCreate.sh)
4. **Mise à jour** : Réexécuter `/install <lang>` pour mettre à jour

---

## Troubleshooting

### "command not found" après installation

```bash
# Recharger le PATH
source ~/.kodflow-env.sh
# ou
export PATH="$HOME/.local/bin:$PATH"
```

### npm/pip permission denied

```bash
# Utiliser --user pour pip
pip install --user <package>

# Ou configurer npm pour global sans sudo
npm config set prefix '~/.local/share/npm-global'
```

### go install échoue

```bash
# Vérifier GOPATH
echo $GOPATH  # Doit être /home/vscode/.cache/go
# Recharger env
source ~/.kodflow-env.sh
```
