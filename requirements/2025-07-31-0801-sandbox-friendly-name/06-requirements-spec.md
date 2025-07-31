# Requirements Specification: Sandbox-Friendly Issue Title Environment Variable

## Problem Statement

When starting tmux sessions for GitHub issues, users need access to a sanitized version of the issue title as an environment variable (`SBS_TITLE`) that is compatible with sandbox environments and contains only alphanumeric characters and hyphens.

## Solution Overview

Implement sandbox-friendly name generation from GitHub issue titles and pass the sanitized name as the `SBS_TITLE` environment variable to all tmux session operations (creation, attachment, command execution).

## Functional Requirements

### FR1: Friendly Name Generation
- **FR1.1**: Generate sandbox-friendly names from GitHub issue titles
- **FR1.2**: Sanitize names to contain only alphanumeric characters and hyphens  
- **FR1.3**: Convert to lowercase for consistency
- **FR1.4**: Limit friendly names to 32 characters maximum
- **FR1.5**: Handle non-English characters by replacing with safe alternatives
- **FR1.6**: Use fallback format `{reponame}-issue-{number}` when issue title unavailable

### FR2: Environment Variable Setting
- **FR2.1**: Set `SBS_TITLE` environment variable for new tmux sessions
- **FR2.2**: Set `SBS_TITLE` environment variable when attaching to existing sessions
- **FR2.3**: Include `SBS_TITLE` when executing commands in sessions
- **FR2.4**: Ensure environment variable is available to all processes in the session

### FR3: Data Persistence
- **FR3.1**: Store friendly title in `SessionMetadata` to avoid re-computation
- **FR3.2**: Maintain consistency between session creation and attachment operations

## Technical Requirements

### TR1: Code Modifications

#### TR1.1: Extend `pkg/repo/manager.go`
```go
// Modify SanitizeName to accept length parameter
func (m *Manager) SanitizeName(name string, maxLength int) string
```

#### TR1.2: Modify `pkg/tmux/manager.go`
```go 
// Update CreateSession signature
func (m *Manager) CreateSession(issueNumber int, workingDir, sessionName string, env map[string]string) (*Session, error)

// Update AttachToSession to set environment
func (m *Manager) AttachToSession(sessionName string, env map[string]string) error

// Update ExecuteCommand methods to include environment
func (m *Manager) ExecuteCommand(sessionName, command string, args []string, env map[string]string) error
```

#### TR1.3: Update `pkg/config/config.go`
```go
type SessionMetadata struct {
    // ... existing fields
    FriendlyTitle  string `json:"friendly_title"`  // New field
}
```

#### TR1.4: Modify `cmd/start.go`
- Generate friendly title from `githubIssue.Title`
- Pass friendly title to tmux operations via environment map
- Store friendly title in session metadata

#### TR1.5: Modify `cmd/attach.go`
- Retrieve friendly title from session metadata
- Pass friendly title to tmux attach operation

### TR2: Implementation Patterns

#### TR2.1: Sanitization Logic
- Use regex pattern: `[^a-zA-Z0-9]+` replaced with `-`
- Trim leading/trailing hyphens
- Apply 32-character length limit
- Handle truncation at word boundaries when possible

#### TR2.2: Environment Variable Handling
- Modify `os.Environ()` slice before tmux operations
- Add `SBS_TITLE={friendly_name}` to environment
- Ensure backward compatibility with existing code

#### TR2.3: Fallback Strategy
```go
func generateFriendlyTitle(repoName string, issueNumber int, issueTitle string) string {
    if issueTitle == "" {
        return fmt.Sprintf("%s-issue-%d", sanitizeRepoName(repoName), issueNumber)
    }
    return sanitizeName(issueTitle, 32)
}
```

## Implementation Hints

### File Modification Order
1. **`pkg/repo/manager.go`** - Extend `SanitizeName()` method
2. **`pkg/config/config.go`** - Add `FriendlyTitle` field to `SessionMetadata`
3. **`pkg/tmux/manager.go`** - Update all methods to accept environment variables
4. **`cmd/start.go`** - Generate and use friendly titles
5. **`cmd/attach.go`** - Retrieve and use stored friendly titles

### Environment Variable Integration
```go
// Helper function to create environment with SBS_TITLE
func createTmuxEnvironment(friendlyTitle string) map[string]string {
    return map[string]string{
        "SBS_TITLE": friendlyTitle,
    }
}

// Usage in session operations
env := createTmuxEnvironment(friendlyTitle)
session, err := tmuxManager.CreateSession(issueNumber, workingDir, sessionName, env)
```

### Backward Compatibility
- Make environment parameter optional by using `map[string]string{}` as default
- Existing callers continue to work without modification
- New functionality only activated when environment map is provided

## Acceptance Criteria

### AC1: Session Creation
- **Given** a GitHub issue with title "Fix user authentication bug"
- **When** running `sbs start 123`  
- **Then** tmux session is created with `SBS_TITLE=fix-user-authentication-bug`

### AC2: Session Attachment  
- **Given** an existing session for issue #123
- **When** running `sbs attach 123`
- **Then** attached session has `SBS_TITLE` environment variable set

### AC3: Command Execution
- **Given** an active session with friendly title
- **When** executing commands via `ExecuteCommand()`
- **Then** commands have access to `SBS_TITLE` environment variable

### AC4: Fallback Handling
- **Given** GitHub API is unavailable for issue #123 in repo "myproject"
- **When** starting a session
- **Then** `SBS_TITLE=myproject-issue-123` is set

### AC5: Length Limitation
- **Given** an issue title exceeding 32 characters
- **When** generating friendly name
- **Then** result is truncated to 32 characters with clean hyphen boundaries

### AC6: Character Sanitization
- **Given** issue title "Fix café login (UTF-8 encoding)"
- **When** generating friendly name  
- **Then** `SBS_TITLE=fix-cafe-login-utf-8-encoding`

## Assumptions

- GitHub CLI (`gh` command) remains the primary method for fetching issue data
- Tmux sessions are the primary execution environment for work-issue scripts
- Repository names are already properly sanitized via existing `SanitizeName()` method
- Session metadata structure can be extended without breaking existing installations
- Environment variables are the preferred method for passing data to sandbox environments

## Related Features

- Existing branch naming using issue titles (`pkg/git/manager.go:formatBranchName`)
- Repository-scoped session naming (`pkg/repo/manager.go`)
- Session metadata persistence (`pkg/config/config.go`)
- Tmux session management (`pkg/tmux/manager.go`)