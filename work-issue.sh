#!/bin/bash

# Function to extract repository name
get_repo_name() {
  # First try to get name from git remote
  local remote_url=$(git remote get-url origin 2>/dev/null)
  if [ -n "$remote_url" ]; then
    # Extract repository name from URL
    # Handle SSH: git@github.com:user/repo.git
    # Handle HTTPS: https://github.com/user/repo.git
    local repo_name=$(echo "$remote_url" | sed -E 's|.*[:/]([^/]+)/([^/]+)$|\2|' | sed 's|\.git$||')
    if [ -n "$repo_name" ]; then
      echo "$repo_name"
      return
    fi
  fi
  
  # Fallback to directory name
  basename "$PWD"
}

# Function to update Claude project trust settings
update_claude_project_trust() {
  local project_path="$PWD"
  local claude_config="$HOME/.claude.json"
  
  # Check if jq is available
  if ! command -v jq >/dev/null 2>&1; then
    echo "Warning: jq not found, skipping Claude trust configuration"
    return 1
  fi
  
  # Check if config file exists
  if [ ! -f "$claude_config" ]; then
    echo "Warning: Claude config file not found at $claude_config"
    return 1
  fi
  
  # Create a temporary file for atomic update
  local temp_file=$(mktemp)
  
  # Update or create the project entry with hasTrustDialogAccepted: true
  jq --arg path "$project_path" '
    .projects[$path] = ((.projects[$path] // {}) + {"allowedTools": [], "hasTrustDialogAccepted": true, "hasCompletedProjectOnboarding": true})
  ' "$claude_config" > "$temp_file"
  
  if [ $? -eq 0 ]; then
    mv "$temp_file" "$claude_config"
    echo "Updated Claude project trust for: $project_path"
  else
    echo "Warning: Failed to update Claude config"
    rm -f "$temp_file"
    return 1
  fi
}

# Create sandbox name with repository name and optional title
REPO_NAME=$(get_repo_name)
SANDBOX_NAME="work-issue-$REPO_NAME-$1"
if [ -n "$SBS_TITLE" ]; then
  # Sanitize title for sandbox name (replace spaces and special chars with dashes)
  SANITIZED_TITLE=$(echo "$SBS_TITLE" | sed 's/[^a-zA-Z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$//g')
  SANDBOX_NAME="work-issue-$REPO_NAME-$1-$SANITIZED_TITLE"
fi

# Function to install Claude Code hook in sandbox
install_claude_hook() {
  local hook_script="/work/scripts/claude-code-stop-hook.sh"
  local claude_config="$HOME/.claude/config.json"
  local sandbox_hook="$HOME/claude-code-stop-hook.sh"
  
  echo "Installing Claude Code hook in sandbox..."
  
  # Check if hook script exists in the project
  if [ ! -f "$hook_script" ]; then
    echo "Warning: Claude Code hook script not found at $hook_script, skipping hook installation"
    return 1
  fi
  
  # Copy hook script to sandbox home directory
  if cp "$hook_script" "$sandbox_hook"; then
    chmod +x "$sandbox_hook"
    echo "Claude Code hook script copied to sandbox: $sandbox_hook"
  else
    echo "Warning: Failed to copy Claude Code hook script, skipping hook installation"
    return 1
  fi
  
  # Check if jq is available for JSON manipulation
  if ! command -v jq >/dev/null 2>&1; then
    echo "Warning: jq not found, skipping Claude Code hook configuration"
    return 1
  fi
  
  # Check if Claude config file exists
  if [ ! -f "$claude_config" ]; then
    echo "Warning: Claude config file not found at $claude_config, creating basic config"
    echo '{"projects": {}, "hooks": {}}' > "$claude_config"
  fi
  
  # Create a temporary file for atomic update
  local temp_file=$(mktemp)
  
  # Configure Claude Code hook - add PostToolUse hook
  jq --arg hook_path "$sandbox_hook" '
    .hooks.PostToolUse = [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": $hook_path
          }
        ]
      }
    ]
  ' "$claude_config" > "$temp_file"
  
  if [ $? -eq 0 ]; then
    mv "$temp_file" "$claude_config"
    echo "Claude Code hook configured in sandbox: PostToolUse -> $sandbox_hook"
  else
    echo "Warning: Failed to configure Claude Code hook"
    rm -f "$temp_file"
    return 1
  fi
  
  # Create .sbs directory in working directory for hook output
  if mkdir -p /work/.sbs; then
    echo "Created .sbs directory for hook output: /work/.sbs"
  else
    echo "Warning: Failed to create .sbs directory"
    return 1
  fi
  
  echo "Claude Code hook installation completed successfully"
  return 0
}

# Update Claude project trust settings for current directory
update_claude_project_trust

# Install Claude Code hook in sandbox environment
install_claude_hook

sandbox \
  --net="host" \
  --bind /tmp/tmux-1000 \
  --name "$SANDBOX_NAME" \
  ~/.claude/local/claude \
  --model sonnet \
  --dangerously-skip-permissions \
  "/work-issue $1"
