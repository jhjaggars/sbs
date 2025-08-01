# Testing Plan: Issue #6 - Improve Resource Tracking and Cleanup for Comprehensive Clean Command

## Overview

This testing plan covers the comprehensive testing strategy for implementing enhanced resource tracking and cleanup functionality in the SBS (Sandbox Sessions) CLI application. The enhancement addresses identified gaps in resource management including branch cleanup, partial failure recovery, resource discovery, and atomic operations for improved reliability.

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
├── config/
│   ├── config.go                       # Updated SessionMetadata with resource tracking
│   ├── config_test.go                  # Updated configuration tests
│   └── resource_tracking_test.go       # New tests for resource tracking fields
├── git/
│   ├── manager.go                      # Updated with branch cleanup methods
│   ├── manager_test.go                 # Updated with branch cleanup tests
│   └── branch_cleanup_test.go          # New tests for branch cleanup functionality
├── sandbox/
│   ├── manager.go                      # Updated with enhanced discovery
│   ├── manager_test.go                 # Updated sandbox tests
│   └── discovery_test.go               # New tests for orphaned resource discovery
cmd/
├── clean.go                           # Enhanced cleanup modes and branch removal
├── clean_test.go                      # New comprehensive clean command tests
├── start.go                           # Updated with atomic resource tracking
├── start_test.go                      # Updated with atomic operation tests
├── start_rollback_test.go             # New tests for failure rollback
├── stop.go                            # Updated with optional branch cleanup
├── stop_test.go                       # Updated stop tests
├── healthcheck.go                     # New health check command
└── healthcheck_test.go                # New health check tests
integration/
├── resource_tracking_test.go          # Integration tests for resource tracking
├── atomic_operations_test.go          # Integration tests for atomic operations
├── cleanup_scenarios_test.go          # Integration tests for cleanup scenarios
└── failure_recovery_test.go           # Integration tests for failure recovery
```

## Testing Priority and Implementation Order

### Phase 1: Foundation Tests (Implement First)
1. **Resource Tracking Unit Tests** - Test enhanced SessionMetadata structure
2. **Branch Management Unit Tests** - Test git branch cleanup methods
3. **Configuration Tests** - Test new resource tracking configuration
4. **Mock Framework Setup** - Establish comprehensive mocking for external dependencies

### Phase 2: Component Tests
5. **Atomic Operations Tests** - Test rollback mechanisms in start command
6. **Enhanced Clean Command Tests** - Test new cleanup modes and functionality
7. **Health Check Command Tests** - Test resource validation and discovery
8. **Orphaned Resource Discovery Tests** - Test detection of abandoned resources

### Phase 3: Integration and System Tests
9. **End-to-End Workflow Tests** - Test complete resource lifecycle
10. **Failure Recovery Tests** - Test partial failure scenarios and recovery
11. **Cross-Command Integration Tests** - Test interaction between commands
12. **Performance Impact Tests** - Validate no significant performance regression

---

## 1. Unit Tests

### 1.1 Enhanced Configuration Tests (`pkg/config/resource_tracking_test.go`)

#### Resource Tracking Metadata Tests

```go
func TestSessionMetadata_ResourceTracking(t *testing.T) {
    t.Run("resource_creation_log_initialization", func(t *testing.T) {
        // Test initialization of ResourceCreationLog field
        // Verify empty slice is properly initialized
        // Assert proper JSON serialization
    })
    
    t.Run("resource_creation_log_operations", func(t *testing.T) {
        // Test adding resource creation entries
        // Verify chronological ordering
        // Assert proper data structure integrity
    })
    
    t.Run("resource_status_tracking", func(t *testing.T) {
        // Test ResourceStatus field updates
        // Verify status transitions (creating, active, cleanup, failed)
        // Assert status validation
    })
    
    t.Run("creation_step_tracking", func(t *testing.T) {
        // Test CurrentCreationStep field
        // Verify step progression tracking
        // Assert rollback step identification
    })
    
    t.Run("failure_point_tracking", func(t *testing.T) {
        // Test FailurePoint field for partial failures
        // Verify failure context capture
        // Assert recovery information storage
    })
}

func TestSessionMetadata_BackwardCompatibility(t *testing.T) {
    t.Run("load_legacy_session_metadata", func(t *testing.T) {
        // Test loading sessions without new resource tracking fields
        // Verify default values are applied appropriately
        // Assert existing functionality is preserved
    })
    
    t.Run("migrate_existing_sessions", func(t *testing.T) {
        // Test migration of existing session metadata
        // Verify new fields are properly initialized
        // Assert no data loss during migration
    })
}

func TestSessionMetadata_ResourceCreationEntry(t *testing.T) {
    t.Run("resource_entry_creation", func(t *testing.T) {
        // Test creation of ResourceCreationEntry structs
        // Verify all required fields are populated
        // Assert proper timestamp handling
    })
    
    t.Run("resource_entry_serialization", func(t *testing.T) {
        // Test JSON serialization of resource entries
        // Verify proper field mapping
        // Assert deserialization accuracy
    })
}
```

#### Test Utilities for Resource Tracking

```go
// pkg/config/test_helpers.go
type MockResourceEntry struct {
    ResourceType string
    ResourceID   string
    CreatedAt    time.Time
    Status       string
    Metadata     map[string]interface{}
}

func CreateTestSessionWithResources(issueNumber int, resources []MockResourceEntry) *SessionMetadata {
    session := &SessionMetadata{
        IssueNumber:         issueNumber,
        ResourceCreationLog: make([]ResourceCreationEntry, 0),
        ResourceStatus:      "active",
        CurrentCreationStep: "completed",
    }
    
    for _, resource := range resources {
        entry := ResourceCreationEntry{
            ResourceType: resource.ResourceType,
            ResourceID:   resource.ResourceID,
            CreatedAt:    resource.CreatedAt,
            Status:       resource.Status,
            Metadata:     resource.Metadata,
        }
        session.ResourceCreationLog = append(session.ResourceCreationLog, entry)
    }
    
    return session
}

func CreateFailedSessionAtStep(issueNumber int, failureStep string) *SessionMetadata {
    return &SessionMetadata{
        IssueNumber:         issueNumber,
        ResourceStatus:      "failed",
        CurrentCreationStep: failureStep,
        FailurePoint:        failureStep,
        ResourceCreationLog: []ResourceCreationEntry{},
    }
}
```

### 1.2 Git Manager Branch Cleanup Tests (`pkg/git/branch_cleanup_test.go`)

#### Branch Management Functionality Tests

```go
func TestGitManager_BranchCleanup(t *testing.T) {
    t.Run("delete_issue_branch_success", func(t *testing.T) {
        // Test successful branch deletion
        // Mock git branch -d command
        // Verify branch deletion is called with correct parameters
    })
    
    t.Run("delete_issue_branch_force", func(t *testing.T) {
        // Test force branch deletion
        // Mock git branch -D command for unmerged branches
        // Verify force deletion when branch has unmerged changes
    })
    
    t.Run("delete_nonexistent_branch", func(t *testing.T) {
        // Test deletion of non-existent branch
        // Mock git command returning "branch not found" error
        // Verify graceful handling without failing operation
    })
    
    t.Run("delete_current_branch_protection", func(t *testing.T) {
        // Test protection against deleting current branch
        // Mock git current branch detection
        // Verify error when attempting to delete current branch
    })
    
    t.Run("delete_branch_with_worktree_attached", func(t *testing.T) {
        // Test branch deletion when worktree is still attached
        // Mock git branch deletion failure
        // Verify proper error handling and user guidance
    })
}

func TestGitManager_BranchDiscovery(t *testing.T) {
    t.Run("list_issue_branches", func(t *testing.T) {
        // Test discovery of issue-* branches
        // Mock git branch listing
        // Verify proper branch pattern matching
    })
    
    t.Run("identify_orphaned_branches", func(t *testing.T) {
        // Test identification of branches without active sessions
        // Mock branch listing and session metadata
        // Verify orphaned branch detection accuracy
    })
    
    t.Run("branch_age_calculation", func(t *testing.T) {
        // Test calculation of branch age for cleanup decisions
        // Mock git log for last commit timestamp
        // Verify age calculation accuracy
    })
}

func TestGitManager_BranchValidation(t *testing.T) {
    t.Run("validate_branch_safety_for_deletion", func(t *testing.T) {
        // Test validation before branch deletion
        // Mock git status and merge checks
        // Verify safety validations are performed
    })
    
    t.Run("detect_unmerged_changes", func(t *testing.T) {
        // Test detection of unmerged branch changes
        // Mock git merge-base and diff operations
        // Verify unmerged change detection
    })
}
```

### 1.3 Enhanced Clean Command Tests (`cmd/clean_test.go`)

#### Multi-Mode Cleanup Functionality Tests

```go
func TestCleanCommand_EnhancedModes(t *testing.T) {
    t.Run("clean_stale_sessions_mode", func(t *testing.T) {
        // Test --stale flag for cleaning stale sessions only
        // Mock stale session detection
        // Verify only stale sessions are targeted
    })
    
    t.Run("clean_orphaned_resources_mode", func(t *testing.T) {
        // Test --orphaned flag for cleaning orphaned resources
        // Mock orphaned resource discovery
        // Verify orphaned resource cleanup
    })
    
    t.Run("clean_branches_mode", func(t *testing.T) {
        // Test --branches flag for branch cleanup
        // Mock branch discovery and deletion
        // Verify branch cleanup functionality
    })
    
    t.Run("clean_all_mode", func(t *testing.T) {
        // Test --all flag for comprehensive cleanup
        // Mock all cleanup operations
        // Verify all resource types are cleaned
    })
    
    t.Run("selective_cleanup_combinations", func(t *testing.T) {
        // Test combinations of cleanup flags
        // Mock different resource states
        // Verify proper flag combination handling
    })
}

func TestCleanCommand_BranchCleanup(t *testing.T) {
    t.Run("branch_cleanup_integration", func(t *testing.T) {
        // Test integration with git manager for branch cleanup
        // Mock branch discovery and deletion operations
        // Verify proper interaction between clean and git manager
    })
    
    t.Run("branch_cleanup_safety_checks", func(t *testing.T) {
        // Test safety checks before branch deletion
        // Mock unmerged branch scenarios
        // Verify user confirmation for risky deletions
    })
    
    t.Run("branch_cleanup_error_handling", func(t *testing.T) {
        // Test error handling during branch cleanup
        // Mock git command failures
        // Verify graceful error handling and user feedback
    })
}

func TestCleanCommand_ImprovedErrorHandling(t *testing.T) {
    t.Run("partial_cleanup_failure_handling", func(t *testing.T) {
        // Test handling when some cleanup operations fail
        // Mock mixed success/failure scenarios
        // Verify partial success reporting
    })
    
    t.Run("resource_dependency_handling", func(t *testing.T) {
        // Test cleanup order to handle resource dependencies
        // Mock resources with dependencies (worktree -> branch)
        // Verify proper cleanup ordering
    })
    
    t.Run("user_confirmation_handling", func(t *testing.T) {
        // Test interactive user confirmation for risky operations
        // Mock user input scenarios
        // Verify proper confirmation handling
    })
}
```

### 1.4 Atomic Start Command Tests (`cmd/start_rollback_test.go`)

#### Atomic Operations and Rollback Tests

```go
func TestStartCommand_AtomicOperations(t *testing.T) {
    t.Run("successful_atomic_resource_creation", func(t *testing.T) {
        // Test successful creation of all resources
        // Mock all external tool successes
        // Verify resource creation log is properly maintained
    })
    
    t.Run("rollback_on_branch_creation_failure", func(t *testing.T) {
        // Test rollback when branch creation fails
        // Mock git branch creation failure
        // Verify no resources are left behind
    })
    
    t.Run("rollback_on_worktree_creation_failure", func(t *testing.T) {
        // Test rollback when worktree creation fails
        // Mock git worktree failure after branch success
        // Verify branch is cleaned up
    })
    
    t.Run("rollback_on_tmux_session_failure", func(t *testing.T) {
        // Test rollback when tmux session creation fails
        // Mock tmux failure after git operations success
        // Verify git resources are cleaned up
    })
    
    t.Run("rollback_on_sandbox_creation_failure", func(t *testing.T) {
        // Test rollback when sandbox creation fails
        // Mock sandbox failure after other successes
        // Verify all previous resources are cleaned up
    })
}

func TestStartCommand_ResourceTracking(t *testing.T) {
    t.Run("resource_creation_logging", func(t *testing.T) {
        // Test logging of each resource creation step
        // Mock successful resource creation
        // Verify resource creation log entries
    })
    
    t.Run("creation_step_progression", func(t *testing.T) {
        // Test progression through creation steps
        // Mock step-by-step resource creation
        // Verify CurrentCreationStep field updates
    })
    
    t.Run("failure_point_capture", func(t *testing.T) {
        // Test capture of failure point for debugging
        // Mock failure at various steps
        // Verify FailurePoint field is set correctly
    })
    
    t.Run("resource_metadata_storage", func(t *testing.T) {
        // Test storage of resource metadata for tracking
        // Mock resource creation with metadata
        // Verify metadata is stored in resource entries
    })
}
```

### 1.5 Health Check Command Tests (`cmd/healthcheck_test.go`)

#### Resource Validation and Discovery Tests

```go
func TestHealthCheckCommand_ResourceValidation(t *testing.T) {
    t.Run("validate_session_resources_healthy", func(t *testing.T) {
        // Test validation of healthy session resources
        // Mock all resources existing and accessible
        // Verify health check passes
    })
    
    t.Run("detect_missing_worktree", func(t *testing.T) {
        // Test detection of missing worktree
        // Mock worktree directory not existing
        // Verify health check identifies issue
    })
    
    t.Run("detect_stale_tmux_session", func(t *testing.T) {
        // Test detection of stale tmux session references
        // Mock tmux session not existing
        // Verify health check identifies stale reference
    })
    
    t.Run("detect_orphaned_sandbox", func(t *testing.T) {
        // Test detection of orphaned sandbox containers
        // Mock sandbox existing without session metadata
        // Verify orphaned resource identification
    })
    
    t.Run("detect_git_worktree_registry_mismatch", func(t *testing.T) {
        // Test detection of git worktree registry inconsistencies
        // Mock worktree not registered in git
        // Verify registry mismatch detection
    })
}

func TestHealthCheckCommand_ComprehensiveDiscovery(t *testing.T) {
    t.Run("discover_all_orphaned_resources", func(t *testing.T) {
        // Test comprehensive orphaned resource discovery
        // Mock various orphaned resource scenarios
        // Verify all orphaned resources are identified
    })
    
    t.Run("discover_inconsistent_metadata", func(t *testing.T) {
        // Test discovery of inconsistent session metadata
        // Mock metadata that doesn't match actual resources
        // Verify inconsistency detection
    })
    
    t.Run("validate_resource_dependencies", func(t *testing.T) {
        // Test validation of resource dependencies
        // Mock missing dependency scenarios
        // Verify dependency validation
    })
}

func TestHealthCheckCommand_ReportGeneration(t *testing.T) {
    t.Run("generate_comprehensive_health_report", func(t *testing.T) {
        // Test generation of detailed health report
        // Mock various health issues
        // Verify report contains all identified issues
    })
    
    t.Run("suggest_remediation_actions", func(t *testing.T) {
        // Test suggestion of remediation actions
        // Mock various health issues
        // Verify appropriate remediation suggestions
    })
    
    t.Run("health_report_formatting", func(t *testing.T) {
        // Test health report formatting and readability
        // Mock various health states
        // Verify report is well-formatted and actionable
    })
}
```

---

## 2. Integration Tests

### 2.1 Resource Lifecycle Integration Tests (`integration/resource_tracking_test.go`)

#### End-to-End Resource Management Tests

```go
func TestResourceLifecycle_CompleteWorkflow(t *testing.T) {
    t.Run("full_lifecycle_with_resource_tracking", func(t *testing.T) {
        // Test complete resource lifecycle from start to clean
        // Mock all external dependencies
        // Verify resource tracking throughout lifecycle
        // Assert proper cleanup at end
    })
    
    t.Run("session_resume_with_resource_validation", func(t *testing.T) {
        // Test session resume with resource validation
        // Mock existing session with tracked resources
        // Verify resource validation during resume
        // Assert session resumes successfully
    })
    
    t.Run("cross_command_resource_consistency", func(t *testing.T) {
        // Test resource consistency across commands
        // Execute start, list, stop, clean sequence
        // Verify resource tracking remains consistent
        // Assert no resource leaks
    })
}

func TestResourceLifecycle_FailureRecovery(t *testing.T) {
    t.Run("recovery_from_partial_creation_failure", func(t *testing.T) {
        // Test recovery from partial resource creation failure
        // Mock failure during resource creation
        // Execute health check and clean operations
        // Verify proper cleanup and recovery
    })
    
    t.Run("recovery_from_metadata_corruption", func(t *testing.T) {
        // Test recovery from corrupted session metadata
        // Mock corrupted metadata scenarios
        // Execute health check for detection
        // Verify recovery mechanisms
    })
}
```

### 2.2 Atomic Operations Integration Tests (`integration/atomic_operations_test.go`)

#### Atomic Operation and Rollback Integration Tests

```go
func TestAtomicOperations_Integration(t *testing.T) {
    t.Run("atomic_start_with_all_dependencies", func(t *testing.T) {
        // Test atomic start with all real dependencies
        // Mock external tools (git, tmux, sandbox, gh)
        // Verify atomic behavior across all components
        // Assert proper rollback on any failure
    })
    
    t.Run("rollback_mechanism_integration", func(t *testing.T) {
        // Test rollback mechanism integration
        // Mock failure at various points in start process
        // Verify rollback cleans up all created resources
        // Assert system state is properly restored
    })
    
    t.Run("resource_cleanup_ordering", func(t *testing.T) {
        // Test proper ordering of resource cleanup during rollback
        // Mock resources with dependencies
        // Verify cleanup occurs in correct dependency order
        // Assert no orphaned resources remain
    })
}
```

### 2.3 Cleanup Scenarios Integration Tests (`integration/cleanup_scenarios_test.go`)

#### Comprehensive Cleanup Scenario Tests

```go
func TestCleanupScenarios_Integration(t *testing.T) {
    t.Run("mixed_resource_states_cleanup", func(t *testing.T) {
        // Test cleanup with mixed resource states
        // Setup sessions with various resource states
        // Execute comprehensive clean operation
        // Verify appropriate cleanup based on resource states
    })
    
    t.Run("cleanup_with_user_interaction", func(t *testing.T) {
        // Test cleanup scenarios requiring user interaction
        // Mock user confirmation scenarios
        // Verify proper user interaction handling
        // Assert cleanup respects user choices
    })
    
    t.Run("cleanup_performance_with_large_datasets", func(t *testing.T) {
        // Test cleanup performance with many sessions
        // Setup large number of sessions and resources
        // Execute cleanup operations
        // Verify acceptable performance characteristics
    })
}
```

---

## 3. Edge Case Tests

### 3.1 Resource State Edge Cases (`test/resource_edge_cases_test.go`)

#### Complex Resource State Scenarios

```go
func TestResourceEdgeCases_ComplexStates(t *testing.T) {
    t.Run("partially_created_resources", func(t *testing.T) {
        // Test handling of partially created resources
        // Mock interrupted resource creation
        // Verify detection and appropriate handling
        // Assert no resource leaks
    })
    
    t.Run("resources_modified_externally", func(t *testing.T) {
        // Test handling of resources modified outside SBS
        // Mock external modification of worktrees, branches, etc.
        // Verify detection of external changes
        // Assert appropriate handling of modified resources
    })
    
    t.Run("concurrent_resource_access", func(t *testing.T) {
        // Test handling of concurrent access to resources
        // Mock multiple SBS processes accessing same resources
        // Verify proper locking and conflict resolution
        // Assert data integrity is maintained
    })
    
    t.Run("resource_permission_issues", func(t *testing.T) {
        // Test handling of resource permission issues
        // Mock permission denied scenarios
        // Verify graceful error handling
        // Assert meaningful error messages
    })
}

func TestResourceEdgeCases_MetadataCorruption(t *testing.T) {
    t.Run("corrupted_session_metadata", func(t *testing.T) {
        // Test handling of corrupted session metadata
        // Mock various corruption scenarios
        // Verify detection and recovery mechanisms
        // Assert system remains stable
    })
    
    t.Run("missing_resource_creation_log", func(t *testing.T) {
        // Test handling of missing resource creation logs
        // Mock sessions without creation logs
        // Verify fallback resource discovery
        // Assert resources are still manageable
    })
    
    t.Run("inconsistent_resource_metadata", func(t *testing.T) {
        // Test handling of inconsistent resource metadata
        // Mock metadata that doesn't match actual resources
        // Verify detection and reconciliation
        // Assert consistency is restored
    })
}
```

### 3.2 Git Repository Edge Cases (`test/git_edge_cases_test.go`)

#### Git-Specific Edge Case Scenarios

```go
func TestGitEdgeCases_RepositoryStates(t *testing.T) {
    t.Run("bare_repository_handling", func(t *testing.T) {
        // Test handling of bare Git repositories
        // Mock bare repository scenarios
        // Verify appropriate error handling
        // Assert clear error messages for unsupported operations
    })
    
    t.Run("repository_corruption_handling", func(t *testing.T) {
        // Test handling of corrupted Git repositories
        // Mock repository corruption scenarios
        // Verify detection and graceful failure
        // Assert no data loss or corruption spread
    })
    
    t.Run("worktree_registry_corruption", func(t *testing.T) {
        // Test handling of corrupted Git worktree registry
        // Mock worktree registry corruption
        // Verify detection and recovery mechanisms
        // Assert registry can be repaired
    })
    
    t.Run("branch_protection_rules", func(t *testing.T) {
        // Test handling of branch protection rules
        // Mock protected branches
        // Verify appropriate handling of protected branches
        // Assert protection rules are respected
    })
}
```

---

## 4. Performance Tests

### 4.1 Resource Tracking Performance (`test/performance_test.go`)

#### Performance Impact Assessment

```go
func TestPerformance_ResourceTracking(t *testing.T) {
    t.Run("resource_tracking_overhead", func(t *testing.T) {
        // Test performance overhead of resource tracking
        // Benchmark operations with and without tracking
        // Verify overhead is within acceptable limits
        // Assert no significant performance regression
    })
    
    t.Run("large_session_metadata_performance", func(t *testing.T) {
        // Test performance with large session metadata
        // Mock sessions with extensive resource logs
        // Verify acceptable performance characteristics
        // Assert scalability with growing metadata
    })
    
    t.Run("cleanup_performance_scaling", func(t *testing.T) {
        // Test cleanup performance scaling
        // Mock increasing numbers of sessions and resources
        // Verify cleanup performance scales appropriately
        // Assert no exponential performance degradation
    })
}

func BenchmarkResourceOperations(b *testing.B) {
    b.Run("resource_creation_tracking", func(b *testing.B) {
        // Benchmark resource creation tracking
        // Measure overhead of tracking operations
    })
    
    b.Run("resource_discovery_scanning", func(b *testing.B) {
        // Benchmark resource discovery scanning
        // Measure time to scan for orphaned resources
    })
    
    b.Run("health_check_validation", func(b *testing.B) {
        // Benchmark health check validation
        // Measure time to validate all resources
    })
}
```

---

## 5. Security and Safety Tests

### 5.1 Resource Safety Tests (`test/security_test.go`)

#### Resource Safety and Security Validation

```go
func TestResourceSafety_SecurityValidation(t *testing.T) {
    t.Run("path_traversal_protection", func(t *testing.T) {
        // Test protection against path traversal attacks
        // Mock malicious path inputs
        // Verify path sanitization and validation
        // Assert no access outside intended directories
    })
    
    t.Run("resource_permission_validation", func(t *testing.T) {
        // Test resource permission validation
        // Mock insufficient permission scenarios
        // Verify proper permission checking
        // Assert no privilege escalation
    })
    
    t.Run("symbolic_link_handling", func(t *testing.T) {
        // Test handling of symbolic links in resource paths
        // Mock symbolic link scenarios
        // Verify safe symbolic link handling
        // Assert no security vulnerabilities
    })
    
    t.Run("resource_cleanup_safety", func(t *testing.T) {
        // Test safety of resource cleanup operations
        // Mock various resource states
        // Verify safe cleanup without data loss
        // Assert no accidental deletion of user data
    })
}
```

---

## 6. Test Data and Fixtures

### 6.1 Resource Test Scenarios

#### Test Data for Resource Management Scenarios

```go
// test/fixtures/resource_scenarios.go
var ResourceTestScenarios = []ResourceTestScenario{
    {
        Name: "complete_healthy_session",
        Session: SessionMetadata{
            IssueNumber:   123,
            IssueTitle:    "Fix authentication bug",
            Branch:        "issue-123-fix-authentication-bug",
            WorktreePath:  "/tmp/test-worktree-123",
            TmuxSession:   "work-issue-123",
            SandboxName:   "work-issue-123",
            ResourceStatus: "active",
            ResourceCreationLog: []ResourceCreationEntry{
                {ResourceType: "branch", ResourceID: "issue-123-fix-authentication-bug", Status: "created"},
                {ResourceType: "worktree", ResourceID: "/tmp/test-worktree-123", Status: "created"},
                {ResourceType: "tmux", ResourceID: "work-issue-123", Status: "created"},
                {ResourceType: "sandbox", ResourceID: "work-issue-123", Status: "created"},
            },
        },
        ExpectedHealth: "healthy",
    },
    {
        Name: "partially_failed_session",
        Session: SessionMetadata{
            IssueNumber:         456,
            IssueTitle:          "Add dark mode support",
            Branch:              "issue-456-add-dark-mode-support",
            WorktreePath:        "/tmp/test-worktree-456",
            TmuxSession:         "work-issue-456",
            ResourceStatus:      "failed",
            CurrentCreationStep: "tmux_creation",
            FailurePoint:        "tmux_creation",
            ResourceCreationLog: []ResourceCreationEntry{
                {ResourceType: "branch", ResourceID: "issue-456-add-dark-mode-support", Status: "created"},
                {ResourceType: "worktree", ResourceID: "/tmp/test-worktree-456", Status: "created"},
                {ResourceType: "tmux", ResourceID: "work-issue-456", Status: "failed"},
            },
        },
        ExpectedHealth: "partially_failed",
    },
    {
        Name: "orphaned_resources_session",
        Session: SessionMetadata{
            IssueNumber:   789,
            IssueTitle:    "Refactor database layer",
            Branch:        "issue-789-refactor-database-layer",
            WorktreePath:  "/tmp/test-worktree-789",
            TmuxSession:   "work-issue-789",
            SandboxName:   "work-issue-789",
            ResourceStatus: "stale",
            ResourceCreationLog: []ResourceCreationEntry{
                {ResourceType: "branch", ResourceID: "issue-789-refactor-database-layer", Status: "created"},
                {ResourceType: "worktree", ResourceID: "/tmp/test-worktree-789", Status: "orphaned"},
                {ResourceType: "sandbox", ResourceID: "work-issue-789", Status: "orphaned"},
            },
        },
        ExpectedHealth: "has_orphaned_resources",
    },
}

var CleanupTestScenarios = []CleanupTestScenario{
    {
        Name: "mixed_resource_states",
        Sessions: []SessionMetadata{
            // Healthy session - should not be cleaned
            ResourceTestScenarios[0].Session,
            // Failed session - should be cleaned up
            ResourceTestScenarios[1].Session,
            // Session with orphaned resources - should be cleaned
            ResourceTestScenarios[2].Session,
        },
        CleanupMode: "comprehensive",
        ExpectedCleaned: []int{456, 789}, // Issue numbers that should be cleaned
        ExpectedRetained: []int{123},     // Issue numbers that should be retained
    },
}
```

### 6.2 Mock External Dependencies

#### Comprehensive Mock Setup for External Tools

```go
// test/mocks/external_tools.go
type MockGitManager struct {
    mock.Mock
    Branches       []string
    Worktrees      []string
    OperationLog   []string
    ShouldFailOn   map[string]bool
}

func (m *MockGitManager) CreateIssueBranch(issueNumber int, issueTitle string) (string, error) {
    args := m.Called(issueNumber, issueTitle)
    if m.ShouldFailOn["CreateIssueBranch"] {
        return "", args.Error(1)
    }
    branchName := fmt.Sprintf("issue-%d-%s", issueNumber, slug.Make(issueTitle))
    m.Branches = append(m.Branches, branchName)
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("created_branch:%s", branchName))
    return branchName, args.Error(1)
}

func (m *MockGitManager) DeleteIssueBranch(branchName string) error {
    args := m.Called(branchName)
    if m.ShouldFailOn["DeleteIssueBranch"] {
        return args.Error(0)
    }
    // Remove branch from mock branches list
    for i, branch := range m.Branches {
        if branch == branchName {
            m.Branches = append(m.Branches[:i], m.Branches[i+1:]...)
            break
        }
    }
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("deleted_branch:%s", branchName))
    return args.Error(0)
}

func (m *MockGitManager) CreateWorktree(branchName string, worktreePath string) error {
    args := m.Called(branchName, worktreePath)
    if m.ShouldFailOn["CreateWorktree"] {
        return args.Error(0)
    }
    m.Worktrees = append(m.Worktrees, worktreePath)
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("created_worktree:%s", worktreePath))
    return args.Error(0)
}

func (m *MockGitManager) RemoveWorktree(worktreePath string) error {
    args := m.Called(worktreePath)
    if m.ShouldFailOn["RemoveWorktree"] {
        return args.Error(0)
    }
    // Remove worktree from mock worktrees list
    for i, worktree := range m.Worktrees {
        if worktree == worktreePath {
            m.Worktrees = append(m.Worktrees[:i], m.Worktrees[i+1:]...)
            break
        }
    }
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("removed_worktree:%s", worktreePath))
    return args.Error(0)
}

func (m *MockGitManager) ListOrphanedBranches() ([]string, error) {
    args := m.Called()
    if m.ShouldFailOn["ListOrphanedBranches"] {
        return nil, args.Error(1)
    }
    // Return mock orphaned branches
    return []string{"issue-999-orphaned-branch"}, args.Error(1)
}

type MockTmuxManager struct {
    mock.Mock
    Sessions     []string
    OperationLog []string
    ShouldFailOn map[string]bool
}

func (m *MockTmuxManager) CreateSession(issueNumber int, workingDir string, sessionName string, env ...map[string]string) (*TmuxSession, error) {
    args := m.Called(issueNumber, workingDir, sessionName, env)
    if m.ShouldFailOn["CreateSession"] {
        return nil, args.Error(1)
    }
    session := &TmuxSession{
        Name:        sessionName,
        WorkingDir:  workingDir,
        IssueNumber: issueNumber,
    }
    m.Sessions = append(m.Sessions, sessionName)
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("created_session:%s", sessionName))
    return session, args.Error(1)
}

func (m *MockTmuxManager) KillSession(sessionName string) error {
    args := m.Called(sessionName)
    if m.ShouldFailOn["KillSession"] {
        return args.Error(0)
    }
    // Remove session from mock sessions list
    for i, session := range m.Sessions {
        if session == sessionName {
            m.Sessions = append(m.Sessions[:i], m.Sessions[i+1:]...)
            break
        }
    }
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("killed_session:%s", sessionName))
    return args.Error(0)
}

func (m *MockTmuxManager) SessionExists(sessionName string) (bool, error) {
    args := m.Called(sessionName)
    if m.ShouldFailOn["SessionExists"] {
        return false, args.Error(1)
    }
    for _, session := range m.Sessions {
        if session == sessionName {
            return true, args.Error(1)
        }
    }
    return false, args.Error(1)
}

type MockSandboxManager struct {
    mock.Mock
    Sandboxes    []string
    OperationLog []string
    ShouldFailOn map[string]bool
}

func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
    args := m.Called(sandboxName)
    if m.ShouldFailOn["SandboxExists"] {
        return false, args.Error(1)
    }
    for _, sandbox := range m.Sandboxes {
        if sandbox == sandboxName {
            return true, args.Error(1)
        }
    }
    return false, args.Error(1)
}

func (m *MockSandboxManager) DeleteSandbox(sandboxName string) error {
    args := m.Called(sandboxName)
    if m.ShouldFailOn["DeleteSandbox"] {
        return args.Error(0)
    }
    // Remove sandbox from mock sandboxes list
    for i, sandbox := range m.Sandboxes {
        if sandbox == sandboxName {
            m.Sandboxes = append(m.Sandboxes[:i], m.Sandboxes[i+1:]...)
            break
        }
    }
    m.OperationLog = append(m.OperationLog, fmt.Sprintf("deleted_sandbox:%s", sandboxName))
    return args.Error(0)
}

func (m *MockSandboxManager) ListOrphanedSandboxes() ([]string, error) {
    args := m.Called()
    if m.ShouldFailOn["ListOrphanedSandboxes"] {
        return nil, args.Error(1)
    }
    // Return mock orphaned sandboxes
    return []string{"work-issue-888-orphaned"}, args.Error(1)
}
```

---

## 7. Test Environment Setup

### 7.1 Integration Test Environment

#### Test Environment Setup Script

```bash
#!/bin/bash
# scripts/setup-resource-tracking-test-env.sh

# Create test environment for resource tracking tests
create_test_environment() {
    local test_root="$1"
    local test_name="$2"
    
    # Create test directory structure
    mkdir -p "$test_root/$test_name"/{repo,worktrees,config,logs}
    
    # Initialize test git repository
    cd "$test_root/$test_name/repo"
    git init
    echo "# Test Repository for Resource Tracking" > README.md
    git add README.md
    git config user.name "Test User"
    git config user.email "test@example.com"
    git commit -m "Initial commit"
    git remote add origin https://github.com/test/repo.git
    
    # Create test configuration
    cat > "$test_root/$test_name/config/config.json" << EOF
{
    "worktree_base_path": "$test_root/$test_name/worktrees",
    "work_issue_script": "./work-issue.sh",
    "command_logging": true,
    "command_log_level": "debug",
    "command_log_path": "$test_root/$test_name/logs/command.log"
}
EOF
    
    # Create test session metadata
    cat > "$test_root/$test_name/config/sessions.json" << 'EOF'
[
    {
        "issue_number": 123,
        "issue_title": "Test Issue 1",
        "branch": "issue-123-test-issue-1",
        "worktree_path": "/tmp/test-worktree-123",
        "tmux_session": "work-issue-123",
        "sandbox_name": "work-issue-123",
        "resource_status": "active",
        "current_creation_step": "completed",
        "resource_creation_log": [
            {
                "resource_type": "branch",
                "resource_id": "issue-123-test-issue-1",
                "created_at": "2025-08-01T10:00:00Z",
                "status": "created",
                "metadata": {"commit_hash": "abc123"}
            },
            {
                "resource_type": "worktree",
                "resource_id": "/tmp/test-worktree-123",
                "created_at": "2025-08-01T10:01:00Z",
                "status": "created",
                "metadata": {"path": "/tmp/test-worktree-123"}
            },
            {
                "resource_type": "tmux",
                "resource_id": "work-issue-123",
                "created_at": "2025-08-01T10:02:00Z",
                "status": "created",
                "metadata": {"session_name": "work-issue-123"}
            },
            {
                "resource_type": "sandbox",
                "resource_id": "work-issue-123",
                "created_at": "2025-08-01T10:03:00Z",
                "status": "created",
                "metadata": {"container_id": "sandbox-123"}
            }
        ]
    }
]
EOF
    
    echo "Test environment created at: $test_root/$test_name"
}

# Setup mock external tools for testing
setup_mock_tools() {
    local bin_dir="$1"
    mkdir -p "$bin_dir"
    
    # Enhanced mock git with resource tracking support
    cat > "$bin_dir/git" << 'EOF'
#!/bin/bash
case "$1" in
    "worktree")
        case "$2" in
            "add")
                echo "Preparing worktree (new branch '$4')"
                echo "HEAD is now at abc123 Initial commit"
                exit 0
                ;;
            "remove")
                echo "Removing worktree '$3'"
                exit 0
                ;;
            "list")
                echo "worktree /path/to/main"
                echo "worktree /tmp/test-worktree-123  abc123 [issue-123-test-issue-1]"
                exit 0
                ;;
            "prune")
                echo "Pruning worktree references"
                exit 0
                ;;
        esac
        ;;
    "branch")
        case "$2" in
            "-d"|"-D")
                echo "Deleted branch '$3'."
                exit 0
                ;;
            "--list")
                echo "  issue-123-test-issue-1"
                echo "  issue-456-orphaned-branch"
                echo "* main"
                exit 0
                ;;
        esac
        ;;
    "log")
        echo "abc123 Test commit"
        exit 0
        ;;
    "--version")
        echo "git version 2.40.0"
        exit 0
        ;;
esac
exit 1
EOF
    chmod +x "$bin_dir/git"
    
    # Enhanced mock tmux with session management
    cat > "$bin_dir/tmux" << 'EOF'
#!/bin/bash
case "$1" in
    "new-session")
        echo "Created session: $4"
        exit 0
        ;;
    "has-session")
        # Simulate session exists check
        case "$3" in
            "work-issue-123") exit 0 ;;
            *) exit 1 ;;
        esac
        ;;
    "kill-session")
        echo "Killed session: $3"
        exit 0
        ;;
    "list-sessions")
        echo "work-issue-123: 1 windows (created Thu Aug  1 10:02:00 2025)"
        exit 0
        ;;
    "-V")
        echo "tmux 3.3a"
        exit 0
        ;;
esac
exit 1
EOF
    chmod +x "$bin_dir/tmux"
    
    # Enhanced mock sandbox with container management
    cat > "$bin_dir/sandbox" << 'EOF'
#!/bin/bash
case "$1" in
    "list")
        echo "work-issue-123"
        echo "work-issue-orphaned"
        exit 0
        ;;
    "delete")
        echo "Deleted sandbox: $2"
        exit 0
        ;;
    "exists")
        case "$2" in
            "work-issue-123") exit 0 ;;
            *) exit 1 ;;
        esac
        ;;
    "--help")
        echo "sandbox - container management tool"
        exit 0
        ;;
esac
exit 1
EOF
    chmod +x "$bin_dir/sandbox"
    
    echo "Mock tools created in: $bin_dir"
}
```

### 7.2 Continuous Integration Setup

#### GitHub Actions Workflow for Resource Tracking Tests

```yaml
# .github/workflows/resource-tracking-tests.yml
name: Resource Tracking and Cleanup Tests

on:
  push:
    paths:
      - 'cmd/clean.go'
      - 'cmd/start.go'
      - 'cmd/stop.go'
      - 'cmd/healthcheck.go'
      - 'pkg/config/config.go'
      - 'pkg/git/manager.go'
      - '**/*_test.go'
  pull_request:
    paths:
      - 'cmd/clean.go'
      - 'cmd/start.go'
      - 'cmd/stop.go'
      - 'cmd/healthcheck.go'
      - 'pkg/config/config.go'
      - 'pkg/git/manager.go'
      - '**/*_test.go'

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24.4'

    - name: Install Dependencies
      run: |
        go mod download
        sudo apt-get update
        sudo apt-get install -y tmux git

    - name: Run Resource Tracking Unit Tests
      run: |
        go test -v ./pkg/config/... -run=".*ResourceTracking.*"
        go test -v ./pkg/git/... -run=".*BranchCleanup.*"
        go test -v ./cmd/... -run=".*ResourceTracking.*"

    - name: Run Enhanced Clean Command Tests
      run: |
        go test -v ./cmd/... -run=".*Clean.*"

    - name: Run Atomic Operations Tests
      run: |
        go test -v ./cmd/... -run=".*Atomic.*"
        go test -v ./cmd/... -run=".*Rollback.*"

    - name: Run Health Check Tests
      run: |
        go test -v ./cmd/... -run=".*HealthCheck.*"

    - name: Test Coverage
      run: |
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
        go tool cover -func=coverage.out

    - name: Upload Coverage
      uses: actions/upload-artifact@v3
      with:
        name: coverage-report
        path: coverage.html

  integration-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24.4'

    - name: Setup Test Environment
      run: |
        ./scripts/setup-resource-tracking-test-env.sh /tmp/test-env integration-test
        export PATH="/tmp/test-env/bin:$PATH"

    - name: Run Integration Tests
      env:
        INTEGRATION_TESTS: "true"
        TEST_ENV_ROOT: "/tmp/test-env"
      run: |
        go test -v -tags=integration ./integration/...

    - name: Test Resource Lifecycle
      env:
        INTEGRATION_TESTS: "true"
      run: |
        go test -v ./integration/... -run=".*ResourceLifecycle.*"

    - name: Test Cleanup Scenarios
      env:
        INTEGRATION_TESTS: "true"  
      run: |
        go test -v ./integration/... -run=".*CleanupScenarios.*"

  edge-case-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24.4'

    - name: Run Edge Case Tests
      run: |
        go test -v ./test/... -run=".*EdgeCase.*"
        go test -v ./test/... -run=".*ResourceState.*"

    - name: Run Security Tests
      run: |
        go test -v ./test/... -run=".*Security.*"
        go test -v ./test/... -run=".*Safety.*"

  performance-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24.4'

    - name: Run Performance Benchmarks
      run: |
        go test -bench=. -benchmem ./pkg/config/...
        go test -bench=. -benchmem ./cmd/...

    - name: Performance Regression Check
      run: |
        go test -bench=. -count=5 ./... > current_bench.txt
        # Compare with baseline if available
        if [ -f baseline_bench.txt ]; then
          benchcmp baseline_bench.txt current_bench.txt
        fi
```

---

## 8. Test Execution Strategy

### 8.1 Test Phases and Timeline

#### Phase 1: Foundation (Week 1-2)
- [ ] Implement enhanced configuration tests for resource tracking
- [ ] Create git manager branch cleanup unit tests
- [ ] Set up comprehensive mock framework for external dependencies
- [ ] Establish test data fixtures and scenarios

#### Phase 2: Component Testing (Week 3-4)
- [ ] Implement atomic operations and rollback tests
- [ ] Create enhanced clean command tests with multi-mode support
- [ ] Implement health check command tests
- [ ] Test resource discovery and validation functionality

#### Phase 3: Integration Testing (Week 5-6)
- [ ] End-to-end resource lifecycle tests
- [ ] Cross-command integration tests
- [ ] Failure recovery and partial success scenarios
- [ ] Performance impact validation

#### Phase 4: Edge Cases and Validation (Week 7-8)
- [ ] Complex resource state edge cases
- [ ] Security and safety validation
- [ ] Backward compatibility verification
- [ ] Final performance and regression testing

### 8.2 Success Criteria and Metrics

#### Test Quality Metrics
- [ ] **Unit Test Coverage**: Minimum 85% for modified components
- [ ] **Integration Test Coverage**: All major resource lifecycle paths covered
- [ ] **Edge Case Coverage**: All identified edge cases have corresponding tests
- [ ] **Performance Validation**: No more than 5% performance regression
- [ ] **Security Validation**: All resource operations are safe and secure

#### Functional Validation
- [ ] **Resource Tracking**: All resources are properly tracked throughout lifecycle
- [ ] **Atomic Operations**: Rollback mechanisms work correctly for all failure points
- [ ] **Enhanced Cleanup**: All cleanup modes work correctly and safely
- [ ] **Health Check**: Resource validation accurately identifies issues
- [ ] **Backward Compatibility**: All existing functionality preserved

---

## 9. Risk Mitigation and Contingency Planning

### 9.1 Identified Testing Risks

#### Major Testing Risks
1. **Complex Resource State Scenarios**: Difficulty testing all possible resource states
   - **Mitigation**: Comprehensive mock framework with state simulation
   - **Contingency**: Phased rollout with extensive manual testing

2. **Atomic Operation Testing Complexity**: Challenging to test rollback mechanisms
   - **Mitigation**: Isolated unit tests for each rollback scenario
   - **Contingency**: Additional integration testing with real external tools

3. **Performance Impact from Enhanced Tracking**: Resource tracking may impact performance
   - **Mitigation**: Extensive benchmarking and performance regression tests
   - **Contingency**: Configurable tracking levels and optimization

4. **Backward Compatibility Issues**: Enhanced functionality may break existing workflows
   - **Mitigation**: Comprehensive regression testing suite
   - **Contingency**: Feature flags and gradual migration path

### 9.2 Test Environment Reliability

#### Test Environment Challenges
- **External Tool Dependencies**: Tests depend on git, tmux, sandbox tools
  - **Solution**: Comprehensive mocking with fallback to real tools for integration tests
- **File System Operations**: Tests involve file system operations that may fail
  - **Solution**: Temporary directories and proper cleanup in test teardown
- **Concurrent Test Execution**: Resource conflicts between parallel tests
  - **Solution**: Isolated test environments and unique resource identifiers

---

## 10. Test Maintenance and Evolution

### 10.1 Test Code Quality Standards

#### Testing Best Practices
- **Clear Test Names**: Test names describe the exact scenario being tested
- **Arrange-Act-Assert Pattern**: Consistent test structure for readability
- **Single Responsibility**: Each test validates one specific behavior
- **Proper Cleanup**: Tests clean up after themselves to avoid side effects
- **Realistic Test Data**: Test data represents actual usage scenarios

#### Mock and Stub Guidelines
- **Interface-Based Mocking**: Use interfaces to enable easy mocking
- **Realistic Mock Behavior**: Mocks behave like real external dependencies
- **State Tracking**: Mocks track state changes for verification
- **Error Simulation**: Mocks can simulate various error conditions

### 10.2 Continuous Improvement Strategy

#### Test Evolution Approach
- **Regular Test Review**: Monthly review of test effectiveness and coverage
- **New Scenario Integration**: Add tests for newly discovered edge cases
- **Performance Monitoring**: Track and optimize test execution performance
- **Documentation Maintenance**: Keep test documentation current with implementation

#### Test Data Management
- **Version Control**: All test fixtures committed to repository
- **Realistic Scenarios**: Test data based on actual usage patterns
- **Privacy Protection**: No real user data in test fixtures
- **Comprehensive Coverage**: Test data covers success, failure, and edge cases

---

This comprehensive testing plan ensures robust validation of the enhanced resource tracking and cleanup functionality while maintaining high code quality, performance, and reliability standards. The plan provides thorough coverage of all implementation phases, edge cases, and integration scenarios while establishing clear success criteria and risk mitigation strategies.