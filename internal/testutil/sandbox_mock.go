package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// MockSandboxManager provides configurable mock behavior for sandbox operations
type MockSandboxManager struct {
	ShouldFail           bool
	FailWithError        error
	SimulateNotInstalled bool
	CommandCalls         [][]string // Track all command calls
	mutex                sync.Mutex

	// Configurable responses
	ListResponse    []string
	ExistsResponse  map[string]bool
	DeleteResponses map[string]error
}

// NewMockSandboxManager creates a new mock sandbox manager
func NewMockSandboxManager() *MockSandboxManager {
	return &MockSandboxManager{
		CommandCalls:    make([][]string, 0),
		ExistsResponse:  make(map[string]bool),
		DeleteResponses: make(map[string]error),
		ListResponse:    make([]string, 0),
	}
}

// CheckSandboxInstalled mocks the sandbox installation check
func (m *MockSandboxManager) CheckSandboxInstalled() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CommandCalls = append(m.CommandCalls, []string{"sandbox", "--help"})

	if m.SimulateNotInstalled {
		return fmt.Errorf("sandbox command not found. Please ensure sandbox is installed and in PATH")
	}
	if m.ShouldFail {
		if m.FailWithError != nil {
			return m.FailWithError
		}
		return fmt.Errorf("mock sandbox failure")
	}
	return nil
}

// ListSandboxes mocks listing sandboxes
func (m *MockSandboxManager) ListSandboxes() ([]string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CommandCalls = append(m.CommandCalls, []string{"sandbox", "list"})

	if m.ShouldFail {
		return nil, m.FailWithError
	}

	return m.ListResponse, nil
}

// SandboxExists mocks sandbox existence check
func (m *MockSandboxManager) SandboxExists(sandboxName string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CommandCalls = append(m.CommandCalls, []string{"sandbox", "list"})

	if m.ShouldFail {
		return false, m.FailWithError
	}

	exists, found := m.ExistsResponse[sandboxName]
	if !found {
		return false, nil // Default to not exists
	}

	return exists, nil
}

// DeleteSandbox mocks sandbox deletion
func (m *MockSandboxManager) DeleteSandbox(sandboxName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// First check if sandbox exists (like real implementation)
	exists, _ := m.ExistsResponse[sandboxName]
	if !exists {
		return nil // Nothing to delete
	}

	m.CommandCalls = append(m.CommandCalls, []string{"sandbox", "delete", sandboxName})

	if err, found := m.DeleteResponses[sandboxName]; found {
		return err
	}

	if m.ShouldFail {
		return m.FailWithError
	}

	return nil
}

// GetSandboxName returns mock sandbox name
func (m *MockSandboxManager) GetSandboxName(issueNumber int) string {
	return fmt.Sprintf("work-issue-%d", issueNumber)
}

// GetRepositorySandboxName returns mock repository-scoped sandbox name
func (m *MockSandboxManager) GetRepositorySandboxName(repoName string, issueNumber int) string {
	return fmt.Sprintf("work-issue-%s-%d", repoName, issueNumber)
}

// GetCommandCalls returns all recorded command calls
func (m *MockSandboxManager) GetCommandCalls() [][]string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return copy to avoid race conditions
	calls := make([][]string, len(m.CommandCalls))
	copy(calls, m.CommandCalls)
	return calls
}

// Reset clears all recorded calls and resets state
func (m *MockSandboxManager) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CommandCalls = make([][]string, 0)
	m.ShouldFail = false
	m.FailWithError = nil
	m.SimulateNotInstalled = false
	m.ExistsResponse = make(map[string]bool)
	m.DeleteResponses = make(map[string]error)
	m.ListResponse = make([]string, 0)
}

// SetSandboxExists configures mock response for sandbox existence
func (m *MockSandboxManager) SetSandboxExists(sandboxName string, exists bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ExistsResponse[sandboxName] = exists
}

// SetDeleteResponse configures mock response for sandbox deletion
func (m *MockSandboxManager) SetDeleteResponse(sandboxName string, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.DeleteResponses[sandboxName] = err
}

// SetListResponse configures mock response for listing sandboxes
func (m *MockSandboxManager) SetListResponse(sandboxes []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ListResponse = make([]string, len(sandboxes))
	copy(m.ListResponse, sandboxes)
}

// SandboxEnvironment provides utilities for managing sandbox test environments
type SandboxEnvironment struct {
	originalPath string
	tempDirs     []string
	mockScripts  map[string]string
	mutex        sync.Mutex
}

// NewSandboxEnvironment creates a new sandbox test environment
func NewSandboxEnvironment() *SandboxEnvironment {
	return &SandboxEnvironment{
		originalPath: os.Getenv("PATH"),
		tempDirs:     make([]string, 0),
		mockScripts:  make(map[string]string),
	}
}

// WithoutSandbox removes sandbox from PATH for testing
func (e *SandboxEnvironment) WithoutSandbox() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Set PATH to exclude potential sandbox locations
	os.Setenv("PATH", "/usr/bin:/bin")
}

// WithMockSandbox creates a mock sandbox command with specified behavior
func (e *SandboxEnvironment) WithMockSandbox(exitCode int, output string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	tempDir, err := os.MkdirTemp("", "mock-sandbox-env")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	e.tempDirs = append(e.tempDirs, tempDir)

	mockScript := filepath.Join(tempDir, "sandbox")
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo "%s"
exit %d
`, output, exitCode)

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create mock script: %w", err)
	}

	e.mockScripts["sandbox"] = mockScript

	// Prepend temp directory to PATH
	newPath := tempDir + ":" + os.Getenv("PATH")
	os.Setenv("PATH", newPath)

	return nil
}

// WithMockSandboxHelp creates a mock sandbox that responds properly to --help
func (e *SandboxEnvironment) WithMockSandboxHelp() error {
	return e.WithMockSandbox(0, "Mock sandbox help\nUsage: sandbox [command]")
}

// WithFailingSandbox creates a mock sandbox that always fails
func (e *SandboxEnvironment) WithFailingSandbox(exitCode int, errorMsg string) error {
	return e.WithMockSandbox(exitCode, errorMsg)
}

// WithHangingSandbox creates a mock sandbox that hangs (for timeout testing)
func (e *SandboxEnvironment) WithHangingSandbox() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	tempDir, err := os.MkdirTemp("", "hanging-sandbox-env")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	e.tempDirs = append(e.tempDirs, tempDir)

	mockScript := filepath.Join(tempDir, "sandbox")
	scriptContent := `#!/bin/bash
# Hang for testing timeout scenarios
sleep 30
exit 0
`

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		return fmt.Errorf("failed to create hanging script: %w", err)
	}

	e.mockScripts["sandbox"] = mockScript

	// Prepend temp directory to PATH
	newPath := tempDir + ":" + os.Getenv("PATH")
	os.Setenv("PATH", newPath)

	return nil
}

// Cleanup restores the original environment
func (e *SandboxEnvironment) Cleanup() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Restore original PATH
	os.Setenv("PATH", e.originalPath)

	// Clean up temporary directories
	for _, tempDir := range e.tempDirs {
		os.RemoveAll(tempDir)
	}

	e.tempDirs = make([]string, 0)
	e.mockScripts = make(map[string]string)
}

// GetMockScriptPath returns the path to a mock script
func (e *SandboxEnvironment) GetMockScriptPath(command string) (string, bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	path, exists := e.mockScripts[command]
	return path, exists
}

// TestScenario defines a test scenario with setup, execution, and validation
type TestScenario struct {
	Name        string
	Description string
	Setup       func(*SandboxEnvironment) error
	Test        func() error
	Validate    func(error) error
}

// RunTestScenario executes a test scenario with proper cleanup
func RunTestScenario(scenario TestScenario) error {
	env := NewSandboxEnvironment()
	defer env.Cleanup()

	// Setup
	if scenario.Setup != nil {
		if err := scenario.Setup(env); err != nil {
			return fmt.Errorf("scenario setup failed: %w", err)
		}
	}

	// Execute test
	var testErr error
	if scenario.Test != nil {
		testErr = scenario.Test()
	}

	// Validate results
	if scenario.Validate != nil {
		if err := scenario.Validate(testErr); err != nil {
			return fmt.Errorf("scenario validation failed: %w", err)
		}
	}

	return nil
}

// Common test scenarios
var CommonSandboxScenarios = []TestScenario{
	{
		Name:        "sandbox_not_found",
		Description: "Test behavior when sandbox command is not found",
		Setup: func(env *SandboxEnvironment) error {
			env.WithoutSandbox()
			return nil
		},
		Test: func() error {
			// This would be replaced with actual test function
			cmd := exec.Command("sandbox", "--help")
			return cmd.Run()
		},
		Validate: func(err error) error {
			if err == nil {
				return fmt.Errorf("expected error when sandbox not found")
			}
			if !strings.Contains(err.Error(), "not found") &&
				!strings.Contains(err.Error(), "command not found") {
				return fmt.Errorf("expected 'not found' error, got: %v", err)
			}
			return nil
		},
	},
	{
		Name:        "sandbox_permission_denied",
		Description: "Test behavior when sandbox command has permission issues",
		Setup: func(env *SandboxEnvironment) error {
			return env.WithFailingSandbox(126, "permission denied")
		},
		Test: func() error {
			cmd := exec.Command("sandbox", "--help")
			return cmd.Run()
		},
		Validate: func(err error) error {
			if err == nil {
				return fmt.Errorf("expected permission error")
			}
			return nil
		},
	},
	{
		Name:        "sandbox_working",
		Description: "Test behavior when sandbox command works properly",
		Setup: func(env *SandboxEnvironment) error {
			return env.WithMockSandboxHelp()
		},
		Test: func() error {
			cmd := exec.Command("sandbox", "--help")
			return cmd.Run()
		},
		Validate: func(err error) error {
			if err != nil {
				return fmt.Errorf("expected sandbox to work, got error: %v", err)
			}
			return nil
		},
	},
}
