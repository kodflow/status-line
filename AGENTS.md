<!-- updated: 2026-05-01T17:55:00Z -->
# Specialist Agents

Status Line is a Go CLI with shell tooling. The agents below are the
ones to invoke from this project.

## Primary

| Agent                              | When to invoke                                                                  |
|------------------------------------|---------------------------------------------------------------------------------|
| `developer-specialist-go`          | Anything in `cmd/` or `internal/` — idiomatic Go, error handling, concurrency, golangci-lint. |
| `developer-specialist-review`      | Code review pass over a branch (correctness + design + quality + security + shell). |
| `developer-executor-correctness`   | Sub-executor for invariants, off-by-one, state-machine bugs, error surfacing.   |
| `developer-executor-design`        | Sub-executor for hexagonal layering — verifies adapters don't leak into domain. |
| `developer-executor-quality`       | Sub-executor for complexity, code smells, ktn-linter pre-checks.                |
| `developer-executor-shell`         | Anything in `.devcontainer/images/.claude/scripts/`, `Makefile`, CI YAML.       |

## Supporting

| Agent                              | When to invoke                                                                  |
|------------------------------------|---------------------------------------------------------------------------------|
| `developer-executor-security`      | Before every release — secrets scan, supply-chain (`go.sum`), CWE patterns.     |
| `devops-specialist-docker`         | If you ever touch the project Dockerfile (currently only the devcontainer).     |
| `devops-orchestrator`              | CI/CD changes — `.github/workflows/{ci,release}.yml`.                           |
| `developer-orchestrator`           | Multi-package refactors that need parallel work (e.g. adding a new adapter).    |

## Usage

The `/review` skill auto-dispatches the executor agents in parallel
(correctness, security, design, quality, shell). You only invoke them
manually when you need a single perspective without the full review
pipeline.

The `/lint` skill is the canonical entry for linting — it routes to
`developer-specialist-go` for `cmd/` and `internal/`, and to
`developer-executor-shell` for shell scripts.

When adding a new external integration (e.g. a Linear adapter), use
`developer-specialist-go` for the adapter implementation and
`developer-executor-design` to verify the hexagonal boundary
(no Linear types leak into `internal/domain/`).
