package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Header renders the application header
func Header(title string, width int) string {
	titleText := TitleStyle.Render(title)
	return titleText + "\n" + HorizontalLine(width)
}

// Footer renders the help footer with auto line wrapping
func Footer(keys [][]string, width int) string {
	var parts []string
	for _, pair := range keys {
		key := KeyStyle.Render(fmt.Sprintf("[%s]", pair[0]))
		desc := MutedStyle.Render(pair[1])
		parts = append(parts, fmt.Sprintf("%s %s", key, desc))
	}

	// Wrap parts into lines based on width
	var lines []string
	var currentLine string
	for i, part := range parts {
		partWidth := lipgloss.Width(part)
		currentWidth := lipgloss.Width(currentLine)

		if currentLine == "" {
			currentLine = part
		} else if currentWidth+2+partWidth <= width {
			currentLine += "  " + part
		} else {
			lines = append(lines, currentLine)
			currentLine = part
		}

		if i == len(parts)-1 && currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	return HorizontalLine(width) + "\n" + strings.Join(lines, "\n")
}

// KeyHint represents a key binding with enabled state
type KeyHint struct {
	Key     string
	Desc    string
	Enabled bool
}

// FooterWithHints renders help footer with disabled keys grayed out and auto line wrapping
func FooterWithHints(hints []KeyHint, width int) string {
	var parts []string
	for _, hint := range hints {
		if hint.Enabled {
			key := KeyStyle.Render(fmt.Sprintf("[%s]", hint.Key))
			desc := MutedStyle.Render(hint.Desc)
			parts = append(parts, fmt.Sprintf("%s %s", key, desc))
		} else {
			// Disabled - fully grayed out
			key := DisabledStyle.Render(fmt.Sprintf("[%s]", hint.Key))
			desc := DisabledStyle.Render(hint.Desc)
			parts = append(parts, fmt.Sprintf("%s %s", key, desc))
		}
	}

	// Wrap parts into lines based on width
	var lines []string
	var currentLine string
	for i, part := range parts {
		partWidth := lipgloss.Width(part)
		currentWidth := lipgloss.Width(currentLine)

		if currentLine == "" {
			currentLine = part
		} else if currentWidth+2+partWidth <= width {
			currentLine += "  " + part
		} else {
			lines = append(lines, currentLine)
			currentLine = part
		}

		if i == len(parts)-1 && currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	return HorizontalLine(width) + "\n" + strings.Join(lines, "\n")
}

// StatusBadge renders a status badge with icon
func StatusBadge(status string) string {
	icon := StatusIcon(status)
	style := GetStatusStyle(status)
	return style.Render(fmt.Sprintf("%s %s", icon, status))
}

// StatusIcon returns the icon for a status
func StatusIcon(status string) string {
	switch status {
	case "pending":
		return "○"
	case "in_progress":
		return "●"
	case "completed":
		return "✓"
	default:
		return "?"
	}
}

// GroupBadge renders a colored group badge
func GroupBadge(name string, color string) string {
	swatch := ColorSwatchStyle(color).Render("██")
	return fmt.Sprintf("%s %s", swatch, name)
}

// CountBadge renders a count badge
func CountBadge(count int) string {
	return MutedStyle.Render(fmt.Sprintf("[%d]", count))
}

// Truncate truncates a string to max length with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// Confirm renders a confirmation dialog
func Confirm(title, message string, confirmKey, cancelKey string) string {
	content := DialogTitleStyle.Render(title) + "\n\n"
	content += message + "\n\n"
	content += fmt.Sprintf("%s %s  %s %s",
		KeyStyle.Render(fmt.Sprintf("[%s]", confirmKey)),
		MutedStyle.Render("Confirm"),
		KeyStyle.Render(fmt.Sprintf("[%s]", cancelKey)),
		MutedStyle.Render("Cancel"),
	)
	return DialogBoxStyle.Render(content)
}

// RenderDropdown renders a dropdown selector
func RenderDropdown(label string, options []string, selected int, focused bool) string {
	var style lipgloss.Style
	if focused {
		style = FocusedInputStyle
	} else {
		style = InputStyle
	}

	selectedText := ""
	if selected >= 0 && selected < len(options) {
		selectedText = options[selected]
	}

	content := fmt.Sprintf("%s ▼", selectedText)
	return fmt.Sprintf("%s\n%s",
		InputLabelStyle.Render(label+":"),
		style.Render(content),
	)
}

// RenderDropdownExpanded renders an expanded dropdown
func RenderDropdownExpanded(label string, options []string, selected int, highlighted int) string {
	var lines []string
	for i, opt := range options {
		prefix := "  "
		if i == selected {
			prefix = "✓ "
		}
		if i == highlighted {
			lines = append(lines, SelectedStyle.Render(prefix+opt))
		} else {
			lines = append(lines, NormalStyle.Render(prefix+opt))
		}
	}

	content := strings.Join(lines, "\n")
	return fmt.Sprintf("%s\n%s",
		InputLabelStyle.Render(label+":"),
		FocusedInputStyle.Render(content),
	)
}

// LabelValue renders a label: value pair
func LabelValue(label, value string) string {
	return fmt.Sprintf("%s %s",
		LabelStyle.Render(label+":"),
		ValueStyle.Render(value),
	)
}

// Section renders a section with title
func Section(title string, content string, width int) string {
	header := MutedStyle.Render(title)
	line := HorizontalLine(width)
	return header + "\n" + line + "\n" + content
}

// Spinner characters for loading animation
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// WordWrap wraps text to fit within width
func WordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) > width {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	textLen := lipgloss.Width(text)
	if textLen >= width {
		return text
	}
	padding := (width - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

// Box renders content in a box
func Box(title, content string, width int) string {
	style := BoxStyle.Width(width)
	if title != "" {
		header := TitleStyle.Render(title)
		return header + "\n" + style.Render(content)
	}
	return style.Render(content)
}
