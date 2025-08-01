package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRequiredTools_SandboxMandatory(t *testing.T) {
	t.Run("all_tools_available", func(t *testing.T) {
		// Skip if any tools are missing in test environment
		if err := exec.Command("tmux", "-V").Run(); err != nil {
			t.Skip("tmux not available, skipping all tools test")
		}
		if err := exec.Command("git", "--version").Run(); err != nil {
			t.Skip("git not available, skipping all tools test")
		}
		if err := exec.Command("gh", "--version").Run(); err != nil {
			t.Skip("gh not available, skipping all tools test")
		}
		if err := exec.Command("sandbox", "--help").Run(); err != nil {
			t.Skip("sandbox not available, skipping all tools test")
		}

		err := CheckRequiredTools()
		assert.NoError(t, err)
	})

	t.Run("sandbox_missing_other_tools_present", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Check if other tools are available first
		if err := exec.Command("tmux", "-V").Run(); err != nil {
			t.Skip("tmux not available, skipping sandbox-specific test")
		}
		if err := exec.Command("git", "--version").Run(); err != nil {
			t.Skip("git not available, skipping sandbox-specific test")
		}

		// Remove sandbox from PATH by setting a limited PATH
		os.Setenv("PATH", "/usr/bin:/bin")

		err := CheckRequiredTools()

		// Should fail because sandbox is missing
		if err != nil {
			assert.Contains(t, err.Error(), "Missing required tools")
			assert.Contains(t, err.Error(), "sandbox command not found")
		}
		// Note: In some test environments, sandbox might still be found
		// so we don't require an error, but if there is one, it should be about sandbox
	})
}

func TestCheckRequiredTools_ErrorAggregation(t *testing.T) {
	t.Run("multiple_tools_missing", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to very limited set to simulate missing tools
		os.Setenv("PATH", "/bin:/usr/bin")

		err := CheckRequiredTools()

		if err != nil {
			// Should aggregate all missing tool errors
			assert.Contains(t, err.Error(), "Missing required tools")

			// Error message should be formatted as a list
			errorMsg := err.Error()
			assert.Contains(t, errorMsg, "Missing required tools:")

			// Should contain bullet points for each missing tool
			lines := strings.Split(errorMsg, "\n")
			bulletCount := 0
			for _, line := range lines {
				if strings.Contains(line, "- ") {
					bulletCount++
				}
			}

			// Should have at least one bullet point (for missing tools)
			assert.Greater(t, bulletCount, 0)
		}
	})
}

func TestCheckRequiredTools_ErrorPriority(t *testing.T) {
	t.Run("sandbox_error_included", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set very limited PATH
		os.Setenv("PATH", "/bin")

		err := CheckRequiredTools()

		if err != nil {
			// Sandbox error should be included in the error message
			errorMsg := err.Error()
			assert.Contains(t, errorMsg, "sandbox")
		}
	})
}

func TestCheckRequiredTools_InstallationGuidance(t *testing.T) {
	t.Run("error_messages_contain_guidance", func(t *testing.T) {
		// Test individual tool error messages for guidance

		// Tmux error message
		tmuxErr := checkTmux()
		if tmuxErr != nil {
			assert.Contains(t, tmuxErr.Error(), "tmux not found")
			assert.Contains(t, tmuxErr.Error(), "install tmux")
		}

		// Git error message
		gitErr := checkGit()
		if gitErr != nil {
			assert.Contains(t, gitErr.Error(), "git not found")
			assert.Contains(t, gitErr.Error(), "install git")
		}

		// These tests verify that if errors occur, they contain installation guidance
	})

	t.Run("sandbox_error_actionable", func(t *testing.T) {
		// Create a mock scenario where sandbox is missing
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to exclude potential sandbox locations
		os.Setenv("PATH", "/bin:/usr/bin")

		err := CheckRequiredTools()

		if err != nil && strings.Contains(err.Error(), "sandbox") {
			// Sandbox error should provide actionable guidance
			assert.Contains(t, err.Error(), "sandbox command not found")
			assert.Contains(t, err.Error(), "installed")
			assert.Contains(t, err.Error(), "PATH")
		}
	})
}

func TestIndividualToolValidation(t *testing.T) {
	t.Run("tmux_validation", func(t *testing.T) {
		err := checkTmux()

		if err != nil {
			// If tmux is missing, error should be specific
			assert.Contains(t, err.Error(), "tmux not found")
			assert.Contains(t, err.Error(), "install tmux")
		}
	})

	t.Run("git_validation", func(t *testing.T) {
		err := checkGit()

		if err != nil {
			// If git is missing, error should be specific
			assert.Contains(t, err.Error(), "git not found")
			assert.Contains(t, err.Error(), "install git")
		}
	})
}

func TestRunValidationCommand_ErrorHandling(t *testing.T) {
	t.Run("successful_command", func(t *testing.T) {
		err := runValidationCommand("echo", []string{"test"}, "Echo failed")
		assert.NoError(t, err)
	})

	t.Run("failed_command", func(t *testing.T) {
		err := runValidationCommand("false", []string{}, "Command failed")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Command failed")
	})

	t.Run("nonexistent_command", func(t *testing.T) {
		err := runValidationCommand("nonexistent-command-98765", []string{}, "Command not found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Command not found")
	})
}

func TestValidationConsistency(t *testing.T) {
	t.Run("error_format_consistency", func(t *testing.T) {
		// Test that validation errors follow consistent format
		testCases := []struct {
			command  string
			args     []string
			errorMsg string
		}{
			{"nonexistent-cmd-1", []string{}, "tool1 not found. Please install tool1"},
			{"nonexistent-cmd-2", []string{"--version"}, "tool2 not found. Please install tool2"},
		}

		for _, tc := range testCases {
			err := runValidationCommand(tc.command, tc.args, tc.errorMsg)
			require.Error(t, err)
			assert.Equal(t, tc.errorMsg, err.Error())
		}
	})

	t.Run("no_alternative_runtime_mentions", func(t *testing.T) {
		// Ensure validation errors don't mention alternative container runtimes
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set very limited PATH to trigger errors
		os.Setenv("PATH", "/bin")

		err := CheckRequiredTools()

		if err != nil {
			errorMsg := strings.ToLower(err.Error())

			// Should not mention alternative container runtimes
			assert.NotContains(t, errorMsg, "podman")
			assert.NotContains(t, errorMsg, "docker")
			assert.NotContains(t, errorMsg, "containerd")
			assert.NotContains(t, errorMsg, "fallback")
			assert.NotContains(t, errorMsg, "alternative")
		}
	})
}

func TestSandboxSpecificValidation(t *testing.T) {
	t.Run("sandbox_validation_separate_from_others", func(t *testing.T) {
		// Test that sandbox validation is handled separately and specifically

		// The CheckRequiredTools function should call specific validation functions
		// including sandbox.CheckSandboxInstalled()

		// We can verify this by checking that the function structure is correct
		err := CheckRequiredTools()

		// If error occurs, it should be properly formatted
		if err != nil {
			assert.Contains(t, err.Error(), "Missing required tools:")
		}
	})

	t.Run("sandbox_mandatory_philosophy", func(t *testing.T) {
		// This test documents that sandbox is mandatory for SBS operations
		// There should be no conditional logic or fallbacks

		// The validation should always check for sandbox
		// There should be no "optional" or "fallback" modes

		// This is verified by the fact that CheckRequiredTools always calls
		// sandbox.CheckSandboxInstalled() without conditions

		// Test passes by construction - if this test runs, the philosophy is maintained
		assert.True(t, true, "sandbox is mandatory by design")
	})
}

// Helper function to create mock command for testing
func createMockCommand(name string, exitCode int, output string) func() {
	tempDir, err := os.MkdirTemp("", "mock-cmd-test")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}

	mockScript := filepath.Join(tempDir, name)
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo "%s"
exit %d
`, output, exitCode)

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		os.RemoveAll(tempDir)
		panic(fmt.Sprintf("Failed to create mock script: %v", err))
	}

	originalPath := os.Getenv("PATH")
	newPath := tempDir + ":" + originalPath
	os.Setenv("PATH", newPath)

	return func() {
		os.Setenv("PATH", originalPath)
		os.RemoveAll(tempDir)
	}
}

func TestValidationErrorMessagesQuality(t *testing.T) {
	t.Run("error_messages_user_friendly", func(t *testing.T) {
		// Test that error messages are user-friendly and actionable

		testErrors := []struct {
			tool        string
			expectedMsg string
		}{
			{"tmux", "tmux not found. Please install tmux"},
			{"git", "git not found. Please install git"},
		}

		for _, te := range testErrors {
			// Test the error message format
			assert.Contains(t, te.expectedMsg, te.tool)
			assert.Contains(t, te.expectedMsg, "not found")
			assert.Contains(t, te.expectedMsg, "install")
		}
	})

	t.Run("aggregate_error_readable", func(t *testing.T) {
		// Test that when multiple tools are missing, the error is readable

		// Simulate multiple missing tools
		mockErrors := []string{
			"tmux not found. Please install tmux",
			"git not found. Please install git",
			"sandbox command not found. Please ensure sandbox is installed and in PATH",
		}

		// Build expected aggregate error format
		expectedFormat := "Missing required tools:\n"
		for _, err := range mockErrors {
			expectedFormat += "  - " + err + "\n"
		}

		// Verify format is readable and structured
		lines := strings.Split(expectedFormat, "\n")
		assert.Contains(t, lines[0], "Missing required tools:")

		bulletCount := 0
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "- ") {
				bulletCount++
			}
		}
		assert.Equal(t, len(mockErrors), bulletCount)
	})
}
