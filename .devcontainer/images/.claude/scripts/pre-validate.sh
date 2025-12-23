#!/bin/bash
# Pre-validation before file modifications
# Usage: pre-validate.sh <file_path>
# Exit code 0 = allow, non-zero = block

FILE="$1"
if [ -z "$FILE" ]; then
    exit 0
fi

# Protected paths (modify as needed)
PROTECTED_PATTERNS=(
    "node_modules/"
    ".git/"
    "vendor/"
    "dist/"
    "build/"
    ".env"
    "*.lock"
    "package-lock.json"
    "yarn.lock"
    "pnpm-lock.yaml"
    "Cargo.lock"
    "poetry.lock"
    "go.sum"
)

# Check if file matches protected patterns
for PATTERN in "${PROTECTED_PATTERNS[@]}"; do
    if [[ "$FILE" == *"$PATTERN"* ]]; then
        echo "🚫 Protected file: $FILE"
        echo "   Pattern: $PATTERN"
        exit 1
    fi
done

# Check for branch protection (prevent commits to main/master)
if [[ "$FILE" == *"git commit"* ]] || [[ "$FILE" == *"git push"* ]]; then
    BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    if [[ "$BRANCH" == "main" ]] || [[ "$BRANCH" == "master" ]]; then
        echo "🚫 Direct commits to $BRANCH are discouraged"
        echo "   Consider using a feature branch"
        # Warning only, don't block (exit 0)
    fi
fi

exit 0
