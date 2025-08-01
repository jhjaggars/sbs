# Context Findings

## Existing TUI Architecture Analysis

### Current List Interface Structure
- **File:** `pkg/tui/model.go`
- **Current Key Bindings:** Up/Down (j/k), Enter (attach), Quit (q), Help (?), Refresh (r), Toggle View (g)
- **Existing Command Pattern:** Uses `tea.Cmd` functions that return messages processed in `Update()` method
- **Current Actions:** Only attach to session via Enter key

### Key TUI Components Found
1. **Key Binding Structure (`pkg/tui/model.go:17-56`)**
   - Uses `github.com/charmbracelet/bubbles/key` for key management
   - Current bindings: Up, Down, Enter, Quit, Help, Refresh, ToggleView
   - Pattern: `key.NewBinding()` with keys, help text

2. **Message Handling Pattern (`pkg/tui/model.go:107-163`)**
   - Uses typed messages: `refreshMsg`, `attachMsg`
   - Commands return functions that return messages
   - Error handling through message structs

3. **Existing Session Operations**
   - **Attach:** `attachToSession()` function (`pkg/tui/model.go:357-362`)
   - **Stop Logic:** Available in `cmd/stop.go:27-101`
   - **Clean Logic:** Available in `cmd/clean.go:28-155`

## Command Implementation Patterns

### Stop Command Analysis (`cmd/stop.go`)
- **Core Logic:** 
  - Load sessions (`config.LoadSessions()`)
  - Find session by issue number (`issueTracker.FindSessionByIssue()`)
  - Kill tmux session (`tmuxManager.KillSession()`)
  - Delete sandbox (`sandboxManager.DeleteSandbox()`)
  - Update session status to "stopped"
  - Save updated sessions (`config.SaveSessions()`)

### Clean Command Analysis (`cmd/clean.go`)
- **Core Logic:**
  - Load all repository sessions (`config.LoadAllRepositorySessions()`)
  - Check which sessions are stale (tmux session doesn't exist)
  - Optionally show confirmation dialog
  - Remove worktree directories and sandboxes
  - Save updated sessions list (removing stale ones)

### Session Management Integration
- **Session Metadata:** `config.SessionMetadata` struct with status tracking
- **Multi-Repository Support:** Uses `LoadAllRepositorySessions()` for global view
- **Status Updates:** Sessions have "active", "stopped", "stale" status values

## Related Features and Patterns

### Similar Enhancement: Attach from List
- **Reference:** `requirements/2025-07-31-1947-attach-from-list/06-requirements-spec.md`
- **Pattern:** Enhanced help text to show existing functionality
- **Implementation:** Modified condensed help text in `pkg/tui/model.go:242-247`

### Key Binding Extension Pattern
Looking at existing key bindings, new actions would follow this pattern:
```go
Stop: key.NewBinding(
    key.WithKeys("s"),
    key.WithHelp("s", "stop session"),
),
Clean: key.NewBinding(
    key.WithKeys("c"),
    key.WithHelp("c", "clean stale sessions"),
),
```

## Technical Integration Points

### Files That Need Modification
1. **`pkg/tui/model.go`** - Add new key bindings, message types, and command functions
2. **`pkg/tui/styles.go`** - Potentially add confirmation dialog styles

### Reusable Components
- **Stop Logic:** Can extract from `cmd/stop.go:27-101` into reusable function
- **Clean Logic:** Can extract from `cmd/clean.go:28-155` into reusable function
- **Session Loading:** Already available via `config.LoadSessions()` and `config.LoadAllRepositorySessions()`
- **Status Updates:** Pattern established in stop command for updating session status

### Confirmation Dialog Requirements
- **For Clean Operations:** Need confirmation dialog similar to `cmd/clean.go:90-98`
- **TUI Implementation:** Would need modal dialog or confirmation prompt
- **User Interaction:** Could use y/n key bindings within confirmation state

## Architecture Considerations

### Message-Based Architecture
The TUI uses Bubble Tea's message-passing architecture:
- Actions trigger commands that return messages
- Messages are processed in `Update()` method
- UI updates happen in response to state changes

### Session Context Awareness
- Current implementation tracks current repository context
- View mode (Repository vs Global) affects which sessions are displayed
- Stop/Clean operations should respect current view mode context

### Error Handling Pattern
Existing pattern shows errors are stored in model state and displayed in `View()` method rather than as separate error dialogs.