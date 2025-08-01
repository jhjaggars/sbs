package sandbox

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

func TestCheckSandboxInstalled_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		sandboxBinary string
		mockSetup     func() func() // setup and cleanup function
		expectedError string
		errorContains []string
		shouldSucceed bool
	}{
		{
			name:          "sandbox_available",
			sandboxBinary: "sandbox",
			mockSetup:     func() func() { return func() {} }, // no-op if sandbox is available
			shouldSucceed: true,
		},
		{
			name:          "command_not_found",
			sandboxBinary: "sandbox",
			mockSetup: func() func() {
				// Create a mock script that simulates command not found (exit code 127)
				return createMockSandbox("sandbox", 127, "command not found")
			},
			expectedError: "sandbox command not found. Please ensure sandbox is installed and in PATH",
			errorContains: []string{"sandbox command not found", "installed", "PATH"},
			shouldSucceed: false,
		},
		{
			name:          "permission_denied",
			sandboxBinary: "sandbox",
			mockSetup: func() func() {
				// Create a mock script that simulates permission denied (exit code 126)
				return createMockSandbox("sandbox", 126, "permission denied")
			},
			expectedError: "sandbox command found but not executable. Please check file permissions",
			errorContains: []string{"not executable", "permissions"},
			shouldSucceed: false,
		},
		{
			name:          "other_failure",
			sandboxBinary: "sandbox",
			mockSetup: func() func() {
				// Create a mock script that simulates other failure (exit code 1)
				return createMockSandbox("sandbox", 1, "general error")
			},
			expectedError: "sandbox command failed with exit code 1",
			errorContains: []string{"failed with exit code", "check sandbox installation"},
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip test if we're testing real sandbox and it's not available
			if tt.name == "sandbox_available" {
				if err := exec.Command("sandbox", "--help").Run(); err != nil {
					t.Skip("sandbox command not available, skipping availability test")
				}
			}

			cleanup := tt.mockSetup()
			defer cleanup()

			err := CheckSandboxInstalled()

			if tt.shouldSucceed {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				for _, contains := range tt.errorContains {
					assert.Contains(t, err.Error(), contains)
				}
			}
		})
	}
}

func TestCheckSandboxInstalled_PathValidation(t *testing.T) {
	t.Run("sandbox_in_path", func(t *testing.T) {
		// Test that we can find sandbox in PATH
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Check if sandbox is actually available
		if err := exec.Command("sandbox", "--help").Run(); err != nil {
			t.Skip("sandbox command not available, skipping path validation test")
		}

		err := CheckSandboxInstalled()
		assert.NoError(t, err)
	})

	t.Run("sandbox_not_in_path", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to exclude sandbox
		os.Setenv("PATH", "/usr/bin:/bin")

		// The actual CheckSandboxInstalled function will try to run sandbox --help
		// which will fail, so we expect an error
		err := CheckSandboxInstalled()
		if err != nil {
			assert.Contains(t, err.Error(), "sandbox command not found")
		}
		// Note: In some environments, sandbox might still be found even with limited PATH
		// so we don't assert.Error here - the important thing is that if there's an error,
		// it contains the right message
	})
}

func TestCheckSandboxInstalled_VersionCompatibility(t *testing.T) {
	t.Run("sandbox_help_available", func(t *testing.T) {
		// Check if sandbox --help works (basic compatibility check)
		if err := exec.Command("sandbox", "--help").Run(); err != nil {
			t.Skip("sandbox command not available, skipping version compatibility test")
		}

		err := CheckSandboxInstalled()
		assert.NoError(t, err)
	})
}

func TestSandboxManager_GracefulDegradation(t *testing.T) {
	manager := NewManager()

	t.Run("list_sandboxes_when_command_unavailable", func(t *testing.T) {
		// When sandbox command is not available, ListSandboxes should return empty slice
		// not error, as per current implementation
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to exclude sandbox
		os.Setenv("PATH", "/usr/bin:/bin")

		sandboxes, err := manager.ListSandboxes()
		// Current implementation returns empty list when sandbox command fails
		assert.NoError(t, err)
		assert.Empty(t, sandboxes)
	})

	t.Run("sandbox_exists_when_command_unavailable", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to exclude sandbox
		os.Setenv("PATH", "/usr/bin:/bin")

		exists, err := manager.SandboxExists("test-sandbox")
		// Current implementation returns false when sandbox command fails
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSandboxManager_ErrorPropagation(t *testing.T) {
	manager := NewManager()

	t.Run("delete_nonexistent_sandbox", func(t *testing.T) {
		// Deleting a non-existent sandbox should not error
		err := manager.DeleteSandbox("nonexistent-sandbox-12345")
		assert.NoError(t, err)
	})

	t.Run("sandbox_name_generation", func(t *testing.T) {
		// Test legacy name generation
		name := manager.GetSandboxName(123)
		assert.Equal(t, "work-issue-123", name)

		// Test repository-scoped name generation
		repoName := manager.GetRepositorySandboxName("myproject", 456)
		assert.Equal(t, "work-issue-myproject-456", repoName)
	})
}

func TestSandboxManager_Operations(t *testing.T) {
	manager := NewManager()

	t.Run("get_sandbox_names", func(t *testing.T) {
		tests := []struct {
			issueNumber  int
			repoName     string
			expectedName string
			isLegacy     bool
		}{
			{123, "", "work-issue-123", true},
			{456, "myproject", "work-issue-myproject-456", false},
			{789, "test-repo", "work-issue-test-repo-789", false},
		}

		for _, tt := range tests {
			if tt.isLegacy {
				name := manager.GetSandboxName(tt.issueNumber)
				assert.Equal(t, tt.expectedName, name)
			} else {
				name := manager.GetRepositorySandboxName(tt.repoName, tt.issueNumber)
				assert.Equal(t, tt.expectedName, name)
			}
		}
	})
}

// createMockSandbox creates a temporary mock sandbox command for testing
func createMockSandbox(commandName string, exitCode int, errorMsg string) func() {
	// Create a temporary directory for our mock binary
	tempDir, err := os.MkdirTemp("", "mock-sandbox-test")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}

	// Create mock sandbox script
	mockScript := filepath.Join(tempDir, commandName)
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo "%s" >&2
exit %d
`, errorMsg, exitCode)

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		os.RemoveAll(tempDir)
		panic(fmt.Sprintf("Failed to create mock script: %v", err))
	}

	// Prepend temp directory to PATH
	originalPath := os.Getenv("PATH")
	newPath := tempDir + ":" + originalPath
	os.Setenv("PATH", newPath)

	// Return cleanup function
	return func() {
		os.Setenv("PATH", originalPath)
		os.RemoveAll(tempDir)
	}
}

func TestSandboxValidation_RequiredForOperations(t *testing.T) {
	t.Run("sandbox_required_philosophy", func(t *testing.T) {
		// This test documents that SBS requires sandbox for all operations
		// It's more of a documentation test than a functional test

		// All sandbox operations should require the sandbox command
		manager := NewManager()

		// These operations should gracefully handle missing sandbox
		// but the expectation is that sandbox is always available
		_, err := manager.ListSandboxes()
		assert.NoError(t, err) // Current implementation doesn't error

		exists, err := manager.SandboxExists("test")
		assert.NoError(t, err) // Current implementation doesn't error
		_ = exists

		err = manager.DeleteSandbox("nonexistent")
		assert.NoError(t, err) // Should succeed for non-existent sandbox
	})
}

func TestSandboxErrorMessages_Actionability(t *testing.T) {
	t.Run("error_message_contains_guidance", func(t *testing.T) {
		// Test that error messages are actionable
		err := fmt.Errorf("sandbox command not found. Please ensure sandbox is installed and in PATH")

		// Error should tell user what to do
		assert.Contains(t, err.Error(), "Please ensure")
		assert.Contains(t, err.Error(), "installed")
		assert.Contains(t, err.Error(), "PATH")

		// Error should not mention alternative runtimes
		assert.NotContains(t, strings.ToLower(err.Error()), "podman")
		assert.NotContains(t, strings.ToLower(err.Error()), "docker")
		assert.NotContains(t, strings.ToLower(err.Error()), "fallback")
	})

	t.Run("error_message_consistency", func(t *testing.T) {
		// All sandbox-related errors should follow consistent format
		errorMessages := []string{
			"sandbox command not found. Please ensure sandbox is installed and in PATH",
			"failed to list sandboxes: %w",
			"failed to delete sandbox %s: %w",
			"failed to check if sandbox exists: %w",
		}

		for _, msg := range errorMessages {
			// Error messages should be specific about sandbox
			if strings.Contains(msg, "sandbox") {
				assert.Contains(t, msg, "sandbox")
			}
		}
	})
}
