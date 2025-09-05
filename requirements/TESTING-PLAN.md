# Testing Plan: Remove executeStaleCleanup Wrapper Function

## Overview

This testing plan covers the validation required for removing the unnecessary `executeStaleCleanup` wrapper function from `cmd/clean.go` (lines 215-219) as described in GitHub issue #37.

## Refactor Summary

The refactor involves:
1. **Removing the wrapper function** (lines 215-219):
   ```go
   func executeStaleCleanup(dryRun, force bool) error {
       fmt.Println("Cleaning up stale sessions only...")
       return executeDefaultCleanup(dryRun, force)
   }
   ```

2. **Updating three call sites** to call `executeDefaultCleanup` directly:
   - Line 103: `case CleanupModeStale:`
   - Line 110: in `CleanupModeStaleAndBranches` case
   - Line 316: in `executeComprehensiveCleanup`

3. **Adding appropriate print messages** at each call site to preserve user-visible behavior

4. **Optional enhancement**: Rename `executeDefaultCleanup` to `executeStaleSessionCleanup` for clarity

## Testing Strategy

### 1. Unit Testing Approach

#### 1.1 Function Removal Validation
- **Test**: Verify `executeStaleCleanup` function no longer exists
- **Method**: Attempt to call the function and expect compilation failure
- **Location**: `cmd/clean_test.go`

#### 1.2 Direct Call Site Testing
- **Test**: Verify all three call sites now call `executeDefaultCleanup` directly
- **Method**: Mock/spy testing to verify correct function calls
- **Expected Behavior**:
  - `CleanupModeStale` case calls `executeDefaultCleanup` with appropriate message
  - `CleanupModeStaleAndBranches` case calls `executeDefaultCleanup` followed by `executeBranchCleanup`
  - `executeComprehensiveCleanup` calls `executeDefaultCleanup` directly

#### 1.3 Message Output Testing
- **Test**: Verify appropriate messages are printed at each call site
- **Method**: Capture stdout/stderr during test execution
- **Expected Messages**:
  - `CleanupModeStale`: "Cleaning up stale sessions only..."
  - `CleanupModeStaleAndBranches`: "Cleaning up stale sessions only..." (before branch cleanup)
  - `executeComprehensiveCleanup`: "Cleaning up stale sessions only..." (within comprehensive flow)

#### 1.4 Function Signature Consistency
- **Test**: Verify all direct calls pass correct parameters (`dryRun`, `force`)
- **Method**: Parameter validation testing
- **Expected**: All calls maintain `(dryRun, force bool) error` signature

### 2. Integration Testing

#### 2.1 Cleanup Mode Integration Tests
Test each cleanup mode end-to-end to ensure the refactor doesn't break functionality:

##### Test Case: Stale-Only Cleanup (`--stale`)
```bash
# Setup test environment with stale sessions
sbs clean --stale --dry-run
sbs clean --stale --force
```
- **Expected**: Identical behavior to pre-refactor
- **Validation**: 
  - Correct message output
  - Same session identification logic
  - Same cleanup operations performed

##### Test Case: Combined Stale and Branch Cleanup (`--stale --branches`)
```bash
# Setup test environment with stale sessions and orphaned branches
sbs clean --stale --branches --dry-run
sbs clean --stale --branches --force
```
- **Expected**: Both cleanup operations execute in sequence
- **Validation**:
  - Stale cleanup executes first with appropriate message
  - Branch cleanup executes second
  - No regression in error handling

##### Test Case: Comprehensive Cleanup (`--all`)
```bash
# Setup comprehensive test environment
sbs clean --all --dry-run
sbs clean --all --force
```
- **Expected**: All cleanup types execute including stale cleanup
- **Validation**:
  - Comprehensive message appears
  - Stale cleanup message appears within comprehensive flow
  - All cleanup operations complete successfully

#### 2.2 Error Handling Integration
- **Test**: Error handling at each refactored call site
- **Method**: Inject errors into `executeDefaultCleanup` and verify proper handling
- **Expected**: Error propagation works identically to pre-refactor

### 3. Regression Testing

#### 3.1 Behavioral Consistency Testing
Run comprehensive comparison tests between pre-refactor and post-refactor behavior:

##### Core Functionality Tests
```bash
# Test matrix covering all flag combinations
sbs clean --dry-run
sbs clean --force
sbs clean --stale --dry-run
sbs clean --stale --force
sbs clean --stale --branches --dry-run
sbs clean --stale --branches --force
sbs clean --all --dry-run
sbs clean --all --force
```

##### Output Format Consistency
- **Test**: Compare stdout/stderr output before and after refactor
- **Method**: Capture and diff outputs for identical test scenarios
- **Expected**: Messages should be identical except for any intentional improvements

##### Session Management Consistency
- **Test**: Verify session persistence and cleanup tracking unchanged
- **Method**: Compare session state files before/after cleanup operations
- **Expected**: Identical session management behavior

#### 3.2 Performance Regression Testing
- **Test**: Verify no performance degradation from refactor
- **Method**: Benchmark cleanup operations pre/post refactor
- **Expected**: Performance should be identical or improved (less function call overhead)

### 4. Edge Case Testing

#### 4.1 Empty Session Scenarios
- **Test**: Clean command behavior when no sessions exist
- **Scenarios**:
  - No sessions at all: `sbs clean --stale`
  - No stale sessions: `sbs clean --stale`
  - No orphaned branches: `sbs clean --stale --branches`
- **Expected**: Appropriate "nothing to clean" messages

#### 4.2 Error Condition Edge Cases
- **Test**: Handle errors in dependency components
- **Scenarios**:
  - Config loading failures
  - Git repository access failures
  - Tmux session detection failures
  - File system permission errors
- **Expected**: Same error handling and user messaging as before refactor

#### 4.3 Concurrent Access Edge Cases
- **Test**: Multiple cleanup operations or concurrent session access
- **Method**: Simulate concurrent `sbs clean` executions
- **Expected**: No new race conditions introduced by refactor

#### 4.4 Large Dataset Edge Cases
- **Test**: Cleanup with large numbers of sessions/branches
- **Method**: Generate test environments with 100+ stale sessions
- **Expected**: Performance and memory usage remain consistent

### 5. Validation Commands

#### 5.1 Pre-Refactor Baseline
```bash
# Capture baseline behavior
make test 2>&1 | tee pre-refactor-tests.log
make build
./sbs clean --stale --dry-run 2>&1 | tee pre-refactor-stale-dry.log
./sbs clean --stale --branches --dry-run 2>&1 | tee pre-refactor-combined-dry.log
./sbs clean --all --dry-run 2>&1 | tee pre-refactor-all-dry.log
```

#### 5.2 Post-Refactor Validation
```bash
# Run full test suite
make test
go test ./cmd/... -v
go test ./pkg/... -v

# Specific clean command testing
make build

# Test basic functionality
./sbs clean --help
./sbs clean --dry-run

# Test refactored call sites
./sbs clean --stale --dry-run
./sbs clean --stale --branches --dry-run  
./sbs clean --all --dry-run

# Test with force flag
./sbs clean --stale --force --dry-run
./sbs clean --all --force --dry-run

# Integration test with real sessions (if available)
# Create test session: sbs start test:cleanup-validation
./sbs clean --stale --dry-run
# Stop test session and rerun: sbs stop test:cleanup-validation
./sbs clean --stale --dry-run
```

#### 5.3 Automated Test Commands
```bash
# Unit tests for clean command
go test ./cmd/clean_test.go -v

# Integration tests
go test ./integration_test.go -v -run TestClean

# Cleanup package tests (dependency verification)
go test ./pkg/cleanup/... -v

# Build verification
make build
make lint
make fmt

# Race condition detection
go test -race ./cmd/...
```

### 6. Test Environment Setup

#### 6.1 Test Data Preparation
Create standardized test environments for consistent validation:

```bash
# Create test sessions for cleanup validation
sbs start test:stale-session-1
sbs start test:stale-session-2
sbs start test:active-session

# Stop some sessions to make them stale
sbs stop test:stale-session-1
sbs stop test:stale-session-2

# Create orphaned git branches
git checkout -b issue-999-orphaned-branch
git checkout main
```

#### 6.2 Test Cleanup
Ensure clean test environment between runs:

```bash
# Clean up test sessions
sbs clean --all --force

# Reset git branches
git branch -D issue-999-orphaned-branch 2>/dev/null || true

# Clean up any test artifacts
rm -f .sbs/test-sessions.json
```

### 7. Success Criteria

The refactor is considered successful when:

1. **All existing tests pass** without modification (except tests specifically testing the removed function)
2. **No behavioral changes** in user-visible functionality
3. **All three call sites** correctly call `executeDefaultCleanup` with appropriate messaging
4. **Error handling** remains identical to pre-refactor behavior
5. **Performance** is identical or improved
6. **Code coverage** is maintained or improved
7. **Integration tests** pass with real session data
8. **Manual testing** confirms expected output messages and cleanup behavior

### 8. Rollback Plan

If any tests fail or regressions are discovered:

1. **Immediate rollback**: Revert the refactor commits
2. **Analysis**: Identify specific failure points
3. **Targeted fixes**: Address specific issues while maintaining refactor goals
4. **Re-validation**: Re-run full test suite after fixes

### 9. Post-Deployment Validation

After the refactor is deployed:

1. **Monitor** for any user-reported issues with cleanup functionality
2. **Verify** that cleanup operations work correctly in production environments  
3. **Validate** that all cleanup modes continue to work as expected
4. **Confirm** that error messages and user experience remain consistent

## Implementation Notes

- **Test Coverage**: Ensure test coverage for cleanup functionality doesn't decrease
- **Documentation**: Update any relevant documentation to reflect function name changes
- **Code Comments**: Update comments that reference the removed function
- **Backward Compatibility**: Ensure no breaking changes to public interfaces