#!/bin/bash
# Combined post-edit hook: format + imports + lint
# Usage: post-edit.sh <file_path>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FILE="$1"

if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

# 1. Format
"$SCRIPT_DIR/format.sh" "$FILE"

# 2. Sort imports
"$SCRIPT_DIR/imports.sh" "$FILE"

# 3. Lint (with auto-fix)
"$SCRIPT_DIR/lint.sh" "$FILE"

exit 0
