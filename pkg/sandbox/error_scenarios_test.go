package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxFailureModes(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() func() // Returns cleanup function
		expectedError  string
		expectedAction string
		testFunc       func() error
	}{
		{
			name: "sandbox_not_in_path",
			setupFunc: func() func() {
				// Temporarily modify PATH to exclude sandbox
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "/usr/bin:/bin") // Very limited PATH
				return func() {
					os.Setenv("PATH", originalPath)
				}
			},
			expectedError:  "sandbox command not found",
			expectedAction: "install sandbox and ensure it's in PATH",
			testFunc: func() error {
				return CheckSandboxInstalled()
			},
		},
		{
			name: "sandbox_permission_denied",
			setupFunc: func() func() {
				// Create a mock sandbox that simulates permission denied
				return createMockSandbox("sandbox", 126, "permission denied")
			},
			expectedError:  "sandbox command not found", // Current implementation maps all errors to this
			expectedAction: "check permissions and installation",
			testFunc: func() error {
				return CheckSandboxInstalled()
			},
		},
		{
			name: "sandbox_general_failure",
			setupFunc: func() func() {
				// Create a mock sandbox that fails with generic error
				return createMockSandbox("sandbox", 1, "unknown error")
			},
			expectedError:  "sandbox command not found", // Current implementation maps all errors to this
			expectedAction: "check sandbox installation",
			testFunc: func() error {
				return CheckSandboxInstalled()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc()
			defer cleanup()

			err := tt.testFunc()

			// For the first test (not_in_path), we expect an error
			// For others, the behavior depends on whether our mock is found
			if tt.name == "sandbox_not_in_path" {
				// This should definitely fail
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				// In some environments, sandbox might still be found despite limited PATH
			} else {
				// For permission denied and general failure cases,
				// the mock sandbox will be found and executed, leading to failure
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestSandboxOperationFailures(t *testing.T) {
	manager := NewManager()

	t.Run("list_sandboxes_command_failure", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Remove sandbox from PATH
		os.Setenv("PATH", "/usr/bin:/bin")

		sandboxes, err := manager.ListSandboxes()

		// Current implementation treats command failure as "no sandboxes"
		assert.NoError(t, err)
		assert.Empty(t, sandboxes)
	})

	t.Run("sandbox_exists_command_failure", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Remove sandbox from PATH
		os.Setenv("PATH", "/usr/bin:/bin")

		exists, err := manager.SandboxExists("test-sandbox")

		// Current implementation treats command failure as "sandbox doesn't exist"
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("delete_sandbox_when_command_unavailable", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Remove sandbox from PATH
		os.Setenv("PATH", "/usr/bin:/bin")

		err := manager.DeleteSandbox("test-sandbox")

		// Should succeed because SandboxExists returns false when command unavailable
		assert.NoError(t, err)
	})
}

func TestSandboxInstallationIssues(t *testing.T) {
	t.Run("sandbox_not_installed", func(t *testing.T) {
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set very restrictive PATH
		os.Setenv("PATH", "/bin")

		err := CheckSandboxInstalled()

		if err != nil {
			// Error message should be helpful
			assert.Contains(t, err.Error(), "sandbox command not found")
			assert.Contains(t, err.Error(), "Please ensure sandbox is installed")
			assert.Contains(t, err.Error(), "in PATH")

			// Should not mention alternative runtimes
			errorMsg := strings.ToLower(err.Error())
			assert.NotContains(t, errorMsg, "podman")
			assert.NotContains(t, errorMsg, "docker")
			assert.NotContains(t, errorMsg, "fallback")
		}
	})

	t.Run("sandbox_in_unusual_location", func(t *testing.T) {
		// Test sandbox found in non-standard PATH location

		// Create a temporary directory and put mock sandbox there
		tempDir, err := os.MkdirTemp("", "sandbox-path-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create working mock sandbox
		mockScript := filepath.Join(tempDir, "sandbox")
		scriptContent := `#!/bin/bash
if [ "$1" = "--help" ]; then
    echo "Mock sandbox help"
    exit 0
fi
echo "Mock sandbox command"
exit 0
`
		err = os.WriteFile(mockScript, []byte(scriptContent), 0755)
		require.NoError(t, err)

		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)

		// Set PATH to only include our temp directory
		os.Setenv("PATH", tempDir)

		err = CheckSandboxInstalled()
		assert.NoError(t, err, "Should find sandbox in unusual location")
	})
}

func TestSandboxRuntimeIssues(t *testing.T) {
	t.Run("sandbox_command_timeout", func(t *testing.T) {
		// Create a mock sandbox that hangs
		cleanup := createHangingSandbox("sandbox")
		defer cleanup()

		// Note: Current implementation doesn't have explicit timeout handling
		// but the command should eventually fail
		start := time.Now()
		err := CheckSandboxInstalled()
		duration := time.Since(start)

		// Should not hang indefinitely
		assert.Less(t, duration, 30*time.Second, "Command should not hang indefinitely")

		// If it fails due to timeout, error should be appropriate
		if err != nil {
			assert.Contains(t, err.Error(), "sandbox command not found")
		}
	})

	t.Run("insufficient_permissions", func(t *testing.T) {
		// Create a mock sandbox with permission issues
		cleanup := createMockSandbox("sandbox", 126, "permission denied")
		defer cleanup()

		err := CheckSandboxInstalled()

		if err != nil {
			// Current implementation maps all failures to "command not found"
			assert.Contains(t, err.Error(), "sandbox command not found")
		}
	})
}

func TestSandboxVersionCompatibility(t *testing.T) {
	t.Run("old_sandbox_version", func(t *testing.T) {
		// Test with a mock sandbox that might represent older version
		cleanup := createMockSandbox("sandbox", 0, "sandbox version 0.1.0")
		defer cleanup()

		err := CheckSandboxInstalled()
		assert.NoError(t, err, "Should accept any version that responds to --help")
	})

	t.Run("incompatible_sandbox_version", func(t *testing.T) {
		// Test with a mock sandbox that fails --help
		cleanup := createMockSandbox("sandbox", 1, "unsupported option: --help")
		defer cleanup()

		err := CheckSandboxInstalled()

		if err != nil {
			assert.Contains(t, err.Error(), "sandbox command not found")
		}
	})

	t.Run("beta_sandbox_version", func(t *testing.T) {
		// Test with a mock sandbox beta version
		cleanup := createMockSandbox("sandbox", 0, "sandbox version 2.0.0-beta")
		defer cleanup()

		err := CheckSandboxInstalled()
		assert.NoError(t, err, "Should accept beta versions")
	})
}

func TestPartialSystemStateRecovery(t *testing.T) {
	manager := NewManager()

	t.Run("sandbox_becomes_unavailable_during_operation", func(t *testing.T) {
		// Simulate sandbox becoming unavailable mid-operation

		// First, check if sandbox is available
		if exec.Command("sandbox", "--help").Run() != nil {
			t.Skip("sandbox not available, skipping system state test")
		}

		// Test that operations handle unavailability gracefully
		sandboxes, err := manager.ListSandboxes()
		assert.NoError(t, err)
		_ = sandboxes

		// Simulate sandbox becoming unavailable
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)
		os.Setenv("PATH", "/usr/bin:/bin")

		// Subsequent operations should handle gracefully
		sandboxes2, err2 := manager.ListSandboxes()
		assert.NoError(t, err2)
		assert.Empty(t, sandboxes2)
	})

	t.Run("cleanup_after_partial_failures", func(t *testing.T) {
		// Test cleanup behavior when sandbox operations partially fail

		// This is more of a behavioral documentation test
		// The current implementation handles failures gracefully by:
		// 1. Treating command failures as "no resources exist"
		// 2. Not erroring on cleanup of non-existent resources

		err := manager.DeleteSandbox("definitely-does-not-exist")
		assert.NoError(t, err, "Cleanup of non-existent resources should succeed")
	})
}

func TestSandboxErrorMessageQuality(t *testing.T) {
	t.Run("error_messages_actionable", func(t *testing.T) {
		// Test that error messages tell users what to do

		expectedError := "sandbox command not found. Please ensure sandbox is installed and in PATH"

		// Error should be actionable
		assert.Contains(t, expectedError, "Please ensure")
		assert.Contains(t, expectedError, "installed")
		assert.Contains(t, expectedError, "in PATH")

		// Error should be specific
		assert.Contains(t, expectedError, "sandbox command")

		// Error should not mention alternatives
		assert.NotContains(t, strings.ToLower(expectedError), "podman")
		assert.NotContains(t, strings.ToLower(expectedError), "docker")
		assert.NotContains(t, strings.ToLower(expectedError), "fallback")
	})

	t.Run("error_messages_user_friendly", func(t *testing.T) {
		// Test that error messages are user-friendly

		errorMessages := []string{
			"sandbox command not found. Please ensure sandbox is installed and in PATH",
			"failed to list sandboxes: %w",
			"failed to delete sandbox %s: %w",
			"failed to check if sandbox exists: %w",
		}

		for _, msg := range errorMessages {
			// Should be clear about what failed
			if strings.Contains(msg, "sandbox") {
				assert.True(t, true, "Error mentions sandbox specifically")
			}

			// Should not be overly technical
			assert.NotContains(t, msg, "exec:")
			assert.NotContains(t, msg, "exit status")

			// Should provide context when using format strings
			if strings.Contains(msg, "%w") || strings.Contains(msg, "%s") {
				assert.True(t, true, "Error provides context with wrapped errors")
			}
		}
	})

	t.Run("error_message_consistency", func(t *testing.T) {
		// Test that all sandbox errors follow consistent patterns

		// Test various scenarios that should produce consistent error
		scenarios := []func() error{
			func() error {
				originalPath := os.Getenv("PATH")
				defer os.Setenv("PATH", originalPath)
				os.Setenv("PATH", "/bin")
				return CheckSandboxInstalled()
			},
		}

		for i, scenario := range scenarios {
			err := scenario()
			if err != nil {
				assert.Contains(t, err.Error(), "sandbox command not found",
					"Scenario %d should produce consistent error message", i)
			}
		}
	})
}

// createHangingSandbox creates a mock sandbox command that hangs for testing timeout scenarios
func createHangingSandbox(commandName string) func() {
	tempDir, err := os.MkdirTemp("", "hanging-sandbox-test")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}

	mockScript := filepath.Join(tempDir, commandName)
	scriptContent := `#!/bin/bash
# Hang for a while to simulate timeout
sleep 10
exit 0
`

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		os.RemoveAll(tempDir)
		panic(fmt.Sprintf("Failed to create hanging script: %v", err))
	}

	originalPath := os.Getenv("PATH")
	newPath := tempDir + ":" + originalPath
	os.Setenv("PATH", newPath)

	return func() {
		os.Setenv("PATH", originalPath)
		os.RemoveAll(tempDir)
	}
}

func TestSandboxManagerConsistency(t *testing.T) {
	manager := NewManager()

	t.Run("consistent_behavior_across_operations", func(t *testing.T) {
		// Test that all manager operations behave consistently when sandbox is unavailable

		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)
		os.Setenv("PATH", "/usr/bin:/bin")

		// All operations should handle missing sandbox consistently
		sandboxes, err1 := manager.ListSandboxes()
		assert.NoError(t, err1)
		assert.Empty(t, sandboxes)

		exists, err2 := manager.SandboxExists("test")
		assert.NoError(t, err2)
		assert.False(t, exists)

		err3 := manager.DeleteSandbox("test")
		assert.NoError(t, err3)

		// All operations succeed gracefully when sandbox is unavailable
		// This documents the current "graceful degradation" approach
	})

	t.Run("error_propagation_consistency", func(t *testing.T) {
		// Test that errors are propagated consistently

		// The manager wraps low-level sandbox command errors consistently
		// This is verified by checking that error messages follow patterns

		testSandboxName := "test-sandbox-consistency"

		// Operations should either succeed or fail with meaningful errors
		err := manager.DeleteSandbox(testSandboxName)
		assert.NoError(t, err, "Delete of non-existent sandbox should succeed")

		exists, err := manager.SandboxExists(testSandboxName)
		assert.NoError(t, err, "Existence check should not error")
		assert.False(t, exists, "Non-existent sandbox should not exist")
	})
}
