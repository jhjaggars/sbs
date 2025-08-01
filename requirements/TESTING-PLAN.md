# Testing Plan: Remove Podman Support and Make Sandbox Command Mandatory

## Overview

This document outlines a comprehensive testing strategy for GitHub issue #17: "Remove Podman support and make sandbox command mandatory". Based on analysis of the current codebase, SBS already uses the `sandbox` command exclusively with no Podman-related code found. This testing plan focuses on:

1. **Strengthening sandbox validation** - Ensure robust error handling when sandbox is unavailable
2. **Improving error messages** - Provide clear, actionable error messages for missing dependencies
3. **Preventing alternative runtime configurations** - Ensure no configuration options exist for other container runtimes
4. **Comprehensive test coverage** - Validate all sandbox-dependent functionality

## Current Implementation Analysis

### Existing Architecture
- **Sandbox Manager**: `pkg/sandbox/manager.go` - Handles all sandbox operations
- **Validation System**: `pkg/validation/tools.go` - Validates required tools including sandbox
- **Command Integration**: All commands (start, stop, clean) use sandbox exclusively
- **No Podman Code**: No references to Podman, Docker, or alternative container runtimes found

### Current Validation
```go
// pkg/validation/tools.go
func CheckRequiredTools() error {
    // ... validates tmux, git, gh, and sandbox
    if err := sandbox.CheckSandboxInstalled(); err != nil {
        errors = append(errors, err.Error())
    }
}
```

## Testing Strategy

### Test-Driven Development Approach

1. **Unit Tests**: Test individual component validation and error handling
2. **Integration Tests**: Test tool validation in realistic scenarios  
3. **Error Scenario Tests**: Test various failure modes and error messages
4. **Configuration Tests**: Prevent introduction of alternative runtime configs
5. **End-to-End Tests**: Test complete workflows with missing dependencies

## 1. Unit Tests

### 1.1 Enhanced Sandbox Validation Tests

**Test File**: `pkg/sandbox/manager_test.go`

#### New Test Functions:

```go
func TestCheckSandboxInstalled_ErrorMessages(t *testing.T)
func TestCheckSandboxInstalled_PathValidation(t *testing.T) 
func TestCheckSandboxInstalled_VersionCompatibility(t *testing.T)
func TestSandboxManager_GracefulDegradation(t *testing.T)
func TestSandboxManager_ErrorPropagation(t *testing.T)
```

#### Test Cases:

| Test Function | Scenario | Expected Behavior | Validation Criteria |
|---------------|----------|-------------------|-------------------|
| `TestCheckSandboxInstalled_ErrorMessages` | Sandbox not in PATH | Clear error with installation instructions | Error message contains "sandbox command not found" and helpful guidance |
| `TestCheckSandboxInstalled_PathValidation` | Sandbox executable not executable | Permission error with clear guidance | Error mentions permission issues and suggests solutions |
| `TestCheckSandboxInstalled_VersionCompatibility` | Old/incompatible sandbox version | Version warning with upgrade instructions | Error includes version info and upgrade guidance |
| `TestSandboxManager_GracefulDegradation` | Sandbox operations fail | Consistent error handling across all operations | All sandbox operations fail gracefully with meaningful messages |
| `TestSandboxManager_ErrorPropagation` | Various sandbox command failures | Error context preserved through call stack | Original error information maintained through all layers |

#### Specific Test Implementations:

```go
func TestCheckSandboxInstalled_ErrorMessages(t *testing.T) {
    tests := []struct {
        name          string
        mockExitCode  int
        mockError     string
        expectedError string
        errorContains []string
    }{
        {
            name:          "command_not_found",
            mockExitCode:  127,
            mockError:     "command not found",
            expectedError: "sandbox command not found. Please ensure sandbox is installed and in PATH",
            errorContains: []string{"sandbox command not found", "installed", "PATH"},
        },
        {
            name:          "permission_denied",
            mockExitCode:  126,
            mockError:     "permission denied",
            expectedError: "sandbox command not executable. Please check file permissions",
            errorContains: []string{"permission", "executable"},
        },
        {
            name:          "general_failure",
            mockExitCode:  1,
            mockError:     "unknown error",
            expectedError: "sandbox command failed: unknown error",
            errorContains: []string{"sandbox command failed"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock sandbox command to return specific error
            // Test that CheckSandboxInstalled returns expected error message
            // Validate error contains all required guidance strings
        })
    }
}
```

### 1.2 Enhanced Tool Validation Tests

**Test File**: `pkg/validation/tools_test.go` (new file)

#### Test Functions:

```go
func TestCheckRequiredTools_SandboxMandatory(t *testing.T)
func TestCheckRequiredTools_ErrorAggregation(t *testing.T)
func TestCheckRequiredTools_ErrorPriority(t *testing.T)
func TestCheckRequiredTools_InstallationGuidance(t *testing.T)
```

#### Test Cases:

| Test Function | Scenario | Expected Behavior |
|---------------|----------|-------------------|
| `TestCheckRequiredTools_SandboxMandatory` | Sandbox missing, other tools present | Validation fails with sandbox-specific error |
| `TestCheckRequiredTools_ErrorAggregation` | Multiple tools missing including sandbox | All errors reported, sandbox error prominent |
| `TestCheckRequiredTools_ErrorPriority` | Various error scenarios | Sandbox errors prioritized in output |
| `TestCheckRequiredTools_InstallationGuidance` | Missing sandbox | Error includes installation/setup instructions |

### 1.3 Configuration Validation Tests

**Test File**: `pkg/config/config_test.go`

#### New Test Functions:

```go
func TestConfig_NoContainerRuntimeOptions(t *testing.T)
func TestConfig_SandboxOnlySupport(t *testing.T)
func TestConfig_PreventAlternativeRuntimes(t *testing.T)
func TestDefaultConfig_SandboxAssumptions(t *testing.T)
```

#### Test Cases:

| Test Function | Purpose | Validation Criteria |
|---------------|---------|-------------------|
| `TestConfig_NoContainerRuntimeOptions` | Ensure no config fields for alternative runtimes | Config struct has no podman/docker/runtime fields |
| `TestConfig_SandboxOnlySupport` | Validate sandbox-only assumptions | All sandbox-related code assumes sandbox availability |
| `TestConfig_PreventAlternativeRuntimes` | Prevent future introduction of alternative runtimes | Configuration validation rejects unknown runtime options |
| `TestDefaultConfig_SandboxAssumptions` | Default configuration assumes sandbox | Default config works with sandbox-only environment |

## 2. Integration Tests

### 2.1 Command Integration with Missing Sandbox

**Test File**: `cmd/start_test.go`, `cmd/stop_test.go`, `cmd/clean_test.go`

#### Enhanced Test Functions:

```go
func TestStartCommand_SandboxMissing(t *testing.T)
func TestStopCommand_SandboxMissing(t *testing.T) 
func TestCleanCommand_SandboxMissing(t *testing.T)
func TestCommandChain_SandboxValidation(t *testing.T)
```

#### Test Scenarios:

| Command | Missing Sandbox Behavior | Expected Outcome |
|---------|-------------------------|------------------|
| `sbs start 123` | Fails at validation step | Clear error about missing sandbox, no partial setup |
| `sbs stop 123` | Fails gracefully | Warning about sandbox operations, continues with other cleanup |
| `sbs clean` | Lists sessions but warns about sandbox cleanup | Shows what would be cleaned, warns about sandbox limitations |
| Command chain | Validation runs before any operations | Early failure prevents inconsistent state |

### 2.2 Tool Validation Integration

**Test File**: `pkg/validation/tools_integration_test.go` (new)

#### Test Functions:

```go
func TestToolValidation_StartupSequence(t *testing.T)
func TestToolValidation_CommandDependencies(t *testing.T)
func TestToolValidation_EnvironmentSetup(t *testing.T)
```

## 3. Error Scenario Tests

### 3.1 Sandbox Command Failure Modes

**Test File**: `pkg/sandbox/error_scenarios_test.go` (new)

#### Test Categories:

1. **Installation Issues**
   - Sandbox not installed
   - Sandbox in non-standard location
   - Sandbox not executable
   - PATH issues

2. **Runtime Issues**
   - Sandbox command timeout
   - Insufficient permissions
   - Disk space issues
   - Network connectivity (if required)

3. **Version Compatibility**
   - Very old sandbox version
   - Incompatible sandbox version
   - Beta/development sandbox version

#### Example Test:

```go
func TestSandboxFailureModes(t *testing.T) {
    tests := []struct {
        name           string
        setupFunc      func() // Simulate failure condition
        cleanupFunc    func() // Restore state
        expectedError  string
        expectedAction string // What user should do
    }{
        {
            name: "sandbox_not_in_path",
            setupFunc: func() {
                // Temporarily modify PATH to exclude sandbox
            },
            expectedError: "sandbox command not found",
            expectedAction: "install sandbox and ensure it's in PATH",
        },
        // ... more scenarios
    }
}
```

### 3.2 Partial System State Tests

**Test File**: `integration/partial_state_test.go` (new)

Test scenarios where sandbox becomes unavailable during operation:

1. **Mid-Operation Failures**
   - Sandbox becomes unavailable during session creation
   - Sandbox fails during cleanup operations
   - Network issues affecting sandbox operations

2. **Recovery Scenarios**
   - System state when sandbox operations fail
   - Cleanup after partial failures
   - User guidance for recovery

## 4. Configuration Prevention Tests

### 4.1 Alternative Runtime Prevention

**Test File**: `pkg/config/runtime_prevention_test.go` (new)

#### Test Functions:

```go
func TestConfig_RejectPodmanConfig(t *testing.T)
func TestConfig_RejectDockerConfig(t *testing.T)
func TestConfig_RejectUnknownRuntimeConfig(t *testing.T)
func TestConfigValidation_SandboxOnly(t *testing.T)
```

#### Test JSON Configurations:

```json
// Should be rejected
{
  "worktree_base_path": "/path/to/worktrees",
  "container_runtime": "podman"  // Should cause validation error
}

{
  "worktree_base_path": "/path/to/worktrees", 
  "podman_command": "podman"     // Should cause validation error
}

{
  "worktree_base_path": "/path/to/worktrees",
  "docker_command": "docker"     // Should cause validation error
}
```

### 4.2 Future-Proofing Tests

**Test File**: `pkg/config/future_proofing_test.go` (new)

Ensure that common alternative runtime configuration patterns are rejected:

```go
func TestConfig_RejectCommonAlternatives(t *testing.T) {
    rejectedConfigs := []string{
        `{"container_runtime": "podman"}`,
        `{"container_runtime": "docker"}`,
        `{"runtime": "podman"}`,
        `{"podman_path": "/usr/bin/podman"}`,
        `{"docker_path": "/usr/bin/docker"}`,
        `{"use_podman": true}`,
        `{"use_docker": true}`,
        `{"container_backend": "podman"}`,
    }
    
    for _, configJSON := range rejectedConfigs {
        t.Run("reject_"+configJSON, func(t *testing.T) {
            // Test that loading this config fails with appropriate error
        })
    }
}
```

## 5. Documentation and User Guidance Tests

### 5.1 Error Message Quality Tests

**Test File**: `pkg/validation/error_messages_test.go` (new)

#### Test Functions:

```go
func TestErrorMessages_Actionable(t *testing.T)
func TestErrorMessages_UserFriendly(t *testing.T)
func TestErrorMessages_InstallationLinks(t *testing.T)
```

#### Error Message Quality Criteria:

1. **Actionable**: Tell user exactly what to do
2. **Specific**: Identify the exact problem
3. **Helpful**: Provide installation/setup guidance
4. **Consistent**: Follow established error message patterns

#### Example Error Message Tests:

```go
func TestErrorMessages_SandboxInstallation(t *testing.T) {
    err := sandbox.CheckSandboxInstalled() // When sandbox is missing
    
    require.Error(t, err)
    
    // Error should be actionable
    assert.Contains(t, err.Error(), "sandbox command not found")
    assert.Contains(t, err.Error(), "Please ensure sandbox is installed")
    assert.Contains(t, err.Error(), "in PATH")
    
    // Error should provide guidance
    // Could include installation instructions or links
}
```

## 6. End-to-End Workflow Tests

### 6.1 Complete Workflow Validation

**Test File**: `e2e/sandbox_required_test.go` (new)

#### Test Scenarios:

1. **New User Experience**
   - Fresh system without sandbox
   - First-time SBS usage
   - Error discovery and resolution

2. **Existing User Migration**
   - Upgrading from older SBS versions
   - Configuration migration
   - Behavior changes

3. **Development Workflows**
   - Developer setup validation
   - CI/CD environment requirements
   - Container environment testing

#### Example E2E Test:

```go
func TestE2E_NewUserWithoutSandbox(t *testing.T) {
    // Simulate fresh system
    originalPath := os.Getenv("PATH")
    defer os.Setenv("PATH", originalPath)
    
    // Remove sandbox from PATH
    os.Setenv("PATH", "/usr/bin:/bin")
    
    // Test that sbs commands fail gracefully with helpful errors
    
    t.Run("start_command_fails_gracefully", func(t *testing.T) {
        cmd := exec.Command("sbs", "start", "123")
        output, err := cmd.CombinedOutput()
        
        require.Error(t, err)
        assert.Contains(t, string(output), "sandbox command not found")
        assert.Contains(t, string(output), "Please ensure sandbox is installed")
    })
    
    t.Run("validation_command_identifies_missing_tools", func(t *testing.T) {
        // Test internal validation function
        err := validation.CheckRequiredTools()
        
        require.Error(t, err)
        assert.Contains(t, err.Error(), "sandbox command not found")
    })
}
```

## 7. Performance and Reliability Tests

### 7.1 Validation Performance Tests

**Test File**: `pkg/validation/performance_test.go` (new)

#### Test Functions:

```go
func BenchmarkCheckRequiredTools(b *testing.B)
func BenchmarkSandboxValidation(b *testing.B)
func TestValidation_Timeout(t *testing.T)
```

Ensure that tool validation:
- Completes quickly (< 1 second typical)
- Has reasonable timeouts
- Doesn't hang on problematic systems

### 7.2 Reliability Tests

**Test File**: `pkg/sandbox/reliability_test.go` (new)

#### Test Scenarios:

1. **Network Issues**: Sandbox commands during network problems
2. **Resource Constraints**: Low disk space, memory pressure
3. **Concurrent Access**: Multiple SBS instances, concurrent sandbox operations
4. **System State Changes**: Sandbox installed/uninstalled during operation

## 8. Mock and Test Utilities

### 8.1 Enhanced Mock Framework

**Test File**: `internal/testutil/sandbox_mock.go` (new)

#### Mock Sandbox Implementation:

```go
type MockSandboxManager struct {
    ShouldFail          bool
    FailWithError       error
    SimulateNotInstalled bool
    CommandCalls        [][]string // Track all command calls
}

func (m *MockSandboxManager) CheckSandboxInstalled() error {
    if m.SimulateNotInstalled {
        return fmt.Errorf("sandbox command not found. Please ensure sandbox is installed and in PATH")
    }
    if m.ShouldFail {
        return m.FailWithError
    }
    return nil
}

// Implement all Manager interface methods with configurable failures
```

### 8.2 Test Environment Setup

**Test File**: `internal/testutil/environment.go` (new)

#### Environment Simulation:

```go
func WithoutSandbox(t *testing.T, testFunc func()) {
    // Temporarily modify PATH to exclude sandbox
    originalPath := os.Getenv("PATH")
    defer os.Setenv("PATH", originalPath)
    
    // Create PATH without sandbox
    cleanPath := removeFromPath(originalPath, "sandbox")
    os.Setenv("PATH", cleanPath)
    
    testFunc()
}

func WithMockSandbox(t *testing.T, mock *MockSandboxManager, testFunc func()) {
    // Replace real sandbox manager with mock
    // Execute test
    // Restore original
}
```

## 9. Test Execution Strategy

### 9.1 Test Categories

| Category | Execution Frequency | Environment Requirements |
|----------|-------------------|-------------------------|
| Unit Tests | Every commit | No external dependencies |
| Integration Tests | Every PR | Mock external tools |
| Error Scenario Tests | Daily/Weekly | Controlled failure simulation |
| E2E Tests | Release candidates | Full development environment |

### 9.2 Continuous Integration

#### Test Matrix:

```yaml
# .github/workflows/test-sandbox-requirements.yml
strategy:
  matrix:
    scenario:
      - name: "with-sandbox"
        setup: "install sandbox"
      - name: "without-sandbox"  
        setup: "remove sandbox from PATH"
      - name: "sandbox-permission-denied"
        setup: "make sandbox non-executable"
```

### 9.3 Test Data Management

#### Error Message Validation Data:

```go
// testdata/expected_errors.json
{
  "sandbox_not_found": {
    "contains": [
      "sandbox command not found",
      "Please ensure sandbox is installed",
      "in PATH"
    ],
    "should_not_contain": [
      "podman",
      "docker",
      "fallback"
    ]
  }
}
```

## 10. Acceptance Criteria Testing

### 10.1 Issue Requirements Validation

Based on issue #17 requirements, tests must verify:

1. **✅ No Podman Support**: 
   - No Podman-related code (verified by analysis)
   - No Podman configuration options
   - No fallback to Podman

2. **✅ Sandbox Command Mandatory**:
   - All operations require sandbox
   - Clear errors when sandbox missing
   - No graceful degradation to non-containerized mode

3. **✅ Clear Error Messages**:
   - Helpful error messages when sandbox unavailable
   - Installation/setup guidance provided
   - Consistent error handling across all commands

4. **✅ Documentation Updates**:
   - All documentation reflects sandbox as hard requirement
   - No references to alternative container runtimes
   - Setup instructions include sandbox requirement

### 10.2 Acceptance Test Suite

**Test File**: `acceptance/issue_17_test.go` (new)

```go
func TestIssue17_SandboxMandatory(t *testing.T) {
    t.Run("no_podman_references", func(t *testing.T) {
        // Scan codebase for any Podman references
        // Fail test if any found
    })
    
    t.Run("sandbox_required_for_all_operations", func(t *testing.T) {
        // Test each major operation requires sandbox
        operations := []string{"start", "stop", "clean", "list"}
        
        for _, op := range operations {
            t.Run(op, func(t *testing.T) {
                // Mock sandbox as unavailable
                // Verify operation fails with proper error
            })
        }
    })
    
    t.Run("clear_error_messages", func(t *testing.T) {
        // Test error message quality
        // Verify messages are actionable and helpful
    })
    
    t.Run("no_fallback_modes", func(t *testing.T) {
        // Ensure no fallback to non-containerized execution
        // Verify failure is clean and complete
    })
}
```

## 11. Test Implementation Priority

### Phase 1: Critical Foundation (Week 1)
1. Enhanced sandbox validation tests
2. Error message quality tests  
3. Configuration prevention tests
4. Basic integration tests

### Phase 2: Comprehensive Coverage (Week 2)
1. End-to-end workflow tests
2. Error scenario tests
3. Mock framework development
4. Performance tests

### Phase 3: Reliability and Polish (Week 3)
1. Reliability tests under various conditions
2. Acceptance criteria validation
3. Documentation validation tests
4. CI/CD integration

## 12. Expected Behavior Changes

### 12.1 Current vs. Target Behavior

| Scenario | Current Behavior | Target Behavior | Test Validation |
|----------|------------------|-----------------|----------------|
| Sandbox missing | Error during validation | Same (already correct) | Verify error message quality |
| Partial sandbox failure | May have inconsistent behavior | Consistent failure with cleanup | Test partial failure scenarios |
| Configuration | No runtime options exist | Prevent future runtime options | Config validation tests |
| Error messages | Basic error reporting | Enhanced, actionable errors | Error message quality tests |

### 12.2 No Breaking Changes Expected

Since analysis shows the codebase already uses sandbox exclusively, this issue primarily focuses on:
- **Strengthening validation**
- **Improving error messages** 
- **Preventing future regression**
- **Comprehensive test coverage**

## 13. Success Metrics

### 13.1 Test Coverage Metrics

- **Unit Test Coverage**: > 90% for sandbox-related code
- **Error Path Coverage**: 100% for sandbox failure scenarios  
- **Integration Coverage**: All major command workflows tested
- **Configuration Coverage**: All config validation paths tested

### 13.2 Quality Metrics

- **Error Message Quality**: All errors actionable and user-friendly
- **Documentation Accuracy**: 100% alignment with sandbox-only approach
- **Performance**: Tool validation < 1 second in normal conditions
- **Reliability**: Graceful handling of all identified failure modes

### 13.3 User Experience Metrics

- **Clear Guidance**: Every error provides next steps
- **Consistent Behavior**: All commands handle missing sandbox consistently
- **No Confusion**: No references to alternative container runtimes
- **Easy Setup**: Clear documentation for sandbox installation

This comprehensive testing plan ensures that SBS maintains its sandbox-only approach with robust error handling, clear user guidance, and prevention of future alternative runtime configurations.