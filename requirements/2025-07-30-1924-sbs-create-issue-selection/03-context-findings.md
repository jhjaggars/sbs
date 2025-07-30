# Context Findings

## Codebase Structure Analysis

### Key Components
- **Go-based CLI application** using Cobra framework at `cmd/` 
- **Current command**: `sbs start <issue-number>` (cmd/start.go:15)
- **GitHub integration**: Uses `gh` CLI via pkg/issue/github.go:32
- **TUI framework**: Charmbracelet Bubble Tea already implemented in pkg/tui/
- **Session management**: Sophisticated tmux and worktree management

### Existing GitHub Integration
- GitHub client in pkg/issue/github.go:31 fetches single issues via `gh issue view`
- Current implementation only supports fetching specific issue numbers
- No existing functionality to list all issues

### TUI Infrastructure Already Exists
- Full TUI model in pkg/tui/model.go with session selection interface
- Styles and formatting already defined in pkg/tui/styles.go
- Key bindings, navigation, and selection patterns established
- Pattern: Up/Down navigation, Enter to select, Q to quit

### Architecture Patterns
- Cobra command structure in cmd/
- Manager pattern for different services (tmux, git, repo, etc.)
- Configuration system with JSON persistence
- Repository-aware session management

### Key Files for Implementation
- **cmd/start.go**: Current implementation of `sbs start <issue-number>`
- **pkg/issue/github.go**: GitHub integration layer
- **pkg/tui/model.go**: TUI interface patterns
- **cmd/root.go**: Command registration

### Missing Components
- No command currently handles `sbs create` (different from `sbs start`)
- No functionality to fetch issue lists (only individual issues)
- No issue selection TUI (current TUI is for session management)

### Technical Constraints
- Requires `gh` CLI to be installed and authenticated
- Uses `gh issue list` command pattern (similar to existing `gh issue view`)
- Must integrate with existing Cobra command structure
- Should follow established TUI patterns and styling

## GitHub CLI Integration Analysis

### Available `gh issue list` Options
- **JSON output**: `--json fields` with fields like `number,title,state,url`
- **Pagination**: `--limit int` (default 30) for controlling number of issues
- **State filtering**: `--state string` (open|closed|all, default "open")  
- **Search capability**: `--search query` for filtering issues
- **Label filtering**: `--label strings` for label-based filtering

### Existing Dependencies
- **Charmbracelet Bubble Tea**: v1.3.6 already available for TUI
- **Charmbracelet Bubbles**: v0.21.0 for UI components (textinput for search)
- **Charmbracelet Lipgloss**: v1.1.0 for styling

### Implementation Requirements Based on Discovery Answers

1. **Command Signature Change**: cmd/start.go:24 needs `Args: cobra.MaxArgs(1)` instead of `cobra.ExactArgs(1)`

2. **GitHub Client Extension**: pkg/issue/github.go needs new `ListIssues()` method using `gh issue list --json number,title,state,url --state open --limit N`

3. **Issue Selection TUI**: New TUI component needed with:
   - Issue list display (similar to pkg/tui/model.go session list)
   - Search/filter functionality (using textinput component)
   - Pagination support for large repositories
   - Consistent keybindings (j/k, enter, q)

4. **Integration Points**:
   - cmd/start.go runStart() function needs branching logic
   - When no args: show issue selection TUI
   - When issue selected: continue with existing workflow
   - Repository validation still required (no multi-repo support)

### Implementation Strategy
- Extend pkg/issue/github.go with ListIssues method
- Create new pkg/tui/issueselect.go for issue selection TUI
- Modify cmd/start.go to handle both argument patterns
- Reuse existing styling and patterns from pkg/tui/