# End-to-End Testing Framework for SBS

This document describes the comprehensive end-to-end (E2E) testing framework for SBS (Sandbox Sessions).

## Overview

The E2E testing framework provides automated testing of the complete SBS workflow, including:

- **Session lifecycle management** (start, stop, attach)
- **Status tracking and time detection**
- **Tmux session management**
- **Git worktree operations**
- **Hook integration** (Claude Code hooks, logging)
- **Cleanup operations**
- **Error handling and recovery**
- **Configuration management**

## Quick Start

### Running E2E Tests

```bash
# Run all E2E tests using Make
make e2e

# Or use the dedicated script
./scripts/run-e2e-tests.sh

# Run with manual setup
E2E_TESTS=1 go test -tags=e2e -v ./e2e_test.go
```

### Prerequisites

- Git (required)
- tmux (recommended, some tests may be skipped without it)
- jq (optional, for pretty JSON output)

## Test Architecture

### Test Suite Structure

The `E2ETestSuite` provides a comprehensive testing environment:

```go
type E2ETestSuite struct {
    t                *testing.T
    tempDir          string          // Isolated test directory
    configDir        string          // SBS configuration directory
    worktreeBaseDir  string          // Git worktree base path
    sbsBinary        string          // Path to SBS binary
    originalWorkDir  string          // Original working directory
    testRepoDir      string          // Test git repository
    sessionIDs       []string        // Track sessions for cleanup
}
```

### Test Environment Setup

Each test suite automatically:

1. **Creates isolated directories** for config, worktrees, and test repository
2. **Builds the SBS binary** if not already present
3. **Initializes a test git repository** with proper configuration
4. **Creates test configuration** with minimal work-issue.sh script
5. **Sets up input source configuration** for test work types

### Automatic Cleanup

The framework provides automatic cleanup:
- Stops all created sessions
- Runs `sbs clean --force` to clean resources
- Removes temporary directories
- Kills leftover tmux sessions

## Test Categories

### 1. Session Lifecycle Tests (`TestE2E_SessionLifecycle`)

Tests the complete session workflow:

```bash
# Creates session
sbs start test:e2e-lifecycle

# Verifies session exists
sbs list --plain

# Verifies worktree and test artifacts
ls ~/.sbs-worktrees/test-e2e-lifecycle/

# Stops session
sbs stop test:e2e-lifecycle
```

**What it tests:**
- Session creation and startup
- Work-issue.sh script execution
- Worktree creation and git operations
- Session listing and status
- Session termination

### 2. Status Tracking Tests (`TestE2E_StatusTracking`)

Tests status detection and time tracking:

```bash
# Start session and let it run
sbs start test:e2e-status

# Check status detection
sbs list --plain

# Verify hook integration
ls ~/.sbs-worktrees/test-e2e-status/.sbs/
```

**What it tests:**
- Session status detection
- Time tracking functionality
- Hook installation and execution
- Status file creation and parsing

### 3. Concurrent Session Tests (`TestE2E_MultipleSessionsConcurrency`)

Tests multiple simultaneous sessions:

```bash
# Start multiple sessions
sbs start test:e2e-concurrent1
sbs start test:e2e-concurrent2
sbs start test:e2e-concurrent3

# Verify all sessions
sbs list --plain

# Test session attachment
sbs attach test:e2e-concurrent1

# Cleanup all sessions
sbs clean --force
```

**What it tests:**
- Multiple concurrent sessions
- Session isolation
- Resource management under load
- Bulk cleanup operations

### 4. Git Integration Tests (`TestE2E_WorktreeAndGitIntegration`)

Tests git worktree and branch management:

```bash
# Create session with git operations
sbs start test:e2e-worktree

# Verify git worktree
cd ~/.sbs-worktrees/test-e2e-worktree/
git status

# Check branch creation
git branch -a
```

**What it tests:**
- Git worktree creation and management
- Branch creation and naming
- Git operations within worktrees
- Repository state consistency

### 5. Error Handling Tests (`TestE2E_ErrorHandling`)

Tests error scenarios and recovery:

```bash
# Invalid session IDs
sbs start invalid:format:with:colons

# Nonexistent sessions
sbs stop test:nonexistent
sbs attach test:nonexistent
```

**What it tests:**
- Invalid input handling
- Graceful error messages
- Recovery from failures
- Input validation

### 6. Configuration Tests (`TestE2E_ConfigurationHandling`)

Tests configuration scenarios:

```bash
# Custom config file
sbs --config /path/to/config.json start test:config

# Missing dependencies
sbs start test:missing-script
```

**What it tests:**
- Custom configuration loading
- Missing file handling
- Configuration validation
- Environment variable handling

### 7. Cleanup Tests (`TestE2E_CleanupOperations`)

Tests comprehensive cleanup:

```bash
# Create multiple sessions
sbs start test:cleanup1
sbs start test:cleanup2

# Test dry-run cleanup
sbs clean --dry-run

# Test force cleanup
sbs clean --force
```

**What it tests:**
- Dry-run operations
- Force cleanup execution
- Resource identification
- Complete cleanup verification

## Test Utilities

### Custom Work-Issue.sh Script

The E2E tests use a minimal work-issue.sh script that:

```bash
#!/bin/bash
echo "Starting work on issue: $1"
echo "Working directory: $(pwd)"
env | grep SBS_ || true

# Create test marker file
echo "E2E test work session for $1" > .sbs-test-marker
echo "Timestamp: $(date)" >> .sbs-test-marker

sleep 2
echo "Work setup complete for issue: $1"
```

This script:
- Provides visible output for verification
- Creates test marker files for validation
- Simulates real work session setup
- Handles environment variables

### Test Input Source Configuration

Tests use the `test` input source type:

```json
{
  "type": "test",
  "settings": {}
}
```

This allows testing with predictable work items without external dependencies.

## Running Tests

### Local Development

```bash
# Run E2E tests with verbose output
make e2e

# Run specific test function
E2E_TESTS=1 go test -tags=e2e -v -run TestE2E_SessionLifecycle ./e2e_test.go

# Run with timeout
E2E_TESTS=1 go test -tags=e2e -v -timeout=15m ./e2e_test.go
```

### Continuous Integration

```bash
# Full test suite including E2E
make test-all

# Just integration and E2E tests
make integration e2e
```

### Script Runner Options

The `./scripts/run-e2e-tests.sh` script supports:

```bash
# Full test run (default)
./scripts/run-e2e-tests.sh

# Show help
./scripts/run-e2e-tests.sh help

# Only build binary
./scripts/run-e2e-tests.sh build

# Only run tests (requires pre-built binary)
./scripts/run-e2e-tests.sh test

# Only cleanup leftover resources
./scripts/run-e2e-tests.sh clean
```

## Environment Variables

- `E2E_TESTS=1` - Enable E2E test execution
- `GO_TEST_TIMEOUT` - Test timeout (default: 10m)
- `XDG_CONFIG_HOME` - Override config directory location

## Troubleshooting

### Tests Fail to Start

1. **Check binary build**: `make build`
2. **Verify git repository**: Ensure you're in a git repository
3. **Check tmux availability**: `which tmux`

### Tests Leave Resources

1. **Manual cleanup**: `./scripts/run-e2e-tests.sh clean`
2. **Kill tmux sessions**: `tmux kill-server`
3. **Remove temp directories**: Check `/tmp/` for test directories

### Timeout Issues

1. **Increase timeout**: `GO_TEST_TIMEOUT=20m make e2e`
2. **Run individual tests**: Use `-run` flag to isolate tests
3. **Check system load**: Ensure adequate resources

## Extending Tests

### Adding New Test Cases

1. Create new test function following naming convention:
   ```go
   func TestE2E_YourFeature(t *testing.T) {
       suite := NewE2ETestSuite(t)
       defer suite.cleanup()
       
       // Your test logic
   }
   ```

2. Add session IDs to cleanup tracking:
   ```go
   sessionID := "test:your-feature"
   suite.sessionIDs = append(suite.sessionIDs, sessionID)
   ```

3. Use test utilities for common operations:
   ```go
   output, err := suite.runSBSCommand("start", sessionID)
   require.NoError(t, err)
   ```

### Testing New Features

1. **Add feature-specific test category**
2. **Create isolated test environment**
3. **Verify all aspects of feature functionality**
4. **Test error conditions and edge cases**
5. **Ensure proper cleanup**

This E2E testing framework provides comprehensive validation of SBS functionality and helps ensure reliability as the project evolves.
