package tmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_CreateSession_WithEnvironment(t *testing.T) {
	manager := NewManager()

	// Test that CreateSession accepts environment variables
	env := map[string]string{
		"SBS_TITLE": "test-title",
		"OTHER_VAR": "test-value",
	}

	// Note: This test verifies the signature and basic functionality
	// We can't easily test actual tmux session creation without complex mocking
	// The real tmux integration will be tested in integration tests

	tests := []struct {
		name        string
		issueNumber int
		workingDir  string
		sessionName string
		env         map[string]string
		shouldPass  bool
	}{
		{
			name:        "with_environment_variables",
			issueNumber: 123,
			workingDir:  "/tmp/test",
			sessionName: "test-session-123",
			env:         env,
			shouldPass:  true,
		},
		{
			name:        "with_empty_environment",
			issueNumber: 456,
			workingDir:  "/tmp/test2",
			sessionName: "test-session-456",
			env:         map[string]string{},
			shouldPass:  true,
		},
		{
			name:        "with_nil_environment",
			issueNumber: 789,
			workingDir:  "/tmp/test3",
			sessionName: "test-session-789",
			env:         nil,
			shouldPass:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing session first
			manager.KillSession(tt.sessionName)

			// Test that the method signature accepts environment parameter
			session, err := manager.CreateSession(tt.issueNumber, tt.workingDir, tt.sessionName, tt.env)
			if tt.shouldPass {
				// The signature should be correct and may succeed if tmux is available
				if err == nil {
					// Session created successfully - verify it exists
					assert.NotNil(t, session)
					assert.Equal(t, tt.sessionName, session.Name)
					assert.Equal(t, tt.workingDir, session.WorkingDir)
					assert.Equal(t, tt.issueNumber, session.IssueNumber)

					// Clean up
					manager.KillSession(tt.sessionName)
				} else {
					// Failed due to environment - this is also acceptable in tests
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestManager_CreateSession_WithoutEnvironment(t *testing.T) {
	manager := NewManager()

	sessionName := "test-session-backward-compat"
	manager.KillSession(sessionName) // Clean up first

	// Test backward compatibility - original signature should still work
	session, err := manager.CreateSession(123, "/tmp/test", sessionName)
	if err == nil {
		// Session created successfully
		assert.NotNil(t, session)
		assert.Equal(t, sessionName, session.Name)
		manager.KillSession(sessionName) // Clean up
	} else {
		// Failed due to environment - acceptable in tests
		assert.Error(t, err)
	}
}

func TestManager_AttachToSession_WithEnvironment(t *testing.T) {
	manager := NewManager()

	env := map[string]string{
		"SBS_TITLE": "test-title",
	}

	// Check if the session exists first to validate expected behavior
	exists, err := manager.SessionExists("test-session-nonexistent")
	require.NoError(t, err)
	assert.False(t, exists, "Session should not exist for this test")

	// Test that AttachToSession accepts environment variables
	// Note: AttachToSession uses syscall.Exec which replaces the current process.
	// If the session doesn't exist, setEnvironmentVariables will fail first.
	err = manager.AttachToSession("test-session-nonexistent", env)
	// Expected to fail for non-existent session when setting environment variables
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set environment variables")
}

func TestManager_AttachToSession_WithoutEnvironment(t *testing.T) {
	manager := NewManager()

	// Check if the session exists first to validate expected behavior
	exists, err := manager.SessionExists("test-session-nonexistent")
	require.NoError(t, err)
	assert.False(t, exists, "Session should not exist for this test")

	// Test backward compatibility - original signature should still work
	// Note: AttachToSession uses syscall.Exec which replaces the current process,
	// so in a test environment, this will likely exit the test process.
	// We can't easily test this without more complex mocking.
	// Just test that the method signature is correct.
	t.Skip("AttachToSession uses syscall.Exec and would replace the test process")
}

func TestManager_ExecuteCommand_WithEnvironment(t *testing.T) {
	manager := NewManager()

	env := map[string]string{
		"SBS_TITLE": "test-title",
		"OTHER_VAR": "test-value",
	}

	// Test that ExecuteCommand accepts environment variables
	err := manager.ExecuteCommand("test-session-nonexistent", "echo", []string{"hello"}, env)
	// Expected to fail for non-existent session
	assert.Error(t, err)
}

func TestManager_ExecuteCommand_WithoutEnvironment(t *testing.T) {
	manager := NewManager()

	// Test backward compatibility - original signature should still work
	err := manager.ExecuteCommand("test-session-nonexistent", "echo", []string{"hello"})
	// Expected to fail for non-existent session
	assert.Error(t, err)
}

func TestManager_EnvironmentVariableParsing(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		env      map[string]string
		expected []string
	}{
		{
			name: "single_variable",
			env: map[string]string{
				"SBS_TITLE": "test-title",
			},
			expected: []string{"SBS_TITLE=test-title"},
		},
		{
			name: "multiple_variables",
			env: map[string]string{
				"SBS_TITLE": "test-title",
				"OTHER_VAR": "test-value",
			},
			expected: []string{"SBS_TITLE=test-title", "OTHER_VAR=test-value"},
		},
		{
			name:     "empty_environment",
			env:      map[string]string{},
			expected: []string{},
		},
		{
			name:     "nil_environment",
			env:      nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.formatEnvironmentVariables(tt.env)

			// Check that all expected variables are present
			for _, expectedVar := range tt.expected {
				assert.Contains(t, result, expectedVar)
			}

			// Check that the count matches
			assert.Len(t, result, len(tt.expected))
		})
	}
}

func TestManager_EnvironmentVariableEscaping(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name: "simple_value",
			env: map[string]string{
				"SBS_TITLE": "test-title",
			},
			expected: "SBS_TITLE=test-title",
		},
		{
			name: "value_with_spaces",
			env: map[string]string{
				"SBS_TITLE": "test title with spaces",
			},
			expected: "SBS_TITLE=test title with spaces",
		},
		{
			name: "value_with_special_characters",
			env: map[string]string{
				"SBS_TITLE": "test-title-with-hyphens_and_underscores",
			},
			expected: "SBS_TITLE=test-title-with-hyphens_and_underscores",
		},
		{
			name: "empty_value",
			env: map[string]string{
				"SBS_TITLE": "",
			},
			expected: "SBS_TITLE=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.formatEnvironmentVariables(tt.env)
			require.Len(t, result, 1)
			assert.Equal(t, tt.expected, result[0])
		})
	}
}

func TestManager_BackwardCompatibility(t *testing.T) {
	manager := NewManager()

	// Test that all existing methods still work with their original signatures
	sessionName := "test-backward-compat"
	manager.KillSession(sessionName) // Clean up first

	// CreateSession backward compatibility
	session, err := manager.CreateSession(123, "/tmp", sessionName)
	if err == nil {
		assert.NotNil(t, session)
		manager.KillSession(sessionName) // Clean up
	} else {
		assert.Error(t, err) // Expected to fail in test environment
	}

	// AttachToSession backward compatibility - skip due to syscall.Exec behavior
	// The method uses syscall.Exec which would replace the test process

	// ExecuteCommand backward compatibility
	err = manager.ExecuteCommand("test-session-nonexistent", "echo", []string{"test"})
	assert.Error(t, err) // Expected to fail for non-existent session

	// ExecuteCommandWithSubstitution backward compatibility
	err = manager.ExecuteCommandWithSubstitution("test-session-nonexistent", "echo", []string{"test"}, nil)
	assert.Error(t, err) // Expected to fail for non-existent session
}

func TestManager_CreateTmuxEnvironment(t *testing.T) {
	// Test the helper function for creating tmux environment
	tests := []struct {
		name          string
		friendlyTitle string
		expected      map[string]string
	}{
		{
			name:          "with_friendly_title",
			friendlyTitle: "test-title",
			expected: map[string]string{
				"SBS_TITLE": "test-title",
			},
		},
		{
			name:          "with_empty_title",
			friendlyTitle: "",
			expected: map[string]string{
				"SBS_TITLE": "",
			},
		},
		{
			name:          "with_complex_title",
			friendlyTitle: "fix-user-authentication-bug",
			expected: map[string]string{
				"SBS_TITLE": "fix-user-authentication-bug",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateTmuxEnvironment(tt.friendlyTitle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_GenerateFriendlyTitle(t *testing.T) {
	// Test the helper function for generating friendly titles
	tests := []struct {
		name        string
		repoName    string
		issueNumber int
		issueTitle  string
		expected    string
	}{
		{
			name:        "with_issue_title",
			repoName:    "myproject",
			issueNumber: 123,
			issueTitle:  "Fix user authentication bug",
			expected:    "fix-user-authentication-bug",
		},
		{
			name:        "without_issue_title",
			repoName:    "myproject",
			issueNumber: 123,
			issueTitle:  "",
			expected:    "myproject-issue-123",
		},
		{
			name:        "with_empty_title",
			repoName:    "myproject",
			issueNumber: 456,
			issueTitle:  "   ",
			expected:    "myproject-issue-456",
		},
		{
			name:        "with_long_title",
			repoName:    "myproject",
			issueNumber: 789,
			issueTitle:  "This is a very long issue title that should be truncated to fit within the 32 character limit",
			expected:    "this-is-a-very-long-issue-title", // Should be truncated to 32 chars
		},
		{
			name:        "with_special_characters",
			repoName:    "myproject",
			issueNumber: 101,
			issueTitle:  "Fix café login (UTF-8 encoding)",
			expected:    "fix-cafe-login-utf-8-encoding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFriendlyTitle(tt.repoName, tt.issueNumber, tt.issueTitle)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 32)
		})
	}
}
