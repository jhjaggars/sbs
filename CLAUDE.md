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

### Development Workflow
```bash
go run . start 123  # Run without building
./sbs start 123     # Start session for issue #123
./sbs list          # List active sessions (interactive TUI)
./sbs attach 123    # Attach to existing session
./sbs stop 123      # Stop a session
./sbs clean         # Clean up stale sessions
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

### Configuration
- Config stored in `~/.config/sbs/config.json`
- Sessions tracked in `~/.config/sbs/sessions.json` (global) and repository-specific files
- Worktrees created in `~/.work-issue-worktrees/` by default

### Session Lifecycle
1. Creates git branch `issue-{number}-{title-slug}`
2. Creates worktree in configured directory
3. Launches tmux session `work-issue-{number}`
4. Executes `work-issue.sh` script in sandboxed environment
5. Tracks session metadata for management

### Dependencies
- Git with worktree support
- tmux for session management
- GitHub CLI (`gh`) for issue metadata
- `sandbox` command for containerized execution
- `work-issue.sh` script integration