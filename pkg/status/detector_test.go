package status

import (
	"encoding/json"
	"fmt"
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
			tmuxSession:    "sbs-123",
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

			// Create mock managers
			mockTmux := &MockTmuxManager{}
			mockSandbox := &MockSandboxManager{}
			if tt.tmuxSession != "" {
				mockTmux.SetSessionExists(tt.tmuxSession, true)
			}

			detector := NewDetector(mockTmux, mockSandbox)
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

			detector := NewDetector(&MockTmuxManager{}, &MockSandboxManager{})
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

	detector := NewDetector(&MockTmuxManager{}, &MockSandboxManager{})
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
	mockTmux.SetSessionExists("sbs-123", true)
	detector := NewDetector(mockTmux, &MockSandboxManager{})
	session := config.SessionMetadata{
		IssueNumber:  123,
		WorktreePath: worktreePath,
		TmuxSession:  "sbs-123",
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

	detector := NewDetector(&MockTmuxManager{}, &MockSandboxManager{})
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

// MockSandboxManager for testing
type MockSandboxManager struct {
	sandboxFiles  map[string]map[string][]byte // sandboxName -> filePath -> fileContent
	sandboxExists map[string]bool              // sandboxName -> exists
}

func (m *MockSandboxManager) ReadFileFromSandbox(sandboxName, filePath string) ([]byte, error) {
	if m.sandboxFiles == nil {
		return nil, fmt.Errorf("sandbox %s does not exist", sandboxName)
	}

	if !m.sandboxExists[sandboxName] {
		return nil, fmt.Errorf("sandbox %s does not exist", sandboxName)
	}

	files, ok := m.sandboxFiles[sandboxName]
	if !ok {
		return nil, fmt.Errorf("file %s not found in sandbox %s", filePath, sandboxName)
	}

	content, ok := files[filePath]
	if !ok {
		return nil, fmt.Errorf("file %s not found in sandbox %s", filePath, sandboxName)
	}

	return content, nil
}

func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
	if m.sandboxExists == nil {
		return false, nil
	}
	return m.sandboxExists[sandboxName], nil
}

func (m *MockSandboxManager) SetSandboxExists(sandboxName string, exists bool) {
	if m.sandboxExists == nil {
		m.sandboxExists = make(map[string]bool)
	}
	m.sandboxExists[sandboxName] = exists
}

func (m *MockSandboxManager) SetFileContent(sandboxName, filePath string, content []byte) {
	if m.sandboxFiles == nil {
		m.sandboxFiles = make(map[string]map[string][]byte)
	}
	if m.sandboxFiles[sandboxName] == nil {
		m.sandboxFiles[sandboxName] = make(map[string][]byte)
	}
	m.sandboxFiles[sandboxName][filePath] = content
}
