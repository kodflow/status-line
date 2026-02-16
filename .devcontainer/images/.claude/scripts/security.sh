#!/bin/bash
# Security scan for secrets and vulnerabilities
# Usage:
#   security.sh <file_path>           # PostToolUse mode (single file)
#   security.sh                       # PreToolUse mode (git commit - scan staged files)
#
# Exit 0 = OK, Exit 2 = blocked (secrets found)

set +e  # Fail-open: don't exit on errors (hook should never block unexpectedly)

# === Single file scan function ===
scan_file() {
    local file="$1"
    local issues=0

    # Skip binary files
    if file "$file" 2>/dev/null | grep -q "binary"; then
        return 0
    fi

    # Check for secrets with detect-secrets
    if command -v detect-secrets &>/dev/null; then
        if detect-secrets scan "$file" 2>/dev/null | grep -q '"results":\s*{[^}]*}'; then
            echo "⚠️  Potential secret detected in $file"
            issues=1
        fi
    fi

    # Check for secrets with trivy
    if command -v trivy &>/dev/null; then
        RESULT=$(trivy fs --scanners secret --quiet "$file" 2>/dev/null || true)
        if [ -n "$RESULT" ] && echo "$RESULT" | grep -qi "secret\|password\|token\|key"; then
            echo "⚠️  Trivy found potential secrets in $file"
            issues=1
        fi
    fi

    # Check for secrets with gitleaks
    if command -v gitleaks &>/dev/null; then
        if ! gitleaks detect --source "$file" --no-git --quiet 2>/dev/null; then
            echo "⚠️  Gitleaks found potential secrets in $file"
            issues=1
        fi
    fi

    # Simple pattern-based checks (fallback)
    if [ $issues -eq 0 ]; then
        PATTERNS=(
            'password\s*=\s*["\047][^"\047]+'
            'api[_-]?key\s*=\s*["\047][^"\047]+'
            'secret[_-]?key\s*=\s*["\047][^"\047]+'
            'aws[_-]?access[_-]?key'
            'private[_-]?key'
            'BEGIN RSA PRIVATE KEY'
            'BEGIN OPENSSH PRIVATE KEY'
            'ghp_[a-zA-Z0-9]{36}'      # GitHub PAT
            'gho_[a-zA-Z0-9]{36}'      # GitHub OAuth
            'github_pat_[a-zA-Z0-9_]+'  # GitHub PAT (new format)
            'sk-[a-zA-Z0-9]{48}'        # OpenAI API key
            'AKIA[0-9A-Z]{16}'          # AWS Access Key ID
        )

        for PATTERN in "${PATTERNS[@]}"; do
            if grep -iEq "$PATTERN" "$file" 2>/dev/null; then
                echo "⚠️  Potential secret pattern found in $file"
                issues=1
                break
            fi
        done
    fi

    return $issues
}

# === Determine mode based on input ===
INPUT=""
FILE="${1:-}"

if [ ! -t 0 ]; then
    # stdin has data, might be JSON from Claude hook
    INPUT=$(cat)
    TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)
    COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null || true)

    # Check if this is a git commit/push command
    if [[ "$TOOL" == "Bash" ]] && [[ "$COMMAND" =~ ^git[[:space:]]+(commit|push) ]]; then
        # === Auto-correct git push --force to --force-with-lease ===
        if [[ "$COMMAND" =~ ^git[[:space:]]+push ]] && \
           [[ "$COMMAND" =~ --force ]] && \
           [[ ! "$COMMAND" =~ --force-with-lease ]]; then
            CORRECTED="${COMMAND/--force/--force-with-lease}"
            echo "⚠️  Auto-corrected: --force → --force-with-lease" >&2
            if command -v jq &>/dev/null; then
                jq -n --arg cmd "$CORRECTED" '{"decision":"allow","updatedInput":{"command":$cmd}}'
            else
                printf '{"decision":"allow","updatedInput":{"command":"%s"}}' "$CORRECTED"
            fi
            exit 0
        fi

        # Scan all staged files
        STAGED_FILES=$(git diff --cached --name-only 2>/dev/null || true)
        if [ -z "$STAGED_FILES" ]; then
            echo "✓ No staged files to scan" >&2
            exit 0
        fi

        ISSUES_FOUND=0
        while IFS= read -r f; do
            if [ -f "$f" ]; then
                if ! scan_file "$f"; then
                    ISSUES_FOUND=1
                fi
            fi
        done <<< "$STAGED_FILES"

        if [ $ISSUES_FOUND -eq 1 ]; then
            echo "═══════════════════════════════════════════════"
            echo "  ⚠️  COMMIT BLOCKED - Secrets detected"
            echo "═══════════════════════════════════════════════"
            echo ""
            echo "  Potential secrets were found in the staged"
            echo "  files. Please remove them before committing."
            echo ""
            echo "═══════════════════════════════════════════════"
            exit 2
        fi
        echo "✓ Security scan passed" >&2
        exit 0
    fi
    # Not a git commit/push command, nothing to do
    exit 0
fi

# === PostToolUse mode: single file ===
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    # No file to scan (stdin mode without git commit, or invalid file)
    exit 0
fi

if scan_file "$FILE"; then
    echo "✓ $FILE: no secrets found" >&2
fi
exit 0
