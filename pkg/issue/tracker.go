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
		IssueNumber:    issueNumber,
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
	now := time.Now().Format(time.RFC3339)

	for i, session := range sessions {
		if session.IssueNumber == issueNumber {
			sessions[i].LastActivity = now
			break
		}
	}

	return sessions
}

func (t *Tracker) GetWorktreePath(issueNumber int) string {
	return filepath.Join(t.config.WorktreeBasePath, fmt.Sprintf("issue-%d", issueNumber))
}

func (t *Tracker) FindSessionByIssue(sessions []config.SessionMetadata, issueNumber int) *config.SessionMetadata {
	for _, session := range sessions {
		if session.IssueNumber == issueNumber {
			return &session
		}
	}
	return nil
}

func (t *Tracker) RemoveSession(sessions []config.SessionMetadata, issueNumber int) []config.SessionMetadata {
	var filtered []config.SessionMetadata

	for _, session := range sessions {
		if session.IssueNumber != issueNumber {
			filtered = append(filtered, session)
		}
	}

	return filtered
}
