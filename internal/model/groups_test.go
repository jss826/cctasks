package model

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jss826/cctasks/internal/data"
)

func setupTestGroups(t *testing.T) (*data.GroupStore, string) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cctasks-test-*")
	if err != nil {
		t.Fatal(err)
	}

	groups := []data.TaskGroup{
		{Name: "Group1", Order: 0, Color: "#ff0000"},
		{Name: "Group2", Order: 1, Color: "#00ff00"},
		{Name: "Group3", Order: 2, Color: "#0000ff"},
	}

	store, err := data.NewGroupStoreForTest(tmpDir, groups)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	return store, tmpDir
}

func TestGroupsModel_MoveDown(t *testing.T) {
	store, tmpDir := setupTestGroups(t)
	defer os.RemoveAll(tmpDir)

	m := NewGroupsModel(store)
	m.cursor = 0 // Start at Group1

	// Press J to move Group1 down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})

	// Verify: Group1 should now be at index 1
	if m.groupStore.Groups[0].Name != "Group2" {
		t.Errorf("Expected Group2 at index 0, got %s", m.groupStore.Groups[0].Name)
	}
	if m.groupStore.Groups[1].Name != "Group1" {
		t.Errorf("Expected Group1 at index 1, got %s", m.groupStore.Groups[1].Name)
	}
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", m.cursor)
	}

	// Press J again to move Group1 further down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})

	// Verify: Group1 should now be at index 2
	if m.groupStore.Groups[2].Name != "Group1" {
		t.Errorf("Expected Group1 at index 2, got %s", m.groupStore.Groups[2].Name)
	}
	if m.cursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", m.cursor)
	}

	// Press J again - should not move (already at bottom)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})

	// Verify: nothing should change
	if m.groupStore.Groups[2].Name != "Group1" {
		t.Errorf("Expected Group1 to stay at index 2, got %s", m.groupStore.Groups[2].Name)
	}
	if m.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", m.cursor)
	}
}

func TestGroupsModel_MoveUp(t *testing.T) {
	store, tmpDir := setupTestGroups(t)
	defer os.RemoveAll(tmpDir)

	m := NewGroupsModel(store)
	m.cursor = 2 // Start at Group3 (bottom)

	// Press K to move Group3 up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})

	// Verify: Group3 should now be at index 1
	if m.groupStore.Groups[1].Name != "Group3" {
		t.Errorf("Expected Group3 at index 1, got %s", m.groupStore.Groups[1].Name)
	}
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", m.cursor)
	}

	// Press K again to move Group3 to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})

	// Verify: Group3 should now be at index 0
	if m.groupStore.Groups[0].Name != "Group3" {
		t.Errorf("Expected Group3 at index 0, got %s", m.groupStore.Groups[0].Name)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	// Press K again - should not move (already at top)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})

	// Verify: nothing should change
	if m.groupStore.Groups[0].Name != "Group3" {
		t.Errorf("Expected Group3 to stay at index 0, got %s", m.groupStore.Groups[0].Name)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestGroupsModel_MoveDownThenUp(t *testing.T) {
	store, tmpDir := setupTestGroups(t)
	defer os.RemoveAll(tmpDir)

	m := NewGroupsModel(store)
	m.cursor = 0 // Start at Group1

	// Move Group1 down twice
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})

	// Group1 should be at bottom
	if m.groupStore.Groups[2].Name != "Group1" {
		t.Errorf("Expected Group1 at index 2, got %s", m.groupStore.Groups[2].Name)
	}

	// Move Group1 back up twice
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})

	// Group1 should be back at top
	if m.groupStore.Groups[0].Name != "Group1" {
		t.Errorf("Expected Group1 at index 0, got %s", m.groupStore.Groups[0].Name)
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}
}
