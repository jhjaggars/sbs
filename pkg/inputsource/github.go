package inputsource

import (
	"fmt"
	"strconv"

	"sbs/pkg/issue"
)

// GitHubClientInterface defines the interface for GitHub client operations
type GitHubClientInterface interface {
	GetIssue(issueNumber int) (*issue.Issue, error)
	ListIssues(searchQuery string, limit int) ([]issue.Issue, error)
}

// GitHubInputSource wraps the existing GitHub issue functionality
type GitHubInputSource struct {
	client GitHubClientInterface
}

// NewGitHubInputSource creates a new GitHubInputSource
func NewGitHubInputSource() *GitHubInputSource {
	return &GitHubInputSource{
		client: issue.NewGitHubClient(),
	}
}

// GetWorkItem retrieves a GitHub issue by its number
func (g *GitHubInputSource) GetWorkItem(id string) (*WorkItem, error) {
	// Parse the ID as an issue number
	issueNumber, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub issue number: %s", id)
	}

	// Get the issue from GitHub
	githubIssue, err := g.client.GetIssue(issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub issue #%d: %w", issueNumber, err)
	}

	// Convert to WorkItem
	return &WorkItem{
		Source: "github",
		ID:     id,
		Title:  githubIssue.Title,
		State:  githubIssue.State,
		URL:    githubIssue.URL,
	}, nil
}

// ListWorkItems retrieves a list of GitHub issues
func (g *GitHubInputSource) ListWorkItems(searchQuery string, limit int) ([]*WorkItem, error) {
	// Get issues from GitHub
	githubIssues, err := g.client.ListIssues(searchQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub issues: %w", err)
	}

	// Convert to WorkItems
	workItems := make([]*WorkItem, len(githubIssues))
	for i, githubIssue := range githubIssues {
		workItems[i] = &WorkItem{
			Source: "github",
			ID:     strconv.Itoa(githubIssue.Number),
			Title:  githubIssue.Title,
			State:  githubIssue.State,
			URL:    githubIssue.URL,
		}
	}

	return workItems, nil
}

// GetType returns the input source type identifier
func (g *GitHubInputSource) GetType() string {
	return "github"
}
