package status

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sbs/pkg/config"
)

func TestStatusTracking_Integration(t *testing.T) {
	// Create a temporary directory structure that mimics a real worktree
	tmpDir := t.TempDir()
	worktreePath := filepath.Join(tmpDir, "issue-123")
	sbsDir := filepath.Join(worktreePath, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	// Create mock managers
	mockTmux := &MockTmuxManager{}
	mockSandbox := &MockSandboxManager{}
	detector := NewDetector(mockTmux, mockSandbox)

	// Test 1: Active session (no stop file, tmux session exists)
	t.Run("active session workflow", func(t *testing.T) {
		mockTmux.SetSessionExists("sbs-123", true)

		session := config.SessionMetadata{
			IssueNumber:  123,
			WorktreePath: worktreePath,
			TmuxSession:  "sbs-123",
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "active", status.Status)
		assert.Equal(t, "now", status.TimeDelta)
		assert.Nil(t, status.LastChange)
	})

	// Test 2: Stopped session (stop file exists, no tmux session)
	t.Run("stopped session workflow", func(t *testing.T) {
		mockTmux.SetSessionExists("sbs-123", false)

		// Create a stop.json file from Claude Code hooks
		stopFile := filepath.Join(sbsDir, "stop.json")
		stopTime := time.Now().Add(-5 * time.Minute)
		stopData := map[string]interface{}{
			"claude_code_hook": map[string]interface{}{
				"timestamp":   stopTime.Format(time.RFC3339),
				"environment": "sandbox",
			},
			"hook_data": map[string]interface{}{
				"tool_executions": 10,
				"last_tool":       "Edit",
			},
		}

		data, err := json.Marshal(stopData)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(stopFile, data, 0644))

		session := config.SessionMetadata{
			IssueNumber:  123,
			WorktreePath: worktreePath,
			TmuxSession:  "sbs-123",
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "stopped", status.Status)
		assert.Equal(t, "5m ago", status.TimeDelta)
		assert.NotNil(t, status.LastChange)
		assert.True(t, stopTime.Sub(*status.LastChange) < time.Minute) // Should be approximately equal
	})

	// Test 3: Stale session (no stop file, no tmux session)
	t.Run("stale session workflow", func(t *testing.T) {
		// Remove stop file
		stopFile := filepath.Join(sbsDir, "stop.json")
		os.Remove(stopFile)

		mockTmux.SetSessionExists("sbs-123", false)

		lastActivity := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
		session := config.SessionMetadata{
			IssueNumber:  123,
			WorktreePath: worktreePath,
			TmuxSession:  "sbs-123",
			LastActivity: lastActivity,
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "stale", status.Status)
		assert.Equal(t, "2h ago", status.TimeDelta)
		assert.NotNil(t, status.LastChange)
	})

	// Test 4: Unknown status (corrupted stop file)
	t.Run("unknown status workflow", func(t *testing.T) {
		mockTmux.SetSessionExists("sbs-123", false)

		// Create a corrupted stop.json file
		stopFile := filepath.Join(sbsDir, "stop.json")
		require.NoError(t, os.WriteFile(stopFile, []byte("invalid json"), 0644))

		session := config.SessionMetadata{
			IssueNumber:  123,
			WorktreePath: worktreePath,
			TmuxSession:  "sbs-123",
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "unknown", status.Status)
		assert.Equal(t, "unknown", status.TimeDelta)
		assert.Nil(t, status.LastChange)
	})

	// Test 5: Sandbox-based stop file detection
	t.Run("sandbox stop file workflow", func(t *testing.T) {
		mockTmux.SetSessionExists("sbs-123", false)

		// Set up sandbox with stop.json file
		sandboxName := "sbs-myrepo-123"
		mockSandbox.SetSandboxExists(sandboxName, true)

		stopTime := time.Now().Add(-10 * time.Minute)
		stopData := map[string]interface{}{
			"claude_code_hook": map[string]interface{}{
				"timestamp":   stopTime.Format(time.RFC3339),
				"environment": "sandbox",
			},
		}
		data, err := json.Marshal(stopData)
		require.NoError(t, err)
		mockSandbox.SetFileContent(sandboxName, ".sbs/stop.json", data)

		session := config.SessionMetadata{
			IssueNumber:    123,
			WorktreePath:   worktreePath,
			TmuxSession:    "sbs-123",
			SandboxName:    sandboxName,
			RepositoryName: "myrepo",
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "stopped", status.Status)
		assert.Equal(t, "10m ago", status.TimeDelta)
		assert.NotNil(t, status.LastChange)
	})

	// Test 6: Sandbox doesn't exist - fallback to direct file access
	t.Run("sandbox not found fallback workflow", func(t *testing.T) {
		mockTmux.SetSessionExists("sbs-123", false)

		// Sandbox doesn't exist
		sandboxName := "sbs-myrepo-456"
		mockSandbox.SetSandboxExists(sandboxName, false)

		// But direct file exists
		stopFile := filepath.Join(sbsDir, "stop.json")
		stopTime := time.Now().Add(-15 * time.Minute)
		stopData := map[string]interface{}{
			"claude_code_hook": map[string]interface{}{
				"timestamp":   stopTime.Format(time.RFC3339),
				"environment": "sandbox",
			},
		}
		data, err := json.Marshal(stopData)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(stopFile, data, 0644))

		session := config.SessionMetadata{
			IssueNumber:    456,
			WorktreePath:   worktreePath,
			TmuxSession:    "sbs-456",
			SandboxName:    sandboxName,
			RepositoryName: "myrepo",
		}

		status := detector.DetectSessionStatus(session)
		assert.Equal(t, "stopped", status.Status)
		assert.Equal(t, "15m ago", status.TimeDelta)
		assert.NotNil(t, status.LastChange)
	})
}

func TestStatusTracking_TimeFormatting_Integration(t *testing.T) {
	formatter := NewTimeFormatter()
	now := time.Now()

	testCases := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "just now",
			timestamp: now.Add(-30 * time.Second),
			expected:  "now",
		},
		{
			name:      "minutes ago",
			timestamp: now.Add(-15 * time.Minute),
			expected:  "15m ago",
		},
		{
			name:      "hours ago",
			timestamp: now.Add(-3 * time.Hour),
			expected:  "3h ago",
		},
		{
			name:      "days ago",
			timestamp: now.Add(-2 * 24 * time.Hour),
			expected:  "2d ago",
		},
		{
			name:      "weeks ago",
			timestamp: now.Add(-10 * 24 * time.Hour),
			expected:  "1w ago",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.FormatTimeDelta(tc.timestamp, now)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestStatusTracking_ConfigurationIntegration(t *testing.T) {
	// Test that the configuration values are properly integrated
	cfg := config.DefaultConfig()

	// Verify default status tracking configuration
	assert.True(t, cfg.StatusTracking, "Status tracking should be enabled by default")
	assert.Equal(t, 60, cfg.StatusRefreshIntervalSecs, "Default refresh interval should be 60 seconds")
	assert.Equal(t, 1048576, cfg.StatusMaxFileSizeBytes, "Default max file size should be 1MB")
	assert.Equal(t, 5, cfg.StatusTimeoutSeconds, "Default timeout should be 5 seconds")

	// Test configuration validation - we'll just test that defaults are reasonable
	assert.True(t, cfg.StatusRefreshIntervalSecs >= 5 && cfg.StatusRefreshIntervalSecs <= 600)
	assert.True(t, cfg.StatusMaxFileSizeBytes >= 1024 && cfg.StatusMaxFileSizeBytes <= 10*1024*1024)
	assert.True(t, cfg.StatusTimeoutSeconds >= 1 && cfg.StatusTimeoutSeconds <= 30)
}
