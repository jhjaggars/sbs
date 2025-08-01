#!/bin/bash

# Research script to discover actual Claude Code trust configuration mechanism

echo "=== Claude Code Configuration Research ==="
echo

# Function to check if Claude Code is installed and where
discover_claude_installations() {
    echo "1. Discovering Claude Code installations..."
    
    # Check common installation locations
    local locations=(
        "$(which claude 2>/dev/null)"
        "$(which claude-code 2>/dev/null)"
        "$HOME/.local/bin/claude"
        "$HOME/.local/bin/claude-code"
        "/usr/local/bin/claude"
        "/usr/local/bin/claude-code"
        "$(npm list -g --depth=0 2>/dev/null | grep -o '[^ ]*claude[^ ]*' | head -1)"
        "$(yarn global list 2>/dev/null | grep -o '[^ ]*claude[^ ]*' | head -1)"
    )
    
    echo "Checking potential Claude Code locations:"
    for location in "${locations[@]}"; do
        if [ -n "$location" ] && [ -f "$location" ]; then
            echo "  ✓ Found: $location"
            
            # Try to get version information
            if "$location" --version 2>/dev/null; then
                echo "    Version: $("$location" --version 2>/dev/null || echo "Unable to get version")"
            fi
        elif [ -n "$location" ]; then
            echo "  ✗ Not found: $location"
        fi
    done
    echo
}

# Function to discover Claude Code configuration file locations
discover_config_locations() {
    echo "2. Discovering Claude Code configuration locations..."
    
    local config_locations=(
        "$HOME/.claude.json"
        "$HOME/.claude/config.json"
        "$HOME/.config/claude/config.json"
        "$HOME/.config/claude.json"
        "$HOME/Library/Application Support/Claude/config.json"
        "$HOME/Library/Preferences/Claude/config.json"
        "$HOME/.local/share/claude/config.json"
    )
    
    echo "Checking potential config file locations:"
    for config in "${config_locations[@]}"; do
        if [ -f "$config" ]; then
            echo "  ✓ Found config: $config"
            echo "    Size: $(wc -c < "$config") bytes"
            echo "    Modified: $(stat -c %y "$config" 2>/dev/null || stat -f %Sm "$config" 2>/dev/null)"
            
            # Try to parse and show structure (first few lines only)
            echo "    Structure preview:"
            if command -v jq >/dev/null 2>&1; then
                jq -C 'keys' "$config" 2>/dev/null | head -10 | sed 's/^/      /'
            else
                head -5 "$config" | sed 's/^/      /'
            fi
        else
            echo "  ✗ Not found: $config"
        fi
    done
    echo
}

# Function to test trust dialog behavior
test_trust_dialog_behavior() {
    echo "3. Testing trust dialog behavior..."
    
    # Create a temporary test directory
    local test_dir=$(mktemp -d)
    echo "Created test directory: $test_dir"
    
    # Create a simple test file
    echo "console.log('Hello from test');" > "$test_dir/test.js"
    
    # Try to determine if Claude Code shows trust dialog
    echo "Manual test required: Navigate to $test_dir and run Claude Code"
    echo "Observe if trust dialog appears and what configuration is created/modified"
    echo "Test directory will remain for manual testing: $test_dir"
    echo
}

# Function to analyze current trust implementation
analyze_current_implementation() {
    echo "4. Analyzing current trust implementation..."
    
    local work_issue_script="/home/jhjaggars/.work-issue-worktrees/work-issue/issue-14/work-issue.sh"
    
    if [ -f "$work_issue_script" ]; then
        echo "Current implementation analysis:"
        echo "  Trust function location: $work_issue_script (lines 23-55)"
        echo "  Target config file: ~/.claude.json"
        echo "  Trust approach: Set hasTrustDialogAccepted and hasCompletedProjectOnboarding to true"
        
        # Show the current implementation
        echo "  Current implementation code:"
        sed -n '23,55p' "$work_issue_script" | sed 's/^/    /'
    else
        echo "ERROR: work-issue.sh script not found at expected location"
    fi
    echo
}

# Function to test the current configuration format
test_config_format() {
    echo "5. Testing current configuration format..."
    
    # Create a test config with the current implementation's format
    local test_config=$(mktemp)
    cat > "$test_config" << 'EOF'
{
  "projects": {
    "/test/project/path": {
      "allowedTools": [],
      "hasTrustDialogAccepted": true,
      "hasCompletedProjectOnboarding": true
    }
  }
}
EOF
    
    echo "Test config created at: $test_config"
    echo "Content:"
    cat "$test_config" | sed 's/^/  /'
    
    if command -v jq >/dev/null 2>&1; then
        echo "JSON validation: $(jq empty "$test_config" 2>&1 && echo "Valid" || echo "Invalid")"
    fi
    
    echo "Manual verification needed: Use this config format with Claude Code to test if trust prompts are bypassed"
    echo
}

# Function to search for Claude Code documentation or examples
search_documentation() {
    echo "6. Searching for Claude Code trust documentation..."
    
    # Check if there are any existing Claude configs on the system
    echo "Looking for existing Claude configurations for reference:"
    find "$HOME" -name "*.json" -path "*claude*" 2>/dev/null | head -10 | while read -r file; do
        if [ -f "$file" ]; then
            echo "  Found: $file"
            if command -v jq >/dev/null 2>&1 && jq empty "$file" 2>/dev/null; then
                echo "    Projects configured: $(jq -r '.projects | keys | length' "$file" 2>/dev/null || echo "N/A")"
                echo "    Has trust settings: $(jq -r '.projects | to_entries[] | select(.value.hasTrustDialogAccepted != null) | .key' "$file" 2>/dev/null | wc -l)"
            fi
        fi
    done
    echo
}

# Main research execution
main() {
    echo "Starting Claude Code configuration research..."
    echo "This will help understand the correct trust configuration mechanism."
    echo
    
    discover_claude_installations
    discover_config_locations
    analyze_current_implementation
    test_config_format
    search_documentation
    test_trust_dialog_behavior
    
    echo "=== Research Complete ==="
    echo "Next steps:"
    echo "1. Manually test Claude Code in the created test directory"
    echo "2. Observe trust dialog behavior and configuration changes"
    echo "3. Compare with current implementation approach"
    echo "4. Update implementation based on findings"
}

# Run research if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi