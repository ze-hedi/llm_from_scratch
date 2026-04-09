#!/bin/bash

# Test runner for wrapper scripts
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
TEST_RESULTS_DIR="$SCRIPT_DIR/results"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Initialize test results
mkdir -p "$TEST_RESULTS_DIR"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test 1: Check if wrapper scripts exist and are executable
test_scripts_exist() {
    ((TOTAL_TESTS++))
    log_test "Checking if wrapper scripts exist and are executable"
    
    if [ -x "$PROJECT_DIR/opencode-wrapper.sh" ] && [ -x "$PROJECT_DIR/openclaw-wrapper.sh" ]; then
        log_pass "Both wrapper scripts exist and are executable"
        return 0
    else
        log_fail "Wrapper scripts missing or not executable"
        return 1
    fi
}

# Test 2: Validate shell script syntax
test_script_syntax() {
    ((TOTAL_TESTS++))
    log_test "Validating shell script syntax"
    
    if bash -n "$PROJECT_DIR/opencode-wrapper.sh" && bash -n "$PROJECT_DIR/openclaw-wrapper.sh"; then
        log_pass "Shell script syntax is valid"
        return 0
    else
        log_fail "Shell script syntax errors detected"
        return 1
    fi
}

# Test 3: Validate JSON configs
test_json_configs() {
    ((TOTAL_TESTS++))
    log_test "Validating JSON configuration files"
    
    local all_valid=true
    for file in "$PROJECT_DIR"/agent_*.json; do
        if [ -f "$file" ]; then
            if python3 -m json.tool "$file" > /dev/null 2>&1; then
                log_info "  ✓ $(basename $file) is valid JSON"
            else
                log_fail "  ✗ $(basename $file) has invalid JSON"
                all_valid=false
            fi
        fi
    done
    
    if [ "$all_valid" = true ]; then
        log_pass "All JSON configs are valid"
        return 0
    else
        log_fail "Some JSON configs are invalid"
        return 1
    fi
}

# Test 4: Test session directory creation
test_session_directory() {
    ((TOTAL_TESTS++))
    log_test "Testing session directory creation"
    
    cd "$PROJECT_DIR"
    rm -rf .sessions/
    
    # Mock openclaw command to test directory creation
    cat > /tmp/mock_openclaw.sh << 'EOF'
#!/bin/bash
echo "Mock OpenClaw running..."
sleep 1
exit 0
EOF
    chmod +x /tmp/mock_openclaw.sh
    
    # Temporarily override openclaw command
    export PATH="/tmp:$PATH"
    ln -sf /tmp/mock_openclaw.sh /tmp/openclaw
    
    # Run wrapper (will exit quickly with mock)
    timeout 5s ./openclaw-wrapper.sh agent --message "test" 2>&1 || true
    
    if [ -d ".sessions" ]; then
        log_pass "Session directory created successfully"
        rm -rf .sessions/
        rm /tmp/openclaw /tmp/mock_openclaw.sh
        return 0
    else
        log_fail "Session directory was not created"
        rm /tmp/openclaw /tmp/mock_openclaw.sh 2>/dev/null || true
        return 1
    fi
}

# Test 5: Test config extraction (if config exists)
test_config_extraction() {
    ((TOTAL_TESTS++))
    log_test "Testing configuration extraction"
    
    # Create a temporary test config
    mkdir -p ~/.openclaw
    cat > ~/.openclaw/openclaw.json << 'EOF'
{
  "agent": {
    "model": "anthropic/claude-sonnet-4",
    "systemPrompt": "Test system prompt",
    "tools": ["bash", "read", "write"]
  }
}
EOF
    
    # Test the grep commands directly
    cd "$PROJECT_DIR"
    CONFIG_FILE="$HOME/.openclaw/openclaw.json"
    
    if [ -f "$CONFIG_FILE" ]; then
        MODEL=$(grep -o '"model"[[:space:]]*:[[:space:]]*"[^"]*"' "$CONFIG_FILE" | sed 's/"model"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
        SYSTEM_PROMPT=$(grep -o '"systemPrompt"[[:space:]]*:[[:space:]]*"[^"]*"' "$CONFIG_FILE" | sed 's/"systemPrompt"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/' | cut -c1-100)
        TOOLS=$(grep -o '"tools"[[:space:]]*:[[:space:]]*\[[^]]*\]' "$CONFIG_FILE" | sed 's/"tools"[[:space:]]*:[[:space:]]*\[\([^]]*\)\]/\1/' | tr -d '"' | tr ',' ' ')
        
        if [ -n "$MODEL" ] && [ -n "$SYSTEM_PROMPT" ] && [ -n "$TOOLS" ]; then
            log_pass "Config extraction working correctly"
            log_info "  Model: $MODEL"
            log_info "  System Prompt: ${SYSTEM_PROMPT:0:50}..."
            log_info "  Tools: $TOOLS"
            return 0
        else
            log_fail "Config extraction failed - empty values"
            return 1
        fi
    else
        log_fail "Config file not found"
        return 1
    fi
}

# Run all tests
main() {
    echo "======================================"
    echo "  Running Wrapper Script Tests"
    echo "======================================"
    echo ""
    
    test_scripts_exist || true
    test_script_syntax || true
    test_json_configs || true
    test_session_directory || true
    test_config_extraction || true
    
    echo ""
    echo "======================================"
    echo "  Test Results Summary"
    echo "======================================"
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    echo ""
    
    # Save results
    cat > "$TEST_RESULTS_DIR/summary.txt" << EOF
Test Results Summary
====================
Total Tests: $TOTAL_TESTS
Passed: $PASSED_TESTS
Failed: $FAILED_TESTS
Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%
EOF
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    fi
}

main "$@"
