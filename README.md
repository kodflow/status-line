# Status Line

A customizable status line for Claude Code, displaying model info, context usage, git status, and more.

## Features

- **Model Display** - Shows current AI model (Sonnet, Opus, Haiku) with color-coded pill
- **Context Progress Bar** - Visual progress bar with burn-rate cursor indicator
- **Git Integration** - Branch name, modified files, untracked files
- **Code Changes** - Lines added/removed in current session
- **MCP Servers** - Display configured MCP server status
- **Taskwarrior** - Project progress tracking (if installed)
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

## Visual Components

```
Line 1: [OS] [Model ━━━━━━━━●━━━━━━ 74%] [/path] [git branch !2 ?1] [+50] [-10]
Line 2: [taskwarrior] [mcp-server] [ v0.4.0]
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
| Taskwarrior | Project progress (if installed) |
| MCP | Configured MCP servers |
| Update | Shows version when update is downloading |

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `STATUSLINE_ICON_OS` | Show OS icon | `true` |
| `STATUSLINE_ICON_MODEL` | Show model icon | `true` |
| `STATUSLINE_ICON_PATH` | Show folder icon | `true` |
| `STATUSLINE_ICON_GIT` | Show git branch icon | `true` |

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

## API Usage Integration

For burn-rate cursor display, configure Anthropic API credentials:

```bash
# ~/.config/anthropic/credentials.json
{
  "default": {
    "api_key": "your-api-key",
    "organization_id": "your-org-id"
  }
}
```

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
