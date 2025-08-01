# Testing Plan: Command Logging Functionality for External Tool Execution

## Overview

This testing plan covers the comprehensive testing strategy for implementing command logging functionality across all external tool executions in the SBS application. The feature introduces centralized command logging with configurable logging levels, performance considerations, and backward compatibility requirements.

## Testing Framework and Tools

### Go Testing Stack
- **Framework**: Go standard `testing` package
- **Assertions**: `testify/assert` and `testify/require` (already in use)
- **Mocking**: `testify/mock` for external dependencies
- **Test Coverage**: `go test -cover` and `go tool cover`
- **Benchmarking**: Go built-in benchmarking for performance tests

### Testing Dependencies (Already Present)
```go
require (
    github.com/stretchr/testify v1.8.4
)
```

## Test Organization Structure

```
pkg/
├── cmdlog/                         # New package for command logging
│   ├── logger.go                   # Command logging implementation  
│   ├── logger_test.go              # Unit tests for command logger
│   └── test_helpers.go             # Test utilities and mocks
├── config/
│   ├── config.go                   # Updated with logging configuration
│   ├── config_test.go              # Updated configuration tests
│   └── cmdlog_config_test.go       # New tests for logging config
├── git/
│   ├── manager.go                  # Updated with command logging
│   ├── manager_test.go             # Updated with logging tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
├── tmux/
│   ├── manager.go                  # Updated with command logging
│   ├── manager_test.go             # Updated with logging tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
├── sandbox/
│   ├── manager.go                  # Updated with command logging
│   ├── sandbox_test.go             # New unit tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
├── issue/
│   ├── github.go                   # Updated with command logging
│   ├── github_test.go              # Updated with logging tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
├── repo/
│   ├── manager.go                  # Updated with command logging
│   ├── manager_test.go             # Updated with logging tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
├── validation/
│   ├── tools.go                    # Updated with command logging
│   ├── tools_test.go               # New unit tests
│   └── cmdlog_integration_test.go  # Command logging integration tests
└── cmd/
    ├── root.go                     # Updated with logging initialization
    ├── root_test.go                # New tests for logging initialization
    └── verbose_flag_test.go        # Tests for --verbose flag
```

## Testing Priority and Implementation Order

### Phase 1: Foundation Tests (Implement First)
1. **Command Logger Unit Tests** - Test core logging functionality
2. **Configuration Tests** - Test new logging configuration options
3. **Performance Baseline Tests** - Establish performance benchmarks

### Phase 2: Integration Tests  
4. **Package-by-Package Integration** - Test logging integration in each package
5. **Cross-Package Integration** - Test logging across multiple components
6. **Configuration Integration** - Test configuration loading and application

### Phase 3: System and Performance Tests
7. **End-to-End Logging Tests** - Test complete command logging workflows
8. **Performance Impact Tests** - Validate no performance regression
9. **Backward Compatibility Tests** - Ensure existing functionality preserved

---

## 1. Unit Tests

### 1.1 Command Logger Tests (`pkg/cmdlog/logger_test.go`)

#### Core Logging Functionality Tests

```go
func TestCommandLogger_BasicLogging(t *testing.T) {
    t.Run("log_command_execution", func(t *testing.T) {
        // Test basic command logging with all parameters
        // Verify log format includes command, arguments, caller context
        // Assert correct timestamp formatting
    })
    
    t.Run("log_command_with_environment", func(t *testing.T) {
        // Test logging commands with environment variables
        // Verify environment variables are logged (sanitized)
        // Assert sensitive data is not logged
    })
    
    t.Run("log_command_execution_time", func(t *testing.T) {
        // Test logging of command execution duration
        // Mock command execution with known duration
        // Verify execution time is captured and logged
    })
    
    t.Run("log_command_exit_code", func(t *testing.T) {
        // Test logging of command exit codes
        // Mock successful and failed command executions
        // Verify exit codes are logged correctly
    })
    
    t.Run("log_command_output_truncation", func(t *testing.T) {
        // Test truncation of large command outputs
        // Mock command with large output
        // Verify output is truncated appropriately
    })
}

func TestCommandLogger_LogLevels(t *testing.T) {
    t.Run("debug_level_logging", func(t *testing.T) {
        // Test DEBUG level logs all details
        // Verify command, arguments, environment, output are logged
        // Assert caller information is included
    })
    
    t.Run("info_level_logging", func(t *testing.T) {
        // Test INFO level logs basic command information
        // Verify command and arguments are logged
        // Assert output and environment are not logged
    })
    
    t.Run("error_level_logging", func(t *testing.T) {
        // Test ERROR level logs only failed commands
        // Mock successful and failed commands
        // Verify only failed commands are logged
    })
    
    t.Run("disabled_logging", func(t *testing.T) {
        // Test that disabled logging produces no output
        // Mock command executions
        // Verify no log entries are created
    })
}

func TestCommandLogger_Configuration(t *testing.T) {
    t.Run("enable_disable_logging", func(t *testing.T) {
        // Test enabling and disabling logging
        // Verify configuration changes take effect
        // Assert no performance impact when disabled
    })
    
    t.Run("log_file_rotation", func(t *testing.T) {
        // Test log file rotation and management
        // Mock large log files
        // Verify rotation occurs correctly
    })
    
    t.Run("concurrent_logging", func(t *testing.T) {
        // Test thread-safe logging from multiple goroutines
        // Execute concurrent command logging
        // Verify no race conditions or corrupted logs
    })
}

func TestCommandLogger_ErrorHandling(t *testing.T) {
    t.Run("log_file_creation_failure", func(t *testing.T) {
        // Test behavior when log file cannot be created
        // Mock filesystem errors
        // Verify graceful degradation (no application failure)
    })
    
    t.Run("log_write_failure", func(t *testing.T) {
        // Test behavior when log writes fail
        // Mock disk full scenarios
        // Verify application continues functioning
    })
    
    t.Run("invalid_log_path", func(t *testing.T) {
        // Test behavior with invalid log paths
        // Mock invalid path configurations
        // Verify fallback behavior
    })
}
```

#### Test Utilities and Fixtures

```go
// pkg/cmdlog/test_helpers.go
type MockLogger struct {
    LogEntries []LogEntry
    Config     LogConfig
}

type LogEntry struct {
    Level       string
    Command     string
    Arguments   []string
    Environment map[string]string
    Output      string
    ExitCode    int
    Duration    time.Duration
    Caller      string
    Timestamp   time.Time
}

type LogConfig struct {
    Enabled   bool
    Level     string
    Path      string
    MaxSize   int64
    MaxAge    int
}

func NewMockLogger(config LogConfig) *MockLogger {
    return &MockLogger{
        LogEntries: make([]LogEntry, 0),
        Config:     config,
    }
}

func (m *MockLogger) LogCommand(cmd, args, caller string, env map[string]string) CommandContext {
    // Mock implementation for testing
}

// Helper to create test command execution scenarios
func CreateTestCommandScenario(name, command string, args []string, exitCode int, output string) TestScenario {
    return TestScenario{
        Name:     name,
        Command:  command,
        Args:     args,
        ExitCode: exitCode,
        Output:   output,
    }
}
```

### 1.2 Configuration Tests (`pkg/config/cmdlog_config_test.go`)

#### Configuration Loading and Validation Tests

```go
func TestConfig_CommandLoggingConfiguration(t *testing.T) {
    t.Run("default_logging_configuration", func(t *testing.T) {
        // Test default logging configuration values
        // Verify CommandLogging defaults to false
        // Assert default log level and path
    })
    
    t.Run("load_logging_config_from_json", func(t *testing.T) {
        // Test loading logging configuration from JSON
        // Mock config file with logging options
        // Verify all logging options are parsed correctly
    })
    
    t.Run("validate_log_level_options", func(t *testing.T) {
        // Test validation of log level values
        // Mock configs with valid/invalid log levels
        // Verify appropriate error handling
    })
    
    t.Run("validate_log_path_options", func(t *testing.T) {
        // Test validation of log path configurations
        // Mock various path scenarios (relative, absolute, invalid)
        // Verify path resolution and validation
    })
    
    t.Run("backward_compatibility_config", func(t *testing.T) {
        // Test that old config files without logging options still work
        // Mock old configuration format
        // Verify default logging values are applied
    })
}

func TestConfig_LoggingConfigurationSerialization(t *testing.T) {
    t.Run("serialize_logging_config", func(t *testing.T) {
        // Test JSON serialization of logging configuration
        // Create config with logging options
        // Verify JSON contains correct logging fields
    })
    
    t.Run("deserialize_logging_config", func(t *testing.T) {
        // Test JSON deserialization of logging configuration
        // Mock JSON with logging configuration
        // Verify config struct is populated correctly
    })
}
```

### 1.3 Package-Specific Wrapper Function Tests

#### Git Manager Logging Tests (`pkg/git/cmdlog_integration_test.go`)

```go
func TestGitManager_CommandLogging(t *testing.T) {
    t.Run("log_git_worktree_add", func(t *testing.T) {
        // Test logging of git worktree add commands
        // Mock git command execution
        // Verify command is logged with correct parameters
    })
    
    t.Run("log_git_worktree_remove", func(t *testing.T) {
        // Test logging of git worktree remove commands
        // Mock git command execution
        // Verify command logging includes force flag
    })
    
    t.Run("log_git_worktree_list", func(t *testing.T) {
        // Test logging of git worktree list commands
        // Mock git command execution
        // Verify porcelain output format is logged
    })
    
    t.Run("log_git_command_failures", func(t *testing.T) {
        // Test logging of failed git commands
        // Mock git command failures
        // Verify error details are logged appropriately
    })
}
```

#### Tmux Manager Logging Tests (`pkg/tmux/cmdlog_integration_test.go`)

```go
func TestTmuxManager_CommandLogging(t *testing.T) {
    t.Run("log_tmux_new_session", func(t *testing.T) {
        // Test logging of tmux new-session commands
        // Mock tmux command execution
        // Verify session parameters are logged
    })
    
    t.Run("log_tmux_has_session", func(t *testing.T) {
        // Test logging of tmux has-session commands
        // Mock tmux session checks
        // Verify session existence checks are logged
    })
    
    t.Run("log_tmux_kill_session", func(t *testing.T) {
        // Test logging of tmux kill-session commands
        // Mock tmux session termination
        // Verify session termination is logged
    })
    
    t.Run("log_tmux_send_keys", func(t *testing.T) {
        // Test logging of tmux send-keys commands
        // Mock tmux key sending
        // Verify commands sent to sessions are logged
    })
    
    t.Run("log_tmux_environment_variables", func(t *testing.T) {
        // Test logging of environment variable setting
        // Mock tmux set-environment commands
        // Verify environment variables are logged (sanitized)
    })
}
```

#### Sandbox Manager Logging Tests (`pkg/sandbox/cmdlog_integration_test.go`)

```go
func TestSandboxManager_CommandLogging(t *testing.T) {
    t.Run("log_sandbox_list", func(t *testing.T) {
        // Test logging of sandbox list commands
        // Mock sandbox command execution
        // Verify sandbox listing is logged
    })
    
    t.Run("log_sandbox_delete", func(t *testing.T) {
        // Test logging of sandbox delete commands
        // Mock sandbox deletion
        // Verify sandbox names are logged correctly
    })
    
    t.Run("log_sandbox_exists_check", func(t *testing.T) {
        // Test logging of sandbox existence checks
        // Mock sandbox existence queries
        // Verify existence checks are logged
    })
}
```

#### GitHub Client Logging Tests (`pkg/issue/cmdlog_integration_test.go`)

```go
func TestGitHubClient_CommandLogging(t *testing.T) {
    t.Run("log_gh_issue_view", func(t *testing.T) {
        // Test logging of gh issue view commands
        // Mock gh command execution
        // Verify issue number and JSON format parameters are logged
    })
    
    t.Run("log_gh_issue_list", func(t *testing.T) {
        // Test logging of gh issue list commands
        // Mock gh command execution
        // Verify search parameters and limits are logged
    })
    
    t.Run("log_gh_authentication_failures", func(t *testing.T) {
        // Test logging of gh authentication failures
        // Mock gh auth failures
        // Verify auth errors are logged without sensitive data
    })
}
```

#### Repository Manager Logging Tests (`pkg/repo/cmdlog_integration_test.go`)

```go
func TestRepositoryManager_CommandLogging(t *testing.T) {
    t.Run("log_git_remote_get_url", func(t *testing.T) {
        // Test logging of git remote get-url commands
        // Mock git remote queries
        // Verify remote URL queries are logged
    })
    
    t.Run("log_git_remote_failures", func(t *testing.T) {
        // Test logging of git remote command failures
        // Mock git remote errors
        // Verify error handling is logged
    })
}
```

#### Validation Tools Logging Tests (`pkg/validation/cmdlog_integration_test.go`)

```go
func TestValidationTools_CommandLogging(t *testing.T) {
    t.Run("log_tool_version_checks", func(t *testing.T) {
        // Test logging of tool version checks (tmux -V, git --version, etc.)
        // Mock tool validation commands
        // Verify version checks are logged
    })
    
    t.Run("log_tool_availability_checks", func(t *testing.T) {
        // Test logging of tool availability checks
        // Mock tool existence verification
        // Verify availability checks are logged
    })
}
```

---

## 2. Integration Tests

### 2.1 Command Flow Integration (`integration_test.go`)

#### End-to-End Command Logging Tests

```go
func TestCommandLogging_EndToEndWorkflow(t *testing.T) {
    t.Run("full_sbs_start_workflow_logging", func(t *testing.T) {
        // Test complete sbs start workflow with logging enabled
        // Mock all external commands (git, tmux, gh, sandbox)
        // Verify all commands are logged in correct sequence
        // Assert log entries contain expected caller context
    })
    
    t.Run("sbs_stop_workflow_logging", func(t *testing.T) {
        // Test sbs stop workflow with logging enabled
        // Mock tmux and sandbox cleanup commands
        // Verify cleanup commands are logged
    })
    
    t.Run("sbs_list_workflow_logging", func(t *testing.T) {
        // Test sbs list workflow with logging enabled
        // Mock tmux list and session queries
        // Verify query commands are logged
    })
    
    t.Run("command_logging_with_failures", func(t *testing.T) {
        // Test command logging when external commands fail
        // Mock command failures at various points
        // Verify failure details are logged appropriately
    })
}

func TestCommandLogging_ConfigurationIntegration(t *testing.T) {
    t.Run("logging_disabled_no_output", func(t *testing.T) {
        // Test that disabled logging produces no log output
        // Configure logging as disabled
        // Execute full workflow
        // Verify no log files are created or written to
    })
    
    t.Run("different_log_levels", func(t *testing.T) {
        // Test different log levels produce appropriate output
        // Execute workflow with DEBUG, INFO, ERROR levels
        // Verify log content matches expected level
    })
    
    t.Run("custom_log_path", func(t *testing.T) {
        // Test custom log path configuration
        // Configure custom log file path
        // Execute workflow and verify logs written to correct location
    })
}
```

### 2.2 Performance Integration Tests

#### Performance Impact Validation

```go
func TestCommandLogging_PerformanceImpact(t *testing.T) {
    t.Run("logging_disabled_performance", func(t *testing.T) {
        // Benchmark command execution with logging disabled
        // Measure baseline performance
        // Assert no measurable overhead
    })
    
    t.Run("logging_enabled_performance", func(t *testing.T) {
        // Benchmark command execution with logging enabled
        // Measure performance impact
        // Assert impact is within acceptable limits (<5% overhead)
    })
    
    t.Run("concurrent_command_logging", func(t *testing.T) {
        // Test performance with concurrent command executions
        // Execute multiple commands simultaneously
        // Verify no performance degradation or race conditions
    })
}

func BenchmarkCommandLogging(b *testing.B) {
    // Benchmark logging functionality
    b.Run("logging_disabled", func(b *testing.B) {
        // Benchmark with logging disabled
    })
    
    b.Run("logging_info_level", func(b *testing.B) {
        // Benchmark with INFO level logging
    })
    
    b.Run("logging_debug_level", func(b *testing.B) {
        // Benchmark with DEBUG level logging
    })
}
```

---

## 3. System Tests

### 3.1 Configuration System Tests

#### Configuration Loading and Application Tests

```go
func TestCommandLogging_ConfigurationSystem(t *testing.T) {
    t.Run("config_file_missing", func(t *testing.T) {
        // Test behavior when config file is missing
        // Verify default logging configuration is applied
        // Assert application starts successfully
    })
    
    t.Run("config_file_invalid_json", func(t *testing.T) {
        // Test behavior with invalid JSON configuration
        // Mock malformed config file
        // Verify graceful error handling
    })
    
    t.Run("config_reload_during_execution", func(t *testing.T) {
        // Test configuration changes during application execution
        // Modify config while application is running
        // Verify changes take effect (if supported)
    })
    
    t.Run("environment_variable_override", func(t *testing.T) {
        // Test environment variable overrides for logging config
        // Set logging environment variables
        // Verify they override config file values
    })
}
```

### 3.2 Verbose Flag System Tests

#### --verbose Flag Implementation Tests

```go
func TestVerboseFlag_CommandLogging(t *testing.T) {
    t.Run("verbose_flag_enables_logging", func(t *testing.T) {
        // Test that --verbose flag enables temporary logging
        // Execute command with --verbose flag
        // Verify logging is enabled regardless of config
    })
    
    t.Run("verbose_flag_with_logging_disabled", func(t *testing.T) {
        // Test --verbose flag overrides disabled logging config
        // Configure logging as disabled
        // Execute with --verbose flag
        // Verify logging occurs
    })
    
    t.Run("verbose_flag_log_level", func(t *testing.T) {
        // Test --verbose flag sets appropriate log level
        // Execute with --verbose flag
        // Verify DEBUG level logging is used
    })
    
    t.Run("verbose_flag_output_location", func(t *testing.T) {
        // Test --verbose flag output goes to stderr or console
        // Execute with --verbose flag
        // Verify output is visible to user
    })
}
```

---

## 4. Edge Case Tests

### 4.1 Error Handling and Recovery Tests

#### Command Execution Edge Cases

```go
func TestCommandLogging_EdgeCases(t *testing.T) {
    t.Run("extremely_long_command_lines", func(t *testing.T) {
        // Test logging of commands with very long argument lists
        // Mock commands with 1000+ character arguments
        // Verify logging handles long commands gracefully
    })
    
    t.Run("commands_with_binary_output", func(t *testing.T) {
        // Test logging of commands that produce binary output
        // Mock commands with non-UTF8 output
        // Verify binary output is handled safely
    })
    
    t.Run("commands_with_sensitive_arguments", func(t *testing.T) {
        // Test that sensitive arguments are sanitized
        // Mock commands with passwords, tokens, etc.
        // Verify sensitive data is not logged
    })
    
    t.Run("rapid_command_execution", func(t *testing.T) {
        // Test logging with rapid command execution
        // Execute many commands in quick succession
        // Verify all commands are logged correctly
    })
    
    t.Run("command_timeout_scenarios", func(t *testing.T) {
        // Test logging of commands that timeout
        // Mock long-running commands
        // Verify timeout handling is logged
    })
}
```

### 4.2 Resource Management Tests

#### Memory and Disk Usage Tests

```go
func TestCommandLogging_ResourceManagement(t *testing.T) {
    t.Run("log_file_size_management", func(t *testing.T) {
        // Test log file size limits and rotation
        // Generate large amounts of log data
        // Verify log rotation occurs correctly
    })
    
    t.Run("memory_usage_with_large_outputs", func(t *testing.T) {
        // Test memory usage with commands producing large outputs
        // Mock commands with multi-MB outputs
        // Verify memory usage remains reasonable
    })
    
    t.Run("disk_space_exhaustion", func(t *testing.T) {
        // Test behavior when disk space is exhausted
        // Mock disk full scenarios
        // Verify graceful degradation
    })
    
    t.Run("log_cleanup_on_application_exit", func(t *testing.T) {
        // Test cleanup behavior when application exits
        // Mock application termination scenarios
        // Verify log files are properly closed
    })
}
```

---

## 5. Security and Privacy Tests

### 5.1 Data Sanitization Tests

#### Sensitive Information Protection

```go
func TestCommandLogging_DataSanitization(t *testing.T) {
    t.Run("sanitize_authentication_tokens", func(t *testing.T) {
        // Test sanitization of GitHub tokens and API keys
        // Mock commands containing auth tokens
        // Verify tokens are redacted from logs
    })
    
    t.Run("sanitize_passwords_and_secrets", func(t *testing.T) {
        // Test sanitization of passwords and secrets
        // Mock commands with password arguments
        // Verify sensitive data is masked
    })
    
    t.Run("sanitize_personal_information", func(t *testing.T) {
        // Test sanitization of personal information
        // Mock commands with user data
        // Verify personal information is protected
    })
    
    t.Run("sanitize_file_paths", func(t *testing.T) {
        // Test sanitization of sensitive file paths
        // Mock commands with private file paths
        // Verify path information is appropriately handled
    })
}
```

### 5.2 Log File Security Tests

#### Log File Access and Permissions

```go
func TestCommandLogging_LogFileSecurity(t *testing.T) {
    t.Run("log_file_permissions", func(t *testing.T) {
        // Test that log files have appropriate permissions
        // Create log files and check permissions
        // Verify files are not world-readable
    })
    
    t.Run("log_file_ownership", func(t *testing.T) {
        // Test log file ownership
        // Verify log files are owned by correct user
        // Assert no privilege escalation issues
    })
    
    t.Run("log_rotation_security", func(t *testing.T) {
        // Test security during log rotation
        // Verify old log files maintain secure permissions
        // Assert no temporary file vulnerabilities
    })
}
```

---

## 6. Backward Compatibility Tests

### 6.1 Existing Functionality Preservation

#### Regression Testing

```go
func TestCommandLogging_BackwardCompatibility(t *testing.T) {
    t.Run("existing_workflows_unchanged", func(t *testing.T) {
        // Test that all existing workflows continue to work
        // Execute sbs start, stop, list, attach, clean
        // Verify identical behavior to pre-logging version
    })
    
    t.Run("existing_configuration_compatibility", func(t *testing.T) {
        // Test that existing config files continue to work
        // Load old configuration formats
        // Verify new logging defaults are applied
    })
    
    t.Run("existing_error_messages_unchanged", func(t *testing.T) {
        // Test that error messages remain the same
        // Mock various error scenarios
        // Verify error output is identical
    })
    
    t.Run("existing_command_line_interface", func(t *testing.T) {
        // Test that existing CLI behavior is preserved
        // Execute all existing commands and flags
        // Verify identical behavior and output
    })
}
```

### 6.2 Performance Regression Tests

#### Performance Baseline Comparison

```go
func TestCommandLogging_PerformanceRegression(t *testing.T) {
    t.Run("startup_time_regression", func(t *testing.T) {
        // Test that application startup time is not impacted
        // Measure startup time with and without logging
        // Assert no significant difference
    })
    
    t.Run("command_execution_time_regression", func(t *testing.T) {
        // Test that command execution time is not impacted
        // Measure execution time for typical workflows
        // Assert performance within acceptable limits
    })
    
    t.Run("memory_usage_regression", func(t *testing.T) {
        // Test that memory usage has not increased significantly
        // Measure memory usage during typical operations
        // Assert memory usage is within acceptable limits
    })
}
```

---

## 7. Test Data and Fixtures

### 7.1 Command Execution Scenarios

#### Test Data for Various Commands

```go
// pkg/cmdlog/test_fixtures.go
var TestCommandScenarios = []TestScenario{
    {
        Name:        "git_worktree_add_success",
        Command:     "git",
        Args:        []string{"worktree", "add", "/path/to/worktree", "branch-name"},
        ExitCode:    0,
        Output:      "Preparing worktree (new branch 'branch-name')",
        Environment: map[string]string{"GIT_DIR": "/path/to/.git"},
    },
    {
        Name:        "tmux_new_session_success",
        Command:     "tmux",
        Args:        []string{"new-session", "-d", "-s", "session-name", "-c", "/path"},
        ExitCode:    0,
        Output:      "",
        Environment: map[string]string{"SBS_TITLE": "test-title"},
    },
    {
        Name:        "gh_issue_view_success",
        Command:     "gh",
        Args:        []string{"issue", "view", "123", "--json", "number,title,state,url"},
        ExitCode:    0,
        Output:      `{"number":123,"title":"Test Issue","state":"open","url":"https://github.com/owner/repo/issues/123"}`,
        Environment: map[string]string{},
    },
    {
        Name:        "sandbox_list_success",
        Command:     "sandbox",
        Args:        []string{"list"},
        ExitCode:    0,
        Output:      "work-issue-123\nwork-issue-456",
        Environment: map[string]string{},
    },
    {
        Name:        "command_failure_scenario",
        Command:     "git",
        Args:        []string{"worktree", "add", "/invalid/path", "nonexistent-branch"},
        ExitCode:    1,
        Output:      "fatal: invalid reference: nonexistent-branch",
        Environment: map[string]string{},
    },
}

var SensitiveDataScenarios = []TestScenario{
    {
        Name:        "github_token_in_environment",
        Command:     "gh",
        Args:        []string{"api", "repos/owner/repo"},
        ExitCode:    0,
        Output:      `{"name":"repo"}`,
        Environment: map[string]string{"GITHUB_TOKEN": "ghp_xxxxxxxxxxxxxxxxxxxx"},
    },
    {
        Name:        "password_in_arguments",
        Command:     "mysql",
        Args:        []string{"-u", "user", "-p", "password123", "database"},
        ExitCode:    0,
        Output:      "Connected to database",
        Environment: map[string]string{},
    },
}
```

### 7.2 Configuration Test Data

#### Test Configuration Files

```go
// Test configurations for various scenarios
var TestConfigurations = struct {
    DefaultConfig      string
    LoggingEnabledConfig string
    DebugLevelConfig   string
    DisabledConfig     string
    InvalidConfig      string
}{
    DefaultConfig: `{
        "worktree_base_path": "~/.work-issue-worktrees",
        "work_issue_script": "./work-issue.sh"
    }`,
    
    LoggingEnabledConfig: `{
        "worktree_base_path": "~/.work-issue-worktrees",
        "work_issue_script": "./work-issue.sh",
        "command_logging": true,
        "command_log_level": "info",
        "command_log_path": "~/.config/sbs/command.log"
    }`,
    
    DebugLevelConfig: `{
        "worktree_base_path": "~/.work-issue-worktrees", 
        "work_issue_script": "./work-issue.sh",
        "command_logging": true,
        "command_log_level": "debug",
        "command_log_path": "~/.config/sbs/debug.log"
    }`,
    
    DisabledConfig: `{
        "worktree_base_path": "~/.work-issue-worktrees",
        "work_issue_script": "./work-issue.sh",
        "command_logging": false
    }`,
    
    InvalidConfig: `{
        "worktree_base_path": "~/.work-issue-worktrees",
        "work_issue_script": "./work-issue.sh",
        "command_logging": true,
        "command_log_level": "invalid_level"
    }`,
}
```

---

## 8. Test Environment Setup

### 8.1 Test Repository and Environment

#### Integration Test Environment Setup

```bash
#!/bin/bash
# scripts/setup-test-environment.sh

# Create test repository for integration tests
create_test_repo() {
    local test_dir="$1"
    mkdir -p "$test_dir"
    cd "$test_dir"
    git init
    echo "# Test Repository for Command Logging" > README.md
    git add README.md
    git commit -m "Initial commit"
    git remote add origin https://github.com/test/repo.git
}

# Setup test configuration files
setup_test_configs() {
    local config_dir="$1"
    mkdir -p "$config_dir"
    
    # Create test configuration files
    cat > "$config_dir/default.json" << EOF
{
    "worktree_base_path": "~/.work-issue-worktrees",
    "work_issue_script": "./work-issue.sh"
}
EOF

    cat > "$config_dir/logging-enabled.json" << EOF
{
    "worktree_base_path": "~/.work-issue-worktrees",
    "work_issue_script": "./work-issue.sh",
    "command_logging": true,
    "command_log_level": "info",
    "command_log_path": "/tmp/sbs-test-command.log"
}
EOF
}

# Setup mock external tools for testing
setup_mock_tools() {
    local bin_dir="$1"
    mkdir -p "$bin_dir"
    
    # Create mock git command
    cat > "$bin_dir/git" << 'EOF'
#!/bin/bash
echo "Mock git called with: $@" >&2
case "$1" in
    "worktree")
        case "$2" in
            "add") echo "Preparing worktree" ;;
            "remove") echo "Removing worktree" ;;
            "list") echo "worktree /path/to/worktree" ;;
        esac
        ;;
    "--version") echo "git version 2.40.0" ;;
    *) echo "Mock git output" ;;
esac
EOF
    chmod +x "$bin_dir/git"
    
    # Create mock tmux command
    cat > "$bin_dir/tmux" << 'EOF'
#!/bin/bash
echo "Mock tmux called with: $@" >&2
case "$1" in
    "new-session") exit 0 ;;
    "has-session") exit 0 ;;
    "kill-session") exit 0 ;;
    "list-sessions") echo "session-1: 1 windows" ;;
    "-V") echo "tmux 3.3a" ;;
esac
EOF
    chmod +x "$bin_dir/tmux"
    
    # Create mock gh command
    cat > "$bin_dir/gh" << 'EOF'
#!/bin/bash
echo "Mock gh called with: $@" >&2
case "$2" in
    "view") echo '{"number":123,"title":"Test Issue","state":"open","url":"https://github.com/owner/repo/issues/123"}' ;;
    "list") echo '[{"number":123,"title":"Test Issue","state":"open","url":"https://github.com/owner/repo/issues/123"}]' ;;
esac
EOF
    chmod +x "$bin_dir/gh"
    
    # Create mock sandbox command
    cat > "$bin_dir/sandbox" << 'EOF'
#!/bin/bash
echo "Mock sandbox called with: $@" >&2
case "$1" in
    "list") echo "work-issue-123" ;;
    "delete") exit 0 ;;
    "--help") echo "sandbox usage" ;;
esac
EOF
    chmod +x "$bin_dir/sandbox"
}
```

### 8.2 Continuous Integration Setup

#### GitHub Actions Workflow for Command Logging Tests

```yaml
# .github/workflows/command-logging-tests.yml
name: Command Logging Tests

on:
  push:
    paths:
      - 'pkg/cmdlog/**'
      - 'pkg/config/**'
      - 'pkg/*/cmdlog_integration_test.go'
      - 'cmd/root.go'
  pull_request:
    paths:
      - 'pkg/cmdlog/**'
      - 'pkg/config/**'
      - 'pkg/*/cmdlog_integration_test.go'
      - 'cmd/root.go'

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Install Dependencies
      run: |
        go mod download
        sudo apt-get update
        sudo apt-get install -y tmux
    
    - name: Run Command Logging Unit Tests
      run: |
        go test -v ./pkg/cmdlog/...
        go test -v ./pkg/config/... -run=".*CommandLogging.*"
    
    - name: Run Integration Tests
      run: |
        go test -v ./pkg/git/... -run=".*CommandLogging.*"
        go test -v ./pkg/tmux/... -run=".*CommandLogging.*"
        go test -v ./pkg/sandbox/... -run=".*CommandLogging.*"
        go test -v ./pkg/issue/... -run=".*CommandLogging.*"
        go test -v ./pkg/repo/... -run=".*CommandLogging.*"
        go test -v ./pkg/validation/... -run=".*CommandLogging.*"
    
    - name: Test Coverage
      run: |
        go test -coverprofile=coverage.out ./pkg/cmdlog/...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload Coverage
      uses: actions/upload-artifact@v3
      with:
        name: coverage-report
        path: coverage.html

  performance-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run Performance Benchmarks
      run: |
        go test -bench=. -benchmem ./pkg/cmdlog/...
        
    - name: Performance Regression Check
      run: |
        # Compare benchmark results with baseline
        go test -bench=. ./... > current_bench.txt
        # Add logic to compare with previous benchmarks

  security-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run Security Tests
      run: |
        go test -v ./pkg/cmdlog/... -run=".*Security.*"
        go test -v ./pkg/cmdlog/... -run=".*Sanitization.*"
    
    - name: Check for Sensitive Data Leaks
      run: |
        # Run tests that verify no sensitive data appears in logs
        go test -v ./... -run=".*SensitiveData.*"
```

---

## 9. Test Execution Strategy

### 9.1 Test Phases and Timeline

#### Phase 1: Foundation (Week 1-2)
- [ ] Implement `pkg/cmdlog/` package unit tests
- [ ] Create configuration system tests
- [ ] Set up test environment and fixtures
- [ ] Establish performance baselines

#### Phase 2: Integration (Week 3-4)
- [ ] Implement package-by-package integration tests
- [ ] Create cross-package integration tests
- [ ] Test verbose flag functionality
- [ ] Validate configuration integration

#### Phase 3: System Testing (Week 5-6)
- [ ] End-to-end workflow tests
- [ ] Performance regression tests
- [ ] Security and privacy tests
- [ ] Backward compatibility validation

#### Phase 4: Edge Cases and Cleanup (Week 7)
- [ ] Edge case testing
- [ ] Resource management tests
- [ ] Final performance validation
- [ ] Documentation and test maintenance

### 9.2 Success Criteria and Metrics

#### Test Quality Metrics
- [ ] **Unit Test Coverage**: Minimum 85% for `pkg/cmdlog/` package
- [ ] **Integration Test Coverage**: All command execution paths tested
- [ ] **Performance Impact**: Less than 5% overhead when logging enabled
- [ ] **Security Validation**: No sensitive data in logs
- [ ] **Backward Compatibility**: 100% existing functionality preserved

#### Functional Validation
- [ ] **Configuration Loading**: All logging configuration options work correctly
- [ ] **Log Output**: Correct format, level filtering, and file management
- [ ] **Error Handling**: Graceful degradation when logging fails
- [ ] **Concurrent Safety**: Thread-safe logging from multiple goroutines

---

## 10. Test Maintenance and Best Practices

### 10.1 Test Code Quality Standards

#### Testing Principles
- **Arrange-Act-Assert**: Structure all tests with clear setup, execution, and validation
- **Single Responsibility**: Each test validates one specific behavior
- **Descriptive Names**: Test names clearly describe the scenario being tested
- **Independent Tests**: No dependencies between test cases
- **Repeatable Results**: Tests produce consistent results across environments

#### Mock and Stub Guidelines
- **Mock External Dependencies**: All external command executions mocked in unit tests
- **Use Interfaces**: Define interfaces for testable abstractions
- **Realistic Test Data**: Use data that represents real-world command scenarios
- **Minimal Mocking**: Focus mocking on system boundaries and external dependencies

### 10.2 Test Data Management

#### Test Data Principles
- **Version Control**: All test fixtures and mock data committed to repository
- **Realistic Scenarios**: Test data based on actual command execution patterns
- **Privacy Protection**: No real user data or sensitive information in tests
- **Comprehensive Coverage**: Test data covers success, failure, and edge cases

### 10.3 Continuous Improvement

#### Test Evolution Strategy
- **Regular Review**: Monthly review of test effectiveness and coverage
- **New Scenario Addition**: Add tests for newly discovered edge cases
- **Performance Monitoring**: Track test execution time and optimize slow tests
- **Documentation Updates**: Keep test documentation current with implementation

---

## 11. Risk Mitigation

### 11.1 Testing Risks and Mitigation

#### Identified Risks
1. **Performance Impact**: Command logging might slow down application
   - **Mitigation**: Comprehensive benchmarking and performance regression tests
   
2. **Security Vulnerabilities**: Sensitive data might be logged
   - **Mitigation**: Extensive data sanitization tests and security review
   
3. **Resource Exhaustion**: Log files might consume excessive disk space
   - **Mitigation**: Log rotation tests and disk usage monitoring
   
4. **Backward Compatibility**: New logging might break existing functionality
   - **Mitigation**: Comprehensive regression testing and feature flagging

### 11.2 Fallback Strategies

#### Graceful Degradation Testing
- [ ] Test application behavior when logging initialization fails
- [ ] Verify application continues functioning when log writes fail
- [ ] Validate fallback to console output when log files are unavailable
- [ ] Ensure no data loss when logging encounters errors

---

This comprehensive testing plan ensures robust validation of the command logging functionality while maintaining the high quality, security, and performance standards of the SBS application. The plan follows test-driven development principles and provides thorough coverage of all implementation aspects, edge cases, and integration scenarios.