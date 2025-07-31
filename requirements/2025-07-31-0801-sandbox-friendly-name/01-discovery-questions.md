# Discovery Questions

Based on the codebase analysis, here are the key questions to understand the problem space:

## Q1: Should the SBS_TITLE environment variable be available to all commands in the tmux session?
**Default if unknown:** Yes (environment variables are typically inherited by all processes in the session)

## Q2: Should the friendly name generation handle non-English characters in issue titles?
**Default if unknown:** Yes (replace non-ASCII characters with safe alternatives for maximum compatibility)

## Q3: Should the friendly name have a maximum length limit for sandbox compatibility?
**Default if unknown:** Yes (limit to 64 characters to ensure compatibility with various systems)

## Q4: Should the friendly name generation be consistent with the existing branch naming patterns?
**Default if unknown:** Yes (follow similar sanitization rules as the existing issue branch creation)

## Q5: Should the SBS_TITLE environment variable be set for both new sessions and when attaching to existing sessions?
**Default if unknown:** Yes (consistent environment across all session interactions)

## Context from Codebase Analysis

Current session management flow:
1. Issue title is fetched from GitHub API via `issue.GetIssue()`
2. Branch names are created using `git.CreateIssueBranch()` with title sanitization
3. Tmux sessions are created via `tmux.CreateSession()` 
4. Commands are executed using `tmux.ExecuteCommand()` with environment variables via `syscall.Exec()`

Key files involved:
- `cmd/start.go:122-130` - Issue fetching and title handling
- `pkg/tmux/manager.go:92-94` - Environment variable handling for tmux attach
- `pkg/repo/manager.go:148-167` - Existing sanitization logic in `SanitizeName()`
- `pkg/git/manager.go` - Branch creation with title sanitization (need to examine)

The codebase already has patterns for:
- Issue title fetching from GitHub
- Name sanitization for filesystem/session compatibility  
- Environment variable passing to tmux sessions
- Repository-scoped session naming