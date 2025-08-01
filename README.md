# SBS (Sandbox Sessions)

A command-line tool that orchestrates GitHub issue work environments with automatic git worktree and tmux session management.

## Features

- **Automated Workflow**: Creates git branches, worktrees, and tmux sessions automatically
- **Issue Integration**: Fetches GitHub issue metadata for better organization
- **Beautiful TUI**: Interactive terminal interface built with Charm.sh tools
- **Session Management**: Track, list, and attach to multiple concurrent work sessions
- **Cleanup Tools**: Automatic cleanup of stale sessions and worktrees

## Installation

```bash
go build -o sbs
```

## Usage

### Start a new work session
```bash
./sbs start 123
```

This will:
1. Create/switch to branch `issue-123-<issue-title-slug>`
2. Create worktree in `~/.work-issue-worktrees/issue-123/`
3. Launch tmux session `work-issue-123`
4. Execute `work-issue.sh 123` in the session

### List active sessions
```bash
./sbs list          # Interactive TUI
./sbs list --plain  # Plain text output
```

### Attach to an existing session
```bash
./sbs attach 123
```

### Stop a session
```bash
./sbs stop 123
```

### Clean up stale sessions
```bash
./sbs clean
./sbs clean --dry-run  # Preview changes
```

## Configuration

Configuration is stored in `~/.config/sbs/config.json`:

```json
{
  "worktree_base_path": "/home/user/.work-issue-worktrees",
  "github_token": "ghp_...",
  "work_issue_script": "/home/user/code/work-issue/work-issue.sh",
  "repo_path": "."
}
```

## Requirements

- Git with worktree support
- tmux
- GitHub CLI (`gh`) for issue metadata
- `work-issue.sh` script (from the original work-issue project)

## Architecture

The application is built with:
- **Cobra** for CLI structure
- **Bubble Tea** for interactive TUI
- **Lipgloss** for terminal styling
- **go-git** for Git operations
- **go-github** for GitHub API integration

Each work session is tracked with metadata including issue number, title, branch name, worktree path, and tmux session name.

## Claude Code Integration

This project includes automatic Claude Code hook installation that captures tool usage data within sandbox environments for analysis and debugging.

### Automatic Hook Installation

The Claude Code hook is automatically installed in each sandbox environment when you start a work session:

1. **During `sbs start`**: The `work-issue.sh` script automatically installs the hook within the sandbox
2. **Sandbox-Only**: The hook runs only inside the sandbox environment, not on the host
3. **No Manual Setup**: No manual configuration required - everything is automatic

### How It Works

When you start a work session:

1. **SBS creates the sandbox** using the `work-issue.sh` script
2. **Hook script is copied** from `scripts/claude-code-stop-hook.sh` to the sandbox
3. **Claude Config is updated** within the sandbox to include the PostToolUse hook
4. **Hook captures data** and writes to `.sbs/stop.json` inside the sandbox
5. **Data is accessible** from the host via sandbox commands

### Accessing Hook Data

To check what Claude Code was doing in any sandbox:

```bash
# Check the most recent Claude Code activity
sandbox --name work-issue-repo-123 cat .sbs/stop.json

# Pretty-print the JSON (if jq is available on host)
sandbox --name work-issue-repo-123 cat .sbs/stop.json | jq .

# Check if hook data exists
sandbox --name work-issue-repo-123 ls -la .sbs/
```

### Output Format

The hook creates `.sbs/stop.json` with the following structure:

```json
{
  "claude_code_hook": {
    "timestamp": "2025-08-01T12:30:45Z",
    "environment": "sandbox",
    "project_directory": "/work",
    "hook_script": "/home/user/claude-code-stop-hook.sh",
    "sandbox_detection": true
  },
  "hook_data": {
    // Original Claude Code hook payload from Claude Code
    // Contains session info, tool data, execution context, etc.
  }
}
```

### Hook Details

The automatic installation:

- **Triggers**: PostToolUse events (after Claude Code tool execution)
- **Captures**: All tool usage within the sandbox environment
- **Isolates**: Hook configuration only affects the sandbox, not your host Claude Code setup
- **Persists**: Hook data remains accessible until the sandbox is deleted
- **Updates**: Each tool execution overwrites the previous `.sbs/stop.json` file

### Troubleshooting

- **Hook Not Installing**: Check that `scripts/claude-code-stop-hook.sh` exists and is executable
- **No Hook Data**: Ensure Claude Code is running within the sandbox environment
- **Permission Issues**: The `work-issue.sh` script handles permissions automatically
- **Missing jq**: Hook will still work but JSON won't be pretty-printed
- **Sandbox Access**: Use `sandbox --name [sandbox-name] ls .sbs/` to verify hook directory exists

### Example Workflow

```bash
# Start a work session
sbs start 123

# Work with Claude Code in the sandbox (hook auto-installed)
# ... Claude Code tool executions happen ...

# From host, check what Claude Code was doing
sandbox --name work-issue-repo-123 cat .sbs/stop.json

# Pretty print the last activity
sandbox --name work-issue-repo-123 cat .sbs/stop.json | jq .hook_data
```