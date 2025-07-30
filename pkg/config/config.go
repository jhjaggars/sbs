package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WorktreeBasePath string `json:"worktree_base_path"`
	GitHubToken      string `json:"github_token"`
	WorkIssueScript  string `json:"work_issue_script"`
	RepoPath         string `json:"repo_path"`
}

type SessionMetadata struct {
	IssueNumber  int    `json:"issue_number"`
	IssueTitle   string `json:"issue_title"`
	Branch       string `json:"branch"`
	WorktreePath string `json:"worktree_path"`
	TmuxSession  string `json:"tmux_session"`
	SandboxName  string `json:"sandbox_name"`
	CreatedAt    string `json:"created_at"`
	LastActivity string `json:"last_activity"`
	Status       string `json:"status"` // active, stopped, stale
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		WorktreeBasePath: filepath.Join(homeDir, ".work-issue-worktrees"),
		GitHubToken:      os.Getenv("GITHUB_TOKEN"),
		WorkIssueScript:  filepath.Join(homeDir, "code/work-issue/work-issue.sh"),
		RepoPath:         ".", // Current directory by default
	}
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	configPath := filepath.Join(homeDir, ".config", "work-orchestrator", "config.json")
	
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
	
	return &config, nil
}

func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(homeDir, ".config", "work-orchestrator")
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

func LoadSessions() ([]SessionMetadata, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	sessionsPath := filepath.Join(homeDir, ".config", "work-orchestrator", "sessions.json")
	
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

func SaveSessions(sessions []SessionMetadata) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(homeDir, ".config", "work-orchestrator")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	sessionsPath := filepath.Join(configDir, "sessions.json")
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(sessionsPath, data, 0644)
}