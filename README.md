# Status Line

A customizable status line for Claude Code, displaying model info, context usage, git status, and more.

## Features

- **Model Display** - Shows current AI model (Sonnet, Opus, Haiku, Fable) with color-coded pill
- **Context Window** - Its own pill, in percent and in tokens, never conflated with a rate limit
- **Rate Limits** - Session, weekly and per-model quotas, each with an even-burn marker,
  an end-of-window projection and a reset countdown
- **Git Integration** - Branch name, modified files, untracked files
- **Code Changes** - Lines added/removed in current session
- **MCP Servers** - Display configured MCP server status
- **Auto-Update** - Automatically updates to latest release

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
  Opus 5 ━━━━────●─ 47%  59m  Weekly ━━━━━●──── 58%  2d23h    Context ━───────── 10% 102k/1M    ~/project   main !3 ?1  +142  -37
 github · gitlab
```

Line one carries everything about the session: the model with its own rate
limits beside it, the context window, then where you are. Line two carries the
ambient MCP server pills.

Each bar marks where an even burn would sit right now (`●`). The fill behind
that mark means room to spare; ahead of it means the quota runs out before it
refills.

A quota scoped to one model family shows only while that model is in use.
dashboard:
  [OS] [Opus 5 ●] [/path] [git branch !2 ?1] [+50] [-10]
  [ctx ██░░ 10% 103k/1M] [session ██░│░ 15% ▸33% ⟳2h46] [⛁ 95% ⟳47m] [$1.83]
  [mcp-server] [ v0.4.0]
```

### Segments

| Segment | Description |
|---------|-------------|
| OS Icon | Linux, macOS, Windows, or Docker |
| Model Pill | Colored by model (pink=Haiku, purple=Sonnet, orange=Opus) |
| Progress Bar | Context window usage with burn-rate cursor (●) |
| Path | Current working directory |
| Git | Branch name, modified (!), untracked (?) |
| Changes | Lines added (+) and removed (-) |
| MCP | Configured MCP servers |
| Update | Shows version when update is downloading |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `STATUSLINE_ICON_OS` | Show OS icon | `true` |
| `STATUSLINE_ICON_MODEL` | Show model icon | `true` |
| `STATUSLINE_ICON_PATH` | Show folder icon | `true` |
| `STATUSLINE_ICON_GIT` | Show git branch icon | `true` |
| `STATUSLINE_LINE_GAP` | Blank lines between the two rows (0-3) | `0` |
| `STATUSLINE_LINKS` | `0` disables the OSC 8 clickable segments | `1` |
| `STATUSLINE_GLYPHS` | `nerd` or `text` (no Nerd Font required) | `nerd` |
| `STATUSLINE_HIDE` | Comma-separated segments to leave out: `context`, `session`, `weekly`, `model` | — |

## Auto-Update

Status-line automatically checks for updates once per hour and downloads newer versions in the background. The update notification appears on line 2 while downloading.

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
