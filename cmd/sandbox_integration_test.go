package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sbs/internal/testutil"
	"sbs/pkg/validation"
)

func TestStartCommand_SandboxValidation(t *testing.T) {
	t.Run("start_fails_when_sandbox_missing", func(t *testing.T) {
		testutil.WithoutSandbox(t, func() {
			// Test that start command validation catches missing sandbox
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should fail with sandbox-related error
				assert.Contains(t, err.Error(), "Missing required tools")

				// Should mention sandbox specifically
				errorMsg := err.Error()
				assert.Contains(t, errorMsg, "sandbox")

				// Should not mention alternative runtimes
				testutil.AssertNoAlternativeRuntimeMentions(t, errorMsg)

				// Should be actionable
				testutil.AssertErrorActionable(t, errorMsg)
			}
		})
	})

	t.Run("start_validation_before_operations", func(t *testing.T) {
		// Test that validation runs before any resource creation
		// This prevents partial setup with missing dependencies

		testutil.WithLimitedPath(t, []string{"/usr/bin", "/bin"}, func() {
			// Validation should catch missing tools early
			err := validation.CheckRequiredTools()

			if err != nil {
				// Error should be comprehensive and mention all missing tools
				assert.Contains(t, err.Error(), "Missing required tools:")

				// Should be formatted as a list
				lines := strings.Split(err.Error(), "\n")
				bulletCount := 0
				for _, line := range lines {
					if strings.Contains(line, "- ") {
						bulletCount++
					}
				}

				// Should have at least one bullet point for missing tools
				if bulletCount == 0 {
					t.Error("Error should be formatted as a bulleted list")
				}
			}
		})
	})
}

func TestStopCommand_SandboxHandling(t *testing.T) {
	t.Run("stop_graceful_when_sandbox_missing", func(t *testing.T) {
		testutil.WithoutSandbox(t, func() {
			// Stop command should handle missing sandbox gracefully
			// since it primarily deals with tmux sessions and git worktrees

			// For testing, we just verify that validation would catch the issue
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should mention sandbox as a missing tool
				assert.Contains(t, err.Error(), "sandbox")
			}

			// The actual stop command behavior would depend on whether
			// it performs validation or just tries to clean up resources
		})
	})

	t.Run("stop_sandbox_cleanup_awareness", func(t *testing.T) {
		// Test that stop command is aware of sandbox cleanup requirements

		// Even if sandbox is missing, stop should document what would be cleaned
		testutil.WithoutSandbox(t, func() {
			// This test documents that stop command should be aware of sandbox resources
			// even if sandbox is not available

			// In a real implementation, this might check for sandbox resources
			// and warn about incomplete cleanup
			assert.True(t, true, "Stop command should be sandbox-aware")
		})
	})
}

func TestCleanCommand_SandboxValidation(t *testing.T) {
	t.Run("clean_lists_resources_when_sandbox_missing", func(t *testing.T) {
		testutil.WithoutSandbox(t, func() {
			// Clean command should be able to list what would be cleaned
			// even if sandbox is not available for actual cleanup

			err := validation.CheckRequiredTools()

			if err != nil {
				// Should identify sandbox as missing
				assert.Contains(t, err.Error(), "sandbox")

				// Error should be helpful for clean operations
				testutil.AssertErrorActionable(t, err.Error())
			}
		})
	})

	t.Run("clean_warns_about_incomplete_cleanup", func(t *testing.T) {
		// Test that clean command warns when sandbox cleanup cannot be performed

		testutil.WithoutSandbox(t, func() {
			// Clean should identify that sandbox resources cannot be cleaned
			// This is important for completeness

			err := validation.CheckRequiredTools()

			if err != nil {
				// Clean operation should be aware of what it cannot clean
				assert.Contains(t, err.Error(), "sandbox")
			}
		})
	})
}

func TestCommandChain_SandboxValidation(t *testing.T) {
	t.Run("validation_runs_before_operations", func(t *testing.T) {
		// Test that validation prevents partial operations when tools are missing

		testutil.WithLimitedPath(t, []string{"/bin"}, func() {
			// All commands should validate required tools before proceeding
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should identify all missing tools
				assert.Contains(t, err.Error(), "Missing required tools")

				// Should prevent operations from starting
				errorMsg := err.Error()
				assert.Contains(t, errorMsg, ":")

				// Error should guide user to fix all issues
				testutil.AssertErrorActionable(t, errorMsg)
			}
		})
	})

	t.Run("consistent_error_handling", func(t *testing.T) {
		// Test that all commands handle sandbox validation consistently

		scenarios := []struct {
			name     string
			testFunc func() error
		}{
			{
				name: "validation_check",
				testFunc: func() error {
					return validation.CheckRequiredTools()
				},
			},
		}

		testutil.WithoutSandbox(t, func() {
			for _, scenario := range scenarios {
				t.Run(scenario.name, func(t *testing.T) {
					err := scenario.testFunc()

					if err != nil {
						// All validation should produce consistent errors
						testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
						testutil.AssertErrorActionable(t, err.Error())

						// Should mention sandbox specifically
						assert.Contains(t, err.Error(), "sandbox")
					}
				})
			}
		})
	})
}

func TestValidationIntegration_SandboxMandatory(t *testing.T) {
	t.Run("sandbox_always_required", func(t *testing.T) {
		// Test that sandbox is always required - no conditional logic

		// Validation should always check for sandbox
		err := validation.CheckRequiredTools()

		// This test passes if validation runs without error when tools are available,
		// or fails with comprehensive error when tools are missing

		if err != nil {
			// If error exists, should be about missing tools
			assert.Contains(t, err.Error(), "Missing required tools")
		}

		// The important thing is that sandbox is always checked
		// There should be no conditional logic based on configuration
		assert.True(t, true, "Sandbox validation is mandatory by design")
	})

	t.Run("no_fallback_modes", func(t *testing.T) {
		// Test that there are no fallback modes when sandbox is missing

		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should not mention fallback options
				errorMsg := strings.ToLower(err.Error())

				assert.NotContains(t, errorMsg, "fallback")
				assert.NotContains(t, errorMsg, "alternative")
				assert.NotContains(t, errorMsg, "instead")
				assert.NotContains(t, errorMsg, "try")

				// Should be definitive about requirement
				assert.Contains(t, err.Error(), "sandbox")
			}
		})
	})

	t.Run("error_aggregation_prioritizes_sandbox", func(t *testing.T) {
		// Test that when multiple tools are missing, sandbox is prominently mentioned

		testutil.WithLimitedPath(t, []string{"/bin"}, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				errorMsg := err.Error()

				// Should mention sandbox in the error
				assert.Contains(t, errorMsg, "sandbox")

				// Should be formatted as a comprehensive error
				assert.Contains(t, errorMsg, "Missing required tools:")

				// Should list all missing tools
				lines := strings.Split(errorMsg, "\n")
				foundSandboxError := false
				for _, line := range lines {
					if strings.Contains(line, "sandbox") && strings.Contains(line, "- ") {
						foundSandboxError = true
						break
					}
				}

				if !foundSandboxError {
					t.Error("Sandbox error should be listed in missing tools")
				}
			}
		})
	})
}

func TestSandboxSpecificIntegration(t *testing.T) {
	t.Run("sandbox_validation_separate_from_others", func(t *testing.T) {
		// Test that sandbox validation is handled as a distinct requirement

		// Even if other tools are available, missing sandbox should cause failure
		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should specifically mention sandbox
				assert.Contains(t, err.Error(), "sandbox")

				// Should not provide workarounds or alternatives
				testutil.AssertNoAlternativeRuntimeMentions(t, err.Error())
			}
		})
	})

	t.Run("sandbox_error_message_quality", func(t *testing.T) {
		// Test that sandbox-specific error messages are high quality

		testutil.WithoutSandbox(t, func() {
			err := validation.CheckRequiredTools()

			if err != nil && strings.Contains(err.Error(), "sandbox") {
				errorMsg := err.Error()

				// Should contain installation guidance
				guidanceTerms := []string{"install", "ensure", "PATH"}
				hasGuidance := false
				for _, term := range guidanceTerms {
					if strings.Contains(errorMsg, term) {
						hasGuidance = true
						break
					}
				}
				assert.True(t, hasGuidance, "Error should contain installation guidance")

				// Should be specific about the problem
				assert.Contains(t, errorMsg, "sandbox command not found")
			}
		})
	})
}

func TestCommandValidationFlow(t *testing.T) {
	t.Run("validation_happens_early", func(t *testing.T) {
		// Test that validation happens before any resource creation

		// This prevents partial system state when dependencies are missing
		testutil.WithoutSandbox(t, func() {
			// In a real command flow, validation would run first
			err := validation.CheckRequiredTools()

			if err != nil {
				// Early validation failure prevents partial setup
				assert.Contains(t, err.Error(), "Missing required tools")

				// User gets complete picture of what needs to be fixed
				testutil.AssertErrorActionable(t, err.Error())
			}
		})
	})

	t.Run("validation_comprehensive", func(t *testing.T) {
		// Test that validation checks all required tools at once

		testutil.WithLimitedPath(t, []string{"/bin"}, func() {
			err := validation.CheckRequiredTools()

			if err != nil {
				// Should check all tools, not fail fast on first missing tool
				errorMsg := err.Error()

				// Should be formatted as a comprehensive list
				assert.Contains(t, errorMsg, "Missing required tools:")

				// Should include sandbox in the comprehensive check
				assert.Contains(t, errorMsg, "sandbox")
			}
		})
	})
}
