package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	Primary     = lipgloss.Color("#8b5cf6") // purple
	Secondary   = lipgloss.Color("#6366f1") // indigo
	Success     = lipgloss.Color("#10b981") // green
	Warning     = lipgloss.Color("#f59e0b") // amber
	Danger      = lipgloss.Color("#ef4444") // red
	Muted       = lipgloss.Color("#6b7280") // gray
	Background  = lipgloss.Color("#1f2937") // dark gray
	Foreground  = lipgloss.Color("#f9fafb") // white
	BorderColor = lipgloss.Color("#4b5563") // gray
)

// Status colors
var (
	PendingColor    = lipgloss.Color("#6b7280") // gray
	InProgressColor = lipgloss.Color("#3b82f6") // blue
	CompletedColor  = lipgloss.Color("#10b981") // green
)

// Base styles
var (
	// App container
	AppStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Foreground).
			Background(Primary).
			Padding(0, 2)

	// Subtitle
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Border box
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	// Selected item
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	// Normal item
	NormalStyle = lipgloss.NewStyle().
			Foreground(Foreground)

	// Muted text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Padding(1, 0)

	// Key style for help
	KeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Value style for details
	ValueStyle = lipgloss.NewStyle().
			Foreground(Foreground)

	// Label style for details
	LabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Width(12)

	// Error style
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	// Success style
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	// Warning style
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)
)

// Status styles
var (
	PendingStyle = lipgloss.NewStyle().
			Foreground(PendingColor)

	InProgressStyle = lipgloss.NewStyle().
			Foreground(InProgressColor)

	CompletedStyle = lipgloss.NewStyle().
			Foreground(CompletedColor)
)

// GetStatusStyle returns the appropriate style for a status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "pending":
		return PendingStyle
	case "in_progress":
		return InProgressStyle
	case "completed":
		return CompletedStyle
	default:
		return MutedStyle
	}
}

// Group header style
var GroupHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(Foreground)

// Task item styles
var (
	TaskItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	TaskSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true).
				Foreground(Primary)

	BlockedByStyle = lipgloss.NewStyle().
			Foreground(Muted).
			PaddingLeft(4).
			Italic(true)
)

// Filter bar style
var FilterBarStyle = lipgloss.NewStyle().
	Foreground(Muted).
	Padding(0, 0, 1, 0)

// Dialog styles
var (
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Width(60)

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Primary).
				MarginBottom(1)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(Foreground).
			Background(Muted).
			Padding(0, 2)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(Foreground).
				Background(Primary).
				Padding(0, 2)
)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Primary).
				Padding(0, 1)

	InputLabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(0)
)

// Color swatch style
func ColorSwatchStyle(color string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color(color)).
		Width(2)
}

// Horizontal line
func HorizontalLine(width int) string {
	return lipgloss.NewStyle().
		Foreground(BorderColor).
		Render(repeatString("─", width))
}

// repeatString repeats a string n times
func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
