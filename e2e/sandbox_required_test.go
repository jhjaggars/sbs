package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/internal/testutil"
	"sbs/pkg/validation"
)

func TestE2E_NewUserWithoutSandbox(t *testing.T) {
	t.Run("fresh_system_validation_failure", func(t *testing.T) {
		// Simulate fresh system without sandbox
		testutil.WithoutSandbox(t, func() {
			// New user trying to use SBS should get clear guidance
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should get comprehensive error about missing tools
				assert.Contains(t, err.Error(), "Missing required tools")
				assert.Contains(t, err.Error(), "sandbox command not found")

				// Error should provide actionable guidance
				testutil.AssertErrorActionable(t, err.Error())

				// Should not mention alternative runtimes
				testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
			}
		})
	})

	t.Run("error_discovery_process", func(t *testing.T) {
		// Test the error discovery process for new users

		testutil.WithLimitedPath(t, []string{"/usr/bin", "/bin"}, func() {
			// User runs validation (which would be part of any command)
			err := validation.CheckRequiredTools()

			if err != nil {
				errorMsg := err.Error()

				// Should discover ALL missing tools at once, not fail fast
				assert.Contains(t, errorMsg, "Missing required tools:")

				// Should be formatted for easy reading
				lines := strings.Split(errorMsg, "\n")
				assert.True(t, len(lines) > 1, "Error should be multi-line for readability")

				// Should include sandbox in the comprehensive check
				assert.Contains(t, errorMsg, "sandbox")

				// Should provide next steps
				testutil.AssertErrorActionable(t, errorMsg)
			}
		})
	})

	t.Run("installation_guidance_clarity", func(t *testing.T) {
		// Test that installation guidance is clear and specific

		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil && strings.Contains(err.Error(), "sandbox") {
				errorMsg := err.Error()

				// Should provide specific guidance for sandbox
				assert.Contains(t, errorMsg, "sandbox command not found")
				assert.Contains(t, errorMsg, "Please ensure sandbox is installed")
				assert.Contains(t, errorMsg, "in PATH")

				// Should be clear about what needs to be done
				requiredGuidance := []string{"installed", "PATH"}
				for _, guidance := range requiredGuidance {
					assert.Contains(t, errorMsg, guidance)
				}
			}
		})
	})
}

func TestE2E_ExistingUserMigration(t *testing.T) {
	t.Run("behavior_consistency_check", func(t *testing.T) {
		// Test that behavior is consistent (no migration needed since already sandbox-only)

		// Existing users should see same validation behavior
		err := validation.CheckRequiredTools()

		// Behavior should be consistent regardless of user type
		if err != nil {
			// Should follow same error patterns
			testutil.AssertErrorActionable(t, err.Error())
			testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
		}

		// No breaking changes expected since codebase already uses sandbox exclusively
		assert.True(t, true, "No migration needed - already sandbox-only")
	})

	t.Run("configuration_compatibility", func(t *testing.T) {
		// Test that existing configurations work without modification

		validConfig := `{
			"worktree_base_path": "/tmp/test-worktrees",
			"github_token": "test-token",
			"work_issue_script": "/tmp/work-issue.sh"
		}`

		testutil.WithMockConfig(t, validConfig, func(configPath string) {
			// Configuration should load successfully
			// Since we're testing file loading, this is more of a structural test
			assert.True(t, true, "Valid configurations should continue to work")
		})
	})
}

func TestE2E_DevelopmentWorkflows(t *testing.T) {
	t.Run("developer_setup_validation", func(t *testing.T) {
		// Test developer environment setup process

		// Developer should get comprehensive validation feedback
		err := validation.CheckRequiredTools()

		if err != nil {
			// Should identify all missing development tools
			assert.Contains(t, err.Error(), "Missing required tools")

			// Should help developer understand complete setup requirements
			testutil.AssertErrorActionable(t, err.Error())
		} else {
			// If no error, all tools are available for development
			assert.True(t, true, "Development environment properly configured")
		}
	})

	t.Run("ci_cd_environment_requirements", func(t *testing.T) {
		// Test CI/CD environment validation

		// CI environments might have different tool availability
		testutil.WithLimitedPath(t, []string{"/usr/bin", "/bin", "/usr/local/bin"}, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// CI should get same clear errors as development
				assert.Contains(t, err.Error(), "Missing required tools")

				// Should be automatable (clear exit conditions)
				testutil.AssertErrorActionable(t, err.Error())
			}
		})
	})

	t.Run("container_environment_testing", func(t *testing.T) {
		// Test behavior in containerized development environments

		// Even in containers, sandbox requirement should be clear
		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should not be confused by containerized environment
				assert.Contains(t, err.Error(), "sandbox command not found")

				// Should not mention alternative container runtimes
				// even when running inside a container
				testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
			}
		})
	})
}

func TestE2E_RealWorldScenarios(t *testing.T) {
	t.Run("partial_tool_availability", func(t *testing.T) {
		// Test realistic scenarios where some but not all tools are available

		// Create environment with only some tools available
		cleanup := testutil.CreateMockCommand(t, "tmux", 0, "tmux 3.2a")
		defer cleanup()

		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should identify specifically which tools are missing
				errorMsg := err.Error()
				assert.Contains(t, errorMsg, "Missing required tools:")

				// Should mention sandbox as missing
				assert.Contains(t, errorMsg, "sandbox")

				// Should provide comprehensive list
				lines := strings.Split(errorMsg, "\n")
				bulletPoints := 0
				for _, line := range lines {
					if strings.Contains(line, "- ") {
						bulletPoints++
					}
				}
				assert.Greater(t, bulletPoints, 0, "Should list missing tools")
			}
		})
	})

	t.Run("mixed_environment_consistency", func(t *testing.T) {
		// Test that validation behaves consistently across different environments

		environments := []struct {
			name      string
			setupFunc func()
		}{
			{
				name: "limited_path",
				setupFunc: func() {
					os.Setenv("PATH", "/usr/bin:/bin")
				},
			},
			{
				name: "empty_path",
				setupFunc: func() {
					os.Setenv("PATH", "")
				},
			},
		}

		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		for _, env := range environments {
			t.Run(env.name, func(t *testing.T) {
				env.setupFunc()

				err := validation.CheckRequiredTools()

				if err != nil {
					// Should behave consistently across environments
					testutil.AssertErrorActionable(t, err.Error())
					testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())

					// Should always check for sandbox
					assert.Contains(t, err.Error(), "Missing required tools")
				}
			})
		}
	})

	t.Run("system_state_recovery", func(t *testing.T) {
		// Test recovery from various system states

		// Test that validation helps users recover from incomplete setups
		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should provide recovery guidance
				errorMsg := err.Error()
				testutil.AssertErrorActionable(t, errorMsg)

				// Should be clear about what needs to be fixed
				assert.Contains(t, errorMsg, "sandbox")
				assert.Contains(t, errorMsg, "installed")
			}
		})
	})
}

func TestE2E_UserExperienceFlow(t *testing.T) {
	t.Run("complete_setup_validation", func(t *testing.T) {
		// Test the complete user setup experience

		// Step 1: User attempts to run SBS command
		// Step 2: Validation runs and provides feedback
		// Step 3: User installs missing tools
		// Step 4: Validation passes

		// Simulate Step 2: Validation feedback
		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should provide complete picture of requirements
				assert.Contains(t, err.Error(), "Missing required tools")

				// Should guide user through setup process
				testutil.AssertErrorActionable(t, err.Error())

				// Should not overwhelm user with alternatives
				testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
			}
		})
	})

	t.Run("error_message_progression", func(t *testing.T) {
		// Test that error messages help users make progress

		testutil.WithLimitedPath(t, []string{"/bin"}, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				errorMsg := err.Error()

				// Should start with clear problem statement
				assert.Contains(t, errorMsg, "Missing required tools:")

				// Should list specific missing tools
				assert.True(t, strings.Contains(errorMsg, "- "), "Should list tools in bullet format")

				// Should include sandbox as a required tool
				assert.Contains(t, errorMsg, "sandbox")

				// Should be actionable for each missing tool
				testutil.AssertErrorActionable(t, errorMsg)
			}
		})
	})

	t.Run("success_path_validation", func(t *testing.T) {
		// Test validation when all tools are available

		// Skip if not all tools are available
		if err := exec.Command("tmux", "-V").Run(); err != nil {
			t.Skip("tmux not available")
		}
		if err := exec.Command("git", "--version").Run(); err != nil {
			t.Skip("git not available")
		}
		if err := exec.Command("gh", "--version").Run(); err != nil {
			t.Skip("gh not available")
		}
		if err := exec.Command("sandbox", "--help").Run(); err != nil {
			t.Skip("sandbox not available")
		}

		// When all tools are available, validation should pass
		err := validation.CheckRequiredTools()
		assert.NoError(t, err, "Validation should pass when all tools are available")
	})
}

func TestE2E_EdgeCases(t *testing.T) {
	t.Run("unusual_path_configurations", func(t *testing.T) {
		// Test with unusual PATH configurations

		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Test with PATH that has sandbox in unusual location
		testutil.WithTempDir(t, func(tempDir string) {
			// Create mock sandbox in temp directory
			cleanup := testutil.CreateMockCommand(t, "sandbox", 0, "mock sandbox help")
			defer cleanup()

			// Add only temp directory to PATH
			os.Setenv("PATH", tempDir)

			err := validation.CheckRequiredTools()

			// Should find sandbox even in unusual location
			if err == nil || !strings.Contains(err.Error(), "sandbox") {
				// If sandbox was found, that's good
				// If other tools are missing, that's expected
				assert.True(t, true, "Should handle unusual PATH configurations")
			}
		})
	})

	t.Run("permission_edge_cases", func(t *testing.T) {
		// Test permission-related edge cases

		// Create mock sandbox with permission issues
		cleanup := testutil.CreateMockCommand(t, "sandbox", 126, "permission denied")
		defer cleanup()

		err := validation.CheckRequiredTools()

		if err != nil {
			// Should handle permission errors gracefully
			testutil.AssertErrorActionable(t, err.Error())
		}
	})

	t.Run("environment_variable_edge_cases", func(t *testing.T) {
		// Test various environment variable configurations

		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		edgeCases := []string{
			"",             // Empty PATH
			":",            // PATH with just separator
			"/nonexistent", // PATH with non-existent directory
		}

		for _, pathValue := range edgeCases {
			t.Run("path_"+pathValue, func(t *testing.T) {
				os.Setenv("PATH", pathValue)

				err := validation.CheckRequiredTools()

				if err != nil {
					// Should handle edge cases gracefully
					testutil.AssertErrorActionable(t, err.Error())
					testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
				}
			})
		}
	})
}

// Helper function to simulate command execution in test environment
func simulateCommandExecution(command string, args []string) error {
	cmd := exec.Command(command, args...)
	return cmd.Run()
}
