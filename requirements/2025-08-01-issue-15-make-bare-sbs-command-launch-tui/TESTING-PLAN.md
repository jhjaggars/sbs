# Testing Plan for Issue #15: Make bare `sbs` command launch TUI and `sbs list` show plain text output

## Overview

This testing plan covers the restructuring of the SBS command interface to make the TUI more accessible as the default interface while providing a quick plain text option for scripting/non-interactive use.

## Requirements Summary

- `sbs` (no subcommand) should launch the interactive TUI (current `sbs list` behavior)
- `sbs list` should output plain text format and exit (equivalent to current `sbs list --plain`)
- Maintain backward compatibility where possible
- Update help text and documentation to reflect new behavior

## Current vs Target Behavior

| Command | Current Behavior | Target Behavior |
|---------|-----------------|-----------------|
| `sbs` | Shows help/usage | Launch interactive TUI |
| `sbs list` | Launch interactive TUI | Show plain text output |
| `sbs list --plain` | Show plain text output | Show plain text output (backward compatibility) |

## Test Strategy

### Phase 1: Unit Tests (Test-First Approach)

#### 1.1 Root Command Tests (`cmd/root_test.go`)

**Test Cases:**
- `TestRootCommand_LaunchesTUI`: Verify bare `sbs` command launches TUI
- `TestRootCommand_WithArgs_ShowsHelp`: Verify invalid args still show help
- `TestRootCommand_HelpText_Updated`: Verify help text reflects new behavior

**Implementation Approach:**
```go
func TestRootCommand_LaunchesTUI(t *testing.T) {
    // Mock TUI model and program
    // Execute root command with no args
    // Verify TUI program is created and run
}
```

#### 1.2 List Command Tests (`cmd/list_test.go` - NEW FILE)

**Test Cases:**
- `TestListCommand_DefaultPlainOutput`: Verify `sbs list` shows plain text
- `TestListCommand_PlainFlag_BackwardCompatibility`: Verify `--plain` still works
- `TestListCommand_PlainOutput_Format`: Verify output format is correct
- `TestListCommand_NoSessions_Message`: Verify "no sessions" message
- `TestListCommand_MultipleSessions_Formatting`: Verify proper session formatting
- `TestListCommand_TerminalWidth_Responsive`: Verify responsive column widths

**Mock Requirements:**
- Session loader mock for consistent test data
- Terminal width mock for testing responsive behavior

#### 1.3 Command Structure Tests

**Test Cases:**
- `TestCommandHierarchy_Unchanged`: Verify other commands unaffected
- `TestGlobalFlags_Preserved`: Verify global flags still work
- `TestCommandHelp_Updated`: Verify help text is accurate

### Phase 2: Integration Tests

#### 2.1 Command Execution Tests (`cmd/command_integration_test.go` - NEW FILE)

**Test Cases:**
- `TestIntegration_RootCommand_TUILaunch`: End-to-end TUI launch
- `TestIntegration_ListCommand_PlainOutput`: End-to-end plain text output
- `TestIntegration_BackwardCompatibility`: Verify existing workflows work

**Test Environment:**
- Use test sessions for consistent data
- Mock terminal environment as needed
- Test with and without active sessions

#### 2.2 TUI Integration Tests

**Test Cases:**
- `TestTUI_LaunchedFromRoot`: Verify TUI model creation from root command
- `TestTUI_Navigation_FromRoot`: Verify TUI navigation works correctly
- `TestTUI_Exit_Behavior`: Verify proper exit handling

### Phase 3: Backward Compatibility Tests

#### 3.1 Compatibility Verification (`regression_test.go` - NEW FILE)

**Test Cases:**
- `TestBackwardCompatibility_ListPlainFlag`: Verify `--plain` flag preserved
- `TestBackwardCompatibility_OtherCommands`: Verify start/stop/attach unchanged
- `TestBackwardCompatibility_GlobalFlags`: Verify global flags work
- `TestBackwardCompatibility_ExistingScripts`: Test scenarios that existing scripts might use

#### 3.2 Migration Path Tests

**Test Cases:**
- `TestMigration_ExistingWorkflows`: Verify common user workflows still work
- `TestMigration_ScriptingUsage`: Verify scripting usage patterns work

### Phase 4: Edge Cases and Error Handling

#### 4.1 Error Condition Tests

**Test Cases:**
- `TestErrorHandling_NoTerminal`: Behavior when terminal detection fails
- `TestErrorHandling_InvalidSessions`: Behavior with corrupted session data
- `TestErrorHandling_MissingDependencies`: Behavior when tools unavailable

#### 4.2 Environment Tests

**Test Cases:**
- `TestEnvironment_NonInteractiveTerminal`: Behavior in CI/automated environments
- `TestEnvironment_DifferentTerminalSizes`: Responsive behavior across sizes
- `TestEnvironment_ColorSupport`: Behavior with/without color support

### Phase 5: Performance and Usability Tests

#### 5.1 Performance Tests

**Test Cases:**
- `TestPerformance_LargeSessions`: Performance with many sessions
- `TestPerformance_StartupTime`: Command startup time benchmarks
- `TestPerformance_TUIResponsiveness`: TUI responsiveness benchmarks

#### 5.2 Usability Tests

**Test Cases:**
- `TestUsability_HelpDiscoverability`: Verify users can discover new behavior
- `TestUsability_ErrorMessages`: Verify clear error messages
- `TestUsability_DefaultBehavior`: Verify default behavior is intuitive

## Test Implementation Order

### Day 1: Foundation Tests
1. Create `cmd/list_test.go` with basic list command tests
2. Update `cmd/root_test.go` with root command TUI launch tests
3. Implement mocks for session loading and TUI components

### Day 2: Integration Tests
1. Create `cmd/command_integration_test.go`
2. Implement end-to-end command execution tests
3. Test TUI integration from root command

### Day 3: Compatibility and Edge Cases
1. Create `regression_test.go` for backward compatibility
2. Implement error handling and edge case tests
3. Add performance benchmarks

## Mock Strategy

### Required Mocks

1. **TUI Program Mock**
   ```go
   type MockTUIProgram struct {
       RunCalled bool
       Model     tea.Model
   }
   ```

2. **Session Loader Mock**
   ```go
   type MockSessionLoader struct {
       Sessions []config.SessionMetadata
       Error    error
   }
   ```

3. **Terminal Environment Mock**
   ```go
   type MockTerminal struct {
       Width       int
       IsInteractive bool
   }
   ```

## Acceptance Criteria Validation

### Automated Test Coverage

- [ ] `sbs` launches interactive TUI → `TestRootCommand_LaunchesTUI`
- [ ] `sbs list` outputs plain text → `TestListCommand_DefaultPlainOutput`
- [ ] `sbs list --plain` still works → `TestListCommand_PlainFlag_BackwardCompatibility`
- [ ] Help text updated → `TestCommandHelp_Updated`
- [ ] No breaking changes → `TestBackwardCompatibility_*` suite
- [ ] Documentation updated → Manual verification

### Quality Gates

- **Unit Test Coverage**: 90%+ for modified files
- **Integration Test Coverage**: All command combinations tested
- **Performance**: No regression in startup time or memory usage
- **Backward Compatibility**: All existing test cases pass

## Test Data Requirements

### Session Test Data
```go
var testSessions = []config.SessionMetadata{
    {
        IssueNumber:    123,
        IssueTitle:     "Test Issue",
        RepositoryName: "test-repo",
        Branch:         "issue-123-test",
        Status:         "active",
        LastActivity:   "2025-08-01T10:00:00Z",
    },
    // Additional test sessions...
}
```

### Terminal Environment Test Data
- Various terminal widths (80, 120, 200+ columns)
- Interactive vs non-interactive environments
- Color vs no-color support scenarios

## Success Metrics

1. **All Tests Pass**: 100% test pass rate
2. **Coverage Target**: 90%+ coverage on modified code
3. **Performance**: No degradation in command execution time
4. **Compatibility**: All existing workflows continue to work
5. **User Experience**: Clear, intuitive default behavior

## Risk Mitigation

### High-Risk Areas
1. **TUI Integration**: Risk of breaking existing TUI functionality
   - Mitigation: Comprehensive TUI integration tests
2. **Backward Compatibility**: Risk of breaking existing scripts
   - Mitigation: Extensive compatibility test suite
3. **Terminal Detection**: Risk of incorrect behavior in different environments
   - Mitigation: Environment-specific test cases

### Testing in CI/CD
- Run tests in various terminal environments
- Test with different terminal sizes and capabilities
- Verify behavior in both interactive and non-interactive modes

This comprehensive testing plan ensures robust implementation of the command interface restructuring while maintaining backward compatibility and code quality.