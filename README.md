# Status Line

A customizable status line for Claude Code, displaying model info, context usage, git status, and more.

## Features

- **Model Display** - Shows current AI model (Sonnet, Opus, Haiku, Fable) with color-coded pill
- **Context Window** - Its own pill, in percent and in tokens, never conflated with a rate limit
- **Rate Limits** - Session, weekly and per-model quotas, each with an even-burn marker,
  an end-of-window projection and a reset countdown
- **Git Integration** - Branch name, modified files, untracked files
- **Code Changes** - Lines added/removed in current session
- **Service Health** - Claude's status page, as a light beside the OS icon
- **Auto-Update** - Updates itself to the latest release, out of band

## Installation

### Download Binary

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -sL https://github.com/kodflow/status-line/releases/latest/download/status-line-linux-amd64 -o status-line
chmod +x status-line
sudo mv status-line /usr/local/bin/

# Linux (arm64)
curl -sL https://github.com/kodflow/status-line/releases/latest/download/status-line-linux-arm64 -o status-line
chmod +x status-line
sudo mv status-line /usr/local/bin/

# macOS (Intel)
curl -sL https://github.com/kodflow/status-line/releases/latest/download/status-line-darwin-amd64 -o status-line
chmod +x status-line
sudo mv status-line /usr/local/bin/

# macOS (Apple Silicon)
curl -sL https://github.com/kodflow/status-line/releases/latest/download/status-line-darwin-arm64 -o status-line
chmod +x status-line
sudo mv status-line /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/kodflow/status-line/releases/latest/download/status-line-windows-amd64.exe -OutFile status-line.exe
```

### Build from Source

```bash
go build -o status-line ./cmd/statusline
```

## Usage

Status-line reads JSON input from stdin and outputs a formatted status line:

```bash
echo '{"model":{"display_name":"Sonnet 4"},"workspace":{"current_dir":"/path"},"context_window":{"total_input_tokens":50000,"total_output_tokens":10000,"context_window_size":200000}}' | status-line
```

### Claude Code Integration

Configure in your Claude Code settings to use as the status line provider.

## The line

```
 󰙴  Opus 5 ━━━━────●─ 47%  59m  Weekly ━━━━━●──── 58%  2d23h    Context ━───────── 10%    ~/project   main !3 ?1  +142  -37   v0.30.0
```

The status line is one line: the OS icon and the health light, the model with
its own rate limits beside it, the context window, then where you are, and
the update notice at the very end when a newer release is known.

The line fits the terminal: the host gives the status line its width in
`COLUMNS` (120 when absent). Each segment has levels, from richest to leanest,
and a weight; a line too wide lowers segments one level at a time, lightest
first, every segment once before any twice. By default: the context, scoped,
weekly and session bars, the update notice, the path and branch to 20, the
countdowns, the path to its last element and the branch to 12, quota names to
their initial, the changes, the path, the model icon, the branch to 8, and
last of all the health light. The OS icon never goes. An update notice that
reports a failure holds out past the first pass, just before the health light.

Each bar marks where an even burn would sit right now (`●`). The fill behind
that mark means room to spare; ahead of it means the quota runs out before it
refills.

A quota scoped to one model family shows only while that model is in use.

### Segments

| Segment | Description |
|---------|-------------|
| OS Icon | Linux, macOS, Windows, or Docker, then the health light of Claude's services |
| Model Pill | Colored by model (pink=Haiku, purple=Sonnet, orange=Opus) |
| Progress Bar | Context window usage with burn-rate cursor (●) |
| Path | Current working directory |
| Git | Branch name, modified (!), untracked (?) |
| Changes | Lines added (+) and removed (-) |
| Update | At the end of the line. Lit while a newer release is known and the running binary is older; turns red with a warning glyph when the install failed or checks have failed for over an hour |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `STATUSLINE_ICON_OS` | Show OS icon | `true` |
| `STATUSLINE_ICON_MODEL` | Show model icon | `true` |
| `STATUSLINE_ICON_PATH` | Show folder icon | `true` |
| `STATUSLINE_ICON_GIT` | Show git branch icon | `true` |
| `STATUSLINE_GLYPHS` | `nerd` or `text` (no Nerd Font required) | `nerd` |
| `STATUSLINE_HIDE` | Comma-separated segments to leave out: `context`, `session`, `weekly`, `model`, `credits`, `health` | — |
| `STATUSLINE_WEIGHTS` | Condensing weights, e.g. `context=10,weekly=30` (lower shrinks first; segments: context, scoped, weekly, session, path, branch, os, changes, model, update) | built-in policy |

## Auto-Update

**A frame never touches the network.** The host cancels a status line command
as soon as the next update triggers, so anything slow done inside a frame is
killed half way: the old in-frame download left partial
`status-line-update-*` files beside the binary and never installed anything.

Instead each frame reads `.status-line-update.json`, kept beside the real
binary (`os.Executable()` resolved through its symlinks), and only when its
`next_attempt_at` is reached starts a detached child, `status-line
--check-update`, in its own session. That child:

- takes a non-blocking `flock` on `.status-line-update.lock` (another child at
  work: it exits; a lock file it cannot even open pushes the next attempt out
  by an hour), then re-reads the state and leaves if the check is no longer
  due, so two children spawned back to back make one request
  (`--check-update=force` skips that test, for a manual run);
- records `next_attempt_at` **before** the request, so a killed child still
  holds frames off: +1 h after a success, 5 min doubled up to 1 h after
  failures in a row, lengthened (never shortened) by `Retry-After` /
  `X-RateLimit-Reset` on a 403 or 429. A `next_attempt_at` more than 24 h
  ahead, which only a clock set in the future could write, counts as stale;
- asks the GitHub API (10 s), accepts only a `vMAJOR.MINOR.PATCH` tag, records
  `check_error` on any failure and clears it on success;
- when the tag is newer, installs it in the same process under the same lock:
  sweeps `status-line-update-*` partial downloads older than an hour,
  downloads (2 min), verifies the checksum, `chmod`, runs the download once
  with `--version` and requires it to name the tag (an asset that dies at load,
  or one built with the wrong version, is refused), `rename`, and records or
  clears `apply_error`.

`status-line --apply-update=<tag>` is the same install as a standalone entry
point. The state file is written to a temporary then renamed, so a frame never
reads half of it (a half-written or unreadable file counts as stale).

The update pill is derived from that file on every frame: lit while the cached
tag is newer than the running version, so it goes out by itself once the new
binary renders; red with a warning glyph when the install failed or checks
have been failing for over an hour. If rendering fails before a single byte
went out, the frame is a plain line, `status-line <version> : erreur de rendu`,
followed by `· mise à jour en attente` when a newer release is known.

Self-replacement runs on Linux and macOS only; on Windows nothing is checked
or replaced.

**Every download is verified** against the `.sha256` asset published alongside
it, fetched from the same release tag. A missing, malformed or mismatching
checksum aborts the update and leaves the running binary untouched — this
process replaces itself in place, so an unverified payload would be executed on
every redraw from then on.

Set `STATUS_LINE_NO_SELF_UPDATE` to any non-empty value to switch the updater
off entirely: no child is started, a child started anyway exits at once, and
the pill is hidden. A frame also never starts a child for a dev build or when
the binary's directory is not writable. Managed
images that verify the binary at build time and rebuild on their own schedule
should set it, otherwise a runtime replacement voids that verification.

To disable auto-update, build without version:
```bash
go build -o status-line ./cmd/statusline  # No -ldflags
```

Release builds include version via ldflags:
```bash
go build -ldflags "-X main.version=v0.4.0" -o status-line ./cmd/statusline
```

## Where the quota figures come from

Two sources, merged, with stdin winning wherever both carry the same bucket:

1. **stdin** — Claude Code >= 2.1.140 pipes `rate_limits` in on every render.
   Free, synchronous, always current.
2. **The OAuth usage API** — enrichment only: per-model weekly quotas, the
   extra-usage credit balance, and whichever buckets stdin did not send.

The API's `limits[]` array is the plan-agnostic source and is read first; the
legacy `five_hour` / `seven_day` fields are a fallback. They are `null` on plans
that do not expose them, so reading them alone loses the weekly quota entirely
on such an account.

**A bucket absent from both sources stays absent.** Plans genuinely differ: one
exposes a weekly cap, another does not. Rendering a missing quota as a
zero-percent bar would be a lie about the account.

The API response is cached for 60s under `$XDG_RUNTIME_DIR` (or `$TMPDIR`) in a
private per-user directory. A stale cache is served immediately and refreshed by
a detached process, so rendering never waits on the network.

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build
make build
```

## Releases

| Version | Features |
|---------|----------|
| v0.4.0 | Update notification pill |
| v0.3.x | Auto-updater, cursor style |
| v0.2.0 | Windows builds |
| v0.1.0 | Initial release |

## License

MIT License - See [LICENSE](LICENSE) for details.
