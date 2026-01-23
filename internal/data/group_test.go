package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGroups(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cctasks-group-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create _groups.json
	gf := groupsFile{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1, Color: "#8b5cf6"},
			{Name: "Frontend", Order: 2, Color: "#3b82f6"},
		},
	}
	data, _ := json.MarshalIndent(gf, "", "  ")
	os.WriteFile(filepath.Join(tmpDir, "_groups.json"), data, 0644)

	// Test loading
	store, err := loadGroupsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadGroups failed: %v", err)
	}

	if len(store.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(store.Groups))
	}

	if store.Groups[0].Name != "Backend" {
		t.Errorf("Expected 'Backend' first, got '%s'", store.Groups[0].Name)
	}
}

func TestLoadGroupsEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cctasks-group-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// No _groups.json file
	store, err := loadGroupsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadGroups failed: %v", err)
	}

	if len(store.Groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(store.Groups))
	}
}

func TestGroupStoreAddGroup(t *testing.T) {
	store := &GroupStore{
		ProjectName: "test",
		Groups:      []TaskGroup{},
	}

	store.AddGroup(TaskGroup{Name: "Backend"})
	if len(store.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(store.Groups))
	}
	if store.Groups[0].Order != 1 {
		t.Errorf("Expected order 1, got %d", store.Groups[0].Order)
	}
	if store.Groups[0].Color == "" {
		t.Error("Expected default color to be set")
	}

	store.AddGroup(TaskGroup{Name: "Frontend", Color: "#ff0000"})
	if store.Groups[1].Order != 2 {
		t.Errorf("Expected order 2, got %d", store.Groups[1].Order)
	}
	if store.Groups[1].Color != "#ff0000" {
		t.Errorf("Expected custom color, got '%s'", store.Groups[1].Color)
	}
}

func TestGroupStoreGetGroup(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1, Color: "#8b5cf6"},
		},
	}

	group := store.GetGroup("Backend")
	if group == nil {
		t.Fatal("Expected to find Backend group")
	}
	if group.Color != "#8b5cf6" {
		t.Errorf("Expected color '#8b5cf6', got '%s'", group.Color)
	}

	notFound := store.GetGroup("NonExistent")
	if notFound != nil {
		t.Error("Expected nil for non-existent group")
	}
}

func TestGroupStoreUpdateGroup(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1, Color: "#8b5cf6"},
		},
	}

	updated := store.UpdateGroup("Backend", TaskGroup{Name: "Backend", Order: 1, Color: "#ff0000"})
	if !updated {
		t.Error("Expected update to succeed")
	}
	if store.Groups[0].Color != "#ff0000" {
		t.Errorf("Expected color '#ff0000', got '%s'", store.Groups[0].Color)
	}

	notUpdated := store.UpdateGroup("NonExistent", TaskGroup{Name: "NonExistent"})
	if notUpdated {
		t.Error("Expected update to fail for non-existent group")
	}
}

func TestGroupStoreDeleteGroup(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1},
			{Name: "Frontend", Order: 2},
		},
	}

	deleted := store.DeleteGroup("Backend")
	if !deleted {
		t.Error("Expected delete to succeed")
	}
	if len(store.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(store.Groups))
	}
	if store.Groups[0].Name != "Frontend" {
		t.Errorf("Expected 'Frontend' to remain, got '%s'", store.Groups[0].Name)
	}

	notDeleted := store.DeleteGroup("NonExistent")
	if notDeleted {
		t.Error("Expected delete to fail for non-existent group")
	}
}

func TestGroupStoreMoveGroup(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1},
			{Name: "Frontend", Order: 2},
			{Name: "DevOps", Order: 3},
		},
	}

	// Move Frontend up
	moved := store.MoveGroupUp("Frontend")
	if !moved {
		t.Error("Expected move up to succeed")
	}
	if store.Groups[0].Name != "Frontend" {
		t.Errorf("Expected Frontend first, got '%s'", store.Groups[0].Name)
	}

	// Can't move first item up
	notMoved := store.MoveGroupUp("Frontend")
	if notMoved {
		t.Error("Expected move up to fail for first item")
	}

	// Move Backend down
	moved = store.MoveGroupDown("Backend")
	if !moved {
		t.Error("Expected move down to succeed")
	}

	// Can't move last item down
	notMoved = store.MoveGroupDown("Backend")
	if notMoved {
		t.Error("Expected move down to fail for last item")
	}
}

func TestGroupStoreGetGroupNames(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Order: 1},
			{Name: "Frontend", Order: 2},
		},
	}

	names := store.GetGroupNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "Backend" || names[1] != "Frontend" {
		t.Errorf("Unexpected names: %v", names)
	}
}

func TestGroupStoreGetGroupColor(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{
			{Name: "Backend", Color: "#8b5cf6"},
		},
	}

	color := store.GetGroupColor("Backend")
	if color != "#8b5cf6" {
		t.Errorf("Expected '#8b5cf6', got '%s'", color)
	}

	defaultColor := store.GetGroupColor("NonExistent")
	if defaultColor != "#6b7280" {
		t.Errorf("Expected default gray color, got '%s'", defaultColor)
	}
}

func TestGroupStoreEnsureGroupExists(t *testing.T) {
	store := &GroupStore{
		Groups: []TaskGroup{},
	}

	store.EnsureGroupExists("Backend")
	if len(store.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(store.Groups))
	}

	// Should not add duplicate
	store.EnsureGroupExists("Backend")
	if len(store.Groups) != 1 {
		t.Errorf("Expected still 1 group, got %d", len(store.Groups))
	}

	// Empty name should be ignored
	store.EnsureGroupExists("")
	if len(store.Groups) != 1 {
		t.Errorf("Expected still 1 group, got %d", len(store.Groups))
	}
}

// Helper function to load groups from a specific directory (for testing)
func loadGroupsFromDir(dir string) (*GroupStore, error) {
	filePath := filepath.Join(dir, "_groups.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GroupStore{Groups: []TaskGroup{}}, nil
		}
		return nil, err
	}

	var gf groupsFile
	if err := json.Unmarshal(data, &gf); err != nil {
		return nil, err
	}

	return &GroupStore{Groups: gf.Groups}, nil
}
