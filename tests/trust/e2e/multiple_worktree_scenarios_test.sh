#!/bin/bash

# Multiple worktree creation scenario tests

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

# Test concurrent worktree creation (simulated)
test_concurrent_worktree_creation() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Create multiple worktrees rapidly
    local pids=()
    local results=()
    
    for i in {1..5}; do
        local worktree_dir="$TEST_HOME/concurrent-$i"
        mkdir -p "$worktree_dir"
        
        # Run trust configuration in background (simulating concurrent execution)
        (
            cd "$worktree_dir"
            update_claude_project_trust >/dev/null 2>&1
            echo $? > "$TEST_HOME/result-$i"
        ) &
        pids+=($!)
    done
    
    # Wait for all background jobs to complete
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    # Check results
    local success_count=0
    for i in {1..5}; do
        local exit_code=$(cat "$TEST_HOME/result-$i" 2>/dev/null || echo "1")
        local worktree_dir="$TEST_HOME/concurrent-$i"
        
        if [ "$exit_code" = "0" ] && verify_project_trust "$worktree_dir"; then
            success_count=$((success_count + 1))
        fi
    done
    
    if [ $success_count -eq 5 ]; then
        echo "All 5 concurrent trust configurations succeeded"
        cleanup_test_environment
        return 0
    else
        echo "Only $success_count out of 5 concurrent trust configurations succeeded"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test rapid sequential worktree creation
test_rapid_sequential_creation() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    local success_count=0
    local total_count=10
    
    # Create worktrees rapidly in sequence
    for i in $(seq 1 $total_count); do
        local worktree_dir="$TEST_HOME/rapid-$i"
        mkdir -p "$worktree_dir"
        cd "$worktree_dir"
        
        if update_claude_project_trust >/dev/null 2>&1 && verify_project_trust "$worktree_dir"; then
            success_count=$((success_count + 1))
        fi
    done
    
    if [ $success_count -eq $total_count ]; then
        echo "All $total_count rapid sequential trust configurations succeeded"
        cleanup_test_environment
        return 0
    else
        echo "Only $success_count out of $total_count rapid sequential trust configurations succeeded"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test deep nested worktree paths
test_deep_nested_paths() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Test very deep nested paths
    local deep_paths=(
        "$TEST_HOME/a/very/deep/nested/path/for/worktree/issue-1"
        "$TEST_HOME/another/extremely/deep/directory/structure/here/issue-2"
        "$TEST_HOME/maximum/depth/testing/with/many/levels/of/nesting/issue-3"
    )
    
    local success_count=0
    for path in "${deep_paths[@]}"; do
        mkdir -p "$path"
        cd "$path"
        
        if update_claude_project_trust >/dev/null 2>&1 && verify_project_trust "$path"; then
            success_count=$((success_count + 1))
        else
            echo "Failed to configure trust for deep path: $path"
        fi
    done
    
    if [ $success_count -eq ${#deep_paths[@]} ]; then
        echo "All deep nested path trust configurations succeeded"
        cleanup_test_environment
        return 0
    else
        echo "Only $success_count out of ${#deep_paths[@]} deep path trust configurations succeeded"
        cleanup_test_environment
        return 1
    fi
}

# Test paths with special characters
test_paths_with_special_characters() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Test paths with various special characters
    local special_paths=(
        "$TEST_HOME/path with spaces/issue-1"
        "$TEST_HOME/path-with-dashes/issue-2"
        "$TEST_HOME/path_with_underscores/issue-3"
        "$TEST_HOME/path.with.dots/issue-4"
        "$TEST_HOME/path@with@symbols/issue-5"
    )
    
    local success_count=0
    for path in "${special_paths[@]}"; do
        mkdir -p "$path"
        cd "$path"
        
        if update_claude_project_trust >/dev/null 2>&1 && verify_project_trust "$path"; then
            success_count=$((success_count + 1))
        else
            echo "Failed to configure trust for special character path: $path"
        fi
    done
    
    if [ $success_count -eq ${#special_paths[@]} ]; then
        echo "All special character path trust configurations succeeded"
        cleanup_test_environment
        return 0
    else
        echo "Only $success_count out of ${#special_paths[@]} special character path trust configurations succeeded"
        cleanup_test_environment
        return 1
    fi
}

# Test mixed existing and new projects
test_mixed_existing_and_new_projects() {
    setup_test_environment
    create_test_config "complex-config"  # Has existing projects
    source_trust_function
    
    # Add several new worktree projects
    local new_paths=(
        "$TEST_HOME/new-project-1/issue-100"
        "$TEST_HOME/new-project-2/issue-200"
        "$TEST_HOME/new-project-3/issue-300"
    )
    
    local success_count=0
    for path in "${new_paths[@]}"; do
        mkdir -p "$path"
        cd "$path"
        
        if update_claude_project_trust >/dev/null 2>&1 && verify_project_trust "$path"; then
            success_count=$((success_count + 1))
        fi
    done
    
    # Verify existing projects are still there
    local config_path="$HOME/.claude.json"
    local existing1=$(jq -r '.projects["/existing/project1"].hasTrustDialogAccepted' "$config_path")
    local existing2=$(jq -r '.projects["/existing/project2"].hasTrustDialogAccepted' "$config_path")
    
    if [ $success_count -eq ${#new_paths[@]} ] && [ "$existing1" = "true" ] && [ "$existing2" = "false" ]; then
        echo "Successfully mixed new and existing project configurations"
        cleanup_test_environment
        return 0
    else
        echo "Mixed project configuration failed. New: $success_count/${#new_paths[@]}, Existing1: $existing1, Existing2: $existing2"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Test large number of projects
test_large_number_of_projects() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    local total_projects=50
    local success_count=0
    
    echo "Creating $total_projects projects..."
    
    for i in $(seq 1 $total_projects); do
        local project_path="$TEST_HOME/mass-project-$i/issue-$i"
        mkdir -p "$project_path"
        cd "$project_path"
        
        if update_claude_project_trust >/dev/null 2>&1; then
            success_count=$((success_count + 1))
        fi
        
        # Print progress every 10 projects
        if [ $((i % 10)) -eq 0 ]; then
            echo "  Created $i/$total_projects projects..."
        fi
    done
    
    # Verify final config has all projects
    local config_path="$HOME/.claude.json"
    local final_count=$(jq -r '.projects | keys | length' "$config_path")
    
    # Should have at least our created projects (may have more from fixture)
    if [ $success_count -eq $total_projects ] && [ $final_count -ge $total_projects ]; then
        echo "Successfully created and configured $total_projects projects (total in config: $final_count)"
        cleanup_test_environment
        return 0
    else
        echo "Large project test failed. Success: $success_count/$total_projects, Final config count: $final_count"
        cleanup_test_environment
        return 1
    fi
}

# Test worktree creation with existing symlinks
test_worktree_with_symlinks() {
    setup_test_environment
    create_test_config "basic-config"
    source_trust_function
    
    # Create a regular directory and a symlink to it
    local real_dir="$TEST_HOME/real-worktree"
    local link_dir="$TEST_HOME/linked-worktree"
    
    mkdir -p "$real_dir"
    ln -s "$real_dir" "$link_dir"
    
    # Test trust configuration on both
    cd "$real_dir"
    local real_result=$(update_claude_project_trust >/dev/null 2>&1 && echo "success" || echo "failed")
    
    cd "$link_dir"
    local link_result=$(update_claude_project_trust >/dev/null 2>&1 && echo "success" || echo "failed")
    
    # Both should succeed, but they might create different config entries
    if [ "$real_result" = "success" ] && [ "$link_result" = "success" ]; then
        echo "Both real and symlinked directories configured successfully"
        cleanup_test_environment
        return 0
    else
        echo "Symlink test failed. Real: $real_result, Link: $link_result"
        print_config
        cleanup_test_environment
        return 1
    fi
}

# Main test execution
main() {
    echo "=== Multiple Worktree Creation Scenario Tests ==="
    echo
    
    # Check dependencies first
    if ! check_jq_dependency; then
        echo "Cannot run tests without jq dependency"
        exit 1
    fi
    
    # Run all tests
    run_test "Concurrent worktree creation" test_concurrent_worktree_creation
    run_test "Rapid sequential creation" test_rapid_sequential_creation
    run_test "Deep nested paths" test_deep_nested_paths
    run_test "Paths with special characters" test_paths_with_special_characters
    run_test "Mixed existing and new projects" test_mixed_existing_and_new_projects
    run_test "Large number of projects" test_large_number_of_projects
    run_test "Worktree with symlinks" test_worktree_with_symlinks
    
    # Print summary
    echo "=== Test Summary ==="
    echo "Tests run: $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "All multiple worktree scenario tests passed!"
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