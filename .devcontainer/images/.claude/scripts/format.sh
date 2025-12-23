#!/bin/bash
# Auto-format files based on extension
# Usage: format.sh <file_path>

set -e

FILE="$1"
if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

EXT="${FILE##*.}"

case "$EXT" in
    # JavaScript/TypeScript - prettier is the standard
    js|jsx|ts|tsx|mjs|cjs)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v npx &>/dev/null; then
            npx prettier --write "$FILE" 2>/dev/null || true
        fi
        ;;

    # Python - black OR ruff (not both, they conflict)
    py)
        if command -v ruff &>/dev/null; then
            ruff format "$FILE" 2>/dev/null || true
        elif command -v black &>/dev/null; then
            black --quiet "$FILE" 2>/dev/null || true
        elif command -v autopep8 &>/dev/null; then
            autopep8 --in-place "$FILE" 2>/dev/null || true
        fi
        ;;

    # Go - prefer goimports (format + imports), fallback to gofmt
    go)
        if command -v goimports &>/dev/null; then
            goimports -w "$FILE" 2>/dev/null || true
        elif command -v gofmt &>/dev/null; then
            gofmt -w "$FILE" 2>/dev/null || true
        fi
        ;;

    # Rust
    rs)
        if command -v rustfmt &>/dev/null; then
            rustfmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # JSON - prettier or jq
    json)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v jq &>/dev/null; then
            TMP=$(mktemp)
            if jq '.' "$FILE" > "$TMP" 2>/dev/null; then
                mv "$TMP" "$FILE"
            else
                rm -f "$TMP"
            fi
        fi
        ;;

    # YAML
    yml|yaml)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        elif command -v yamlfmt &>/dev/null; then
            yamlfmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Markdown
    md)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        fi
        ;;

    # Terraform
    tf|tfvars)
        if command -v terraform &>/dev/null; then
            terraform fmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Shell
    sh|bash)
        if command -v shfmt &>/dev/null; then
            shfmt -w "$FILE" 2>/dev/null || true
        fi
        ;;

    # C/C++
    c|cpp|cc|cxx|h|hpp)
        if command -v clang-format &>/dev/null; then
            clang-format -i "$FILE" 2>/dev/null || true
        fi
        ;;

    # Java
    java)
        if command -v google-java-format &>/dev/null; then
            google-java-format --replace "$FILE" 2>/dev/null || true
        fi
        ;;

    # HTML/CSS/SCSS
    html|htm|css|scss|less)
        if command -v prettier &>/dev/null; then
            prettier --write "$FILE" 2>/dev/null || true
        fi
        ;;

    # XML
    xml)
        if command -v xmllint &>/dev/null; then
            xmllint --format "$FILE" --output "$FILE" 2>/dev/null || true
        fi
        ;;

    # SQL
    sql)
        if command -v sql-formatter &>/dev/null; then
            sql-formatter "$FILE" -o "$FILE" 2>/dev/null || true
        elif command -v pg_format &>/dev/null; then
            pg_format -i "$FILE" 2>/dev/null || true
        fi
        ;;

    # Lua
    lua)
        if command -v stylua &>/dev/null; then
            stylua "$FILE" 2>/dev/null || true
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
        if command -v php-cs-fixer &>/dev/null; then
            php-cs-fixer fix "$FILE" --quiet 2>/dev/null || true
        fi
        ;;

    # Kotlin
    kt|kts)
        if command -v ktlint &>/dev/null; then
            ktlint -F "$FILE" 2>/dev/null || true
        fi
        ;;

    # Swift
    swift)
        if command -v swiftformat &>/dev/null; then
            swiftformat "$FILE" 2>/dev/null || true
        fi
        ;;

    # Dart
    dart)
        if command -v dart &>/dev/null; then
            dart format "$FILE" 2>/dev/null || true
        fi
        ;;

    # Elixir
    ex|exs)
        if command -v mix &>/dev/null; then
            mix format "$FILE" 2>/dev/null || true
        fi
        ;;

    # Zig
    zig)
        if command -v zig &>/dev/null; then
            zig fmt "$FILE" 2>/dev/null || true
        fi
        ;;

    # Nim
    nim)
        if command -v nimpretty &>/dev/null; then
            nimpretty "$FILE" 2>/dev/null || true
        fi
        ;;

    # TOML
    toml)
        if command -v taplo &>/dev/null; then
            taplo fmt "$FILE" 2>/dev/null || true
        fi
        ;;
esac

exit 0
