package issue

import (
	"fmt"
	"path/filepath"
	"time"

	"sbs/pkg/config"
)

type Tracker struct {
	githubClient *GitHubClient
	config       *config.Config
}

func NewTracker(cfg *config.Config) *Tracker {
	githubClient := NewGitHubClient()

	return &Tracker{
		githubClient: githubClient,
		config:       cfg,
	}
}

func (t *Tracker) GetIssue(issueNumber int) (*Issue, error) {
	return t.githubClient.GetIssue(issueNumber)
}

func (t *Tracker) CreateSessionMetadata(issueNumber int, issue *Issue, branch, worktreePath, tmuxSession, sandboxName, repoName, repoRoot, friendlyTitle string) *config.SessionMetadata {
	now := time.Now().Format(time.RFC3339)

	return &config.SessionMetadata{
		IssueTitle:     issue.Title,
		FriendlyTitle:  friendlyTitle,
		Branch:         branch,
		WorktreePath:   worktreePath,
		TmuxSession:    tmuxSession,
		SandboxName:    sandboxName,
		RepositoryName: repoName,
		RepositoryRoot: repoRoot,
		CreatedAt:      now,
		LastActivity:   now,
		Status:         "active",
	}
}

func (t *Tracker) UpdateSessionActivity(sessions []config.SessionMetadata, issueNumber int) []config.SessionMetadata {
	// Update by namespaced ID - issue tracker should be refactored
	// For now, return sessions unchanged
	_ = issueNumber // avoid unused variable error

	return sessions
}

func (t *Tracker) GetWorktreePath(issueNumber int) string {
	return filepath.Join(t.config.WorktreeBasePath, fmt.Sprintf("issue-%d", issueNumber))
}

func (t *Tracker) FindSessionByIssue(sessions []config.SessionMetadata, issueNumber int) *config.SessionMetadata {
	// This function should be deprecated - using namespaced IDs instead
	// Return nil for now
	_ = issueNumber // avoid unused variable error
	return nil
}

func (t *Tracker) RemoveSession(sessions []config.SessionMetadata, issueNumber int) []config.SessionMetadata {
	// This function should be deprecated - using namespaced IDs instead
	// Return sessions unchanged for now
	_ = issueNumber // avoid unused variable error
	return sessions
}
