package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// WithoutSandbox temporarily removes sandbox from PATH for testing
func WithoutSandbox(t *testing.T, testFunc func()) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Create PATH without sandbox by using limited directories
	cleanPath := "/usr/bin:/bin"
	os.Setenv("PATH", cleanPath)

	testFunc()
}

// WithMockSandbox replaces sandbox with a mock implementation
func WithMockSandbox(t *testing.T, mock *MockSandboxManager, testFunc func()) {
	// This is a simplified version - in a real implementation,
	// you might use dependency injection or interface replacement
	testFunc()
}

// WithLimitedPath sets PATH to a limited set of directories
func WithLimitedPath(t *testing.T, pathDirs []string, testFunc func()) {
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	newPath := strings.Join(pathDirs, ":")
	os.Setenv("PATH", newPath)

	testFunc()
}

// WithTempDir creates a temporary directory for testing
func WithTempDir(t *testing.T, testFunc func(tempDir string)) {
	tempDir, err := os.MkdirTemp("", "sbs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFunc(tempDir)
}

// WithMockConfig creates a temporary config for testing
func WithMockConfig(t *testing.T, configContent string, testFunc func(configPath string)) {
	WithTempDir(t, func(tempDir string) {
		configDir := filepath.Join(tempDir, ".config", "sbs")
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		configPath := filepath.Join(configDir, "config.json")
		err = os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Temporarily change HOME to use our test config
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)
		os.Setenv("HOME", tempDir)

		testFunc(configPath)
	})
}

// WithCleanEnvironment provides a clean environment for testing
func WithCleanEnvironment(t *testing.T, testFunc func()) {
	// Save original environment
	originalPath := os.Getenv("PATH")
	originalHome := os.Getenv("HOME")

	defer func() {
		os.Setenv("PATH", originalPath)
		os.Setenv("HOME", originalHome)
	}()

	// Create temporary home directory
	WithTempDir(t, func(tempHome string) {
		os.Setenv("HOME", tempHome)
		testFunc()
	})
}

// CreateMockCommand creates a mock command in a temporary directory and adds it to PATH
func CreateMockCommand(t *testing.T, commandName string, exitCode int, output string) func() {
	tempDir, err := os.MkdirTemp("", "mock-cmd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	mockScript := filepath.Join(tempDir, commandName)
	scriptContent := "#!/bin/bash\n"

	if output != "" {
		scriptContent += "echo \"" + output + "\"\n"
	}

	scriptContent += fmt.Sprintf("exit %d\n", exitCode)

	if err := os.WriteFile(mockScript, []byte(scriptContent), 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create mock script: %v", err)
	}

	originalPath := os.Getenv("PATH")
	newPath := tempDir + ":" + originalPath
	os.Setenv("PATH", newPath)

	return func() {
		os.Setenv("PATH", originalPath)
		os.RemoveAll(tempDir)
	}
}

// SkipIfSandboxMissing skips test if sandbox is not available
func SkipIfSandboxMissing(t *testing.T) {
	if err := CheckSandboxAvailable(); err != nil {
		t.Skipf("sandbox not available: %v", err)
	}
}

// SkipIfSandboxPresent skips test if sandbox is available (for testing missing scenarios)
func SkipIfSandboxPresent(t *testing.T) {
	if err := CheckSandboxAvailable(); err == nil {
		t.Skip("sandbox is available, skipping test for missing sandbox scenario")
	}
}

// CheckSandboxAvailable checks if sandbox command is available
func CheckSandboxAvailable() error {
	// Import the actual CheckSandboxInstalled function would create circular import
	// So we implement a simple check here
	cmd := exec.Command("sandbox", "--help")
	return cmd.Run()
}

// AssertNoAlternativeRuntimeMentions verifies error messages don't mention alternative runtimes
func AssertNoAlternativeRuntimeMentions(t *testing.T, errorMsg string) {
	if errorMsg == "" {
		return
	}

	lowerMsg := strings.ToLower(errorMsg)

	prohibitedTerms := []string{
		"podman",
		"docker",
		"containerd",
		"runc",
		"fallback",
		"alternative",
		"instead",
	}

	for _, term := range prohibitedTerms {
		if strings.Contains(lowerMsg, term) {
			t.Errorf("Error message should not mention '%s': %s", term, errorMsg)
		}
	}
}

// AssertErrorActionable verifies error messages are actionable
func AssertErrorActionable(t *testing.T, errorMsg string) {
	if errorMsg == "" {
		t.Error("Expected actionable error message, got empty string")
		return
	}

	// Should contain guidance words
	guidanceWords := []string{
		"please",
		"ensure",
		"install",
		"check",
		"verify",
	}

	hasGuidance := false
	lowerMsg := strings.ToLower(errorMsg)
	for _, word := range guidanceWords {
		if strings.Contains(lowerMsg, word) {
			hasGuidance = true
			break
		}
	}

	if !hasGuidance {
		t.Errorf("Error message should be actionable and contain guidance: %s", errorMsg)
	}
}

// AssertErrorContains verifies error contains expected substrings
func AssertErrorContains(t *testing.T, err error, expectedSubstrings []string) {
	if err == nil {
		t.Error("Expected error, got nil")
		return
	}

	errorMsg := err.Error()
	for _, substring := range expectedSubstrings {
		if !strings.Contains(errorMsg, substring) {
			t.Errorf("Error should contain '%s': %s", substring, errorMsg)
		}
	}
}

// MockEnvironmentScenario defines a test environment scenario
type MockEnvironmentScenario struct {
	Name            string
	SetupFunc       func(*testing.T)
	CleanupFunc     func(*testing.T)
	SandboxBehavior string // "missing", "failing", "working"
}

// RunWithMockEnvironment runs a test with a specific mock environment
func RunWithMockEnvironment(t *testing.T, scenario MockEnvironmentScenario, testFunc func()) {
	if scenario.SetupFunc != nil {
		scenario.SetupFunc(t)
	}

	if scenario.CleanupFunc != nil {
		defer scenario.CleanupFunc(t)
	}

	testFunc()
}

// Common mock environment scenarios
var (
	SandboxMissingScenario = MockEnvironmentScenario{
		Name: "sandbox_missing",
		SetupFunc: func(t *testing.T) {
			WithoutSandbox(t, func() {})
		},
		SandboxBehavior: "missing",
	}

	SandboxFailingScenario = MockEnvironmentScenario{
		Name: "sandbox_failing",
		SetupFunc: func(t *testing.T) {
			CreateMockCommand(t, "sandbox", 1, "sandbox error")
		},
		SandboxBehavior: "failing",
	}

	SandboxWorkingScenario = MockEnvironmentScenario{
		Name: "sandbox_working",
		SetupFunc: func(t *testing.T) {
			CreateMockCommand(t, "sandbox", 0, "sandbox help")
		},
		SandboxBehavior: "working",
	}
)
