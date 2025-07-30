package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#F25D94")
	accentColor    = lipgloss.Color("#04B575")
	warningColor   = lipgloss.Color("#FF8C00")
	errorColor     = lipgloss.Color("#FF6B6B")
	mutedColor     = lipgloss.Color("#6C7086")
	
	// Base styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(0, 1)
	
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(primaryColor).
		Padding(0, 1).
		MarginBottom(1)
	
	selectedItemStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(primaryColor).
		Padding(0, 1)
	
	normalItemStyle = lipgloss.NewStyle().
		Padding(0, 1)
	
	statusActiveStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor)
	
	statusStoppedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(warningColor)
	
	statusStaleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(errorColor)
	
	mutedStyle = lipgloss.NewStyle().
		Foreground(mutedColor)
	
	helpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		MarginTop(1)
	
	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(primaryColor).
		Padding(0, 1).
		AlignHorizontal(lipgloss.Left)
	
	tableCellStyle = lipgloss.NewStyle().
		Padding(0, 1).
		AlignHorizontal(lipgloss.Left)
	
	selectedRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Bold(true)
)

func FormatStatus(status string) string {
	switch status {
	case "active":
		return statusActiveStyle.Render("●")
	case "stopped":
		return statusStoppedStyle.Render("●")
	case "stale":
		return statusStaleStyle.Render("●")
	default:
		return mutedStyle.Render("●")
	}
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}