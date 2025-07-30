package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// GetSandboxName returns the expected sandbox name for an issue
func (m *Manager) GetSandboxName(issueNumber int) string {
	return fmt.Sprintf("work-issue-%d", issueNumber)
}

// SandboxExists checks if a sandbox with the given name exists
func (m *Manager) SandboxExists(sandboxName string) (bool, error) {
	cmd := exec.Command("sandbox", "list")
	output, err := cmd.Output()
	if err != nil {
		// If sandbox command fails, assume sandboxes don't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, fmt.Errorf("failed to list sandboxes: %w", err)
	}
	
	// Check if sandbox name appears in output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, sandboxName) {
			return true, nil
		}
	}
	
	return false, nil
}

// DeleteSandbox removes a sandbox with the given name
func (m *Manager) DeleteSandbox(sandboxName string) error {
	// Check if sandbox exists first
	exists, err := m.SandboxExists(sandboxName)
	if err != nil {
		return fmt.Errorf("failed to check if sandbox exists: %w", err)
	}
	
	if !exists {
		// Sandbox doesn't exist, nothing to do
		return nil
	}
	
	cmd := exec.Command("sandbox", "delete", sandboxName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete sandbox %s: %w", sandboxName, err)
	}
	
	return nil
}

// ListSandboxes returns all work-issue sandboxes
func (m *Manager) ListSandboxes() ([]string, error) {
	cmd := exec.Command("sandbox", "list")
	output, err := cmd.Output()
	if err != nil {
		// If sandbox command fails, return empty list
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list sandboxes: %w", err)
	}
	
	var workIssueSandboxes []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "work-issue-") {
			// Extract just the sandbox name (first field)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				workIssueSandboxes = append(workIssueSandboxes, fields[0])
			}
		}
	}
	
	return workIssueSandboxes, nil
}

// CheckSandboxInstalled verifies that the sandbox command is available
func CheckSandboxInstalled() error {
	cmd := exec.Command("sandbox", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sandbox command not found. Please ensure sandbox is installed and in PATH")
	}
	return nil
}