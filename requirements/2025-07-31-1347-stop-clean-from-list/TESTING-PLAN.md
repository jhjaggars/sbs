# Testing Plan: Stop and Clean Sessions from List Interface

## Overview

This testing plan implements a Test-Driven Development (TDD) approach for adding stop ('s' key) and clean ('c' key) session management functionality to the interactive TUI list interface. The plan covers unit tests, integration tests, edge cases, and user interaction workflows.

## Test Structure & Organization

### Test Files
```
pkg/tui/
├── model_test.go           # Extended for new help text tests
├── stop_clean_test.go      # New - core stop/clean functionality  
├── modal_test.go           # New - modal dialog testing
└── integration_test.go     # New - end-to-end workflows
```

### Test Categories

1. **Unit Tests** - Core functionality with mocking
2. **Integration Tests** - TUI interactions and message flow
3. **Edge Case Tests** - Error scenarios and boundary conditions
4. **View Rendering Tests** - Help text and modal display

## TDD Implementation Approach

### Phase 1: Red - Write Failing Tests

#### 1.1 Key Binding Tests
```go
func TestStopCleanKeyBindings(t *testing.T) {
    // Test 's' key binding exists and configured correctly
    // Test 'c' key binding exists and configured correctly
    // Test help text includes new shortcuts
}
```

#### 1.2 Message Type Tests
```go
func TestNewMessageTypes(t *testing.T) {
    // Test stopSessionMsg struct creation
    // Test cleanSessionsMsg struct creation  
    // Test confirmationDialogMsg struct creation
}
```

#### 1.3 Modal Dialog Tests
```go
func TestModalDialogState(t *testing.T) {
    // Test modal dialog visibility toggle
    // Test confirmation message content
    // Test pending clean sessions storage
}
```

### Phase 2: Green - Implement Minimal Functionality

#### 2.1 Add Key Bindings
- Extend keyMap struct with Stop and Clean bindings
- Add to keys variable following existing pattern
- Update help text generation

#### 2.2 Add Message Types
- Implement stopSessionMsg with error handling
- Implement cleanSessionsMsg with results
- Implement confirmationDialogMsg for modal state

#### 2.3 Add Model State
- Add modal dialog state fields to Model struct
- Implement state management methods

### Phase 3: Refactor - Extract and Optimize

#### 3.1 Extract Reusable Logic
- Extract stop logic from cmd/stop.go
- Extract clean logic from cmd/clean.go
- Create shared utility functions

#### 3.2 Optimize Performance
- Efficient session filtering for clean operations
- Minimal UI redraws during operations

## Unit Tests

### Core Stop Functionality
```go
func TestStopSelectedSession(t *testing.T) {
    tests := []struct {
        name           string
        selectedIndex  int
        sessions       []config.SessionMetadata  
        tmuxError      error
        sandboxError   error
        expectedError  error
    }{
        {
            name: "successful stop",
            selectedIndex: 0,
            sessions: []config.SessionMetadata{{IssueNumber: 123}},
            tmuxError: nil,
            sandboxError: nil,
            expectedError: nil,
        },
        {
            name: "tmux kill fails",
            selectedIndex: 0, 
            sessions: []config.SessionMetadata{{IssueNumber: 123}},
            tmuxError: errors.New("tmux session not found"),
            sandboxError: nil,
            expectedError: errors.New("tmux session not found"),
        },
        {
            name: "sandbox cleanup fails",
            selectedIndex: 0,
            sessions: []config.SessionMetadata{{IssueNumber: 123}},
            tmuxError: nil,
            sandboxError: errors.New("sandbox removal failed"),
            expectedError: errors.New("sandbox removal failed"),
        },
        {
            name: "no session selected",
            selectedIndex: -1,
            sessions: []config.SessionMetadata{},
            expectedError: errors.New("no session selected"),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockTmux := &MockTmuxManager{killError: tt.tmuxError}
            mockSandbox := &MockSandboxManager{removeError: tt.sandboxError}
            
            // Test implementation
            model := setupTestModel(mockTmux, mockSandbox)
            model.sessions = tt.sessions
            model.selectedIndex = tt.selectedIndex
            
            cmd := model.stopSelectedSession()
            msg := executeCommand(cmd)
            
            if tt.expectedError == nil {
                assert.NoError(t, msg.(stopSessionMsg).err)
            } else {
                assert.Equal(t, tt.expectedError.Error(), msg.(stopSessionMsg).err.Error())
            }
        })
    }
}
```

### Core Clean Functionality
```go
func TestCleanStaleSessions(t *testing.T) {
    tests := []struct{
        name              string
        viewMode          ViewMode
        sessions          []config.SessionMetadata
        staleSessions     []string
        expectedCleaned   int
        expectedError     error
    }{
        {
            name: "clean stale sessions in repository view",
            viewMode: RepositoryView,
            sessions: []config.SessionMetadata{
                {IssueNumber: 123, TmuxSession: "work-issue-123"},
                {IssueNumber: 124, TmuxSession: "work-issue-124"},
            },
            staleSessions: []string{"work-issue-124"},
            expectedCleaned: 1,
            expectedError: nil,
        },
        {
            name: "no stale sessions found",
            viewMode: GlobalView,
            sessions: []config.SessionMetadata{
                {IssueNumber: 123, TmuxSession: "work-issue-123"},
            },
            staleSessions: []string{},
            expectedCleaned: 0,
            expectedError: nil,
        },
        {
            name: "cleanup fails for some sessions",
            viewMode: GlobalView,
            sessions: []config.SessionMetadata{
                {IssueNumber: 123, TmuxSession: "work-issue-123"},
                {IssueNumber: 124, TmuxSession: "work-issue-124"},
            },
            staleSessions: []string{"work-issue-123", "work-issue-124"},
            expectedCleaned: 1,
            expectedError: errors.New("failed to clean some sessions"),
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test clean functionality
            mockTmux := &MockTmuxManager{staleSessions: tt.staleSessions}
            model := setupTestModel(mockTmux, nil)
            model.viewMode = tt.viewMode
            model.sessions = tt.sessions
            
            result := model.identifyAndCleanStaleSessions()
            
            assert.Equal(t, tt.expectedCleaned, len(result.cleanedSessions))
            if tt.expectedError != nil {
                assert.Contains(t, result.err.Error(), tt.expectedError.Error())
            }
        })
    }
}
```

### Modal Dialog Tests
```go
func TestModalDialogInteractions(t *testing.T) {
    tests := []struct{
        name              string
        key               string
        initialState      bool
        expectedState     bool
        expectedAction    string
    }{
        {
            name: "show confirmation dialog",
            key: "c",
            initialState: false,
            expectedState: true,
            expectedAction: "show_dialog",
        },
        {
            name: "confirm with y key",
            key: "y", 
            initialState: true,
            expectedState: false,
            expectedAction: "execute_cleanup",
        },
        {
            name: "cancel with n key",
            key: "n",
            initialState: true, 
            expectedState: false,
            expectedAction: "cancel_cleanup",
        },
        {
            name: "cancel with escape",
            key: "esc",
            initialState: true,
            expectedState: false,
            expectedAction: "cancel_cleanup", 
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            model := setupTestModel(nil, nil)
            model.showConfirmationDialog = tt.initialState
            
            newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key[0]}})
            
            assert.Equal(t, tt.expectedState, newModel.(Model).showConfirmationDialog)
            // Verify appropriate command is returned based on expectedAction
        })
    }
}
```

## Integration Tests

### End-to-End User Workflows
```go
func TestCompleteStopWorkflow(t *testing.T) {
    // Setup test environment with active sessions
    mockTmux := &MockTmuxManager{}
    mockSandbox := &MockSandboxManager{}
    model := setupTestModel(mockTmux, mockSandbox)
    
    // Simulate user pressing 's' key
    model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
    
    // Execute the stop command
    msg := executeCommand(cmd)
    
    // Process the stop result message
    model, refreshCmd := model.Update(msg)
    
    // Verify session was stopped and list refreshed
    assert.Equal(t, 0, len(model.sessions))
    assert.NotNil(t, refreshCmd)
}

func TestCompleteCleanWorkflow(t *testing.T) {
    // Setup test environment with stale sessions
    mockTmux := &MockTmuxManager{staleSessions: []string{"work-issue-123"}}
    model := setupTestModel(mockTmux, nil)
    
    // Simulate user pressing 'c' key  
    model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
    assert.True(t, model.showConfirmationDialog)
    
    // Simulate user confirming with 'y'
    model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
    
    // Execute the clean command
    msg := executeCommand(cmd)
    
    // Process the clean result message
    model, refreshCmd := model.Update(msg)
    
    // Verify cleanup occurred and dialog closed
    assert.False(t, model.showConfirmationDialog)
    assert.NotNil(t, refreshCmd)
}
```

### Cross-View Mode Testing
```go
func TestViewModeConsistency(t *testing.T) {
    viewModes := []ViewMode{RepositoryView, GlobalView}
    
    for _, mode := range viewModes {
        t.Run(fmt.Sprintf("operations in %v mode", mode), func(t *testing.T) {
            model := setupTestModel(nil, nil)
            model.viewMode = mode
            
            // Test that operations respect view mode
            // Test help text is consistent across modes
            // Test session filtering works correctly
        })
    }
}
```

## Edge Cases & Error Handling

### Boundary Conditions
```go
func TestEmptySessionsList(t *testing.T) {
    model := setupTestModel(nil, nil)
    model.sessions = []config.SessionMetadata{}
    
    // Test 's' key with no sessions
    _, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
    assert.Nil(t, cmd, "Should not execute stop command with no sessions")
    
    // Test 'c' key with no sessions  
    _, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
    assert.Nil(t, cmd, "Should not show confirmation dialog with no sessions")
}

func TestLargeDatasets(t *testing.T) {
    // Test with 100+ sessions
    sessions := make([]config.SessionMetadata, 100)
    for i := range sessions {
        sessions[i] = config.SessionMetadata{IssueNumber: i + 1}
    }
    
    model := setupTestModel(nil, nil)
    model.sessions = sessions
    
    // Test operations complete within reasonable time
    start := time.Now()
    model.stopSelectedSession()
    duration := time.Since(start)
    assert.Less(t, duration, time.Second, "Stop operation should complete quickly")
}
```

### Error Scenarios
```go
func TestConcurrentModifications(t *testing.T) {
    // Test behavior when sessions are modified externally during operations
}

func TestPermissionErrors(t *testing.T) {  
    // Test handling of permission denied errors
}

func TestNetworkFailures(t *testing.T) {
    // Test behavior when sandbox operations fail due to network issues
}
```

## View Rendering Tests

### Help Text Updates
```go
func TestHelpTextInclusion(t *testing.T) {
    model := setupTestModel(nil, nil)
    
    // Test repository view help text
    model.viewMode = RepositoryView
    view := model.View()
    assert.Contains(t, view, "s to stop")
    assert.Contains(t, view, "c to clean")
    
    // Test global view help text  
    model.viewMode = GlobalView
    view = model.View()
    assert.Contains(t, view, "s to stop")
    assert.Contains(t, view, "c to clean")
}
```

### Modal Dialog Rendering
```go
func TestModalDialogAppearance(t *testing.T) {
    model := setupTestModel(nil, nil)
    model.showConfirmationDialog = true
    model.confirmationMessage = "Clean 2 stale sessions?\nIssue #123, Issue #124"
    
    view := model.View()
    
    // Verify modal content is present
    assert.Contains(t, view, "Clean 2 stale sessions?")
    assert.Contains(t, view, "Issue #123")
    assert.Contains(t, view, "Issue #124")
    
    // Verify modal styling is applied
    assert.Contains(t, view, "y/n") // Confirmation options
}
```

## Test Helpers & Mocks

### Mock Implementations
```go
type MockTmuxManager struct {
    sessions     []string
    staleSessions []string
    killError    error
}

func (m *MockTmuxManager) KillSession(sessionName string) error {
    return m.killError
}

func (m *MockTmuxManager) ListSessions() ([]string, error) {
    return m.sessions, nil
}

type MockSandboxManager struct {
    removeError error
}

func (m *MockSandboxManager) RemoveSandbox(sessionID string) error {
    return m.removeError
}

type MockConfigManager struct {
    sessions    []config.SessionMetadata
    saveError   error
}

func (m *MockConfigManager) SaveSessions(sessions []config.SessionMetadata) error {
    return m.saveError
}
```

### Test Setup Helpers
```go
func setupTestModel(tmux TmuxManager, sandbox SandboxManager) Model {
    return Model{
        tmuxManager:    tmux,
        sandboxManager: sandbox,
        sessions:       []config.SessionMetadata{},
        selectedIndex:  0,
    }
}

func executeCommand(cmd tea.Cmd) tea.Msg {
    if cmd == nil {
        return nil
    }
    return cmd()
}
```

## Coverage Goals

- **90%+ unit test coverage** for new functionality
- **100% coverage** of error scenarios and edge cases  
- **Complete integration testing** of user interaction paths
- **Performance validation** with large datasets

## Test Execution Strategy

### Development Phase
```bash
# Run tests during development
go test ./pkg/tui/... -v
go test ./pkg/tui/... -race
go test ./pkg/tui/... -cover
```

### CI/CD Integration
```bash
# Full test suite with coverage
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Performance Benchmarks
```go
func BenchmarkStopSession(b *testing.B) {
    model := setupTestModel(nil, nil)
    for i := 0; i < b.N; i++ {
        model.stopSelectedSession()
    }
}

func BenchmarkCleanSessions(b *testing.B) {
    model := setupTestModel(nil, nil)
    for i := 0; i < b.N; i++ {
        model.cleanStaleSessions()
    }
}
```

## Backwards Compatibility Verification

### Regression Tests
```go
func TestExistingFunctionalityPreserved(t *testing.T) {
    // Test that existing key bindings still work
    // Test that view switching is unaffected
    // Test that session listing behavior is unchanged
    // Test that attach functionality works as before
}
```

This comprehensive testing plan ensures robust implementation of the stop and clean session management features while maintaining the reliability and performance of the existing TUI interface.