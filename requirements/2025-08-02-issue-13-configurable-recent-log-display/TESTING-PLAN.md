# Testing Plan: Add Configurable Recent Log Display with Project-Specific Loghook Scripts

## Overview

This testing plan covers comprehensive validation of GitHub issue #13: implementing configurable recent log display with project-specific loghook scripts. The feature adds a new 'l' key binding to the TUI that opens a dedicated log view, executes `.sbs/loghook` scripts from session worktrees, and displays the output with configurable auto-refresh intervals.

## Test Scope

### In Scope
- Unit tests for new log view functionality
- Integration tests for TUI log view interactions
- Script execution and output display validation
- Auto-refresh mechanism testing
- Configuration parameter validation
- Error handling for missing/failed scripts
- Key binding integration with existing TUI
- Log view state management and lifecycle

### Out of Scope
- Content validation of specific loghook script implementations
- Performance testing of long-running scripts
- Cross-platform script execution differences
- Network-dependent log sources (test with mock data)

## Architecture Understanding

Based on codebase analysis, the implementation will require:

1. **TUI Architecture**: Uses Bubble Tea framework with key-based message handling
2. **View States**: Current ViewMode enum manages Repository/Global views
3. **Key Bindings**: Defined in keyMap struct with corresponding handlers
4. **Session Management**: Uses config.SessionMetadata for worktree paths
5. **File System Access**: Direct file system operations for worktree access
6. **Auto-refresh**: Existing tickAutoRefresh pattern for status updates

## Test Strategy

Following TDD (Test Driven Development) practices:
1. **Red Phase**: Tests fail with current implementation
2. **Green Phase**: Tests pass after implementing the changes  
3. **Refactor Phase**: Code and tests are clean and maintainable

## Test Categories

### 1. Unit Tests

#### 1.1 Key Binding Tests
**File**: `pkg/tui/model_test.go` (extend existing)

**Test Cases**:

```go
func TestModel_LogViewKeyBinding(t *testing.T) {
    t.Run("l_key_binding_triggers_log_view", func(t *testing.T) {
        // Test that 'l' key creates log view message
        // Verify key binding is properly registered
        // Test with active sessions
    })
    
    t.Run("l_key_binding_help_text", func(t *testing.T) {
        // Test that help text includes 'l: logs' option
        // Verify placement in help text order
        // Test both condensed and full help views
    })
    
    t.Run("l_key_binding_disabled_when_no_sessions", func(t *testing.T) {
        // Test behavior when no sessions are available
        // Verify appropriate error handling or no-op behavior
    })
    
    t.Run("l_key_binding_disabled_when_no_current_session", func(t *testing.T) {
        // Test when sessions exist but none is selected
        // Verify graceful handling
    })
}
```

#### 1.2 Log View State Management Tests
**File**: `pkg/tui/log_view_test.go` (new file)

**Test Cases**:

```go
func TestLogView_StateManagement(t *testing.T) {
    t.Run("log_view_creation", func(t *testing.T) {
        // Test LogView struct initialization
        // Verify default values and configuration
    })
    
    t.Run("log_view_mode_toggle", func(t *testing.T) {
        // Test switching between list view and log view
        // Verify state preservation
    })
    
    t.Run("log_view_exit_on_escape", func(t *testing.T) {
        // Test ESC key exits log view
        // Returns to previous view state
    })
    
    t.Run("log_view_exit_on_q", func(t *testing.T) {
        // Test 'q' key exits log view
        // Returns to list view
    })
    
    t.Run("log_view_scroll_functionality", func(t *testing.T) {
        // Test up/down arrow scrolling
        // Test page up/down scrolling
        // Test scroll bounds handling
    })
}
```

#### 1.3 Loghook Script Execution Tests
**File**: `pkg/tui/loghook_test.go` (new file)

**Test Cases**:

```go
func TestLoghook_ScriptExecution(t *testing.T) {
    t.Run("execute_existing_loghook_script", func(t *testing.T) {
        // Create test worktree with .sbs/loghook script
        // Execute script and capture output
        // Verify output is returned correctly
    })
    
    t.Run("handle_missing_loghook_script", func(t *testing.T) {
        // Test behavior when .sbs/loghook doesn't exist
        // Verify appropriate error message or fallback
    })
    
    t.Run("handle_non_executable_loghook_script", func(t *testing.T) {
        // Test with loghook file that lacks execute permissions
        // Verify error handling and user feedback
    })
    
    t.Run("handle_script_execution_failure", func(t *testing.T) {
        // Test with script that returns non-zero exit code
        // Verify error capture and display
    })
    
    t.Run("handle_script_execution_timeout", func(t *testing.T) {
        // Test with long-running script
        // Verify timeout handling (if implemented)
    })
    
    t.Run("script_working_directory", func(t *testing.T) {
        // Verify script executes from correct working directory
        // Test that script has access to worktree context
    })
}
```

#### 1.4 Auto-Refresh Mechanism Tests
**File**: `pkg/tui/log_refresh_test.go` (new file)

**Test Cases**:

```go
func TestLogView_AutoRefresh(t *testing.T) {
    t.Run("auto_refresh_enabled_when_log_view_active", func(t *testing.T) {
        // Test that auto-refresh starts when entering log view
        // Verify refresh interval configuration
    })
    
    t.Run("auto_refresh_disabled_when_log_view_inactive", func(t *testing.T) {
        // Test that auto-refresh stops when exiting log view
        // Verify no background refresh activity
    })
    
    t.Run("configurable_refresh_interval", func(t *testing.T) {
        // Test default 5-second interval
        // Test custom intervals from configuration
        // Test minimum/maximum interval bounds
    })
    
    t.Run("manual_refresh_with_r_key", func(t *testing.T) {
        // Test 'r' key triggers immediate refresh
        // Test refresh indication to user
    })
    
    t.Run("refresh_error_handling", func(t *testing.T) {
        // Test behavior when refresh fails
        // Verify error display to user
        // Test retry behavior
    })
}
```

#### 1.5 Configuration Tests
**File**: `pkg/config/log_config_test.go` (new file)

**Test Cases**:

```go
func TestConfig_LogDisplay(t *testing.T) {
    t.Run("default_log_refresh_interval", func(t *testing.T) {
        // Test default 5-second refresh interval
        // Verify reasonable default values
    })
    
    t.Run("custom_log_refresh_interval", func(t *testing.T) {
        // Test setting custom refresh intervals
        // Test validation of interval bounds (1-300 seconds)
    })
    
    t.Run("log_display_configuration_loading", func(t *testing.T) {
        // Test loading log display config from files
        // Test both global and repository-specific config
    })
    
    t.Run("log_display_configuration_merging", func(t *testing.T) {
        // Test repository config overrides global config
        // Test precedence handling
    })
}
```

### 2. Integration Tests

#### 2.1 TUI Integration Tests
**File**: `pkg/tui/log_integration_test.go` (new file)

**Test Cases**:

```go
func TestLogView_TUIIntegration(t *testing.T) {
    t.Run("full_workflow_list_to_log_to_list", func(t *testing.T) {
        // Test complete user workflow
        // Start in list view, press 'l', view logs, press ESC
        // Verify state transitions and data preservation
    })
    
    t.Run("log_view_with_multiple_sessions", func(t *testing.T) {
        // Test log view behavior with multiple active sessions
        // Verify correct session selection and script execution
    })
    
    t.Run("log_view_terminal_resize_handling", func(t *testing.T) {
        // Test log display adjusts to terminal size changes
        // Verify text wrapping and layout adaptation
    })
    
    t.Run("log_view_concurrent_with_background_refresh", func(t *testing.T) {
        // Test log view behavior during background session updates
        // Verify no conflicts between different refresh mechanisms
    })
}
```

#### 2.2 Session Integration Tests
**File**: `pkg/tui/log_session_integration_test.go` (new file)

**Test Cases**:

```go
func TestLogView_SessionIntegration(t *testing.T) {
    t.Run("log_view_for_repository_mode_sessions", func(t *testing.T) {
        // Test log view with repository-scoped sessions
        // Verify correct worktree path resolution
    })
    
    t.Run("log_view_for_global_mode_sessions", func(t *testing.T) {
        // Test log view with global session list
        // Verify cross-repository session handling
    })
    
    t.Run("log_view_session_selection_persistence", func(t *testing.T) {
        // Test that selected session remains selected after log view
        // Verify cursor position preservation
    })
    
    t.Run("log_view_with_stale_sessions", func(t *testing.T) {
        // Test log view behavior with inactive/stale sessions
        // Verify appropriate error handling
    })
}
```

### 3. Error Handling and Edge Case Tests

#### 3.1 Error Scenario Tests
**File**: `pkg/tui/log_error_test.go` (new file)

**Test Cases**:

```go
func TestLogView_ErrorHandling(t *testing.T) {
    t.Run("missing_worktree_directory", func(t *testing.T) {
        // Test when session worktree directory doesn't exist
        // Verify appropriate error message display
    })
    
    t.Run("missing_sbs_directory", func(t *testing.T) {
        // Test when .sbs directory doesn't exist in worktree
        // Verify fallback behavior or error handling
    })
    
    t.Run("permission_denied_on_script", func(t *testing.T) {
        // Test when user lacks permission to execute script
        // Verify error message and graceful handling
    })
    
    t.Run("script_output_too_large", func(t *testing.T) {
        // Test with script that outputs very large amounts of data
        // Verify memory handling and display truncation
    })
    
    t.Run("concurrent_script_execution", func(t *testing.T) {
        // Test behavior when multiple refresh attempts occur
        // Verify proper synchronization and no race conditions
    })
}
```

#### 3.2 Edge Case Tests
**File**: `pkg/tui/log_edge_case_test.go` (new file)

**Test Cases**:

```go
func TestLogView_EdgeCases(t *testing.T) {
    t.Run("empty_script_output", func(t *testing.T) {
        // Test with script that produces no output
        // Verify appropriate display message
    })
    
    t.Run("script_with_ansi_colors", func(t *testing.T) {
        // Test script output containing ANSI color codes
        // Verify proper handling/stripping if needed
    })
    
    t.Run("script_with_unicode_characters", func(t *testing.T) {
        // Test script output with Unicode/UTF-8 content
        // Verify proper display and terminal compatibility
    })
    
    t.Run("very_narrow_terminal", func(t *testing.T) {
        // Test log view display in very narrow terminal (40 chars)
        // Verify text wrapping and layout handling
    })
    
    t.Run("rapid_key_presses", func(t *testing.T) {
        // Test rapid 'l' key presses or view switching
        // Verify no crashes or state corruption
    })
}
```

### 4. Performance and Resource Tests

#### 4.1 Resource Management Tests
**File**: `pkg/tui/log_performance_test.go` (new file)

**Test Cases**:

```go
func TestLogView_Performance(t *testing.T) {
    t.Run("memory_usage_with_large_output", func(t *testing.T) {
        // Test memory consumption with large script output
        // Verify no memory leaks with repeated executions
    })
    
    t.Run("script_process_cleanup", func(t *testing.T) {
        // Verify script processes are properly cleaned up
        // Test for orphaned processes after view exits
    })
    
    t.Run("auto_refresh_resource_usage", func(t *testing.T) {
        // Test resource usage with auto-refresh enabled
        // Verify reasonable CPU and memory consumption
    })
    
    t.Run("concurrent_log_views", func(t *testing.T) {
        // Test behavior with multiple SBS instances
        // Verify no conflicts or resource contention
    })
}
```

### 5. Configuration Integration Tests

#### 5.1 Configuration Tests
**File**: `pkg/config/log_display_integration_test.go` (new file)

**Test Cases**:

```go
func TestConfig_LogDisplayIntegration(t *testing.T) {
    t.Run("global_config_log_display_settings", func(t *testing.T) {
        // Test loading log display settings from global config
        // Verify default values and custom overrides
    })
    
    t.Run("repository_specific_log_config", func(t *testing.T) {
        // Test repository-specific .sbs/config.json overrides
        // Verify per-project log display customization
    })
    
    t.Run("invalid_config_values_handling", func(t *testing.T) {
        // Test with invalid refresh intervals or settings
        // Verify fallback to defaults and error reporting
    })
    
    t.Run("config_changes_during_runtime", func(t *testing.T) {
        // Test behavior when config files change during execution
        // Verify if changes are picked up (if supported)
    })
}
```

## Test Data and Fixtures

### Mock Session Data
```go
var testLogSessions = []config.SessionMetadata{
    {
        IssueNumber:    123,
        IssueTitle:     "Fix authentication bug",
        RepositoryName: "test-repo",
        Branch:         "issue-123-fix-auth-bug",
        WorktreePath:   "/tmp/test-worktrees/issue-123",
        TmuxSession:    "work-issue-123",
        SandboxName:    "work-issue-test-repo-123",
        LastActivity:   "2025-08-02T10:00:00Z",
    },
    {
        IssueNumber:    124,
        IssueTitle:     "Add dark mode support",
        RepositoryName: "test-repo",
        Branch:         "issue-124-dark-mode",
        WorktreePath:   "/tmp/test-worktrees/issue-124",
        TmuxSession:    "work-issue-124",
        SandboxName:    "work-issue-test-repo-124",
        LastActivity:   "2025-08-02T09:30:00Z",
    },
}
```

### Test Loghook Scripts
```bash
# Basic successful script
#!/bin/bash
echo "Recent activity:"
echo "$(date): Script executed successfully"
git log --oneline -5

# Script with error
#!/bin/bash
echo "Error: Something went wrong" >&2
exit 1

# Long-running script
#!/bin/bash
echo "Starting long operation..."
sleep 10
echo "Operation completed"

# Large output script
#!/bin/bash
for i in {1..1000}; do
    echo "Log line $i: Some log information here"
done
```

### Test Directory Structure Setup
```go
func setupTestWorktree(t *testing.T) string {
    worktreePath := filepath.Join(t.TempDir(), "test-worktree")
    sbsDir := filepath.Join(worktreePath, ".sbs")
    
    require.NoError(t, os.MkdirAll(sbsDir, 0755))
    
    // Create test loghook script
    loghookPath := filepath.Join(sbsDir, "loghook")
    script := `#!/bin/bash
echo "Test log output"
echo "Timestamp: $(date)"
`
    require.NoError(t, os.WriteFile(loghookPath, []byte(script), 0755))
    
    return worktreePath
}
```

## Test Implementation Strategy

### Phase 1: Core Infrastructure Tests (Day 1)
1. Create log view state management tests
2. Implement key binding tests
3. Set up test utilities and mock objects
4. Ensure all tests fail with current implementation (Red phase)

### Phase 2: Script Execution Tests (Day 1-2)
1. Implement loghook script execution tests
2. Create error handling tests for script failures
3. Add edge case testing for various script scenarios

### Phase 3: TUI Integration Tests (Day 2)
1. Implement view transition tests
2. Create auto-refresh mechanism tests
3. Add terminal handling and display tests

### Phase 4: Configuration and Session Tests (Day 2-3)
1. Implement configuration loading and validation tests
2. Create session integration tests
3. Add repository vs global mode testing

### Phase 5: Performance and Edge Cases (Day 3)
1. Implement resource management tests
2. Create comprehensive edge case coverage
3. Add performance and memory usage tests

## Expected Test Results

### Before Implementation (Red Phase)
- All new log view tests should fail (feature doesn't exist)
- Key binding tests should fail (no 'l' key handler)
- Script execution tests should fail (no execution logic)
- Configuration tests should fail (no log display config)
- Existing TUI tests should pass (no regressions)

### After Implementation (Green Phase)
- All log view functionality tests should pass
- All script execution tests should pass
- All integration tests should pass
- All configuration tests should pass
- All existing tests should continue to pass

## Test Execution Commands

```bash
# Run all log view tests
go test ./pkg/tui/ -v -run TestLogView

# Run specific test categories
go test ./pkg/tui/ -v -run TestLogView_StateManagement
go test ./pkg/tui/ -v -run TestLoghook_ScriptExecution
go test ./pkg/config/ -v -run TestConfig_LogDisplay

# Run integration tests
go test ./pkg/tui/ -v -run TestLogView_.*Integration

# Run with coverage
go test ./pkg/tui/ -v -coverprofile=coverage.out -run TestLogView
go tool cover -html=coverage.out -o coverage.html

# Run all tests to ensure no regressions
go test ./... -v
make test
```

## Success Criteria

1. **100% Test Coverage**: All new log view code paths have test coverage
2. **All Tests Pass**: Complete test suite passes after implementation
3. **No Regressions**: All existing TUI functionality continues to work
4. **Error Handling**: Graceful handling of all error scenarios
5. **Performance**: Acceptable memory and CPU usage with auto-refresh
6. **Configuration**: Proper integration with existing config system
7. **User Experience**: Intuitive key bindings and responsive interface

## Risk Mitigation

### High-Risk Areas
- Script execution security and process management
- Auto-refresh resource usage and performance impact
- Terminal compatibility and display handling
- State management between different view modes

### Medium-Risk Areas
- Configuration loading and validation
- Session selection and worktree path resolution
- Error message display and user feedback

### Low-Risk Areas
- Key binding registration (well-established pattern)
- Basic text display functionality
- Help text integration

### Mitigation Strategies
- Comprehensive error handling tests for all script execution scenarios
- Resource usage monitoring tests to prevent memory leaks
- Extensive terminal compatibility testing across different sizes
- Clear separation of concerns between view state and business logic
- Security-focused script execution with proper process cleanup
- Fallback mechanisms for missing or failing scripts

## Maintenance and Future Considerations

1. **Test Maintenance**: Update tests when adding new log display features
2. **Script Security**: Consider sandboxing or security restrictions for loghook scripts
3. **Performance Monitoring**: Add metrics for script execution times and resource usage
4. **Documentation**: Keep test documentation updated with new patterns
5. **Configuration Evolution**: Design tests to accommodate future config options
6. **Cross-Platform**: Ensure test compatibility across different operating systems
7. **Accessibility**: Consider future accessibility requirements for log display

## Dependencies and Prerequisites

### Required for Testing
- Go testing framework and testify package
- Mock file system utilities for test fixtures
- Process execution mocking for script tests
- Terminal simulation for TUI testing
- Configuration loading test utilities

### Test Environment Setup
- Temporary directory creation for test worktrees
- Mock script creation with various scenarios
- Test configuration file generation
- Session metadata mock data preparation
- Terminal size simulation capabilities

This comprehensive testing plan ensures robust implementation of the configurable recent log display feature while maintaining high code quality and preventing regressions in the existing TUI functionality.