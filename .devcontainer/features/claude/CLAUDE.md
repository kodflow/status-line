# Kodflow DevContainer Template

## Project Structure (MANDATORY)

```
/workspace
├── src/                    # ALL source code (mandatory)
│   ├── components/
│   ├── services/
│   └── ...
├── tests/                  # Unit tests (optional, not for Go)
├── docs/                   # Documentation
└── CLAUDE.md
```

**Rules:**
- ALL code MUST be in `/src` regardless of language
- Tests in `/tests` (except Go: tests alongside code in `/src`)
- Never put code at project root

## Language Rules

**STRICT**: Follow rules in `.devcontainer/features/languages/<lang>/RULES.md`

Each RULES.md contains:
1. **Line 1**: Required version (NEVER downgrade)
2. Code style and conventions
3. Project structure requirements
4. Testing standards

## Workflow (MANDATORY)

### 1. Context Generation
```
/build --context
```
Generates CLAUDE.md in all subdirectories + fetches latest language versions.

### 2. Feature Development
```
/feature <description>
```
Creates `feat/<description>` branch, **mandatory planning mode**, CI check, PR creation (no auto-merge).

### 3. Bug Fixes
```
/fix <description>
```
Creates `fix/<description>` branch, **mandatory planning mode**, CI check, PR creation (no auto-merge).

**Flow:**
```
/build --context → /feature "..." ou /fix "..."
```

## Branch Conventions

| Type | Branch | Commit prefix |
|------|--------|---------------|
| Feature | `feat/<desc>` | `feat(scope): message` |
| Bugfix | `fix/<desc>` | `fix(scope): message` |

## Code Quality

- Latest stable version ONLY (see RULES.md)
- No deprecated APIs
- No legacy patterns
- Security-first approach
- Full test coverage

## SAFEGUARDS (ABSOLUTE - NO BYPASS)

**NEVER without EXPLICIT user approval:**
- Delete files in `.claude/` directory
- Delete files in `.devcontainer/` directory
- Modify `.claude/commands/*.md` destructively (removing features/logic)
- Remove hooks from `.devcontainer/hooks/`

**When simplifying/refactoring:**
- Move content to separate files, NEVER delete logic
- Ask before removing any feature, even if it seems redundant

## Hooks (Auto-applied)

| Hook | Action |
|------|--------|
| `pre-validate.sh` | Protect sensitive files |
| `post-edit.sh` | Format + Imports + Lint |
| `security.sh` | Secret detection |
| `test.sh` | Run related tests |

## Context Hierarchy

```
/CLAUDE.md              → Overview (committed)
/src/CLAUDE.md          → src details (gitignored)
/src/api/CLAUDE.md      → API details (gitignored)
```

**Principle:** More details deeper in tree, <60 lines each.
