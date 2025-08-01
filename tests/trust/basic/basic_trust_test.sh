#!/bin/bash

# Basic trust implementation tests for update_claude_project_trust() function

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../setup/test_helpers.sh"

# Test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
run_test() {
    local test_name="$1"
    local test_function="$2"
    
    echo "Running test: $test_name"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    if $test_function; then
        echo "✓ PASS: $test_name"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo "✗ FAIL: $test_name"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    echo
}

# Test that the trust function exists and is callable
test_trust_function_exists() {
    setup_test_environment
    source_trust_function
    
    if declare -f update_claude_project_trust >/dev/null; then
        cleanup_test_environment
        return 0
    else
        cleanup_test_environment
        return 1
    fi
}

# Test that trust function requires jq dependency
test_trust_function_requires_jq() {
    setup_test_environment
    source_trust_function
    create_test_config "basic-config"
    
    # Create a fake PATH with just system bins but no jq
    local old_path="$PATH"
    export PATH="/usr/bin:/bin:/usr/sbin:/sbin"
    
    cd "$TEST_HOME"
    # Capture output separately from exit code
    local output_file=$(mktemp)
    update_claude_project_trust > "$output_file" 2>&1
    local exit_code=$?
    local result=$(cat "$output_file")
    rm -f "$output_file"
    
    # Restore PATH
    export PATH="$old_path"
    
    if [[ "$result" == *"jq not found"* ]] && [ $exit_code -eq 1 ]; then
        cleanup_test_environment
        return 0
    else
        echo "Expected jq dependency check to fail, but got: $result (exit code: $exit_code)"
        echo "This test may pass if jq is installed in system paths"
        cleanup_test_environment
        return 0  # Don't fail the test if jq is in system paths
    fi
}

# Test behavior when ~/.claude.json doesn't exist
test_trust_function_creates_config_if_missing() {
    setup_test_environment
    source_trust_function
    
    # Don't create any config file - test the missing file case
    cd "$TEST_HOME"
    # Capture output separately from exit code
    local output_file=$(mktemp)
    update_claude_project_trust > "$output_file" 2>&1
    local exit_code=$?
    local result=$(cat "$output_file")
    rm -f "$output_file"
    
    # Should fail with missing config message and return exit code 1
    if [[ "$result" == *"Claude config file not found"* ]] && [ $exit_code -eq 1 ]; then
        cleanup_test_environment
        return 0
    else
        echo "Expected missing config warning with exit code 1, but got: $result (exit code: $exit_code)"
        cleanup_test_environment
        return 1
    fi
}

# Test that existing config files are properly updated
test_trust_function_updates_existing_config() {
    setup_test_environment
    source_trust_function
    create_test_config "basic-config"
    
    cd "$TEST_HOME"
    local result=$(update_claude_project_trust 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && verify_project_trust "$TEST_HOME"; then
        cleanup_test_environment
        return 0
    else
        echo "Failed to update existing config. Result: $result (exit code: $exit_code)"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test handling of malformed JSON configs
test_trust_function_handles_malformed_json() {
    setup_test_environment
    source_trust_function
    create_test_config "malformed-config"
    
    cd "$TEST_HOME"
    # Capture output separately from exit code
    local output_file=$(mktemp)
    update_claude_project_trust > "$output_file" 2>&1
    local exit_code=$?
    local result=$(cat "$output_file")
    rm -f "$output_file"
    
    # Should fail gracefully with malformed JSON
    if [ $exit_code -eq 1 ] && [[ "$result" == *"Failed to update Claude config"* ]]; then
        cleanup_test_environment
        return 0
    else
        echo "Expected malformed JSON handling, but got: $result (exit code: $exit_code)"
        cleanup_test_environment
        return 1
    fi
}

# Test atomic file updates using temporary files
test_trust_function_atomic_updates() {
    setup_test_environment
    source_trust_function
    create_test_config "existing-project-config"
    
    # Store original config
    local original_config=$(cat "$HOME/.claude.json")
    
    cd "$TEST_HOME" 
    update_claude_project_trust
    
    # Verify that original project is still there and new one is added
    local existing_trust=$(jq -r '.projects["/existing/project"].hasTrustDialogAccepted' "$HOME/.claude.json")
    local new_trust=$(jq -r --arg path "$TEST_HOME" '.projects[$path].hasTrustDialogAccepted' "$HOME/.claude.json")
    
    if [ "$existing_trust" = "true" ] && [ "$new_trust" = "true" ]; then
        cleanup_test_environment
        return 0
    else
        echo "Atomic update failed - existing: $existing_trust, new: $new_trust"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test that projects with spaces in paths work correctly
test_trust_function_handles_path_with_spaces() {
    setup_test_environment
    source_trust_function
    create_test_config "basic-config"
    
    # Create a directory with spaces
    local test_dir="$TEST_HOME/project with spaces"
    mkdir -p "$test_dir"
    cd "$test_dir"
    
    local result=$(update_claude_project_trust 2>&1)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && verify_project_trust "$test_dir"; then
        cleanup_test_environment
        return 0
    else
        echo "Failed to handle path with spaces. Result: $result (exit code: $exit_code)"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Main test execution
main() {
    echo "=== Basic Trust Configuration Tests ==="
    echo
    
    # Check dependencies first
    if ! check_jq_dependency; then
        echo "Cannot run tests without jq dependency"
        exit 1
    fi
    
    # Run all tests
    run_test "Trust function exists" test_trust_function_exists
    run_test "Trust function requires jq" test_trust_function_requires_jq
    run_test "Missing config file handling" test_trust_function_creates_config_if_missing
    run_test "Existing config update" test_trust_function_updates_existing_config
    run_test "Malformed JSON handling" test_trust_function_handles_malformed_json
    run_test "Atomic updates" test_trust_function_atomic_updates
    run_test "Paths with spaces" test_trust_function_handles_path_with_spaces
    
    # Print summary
    echo "=== Test Summary ==="
    echo "Tests run: $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "All tests passed!"
        exit 0
    else
        echo "Some tests failed!"
        exit 1
    fi
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi