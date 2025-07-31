# Expert Detail Questions

Based on the detailed codebase analysis, here are the specific implementation questions:

## Q1: Should we modify the existing `tmux.CreateSession()` method to accept environment variables, or create a new method variant?
**Default if unknown:** Modify existing method to accept optional environment map (maintains backward compatibility while adding functionality)

## Q2: Should the friendly name be stored in `SessionMetadata` to avoid re-sanitization on attach operations?
**Default if unknown:** Yes (store as new field `FriendlyTitle` for performance and consistency)

## Q3: Should we create a new sanitization utility or extend the existing `repo.Manager.SanitizeName()` method?
**Default if unknown:** Create new utility in `pkg/utils` package (separates concerns and allows different length limits)

## Q4: Should the SBS_TITLE environment variable be set only during session creation or also when executing commands via `ExecuteCommand()`?
**Default if unknown:** Both (ensures environment consistency across all tmux operations)

## Q5: Should we handle the case where issue title is unavailable (e.g., network issues) by using a fallback friendly name?
**Default if unknown:** Yes (fallback to "issue-{number}" format for reliability)

## Technical Context

### Current Environment Handling:
- `pkg/tmux/manager.go:92` - `env := os.Environ()` before `syscall.Exec()`
- No environment modification for session creation currently
- Command execution doesn't modify environment

### Existing Sanitization Patterns:
- `pkg/git/manager.go:139-157` - Issue title to branch slug (50 char limit)
- `pkg/repo/manager.go:148-167` - Repository name sanitization (30 char limit)
- Both use `[^a-zA-Z0-9]+` regex pattern

### Session Metadata Schema:
```go
type SessionMetadata struct {
    IssueTitle     string `json:"issue_title"`     // Already exists
    // Proposed addition:
    // FriendlyTitle  string `json:"friendly_title"`  // New field
}
```

### Integration Points:
- `cmd/start.go:122` - Issue title available as `githubIssue.Title`
- `cmd/attach.go:41` - Session metadata available with `session.IssueTitle`
- Both commands need to generate and pass friendly name to tmux operations