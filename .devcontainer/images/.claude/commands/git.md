---
name: git
description: |
  Workflow Git Automation with RLM decomposition.
  Handles branch management, conventional commits, and CI validation.
  Use when: committing changes, creating PRs/MRs, or merging with CI checks.
  Supports GitHub (PRs) and GitLab (MRs) - auto-detected from git remote.
allowed-tools:
  - "Bash(git:*)"
  - "Bash(gh:*)"
  - "Bash(glab:*)"
  - "mcp__github__*"
  - "mcp__gitlab__*"
  - "Read(**/*)"
  - "Write(.env)"
  - "Edit(.env)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Task(*)"
  - "TaskCreate(*)"
  - "TaskUpdate(*)"
  - "TaskList(*)"
  - "TaskGet(*)"
  - "AskUserQuestion(*)"
---

# /git - Workflow Git Automation (RLM Architecture)

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

$ARGUMENTS

---

## Overview

Git automation with **RLM** patterns:

- **Identity** - Verify/configure git identity via `.env`
- **Peek** - Analyze git state before action
- **Decompose** - Identify files by category
- **Parallelize** - Parallel checks (lint, test, CI)
- **Context** - `/warmup --update` on modified files (branch diff)
- **Synthesize** - Consolidated report

**Note:** The git identity (user.name/user.email) is stored in `/workspace/.env` and automatically synchronized with git config on each execution.

---

## Arguments

| Pattern | Action |
|---------|--------|
| `--commit` | Full workflow: branch, commit, push, PR/MR |
| `--merge` | Merge the PR/MR with CI validation |
| `--help` | Display help |

### Options --commit

| Option | Action |
|--------|--------|
| `--branch <name>` | Force the branch name |
| `--no-pr` | Skip PR/MR creation |
| `--amend` | Amend the last commit |
| `--skip-identity` | Skip git identity verification |

### Options --merge

| Option | Action |
|--------|--------|
| `--pr <number>` | Merge a specific PR (GitHub) |
| `--mr <number>` | Merge a specific MR (GitLab) |
| `--strategy <type>` | Method: merge/squash/rebase (default: squash) |
| `--dry-run` | Verify without merging |

---

## --help

```
═══════════════════════════════════════════════════════════════
  /git - Workflow Git Automation (RLM)
═══════════════════════════════════════════════════════════════

Usage: /git <action> [options]

Actions:
  --commit          Full workflow (branch, commit, push, PR/MR)
  --merge           Merge with CI validation and auto-fix

RLM Patterns:
  0.5. Identity    - Verify/configure git user via .env
  1. Peek          - Analyze git state
  2. Decompose     - Categorize files
  3. Parallelize   - Simultaneous checks
  3.8. Context     - /warmup --update (branch diff, 5min staleness)
  4. Synthesize    - Consolidated report

Options --commit:
  --branch <name>   Force the branch name
  --no-pr           Skip PR/MR creation
  --amend           Amend the last commit
  --skip-identity   Skip identity verification

Options --merge:
  --pr <number>     Merge a specific PR (GitHub)
  --mr <number>     Merge a specific MR (GitLab)
  --strategy <type> Method: merge/squash/rebase (default: squash)
  --dry-run         Verify without merging

Identity (.env):
  - GIT_USER and GIT_EMAIL stored in /workspace/.env
  - Automatically synchronized with git config
  - Prompted to user if missing

Examples:
  /git --commit                 Automatic commit + PR
  /git --commit --no-pr         Commit without creating PR
  /git --commit --skip-identity Skip identity verification
  /git --merge                  Merge current PR/MR
  /git --merge --pr 42          Merge PR #42

═══════════════════════════════════════════════════════════════
```

---

## MCP vs CLI Priority

**IMPORTANT**: Always prefer MCP tools when available.

**Platform auto-detected:** `git remote get-url origin` → github.com | gitlab.*

### GitHub (PRs)

| Action | Priority 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Create branch | `mcp__github__create_branch` | `git checkout -b` |
| Create PR | `mcp__github__create_pull_request` | `gh pr create` |
| List PRs | `mcp__github__list_pull_requests` | `gh pr list` |
| View PR | `mcp__github__get_pull_request` | `gh pr view` |
| CI Status | `mcp__github__get_pull_request_status` | `gh pr checks` |
| Merge PR | `mcp__github__merge_pull_request` | `gh pr merge` |

### GitLab (MRs)

| Action | Priority 1 (MCP) | Fallback (CLI) |
|--------|------------------|----------------|
| Create branch | `git checkout -b` + push | `git checkout -b` |
| Create MR | `mcp__gitlab__create_merge_request` | `glab mr create` |
| List MRs | `mcp__gitlab__list_merge_requests` | `glab mr list` |
| View MR | `mcp__gitlab__get_merge_request` | `glab mr view` |
| CI Status | `mcp__gitlab__list_pipelines` | `glab ci status` |
| Merge MR | `mcp__gitlab__merge_merge_request` | `glab mr merge` |

---

## Action: --commit

### Phase 1.0: Git Identity Validation (MANDATORY)

**Verify and configure git identity BEFORE any action:**

```yaml
identity_validation:
  env_file: "/workspace/.env"

  1_check_env:
    action: "Check if .env exists and contains GIT_USER/GIT_EMAIL"
    tool: Read("/workspace/.env")
    fallback: "File not found → create"

  2_extract_or_ask:
    rule: |
      IF .env exists AND contains GIT_USER AND GIT_EMAIL:
        user = extract(GIT_USER)
        email = extract(GIT_EMAIL)
      ELSE:
        → AskUserQuestion (see below)
        → Create/Update .env

  3_verify_git_config:
    action: "Compare with current git config"
    commands:
      - "git config user.name"
      - "git config user.email"
    decision:
      if_match: "→ Continue to Phase 1"
      if_mismatch: "→ Fix git config"

  4_fix_if_needed:
    action: "Apply the correct configuration"
    commands:
      - "git config user.name '{user}'"
      - "git config user.email '{email}'"

  5_check_gpg:
    action: "Check if GPG signing is configured"
    commands:
      - "git config --get commit.gpgsign"
      - "git config --get user.signingkey"

  6_configure_gpg_if_missing:
    condition: "commit.gpgsign != true OR user.signingkey is empty"
    action: "List GPG keys and prompt for selection if needed"
    workflow:
      1_list_keys: "gpg --list-secret-keys --keyid-format LONG"
      2_find_matching:
        rule: "Find key matching GIT_EMAIL"
        action: "grep -B1 '{email}' in gpg output"
      3_if_no_match_but_keys_exist:
        tool: AskUserQuestion
        questions:
          - question: "Which GPG key to use for signing commits?"
            header: "GPG Key"
            options: "<dynamically generated from gpg output>"
      4_configure:
        commands:
          - "git config --global user.signingkey {selected_key}"
          - "git config --global commit.gpgsign true"
          - "git config --global tag.forceSignAnnotated true"
```

**Prompt if .env is missing or incomplete:**

```yaml
ask_identity:
  tool: AskUserQuestion
  questions:
    - question: "What name to use for git commits?"
      header: "Git User"
      options:
        - label: "{detected_user}"
          description: "Detected from global git config"
        - label: "{github_user}"
          description: "Detected from GitHub/GitLab"
      # User can also enter "Other" with custom value

    - question: "What email address to use for commits?"
      header: "Git Email"
      options:
        - label: "{detected_email}"
          description: "Detected from global git config"
        - label: "{noreply_email}"
          description: "GitHub/GitLab noreply email"
```

**Generated/updated .env format:**

```bash
# Git identity for commits (managed by /git)
GIT_USER="John Doe"
GIT_EMAIL="john.doe@example.com"
```

**Output Phase 0.5:**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity & GPG Validation
═══════════════════════════════════════════════════════════════

  .env check:
    ├─ File: /workspace/.env
    ├─ GIT_USER: "John Doe" ✓
    └─ GIT_EMAIL: "john.doe@example.com" ✓

  Git config:
    ├─ user.name: "John Doe" ✓ (match)
    └─ user.email: "john.doe@example.com" ✓ (match)

  GPG config:
    ├─ commit.gpgsign: true ✓
    └─ user.signingkey: ABCD1234EF567890 ✓

  Status: ✓ Identity & GPG validated, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

**Output if correction needed:**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity Validation
═══════════════════════════════════════════════════════════════

  .env check:
    ├─ File: /workspace/.env
    ├─ GIT_USER: "John Doe" ✓
    └─ GIT_EMAIL: "john.doe@example.com" ✓

  Git config:
    ├─ user.name: "johndoe" ✗ (mismatch)
    └─ user.email: "old@email.com" ✗ (mismatch)

  Action: Correcting git config...
    ├─ git config user.name "John Doe"
    └─ git config user.email "john.doe@example.com"

  Status: ✓ Identity corrected, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

**Output if .env missing:**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Git Identity Validation
═══════════════════════════════════════════════════════════════

  .env check:
    └─ File: NOT FOUND → Creating...

  User input required...
    ├─ Git User: "John Doe" (entered)
    └─ Git Email: "john.doe@example.com" (entered)

  Actions:
    ├─ Created /workspace/.env with GIT_USER, GIT_EMAIL
    ├─ git config user.name "John Doe"
    └─ git config user.email "john.doe@example.com"

  Status: ✓ Identity configured, proceeding to Phase 1

═══════════════════════════════════════════════════════════════
```

---

### Phase 2.0: Peek (RLM Pattern)

**Analyze git state BEFORE any action:**

```yaml
peek_workflow:
  1_status:
    action: "Check repo state (ALL modifications, not just current task)"
    commands:
      - "git status --porcelain"
      - "git branch --show-current"
      - "git log -1 --format='%h %s'"
    critical_rule: |
      LIST ALL modified files — including CLAUDE.md, .devcontainer/,
      .claude/commands/. NEVER ignore tracked modified files.
      git status --porcelain shows EVERYTHING that is tracked and modified.
      Gitignored files DO NOT APPEAR → no risk of including them.

  2_changes:
    action: "Analyze changes"
    tools: [Bash(git diff --stat)]

  3_branch_check:
    action: "Check current branch"
    decision:
      - "main/master → MUST create new branch"
      - "feat/* | fix/* → Check coherence"
```

**Output Phase 1:**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Peek Analysis
═══════════════════════════════════════════════════════════════

  Branch: main (protected)
  Status: 5 files modified, 2 untracked

  Changes detected:
    ├─ src/auth/login.ts (+45, -12)
    ├─ src/auth/logout.ts (+23, -5)
    ├─ tests/auth.test.ts (+80, -0) [new]
    ├─ package.json (+2, -1)
    └─ README.md (+15, -3)

  Decision: CREATE new branch (on protected main)

═══════════════════════════════════════════════════════════════
```

---

### Phase 3.0: Decompose (RLM Pattern)

**Categorize modified files:**

```yaml
decompose_workflow:
  categories:
    features:
      patterns: ["src/**/*.ts", "src/**/*.js", "src/**/*.go", "src/**/*.rs", "src/**/*.py"]
      prefix: "feat"

    fixes:
      patterns: ["*fix*", "*bug*"]
      prefix: "fix"

    tests:
      patterns: ["tests/**", "**/*.test.*", "**/*_test.go"]
      prefix: "test"

    docs:
      patterns: ["*.md", "docs/**", "**/CLAUDE.md", ".claude/commands/*.md"]
      prefix: "docs"

    config:
      patterns: ["*.json", "*.yaml", "*.toml", ".devcontainer/**"]
      prefix: "chore"

    hooks:
      patterns: [".devcontainer/hooks/**", ".claude/scripts/**", ".githooks/**"]
      prefix: "fix"

  auto_detect:
    action: "Infer the dominant type"
    output: "commit_type, scope, branch_name"

  gitignore_awareness:
    rule: |
      BEFORE categorizing, check the gitignore status of each file.
      Use `git status --porcelain` to list ALL modified files.
      Gitignored files do not appear in git status → no risk.
      Tracked modified files MUST be included, even if they are in .claude/ or CLAUDE.md.
    check: |
      # List ALL modifications (staged + unstaged + untracked non-ignored)
      git status --porcelain
      # Verify that nothing tracked is forgotten after staging
      git diff --name-only  # Must be empty after git add -A
```

---

### Phase 4.0: Parallelize (RLM Pattern) - Multi-Language Pre-commit

**Auto-detect ALL project languages and run checks for each:**

```yaml
language_detection:
  script: ".claude/scripts/pre-commit-checks.sh"

  detection_files:
    go.mod: "Go"
    Cargo.toml: "Rust"
    package.json: "Node.js"
    pyproject.toml: "Python"
    requirements.txt: "Python"
    Gemfile: "Ruby"
    pom.xml: "Java (Maven)"
    build.gradle: "Java/Kotlin (Gradle)"
    mix.exs: "Elixir"
    composer.json: "PHP"
    pubspec.yaml: "Dart/Flutter"
    build.sbt: "Scala"
    CMakeLists.txt: "C/C++ (CMake)"
    meson.build: "C/C++ (Meson)"
    "*.csproj": "C# (.NET)"
    "*.sln": "C# (.NET)"
    Package.swift: "Swift"
    DESCRIPTION: "R"
    cpanfile: "Perl"
    Makefile.PL: "Perl"
    "*.rockspec": "Lua"
    .luacheckrc: "Lua"
    fpm.toml: "Fortran"
    alire.toml: "Ada"
    "*.gpr": "Ada"
    "*.cob": "COBOL"
    "*.cbl": "COBOL"
    "*.lpi": "Pascal"
    "*.vbproj": "VB.NET"

parallel_checks:
  mode: "PARALLEL (single message, multiple calls)"

  for_each_detected_language:
    - task: "lint-check"
      priority: "Makefile target > language-specific tool"
      commands:
        go: "golangci-lint run ./..."
        rust: "cargo clippy -- -D warnings"
        nodejs: "npm run lint"
        python: "ruff check ."
        ruby: "bundle exec rubocop"
        java-maven: "mvn checkstyle:check"
        java-gradle: "./gradlew check"
        elixir: "mix credo --strict"
        php: "vendor/bin/phpstan analyse"
        dart: "dart analyze --fatal-infos"

    - task: "build-check"
      commands:
        go: "go build ./..."
        rust: "cargo build --release"
        nodejs: "npm run build"
        java-maven: "mvn compile -q"
        java-gradle: "./gradlew build -x test"
        elixir: "mix compile --warnings-as-errors"

    - task: "test-check"
      commands:
        go: "go test -race ./..."
        rust: "cargo test"
        nodejs: "npm test"
        python: "pytest"
        ruby: "bundle exec rspec"
        java-maven: "mvn test -q"
        java-gradle: "./gradlew test"
        elixir: "mix test"
        php: "vendor/bin/phpunit"
        dart: "dart test"
```

**Output Multi-Language:**

```
═══════════════════════════════════════════════════════════════
   Pre-commit Checks
═══════════════════════════════════════════════════════════════

  Languages detected: Go, Rust

--- Rust Checks ---
[CHECK] Rust lint (clippy)...
[PASS] Rust lint (clippy)
[CHECK] Rust build...
[PASS] Rust build
[CHECK] Rust tests...
[PASS] Rust tests

--- Go Checks ---
[CHECK] Go lint (golangci-lint)...
[PASS] Go lint (golangci-lint)
[CHECK] Go build...
[PASS] Go build
[CHECK] Go tests (with race detection)...
[PASS] Go tests (with race detection)

═══════════════════════════════════════════════════════════════
   All pre-commit checks passed
═══════════════════════════════════════════════════════════════
```

**IMPORTANT**: Run `.claude/scripts/pre-commit-checks.sh` which auto-detects languages.

---

### Phase 5.0: Secret Scan (1Password Integration)

**ABSOLUTE RULE: No real secret/password must leak into a commit.**

**Secrets policy:**

| Type | Action | Example |
|------|--------|---------|
| Real secret (token, prod password) | **BLOCK the commit** | `ghp_abc123...`, `postgres://user:realpass@prod/db` |
| Test password | **OK if in `.example` file** | `.env.example`, `config.example.yaml` |
| Test password in code | **OK if explicitly commented** | `// TEST ONLY - not a real credential` |
| `.env` file with real secrets | **NEVER committed** | Must be in `.gitignore` |

**`.example` files:** Test passwords in `.example` files are accepted because they serve as documentation. They MUST have a comment explaining they are test values:

```bash
# .env.example - Test/default values only, NOT real credentials
DB_PASSWORD=test_password_change_me    # TEST ONLY
API_KEY=sk-test-fake-key-for-dev       # TEST ONLY
```

**Scan staged files for hardcoded secrets:**

```yaml
secret_scan:
  trigger: "ALWAYS run in parallel with language checks"
  blocking: true  # BLOCKS the commit if real secret detected

  0_policy:
    real_secrets: "BLOCK - never commit real tokens, passwords, API keys"
    test_passwords_in_example_files: "ALLOW - .example files are documentation"
    test_passwords_in_code: "ALLOW if commented with '// TEST ONLY' or '# TEST ONLY'"
    env_files: "BLOCK - .env must be in .gitignore, use .env.example instead"

  1_get_staged_files:
    command: "git diff --cached --name-only"
    exclude: [".env", ".env.*", "*.lock", "*.sum"]

  1b_check_env_not_staged:
    command: "git diff --cached --name-only | grep -E '^\.env$' || true"
    action: |
      IF .env is staged:
        BLOCK the commit
        Message: ".env potentially contains real secrets. Use .env.example for default values."

  2_scan_patterns:
    patterns:
      tokens:
        - 'ghp_[a-zA-Z0-9]{36}'           # GitHub PAT
        - 'glpat-[a-zA-Z0-9\-]{20}'       # GitLab PAT
        - 'sk-[a-zA-Z0-9]{48}'            # OpenAI/Stripe secret key
        - 'pk_[a-zA-Z0-9]{24,}'           # Stripe publishable key
        - 'ops_[a-zA-Z0-9]{50,}'          # 1Password service account
        - 'AKIA[0-9A-Z]{16}'             # AWS access key
      connection_strings:
        - 'postgres://[^\s]+'
        - 'mysql://[^\s]+'
        - 'mongodb(\+srv)?://[^\s]+'
      generic:
        - '[a-zA-Z0-9+/]{40,}={0,2}'     # Long base64 (potential secrets)

    exceptions:
      - file_pattern: "*.example*"         # .env.example, config.example.yaml
      - file_pattern: "*_example.*"
      - file_pattern: "*.sample*"
      - comment_marker: "TEST ONLY"        # Inline comment marks test value
      - comment_marker: "FAKE"
      - comment_marker: "PLACEHOLDER"
      - value_pattern: "test_*"            # test_password, test_token
      - value_pattern: "fake_*"
      - value_pattern: "dummy_*"
      - value_pattern: "changeme"
      - value_pattern: "TODO:*"

  3_if_secrets_found:
    action: "BLOCK commit + suggestion"
    output: |
      ═══════════════════════════════════════════════════════════════
        ⛔ REAL SECRETS DETECTED - COMMIT BLOCKED
      ═══════════════════════════════════════════════════════════════

        Found {count} potential secret(s) in staged files:

        File: src/config.go
          Line 42: ghp_xxxx... (GitHub PAT)
          Suggestion: /secret --push GITHUB_TOKEN=<value>
                      Replace with: os.Getenv("GITHUB_TOKEN")

        File: .env.production
          Line 5: postgres://user:pass@host/db
          Suggestion: /secret --push DATABASE_URL=<value>

        Action: Use /secret --push to store in 1Password
                Then replace with env var reference

        Test passwords? Put them in .env.example with comment:
          DB_PASSWORD=test_pass  # TEST ONLY

      ═══════════════════════════════════════════════════════════════

  4_if_no_secrets:
    output: "[PASS] No hardcoded secrets detected"
```

---

### Phase 6.0: Context Update (MANDATORY before commit)

**Updates CLAUDE.md files to reflect the branch modifications.**

**IMPORTANT**: This phase runs AFTER lint/test/build (Phase 3) to avoid
re-running `/warmup --update` if checks fail and require corrections.

```yaml
context_update_workflow:
  trigger: "ALWAYS (mandatory before commit)"
  position: "After Phase 3 + 3.5 (all checks pass), before Phase 4 (commit)"
  tool: "/warmup --update"

  1_collect_branch_diff:
    action: "Identify ALL files modified on the branch"
    command: |
      # Files modified in the entire branch (vs main)
      git diff main...HEAD --name-only 2>/dev/null || git diff HEAD --name-only
      # + unstaged/uncommitted files (in progress)
      git diff --name-only
      git diff --cached --name-only
      # Deduplicate
    output: "changed_files[] (unique list)"

  2_resolve_claude_files:
    action: "Find CLAUDE.md files affected by modified files"
    algorithm: |
      FOR each modified file:
        dir = dirname(file)
        WHILE dir != /workspace:
          IF exists(dir/CLAUDE.md):
            add(dir/CLAUDE.md) to set
          dir = parent(dir)
      # Always include /workspace/CLAUDE.md (root)
    output: "claude_files_to_update[] (unique set)"

  3_check_staleness:
    action: "Check the last update timestamp"
    algorithm: |
      FOR each claude_file IN claude_files_to_update:
        first_line = read_first_line(claude_file)
        IF first_line matches '<!-- updated: YYYY-MM-DDTHH:MM:SSZ -->':
          timestamp = parse_iso(first_line)
          age = now() - timestamp
          IF age < 5 minutes:
            skip(claude_file)  # Already up to date
            log("Skipping {claude_file} (updated {age} ago)")
        ELSE:
          include(claude_file)  # No timestamp = always update
    output: "stale_claude_files[] (files needing update)"

  4_run_warmup_update:
    condition: "stale_claude_files is not empty"
    action: "Run /warmup --update on stale files"
    tool: "Skill(warmup, --update)"
    scope: "Limited to directories of stale_claude_files"
    note: |
      /warmup --update will automatically add the ISO timestamp
      as the first line of each updated CLAUDE.md:
        <!-- updated: 2026-02-11T14:30:00Z -->

  5_stage_updated_docs:
    action: "Add updated CLAUDE.md files to staging"
    command: "git add **/CLAUDE.md"
    note: "Included in the same commit as code modifications"

  timestamp_format:
    format: "<!-- updated: YYYY-MM-DDTHH:MM:SSZ -->"
    example: "<!-- updated: 2026-02-11T14:30:00Z -->"
    position: "First line of CLAUDE.md file"
    purpose: "Freshness detection (staleness check 5 minutes)"
    parse: "ISO 8601 - easiest format to parse programmatically"
```

**Output Phase 3.8:**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Context Update (Phase 3.8)
═══════════════════════════════════════════════════════════════

  Branch diff: 12 files changed

  CLAUDE.md resolution:
    ├─ /workspace/CLAUDE.md (stale, 2h ago)
    ├─ .devcontainer/CLAUDE.md (stale, no timestamp)
    ├─ .devcontainer/images/CLAUDE.md (fresh, 3m ago) → SKIP
    └─ .devcontainer/hooks/CLAUDE.md (stale, 45m ago)

  /warmup --update:
    ✓ 3 CLAUDE.md files updated
    ✓ Timestamps refreshed
    ✓ Staged for commit

═══════════════════════════════════════════════════════════════
```

---

### Phase 7.0: Execute & Synthesize

```yaml
execute_workflow:
  1_branch:
    action: "Create or use branch"
    auto: true

  2_stage:
    action: "Stage ALL tracked modified files"
    steps:
      - command: "git add -A"
        note: "git add -A respects .gitignore automatically — no ignored file will be staged"
      - command: "git diff --name-only"
        verify: "MUST be empty — otherwise tracked files have been missed"
        on_failure: |
          If tracked files remain unstaged after git add -A:
          → Add them explicitly with git add <file>
          → NEVER ignore modifications to tracked files (CLAUDE.md, .claude/commands/, hooks/)
    rules:
      - "ALWAYS use git add -A (never selective staging by filename)"
      - "git add -A automatically includes: CLAUDE.md, .devcontainer/, .claude/commands/"
      - "git add -A automatically excludes: .env, mcp.json, .grepai/, .claude/* (except gitignore exceptions)"
      - "Check git diff --name-only after staging — if non-empty, there is a problem"
      - "If a tracked file should NOT be committed → git restore <file> BEFORE staging, not after"

  3_commit:
    action: "Create the commit"
    format: |
      <type>(<scope>): <description>

      [optional body]

  4_push:
    action: "Push to origin"
    command: "git push -u origin <branch>"

  5_pr_mr:
    action: "Create the PR/MR"
    tools:
      github: mcp__github__create_pull_request
      gitlab: mcp__gitlab__create_merge_request
    skip_if: "--no-pr"
```

**Final Output (GitHub):**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Completed (GitHub)
═══════════════════════════════════════════════════════════════

| Step    | Status                           |
|---------|----------------------------------|
| Peek    | ✓ 5 files analyzed               |
| Checks  | ✓ lint, test, build PASS         |
| Context | ✓ 3 CLAUDE.md updated            |
| Branch  | `feat/add-user-auth`             |
| Commit  | `feat(auth): add user auth`      |
| Push    | origin/feat/add-user-auth        |
| PR      | #42 - feat(auth): add user auth  |

URL: https://github.com/<owner>/<repo>/pull/42

═══════════════════════════════════════════════════════════════
```

**Final Output (GitLab):**

```
═══════════════════════════════════════════════════════════════
  /git --commit - Completed (GitLab)
═══════════════════════════════════════════════════════════════

| Step    | Status                           |
|---------|----------------------------------|
| Peek    | ✓ 5 files analyzed               |
| Checks  | ✓ lint, test, build PASS         |
| Context | ✓ 3 CLAUDE.md updated            |
| Branch  | `feat/add-user-auth`             |
| Commit  | `feat(auth): add user auth`      |
| Push    | origin/feat/add-user-auth        |
| MR      | !42 - feat(auth): add user auth  |

URL: https://gitlab.com/<owner>/<repo>/-/merge_requests/42

═══════════════════════════════════════════════════════════════
```

---

## Action: --merge

### MCP-ONLY Policy (STRICT - Issue #142)

**NEVER use CLI for pipeline status. Always use MCP tools:**

```yaml
mcp_only_policy:
  MANDATORY:
    github:
      pipeline_status: "mcp__github__get_pull_request"
      check_runs: "mcp__github__list_check_runs (via pull_request_read)"
    gitlab:
      pipeline_status: "mcp__gitlab__list_pipelines"
      pipeline_jobs: "mcp__gitlab__list_pipeline_jobs"

  FORBIDDEN:
    - "gh pr checks"
    - "gh run view"
    - "glab ci status"
    - "glab ci view"
    - "curl api.github.com"
    - "curl gitlab.com/api"

  rationale: |
    CLI commands return stale/cached data and require parsing
    MCP provides structured JSON with real-time status
```

---

### Phase 1.0: Peek + Commit-Pinned Tracking

**CRITICAL: Track pipeline for SPECIFIC commit SHA**

```yaml
peek_workflow:
  0_get_pushed_commit:
    action: "Get SHA of just-pushed commit"
    command: "git rev-parse HEAD"
    store: "pushed_commit_sha"
    critical: true

  1_pr_mr_info:
    action: "Retrieve PR/MR info"
    tools:
      github: mcp__github__get_pull_request
      gitlab: mcp__gitlab__get_merge_request
    verify: "head_sha == pushed_commit_sha"
    output: "pr_mr_number, head_sha, status, checks"

  2_find_pipeline:
    action: "Find pipeline triggered by THIS commit"
    github: |
      # Verify: check_run.head_sha == pushed_commit_sha
      mcp__github__pull_request_read(method="get")
    gitlab: |
      # Filter: pipeline.sha == pushed_commit_sha
      mcp__gitlab__list_pipelines(sha=pushed_commit_sha)

  3_validate_pipeline:
    action: "Abort if pipeline not found within 60s"
    timeout: 60s
    on_timeout: "ERROR: No pipeline triggered for commit {sha}"

  4_conflicts:
    action: "Check for conflicts"
    command: "git fetch && git merge-base..."
```

**Output Phase 1:**

```
═══════════════════════════════════════════════════════════════
  /git --merge - Pipeline Tracking
═══════════════════════════════════════════════════════════════

  Commit: abc1234 (verified)
  PR: #42

  Pipeline found:
    ├─ ID: 12345
    ├─ SHA: abc1234 ✓ (matches pushed commit)
    ├─ Triggered: 15s ago
    └─ Status: running

═══════════════════════════════════════════════════════════════
```

---

### Phase 2.0: Job-Level Status Parsing (CRITICAL)

**Parse EACH job individually, not overall status:**

```yaml
status_parsing:
  github:
    statuses:
      success: ["success", "neutral"]
      pending: ["queued", "in_progress", "waiting", "pending"]
      failure: ["failure", "action_required", "timed_out"]
      cancelled: ["cancelled", "stale"]
      skipped: ["skipped"]

    aggregation_rule: |
      # CRITICAL: A single failed job = PIPELINE FAILED
      pipeline_success = ALL jobs in [success, skipped, neutral]
      pipeline_failure = ANY job in [failure, cancelled, timed_out]
      pipeline_pending = ANY job in [pending, queued, in_progress]

      # DO NOT report success if any job failed!

  gitlab:
    statuses:
      success: ["success", "manual"]
      pending: ["created", "waiting_for_resource", "preparing", "pending", "running"]
      failure: ["failed"]
      cancelled: ["canceled"]
      skipped: ["skipped"]

job_by_job_output:
  format: |
    ═══════════════════════════════════════════════════════════════
      CI Status - Commit {sha}
    ═══════════════════════════════════════════════════════════════

      Pipeline: #{id} (triggered {time_ago})
      Branch:   {branch}
      Commit:   {sha} ✓ (verified)

      Jobs:
        ├─ lint      : ✓ passed (45s)
        ├─ build     : ✓ passed (1m 23s)
        ├─ test      : ✗ FAILED (2m 15s)    <-- FAILED
        └─ deploy    : ⊘ skipped

      Overall: ✗ FAILED (1 job failed)

    ═══════════════════════════════════════════════════════════════
```

---

### Phase 3.0: CI Monitoring with Exponential Backoff and Hard Timeout

**ABSOLUTE LIMIT: 10 minutes / 30 polls**

```yaml
ci_monitoring:
  description: "Intelligent CI status tracking with adaptive polling"

  #---------------------------------------------------------------------------
  # CONFIGURATION
  #---------------------------------------------------------------------------
  config:
    initial_interval: 10s          # Initial interval
    max_interval: 120s             # Capped at 2 minutes
    backoff_multiplier: 1.5        # 10s → 15s → 22s → 33s → 50s → 75s → 112s → 120s
    jitter_percent: 20             # +/- 20% random (prevents thundering herd)
    timeout: 600s                  # 10 minutes HARD timeout total
    max_poll_attempts: 30          # Safety limit

  #---------------------------------------------------------------------------
  # POLLING STRATEGY (MCP-ONLY - NO CLI FALLBACK)
  #---------------------------------------------------------------------------
  polling_strategy:
    github:
      tool: mcp__github__get_pull_request
      params:
        pull_number: "{pr_number}"
      response_fields: ["state", "statuses[]", "check_runs[]"]
      # NO FALLBACK - CLI FORBIDDEN

    gitlab:
      tool: mcp__gitlab__list_pipelines
      params:
        project_id: "{project_id}"
        ref: "{branch}"
        per_page: 1
      response_fields: ["status", "id", "web_url"]
      # NO FALLBACK - CLI FORBIDDEN

  #---------------------------------------------------------------------------
  # EXPONENTIAL BACKOFF ALGORITHM
  #---------------------------------------------------------------------------
  backoff_algorithm:
    pseudocode: |
      interval = initial_interval
      elapsed = 0
      attempt = 0

      WHILE elapsed < timeout AND attempt < max_poll_attempts:
        status = poll_ci_status()  # MCP ONLY

        IF status == SUCCESS:
          RETURN {status: "passed", duration: elapsed}
        IF status in [FAILURE, ERROR, CANCELED]:
          RETURN {status: "failed", duration: elapsed, details: get_failure_details()}
        IF status in [PENDING, RUNNING]:
          # Apply jitter
          jitter = interval * (random(-jitter_percent, +jitter_percent) / 100)
          sleep(interval + jitter)
          elapsed += interval + jitter

          # Exponential backoff
          interval = min(interval * backoff_multiplier, max_interval)
          attempt++

      RETURN {status: "timeout", duration: elapsed}

  #---------------------------------------------------------------------------
  # ON TIMEOUT
  #---------------------------------------------------------------------------
  on_timeout:
    action: "ABORT immediately"
    output: |
      ═══════════════════════════════════════════════════════════════
        ⛔ Pipeline Timeout
      ═══════════════════════════════════════════════════════════════

        Waited: 10 minutes
        Polls:  30 attempts
        Status: Still pending

        This usually means:
        - Pipeline is stuck
        - Pipeline was cancelled externally
        - Wrong pipeline being monitored

        Actions:
        1. Check pipeline manually: {pipeline_url}
        2. Re-run: /git --merge
        3. Force: /git --merge --skip-ci (if CI is broken)

      ═══════════════════════════════════════════════════════════════

  #---------------------------------------------------------------------------
  # PARALLEL TASKS (during polling)
  #---------------------------------------------------------------------------
  parallel_tasks:
    - task: "Check conflicts"
      action: "git fetch && git merge-base --is-ancestor origin/main HEAD"
      on_conflict: "Automatic rebase if --auto-rebase"

    - task: "Sync with main"
      action: "Rebase if behind (max 10 commits)"
      on_behind: "git rebase origin/main"
```

**Output Phase 2.5:**

```
═══════════════════════════════════════════════════════════════
  /git --merge - CI Monitoring (Phase 2.5)
═══════════════════════════════════════════════════════════════

  PR/MR    : #42 (feat/add-auth)
  Platform : GitHub
  Timeout  : 10 minutes (HARD LIMIT)

  Polling CI status (MCP-ONLY)...
    [10:30:15] Poll #1: pending (10s elapsed, next in 10s)
    [10:30:27] Poll #2: running (22s elapsed, next in 15s)
    [10:30:45] Poll #3: running (40s elapsed, next in 22s)
    [10:31:12] Poll #4: running (67s elapsed, next in 33s)
    [10:31:50] ✓ CI PASSED (95s)

  Job-level verification:
    ├─ lint: ✓ passed (45s)
    ├─ build: ✓ passed (1m 23s)
    └─ test: ✓ passed (2m 45s)

  Proceeding to Phase 3...

═══════════════════════════════════════════════════════════════
```

---

### Phase 4.0: Error Log Extraction (on failure)

**When pipeline fails, extract actionable information:**

```yaml
error_extraction:
  step_1_identify:
    action: "Get list of failed jobs"
    output: "[job_name, job_id, failure_reason]"

  step_2_parse_error:
    patterns:
      lint_error:
        - "eslint.*error"
        - "golangci-lint"
        - "clippy::"
        - "ruff.*error"
      build_error:
        - "cannot find module"
        - "compilation failed"
        - "cargo build.*error"
        - "tsc.*error"
      test_error:
        - "FAIL.*test"
        - "AssertionError"
        - "--- FAIL:"
        - "pytest.*FAILED"
      security_error:
        - "CRITICAL.*vulnerability"
        - "CVE-"
        - "HIGH.*severity"

  step_3_generate_debug_plan:
    output: |
      ═══════════════════════════════════════════════════════════════
        Pipeline Failed - Debug Plan
      ═══════════════════════════════════════════════════════════════

        Failed Job: {job_name}
        Error Type: {error_type}
        Exit Code:  {exit_code}

        Error Summary:
        ┌─────────────────────────────────────────────────────────────
        │ {error_excerpt_20_lines}
        └─────────────────────────────────────────────────────────────

        Suggested Actions:
        1. {action_1_based_on_error_type}
        2. {action_2_based_on_error_type}
        3. Run locally: {local_command}

        Next Step: Run `/plan debug {error_type}` to investigate

      ═══════════════════════════════════════════════════════════════
```

---

### Phase 5.0: Auto-fix Loop with Error Categories

```yaml
autofix_loop:
  description: "Detection, categorization and automatic correction of CI errors"

  #---------------------------------------------------------------------------
  # CONFIGURATION
  #---------------------------------------------------------------------------
  config:
    max_attempts: 3
    cooldown_between_attempts: 30s    # Wait before re-triggering CI
    autofix_per_attempt_timeout: 120s # 2 min max per fix attempt
    require_human_for:
      - security_scan
      - timeout
      - "confidence == LOW after 2 attempts"

  #---------------------------------------------------------------------------
  # ERROR CATEGORIES
  #---------------------------------------------------------------------------
  error_categories:
    #-------------------------------------------------------------------------
    # LINT ERRORS - Auto-fixable (HIGH confidence)
    #-------------------------------------------------------------------------
    lint_error:
      patterns:
        - "eslint.*error"
        - "prettier.*differ"
        - "golangci-lint.*"
        - "ruff.*error"
        - "shellcheck.*SC[0-9]+"
        - "stylelint.*"
      severity: LOW
      auto_fixable: true
      confidence: HIGH
      fix_strategy: "run_linter_fix"

    #-------------------------------------------------------------------------
    # TYPE ERRORS - Partially auto-fixable
    #-------------------------------------------------------------------------
    type_error:
      patterns:
        - "TS[0-9]+:"                    # TypeScript errors
        - "type.*incompatible"
        - "cannot find name"
        - "go build.*undefined:"         # Go type errors
        - "mypy.*error:"                 # Python mypy
      severity: MEDIUM
      auto_fixable: partial
      confidence: MEDIUM
      fix_strategy: "type_fix"

    #-------------------------------------------------------------------------
    # TEST FAILURES - Conditional auto-fix
    #-------------------------------------------------------------------------
    test_failure:
      patterns:
        - "FAIL.*test"
        - "AssertionError"
        - "expected.*but got"
        - "Error: expect\\("
        - "--- FAIL:"                    # Go test failures
        - "FAILED.*::.*::"               # pytest
      severity: HIGH
      auto_fixable: conditional
      confidence: MEDIUM
      fix_strategy: "test_analysis"

    #-------------------------------------------------------------------------
    # BUILD ERRORS - Requires careful analysis
    #-------------------------------------------------------------------------
    build_error:
      patterns:
        - "error: cannot find module"
        - "Module not found"
        - "compilation failed"
        - "SyntaxError:"
        - "package.*not found"
      severity: HIGH
      auto_fixable: partial
      confidence: LOW
      fix_strategy: "build_analysis"

    #-------------------------------------------------------------------------
    # SECURITY SCAN - NEVER auto-fix
    #-------------------------------------------------------------------------
    security_scan:
      patterns:
        - "CRITICAL.*vulnerability"
        - "HIGH.*CVE-"
        - "security.*violation"
        - "secret.*detected"
        - "trivy.*CRITICAL"
      severity: CRITICAL
      auto_fixable: false
      confidence: N/A
      fix_strategy: "user_intervention_required"

    #-------------------------------------------------------------------------
    # DEPENDENCY ERRORS - Often auto-fixable
    #-------------------------------------------------------------------------
    dependency_error:
      patterns:
        - "npm ERR!.*peer dep"
        - "cannot resolve dependency"
        - "go: module.*not found"
        - "pip.*ResolutionImpossible"
      severity: MEDIUM
      auto_fixable: true
      confidence: MEDIUM
      fix_strategy: "dependency_fix"

    #-------------------------------------------------------------------------
    # INFRASTRUCTURE ERRORS - Retry only
    #-------------------------------------------------------------------------
    infrastructure_error:
      patterns:
        - "rate limit"
        - "connection refused"
        - "503 Service Unavailable"
        - "ECONNRESET"
      severity: LOW
      auto_fixable: retry
      confidence: HIGH
      fix_strategy: "retry_ci"

  #---------------------------------------------------------------------------
  # LOOP ALGORITHM
  #---------------------------------------------------------------------------
  loop_algorithm:
    pseudocode: |
      attempt = 0
      fix_history = []

      WHILE attempt < max_attempts:
        attempt++

        # Step 1: Retrieve CI failure details
        failure = get_ci_failure_details()
        category = categorize_error(failure)

        # Step 2: Check if auto-fixable
        IF NOT category.auto_fixable:
          RETURN abort_with_report(category, failure)

        # Step 3: Detect circular fix
        IF is_circular_fix(category, fix_history):
          RETURN abort_with_circular_warning(fix_history)

        # Step 4: Apply fix strategy
        fix_result = apply_fix_strategy(category)
        fix_history.append({category, fix_result})

        IF fix_result.success:
          # Step 5: Commit and push
          commit_fix(fix_result)
          push_to_remote()

          # Step 6: Wait cooldown then re-poll CI
          sleep(cooldown_between_attempts)
          ci_status = poll_ci_with_backoff()  # Re-use Phase 2.5

          IF ci_status == SUCCESS:
            RETURN success_report(attempt, fix_history)
        ELSE:
          RETURN abort_with_fix_failure(fix_result)

      # Max attempts reached
      RETURN abort_max_attempts(fix_history)

  #---------------------------------------------------------------------------
  # FIX STRATEGIES
  #---------------------------------------------------------------------------
  fix_strategies:
    run_linter_fix:
      detect_linter:
        - check: "package.json"
          command: "npm run lint -- --fix"
        - check: ".golangci.yml"
          command: "golangci-lint run --fix"
        - check: "pyproject.toml [tool.ruff]"
          command: "ruff check --fix"
      commit_format: "fix(lint): auto-fix {linter} errors"

    type_fix:
      workflow:
        1_extract: "Parse CI log for specific type errors"
        2_analyze: "Identify the file and line"
        3_fix: "Apply minimal correction"
        4_verify: "npm run typecheck OR go build"
      commit_format: "fix(types): resolve {error_code} in {file}"

    test_analysis:
      conditions:
        assertion_mismatch:
          pattern: "expected.*but got"
          auto_fix: true
          strategy: "Update assertion if implementation changed"
        snapshot_mismatch:
          pattern: "snapshot.*differ"
          auto_fix: true
          strategy: "npm test -- -u"
        timeout:
          pattern: "exceeded timeout"
          auto_fix: false
      commit_format: "fix(test): update {test_name}"

    dependency_fix:
      strategies:
        npm: "npm install --legacy-peer-deps"
        go: "go mod tidy"
        pip: "pip install --upgrade"
      commit_format: "fix(deps): resolve {package} conflict"

    retry_ci:
      wait: 60s
      retrigger:
        github: "gh run rerun --failed"
        gitlab: "glab ci retry"

    user_intervention_required:
      action: "Generate detailed failure report"
      include:
        - "Error category and severity"
        - "Relevant CI log snippets (max 50 lines)"
        - "Affected files"
        - "Suggested manual steps"
      block_merge: true
```

**Output Phase 4 (Auto-fix Success):**

```
═══════════════════════════════════════════════════════════════
  /git --merge - Auto-fix Loop (Phase 4)
═══════════════════════════════════════════════════════════════

  Attempt 1/3 - lint_error
  -------------------------
    Category : lint_error (LOW severity)
    Confidence: HIGH
    Auto-fix : YES

    Error: eslint: 3 errors in src/utils/parser.ts
      ├─ Line 45: no-unused-vars
      ├─ Line 67: prefer-const
      └─ Line 89: no-console

    Fix: npm run lint -- --fix
    Result: ✓ Fixed

    Commit: fix(lint): auto-fix eslint errors in parser.ts
    Push: origin/feat/add-parser

    Re-polling CI...
      [10:32:45] ✓ CI PASSED (67s)

═══════════════════════════════════════════════════════════════
  ✓ Auto-fix Successful (1 attempt)
═══════════════════════════════════════════════════════════════

  Commits added: 1
    └─ fix(lint): auto-fix eslint errors in parser.ts

  Proceeding to Phase 5 (Merge)...

═══════════════════════════════════════════════════════════════
```

**Output Phase 4 (Security Block):**

```
═══════════════════════════════════════════════════════════════
  /git --merge - BLOCKED (Security Issue)
═══════════════════════════════════════════════════════════════

  ⛔ AUTO-FIX DISABLED for security issues

  Category: security_scan
  Severity: CRITICAL

  Vulnerability:
    ┌─────────────────────────────────────────────────────────┐
    │ CRITICAL CVE-2023-44487                                 │
    │ Package: golang.org/x/net v0.7.0                        │
    │ Fixed in: v0.17.0                                       │
    └─────────────────────────────────────────────────────────┘

  Required Actions:
    1. go get golang.org/x/net@v0.17.0 && go mod tidy
    2. trivy fs --severity CRITICAL .
    3. Re-run /git --merge

  ⚠️  Force merge NOT available for security issues.

═══════════════════════════════════════════════════════════════
```

---

### Phase 6.0: Synthesize (Merge & Cleanup)

```yaml
merge_workflow:
  1_final_verify:
    action: "Verify ALL jobs passed (job-level check)"
    tools:
      github: mcp__github__get_pull_request
      gitlab: mcp__gitlab__get_merge_request
    condition: "ALL check_runs.conclusion == 'success'"

  2_merge:
    tools:
      github: mcp__github__merge_pull_request
      gitlab: mcp__gitlab__merge_merge_request
    method: "squash"

  3_cleanup:
    actions:
      - "git push origin --delete <branch>"
      - "git branch -D <branch>"
      - "git checkout main"
      - "git pull origin main"
```

**Final Output (GitHub):**

```
═══════════════════════════════════════════════════════════════
  ✓ PR #42 merged successfully
═══════════════════════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  Rebase  : ✓ Synced (was 3 commits behind)

  CI (job-level verification):
    ├─ lint      : ✓ passed
    ├─ build     : ✓ passed
    ├─ test      : ✓ passed
    └─ security  : ✓ passed

  Total CI Time: 2m 34s

  Commits : 5 commits → 1 squashed

  Cleanup:
    ✓ Remote branch deleted
    ✓ Local branch deleted
    ✓ Switched to main
    ✓ Pulled latest (now at abc1234)

═══════════════════════════════════════════════════════════════
```

**Final Output (GitLab):**

```
═══════════════════════════════════════════════════════════════
  ✓ MR !42 merged successfully
═══════════════════════════════════════════════════════════════

  Branch  : feat/add-auth → main
  Method  : squash
  Pipeline: ✓ Passed (#12345, 2m 34s)
  Commits : 5 commits → 1 squashed

  Cleanup:
    ✓ Remote branch deleted
    ✓ Local branch deleted
    ✓ Switched to main
    ✓ Pulled latest (now at abc1234)

═══════════════════════════════════════════════════════════════
```

---

## Conventional Commits

| Type | Usage |
|------|-------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Refactoring |
| `docs` | Documentation |
| `test` | Tests |
| `chore` | Maintenance |
| `ci` | CI/CD |

---

## Guardrails (ABSOLUTE)

| Action | Status | Reason |
|--------|--------|--------|
| Skip Phase 0.5 (Identity) without flag | **FORBIDDEN** | Git identity required |
| Skip Phase 1 (Peek) | **FORBIDDEN** | git status before action |
| Skip Phase 3.8 (Context) | **FORBIDDEN** | CLAUDE.md must reflect changes |
| Skip Phase 2 (CI Polling) | **FORBIDDEN** | CI validation mandatory |
| Automatic merge without CI | **FORBIDDEN** | Code quality |
| Push to main/master | **FORBIDDEN** | Protected branch |
| Force merge if CI fails x3 | **FORBIDDEN** | Attempt limit |
| Push without --force-with-lease | **FORBIDDEN** | Safety |
| AI mentions in commits | **FORBIDDEN** | Discretion |
| Commit without validated identity | **FORBIDDEN** | Traceability |
| CLI for CI status | **FORBIDDEN** | MCP-ONLY policy |
| Report success if ANY job failed | **FORBIDDEN** | Job-level parsing |
| Wait > 10 min for pipeline | **FORBIDDEN** | Hard timeout |
| Monitor wrong commit's pipeline | **FORBIDDEN** | Commit-pinned tracking |

### Auto-fix Safeguards

| Action | Status | Reason |
|--------|--------|--------|
| Auto-fix security vulnerabilities | **FORBIDDEN** | Human review required |
| Merge with CRITICAL issues | **FORBIDDEN** | Security first |
| Circular fix (same error 3x) | **FORBIDDEN** | Prevents infinite loop |
| Modify .claude/ via auto-fix | **FORBIDDEN** | Protected config |
| Modify .devcontainer/ via auto-fix | **FORBIDDEN** | Protected config |
| Auto-fix without commit message | **FORBIDDEN** | Traceability |

### Auto-fix Timeouts

| Element | Value | Reason |
|---------|-------|--------|
| CI Polling total | 600s (10min) | Prevent infinite wait |
| Per fix attempt | 120s (2min) | Prevent blocking |
| Cooldown between attempts | 30s | Allow CI to start |
| Polling jitter | ±20% | Prevent thundering herd |

### CLI Commands FORBIDDEN for CI Monitoring

```yaml
forbidden_cli:
  github:
    - "gh pr checks"
    - "gh run view"
    - "gh run list"
    - "gh api repos/.../check-runs"
  gitlab:
    - "glab ci status"
    - "glab ci view"
    - "glab pipeline status"
  generic:
    - "curl *api.github.com*"
    - "curl *gitlab.com/api*"

required_mcp:
  github: "mcp__github__get_pull_request, mcp__github__pull_request_read"
  gitlab: "mcp__gitlab__list_pipelines, mcp__gitlab__list_pipeline_jobs"
```

### Legitimate Parallelization

| Element | Parallel? | Reason |
|---------|-----------|--------|
| Pre-commit checks (lint+test+build) | Parallel | Independent |
| Language checks (Go+Rust+Node) | Parallel | Independent |
| CI polling + conflict check | Parallel | Independent |
| Git operations (branch→commit→push→PR) | Sequential | Dependency chain |
| Auto-fix attempts | Sequential | Depends on CI result |
| CI checks waiting | Sequential | Wait for result |
| Pipeline polling | Sequential | State changes between polls |
