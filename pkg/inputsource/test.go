package inputsource

import (
	"fmt"
	"strings"
)

// TestInputSource provides test work items for development and testing.
// It accepts any arbitrary ID and creates test work items dynamically.
type TestInputSource struct{}

// NewTestInputSource creates a new TestInputSource
func NewTestInputSource() *TestInputSource {
	return &TestInputSource{}
}

// GetWorkItem retrieves a specific test work item by its ID.
// Accepts any arbitrary ID and creates a test work item dynamically.
func (t *TestInputSource) GetWorkItem(id string) (*WorkItem, error) {
	// Validate input
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("work item ID cannot be empty")
	}

	// Validate ID format - should only contain alphanumeric characters, hyphens, and underscores
	if !isValidTestID(id) {
		return nil, fmt.Errorf("invalid test work item ID: %s (must contain only alphanumeric characters, hyphens, and underscores)", id)
	}

	// Create a dynamic test work item
	return &WorkItem{
		Source: "test",
		ID:     id,
		Title:  fmt.Sprintf("Test work item: %s", id),
		State:  "open",
		URL:    "",
	}, nil
}

// isValidTestID validates that a test ID contains only allowed characters
func isValidTestID(id string) bool {
	if len(id) == 0 {
		return false
	}

	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// ListWorkItems retrieves a list of test work items, optionally filtered by search query.
// Since test work items are created dynamically, this returns an empty list.
// Users should specify the exact test ID they want to use.
func (t *TestInputSource) ListWorkItems(searchQuery string, limit int) ([]*WorkItem, error) {
	// Test work items are created dynamically, so there's no predefined list to return.
	// Return empty list to indicate users should specify exact test IDs.
	return []*WorkItem{}, nil
}

// GetType returns the input source type identifier
func (t *TestInputSource) GetType() string {
	return "test"
}
