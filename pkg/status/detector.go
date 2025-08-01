package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sbs/pkg/config"
)

// SessionStatus represents the status of a work session
type SessionStatus struct {
	Status     string     // active, stopped, stale, unknown
	LastChange *time.Time // timestamp when status last changed
	TimeDelta  string     // human-readable time since last change
}

// TmuxManager interface for tmux operations (for dependency injection/testing)
type TmuxManager interface {
	SessionExists(sessionName string) (bool, error)
}

// SandboxManager interface for sandbox operations (for dependency injection/testing)
type SandboxManager interface {
	ReadFileFromSandbox(sandboxName, filePath string) ([]byte, error)
	SandboxExists(sandboxName string) (bool, error)
}

// Detector handles status detection for work sessions
type Detector struct {
	tmuxManager    TmuxManager
	sandboxManager SandboxManager
	timeFormatter  *TimeFormatter
}

// NewDetector creates a new status detector
func NewDetector(tmuxManager TmuxManager, sandboxManager SandboxManager) *Detector {
	return &Detector{
		tmuxManager:    tmuxManager,
		sandboxManager: sandboxManager,
		timeFormatter:  NewTimeFormatter(),
	}
}

// DetectSessionStatus determines the current status of a session
func (d *Detector) DetectSessionStatus(session config.SessionMetadata) SessionStatus {
	// Check if tmux session exists
	tmuxExists := false
	if session.TmuxSession != "" {
		exists, err := d.tmuxManager.SessionExists(session.TmuxSession)
		if err == nil {
			tmuxExists = exists
		}
	}

	// Check for stop.json file in sandbox/.sbs/ first, then fallback to direct file access
	var stopTime time.Time
	var err error
	if session.SandboxName != "" {
		stopTime, err = d.ParseStopJsonFromSandbox(session.SandboxName, ".sbs/stop.json")
		// If sandbox reading fails, fallback to direct file access
		if err != nil {
			stopFilePath := filepath.Join(session.WorktreePath, ".sbs", "stop.json")
			stopTime, err = d.ParseStopJsonFile(stopFilePath)
		}
	} else {
		// No sandbox name available, use direct file access (for backward compatibility)
		stopFilePath := filepath.Join(session.WorktreePath, ".sbs", "stop.json")
		stopTime, err = d.ParseStopJsonFile(stopFilePath)
	}

	now := time.Now()

	if err == nil && !stopTime.IsZero() {
		// stop.json exists and is valid - session is stopped
		timeDelta := d.timeFormatter.FormatTimeDelta(stopTime, now)
		return SessionStatus{
			Status:     "stopped",
			LastChange: &stopTime,
			TimeDelta:  timeDelta,
		}
	}

	if tmuxExists {
		// Tmux session exists, no valid stop file - session is active
		return SessionStatus{
			Status:    "active",
			TimeDelta: "now",
		}
	}

	if err != nil && !os.IsNotExist(err) {
		// Stop file exists but can't be parsed - unknown status
		return SessionStatus{
			Status:    "unknown",
			TimeDelta: "unknown",
		}
	}

	// No tmux session and no stop file - session is stale
	// Use last activity from session metadata if available
	var lastActivity time.Time
	if session.LastActivity != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, session.LastActivity); parseErr == nil {
			lastActivity = parsed
		}
	}

	timeDelta := "unknown"
	if !lastActivity.IsZero() {
		timeDelta = d.timeFormatter.FormatTimeDelta(lastActivity, now)
	}

	return SessionStatus{
		Status:     "stale",
		LastChange: &lastActivity,
		TimeDelta:  timeDelta,
	}
}

// ParseStopJsonFile parses a stop.json file and extracts the timestamp
func (d *Detector) ParseStopJsonFile(filePath string) (time.Time, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return time.Time{}, err
	}

	if len(data) == 0 {
		return time.Time{}, fmt.Errorf("empty stop.json file")
	}

	var stopData map[string]interface{}
	if err := json.Unmarshal(data, &stopData); err != nil {
		return time.Time{}, fmt.Errorf("invalid JSON in stop.json: %w", err)
	}

	// Try to extract timestamp from different formats
	timestamp := extractTimestamp(stopData)
	if timestamp.IsZero() {
		return time.Time{}, fmt.Errorf("no valid timestamp found in stop.json")
	}

	return timestamp, nil
}

// ParseStopJsonFromSandbox parses a stop.json file from within a sandbox and extracts the timestamp
func (d *Detector) ParseStopJsonFromSandbox(sandboxName, filePath string) (time.Time, error) {
	data, err := d.sandboxManager.ReadFileFromSandbox(sandboxName, filePath)
	if err != nil {
		return time.Time{}, err
	}

	if len(data) == 0 {
		return time.Time{}, fmt.Errorf("empty stop.json file")
	}

	var stopData map[string]interface{}
	if err := json.Unmarshal(data, &stopData); err != nil {
		return time.Time{}, fmt.Errorf("invalid JSON in stop.json: %w", err)
	}

	// Try to extract timestamp from different formats
	timestamp := extractTimestamp(stopData)
	if timestamp.IsZero() {
		return time.Time{}, fmt.Errorf("no valid timestamp found in stop.json")
	}

	return timestamp, nil
}

// CalculateTimeDelta calculates human-readable time delta from a timestamp
func (d *Detector) CalculateTimeDelta(timestamp time.Time) string {
	return d.timeFormatter.FormatTimeDelta(timestamp, time.Now())
}

// extractTimestamp extracts timestamp from various stop.json formats
func extractTimestamp(data map[string]interface{}) time.Time {
	// Try Claude Code hook format first
	if hookData, ok := data["claude_code_hook"].(map[string]interface{}); ok {
		if timestampStr, ok := hookData["timestamp"].(string); ok {
			if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				return timestamp
			}
		}
	}

	// Try direct timestamp field
	if timestampStr, ok := data["timestamp"].(string); ok {
		if timestamp, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			return timestamp
		}
	}

	return time.Time{}
}
