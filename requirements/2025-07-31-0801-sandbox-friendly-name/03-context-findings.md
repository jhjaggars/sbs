# Context Findings

## Current Architecture Analysis

### Issue Title Sanitization Pattern
The codebase already has title sanitization in `pkg/git/manager.go:139-157`:
```go
func (m *Manager) formatBranchName(issueNumber int, issueTitle string) string {
    slug := strings.ToLower(issueTitle)
    reg := regexp.MustCompile(`[^a-z0-9]+`)
    slug = reg.ReplaceAllString(slug, "-")
    slug = strings.Trim(slug, "-")
    if len(slug) > 50 {
        slug = slug[:50]
        slug = strings.TrimSuffix(slug, "-")
    }
    return fmt.Sprintf("issue-%d-%s", issueNumber, slug)
}
```

### Environment Variable Handling in Tmux
Current tmux attachment in `pkg/tmux/manager.go:83-101`:
```go
func (m *Manager) AttachToSession(sessionName string) error {
    tmuxPath, err := exec.LookPath("tmux")
    if err != nil {
        return fmt.Errorf("tmux command not found: %w", err)
    }
    
    args := []string{"tmux", "attach-session", "-t", sessionName}
    env := os.Environ()  // <-- This is where we can add SBS_TITLE
    
    err = syscall.Exec(tmuxPath, args, env)
    // ...
}
```

### Session Creation Flow
Current session creation in `pkg/tmux/manager.go:28-68`:
```go
func (m *Manager) CreateSession(issueNumber int, workingDir, sessionName string) (*Session, error) {
    // Creates new detached session
    cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir)
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("failed to create tmux session %s: %w", sessionName, err)
    }
    // ...
}
```

## Key Integration Points

### 1. Title Sanitization Function Location
- **File**: `pkg/git/manager.go` or new utility package
- **Pattern**: Similar to `formatBranchName()` but with 32-char limit
- **Input**: Issue title string
- **Output**: Sanitized friendly name

### 2. Environment Variable Setting Locations
- **Session Creation**: `pkg/tmux/manager.go:CreateSession()` - needs env vars for initial session
- **Session Attachment**: `pkg/tmux/manager.go:AttachToSession()` - line 92 env modification
- **Command Execution**: `pkg/tmux/manager.go:ExecuteCommand()` - needs env context

### 3. Issue Title Access Points
- **Start Command**: `cmd/start.go:122-130` - `githubIssue.Title` is already available
- **Attach Command**: `cmd/attach.go:40-44` - session metadata has `IssueTitle` field
- **Session Metadata**: `pkg/config/config.go:19-31` - `SessionMetadata.IssueTitle` stores the title

### 4. Repository-Scoped Integration
- **Current Pattern**: `pkg/repo/manager.go:47-55` - repository-scoped naming
- **Sanitization**: `pkg/repo/manager.go:148-167` - existing `SanitizeName()` method (30-char limit)

## Implementation Strategy

### Option 1: Extend Existing Sanitization
Modify `pkg/repo/manager.go:SanitizeName()` to accept length parameter and use for both repo names and issue titles.

### Option 2: New Utility Function
Create new sanitization function specifically for issue titles with 32-char limit.

### Option 3: Issue Title Caching
Store sanitized title in `SessionMetadata` to avoid re-sanitization on attach.

## Required Changes Summary

### Files to Modify:
1. **`pkg/tmux/manager.go`**:
   - `CreateSession()`: Add environment variable setting for new sessions
   - `AttachToSession()`: Add SBS_TITLE to environment before exec
   - New method: `createSessionEnvironment()` helper

2. **`cmd/start.go`**:
   - Generate friendly name from issue title
   - Pass friendly name to tmux operations

3. **`cmd/attach.go`**:
   - Retrieve issue title from session metadata
   - Generate/retrieve friendly name for environment

4. **New/Modified Utility**:
   - Create sanitization function for issue titles with 32-char limit

### Environment Variable Flow:
```
Issue Title → Sanitize (32 chars) → SBS_TITLE env var → Tmux Session → All Commands
```

### Existing Patterns to Follow:
- Repository-scoped session naming: `work-issue-{repo}-{number}`
- Sanitization regex: `[^a-zA-Z0-9]+` replaced with `-`  
- Length trimming with hyphen cleanup
- Environment passing via `os.Environ()` modification