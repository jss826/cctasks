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
