#!/bin/bash
# Lint files based on extension
# Usage: lint.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"
DIR=$(dirname "$FILE")
BASENAME=$(basename "$FILE")

case "$EXT" in
    # JavaScript/TypeScript
    js|jsx|ts|tsx|mjs|cjs)
        if command -v eslint &>/dev/null; then
            eslint --fix "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null && [ -f "package.json" ]; then
            npx eslint --fix "$FILE" 2>/dev/null || true
        fi
        ;;

    # Python - ruff is faster and preferred, pylint as fallback
    py)
        if command -v ruff &>/dev/null; then
            ruff check --fix "$FILE" 2>/dev/null || true
        elif command -v pylint &>/dev/null; then
            pylint --errors-only "$FILE" 2>/dev/null || true
        fi
        # Type checking is separate (in typecheck.sh)
        ;;

    # Go - golangci-lint is comprehensive, go vet is basic
    go)
        if command -v golangci-lint &>/dev/null; then
            golangci-lint run --fix "$FILE" 2>/dev/null || true
        elif command -v staticcheck &>/dev/null; then
            staticcheck "$FILE" 2>/dev/null || true
        elif command -v go &>/dev/null; then
            go vet "$FILE" 2>/dev/null || true
        fi
        ;;

    # Rust - clippy is the standard linter
    rs)
        [[ -f "$HOME/.cache/cargo/env" ]] && source "$HOME/.cache/cargo/env"
        if command -v cargo &>/dev/null; then
            (cd "$DIR" && cargo clippy --fix --allow-dirty --allow-staged -- -D warnings 2>/dev/null) || true
        fi
        ;;

    # Shell
    sh|bash)
        if command -v shellcheck &>/dev/null; then
            shellcheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # Dockerfile
    Dockerfile*)
        if command -v hadolint &>/dev/null; then
            hadolint "$FILE" 2>/dev/null || true
        fi
        ;;

    # YAML
    yml|yaml)
        if command -v yamllint &>/dev/null; then
            yamllint -d relaxed "$FILE" 2>/dev/null || true
        fi
        ;;

    # Terraform
    tf|tfvars)
        if command -v tflint &>/dev/null; then
            tflint "$FILE" 2>/dev/null || true
        fi
        # terraform validate needs to run in module directory
        ;;

    # C/C++
    c|cpp|cc|cxx|h|hpp)
        if command -v clang-tidy &>/dev/null; then
            clang-tidy "$FILE" --fix 2>/dev/null || true
        elif command -v cppcheck &>/dev/null; then
            cppcheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # Java
    java)
        if command -v checkstyle &>/dev/null; then
            checkstyle "$FILE" 2>/dev/null || true
        fi
        ;;

    # Ruby
    rb)
        if command -v rubocop &>/dev/null; then
            rubocop -a "$FILE" 2>/dev/null || true
        fi
        ;;

    # PHP
    php)
        if command -v phpstan &>/dev/null; then
            phpstan analyse "$FILE" 2>/dev/null || true
        elif command -v php &>/dev/null; then
            php -l "$FILE" 2>/dev/null || true
        fi
        ;;

    # Kotlin
    kt|kts)
        if command -v ktlint &>/dev/null; then
            ktlint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Swift
    swift)
        if command -v swiftlint &>/dev/null; then
            swiftlint lint --path "$FILE" 2>/dev/null || true
        fi
        ;;

    # Lua
    lua)
        if command -v luacheck &>/dev/null; then
            luacheck "$FILE" 2>/dev/null || true
        fi
        ;;

    # SQL
    sql)
        if command -v sqlfluff &>/dev/null; then
            sqlfluff lint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Markdown
    md)
        if command -v markdownlint &>/dev/null; then
            markdownlint "$FILE" 2>/dev/null || true
        fi
        ;;

    # JSON
    json)
        if command -v jsonlint &>/dev/null; then
            jsonlint -q "$FILE" 2>/dev/null || true
        fi
        ;;

    # HTML
    html|htm)
        if command -v htmlhint &>/dev/null; then
            htmlhint "$FILE" 2>/dev/null || true
        fi
        ;;

    # CSS/SCSS
    css|scss|less)
        if command -v stylelint &>/dev/null; then
            stylelint --fix "$FILE" 2>/dev/null || true
        fi
        ;;

    # Elixir
    ex|exs)
        if command -v mix &>/dev/null; then
            mix credo "$FILE" 2>/dev/null || true
        fi
        ;;

    # Dart
    dart)
        if command -v dart &>/dev/null; then
            dart analyze "$FILE" 2>/dev/null || true
        fi
        ;;

    # Zig
    zig)
        if command -v zig &>/dev/null; then
            zig build --summary all 2>/dev/null || true
        fi
        ;;

    # TOML
    toml)
        if command -v taplo &>/dev/null; then
            taplo lint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Protobuf
    proto)
        if command -v buf &>/dev/null; then
            buf lint "$FILE" 2>/dev/null || true
        fi
        ;;

    # Ansible
    yml|yaml)
        # Already handled above, but ansible-lint for playbooks
        if [[ "$FILE" == *"playbook"* ]] || [[ "$FILE" == *"ansible"* ]]; then
            if command -v ansible-lint &>/dev/null; then
                ansible-lint "$FILE" 2>/dev/null || true
            fi
        fi
        ;;
esac

exit 0
