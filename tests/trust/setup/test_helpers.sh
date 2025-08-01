#!/bin/bash

# Test helper functions for Claude trust configuration testing

# Set up test environment with isolated home directory
setup_test_environment() {
    export TEST_HOME=$(mktemp -d)
    export ORIGINAL_HOME="$HOME"
    export HOME="$TEST_HOME"
    echo "Test environment set up at: $TEST_HOME"
}

# Clean up test environment
cleanup_test_environment() {
    if [ -n "$TEST_HOME" ] && [ -d "$TEST_HOME" ]; then
        rm -rf "$TEST_HOME"
        echo "Test environment cleaned up: $TEST_HOME"
    fi
    if [ -n "$ORIGINAL_HOME" ]; then
        export HOME="$ORIGINAL_HOME"
    fi
}

# Create a test Claude config file from fixture
create_test_config() {
    local fixture_name="$1"
    local fixture_path="/home/jhjaggars/.work-issue-worktrees/work-issue/issue-14/tests/trust/fixtures/${fixture_name}.json"
    local config_path="$HOME/.claude.json"
    
    if [ ! -f "$fixture_path" ]; then
        echo "ERROR: Fixture not found: $fixture_path"
        return 1
    fi
    
    mkdir -p "$(dirname "$config_path")"
    cp "$fixture_path" "$config_path"
    echo "Created test config from fixture: $fixture_name"
}

# Source the work-issue.sh script to get the trust function
source_trust_function() {
    # Create a safe version of the trust function that can be sourced
    local work_issue_script="/home/jhjaggars/.work-issue-worktrees/work-issue/issue-14/work-issue.sh"
    
    # Extract just the update_claude_project_trust function from lines 23-55
    sed -n '23,55p' "$work_issue_script" > "$TEST_HOME/trust_function.sh"
    source "$TEST_HOME/trust_function.sh"
}

# Verify JSON structure of config file
verify_json_structure() {
    local config_path="$HOME/.claude.json"
    
    if [ ! -f "$config_path" ]; then
        echo "ERROR: Config file not found: $config_path"
        return 1
    fi
    
    if ! jq empty "$config_path" 2>/dev/null; then
        echo "ERROR: Invalid JSON in config file"
        return 1
    fi
    
    echo "JSON structure is valid"
    return 0
}

# Check if project path exists in config with correct trust settings
verify_project_trust() {
    local project_path="$1"
    local config_path="$HOME/.claude.json"
    
    if [ ! -f "$config_path" ]; then
        echo "ERROR: Config file not found: $config_path"
        return 1
    fi
    
    local has_trust_dialog=$(jq -r --arg path "$project_path" '.projects[$path].hasTrustDialogAccepted // false' "$config_path")
    local has_onboarding=$(jq -r --arg path "$project_path" '.projects[$path].hasCompletedProjectOnboarding // false' "$config_path")
    local allowed_tools=$(jq -r --arg path "$project_path" '.projects[$path].allowedTools // []' "$config_path")
    
    if [ "$has_trust_dialog" = "true" ] && [ "$has_onboarding" = "true" ] && [ "$allowed_tools" = "[]" ]; then
        echo "Project trust verified for: $project_path"
        return 0
    else
        echo "ERROR: Project trust not properly configured"
        echo "  hasTrustDialogAccepted: $has_trust_dialog"
        echo "  hasCompletedProjectOnboarding: $has_onboarding"
        echo "  allowedTools: $allowed_tools"
        return 1
    fi
}

# Print the current config for debugging
print_config() {
    local config_path="$HOME/.claude.json"
    
    if [ -f "$config_path" ]; then
        echo "Current config:"
        jq . "$config_path" 2>/dev/null || cat "$config_path"
    else
        echo "No config file found at: $config_path"
    fi
}

# Check if jq is available
check_jq_dependency() {
    if ! command -v jq >/dev/null 2>&1; then
        echo "ERROR: jq is required for tests but not found"
        return 1
    fi
    echo "jq dependency satisfied"
    return 0
}