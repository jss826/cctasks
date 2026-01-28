package model

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
)

func setupTestTasks(t *testing.T) (*data.TaskStore, *data.GroupStore, string) {
	tmpDir, err := os.MkdirTemp("", "cctasks-tasks-test-*")
	if err != nil {
		t.Fatal(err)
	}

	tasks := []data.Task{
		{ID: "1", Subject: "Task 1", Status: "pending", Blocks: []string{}, BlockedBy: []string{}, Metadata: map[string]interface{}{"group": "Backend"}},
		{ID: "2", Subject: "Task 2", Status: "in_progress", Blocks: []string{}, BlockedBy: []string{}, Metadata: map[string]interface{}{"group": "Frontend"}},
		{ID: "3", Subject: "Task 3", Status: "completed", Blocks: []string{}, BlockedBy: []string{}, Metadata: map[string]interface{}{"group": "Backend"}},
		{ID: "4", Subject: "Task 4", Status: "pending", Blocks: []string{}, BlockedBy: []string{}},
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

func TestTasksModel_Navigation(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Initial cursor should be at 0
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", m.cursor)
	}

	// Move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Can't go above 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestTasksModel_JKNavigation(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// j moves down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1 after 'j', got %d", m.cursor)
	}

	// k moves up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after 'k', got %d", m.cursor)
	}
}

func TestTasksModel_HomeEnd(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Get visible items count
	itemCount := len(m.items)
	if itemCount == 0 {
		t.Skip("No items in list")
	}

	// End goes to last
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if m.cursor != itemCount-1 {
		t.Errorf("Expected cursor at %d after End, got %d", itemCount-1, m.cursor)
	}

	// Home goes to first
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0 after Home, got %d", m.cursor)
	}
}

func TestTasksModel_StatusFilter(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Initial status filter is "all" (empty string represents all)
	initialFilter := m.statusFilter

	// Press f to cycle status filter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if m.statusFilter == initialFilter {
		t.Error("Expected statusFilter to change after 'f'")
	}

	// Keep cycling and verify it changes
	prevFilter := m.statusFilter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if m.statusFilter == prevFilter {
		t.Error("Expected statusFilter to change on second 'f'")
	}

	prevFilter = m.statusFilter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if m.statusFilter == prevFilter {
		t.Error("Expected statusFilter to change on third 'f'")
	}

	// Fourth press should cycle back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if m.statusFilter != initialFilter {
		t.Errorf("Expected statusFilter to cycle back to initial '%s', got '%s'", initialFilter, m.statusFilter)
	}
}

func TestTasksModel_GroupFilter(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Get initial group filter
	initialFilter := m.groupFilter

	// Press g to cycle group filter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	// Should cycle to a different value
	if m.groupFilter == initialFilter {
		t.Error("Expected groupFilter to change after 'g'")
	}
}

func TestTasksModel_HideCompleted(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Initial hideCompleted is true (default)
	if !m.hideCompleted {
		t.Error("Expected initial hideCompleted to be true")
	}

	// Expand all groups first to see the tasks
	for groupName := range m.collapsedGroups {
		m.collapsedGroups[groupName] = false
	}
	m.rebuildItems()

	initialCount := len(m.items)

	// Press h to toggle (show completed)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if m.hideCompleted {
		t.Error("Expected hideCompleted to be false after 'h'")
	}

	// Should have more visible items (we have 1 completed task)
	newCount := len(m.items)
	if newCount <= initialCount {
		t.Errorf("Expected more items after showing completed, got %d (was %d)", newCount, initialCount)
	}

	// Toggle back to hide
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if !m.hideCompleted {
		t.Error("Expected hideCompleted to be true after toggling again")
	}
}

func TestTasksModel_SortMode(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Get initial sortMode
	initialMode := m.sortMode

	// Press o to cycle sort mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if m.sortMode == initialMode {
		t.Error("Expected sortMode to change after 'o'")
	}

	// Press o again to cycle back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if m.sortMode != initialMode {
		t.Errorf("Expected sortMode to cycle back to initial '%s', got '%s'", initialMode, m.sortMode)
	}
}

func TestTasksModel_Search(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Press / to start search
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searchActive {
		t.Error("Expected searchActive to be true after '/'")
	}

	// Type search query
	m.searchInput.SetValue("Task 1")

	// Escape to exit search
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.searchActive {
		t.Error("Expected searchActive to be false after Esc")
	}
}

func TestTasksModel_ToggleGroup(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Find a group header and toggle it
	for i, item := range m.items {
		if item.isGroup {
			m.cursor = i
			initialCount := len(m.items)

			// Group should be collapsed initially (default)
			if !m.collapsedGroups[item.groupName] {
				t.Error("Expected group to be collapsed initially")
			}

			// Press Enter to toggle (expand)
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Group should now be expanded
			if m.collapsedGroups[item.groupName] {
				t.Error("Expected group to be expanded after Enter")
			}

			// Should have more items visible
			newCount := len(m.items)
			if newCount <= initialCount {
				t.Errorf("Expected more items after expand, got %d (was %d)", newCount, initialCount)
			}

			// Toggle back to collapse
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if !m.collapsedGroups[item.groupName] {
				t.Error("Expected group to be collapsed after second Enter")
			}

			break
		}
	}
}

func TestTasksModel_Items(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Groups are collapsed by default, so we should have group headers
	hasGroups := false
	for _, item := range m.items {
		if item.isGroup {
			hasGroups = true
			break
		}
	}

	if !hasGroups {
		t.Error("Expected to have group items")
	}

	// Expand all groups to verify tasks exist
	for groupName := range m.collapsedGroups {
		m.collapsedGroups[groupName] = false
	}
	m.rebuildItems()

	hasTasks := false
	for _, item := range m.items {
		if !item.isGroup {
			hasTasks = true
			break
		}
	}

	if !hasTasks {
		t.Error("Expected to have task items after expanding groups")
	}
}

func TestTasksModel_QuickStatusChange(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Expand all groups first to see tasks
	for groupName := range m.collapsedGroups {
		m.collapsedGroups[groupName] = false
	}
	m.rebuildItems()

	// Find first task (not a group)
	foundTask := false
	for i, item := range m.items {
		if !item.isGroup && item.task != nil {
			m.cursor = i
			foundTask = true
			taskID := item.task.ID

			// Press s to enter status change mode
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
			if !m.statusChangeMode {
				t.Error("Expected statusChangeMode to be true after 's'")
			}

			// Press 'i' to set to in_progress
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
			if m.statusChangeMode {
				t.Error("Expected statusChangeMode to be false after selecting status")
			}

			// Verify status changed in taskStore
			task := m.taskStore.GetTask(taskID)
			if task != nil && task.Status != "in_progress" {
				t.Errorf("Expected status 'in_progress', got '%s'", task.Status)
			}
			break
		}
	}

	if !foundTask {
		t.Skip("No task items found")
	}
}

func TestTasksModel_QuickStatusChangeCancel(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	// Expand all groups first to see tasks
	for groupName := range m.collapsedGroups {
		m.collapsedGroups[groupName] = false
	}
	m.rebuildItems()

	// Find first task
	for i, item := range m.items {
		if !item.isGroup && item.task != nil {
			m.cursor = i
			originalStatus := item.task.Status

			// Enter status change mode
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
			if !m.statusChangeMode {
				t.Error("Expected statusChangeMode to be true")
			}

			// Cancel with Esc
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			if m.statusChangeMode {
				t.Error("Expected statusChangeMode to be false after Esc")
			}

			// Status should not have changed
			task := m.taskStore.GetTask(item.task.ID)
			if task != nil && task.Status != originalStatus {
				t.Errorf("Expected status to remain '%s', got '%s'", originalStatus, task.Status)
			}
			break
		}
	}
}

func TestTasksModel_ViewOutput(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	view := m.View()

	// View should not be empty
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain project name
	if !containsStr(view, "test") {
		t.Error("Expected view to contain project name")
	}
}

func TestTasksModel_CursorBounds(t *testing.T) {
	taskStore, groupStore, tmpDir := setupTestTasks(t)
	defer os.RemoveAll(tmpDir)

	m := NewTasksModel("test", taskStore, groupStore)
	m.width = 80
	m.height = 24

	itemCount := len(m.items)
	if itemCount == 0 {
		t.Skip("No items")
	}

	// Go to end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if m.cursor != itemCount-1 {
		t.Errorf("Expected cursor at end (%d), got %d", itemCount-1, m.cursor)
	}

	// Try to go beyond end - should stay at end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != itemCount-1 {
		t.Errorf("Expected cursor to stay at end (%d), got %d", itemCount-1, m.cursor)
	}

	// Go to start
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	if m.cursor != 0 {
		t.Errorf("Expected cursor at start (0), got %d", m.cursor)
	}

	// Try to go before start - should stay at start
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("Expected cursor to stay at start (0), got %d", m.cursor)
	}
}

// Helper function
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
