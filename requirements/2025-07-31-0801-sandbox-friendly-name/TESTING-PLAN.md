# Testing Plan: Sandbox-Friendly Issue Title Environment Variable

## Overview

This document outlines a comprehensive testing strategy for implementing the SBS_TITLE environment variable feature in the SBS CLI application. The feature extends the application to generate sandbox-friendly names from GitHub issue titles and pass them as environment variables to tmux session operations.

## Testing Strategy

### Test-Driven Development Approach

1. **Unit Tests First**: Write comprehensive unit tests for each component before implementation
2. **Integration Tests**: Test component interactions and data flow
3. **End-to-End Tests**: Verify complete workflows from user commands to environment variable availability
4. **Edge Case Coverage**: Test boundary conditions, error scenarios, and fallback behaviors
5. **Performance Testing**: Ensure no significant performance regression

## 1. Unit Tests

### 1.1 pkg/repo/manager.go - SanitizeName Method Extension

**Test File**: `pkg/repo/manager_test.go`

#### Test Functions:

```go
func TestManager_SanitizeName_WithMaxLength(t *testing.T)
func TestManager_SanitizeName_AlphanumericOnly(t *testing.T) 
func TestManager_SanitizeName_LowercaseConversion(t *testing.T)
func TestManager_SanitizeName_HyphenReplacement(t *testing.T)
func TestManager_SanitizeName_LeadingTrailingHyphens(t *testing.T)
func TestManager_SanitizeName_UnicodeCharacters(t *testing.T)
func TestManager_SanitizeName_EmptyString(t *testing.T)
func TestManager_SanitizeName_WordBoundaryTruncation(t *testing.T)
func TestManager_SanitizeName_BackwardCompatibility(t *testing.T)
```

#### Test Cases:

| Test Function | Input | Max Length | Expected Output | Description |
|---------------|-------|------------|-----------------|-------------|
| `TestManager_SanitizeName_WithMaxLength` | "Fix user authentication bug in login system" | 32 | "fix-user-authentication-bug-in" | Basic length limiting |
| `TestManager_SanitizeName_AlphanumericOnly` | "Fix café login (UTF-8 encoding)" | 32 | "fix-cafe-login-utf-8-encoding" | Special character sanitization |
| `TestManager_SanitizeName_LowercaseConversion` | "Fix User Authentication BUG" | 32 | "fix-user-authentication-bug" | Case normalization |
| `TestManager_SanitizeName_HyphenReplacement` | "Fix  multiple   spaces & symbols!" | 32 | "fix-multiple-spaces-symbols" | Multiple separator handling |
| `TestManager_SanitizeName_LeadingTrailingHyphens` | "  Fix bug  " | 32 | "fix-bug" | Trim edge hyphens |
| `TestManager_SanitizeName_UnicodeCharacters` | "修复登录错误" | 32 | "fix-login-error" | Unicode to ASCII fallback |
| `TestManager_SanitizeName_EmptyString` | "" | 32 | "" | Empty input handling |
| `TestManager_SanitizeName_WordBoundaryTruncation` | "Fix authentication system completely" | 20 | "fix-authentication" | Smart truncation at word boundaries |
| `TestManager_SanitizeName_BackwardCompatibility` | "myproject-name" | 0 | "myproject-name" | Existing behavior when maxLength is 0 |

### 1.2 pkg/config/config.go - SessionMetadata Extension

**Test File**: `pkg/config/config_test.go`

#### Test Functions:

```go
func TestSessionMetadata_FriendlyTitleField(t *testing.T)
func TestSessionMetadata_JSONSerialization(t *testing.T)
func TestSessionMetadata_JSONDeserialization(t *testing.T)
func TestSessionMetadata_BackwardCompatibility(t *testing.T)
func TestSessionMetadata_DefaultFriendlyTitle(t *testing.T)
```

#### Test Cases:

| Test Function | Description | Expected Behavior |
|---------------|-------------|-------------------|
| `TestSessionMetadata_FriendlyTitleField` | Verify FriendlyTitle field exists | Field accessible and modifiable |
| `TestSessionMetadata_JSONSerialization` | Test JSON marshaling with FriendlyTitle | Field included in JSON output |
| `TestSessionMetadata_JSONDeserialization` | Test JSON unmarshaling | Missing field doesn't break deserialization |
| `TestSessionMetadata_BackwardCompatibility` | Load old JSON without FriendlyTitle | Graceful handling of missing field |
| `TestSessionMetadata_DefaultFriendlyTitle` | Test default value behavior | Empty string or appropriate default |

### 1.3 pkg/tmux/manager.go - Environment Variable Support

**Test File**: `pkg/tmux/manager_test.go`

#### Test Functions:

```go
func TestManager_CreateSession_WithEnvironment(t *testing.T)
func TestManager_CreateSession_WithoutEnvironment(t *testing.T)
func TestManager_AttachToSession_WithEnvironment(t *testing.T)
func TestManager_AttachToSession_WithoutEnvironment(t *testing.T)
func TestManager_ExecuteCommand_WithEnvironment(t *testing.T)
func TestManager_ExecuteCommand_WithoutEnvironment(t *testing.T)
func TestManager_EnvironmentVariableParsing(t *testing.T)
func TestManager_EnvironmentVariableEscaping(t *testing.T)
func TestManager_BackwardCompatibility(t *testing.T)
```

#### Test Cases:

| Test Function | Input Environment | Expected Behavior |
|---------------|-------------------|-------------------|
| `TestManager_CreateSession_WithEnvironment` | `map[string]string{"SBS_TITLE": "test-title"}` | tmux command includes environment variables |
| `TestManager_CreateSession_WithoutEnvironment` | `map[string]string{}` or `nil` | Default behavior, no environment changes |
| `TestManager_AttachToSession_WithEnvironment` | `map[string]string{"SBS_TITLE": "test-title"}` | Environment variables set during attach |
| `TestManager_ExecuteCommand_WithEnvironment` | `map[string]string{"SBS_TITLE": "test-title", "OTHER": "value"}` | Multiple environment variables handled |
| `TestManager_EnvironmentVariableParsing` | Environment with special characters | Proper escaping and quoting |
| `TestManager_BackwardCompatibility` | No environment parameter | Existing functionality unchanged |

### 1.4 Helper Functions and Utilities

**Test File**: `pkg/helpers/friendly_title_test.go`

#### Test Functions:

```go
func TestGenerateFriendlyTitle_WithTitle(t *testing.T)
func TestGenerateFriendlyTitle_WithoutTitle(t *testing.T)
func TestGenerateFriendlyTitle_EmptyTitle(t *testing.T)
func TestCreateTmuxEnvironment(t *testing.T)
func TestCreateTmuxEnvironment_EmptyTitle(t *testing.T)
```

#### Test Cases:

| Test Function | Repo Name | Issue Number | Issue Title | Expected Output |
|---------------|-----------|--------------|-------------|-----------------|
| `TestGenerateFriendlyTitle_WithTitle` | "myproject" | 123 | "Fix user authentication bug" | "fix-user-authentication-bug" |
| `TestGenerateFriendlyTitle_WithoutTitle` | "myproject" | 123 | "" | "myproject-issue-123" |
| `TestGenerateFriendlyTitle_EmptyTitle` | "myproject" | 123 | "   " | "myproject-issue-123" |

## 2. Integration Tests

### 2.1 Command Integration Tests

**Test File**: `cmd/start_integration_test.go`

#### Test Functions:

```go
func TestStartCommand_FriendlyTitleGeneration(t *testing.T)
func TestStartCommand_EnvironmentVariableFlow(t *testing.T)
func TestStartCommand_SessionMetadataPersistence(t *testing.T)
func TestStartCommand_GitHubAPIIntegration(t *testing.T)
func TestStartCommand_FallbackBehavior(t *testing.T)
```

**Test File**: `cmd/attach_integration_test.go`

#### Test Functions:

```go
func TestAttachCommand_FriendlyTitleRetrieval(t *testing.T)
func TestAttachCommand_EnvironmentVariableFlow(t *testing.T)
func TestAttachCommand_MissingMetadata(t *testing.T)
```

### 2.2 Data Flow Integration Tests

#### Test Scenarios:

1. **Complete Workflow Test**:
   - Start session with GitHub issue title
   - Verify friendly title generation
   - Verify environment variable setting
   - Verify metadata persistence
   - Attach to session
   - Verify environment variable availability

2. **Fallback Workflow Test**:
   - Simulate GitHub API failure
   - Verify fallback title generation
   - Verify environment variable setting with fallback

3. **Session Recovery Test**:
   - Create session with friendly title
   - Simulate application restart
   - Attach to existing session
   - Verify friendly title retrieval from metadata

## 3. End-to-End Tests

### 3.1 Full Workflow Tests

**Test File**: `e2e/sbs_title_test.go`

#### Test Functions:

```go
func TestE2E_StartSessionWithSBSTitle(t *testing.T)
func TestE2E_AttachSessionWithSBSTitle(t *testing.T)
func TestE2E_CommandExecutionWithSBSTitle(t *testing.T)
func TestE2E_FallbackTitleGeneration(t *testing.T)
```

#### Test Scenarios:

1. **Full Start Workflow**:
   ```bash
   # Setup test repository and issue
   sbs start 123
   # Verify tmux session has SBS_TITLE environment variable
   # Verify work-issue.sh receives SBS_TITLE
   ```

2. **Full Attach Workflow**:
   ```bash
   # Start session
   sbs start 123
   # Detach from session
   # Attach to session
   sbs attach 123
   # Verify SBS_TITLE is available in attached session
   ```

3. **Command Execution Test**:
   ```bash
   # Start session
   sbs start 123
   # Execute command that uses SBS_TITLE
   # Verify environment variable is available to command
   ```

### 3.2 Environment Variable Verification

#### Test Utilities:

```go
func verifyEnvironmentVariable(t *testing.T, sessionName, expectedValue string) {
    // Use tmux show-environment to verify SBS_TITLE is set
    cmd := exec.Command("tmux", "show-environment", "-t", sessionName, "SBS_TITLE")
    output, err := cmd.Output()
    require.NoError(t, err)
    assert.Contains(t, string(output), fmt.Sprintf("SBS_TITLE=%s", expectedValue))
}

func executeCommandInSession(t *testing.T, sessionName, command string) string {
    // Execute command in tmux session and capture output
    // Return output for verification
}
```

## 4. Edge Case Tests

### 4.1 Boundary Conditions

#### Test Functions:

```go
func TestEdgeCase_VeryLongIssueTitle(t *testing.T)
func TestEdgeCase_IssueWithOnlySpecialCharacters(t *testing.T)
func TestEdgeCase_IssueWithOnlyWhitespace(t *testing.T)
func TestEdgeCase_UnicodeOnlyTitle(t *testing.T)
func TestEdgeCase_ExactlyMaxLength(t *testing.T)
func TestEdgeCase_EmptyRepositoryName(t *testing.T)
```

#### Test Cases:

| Test Function | Input | Expected Behavior |
|---------------|-------|-------------------|
| `TestEdgeCase_VeryLongIssueTitle` | 200-character title | Truncated to 32 chars at word boundary |
| `TestEdgeCase_IssueWithOnlySpecialCharacters` | "!@#$%^&*()" | Falls back to repo-issue-number format |
| `TestEdgeCase_IssueWithOnlyWhitespace` | "   \t\n   " | Falls back to repo-issue-number format |
| `TestEdgeCase_UnicodeOnlyTitle` | "修复错误" | Converted to ASCII or falls back |
| `TestEdgeCase_ExactlyMaxLength` | 32-character title | No truncation, preserved exactly |
| `TestEdgeCase_EmptyRepositoryName` | "" repo name | Graceful handling or default value |

### 4.2 Error Handling Tests

#### Test Functions:

```go
func TestErrorHandling_GitHubAPIFailure(t *testing.T)
func TestErrorHandling_TmuxSessionCreationFailure(t *testing.T)
func TestErrorHandling_MetadataPersistenceFailure(t *testing.T)
func TestErrorHandling_CorruptedMetadata(t *testing.T)
func TestErrorHandling_MissingTmuxSession(t *testing.T)
```

#### Test Scenarios:

1. **GitHub API Unavailable**:
   - Mock GitHub API failure
   - Verify fallback title generation
   - Verify session creation continues

2. **Tmux Command Failure**:
   - Mock tmux command failure
   - Verify graceful error handling
   - Verify appropriate error messages

3. **Metadata Corruption**:
   - Create corrupted session metadata
   - Verify graceful handling
   - Verify recovery mechanisms

## 5. Performance Tests

### 5.1 Performance Benchmarks

**Test File**: `performance/sbs_title_benchmark_test.go`

#### Benchmark Functions:

```go
func BenchmarkSanitizeName(b *testing.B)
func BenchmarkFriendlyTitleGeneration(b *testing.B)
func BenchmarkEnvironmentVariableCreation(b *testing.B)
func BenchmarkMetadataSerialization(b *testing.B)
```

#### Performance Criteria:

- `SanitizeName` with 32-char limit: < 1ms for typical titles
- Friendly title generation: < 5ms total
- Environment variable setup: < 10ms
- Metadata serialization: < 50ms

### 5.2 Memory Usage Tests

#### Test Functions:

```go
func TestMemoryUsage_LargeIssueTitle(t *testing.T)
func TestMemoryUsage_ManyEnvironmentVariables(t *testing.T)
func TestMemoryUsage_MetadataGrowth(t *testing.T)
```

## 6. Acceptance Criteria Test Coverage

### 6.1 AC1: Session Creation Test

```go
func TestAC1_SessionCreationWithSBSTitle(t *testing.T) {
    // Given: GitHub issue with title "Fix user authentication bug"
    mockIssue := &issue.Issue{
        Number: 123,
        Title:  "Fix user authentication bug",
        State:  "open",
    }
    
    // When: running `sbs start 123`
    // Then: tmux session is created with SBS_TITLE=fix-user-authentication-bug
    
    expectedTitle := "fix-user-authentication-bug"
    verifyEnvironmentVariable(t, sessionName, expectedTitle)
}
```

### 6.2 AC2: Session Attachment Test

```go
func TestAC2_SessionAttachmentWithSBSTitle(t *testing.T) {
    // Given: existing session for issue #123
    // When: running `sbs attach 123`
    // Then: attached session has SBS_TITLE environment variable set
}
```

### 6.3 AC3: Command Execution Test

```go
func TestAC3_CommandExecutionWithSBSTitle(t *testing.T) {
    // Given: active session with friendly title
    // When: executing commands via ExecuteCommand()
    // Then: commands have access to SBS_TITLE environment variable
}
```

### 6.4 AC4: Fallback Handling Test

```go
func TestAC4_FallbackHandling(t *testing.T) {
    // Given: GitHub API is unavailable for issue #123 in repo "myproject"
    // When: starting a session
    // Then: SBS_TITLE=myproject-issue-123 is set
}
```

### 6.5 AC5: Length Limitation Test

```go
func TestAC5_LengthLimitation(t *testing.T) {
    // Given: issue title exceeding 32 characters
    longTitle := "This is a very long issue title that exceeds the thirty-two character limit"
    // When: generating friendly name
    // Then: result is truncated to 32 characters with clean hyphen boundaries
}
```

### 6.6 AC6: Character Sanitization Test

```go
func TestAC6_CharacterSanitization(t *testing.T) {
    // Given: issue title "Fix café login (UTF-8 encoding)"
    input := "Fix café login (UTF-8 encoding)"
    // When: generating friendly name
    // Then: SBS_TITLE=fix-cafe-login-utf-8-encoding
    expected := "fix-cafe-login-utf-8-encoding"
}
```

## 7. Test Data and Fixtures

### 7.1 Test Issue Titles

```go
var testIssueTitles = map[string]struct {
    input    string
    expected string
}{
    "basic_title":           {"Fix user authentication", "fix-user-authentication"},
    "with_special_chars":    {"Fix café login (UTF-8)", "fix-cafe-login-utf-8"},
    "very_long_title":       {"This is a very long issue title that should be truncated properly", "this-is-a-very-long-issue-title"},
    "unicode_title":         {"修复用户认证错误", "fix-user-auth-error"}, // With fallback
    "numbers_and_chars":     {"Fix API v2.1 integration", "fix-api-v2-1-integration"},
    "multiple_spaces":       {"Fix   multiple    spaces", "fix-multiple-spaces"},
    "leading_trailing":      "  Fix leading trailing  ", "fix-leading-trailing"},
    "only_special_chars":    {"!@#$%^&*()", ""}, // Should trigger fallback
    "empty_title":           {"", ""}, // Should trigger fallback
    "whitespace_only":       {"   \t\n   ", ""}, // Should trigger fallback
}
```

### 7.2 Test Repository Names

```go
var testRepositoryNames = []string{
    "myproject",
    "my-awesome-project",
    "project123",
    "Project_Name",
    "café-repo", // With special characters
}
```

### 7.3 Mock Session Metadata

```go
func createTestSessionMetadata(issueNumber int, friendlyTitle string) *config.SessionMetadata {
    return &config.SessionMetadata{
        IssueNumber:    issueNumber,
        IssueTitle:     "Test Issue Title",
        FriendlyTitle:  friendlyTitle,
        Branch:         fmt.Sprintf("issue-%d-test", issueNumber),
        WorktreePath:   "/tmp/test-worktree",
        TmuxSession:    fmt.Sprintf("work-issue-test-%d", issueNumber),
        SandboxName:    fmt.Sprintf("work-issue-test-%d", issueNumber),
        RepositoryName: "test-repo",
        RepositoryRoot: "/tmp/test-repo",
        CreatedAt:      "2025-07-31T08:00:00Z",
        LastActivity:   "2025-07-31T08:00:00Z",
        Status:         "active",
    }
}
```

## 8. Test Environment Setup

### 8.1 Prerequisites

- Go 1.19+ for testing
- tmux installed and available
- git repository for testing
- Mock GitHub CLI or real gh command for integration tests

### 8.2 Test Categories

#### Unit Tests
```bash
go test ./pkg/repo -v
go test ./pkg/config -v
go test ./pkg/tmux -v
go test ./pkg/helpers -v
```

#### Integration Tests
```bash
go test ./cmd -tags=integration -v
go test ./... -tags=integration -v
```

#### End-to-End Tests
```bash
go test ./e2e -tags=e2e -v
```

#### Performance Tests
```bash
go test ./performance -bench=. -v
```

### 8.3 Test Isolation

- Use temporary directories for worktrees
- Use unique tmux session names per test
- Clean up resources after each test
- Mock external dependencies (GitHub API, file system)

### 8.4 Continuous Integration

#### Test Pipeline

1. **Unit Tests**: Run on every commit
2. **Integration Tests**: Run on pull requests
3. **E2E Tests**: Run on main branch
4. **Performance Tests**: Run weekly or on major changes

#### Test Coverage Requirements

- Unit test coverage: > 90%
- Integration test coverage: > 80%
- Critical path coverage: 100%

## 9. Test Implementation Order

1. **Phase 1**: Unit tests for core functionality
   - pkg/repo/manager.go SanitizeName extension
   - pkg/config/config.go SessionMetadata extension
   - Helper functions for friendly title generation

2. **Phase 2**: Unit tests for tmux integration
   - pkg/tmux/manager.go environment variable support
   - Environment variable handling and escaping

3. **Phase 3**: Integration tests
   - Command-level integration
   - Data flow between components
   - Error handling scenarios

4. **Phase 4**: End-to-end tests
   - Full workflow testing
   - Real tmux session testing
   - Environment variable verification

5. **Phase 5**: Edge cases and performance
   - Boundary condition testing
   - Performance benchmarking
   - Memory usage verification

## 10. Success Criteria

### Functional Success
- All acceptance criteria tests pass
- No regression in existing functionality
- All edge cases handled gracefully
- Fallback mechanisms work correctly

### Quality Success
- Test coverage > 90% for new code
- No memory leaks in performance tests
- All error conditions properly tested
- Documentation updated with test examples

### Integration Success
- End-to-end workflows complete successfully
- Environment variables available in tmux sessions
- Session metadata persistence works correctly
- Backward compatibility maintained

This comprehensive testing plan ensures that the SBS_TITLE environment variable feature is thoroughly tested from unit level through end-to-end scenarios, with particular attention to edge cases, error handling, and performance characteristics.