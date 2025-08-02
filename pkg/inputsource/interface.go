package inputsource

// InputSource defines the interface for different work item backends
type InputSource interface {
	// GetWorkItem retrieves a specific work item by its ID
	// The ID should be the source-specific identifier (e.g., "123" for GitHub, "quick" for test)
	GetWorkItem(id string) (*WorkItem, error)

	// ListWorkItems retrieves a list of work items, optionally filtered by search query
	// searchQuery can be empty to list all items
	// limit specifies the maximum number of items to return
	ListWorkItems(searchQuery string, limit int) ([]*WorkItem, error)

	// GetType returns the source type identifier (e.g., "github", "test", "jira")
	GetType() string
}
