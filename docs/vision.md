<!-- updated: 2026-05-01T17:55:00Z -->
# Vision: Status Line

## Purpose

Status Line is a small, fast Go CLI that renders a Powerline-style status bar
for [Claude Code](https://claude.com/claude-code). It reads the JSON event
that Claude Code emits on every render, queries a handful of local sources
(git, MCP config, the Anthropic OAuth API), and writes ANSI
back to stdout — once per Claude Code interaction, in tens of milliseconds.

The output is two lines:

- **Line 1** — OS, model pill + session burn-rate, weekly burn-rate, path,
  git branch + dirty count, lines added/removed.
- **Line 2** — active MCP servers,
  available-update notice.

## Problem Statement

Heavy Claude Code users juggle several pieces of state that the default UI
doesn't expose:

- **Which model is active** — Sonnet vs Opus vs Haiku changes cost and
  latency, and switching is silent.
- **Where you are in the rate-limit window** — both the 5-hour session
  bucket (real Anthropic OAuth `five_hour.utilization`, not a context
  approximation) and the rolling 7-day weekly bucket.
- **Git state** — branch, modified count, untracked count, lines added/
  removed in the current session.
- **Active MCP servers** — what tools the agent currently has access to,
  derived from `~/.claude/.claude.json` and `mcp.json`.

Without these, you discover problems after the fact: the rate-limit cap, a
forgotten unstaged file, an MCP server that didn't load. Status Line surfaces
them inline so the next decision is informed.

## Target Users

- Developers using Claude Code in a terminal — every render gets the bar.
- Heavy users who hit the 5-hour or weekly rate limit and want a leading
  indicator instead of an error.
  segmented progress).
- Anyone who wants Powerline aesthetics integrated with the agent state.

## Goals

1. **Correct** — never crash, never block; degrade gracefully when any
   external dependency (git, MCP, Anthropic API) is missing.
2. **Fast** — runs on every Claude Code render tick; budget is tens of
   milliseconds, not hundreds.
3. **Readable** — Powerline segments + colors that survive narrow terminals
   (auto-collapse based on terminal width).
4. **Self-updating** — single static binary that can update itself from
   GitHub Releases (`internal/adapter/updater`).
5. **Hexagonal** — domain in the middle, adapters on the edge; every
   external dependency is replaceable for tests.

## Success Criteria

| Metric                            | Target                                  |
|-----------------------------------|-----------------------------------------|
| Cold-start latency                | < 50 ms p95 on a warm filesystem        |
| Stdout never blocks Claude Code   | 0 reports of render hangs               |
| Behaviour with all deps missing   | Degraded line still renders (no panic)  |
| Test coverage on `internal/`      | ≥ 80 % statement coverage               |
| Binary size                       | ≤ 10 MB stripped                        |
| Anthropic API failures            | Auto-hide weekly segment, no error to user |

## Design Principles

- **Stateless** — no on-disk state of our own; everything is recomputed.
- **Stdlib-first** — minimize deps. Current footprint: `golang.org/x/term`,
  `golang.org/x/sys`. Adding a dep needs a one-line justification.
- **Adapter isolation** — every external (git, MCP, Anthropic,
  filesystem, terminal) is behind a port interface in `internal/domain/port`.
- **No network on the hot path unless cached** — Anthropic OAuth usage is
  cached to disk; refresh happens out-of-band.
- **Fail open** — if an adapter errors, its segment is suppressed, not
  rendered as an error.
- **Tests as documentation** — `*_test.go` in the same package, with the
  `_internal_test` / `_external_test` split convention.

## Non-Goals

- **Not a TUI** — no interactivity, no curses, no input loop. Stdin in,
  stdout out, exit.
- **Not a notification system** — not surfacing alerts, just state.
- **Not a Claude Code plugin** — it's invoked via Claude Code's status-line
  hook; nothing else.
- **No persistent storage** — we don't own a database; if a source needs
  caching, the cache lives in `~/.cache/...` and is invalidated by mtime,
  not maintained by us.
- **No multi-line history** — exactly two lines, every render.

## Key Decisions

| Decision                                      | Rationale                                                                 |
|-----------------------------------------------|---------------------------------------------------------------------------|
| Go 1.25.5 (toolchain 1.26)                     | Fast cold-start, single static binary, easy cross-compile.               |
| Hexagonal layout (`cmd/` + `internal/{...}`)   | Adapters are independently testable; new sources (e.g. JIRA) plug in cleanly. |
| ktn-linter via `make lint`                     | Aligns with the rest of the kodflow ecosystem; 148 rules across 8 phases. |
| Powerline ANSI in `presentation/renderer/`     | Pure rendering layer — no business logic; testable on string output.     |
| Real Anthropic OAuth `five_hour.utilization`   | Replaces the heuristic context approximation; matches what the user sees in Anthropic's dashboard. |
| `_test.go` in same package, internal/external split | Allows white-box tests for unexported helpers without polluting the public API. |
| Auto-update from GitHub Releases               | One install path, no package-manager fragmentation.                       |
