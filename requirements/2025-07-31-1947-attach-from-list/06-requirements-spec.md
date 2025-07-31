# Requirements Specification: List View Attach Indication

## Problem Statement

Users can attach to running sessions from the list view by pressing Enter, but this functionality is not clearly indicated in the condensed help text. The attach capability is only visible when users press `?` to view the full help, leading to poor discoverability of this primary feature.

## Solution Overview

Enhance the condensed help text in the TUI list view to clearly indicate that pressing Enter will attach to the selected session, following existing UX patterns from the issue selection interface.

## Functional Requirements

### FR-1: Condensed Help Text Enhancement
- **MUST** modify the condensed help text to include "enter to attach" at the beginning
- **MUST** apply to both repository view mode and global view mode
- **MUST** maintain consistency with existing issueselect.go help text pattern
- **MUST** keep text concise to fit on one terminal line

### FR-2: View Mode Consistency  
- **MUST** update both help text variations in pkg/tui/model.go:
  - Standard condensed help (line 247)
  - Non-repository condensed help (line 249)

## Technical Requirements

### TR-1: File Modifications
- **File:** `pkg/tui/model.go`
- **Lines to modify:** 247, 249
- **Pattern to follow:** Similar to `pkg/tui/issueselect.go:335`

### TR-2: Help Text Format
- **Current format:** "Press ? for help, g to toggle view, r to refresh, q to quit"
- **New format:** "Press enter to attach, ? for help, g to toggle view, r to refresh, q to quit"
- **Consistency:** Follow same comma-separated format as issue selection TUI

## Implementation Hints

### Existing Pattern Reference
In `pkg/tui/issueselect.go:335`:
```
helpText := "\nPress ? for help, tab to search, enter to select, q to quit"
```

### Target Implementation
```go
// Line 247
helpText := "\nPress enter to attach, ? for help, g to toggle view, r to refresh, q to quit"

// Line 249  
helpText := "\nNot in git repository - showing global view. Press enter to attach, ? for help, r to refresh, q to quit"
```

## Acceptance Criteria

1. **AC-1:** When users view the list interface, they can see "enter to attach" in the condensed help
2. **AC-2:** Help text appears in both repository and global view modes
3. **AC-3:** Text remains concise and fits on standard terminal width
4. **AC-4:** "enter to attach" appears first in the help sequence for prominence
5. **AC-5:** Format matches existing application help text patterns

## Assumptions

- No changes needed to the actual attachment functionality (already working)
- No changes needed to the full help view (already contains correct text)
- Terminal width supports the slightly longer condensed help text
- Users expect primary actions to be mentioned in condensed help

## Out of Scope

- Modifying attachment behavior or error handling
- Adding visual feedback during attachment process
- Changing full help view content
- Adding new keyboard shortcuts