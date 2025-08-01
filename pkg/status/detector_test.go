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

func TestStatusDetector_DetectSessionStatus(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, tmpDir string) string
		hasStopFile    bool
		tmuxSession    string
		expected       SessionStatus
		expectedStatus string
	}{
		{
			name: "active session without stop file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				worktreeDir := filepath.Join(tmpDir, "issue-123")
				require.NoError(t, os.MkdirAll(worktreeDir, 0755))
				return worktreeDir
			},
			hasStopFile:    false,
			tmuxSession:    "work-issue-123",
			expectedStatus: "active",
		},
		{
			name: "stopped session with stop file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				worktreeDir := filepath.Join(tmpDir, "issue-124")
				sbsDir := filepath.Join(worktreeDir, ".sbs")
				require.NoError(t, os.MkdirAll(sbsDir, 0755))

				stopFile := filepath.Join(sbsDir, "stop.json")
				stopData := map[string]interface{}{
					"claude_code_hook": map[string]interface{}{
						"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
					},
				}
				data, _ := json.Marshal(stopData)
				require.NoError(t, os.WriteFile(stopFile, data, 0644))
				return worktreeDir
			},
			hasStopFile:    true,
			tmuxSession:    "",
			expectedStatus: "stopped",
		},
		{
			name: "stale session without tmux and no stop file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				worktreeDir := filepath.Join(tmpDir, "issue-125")
				require.NoError(t, os.MkdirAll(worktreeDir, 0755))
				return worktreeDir
			},
			hasStopFile:    false,
			tmuxSession:    "",
			expectedStatus: "stale",
		},
		{
			name: "unknown status with corrupted stop file",
			setupFunc: func(t *testing.T, tmpDir string) string {
				worktreeDir := filepath.Join(tmpDir, "issue-126")
				sbsDir := filepath.Join(worktreeDir, ".sbs")
				require.NoError(t, os.MkdirAll(sbsDir, 0755))

				stopFile := filepath.Join(sbsDir, "stop.json")
				require.NoError(t, os.WriteFile(stopFile, []byte("invalid json"), 0644))
				return worktreeDir
			},
			hasStopFile:    true,
			tmuxSession:    "",
			expectedStatus: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			worktreePath := tt.setupFunc(t, tmpDir)

			// Create mock tmux manager
			mockTmux := &MockTmuxManager{}
			if tt.tmuxSession != "" {
				mockTmux.SetSessionExists(tt.tmuxSession, true)
			}

			detector := NewDetector(mockTmux)
			session := config.SessionMetadata{
				IssueNumber:  123,
				WorktreePath: worktreePath,
				TmuxSession:  tt.tmuxSession,
			}

			status := detector.DetectSessionStatus(session)

			assert.Equal(t, tt.expectedStatus, status.Status)
			if tt.hasStopFile && tt.expectedStatus != "unknown" {
				assert.NotNil(t, status.LastChange)
			} else if tt.expectedStatus == "unknown" {
				assert.Nil(t, status.LastChange)
			}
		})
	}
}

func TestStatusDetector_ParseStopJsonFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectTime  bool
	}{
		{
			name: "basic claude code hook format",
			content: `{
				"claude_code_hook": {
					"timestamp": "2025-08-01T12:30:45Z",
					"environment": "sandbox"
				}
			}`,
			expectError: false,
			expectTime:  true,
		},
		{
			name: "minimal format with timestamp",
			content: `{
				"timestamp": "2025-08-01T10:15:30Z",
				"status": "stopped"
			}`,
			expectError: false,
			expectTime:  true,
		},
		{
			name: "corrupted json",
			content: `{
				"timestamp": "2025-08-01T10:15:30Z",
				"status": "stopped"
			`,
			expectError: true,
			expectTime:  false,
		},
		{
			name:        "empty file",
			content:     "",
			expectError: true,
			expectTime:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "stop.json")
			require.NoError(t, os.WriteFile(tmpFile, []byte(tt.content), 0644))

			detector := NewDetector(&MockTmuxManager{})
			timestamp, err := detector.ParseStopJsonFile(tmpFile)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectTime {
					assert.False(t, timestamp.IsZero())
				}
			}
		})
	}
}

func TestStatusDetector_CalculateTimeDelta(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
		expectNow bool
	}{
		{
			name:      "30 seconds ago",
			timestamp: now.Add(-30 * time.Second),
			expected:  "now",
			expectNow: true,
		},
		{
			name:      "5 minutes ago",
			timestamp: now.Add(-5 * time.Minute),
			expected:  "5m ago",
		},
		{
			name:      "2 hours ago",
			timestamp: now.Add(-2 * time.Hour),
			expected:  "2h ago",
		},
		{
			name:      "3 days ago",
			timestamp: now.Add(-3 * 24 * time.Hour),
			expected:  "3d ago",
		},
		{
			name:      "2 weeks ago",
			timestamp: now.Add(-14 * 24 * time.Hour),
			expected:  "2w ago",
		},
	}

	detector := NewDetector(&MockTmuxManager{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.CalculateTimeDelta(tt.timestamp)
			if tt.expectNow {
				assert.Equal(t, "now", result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStatusDetector_HandleMissingStopFile(t *testing.T) {
	tmpDir := t.TempDir()
	worktreePath := filepath.Join(tmpDir, "issue-123")
	require.NoError(t, os.MkdirAll(worktreePath, 0755))

	mockTmux := &MockTmuxManager{}
	mockTmux.SetSessionExists("work-issue-123", true)
	detector := NewDetector(mockTmux)
	session := config.SessionMetadata{
		IssueNumber:  123,
		WorktreePath: worktreePath,
		TmuxSession:  "work-issue-123",
	}

	status := detector.DetectSessionStatus(session)

	// Should default to active when no stop file exists and tmux session is present
	assert.Equal(t, "active", status.Status)
	assert.Nil(t, status.LastChange)
}

func TestStatusDetector_HandlePermissionErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test as root user")
	}

	tmpDir := t.TempDir()
	worktreePath := filepath.Join(tmpDir, "issue-123")
	sbsDir := filepath.Join(worktreePath, ".sbs")
	require.NoError(t, os.MkdirAll(sbsDir, 0755))

	stopFile := filepath.Join(sbsDir, "stop.json")
	require.NoError(t, os.WriteFile(stopFile, []byte(`{"timestamp":"2025-08-01T12:00:00Z"}`), 0644))

	// Remove read permissions
	require.NoError(t, os.Chmod(stopFile, 0000))
	defer os.Chmod(stopFile, 0644) // Cleanup

	detector := NewDetector(&MockTmuxManager{})
	session := config.SessionMetadata{
		IssueNumber:  123,
		WorktreePath: worktreePath,
		TmuxSession:  "",
	}

	status := detector.DetectSessionStatus(session)

	// Should handle permission errors gracefully
	assert.Equal(t, "unknown", status.Status)
}

// MockTmuxManager for testing
type MockTmuxManager struct {
	sessions map[string]bool
}

func (m *MockTmuxManager) SessionExists(sessionName string) (bool, error) {
	if m.sessions == nil {
		return false, nil
	}
	exists, ok := m.sessions[sessionName]
	return exists && ok, nil
}

func (m *MockTmuxManager) SetSessionExists(sessionName string, exists bool) {
	if m.sessions == nil {
		m.sessions = make(map[string]bool)
	}
	m.sessions[sessionName] = exists
}
