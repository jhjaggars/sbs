#!/bin/bash

# Comprehensive test runner for Claude Code trust configuration

echo "=== Claude Code Trust Configuration Test Suite ==="
echo "Running comprehensive tests for automatic worktree trust configuration"
echo

# Test directories
TRUST_TESTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/trust"

# Test results
TOTAL_TESTS=0
TOTAL_PASSED=0  
TOTAL_FAILED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to run a test suite
run_test_suite() {
    local test_script="$1"
    local suite_name="$2"
    
    echo -e "${YELLOW}Running $suite_name...${NC}"
    echo "=============================================="
    
    if [ -f "$test_script" ] && [ -x "$test_script" ]; then
        # Run the test and capture results
        local output
        local exit_code
        
        output=$("$test_script" 2>&1)
        exit_code=$?
        
        echo "$output"
        
        # Extract test counts from output
        local tests_run=$(echo "$output" | grep "Tests run:" | sed 's/Tests run: //')
        local tests_passed=$(echo "$output" | grep "Tests passed:" | sed 's/Tests passed: //')
        local tests_failed=$(echo "$output" | grep "Tests failed:" | sed 's/Tests failed: //')
        
        # Update totals
        if [ -n "$tests_run" ]; then
            TOTAL_TESTS=$((TOTAL_TESTS + tests_run))
            TOTAL_PASSED=$((TOTAL_PASSED + tests_passed))
            TOTAL_FAILED=$((TOTAL_FAILED + tests_failed))
        fi
        
        if [ $exit_code -eq 0 ]; then
            echo -e "${GREEN}✓ $suite_name: PASSED${NC}"
        else
            echo -e "${RED}✗ $suite_name: FAILED${NC}"
        fi
    else
        echo -e "${RED}✗ $suite_name: Test script not found or not executable: $test_script${NC}"
        TOTAL_FAILED=$((TOTAL_FAILED + 1))
    fi
    
    echo
}

# Main execution
main() {
    echo "Starting comprehensive trust configuration tests..."
    echo "Test suite location: $TRUST_TESTS_DIR"
    echo
    
    # Check if jq is available
    if ! command -v jq >/dev/null 2>&1; then
        echo -e "${RED}ERROR: jq is required for tests but not found${NC}"
        echo "Please install jq to run the test suite"
        exit 1
    fi
    
    # Run test suites in order
    run_test_suite "$TRUST_TESTS_DIR/basic/basic_trust_test.sh" "Basic Trust Function Tests"
    run_test_suite "$TRUST_TESTS_DIR/e2e/session_trust_e2e_test.sh" "End-to-End Trust Workflow Tests"
    run_test_suite "$TRUST_TESTS_DIR/e2e/multiple_worktree_scenarios_test.sh" "Multiple Worktree Scenario Tests"
    
    # Print final summary
    echo "=============================================="
    echo -e "${YELLOW}FINAL TEST RESULTS${NC}"
    echo "=============================================="
    echo "Total tests run: $TOTAL_TESTS"
    echo -e "Total passed: ${GREEN}$TOTAL_PASSED${NC}"
    echo -e "Total failed: ${RED}$TOTAL_FAILED${NC}"
    
    if [ $TOTAL_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
        echo
        echo "The Claude Code automatic trust configuration is working correctly."
        echo "Worktree directories will be automatically trusted without manual intervention."
        exit 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED!${NC}"
        echo
        echo "There are issues with the trust configuration that need attention."
        echo "See individual test outputs above for details."
        exit 1
    fi
}

# Show help if requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: $0 [options]"
    echo
    echo "Runs the complete Claude Code trust configuration test suite."
    echo
    echo "Options:"
    echo "  --help, -h    Show this help message"
    echo
    echo "Test Suites:"
    echo "  1. Basic Trust Function Tests       - Unit tests for trust function"
    echo "  2. End-to-End Trust Workflow Tests  - Complete workflow simulation"
    echo "  3. Multiple Worktree Scenarios      - Edge cases and concurrent access"
    echo
    echo "Requirements:"
    echo "  - jq (JSON processor)"
    echo "  - bash (version 4.0+)"
    echo
    exit 0
fi

# Run main execution
main "$@"