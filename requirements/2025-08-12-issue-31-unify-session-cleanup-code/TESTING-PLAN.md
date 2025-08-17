# Testing Plan: Unify Session Cleanup Code Between TUI and CLI

## Overview

This testing plan implements a Test-Driven Development (TDD) approach for creating a unified session cleanup system between the CLI (`cmd/clean.go`) and TUI (`pkg/tui/model.go`) interfaces. The plan covers the creation of a new `pkg/cleanup` package with a `CleanupManager` that consolidates ~150+ lines of duplicated cleanup logic while preserving the unique characteristics of each interface.

## Problem Analysis

### Current State
- **CLI cleanup (cmd/clean.go)**: Comprehensive (worktrees + sandboxes), user interaction, detailed logging, dry-run support, force mode
- **TUI cleanup (pkg/tui/model.go)**: Sandbox-only cleanup, async tea.Cmd pattern, silent operation
- **Duplication**: ~150+ lines of similar cleanup logic across both interfaces
- **Inconsistency**: Different behaviors for the same underlying operations

### Solution Approach
- Create `pkg/cleanup` package with unified `CleanupManager`
- Maintain interface-specific characteristics (CLI: interactive/verbose, TUI: silent/async)
- Eliminate code duplication while preserving functionality
- Ensure consistent cleanup behavior across interfaces

## Test Structure & Organization

### Test Files
```
pkg/cleanup/
├── manager.go              # New - unified cleanup manager
├── manager_test.go         # New - core cleanup functionality tests
├── integration_test.go     # New - CLI/TUI integration tests
├── mock_test.go           # New - test mocks and helpers
└── benchmark_test.go      # New - performance tests

cmd/
├── clean.go               # Modified - use new cleanup package
└── clean_test.go          # Extended - integration with cleanup manager

pkg/tui/
├── model.go               # Modified - use new cleanup package  
└── stop_clean_test.go     # Extended - integration with cleanup manager
```

### Test Categories

1. **Unit Tests** - Core cleanup functionality with mocking
2. **Integration Tests** - CLI/TUI interface compatibility
3. **Regression Tests** - Ensure no functionality loss
4. **Performance Tests** - Verify efficiency gains from unification
5. **Edge Case Tests** - Error scenarios and boundary conditions

## TDD Implementation Approach

### Phase 1: Red - Write Failing Tests

#### 1.1 Core Cleanup Manager Tests
```go
func TestCleanupManager_Creation(t *testing.T) {
    // Test CleanupManager struct exists and initializes correctly
    // Test dependency injection for tmux, sandbox, and git managers
    // Test configuration options setup
}

func TestCleanupManager_SessionIdentification(t *testing.T) {
    // Test identification of stale sessions across repositories
    // Test filtering by repository context
    // Test handling of namespaced session IDs
}

func TestCleanupManager_ResourceCleanup(t *testing.T) {
    // Test worktree cleanup functionality
    // Test sandbox cleanup functionality
    // Test orphaned branch cleanup functionality
    // Test combined resource cleanup
}
```

#### 1.2 Interface Integration Tests
```go
func TestCLI_CleanupManagerIntegration(t *testing.T) {
    // Test CLI cleanup command uses CleanupManager
    // Test dry-run mode compatibility
    // Test force mode compatibility
    // Test user confirmation flow
    // Test detailed logging output
}

func TestTUI_CleanupManagerIntegration(t *testing.T) {
    // Test TUI cleanup uses CleanupManager
    // Test async tea.Cmd pattern compatibility
    // Test silent operation mode
    // Test view-specific session filtering
}
```

#### 1.3 Behavior Consistency Tests
```go
func TestCleanup_BehaviorConsistency(t *testing.T) {
    // Test identical cleanup results between CLI and TUI
    // Test same sessions identified as stale
    // Test same resources cleaned up
    // Test error handling consistency
}
```

### Phase 2: Green - Implement Minimal Functionality

#### 2.1 Create CleanupManager Package
- Implement basic `CleanupManager` struct with dependencies
- Add session identification methods
- Add resource cleanup methods
- Implement interface-specific options

#### 2.2 Extract Common Logic
- Move stale session detection from both interfaces
- Move sandbox cleanup logic to shared implementation  
- Move worktree cleanup logic to shared implementation
- Add branch cleanup support

#### 2.3 Update Interfaces
- Modify `cmd/clean.go` to use `CleanupManager`
- Modify `pkg/tui/model.go` to use `CleanupManager`
- Preserve interface-specific characteristics

### Phase 3: Refactor - Optimize and Polish

#### 3.1 Performance Optimization
- Optimize session loading and filtering
- Minimize duplicate filesystem operations
- Improve error handling and recovery

#### 3.2 Code Quality
- Ensure consistent error messages
- Add comprehensive logging
- Validate security considerations

## Unit Tests

### Core CleanupManager Functionality
```go
func TestCleanupManager_StaleSessionDetection(t *testing.T) {
    tests := []struct {
        name              string
        sessions          []config.SessionMetadata
        activeTmuxSessions []string
        expectedStale     []string
        expectedError     error
    }{
        {
            name: "identify stale sessions correctly",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
                {NamespacedID: "124", TmuxSession: "sbs-124"},
            },
            activeTmuxSessions: []string{"sbs-123"},
            expectedStale:      []string{"124"},
            expectedError:      nil,
        },
        {
            name: "handle tmux manager errors",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            },
            activeTmuxSessions: nil,
            expectedStale:      []string{},
            expectedError:      errors.New("tmux manager error"),
        },
        {
            name: "no stale sessions found",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            },
            activeTmuxSessions: []string{"sbs-123"},
            expectedStale:      []string{},
            expectedError:      nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTmux := &MockTmuxManager{
                sessions: tt.activeTmuxSessions,
                error:    tt.expectedError,
            }
            
            manager := NewCleanupManager(mockTmux, nil, nil, nil)
            staleSessions, err := manager.IdentifyStaleSessionsInView(tt.sessions, ViewModeGlobal)
            
            if tt.expectedError != nil {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError.Error())
            } else {
                assert.NoError(t, err)
                staleIDs := extractSessionIDs(staleSessions)
                assert.ElementsMatch(t, tt.expectedStale, staleIDs)
            }
        })
    }
}

func TestCleanupManager_ResourceCleanup(t *testing.T) {
    tests := []struct {
        name            string
        sessions        []config.SessionMetadata
        sandboxExists   map[string]bool
        worktreeExists  map[string]bool
        cleanupOptions  CleanupOptions
        expectedResults CleanupResults
        expectedError   error
    }{
        {
            name: "successful comprehensive cleanup",
            sessions: []config.SessionMetadata{
                {
                    NamespacedID:   "123",
                    SandboxName:    "sbs-repo-123",
                    WorktreePath:   "/path/to/worktree-123",
                    TmuxSession:    "sbs-123",
                },
            },
            sandboxExists:  map[string]bool{"sbs-repo-123": true},
            worktreeExists: map[string]bool{"/path/to/worktree-123": true},
            cleanupOptions: CleanupOptions{
                CleanSandboxes: true,
                CleanWorktrees: true,
                DryRun:        false,
            },
            expectedResults: CleanupResults{
                CleanedSessions:  1,
                CleanedSandboxes: 1,
                CleanedWorktrees: 1,
                Errors:          []error{},
            },
            expectedError: nil,
        },
        {
            name: "dry run mode",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", SandboxName: "sbs-repo-123"},
            },
            sandboxExists: map[string]bool{"sbs-repo-123": true},
            cleanupOptions: CleanupOptions{
                CleanSandboxes: true,
                DryRun:        true,
            },
            expectedResults: CleanupResults{
                CleanedSessions:  0, // Nothing actually cleaned in dry run
                CleanedSandboxes: 0,
                WouldClean:      1,
                Errors:         []error{},
            },
            expectedError: nil,
        },
        {
            name: "partial cleanup failures",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", SandboxName: "sbs-repo-123"},
                {NamespacedID: "124", SandboxName: "sbs-repo-124"},
            },
            sandboxExists: map[string]bool{
                "sbs-repo-123": true,
                "sbs-repo-124": true,
            },
            cleanupOptions: CleanupOptions{CleanSandboxes: true},
            expectedResults: CleanupResults{
                CleanedSessions:  1,
                CleanedSandboxes: 1,
                Errors:          []error{errors.New("sandbox cleanup failed")},
            },
            expectedError: nil, // Partial failures don't fail entire operation
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSandbox := &MockSandboxManager{
                sandboxes: tt.sandboxExists,
                deleteErrors: map[string]error{
                    "sbs-repo-124": errors.New("sandbox cleanup failed"),
                },
            }
            
            mockGit := &MockGitManager{
                worktrees: tt.worktreeExists,
            }
            
            manager := NewCleanupManager(nil, mockSandbox, mockGit, nil)
            results, err := manager.CleanupSessions(tt.sessions, tt.cleanupOptions)
            
            if tt.expectedError != nil {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            assert.Equal(t, tt.expectedResults.CleanedSessions, results.CleanedSessions)
            assert.Equal(t, tt.expectedResults.CleanedSandboxes, results.CleanedSandboxes)
            assert.Equal(t, tt.expectedResults.CleanedWorktrees, results.CleanedWorktrees)
            assert.Len(t, results.Errors, len(tt.expectedResults.Errors))
        })
    }
}
```

### Interface-Specific Options
```go
func TestCleanupManager_InterfaceOptions(t *testing.T) {
    tests := []struct {
        name           string
        options        CleanupOptions
        expectedConfig CleanupConfig
    }{
        {
            name: "CLI options with confirmation and logging",
            options: CleanupOptions{
                RequireConfirmation: true,
                VerboseLogging:     true,
                DryRun:            true,
                Force:             false,
            },
            expectedConfig: CleanupConfig{
                Interactive: true,
                Verbose:    true,
                Preview:    true,
            },
        },
        {
            name: "TUI options with silent operation",
            options: CleanupOptions{
                RequireConfirmation: false,
                VerboseLogging:     false,
                SilentMode:         true,
                AsyncOperation:     true,
            },
            expectedConfig: CleanupConfig{
                Interactive: false,
                Verbose:    false,
                Silent:     true,
                Async:      true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := NewCleanupManager(nil, nil, nil, nil)
            config := manager.BuildCleanupConfig(tt.options)
            
            assert.Equal(t, tt.expectedConfig.Interactive, config.Interactive)
            assert.Equal(t, tt.expectedConfig.Verbose, config.Verbose)
            assert.Equal(t, tt.expectedConfig.Silent, config.Silent)
        })
    }
}
```

## Integration Tests

### CLI Integration
```go
func TestCLI_UnifiedCleanupIntegration(t *testing.T) {
    tests := []struct {
        name          string
        args          []string
        sessions      []config.SessionMetadata
        expectedOutput string
        expectedError  error
    }{
        {
            name: "clean command uses unified manager",
            args: []string{"clean", "--dry-run"},
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            },
            expectedOutput: "Found 1 stale session(s):\n  Work Item 123:",
            expectedError:  nil,
        },
        {
            name: "force cleanup without confirmation",
            args: []string{"clean", "--force"},
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            },
            expectedOutput: "Cleaning up stale sessions...",
            expectedError:  nil,
        },
        {
            name: "comprehensive cleanup with branches",
            args: []string{"clean", "--all"},
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            },
            expectedOutput: "Performing comprehensive cleanup",
            expectedError:  nil,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test environment
            setupTestSessions(tt.sessions)
            
            // Capture output
            output, err := executeCommand(tt.args)
            
            if tt.expectedError != nil {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Contains(t, output, tt.expectedOutput)
            }
        })
    }
}
```

### TUI Integration
```go
func TestTUI_UnifiedCleanupIntegration(t *testing.T) {
    tests := []struct {
        name              string
        initialSessions   []config.SessionMetadata
        keySequence       []tea.Msg
        expectedDialogMsg string
        expectedCleaned   int
    }{
        {
            name: "TUI cleanup uses unified manager",
            initialSessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
                {NamespacedID: "124", TmuxSession: "sbs-124"},
            },
            keySequence: []tea.Msg{
                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
            },
            expectedDialogMsg: "Clean 2 stale sessions?",
            expectedCleaned:   2,
        },
        {
            name: "TUI respects view mode filtering",
            initialSessions: []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123", RepositoryName: "repo1"},
                {NamespacedID: "124", TmuxSession: "sbs-124", RepositoryName: "repo2"},
            },
            keySequence: []tea.Msg{
                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}},
                tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
            },
            expectedDialogMsg: "Clean 1 stale session?", // Only current repo
            expectedCleaned:   1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            model := setupTestModel()
            model.sessions = tt.initialSessions
            
            // Execute key sequence
            for _, key := range tt.keySequence {
                var cmd tea.Cmd
                model, cmd = model.Update(key)
                
                if cmd != nil {
                    msg := executeCommand(cmd)
                    if msg != nil {
                        model, _ = model.Update(msg)
                    }
                }
            }
            
            // Verify expected behavior based on test case
            if tt.expectedDialogMsg != "" {
                assert.Contains(t, model.confirmationMessage, tt.expectedDialogMsg)
            }
        })
    }
}
```

### Cross-Interface Consistency
```go
func TestCrossInterface_CleanupConsistency(t *testing.T) {
    sessions := []config.SessionMetadata{
        {NamespacedID: "123", TmuxSession: "sbs-123", SandboxName: "sbs-repo-123"},
        {NamespacedID: "124", TmuxSession: "sbs-124", SandboxName: "sbs-repo-124"},
    }
    
    // Setup identical test environment
    setupIdenticalTestEnvironment(sessions)
    
    // Execute CLI cleanup
    cliResults := executeCLICleanup([]string{"clean", "--force"})
    
    // Reset environment
    setupIdenticalTestEnvironment(sessions)
    
    // Execute TUI cleanup  
    tuiResults := executeTUICleanup()
    
    // Verify identical results
    assert.Equal(t, cliResults.cleanedSessions, tuiResults.cleanedSessions)
    assert.Equal(t, cliResults.cleanedSandboxes, tuiResults.cleanedSandboxes)
    assert.Equal(t, len(cliResults.errors), len(tuiResults.errors))
}
```

## Regression Tests

### Preserve Existing Functionality
```go
func TestRegression_CLIFunctionalityPreserved(t *testing.T) {
    tests := []struct {
        name        string
        command     []string
        expectation string
    }{
        {
            name:        "dry-run still works",
            command:     []string{"clean", "--dry-run"},
            expectation: "Dry run - no changes made",
        },
        {
            name:        "force mode still works",
            command:     []string{"clean", "--force"},
            expectation: "Cleaning up stale sessions",
        },
        {
            name:        "branch cleanup still works",
            command:     []string{"clean", "--branches"},
            expectation: "Cleaning up orphaned branches",
        },
        {
            name:        "comprehensive cleanup still works",
            command:     []string{"clean", "--all"},
            expectation: "Performing comprehensive cleanup",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            output, err := executeCommand(tt.command)
            assert.NoError(t, err)
            assert.Contains(t, output, tt.expectation)
        })
    }
}

func TestRegression_TUIFunctionalityPreserved(t *testing.T) {
    tests := []struct {
        name           string
        action         func(Model) (Model, tea.Cmd)
        expectedResult string
    }{
        {
            name: "c key shows confirmation dialog",
            action: func(m Model) (Model, tea.Cmd) {
                return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
            },
            expectedResult: "confirmation_dialog_shown",
        },
        {
            name: "y key executes cleanup",
            action: func(m Model) (Model, tea.Cmd) {
                m.showConfirmationDialog = true
                return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
            },
            expectedResult: "cleanup_executed",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            model := setupTestModel()
            newModel, cmd := tt.action(model)
            
            switch tt.expectedResult {
            case "confirmation_dialog_shown":
                assert.True(t, newModel.(Model).showConfirmationDialog)
            case "cleanup_executed":
                assert.NotNil(t, cmd)
                assert.False(t, newModel.(Model).showConfirmationDialog)
            }
        })
    }
}
```

### Backwards Compatibility
```go
func TestBackwardsCompatibility_ExistingBehavior(t *testing.T) {
    // Test that session metadata format is preserved
    // Test that configuration file format is preserved  
    // Test that command line interface is preserved
    // Test that TUI key bindings are preserved
}
```

## Performance Tests

### Cleanup Operation Efficiency
```go
func BenchmarkCleanupManager_StaleDetection(b *testing.B) {
    sessions := generateTestSessions(100) // 100 sessions
    manager := setupTestCleanupManager()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := manager.IdentifyStaleSessionsInView(sessions, ViewModeGlobal)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCleanupManager_ResourceCleanup(b *testing.B) {
    sessions := generateTestSessions(50) // 50 sessions
    manager := setupTestCleanupManager()
    options := CleanupOptions{
        CleanSandboxes: true,
        CleanWorktrees: true,
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := manager.CleanupSessions(sessions, options)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func TestPerformance_LargeDatasets(t *testing.T) {
    sizes := []int{10, 50, 100, 500, 1000}
    
    for _, size := range sizes {
        t.Run(fmt.Sprintf("sessions_%d", size), func(t *testing.T) {
            sessions := generateTestSessions(size)
            manager := setupTestCleanupManager()
            
            start := time.Now()
            _, err := manager.IdentifyStaleSessionsInView(sessions, ViewModeGlobal)
            duration := time.Since(start)
            
            assert.NoError(t, err)
            
            // Performance expectations based on dataset size
            maxDuration := time.Duration(size) * time.Millisecond
            assert.Less(t, duration, maxDuration,
                "Operation should complete within reasonable time for %d sessions", size)
        })
    }
}
```

## Edge Cases & Error Handling

### Boundary Conditions
```go
func TestEdgeCases_EmptyInputs(t *testing.T) {
    manager := NewCleanupManager(nil, nil, nil, nil)
    
    // Test empty session list
    results, err := manager.CleanupSessions([]config.SessionMetadata{}, CleanupOptions{})
    assert.NoError(t, err)
    assert.Equal(t, 0, results.CleanedSessions)
    
    // Test nil session list
    results, err = manager.CleanupSessions(nil, CleanupOptions{})
    assert.NoError(t, err)
    assert.Equal(t, 0, results.CleanedSessions)
}

func TestEdgeCases_CorruptedData(t *testing.T) {
    tests := []struct {
        name     string
        sessions []config.SessionMetadata
        expected string
    }{
        {
            name: "missing sandbox name",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", SandboxName: ""},
            },
            expected: "invalid sandbox name",
        },
        {
            name: "missing worktree path",
            sessions: []config.SessionMetadata{
                {NamespacedID: "123", WorktreePath: ""},
            },
            expected: "invalid worktree path",
        },
        {
            name: "invalid namespaced ID",
            sessions: []config.SessionMetadata{
                {NamespacedID: "", TmuxSession: "sbs-123"},
            },
            expected: "invalid namespaced ID",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manager := setupTestCleanupManager()
            results, err := manager.CleanupSessions(tt.sessions, CleanupOptions{
                CleanSandboxes: true,
                CleanWorktrees: true,
            })
            
            // Should handle gracefully, not crash
            assert.NotNil(t, results)
            if err != nil {
                assert.Contains(t, err.Error(), tt.expected)
            }
        })
    }
}
```

### Error Scenarios
```go
func TestErrorHandling_DependencyFailures(t *testing.T) {
    tests := []struct {
        name        string
        tmuxError   error
        sandboxError error
        gitError     error
        expectedErr  string
    }{
        {
            name:        "tmux manager failure",
            tmuxError:   errors.New("tmux connection failed"),
            expectedErr: "tmux connection failed",
        },
        {
            name:         "sandbox manager failure",
            sandboxError: errors.New("sandbox service unavailable"),
            expectedErr:  "sandbox service unavailable",
        },
        {
            name:        "git manager failure",
            gitError:    errors.New("git repository not found"),
            expectedErr: "git repository not found",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockTmux := &MockTmuxManager{error: tt.tmuxError}
            mockSandbox := &MockSandboxManager{error: tt.sandboxError}
            mockGit := &MockGitManager{error: tt.gitError}
            
            manager := NewCleanupManager(mockTmux, mockSandbox, mockGit, nil)
            
            sessions := []config.SessionMetadata{
                {NamespacedID: "123", TmuxSession: "sbs-123"},
            }
            
            results, err := manager.CleanupSessions(sessions, CleanupOptions{
                CleanSandboxes: true,
                CleanWorktrees: true,
            })
            
            if tt.expectedErr != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
            }
            
            // Results should still be valid even with errors
            assert.NotNil(t, results)
        })
    }
}
```

## Test Helpers & Mocks

### Mock Implementations
```go
type MockTmuxManager struct {
    sessions []string
    error    error
}

func (m *MockTmuxManager) SessionExists(sessionName string) (bool, error) {
    if m.error != nil {
        return false, m.error
    }
    for _, session := range m.sessions {
        if session == sessionName {
            return true, nil
        }
    }
    return false, nil
}

func (m *MockTmuxManager) KillSession(sessionName string) error {
    return m.error
}

func (m *MockTmuxManager) ListSessions() ([]*tmux.Session, error) {
    if m.error != nil {
        return nil, m.error
    }
    var sessions []*tmux.Session
    for _, name := range m.sessions {
        sessions = append(sessions, &tmux.Session{Name: name})
    }
    return sessions, nil
}

type MockSandboxManager struct {
    sandboxes    map[string]bool
    deleteErrors map[string]error
    error        error
}

func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
    if m.error != nil {
        return false, m.error
    }
    if m.sandboxes == nil {
        return false, nil
    }
    exists, ok := m.sandboxes[sandboxName]
    return exists && ok, nil
}

func (m *MockSandboxManager) DeleteSandbox(sandboxName string) error {
    if m.error != nil {
        return m.error
    }
    if m.deleteErrors != nil {
        if err, exists := m.deleteErrors[sandboxName]; exists {
            return err
        }
    }
    return nil
}

type MockGitManager struct {
    worktrees map[string]bool
    branches  []string
    error     error
}

func (m *MockGitManager) FindOrphanedIssueBranches(activeWorkItems []string) ([]string, error) {
    if m.error != nil {
        return nil, m.error
    }
    return m.branches, nil
}

func (m *MockGitManager) DeleteMultipleBranches(branches []string, force bool) ([]git.BranchDeleteResult, error) {
    if m.error != nil {
        return nil, m.error
    }
    var results []git.BranchDeleteResult
    for _, branch := range branches {
        results = append(results, git.BranchDeleteResult{
            BranchName: branch,
            Success:    true,
            Message:    "deleted",
        })
    }
    return results, nil
}

type MockConfigManager struct {
    sessions  []config.SessionMetadata
    saveError error
}

func (m *MockConfigManager) LoadAllRepositorySessions() ([]config.SessionMetadata, error) {
    return m.sessions, nil
}

func (m *MockConfigManager) SaveSessions(sessions []config.SessionMetadata) error {
    return m.saveError
}
```

### Test Setup Helpers
```go
func setupTestCleanupManager() *CleanupManager {
    return NewCleanupManager(
        &MockTmuxManager{sessions: []string{}},
        &MockSandboxManager{sandboxes: make(map[string]bool)},
        &MockGitManager{worktrees: make(map[string]bool)},
        &MockConfigManager{},
    )
}

func setupTestModel() Model {
    return NewModel()
}

func generateTestSessions(count int) []config.SessionMetadata {
    sessions := make([]config.SessionMetadata, count)
    for i := 0; i < count; i++ {
        sessions[i] = config.SessionMetadata{
            NamespacedID: fmt.Sprintf("%d", i+1),
            TmuxSession:  fmt.Sprintf("sbs-%d", i+1),
            SandboxName:  fmt.Sprintf("sbs-repo-%d", i+1),
            WorktreePath: fmt.Sprintf("/path/to/worktree-%d", i+1),
        }
    }
    return sessions
}

func executeCommand(cmd tea.Cmd) tea.Msg {
    if cmd == nil {
        return nil
    }
    return cmd()
}

func setupTestSessions(sessions []config.SessionMetadata) {
    // Setup test environment with provided sessions
    // This would involve creating temporary directories, etc.
}

func setupIdenticalTestEnvironment(sessions []config.SessionMetadata) {
    // Setup identical test conditions for cross-interface testing
}
```

## Package Structure Definition

### CleanupManager Interface
```go
type CleanupManager interface {
    // Core functionality
    IdentifyStaleSessionsInView(sessions []config.SessionMetadata, viewMode ViewMode) ([]config.SessionMetadata, error)
    CleanupSessions(sessions []config.SessionMetadata, options CleanupOptions) (CleanupResults, error)
    
    // Interface-specific methods
    BuildCLICleanupOptions(dryRun, force bool, mode CleanupMode) CleanupOptions
    BuildTUICleanupOptions(viewMode ViewMode, silent bool) CleanupOptions
    
    // Resource-specific cleanup
    CleanSandboxes(sessions []config.SessionMetadata, dryRun bool) (CleanupResults, error)
    CleanWorktrees(sessions []config.SessionMetadata, dryRun bool) (CleanupResults, error)
    CleanBranches(activeWorkItems []string, dryRun bool) (CleanupResults, error)
}

type CleanupOptions struct {
    // Behavior options
    CleanSandboxes      bool
    CleanWorktrees      bool
    CleanBranches       bool
    DryRun             bool
    Force              bool
    
    // Interface options
    RequireConfirmation bool
    VerboseLogging     bool
    SilentMode         bool
    AsyncOperation     bool
    
    // Context options
    ViewMode           ViewMode
    RepositoryFilter   string
}

type CleanupResults struct {
    CleanedSessions  int
    CleanedSandboxes int
    CleanedWorktrees int
    CleanedBranches  int
    WouldClean       int // For dry run
    Errors           []error
    Details          []string // For verbose output
}
```

## Coverage Goals

- **95%+ unit test coverage** for new `pkg/cleanup` package
- **90%+ integration test coverage** for CLI/TUI interface integration
- **100% coverage** of error scenarios and edge cases
- **Complete regression testing** of existing functionality
- **Performance validation** with large datasets (1000+ sessions)

## Test Execution Strategy

### Development Phase
```bash
# Run cleanup package tests during development
go test ./pkg/cleanup/... -v
go test ./pkg/cleanup/... -race
go test ./pkg/cleanup/... -cover

# Run integration tests
go test ./cmd/... -v -tags=integration
go test ./pkg/tui/... -v -tags=integration
```

### Pre-commit Validation
```bash
# Full test suite with coverage
make test
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Performance benchmarks
go test ./pkg/cleanup/... -bench=. -benchmem
```

### CI/CD Integration
```bash
# Automated testing pipeline
go test ./... -v -race -coverprofile=coverage.out
go test ./... -v -tags=integration
go test ./pkg/cleanup/... -bench=. -benchmem -count=3

# Coverage reporting
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Implementation Phases

### Phase 1: Core Package Creation (TDD)
1. Write failing tests for `CleanupManager` interface
2. Implement minimal `pkg/cleanup` package structure
3. Extract common logic from existing implementations
4. Ensure all tests pass

### Phase 2: CLI Integration (TDD)
1. Write failing tests for CLI integration
2. Modify `cmd/clean.go` to use `CleanupManager`
3. Preserve all existing CLI functionality
4. Ensure all tests pass

### Phase 3: TUI Integration (TDD)
1. Write failing tests for TUI integration
2. Modify `pkg/tui/model.go` to use `CleanupManager`
3. Preserve all existing TUI functionality
4. Ensure all tests pass

### Phase 4: Cross-Interface Validation
1. Write consistency tests between CLI and TUI
2. Verify identical cleanup behavior
3. Performance testing and optimization
4. Final regression testing

## Success Criteria

1. **Code Duplication Eliminated**: ~150+ lines of duplicated cleanup logic consolidated
2. **Functionality Preserved**: All existing CLI and TUI features work identically
3. **Behavior Consistency**: CLI and TUI produce identical cleanup results
4. **Test Coverage**: 95%+ coverage for new package, 90%+ for integrations
5. **Performance**: No regression in cleanup operation performance
6. **Error Handling**: Robust error handling with graceful degradation
7. **Interface Characteristics Maintained**: 
   - CLI: Interactive, verbose, dry-run support
   - TUI: Silent, async, view-aware operations

This comprehensive testing plan ensures a robust, test-driven implementation of the unified session cleanup system that eliminates code duplication while preserving the unique characteristics and functionality of both CLI and TUI interfaces.