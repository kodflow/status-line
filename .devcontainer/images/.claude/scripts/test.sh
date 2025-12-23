#!/bin/bash
# Run tests for modified files
# Usage: test.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
BASENAME=$(basename "$FILE")
DIR=$(dirname "$FILE")

# Check if this is a test file
IS_TEST=0
case "$BASENAME" in
    *.test.*|*.spec.*|*_test.*|test_*)
        IS_TEST=1
        ;;
esac

case "$EXT" in
    # JavaScript/TypeScript
    js|jsx|ts|tsx)
        if [ $IS_TEST -eq 1 ]; then
            if command -v jest &>/dev/null; then
                jest "$FILE" --passWithNoTests 2>/dev/null || true
            elif command -v npx &>/dev/null; then
                npx jest "$FILE" --passWithNoTests 2>/dev/null || true
            elif command -v vitest &>/dev/null; then
                vitest run "$FILE" 2>/dev/null || true
            fi
        fi
        ;;

    # Python
    py)
        if [ $IS_TEST -eq 1 ]; then
            if command -v pytest &>/dev/null; then
                pytest "$FILE" -v 2>/dev/null || true
            elif command -v python &>/dev/null; then
                python -m pytest "$FILE" -v 2>/dev/null || true
            fi
        fi
        ;;

    # Go
    go)
        if [[ "$BASENAME" == *"_test.go" ]]; then
            if command -v go &>/dev/null; then
                (cd "$DIR" && go test -v -run . 2>/dev/null) || true
            fi
        fi
        ;;

    # Rust
    rs)
        if command -v cargo &>/dev/null; then
            if [[ "$FILE" == *"tests"* ]] || grep -q "#\[test\]" "$FILE" 2>/dev/null; then
                (cd "$DIR" && cargo test 2>/dev/null) || true
            fi
        fi
        ;;
esac

exit 0
