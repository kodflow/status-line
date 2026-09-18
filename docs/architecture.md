<!-- updated: 2026-05-01T17:55:00Z -->
# Architecture: Status Line

## System Context

```
┌────────────────────┐
│   Claude Code      │  emits {model, cwd, ...} JSON on every render
│  (status-line hook)│
└─────────┬──────────┘
          │ stdin
          ▼
┌─────────────────────────────────────────────────────┐
│ cmd/statusline/                                     │
│   parses input → invokes application service        │
└─────────┬───────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────┐
│ internal/application/StatusLineService              │
│   orchestrates port calls, builds domain model      │
└─────┬─────────────┬──────────────┬──────────────────┘
      │             │              │
      ▼             ▼              ▼
┌──────────┐ ┌──────────┐  ┌──────────────────┐
│ git      │ │ taskw.   │  │ usage (Anthropic │
│ adapter  │ │ adapter  │  │  OAuth API)      │
└──────────┘ └──────────┘  └──────────────────┘
      │             │              │
      └────────┬────┴──────────────┘
               ▼
      ┌────────────────────────────────────┐
      │ internal/presentation/renderer     │
      │   builds Powerline ANSI segments   │
      └─────────┬──────────────────────────┘
                │ stdout
                ▼
         ANSI string back to Claude Code
```

## Components

### `cmd/statusline/`

Entry point. Responsibilities:

- Parse stdin JSON (Claude Code's `Input` schema).
- Construct adapters (default wiring; tests override with fakes).
- Invoke `application.StatusLineService.Render(ctx, input)`.
- Write the resulting ANSI string to stdout.
- Exit 0 always — never propagate adapter errors to Claude Code.

### `internal/domain/`

Pure data types and interfaces. No I/O, no concurrency.

- `model/` — `Input`, `Progress`, `Usage`, `Git`, `MCPServer`,
  `OSInfo`, `TerminalInfo`, etc. All value types.
- `port/` — `InputProvider`, `Renderer`, `GitRepository`,
  `MCPDetector`, `SystemInfoProvider`,
  `TerminalDetector`, `Updater`, `UsageProvider`. Every external
  collaborator the application service needs.

### `internal/application/`

`StatusLineService` orchestrates the ports to build the domain model
that the renderer consumes. Steps:

1. Run port calls in parallel where possible (git + MCP + system + usage).
2. Tolerate per-port errors (skip that segment).
3. Pass the assembled domain model to the renderer.

### `internal/adapter/`

One adapter per external concern. Each implements one port from
`internal/domain/port` and contains its own integration logic.

| Adapter        | Port implemented      | What it talks to                                |
|----------------|-----------------------|-------------------------------------------------|
| `git`          | `GitRepository`       | `git` CLI (status, diff stats)                  |
| `mcp`          | `MCPDetector`         | `~/.claude/.claude.json`, project `mcp.json`    |
| `system`       | `SystemInfoProvider`  | `runtime.GOOS`, `/.dockerenv`                   |
| `terminal`     | `TerminalDetector`    | `golang.org/x/term` (width, color depth)        |
| `updater`      | `Updater`             | GitHub Releases API                             |
| `usage`        | `UsageProvider`       | Anthropic OAuth API (`five_hour.utilization`)   |

### `internal/presentation/renderer/`

Pure rendering layer. Takes a fully-built domain model, returns an
ANSI string. Knows about Powerline glyphs, color codes, segment
collapsing rules (responds to `TerminalInfo.Width`), and pill styling
(model badge, progress bar, burn-rate cursor).

## Data Flow

1. Claude Code emits JSON to `stdin` (model, session id, cwd, ...).
2. `cmd/statusline/` parses to `domain.Input`.
3. `application.StatusLineService` fans out port calls in parallel:
   - `git.Status()`, `git.DiffStats()` for line 1.
   - `mcp.ActiveServers()` for line 2.
   - `usage.SessionAndWeekly()` for the burn-rate bars.
   - `terminal.Width()` for collapse decisions.
4. Per-port errors are dropped; the corresponding segment is omitted.
5. Renderer composes the two-line output as a single ANSI string.
6. `stdout` ← rendered string. Exit 0.

## Technology Stack

- **Language** — Go 1.25.5, toolchain 1.26.
- **Direct deps** — `golang.org/x/term` (terminal width/colors),
  `golang.org/x/sys` (transitive).
- **Build** — `make build` → `bin/status-line`. Cross-compiled binaries
  ship via GitHub Releases (`.github/workflows/release.yml`).
- **Testing** — stdlib `testing`. Convention: `*_test.go` in the same
  package, with `_internal_test.go` (white-box) and `_external_test.go`
  (black-box, package suffix `_test`).
- **Linting** — ktn-linter via `make lint` (148 rules, 8 phases). Run by
  `/lint` skill in CI.
- **CI** — `.github/workflows/ci.yml` runs `make test && make lint` on
  every push.
- **Release** — `.github/workflows/release.yml` builds, tags, and
  triggers `kodflow/devcontainer-template` rebuild on each release.

## Constraints

- **Latency budget** — Status Line runs every Claude Code interaction.
  Cold-start + render must stay under 50 ms p95. Anything network-bound
  (Anthropic OAuth usage) MUST cache to disk and refresh out-of-band.
- **Must not block** — if any port hangs, the surrounding segment is
  skipped; the binary always exits.
- **No persistent state** — we own no database. Caches live in
  `~/.cache/...` with mtime-based invalidation.
- **Single static binary** — no runtime dependencies on the user's
  system (apart from the optional `git` and `task` CLIs the adapters
  query — both gracefully skipped if absent).
- **Two-line output** — exactly two lines. Adding a third changes the
  Claude Code render layout and breaks expectation.
