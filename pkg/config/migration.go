package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MigrateSessionMetadata migrates session metadata to include input source fields
// This ensures backward compatibility while adding support for pluggable input sources
func MigrateSessionMetadata(sessions []SessionMetadata) ([]SessionMetadata, error) {
	if sessions == nil {
		return []SessionMetadata{}, nil
	}

	migratedSessions := make([]SessionMetadata, len(sessions))

	for i, session := range sessions {
		if sessionNeedsMigration(&session) {
			migrated, err := migrateSession(&session)
			if err != nil {
				return nil, fmt.Errorf("failed to migrate session for issue %d: %w", session.IssueNumber, err)
			}
			migratedSessions[i] = *migrated
		} else {
			// Session already migrated or doesn't need migration
			migratedSessions[i] = session
		}
	}

	return migratedSessions, nil
}

// sessionNeedsMigration determines if a session needs to be migrated
func sessionNeedsMigration(session *SessionMetadata) bool {
	// Session needs migration if it's missing source type or namespaced ID
	return strings.TrimSpace(session.SourceType) == "" || strings.TrimSpace(session.NamespacedID) == ""
}

// migrateSession migrates a single session to the new format
func migrateSession(session *SessionMetadata) (*SessionMetadata, error) {
	// Create a copy of the session to avoid modifying the original
	migrated := *session

	// Determine source type and namespaced ID based on existing data
	if migrated.SourceType == "" || migrated.NamespacedID == "" {
		// For legacy sessions, we assume GitHub source since that was the only option
		if session.IssueNumber > 0 {
			// Regular GitHub issue
			migrated.SourceType = "github"
			migrated.NamespacedID = fmt.Sprintf("github:%d", session.IssueNumber)
		} else {
			// Handle edge case where issue number is 0 or negative
			// This might be a test session or corrupted data
			// Try to infer from branch name or default to GitHub
			if strings.Contains(session.Branch, "test-") {
				migrated.SourceType = "test"
				// Try to extract test ID from branch
				if testID := extractTestIDFromBranch(session.Branch); testID != "" {
					migrated.NamespacedID = fmt.Sprintf("test:%s", testID)
				} else {
					migrated.NamespacedID = "test:unknown"
				}
			} else {
				// Default to GitHub with unknown ID
				migrated.SourceType = "github"
				migrated.NamespacedID = "github:unknown"
			}
		}
	}

	return &migrated, nil
}

// extractTestIDFromBranch attempts to extract a test ID from a branch name
// Expected format: issue-test-{id}-{title-slug} or similar
func extractTestIDFromBranch(branch string) string {
	// Split by hyphens and look for test pattern
	parts := strings.Split(branch, "-")

	// Look for "test" followed by an ID
	for i, part := range parts {
		if part == "test" && i+1 < len(parts) {
			// Next part should be the test ID
			candidate := parts[i+1]
			// Validate it's a reasonable test ID (not empty, not a title word)
			if candidate != "" && !isCommonTitleWord(candidate) {
				return candidate
			}
		}
	}

	return ""
}

// isCommonTitleWord checks if a string is likely a title word rather than an ID
func isCommonTitleWord(word string) bool {
	commonWords := map[string]bool{
		"development": true,
		"test":        true,
		"fix":         true,
		"bug":         true,
		"feature":     true,
		"add":         true,
		"update":      true,
		"remove":      true,
		"implement":   true,
		"create":      true,
	}

	return commonWords[strings.ToLower(word)]
}

// MigrateSessionsInPlace migrates sessions and saves them to the same location
// This is a convenience function for automatic migration during session loading
func MigrateSessionsInPlace(sessionsPath string) error {
	// Load existing sessions
	sessions, err := LoadSessionsFromPath(sessionsPath)
	if err != nil {
		return fmt.Errorf("failed to load sessions for migration: %w", err)
	}

	// Check if any sessions need migration
	needsMigration := false
	for _, session := range sessions {
		if sessionNeedsMigration(&session) {
			needsMigration = true
			break
		}
	}

	// If no migration needed, return early
	if !needsMigration {
		return nil
	}

	// Migrate sessions
	migratedSessions, err := MigrateSessionMetadata(sessions)
	if err != nil {
		return fmt.Errorf("failed to migrate sessions: %w", err)
	}

	// Save migrated sessions
	if err := SaveSessionsToPath(migratedSessions, sessionsPath); err != nil {
		return fmt.Errorf("failed to save migrated sessions: %w", err)
	}

	return nil
}

// GetGlobalSessionsPath returns the path to the global sessions file
func GetGlobalSessionsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "sbs", "sessions.json"), nil
}

// LoadSessionsWithMigration loads sessions and automatically migrates them if needed
// This should be used instead of LoadSessionsFromPath to ensure migration happens
func LoadSessionsWithMigration(sessionsPath string) ([]SessionMetadata, error) {
	// Load existing sessions
	sessions, err := LoadSessionsFromPath(sessionsPath)
	if err != nil {
		return nil, err
	}

	// Migrate if needed
	migratedSessions, err := MigrateSessionMetadata(sessions)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate sessions: %w", err)
	}

	// Check if migration occurred and save if needed
	migrationOccurred := false
	for i := range sessions {
		if i < len(migratedSessions) {
			if sessions[i].SourceType != migratedSessions[i].SourceType ||
				sessions[i].NamespacedID != migratedSessions[i].NamespacedID {
				migrationOccurred = true
				break
			}
		}
	}

	// Save migrated sessions if migration occurred
	if migrationOccurred {
		if err := SaveSessionsToPath(migratedSessions, sessionsPath); err != nil {
			// Log warning but don't fail - we can still return the migrated sessions
			// In a real implementation, we might use a logger here
			// fmt.Printf("Warning: failed to save migrated sessions: %v\n", err)
		}
	}

	return migratedSessions, nil
}
