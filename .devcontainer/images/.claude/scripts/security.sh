#!/bin/bash
# Security scan for secrets and vulnerabilities
# Usage: security.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

ISSUES_FOUND=0

# Check for secrets with detect-secrets
if command -v detect-secrets &>/dev/null; then
    if detect-secrets scan "$FILE" 2>/dev/null | grep -q '"results":\s*{[^}]*}'; then
        echo "⚠️  Potential secret detected in $FILE"
        ISSUES_FOUND=1
    fi
fi

# Check for secrets with trivy
if command -v trivy &>/dev/null; then
    RESULT=$(trivy fs --scanners secret --quiet "$FILE" 2>/dev/null || true)
    if [ -n "$RESULT" ] && echo "$RESULT" | grep -qi "secret\|password\|token\|key"; then
        echo "⚠️  Trivy found potential secrets in $FILE"
        ISSUES_FOUND=1
    fi
fi

# Check for secrets with gitleaks
if command -v gitleaks &>/dev/null; then
    if ! gitleaks detect --source "$FILE" --no-git --quiet 2>/dev/null; then
        echo "⚠️  Gitleaks found potential secrets in $FILE"
        ISSUES_FOUND=1
    fi
fi

# Simple pattern-based checks (fallback)
if [ $ISSUES_FOUND -eq 0 ]; then
    # Check for common secret patterns
    PATTERNS=(
        'password\s*=\s*["\047][^"\047]+'
        'api[_-]?key\s*=\s*["\047][^"\047]+'
        'secret[_-]?key\s*=\s*["\047][^"\047]+'
        'aws[_-]?access[_-]?key'
        'private[_-]?key'
        'BEGIN RSA PRIVATE KEY'
        'BEGIN OPENSSH PRIVATE KEY'
    )

    for PATTERN in "${PATTERNS[@]}"; do
        if grep -iEq "$PATTERN" "$FILE" 2>/dev/null; then
            echo "⚠️  Potential secret pattern found in $FILE: $PATTERN"
            ISSUES_FOUND=1
            break
        fi
    done
fi

exit 0
