package model

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
)

func setupTestEdit(t *testing.T) (*data.TaskStore, *data.GroupStore, string) {
	tmpDir, err := os.MkdirTemp("", "cctasks-edit-test-*")
	if err != nil {
		t.Fatal(err)
	}

	tasks := []data.Task{
		{ID: "1", Subject: "Task 1", Status: "pending", Blocks: []string{}, BlockedBy: []string{}},
		{ID: "2", Subject: "Task 2", Status: "in_progress", Blocks: []string{}, BlockedBy: []string{}},
		{ID: "3", Subject: "Task 3", Status: "completed", Blocks: []string{"1"}, BlockedBy: []string{"2"}},
	}

	groups := []data.TaskGroup{
		{Name: "Backend", Order: 0, Color: "#8b5cf6"},
		{Name: "Frontend", Order: 1, Color: "#3b82f6"},
	}

	taskStore, err := data.NewTaskStoreForTest(tmpDir, tasks)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	groupStore, err := data.NewGroupStoreForTest(tmpDir, groups)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	return taskStore, groupStore, tmpDir
}

func TestParseTaskIDs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"   ", []string{}},
		{"1", []string{"1"}},
		{"1,2,3", []string{"1", "2", "3"}},
		{"1, 2, 3", []string{"1", "2", "3"}},
		{"  1  ,  2  ,  3  ", []string{"1", "2", "3"}},
		{"1,,2", []string{"1", "2"}},
		{",1,2,", []string{"1", "2"}},
	}

	for _, tt := range tests {
		result := parseTaskIDs(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parseTaskIDs(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, id := range result {
			if id != tt.expected[i] {
				t.Errorf("parseTaskIDs(%q)[%d] = %q, want %q", tt.input, i, id, tt.expected[i])
			}
		}
	}
}

func TestEditModel_NewTask(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Should be new task mode
	if !m.isNew {
		t.Error("Expected isNew to be true")
	}

	// Default status should be pending
	if m.statuses[m.statusIdx] != "pending" {
		t.Errorf("Expected default status 'pending', got '%s'", m.statuses[m.statusIdx])
	}

	// Subject input should be focused
	if !m.subjectInput.Focused() {
		t.Error("Expected subject input to be focused")
	}
}

func TestEditModel_EditExistingTask(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	task := &data.Task{
		ID:          "3",
		Subject:     "Task 3",
		Description: "Description",
		Status:      "completed",
		Owner:       "john",
		Blocks:      []string{"1"},
		BlockedBy:   []string{"2"},
	}

	m := NewEditModel(task, taskStore, groupStore, false)

	// Should not be new task mode
	if m.isNew {
		t.Error("Expected isNew to be false")
	}

	// Values should be populated
	if m.subjectInput.Value() != "Task 3" {
		t.Errorf("Expected subject 'Task 3', got '%s'", m.subjectInput.Value())
	}

	if m.ownerInput.Value() != "john" {
		t.Errorf("Expected owner 'john', got '%s'", m.ownerInput.Value())
	}

	if m.blocksInput.Value() != "1" {
		t.Errorf("Expected blocks '1', got '%s'", m.blocksInput.Value())
	}

	if m.blockedByInput.Value() != "2" {
		t.Errorf("Expected blockedBy '2', got '%s'", m.blockedByInput.Value())
	}
}

func TestEditModel_TabNavigation(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Initial focus is on subject (0)
	if m.focusIdx != 0 {
		t.Errorf("Expected initial focusIdx 0, got %d", m.focusIdx)
	}

	// Tab should move to next field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIdx != 1 {
		t.Errorf("Expected focusIdx 1 after Tab, got %d", m.focusIdx)
	}

	// Continue tabbing through all fields
	for i := 2; i <= 6; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		if m.focusIdx != i {
			t.Errorf("Expected focusIdx %d after Tab, got %d", i, m.focusIdx)
		}
	}

	// Tab from last field should wrap to first
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focusIdx != 0 {
		t.Errorf("Expected focusIdx 0 after wrapping, got %d", m.focusIdx)
	}
}

func TestEditModel_ShiftTabNavigation(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Shift+Tab from first field should wrap to last
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focusIdx != 6 {
		t.Errorf("Expected focusIdx 6 after Shift+Tab from 0, got %d", m.focusIdx)
	}
}

func TestEditModel_StatusSelector(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Navigate to status field (index 2)
	m.focusIdx = 2

	// Initial status is pending (index 0)
	if m.statusIdx != 0 {
		t.Errorf("Expected initial statusIdx 0, got %d", m.statusIdx)
	}

	// Down arrow should change status
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.statusIdx != 1 {
		t.Errorf("Expected statusIdx 1 after Down, got %d", m.statusIdx)
	}

	// Up arrow should change status back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.statusIdx != 0 {
		t.Errorf("Expected statusIdx 0 after Up, got %d", m.statusIdx)
	}

	// Up arrow at 0 should stay at 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.statusIdx != 0 {
		t.Errorf("Expected statusIdx to stay at 0, got %d", m.statusIdx)
	}
}

func TestEditModel_GroupSelector(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Navigate to group field (index 3)
	m.focusIdx = 3

	// Initial group is none (index 0)
	if m.groupIdx != 0 {
		t.Errorf("Expected initial groupIdx 0, got %d", m.groupIdx)
	}

	// Down arrow should change group
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.groupIdx != 1 {
		t.Errorf("Expected groupIdx 1 after Down, got %d", m.groupIdx)
	}
}

func TestEditModel_OpenPicker(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Navigate to blocks field (index 5)
	m.focusIdx = 5

	// Press / to open picker
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})

	if !m.pickerActive {
		t.Error("Expected picker to be active after '/'")
	}

	if m.pickerForField != 5 {
		t.Errorf("Expected pickerForField 5, got %d", m.pickerForField)
	}
}

func TestEditModel_PickerNavigation(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Navigate to blockedBy field and open picker
	m.focusIdx = 6
	m.openPicker(6)

	// Should have tasks in picker (excluding self, but this is new so all tasks)
	if len(m.pickerTasks) == 0 {
		t.Error("Expected picker to have tasks")
	}

	// Initial cursor should be 0
	if m.pickerCursor != 0 {
		t.Errorf("Expected pickerCursor 0, got %d", m.pickerCursor)
	}

	// Down should move cursor
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.pickerCursor != 1 && len(m.pickerTasks) > 1 {
		t.Errorf("Expected pickerCursor 1 after Down, got %d", m.pickerCursor)
	}

	// Up should move cursor back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.pickerCursor != 0 {
		t.Errorf("Expected pickerCursor 0 after Up, got %d", m.pickerCursor)
	}
}

func TestEditModel_PickerToggleSelection(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Open picker
	m.focusIdx = 5
	m.openPicker(5)

	if len(m.pickerTasks) == 0 {
		t.Skip("No tasks in picker")
	}

	taskID := m.pickerTasks[0].ID

	// Initially not selected
	if m.pickerSelected[taskID] {
		t.Error("Expected task to not be selected initially")
	}

	// Enter should toggle selection
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.pickerSelected[taskID] {
		t.Error("Expected task to be selected after Enter")
	}

	// Enter again should deselect
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.pickerSelected[taskID] {
		t.Error("Expected task to be deselected after second Enter")
	}
}

func TestEditModel_PickerConfirm(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Open picker for blocks
	m.focusIdx = 5
	m.openPicker(5)

	if len(m.pickerTasks) == 0 {
		t.Skip("No tasks in picker")
	}

	// Select first task
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Tab to confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Picker should be closed
	if m.pickerActive {
		t.Error("Expected picker to be closed after Tab")
	}

	// Blocks input should have the selected task ID
	if m.blocksInput.Value() == "" {
		t.Error("Expected blocksInput to have selected task ID")
	}
}

func TestEditModel_PickerCancel(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Set initial value
	m.blocksInput.SetValue("1, 2")

	// Open picker
	m.focusIdx = 5
	m.openPicker(5)

	// Select a different task
	if len(m.pickerTasks) > 0 {
		m.pickerCursor = len(m.pickerTasks) - 1
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	}

	// Escape to cancel
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Picker should be closed
	if m.pickerActive {
		t.Error("Expected picker to be closed after Esc")
	}

	// Original value should be preserved
	if m.blocksInput.Value() != "1, 2" {
		t.Errorf("Expected blocksInput to preserve original value, got '%s'", m.blocksInput.Value())
	}
}

func TestEditModel_PickerSearch(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Open picker
	m.focusIdx = 5
	m.openPicker(5)

	initialCount := len(m.pickerTasks)
	if initialCount == 0 {
		t.Skip("No tasks in picker")
	}

	// Type search query
	m.pickerSearch.SetValue("Task 1")
	m.filterPickerTasks()

	// Should have filtered results
	if len(m.pickerTasks) >= initialCount && initialCount > 1 {
		t.Error("Expected search to filter tasks")
	}
}

func TestEditModel_CancelEdit(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)

	// Press Escape
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should return cancel message
	if cmd == nil {
		t.Error("Expected a command from Esc")
	}

	msg := cmd()
	if _, ok := msg.(CancelEditMsg); !ok {
		t.Error("Expected CancelEditMsg from Esc")
	}
}

func TestEditModel_View(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)
	m.width = 80
	m.height = 24

	view := m.View()

	// View should contain expected elements
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain field labels
	expectedLabels := []string{"Subject:", "Description:", "Status:", "Group:", "Owner:", "Blocks:", "Blocked By:"}
	for _, label := range expectedLabels {
		if !containsString(view, label) {
			t.Errorf("Expected view to contain '%s'", label)
		}
	}
}

func TestEditModel_PickerView(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestEdit(t)
	defer os.RemoveAll(tmpDir)

	m := NewEditModel(nil, taskStore, groupStore, true)
	m.width = 80
	m.height = 24

	// Open picker
	m.focusIdx = 5
	m.openPicker(5)

	view := m.View()

	// Should show picker view
	if !containsString(view, "Select Tasks") {
		t.Error("Expected picker view to contain 'Select Tasks'")
	}

	if !containsString(view, "Search:") {
		t.Error("Expected picker view to contain 'Search:'")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
