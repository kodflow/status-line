<!-- updated: 2026-05-01T17:55:00Z -->
# Development Workflows

## Setup

Prerequisites:

- Go 1.25.5+ (toolchain 1.26)
- `make`
- Optional: `git`, `op` (1Password CLI) — used by
  adapters at runtime; missing tools are gracefully skipped.

Bootstrap:

```bash
git clone https://github.com/kodflow/status-line.git
cd status-line
make build      # → bin/status-line
make demo       # render once with sample input
```

## Development Loop

```
edit → make build → make demo (sanity check) → make test → make lint → commit via /git
```

| Command       | Purpose                                                              |
|---------------|----------------------------------------------------------------------|
| `make build`  | Compile to `bin/status-line` (no install)                            |
| `make test`   | `go test ./...` — all packages, race detector when CI                |
| `make lint`   | ktn-linter via `/lint` (148 rules across 8 phases)                   |
| `make demo`   | Build + pipe sample JSON to render once for visual check             |
| `make run`    | Build + demo (alias)                                                 |
| `make clean`  | Remove `bin/`                                                        |

Use the `/git` skill for commits (conventional format, branch creation,
PR open). Never commit to `main` directly — branch protection enforces
this.

## Testing Strategy

### Unit tests

Convention: `*_test.go` in the same directory as the code under test.

- **`xxx_internal_test.go`** — package `xxx` — white-box tests with
  access to unexported identifiers. Use for invariants, table-driven
  helper tests, and edge-case dispatch.
- **`xxx_external_test.go`** — package `xxx_test` — black-box tests
  against the public API. Use for end-to-end-of-package contracts and
  to prove the API surface is sufficient without leaking internals.

### Adapter tests

Adapters in `internal/adapter/*` are tested via their port interface.
Use fakes from `internal/domain/port` test doubles when exercising
`internal/application/StatusLineService`. Real CLI adapters (`git`,
`task`) get an integration test guarded by build tag if they need a
real binary.

### Render tests

`internal/presentation/renderer` works on string output. Tests assemble
a domain model, call `Render`, and compare the ANSI bytes. Snapshot
tests welcome but only with the snapshot committed and reviewed.

### Coverage target

`make test-coverage` (when present) generates `coverage.out`. Target:
≥ 80 % statement coverage on `internal/` (excluding `cmd/` which is
glue).

## Deployment

Single static binary distribution.

- **Tag a release** — `git tag vX.Y.Z && git push --tags`. The
  `.github/workflows/release.yml` workflow builds for `linux/amd64`,
  `linux/arm64`, `darwin/amd64`, `darwin/arm64`, attaches binaries,
  publishes the release, and dispatches a rebuild event to
  `kodflow/devcontainer-template`.
- **Self-update** — `internal/adapter/updater` polls the GitHub
  Releases API and surfaces an update segment on line 2 when a newer
  version is available. The user runs `status-line --update` (or the
  shipped equivalent) to swap the binary in place.

## CI/CD

`.github/workflows/ci.yml`:

| Stage     | Action                                                  |
|-----------|---------------------------------------------------------|
| `test`    | `make test` on `ubuntu-latest`, Go 1.25.5               |
| `lint`    | `make lint` (ktn-linter)                                |
| `build`   | `make build` — compile sanity                           |

Branch protection on `main` requires:

- All `ci.yml` checks green.
- `dismiss_stale_reviews: true` — when a new commit is pushed,
  CodeRabbit's previous CHANGES_REQUESTED reviews are auto-dismissed
  so the PR can re-evaluate against the new state.

`.github/workflows/release.yml` runs only on tag push and triggers the
downstream devcontainer-template rebuild via `repository_dispatch`
event `status-line-release`.

## Skill-driven workflow

This project uses `kodflow/devcontainer-template` skills:

- `/init` — project bootstrap (run once; you ran it).
- `/feature <name>` — plan + apply a feature in one flow.
- `/fix <bug>` — same flow, scoped to a bug.
- `/git --commit` — branch + conventional commit + PR.
- `/git --merge` — readiness check + squash merge.
- `/lint` — ktn-linter run scoped to changed files.
- `/test` — incremental test loop scoped to changed files.
- `/review` — local multi-agent review against `main`.
- `/update` — sync `.devcontainer/` from upstream template.

Two AI review tools are wired but **label-triggered**:

- **CodeRabbit** — add label `coderabbit` on a PR to request a review;
  no automatic noise on every push.
- **Qodo Merge** — add label `qodo` on a PR to request a Qodo review.
