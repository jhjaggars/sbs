package inputsource

import (
	"fmt"
	"strings"

	"sbs/pkg/config"
)

// InputSourceFactory creates InputSource instances based on configuration
type InputSourceFactory struct {
	supportedTypes map[string]func() InputSource
}

// NewInputSourceFactory creates a new InputSourceFactory with all supported types
func NewInputSourceFactory() *InputSourceFactory {
	return &InputSourceFactory{
		supportedTypes: map[string]func() InputSource{
			"github": func() InputSource { return NewGitHubInputSource() },
			"test":   func() InputSource { return NewTestInputSource() },
		},
	}
}

// Create creates an InputSource based on the provided configuration
func (f *InputSourceFactory) Create(cfg *config.InputSourceConfig) (InputSource, error) {
	// Handle nil config - default to GitHub
	if cfg == nil {
		return f.createGitHubSource(), nil
	}

	// Handle empty type - default to GitHub
	sourceType := strings.TrimSpace(cfg.Type)
	if sourceType == "" {
		return f.createGitHubSource(), nil
	}

	// Look up the creator function
	creator, exists := f.supportedTypes[sourceType]
	if !exists {
		supportedTypes := f.GetSupportedTypes()
		return nil, fmt.Errorf("unsupported input source type: %s (supported types: %s)",
			sourceType, strings.Join(supportedTypes, ", "))
	}

	// Create the input source
	return creator(), nil
}

// CreateFromProject creates an InputSource by loading configuration from project root
func (f *InputSourceFactory) CreateFromProject(projectRoot string) (InputSource, error) {
	// Load configuration from project
	cfg, err := config.LoadInputSourceConfig(projectRoot)
	if err != nil {
		// If we can't load config, fall back to GitHub default
		// This maintains backward compatibility for projects without input source config
		return f.createGitHubSource(), nil
	}

	// Create source from configuration
	return f.Create(cfg)
}

// GetSupportedTypes returns a list of supported input source types
func (f *InputSourceFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.supportedTypes))
	for sourceType := range f.supportedTypes {
		types = append(types, sourceType)
	}
	return types
}

// createGitHubSource is a helper to create GitHub sources consistently
func (f *InputSourceFactory) createGitHubSource() InputSource {
	creator := f.supportedTypes["github"]
	return creator()
}
