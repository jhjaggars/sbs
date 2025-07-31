package issue

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// commandExecutor interface for testing
type commandExecutor interface {
	executeCommand(name string, args ...string) ([]byte, error)
}

// realCommandExecutor implements commandExecutor using os/exec
type realCommandExecutor struct{}

func (r *realCommandExecutor) executeCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

type GitHubClient struct {
	executor commandExecutor
}

type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

type ghIssueJSON struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		executor: &realCommandExecutor{},
	}
}

func (g *GitHubClient) GetIssue(issueNumber int) (*Issue, error) {
	// Use gh command to fetch issue data
	output, err := g.executor.executeCommand("gh", "issue", "view", strconv.Itoa(issueNumber), "--json", "number,title,state,url")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "Could not resolve to an Issue") {
				return nil, fmt.Errorf("issue #%d not found", issueNumber)
			}
		}
		return nil, fmt.Errorf("failed to fetch issue #%d with gh command: %w", issueNumber, err)
	}
	
	var ghIssue ghIssueJSON
	if err := json.Unmarshal(output, &ghIssue); err != nil {
		return nil, fmt.Errorf("failed to parse gh command output: %w", err)
	}
	
	return &Issue{
		Number: ghIssue.Number,
		Title:  ghIssue.Title,
		State:  ghIssue.State,
		URL:    ghIssue.URL,
	}, nil
}

// ListIssues fetches a list of open issues from the current repository
func (g *GitHubClient) ListIssues(searchQuery string, limit int) ([]Issue, error) {
	// Build gh command arguments
	args := []string{"issue", "list", "--json", "number,title,state,url", "--state", "open", "--limit", strconv.Itoa(limit)}
	
	// Add search query if provided
	if searchQuery != "" {
		args = append(args, "--search", searchQuery)
	}
	
	// Execute gh command
	output, err := g.executor.executeCommand("gh", args...)
	if err != nil {
		// Handle specific error cases
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			return nil, fmt.Errorf("gh command not found. Please install GitHub CLI: https://cli.github.com/")
		}
		
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "gh auth login") {
				return nil, fmt.Errorf("GitHub CLI authentication required. Please run: gh auth login")
			}
		}
		
		return nil, fmt.Errorf("failed to list issues with gh command: %w", err)
	}
	
	// Parse JSON response
	var ghIssues []ghIssueJSON
	if err := json.Unmarshal(output, &ghIssues); err != nil {
		return nil, fmt.Errorf("failed to parse gh command output: %w", err)
	}
	
	// Convert to Issue structs
	issues := make([]Issue, len(ghIssues))
	for i, ghIssue := range ghIssues {
		issues[i] = Issue{
			Number: ghIssue.Number,
			Title:  ghIssue.Title,
			State:  ghIssue.State,
			URL:    ghIssue.URL,
		}
	}
	
	return issues, nil
}

// CheckGHInstalled verifies that the gh command is available
func CheckGHInstalled() error {
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh command not found. Please install GitHub CLI: https://cli.github.com/")
	}
	return nil
}

// ParseIssueNumber extracts issue number from various formats
func ParseIssueNumber(input string) (int, error) {
	// Remove common prefixes
	if input[0] == '#' {
		input = input[1:]
	}
	
	return strconv.Atoi(input)
}