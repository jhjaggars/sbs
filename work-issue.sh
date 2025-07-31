#!/bin/bash

# set_title() {
#     if [ -n "$TMUX" ]; then
#         # We're in tmux - use tmux command
#         tmux rename-window "$1"
#     else
#         # Regular terminal - use ANSI escape sequence
#         printf '\033]0;%s\007' "$1"
#     fi
# }
#
# TITLE=(gh issue view $1 --json title -q .title)
#
# set_title "✻ $TITLE"

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
    .projects[$path] = ((.projects[$path] // {}) + {"hasTrustDialogAccepted": true})
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

# Update Claude project trust settings for current directory
update_claude_project_trust

sandbox \
  --net="host" \
  --bind /tmp/tmux-1000 \
  --name "$SANDBOX_NAME" \
  ~/.claude/local/claude \
  --model sonnet \
  --dangerously-skip-permissions \
  "/work-issue $1"
