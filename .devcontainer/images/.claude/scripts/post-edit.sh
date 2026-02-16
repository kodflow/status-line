#!/bin/bash
# Combined post-edit hook: format + lint + typecheck
# Usage: post-edit.sh <file_path>
# Note: format.sh handles imports (goimports, ruff, rustfmt, etc.)

set +e  # Fail-open: hooks should never block unexpectedly

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FILE="${1:-}"

if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
    exit 0
fi

# Skip format/lint for documentation and config files
if [[ "$FILE" == *".claude/plans/"* ]] || \
   [[ "$FILE" == *".claude/sessions/"* ]] || \
   [[ "$FILE" == */plans/* ]] || \
   [[ "$FILE" == *.md ]] || \
   [[ "$FILE" == /tmp/* ]] || \
   [[ "$FILE" == /home/vscode/.claude/* ]]; then
    exit 0
fi

# === Format/Lint/Types pipeline ===

# 1. Format (includes import sorting via goimports, ruff, rustfmt, etc.)
"$SCRIPT_DIR/format.sh" "$FILE"

# 2. Lint (with auto-fix)
"$SCRIPT_DIR/lint.sh" "$FILE"

# 3. Type check (academic rigor)
"$SCRIPT_DIR/typecheck.sh" "$FILE"

exit 0
