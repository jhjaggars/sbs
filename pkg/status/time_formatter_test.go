package status

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeFormatter_FormatTimeDelta(t *testing.T) {
	baseTime := time.Date(2025, 8, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		baseTime  time.Time
		expected  string
	}{
		{
			name:      "just now - 30 seconds",
			timestamp: baseTime.Add(-30 * time.Second),
			baseTime:  baseTime,
			expected:  "now",
		},
		{
			name:      "5 minutes ago",
			timestamp: baseTime.Add(-5 * time.Minute),
			baseTime:  baseTime,
			expected:  "5m ago",
		},
		{
			name:      "1 minute ago",
			timestamp: baseTime.Add(-1 * time.Minute),
			baseTime:  baseTime,
			expected:  "1m ago",
		},
		{
			name:      "59 minutes ago",
			timestamp: baseTime.Add(-59 * time.Minute),
			baseTime:  baseTime,
			expected:  "59m ago",
		},
		{
			name:      "1 hour ago",
			timestamp: baseTime.Add(-1 * time.Hour),
			baseTime:  baseTime,
			expected:  "1h ago",
		},
		{
			name:      "2 hours ago",
			timestamp: baseTime.Add(-2 * time.Hour),
			baseTime:  baseTime,
			expected:  "2h ago",
		},
		{
			name:      "23 hours ago",
			timestamp: baseTime.Add(-23 * time.Hour),
			baseTime:  baseTime,
			expected:  "23h ago",
		},
		{
			name:      "1 day ago",
			timestamp: baseTime.Add(-24 * time.Hour),
			baseTime:  baseTime,
			expected:  "1d ago",
		},
		{
			name:      "3 days ago",
			timestamp: baseTime.Add(-3 * 24 * time.Hour),
			baseTime:  baseTime,
			expected:  "3d ago",
		},
		{
			name:      "6 days ago",
			timestamp: baseTime.Add(-6 * 24 * time.Hour),
			baseTime:  baseTime,
			expected:  "6d ago",
		},
		{
			name:      "1 week ago",
			timestamp: baseTime.Add(-7 * 24 * time.Hour),
			baseTime:  baseTime,
			expected:  "1w ago",
		},
		{
			name:      "2 weeks ago",
			timestamp: baseTime.Add(-14 * 24 * time.Hour),
			baseTime:  baseTime,
			expected:  "2w ago",
		},
		{
			name:      "4 weeks ago",
			timestamp: baseTime.Add(-28 * 24 * time.Hour),
			baseTime:  baseTime,
			expected:  "4w ago",
		},
	}

	formatter := NewTimeFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTimeDelta(tt.timestamp, tt.baseTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeFormatter_FormatNow(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	// Test various "now" scenarios
	nowCases := []time.Duration{
		0 * time.Second,
		15 * time.Second,
		30 * time.Second,
		45 * time.Second,
		59 * time.Second,
	}

	for _, duration := range nowCases {
		t.Run(duration.String(), func(t *testing.T) {
			timestamp := baseTime.Add(-duration)
			result := formatter.FormatTimeDelta(timestamp, baseTime)
			assert.Equal(t, "now", result)
		})
	}
}

func TestTimeFormatter_FormatMinutes(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	minuteCases := []struct {
		minutes  int
		expected string
	}{
		{1, "1m ago"},
		{2, "2m ago"},
		{15, "15m ago"},
		{30, "30m ago"},
		{45, "45m ago"},
		{59, "59m ago"},
	}

	for _, tc := range minuteCases {
		t.Run(tc.expected, func(t *testing.T) {
			timestamp := baseTime.Add(-time.Duration(tc.minutes) * time.Minute)
			result := formatter.FormatTimeDelta(timestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_FormatHours(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	hourCases := []struct {
		hours    int
		expected string
	}{
		{1, "1h ago"},
		{2, "2h ago"},
		{6, "6h ago"},
		{12, "12h ago"},
		{18, "18h ago"},
		{23, "23h ago"},
	}

	for _, tc := range hourCases {
		t.Run(tc.expected, func(t *testing.T) {
			timestamp := baseTime.Add(-time.Duration(tc.hours) * time.Hour)
			result := formatter.FormatTimeDelta(timestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_FormatDays(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	dayCases := []struct {
		days     int
		expected string
	}{
		{1, "1d ago"},
		{2, "2d ago"},
		{3, "3d ago"},
		{4, "4d ago"},
		{5, "5d ago"},
		{6, "6d ago"},
	}

	for _, tc := range dayCases {
		t.Run(tc.expected, func(t *testing.T) {
			timestamp := baseTime.Add(-time.Duration(tc.days) * 24 * time.Hour)
			result := formatter.FormatTimeDelta(timestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_FormatWeeks(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	weekCases := []struct {
		weeks    int
		expected string
	}{
		{1, "1w ago"},
		{2, "2w ago"},
		{3, "3w ago"},
		{4, "4w ago"},
		{8, "8w ago"},
		{12, "12w ago"},
	}

	for _, tc := range weekCases {
		t.Run(tc.expected, func(t *testing.T) {
			timestamp := baseTime.Add(-time.Duration(tc.weeks) * 7 * 24 * time.Hour)
			result := formatter.FormatTimeDelta(timestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_HandleFutureTimestamps(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	// Test future timestamps (clock skew scenarios)
	futureCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "5 minutes in future",
			duration: 5 * time.Minute,
			expected: "now", // Should handle gracefully
		},
		{
			name:     "1 hour in future",
			duration: 1 * time.Hour,
			expected: "now", // Should handle gracefully
		},
		{
			name:     "1 day in future",
			duration: 24 * time.Hour,
			expected: "now", // Should handle gracefully
		},
	}

	for _, tc := range futureCases {
		t.Run(tc.name, func(t *testing.T) {
			futureTimestamp := baseTime.Add(tc.duration)
			result := formatter.FormatTimeDelta(futureTimestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_EdgeCases(t *testing.T) {
	formatter := NewTimeFormatter()
	baseTime := time.Now()

	// Test edge cases
	edgeCases := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "exactly 1 minute",
			timestamp: baseTime.Add(-1 * time.Minute),
			expected:  "1m ago",
		},
		{
			name:      "exactly 1 hour",
			timestamp: baseTime.Add(-1 * time.Hour),
			expected:  "1h ago",
		},
		{
			name:      "exactly 1 day",
			timestamp: baseTime.Add(-24 * time.Hour),
			expected:  "1d ago",
		},
		{
			name:      "exactly 1 week",
			timestamp: baseTime.Add(-7 * 24 * time.Hour),
			expected:  "1w ago",
		},
		{
			name:      "zero time",
			timestamp: time.Time{},
			expected:  "unknown",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.FormatTimeDelta(tc.timestamp, baseTime)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeFormatter_WithCustomBaseTime(t *testing.T) {
	formatter := NewTimeFormatter()

	// Use a fixed base time for consistent testing
	baseTime := time.Date(2025, 8, 1, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "2 hours before base time",
			timestamp: time.Date(2025, 8, 1, 13, 30, 0, 0, time.UTC),
			expected:  "2h ago",
		},
		{
			name:      "1 day before base time",
			timestamp: time.Date(2025, 7, 31, 15, 30, 0, 0, time.UTC),
			expected:  "1d ago",
		},
		{
			name:      "1 week before base time",
			timestamp: time.Date(2025, 7, 25, 15, 30, 0, 0, time.UTC),
			expected:  "1w ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatTimeDelta(tt.timestamp, baseTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}
