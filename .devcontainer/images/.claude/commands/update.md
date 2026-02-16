---
name: update
description: |
  DevContainer Environment Update from official template.
  Updates hooks, commands, agents, and settings from kodflow/devcontainer-template.
  Use when: syncing local devcontainer with latest template improvements.
allowed-tools:
  - "Bash(curl:*)"
  - "Bash(git:*)"
  - "Bash(jq:*)"
  - "Read(**/*)"
  - "Write(.devcontainer/**/*)"
  - "WebFetch(*)"
  - "Task(*)"
---

# Update - DevContainer Environment Update

$ARGUMENTS

---

## Description

Updates the DevContainer environment from the official template.

**API-FIRST approach**: Uses the GitHub API to dynamically discover
existing files instead of hardcoded lists.

**Updated components:**

- **Hooks** - Claude scripts (format, lint, security, etc.)
- **Commands** - Slash commands (/git, /search, etc.)
- **Agents** - Agent definitions (specialists, executors)
- **Lifecycle** - Lifecycle hooks (delegation stubs)
- **Image-hooks** - Hooks embedded in the Docker image (real logic)
- **Shared-utils** - Shared utilities (utils.sh)
- **Config** - p10k, settings.json
- **Compose** - docker-compose.yml (update devcontainer, preserve custom)
- **Grepai** - Optimized grepai configuration

**Source**: `github.com/kodflow/devcontainer-template`

---

## Arguments

| Pattern | Action |
|---------|--------|
| (none) | Full update |
| `--check` | Check for available updates |
| `--component <name>` | Update a specific component |
| `--help` | Show help |

### Available components

| Component | Path | Description |
|-----------|------|-------------|
| `hooks` | `.devcontainer/images/.claude/scripts/` | Claude scripts |
| `commands` | `.devcontainer/images/.claude/commands/` | Slash commands |
| `agents` | `.devcontainer/images/.claude/agents/` | Agent definitions |
| `lifecycle` | `.devcontainer/hooks/lifecycle/` | Lifecycle hooks (stubs) |
| `image-hooks` | `.devcontainer/images/hooks/` | Image-embedded lifecycle hooks |
| `shared-utils` | `.devcontainer/hooks/shared/utils.sh` | Shared hook utilities |
| `p10k` | `.devcontainer/images/.p10k.zsh` | Powerlevel10k config |
| `settings` | `.../images/.claude/settings.json` | Claude config |
| `compose` | `.devcontainer/docker-compose.yml` | Update devcontainer service |
| `grepai` | `.devcontainer/images/grepai.config.yaml` | grepai config |

---

## --help

```
═══════════════════════════════════════════════
  /update - DevContainer Environment Update
═══════════════════════════════════════════════

Usage: /update [options]

Options:
  (none)              Full update
  --check             Check for updates
  --component <name>  Update a component
  --help              Show this help

Components:
  hooks        Claude scripts (format, lint...)
  commands     Slash commands (/git, /search)
  agents       Agent definitions (specialists)
  lifecycle    Lifecycle hooks (delegation stubs)
  image-hooks  Image-embedded lifecycle hooks
  shared-utils Shared hook utilities (utils.sh)
  p10k         Powerlevel10k config
  settings     Claude settings.json
  compose      docker-compose.yml (devcontainer service)
  grepai       grepai config (provider, model)

Examples:
  /update                       Update everything
  /update --check               Check for updates
  /update --component hooks     Hooks only

Source: kodflow/devcontainer-template (main)
═══════════════════════════════════════════════
```

---

## Overview

DevContainer environment update using **RLM** patterns:

- **Peek** - Verify connectivity and versions
- **Discover** - Dynamically discover files via GitHub API
- **Validate** - Validate each download (no 404)
- **Synthesize** - Apply updates and consolidated report

---

## Configuration

```yaml
REPO: "kodflow/devcontainer-template"
BRANCH: "main"
BASE_URL: "https://raw.githubusercontent.com/${REPO}/${BRANCH}"
API_URL: "https://api.github.com/repos/${REPO}/contents"
```

---

## ZSH Compatibility (CRITICAL)

**The default shell is `zsh` (set via `chsh -s /bin/zsh` in Dockerfile).**
Claude Code's Bash tool executes commands using `$SHELL` (zsh), not bash.

**RULE: All inline scripts MUST be zsh-compatible.**

| Pattern | Status | Reason |
|---------|--------|--------|
| `for x in $VAR` | **BROKEN in zsh** | zsh does not split variables on IFS |
| `while IFS= read -r x; do` | **WORKS everywhere** | Portable bash/zsh |
| `for x in literal1 literal2` | **WORKS everywhere** | No variable expansion |

**Always use `while read` for iterating over command output:**

```bash
# CORRECT (works in both bash and zsh):
curl ... | jq ... | while IFS= read -r item; do
    [ -z "$item" ] && continue
    echo "$item"
done

# INCORRECT (breaks in zsh - variable not split):
ITEMS=$(curl ... | jq ...)
for item in $ITEMS; do
    echo "$item"
done
```

**For the reference script:** Write to a temp file and execute with `bash` explicitly:
```bash
# Write script to temp file, then run with bash
cat > /tmp/update-script.sh << 'SCRIPT'
#!/bin/bash
# ... script content ...
SCRIPT
bash /tmp/update-script.sh && rm -f /tmp/update-script.sh
```

---

## Phase 1.0: Environment Detection (NEW)

**MANDATORY: Detect execution context before any operation.**

```yaml
environment_detection:
  1_container_check:
    action: "Detect if running inside container"
    method: "[ -f /.dockerenv ]"
    output: "IS_CONTAINER (true|false)"

  2_devcontainer_check:
    action: "Check DEVCONTAINER env var"
    method: "[ -n \"${DEVCONTAINER:-}\" ]"
    note: "Set by VS Code when attached to devcontainer"

  3_determine_target:
    container_mode:
      target: "/workspace/.devcontainer/images/.claude"
      behavior: "Update template source (requires rebuild)"
      propagation: "Changes applied at next container start"

    host_mode:
      target: "$HOME/.claude"
      behavior: "Update user Claude configuration"
      propagation: "Immediate (no rebuild needed)"

  4_display_context:
    output: |
      Environment: {CONTAINER|HOST}
      Update target: {path}
      Mode: {template|user}
```

**Implementation:**

```bash
# Detect environment context
detect_context() {
    # Check if running inside container
    if [ -f /.dockerenv ]; then
        CONTEXT="container"
        UPDATE_TARGET="/workspace/.devcontainer/images/.claude"
        echo "Detected: Container environment"
    else
        CONTEXT="host"
        UPDATE_TARGET="$HOME/.claude"
        echo "Detected: Host machine"
    fi

    # Additional checks
    if [ -n "${DEVCONTAINER:-}" ]; then
        echo "  (DevContainer detected via DEVCONTAINER env var)"
    fi

    echo "Update target: $UPDATE_TARGET"
    echo "Mode: $CONTEXT"
}

# Call at start of update
detect_context
```

**Output Phase 1.0:**

```
═══════════════════════════════════════════════
  /update - Environment Detection
═══════════════════════════════════════════════

  Environment: HOST MACHINE
  Update target: /home/user/.claude
  Mode: user configuration

  Changes will be:
    - Applied immediately
    - No container rebuild needed
    - Synced to container via postStart.sh

═══════════════════════════════════════════════
```

Or in container:

```
═══════════════════════════════════════════════
  /update - Environment Detection
═══════════════════════════════════════════════

  Environment: DEVCONTAINER
  Update target: /workspace/.devcontainer/images/.claude
  Mode: template source

  Changes will be:
    - Applied to template files
    - Require container rebuild to propagate
    - Or wait for next postStart.sh sync

═══════════════════════════════════════════════
```

---

## Phase 2.0: Peek (Version Check)

```yaml
peek_workflow:
  1_connectivity:
    action: "Verify GitHub connectivity"
    tool: WebFetch
    url: "https://api.github.com/repos/kodflow/devcontainer-template/commits/main"

  2_local_version:
    action: "Read local version"
    tool: Read
    file: ".devcontainer/.template-version"
```

**Output Phase 2.0:**

```
═══════════════════════════════════════════════
  /update - Peek Analysis
═══════════════════════════════════════════════

  Connectivity   : ✓ GitHub API accessible
  Local version  : abc1234 (2024-01-15)
  Remote version : def5678 (2024-01-20)

  Status: UPDATE AVAILABLE

═══════════════════════════════════════════════
```

---

## Phase 3.0: Discover (API-FIRST - Dynamic Discovery)

**CRITICAL RULE: Always use the GitHub API to discover files.**

NEVER use hardcoded lists. Files can be added,
renamed, or deleted in the source template.

```yaml
discover_workflow:
  strategy: "API-FIRST"

  components:
    hooks:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/scripts"
      filter: "*.sh"
      local_path: ".devcontainer/images/.claude/scripts/"

    commands:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/commands"
      filter: "*.md"
      local_path: ".devcontainer/images/.claude/commands/"

    agents:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/agents"
      filter: "*.md"
      local_path: ".devcontainer/images/.claude/agents/"

    lifecycle:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/hooks/lifecycle"
      filter: "*.sh"
      local_path: ".devcontainer/hooks/lifecycle/"

    image-hooks:
      api: "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/hooks"
      recursive: true
      local_path: ".devcontainer/images/hooks/"
      note: "Image-embedded lifecycle hooks (real logic)"

    shared-utils:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/hooks/shared/utils.sh"
      local_path: ".devcontainer/hooks/shared/utils.sh"
      note: "Needed by initialize.sh (runs on host)"

    p10k:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/.p10k.zsh"
      local_path: ".devcontainer/images/.p10k.zsh"

    settings:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/.claude/settings.json"
      local_path: ".devcontainer/images/.claude/settings.json"

    compose:
      strategy: "REPLACE from template, PRESERVE custom services"
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/docker-compose.yml"
      local_path: ".devcontainer/docker-compose.yml"
      note: |
        - If file absent -> download complete
        - If file exists:
          1. Extract custom services (not devcontainer)
          2. Replace entirely from template (preserves order/comments)
          3. Merge extracted custom services back
        - Order: devcontainer -> custom
        - Backup created before modification, restored on failure
        - Uses mikefarah/yq (Go version) for merge
        - Note: Ollama runs on HOST (installed via initialize.sh)

    grepai:
      raw_url: "https://raw.githubusercontent.com/kodflow/devcontainer-template/main/.devcontainer/images/grepai.config.yaml"
      local_path: ".devcontainer/images/grepai.config.yaml"
      note: "Optimized config with bge-m3 model (best accuracy)"
```

**Discover implementation:**

```bash
# Function to list files from a directory via GitHub API
list_remote_files() {
    local api_url="$1"
    local filter="$2"

    curl -sL "$api_url" | jq -r '.[].name' | grep -E "$filter" || true
}

# Example: discover scripts
SCRIPTS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/scripts" \
    '\.sh$')

# Example: discover commands
COMMANDS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/commands" \
    '\.md$')

# Example: discover agents
AGENTS=$(list_remote_files \
    "https://api.github.com/repos/kodflow/devcontainer-template/contents/.devcontainer/images/.claude/agents" \
    '\.md$')
```

---

## Phase 4.0: Validate (Download with Verification)

**CRITICAL RULE: Always validate downloads before writing.**

NEVER write a file without verifying that the download succeeded.
Detect 404 errors and other failures.

```yaml
validate_workflow:
  rule: "NEVER write files without validation"

  checks:
    - "HTTP status 200 (not 404)"
    - "Content is not empty"
    - "Content is not HTML error page"
    - "Content starts with expected pattern"

  on_failure:
    - "Log error"
    - "Skip file"
    - "Continue with next file"
```

**Validate implementation:**

```bash
# Secure download function with validation
safe_download() {
    local url="$1"
    local output="$2"
    local temp_file=$(mktemp)

    # Download with HTTP code
    http_code=$(curl -sL -w "%{http_code}" -o "$temp_file" "$url")

    # Validate the download
    if [ "$http_code" != "200" ]; then
        echo "✗ $output (HTTP $http_code)"
        rm -f "$temp_file"
        return 1
    fi

    # Check it is not a disguised 404 page
    if head -1 "$temp_file" | grep -qE "^404|^<!DOCTYPE|^<html"; then
        echo "✗ $output (invalid content)"
        rm -f "$temp_file"
        return 1
    fi

    # Check the file is not empty
    if [ ! -s "$temp_file" ]; then
        echo "✗ $output (empty)"
        rm -f "$temp_file"
        return 1
    fi

    # All OK, move the file
    mv "$temp_file" "$output"
    echo "✓ $output"
    return 0
}
```

---

## Phase 5.0: Synthesize (Apply Updates)

### 5.1: Download components

**IMPORTANT**: Use `safe_download` for each file.

#### Hooks (scripts)

```bash
BASE="https://raw.githubusercontent.com/kodflow/devcontainer-template/main"
API="https://api.github.com/repos/kodflow/devcontainer-template/contents"

# Discover and download scripts via API (zsh-compatible)
curl -sL "$API/.devcontainer/images/.claude/scripts" | jq -r '.[].name' | grep '\.sh$' \
| while IFS= read -r script; do
    [ -z "$script" ] && continue
    safe_download \
        "$BASE/.devcontainer/images/.claude/scripts/$script" \
        ".devcontainer/images/.claude/scripts/$script" \
    && chmod +x ".devcontainer/images/.claude/scripts/$script"
done
```

#### Commands

```bash
# Discover and download commands via API (zsh-compatible)
curl -sL "$API/.devcontainer/images/.claude/commands" | jq -r '.[].name' | grep '\.md$' \
| while IFS= read -r cmd; do
    [ -z "$cmd" ] && continue
    safe_download \
        "$BASE/.devcontainer/images/.claude/commands/$cmd" \
        ".devcontainer/images/.claude/commands/$cmd"
done
```

#### Agents

```bash
# Discover and download agents via API (zsh-compatible)
mkdir -p ".devcontainer/images/.claude/agents"
curl -sL "$API/.devcontainer/images/.claude/agents" | jq -r '.[].name' | grep '\.md$' \
| while IFS= read -r agent; do
    [ -z "$agent" ] && continue
    safe_download \
        "$BASE/.devcontainer/images/.claude/agents/$agent" \
        ".devcontainer/images/.claude/agents/$agent"
done
```

#### Lifecycle Hooks

```bash
# Discover and download lifecycle hooks via API (zsh-compatible)
mkdir -p ".devcontainer/hooks/lifecycle"
curl -sL "$API/.devcontainer/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$' \
| while IFS= read -r hook; do
    [ -z "$hook" ] && continue
    safe_download \
        "$BASE/.devcontainer/hooks/lifecycle/$hook" \
        ".devcontainer/hooks/lifecycle/$hook" \
    && chmod +x ".devcontainer/hooks/lifecycle/$hook"
done
```

#### Image-Embedded Hooks

```bash
# Discover image hooks via API (recursive: shared/ + lifecycle/)
mkdir -p ".devcontainer/images/hooks/shared" ".devcontainer/images/hooks/lifecycle"

# shared/utils.sh
safe_download \
    "$BASE/.devcontainer/images/hooks/shared/utils.sh" \
    ".devcontainer/images/hooks/shared/utils.sh" \
&& chmod +x ".devcontainer/images/hooks/shared/utils.sh"

# lifecycle hooks (zsh-compatible)
curl -sL "$API/.devcontainer/images/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$' \
| while IFS= read -r hook; do
    [ -z "$hook" ] && continue
    safe_download \
        "$BASE/.devcontainer/images/hooks/lifecycle/$hook" \
        ".devcontainer/images/hooks/lifecycle/$hook" \
    && chmod +x ".devcontainer/images/hooks/lifecycle/$hook"
done
```

#### Shared Utils (workspace copy for initialize.sh)

```bash
# Update workspace utils.sh (needed by initialize.sh on host)
safe_download \
    "$BASE/.devcontainer/hooks/shared/utils.sh" \
    ".devcontainer/hooks/shared/utils.sh"
```

#### Migration: Old Full Hooks → Delegation Stubs

```bash
# Detect old full hooks (hooks without "Delegation stub" marker) and replace with stubs
# Note: literal list works in zsh, no variable expansion needed
for hook in onCreate postCreate postStart postAttach updateContent; do
    hook_file=".devcontainer/hooks/lifecycle/${hook}.sh"
    if [ -f "$hook_file" ] && ! grep -q "Delegation stub" "$hook_file"; then
        echo "  Migrating ${hook}.sh to delegation stub..."
        safe_download \
            "$BASE/.devcontainer/hooks/lifecycle/${hook}.sh" \
            "$hook_file" \
        && chmod +x "$hook_file"
    fi
done
```

#### Config Files (p10k, settings, compose, grepai)

```bash
# p10k
safe_download \
    "$BASE/.devcontainer/images/.p10k.zsh" \
    ".devcontainer/images/.p10k.zsh"

# settings.json
safe_download \
    "$BASE/.devcontainer/images/.claude/settings.json" \
    ".devcontainer/images/.claude/settings.json"

# docker-compose.yml (update devcontainer service, PRESERVE custom services)
# Note: Uses mikefarah/yq (Go version) - simpler syntax with -i for in-place
# Strategy: Start fresh from template, merge back custom services
# Note: Ollama runs on HOST (installed via initialize.sh), not in container
update_compose_services() {
    local compose_file=".devcontainer/docker-compose.yml"
    local temp_template=$(mktemp --suffix=.yaml)
    local temp_custom=$(mktemp --suffix=.yaml)
    local backup_file="${compose_file}.backup"

    # Download template
    if ! curl -sL -o "$temp_template" "$BASE/.devcontainer/docker-compose.yml"; then
        echo "  ✗ Failed to download template"
        rm -f "$temp_template"
        return 1
    fi

    # Validate downloaded template is not empty and contains expected content
    if [ ! -s "$temp_template" ] || ! grep -q "^services:" "$temp_template"; then
        echo "  ✗ Downloaded template is empty or invalid (check network/rate limit)"
        rm -f "$temp_template"
        return 1
    fi

    # Backup original
    cp "$compose_file" "$backup_file"

    # Extract custom services (anything that's NOT devcontainer)
    yq '.services | to_entries | map(select(.key != "devcontainer")) | from_entries' "$compose_file" > "$temp_custom"

    # Start fresh from template (devcontainer service)
    cp "$temp_template" "$compose_file"

    # Merge back custom services if any exist
    if [ -s "$temp_custom" ] && [ "$(yq '. | length' "$temp_custom")" != "0" ]; then
        yq -i ".services *= load(\"$temp_custom\")" "$compose_file"
        echo "    - Preserved custom services"
    fi

    # Cleanup temp files
    rm -f "$temp_template" "$temp_custom"

    # Verify file is not empty and contains required structure
    if [ ! -s "$compose_file" ] || ! grep -q "^services:" "$compose_file"; then
        # File is empty or missing services - restore backup
        mv "$backup_file" "$compose_file"
        echo "  ✗ Result file is empty or invalid, restored backup"
        return 1
    fi

    # Verify YAML is valid and has expected content
    if yq '.services.devcontainer' "$compose_file" > /dev/null 2>&1; then
        rm -f "$backup_file"
        echo "  ✓ docker-compose.yml updated"
        echo "    - REPLACED: devcontainer from template"
        echo "    - PRESERVED: custom services (if any)"
        return 0
    else
        # Restore backup on failure
        mv "$backup_file" "$compose_file"
        echo "  ✗ YAML validation failed (missing devcontainer service), restored backup"
        return 1
    fi
}

if [ ! -f ".devcontainer/docker-compose.yml" ]; then
    # No file exists - download full template
    safe_download \
        "$BASE/.devcontainer/docker-compose.yml" \
        ".devcontainer/docker-compose.yml"
    echo "  ✓ docker-compose.yml created from template"
else
    # File exists - update devcontainer service
    echo "  Updating devcontainer service..."
    update_compose_services
fi

# grepai config (optimized with bge-m3)
safe_download \
    "$BASE/.devcontainer/images/grepai.config.yaml" \
    ".devcontainer/images/grepai.config.yaml"
```

### 5.2: Cleanup deprecated files

```bash
# Remove deprecated configuration files
[ -f ".coderabbit.yaml" ] && rm -f ".coderabbit.yaml" \
    && echo "Removed deprecated .coderabbit.yaml"
```

### 5.3: Update version file

```bash
COMMIT=$(curl -sL "https://api.github.com/repos/kodflow/devcontainer-template/commits/main" | jq -r '.sha[:7]')
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > .devcontainer/.template-version
```

### 5.4: Consolidated report

**Final output:**

```
═══════════════════════════════════════════════
  ✓ DevContainer updated successfully
═══════════════════════════════════════════════

  Template: kodflow/devcontainer-template
  Version : def5678 (2024-01-20)

  Updated components:
    ✓ hooks        (10 scripts)
    ✓ commands     (10 commands)
    ✓ agents       (35 agents)
    ✓ lifecycle    (6 delegation stubs)
    ✓ image-hooks  (6 image-embedded hooks)
    ✓ shared-utils (1 file)
    ✓ p10k         (1 file)
    ✓ settings     (1 file)
    ✓ compose      (devcontainer service updated)
    ✓ grepai       (1 file - bge-m3 config)
    ✓ user-hooks   (synchronized with template)
    ✓ validation   (all scripts exist)

  Grepai config:
    provider: ollama
    model: bge-m3
    endpoint: host.docker.internal:11434 (GPU-accelerated)

  Cleanup:
    ✓ .coderabbit.yaml removed (if existed)

  Note: Restart terminal to apply p10k changes.

═══════════════════════════════════════════════
```

---

## Phase 6.0: Hook Synchronization

**Goal:** Synchronize hooks from `~/.claude/settings.json` with the template.

**Problem solved:** Users with an older `settings.json` may have
references to obsolete scripts (bash-validate.sh, phase-validate.sh, etc.)
because `postStart.sh` only copies `settings.json` if it does not exist.

```yaml
hook_sync_workflow:
  1_backup:
    action: "Backup user settings.json"
    command: "cp ~/.claude/settings.json ~/.claude/settings.json.backup"

  2_merge_hooks:
    action: "Replace the hooks section with the template"
    strategy: "REPLACE (not merge) - the template is the source of truth"
    tool: jq
    preserves:
      - permissions
      - model
      - env
      - statusLine
      - disabledMcpjsonServers

  3_restore_on_failure:
    action: "Restore backup if merge fails"
```

**Implementation:**

```bash
sync_user_hooks() {
    local user_settings="$HOME/.claude/settings.json"
    local template_settings=".devcontainer/images/.claude/settings.json"

    if [ ! -f "$user_settings" ]; then
        echo "  ⚠ No user settings.json, skipping hook sync"
        return 0
    fi

    if [ ! -f "$template_settings" ]; then
        echo "  ✗ Template settings.json not found"
        return 1
    fi

    echo "  Synchronizing user hooks with template..."

    # Backup
    cp "$user_settings" "${user_settings}.backup"

    # Replace hooks section only (preserve all other settings)
    if jq --slurpfile tpl "$template_settings" '.hooks = $tpl[0].hooks' \
       "$user_settings" > "${user_settings}.tmp"; then

        # Validate JSON
        if jq empty "${user_settings}.tmp" 2>/dev/null; then
            mv "${user_settings}.tmp" "$user_settings"
            rm -f "${user_settings}.backup"
            echo "  ✓ User hooks synchronized with template"
            return 0
        else
            mv "${user_settings}.backup" "$user_settings"
            rm -f "${user_settings}.tmp"
            echo "  ✗ Hook merge produced invalid JSON, restored backup"
            return 1
        fi
    else
        mv "${user_settings}.backup" "$user_settings"
        echo "  ✗ Hook merge failed, restored backup"
        return 1
    fi
}
```

---

## Phase 7.0: Script Validation

**Goal:** Validate that all scripts referenced in hooks exist.

```yaml
validate_workflow:
  1_extract:
    action: "Extract all script paths from hooks"
    tool: jq
    pattern: ".hooks | .. | .command? // empty"

  2_verify:
    action: "Verify that each script exists"
    for_each: script_path
    check: "[ -f $script_path ]"

  3_report:
    on_missing: "List missing scripts with fix suggestion"
    on_success: "All scripts validated"
```

**Implementation:**

```bash
validate_hook_scripts() {
    local settings_file="$HOME/.claude/settings.json"
    local scripts_dir="$HOME/.claude/scripts"
    local missing_count=0

    if [ ! -f "$settings_file" ]; then
        echo "  ⚠ No settings.json to validate"
        return 0
    fi

    # Extract all script paths from hooks
    local scripts
    scripts=$(jq -r '.hooks | .. | .command? // empty' "$settings_file" 2>/dev/null \
        | grep -oE '/home/vscode/.claude/scripts/[^ "]+' \
        | sed 's/ .*//' \
        | sort -u)

    if [ -z "$scripts" ]; then
        echo "  ⚠ No hook scripts found in settings.json"
        return 0
    fi

    echo "  Validating hook scripts..."

    # Use while read for zsh compatibility (for x in $VAR breaks in zsh)
    echo "$scripts" | while IFS= read -r script_path; do
        [ -z "$script_path" ] && continue
        local script_name=$(basename "$script_path")

        if [ -f "$script_path" ]; then
            echo "    ✓ $script_name"
        else
            echo "    ✗ $script_name (MISSING)"
            missing_count=$((missing_count + 1))
        fi
    done

    if [ $missing_count -gt 0 ]; then
        echo ""
        echo "  ⚠ $missing_count missing script(s) detected!"
        echo "  → Run: /update --component hooks"
        return 1
    fi

    echo "  ✓ All hook scripts validated"
    return 0
}
```

---

## Guardrails (ABSOLUTE)

| Action | Status | Reason |
|--------|--------|--------|
| Use hardcoded lists | **FORBIDDEN** | API-FIRST MANDATORY |
| Write without validation | **FORBIDDEN** | Corruption risk |
| Skip HTTP verification | **FORBIDDEN** | 404 files possible |
| Non-official source | **FORBIDDEN** | Security |
| Hook sync without backup | **FORBIDDEN** | Always backup first |
| Delete user settings | **FORBIDDEN** | Only merge hooks |
| Skip script validation | **FORBIDDEN** | Error detection MANDATORY |
| `for x in $VAR` pattern | **FORBIDDEN** | Breaks in zsh ($SHELL=zsh) |
| Inline execution without bash | **FORBIDDEN** | Always `bash /tmp/script.sh` |

---

## Affected files

**Updated by /update:**
```
.devcontainer/
├── docker-compose.yml            # Update devcontainer service
├── hooks/
│   ├── lifecycle/*.sh            # Delegation stubs
│   └── shared/utils.sh          # Shared utilities (host)
├── images/
│   ├── .p10k.zsh
│   ├── grepai.config.yaml       # grepai config (provider, model)
│   ├── hooks/                    # Image-embedded hooks (real logic)
│   │   ├── shared/utils.sh
│   │   └── lifecycle/*.sh
│   └── .claude/
│       ├── agents/*.md
│       ├── commands/*.md
│       ├── scripts/*.sh
│       └── settings.json
└── .template-version
```

**In the Docker image (restored at startup):**
```
/etc/grepai/config.yaml            # GrepAI config template
/etc/mcp/mcp.json.tpl              # MCP template
/etc/claude-defaults/*             # Claude defaults
```

**NEVER modified:**
```
.devcontainer/
├── devcontainer.json      # Project config (customizations)
└── Dockerfile             # Image customizations
```

---

## Complete script (reference)

**IMPORTANT: This script uses `#!/bin/bash`. Always write to a temp file and execute with `bash`:**
```bash
cat > /tmp/update-devcontainer.sh << 'SCRIPT'
# ... (script below) ...
SCRIPT
bash /tmp/update-devcontainer.sh && rm -f /tmp/update-devcontainer.sh
```

```bash
#!/bin/bash
# /update implementation - API-FIRST with validation + Environment Detection
# NOTE: Must be executed with bash (not zsh) due to word splitting in for loops.
# If running from Claude Code (zsh), write to temp file first: bash /tmp/script.sh

set -uo pipefail
set +H 2>/dev/null || true  # Disable bash history expansion (! in YAML causes errors)

BASE="https://raw.githubusercontent.com/kodflow/devcontainer-template/main"
API="https://api.github.com/repos/kodflow/devcontainer-template/contents"

# Environment detection function (Phase 0)
detect_context() {
    # Check if running inside container
    if [ -f /.dockerenv ]; then
        CONTEXT="container"
        UPDATE_TARGET="/workspace/.devcontainer/images/.claude"
        echo "Detected: Container environment"
    else
        CONTEXT="host"
        UPDATE_TARGET="$HOME/.claude"
        echo "Detected: Host machine"
    fi

    # Additional checks
    if [ -n "${DEVCONTAINER:-}" ]; then
        echo "  (DevContainer detected via DEVCONTAINER env var)"
    fi

    echo "Update target: $UPDATE_TARGET"
    echo "Mode: $CONTEXT"
}

# Safe download function
safe_download() {
    local url="$1"
    local output="$2"
    local temp_file=$(mktemp)

    http_code=$(curl -sL -w "%{http_code}" -o "$temp_file" "$url")

    if [ "$http_code" != "200" ]; then
        echo "✗ $(basename "$output") (HTTP $http_code)"
        rm -f "$temp_file"
        return 1
    fi

    if head -1 "$temp_file" | grep -qE "^404|^<!DOCTYPE|^<html"; then
        echo "✗ $(basename "$output") (invalid content)"
        rm -f "$temp_file"
        return 1
    fi

    if [ ! -s "$temp_file" ]; then
        echo "✗ $(basename "$output") (empty)"
        rm -f "$temp_file"
        return 1
    fi

    mkdir -p "$(dirname "$output")"
    mv "$temp_file" "$output"
    echo "✓ $(basename "$output")"
    return 0
}

# Hook synchronization function (Phase 5)
sync_user_hooks() {
    local user_settings="$HOME/.claude/settings.json"
    local template_settings=".devcontainer/images/.claude/settings.json"

    if [ ! -f "$user_settings" ]; then
        echo "  ⚠ No user settings.json, skipping hook sync"
        return 0
    fi

    if [ ! -f "$template_settings" ]; then
        echo "  ✗ Template settings.json not found"
        return 1
    fi

    echo "  Synchronizing user hooks with template..."
    cp "$user_settings" "${user_settings}.backup"

    if jq --slurpfile tpl "$template_settings" '.hooks = $tpl[0].hooks' \
       "$user_settings" > "${user_settings}.tmp"; then
        if jq empty "${user_settings}.tmp" 2>/dev/null; then
            mv "${user_settings}.tmp" "$user_settings"
            rm -f "${user_settings}.backup"
            echo "  ✓ User hooks synchronized"
            return 0
        else
            mv "${user_settings}.backup" "$user_settings"
            rm -f "${user_settings}.tmp"
            echo "  ✗ Invalid JSON, restored backup"
            return 1
        fi
    else
        mv "${user_settings}.backup" "$user_settings"
        echo "  ✗ Hook merge failed, restored backup"
        return 1
    fi
}

# Script validation function (Phase 6)
validate_hook_scripts() {
    local settings_file="$HOME/.claude/settings.json"
    local missing_count=0

    if [ ! -f "$settings_file" ]; then
        echo "  ⚠ No settings.json to validate"
        return 0
    fi

    local scripts
    scripts=$(jq -r '.hooks | .. | .command? // empty' "$settings_file" 2>/dev/null \
        | grep -oE '/home/vscode/.claude/scripts/[^ "]+' \
        | sed 's/ .*//' | sort -u)

    if [ -z "$scripts" ]; then
        echo "  ⚠ No hook scripts found"
        return 0
    fi

    echo "  Validating hook scripts..."
    while IFS= read -r script_path; do
        [ -z "$script_path" ] && continue
        local script_name=$(basename "$script_path")
        if [ -f "$script_path" ]; then
            echo "    ✓ $script_name"
        else
            echo "    ✗ $script_name (MISSING)"
            missing_count=$((missing_count + 1))
        fi
    done <<< "$scripts"

    if [ $missing_count -gt 0 ]; then
        echo "  ⚠ $missing_count missing script(s)!"
        echo "  → Run: /update --component hooks"
        return 1
    fi

    echo "  ✓ All scripts validated"
    return 0
}

echo "═══════════════════════════════════════════════"
echo "  /update - DevContainer Environment Update"
echo "═══════════════════════════════════════════════"
echo ""

# Phase 1.0: Environment Detection
detect_context
echo ""

# Hooks
echo "Updating hooks..."
hook_count=0
while IFS= read -r script; do
    [ -z "$script" ] && continue
    if safe_download "$BASE/.devcontainer/images/.claude/scripts/$script" \
                     "$UPDATE_TARGET/scripts/$script"; then
        chmod +x "$UPDATE_TARGET/scripts/$script"
        hook_count=$((hook_count + 1))
    fi
done < <(curl -sL "$API/.devcontainer/images/.claude/scripts" | jq -r '.[].name' | grep '\.sh$')
echo "  ($hook_count scripts)"

# Commands
echo ""
echo "Updating commands..."
cmd_count=0
while IFS= read -r cmd; do
    [ -z "$cmd" ] && continue
    if safe_download "$BASE/.devcontainer/images/.claude/commands/$cmd" \
                     "$UPDATE_TARGET/commands/$cmd"; then
        cmd_count=$((cmd_count + 1))
    fi
done < <(curl -sL "$API/.devcontainer/images/.claude/commands" | jq -r '.[].name' | grep '\.md$')
echo "  ($cmd_count commands)"

# Agents
echo ""
echo "Updating agents..."
mkdir -p "$UPDATE_TARGET/agents"
agent_count=0
while IFS= read -r agent; do
    [ -z "$agent" ] && continue
    if safe_download "$BASE/.devcontainer/images/.claude/agents/$agent" \
                     "$UPDATE_TARGET/agents/$agent"; then
        agent_count=$((agent_count + 1))
    fi
done < <(curl -sL "$API/.devcontainer/images/.claude/agents" | jq -r '.[].name' | grep '\.md$')
echo "  ($agent_count agents)"

# Lifecycle stubs (only in container mode - skip on host)
if [ "$CONTEXT" = "container" ]; then
    echo ""
    echo "Updating lifecycle hooks (delegation stubs)..."
    mkdir -p ".devcontainer/hooks/lifecycle"
    lifecycle_count=0
    while IFS= read -r hook; do
        [ -z "$hook" ] && continue
        if safe_download "$BASE/.devcontainer/hooks/lifecycle/$hook" \
                         ".devcontainer/hooks/lifecycle/$hook"; then
            chmod +x ".devcontainer/hooks/lifecycle/$hook"
            lifecycle_count=$((lifecycle_count + 1))
        fi
    done < <(curl -sL "$API/.devcontainer/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$')
    echo "  ($lifecycle_count stubs)"

    # Image-embedded hooks (real logic)
    echo ""
    echo "Updating image-embedded hooks..."
    mkdir -p ".devcontainer/images/hooks/shared" ".devcontainer/images/hooks/lifecycle"
    safe_download "$BASE/.devcontainer/images/hooks/shared/utils.sh" \
                  ".devcontainer/images/hooks/shared/utils.sh" \
    && chmod +x ".devcontainer/images/hooks/shared/utils.sh"
    while IFS= read -r hook; do
        [ -z "$hook" ] && continue
        safe_download "$BASE/.devcontainer/images/hooks/lifecycle/$hook" \
                      ".devcontainer/images/hooks/lifecycle/$hook" \
        && chmod +x ".devcontainer/images/hooks/lifecycle/$hook"
    done < <(curl -sL "$API/.devcontainer/images/hooks/lifecycle" | jq -r '.[].name' | grep '\.sh$')

    # Shared utils (workspace copy for initialize.sh on host)
    echo ""
    echo "Updating shared utilities..."
    safe_download "$BASE/.devcontainer/hooks/shared/utils.sh" \
                  ".devcontainer/hooks/shared/utils.sh"

    # Migration: detect old full hooks and replace with stubs
    for h in onCreate postCreate postStart postAttach updateContent; do
        hook_file=".devcontainer/hooks/lifecycle/${h}.sh"
        if [ -f "$hook_file" ] && ! grep -q "Delegation stub" "$hook_file"; then
            echo "  Migrating ${h}.sh to delegation stub..."
            safe_download "$BASE/.devcontainer/hooks/lifecycle/${h}.sh" "$hook_file" \
            && chmod +x "$hook_file"
        fi
    done
fi

# Config files
echo ""
echo "Updating config files..."
if [ "$CONTEXT" = "container" ]; then
    safe_download "$BASE/.devcontainer/images/.p10k.zsh" ".devcontainer/images/.p10k.zsh"
fi
safe_download "$BASE/.devcontainer/images/.claude/settings.json" "$UPDATE_TARGET/settings.json"

# Docker compose (only in container mode - not applicable on host)
if [ "$CONTEXT" = "container" ]; then
    # Note: Uses mikefarah/yq (Go version) - simpler syntax with -i for in-place
    # Ollama runs on HOST (installed via initialize.sh), not in container
    echo ""
    echo "Updating docker-compose.yml..."

    update_compose_services() {
    local compose_file=".devcontainer/docker-compose.yml"
    local temp_template=$(mktemp --suffix=.yaml)
    local temp_custom=$(mktemp --suffix=.yaml)
    local backup_file="${compose_file}.backup"

    # Download template
    if ! curl -sL -o "$temp_template" "$BASE/.devcontainer/docker-compose.yml"; then
        echo "  ✗ Failed to download template"
        rm -f "$temp_template"
        return 1
    fi

    # Validate downloaded template is not empty and contains expected content
    if [ ! -s "$temp_template" ] || ! grep -q "^services:" "$temp_template"; then
        echo "  ✗ Downloaded template is empty or invalid (check network/rate limit)"
        rm -f "$temp_template"
        return 1
    fi

    # Backup original
    cp "$compose_file" "$backup_file"

    # Extract custom services (anything that's NOT devcontainer)
    yq '.services | to_entries | map(select(.key != "devcontainer")) | from_entries' "$compose_file" > "$temp_custom"

    # Extract custom volumes (anything that's NOT in template)
    local template_volumes=$(yq '.volumes | keys | .[]' "$temp_template" 2>/dev/null | tr '\n' '|')

    # Start fresh from template (devcontainer service)
    cp "$temp_template" "$compose_file"

    # Merge back custom services if any exist
    if [ -s "$temp_custom" ] && [ "$(yq '. | length' "$temp_custom")" != "0" ]; then
        yq -i ".services *= load(\"$temp_custom\")" "$compose_file"
        echo "    - Preserved custom services"
    fi

    # Cleanup temp files
    rm -f "$temp_template" "$temp_custom"

    # Verify file is not empty and contains required structure
    if [ ! -s "$compose_file" ] || ! grep -q "^services:" "$compose_file"; then
        # File is empty or missing services - restore backup
        mv "$backup_file" "$compose_file"
        echo "  ✗ Result file is empty or invalid, restored backup"
        return 1
    fi

    # Verify YAML is valid and has expected content
    if yq '.services.devcontainer' "$compose_file" > /dev/null 2>&1; then
        rm -f "$backup_file"
        echo "  ✓ devcontainer service updated"
        return 0
    else
        mv "$backup_file" "$compose_file"
        echo "  ✗ YAML validation failed (missing devcontainer service), restored backup"
        return 1
    fi
}

    if [ ! -f ".devcontainer/docker-compose.yml" ]; then
        echo "  No docker-compose.yml found, downloading template..."
        safe_download "$BASE/.devcontainer/docker-compose.yml" ".devcontainer/docker-compose.yml"
    else
        echo "  Updating devcontainer service..."
        update_compose_services
    fi

    # Grepai config
    echo ""
    echo "Updating grepai config..."
    safe_download "$BASE/.devcontainer/images/grepai.config.yaml" ".devcontainer/images/grepai.config.yaml"
fi  # End container-only updates

# Phase 6.0: Synchronize user hooks (both container and host)
echo ""
echo "Phase 6.0: Synchronizing user hooks..."
sync_user_hooks

# Phase 7.0: Validate hook scripts
echo ""
echo "Phase 7.0: Validating hook scripts..."
validate_hook_scripts

# Version
COMMIT=$(curl -sL "https://api.github.com/repos/kodflow/devcontainer-template/commits/main" | jq -r '.sha[:7]')
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

if [ "$CONTEXT" = "container" ]; then
    echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > .devcontainer/.template-version
else
    echo "{\"commit\": \"$COMMIT\", \"updated\": \"$DATE\"}" > "$UPDATE_TARGET/.template-version"
fi

echo ""
echo "═══════════════════════════════════════════════"
echo "  ✓ Update complete - version: $COMMIT"
echo "  Context: $CONTEXT"
echo "  Target: $UPDATE_TARGET"
echo "═══════════════════════════════════════════════"
```
