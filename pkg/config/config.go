package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	WorktreeBasePath string `json:"worktree_base_path"`
	GitHubToken      string `json:"github_token"`
	WorkIssueScript  string `json:"work_issue_script"`
	RepoPath         string `json:"repo_path"`
}

type SessionMetadata struct {
	IssueNumber    int    `json:"issue_number"`
	IssueTitle     string `json:"issue_title"`
	Branch         string `json:"branch"`
	WorktreePath   string `json:"worktree_path"`
	TmuxSession    string `json:"tmux_session"`
	SandboxName    string `json:"sandbox_name"`
	RepositoryName string `json:"repository_name"`
	RepositoryRoot string `json:"repository_root"`
	CreatedAt      string `json:"created_at"`
	LastActivity   string `json:"last_activity"`
	Status         string `json:"status"` // active, stopped, stale
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
	
	sessionsPath := filepath.Join(homeDir, ".config", "work-orchestrator", "sessions.json")
	return LoadSessionsFromPath(sessionsPath)
}

// SaveSessions saves sessions to the global location (for backward compatibility)
func SaveSessions(sessions []SessionMetadata) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configDir := filepath.Join(homeDir, ".config", "work-orchestrator")
	sessionsPath := filepath.Join(configDir, "sessions.json")
	return SaveSessionsToPath(sessions, sessionsPath)
}

// LoadAllRepositorySessions loads sessions from all known repositories
func LoadAllRepositorySessions() ([]SessionMetadata, error) {
	var allSessions []SessionMetadata
	
	// Load global sessions first (for backward compatibility)
	globalSessions, err := LoadSessions()
	if err == nil {
		allSessions = append(allSessions, globalSessions...)
	}
	
	// Discover and load repository-specific sessions
	repoSessions, err := discoverRepositorySessions()
	if err == nil {
		allSessions = append(allSessions, repoSessions...)
	}
	
	return allSessions, nil
}

// discoverRepositorySessions finds and loads sessions from repository-specific locations
func discoverRepositorySessions() ([]SessionMetadata, error) {
	var allSessions []SessionMetadata
	
	// Get common workspace directories to search
	searchPaths := getWorkspaceSearchPaths()
	
	for _, basePath := range searchPaths {
		sessions, err := scanForSessionFiles(basePath)
		if err == nil {
			allSessions = append(allSessions, sessions...)
		}
	}
	
	return allSessions, nil
}

// getWorkspaceSearchPaths returns common directories where repositories might be located
func getWorkspaceSearchPaths() []string {
	homeDir, _ := os.UserHomeDir()
	
	// Common workspace locations
	searchPaths := []string{
		filepath.Join(homeDir, "code"),
		filepath.Join(homeDir, "projects"),
		filepath.Join(homeDir, "workspace"),
		filepath.Join(homeDir, "dev"),
		filepath.Join(homeDir, "src"),
		filepath.Join(homeDir, "git"),
		homeDir, // Also search home directory itself
	}
	
	// Add current working directory and its parent
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, cwd)
		searchPaths = append(searchPaths, filepath.Dir(cwd))
	}
	
	return searchPaths
}

// scanForSessionFiles recursively searches for .work-orchestrator/sessions.json files
func scanForSessionFiles(basePath string) ([]SessionMetadata, error) {
	var allSessions []SessionMetadata
	
	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return allSessions, nil
	}
	
	// Walk through directories looking for .work-orchestrator/sessions.json
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there are permission errors
		}
		
		// Skip hidden directories (except .work-orchestrator itself)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".work-orchestrator" {
			return filepath.SkipDir
		}
		
		// Look for sessions.json files in .work-orchestrator directories
		if info.Name() == "sessions.json" && strings.HasSuffix(filepath.Dir(path), ".work-orchestrator") {
			sessions, err := LoadSessionsFromPath(path)
			if err == nil {
				allSessions = append(allSessions, sessions...)
			}
		}
		
		// Limit recursion depth to avoid scanning too deep
		depth := strings.Count(strings.TrimPrefix(path, basePath), string(os.PathSeparator))
		if depth > 3 {
			return filepath.SkipDir
		}
		
		return nil
	})
	
	return allSessions, err
}