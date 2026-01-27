package ui

import (
	"strings"
	"testing"
)

func TestFooter(t *testing.T) {
	keys := [][]string{
		{"a", "Action A"},
		{"b", "Action B"},
	}

	result := Footer(keys, 80)

	if !strings.Contains(result, "[a]") {
		t.Error("Expected footer to contain [a]")
	}
	if !strings.Contains(result, "[b]") {
		t.Error("Expected footer to contain [b]")
	}
	if !strings.Contains(result, "Action A") {
		t.Error("Expected footer to contain 'Action A'")
	}
}

func TestFooterWithHints(t *testing.T) {
	hints := []KeyHint{
		{Key: "a", Desc: "Enabled", Enabled: true},
		{Key: "b", Desc: "Disabled", Enabled: false},
	}

	result := FooterWithHints(hints, 80)

	if !strings.Contains(result, "[a]") {
		t.Error("Expected footer to contain [a]")
	}
	if !strings.Contains(result, "[b]") {
		t.Error("Expected footer to contain [b]")
	}
	if !strings.Contains(result, "Enabled") {
		t.Error("Expected footer to contain 'Enabled'")
	}
	if !strings.Contains(result, "Disabled") {
		t.Error("Expected footer to contain 'Disabled'")
	}
}

func TestStatusIcon(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"pending", "○"},
		{"in_progress", "●"},
		{"completed", "✓"},
		{"unknown", "?"},
	}

	for _, tt := range tests {
		result := StatusIcon(tt.status)
		if result != tt.expected {
			t.Errorf("StatusIcon(%s) = %s, want %s", tt.status, result, tt.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"ab", 2, "ab"},
		{"abc", 2, "ab"},
	}

	for _, tt := range tests {
		result := Truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestStatusBadge(t *testing.T) {
	result := StatusBadge("pending")
	if !strings.Contains(result, "pending") {
		t.Error("Expected StatusBadge to contain 'pending'")
	}
	if !strings.Contains(result, "○") {
		t.Error("Expected StatusBadge to contain pending icon")
	}
}

func TestGroupBadge(t *testing.T) {
	result := GroupBadge("Backend", "#8b5cf6")
	if !strings.Contains(result, "Backend") {
		t.Error("Expected GroupBadge to contain 'Backend'")
	}
}

func TestWordWrap(t *testing.T) {
	text := "This is a long line that should be wrapped"
	result := WordWrap(text, 20)

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if len(line) > 20 {
			t.Errorf("Line too long: %q (len=%d)", line, len(line))
		}
	}
}

func TestHorizontalLine(t *testing.T) {
	result := HorizontalLine(10)
	if len(result) == 0 {
		t.Error("Expected non-empty horizontal line")
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected string
	}{
		{"hi", 6, "  hi"},
		{"hello", 5, "hello"},
		{"ab", 10, "    ab"},
		{"", 4, "  "},
	}

	for _, tt := range tests {
		result := CenterText(tt.text, tt.width)
		if result != tt.expected {
			t.Errorf("CenterText(%q, %d) = %q, want %q", tt.text, tt.width, result, tt.expected)
		}
	}
}

func TestCenterPad(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected string
	}{
		{"hi", 6, "  hi  "},
		{"hello", 5, "hello"},
		{"ab", 10, "    ab    "},
		{"x", 4, " x  "},
	}

	for _, tt := range tests {
		result := CenterPad(tt.text, tt.width)
		if result != tt.expected {
			t.Errorf("CenterPad(%q, %d) = %q, want %q", tt.text, tt.width, result, tt.expected)
		}
		// Verify length matches width
		if len(result) != tt.width && len(tt.text) < tt.width {
			t.Errorf("CenterPad result length = %d, want %d", len(result), tt.width)
		}
	}
}

func TestCountBadge(t *testing.T) {
	result := CountBadge(5)
	if !strings.Contains(result, "5") {
		t.Error("Expected CountBadge to contain the count")
	}
	if !strings.Contains(result, "[") || !strings.Contains(result, "]") {
		t.Error("Expected CountBadge to have brackets")
	}
}

func TestLabelValue(t *testing.T) {
	result := LabelValue("Name", "John")
	if !strings.Contains(result, "Name:") {
		t.Error("Expected LabelValue to contain label")
	}
	if !strings.Contains(result, "John") {
		t.Error("Expected LabelValue to contain value")
	}
}

func TestSection(t *testing.T) {
	result := Section("Title", "Content here", 40)
	if !strings.Contains(result, "Title") {
		t.Error("Expected Section to contain title")
	}
	if !strings.Contains(result, "Content here") {
		t.Error("Expected Section to contain content")
	}
}

func TestWordWrapMultiline(t *testing.T) {
	text := "Line one\nLine two is longer than expected"
	result := WordWrap(text, 15)

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if len(line) > 15 {
			t.Errorf("Line too long after wrap: %q (len=%d)", line, len(line))
		}
	}
}

func TestWordWrapZeroWidth(t *testing.T) {
	text := "Some text"
	result := WordWrap(text, 0)
	if result != text {
		t.Errorf("WordWrap with 0 width should return original, got %q", result)
	}
}

func TestRenderDropdown(t *testing.T) {
	options := []string{"Option A", "Option B", "Option C"}

	result := RenderDropdown("Label", options, 1, true)
	if !strings.Contains(result, "Label:") {
		t.Error("Expected dropdown to contain label")
	}
	if !strings.Contains(result, "Option B") {
		t.Error("Expected dropdown to contain selected option")
	}
	if !strings.Contains(result, "▼") {
		t.Error("Expected dropdown to contain arrow")
	}
}

func TestRenderDropdownExpanded(t *testing.T) {
	options := []string{"A", "B", "C"}

	result := RenderDropdownExpanded("Label", options, 1, 2)
	if !strings.Contains(result, "Label:") {
		t.Error("Expected expanded dropdown to contain label")
	}
	if !strings.Contains(result, "A") || !strings.Contains(result, "B") || !strings.Contains(result, "C") {
		t.Error("Expected expanded dropdown to contain all options")
	}
}

func TestConfirm(t *testing.T) {
	result := Confirm("Delete?", "Are you sure?", "y", "n")
	if !strings.Contains(result, "Delete?") {
		t.Error("Expected confirm to contain title")
	}
	if !strings.Contains(result, "Are you sure?") {
		t.Error("Expected confirm to contain message")
	}
	if !strings.Contains(result, "[y]") || !strings.Contains(result, "[n]") {
		t.Error("Expected confirm to contain key hints")
	}
}

func TestBox(t *testing.T) {
	result := Box("Title", "Content", 40)
	if !strings.Contains(result, "Title") {
		t.Error("Expected box to contain title")
	}
	if !strings.Contains(result, "Content") {
		t.Error("Expected box to contain content")
	}

	// Test without title
	resultNoTitle := Box("", "Content only", 40)
	if !strings.Contains(resultNoTitle, "Content only") {
		t.Error("Expected box without title to contain content")
	}
}

func TestHeader(t *testing.T) {
	result := Header("My App", 40)
	if !strings.Contains(result, "My App") {
		t.Error("Expected header to contain title")
	}
}
