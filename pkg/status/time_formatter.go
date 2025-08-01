package status

import (
	"fmt"
	"time"
)

// TimeFormatter handles formatting of time deltas into human-readable strings
type TimeFormatter struct{}

// NewTimeFormatter creates a new TimeFormatter instance
func NewTimeFormatter() *TimeFormatter {
	return &TimeFormatter{}
}

// FormatTimeDelta formats a timestamp relative to a base time into a human-readable string
func (tf *TimeFormatter) FormatTimeDelta(timestamp, baseTime time.Time) string {
	// Handle zero time
	if timestamp.IsZero() {
		return "unknown"
	}

	duration := baseTime.Sub(timestamp)

	// Handle future timestamps (clock skew) - treat as "now"
	if duration < 0 {
		return "now"
	}

	// Less than 1 minute
	if duration < time.Minute {
		return "now"
	}

	// Minutes (1-59)
	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return formatTime(minutes, "m")
	}

	// Hours (1-23)
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return formatTime(hours, "h")
	}

	// Days (1-6)
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		return formatTime(days, "d")
	}

	// Weeks (7+ days)
	weeks := int(duration.Hours() / (24 * 7))
	return formatTime(weeks, "w")
}

// formatTime formats a number with a unit and " ago" suffix
func formatTime(value int, unit string) string {
	return fmt.Sprintf("%d%s ago", value, unit)
}
