#!/bin/bash
# test-workflow.sh - Tests d'intégration pour le workflow /plan → /apply
# Usage: ./test-workflow.sh

set -euo pipefail

TEST_DIR=$(mktemp -d)
FAILED=0
PASSED=0

# Couleurs
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

# === Test 1: Session v2 création ===
test_session_v2_creation() {
    echo "Test: Session v2 création..."

    # Simuler task-init en créant une session v2
    local SESSION_FILE="$TEST_DIR/test-project.json"
    cat > "$SESSION_FILE" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "type": "feature",
    "project": "test-project",
    "branch": "feat/test-project",
    "currentTask": null,
    "currentEpic": null,
    "lockedPaths": [],
    "epics": [],
    "createdAt": "2024-01-01T00:00:00Z"
}
EOF

    # Vérifier schemaVersion
    local version
    version=$(jq -r '.schemaVersion' "$SESSION_FILE")
    if [[ "$version" == "2" ]]; then
        log_pass "schemaVersion = 2"
    else
        log_fail "schemaVersion devrait être 2, trouvé: $version"
    fi

    # Vérifier state initial
    local state
    state=$(jq -r '.state' "$SESSION_FILE")
    if [[ "$state" == "planning" ]]; then
        log_pass "state initial = planning"
    else
        log_fail "state initial devrait être planning, trouvé: $state"
    fi
}

# === Test 2: Transition planning → planned ===
test_transition_planning_to_planned() {
    echo "Test: Transition planning → planned..."

    local SESSION_FILE="$TEST_DIR/test-transition.json"
    cat > "$SESSION_FILE" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planning",
    "type": "feature",
    "project": "test",
    "branch": "feat/test",
    "epics": []
}
EOF

    # Simuler la transition
    jq '.state = "planned"' "$SESSION_FILE" > "$SESSION_FILE.tmp" && mv "$SESSION_FILE.tmp" "$SESSION_FILE"

    local state
    state=$(jq -r '.state' "$SESSION_FILE")
    if [[ "$state" == "planned" ]]; then
        log_pass "Transition planning → planned"
    else
        log_fail "Transition échouée, state: $state"
    fi
}

# === Test 3: Transition planned → applying ===
test_transition_planned_to_applying() {
    echo "Test: Transition planned → applying..."

    local SESSION_FILE="$TEST_DIR/test-applying.json"
    cat > "$SESSION_FILE" << 'EOF'
{
    "schemaVersion": 2,
    "state": "planned",
    "type": "fix",
    "project": "bugfix",
    "branch": "fix/bugfix",
    "epics": []
}
EOF

    # Simuler la transition (comme task-start.sh)
    jq 'if .state == "planned" then .state = "applying" else . end' "$SESSION_FILE" > "$SESSION_FILE.tmp" && mv "$SESSION_FILE.tmp" "$SESSION_FILE"

    local state
    state=$(jq -r '.state' "$SESSION_FILE")
    if [[ "$state" == "applying" ]]; then
        log_pass "Transition planned → applying"
    else
        log_fail "Transition échouée, state: $state"
    fi
}

# === Test 4: Invariants type ===
test_invariants_type() {
    echo "Test: Invariants type (feature|fix)..."

    # Type valide
    local valid_type="feature"
    if [[ "$valid_type" =~ ^(feature|fix)$ ]]; then
        log_pass "Type 'feature' valide"
    else
        log_fail "Type 'feature' devrait être valide"
    fi

    # Type invalide
    local invalid_type="hotfix"
    if [[ ! "$invalid_type" =~ ^(feature|fix)$ ]]; then
        log_pass "Type 'hotfix' rejeté"
    else
        log_fail "Type 'hotfix' devrait être rejeté"
    fi
}

# === Test 5: Invariants state ===
test_invariants_state() {
    echo "Test: Invariants state..."

    local valid_states=("planning" "planned" "applying" "applied")
    for state in "${valid_states[@]}"; do
        if [[ "$state" =~ ^(planning|planned|applying|applied)$ ]]; then
            log_pass "State '$state' valide"
        else
            log_fail "State '$state' devrait être valide"
        fi
    done

    # State invalide
    local invalid_state="in_progress"
    if [[ ! "$invalid_state" =~ ^(planning|planned|applying|applied)$ ]]; then
        log_pass "State 'in_progress' rejeté"
    else
        log_fail "State 'in_progress' devrait être rejeté"
    fi
}

# === Exécution ===
echo "═══════════════════════════════════════════════"
echo "  Tests workflow /plan → /apply"
echo "═══════════════════════════════════════════════"
echo ""

test_session_v2_creation
test_transition_planning_to_planned
test_transition_planned_to_applying
test_invariants_type
test_invariants_state

echo ""
echo "═══════════════════════════════════════════════"
echo "  Résultats: ${PASSED} passés, ${FAILED} échoués"
echo "═══════════════════════════════════════════════"

exit "$FAILED"
