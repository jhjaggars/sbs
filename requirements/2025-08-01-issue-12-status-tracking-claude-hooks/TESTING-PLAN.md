# Testing Plan for Issue #12: Test recently added status tracking technique that uses Claude Code hooks

## Overview

This testing plan covers the comprehensive test strategy for implementing status tracking functionality that monitors `.sbs/stop.json` files in active session worktrees and displays session status with human-readable time deltas in the TUI.

## Feature Requirements Summary

**Status Tracking Implementation:**
- Monitor `.sbs/stop.json` file in active session worktrees
- Check for status files when TUI launches and every 60 seconds (configurable interval)
- Add new status column to list TUI showing current state
- Display human-readable time delta since status change (e.g., "stopped 5m ago")
- Integration with Claude Code hooks for status detection
- Minimal performance impact during status checks

**Expected Status States:**
- `active` - Session running, no stop.json file or recent activity
- `stopped` - stop.json file exists with recent timestamp
- `stale` - No tmux session, no recent activity
- `unknown` - Unable to determine status

## Test Structure and Organization

### 1. Unit Tests

#### 1.1 Status Detection Logic Tests (`pkg/status/detector_test.go` - NEW FILE)

**Core status detection functionality:**

```go
func TestStatusDetector_DetectSessionStatus(t *testing.T)
func TestStatusDetector_ParseStopJsonFile(t *testing.T)  
func TestStatusDetector_CalculateTimeDelta(t *testing.T)
func TestStatusDetector_HandleMissingStopFile(t *testing.T)
func TestStatusDetector_HandleCorruptedStopFile(t *testing.T)
func TestStatusDetector_HandlePermissionErrors(t *testing.T)
```

**Test Coverage:**
- Verify correct status detection based on stop.json file presence/absence
- Test parsing of various stop.json file formats (Claude Code hook outputs)
- Test time delta calculations for different durations
- Handle missing files gracefully (default to active/unknown status)
- Handle corrupted JSON files without crashing
- Handle file permission errors appropriately

#### 1.2 Time Formatting Tests (`pkg/status/time_formatter_test.go` - NEW FILE)

**Human-readable time formatting:**

```go
func TestTimeFormatter_FormatTimeDelta(t *testing.T)
func TestTimeFormatter_FormatNow(t *testing.T)
func TestTimeFormatter_FormatMinutes(t *testing.T)
func TestTimeFormatter_FormatHours(t *testing.T)
func TestTimeFormatter_FormatDays(t *testing.T)
func TestTimeFormatter_FormatWeeks(t *testing.T)
func TestTimeFormatter_HandleFutureTimestamps(t *testing.T)
```

**Test Coverage:**
- Test "now" for deltas < 1 minute
- Test "5m ago" format for minutes (1-59)
- Test "2h ago" format for hours (1-23)
- Test "3d ago" format for days (1-6)
- Test "2w ago" format for weeks (7+ days)
- Handle edge cases like future timestamps (clock skew)

#### 1.3 Configuration Tests (`pkg/config/status_config_test.go` - NEW FILE)

**Status tracking configuration:**

```go
func TestStatusConfig_DefaultRefreshInterval(t *testing.T)
func TestStatusConfig_ConfigurableRefreshInterval(t *testing.T)
func TestStatusConfig_ValidationRefreshInterval(t *testing.T)
func TestStatusConfig_StatusTrackingEnabled(t *testing.T)
func TestStatusConfig_LoadConfigurationWithStatusTracking(t *testing.T)
```

**Test Coverage:**
- Default refresh interval of 60 seconds
- Custom refresh intervals (5s, 30s, 120s, etc.)
- Validation of refresh intervals (minimum 5s, maximum 600s)
- Status tracking enable/disable toggle
- Configuration loading with status tracking settings

#### 1.4 TUI Status Display Tests (`pkg/tui/status_display_test.go` - NEW FILE)

**Status column rendering and formatting:**

```go
func TestStatusDisplay_StatusColumnRendering(t *testing.T)
func TestStatusDisplay_StatusColorFormatting(t *testing.T)
func TestStatusDisplay_StatusColumnWidth(t *testing.T)
func TestStatusDisplay_TimeDeltaDisplay(t *testing.T)
func TestStatusDisplay_StatusColumnResponsiveLayout(t *testing.T)
```

**Test Coverage:**
- Status column appears in both repository and global views
- Correct color coding for different status states
- Proper column width calculations including status
- Time delta appears next to status indicator
- Responsive layout adjustments with status column

### 2. Integration Tests

#### 2.1 Status Tracking Integration Tests (`integration_status_tracking_test.go` - NEW FILE)

**End-to-end status tracking workflow:**

```go
func TestE2E_StatusTrackingWorkflow(t *testing.T)
func TestE2E_StatusUpdateCycle(t *testing.T) 
func TestE2E_MultipleSessionStatusTracking(t *testing.T)
func TestE2E_StatusTrackingWithSandboxIntegration(t *testing.T)
func TestE2E_StatusTrackingPerformanceImpact(t *testing.T)
```

**Test Coverage:**
- Complete workflow from session start to status detection
- Status updates every 60 seconds in TUI
- Multiple sessions with different status states
- Integration with sandbox commands and Claude Code hooks
- Performance impact measurement during status checks

#### 2.2 TUI Integration Tests (`pkg/tui/status_tui_integration_test.go` - NEW FILE)

**TUI integration with status tracking:**

```go
func TestTUI_StatusTrackingInitialization(t *testing.T)
func TestTUI_StatusUpdateMessages(t *testing.T)
func TestTUI_StatusDisplayUpdates(t *testing.T)
func TestTUI_RefreshWithStatusTracking(t *testing.T)
func TestTUI_StatusTrackingErrorHandling(t *testing.T)
```

**Test Coverage:**
- TUI initializes status tracking on startup
- Status update messages trigger UI refreshes
- Status display updates correctly in real-time
- Manual refresh (r key) includes status checks
- Error handling when status tracking fails

#### 2.3 Claude Code Hooks Integration Tests (`integration_claude_hooks_test.go` - NEW FILE)

**Integration with Claude Code hooks:**

```go
func TestClaudeHooks_StopJsonCreation(t *testing.T)
func TestClaudeHooks_StopJsonParsing(t *testing.T)
func TestClaudeHooks_StatusDetectionFromHooks(t *testing.T)
func TestClaudeHooks_SandboxIntegration(t *testing.T)
```

**Test Coverage:**
- stop.json files created by Claude Code hooks
- Parsing various hook output formats
- Status detection from hook data timestamps
- Integration with sandbox file access

### 3. Performance Tests

#### 3.1 Status Checking Performance Tests (`pkg/status/performance_test.go` - NEW FILE)

**Performance impact measurement:**

```go
func BenchmarkStatusDetection_SingleSession(b *testing.B)
func BenchmarkStatusDetection_MultipleSessionsSerial(b *testing.B)
func BenchmarkStatusDetection_MultipleSessionsParallel(b *testing.B)
func BenchmarkStatusDetection_FileSystemAccess(b *testing.B)
func TestPerformance_StatusCheckingLatency(t *testing.T)
func TestPerformance_MemoryUsageImpact(t *testing.T)
```

**Test Coverage:**
- Single session status detection performance
- Multiple sessions processed serially vs parallel
- File system access overhead measurement
- Latency targets for status checking (< 100ms per session)
- Memory usage impact (< 5MB increase)

#### 3.2 TUI Refresh Performance Tests (`pkg/tui/refresh_performance_test.go` - NEW FILE)

**TUI refresh performance with status tracking:**

```go
func BenchmarkTUI_RefreshWithStatusTracking(b *testing.B)
func BenchmarkTUI_StatusColumnRendering(b *testing.B)
func TestPerformance_AutoRefreshOverhead(t *testing.T)
func TestPerformance_StatusTrackingDisabled(t *testing.T)
```

**Test Coverage:**
- TUI refresh performance with status tracking enabled
- Status column rendering performance
- Auto-refresh overhead measurement
- Performance comparison with status tracking disabled

### 4. Edge Cases and Error Conditions

#### 4.1 File System Edge Cases (`pkg/status/filesystem_edge_cases_test.go` - NEW FILE)

**File system error handling:**

```go
func TestEdgeCases_MissingWorktreeDirectory(t *testing.T)
func TestEdgeCases_MissingSbsDirectory(t *testing.T)
func TestEdgeCases_PermissionDeniedStopFile(t *testing.T)
func TestEdgeCases_CorruptedStopJsonFile(t *testing.T)
func TestEdgeCases_EmptyStopJsonFile(t *testing.T)
func TestEdgeCases_VeryLargeStopJsonFile(t *testing.T)
func TestEdgeCases_StopFileLockedByProcess(t *testing.T)
func TestEdgeCases_NetworkFileSystemLatency(t *testing.T)
```

**Test Coverage:**
- Missing worktree directories (sessions without worktrees)
- Missing .sbs directories within worktrees
- Permission denied when accessing stop.json files
- Corrupted or invalid JSON in stop.json files
- Empty stop.json files
- Abnormally large stop.json files (>1MB)
- Files locked by other processes
- Network file system access delays

#### 4.2 Timing Edge Cases (`pkg/status/timing_edge_cases_test.go` - NEW FILE)

**Time-related edge cases:**

```go
func TestEdgeCases_FutureTimestamps(t *testing.T)
func TestEdgeCases_ClockSkewHandling(t *testing.T)
func TestEdgeCases_TimezoneChanges(t *testing.T)
func TestEdgeCases_DaylightSavingTransitions(t *testing.T)
func TestEdgeCases_SystemClockAdjustment(t *testing.T)
func TestEdgeCases_VeryOldTimestamps(t *testing.T)
```

**Test Coverage:**
- Future timestamps in stop.json (clock skew)
- System clock adjustments during status tracking
- Timezone changes affecting time calculations
- Daylight saving time transitions
- Very old timestamps (months/years ago)

#### 4.3 Concurrent Access Edge Cases (`pkg/status/concurrency_test.go` - NEW FILE)

**Concurrent access scenarios:**

```go
func TestConcurrency_MultipleStatusChecksSimultaneous(t *testing.T)
func TestConcurrency_StatusUpdateDuringTUIRefresh(t *testing.T)
func TestConcurrency_FileModificationDuringRead(t *testing.T)
func TestConcurrency_RaceConditionPrevention(t *testing.T)
```

**Test Coverage:**
- Multiple status checks running simultaneously
- Status updates occurring during TUI refresh
- stop.json files modified while being read
- Race condition prevention in status tracking

### 5. Mock Requirements and Test Data

#### 5.1 Status Detection Mocks

```go
type MockStatusDetector struct {
    sessionStatuses map[string]StatusInfo
    detectError     error
    detectLatency   time.Duration
}

type StatusInfo struct {
    Status     string
    LastChange time.Time
    Error      error
}

func (m *MockStatusDetector) DetectStatus(sessionPath string) (StatusInfo, error)
func (m *MockStatusDetector) SetSessionStatus(sessionId string, status StatusInfo)
func (m *MockStatusDetector) SetDetectError(err error)
func (m *MockStatusDetector) SetDetectLatency(latency time.Duration)
```

#### 5.2 File System Mocks

```go
type MockFileSystem struct {
    files         map[string][]byte
    permissions   map[string]os.FileMode
    readLatency   time.Duration
    readError     error
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error)
func (m *MockFileSystem) Stat(path string) (os.FileInfo, error)
func (m *MockFileSystem) SetFile(path string, content []byte)
func (m *MockFileSystem) SetPermissions(path string, mode os.FileMode)
func (m *MockFileSystem) SetReadLatency(latency time.Duration)
func (m *MockFileSystem) SetReadError(err error)
```

#### 5.3 Time Mocks

```go
type MockTimeProvider struct {
    currentTime   time.Time
    timeAdvance   time.Duration
}

func (m *MockTimeProvider) Now() time.Time
func (m *MockTimeProvider) SetCurrentTime(t time.Time)
func (m *MockTimeProvider) AdvanceTime(d time.Duration)
```

### 6. Test Data and Fixtures

#### 6.1 Sample stop.json Files

**File:** `testdata/stop_json_samples/`

```json
// basic_stop.json - Basic Claude Code hook output
{
  "claude_code_hook": {
    "timestamp": "2025-08-01T12:30:45Z",
    "environment": "sandbox",
    "project_directory": "/work",
    "hook_script": "/home/user/claude-code-stop-hook.sh",
    "sandbox_detection": true
  },
  "hook_data": {
    "session_id": "work-issue-123",
    "tool_executions": 15,
    "last_tool": "Edit",
    "project_files_modified": 3
  }
}
```

```json
// minimal_stop.json - Minimal valid format
{
  "timestamp": "2025-08-01T10:15:30Z",
  "status": "stopped"
}
```

```json
// corrupted_stop.json - Invalid JSON for error testing
{
  "timestamp": "2025-08-01T10:15:30Z",
  "status": "stopped"
  // Missing closing brace - invalid JSON
```

#### 6.2 Test Session Configurations

**File:** `testdata/test_sessions.json`

```json
[
  {
    "issue_number": 123,
    "issue_title": "Test status tracking with active session",
    "repository_name": "test-repo",
    "worktree_path": "/tmp/test-worktrees/issue-123",
    "tmux_session": "work-issue-123",
    "status": "active",
    "expected_status_file": "testdata/stop_json_samples/basic_stop.json",
    "expected_status": "stopped",
    "expected_time_delta": "5m ago"
  },
  {
    "issue_number": 124,
    "issue_title": "Test status tracking with stale session",
    "repository_name": "test-repo",
    "worktree_path": "/tmp/test-worktrees/issue-124",
    "tmux_session": "work-issue-124",
    "status": "stale",
    "expected_status_file": null,
    "expected_status": "stale",
    "expected_time_delta": "2h ago"
  }
]
```

### 7. Specific Test Scenarios

#### 7.1 Status Detection Scenarios

| Scenario | stop.json Exists | Tmux Session | Expected Status | Expected Display |
|----------|------------------|--------------|-----------------|------------------|
| Active Session | No | Yes | active | ● now |
| Recently Stopped | Yes (5m ago) | No | stopped | ● stopped 5m ago |
| Stale Session | No | No | stale | ● stale 2h ago |
| Fresh Stop | Yes (30s ago) | No | stopped | ● stopped now |
| Old Stop | Yes (2d ago) | No | stopped | ● stopped 2d ago |

#### 7.2 TUI Display Scenarios

**Status Column Integration:**

```go
func TestStatusColumn_RepositoryView(t *testing.T)
func TestStatusColumn_GlobalView(t *testing.T) 
func TestStatusColumn_ResponsiveLayout(t *testing.T)
func TestStatusColumn_ColorCoding(t *testing.T)
func TestStatusColumn_TimeDeltaFormatting(t *testing.T)
```

**Expected Repository View Layout:**
```
Issue   Title                             Branch                      Status           Last Activity
============================================================================================
123     Fix authentication bug            issue-123-fix-auth-bug      ● stopped 5m ago 2025-08-01
124     Add dark mode support            issue-124-dark-mode         ● active now     2025-08-01
```

**Expected Global View Layout:**
```
Issue   Title                             Repository    Branch                      Status           Last Activity
================================================================================================================
123     Fix authentication bug            test-repo     issue-123-fix-auth-bug      ● stopped 5m ago 2025-08-01  
124     Add dark mode support            other-repo    issue-124-dark-mode         ● active now     2025-08-01
```

#### 7.3 Performance Test Scenarios

**Latency Requirements:**
- Single session status check: < 50ms
- 10 sessions status check: < 200ms
- 50 sessions status check: < 1000ms
- TUI refresh with status: < 500ms

**Memory Requirements:**
- Status tracking overhead: < 5MB
- Per-session status data: < 1KB

### 8. Auto-Refresh Testing

#### 8.1 Auto-Refresh Mechanism Tests (`pkg/tui/auto_refresh_test.go` - NEW FILE)

**Automatic status updates:**

```go
func TestAutoRefresh_StatusUpdateInterval(t *testing.T)
func TestAutoRefresh_ConfigurableInterval(t *testing.T)
func TestAutoRefresh_DisableAutoRefresh(t *testing.T)
func TestAutoRefresh_ManualRefreshPreservation(t *testing.T)
func TestAutoRefresh_ErrorHandlingDuringRefresh(t *testing.T)
```

**Test Coverage:**
- Status updates occur every 60 seconds by default
- Custom refresh intervals respect configuration
- Auto-refresh can be disabled
- Manual refresh (r key) still works when auto-refresh is disabled
- Errors during auto-refresh don't crash TUI

### 9. Configuration Integration Tests

#### 9.1 Status Configuration Tests (`pkg/config/status_integration_test.go` - NEW FILE)

**Configuration integration:**

```go
func TestStatusConfig_GlobalConfiguration(t *testing.T)
func TestStatusConfig_RepositoryOverrides(t *testing.T)
func TestStatusConfig_EnvironmentVariables(t *testing.T)
func TestStatusConfig_ConfigurationValidation(t *testing.T)
```

**Sample Configuration:**
```json
{
  "status_tracking": {
    "enabled": true,
    "refresh_interval_seconds": 60,
    "max_file_size_bytes": 1048576,
    "timeout_seconds": 5
  }
}
```

### 10. Backward Compatibility Tests

#### 10.1 Legacy Compatibility Tests (`backward_compatibility_status_test.go` - NEW FILE)

**Ensure existing functionality is preserved:**

```go
func TestBackwardCompatibility_ExistingTUIBehavior(t *testing.T)
func TestBackwardCompatibility_ExistingStatusDisplay(t *testing.T)
func TestBackwardCompatibility_ConfigurationUpgrade(t *testing.T)
func TestBackwardCompatibility_SessionMetadataFormat(t *testing.T)
```

**Test Coverage:**
- Existing TUI behavior unchanged when status tracking disabled
- Original status detection (tmux-based) still works
- Configuration files without status tracking settings still work
- Session metadata format is backward compatible

### 11. Test Implementation Strategy

#### 11.1 Phase 1: Core Status Detection (TDD Approach)
1. Write failing tests for status detection logic
2. Write failing tests for time formatting functions
3. Implement minimal status detector to make tests pass
4. Refactor while keeping tests green

#### 11.2 Phase 2: TUI Integration
1. Add failing tests for status column display
2. Add failing tests for status column width calculations
3. Implement TUI changes to display status
4. Add auto-refresh mechanism

#### 11.3 Phase 3: Configuration and Performance
1. Add configuration tests for status tracking settings
2. Add performance tests with benchmarks
3. Implement configuration integration
4. Optimize for performance requirements

#### 11.4 Phase 4: Edge Cases and Integration
1. Add comprehensive edge case tests
2. Add Claude Code hooks integration tests
3. Add end-to-end workflow tests
4. Manual testing of all scenarios

### 12. Test Execution Commands

#### 12.1 Unit Tests
```bash
# Core status detection tests
go test ./pkg/status/... -v

# TUI status display tests  
go test ./pkg/tui/... -v -run TestStatus

# Configuration tests
go test ./pkg/config/... -v -run TestStatus
```

#### 12.2 Integration Tests
```bash
# Integration tests (with build tag)
go test -tags integration ./integration_status_test.go -v

# End-to-end tests
go test -tags e2e ./e2e_status_test.go -v
```

#### 12.3 Performance Tests
```bash
# Performance benchmarks
go test -bench=BenchmarkStatus ./pkg/status/... -benchmem

# Performance validation tests
go test -tags performance ./pkg/status/performance_test.go -v
```

#### 12.4 All Tests
```bash
# Complete test suite
make test

# Test with coverage
go test -cover ./...

# Race condition detection
go test -race ./...
```

### 13. Success Criteria

#### 13.1 Functional Requirements
- [ ] Status tracking detects session states correctly
- [ ] TUI displays status column in both repository and global views
- [ ] Time deltas display in human-readable format (5m ago, 2h ago, etc.)
- [ ] Auto-refresh updates status every 60 seconds (configurable)
- [ ] Integration with Claude Code hooks works correctly
- [ ] Manual refresh (r key) includes status updates

#### 13.2 Performance Requirements
- [ ] Single session status check: < 50ms
- [ ] Multiple sessions status check: < 20ms per session
- [ ] TUI refresh with status tracking: < 500ms total
- [ ] Memory overhead: < 5MB for status tracking
- [ ] No noticeable impact on TUI responsiveness

#### 13.3 Quality Requirements
- [ ] Unit test coverage: 90%+ for new status tracking code
- [ ] Integration test coverage: 80%+ for status workflows
- [ ] All edge cases covered with tests
- [ ] Error handling prevents crashes
- [ ] Backward compatibility maintained

#### 13.4 Configuration Requirements
- [ ] Status tracking can be enabled/disabled
- [ ] Refresh interval is configurable (5-600 seconds)
- [ ] Configuration validation prevents invalid settings
- [ ] Repository-specific overrides work correctly

### 14. CI/CD Integration

#### 14.1 GitHub Actions Workflow
```yaml
# .github/workflows/status_tracking_tests.yml
name: Status Tracking Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run status tracking unit tests
        run: |
          go test ./pkg/status/... -v
          go test ./pkg/tui/... -v -run TestStatus
          go test ./pkg/config/... -v -run TestStatus

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y tmux
      - name: Run integration tests
        run: |
          go test -tags integration ./integration_status_test.go -v

  performance-tests:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run performance tests
        run: |
          go test -bench=BenchmarkStatus ./pkg/status/... -benchmem
          go test -tags performance ./pkg/status/performance_test.go -v
```

### 15. Test Maintenance and Documentation

#### 15.1 Test Documentation Updates
- Update test documentation for status tracking patterns
- Add examples of mock usage for status detection
- Document performance test expectations
- Add troubleshooting guide for test failures

#### 15.2 Future Test Considerations
- Plan for additional status tracking features (custom status types)
- Consider accessibility testing for status indicators
- Plan for performance testing with larger datasets (100+ sessions)
- Consider integration with external monitoring systems

## Implementation Notes

1. **Test-Driven Development**: Write comprehensive tests before implementing features
2. **Mock Strategy**: Use dependency injection for testable code, especially for file system access
3. **Performance Focus**: Include performance tests from the beginning, not as an afterthought
4. **Error Handling**: Test all error conditions and edge cases thoroughly
5. **Configuration Testing**: Ensure all configuration options are validated and tested
6. **Integration Testing**: Use build tags to separate unit and integration tests
7. **Documentation**: Keep test documentation up-to-date with implementation

## Dependencies

- `github.com/stretchr/testify` for assertions and mocking
- `github.com/charmbracelet/bubbletea` for TUI testing
- Build tags for test separation (`integration`, `e2e`, `performance`)
- Mock interfaces for dependency injection
- Test data fixtures for consistent testing scenarios

This comprehensive testing plan ensures that the status tracking feature implementation is robust, performant, maintainable, and integrates seamlessly with the existing SBS functionality while providing the enhanced status visibility required by Issue #12.