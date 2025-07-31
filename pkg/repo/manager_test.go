package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_SanitizeName_WithMaxLength(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "basic_length_limiting",
			input:     "Fix user authentication bug in login system",
			maxLength: 32,
			expected:  "fix-user-authentication-bug-in",
		},
		{
			name:      "exactly_max_length",
			input:     "fix-user-authentication-bug-sys",
			maxLength: 32,
			expected:  "fix-user-authentication-bug-sys",
		},
		{
			name:      "under_max_length",
			input:     "fix-bug",
			maxLength: 32,
			expected:  "fix-bug",
		},
		{
			name:      "zero_max_length_uses_existing_behavior",
			input:     "myproject-name",
			maxLength: 0,
			expected:  "myproject-name",
		},
		{
			name:      "word_boundary_truncation",
			input:     "Fix authentication system completely",
			maxLength: 20,
			expected:  "fix-authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)
			if tt.maxLength > 0 {
				assert.LessOrEqual(t, len(result), tt.maxLength)
			}
		})
	}
}

func TestManager_SanitizeName_AlphanumericOnly(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "special_characters",
			input:    "Fix café login (UTF-8 encoding)",
			expected: "fix-cafe-login-utf-8-encoding",
		},
		{
			name:     "numbers_and_chars",
			input:    "Fix API v2.1 integration",
			expected: "fix-api-v2-1-integration",
		},
		{
			name:     "only_alphanumeric_and_hyphens",
			input:    "Fix-user-auth-123",
			expected: "fix-user-auth-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, 32)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SanitizeName_LowercaseConversion(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mixed_case",
			input:    "Fix User Authentication BUG",
			expected: "fix-user-authentication-bug",
		},
		{
			name:     "all_uppercase",
			input:    "FIX AUTHENTICATION",
			expected: "fix-authentication",
		},
		{
			name:     "already_lowercase",
			input:    "fix authentication",
			expected: "fix-authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, 32)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SanitizeName_HyphenReplacement(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple_spaces",
			input:    "Fix  multiple   spaces",
			expected: "fix-multiple-spaces",
		},
		{
			name:     "mixed_separators",
			input:    "Fix  multiple   spaces & symbols!",
			expected: "fix-multiple-spaces-symbols",
		},
		{
			name:     "tabs_and_newlines",
			input:    "Fix\tmultiple\nspaces",
			expected: "fix-multiple-spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, 32)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SanitizeName_LeadingTrailingHyphens(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "leading_trailing_spaces",
			input:    "  Fix bug  ",
			expected: "fix-bug",
		},
		{
			name:     "leading_trailing_special_chars",
			input:    "!!!Fix bug???",
			expected: "fix-bug",
		},
		{
			name:     "only_special_chars_at_ends",
			input:    "---Fix bug---",
			expected: "fix-bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, 32)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SanitizeName_UnicodeCharacters(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "accented_characters",
			input:    "Fix café login",
			expected: "fix-cafe-login",
		},
		{
			name:     "mixed_unicode",
			input:    "Fix café & résumé",
			expected: "fix-cafe-resume",
		},
		{
			name:     "emoji_and_symbols",
			input:    "Fix bug 🐛 now!",
			expected: "fix-bug-now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, 32)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_SanitizeName_EmptyString(t *testing.T) {
	manager := NewManager()

	result := manager.SanitizeName("", 32)
	assert.Equal(t, "", result)
}

func TestManager_SanitizeName_WordBoundaryTruncation(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "truncate_at_word_boundary",
			input:     "Fix authentication system completely",
			maxLength: 20,
			expected:  "fix-authentication",
		},
		{
			name:      "truncate_mid_word_if_necessary",
			input:     "superlongwordwithouthyphens",
			maxLength: 10,
			expected:  "superlong-", // Should truncate and trim trailing hyphen
		},
		{
			name:      "single_long_word",
			input:     "verylongauthenticationsystem",
			maxLength: 15,
			expected:  "verylongauthent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.SanitizeName(tt.input, tt.maxLength)
			assert.LessOrEqual(t, len(result), tt.maxLength)
			// Result should not end with hyphen unless it's a single hyphen
			if len(result) > 1 {
				assert.NotEqual(t, "-", result[len(result)-1:])
			}
		})
	}
}

func TestManager_SanitizeName_BackwardCompatibility(t *testing.T) {
	manager := NewManager()

	// Test that the original SanitizeName behavior is preserved when maxLength is 0
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "repository_name",
			input:    "myproject-name",
			expected: "myproject-name",
		},
		{
			name:     "with_special_chars",
			input:    "My Project Name!",
			expected: "my-project-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test original method
			originalResult := manager.SanitizeName(tt.input)
			// Test new method with maxLength 0 (should use existing behavior)
			newResult := manager.SanitizeName(tt.input, 0)
			assert.Equal(t, originalResult, newResult)
			assert.Equal(t, tt.expected, newResult)
		})
	}
}
