#!/bin/bash
# test-validation.sh - Tests de validation de schéma invalide
# Usage: ./test-validation.sh

set -euo pipefail

TEST_DIR=$(mktemp -d)
FAILED=0
PASSED=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

log_pass() {
    echo -e "${GREEN}✓${NC} $1"
    PASSED=$((PASSED + 1))
}

log_fail() {
    echo -e "${RED}✗${NC} $1"
    FAILED=$((FAILED + 1))
}

# Fonction de validation (simule task-validate.sh)
validate_session() {
    local session_file="$1"

    # Vérifier schéma v2
    local schema_version
    schema_version=$(jq -r '.schemaVersion // 1' "$session_file" 2>/dev/null || echo "1")

    if [[ "$schema_version" == "2" ]]; then
        # Vérifier invariants
        local type project state
        type=$(jq -r '.type // ""' "$session_file" 2>/dev/null)
        project=$(jq -r '.project // ""' "$session_file" 2>/dev/null)
        state=$(jq -r '.state // ""' "$session_file" 2>/dev/null)

        # Type obligatoire
        if [[ -z "$type" ]]; then
            echo "INVALID:type_missing"
            return 1
        fi

        # Type valide
        if [[ ! "$type" =~ ^(feature|fix)$ ]]; then
            echo "INVALID:type_invalid"
            return 1
        fi

        # Project obligatoire
        if [[ -z "$project" ]]; then
            echo "INVALID:project_missing"
            return 1
        fi

        # State valide
        if [[ ! "$state" =~ ^(planning|planned|applying|applied)$ ]]; then
            echo "INVALID:state_invalid"
            return 1
        fi
    fi

    echo "VALID"
    return 0
}

# === Test 1: Session valide ===
test_valid_session() {
    echo "Test: Session v2 valide..."

    cat > "$TEST_DIR/valid.json" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "type": "feature",
    "project": "test-project",
    "branch": "feat/test-project"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/valid.json")
    if [[ "$result" == "VALID" ]]; then
        log_pass "Session valide acceptée"
    else
        log_fail "Session valide rejetée: $result"
    fi
}

# === Test 2: Type manquant ===
test_missing_type() {
    echo "Test: Type manquant..."

    cat > "$TEST_DIR/no-type.json" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "project": "test"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/no-type.json") || true
    if [[ "$result" == "INVALID:type_missing" ]]; then
        log_pass "Type manquant détecté"
    else
        log_fail "Type manquant non détecté: $result"
    fi
}

# === Test 3: Type invalide ===
test_invalid_type() {
    echo "Test: Type invalide..."

    cat > "$TEST_DIR/bad-type.json" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "type": "hotfix",
    "project": "test"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/bad-type.json") || true
    if [[ "$result" == "INVALID:type_invalid" ]]; then
        log_pass "Type invalide détecté"
    else
        log_fail "Type invalide non détecté: $result"
    fi
}

# === Test 4: Project manquant ===
test_missing_project() {
    echo "Test: Project manquant..."

    cat > "$TEST_DIR/no-project.json" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "type": "feature"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/no-project.json") || true
    if [[ "$result" == "INVALID:project_missing" ]]; then
        log_pass "Project manquant détecté"
    else
        log_fail "Project manquant non détecté: $result"
    fi
}

# === Test 5: State invalide ===
test_invalid_state() {
    echo "Test: State invalide..."

    cat > "$TEST_DIR/bad-state.json" << 'EOF'
{
    "schemaVersion": 2,
    "state": "in_progress",
    "type": "feature",
    "project": "test"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/bad-state.json") || true
    if [[ "$result" == "INVALID:state_invalid" ]]; then
        log_pass "State invalide détecté"
    else
        log_fail "State invalide non détecté: $result"
    fi
}

# === Test 6: Schema v1 (pas de validation stricte) ===
test_schema_v1_no_validation() {
    echo "Test: Schema v1 pas de validation stricte..."

    cat > "$TEST_DIR/v1.json" << 'EOF'
{
    "schemaVersion": 1,
    "mode": "plan"
}
EOF

    local result
    result=$(validate_session "$TEST_DIR/v1.json")
    if [[ "$result" == "VALID" ]]; then
        log_pass "Schema v1 accepté sans validation stricte"
    else
        log_fail "Schema v1 devrait être accepté: $result"
    fi
}

# === Exécution ===
echo "═══════════════════════════════════════════════"
echo "  Tests validation schema"
echo "═══════════════════════════════════════════════"
echo ""

test_valid_session
test_missing_type
test_invalid_type
test_missing_project
test_invalid_state
test_schema_v1_no_validation

echo ""
echo "═══════════════════════════════════════════════"
echo "  Résultats: ${PASSED} passés, ${FAILED} échoués"
echo "═══════════════════════════════════════════════"

exit "$FAILED"
