package tui

import (
	"fmt"
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

	// Modal dialog styles
	modalBackgroundStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#282828")).
				Foreground(lipgloss.Color("#F8F8F2"))

	modalContentStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#44475A")).
				Foreground(lipgloss.Color("#F8F8F2")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2).
				Bold(true)

	confirmationTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F8F8F2")).
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

// ColumnWidths represents the calculated widths for table columns
type ColumnWidths struct {
	Issue        int
	Title        int
	Repository   int
	Branch       int
	Status       int
	LastActivity int
}

// CalculateRepositoryViewWidths calculates column widths for repository view based on terminal width
func CalculateRepositoryViewWidths(terminalWidth int) ColumnWidths {
	// Account for padding and spacing between columns (roughly 6 spaces for 5 columns)
	availableWidth := terminalWidth - 6

	// Minimum widths for readability
	const (
		minIssue        = 6
		minTitle        = 20
		minBranch       = 12
		minStatus       = 10
		minLastActivity = 10
	)

	// Start with minimum widths
	widths := ColumnWidths{
		Issue:        minIssue,
		Title:        minTitle,
		Branch:       minBranch,
		Status:       minStatus,
		LastActivity: minLastActivity,
	}

	usedWidth := minIssue + minTitle + minBranch + minStatus + minLastActivity

	// If we have extra space, allocate it to the title column
	if availableWidth > usedWidth {
		extraSpace := availableWidth - usedWidth
		widths.Title += extraSpace

		// Cap title width at reasonable maximum
		if widths.Title > 60 {
			excess := widths.Title - 60
			widths.Title = 60
			// Distribute excess to branch if there's still space
			if widths.Branch < 25 {
				branchIncrease := min(excess, 25-widths.Branch)
				widths.Branch += branchIncrease
			}
		}
	}

	return widths
}

// CalculateGlobalViewWidths calculates column widths for global view based on terminal width
func CalculateGlobalViewWidths(terminalWidth int) ColumnWidths {
	// Account for padding and spacing between columns (roughly 7 spaces for 6 columns)
	availableWidth := terminalWidth - 7

	// Minimum widths for readability
	const (
		minIssue        = 6
		minTitle        = 15
		minRepository   = 12
		minBranch       = 10
		minStatus       = 10
		minLastActivity = 10
	)

	// Start with minimum widths
	widths := ColumnWidths{
		Issue:        minIssue,
		Title:        minTitle,
		Repository:   minRepository,
		Branch:       minBranch,
		Status:       minStatus,
		LastActivity: minLastActivity,
	}

	usedWidth := minIssue + minTitle + minRepository + minBranch + minStatus + minLastActivity

	// If we have extra space, distribute it intelligently
	if availableWidth > usedWidth {
		extraSpace := availableWidth - usedWidth

		// Give most space to title, some to repository and branch
		titleShare := extraSpace * 60 / 100                // 60% to title
		repoShare := extraSpace * 25 / 100                 // 25% to repository
		branchShare := extraSpace - titleShare - repoShare // remainder to branch

		widths.Title += titleShare
		widths.Repository += repoShare
		widths.Branch += branchShare

		// Apply reasonable maximums
		if widths.Title > 40 {
			excess := widths.Title - 40
			widths.Title = 40
			widths.Repository += excess / 2
			widths.Branch += excess / 2
		}
		if widths.Repository > 20 {
			excess := widths.Repository - 20
			widths.Repository = 20
			widths.Branch += excess
		}
		if widths.Branch > 20 {
			widths.Branch = 20
		}
	}

	return widths
}

// FormatRepositoryViewRow formats a row for repository view with given column widths
func FormatRepositoryViewRow(widths ColumnWidths, issue int, title, branch, status, lastActivity string) string {
	return fmt.Sprintf("%-*d %-*s %-*s %-*s %-*s",
		widths.Issue, issue,
		widths.Title, TruncateString(title, widths.Title),
		widths.Branch, TruncateString(branch, widths.Branch),
		widths.Status, status,
		widths.LastActivity, lastActivity,
	)
}

// FormatGlobalViewRow formats a row for global view with given column widths
func FormatGlobalViewRow(widths ColumnWidths, issue int, title, repository, branch, status, lastActivity string) string {
	return fmt.Sprintf("%-*d %-*s %-*s %-*s %-*s %-*s",
		widths.Issue, issue,
		widths.Title, TruncateString(title, widths.Title),
		widths.Repository, TruncateString(repository, widths.Repository),
		widths.Branch, TruncateString(branch, widths.Branch),
		widths.Status, status,
		widths.LastActivity, lastActivity,
	)
}

// FormatRepositoryViewHeader formats the header for repository view with given column widths
func FormatRepositoryViewHeader(widths ColumnWidths) string {
	return fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		widths.Issue, "Issue",
		widths.Title, "Title",
		widths.Branch, "Branch",
		widths.Status, "Status",
		widths.LastActivity, "Last Activity",
	)
}

// FormatGlobalViewHeader formats the header for global view with given column widths
func FormatGlobalViewHeader(widths ColumnWidths) string {
	return fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s",
		widths.Issue, "Issue",
		widths.Title, "Title",
		widths.Repository, "Repository",
		widths.Branch, "Branch",
		widths.Status, "Status",
		widths.LastActivity, "Last Activity",
	)
}

// CalculateIssueSelectWidths calculates column widths for issue selection view (Issue + Title only)
func CalculateIssueSelectWidths(terminalWidth int) ColumnWidths {
	// Account for padding and spacing between columns (roughly 3 spaces for 2 columns)
	availableWidth := terminalWidth - 3

	// Minimum widths for readability
	const (
		minIssue = 8
		minTitle = 20
	)

	// Start with minimum widths
	widths := ColumnWidths{
		Issue: minIssue,
		Title: minTitle,
	}

	usedWidth := minIssue + minTitle

	// If we have extra space, give it all to the title column
	if availableWidth > usedWidth {
		extraSpace := availableWidth - usedWidth
		widths.Title += extraSpace

		// Cap title width at reasonable maximum to prevent overly long lines
		if widths.Title > 100 {
			widths.Title = 100
		}
	}

	return widths
}

// FormatIssueSelectRow formats a row for issue selection view with given column widths
func FormatIssueSelectRow(widths ColumnWidths, issue int, title string) string {
	return fmt.Sprintf("%-*d %-*s",
		widths.Issue, issue,
		widths.Title, TruncateString(title, widths.Title),
	)
}

// FormatIssueSelectHeader formats the header for issue selection view with given column widths
func FormatIssueSelectHeader(widths ColumnWidths) string {
	return fmt.Sprintf("%-*s %-*s",
		widths.Issue, "Issue",
		widths.Title, "Title",
	)
}
