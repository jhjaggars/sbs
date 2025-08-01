# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

SBS (Sandbox Sessions) is a Go CLI application that orchestrates GitHub issue work environments with automatic git worktree and tmux session management. It creates isolated development environments for each GitHub issue.

## Common Development Commands

### Build and Install
```bash
make build          # Build the sbs binary
make install        # Install to ~/bin
make dev            # Build with race detection
```

### Testing and Code Quality
```bash
make test           # Run all tests
make fmt            # Format Go code
make lint           # Run golangci-lint (requires golangci-lint installed)
go test ./...       # Run tests directly
```

### Git Hooks Setup
```bash
./scripts/install-hooks.sh  # Install pre-commit hook for automatic code formatting
```

The pre-commit hook automatically runs `make fmt` before each commit to ensure consistent code formatting. To temporarily bypass the hook, use `git commit --no-verify`.

### SBS Command Usage

#### Interactive Mode (TUI)
```bash
sbs                    # Launch interactive TUI for session management
go run .               # Run TUI without building
```

#### Start Command
```bash
sbs start 123                           # Start session for issue #123
sbs start                              # Interactive issue selection
sbs start 123 --resume                # Resume existing session without work-issue.sh
sbs start 123 --no-command            # Start without executing any command
sbs start 123 --command "make test"   # Custom command instead of work-issue.sh
sbs start 123 --verbose               # Enable verbose debug output
go run . start 123                     # Run without building
```

#### List and Management
```bash
sbs list              # List sessions in plain text format
sbs list --plain      # Same as above (default behavior)
sbs attach 123        # Attach to existing tmux session
sbs stop 123          # Stop tmux session (preserves worktree)
```

#### Cleanup Operations
```bash
sbs clean             # Clean stale sessions (with confirmation)
sbs clean --dry-run   # Preview what would be cleaned
sbs clean --force     # Force cleanup without confirmation
```

#### Global Options
```bash
sbs --config ~/.config/sbs/custom.json  # Use custom config file
sbs --verbose                           # Enable verbose logging
sbs --help                             # Show help for any command
```

## Architecture

### Core Components
- **CLI Framework**: Built with Cobra for command structure
- **Interactive TUI**: Uses Bubble Tea and Lipgloss for terminal UI
- **Git Integration**: go-git for worktree and branch management
- **Session Management**: Tracks metadata in JSON files

### Package Structure
- `cmd/`: Cobra command definitions (start, stop, list, attach, clean)
- `pkg/config/`: Configuration management and session metadata
- `pkg/git/`: Git operations and worktree management
- `pkg/tmux/`: Tmux session management
- `pkg/sandbox/`: Sandbox environment coordination
- `pkg/tui/`: Terminal UI components and styling
- `pkg/issue/`: GitHub issue integration
- `pkg/repo/`: Repository management
- `pkg/validation/`: Tool validation utilities

### Sandbox Integration

SBS integrates with the `sandbox` command to provide isolated development environments. Each work session creates a named sandbox that contains the development environment.

#### Sandbox Naming Convention
```bash
# Format: work-issue-{repo-name}-{issue-number}[-{title-slug}]
work-issue-myproject-123                    # Basic format
work-issue-myproject-123-fix-login-bug      # With SBS_TITLE environment variable
```

#### Sandbox Management Commands
```bash
# List all running sandboxes
sandbox list

# Check status of specific sandbox
sandbox --name work-issue-myproject-123 status

# Access files in sandbox
sandbox --name work-issue-myproject-123 ls /work
sandbox --name work-issue-myproject-123 cat /work/README.md

# Execute commands in sandbox
sandbox --name work-issue-myproject-123 pwd
sandbox --name work-issue-myproject-123 git status

# Show changes made in sandbox
sandbox --name work-issue-myproject-123 diff

# Accept/sync changes from sandbox to host
sandbox --name work-issue-myproject-123 accept

# Stop sandbox (preserves files)
sandbox --name work-issue-myproject-123 stop

# Delete sandbox and all files
sandbox --name work-issue-myproject-123 delete
```

#### Sandbox Network and Bind Mounts
The `work-issue.sh` script configures sandboxes with:
```bash
--net="host"                    # Host network access
--bind /tmp/tmux-1000          # Tmux socket access
```

#### Environment Variables in Sandbox
```bash
# SBS_TITLE is automatically passed to sandbox for friendly naming
export SBS_TITLE="Fix login bug"
sbs start 123  # Creates: work-issue-myproject-123-fix-login-bug
```

### Configuration

#### Configuration Files
- Config stored in `~/.config/sbs/config.json`
- Sessions tracked in `~/.config/sbs/sessions.json` (global) and repository-specific files
- Worktrees created in `~/.work-issue-worktrees/` by default
- Sandbox storage in `~/.sandboxes/` (default sandbox location)

#### Example config.json
```json
{
  "worktree_base_path": "/home/user/.work-issue-worktrees",
  "github_token": "ghp_your_github_token_here",
  "work_issue_script": "/home/user/code/work-issue/work-issue.sh",
  "repo_path": "."
}
```

#### Configuration Options
- **worktree_base_path**: Directory where git worktrees are created (default: `~/.work-issue-worktrees/`)
- **github_token**: GitHub personal access token for API access (optional, falls back to `gh` CLI)
- **work_issue_script**: Path to work-issue.sh script (optional, defaults to current directory)
- **repo_path**: Repository path to use (default: current directory ".")

#### Environment Variables
```bash
# Set issue title for sandbox naming
export SBS_TITLE="Fix authentication bug"

# Custom config file location
sbs --config /path/to/custom/config.json start 123

# Enable verbose logging
export SBS_VERBOSE=1
# or use flag
sbs --verbose start 123
```

### Session Management

#### Session Lifecycle
1. Creates git branch `issue-{number}-{title-slug}`
2. Creates worktree in configured directory
3. Launches tmux session `work-issue-{number}`
4. Executes `work-issue.sh` script in sandboxed environment
5. Tracks session metadata for management

#### Typical Workflow Patterns

**Single Issue Development:**
```bash
# Start work on issue #123
sbs start 123

# Work in the sandbox environment (Claude Code, git, etc.)
# ... development happens in tmux session ...

# When done, stop the session (keeps worktree)
sbs stop 123

# Later, resume work
sbs start 123 --resume    # Skip work-issue.sh execution
```

**Multiple Concurrent Issues:**
```bash
# Work on multiple issues simultaneously
sbs start 123    # Issue #123 in work-issue-123 tmux session
sbs start 456    # Issue #456 in work-issue-456 tmux session

# List active sessions
sbs list         # Or just 'sbs' for TUI

# Switch between sessions
sbs attach 123   # Attach to issue #123 session
sbs attach 456   # Attach to issue #456 session
```

**Session Cleanup:**
```bash
# Regular cleanup of stale sessions
sbs clean --dry-run    # Preview what will be cleaned
sbs clean              # Clean with confirmation
sbs clean --force      # Clean without confirmation

# Manual cleanup if needed
sbs stop 123          # Stop specific session
# Worktree remains in ~/.work-issue-worktrees/issue-123/
```

#### Session Metadata Tracking
Sessions are tracked with the following information:
- Issue number and title
- Git branch name (`issue-{number}-{title-slug}`)
- Worktree path (`~/.work-issue-worktrees/issue-{number}/`)
- Tmux session name (`work-issue-{number}`)
- Sandbox name (`work-issue-{repo}-{number}[-{title}]`)
- Creation timestamp and status

### Claude Code Hook Integration

SBS automatically installs Claude Code hooks within sandbox environments to capture development activity and tool usage data.

#### Automatic Hook Installation
The hook installation is completely automatic:
1. **During `sbs start`**: The `work-issue.sh` script installs the hook in the sandbox
2. **Sandbox-Only**: Hook runs only inside sandbox, not on host system
3. **No Manual Setup**: Everything configured automatically with proper permissions

#### Hook Functionality
- **Triggers**: PostToolUse events (after each Claude Code tool execution)
- **Captures**: All Claude Code tool usage within the sandbox environment  
- **Data Storage**: Creates `.sbs/stop.json` file inside the sandbox
- **Isolation**: Hook configuration only affects the sandbox Claude Code setup

#### Accessing Hook Data
```bash
# Check what Claude Code was doing in a sandbox
sandbox --name work-issue-myproject-123 cat .sbs/stop.json

# Pretty-print the hook data (if jq available on host)
sandbox --name work-issue-myproject-123 cat .sbs/stop.json | jq .

# List hook output files
sandbox --name work-issue-myproject-123 ls -la .sbs/

# Check if hook is working
sandbox --name work-issue-myproject-123 ls .sbs/
```

#### Hook Output Format
The hook creates `.sbs/stop.json` with structure like:
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
    // Claude Code session info, tool data, execution context, etc.
  }
}
```

#### Troubleshooting Hook Issues
- **Hook Not Installing**: Verify `scripts/claude-code-stop-hook.sh` exists and is executable
- **No Hook Data**: Ensure Claude Code is actually running within the sandbox environment
- **Permission Issues**: The `work-issue.sh` script handles all permissions automatically
- **Missing Dependencies**: Hook will warn if `jq` is not available but will still function
- **Sandbox Access**: Use `sandbox --name [name] ls .sbs/` to verify hook directory exists

### Dependencies
- Git with worktree support
- tmux for session management
- GitHub CLI (`gh`) for issue metadata
- `sandbox` command for containerized execution
- `work-issue.sh` script integration
- `jq` (optional, for JSON formatting in hooks)