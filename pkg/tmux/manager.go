package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"sbs/pkg/cmdlog"
)

type Session struct {
	Name         string
	WorkingDir   string
	IssueNumber  int
	Created      time.Time
	LastActivity time.Time
	Status       string // "active", "stopped"
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) CreateSession(issueNumber int, workingDir, sessionName string, env ...map[string]string) (*Session, error) {
	// sessionName is now provided by the caller (repository-aware)

	// Check if session already exists
	exists, err := m.SessionExists(sessionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if session exists: %w", err)
	}

	if exists {
		// Session exists, update its working directory
		if err := m.setWorkingDirectory(sessionName, workingDir); err != nil {
			return nil, fmt.Errorf("failed to update session working directory: %w", err)
		}

		// Set environment variables if provided
		if len(env) > 0 && env[0] != nil {
			if err := m.setEnvironmentVariables(sessionName, env[0]); err != nil {
				return nil, fmt.Errorf("failed to set environment variables: %w", err)
			}
		}

		return &Session{
			Name:         sessionName,
			WorkingDir:   workingDir,
			IssueNumber:  issueNumber,
			Created:      time.Now(), // We don't know actual creation time
			LastActivity: time.Now(),
			Status:       "active",
		}, nil
	}

	// Create new detached session with environment variables
	args := []string{"new-session", "-d", "-s", sessionName, "-c", workingDir}
	if err := m.runTmuxCommandWithEnv(args, env...); err != nil {
		return nil, fmt.Errorf("failed to create tmux session %s: %w", sessionName, err)
	}

	// Set environment variables in the session after creation
	if len(env) > 0 && env[0] != nil {
		if err := m.setEnvironmentVariables(sessionName, env[0]); err != nil {
			return nil, fmt.Errorf("failed to set environment variables: %w", err)
		}
	}

	now := time.Now()
	return &Session{
		Name:         sessionName,
		WorkingDir:   workingDir,
		IssueNumber:  issueNumber,
		Created:      now,
		LastActivity: now,
		Status:       "active",
	}, nil
}

func (m *Manager) SessionExists(sessionName string) (bool, error) {
	args := []string{"has-session", "-t", sessionName}
	err := m.runTmuxCommandRun(args)
	if err != nil {
		// Exit code 1 means session doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("error checking session existence: %w", err)
	}
	return true, nil
}

func (m *Manager) AttachToSession(sessionName string, env ...map[string]string) error {
	// Find tmux executable path
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux command not found: %w", err)
	}

	// Set environment variables in the session before attaching
	if len(env) > 0 && env[0] != nil {
		if err := m.setEnvironmentVariables(sessionName, env[0]); err != nil {
			return fmt.Errorf("failed to set environment variables: %w", err)
		}
	}

	// Replace current process with tmux attach
	args := []string{"tmux", "attach-session", "-t", sessionName}
	execEnv := os.Environ()

	// Add environment variables to the exec environment
	if len(env) > 0 && env[0] != nil {
		for key, value := range env[0] {
			execEnv = append(execEnv, fmt.Sprintf("%s=%s", key, value))
		}
	}

	err = syscall.Exec(tmuxPath, args, execEnv)
	if err != nil {
		return fmt.Errorf("failed to exec tmux attach: %w", err)
	}

	// This line should never be reached if exec succeeds
	return nil
}

func (m *Manager) KillSession(sessionName string) error {
	args := []string{"kill-session", "-t", sessionName}
	if err := m.runTmuxCommandRun(args); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", sessionName, err)
	}
	return nil
}

func (m *Manager) ListSessions() ([]*Session, error) {
	args := []string{"list-sessions", "-F", "#{session_name}|#{session_created}|#{session_last_attached}"}
	output, err := m.runTmuxCommand(args)
	if err != nil {
		// No sessions exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to list tmux sessions: %w", err)
	}

	var sessions []*Session
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}

		sessionName := parts[0]

		// Only process sbs sessions
		if !strings.HasPrefix(sessionName, "sbs-") {
			continue
		}

		// Extract issue number from different formats:
		// sbs-123 (legacy)
		// sbs-repo-123 (new format)
		issueNumber := m.extractIssueNumber(sessionName)
		if issueNumber == 0 {
			continue
		}

		// Parse timestamps
		created, _ := strconv.ParseInt(parts[1], 10, 64)
		lastActivity, _ := strconv.ParseInt(parts[2], 10, 64)

		// Get working directory
		workingDir, _ := m.getSessionWorkingDir(sessionName)

		sessions = append(sessions, &Session{
			Name:         sessionName,
			WorkingDir:   workingDir,
			IssueNumber:  issueNumber,
			Created:      time.Unix(created, 0),
			LastActivity: time.Unix(lastActivity, 0),
			Status:       "active",
		})
	}

	return sessions, nil
}

func (m *Manager) StartWorkIssue(sessionName string, issueNumber int, workIssueScript string, env ...map[string]string) error {
	// Set environment variables in the session before executing command
	if len(env) > 0 && env[0] != nil {
		if err := m.setEnvironmentVariables(sessionName, env[0]); err != nil {
			return fmt.Errorf("failed to set environment variables: %w", err)
		}
	}

	// Send command to run work-issue script in the session
	command := fmt.Sprintf("%s %d", workIssueScript, issueNumber)
	args := []string{"send-keys", "-t", sessionName, command, "Enter"}

	if err := m.runTmuxCommandRun(args); err != nil {
		return fmt.Errorf("failed to start work-issue in session %s: %w", sessionName, err)
	}

	return nil
}

// ExecuteCommand executes a flexible command in the tmux session
func (m *Manager) ExecuteCommand(sessionName, command string, args []string, env ...map[string]string) error {
	var envVars map[string]string
	if len(env) > 0 {
		envVars = env[0]
	}
	return m.ExecuteCommandWithSubstitution(sessionName, command, args, nil, envVars)
}

// ExecuteCommandWithSubstitution executes a command with parameter substitution
func (m *Manager) ExecuteCommandWithSubstitution(sessionName, command string, args []string, substitutions map[string]string, env ...map[string]string) error {
	// Set environment variables in the session before executing command
	if len(env) > 0 && env[0] != nil {
		if err := m.setEnvironmentVariables(sessionName, env[0]); err != nil {
			return fmt.Errorf("failed to set environment variables: %w", err)
		}
	}

	// Build the full command string
	var fullCommand string
	if len(args) > 0 {
		// Apply substitutions to args if provided
		processedArgs := make([]string, len(args))
		for i, arg := range args {
			processedArgs[i] = m.substituteParameters(arg, substitutions)
		}

		// Use command + processed args
		cmdParts := append([]string{command}, processedArgs...)
		fullCommand = strings.Join(cmdParts, " ")
	} else {
		fullCommand = command
	}

	// Send command to the session
	tmuxArgs := []string{"send-keys", "-t", sessionName, fullCommand, "Enter"}

	if err := m.runTmuxCommandRun(tmuxArgs); err != nil {
		return fmt.Errorf("failed to execute command in session %s: %w", sessionName, err)
	}

	return nil
}

// substituteParameters replaces parameter placeholders in a string
func (m *Manager) substituteParameters(input string, substitutions map[string]string) string {
	if substitutions == nil {
		return input
	}

	result := input
	for placeholder, value := range substitutions {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func (m *Manager) setWorkingDirectory(sessionName, workingDir string) error {
	// Send cd command to the session
	args := []string{"send-keys", "-t", sessionName, fmt.Sprintf("cd %s", workingDir), "Enter"}
	return m.runTmuxCommandRun(args)
}

func (m *Manager) getSessionWorkingDir(sessionName string) (string, error) {
	args := []string{"display-message", "-t", sessionName, "-p", "#{pane_current_path}"}
	output, err := m.runTmuxCommand(args)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractIssueNumber extracts issue number from session names
// Since we now use namespaced format, this always returns 0 for generic handling
func (m *Manager) extractIssueNumber(sessionName string) int {
	// With namespaced work items, we don't extract numeric issue numbers anymore
	// All work items are handled generically
	return 0
}

// setEnvironmentVariables sets environment variables in a tmux session
func (m *Manager) setEnvironmentVariables(sessionName string, env map[string]string) error {
	if env == nil || len(env) == 0 {
		return nil
	}

	for key, value := range env {
		args := []string{"set-environment", "-t", sessionName, key, value}
		if err := m.runTmuxCommandRun(args); err != nil {
			return fmt.Errorf("failed to set environment variable %s=%s: %w", key, value, err)
		}
	}

	return nil
}

// formatEnvironmentVariables formats environment variables for testing and display
func (m *Manager) formatEnvironmentVariables(env map[string]string) []string {
	if env == nil || len(env) == 0 {
		return []string{}
	}

	var result []string
	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}

// CreateTmuxEnvironment creates a map with SBS_TITLE environment variable
func CreateTmuxEnvironment(friendlyTitle string) map[string]string {
	return map[string]string{
		"SBS_TITLE": friendlyTitle,
	}
}

// GenerateFriendlyTitle generates a sandbox-friendly title from issue information
func GenerateFriendlyTitle(repoName string, issueNumber int, issueTitle string) string {
	// Import repo manager to use SanitizeName
	if strings.TrimSpace(issueTitle) == "" {
		// Fallback format when title is unavailable
		return fmt.Sprintf("%s-issue-%d", sanitizeRepoName(repoName), issueNumber)
	}

	// Use sanitization logic similar to repo manager
	return sanitizeTitle(issueTitle, 32)
}

// sanitizeRepoName sanitizes repository name for use in fallback title
func sanitizeRepoName(name string) string {
	if name == "" {
		return "repo"
	}
	return sanitizeTitle(name, 20) // Shorter limit for repo name part
}

// sanitizeTitle sanitizes a title string for sandbox use
func sanitizeTitle(title string, maxLength int) string {
	if title == "" {
		return ""
	}

	// Normalize and remove special characters
	result := normalizeTitle(title)

	// Replace non-alphanumeric with hyphens
	result = strings.ToLower(result)
	result = strings.ReplaceAll(result, " ", "-")

	// Remove non-alphanumeric except hyphens
	var clean strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			clean.WriteRune(r)
		} else {
			clean.WriteRune('-')
		}
	}
	result = clean.String()

	// Remove duplicate hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim hyphens from ends
	result = strings.Trim(result, "-")

	// Apply length limit with word boundary consideration
	if len(result) > maxLength {
		truncated := result[:maxLength]
		// Try to find last hyphen for clean break
		if lastHyphen := strings.LastIndex(truncated, "-"); lastHyphen > 0 && lastHyphen < maxLength-1 {
			result = truncated[:lastHyphen]
		} else {
			result = truncated
		}
		result = strings.TrimSuffix(result, "-")
	}

	return result
}

// normalizeTitle normalizes Unicode characters in title
func normalizeTitle(title string) string {
	// Simple ASCII replacements for common characters
	replacements := map[rune]string{
		'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
		'á': "a", 'à': "a", 'â': "a", 'ä': "a",
		'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
		'ó': "o", 'ò': "o", 'ô': "o", 'ö': "o",
		'ú': "u", 'ù': "u", 'û': "u", 'ü': "u",
		'ç': "c", 'ñ': "n",
		'É': "E", 'È': "E", 'Ê': "E", 'Ë': "E",
		'Á': "A", 'À': "A", 'Â': "A", 'Ä': "A",
		'Í': "I", 'Ì': "I", 'Î': "I", 'Ï': "I",
		'Ó': "O", 'Ò': "O", 'Ô': "O", 'Ö': "O",
		'Ú': "U", 'Ù': "U", 'Û': "U", 'Ü': "U",
		'Ç': "C", 'Ñ': "N",
	}

	var result strings.Builder
	for _, r := range title {
		if replacement, exists := replacements[r]; exists {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// runTmuxCommand executes a tmux command with logging and returns output
func (m *Manager) runTmuxCommand(args []string) ([]byte, error) {
	ctx := cmdlog.LogCommandGlobal("tmux", args, cmdlog.GetCaller())

	cmd := exec.Command("tmux", args...)
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

// runTmuxCommandRun executes a tmux command with logging without capturing output
func (m *Manager) runTmuxCommandRun(args []string) error {
	ctx := cmdlog.LogCommandGlobal("tmux", args, cmdlog.GetCaller())

	cmd := exec.Command("tmux", args...)
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

// runTmuxCommandWithEnv executes a tmux command with custom environment variables
func (m *Manager) runTmuxCommandWithEnv(args []string, env ...map[string]string) error {
	ctx := cmdlog.LogCommandGlobal("tmux", args, cmdlog.GetCaller())

	cmd := exec.Command("tmux", args...)

	// Set environment variables for the tmux command
	if len(env) > 0 && env[0] != nil {
		cmdEnv := os.Environ()
		for key, value := range env[0] {
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = cmdEnv
	}

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

// getExitCode extracts exit code from command error
func getExitCode(cmd *exec.Cmd) int {
	if cmd.ProcessState == nil {
		return -1
	}
	return cmd.ProcessState.ExitCode()
}
