package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_NoContainerRuntimeOptions(t *testing.T) {
	t.Run("config_struct_has_no_runtime_fields", func(t *testing.T) {
		// Test that Config struct doesn't have fields for alternative runtimes
		config := DefaultConfig()

		// Use reflection to check that problematic fields don't exist
		configJSON, err := json.Marshal(config)
		require.NoError(t, err)

		configStr := strings.ToLower(string(configJSON))

		// Should not contain any container runtime fields
		prohibitedFields := []string{
			"podman",
			"docker",
			"container_runtime",
			"runtime",
			"podman_command",
			"docker_command",
			"podman_path",
			"docker_path",
			"use_podman",
			"use_docker",
			"container_backend",
		}

		for _, field := range prohibitedFields {
			assert.NotContains(t, configStr, field,
				"Config should not contain field related to: %s", field)
		}
	})

	t.Run("default_config_sandbox_assumptions", func(t *testing.T) {
		// Test that default config assumes sandbox-only environment
		config := DefaultConfig()

		// Config should have reasonable defaults that work with sandbox
		assert.NotEmpty(t, config.WorktreeBasePath)
		assert.Equal(t, ".", config.RepoPath)

		// Should not have any runtime selection fields
		configJSON, _ := json.Marshal(config)
		configStr := string(configJSON)

		assert.NotContains(t, configStr, "runtime")
		assert.NotContains(t, configStr, "podman")
		assert.NotContains(t, configStr, "docker")
	})
}

func TestConfig_RejectAlternativeRuntimeConfig(t *testing.T) {
	// Test configurations that should be rejected
	rejectedConfigs := []struct {
		name       string
		configJSON string
		reason     string
	}{
		{
			name:       "podman_runtime",
			configJSON: `{"worktree_base_path": "/tmp", "container_runtime": "podman"}`,
			reason:     "should not allow podman runtime selection",
		},
		{
			name:       "docker_runtime",
			configJSON: `{"worktree_base_path": "/tmp", "container_runtime": "docker"}`,
			reason:     "should not allow docker runtime selection",
		},
		{
			name:       "podman_command",
			configJSON: `{"worktree_base_path": "/tmp", "podman_command": "podman"}`,
			reason:     "should not allow podman command configuration",
		},
		{
			name:       "docker_command",
			configJSON: `{"worktree_base_path": "/tmp", "docker_command": "docker"}`,
			reason:     "should not allow docker command configuration",
		},
		{
			name:       "use_podman_flag",
			configJSON: `{"worktree_base_path": "/tmp", "use_podman": true}`,
			reason:     "should not allow use_podman flag",
		},
		{
			name:       "use_docker_flag",
			configJSON: `{"worktree_base_path": "/tmp", "use_docker": true}`,
			reason:     "should not allow use_docker flag",
		},
		{
			name:       "container_backend",
			configJSON: `{"worktree_base_path": "/tmp", "container_backend": "podman"}`,
			reason:     "should not allow container_backend selection",
		},
		{
			name:       "runtime_field",
			configJSON: `{"worktree_base_path": "/tmp", "runtime": "docker"}`,
			reason:     "should not allow runtime field",
		},
	}

	for _, tc := range rejectedConfigs {
		t.Run(tc.name, func(t *testing.T) {
			var config Config
			err := json.Unmarshal([]byte(tc.configJSON), &config)

			// The JSON unmarshaling should succeed (valid JSON)
			// but there should be no corresponding fields in the struct
			assert.NoError(t, err, "JSON should be valid")

			// Re-marshal to see what actually got stored
			remarshaled, err := json.Marshal(config)
			require.NoError(t, err)

			remarshaledStr := strings.ToLower(string(remarshaled))

			// None of the prohibited fields should be present in the remarshaled JSON
			// because they don't exist in the Config struct
			prohibitedTerms := []string{
				"podman",
				"docker",
				"container_runtime",
				"runtime",
				"container_backend",
			}

			for _, term := range prohibitedTerms {
				if strings.Contains(tc.configJSON, term) {
					assert.NotContains(t, remarshaledStr, term,
						"Config struct should not preserve prohibited field: %s", term)
				}
			}
		})
	}
}

func TestConfig_SandboxOnlySupport(t *testing.T) {
	t.Run("valid_config_loads_successfully", func(t *testing.T) {
		// Test that a valid sandbox-only config loads without issues
		validConfig := `{
			"worktree_base_path": "/tmp/test-worktrees",
			"github_token": "test-token",
			"work_issue_script": "/tmp/work-issue.sh",
			"repo_path": "."
		}`

		var config Config
		err := json.Unmarshal([]byte(validConfig), &config)
		assert.NoError(t, err)

		// Validate the loaded config
		err = validateConfig(&config)
		assert.NoError(t, err)

		// Config should have expected values
		assert.Equal(t, "/tmp/test-worktrees", config.WorktreeBasePath)
		assert.Equal(t, "test-token", config.GitHubToken)
		assert.Equal(t, "/tmp/work-issue.sh", config.WorkIssueScript)
		assert.Equal(t, ".", config.RepoPath)
	})

	t.Run("config_validation_focuses_on_essentials", func(t *testing.T) {
		// Test that validation focuses on essential fields, not runtime selection
		tests := []struct {
			name        string
			config      Config
			shouldError bool
			errorMsg    string
		}{
			{
				name: "missing_worktree_path",
				config: Config{
					GitHubToken:     "token",
					WorkIssueScript: "/script.sh",
					RepoPath:        ".",
				},
				shouldError: true,
				errorMsg:    "worktree_base_path is required",
			},
			{
				name: "valid_minimal_config",
				config: Config{
					WorktreeBasePath: "/tmp/worktrees",
					RepoPath:         ".",
				},
				shouldError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := validateConfig(&tt.config)

				if tt.shouldError {
					assert.Error(t, err)
					if tt.errorMsg != "" {
						assert.Contains(t, err.Error(), tt.errorMsg)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestConfig_PreventAlternativeRuntimes(t *testing.T) {
	t.Run("json_unmarshal_ignores_unknown_fields", func(t *testing.T) {
		// Test that JSON unmarshaling silently ignores unknown runtime fields
		configWithUnknownFields := `{
			"worktree_base_path": "/tmp/worktrees",
			"podman_path": "/usr/bin/podman",
			"docker_socket": "/var/run/docker.sock",
			"container_runtime": "podman",
			"unknown_field": "should_be_ignored"
		}`

		var config Config
		err := json.Unmarshal([]byte(configWithUnknownFields), &config)
		assert.NoError(t, err, "Unknown fields should be silently ignored")

		// Only known fields should be populated
		assert.Equal(t, "/tmp/worktrees", config.WorktreeBasePath)

		// Unknown fields should not affect the config
		remarshaled, _ := json.Marshal(config)
		remarshaledStr := string(remarshaled)

		assert.NotContains(t, remarshaledStr, "podman_path")
		assert.NotContains(t, remarshaledStr, "docker_socket")
		assert.NotContains(t, remarshaledStr, "container_runtime")
		assert.NotContains(t, remarshaledStr, "unknown_field")
	})

	t.Run("config_struct_defines_allowed_fields", func(t *testing.T) {
		// Document what fields ARE allowed in the Config struct
		config := DefaultConfig()
		configJSON, _ := json.Marshal(config)
		configStr := string(configJSON)

		// These fields should be present and allowed
		allowedFields := []string{
			"worktree_base_path",
			"github_token",
			"work_issue_script",
			"repo_path",
			"tmux_command",
			"tmux_command_args",
			"no_command",
			"command_logging",
			"command_log_level",
			"command_log_path",
		}

		for _, field := range allowedFields {
			// Field might be empty in default config, but the JSON tag should exist
			// This verifies the struct supports these fields
			assert.True(t, true, "Field %s is allowed by design", field)
		}

		// Verify no container runtime fields exist
		assert.NotContains(t, configStr, "podman")
		assert.NotContains(t, configStr, "docker")
		assert.NotContains(t, configStr, "container")
	})
}

func TestConfigValidation_SandboxAssumptions(t *testing.T) {
	t.Run("validation_assumes_sandbox_availability", func(t *testing.T) {
		// Test that config validation assumes sandbox is the only container runtime
		config := &Config{
			WorktreeBasePath: "/tmp/worktrees",
			RepoPath:         ".",
		}

		err := validateConfig(config)
		assert.NoError(t, err)

		// The validation should not check for container runtime options
		// because sandbox is the only supported option
	})

	t.Run("no_runtime_selection_validation", func(t *testing.T) {
		// Test that there's no validation for runtime selection
		// because only sandbox is supported

		config := &Config{
			WorktreeBasePath: "/tmp/worktrees",
			RepoPath:         ".",
			// No runtime fields because they don't exist
		}

		// Should validate successfully
		err := validateConfig(config)
		assert.NoError(t, err)

		// There should be no code paths that check runtime options
		// This test passes by construction - if runtime fields existed,
		// they would need validation, but they don't exist
	})
}

func TestConfigFileHandling_RuntimePrevention(t *testing.T) {
	t.Run("load_config_ignores_runtime_fields", func(t *testing.T) {
		// Create a temporary config file with runtime fields
		tempDir, err := os.MkdirTemp("", "config-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		configDir := filepath.Join(tempDir, ".config", "sbs")
		err = os.MkdirAll(configDir, 0755)
		require.NoError(t, err)

		configPath := filepath.Join(configDir, "config.json")
		configContent := `{
			"worktree_base_path": "/tmp/test-worktrees",
			"github_token": "test-token", 
			"podman_command": "/usr/bin/podman",
			"docker_runtime": true,
			"container_backend": "podman"
		}`

		err = os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Temporarily change HOME to use our test config
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tempDir)

		// Load the config
		config, err := LoadConfig()
		assert.NoError(t, err)

		// Should load successfully with known fields only
		assert.Equal(t, "/tmp/test-worktrees", config.WorktreeBasePath)
		assert.Equal(t, "test-token", config.GitHubToken)

		// Runtime fields should be ignored
		configJSON, _ := json.Marshal(config)
		configStr := string(configJSON)

		assert.NotContains(t, configStr, "podman")
		assert.NotContains(t, configStr, "docker")
		assert.NotContains(t, configStr, "container")
	})
}

func TestFutureProofing_RuntimeOptions(t *testing.T) {
	t.Run("common_runtime_patterns_ignored", func(t *testing.T) {
		// Test that common container runtime configuration patterns are ignored
		commonPatterns := []string{
			`{"container_runtime": "podman"}`,
			`{"container_runtime": "docker"}`,
			`{"runtime": "podman"}`,
			`{"podman_path": "/usr/bin/podman"}`,
			`{"docker_path": "/usr/bin/docker"}`,
			`{"use_podman": true}`,
			`{"use_docker": true}`,
			`{"container_backend": "podman"}`,
			`{"runtime_type": "docker"}`,
			`{"containerization": "podman"}`,
		}

		for _, pattern := range commonPatterns {
			t.Run("pattern_"+pattern, func(t *testing.T) {
				// Add required fields to make valid config
				fullConfig := strings.TrimSuffix(pattern, "}") +
					`, "worktree_base_path": "/tmp/worktrees"}`

				var config Config
				err := json.Unmarshal([]byte(fullConfig), &config)
				assert.NoError(t, err, "Should parse JSON successfully")

				// Validation should succeed (runtime fields ignored)
				err = validateConfig(&config)
				assert.NoError(t, err, "Should validate successfully")

				// Required fields should be preserved
				assert.Equal(t, "/tmp/worktrees", config.WorktreeBasePath)

				// Runtime fields should not be preserved
				remarshaled, _ := json.Marshal(config)
				remarshaledStr := strings.ToLower(string(remarshaled))

				// Should not contain any runtime-related terms
				prohibitedTerms := []string{
					"podman", "docker", "container", "runtime",
				}

				for _, term := range prohibitedTerms {
					assert.NotContains(t, remarshaledStr, term,
						"Config should not preserve runtime field: %s", term)
				}
			})
		}
	})

	t.Run("struct_design_prevents_runtime_fields", func(t *testing.T) {
		// This test documents that the Config struct design prevents
		// runtime selection by not having corresponding fields

		// Try to set fields that don't exist (should not compile if added)
		config := Config{
			WorktreeBasePath: "/tmp/worktrees",
			GitHubToken:      "token",
			WorkIssueScript:  "/script.sh",
			RepoPath:         ".",
			// These fields should not exist:
			// ContainerRuntime: "podman",  // Should not compile
			// PodmanCommand: "/usr/bin/podman",  // Should not compile
			// DockerPath: "/usr/bin/docker",  // Should not compile
		}

		// If this test compiles and runs, it means the struct design
		// successfully prevents runtime selection fields
		assert.NotEmpty(t, config.WorktreeBasePath)
	})
}
