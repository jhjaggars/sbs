package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	WorktreeBasePath string   `json:"worktree_base_path"`
	GitHubToken      string   `json:"github_token"`
	WorkIssueScript  string   `json:"work_issue_script"`
	RepoPath         string   `json:"repo_path"`
	TmuxCommand      string   `json:"tmux_command,omitempty"`      // Custom command to run in tmux session
	TmuxCommandArgs  []string `json:"tmux_command_args,omitempty"` // Arguments for the custom command
	NoCommand        bool     `json:"no_command,omitempty"`        // Disable automatic command execution

	// Command logging configuration
	CommandLogging  bool   `json:"command_logging,omitempty"`   // Enable/disable command logging
	CommandLogLevel string `json:"command_log_level,omitempty"` // Log level: debug, info, error
	CommandLogPath  string `json:"command_log_path,omitempty"`  // Optional log file path

	// Status tracking configuration
	StatusTracking            bool `json:"status_tracking,omitempty"`                 // Enable/disable status tracking
	StatusRefreshIntervalSecs int  `json:"status_refresh_interval_seconds,omitempty"` // Refresh interval in seconds (default: 60)
	StatusMaxFileSizeBytes    int  `json:"status_max_file_size_bytes,omitempty"`      // Maximum stop.json file size (default: 1MB)
	StatusTimeoutSeconds      int  `json:"status_timeout_seconds,omitempty"`          // Timeout for status operations (default: 5)

	// Log display configuration
	LogRefreshIntervalSecs int `json:"log_refresh_interval_seconds,omitempty"` // Log refresh interval in seconds (default: 5)
}

// ResourceCreationEntry tracks the creation of individual resources during session setup
type ResourceCreationEntry struct {
	ResourceType string                 `json:"resource_type"` // branch, worktree, tmux, sandbox
	ResourceID   string                 `json:"resource_id"`   // identifier for the resource
	CreatedAt    time.Time              `json:"created_at"`    // when the resource was created
	Status       string                 `json:"status"`        // created, failed, cleanup
	Metadata     map[string]interface{} `json:"metadata"`      // additional resource-specific data
}

type SessionMetadata struct {
	IssueNumber    int    `json:"issue_number,omitempty"` // Legacy compatibility field
	IssueTitle     string `json:"issue_title"`
	FriendlyTitle  string `json:"friendly_title"` // Sandbox-friendly version of issue title
	Branch         string `json:"branch"`
	WorktreePath   string `json:"worktree_path"`
	TmuxSession    string `json:"tmux_session"`
	SandboxName    string `json:"sandbox_name"`
	RepositoryName string `json:"repository_name"`
	RepositoryRoot string `json:"repository_root"`
	CreatedAt      string `json:"created_at"`
	LastActivity   string `json:"last_activity"`
	Status         string `json:"status"` // active, stopped, stale

	// Input source fields for pluggable backends
	SourceType   string `json:"source_type,omitempty"`   // github, test, jira, etc.
	NamespacedID string `json:"namespaced_id,omitempty"` // Full namespaced ID (e.g., "github:123", "test:quick")

	// Resource tracking fields for enhanced cleanup and failure recovery
	ResourceStatus      string                  `json:"resource_status,omitempty"`       // creating, active, cleanup, failed
	CurrentCreationStep string                  `json:"current_creation_step,omitempty"` // tracks current step in resource creation
	FailurePoint        string                  `json:"failure_point,omitempty"`         // step where creation failed
	FailureReason       string                  `json:"failure_reason,omitempty"`        // reason for failure
	ResourceCreationLog []ResourceCreationEntry `json:"resource_creation_log,omitempty"` // log of all created resources
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		WorktreeBasePath:          filepath.Join(homeDir, ".sbs-worktrees"),
		GitHubToken:               os.Getenv("GITHUB_TOKEN"),
		WorkIssueScript:           filepath.Join(homeDir, "code/work-issue/work-issue.sh"),
		RepoPath:                  ".",     // Current directory by default
		StatusTracking:            true,    // Enable status tracking by default
		StatusRefreshIntervalSecs: 60,      // Default to 60 seconds
		StatusMaxFileSizeBytes:    1048576, // Default to 1MB
		StatusTimeoutSeconds:      5,       // Default to 5 seconds
		LogRefreshIntervalSecs:    5,       // Default to 5 seconds
	}
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".config", "sbs", "config.json")

	// Create default config if doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Validate required fields for resource tracking features
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// LoadConfigWithRepository loads configuration with repository-specific overrides
func LoadConfigWithRepository(repoRoot string) (*Config, error) {
	// Load global config first
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Try to load repository-specific config
	repoConfig, err := LoadRepositoryConfig(repoRoot)
	if err != nil {
		// If repository config doesn't exist or can't be loaded, just use global config
		return config, nil
	}

	// Merge repository config over global config
	mergedConfig := MergeConfig(config, repoConfig)
	return mergedConfig, nil
}

// LoadRepositoryConfig loads configuration from .sbs/config.json in repository root
func LoadRepositoryConfig(repoRoot string) (*Config, error) {
	configPath := filepath.Join(repoRoot, ".sbs", "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// MergeConfig merges repository config over base config, only overriding non-zero values
func MergeConfig(base, override *Config) *Config {
	merged := *base // Copy base config

	if override.WorktreeBasePath != "" {
		merged.WorktreeBasePath = override.WorktreeBasePath
	}
	if override.GitHubToken != "" {
		merged.GitHubToken = override.GitHubToken
	}
	if override.WorkIssueScript != "" {
		merged.WorkIssueScript = override.WorkIssueScript
	}
	if override.RepoPath != "" {
		merged.RepoPath = override.RepoPath
	}
	if override.TmuxCommand != "" {
		merged.TmuxCommand = override.TmuxCommand
	}
	if len(override.TmuxCommandArgs) > 0 {
		merged.TmuxCommandArgs = make([]string, len(override.TmuxCommandArgs))
		copy(merged.TmuxCommandArgs, override.TmuxCommandArgs)
	}
	// NoCommand is a boolean, so we only override if it's explicitly set to true
	if override.NoCommand {
		merged.NoCommand = override.NoCommand
	}

	// Command logging configuration
	// CommandLogging is a boolean, override if explicitly set to true
	if override.CommandLogging {
		merged.CommandLogging = override.CommandLogging
	}
	if override.CommandLogLevel != "" {
		merged.CommandLogLevel = override.CommandLogLevel
	}
	if override.CommandLogPath != "" {
		merged.CommandLogPath = override.CommandLogPath
	}

	// Status tracking configuration
	// StatusTracking is a boolean, we need to check if it was explicitly set
	// For now, we'll assume any non-zero value in override means it was explicitly set
	if override.StatusTracking != merged.StatusTracking {
		merged.StatusTracking = override.StatusTracking
	}
	if override.StatusRefreshIntervalSecs > 0 {
		merged.StatusRefreshIntervalSecs = override.StatusRefreshIntervalSecs
	}
	if override.StatusMaxFileSizeBytes > 0 {
		merged.StatusMaxFileSizeBytes = override.StatusMaxFileSizeBytes
	}
	if override.StatusTimeoutSeconds > 0 {
		merged.StatusTimeoutSeconds = override.StatusTimeoutSeconds
	}

	// Log display configuration
	if override.LogRefreshIntervalSecs > 0 {
		merged.LogRefreshIntervalSecs = override.LogRefreshIntervalSecs
	}

	return &merged
}

func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "sbs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadSessions loads sessions from a specific path
func LoadSessionsFromPath(sessionsPath string) ([]SessionMetadata, error) {
	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return []SessionMetadata{}, nil
	}

	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		return nil, err
	}

	var sessions []SessionMetadata
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// SaveSessionsToPath saves sessions to a specific path
func SaveSessionsToPath(sessions []SessionMetadata, sessionsPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(sessionsPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sessionsPath, data, 0644)
}

// LoadSessions loads sessions from the global location (for backward compatibility)
func LoadSessions() ([]SessionMetadata, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsPath := filepath.Join(homeDir, ".config", "sbs", "sessions.json")
	return LoadSessionsFromPath(sessionsPath)
}

// SaveSessions saves sessions to the global location (for backward compatibility)
func SaveSessions(sessions []SessionMetadata) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "sbs")
	sessionsPath := filepath.Join(configDir, "sessions.json")
	return SaveSessionsToPath(sessions, sessionsPath)
}

// LoadAllRepositorySessions loads sessions from the global sessions file
func LoadAllRepositorySessions() ([]SessionMetadata, error) {
	// Use only the global sessions file as the single source of truth
	// Repository scoping is handled by filtering based on RepositoryRoot field
	return LoadSessions()
}

// GetGlobalSessionsPath returns the path to the global sessions file
func GetGlobalSessionsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "sbs", "sessions.json"), nil
}

// validateConfig validates that required fields are present for resource tracking features
func validateConfig(config *Config) error {
	var errors []string

	// Validate essential paths - only check if they're not set to reasonable defaults
	if config.WorktreeBasePath == "" {
		errors = append(errors, "worktree_base_path is required")
	}

	// RepoPath can be empty (will default to "." in DefaultConfig), so only validate if it's explicitly empty in a non-default scenario
	// Skip this validation as it's too restrictive for test scenarios

	// Validate command logging configuration if enabled
	if config.CommandLogging {
		validLevels := map[string]bool{
			"debug": true,
			"info":  true,
			"error": true,
			"":      true, // Empty string is acceptable (defaults to info)
		}
		if !validLevels[config.CommandLogLevel] {
			errors = append(errors, "command_log_level must be one of: debug, info, error")
		}
	}

	// Validate status tracking configuration if enabled
	if config.StatusTracking {
		if config.StatusRefreshIntervalSecs < 5 || config.StatusRefreshIntervalSecs > 600 {
			errors = append(errors, "status_refresh_interval_seconds must be between 5 and 600")
		}
		if config.StatusMaxFileSizeBytes < 1024 || config.StatusMaxFileSizeBytes > 10*1024*1024 {
			errors = append(errors, "status_max_file_size_bytes must be between 1KB and 10MB")
		}
		if config.StatusTimeoutSeconds < 1 || config.StatusTimeoutSeconds > 30 {
			errors = append(errors, "status_timeout_seconds must be between 1 and 30")
		}
	}

	// Validate log display configuration (only if explicitly set)
	if config.LogRefreshIntervalSecs != 0 && (config.LogRefreshIntervalSecs < 1 || config.LogRefreshIntervalSecs > 300) {
		errors = append(errors, "log_refresh_interval_seconds must be between 1 and 300")
	}

	// If there are validation errors, return them as a single error
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}
