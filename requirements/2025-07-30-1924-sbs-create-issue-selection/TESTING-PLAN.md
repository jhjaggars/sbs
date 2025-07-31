# Testing Plan: Interactive Issue Selection for `sbs start`

## Overview

This testing plan covers the comprehensive testing strategy for implementing interactive issue selection in the `sbs start` command. The plan follows test-driven development (TDD) principles and is organized by testing levels, from unit tests to end-to-end scenarios.

## Testing Framework and Tools

### Go Testing Stack
- **Framework**: Go standard `testing` package
- **Assertions**: `testify/assert` and `testify/require` (to be added to dependencies)
- **Mocking**: `testify/mock` for external dependencies
- **Test Coverage**: `go test -cover` and `go tool cover`

### Dependencies to Add to go.mod
```go
require (
    github.com/stretchr/testify v1.8.4
)
```

## Test Organization Structure

```
pkg/
├── issue/
│   ├── github.go
│   ├── github_test.go          # Unit tests for GitHub client
│   └── test_helpers.go         # Test utilities and fixtures
├── tui/
│   ├── issueselect.go          # New TUI component (to be implemented)
│   ├── issueselect_test.go     # Unit tests for issue selection TUI
│   └── tui_test_helpers.go     # TUI testing utilities
└── cmd/
    ├── start.go
    └── start_test.go           # Integration tests for start command
```

## Testing Priority and Implementation Order

### Phase 1: Foundation Tests (Implement First)
1. **GitHub Client Unit Tests** - Test `ListIssues` method
2. **Command Argument Parsing Tests** - Test optional argument handling
3. **Error Handling Tests** - Test failure scenarios

### Phase 2: Component Tests
4. **TUI Component Unit Tests** - Test issue selection interface
5. **Integration Tests** - Test component interactions

### Phase 3: System Tests
6. **End-to-End Tests** - Test complete workflows
7. **Manual Testing Scenarios** - Human verification

---

## 1. Unit Tests

### 1.1 GitHub Client Tests (`pkg/issue/github_test.go`)

#### Test Coverage for `ListIssues` Method

```go
func TestGitHubClient_ListIssues(t *testing.T) {
    // Test cases to implement:
    
    // Success cases
    t.Run("successful_list_with_no_search", func(t *testing.T) {
        // Mock gh command returning JSON issue list
        // Verify correct parsing of issue data
        // Assert expected number of issues returned
    })
    
    t.Run("successful_list_with_search_query", func(t *testing.T) {
        // Mock gh command with --search parameter
        // Verify search query is passed correctly
        // Assert filtered results
    })
    
    t.Run("successful_list_with_limit", func(t *testing.T) {
        // Mock gh command with --limit parameter
        // Verify limit is applied correctly
        // Assert result count respects limit
    })
    
    // Edge cases
    t.Run("empty_issue_list", func(t *testing.T) {
        // Mock gh command returning empty JSON array
        // Verify graceful handling of no issues
        // Assert empty slice returned, no error
    })
    
    t.Run("single_issue_result", func(t *testing.T) {
        // Mock gh command returning single issue
        // Verify single issue is parsed correctly
    })
    
    // Error cases
    t.Run("gh_command_not_found", func(t *testing.T) {
        // Mock exec.Command to return "command not found"
        // Verify appropriate error message
    })
    
    t.Run("gh_authentication_error", func(t *testing.T) {
        // Mock gh command returning auth error in stderr
        // Verify specific auth error handling
    })
    
    t.Run("invalid_json_response", func(t *testing.T) {
        // Mock gh command returning malformed JSON
        // Verify JSON parse error handling
    })
    
    t.Run("gh_command_exit_error", func(t *testing.T) {
        // Mock gh command with non-zero exit code
        // Verify error propagation
    })
    
    t.Run("network_connectivity_error", func(t *testing.T) {
        // Mock network-related error scenarios
        // Verify appropriate error messages
    })
}
```

#### Test Utilities and Fixtures

```go
// Test fixtures for GitHub API responses
var sampleIssuesJSON = `[
    {
        "number": 123,
        "title": "Fix authentication bug",
        "state": "open",
        "url": "https://github.com/owner/repo/issues/123"
    },
    {
        "number": 124,
        "title": "Add dark mode support",
        "state": "open", 
        "url": "https://github.com/owner/repo/issues/124"
    }
]`

// Mock command executor for testing
type mockCommandExecutor struct {
    mockOutput []byte
    mockError  error
}

func (m *mockCommandExecutor) executeCommand(cmd *exec.Cmd) ([]byte, error) {
    return m.mockOutput, m.mockError
}
```

### 1.2 TUI Component Tests (`pkg/tui/issueselect_test.go`)

#### Test Coverage for Issue Selection Interface

```go
func TestIssueSelectModel(t *testing.T) {
    t.Run("initialization", func(t *testing.T) {
        // Test initial model state
        // Verify default values
        // Assert search box is empty
        // Assert no issues initially loaded
    })
    
    t.Run("loading_issues", func(t *testing.T) {
        // Test issue loading into model
        // Mock issue data
        // Verify display formatting
    })
    
    t.Run("keyboard_navigation", func(t *testing.T) {
        // Test up/down arrow keys
        // Test j/k key bindings
        // Verify cursor movement
        // Test boundary conditions (first/last item)
    })
    
    t.Run("issue_selection", func(t *testing.T) {
        // Test Enter key selection
        // Verify correct issue is selected
        // Test selection with empty list
    })
    
    t.Run("search_functionality", func(t *testing.T) {
        // Test search input handling
        // Mock search API calls
        // Verify real-time filtering
        // Test search clear functionality
    })
    
    t.Run("quit_behavior", func(t *testing.T) {
        // Test 'q' key quit
        // Test Ctrl+C quit
        // Verify quit message propagation
    })
    
    t.Run("error_display", func(t *testing.T) {
        // Test error state rendering
        // Verify error message display
        // Test recovery from error state
    })
    
    t.Run("window_resize_handling", func(t *testing.T) {
        // Test responsive layout
        // Verify table width adjustments
        // Test with various terminal sizes
    })
}
```

#### TUI Testing Utilities

```go
// Mock GitHub client for TUI testing
type mockGitHubClient struct {
    issues []issue.Issue
    err    error
}

func (m *mockGitHubClient) ListIssues(searchQuery string, limit int) ([]issue.Issue, error) {
    if m.err != nil {
        return nil, m.err
    }
    
    // Filter issues based on search query for testing
    if searchQuery == "" {
        return m.issues, nil
    }
    
    var filtered []issue.Issue
    for _, issue := range m.issues {
        if strings.Contains(strings.ToLower(issue.Title), strings.ToLower(searchQuery)) {
            filtered = append(filtered, issue)
        }
    }
    return filtered, nil
}

// Helper to create test model with mock data
func createTestModel(issues []issue.Issue, err error) *IssueSelectModel {
    mockClient := &mockGitHubClient{issues: issues, err: err}
    return NewIssueSelectModel(mockClient)
}

// Key press simulation helper
func simulateKeyPress(model tea.Model, key string) tea.Model {
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
    newModel, _ := model.Update(msg)
    return newModel
}
```

### 1.3 Command Argument Parsing Tests (`cmd/start_test.go`)

```go
func TestStartCommand_ArgumentParsing(t *testing.T) {
    t.Run("with_issue_number_argument", func(t *testing.T) {
        // Test existing behavior with issue number
        // Verify argument is parsed correctly
        // Assert existing workflow is triggered
    })
    
    t.Run("without_arguments", func(t *testing.T) {
        // Test new behavior with no arguments
        // Verify interactive selection is triggered
        // Mock repository detection
    })
    
    t.Run("with_resume_flag", func(t *testing.T) {
        // Test --resume flag behavior
        // Test with and without issue number
        // Verify flag precedence
    })
    
    t.Run("invalid_issue_number", func(t *testing.T) {
        // Test non-numeric issue number
        // Verify error handling
    })
    
    t.Run("too_many_arguments", func(t *testing.T) {
        // Test cobra.MaxArgs(1) validation
        // Verify appropriate error message
    })
}
```

---

## 2. Integration Tests

### 2.1 Command Flow Integration (`cmd/start_integration_test.go`)

```go
func TestStartCommand_Integration(t *testing.T) {
    t.Run("full_interactive_workflow", func(t *testing.T) {
        // Setup: Mock repository, GitHub client, TUI
        // Execute: Run sbs start with no args
        // Verify: Issue selection UI launches
        // Simulate: User selects issue
        // Assert: Normal workflow continues with selected issue
    })
    
    t.Run("repository_validation_before_issue_selection", func(t *testing.T) {
        // Setup: Mock non-git directory
        // Execute: Run sbs start with no args
        // Assert: Repository error occurs before TUI launch
    })
    
    t.Run("github_client_integration", func(t *testing.T) {
        // Setup: Mock gh command execution
        // Execute: Launch issue selection
        // Verify: GitHub API calls are made correctly
        // Assert: Issues are fetched and displayed
    })
    
    t.Run("issue_selection_to_workflow_handoff", func(t *testing.T) {
        // Setup: Mock full environment
        // Execute: Complete issue selection
        // Verify: Selected issue number is passed to existing workflow
        // Assert: Branch creation, worktree, tmux session creation
    })
}
```

### 2.2 GitHub CLI Integration Tests (`pkg/issue/github_integration_test.go`)

```go
func TestGitHubClient_CLIIntegration(t *testing.T) {
    // Note: These tests require gh command to be available
    // Use build tags to conditionally run
    
    t.Run("real_gh_command_execution", func(t *testing.T) {
        // Skip if INTEGRATION_TESTS env var not set
        if os.Getenv("INTEGRATION_TESTS") == "" {
            t.Skip("Skipping integration test")
        }
        
        // Test with real gh command
        // Verify JSON parsing works with actual API response
        // Test rate limiting and authentication
    })
}
```

### 2.3 TUI to Command Integration (`pkg/tui/integration_test.go`)

```go
func TestTUICommandIntegration(t *testing.T) {
    t.Run("issue_selection_result_handling", func(t *testing.T) {
        // Setup: Mock TUI with issues
        // Execute: Simulate issue selection
        // Verify: Correct issue number is returned
        // Assert: Command can continue with result
    })
    
    t.Run("tui_error_propagation", func(t *testing.T) {
        // Setup: Mock TUI with error state
        // Execute: Handle TUI errors
        // Verify: Errors are propagated to command level
        // Assert: Appropriate error messages displayed
    })
}
```

---

## 3. End-to-End Tests

### 3.1 Complete User Workflows (`e2e/workflows_test.go`)

```go
func TestE2E_UserWorkflows(t *testing.T) {
    // Note: Requires test environment setup
    
    t.Run("complete_issue_selection_workflow", func(t *testing.T) {
        // Setup: Test repository with real .git
        // Mock: GitHub CLI responses
        // Execute: Full sbs start command
        // Verify: 
        //   - Repository detection works
        //   - Issues are fetched from GitHub
        //   - TUI displays correctly
        //   - Issue selection works
        //   - Branch/worktree/tmux creation succeeds
        //   - work-issue.sh launches
    })
    
    t.Run("search_and_select_workflow", func(t *testing.T) {
        // Setup: Large issue list
        // Execute: Search for specific issue
        // Verify: Search filters correctly
        // Select: Issue from filtered results
        // Assert: Correct issue is selected
    })
    
    t.Run("quit_from_issue_selection", func(t *testing.T) {
        // Setup: Launch issue selection
        // Execute: Press 'q' to quit
        // Verify: Command exits cleanly
        // Assert: No artifacts created
    })
}
```

### 3.2 Error Handling Scenarios (`e2e/error_scenarios_test.go`)

```go
func TestE2E_ErrorScenarios(t *testing.T) {
    t.Run("no_open_issues_scenario", func(t *testing.T) {
        // Setup: Repository with no open issues
        // Mock: Empty issue list from GitHub
        // Execute: sbs start (no args)
        // Verify: Appropriate "no issues" message
        // Assert: Clean exit
    })
    
    t.Run("github_authentication_failure", func(t *testing.T) {
        // Setup: Unauthenticated gh CLI
        // Mock: Authentication error
        // Execute: sbs start (no args)
        // Verify: Clear auth error message
        // Assert: Instructions for authentication
    })
    
    t.Run("network_connectivity_failure", func(t *testing.T) {
        // Setup: Network unavailable scenario
        // Mock: Network timeout/connection error
        // Execute: sbs start (no args)
        // Verify: Network error handling
        // Assert: Helpful error message
    })
    
    t.Run("invalid_repository_scenario", func(t *testing.T) {
        // Setup: Directory that's not a git repository
        // Execute: sbs start (no args)
        // Verify: Repository validation error
        // Assert: Clear error message before TUI launch
    })
}
```

---

## 4. Edge Case Tests

### 4.1 Boundary Conditions (`test/edge_cases_test.go`)

```go
func TestEdgeCases(t *testing.T) {
    t.Run("large_issue_list_pagination", func(t *testing.T) {
        // Setup: Repository with 200+ issues
        // Mock: Paginated GitHub API response
        // Execute: Issue selection with limit
        // Verify: Only specified number of issues shown
        // Assert: Pagination indicator displayed
    })
    
    t.Run("very_long_issue_titles", func(t *testing.T) {
        // Setup: Issues with extremely long titles
        // Execute: Display in TUI
        // Verify: Titles are truncated appropriately
        // Assert: UI remains readable
    })
    
    t.Run("special_characters_in_titles", func(t *testing.T) {
        // Setup: Issues with Unicode, emojis, special chars
        // Execute: Display and search functionality
        // Verify: Characters render correctly
        // Assert: Search works with special characters
    })
    
    t.Run("empty_search_results", func(t *testing.T) {
        // Setup: Issues loaded
        // Execute: Search for non-existent term
        // Verify: Empty results handled gracefully
        // Assert: Clear "no results" message
    })
    
    t.Run("rapid_search_input", func(t *testing.T) {
        // Setup: Issues loaded
        // Execute: Rapid typing in search box
        // Verify: Search debouncing works
        // Assert: Performance remains acceptable
    })
    
    t.Run("terminal_resize_during_operation", func(t *testing.T) {
        // Setup: TUI running
        // Execute: Simulate terminal resize
        // Verify: Layout adapts correctly
        // Assert: No crashes or artifacts
    })
}
```

### 4.2 Performance and Resource Tests (`test/performance_test.go`)

```go
func TestPerformance(t *testing.T) {
    t.Run("large_issue_list_rendering", func(t *testing.T) {
        // Setup: 100 issues
        // Execute: TUI rendering
        // Measure: Render time
        // Assert: Acceptable performance (<100ms)
    })
    
    t.Run("search_performance", func(t *testing.T) {
        // Setup: Many issues
        // Execute: Search operations
        // Measure: Search response time
        // Assert: Real-time feel maintained
    })
    
    t.Run("memory_usage_with_large_dataset", func(t *testing.T) {
        // Setup: Maximum issue limit
        // Execute: Full workflow
        // Monitor: Memory consumption
        // Assert: Reasonable memory usage
    })
}
```

---

## 5. Manual Testing Scenarios

### 5.1 UI/UX Validation Checklist

#### Visual and Interactive Testing
- [ ] **Issue List Display**
  - Issues display with proper formatting (number, title)
  - Long titles are truncated with ellipsis
  - List scrolls properly with keyboard navigation
  - Selection highlighting is visible and clear

- [ ] **Search Functionality**
  - Search box is prominently displayed at top
  - Typing immediately filters results
  - Search works with partial matches
  - Clear search (Ctrl+A, Delete) shows all issues again
  - Search placeholder text is helpful

- [ ] **Keyboard Navigation**
  - ↑/↓ arrow keys move selection
  - j/k keys work as alternatives
  - Page Up/Page Down work for long lists
  - Home/End keys jump to first/last item
  - Tab moves between search and list areas

- [ ] **Responsive Design**
  - Interface adapts to different terminal sizes
  - Minimum usable width is reasonable
  - Very wide terminals use space effectively
  - Terminal resize is handled gracefully

#### Error State Testing
- [ ] **No Issues Available**
  - Clear message when repository has no open issues
  - Provides guidance on what to do next
  - Exit behavior is intuitive

- [ ] **GitHub API Errors**
  - Authentication errors show clear instructions
  - Network errors are user-friendly
  - Rate limiting is handled gracefully
  - Repository access errors are informative

- [ ] **Search Edge Cases**
  - Empty search results show helpful message
  - Special characters in search work correctly
  - Search with no matching results handles well
  - Very long search terms are handled

### 5.2 Integration Testing Checklist

#### Workflow Continuity
- [ ] **Issue Selection to Branch Creation**
  - Selected issue proceeds with normal workflow
  - Branch naming follows existing patterns
  - No data loss between TUI and command execution

- [ ] **Session Management**
  - Selected issue creates proper session metadata
  - Session appears in `sbs list` correctly
  - Resuming session works as expected

- [ ] **work-issue.sh Integration**
  - Script receives correct issue number
  - Terminal/tmux title updates properly
  - All environment setup is identical to direct `sbs start <number>`

#### Backward Compatibility
- [ ] **Existing Command Behavior**
  - `sbs start <number>` works exactly as before
  - All flags and options remain functional
  - No performance regression in existing workflows
  - Error messages remain consistent

### 5.3 Accessibility Testing

#### Keyboard-Only Navigation
- [ ] All functionality accessible via keyboard
- [ ] Tab order is logical and intuitive
- [ ] No mouse-dependent features
- [ ] Screen reader compatibility (basic test)

#### Terminal Compatibility
- [ ] Works in various terminal emulators (iTerm, Terminal.app, xterm, etc.)
- [ ] Color scheme works in dark/light terminals
- [ ] Functions properly over SSH connections
- [ ] Works with tmux/screen multiplexers

---

## 6. Test Data and Fixtures

### 6.1 Test Repository Setup

```bash
# Script to create test repository for integration tests
#!/bin/bash
mkdir -p test_repo
cd test_repo
git init
echo "# Test Repository" > README.md
git add README.md
git commit -m "Initial commit"
git remote add origin https://github.com/test/repo.git
```

### 6.2 Mock GitHub API Responses

```go
// pkg/issue/test_fixtures.go
var TestIssues = []Issue{
    {
        Number: 1,
        Title:  "Fix authentication bug in login flow",
        State:  "open",
        URL:    "https://github.com/owner/repo/issues/1",
    },
    {
        Number: 2,
        Title:  "Add dark mode support to user interface",
        State:  "open", 
        URL:    "https://github.com/owner/repo/issues/2",
    },
    {
        Number: 15,
        Title:  "Refactor database connection pooling for better performance and reliability",
        State:  "open",
        URL:    "https://github.com/owner/repo/issues/15",
    },
    {
        Number: 42,
        Title:  "🐛 Unicode handling in search functionality needs improvement",
        State:  "open",
        URL:    "https://github.com/owner/repo/issues/42",
    },
}

var EmptyIssueListResponse = `[]`

var ErrorResponses = struct {
    AuthError    string
    NetworkError string
    InvalidJSON  string
}{
    AuthError:    "gh: authentication failed",
    NetworkError: "network unreachable",
    InvalidJSON:  `{"invalid": json}`,
}
```

---

## 7. Continuous Integration Testing

### 7.1 GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.24.4
    
    - name: Install GitHub CLI
      run: |
        curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
        sudo apt update
        sudo apt install gh
    
    - name: Run Unit Tests
      run: go test -v ./pkg/...
    
    - name: Run Integration Tests
      env:
        INTEGRATION_TESTS: "true"
      run: go test -v -tags=integration ./...
    
    - name: Test Coverage
      run: |
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload Coverage
      uses: actions/upload-artifact@v3
      with:
        name: coverage-report
        path: coverage.html
```

### 7.2 Test Coverage Targets

- **Unit Tests**: Minimum 80% coverage for new code
- **Integration Tests**: Cover all major user workflows
- **Critical Paths**: 100% coverage for error handling and edge cases

---

## 8. Testing Timeline and Milestones

### Phase 1: Foundation (Week 1)
- [ ] Set up testing framework and utilities
- [ ] Implement GitHub client unit tests
- [ ] Create command argument parsing tests
- [ ] Set up CI pipeline

### Phase 2: Component Testing (Week 2)
- [ ] Implement TUI component unit tests
- [ ] Create integration test framework
- [ ] Implement GitHub CLI integration tests

### Phase 3: System Testing (Week 3)
- [ ] Implement end-to-end tests
- [ ] Create error scenario tests
- [ ] Implement edge case tests
- [ ] Performance testing

### Phase 4: Validation (Week 4)
- [ ] Manual testing scenarios
- [ ] Accessibility testing
- [ ] Backward compatibility verification
- [ ] Documentation and test maintenance

---

## 9. Test Maintenance and Best Practices

### 9.1 Test Organization Principles
- **Arrange-Act-Assert**: Structure all tests clearly
- **One Assertion Per Test**: Keep tests focused and readable
- **Descriptive Names**: Test names should explain the scenario
- **Fast Execution**: Unit tests should run in milliseconds
- **Isolated Tests**: No dependencies between test cases

### 9.2 Mock and Stub Guidelines
- **Mock External Dependencies**: GitHub CLI, file system, network calls
- **Use Interfaces**: Define interfaces for mockable components
- **Minimal Mocking**: Don't mock everything, focus on boundaries
- **Realistic Test Data**: Use data that represents real-world scenarios

### 9.3 Test Data Management
- **Version Control**: Include test fixtures in repository
- **Realistic Data**: Based on actual GitHub issue formats
- **Edge Case Coverage**: Include boundary and error conditions
- **Privacy**: Ensure no sensitive data in test fixtures

---

## 10. Success Criteria

### Test Quality Metrics
- [ ] **Code Coverage**: Minimum 80% for new functionality
- [ ] **Test Execution Time**: Full test suite completes in under 30 seconds
- [ ] **Test Reliability**: Zero flaky tests in CI pipeline
- [ ] **Error Coverage**: All error conditions have corresponding tests

### Functional Validation
- [ ] **Backward Compatibility**: All existing functionality preserved
- [ ] **Performance**: No regression in existing command performance
- [ ] **User Experience**: Manual testing scenarios pass validation
- [ ] **Documentation**: All test procedures documented and maintained

This comprehensive testing plan ensures robust validation of the interactive issue selection feature while maintaining the high quality and reliability of the existing `sbs` command-line tool.