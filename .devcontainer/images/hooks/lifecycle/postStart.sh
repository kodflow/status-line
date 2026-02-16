#!/bin/bash
# shellcheck disable=SC1090,SC1091
# ============================================================================
# postStart.sh - Runs EVERY TIME the container starts
# ============================================================================
# This script runs after postCreateCommand and before postAttachCommand.
# Runs each time the container is successfully started.
# Use it for: MCP setup, services startup, recurring initialization.
#
# Uses run_step pattern: each step runs in an isolated subshell so that
# failures never block the DevContainer lifecycle.
# ============================================================================

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../shared/utils.sh"

log_info "postStart: Container starting..."

init_steps

# ============================================================================
# Step functions
# ============================================================================

# Restore Claude commands/scripts from image defaults OR host installation
# Priority:
#   1. Host installation ($HOME/.claude/ with .template-version)
#   2. Image defaults (/etc/claude-defaults/)
step_restore_claude_config() {
    local CLAUDE_DEFAULTS="/etc/claude-defaults"

    # Check if user has a host-installed Claude configuration
    # (installed via install.sh - indicated by .template-version file)
    if [ -f "$HOME/.claude/.template-version" ]; then
        log_info "Detected host-installed Claude configuration (via install.sh)"
        log_info "Skipping image defaults restore - using host installation"
        log_success "Claude configuration from host (priority)"
        return 0
    fi

    if [ ! -d "$CLAUDE_DEFAULTS" ]; then
        log_warning "No Claude configuration found (neither host nor image defaults)"
        return 0
    fi

    log_info "Restoring Claude configuration from image defaults..."

    # Ensure base directory exists
    mkdir -p "$HOME/.claude"

    # CLEAN commands, scripts, agents and docs to avoid legacy pollution
    # Only these directories are managed by the image - sessions/plans are user data
    rm -rf "$HOME/.claude/commands" "$HOME/.claude/scripts" "$HOME/.claude/agents" "$HOME/.claude/docs"

    # Restore commands (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/commands" ]; then
        mkdir -p "$HOME/.claude/commands"
        cp -r "$CLAUDE_DEFAULTS/commands/"* "$HOME/.claude/commands/" 2>/dev/null || true
    fi

    # Restore scripts (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/scripts" ]; then
        mkdir -p "$HOME/.claude/scripts"
        cp -r "$CLAUDE_DEFAULTS/scripts/"* "$HOME/.claude/scripts/" 2>/dev/null || true
        chmod -R 755 "$HOME/.claude/scripts/"
    fi

    # Restore agents (fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/agents" ]; then
        mkdir -p "$HOME/.claude/agents"
        cp -r "$CLAUDE_DEFAULTS/agents/"* "$HOME/.claude/agents/" 2>/dev/null || true
        chmod -R 755 "$HOME/.claude/agents/"
    fi

    # Restore docs (Design Patterns Knowledge Base - fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/docs" ]; then
        mkdir -p "$HOME/.claude/docs"
        cp -r "$CLAUDE_DEFAULTS/docs/"* "$HOME/.claude/docs/" 2>/dev/null || true
    fi

    # Restore templates (Documentation and C4 templates - fresh copy from image)
    if [ -d "$CLAUDE_DEFAULTS/templates" ]; then
        mkdir -p "$HOME/.claude/templates"
        cp -r "$CLAUDE_DEFAULTS/templates/"* "$HOME/.claude/templates/" 2>/dev/null || true
    fi

    # Restore settings.json only if it does not exist (user customizations preserved)
    if [ -f "$CLAUDE_DEFAULTS/settings.json" ] && [ ! -f "$HOME/.claude/settings.json" ]; then
        cp "$CLAUDE_DEFAULTS/settings.json" "$HOME/.claude/settings.json"
    fi

    log_success "Claude configuration restored from image defaults"
}

# Ensure Claude directories exist (volume mount point)
step_init_claude_dirs() {
    mkdir -p "$HOME/.claude/sessions" "$HOME/.claude/plans"
    log_success "Claude directories initialized"
}

# Ensure shell environment is properly configured (repair mechanism)
# postCreate.sh creates ~/.devcontainer-env.sh with super-claude() and other
# shell functions. If the source line is missing from RC files (stale image,
# volume issue, or .bashrc never configured), inject it here on every start.
step_shell_env_repair() {
    local DC_SOURCE_LINE='[[ -f ~/.devcontainer-env.sh ]] && source ~/.devcontainer-env.sh'

    for rc_file in "$HOME/.zshrc" "$HOME/.bashrc"; do
        if [ -f "$rc_file" ] && ! grep -q "devcontainer-env.sh" "$rc_file" 2>/dev/null; then
            printf '\n%s\n' "$DC_SOURCE_LINE" >> "$rc_file"
            log_info "Added devcontainer-env source to $(basename "$rc_file")"
        fi
    done

    # Upgrade devcontainer-env.sh to two-phase loading (v2)
    # v1 loaded heavy init (nvm.sh, eval pyenv/rbenv, 15+ completions) unconditionally,
    # causing VS Code ptyHost to timeout resolving shell environment.
    # v2 splits into: Phase 1 (fast PATH exports) + Phase 2 (heavy init, terminal only).
    local DC_ENV="$HOME/.devcontainer-env.sh"
    if [ -f "$DC_ENV" ] && ! grep -q "two-phase" "$DC_ENV" 2>/dev/null; then
        log_info "Upgrading devcontainer-env.sh to two-phase loading (v2)..."

        # Trigger postCreate.sh to regenerate the env file
        # Remove the guard marker so postCreate regenerates, but preserve initialization state
        if [ -f /home/vscode/.devcontainer-initialized ]; then
            rm -f /home/vscode/.devcontainer-initialized
            if [ -f /workspace/.devcontainer/hooks/lifecycle/postCreate.sh ]; then
                bash /workspace/.devcontainer/hooks/lifecycle/postCreate.sh 2>/dev/null || true
            fi
            # Restore marker (postCreate.sh recreates it, but ensure it exists)
            touch /home/vscode/.devcontainer-initialized
        fi
        log_success "devcontainer-env.sh upgraded to v2 (two-phase)"
    fi

    # Remove duplicate NVM from .zshrc (added by nodejs feature, already in env.sh)
    if [ -f "$HOME/.zshrc" ] && grep -c "NVM_DIR" "$HOME/.zshrc" 2>/dev/null | grep -q "^[2-9]"; then
        sed -i '/^# NVM (Node Version Manager)$/,/^\[ -s "\$NVM_DIR\/nvm\.sh" \]/d' "$HOME/.zshrc" 2>/dev/null || true
        sed -i '/^export NVM_SYMLINK_CURRENT=true$/d' "$HOME/.zshrc" 2>/dev/null || true
        log_info "Removed duplicate NVM from .zshrc (handled by devcontainer-env.sh)"
    fi

    # Ensure zsh is the default login shell
    if [ "$(getent passwd "$(whoami)" | cut -d: -f7)" != "/bin/zsh" ]; then
        sudo chsh -s /bin/zsh "$(whoami)" 2>/dev/null || true
        log_success "Default shell set to zsh"
    fi
}

# Reload .env file to get updated tokens
step_reload_env() {
    local ENV_FILE="/workspace/.devcontainer/.env"
    if [ -f "$ENV_FILE" ]; then
        log_info "Reloading environment from .env..."
        set -a
        source "$ENV_FILE"
        set +a
    fi
}

# Fix 1Password CLI config directory permissions
# Docker named volumes create directories with root ownership.
# 1Password CLI requires: ownership by current user + permissions 700.
step_1password_permissions() {
    local OP_CONFIG_DIRS=("$HOME/.config/op" "$HOME/.op")

    for OP_DIR in "${OP_CONFIG_DIRS[@]}"; do
        if [ -d "$OP_DIR" ]; then
            # Fix ownership if not current user
            if [ "$(stat -c '%U' "$OP_DIR" 2>/dev/null)" != "$(whoami)" ]; then
                log_info "Fixing ownership of $OP_DIR..."
                sudo chown -R "$(whoami):$(whoami)" "$OP_DIR"
            fi
            # Ensure correct permissions (700 = owner only)
            chmod 700 "$OP_DIR"
        fi
    done
    log_success "1Password config directories configured"
}

# Fix npm cache permissions
# Docker named volumes create directories with root ownership.
# npm requires write access to its cache for npx/MCP servers to work.
step_npm_cache_permissions() {
    local NPM_CACHE_DIR="$HOME/.cache/npm"

    if [ -d "$NPM_CACHE_DIR" ]; then
        # Fix ownership if not current user
        if [ "$(stat -c '%U' "$NPM_CACHE_DIR" 2>/dev/null)" != "$(whoami)" ]; then
            log_info "Fixing ownership of npm cache..."
            sudo chown -R "$(whoami):$(whoami)" "$NPM_CACHE_DIR"
        fi
    fi
    log_success "npm cache configured"
}

# MCP configuration setup (inject secrets into template)
step_mcp_configuration() {
    # 1Password vault ID (can be overridden via OP_VAULT_ID env var)
    local VAULT_ID="${OP_VAULT_ID:-ypahjj334ixtiyjkytu5hij2im}"
    local MCP_TPL="/etc/mcp/mcp.json.tpl"
    local MCP_OUTPUT="/workspace/mcp.json"

    # Helper function to get 1Password field (tries multiple field names)
    get_1password_field() {
        local item="$1"
        local vault="$2"
        local fields=("credential" "password" "identifiant" "mot de passe")

        for field in "${fields[@]}"; do
            local value
            value=$(op item get "$item" --vault "$vault" --fields "$field" --reveal 2>/dev/null || echo "")
            if [ -n "$value" ]; then
                echo "$value"
                return 0
            fi
        done
        echo ""
    }

    # Initialize tokens from environment variables (fallback)
    local CODACY_TOKEN="${CODACY_API_TOKEN:-}"
    local GITHUB_TOKEN="${GITHUB_API_TOKEN:-}"

    # Try 1Password if OP_SERVICE_ACCOUNT_TOKEN is defined
    if [ -n "${OP_SERVICE_ACCOUNT_TOKEN:-}" ] && command -v op &> /dev/null; then
        log_info "Retrieving secrets from 1Password..."

        local OP_CODACY
        OP_CODACY=$(get_1password_field "mcp-codacy" "$VAULT_ID")
        local OP_GITHUB
        OP_GITHUB=$(get_1password_field "mcp-github" "$VAULT_ID")

        [ -n "$OP_CODACY" ] && CODACY_TOKEN="$OP_CODACY"
        [ -n "$OP_GITHUB" ] && GITHUB_TOKEN="$OP_GITHUB"
    fi

    # Show status of tokens (INFO for optional, WARNING for essential)
    [ -z "$CODACY_TOKEN" ] && log_info "Codacy token not configured (optional)"
    [ -z "$GITHUB_TOKEN" ] && log_warning "GitHub token not available"

    # Helper: escape special chars for sed replacement
    escape_for_sed() {
        LC_ALL=C printf '%s' "$1" | tr -d '\n\r' | sed -e 's/[&/|\\]/\\&/g'
    }

    # Security: refuse to write secrets through symlinks or unsafe directories
    local MCP_DIR
    MCP_DIR=$(dirname -- "$MCP_OUTPUT")
    if [ ! -d "$MCP_DIR" ] || [ -L "$MCP_DIR" ]; then
        log_error "Refusing to write mcp.json: unsafe parent directory ($MCP_DIR)"
        return 0
    elif [ -e "$MCP_OUTPUT" ] && { [ -L "$MCP_OUTPUT" ] || [ ! -f "$MCP_OUTPUT" ]; }; then
        log_error "Refusing to write mcp.json: not a regular file ($MCP_OUTPUT)"
        return 0
    fi

    # Migrate legacy .mcp.json to mcp.json (renamed in v2)
    if [ -f "/workspace/.mcp.json" ] && [ ! -e "$MCP_OUTPUT" ]; then
        log_info "Migrating legacy .mcp.json to mcp.json..."

        if ! command -v jq >/dev/null 2>&1; then
            log_warning "jq not found; migrating without JSON validation"
            if cp "/workspace/.mcp.json" "$MCP_OUTPUT"; then
                chown "$(id -u):$(id -g)" "$MCP_OUTPUT" 2>/dev/null || true
                chmod 600 "$MCP_OUTPUT"
                rm -f "/workspace/.mcp.json" || log_warning "Could not remove legacy .mcp.json (permissions?)"
                log_success "Migration complete: .mcp.json → mcp.json"
            else
                log_error "Migration failed: unable to copy legacy file"
            fi
        else
            local MCP_MIG_TMP
            MCP_MIG_TMP=$(mktemp "${MCP_OUTPUT}.migrate.XXXXXX") || {
                log_error "Migration failed: unable to create temp file"
                MCP_MIG_TMP=""
            }
            if [ -n "$MCP_MIG_TMP" ] && cp "/workspace/.mcp.json" "$MCP_MIG_TMP"; then
                if jq empty "$MCP_MIG_TMP" 2>/dev/null; then
                    mv "$MCP_MIG_TMP" "$MCP_OUTPUT"
                    chown "$(id -u):$(id -g)" "$MCP_OUTPUT" 2>/dev/null || true
                    chmod 600 "$MCP_OUTPUT"
                    rm -f "/workspace/.mcp.json" || log_warning "Could not remove legacy .mcp.json (permissions?)"
                    log_success "Migration complete: .mcp.json → mcp.json"
                else
                    log_error "Legacy .mcp.json is invalid JSON; keeping legacy file"
                    rm -f "$MCP_MIG_TMP"
                fi
            elif [ -n "$MCP_MIG_TMP" ]; then
                log_error "Migration failed"
                rm -f "$MCP_MIG_TMP"
            fi
        fi
    fi

    # Generate mcp.json from template (baked in Docker image)
    # ALWAYS regenerate from template to ensure latest MCP config is applied
    if [ -f "$MCP_TPL" ]; then
        if [ -z "$CODACY_TOKEN" ] && [ -z "$GITHUB_TOKEN" ]; then
            log_warning "No tokens available, creating minimal mcp.json"
            printf '%s\n' '{"mcpServers":{}}' > "$MCP_OUTPUT"
            chown "$(id -u):$(id -g)" "$MCP_OUTPUT" 2>/dev/null || true
            chmod 600 "$MCP_OUTPUT"
            log_info "Created minimal mcp.json (optional MCPs will be added below)"
        else
            generate_mcp_from_template() {
                local escaped_codacy escaped_github mcp_tmp
                escaped_codacy=$(escape_for_sed "${CODACY_TOKEN}")
                escaped_github=$(escape_for_sed "${GITHUB_TOKEN}")

                mcp_tmp=$(mktemp "${MCP_OUTPUT}.tmp.XXXXXX") || {
                    log_error "Failed to create temp file for mcp.json generation"
                    return 0
                }

                trap 'rm -f "${mcp_tmp:-}" 2>/dev/null || true' RETURN

                if ! sed -e "s|{{CODACY_TOKEN}}|${escaped_codacy}|g" \
                        -e "s|{{GITHUB_TOKEN}}|${escaped_github}|g" \
                        "$MCP_TPL" > "$mcp_tmp"; then
                    log_error "Failed to render mcp.json template"
                    return 0
                fi

                if jq empty "$mcp_tmp" 2>/dev/null; then
                    mv "$mcp_tmp" "$MCP_OUTPUT"
                    chown "$(id -u):$(id -g)" "$MCP_OUTPUT" 2>/dev/null || true
                    chmod 600 "$MCP_OUTPUT"
                    log_success "mcp.json generated successfully"
                else
                    log_error "Generated mcp.json is invalid JSON, keeping original"
                fi
            }
            log_info "Regenerating mcp.json from template (forced)..."
            generate_mcp_from_template
        fi

        # Add optional MCPs based on installed features
        add_optional_mcp() {
            local name="$1"
            local binary="$2"
            local output="$3"

            [ -f "$output" ] || return 0

            if ! command -v jq >/dev/null 2>&1; then
                log_warning "Skipping $name MCP injection (jq not found)"
                return 0
            fi

            if [ -x "$binary" ]; then
                log_info "Adding $name MCP (binary found at $binary)"
                local tmp_file
                tmp_file=$(mktemp "${output}.tmp.XXXXXX") || {
                    log_warning "Failed to add $name MCP (unable to create temp file)"
                    return 0
                }
                if jq --arg name "$name" --arg bin "$binary" \
                   '.mcpServers = (.mcpServers // {}) | .mcpServers[$name] = {"command": $bin, "args": [], "env": {}}' \
                   "$output" > "$tmp_file" && jq empty "$tmp_file" 2>/dev/null; then
                    mv "$tmp_file" "$output"
                    chown "$(id -u):$(id -g)" "$output" 2>/dev/null || true
                    chmod 600 "$output" 2>/dev/null || true
                else
                    log_warning "Failed to add $name MCP, keeping original"
                    rm -f "$tmp_file"
                fi
            else
                log_info "Skipping $name MCP (binary not found)"
            fi
        }

        add_optional_mcp "rust-analyzer" "$HOME/.cache/cargo/bin/rust-analyzer-mcp" "$MCP_OUTPUT"
    else
        log_warning "MCP template not found at $MCP_TPL"
    fi
}

# Clean git credential helpers (remove macOS-specific helpers)
step_git_credential_cleanup() {
    log_info "Cleaning git credential helpers..."
    git config --global --unset-all credential.https://github.com.helper 2>/dev/null || true
    git config --global --unset-all credential.https://gist.github.com.helper 2>/dev/null || true
    log_success "Git credential helpers cleaned"
}

# Auto-run /init for project initialization check
# Runs at every container start to verify project is properly initialized
# Skipped in CI environment
step_auto_init_check() {
    local INIT_LOG="$HOME/.devcontainer-init.log"

    if command -v claude &> /dev/null && [ -z "${CI:-}" ]; then
        log_info "Running project initialization check..."
        nohup bash -c "sleep 2 && claude \"/init\" || echo \"[\$(date -Iseconds)] Init check failed with exit code \$?\" >> \"$INIT_LOG\"" >> "$INIT_LOG" 2>&1 &
        log_success "Init check scheduled (logs: ~/.devcontainer-init.log)"
    elif [ -n "${CI:-}" ]; then
        log_info "CI environment detected, skipping init"
    fi
}

# ============================================================================
# Ollama + grepai Initialization (for semantic code search MCP)
# ============================================================================
GREPAI_BIN="/usr/local/bin/grepai"
GREPAI_CONFIG_TPL="/etc/grepai/config.yaml"
OLLAMA_HOST_ENDPOINT="host.docker.internal:11434"

detect_ollama_endpoint() {
    local endpoint=""
    local source=""

    if [ -n "${OLLAMA_HOST:-}" ]; then
        endpoint="$OLLAMA_HOST"
        source="OLLAMA_HOST env var"
        if curl -sf --connect-timeout 3 "http://${endpoint}/api/tags" >/dev/null 2>&1; then
            echo "$endpoint|$source"
            return 0
        else
            log_warning "OLLAMA_HOST=$endpoint not responding"
        fi
    fi

    endpoint="$OLLAMA_HOST_ENDPOINT"
    if curl -sf --connect-timeout 3 "http://${endpoint}/api/tags" >/dev/null 2>&1; then
        source="host (GPU-accelerated)"
        echo "$endpoint|$source"
        return 0
    fi

    echo ""
    return 1
}

check_model_available() {
    local endpoint="$1"
    local model="$2"
    curl -sf "http://${endpoint}/api/tags" 2>/dev/null | grep -q "$model"
}

show_ollama_instructions() {
    log_warning "==============================================================================="
    log_warning "  Ollama not running - Semantic search (grepai) will be disabled"
    log_warning "==============================================================================="
    log_info ""
    log_info "Ollama should be installed automatically via initialize.sh"
    log_info "If not running, start it manually on your host machine:"
    log_info ""
    log_info "  macOS:"
    log_info "    brew services start ollama"
    log_info "    # or: ollama serve"
    log_info ""
    log_info "  Linux:"
    log_info "    sudo systemctl start ollama"
    log_info "    # or: ollama serve"
    log_info ""
    log_info "  Then restart the DevContainer or run: /init"
    log_info ""
    log_warning "==============================================================================="
}

# --- Health stamp helpers ---
# The health stamp tracks 3 invalidation factors: model, binary version, config hash.
# If any factor changes between container starts, the index is purged and rebuilt.

compute_config_hash() {
    # Hash the config EXCLUDING the endpoint line (which changes per environment)
    grep -v '^\s*endpoint:' "$1" 2>/dev/null | md5sum | awk '{print $1}'
}

get_grepai_version() {
    "$GREPAI_BIN" version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown"
}

read_health_stamp() {
    # Reads .health-stamp into STAMP_* variables in the caller's scope
    # Returns 1 if stamp doesn't exist
    local stamp_file="$1"
    STAMP_MODEL=""
    STAMP_GREPAI_VERSION=""
    STAMP_CONFIG_HASH=""
    STAMP_DAEMON_PID=""
    STAMP_LAST_HEALTHY=""

    [ -f "$stamp_file" ] || return 1

    STAMP_MODEL=$(grep '^MODEL=' "$stamp_file" 2>/dev/null | cut -d= -f2-)
    STAMP_GREPAI_VERSION=$(grep '^GREPAI_VERSION=' "$stamp_file" 2>/dev/null | cut -d= -f2-)
    STAMP_CONFIG_HASH=$(grep '^CONFIG_HASH=' "$stamp_file" 2>/dev/null | cut -d= -f2-)
    # shellcheck disable=SC2034  # STAMP_DAEMON_PID used by grepai_watchdog via read_health_stamp
    STAMP_DAEMON_PID=$(grep '^DAEMON_PID=' "$stamp_file" 2>/dev/null | cut -d= -f2-)
    # shellcheck disable=SC2034  # STAMP_LAST_HEALTHY used by write_health_stamp
    STAMP_LAST_HEALTHY=$(grep '^LAST_HEALTHY=' "$stamp_file" 2>/dev/null | cut -d= -f2-)
    return 0
}

write_health_stamp() {
    local stamp_file="$1"
    local model="$2"
    local version="$3"
    local config_hash="$4"
    local daemon_pid="$5"

    cat > "$stamp_file" <<STAMP_EOF
MODEL=$model
GREPAI_VERSION=$version
CONFIG_HASH=$config_hash
DAEMON_PID=$daemon_pid
LAST_HEALTHY=$(date +%s)
STAMP_EOF
    chmod 644 "$stamp_file"
}

stop_grepai_daemon() {
    local pid
    pid=$(pgrep -f "$GREPAI_BIN watch" 2>/dev/null || true)
    if [ -n "$pid" ]; then
        log_info "Stopping grepai daemon (PID: $pid)..."
        kill "$pid" 2>/dev/null || true
        sleep 2
        # Force kill if still alive
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
            sleep 1
        fi
    fi
    # Clean stale lock (daemon may have held it when killed)
    rm -f "/workspace/.grepai/index.gob.lock"
}

_grepai_init_core() {
    local quiet="${1:-false}"
    local grepai_dir="/workspace/.grepai"
    local grepai_config="${grepai_dir}/config.yaml"
    local health_stamp="${grepai_dir}/.health-stamp"
    local grepai_log="/tmp/grepai.log"
    local detected_result=""
    local ollama_endpoint=""
    local ollama_source=""

    # --- Step 1: Detect Ollama endpoint ---
    [ "$quiet" = "false" ] && log_info "Checking Ollama on host ($OLLAMA_HOST_ENDPOINT)..."
    detected_result=$(detect_ollama_endpoint)

    if [ -n "$detected_result" ]; then
        ollama_endpoint=$(echo "$detected_result" | cut -d'|' -f1)
        ollama_source=$(echo "$detected_result" | cut -d'|' -f2)
        log_success "Ollama connected: $ollama_endpoint ($ollama_source)"
    else
        [ "$quiet" = "false" ] && show_ollama_instructions
        # Pre-initialize config for when Ollama becomes available
        if [ -x "$GREPAI_BIN" ] && [ -f "$GREPAI_CONFIG_TPL" ]; then
            mkdir -p "$grepai_dir"
            cp "$GREPAI_CONFIG_TPL" "$grepai_config" 2>/dev/null || true
            sed -i -E "s|(endpoint: http://)[^[:space:]]+|\1${OLLAMA_HOST_ENDPOINT}|" "$grepai_config" 2>/dev/null || true
            [ "$quiet" = "false" ] && log_info "grepai config initialized (waiting for Ollama)"
        fi
        return 1
    fi

    # --- Step 2: Verify grepai binary ---
    if [ ! -x "$GREPAI_BIN" ]; then
        [ "$quiet" = "false" ] && log_warning "grepai binary not found at $GREPAI_BIN"
        return 2
    fi

    # --- Step 3: Sync config from template (always, to prevent drift) ---
    mkdir -p "$grepai_dir"

    if [ -f "$GREPAI_CONFIG_TPL" ]; then
        cp "$GREPAI_CONFIG_TPL" "$grepai_config"
        sed -i -E "s|(endpoint: http://)[^[:space:]]+|\1${ollama_endpoint}|" "$grepai_config"
        log_success "grepai config synced from template (endpoint: http://$ollama_endpoint)"
    else
        log_warning "Config template not found at $GREPAI_CONFIG_TPL, using grepai init..."
        (cd /workspace && "$GREPAI_BIN" init --provider ollama --backend gob --yes 2>/dev/null) || true
        if [ -f "$grepai_config" ]; then
            sed -i -E "s|(endpoint: http://)[^[:space:]]+|\1${ollama_endpoint}|" "$grepai_config"
        fi
    fi

    # --- Step 4: Compute current state ---
    local current_model current_version current_config_hash
    current_model=$(grep -E '^\s+model:' "$grepai_config" 2>/dev/null | awk '{print $2}' | head -1)
    current_version=$(get_grepai_version)
    current_config_hash=$(compute_config_hash "$grepai_config")

    [ "$quiet" = "false" ] && log_info "grepai state: model=$current_model version=$current_version config=$current_config_hash"

    # --- Step 5-6: Multi-factor invalidation detection ---
    local need_rebuild=false
    local rebuild_reasons=""

    if read_health_stamp "$health_stamp"; then
        # Factor 1: model change (embeddings become incompatible)
        if [ -n "$current_model" ] && [ "$STAMP_MODEL" != "$current_model" ]; then
            log_warning "Model changed: ${STAMP_MODEL:-unknown} -> $current_model"
            need_rebuild=true
            rebuild_reasons="${rebuild_reasons}model_change "
        fi

        # Factor 2: grepai binary version change (index format may change)
        if [ "$STAMP_GREPAI_VERSION" != "$current_version" ]; then
            log_warning "grepai version changed: ${STAMP_GREPAI_VERSION:-unknown} -> $current_version"
            need_rebuild=true
            rebuild_reasons="${rebuild_reasons}version_change "
        fi

        # Factor 3: config change (chunk size, ignore patterns, etc.)
        if [ "$STAMP_CONFIG_HASH" != "$current_config_hash" ]; then
            log_warning "Config changed: ${STAMP_CONFIG_HASH:-unknown} -> $current_config_hash"
            need_rebuild=true
            rebuild_reasons="${rebuild_reasons}config_change "
        fi
    else
        # No health stamp = fresh install or first run after migration
        [ "$quiet" = "false" ] && log_info "No health stamp found (fresh install or first run)"

        # Migrate from legacy .model-stamp if present
        local legacy_stamp="${grepai_dir}/.model-stamp"
        if [ -f "$legacy_stamp" ]; then
            local legacy_model
            legacy_model=$(cat "$legacy_stamp" 2>/dev/null || echo "")
            if [ -n "$legacy_model" ] && [ "$legacy_model" != "$current_model" ]; then
                log_warning "Legacy stamp model mismatch: $legacy_model -> $current_model"
                need_rebuild=true
                rebuild_reasons="${rebuild_reasons}legacy_model_change "
            fi
            rm -f "$legacy_stamp"
            log_info "Migrated from legacy .model-stamp"
        fi

        # If index exists but no stamp, we can't trust it — rebuild
        if [ -f "${grepai_dir}/index.gob" ] && [ "$need_rebuild" = "false" ]; then
            log_warning "Index exists but no health stamp — cannot verify integrity"
            need_rebuild=true
            rebuild_reasons="${rebuild_reasons}missing_stamp "
        fi
    fi

    # --- Step 7: Handle invalidation ---
    if [ "$need_rebuild" = "true" ]; then
        log_warning "Index rebuild required: ${rebuild_reasons}"
        stop_grepai_daemon
        rm -f "${grepai_dir}/index.gob" "${grepai_dir}/symbols.gob" "${grepai_dir}/index.gob.lock"
        log_success "Index cleared — will rebuild from scratch"
    fi

    # --- Step 8: Verify model is available before starting daemon ---
    if [ -n "$current_model" ]; then
        if check_model_available "$ollama_endpoint" "$current_model"; then
            log_success "Model $current_model available on Ollama"
        else
            log_warning "Model $current_model not found on Ollama"
            [ "$quiet" = "false" ] && log_info "Pull the model on your host: ollama pull $current_model"
            [ "$quiet" = "false" ] && log_warning "Skipping daemon start until model is available"
            return 2
        fi
    fi

    # --- Step 9: Start or verify daemon ---
    local grepai_pid
    grepai_pid=$(pgrep -f "$GREPAI_BIN watch" 2>/dev/null || true)

    if [ -z "$grepai_pid" ]; then
        # Clean stale lock from previous crashed daemon (no process = lock is stale)
        rm -f "${grepai_dir}/index.gob.lock"
        log_info "Starting grepai watch daemon..."
        : > "$grepai_log"
        (cd /workspace && nohup "$GREPAI_BIN" watch > "$grepai_log" 2>&1 &)

        # Retry loop: wait up to 5 seconds for daemon to start
        local retries=10
        while [ $retries -gt 0 ]; do
            grepai_pid=$(pgrep -f "$GREPAI_BIN watch" 2>/dev/null || true)
            [ -n "$grepai_pid" ] && break
            retries=$((retries - 1))
            sleep 0.5
        done

        if [ -n "$grepai_pid" ]; then
            log_success "grepai watch daemon started (PID: $grepai_pid)"
        else
            log_warning "grepai daemon failed to start (check $grepai_log)"
            return 2
        fi
    else
        log_info "grepai watch daemon already running (PID: $grepai_pid)"
    fi

    # --- Step 10: Save health stamp ---
    write_health_stamp "$health_stamp" \
        "$current_model" "$current_version" "$current_config_hash" "$grepai_pid"
    log_success "Health stamp saved (model=$current_model ver=$current_version)"
}

init_semantic_search() {
    _grepai_init_core "false"
    # Always launch watchdog (handles both daemon monitoring AND deferred init)
    grepai_watchdog &
}

# Watchdog: monitors grepai daemon and handles deferred initialization.
# When health stamp exists: restarts crashed daemon.
# When no health stamp: retries init when Ollama becomes available.
# Runs in background for the lifetime of the container.
grepai_watchdog() {
    local grepai_dir="/workspace/.grepai"
    local health_stamp="${grepai_dir}/.health-stamp"
    local grepai_log="/tmp/grepai.log"
    local deferred_attempts=0

    # Write PID file for discoverability (shell functions are invisible to pgrep)
    echo $$ > /tmp/grepai-watchdog.pid

    # Let the container finish starting before first check
    sleep 30

    while true; do
        sleep 60

        if [ ! -f "$health_stamp" ]; then
            # Deferred init: Ollama may have become available since startup
            if detect_ollama_endpoint >/dev/null 2>&1; then
                deferred_attempts=$((deferred_attempts + 1))
                if [ "$deferred_attempts" -eq 1 ]; then
                    log_info "[WATCHDOG] Ollama now available — initializing semantic search"
                elif [ $((deferred_attempts % 5)) -eq 0 ]; then
                    log_warning "[WATCHDOG] Deferred init retry #${deferred_attempts}"
                fi

                if _grepai_init_core "true"; then
                    log_success "[WATCHDOG] Deferred initialization complete"
                    deferred_attempts=0
                fi
            fi
            continue
        fi

        # Reset counter once healthy
        deferred_attempts=0

        local current_pid
        current_pid=$(pgrep -f "$GREPAI_BIN watch" 2>/dev/null || true)

        if [ -z "$current_pid" ]; then
            log_warning "[WATCHDOG] grepai daemon not running — restarting..."

            # Verify Ollama is still reachable before restarting
            if ! detect_ollama_endpoint >/dev/null 2>&1; then
                log_warning "[WATCHDOG] Ollama not reachable, skipping restart"
                continue
            fi

            # Clean stale lock from crashed daemon before restart
            rm -f "${grepai_dir}/index.gob.lock"
            (cd /workspace && nohup "$GREPAI_BIN" watch >> "$grepai_log" 2>&1 &)
            sleep 3

            current_pid=$(pgrep -f "$GREPAI_BIN watch" 2>/dev/null || true)
            if [ -n "$current_pid" ]; then
                log_success "[WATCHDOG] Daemon restarted (PID: $current_pid)"
                # Update stamp with new PID
                if read_health_stamp "$health_stamp"; then
                    write_health_stamp "$health_stamp" \
                        "$STAMP_MODEL" "$STAMP_GREPAI_VERSION" \
                        "$STAMP_CONFIG_HASH" "$current_pid"
                fi
            else
                log_warning "[WATCHDOG] Failed to restart daemon (check $grepai_log)"
            fi
        fi
    done
}

# ============================================================================
# VPN Auto-Connect (optional - skipped if no config found)
# ============================================================================
# Multi-protocol VPN support: OpenVPN, WireGuard, IPsec/IKEv2, PPTP
# Config source: VPN_CONFIG_REF=op://VAULT/PROFILE (or legacy OPENVPN_CONFIG_REF)
#   - DOCUMENT "PROFILE" in vault → config file (protocol determined by tags)
#   - LOGIN "PROFILE" in vault → username/password (not needed for WireGuard)
#   - Tags: "openvpn" (default), "wireguard", "ipsec", "pptp"
#   - Fallback: file on disk at $OPENVPN_CONFIG
#   - Nothing found → skip silently

# --- OpenVPN connect (extracted for multi-protocol support) ---
connect_openvpn() {
    local vault="$1" profile="$2" doc_uuid="$3" login_uuid="$4"
    local ovpn_config="${OPENVPN_CONFIG:-/home/vscode/.config/openvpn/client.ovpn}"
    local ovpn_auth="${OPENVPN_AUTH:-/tmp/vpn-auth.txt}"
    local ovpn_dir
    ovpn_dir=$(dirname "$ovpn_config")

    mkdir -p "$ovpn_dir"

    # Download .ovpn config
    if [ -n "$doc_uuid" ] && op document get "$doc_uuid" --vault "$vault" > "$ovpn_config" 2>/dev/null; then
        chmod 600 "$ovpn_config"
        log_success "Resolved .ovpn config ($vault/$profile)"
    else
        log_warning "No DOCUMENT '$profile' in vault '$vault', skipping VPN"
        return 0
    fi

    # Resolve credentials
    if [ -n "$login_uuid" ]; then
        local vpn_user vpn_pass
        vpn_user=$(op read "op://$vault/$login_uuid/username" 2>/dev/null || echo "")
        vpn_pass=$(op read "op://$vault/$login_uuid/password" 2>/dev/null || echo "")
        if [ -n "$vpn_user" ] && [ -n "$vpn_pass" ]; then
            printf '%s\n%s\n' "$vpn_user" "$vpn_pass" > "$ovpn_auth"
            chmod 600 "$ovpn_auth"
            log_success "Resolved VPN credentials ($vault/$profile)"
        fi
    fi

    # Validate config
    if [ ! -s "$ovpn_config" ]; then
        log_warning "OpenVPN config is empty: $ovpn_config"
        return 0
    fi

    local -a vpn_args=(
        --config "$ovpn_config"
        --daemon ovpn-client
        --log /tmp/openvpn.log
        --script-security 2
        --up /etc/openvpn/update-dns
        --down /etc/openvpn/update-dns
        --keepalive 10 60
        --connect-retry 5
        --connect-retry-max 0
        --persist-tun
        --persist-key
        --resolv-retry infinite
    )
    if [ -f "$ovpn_auth" ] && [ -s "$ovpn_auth" ]; then
        vpn_args+=(--auth-user-pass "$ovpn_auth")
    fi

    log_info "Starting OpenVPN..."
    if sudo openvpn "${vpn_args[@]}"; then
        local attempt=0
        while [ $attempt -lt 15 ]; do
            if ip link show tun0 &>/dev/null; then
                local vpn_ip
                vpn_ip=$(ip -4 addr show tun0 2>/dev/null | grep -oP 'inet \K[\d.]+' || echo "unknown")
                log_success "VPN connected via OpenVPN (tun0: $vpn_ip)"
                return 0
            fi
            sleep 1
            ((attempt++))
        done
        log_warning "OpenVPN started but tun0 not detected after 15s (check /tmp/openvpn.log)"
    else
        log_warning "OpenVPN failed to start (check /tmp/openvpn.log)"
    fi
}

# --- WireGuard connect ---
connect_wireguard() {
    local vault="$1" profile="$2" doc_uuid="$3"
    local wg_config="/home/vscode/.config/wireguard/wg0.conf"

    mkdir -p "$(dirname "$wg_config")"

    if [ -n "$doc_uuid" ] && op document get "$doc_uuid" --vault "$vault" > "$wg_config" 2>/dev/null; then
        chmod 600 "$wg_config"
        log_success "Resolved WireGuard config ($vault/$profile)"
    else
        log_warning "No DOCUMENT '$profile' in vault '$vault', skipping WireGuard"
        return 0
    fi

    log_info "Starting WireGuard..."
    if sudo wg-quick up "$wg_config" 2>/dev/null; then
        local attempt=0
        while [ $attempt -lt 10 ]; do
            if ip link show wg0 &>/dev/null; then
                local vpn_ip
                vpn_ip=$(ip -4 addr show wg0 2>/dev/null | grep -oP 'inet \K[\d.]+' || echo "unknown")
                log_success "VPN connected via WireGuard (wg0: $vpn_ip)"
                return 0
            fi
            sleep 1
            ((attempt++))
        done
        log_warning "WireGuard started but wg0 not detected after 10s"
    else
        log_warning "WireGuard failed to start"
    fi
}

# --- IPsec/IKEv2 connect ---
connect_ipsec() {
    local vault="$1" profile="$2" doc_uuid="$3" login_uuid="$4"
    local ipsec_config="/home/vscode/.config/strongswan/ipsec.conf"

    mkdir -p "$(dirname "$ipsec_config")"

    if [ -n "$doc_uuid" ] && op document get "$doc_uuid" --vault "$vault" > "$ipsec_config" 2>/dev/null; then
        chmod 600 "$ipsec_config"
        log_success "Resolved IPsec config ($vault/$profile)"
    else
        log_warning "No DOCUMENT '$profile' in vault '$vault', skipping IPsec"
        return 0
    fi

    # Copy config and secrets to strongswan dir
    sudo cp "$ipsec_config" /etc/ipsec.d/profile.conf
    if [ -n "$login_uuid" ]; then
        local vpn_user vpn_pass
        vpn_user=$(op read "op://$vault/$login_uuid/username" 2>/dev/null || echo "")
        vpn_pass=$(op read "op://$vault/$login_uuid/password" 2>/dev/null || echo "")
        if [ -n "$vpn_user" ] && [ -n "$vpn_pass" ]; then
            printf '%s : EAP "%s"\n' "$vpn_user" "$vpn_pass" | sudo tee /etc/ipsec.d/profile.secrets > /dev/null
            sudo chmod 600 /etc/ipsec.d/profile.secrets
        fi
    fi

    local conn_name
    conn_name=$(grep -oP '(?<=^conn )\S+' "$ipsec_config" | head -1)
    if [ -z "$conn_name" ]; then
        log_warning "No connection name found in IPsec config"
        return 0
    fi

    log_info "Starting IPsec ($conn_name)..."
    sudo ipsec restart 2>/dev/null
    sleep 2
    if sudo ipsec up "$conn_name" 2>/dev/null; then
        log_success "VPN connected via IPsec ($conn_name)"
    else
        log_warning "IPsec connection '$conn_name' failed"
    fi
}

# --- PPTP connect ---
connect_pptp() {
    local vault="$1" profile="$2" doc_uuid="$3" login_uuid="$4"
    local pptp_config="/home/vscode/.config/pptp/tunnel.conf"

    mkdir -p "$(dirname "$pptp_config")"

    if [ -n "$doc_uuid" ] && op document get "$doc_uuid" --vault "$vault" > "$pptp_config" 2>/dev/null; then
        chmod 600 "$pptp_config"
        log_success "Resolved PPTP config ($vault/$profile)"
    else
        log_warning "No DOCUMENT '$profile' in vault '$vault', skipping PPTP"
        return 0
    fi

    if [ -n "$login_uuid" ]; then
        local vpn_user vpn_pass
        vpn_user=$(op read "op://$vault/$login_uuid/username" 2>/dev/null || echo "")
        vpn_pass=$(op read "op://$vault/$login_uuid/password" 2>/dev/null || echo "")
        if [ -n "$vpn_user" ] && [ -n "$vpn_pass" ]; then
            printf '%s\n%s\n' "$vpn_user" "$vpn_pass" > /tmp/vpn-auth.txt
            chmod 600 /tmp/vpn-auth.txt
        fi
    fi

    log_info "Starting PPTP..."
    # shellcheck disable=SC2024
    sudo pppd call tunnel nodetach < /dev/null > /tmp/pptp.log 2>&1 &

    local attempt=0
    while [ $attempt -lt 15 ]; do
        if ip link show ppp0 &>/dev/null; then
            local vpn_ip
            vpn_ip=$(ip -4 addr show ppp0 2>/dev/null | grep -oP 'inet \K[\d.]+' || echo "unknown")
            log_success "VPN connected via PPTP (ppp0: $vpn_ip)"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    log_warning "PPTP started but ppp0 not detected after 15s"
}

# --- Main VPN auto-connect orchestrator ---
init_vpn() {
    # Skip if no VPN tools installed at all
    local has_vpn=false
    command -v openvpn &>/dev/null && has_vpn=true
    command -v wg &>/dev/null && has_vpn=true
    command -v ipsec &>/dev/null && has_vpn=true
    command -v pptp &>/dev/null && has_vpn=true
    if [ "$has_vpn" = "false" ]; then
        log_debug "No VPN clients installed, skipping"
        return 0
    fi

    log_info "VPN clients detected, checking configuration..."

    # Skip if already connected (any protocol)
    if pgrep -x openvpn &>/dev/null || ip link show wg0 &>/dev/null 2>&1 || \
       pgrep -x charon &>/dev/null || pgrep -x pppd &>/dev/null; then
        log_info "VPN already connected, skipping"
        return 0
    fi

    # Backward compatible: support both VPN_CONFIG_REF and OPENVPN_CONFIG_REF
    local vpn_ref="${VPN_CONFIG_REF:-${OPENVPN_CONFIG_REF:-}}"

    # Source 1: Resolve from 1Password vault
    if [ -n "$vpn_ref" ] && [ -n "${OP_SERVICE_ACCOUNT_TOKEN:-}" ] && command -v op &> /dev/null; then
        local ref="${vpn_ref#op://}"
        local vault profile
        vault=$(echo "$ref" | cut -d'/' -f1)
        profile=$(echo "$ref" | cut -d'/' -f2)

        log_info "Resolving VPN profile '$profile' from 1Password ($vault)..."

        # Resolve DOCUMENT UUID and detect protocol from tags
        local doc_item doc_uuid protocol tags
        doc_item=$(op item list --vault "$vault" --categories DOCUMENT --format json 2>/dev/null \
            | jq -r --arg t "$profile" '.[] | select(.title==$t)' 2>/dev/null || echo "")
        doc_uuid=$(echo "$doc_item" | jq -r '.id // empty' 2>/dev/null || echo "")
        tags=$(echo "$doc_item" | jq -r '.tags // [] | .[]' 2>/dev/null || echo "")

        # Detect protocol (default: openvpn)
        protocol="openvpn"
        for tag in $tags; do
            case "$tag" in
                wireguard|ipsec|pptp) protocol="$tag"; break ;;
            esac
        done

        # Resolve LOGIN UUID
        local login_uuid
        login_uuid=$(op item list --vault "$vault" --categories LOGIN --format json 2>/dev/null \
            | jq -r --arg t "$profile" '.[] | select(.title==$t) | .id' 2>/dev/null || echo "")

        log_info "VPN profile '$profile' → protocol: $protocol"

        # Protocol-specific connect
        case "$protocol" in
            openvpn)   connect_openvpn "$vault" "$profile" "$doc_uuid" "$login_uuid" ;;
            wireguard) connect_wireguard "$vault" "$profile" "$doc_uuid" ;;
            ipsec)     connect_ipsec "$vault" "$profile" "$doc_uuid" "$login_uuid" ;;
            pptp)      connect_pptp "$vault" "$profile" "$doc_uuid" "$login_uuid" ;;
        esac
        return 0
    fi

    # Source 2: File on disk (OpenVPN fallback for backward compat)
    local ovpn_config="${OPENVPN_CONFIG:-/home/vscode/.config/openvpn/client.ovpn}"
    if [ -f "$ovpn_config" ] && [ -s "$ovpn_config" ] && command -v openvpn &>/dev/null; then
        log_info "Found OpenVPN config on disk, connecting..."
        connect_openvpn "" "" "" ""
    else
        log_info "No VPN config found, skipping"
    fi

    return 0
}

# ============================================================================
# Execution
# ============================================================================

run_step "Restore Claude config"    step_restore_claude_config
run_step "Init Claude dirs"         step_init_claude_dirs
run_step "Shell env repair"         step_shell_env_repair
run_step "Reload environment"       step_reload_env
run_step "1Password permissions"    step_1password_permissions
run_step "npm cache permissions"    step_npm_cache_permissions
run_step "MCP configuration"        step_mcp_configuration
run_step "Git credential cleanup"   step_git_credential_cleanup

# Background tasks (have internal error handling, not tracked in summary)
init_semantic_search >> /tmp/grepai-init.log 2>&1 &
init_vpn &

# Export dynamic environment variables (appended to ~/.devcontainer-env.sh)
# Note: ~/.devcontainer-env.sh is created by postCreate.sh with static content

run_step "Project init check"       step_auto_init_check

print_step_summary "postStart"

log_success "postStart: Container ready!"
