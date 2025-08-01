#!/bin/bash

# End-to-end tests for complete trust workflow

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

# Test that new session creates trust configuration
test_new_session_creates_trust_config() {
    setup_test_environment
    create_test_config "basic-config"
    
    # Simulate worktree directory creation
    local worktree_dir="$TEST_HOME/work-issue-test-123"
    mkdir -p "$worktree_dir"
    cd "$worktree_dir"
    
    # Source and run the trust function as it would be called from work-issue.sh
    source_trust_function
    local output_file=$(mktemp)
    update_claude_project_trust > "$output_file" 2>&1
    local exit_code=$?
    local result=$(cat "$output_file")
    rm -f "$output_file"
    
    # Verify trust was configured
    if [ $exit_code -eq 0 ] && verify_project_trust "$worktree_dir"; then
        cleanup_test_environment
        return 0
    else
        echo "Failed to create trust config for new session. Result: $result (exit code: $exit_code)"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test multiple sessions create unique trust entries
test_multiple_sessions_unique_trust_entries() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Create multiple worktree directories
    local worktree1="$TEST_HOME/work-issue-test-123"
    local worktree2="$TEST_HOME/work-issue-test-456"
    mkdir -p "$worktree1" "$worktree2"
    
    # Configure trust for first worktree
    cd "$worktree1"
    update_claude_project_trust >/dev/null 2>&1
    
    # Configure trust for second worktree
    cd "$worktree2"
    update_claude_project_trust >/dev/null 2>&1
    
    # Verify both are configured
    if verify_project_trust "$worktree1" && verify_project_trust "$worktree2"; then
        # Verify they are separate entries
        local config_path="$HOME/.claude.json"
        local count=$(jq -r --arg path1 "$worktree1" --arg path2 "$worktree2" '
            [.projects | keys[] | select(. == $path1 or . == $path2)] | length
        ' "$config_path")
        
        if [ "$count" = "2" ]; then
            cleanup_test_environment
            return 0
        else
            echo "Expected 2 separate trust entries, but found $count"
            print_config
            cleanup_test_environment
            return 1
        fi
    else
        echo "Failed to configure trust for multiple sessions"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test that trust persists across directory changes
test_trust_persists_across_session_restarts() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    local worktree_dir="$TEST_HOME/work-issue-test-789"
    mkdir -p "$worktree_dir"
    cd "$worktree_dir"
    
    # Configure trust
    update_claude_project_trust >/dev/null 2>&1
    
    # Verify initial trust
    if ! verify_project_trust "$worktree_dir"; then
        echo "Initial trust configuration failed"
        cleanup_test_environment
        return 1
    fi
    
    # Simulate session restart by changing directories and coming back
    cd "$TEST_HOME"
    cd "$worktree_dir"
    
    # Verify trust still exists
    if verify_project_trust "$worktree_dir"; then
        cleanup_test_environment
        return 0
    else
        echo "Trust configuration did not persist across session restart"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test trust config survives config file updates
test_trust_config_survives_updates() {
    setup_test_environment
    create_test_config "existing-project-config"
    source_trust_function
    
    local worktree_dir="$TEST_HOME/work-issue-test-update"
    mkdir -p "$worktree_dir"
    cd "$worktree_dir"
    
    # Configure trust for new worktree
    update_claude_project_trust >/dev/null 2>&1
    
    # Verify both original and new project exist
    local config_path="$HOME/.claude.json"
    local original_exists=$(jq -r '.projects["/existing/project"].hasTrustDialogAccepted' "$config_path")
    local new_exists=$(jq -r --arg path "$worktree_dir" '.projects[$path].hasTrustDialogAccepted' "$config_path")
    
    if [ "$original_exists" = "true" ] && [ "$new_exists" = "true" ] && verify_project_trust "$worktree_dir"; then
        cleanup_test_environment
        return 0
    else
        echo "Trust config update affected existing entries. Original: $original_exists, New: $new_exists"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test the complete workflow as it happens in practice
test_complete_workflow_simulation() {
    setup_test_environment
    create_test_config "basic-config"
    
    # Simulate the complete SBS workflow
    local repo_name="test-repo"
    local issue_number="123"
    local worktree_dir="$TEST_HOME/.work-issue-worktrees/$repo_name/issue-$issue_number"
    
    # Create worktree directory structure as SBS would
    mkdir -p "$worktree_dir"
    cd "$worktree_dir"
    
    # Create some project files to make it realistic
    echo "# Test Project" > README.md
    echo '{"name": "test-project"}' > package.json
    
    # Run the trust configuration as work-issue.sh would
    source_trust_function
    local output_file=$(mktemp)
    update_claude_project_trust > "$output_file" 2>&1
    local exit_code=$?
    local result=$(cat "$output_file")
    rm -f "$output_file"
    
    # Verify the complete setup
    if [ $exit_code -eq 0 ] && verify_project_trust "$worktree_dir"; then
        # Check that the config contains expected structure
        local config_path="$HOME/.claude.json"
        local has_allowed_tools=$(jq -r --arg path "$worktree_dir" '.projects[$path].allowedTools | type' "$config_path")
        local has_trust=$(jq -r --arg path "$worktree_dir" '.projects[$path].hasTrustDialogAccepted' "$config_path")
        local has_onboarding=$(jq -r --arg path "$worktree_dir" '.projects[$path].hasCompletedProjectOnboarding' "$config_path")
        
        if [ "$has_allowed_tools" = "array" ] && [ "$has_trust" = "true" ] && [ "$has_onboarding" = "true" ]; then
            echo "Complete workflow simulation successful"
            cleanup_test_environment
            return 0
        else
            echo "Trust config structure incorrect: allowedTools=$has_allowed_tools, trust=$has_trust, onboarding=$has_onboarding"
            print_config
            cleanup_test_environment
            return 1
        fi
    else
        echo "Complete workflow failed. Result: $result (exit code: $exit_code)"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test with realistic worktree paths
test_realistic_worktree_paths() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Test paths that match real SBS usage
    local base_path="$TEST_HOME/.work-issue-worktrees"
    local paths=(
        "$base_path/my-project/issue-123"
        "$base_path/another-repo/issue-456"
        "$base_path/project-with-dashes/issue-789"
        "$base_path/project_with_underscores/issue-101"
    )
    
    # Configure trust for all paths
    for path in "${paths[@]}"; do
        mkdir -p "$path"
        cd "$path"
        update_claude_project_trust >/dev/null 2>&1
        
        if ! verify_project_trust "$path"; then
            echo "Failed to configure trust for path: $path"
            cleanup_test_environment
            return 1
        fi
    done
    
    # Verify all paths are in config
    local config_path="$HOME/.claude.json"
    local total_projects=$(jq -r '.projects | keys | length' "$config_path")
    
    # Should have original empty projects plus our 4 new ones
    if [ "$total_projects" -ge 4 ]; then
        echo "Successfully configured trust for $total_projects projects"
        cleanup_test_environment
        return 0
    else
        echo "Expected at least 4 projects, but found $total_projects"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Main test execution
main() {
    echo "=== End-to-End Trust Configuration Tests ==="
    echo
    
    # Check dependencies first
    if ! check_jq_dependency; then
        echo "Cannot run tests without jq dependency"
        exit 1
    fi
    
    # Run all tests
    run_test "New session creates trust config" test_new_session_creates_trust_config
    run_test "Multiple sessions create unique entries" test_multiple_sessions_unique_trust_entries
    run_test "Trust persists across restarts" test_trust_persists_across_session_restarts
    run_test "Trust survives config updates" test_trust_config_survives_updates
    run_test "Complete workflow simulation" test_complete_workflow_simulation
    run_test "Realistic worktree paths" test_realistic_worktree_paths
    
    # Print summary
    echo "=== Test Summary ==="
    echo "Tests run: $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "All end-to-end tests passed!"
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