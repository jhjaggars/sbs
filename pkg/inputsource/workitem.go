package inputsource

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// MaxTitleSlugLength defines the maximum length for title slugs in branch names
	MaxTitleSlugLength = 100
	// MaxFriendlyTitleLength defines the maximum length for friendly titles
	MaxFriendlyTitleLength = 50
	// DefaultSearchLimit defines the default limit for search operations in tests
	DefaultSearchLimit = 10
)

var (
	// titleSlugRegex is a compiled regex for creating title slugs
	// This is cached to avoid recompiling on each call
	titleSlugRegex = regexp.MustCompile(`[^a-z0-9]+`)
)

// WorkItem represents a work item from any input source with namespaced ID
type WorkItem struct {
	Source string `json:"source"` // github, test, jira, etc.
	ID     string `json:"id"`     // The source-specific identifier
	Title  string `json:"title"`
	State  string `json:"state"` // open, closed, etc.
	URL    string `json:"url"`   // Optional URL to the work item
}

// FullID returns the full namespaced ID in the format "source:id"
func (w *WorkItem) FullID() string {
	return fmt.Sprintf("%s:%s", w.Source, w.ID)
}

// GetBranchName returns the git branch name using namespaced format
// Format: issue-{source}-{id}-{title-slug}
func (w *WorkItem) GetBranchName() string {
	titleSlug := createTitleSlug(w.Title)
	if titleSlug == "" {
		return fmt.Sprintf("issue-%s-%s", w.Source, w.ID)
	}
	return fmt.Sprintf("issue-%s-%s-%s", w.Source, w.ID, titleSlug)
}

// GetLegacyBranchName returns the branch name for backward compatibility
// For GitHub sources from legacy input, omits the source prefix
// For non-GitHub sources, same as GetBranchName()
func (w *WorkItem) GetLegacyBranchName() string {
	if w.Source == "github" {
		titleSlug := createTitleSlug(w.Title)
		if titleSlug == "" {
			return fmt.Sprintf("issue-%s", w.ID)
		}
		return fmt.Sprintf("issue-%s-%s", w.ID, titleSlug)
	}
	return w.GetBranchName()
}

// ParseWorkItemID parses a work item ID and returns a WorkItem
// Supports both namespaced format "source:id" and legacy numeric format
func ParseWorkItemID(input string) (*WorkItem, error) {
	if input == "" {
		return nil, fmt.Errorf("work item ID cannot be empty")
	}

	// Handle legacy numeric format (including #123 format)
	if IsLegacyFormat(input) {
		// Remove # prefix if present
		cleanInput := strings.TrimPrefix(input, "#")

		// Validate it's a number
		if _, err := strconv.Atoi(cleanInput); err != nil {
			return nil, fmt.Errorf("invalid legacy issue number: %s", input)
		}

		return &WorkItem{
			Source: "github", // Legacy format defaults to GitHub
			ID:     cleanInput,
		}, nil
	}

	// Parse namespaced format "source:id"
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid work item ID format: %s (expected 'source:id' or numeric)", input)
	}

	source := strings.TrimSpace(parts[0])
	id := strings.TrimSpace(parts[1])

	if source == "" {
		return nil, fmt.Errorf("source cannot be empty in work item ID: %s", input)
	}
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty in work item ID: %s", input)
	}

	// Validate ID doesn't contain spaces or special characters that would break parsing
	if strings.Contains(id, " ") {
		return nil, fmt.Errorf("work item ID cannot contain spaces: %s", input)
	}

	return &WorkItem{
		Source: source,
		ID:     id,
	}, nil
}

// IsLegacyFormat checks if the input is in legacy numeric format
func IsLegacyFormat(input string) bool {
	// Remove # prefix if present
	cleanInput := strings.TrimPrefix(input, "#")

	// Check if it's a pure number
	_, err := strconv.Atoi(cleanInput)
	return err == nil
}

// createTitleSlug creates a URL-safe slug from a title
func createTitleSlug(title string) string {
	// Trim whitespace
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}

	// Convert to lowercase
	title = strings.ToLower(title)

	// Replace non-alphanumeric characters with hyphens using cached regex
	title = titleSlugRegex.ReplaceAllString(title, "-")

	// Remove leading/trailing hyphens
	title = strings.Trim(title, "-")

	// Limit length for practical git branch naming
	if len(title) > MaxTitleSlugLength {
		title = title[:MaxTitleSlugLength]
		// Remove trailing hyphen if we cut in the middle of a word
		title = strings.TrimSuffix(title, "-")
	}

	return title
}
