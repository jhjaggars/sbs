package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
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

func (m *Manager) CreateSession(issueNumber int, workingDir string) (*Session, error) {
	sessionName := fmt.Sprintf("work-issue-%d", issueNumber)
	
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
		
		return &Session{
			Name:         sessionName,
			WorkingDir:   workingDir,
			IssueNumber:  issueNumber,
			Created:      time.Now(), // We don't know actual creation time
			LastActivity: time.Now(),
			Status:       "active",
		}, nil
	}
	
	// Create new detached session
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create tmux session %s: %w", sessionName, err)
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
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	if err != nil {
		// Exit code 1 means session doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, fmt.Errorf("error checking session existence: %w", err)
	}
	return true, nil
}

func (m *Manager) AttachToSession(sessionName string) error {
	// Find tmux executable path
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux command not found: %w", err)
	}
	
	// Replace current process with tmux attach
	args := []string{"tmux", "attach-session", "-t", sessionName}
	env := os.Environ()
	
	err = syscall.Exec(tmuxPath, args, env)
	if err != nil {
		return fmt.Errorf("failed to exec tmux attach: %w", err)
	}
	
	// This line should never be reached if exec succeeds
	return nil
}

func (m *Manager) KillSession(sessionName string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session %s: %w", sessionName, err)
	}
	return nil
}

func (m *Manager) ListSessions() ([]*Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}|#{session_created}|#{session_last_attached}")
	output, err := cmd.Output()
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
		
		// Only process work-issue sessions
		if !strings.HasPrefix(sessionName, "work-issue-") {
			continue
		}
		
		// Extract issue number
		issueNumStr := strings.TrimPrefix(sessionName, "work-issue-")
		issueNumber, err := strconv.Atoi(issueNumStr)
		if err != nil {
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

func (m *Manager) StartWorkIssue(sessionName string, issueNumber int, workIssueScript string) error {
	// Send command to run work-issue script in the session
	command := fmt.Sprintf("%s %d", workIssueScript, issueNumber)
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, command, "Enter")
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start work-issue in session %s: %w", sessionName, err)
	}
	
	return nil
}

func (m *Manager) setWorkingDirectory(sessionName, workingDir string) error {
	// Send cd command to the session
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, fmt.Sprintf("cd %s", workingDir), "Enter")
	return cmd.Run()
}

func (m *Manager) getSessionWorkingDir(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{pane_current_path}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}