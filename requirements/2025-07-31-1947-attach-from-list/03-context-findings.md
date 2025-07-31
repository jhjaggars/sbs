# Context Findings

## Current Implementation Analysis

### Existing Functionality
- **Location**: `pkg/tui/model.go:131-136`
- **Current Behavior**: Enter key already triggers `m.attachToSession(sessionName)` 
- **Implementation**: Uses `tmux.Manager.AttachToSession()` which executes `syscall.Exec()` to replace current process

### Help Text Status
- **Full Help View** (`pkg/tui/model.go:265`): "enter  - Attach to selected session"
- **Condensed Help** (`pkg/tui/model.go:247`): "Press ? for help, g to toggle view, r to refresh, q to quit"
- **Missing**: No mention of "enter to attach" in the condensed help text

### Issue Identification
The functionality already exists and works correctly. The problem is that users don't see what the Enter key does unless they:
1. Press `?` to toggle full help, OR
2. Read the command description when running `sbs list --help`

### Key Files Requiring Changes
- `pkg/tui/model.go:247` - Add "enter to attach" to condensed help text
- `pkg/tui/model.go:249` - Add "enter to attach" to non-repo condensed help text

### Similar Patterns in Codebase
- `pkg/tui/issueselect.go:335` shows similar condensed help: "Press ? for help, tab to search, enter to select, q to quit"
- Pattern: Shows primary action (enter) in condensed help

### Technical Constraints
- Help text must fit on one line in condensed view
- Should maintain consistency with existing help text patterns
- No changes needed to attachment logic itself