# AGENT.md - Development Guidelines for SBS

## Build/Lint/Test Commands
- `make build` - Build the sbs binary
- `make test` - Run all tests (`go test ./...`)
- `make fmt` - Format Go code
- `make lint` - Run golangci-lint
- `go test ./pkg/[package]` - Run single package tests
- `go run . start 123` - Run without building
- Install pre-commit hooks: `./scripts/install-hooks.sh`

## Architecture & Structure
- **CLI Framework**: Cobra commands in `cmd/` (start, stop, list, attach, clean)
- **Core Packages**: 
  - `pkg/config/` - Configuration and session metadata
  - `pkg/git/` - Git operations and worktree management  
  - `pkg/tmux/` - Tmux session management
  - `pkg/sandbox/` - Sandbox environment coordination
  - `pkg/tui/` - Terminal UI (Bubble Tea + Lipgloss)
  - `pkg/inputsource/` - Pluggable input sources (GitHub, JIRA, test)
  - `pkg/issue/`, `pkg/repo/`, `pkg/validation/` - External integrations

## Code Style & Conventions
- **Imports**: Standard library first, then external deps, then internal packages
- **Error Handling**: Return errors, don't log and continue
- **Naming**: Use Go conventions (camelCase for unexported, PascalCase for exported)
- **Testing**: Use testify/assert, table-driven tests for multiple scenarios
- **Configuration**: JSON files in `~/.config/sbs/`, structs with json tags
- **Logging**: Use command logging via `pkg/cmdlog` for external commands
