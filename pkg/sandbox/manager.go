package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"sbs/pkg/cmdlog"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// GetSandboxName returns the expected sandbox name for an issue (legacy method)
func (m *Manager) GetSandboxName(issueNumber int) string {
	return fmt.Sprintf("sbs-%d", issueNumber)
}

// GetRepositorySandboxName returns the repository-scoped sandbox name
func (m *Manager) GetRepositorySandboxName(repoName string, issueNumber int) string {
	return fmt.Sprintf("sbs-%s-%d", repoName, issueNumber)
}

// SandboxExists checks if a sandbox with the given name exists
func (m *Manager) SandboxExists(sandboxName string) (bool, error) {
	output, err := m.runSandboxCommand([]string{"list"})
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

	if err := m.runSandboxCommandRun([]string{"delete", sandboxName, "-y"}); err != nil {
		return fmt.Errorf("failed to delete sandbox %s: %w", sandboxName, err)
	}

	return nil
}

// ListSandboxes returns all sbs sandboxes
func (m *Manager) ListSandboxes() ([]string, error) {
	output, err := m.runSandboxCommand([]string{"list"})
	if err != nil {
		// If sandbox command fails, return empty list
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list sandboxes: %w", err)
	}

	var sbsSandboxes []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "sbs-") {
			// Extract just the sandbox name (first field)
			fields := strings.Fields(line)
			if len(fields) > 0 {
				sbsSandboxes = append(sbsSandboxes, fields[0])
			}
		}
	}

	return sbsSandboxes, nil
}

// CheckSandboxInstalled verifies that the sandbox command is available
func CheckSandboxInstalled() error {
	ctx := cmdlog.LogCommandGlobal("sandbox", []string{"--help"}, cmdlog.GetCaller())

	cmd := exec.Command("sandbox", "--help")
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return fmt.Errorf("sandbox command not found. Please ensure sandbox is installed and in PATH")
	}

	ctx.LogCompletion(true, 0, "", duration)
	return nil
}

// runSandboxCommand executes a sandbox command with logging
func (m *Manager) runSandboxCommand(args []string) ([]byte, error) {
	ctx := cmdlog.LogCommandGlobal("sandbox", args, cmdlog.GetCaller())

	cmd := exec.Command("sandbox", args...)
	start := time.Now()
	output, err := cmd.Output()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return output, err
	}

	ctx.LogCompletion(true, 0, "", duration)
	return output, nil
}

// runSandboxCommandRun executes a sandbox command without capturing output, with logging
func (m *Manager) runSandboxCommandRun(args []string) error {
	ctx := cmdlog.LogCommandGlobal("sandbox", args, cmdlog.GetCaller())

	cmd := exec.Command("sandbox", args...)
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		ctx.LogCompletion(false, getExitCode(cmd), err.Error(), duration)
		return err
	}

	ctx.LogCompletion(true, 0, "", duration)
	return nil
}

// ReadFileFromSandbox reads a file from within a sandbox using 'sandbox --name <name> cat <path>'
func (m *Manager) ReadFileFromSandbox(sandboxName, filePath string) ([]byte, error) {
	// Check if sandbox exists first
	exists, err := m.SandboxExists(sandboxName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if sandbox exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("sandbox %s does not exist", sandboxName)
	}

	// Execute sandbox command to read file
	output, err := m.runSandboxCommand([]string{"--name", sandboxName, "cat", filePath})
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from sandbox %s: %w", filePath, sandboxName, err)
	}

	return output, nil
}

// getExitCode extracts exit code from exec.Cmd
func getExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode()
	}
	return -1
}
