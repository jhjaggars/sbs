# Requirements Specification: Interactive Issue Selection for `sbs start`

## Problem Statement

Currently, `sbs start` requires an issue number as a mandatory argument. Users must know the specific issue number they want to work on, requiring them to check GitHub separately or remember issue numbers. This creates friction in the workflow and breaks the seamless experience of starting work on an issue.

## Solution Overview

Enhance the existing `sbs start` command to support interactive issue selection when no issue number is provided. When run without arguments, `sbs start` will:

1. Fetch open issues from the current repository using GitHub CLI
2. Present an interactive TUI for issue selection with search capabilities  
3. Once an issue is selected, proceed with the existing workflow (branch creation, worktree setup, tmux session, work-issue.sh launch)

## Functional Requirements

### FR1: Command Signature Enhancement
- **Current**: `sbs start <issue-number>` (required argument)
- **Enhanced**: `sbs start [issue-number]` (optional argument)
- **Behavior**: 
  - With argument: existing behavior unchanged
  - Without argument: launch interactive issue selection

### FR2: Repository Context Validation
- Must be run from within a git repository (both modes)
- Repository validation occurs before argument checking
- Clear error message when run outside a repository

### FR3: Issue Fetching
- Fetch only open issues from current repository
- Use GitHub CLI (`gh issue list`) with JSON output
- Handle authentication errors from gh CLI
- Handle repositories with no open issues gracefully

### FR4: Interactive Issue Selection Interface
- Display issue numbers and titles in a list format
- Support keyboard navigation (↑/↓ or j/k)
- Enter key selects an issue and proceeds with workflow
- 'q' key exits the command entirely
- Consistent styling with existing TUI components

### FR5: Search and Filtering
- Always-visible search input field at top of interface
- Server-side search using GitHub's search API (`gh issue list --search`)
- Real-time filtering as user types
- Clear search results when search is cleared

### FR6: Pagination Support
- Hardcoded limit (reasonable default like 100 issues)
- Handle repositories with many issues
- Show indication when results are limited

### FR7: Error Handling
- No open issues: display message and exit cleanly
- GitHub API errors: display helpful error messages
- Network connectivity issues: graceful error handling
- Invalid repository: clear error message

## Technical Requirements

### TR1: GitHub Client Extension
- **File**: `pkg/issue/github.go`
- **New Method**: `ListIssues(searchQuery string, limit int) ([]Issue, error)`
- **Implementation**: Use `gh issue list --json number,title,state,url --state open --limit N [--search QUERY]`
- **Error Handling**: Follow existing pattern from `GetIssue()` method

### TR2: Issue Selection TUI Component
- **File**: `pkg/tui/issueselect.go` (new file)
- **Framework**: Charmbracelet Bubble Tea (existing dependency)
- **Components**: 
  - List display using existing table patterns
  - TextInput component for search (from bubbles package)
  - Consistent key bindings with existing TUI
- **Styling**: Reuse existing styles from `pkg/tui/styles.go`

### TR3: Command Integration
- **File**: `cmd/start.go`
- **Change**: Modify `Args: cobra.ExactArgs(1)` to `Args: cobra.MaxArgs(1)`
- **Logic**: Add branching in `runStart()` function:
  ```
  1. Validate repository context (existing logic)
  2. Check argument count:
     - If args provided: existing workflow
     - If no args: launch issue selection TUI
  3. After issue selection: continue with existing workflow
  ```

### TR4: Data Flow
1. User runs `sbs start` (no args)
2. Validate repository context 
3. Launch issue selection TUI
4. TUI fetches issues via GitHub client
5. User searches/selects issue
6. Return selected issue number to main workflow
7. Continue with existing `sbs start <number>` logic

## Implementation Hints and Patterns

### GitHub Client Pattern
Follow existing pattern in `pkg/issue/github.go:31-56`:
- Use `exec.Command("gh", ...)` for CLI interaction
- Parse JSON output with struct unmarshaling
- Handle stderr for specific error conditions
- Return custom errors for "no issues found" case

### TUI Component Pattern  
Follow existing pattern in `pkg/tui/model.go`:
- Implement `tea.Model` interface (Init, Update, View methods)
- Use existing key binding patterns from `pkg/tui/model.go:27-56`
- Reuse table display patterns from `pkg/tui/model.go:186-240`
- Handle window resize messages

### Integration Pattern
Follow existing validation pattern in `cmd/start.go:42-47`:
- Repository detection before other operations
- Consistent error message formatting
- Use existing manager initialization pattern

## Acceptance Criteria

### AC1: Backward Compatibility
- [ ] `sbs start <issue-number>` works exactly as before
- [ ] All existing command flags and options remain functional
- [ ] No breaking changes to existing workflows

### AC2: Interactive Selection
- [ ] `sbs start` (no args) launches issue selection interface
- [ ] Interface displays issue numbers and titles clearly
- [ ] Navigation with arrow keys and j/k works
- [ ] Enter key selects issue and continues workflow
- [ ] 'q' key exits cleanly

### AC3: Search Functionality
- [ ] Search box is visible and functional
- [ ] Typing updates results in real-time via server-side search
- [ ] Clearing search shows all issues again
- [ ] Search works with GitHub's search syntax

### AC4: Error Handling
- [ ] Clear error when run outside git repository
- [ ] Helpful message when no open issues exist
- [ ] GitHub authentication errors are handled gracefully
- [ ] Network errors display appropriate messages

### AC5: Integration
- [ ] Selected issue proceeds with identical workflow to `sbs start <number>`
- [ ] Branch creation, worktree setup, tmux session work as expected
- [ ] work-issue.sh launches with selected issue number

## Assumptions

- GitHub CLI (`gh`) is installed and authenticated (existing requirement)
- User has appropriate repository permissions for issue access
- Terminal supports TUI rendering (existing assumption)
- Network connectivity for GitHub API access
- Repository has reasonable number of issues (pagination handles edge cases)

## Out of Scope

- Multi-repository issue selection
- Issue creation from selection interface  
- Advanced filtering by labels/assignees in TUI
- Local caching of issue data
- Configuration of pagination limits
- Closed issue selection
- Issue preview/description display