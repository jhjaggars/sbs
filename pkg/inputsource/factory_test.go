package inputsource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sbs/pkg/config"
)

func TestInputSourceFactory_CreateFromConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.InputSourceConfig
		expectedType  string
		expectedError bool
	}{
		{
			name:          "github_source",
			config:        &config.InputSourceConfig{Type: "github"},
			expectedType:  "github",
			expectedError: false,
		},
		{
			name:          "test_source",
			config:        &config.InputSourceConfig{Type: "test"},
			expectedType:  "test",
			expectedError: false,
		},
		{
			name:          "unknown_source",
			config:        &config.InputSourceConfig{Type: "unknown"},
			expectedType:  "",
			expectedError: true,
		},
		{
			name:          "nil_config_defaults_to_github",
			config:        nil,
			expectedType:  "github",
			expectedError: false,
		},
		{
			name:          "empty_type_defaults_to_github",
			config:        &config.InputSourceConfig{Type: ""},
			expectedType:  "github",
			expectedError: false,
		},
	}

	factory := NewInputSourceFactory()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := factory.Create(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, source)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, source)
				assert.Equal(t, tt.expectedType, source.GetType())
			}
		})
	}
}

func TestInputSourceFactory_CreateFromProject(t *testing.T) {
	// Test factory creating input sources from project configuration
	factory := NewInputSourceFactory()

	t.Run("project_with_test_config", func(t *testing.T) {
		// Create temporary project directory with test config
		tempDir := t.TempDir()

		testConfig := &config.InputSourceConfig{
			Type: "test",
			Settings: map[string]interface{}{
				"description": "Test project",
			},
		}

		err := config.SaveInputSourceConfig(tempDir, testConfig)
		require.NoError(t, err)

		// Create source from project
		source, err := factory.CreateFromProject(tempDir)
		require.NoError(t, err)
		assert.Equal(t, "test", source.GetType())
	})

	t.Run("project_with_github_config", func(t *testing.T) {
		// Create temporary project directory with GitHub config
		tempDir := t.TempDir()

		githubConfig := &config.InputSourceConfig{
			Type: "github",
			Settings: map[string]interface{}{
				"repository": "owner/repo",
			},
		}

		err := config.SaveInputSourceConfig(tempDir, githubConfig)
		require.NoError(t, err)

		// Create source from project
		source, err := factory.CreateFromProject(tempDir)
		require.NoError(t, err)
		assert.Equal(t, "github", source.GetType())
	})

	t.Run("project_without_config", func(t *testing.T) {
		// Create temporary project directory without config
		tempDir := t.TempDir()

		// Should default to GitHub
		source, err := factory.CreateFromProject(tempDir)
		require.NoError(t, err)
		assert.Equal(t, "github", source.GetType())
	})
}

func TestInputSourceFactory_SupportedTypes(t *testing.T) {
	factory := NewInputSourceFactory()
	supportedTypes := factory.GetSupportedTypes()

	// Should include at least GitHub and test sources
	assert.Contains(t, supportedTypes, "github")
	assert.Contains(t, supportedTypes, "test")
	assert.GreaterOrEqual(t, len(supportedTypes), 2)
}

func TestInputSourceFactory_ErrorHandling(t *testing.T) {
	factory := NewInputSourceFactory()

	t.Run("unsupported_type", func(t *testing.T) {
		config := &config.InputSourceConfig{Type: "unsupported-source"}
		source, err := factory.Create(config)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported input source type")
		assert.Contains(t, err.Error(), "unsupported-source")
		assert.Nil(t, source)
	})

	t.Run("invalid_project_path", func(t *testing.T) {
		source, err := factory.CreateFromProject("/nonexistent/path")

		// Should not error, should fall back to default (GitHub)
		require.NoError(t, err)
		assert.Equal(t, "github", source.GetType())
	})
}

func TestInputSourceFactory_ConfigurationParsing(t *testing.T) {
	factory := NewInputSourceFactory()

	t.Run("config_with_settings", func(t *testing.T) {
		config := &config.InputSourceConfig{
			Type: "test",
			Settings: map[string]interface{}{
				"description": "Factory test",
			},
		}

		source, err := factory.Create(config)
		require.NoError(t, err)
		assert.Equal(t, "test", source.GetType())

		// Test source doesn't use settings, but factory should handle them
		assert.NotNil(t, source)
	})

	t.Run("config_without_settings", func(t *testing.T) {
		config := &config.InputSourceConfig{
			Type:     "test",
			Settings: nil,
		}

		source, err := factory.Create(config)
		require.NoError(t, err)
		assert.Equal(t, "test", source.GetType())
	})
}

func TestInputSourceFactory_Consistency(t *testing.T) {
	// Test that factory creates consistent instances
	factory := NewInputSourceFactory()

	t.Run("multiple_github_instances", func(t *testing.T) {
		config := &config.InputSourceConfig{Type: "github"}

		source1, err1 := factory.Create(config)
		source2, err2 := factory.Create(config)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, source1.GetType(), source2.GetType())
	})

	t.Run("multiple_test_instances", func(t *testing.T) {
		config := &config.InputSourceConfig{Type: "test"}

		source1, err1 := factory.Create(config)
		source2, err2 := factory.Create(config)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, source1.GetType(), source2.GetType())

		// Test sources should return the same predefined items
		items1, _ := source1.ListWorkItems("", 10)
		items2, _ := source2.ListWorkItems("", 10)
		assert.Equal(t, len(items1), len(items2))
	})
}
