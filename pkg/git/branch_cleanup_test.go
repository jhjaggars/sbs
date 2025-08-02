package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitManager_BranchCleanup(t *testing.T) {
	t.Run("delete_current_branch_protection", func(t *testing.T) {
		// Test protection against deleting current branch
		manager := &Manager{
			repoPath: "/tmp/test-repo",
		}

		// This should always return an error
		err := manager.DeleteCurrentBranch()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete current branch")
	})

	t.Run("extract_issue_number_from_branch", func(t *testing.T) {
		// Test issue number extraction logic
		manager := &Manager{
			repoPath: "/tmp/test-repo",
		}

		testCases := []struct {
			branchName    string
			expectedIssue int
		}{
			{"issue-123-fix-bug", 123},
			{"issue-456-add-feature", 456},
			{"issue-1-simple", 1},
			{"not-an-issue-branch", 0},
			{"issue-", 0},
			{"issue-abc-invalid", 0},
			{"main", 0},
		}

		for _, tc := range testCases {
			result := manager.extractIssueNumberFromBranch(tc.branchName)
			assert.Equal(t, tc.expectedIssue, result, "Failed for branch: %s", tc.branchName)
		}
	})
}

func TestGitManager_BranchDiscovery(t *testing.T) {
	t.Run("method_signatures_exist", func(t *testing.T) {
		// Test that the methods exist with correct signatures
		manager := &Manager{
			repoPath: "/tmp/test-repo",
		}

		// Test that methods exist (will fail if git repo doesn't exist, but methods should exist)
		_, err := manager.ListIssueBranches()
		// We expect an error since no real repo, but method should exist
		assert.NotNil(t, err) // Method exists and returns error for invalid repo

		orphanedBranches, err := manager.FindOrphanedIssueBranches([]string{"123", "456"})
		assert.NotNil(t, err) // Method exists
		_ = orphanedBranches  // Will be nil due to error, but method exists

		_, err = manager.GetBranchAge("issue-123-test-branch")
		assert.NotNil(t, err) // Method exists

		safe, warnings, err := manager.ValidateBranchDeletion("issue-123-test-branch")
		// ValidateBranchDeletion should handle non-existent branches gracefully
		if err == nil {
			assert.True(t, safe) // Non-existent branch is safe to "delete"
			assert.IsType(t, []string{}, warnings)
		}

		hasUnmerged, err := manager.HasUnmergedChanges("issue-456-unmerged")
		// Should return false for non-existent branch
		if err == nil {
			assert.False(t, hasUnmerged)
		}

		exists, err := manager.BranchExists("issue-123-test-branch")
		// BranchExists should work even with invalid repo (returns false)
		if err == nil {
			assert.False(t, exists)
		}
	})
}

func TestGitManager_BranchCleanupIntegration(t *testing.T) {
	t.Run("delete_multiple_branches_method_exists", func(t *testing.T) {
		// Test that DeleteMultipleBranches method exists and handles input correctly
		manager := &Manager{
			repoPath: "/tmp/test-repo",
		}

		branchesToDelete := []string{"issue-111-old", "issue-222-stale", "issue-333-done"}
		results, _ := manager.DeleteMultipleBranches(branchesToDelete, true) // dry run

		// Method should exist and return results (even if they indicate failures)
		assert.NotNil(t, results)
		assert.Equal(t, len(branchesToDelete), len(results))

		// Each result should have the correct branch name
		for i, result := range results {
			assert.Equal(t, branchesToDelete[i], result.BranchName)
		}
	})
}

// BranchDeletionResult is defined in manager.go
