# Requirements Specification: Stop and Clean Sessions from List

## Problem Statement

Users currently need to exit the interactive list interface and run separate `sbs stop` and `sbs clean` commands to manage sessions. This creates workflow friction and context switching that reduces efficiency. Users should be able to perform session management operations directly from the list interface where they can see session status and make informed decisions.

## Solution Overview

Extend the interactive TUI list interface with keyboard shortcuts to stop individual sessions ('s' key) and clean stale sessions ('c' key). Operations will work on the currently selected session (stop) or visible sessions in current view mode (clean), with automatic UI refresh and proper error handling.

## Functional Requirements

### FR-1: Stop Session from List
- **MUST** add 's' keyboard shortcut to stop the currently selected session
- **MUST** work on the session highlighted by the cursor in the list
- **MUST** execute the same logic as `cmd/stop.go` (kill tmux session, delete sandbox, update status)
- **MUST** display operation results (success/error) to the user
- **MUST** automatically refresh the session list after operation completes
- **MUST** handle cases where session is already stopped gracefully

### FR-2: Clean Stale Sessions from List  
- **MUST** add 'c' keyboard shortcut to clean stale sessions
- **MUST** work only on sessions visible in current view mode (repository vs global)
- **MUST** identify stale sessions (tmux session no longer exists)
- **MUST** show modal confirmation dialog before proceeding
- **MUST** execute same cleanup logic as `cmd/clean.go` (remove worktrees, sandboxes, update sessions)
- **MUST** display cleanup results to user
- **MUST** automatically refresh session list after cleanup

### FR-3: Modal Confirmation Dialog
- **MUST** implement modal overlay for clean operation confirmation
- **MUST** show number of stale sessions that will be cleaned
- **MUST** list affected sessions (issue numbers and titles)
- **MUST** support y/n or Enter/Escape key bindings for confirmation
- **MUST** be dismissible without performing cleanup

### FR-4: Enhanced Help Text
- **MUST** update condensed help text to include 's' and 'c' shortcuts
- **MUST** maintain existing help text format and length constraints
- **MUST** apply to both repository and global view modes
- **MUST** follow pattern established in attach-from-list enhancement

### FR-5: Error Handling and Feedback
- **MUST** display errors prominently in the UI (similar to existing error display pattern)
- **MUST** refresh session list even when operations fail
- **MUST** provide clear feedback on operation success/failure
- **MUST** handle edge cases (no sessions selected, no stale sessions found, etc.)

## Technical Requirements

### TR-1: Key Binding Extensions
- **File:** `pkg/tui/model.go`
- **Add to keyMap struct (lines 17-25):**
  - `Stop` binding for 's' key with help text "stop session"  
  - `Clean` binding for 'c' key with help text "clean stale"
- **Add to keys variable (lines 27-56):** New key bindings following existing pattern

### TR-2: Message Types
- **File:** `pkg/tui/model.go`
- **Add new message types:**
  - `stopSessionMsg` struct with error field and success indicator
  - `cleanSessionsMsg` struct with error field and cleanup results
  - `confirmationDialogMsg` struct for modal state management

### TR-3: Model State Extensions
- **File:** `pkg/tui/model.go`
- **Add to Model struct (lines 65-77):**
  - `showConfirmationDialog bool` - tracks modal dialog state
  - `confirmationMessage string` - content for confirmation dialog
  - `pendingCleanSessions []config.SessionMetadata` - sessions to be cleaned

### TR-4: Command Functions
- **File:** `pkg/tui/model.go`
- **Implement functions:**
  - `stopSelectedSession()` - extracts stop logic from `cmd/stop.go`
  - `cleanStaleSessions()` - extracts clean logic from `cmd/clean.go`  
  - `showCleanConfirmation()` - prepares confirmation dialog
  - `executeCleanup()` - performs actual cleanup after confirmation

### TR-5: Update Method Extensions
- **File:** `pkg/tui/model.go`
- **Modify Update method (lines 107-163):**
  - Add cases for 's' and 'c' key handling
  - Add message handling for new message types
  - Add confirmation dialog state management
  - Integrate with existing refresh pattern

### TR-6: View Method Extensions  
- **File:** `pkg/tui/model.go`
- **Modify View method (lines 165-253):**
  - Add modal confirmation dialog rendering
  - Update help text to include new shortcuts (lines 242-247)

### TR-7: Styling for Modal Dialog
- **File:** `pkg/tui/styles.go`
- **Add new styles:**
  - `modalBackgroundStyle` - semi-transparent background
  - `modalContentStyle` - dialog box styling
  - `confirmationTextStyle` - content formatting

## Implementation Hints

### Reusable Logic Extraction
Extract core logic from existing commands:
```go
// From cmd/stop.go:27-101
func stopSessionByMetadata(session config.SessionMetadata) error {
    // Extract tmux and sandbox stopping logic
}

// From cmd/clean.go:28-155  
func cleanStaleSessionsInView(sessions []config.SessionMetadata) ([]config.SessionMetadata, error) {
    // Extract stale session detection and cleanup logic
}
```

### Key Binding Pattern
Follow existing pattern in `pkg/tui/model.go:27-56`:
```go
Stop: key.NewBinding(
    key.WithKeys("s"),
    key.WithHelp("s", "stop session"),
),
Clean: key.NewBinding(
    key.WithKeys("c"),  
    key.WithHelp("c", "clean stale"),
),
```

### Message Handling Pattern
Follow existing pattern in `pkg/tui/model.go:149-163`:
```go
case stopSessionMsg:
    m.error = msg.err
    return m, m.refreshSessions()

case cleanSessionsMsg:
    m.error = msg.err
    m.showConfirmationDialog = false
    return m, m.refreshSessions()
```

### Modal Dialog Implementation
Render over existing content using lipgloss positioning:
```go
if m.showConfirmationDialog {
    dialog := modalContentStyle.Render(m.confirmationMessage)
    content = modalBackgroundStyle.Render(content) 
    // Position dialog in center
}
```

## Acceptance Criteria

1. **AC-1:** Press 's' on any session in list to stop it immediately
2. **AC-2:** Press 'c' to see confirmation dialog listing stale sessions in current view
3. **AC-3:** Confirmation dialog shows session count and titles, responds to y/n keys
4. **AC-4:** Both operations refresh the list automatically and show success/error feedback
5. **AC-5:** Help text includes "s to stop, c to clean" in condensed help
6. **AC-6:** Operations work correctly in both repository and global view modes
7. **AC-7:** Error states (no selection, no stale sessions) are handled gracefully
8. **AC-8:** Modal dialog can be dismissed without performing cleanup

## Assumptions

- Users prefer keyboard shortcuts over mouse interactions in terminal interfaces
- Current session selection (cursor position) is the intended target for stop operations
- Clean operations should respect current view context rather than always working globally
- Modal confirmation is preferred over status line prompts for destructive operations
- Existing error display pattern in `pkg/tui/model.go:166-168` is adequate for new errors

## Out of Scope

- Bulk operations (selecting multiple sessions)
- Custom confirmation messages or settings
- Undo functionality for cleanup operations
- Progress indicators for long-running operations
- Integration with external notification systems
- Keyboard shortcuts other than 's' and 'c'