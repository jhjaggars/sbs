# Testing Plan: Pluggable Input Sources with Project-Specific Configuration

## Overview

This testing plan covers the comprehensive testing strategy for implementing pluggable input sources in SBS, enabling work with different types of work items beyond GitHub issues. The plan follows test-driven development (TDD) principles and ensures full backward compatibility with existing GitHub issue workflows.

## Architecture Under Test

### Core Components
1. **WorkItem Struct** - Namespaced work items (e.g., "github:123", "test:quick", "jira:PROJ-456")
2. **InputSource Interface** - Abstract interface for different backends
3. **GitHubInputSource** - Existing GitHub integration wrapped in new interface
4. **TestInputSource** - Built-in test source for rapid development/testing
5. **Project Configuration** - `.sbs/input-source.json` configuration files
6. **Session Migration** - Backward compatibility for existing sessions

### Key Features Under Test
- ID namespacing system to prevent conflicts
- Test input source with predefined work items
- Project-specific input source configuration
- Backward compatibility for existing `sbs start 123` commands
- Git branch naming: `issue-{source}-{id}-{title-slug}`
- Session metadata migration and enhancement

## Testing Framework and Tools

### Go Testing Stack
- **Framework**: Go standard `testing` package
- **Assertions**: `testify/assert` and `testify/require` (already in use)
- **Mocking**: `testify/mock` for external dependencies
- **Test Coverage**: `go test -cover` and `go tool cover`
- **Integration**: Custom integration test framework with build tag

### File Structure for Tests
```
pkg/
├── inputsource/
│   ├── interface.go              # InputSource interface definition
│   ├── interface_test.go         # Interface compliance tests
│   ├── workitem.go               # WorkItem struct and utilities
│   ├── workitem_test.go          # WorkItem unit tests
│   ├── github.go                 # GitHub InputSource implementation
│   ├── github_test.go            # GitHub InputSource unit tests
│   ├── test.go                   # Test InputSource implementation
│   ├── test_test.go              # Test InputSource unit tests
│   ├── factory.go                # InputSource factory
│   ├── factory_test.go           # Factory unit tests
│   └── test_helpers.go           # Shared test utilities
├── config/
│   ├── inputsource_config.go     # Input source configuration
│   ├── inputsource_config_test.go # Configuration tests
│   └── migration.go              # Session metadata migration
│   └── migration_test.go         # Migration tests
└── cmd/
    ├── start.go                  # Updated start command
    └── start_test.go             # Integration tests
```

## Testing Strategy by Component

## 1. Unit Testing Strategy

### 1.1 WorkItem Structure Tests

**File**: `pkg/inputsource/workitem_test.go`

```go
func TestWorkItem_ParseID(t *testing.T) {
    tests := []struct {
        name           string
        input          string
        expectedSource string
        expectedID     string
        expectedError  bool
    }{
        {"github_namespaced", "github:123", "github", "123", false},
        {"test_namespaced", "test:quick", "test", "quick", false},
        {"jira_namespaced", "jira:PROJ-456", "jira", "PROJ-456", false},
        {"legacy_github", "123", "github", "123", false}, // backward compatibility
        {"invalid_format", "invalid-format", "", "", true},
        {"empty_source", ":123", "", "", true},
        {"empty_id", "github:", "", "", true},
    }
    // Implementation tests...
}

func TestWorkItem_FullID(t *testing.T) {
    // Test full ID generation
}

func TestWorkItem_IsLegacyFormat(t *testing.T) {
    // Test legacy format detection
}

func TestWorkItem_GetBranchName(t *testing.T) {
    // Test branch name generation: issue-{source}-{id}-{title-slug}
}
```

**Priority**: High - Foundation component
**TDD Phase**: Implement first

### 1.2 InputSource Interface Tests

**File**: `pkg/inputsource/interface_test.go`

```go
func TestInputSource_InterfaceCompliance(t *testing.T) {
    // Test that all implementations satisfy the interface
    var _ InputSource = (*GitHubInputSource)(nil)
    var _ InputSource = (*TestInputSource)(nil)
}

func TestInputSource_GetWorkItem_ErrorHandling(t *testing.T) {
    // Test error scenarios across all implementations
}

func TestInputSource_ListWorkItems_Consistency(t *testing.T) {
    // Test consistent behavior across implementations
}
```

**Priority**: High - Interface contract validation
**TDD Phase**: Implement with interface definition

### 1.3 TestInputSource Implementation Tests

**File**: `pkg/inputsource/test_test.go`

```go
func TestTestInputSource_PredefinedItems(t *testing.T) {
    tests := []struct {
        name        string
        id          string
        expectFound bool
        expectTitle string
    }{
        {"quick_test", "quick", true, "Quick development test"},
        {"hooks_test", "hooks", true, "Test Claude Code hooks"},
        {"sandbox_test", "sandbox", true, "Test sandbox integration"},
        {"invalid_id", "nonexistent", false, ""},
    }
    
    source := NewTestInputSource()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            item, err := source.GetWorkItem(tt.id)
            if tt.expectFound {
                require.NoError(t, err)
                assert.Equal(t, tt.expectTitle, item.Title)
                assert.Equal(t, "test", item.Source)
                assert.Equal(t, tt.id, item.ID)
            } else {
                assert.Error(t, err)
                assert.Nil(t, item)
            }
        })
    }
}

func TestTestInputSource_ListWorkItems(t *testing.T) {
    source := NewTestInputSource()
    items, err := source.ListWorkItems("", 10)
    
    require.NoError(t, err)
    assert.Len(t, items, 3) // quick, hooks, sandbox
    
    // Verify all items have correct structure
    for _, item := range items {
        assert.Equal(t, "test", item.Source)
        assert.NotEmpty(t, item.ID)
        assert.NotEmpty(t, item.Title)
        assert.Equal(t, "open", item.State)
    }
}

func TestTestInputSource_SearchFiltering(t *testing.T) {
    source := NewTestInputSource()
    
    // Test search functionality
    items, err := source.ListWorkItems("hooks", 10)
    require.NoError(t, err)
    
    found := false
    for _, item := range items {
        if strings.Contains(strings.ToLower(item.Title), "hooks") {
            found = true
            break
        }
    }
    assert.True(t, found, "Search should find items containing 'hooks'")
}
```

**Priority**: High - Critical for development workflow
**TDD Phase**: Implement early for rapid testing

### 1.4 GitHubInputSource Wrapper Tests

**File**: `pkg/inputsource/github_test.go`

```go
func TestGitHubInputSource_GetWorkItem(t *testing.T) {
    // Mock the underlying GitHub client
    mockClient := &mockGitHubClient{
        issues: map[int]*issue.Issue{
            123: {Number: 123, Title: "Fix auth bug", State: "open", URL: "https://github.com/test/repo/issues/123"},
        },
    }
    
    source := &GitHubInputSource{client: mockClient}
    
    t.Run("successful_get", func(t *testing.T) {
        item, err := source.GetWorkItem("123")
        require.NoError(t, err)
        assert.Equal(t, "github", item.Source)
        assert.Equal(t, "123", item.ID)
        assert.Equal(t, "Fix auth bug", item.Title)
        assert.Equal(t, "github:123", item.FullID())
    })
    
    t.Run("invalid_id_format", func(t *testing.T) {
        item, err := source.GetWorkItem("invalid")
        assert.Error(t, err)
        assert.Nil(t, item)
    })
    
    t.Run("issue_not_found", func(t *testing.T) {
        item, err := source.GetWorkItem("999")
        assert.Error(t, err)
        assert.Nil(t, item)
    })
}

func TestGitHubInputSource_BackwardCompatibility(t *testing.T) {
    // Test that existing GitHub workflows still work
    mockClient := &mockGitHubClient{
        issues: map[int]*issue.Issue{
            456: {Number: 456, Title: "Legacy issue", State: "open"},
        },
    }
    
    source := &GitHubInputSource{client: mockClient}
    item, err := source.GetWorkItem("456")
    
    require.NoError(t, err)
    assert.Equal(t, "github:456", item.FullID())
    assert.Equal(t, "issue-github-456-legacy-issue", item.GetBranchName())
}

func TestGitHubInputSource_ListWorkItems(t *testing.T) {
    // Test list functionality with GitHub client
    mockClient := &mockGitHubClient{
        listResult: []issue.Issue{
            {Number: 1, Title: "First issue", State: "open"},
            {Number: 2, Title: "Second issue", State: "open"},
        },
    }
    
    source := &GitHubInputSource{client: mockClient}
    items, err := source.ListWorkItems("", 10)
    
    require.NoError(t, err)
    assert.Len(t, items, 2)
    
    for _, item := range items {
        assert.Equal(t, "github", item.Source)
        assert.NotEmpty(t, item.ID)
    }
}
```

**Priority**: High - Ensures no regression in existing functionality
**TDD Phase**: Implement with interface wrapper

### 1.5 InputSource Factory Tests

**File**: `pkg/inputsource/factory_test.go`

```go
func TestInputSourceFactory_CreateFromConfig(t *testing.T) {
    tests := []struct {
        name           string
        config         *InputSourceConfig
        expectedType   string
        expectedError  bool
    }{
        {
            name:         "github_source",
            config:       &InputSourceConfig{Type: "github"},
            expectedType: "github",
            expectedError: false,
        },
        {
            name:         "test_source",
            config:       &InputSourceConfig{Type: "test"},
            expectedType: "test",
            expectedError: false,
        },
        {
            name:         "unknown_source",
            config:       &InputSourceConfig{Type: "unknown"},
            expectedType: "",
            expectedError: true,
        },
        {
            name:         "nil_config_defaults_to_github",
            config:       nil,
            expectedType: "github",
            expectedError: false,
        },
    }
    
    factory := NewInputSourceFactory()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            source, err := factory.Create(tt.config)
            
            if tt.expectedError {
                assert.Error(t, err)
                assert.Nil(t, source)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, source)
                // Verify type through interface behavior
            }
        })
    }
}
```

**Priority**: Medium - Important for configuration system
**TDD Phase**: Implement with configuration

## 2. Configuration Testing Strategy

### 2.1 Input Source Configuration Tests

**File**: `pkg/config/inputsource_config_test.go`

```go
func TestInputSourceConfig_LoadFromProject(t *testing.T) {
    // Test loading .sbs/input-source.json from project root
    tempDir := t.TempDir()
    
    // Create test configuration
    sbsDir := filepath.Join(tempDir, ".sbs")
    require.NoError(t, os.MkdirAll(sbsDir, 0755))
    
    configData := `{
        "type": "test",
        "settings": {
            "description": "Test project configuration"
        }
    }`
    
    configPath := filepath.Join(sbsDir, "input-source.json")
    require.NoError(t, os.WriteFile(configPath, []byte(configData), 0644))
    
    // Test loading
    config, err := LoadInputSourceConfig(tempDir)
    require.NoError(t, err)
    assert.Equal(t, "test", config.Type)
}

func TestInputSourceConfig_FallbackToDefault(t *testing.T) {
    // Test that missing config falls back to GitHub
    tempDir := t.TempDir()
    
    config, err := LoadInputSourceConfig(tempDir)
    require.NoError(t, err)
    assert.Equal(t, "github", config.Type)
}

func TestInputSourceConfig_ValidationError(t *testing.T) {
    // Test invalid configuration
    tempDir := t.TempDir()
    sbsDir := filepath.Join(tempDir, ".sbs")
    require.NoError(t, os.MkdirAll(sbsDir, 0755))
    
    // Invalid JSON
    configPath := filepath.Join(sbsDir, "input-source.json")
    require.NoError(t, os.WriteFile(configPath, []byte(`{invalid json}`), 0644))
    
    config, err := LoadInputSourceConfig(tempDir)
    assert.Error(t, err)
    assert.Nil(t, config)
}

func TestInputSourceConfig_ConfigDirectoryCreation(t *testing.T) {
    // Test that .sbs directory is created if it doesn't exist
    tempDir := t.TempDir()
    
    config := &InputSourceConfig{Type: "test"}
    err := SaveInputSourceConfig(tempDir, config)
    require.NoError(t, err)
    
    // Verify directory and file were created
    configPath := filepath.Join(tempDir, ".sbs", "input-source.json")
    assert.FileExists(t, configPath)
}
```

**Priority**: High - Core configuration functionality
**TDD Phase**: Implement with configuration loading

### 2.2 Session Migration Tests

**File**: `pkg/config/migration_test.go`

```go
func TestSessionMigration_LegacyToNamespaced(t *testing.T) {
    // Test migration of existing sessions to use namespaced IDs
    legacySessions := []SessionMetadata{
        {
            IssueNumber: 123,
            IssueTitle:  "Legacy issue",
            Branch:      "issue-123-legacy-issue",
            // Missing SourceType, NamespacedID fields
        },
    }
    
    migratedSessions, err := MigrateSessionMetadata(legacySessions)
    require.NoError(t, err)
    assert.Len(t, migratedSessions, 1)
    
    migrated := migratedSessions[0]
    assert.Equal(t, "github", migrated.SourceType)
    assert.Equal(t, "github:123", migrated.NamespacedID)
    assert.Equal(t, 123, migrated.IssueNumber) // Preserved for compatibility
}

func TestSessionMigration_AlreadyMigrated(t *testing.T) {
    // Test that already migrated sessions are not modified
    existingSessions := []SessionMetadata{
        {
            IssueNumber:  456,
            SourceType:   "test",
            NamespacedID: "test:quick",
            Branch:       "issue-test-quick-dev-test",
        },
    }
    
    migratedSessions, err := MigrateSessionMetadata(existingSessions)
    require.NoError(t, err)
    assert.Len(t, migratedSessions, 1)
    
    // Should remain unchanged
    assert.Equal(t, existingSessions[0], migratedSessions[0])
}

func TestSessionMigration_MixedSessions(t *testing.T) {
    // Test migration of mixed legacy and new sessions
    mixedSessions := []SessionMetadata{
        // Legacy session
        {IssueNumber: 123, Branch: "issue-123-legacy"},
        // Already migrated session
        {IssueNumber: 456, SourceType: "github", NamespacedID: "github:456"},
    }
    
    migratedSessions, err := MigrateSessionMetadata(mixedSessions)
    require.NoError(t, err)
    assert.Len(t, migratedSessions, 2)
    
    // Check first session was migrated
    assert.Equal(t, "github", migratedSessions[0].SourceType)
    assert.Equal(t, "github:123", migratedSessions[0].NamespacedID)
    
    // Check second session unchanged
    assert.Equal(t, "github:456", migratedSessions[1].NamespacedID)
}
```

**Priority**: High - Critical for backward compatibility
**TDD Phase**: Implement early to ensure compatibility

## 3. Integration Testing Strategy

### 3.1 Start Command Integration Tests

**File**: `cmd/start_test.go` (additions)

```go
func TestStartCommand_InputSourceIntegration(t *testing.T) {
    // Integration test with build tag
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    t.Run("test_input_source_quick", func(t *testing.T) {
        // Setup test repository with test input source config
        tempRepo := setupTestRepository(t)
        defer cleanupTestRepository(t, tempRepo)
        
        // Create test input source configuration
        configureTestInputSource(t, tempRepo.Root)
        
        // Mock tmux and git operations
        mockTmux := &mockTmuxManager{}
        mockGit := &mockGitManager{}
        
        // Execute start command with test source
        err := runStartWithMocks([]string{"test:quick"}, mockTmux, mockGit)
        require.NoError(t, err)
        
        // Verify correct branch creation
        assert.Contains(t, mockGit.createdBranches, "issue-test-quick-quick-development-test")
        
        // Verify tmux session creation
        assert.Len(t, mockTmux.createdSessions, 1)
        assert.Contains(t, mockTmux.createdSessions[0].Name, "work-issue-test-quick")
    })
    
    t.Run("github_backward_compatibility", func(t *testing.T) {
        // Test that existing `sbs start 123` still works
        tempRepo := setupTestRepository(t)
        defer cleanupTestRepository(t, tempRepo)
        
        // No input source config = defaults to GitHub
        mockGitHub := &mockGitHubClient{
            issues: map[int]*issue.Issue{
                123: {Number: 123, Title: "GitHub issue", State: "open"},
            },
        }
        
        err := runStartWithGitHubMock([]string{"123"}, mockGitHub)
        require.NoError(t, err)
        
        // Verify GitHub issue was fetched
        assert.True(t, mockGitHub.getIssueCalled)
        assert.Equal(t, 123, mockGitHub.lastIssueNumber)
    })
    
    t.Run("namespaced_id_input", func(t *testing.T) {
        // Test explicit namespaced ID input
        tempRepo := setupTestRepository(t)
        defer cleanupTestRepository(t, tempRepo)
        
        configureTestInputSource(t, tempRepo.Root)
        
        err := runStartWithMocks([]string{"test:hooks"}, nil, nil)
        require.NoError(t, err)
        
        // Verify session metadata uses namespaced ID
        sessions, err := config.LoadSessions()
        require.NoError(t, err)
        
        found := false
        for _, session := range sessions {
            if session.NamespacedID == "test:hooks" {
                found = true
                assert.Equal(t, "test", session.SourceType)
                break
            }
        }
        assert.True(t, found, "Session should have namespaced ID")
    })
}

func TestStartCommand_ProjectSpecificConfiguration(t *testing.T) {
    t.Run("different_projects_different_sources", func(t *testing.T) {
        // Test that different projects can use different input sources
        project1 := setupTestRepository(t)
        project2 := setupTestRepository(t)
        defer cleanupTestRepository(t, project1)
        defer cleanupTestRepository(t, project2)
        
        // Configure project1 for test source
        configureTestInputSource(t, project1.Root)
        
        // Configure project2 for GitHub source
        configureGitHubInputSource(t, project2.Root)
        
        // Test project1 uses test source
        os.Chdir(project1.Root)
        err := runStartWithMocks([]string{"test:quick"}, nil, nil)
        require.NoError(t, err)
        
        // Test project2 uses GitHub source
        os.Chdir(project2.Root)
        mockGitHub := &mockGitHubClient{
            issues: map[int]*issue.Issue{
                456: {Number: 456, Title: "GitHub issue", State: "open"},
            },
        }
        err = runStartWithGitHubMock([]string{"456"}, mockGitHub)
        require.NoError(t, err)
    })
}
```

**Priority**: High - Validates end-to-end functionality
**TDD Phase**: Implement after unit tests pass

### 3.2 Session Management Integration Tests

**File**: `pkg/config/config_test.go` (additions)

```go
func TestSessionManagement_InputSourceIntegration(t *testing.T) {
    t.Run("list_sessions_mixed_sources", func(t *testing.T) {
        // Test listing sessions from different input sources
        sessions := []SessionMetadata{
            {
                IssueNumber:  123,
                SourceType:   "github",
                NamespacedID: "github:123",
                IssueTitle:   "GitHub issue",
                Status:       "active",
            },
            {
                IssueNumber:  0, // Test items don't have issue numbers
                SourceType:   "test",
                NamespacedID: "test:quick",
                IssueTitle:   "Quick development test",
                Status:       "active",
            },
        }
        
        // Save sessions
        err := config.SaveSessions(sessions)
        require.NoError(t, err)
        
        // Load and verify
        loadedSessions, err := config.LoadSessions()
        require.NoError(t, err)
        assert.Len(t, loadedSessions, 2)
        
        // Verify both source types are represented
        sourceTypes := make(map[string]bool)
        for _, session := range loadedSessions {
            sourceTypes[session.SourceType] = true
        }
        assert.True(t, sourceTypes["github"])
        assert.True(t, sourceTypes["test"])
    })
}
```

**Priority**: Medium - Important for mixed environment support
**TDD Phase**: Implement after core functionality

## 4. Backward Compatibility Testing

### 4.1 Legacy Command Compatibility Tests

**File**: `cmd/start_test.go` (additions)

```go
func TestStartCommand_BackwardCompatibility(t *testing.T) {
    t.Run("legacy_numeric_argument", func(t *testing.T) {
        // Test that `sbs start 123` still works exactly as before
        tempRepo := setupTestRepository(t)
        defer cleanupTestRepository(t, tempRepo)
        
        // No .sbs/input-source.json config (defaults to GitHub)
        mockGitHub := &mockGitHubClient{
            issues: map[int]*issue.Issue{
                123: {Number: 123, Title: "Legacy issue", State: "open"},
            },
        }
        
        // Run legacy command
        err := runStartWithGitHubMock([]string{"123"}, mockGitHub)
        require.NoError(t, err)
        
        // Verify legacy behavior
        sessions, err := config.LoadSessions()
        require.NoError(t, err)
        
        session := findSessionByIssueNumber(sessions, 123)
        require.NotNil(t, session)
        
        // Should have been migrated to namespaced format internally
        assert.Equal(t, "github", session.SourceType)
        assert.Equal(t, "github:123", session.NamespacedID)
        assert.Equal(t, 123, session.IssueNumber) // Preserved for compatibility
        
        // Branch should use legacy format for backward compatibility
        assert.Equal(t, "issue-123-legacy-issue", session.Branch)
    })
    
    t.Run("existing_sessions_still_work", func(t *testing.T) {
        // Test that existing sessions created before this feature still work
        legacySession := SessionMetadata{
            IssueNumber:    456,
            IssueTitle:     "Existing session",
            Branch:         "issue-456-existing-session",
            WorktreePath:   "/path/to/worktree",
            TmuxSession:    "work-issue-456",
            Status:         "active",
            // Missing SourceType and NamespacedID (legacy format)
        }
        
        sessions := []SessionMetadata{legacySession}
        err := config.SaveSessions(sessions)
        require.NoError(t, err)
        
        // Load sessions (should trigger migration)
        loadedSessions, err := config.LoadSessions()
        require.NoError(t, err)
        
        assert.Len(t, loadedSessions, 1)
        migrated := loadedSessions[0]
        
        // Should have been migrated
        assert.Equal(t, "github", migrated.SourceType)
        assert.Equal(t, "github:456", migrated.NamespacedID)
        assert.Equal(t, 456, migrated.IssueNumber) // Preserved
    })
}

func TestStartCommand_LegacyConfigCompatibility(t *testing.T) {
    t.Run("no_input_source_config", func(t *testing.T) {
        // Projects without .sbs/input-source.json should default to GitHub
        tempRepo := setupTestRepository(t)
        defer cleanupTestRepository(t, tempRepo)
        
        // Explicitly don't create .sbs/input-source.json
        
        // Should default to GitHub input source
        source, err := inputsource.LoadProjectInputSource(tempRepo.Root)
        require.NoError(t, err)
        assert.Equal(t, "github", source.GetType())
    })
}
```

**Priority**: Critical - Must not break existing workflows
**TDD Phase**: Implement early and run frequently

### 4.2 Branch Naming Compatibility Tests

**File**: `pkg/inputsource/workitem_test.go` (additions)

```go
func TestWorkItem_BranchNamingCompatibility(t *testing.T) {
    t.Run("github_legacy_format", func(t *testing.T) {
        // GitHub issues should maintain backward-compatible branch names for legacy sessions
        item := &WorkItem{
            Source: "github",
            ID:     "123",
            Title:  "Fix authentication bug",
        }
        
        // For legacy compatibility, GitHub branches should omit source prefix
        // when created from legacy numeric input
        legacyBranch := item.GetLegacyBranchName()
        assert.Equal(t, "issue-123-fix-authentication-bug", legacyBranch)
        
        // New namespaced format includes source
        namespacedBranch := item.GetBranchName()
        assert.Equal(t, "issue-github-123-fix-authentication-bug", namespacedBranch)
    })
    
    t.Run("new_sources_use_namespaced_format", func(t *testing.T) {
        // Non-GitHub sources should always use namespaced format
        item := &WorkItem{
            Source: "test",
            ID:     "quick",
            Title:  "Quick development test",
        }
        
        branch := item.GetBranchName()
        assert.Equal(t, "issue-test-quick-quick-development-test", branch)
        
        // No legacy format for non-GitHub sources
        assert.Equal(t, branch, item.GetLegacyBranchName())
    })
}
```

**Priority**: High - Ensures smooth migration
**TDD Phase**: Implement with WorkItem structure

## 5. Edge Cases and Error Handling

### 5.1 Configuration Error Tests

**File**: `pkg/config/inputsource_config_test.go` (additions)

```go
func TestInputSourceConfig_ErrorScenarios(t *testing.T) {
    t.Run("corrupted_config_file", func(t *testing.T) {
        tempDir := t.TempDir()
        sbsDir := filepath.Join(tempDir, ".sbs")
        require.NoError(t, os.MkdirAll(sbsDir, 0755))
        
        // Create corrupted JSON
        configPath := filepath.Join(sbsDir, "input-source.json")
        require.NoError(t, os.WriteFile(configPath, []byte(`{corrupted json`), 0644))
        
        config, err := LoadInputSourceConfig(tempDir)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to parse")
        assert.Nil(t, config)
    })
    
    t.Run("permission_denied", func(t *testing.T) {
        if runtime.GOOS == "windows" {
            t.Skip("Permission tests not reliable on Windows")
        }
        
        tempDir := t.TempDir()
        sbsDir := filepath.Join(tempDir, ".sbs")
        require.NoError(t, os.MkdirAll(sbsDir, 0755))
        
        configPath := filepath.Join(sbsDir, "input-source.json")
        require.NoError(t, os.WriteFile(configPath, []byte(`{"type": "test"}`), 0000)) // No read permissions
        
        config, err := LoadInputSourceConfig(tempDir)
        assert.Error(t, err)
        assert.Nil(t, config)
    })
    
    t.Run("unsupported_input_source_type", func(t *testing.T) {
        tempDir := t.TempDir()
        sbsDir := filepath.Join(tempDir, ".sbs")
        require.NoError(t, os.MkdirAll(sbsDir, 0755))
        
        configData := `{"type": "unsupported-source"}`
        configPath := filepath.Join(sbsDir, "input-source.json")
        require.NoError(t, os.WriteFile(configPath, []byte(configData), 0644))
        
        factory := NewInputSourceFactory()
        config, _ := LoadInputSourceConfig(tempDir)
        
        source, err := factory.Create(config)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "unsupported input source type")
        assert.Nil(t, source)
    })
}
```

**Priority**: Medium - Important for robustness
**TDD Phase**: Implement after core functionality

### 5.2 Input Validation Edge Cases

**File**: `pkg/inputsource/workitem_test.go` (additions)

```go
func TestWorkItem_EdgeCases(t *testing.T) {
    t.Run("special_characters_in_id", func(t *testing.T) {
        tests := []struct {
            name      string
            fullID    string
            expectErr bool
        }{
            {"jira_format", "jira:PROJ-123", false},
            {"underscores", "test:quick_test", false},
            {"spaces_invalid", "test:quick test", true},
            {"multiple_colons", "test:quick:extra", true},
            {"unicode_characters", "test:café", true},
        }
        
        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                item, err := ParseWorkItemID(tt.fullID)
                if tt.expectErr {
                    assert.Error(t, err)
                    assert.Nil(t, item)
                } else {
                    assert.NoError(t, err)
                    assert.NotNil(t, item)
                }
            })
        }
    })
    
    t.Run("extremely_long_titles", func(t *testing.T) {
        longTitle := strings.Repeat("a", 500) // Very long title
        item := &WorkItem{
            Source: "test",
            ID:     "long",
            Title:  longTitle,
        }
        
        branch := item.GetBranchName()
        // Branch name should be truncated to reasonable length
        assert.Less(t, len(branch), 200) // Git has practical limits
        assert.Contains(t, branch, "issue-test-long")
    })
    
    t.Run("empty_and_whitespace_titles", func(t *testing.T) {
        tests := []struct {
            title           string
            expectedBranch  string
        }{
            {"", "issue-test-empty"},
            {"   ", "issue-test-empty"},
            {"   Mixed   Spaces   ", "issue-test-empty-mixed-spaces"},
        }
        
        for _, tt := range tests {
            item := &WorkItem{
                Source: "test",
                ID:     "empty",
                Title:  tt.title,
            }
            
            branch := item.GetBranchName()
            assert.Contains(t, branch, "issue-test-empty")
        }
    })
}
```

**Priority**: Medium - Edge case handling
**TDD Phase**: Implement after core functionality

## 6. Test Data Setup and Fixtures

### 6.1 Test Fixtures

**File**: `pkg/inputsource/test_helpers.go`

```go
// Test data fixtures for consistent testing
var (
    TestWorkItems = map[string]*WorkItem{
        "github:123": {
            Source: "github",
            ID:     "123",
            Title:  "Fix authentication bug",
            State:  "open",
            URL:    "https://github.com/test/repo/issues/123",
        },
        "test:quick": {
            Source: "test",
            ID:     "quick",
            Title:  "Quick development test",
            State:  "open",
            URL:    "", // Test items don't have URLs
        },
        "jira:PROJ-456": {
            Source: "jira",
            ID:     "PROJ-456",
            Title:  "JIRA integration task",
            State:  "open",
            URL:    "https://company.atlassian.net/browse/PROJ-456",
        },
    }
    
    TestSessions = []SessionMetadata{
        {
            IssueNumber:    123,
            SourceType:     "github",
            NamespacedID:   "github:123",
            IssueTitle:     "Fix authentication bug",
            Branch:         "issue-123-fix-authentication-bug",
            Status:         "active",
        },
        {
            SourceType:     "test",
            NamespacedID:   "test:quick",
            IssueTitle:     "Quick development test",
            Branch:         "issue-test-quick-quick-development-test",
            Status:         "active",
        },
    }
)

// Helper functions for test setup
func SetupTestRepository(t *testing.T) *TestRepository {
    tempDir := t.TempDir()
    
    // Initialize git repository
    cmd := exec.Command("git", "init")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())
    
    // Create initial commit
    cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
    cmd.Dir = tempDir
    require.NoError(t, cmd.Run())
    
    return &TestRepository{
        Root: tempDir,
        Name: "test-repo",
    }
}

func ConfigureTestInputSource(t *testing.T, repoRoot string) {
    sbsDir := filepath.Join(repoRoot, ".sbs")
    require.NoError(t, os.MkdirAll(sbsDir, 0755))
    
    config := &InputSourceConfig{
        Type: "test",
        Settings: map[string]interface{}{
            "description": "Test project for development",
        },
    }
    
    err := SaveInputSourceConfig(repoRoot, config)
    require.NoError(t, err)
}

func ConfigureGitHubInputSource(t *testing.T, repoRoot string) {
    sbsDir := filepath.Join(repoRoot, ".sbs")
    require.NoError(t, os.MkdirAll(sbsDir, 0755))
    
    config := &InputSourceConfig{
        Type: "github",
        Settings: map[string]interface{}{
            "repository": "test/repo",
        },
    }
    
    err := SaveInputSourceConfig(repoRoot, config)
    require.NoError(t, err)
}

// Mock implementations for testing
type MockInputSource struct {
    WorkItems map[string]*WorkItem
    ListResult []*WorkItem
    GetCalls  []string
    ListCalls []ListCall
}

type ListCall struct {
    SearchQuery string
    Limit      int
}

func (m *MockInputSource) GetWorkItem(id string) (*WorkItem, error) {
    m.GetCalls = append(m.GetCalls, id)
    
    if item, exists := m.WorkItems[id]; exists {
        return item, nil
    }
    
    return nil, fmt.Errorf("work item not found: %s", id)
}

func (m *MockInputSource) ListWorkItems(searchQuery string, limit int) ([]*WorkItem, error) {
    m.ListCalls = append(m.ListCalls, ListCall{
        SearchQuery: searchQuery,
        Limit:      limit,
    })
    
    return m.ListResult, nil
}

func (m *MockInputSource) GetType() string {
    return "mock"
}
```

**Priority**: Medium - Essential for comprehensive testing
**TDD Phase**: Implement early for test infrastructure

### 6.2 Integration Test Fixtures

**File**: `testdata/inputsource/`

```
testdata/
├── inputsource/
│   ├── github-config.json          # Sample GitHub configuration
│   ├── test-config.json            # Sample test configuration
│   ├── invalid-config.json         # Invalid configuration for error testing
│   ├── legacy-sessions.json        # Legacy session metadata
│   └── mixed-sessions.json         # Mixed legacy and new sessions
└── repositories/
    ├── github-project/             # Test project with GitHub config
    │   └── .sbs/
    │       └── input-source.json
    └── test-project/               # Test project with test config
        └── .sbs/
            └── input-source.json
```

**Priority**: Low - Nice to have for realistic testing
**TDD Phase**: Implement for comprehensive test coverage

## 7. Manual Testing Procedures

### 7.1 Step-by-Step Validation Workflows

#### Workflow 1: New Test Input Source
```bash
# Setup test project
mkdir test-project && cd test-project
git init && git commit --allow-empty -m "Initial"

# Configure for test input source
mkdir .sbs
cat > .sbs/input-source.json << 'EOF'
{
    "type": "test",
    "settings": {
        "description": "Development testing project"
    }
}
EOF

# Test commands
sbs start test:quick              # Should work with namespaced ID
sbs start quick                   # Should work with short ID (test source context)
sbs list                         # Should show test session
sbs attach test:quick            # Should attach to test session
sbs stop test:quick              # Should stop test session
```

#### Workflow 2: Backward Compatibility
```bash
# Setup existing project (no input source config)
mkdir legacy-project && cd legacy-project
git init && git commit --allow-empty -m "Initial"

# Legacy commands should still work
sbs start 123                    # Should default to GitHub, fetch issue #123
sbs list                         # Should show GitHub session
sbs attach 123                   # Should attach to GitHub session
```

#### Workflow 3: Mixed Environment
```bash
# Project 1: Test source
cd test-project
sbs start test:hooks             # Test session

# Project 2: GitHub source  
cd ../legacy-project
sbs start 456                    # GitHub session

# Global list should show both
sbs list                         # Mixed sources in output
```

### 7.2 Error Condition Testing

#### Test Invalid Configurations
```bash
# Invalid JSON in config
echo '{invalid json}' > .sbs/input-source.json
sbs start test:quick             # Should show helpful error

# Unsupported source type
echo '{"type": "unknown"}' > .sbs/input-source.json
sbs start anything               # Should show supported types

# Permission issues
chmod 000 .sbs/input-source.json
sbs start test:quick             # Should handle permission error gracefully
```

#### Test Invalid Work Item IDs
```bash
# Invalid namespaced format
sbs start invalid-format         # Should show format help
sbs start test:nonexistent       # Should show available test items
sbs start github:invalid         # Should show GitHub error
```

**Priority**: High - Critical for user experience
**Manual Testing Phase**: After integration tests pass

## 8. Performance Considerations

### 8.1 Performance Test Cases

**File**: `pkg/inputsource/performance_test.go`

```go
func BenchmarkWorkItem_ParseID(b *testing.B) {
    testIDs := []string{
        "github:123",
        "test:quick",
        "jira:PROJ-456",
        "123", // Legacy format
    }
    
    for i := 0; i < b.N; i++ {
        for _, id := range testIDs {
            ParseWorkItemID(id)
        }
    }
}

func BenchmarkInputSourceFactory_Create(b *testing.B) {
    factory := NewInputSourceFactory()
    config := &InputSourceConfig{Type: "test"}
    
    for i := 0; i < b.N; i++ {
        source, _ := factory.Create(config)
        _ = source
    }
}

func TestInputSource_ResponseTime(t *testing.T) {
    // Test that input source operations complete within reasonable time
    source := NewTestInputSource()
    
    start := time.Now()
    items, err := source.ListWorkItems("", 100)
    duration := time.Since(start)
    
    require.NoError(t, err)
    assert.NotEmpty(t, items)
    assert.Less(t, duration, 100*time.Millisecond) // Should be very fast for test source
}
```

### 8.2 Memory Usage Tests

```go
func TestInputSource_MemoryUsage(t *testing.T) {
    // Test that input source doesn't leak memory
    factory := NewInputSourceFactory()
    
    var m1, m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Create and use many input sources
    for i := 0; i < 1000; i++ {
        source, _ := factory.Create(&InputSourceConfig{Type: "test"})
        _, _ = source.ListWorkItems("", 10)
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    // Memory growth should be reasonable
    memoryGrowth := m2.Alloc - m1.Alloc
    assert.Less(t, memoryGrowth, uint64(10*1024*1024)) // Less than 10MB
}
```

**Priority**: Low - Performance optimization
**Testing Phase**: After functional tests pass

## 9. Test Execution Strategy

### 9.1 Test Organization and Execution

#### Unit Tests (Fast)
```bash
# Run all unit tests
go test ./pkg/inputsource/... -v

# Run with coverage
go test ./pkg/inputsource/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test suites
go test ./pkg/inputsource/workitem_test.go -v
go test ./pkg/inputsource/factory_test.go -v
```

#### Integration Tests (Slower)
```bash
# Run integration tests with build tag
go test -tags=integration ./... -v

# Run integration tests with external dependencies
INTEGRATION_TESTS=1 go test ./cmd/... -v
```

#### Full Test Suite
```bash
# Complete test run
make test                        # All unit tests
make test-integration           # Integration tests
make test-all                   # Unit + integration tests
```

### 9.2 Continuous Integration Configuration

```yaml
# .github/workflows/test.yml (addition)
- name: Test Input Source Feature
  run: |
    go test ./pkg/inputsource/... -v -cover
    go test ./pkg/config/... -v -cover
    INTEGRATION_TESTS=1 go test ./cmd/... -v
```

### 9.3 Test Data Management

#### Test Environment Setup
```bash
# Setup script for test environment
#!/bin/bash
# scripts/setup-test-env.sh

# Create test repositories
mkdir -p testdata/repositories/github-project/.sbs
mkdir -p testdata/repositories/test-project/.sbs

# Create test configurations
cat > testdata/repositories/github-project/.sbs/input-source.json << 'EOF'
{"type": "github"}
EOF

cat > testdata/repositories/test-project/.sbs/input-source.json << 'EOF'
{"type": "test"}
EOF

# Initialize git repositories
cd testdata/repositories/github-project
git init && git commit --allow-empty -m "Initial"

cd ../test-project
git init && git commit --allow-empty -m "Initial"
```

## 10. Success Criteria and Definition of Done

### 10.1 Test Coverage Targets
- **Unit Tests**: 90% line coverage for new code
- **Integration Tests**: All major workflows covered
- **Edge Cases**: All error conditions tested
- **Backward Compatibility**: All existing functionality preserved

### 10.2 Quality Gates
1. **All unit tests pass** ✅
2. **All integration tests pass** ✅
3. **No regression in existing functionality** ✅
4. **Performance benchmarks meet requirements** ✅
5. **Manual test scenarios execute successfully** ✅

### 10.3 Documentation and Examples
- Comprehensive test documentation (this plan)
- Example configurations in testdata/
- Error message testing for user experience
- Performance characteristics documented

## 11. Risk Mitigation

### 11.1 High-Risk Areas
1. **Session Migration**: Existing sessions must continue working
2. **GitHub Integration**: No changes to GitHub API interaction
3. **Git Operations**: Branch naming changes must not break existing workflows
4. **Configuration Loading**: Robust error handling for invalid configs

### 11.2 Mitigation Strategies
- Extensive backward compatibility testing
- Migration testing with real legacy session data
- Rollback capability if issues discovered
- Staged deployment with feature flags if needed

## Implementation Priority Order

1. **Phase 1**: Core interfaces and WorkItem structure
2. **Phase 2**: TestInputSource implementation
3. **Phase 3**: Configuration system and factory
4. **Phase 4**: GitHubInputSource wrapper
5. **Phase 5**: Session migration functionality
6. **Phase 6**: Start command integration
7. **Phase 7**: Comprehensive integration testing
8. **Phase 8**: Manual testing and edge cases

This testing plan ensures comprehensive coverage of the pluggable input sources feature while maintaining backward compatibility and following test-driven development practices.