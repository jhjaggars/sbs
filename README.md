# Work Orchestrator

A command-line tool that orchestrates GitHub issue work environments with automatic git worktree and tmux session management.

## Features

- **Automated Workflow**: Creates git branches, worktrees, and tmux sessions automatically
- **Issue Integration**: Fetches GitHub issue metadata for better organization
- **Beautiful TUI**: Interactive terminal interface built with Charm.sh tools
- **Session Management**: Track, list, and attach to multiple concurrent work sessions
- **Cleanup Tools**: Automatic cleanup of stale sessions and worktrees

## Installation

```bash
go build -o work-orchestrator
```

## Usage

### Start a new work session
```bash
./work-orchestrator start 123
```

This will:
1. Create/switch to branch `issue-123-<issue-title-slug>`
2. Create worktree in `~/.work-issue-worktrees/issue-123/`
3. Launch tmux session `work-issue-123`
4. Execute `work-issue.sh 123` in the session

### List active sessions
```bash
./work-orchestrator list          # Interactive TUI
./work-orchestrator list --plain  # Plain text output
```

### Attach to an existing session
```bash
./work-orchestrator attach 123
```

### Stop a session
```bash
./work-orchestrator stop 123
```

### Clean up stale sessions
```bash
./work-orchestrator clean
./work-orchestrator clean --dry-run  # Preview changes
```

## Configuration

Configuration is stored in `~/.config/work-orchestrator/config.json`:

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