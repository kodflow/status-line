---
name: init
description: |
  Conversational project discovery + doc generation.
  Open-ended dialogue builds rich context, then synthesizes all project docs.
  Use when: creating new project, starting work, verifying setup.
allowed-tools:
  - Write
  - Edit
  - "Bash(git:*)"
  - "Bash(docker:*)"
  - "Bash(terraform:*)"
  - "Bash(kubectl:*)"
  - "Bash(node:*)"
  - "Bash(python:*)"
  - "Bash(go:*)"
  - "Bash(grepai:*)"
  - "Bash(curl:*)"
  - "Bash(pgrep:*)"
  - "Bash(nohup:*)"
  - "Bash(mkdir:*)"
  - "Bash(rm:*)"
  - "Bash(wc:*)"
  - "Read(**/*)"
  - "Glob(**/*)"
  - "mcp__grepai__*"
  - "Grep(**/*)"
  - "Task(*)"
  - "TaskCreate(*)"
  - "TaskUpdate(*)"
  - "TaskList(*)"
  - "TaskGet(*)"
  - "mcp__github__*"
  - "mcp__codacy__*"
---

# /init - Conversational Project Discovery

$ARGUMENTS

## GREPAI-FIRST (MANDATORY)

Use `grepai_search` for ALL semantic/meaning-based queries BEFORE Grep.
Use `grepai_trace_callers`/`grepai_trace_callees` for impact analysis.
Fallback to Grep ONLY for exact string matches or regex patterns.

---

## Overview

Conversational initialization with **progressive context building**:

1. **Detect** - Template or already personalized?
2. **Discover** - Open-ended conversation to understand the project
3. **Synthesize** - Review accumulated context with user
4. **Generate** - Produce all project docs from rich context
5. **Validate** - Environment, tools, deps, config

---

## Usage

```
/init                # Everything automatic
```

**Intelligent behavior:**
- Detects template → starts discovery conversation
- Detects personalized → skips to validation
- Detects problems → auto-fix when possible
- No flags, no unnecessary questions

---

## Phase 1.0: Detect (Repository Identity → Template vs Personalized)

**Step 1: Identify the repository via git remote.**

```yaml
detect_repository:
  command: "git remote get-url origin 2>/dev/null"
  check: "does the URL contain 'kodflow/devcontainer-template'?"

  decision:
    if_is_devcontainer_template:
      action: "Continue to Step 2 (template marker check)"
      message: "devcontainer-template repo detected."
    if_is_other_project:
      action: "RESET — erase all generated docs, restart Phase 1 from scratch"
      message: "Different project detected. Resetting for fresh initialization."
      reset_files:
        - "/workspace/CLAUDE.md"
        - "/workspace/AGENTS.md"
      reset_directories:
        - "/workspace/docs/"    # rm -rf — template docs don't apply to new projects
      note: "README.md is NOT erased — only its description will be updated in Phase 3"
```

**Step 2 (only for devcontainer-template repo): Check template markers.**

```yaml
detect_template:
  check_markers:
    - file: "/workspace/CLAUDE.md"
      template_marker: "Kodflow DevContainer Template"
    - file: "/workspace/docs/vision.md"
      template_marker: "batteries-included VS Code Dev Container"

  decision:
    if_template_detected:
      action: "Run Phase 1 (Discovery Conversation)"
      message: "Template detected. Let's discover your project."
    if_personalized:
      action: "Skip to Phase 4 (Validation)"
      message: "Project already personalized. Validating..."
```

**Output Phase 0 (other project — reset):**

```
═══════════════════════════════════════════════════════════════
  /init - Project Detection
═══════════════════════════════════════════════════════════════

  Checking: git remote origin
  Result  : {remote_url} (NOT devcontainer-template)

  → Different project detected
  → Resetting docs for fresh initialization...
    ✗ CLAUDE.md        (reset)
    ✗ AGENTS.md        (reset)
    ✗ docs/            (removed)

  → Starting discovery conversation...

═══════════════════════════════════════════════════════════════
```

**Output Phase 0 (devcontainer-template — template markers):**

```
═══════════════════════════════════════════════════════════════
  /init - Project Detection
═══════════════════════════════════════════════════════════════

  Checking: git remote origin
  Result  : kodflow/devcontainer-template

  Checking: /workspace/CLAUDE.md
  Result  : Template markers found

  → Project needs personalization
  → Starting discovery conversation...

═══════════════════════════════════════════════════════════════
```

---

## Phase 2.0: Discovery Conversation

**RULES (ABSOLUTE):**

- Ask **ONE question at a time** as plain text output
- **NEVER** use AskUserQuestion tool
- **NEVER** offer predefined options or multiple-choice lists
- After **EACH** user response, display the updated **Project Context** block
- Adapt the next question based on accumulated context
- Minimum **4** exchanges, maximum **10**
- Questions must be open-ended and conversational

### Question Strategy

**Fixed questions (always asked first):**

```yaml
round_1:
  question: |
    Tell me about your project. What are you building
    and what problem does it solve?
  extracts: [purpose, problem]

round_2:
  question: |
    Who will use this? Describe the people or systems
    that will interact with it.
  extracts: [users]

round_3:
  question: |
    What should we call this project?
  extracts: [name]
```

**Adaptive questions (selected based on gaps in context):**

```yaml
adaptive_pool:
  tech_stack:
    trigger: "tech stack unknown"
    question: "What languages, frameworks, or tools are you planning to use?"
    extracts: [tech_stack]

  data_storage:
    trigger: "data storage relevant AND unknown"
    question: "How will your project store and manage data?"
    extracts: [database]

  deployment:
    trigger: "deployment unknown"
    question: "Where and how will this run in production?"
    extracts: [deployment]

  quality:
    trigger: "quality priorities unknown"
    question: "What matters most for quality — test coverage, performance, security, or something else?"
    extracts: [quality]

  constraints:
    trigger: "constraints unknown"
    question: "Are there any constraints I should know about — team size, timeline, compliance requirements?"
    extracts: [constraints]

  architecture:
    trigger: "complex project AND architecture unclear"
    question: "Do you have a particular architecture in mind — monolith, microservices, event-driven, or something else?"
    extracts: [architecture]

  follow_up:
    trigger: "previous answer was brief"
    question: "Can you tell me more about {topic}? I want to make sure I capture the full picture."
    extracts: [varies]
```

### Project Context Block

**Display this block after EVERY exchange, updated with new information:**

```
═════════════════════════════════════════════════════
  PROJECT CONTEXT
═════════════════════════════════════════════════════
  Name        : {name or "---"}
  Purpose     : {1-2 sentence summary or "---"}
  Problem     : {problem statement or "---"}
  Users       : {target users or "---"}
  Tech Stack  : {languages, frameworks or "---"}
  Database    : {database choices or "---"}
  Deployment  : {cloud/hosting or "---"}
  Architecture: {architecture approach or "---"}
  Quality     : {quality priorities or "---"}
  Constraints : {known constraints or "---"}
  [Discovery — exchange {N}/10]
═════════════════════════════════════════════════════
```

### Transition Criteria

Move to Phase 2 when **ALL** of these are true:

- Name is known
- Purpose/Problem is known
- Users are known
- At least one tech element is concrete
- At least 4 exchanges completed

**OR:** User signals readiness / 10 exchanges reached.

---

## Phase 3.0: Vision Synthesis

**Review the accumulated context with the user before generating files.**

```yaml
synthesis_workflow:
  step_1:
    action: "Display FINAL Project Context with all fields populated"
    output: |
      ═════════════════════════════════════════════════════
        FINAL PROJECT CONTEXT
      ═════════════════════════════════════════════════════
        Name        : {name}
        Purpose     : {purpose}
        Problem     : {problem}
        Users       : {users}
        Tech Stack  : {tech_stack}
        Database    : {database}
        Deployment  : {deployment}
        Architecture: {architecture}
        Quality     : {quality}
        Constraints : {constraints}
      ═════════════════════════════════════════════════════

  step_2:
    message: |
      Here is what I understand about your project.
      Review and tell me if anything needs to change.
      Say "generate" when you're ready for me to create
      your project documentation.

  step_3:
    loop: "Process any refinements, update context, repeat"
    exit: "User says 'generate' or confirms"
```

---

## Phase 4.0: File Generation

**Generate all files DIRECTLY from accumulated context. No templates.**

```yaml
generation_rules:
  - NO mustache/handlebars placeholders
  - NO template files referenced
  - Content is SYNTHESIZED from the full conversation context
  - Every file must contain real, specific, actionable content
  - Write vision.md FIRST, then remaining files in parallel
```

### Files to Generate

```yaml
files:
  # PRIMARY OUTPUT - written first
  - path: "/workspace/docs/vision.md"
    description: "Rich project vision synthesized from conversation"
    structure:
      - "# Vision: {name}"
      - "## Purpose — what and why"
      - "## Problem Statement — pain points addressed"
      - "## Target Users — who benefits and how"
      - "## Goals — prioritized list"
      - "## Success Criteria — measurable targets table"
      - "## Design Principles — guiding decisions"
      - "## Non-Goals — explicit exclusions"
      - "## Key Decisions — tech choices with rationale"

  # SUPPORTING FILES - written in parallel after vision.md
  - path: "/workspace/CLAUDE.md"
    description: "Project overview, tech stack, how to work"
    structure:
      - "# {name}"
      - "## Purpose — 2-3 sentences"
      - "## Tech Stack — languages, frameworks, databases"
      - "## How to Work — /init, /feature, /fix"
      - "## Key Principles — MCP-first, semantic search, specialists"
      - "## Verification — test, lint, security commands"
      - "## Documentation — links to vision, architecture, workflows"

  - path: "/workspace/AGENTS.md"
    description: "Map tech stack to available specialist agents"
    structure:
      - "# Specialist Agents"
      - "## Primary — agents matching tech stack"
      - "## Supporting — review, devops, security agents"
      - "## Usage — when to invoke each agent"

  - path: "/workspace/docs/architecture.md"
    description: "System context, components, data flow"
    structure:
      - "# Architecture: {name}"
      - "## System Context — high-level view"
      - "## Components — key modules/services"
      - "## Data Flow — how data moves"
      - "## Technology Stack — detailed breakdown"
      - "## Constraints — technical boundaries"

  - path: "/workspace/docs/workflows.md"
    description: "Development processes adapted to tech stack"
    structure:
      - "# Development Workflows"
      - "## Setup — prerequisites, installation"
      - "## Development Loop — code, test, commit"
      - "## Testing Strategy — unit, integration, e2e"
      - "## Deployment — build, release process"
      - "## CI/CD — pipeline stages"

  - path: "/workspace/README.md"
    description: "Update description section only, preserve existing structure"
    mode: "edit"
    note: "Only update the project description. Keep all other content."

  # CONDITIONAL FILES
  - path: "/workspace/.env.example"
    condition: "database OR cloud services mentioned"
    description: "Environment variable template"
    structure:
      - "# {name} Environment Variables"
      - "APP_NAME={name}"
      - "# Database, cloud, API vars as relevant"

  - path: "/workspace/Makefile"
    condition: "language with build tooling (Go, Rust, Python, Node)"
    description: "Build targets adapted to tech stack"
    structure:
      - "# {name} targets"
      - "Standard targets: build, test, lint, fmt, clean"
      - "Language-specific targets as relevant"
```

---

## Phase 5.0: Environment Validation

**Verify the environment (parallel via Task agents).**

```yaml
parallel_checks:
  agents:
    - name: "tools-checker"
      checks: [git, node, go, terraform, docker, grepai]
      output: "{tool, required, installed, status}"

    - name: "deps-checker"
      checks: [npm ci, go mod, terraform init]
      output: "{manager, status, issues}"

    - name: "config-checker"
      checks: [.env, CLAUDE.md, mcp.json]
      output: "{file, status, issue}"

    - name: "grepai-checker"
      checks: [Ollama, daemon, index]
      output: "{component, status, details}"

    - name: "secret-checker"
      checks: [op CLI, OP_SERVICE_ACCOUNT_TOKEN, vault access, project secrets]
      output: "{op_installed, token_set, vault_name, project_path, secrets_count, status}"
```

---

## Phase 6.0: Report

```
═══════════════════════════════════════════════════════════════
  /init - Complete
═══════════════════════════════════════════════════════════════

  Project: {name}
  Purpose: {purpose summary}

  Generated:
    ✓ docs/vision.md
    ✓ CLAUDE.md
    ✓ AGENTS.md
    ✓ docs/architecture.md
    ✓ docs/workflows.md
    ✓ README.md (updated)
    {conditional files}

  Environment:
    ✓ Tools installed ({tool list})
    ✓ Dependencies ready
    ✓ grepai indexed ({N} files)

  1Password:
    ✓ op CLI installed
    ✓ Vault connected ({N} project secrets)

  Ready to develop!
    → /feature "description" to start

═══════════════════════════════════════════════════════════════
```

---

## Phase 7.0: GrepAI Calibration

**MANDATORY** after project discovery. Calibrate grepai config based on project size and structure.

```yaml
grepai_calibration:
  1_count_files:
    command: |
      find /workspace -type f \
        -not -path '*/.git/*' -not -path '*/node_modules/*' \
        -not -path '*/vendor/*' -not -path '*/.grepai/*' \
        -not -path '*/__pycache__/*' -not -path '*/target/*' \
        -not -path '*/.venv/*' -not -path '*/dist/*' | wc -l
    output: file_count

  2_select_profile:
    rules:
      - "file_count < 10000   → profile: small"
      - "file_count < 100000  → profile: medium"
      - "file_count < 500000  → profile: large"
      - "file_count >= 500000 → profile: massive"

    profiles:
      small:
        chunking: { size: 1024, overlap: 100 }
        hybrid: { enabled: true, k: 60 }
        debounce_ms: 1000
      medium:
        chunking: { size: 1024, overlap: 100 }
        hybrid: { enabled: true, k: 60 }
        debounce_ms: 2000
      large:
        chunking: { size: 512, overlap: 50 }
        hybrid: { enabled: true, k: 60 }
        debounce_ms: 3000
      massive:
        chunking: { size: 512, overlap: 50 }
        hybrid: { enabled: false }
        debounce_ms: 5000

  3_detect_languages:
    action: "Scan for go.mod, package.json, Cargo.toml, etc."
    output: "Filter trace.enabled_languages to only detected languages"

  4_customize_boost:
    action: |
      Scan project structure (ls -d */):
      - If src/ exists → bonus /src/ 1.2
      - If pkg/ exists → bonus /pkg/ 1.15
      - If internal/ exists → bonus /internal/ 1.1
      - If app/ exists → bonus /app/ 1.15
      - If lib/ exists → bonus /lib/ 1.15
      Add project-specific ignore patterns (e.g., .next/, .nuxt/, .angular/)

  5_write_config:
    action: "Generate .grepai/config.yaml with selected profile"
    template: "/etc/grepai/config.yaml (base) + profile overrides"

  6_restart_daemon:
    action: |
      pkill -f 'grepai watch' 2>/dev/null || true
      rm -f /workspace/.grepai/index.gob /workspace/.grepai/symbols.gob
      nohup grepai watch >/tmp/grepai.log 2>&1 &
      sleep 3
      grepai status
```

**Output Phase 6:**

```
═══════════════════════════════════════════════════════════════
  GrepAI Calibration
═══════════════════════════════════════════════════════════════

  Files detected : 47,230
  Profile        : medium
  Model          : bge-m3 (1024d, 72% accuracy)

  Config applied:
    chunking    : 1024 tokens / 100 overlap
    hybrid      : ON (k=60)
    debounce    : 2000ms
    languages   : .go, .ts, .py (3 detected)

  Boost customized:
    +1.2  /src/
    +1.15 /pkg/
    +1.1  /internal/

  Daemon: restarted (indexing 47,230 files...)

═══════════════════════════════════════════════════════════════
```

---

## Auto-fix (automatic)

When a problem is detected, auto-fix if possible:

| Problem | Auto Action |
|---------|-------------|
| `.env` missing | `cp .env.example .env` |
| deps not installed | `npm ci` / `go mod download` |
| grepai not running | `nohup grepai watch &` |
| Ollama not reachable | Display HOST instructions |
| grepai uncalibrated | Run Phase 6 calibration |

---

## Guardrails

| Action | Status |
|--------|--------|
| Skip detection | FORBIDDEN |
| Closed questions / AskUserQuestion | FORBIDDEN |
| Placeholders in generated files | FORBIDDEN |
| Skip vision synthesis review | FORBIDDEN |
| Destructive fix without asking | FORBIDDEN |
