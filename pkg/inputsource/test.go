package inputsource

import (
	"fmt"
	"strings"
)

// TestInputSource provides predefined test work items for development and testing
type TestInputSource struct {
	items map[string]*WorkItem
}

// NewTestInputSource creates a new TestInputSource with predefined test items
func NewTestInputSource() *TestInputSource {
	items := map[string]*WorkItem{
		"quick": {
			Source: "test",
			ID:     "quick",
			Title:  "Quick development test",
			State:  "open",
			URL:    "",
		},
		"hooks": {
			Source: "test",
			ID:     "hooks",
			Title:  "Test Claude Code hooks",
			State:  "open",
			URL:    "",
		},
		"sandbox": {
			Source: "test",
			ID:     "sandbox",
			Title:  "Test sandbox integration",
			State:  "open",
			URL:    "",
		},
	}

	return &TestInputSource{
		items: items,
	}
}

// GetWorkItem retrieves a specific test work item by its ID
func (t *TestInputSource) GetWorkItem(id string) (*WorkItem, error) {
	// Validate input
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("work item ID cannot be empty")
	}

	// Look up the item
	item, exists := t.items[id]
	if !exists {
		return nil, fmt.Errorf("test work item not found: %s (available: quick, hooks, sandbox)", id)
	}

	// Return a copy to prevent modification of the original
	return &WorkItem{
		Source: item.Source,
		ID:     item.ID,
		Title:  item.Title,
		State:  item.State,
		URL:    item.URL,
	}, nil
}

// ListWorkItems retrieves a list of test work items, optionally filtered by search query
func (t *TestInputSource) ListWorkItems(searchQuery string, limit int) ([]*WorkItem, error) {
	var results []*WorkItem

	// If limit is 0, return empty list
	if limit <= 0 {
		return results, nil
	}

	searchQuery = strings.ToLower(strings.TrimSpace(searchQuery))

	// Iterate through all items
	for _, item := range t.items {
		// Apply search filter if provided
		if searchQuery != "" {
			titleLower := strings.ToLower(item.Title)
			idLower := strings.ToLower(item.ID)

			// Skip if neither title nor ID contains the search query
			if !strings.Contains(titleLower, searchQuery) && !strings.Contains(idLower, searchQuery) {
				continue
			}
		}

		// Return a copy to prevent modification
		itemCopy := &WorkItem{
			Source: item.Source,
			ID:     item.ID,
			Title:  item.Title,
			State:  item.State,
			URL:    item.URL,
		}
		results = append(results, itemCopy)

		// Apply limit
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// GetType returns the input source type identifier
func (t *TestInputSource) GetType() string {
	return "test"
}
