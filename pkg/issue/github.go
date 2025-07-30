package issue

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type GitHubClient struct{}

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
	return &GitHubClient{}
}

func (g *GitHubClient) GetIssue(issueNumber int) (*Issue, error) {
	// Use gh command to fetch issue data
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber), "--json", "number,title,state,url")
	output, err := cmd.Output()
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