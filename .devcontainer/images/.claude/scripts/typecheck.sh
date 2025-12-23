#!/bin/bash
# Type check files based on extension
# Usage: typecheck.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
DIR=$(dirname "$FILE")

case "$EXT" in
    # TypeScript
    ts|tsx)
        if command -v tsc &>/dev/null; then
            tsc --noEmit "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null && [ -f "tsconfig.json" ]; then
            npx tsc --noEmit 2>/dev/null || true
        fi
        ;;

    # Python (with type hints)
    py)
        if command -v mypy &>/dev/null; then
            mypy --ignore-missing-imports "$FILE" 2>/dev/null || true
        elif command -v pyright &>/dev/null; then
            pyright "$FILE" 2>/dev/null || true
        fi
        ;;

    # Go (built-in type checking via go build)
    go)
        if command -v go &>/dev/null; then
            go build -o /dev/null "$FILE" 2>/dev/null || true
        fi
        ;;

    # Rust (cargo check)
    rs)
        if command -v cargo &>/dev/null; then
            (cd "$DIR" && cargo check 2>/dev/null) || true
        fi
        ;;
esac

exit 0
