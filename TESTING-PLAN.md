# Testing Plan for Issue #15: Make bare `sbs` command launch TUI and `sbs list` show plain text output

## Overview

This testing plan covers the comprehensive test strategy for implementing the feature that restructures the command interface to make the TUI more accessible as the default interface while providing a quick plain text option for scripting/non-interactive use.

## Feature Requirements Summary

**Current Behavior:**
- `sbs` → Shows help/usage
- `sbs list` → Launches interactive TUI
- `sbs list --plain` → Shows plain text output

**Target Behavior:**
- `sbs` → Launch interactive TUI (current `sbs list` behavior)
- `sbs list` → Show plain text output (current `sbs list --plain` behavior)
- `sbs list --plain` → Still show plain text output (backward compatibility)

## Test Structure and Organization

### 1. Unit Tests

#### 1.1 Root Command Tests (`cmd/root_test.go`)

**New test cases to add:**

```go
func TestRootCommand_LaunchesTUI(t *testing.T)
func TestRootCommand_NoArguments(t *testing.T)
func TestRootCommand_HelpOutput(t *testing.T)
func TestRootCommand_TUIInitialization(t *testing.T)
```

**Test Coverage:**
- Verify root command now has a RunE function that launches TUI
- Test that root command with no arguments launches TUI instead of showing help
- Verify help text is updated to reflect new behavior
- Test TUI initialization without side effects (using mocks)

#### 1.2 List Command Tests (`cmd/list_test.go` - NEW FILE)

**Primary test cases:**

```go
func TestListCommand_ShowsPlainTextByDefault(t *testing.T)
func TestListCommand_PlainFlagBackwardCompatibility(t *testing.T)
func TestListCommand_HelpTextUpdated(t *testing.T)
func TestListCommand_ArgumentParsing(t *testing.T)
func TestListCommand_Structure(t *testing.T)
func TestListCommand_OutputFormatting(t *testing.T)
```

**Test Coverage:**
- Verify `sbs list` outputs plain text by default (no TUI)
- Test `sbs list --plain` still works (backward compatibility)
- Verify help text reflects new behavior
- Test command structure and flag handling
- Test plain text output formatting consistency

#### 1.3 Command Integration Tests

**Test file:** `cmd/command_integration_test.go` (NEW FILE)

```go
func TestCommandBehaviorTransition(t *testing.T)
func TestBackwardCompatibility(t *testing.T)
func TestHelpTextConsistency(t *testing.T)
```

**Test Coverage:**
- Verify the behavior transition doesn't break existing functionality
- Test all combinations of command usage
- Ensure help text is consistent across commands

### 2. Integration Tests

#### 2.1 End-to-End Behavior Tests

**Test file:** `integration_command_behavior_test.go` (NEW FILE)

**Test Cases:**
```go
func TestE2E_RootCommandLaunchesTUI(t *testing.T)
func TestE2E_ListCommandPlainOutput(t *testing.T)
func TestE2E_BackwardCompatibilityPreserved(t *testing.T)
func TestE2E_TerminalDetection(t *testing.T)
```

**Test Coverage:**
- Test actual command execution with process spawning
- Verify TUI launches correctly from root command
- Test plain text output generation
- Test behavior in different terminal environments

#### 2.2 TUI Integration Tests

**Enhancement to existing:** `pkg/tui/model_test.go`

**New test cases:**
```go
func TestTUI_LaunchedFromRootCommand(t *testing.T)
func TestTUI_SessionDataLoading(t *testing.T)
func TestTUI_ErrorHandlingFromRoot(t *testing.T)
```

### 3. Backward Compatibility Tests

#### 3.1 Regression Test Suite

**Test file:** `regression_test.go` (NEW FILE)

**Test Categories:**
```go
func TestRegression_ExistingCommands(t *testing.T)
func TestRegression_ConfigLoading(t *testing.T)
func TestRegression_SessionManagement(t *testing.T)
func TestRegression_OtherSubcommands(t *testing.T)
```

**Test Coverage:**
- Verify `start`, `stop`, `attach`, `clean`, `version` commands unchanged
- Test configuration loading still works
- Verify session management functionality preserved
- Test that no existing workflows are broken

#### 3.2 Flag Compatibility Tests

**Test file:** `cmd/flag_compatibility_test.go` (NEW FILE)

```go
func TestFlagCompatibility_ListPlainFlag(t *testing.T)
func TestFlagCompatibility_GlobalFlags(t *testing.T)
func TestFlagCompatibility_HelpBehavior(t *testing.T)
```

### 4. Edge Cases and Error Conditions

#### 4.1 Error Handling Tests

**Test cases to add across relevant files:**

```go
func TestErrorHandling_NoSessions(t *testing.T)
func TestErrorHandling_ConfigLoadFailure(t *testing.T)
func TestErrorHandling_TUIInitFailure(t *testing.T)
func TestErrorHandling_InvalidTerminal(t *testing.T)
```

**Test Coverage:**
- Test behavior when no sessions exist
- Test graceful handling of configuration errors
- Test TUI initialization failures
- Test behavior in invalid terminal environments

#### 4.2 Terminal Environment Tests

**Test file:** `cmd/terminal_test.go` (NEW FILE)

```go
func TestTerminal_InteractiveDetection(t *testing.T)
func TestTerminal_NonInteractiveEnvironment(t *testing.T)
func TestTerminal_WidthDetection(t *testing.T)
func TestTerminal_PipeOutput(t *testing.T)
```

**Test Coverage:**
- Test interactive vs non-interactive terminal detection
- Test behavior when stdout is piped
- Test terminal width detection for formatting
- Test output formatting in various terminal conditions

### 5. Mock Requirements

#### 5.1 TUI Mocks

```go
type MockTUIModel struct {
    launched bool
    sessions []config.SessionMetadata
    error    error
}

func (m *MockTUIModel) Init() tea.Cmd
func (m *MockTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *MockTUIModel) View() string
```

#### 5.2 Terminal Mocks

```go
type MockTerminal struct {
    isInteractive bool
    width         int
    height        int
}

func (m *MockTerminal) IsInteractive() bool
func (m *MockTerminal) Size() (int, int)
```

#### 5.3 Session Loading Mocks

```go
type MockSessionLoader struct {
    sessions []config.SessionMetadata
    error    error
}

func (m *MockSessionLoader) LoadAllRepositorySessions() ([]config.SessionMetadata, error)
```

### 6. Test Data and Fixtures

#### 6.1 Test Session Data

**File:** `testdata/sessions.json`

```json
[
  {
    "issue_number": 123,
    "issue_title": "Test issue for TUI display",
    "repository_name": "test-repo",
    "branch": "issue-123-test-branch",
    "tmux_session": "work-issue-123",
    "last_activity": "2025-07-31T10:00:00Z",
    "status": "active"
  },
  {
    "issue_number": 124,
    "issue_title": "Another test issue with a very long title that might wrap",
    "repository_name": "another-repo",
    "branch": "issue-124-long-title-test",
    "tmux_session": "work-issue-124",
    "last_activity": "2025-07-31T09:30:00Z",
    "status": "active"
  }
]
```

#### 6.2 Expected Plain Text Output

**File:** `testdata/expected_plain_output.txt`

```
Issue   Title                             Repository    Branch                      Status  Last Activity
===================================================================================================
123     Test issue for TUI display        test-repo     issue-123-test-branch       active  2025-07-31
124     Another test issue with a very... another-repo  issue-124-long-title-test   active  2025-07-31
```

### 7. Performance and Load Tests

#### 7.1 TUI Launch Performance

```go
func TestPerformance_TUILaunchTime(t *testing.T)
func TestPerformance_SessionLoadingSpeed(t *testing.T)
func TestPerformance_PlainOutputGeneration(t *testing.T)
```

#### 7.2 Memory Usage Tests

```go
func TestMemory_TUIResourceUsage(t *testing.T)
func TestMemory_SessionDataHandling(t *testing.T)
```

### 8. Specific Test Scenarios

#### 8.1 Command Line Behavior Matrix

| Command | Expected Behavior | Test Function |
|---------|------------------|---------------|
| `sbs` | Launch TUI | `TestRootCommand_LaunchesTUI` |
| `sbs -h` | Show help | `TestRootCommand_ShowsHelp` |
| `sbs --help` | Show help | `TestRootCommand_ShowsHelp` |
| `sbs list` | Plain text output | `TestListCommand_PlainOutput` |
| `sbs list -p` | Plain text output | `TestListCommand_PlainFlag` |
| `sbs list --plain` | Plain text output | `TestListCommand_PlainFlagLong` |
| `sbs list -h` | Show list help | `TestListCommand_ShowsHelp` |

#### 8.2 Output Format Tests

```go
func TestOutputFormat_PlainTextColumns(t *testing.T)
func TestOutputFormat_EmptySessionsList(t *testing.T)
func TestOutputFormat_SingleSession(t *testing.T)
func TestOutputFormat_MultipleRepositories(t *testing.T)
func TestOutputFormat_TerminalWidthAdaptation(t *testing.T)
```

#### 8.3 User Experience Tests

```go
func TestUX_HelpTextClarity(t *testing.T)
func TestUX_ErrorMessageConsistency(t *testing.T)
func TestUX_CommandDiscoverability(t *testing.T)
```

### 9. Test Implementation Strategy

#### 9.1 Phase 1: Unit Tests (TDD Approach)
1. Write failing tests for new root command behavior
2. Write failing tests for modified list command behavior
3. Implement minimal changes to make tests pass
4. Refactor while keeping tests green

#### 9.2 Phase 2: Integration Tests
1. Add integration tests for command behavior
2. Test with mock TUI and session data
3. Verify backward compatibility

#### 9.3 Phase 3: End-to-End Tests
1. Add full process execution tests
2. Test in various terminal environments
3. Performance and memory testing

#### 9.4 Phase 4: Regression Testing
1. Run full test suite to ensure no breakage
2. Manual testing of all commands
3. Documentation updates verification

### 10. Test Execution

#### 10.1 Test Commands

```bash
# Unit tests
go test ./cmd/... -v

# Integration tests (with build tag)
go test -tags integration ./... -v

# All tests
make test

# Specific test coverage
go test -cover ./cmd/...

# Race condition detection
go test -race ./...
```

#### 10.2 CI/CD Integration

```yaml
# .github/workflows/test.yml additions
- name: Test command behavior changes
  run: |
    go test ./cmd/root_test.go -v
    go test ./cmd/list_test.go -v
    go test -tags integration ./integration_command_behavior_test.go -v
```

### 11. Success Criteria

#### 11.1 Test Coverage Goals
- Unit test coverage: 90%+ for modified files
- Integration test coverage: 80%+ for command behavior
- All edge cases covered with tests

#### 11.2 Performance Benchmarks
- TUI launch time: < 100ms for typical session count
- Plain text output generation: < 50ms for 100 sessions
- Memory usage increase: < 10% from baseline

#### 11.3 Quality Gates
- All existing tests continue to pass
- No regression in other command functionality
- Help text and documentation updated
- Backward compatibility maintained for `sbs list --plain`

### 12. Test Maintenance

#### 12.1 Documentation Updates
- Update test documentation for new behavior
- Add examples of new test patterns
- Document mock usage patterns

#### 12.2 Future Test Considerations
- Plan for additional TUI features
- Consider accessibility testing
- Plan for internationalization testing

## Implementation Notes

1. **Test-Driven Development**: Write tests first, then implement features
2. **Mock Strategy**: Use dependency injection for testable code
3. **Integration Testing**: Use build tags to separate unit and integration tests
4. **Backward Compatibility**: Maintain all existing test scenarios
5. **Performance**: Include performance tests for TUI launch and text output
6. **Error Handling**: Test all error conditions and edge cases

## Dependencies

- `github.com/stretchr/testify` for assertions and mocking
- `github.com/charmbracelet/bubbletea` for TUI testing
- Build tags for integration test separation
- Mock interfaces for dependency injection

This comprehensive testing plan ensures that the feature implementation is robust, maintainable, and preserves existing functionality while adding the new behavior as specified in GitHub Issue #15.