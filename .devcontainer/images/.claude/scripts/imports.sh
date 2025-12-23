#!/bin/bash
# Sort and organize imports based on extension
# Usage: imports.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"

case "$EXT" in
    # JavaScript/TypeScript
    js|jsx|ts|tsx|mjs|cjs)
        # Try eslint with import plugin, then prettier-plugin-organize-imports
        if command -v eslint &>/dev/null; then
            eslint --fix --rule 'import/order: error' "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null; then
            npx eslint --fix --rule 'import/order: error' "$FILE" 2>/dev/null || true
        fi
        ;;

    # Python
    py)
        # isort is the standard, ruff can also do it
        if command -v isort &>/dev/null; then
            isort --quiet "$FILE" 2>/dev/null || true
        fi
        # ruff is complementary - also handles import sorting
        if command -v ruff &>/dev/null; then
            ruff check --select I --fix "$FILE" 2>/dev/null || true
        fi
        ;;

    # Go - goimports is the ONLY tool that sorts imports
    # gofmt does NOT sort imports, only formats
    go)
        if command -v goimports &>/dev/null; then
            goimports -w "$FILE" 2>/dev/null || true
        fi
        # No fallback - gofmt doesn't sort imports
        ;;

    # Rust - rustfmt handles imports ordering automatically
    rs)
        if command -v rustfmt &>/dev/null; then
            rustfmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Java
    java)
        if command -v google-java-format &>/dev/null; then
            google-java-format --replace "$FILE" 2>/dev/null || true
        fi
        ;;

    # C/C++ - clang-format can sort includes
    c|cpp|cc|cxx|h|hpp)
        if command -v clang-format &>/dev/null; then
            clang-format -i --sort-includes "$FILE" 2>/dev/null || true
        fi
        ;;
esac

exit 0
