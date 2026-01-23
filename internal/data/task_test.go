package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTasks(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "cctasks-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test task files
	tasks := []Task{
		{ID: "1", Subject: "Task 1", Status: "pending", Blocks: []string{}, BlockedBy: []string{}},
		{ID: "2", Subject: "Task 2", Status: "in_progress", Blocks: []string{}, BlockedBy: []string{"1"}},
		{ID: "3", Subject: "Task 3", Status: "completed", Blocks: []string{}, BlockedBy: []string{}},
	}

	for _, task := range tasks {
		data, _ := json.MarshalIndent(task, "", "  ")
		filePath := filepath.Join(tmpDir, task.ID+".json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			t.Fatalf("Failed to write task file: %v", err)
		}
	}

	// Test loading
	store, err := loadTasksFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadTasks failed: %v", err)
	}

	if len(store.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(store.Tasks))
	}

	// Verify order (should be sorted by ID)
	if store.Tasks[0].ID != "1" || store.Tasks[1].ID != "2" || store.Tasks[2].ID != "3" {
		t.Errorf("Tasks not sorted by ID")
	}
}

func TestCountTaskFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cctasks-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "1.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "2.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "_groups.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte(""), 0644)

	count := countTaskFiles(tmpDir)
	if count != 2 {
		t.Errorf("Expected 2 task files, got %d", count)
	}
}

func TestTaskStoreAddAndDelete(t *testing.T) {
	store := &TaskStore{
		ProjectName: "test",
		Tasks:       []Task{},
	}

	// Add task
	id := store.AddTask(Task{Subject: "New Task"})
	if id != "1" {
		t.Errorf("Expected ID '1', got '%s'", id)
	}
	if len(store.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(store.Tasks))
	}

	// Add another
	id2 := store.AddTask(Task{Subject: "Another Task"})
	if id2 != "2" {
		t.Errorf("Expected ID '2', got '%s'", id2)
	}

	// Delete first task
	err := store.DeleteTask("1")
	if err != nil {
		t.Errorf("DeleteTask failed: %v", err)
	}
	if len(store.Tasks) != 1 {
		t.Errorf("Expected 1 task after delete, got %d", len(store.Tasks))
	}
}

func TestGetTaskGroup(t *testing.T) {
	task := Task{
		ID:      "1",
		Subject: "Test",
		Metadata: map[string]interface{}{
			"group": "Backend",
		},
	}

	group := GetTaskGroup(task)
	if group != "Backend" {
		t.Errorf("Expected 'Backend', got '%s'", group)
	}

	// Test without metadata
	task2 := Task{ID: "2", Subject: "Test2"}
	group2 := GetTaskGroup(task2)
	if group2 != "" {
		t.Errorf("Expected empty string, got '%s'", group2)
	}
}

func TestSetTaskGroup(t *testing.T) {
	task := Task{ID: "1", Subject: "Test"}

	SetTaskGroup(&task, "Frontend")
	if task.Metadata["group"] != "Frontend" {
		t.Errorf("Expected 'Frontend', got '%v'", task.Metadata["group"])
	}

	// Clear group
	SetTaskGroup(&task, "")
	if _, exists := task.Metadata["group"]; exists {
		t.Error("Expected group to be removed")
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

func TestSearchTasks(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Subject: "Fix login bug", Description: "Authentication issue"},
			{ID: "2", Subject: "Add feature", Description: "New dashboard"},
			{ID: "3", Subject: "Update docs", Description: "Fix typos"},
		},
	}

	// Search by subject
	results := store.SearchTasks("login")
	if len(results) != 1 || results[0].ID != "1" {
		t.Errorf("Expected 1 result for 'login', got %d", len(results))
	}

	// Search by description
	results = store.SearchTasks("dashboard")
	if len(results) != 1 || results[0].ID != "2" {
		t.Errorf("Expected 1 result for 'dashboard', got %d", len(results))
	}

	// Case insensitive
	results = store.SearchTasks("FIX")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'FIX', got %d", len(results))
	}

	// Empty query returns all
	results = store.SearchTasks("")
	if len(results) != 3 {
		t.Errorf("Expected 3 results for empty query, got %d", len(results))
	}
}

func TestGetTasksByStatus(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Status: "pending"},
			{ID: "2", Status: "in_progress"},
			{ID: "3", Status: "completed"},
			{ID: "4", Status: "pending"},
		},
	}

	pending := store.GetTasksByStatus("pending")
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", len(pending))
	}

	all := store.GetTasksByStatus("all")
	if len(all) != 4 {
		t.Errorf("Expected 4 tasks for 'all', got %d", len(all))
	}

	empty := store.GetTasksByStatus("")
	if len(empty) != 4 {
		t.Errorf("Expected 4 tasks for empty filter, got %d", len(empty))
	}
}

func TestGetTasksByGroup(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Metadata: map[string]interface{}{"group": "Backend"}},
			{ID: "2", Metadata: map[string]interface{}{"group": "Frontend"}},
			{ID: "3", Metadata: map[string]interface{}{"group": "Backend"}},
			{ID: "4"}, // No group
		},
	}

	backend := store.GetTasksByGroup("Backend")
	if len(backend) != 2 {
		t.Errorf("Expected 2 Backend tasks, got %d", len(backend))
	}

	all := store.GetTasksByGroup("all")
	if len(all) != 4 {
		t.Errorf("Expected 4 tasks for 'all', got %d", len(all))
	}
}

func TestGetAllGroups(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Metadata: map[string]interface{}{"group": "Backend"}},
			{ID: "2", Metadata: map[string]interface{}{"group": "Frontend"}},
			{ID: "3", Metadata: map[string]interface{}{"group": "Backend"}},
			{ID: "4"}, // No group
		},
	}

	groups := store.GetAllGroups()
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
	// Should be sorted
	if groups[0] != "Backend" || groups[1] != "Frontend" {
		t.Errorf("Expected [Backend, Frontend], got %v", groups)
	}
}

func TestGetTask(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Subject: "Task 1"},
			{ID: "2", Subject: "Task 2"},
		},
	}

	task := store.GetTask("1")
	if task == nil || task.Subject != "Task 1" {
		t.Error("Expected to find Task 1")
	}

	notFound := store.GetTask("999")
	if notFound != nil {
		t.Error("Expected nil for non-existent task")
	}
}

func TestUpdateTask(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Subject: "Old Subject", Status: "pending"},
		},
	}

	err := store.UpdateTask(Task{ID: "1", Subject: "New Subject", Status: "completed"})
	if err != nil {
		t.Errorf("UpdateTask failed: %v", err)
	}
	if store.Tasks[0].Subject != "New Subject" {
		t.Errorf("Expected 'New Subject', got '%s'", store.Tasks[0].Subject)
	}

	err = store.UpdateTask(Task{ID: "999", Subject: "Nonexistent"})
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestDeleteTaskWithDependencies(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Subject: "Task 1", Blocks: []string{"2"}, BlockedBy: []string{}},
			{ID: "2", Subject: "Task 2", Blocks: []string{}, BlockedBy: []string{"1"}},
		},
	}

	// Delete task 1, should remove from task 2's blockedBy
	err := store.DeleteTask("1")
	if err != nil {
		t.Errorf("DeleteTask failed: %v", err)
	}

	if len(store.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(store.Tasks))
	}

	// Task 2 should no longer be blocked by 1
	if len(store.Tasks[0].BlockedBy) != 0 {
		t.Errorf("Expected empty BlockedBy, got %v", store.Tasks[0].BlockedBy)
	}
}

func TestDeleteTaskNonExistent(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "1", Subject: "Task 1"},
		},
	}

	err := store.DeleteTask("999")
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestAddTaskDefaults(t *testing.T) {
	store := &TaskStore{
		ProjectName: "test",
		Tasks:       []Task{},
	}

	// Add task without status, blocks, blockedBy
	id := store.AddTask(Task{Subject: "New Task"})

	task := store.GetTask(id)
	if task.Status != "pending" {
		t.Errorf("Expected default status 'pending', got '%s'", task.Status)
	}
	if task.Blocks == nil {
		t.Error("Expected Blocks to be initialized")
	}
	if task.BlockedBy == nil {
		t.Error("Expected BlockedBy to be initialized")
	}
}

func TestGenerateID(t *testing.T) {
	store := &TaskStore{
		Tasks: []Task{
			{ID: "5"},
			{ID: "3"},
			{ID: "10"},
		},
	}

	id := store.generateID()
	if id != "11" {
		t.Errorf("Expected ID '11', got '%s'", id)
	}

	// Empty store should generate "1"
	emptyStore := &TaskStore{Tasks: []Task{}}
	id = emptyStore.generateID()
	if id != "1" {
		t.Errorf("Expected ID '1' for empty store, got '%s'", id)
	}
}

func TestRemoveFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c", "b", "d"}
	result := removeFromSlice(slice, "b")

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}

	for _, item := range result {
		if item == "b" {
			t.Error("Expected 'b' to be removed")
		}
	}
}

// Helper function to load tasks from a specific directory (for testing)
func loadTasksFromDir(dir string) (*TaskStore, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name[0] == '_' || filepath.Ext(name) != ".json" {
			continue
		}

		filePath := filepath.Join(dir, name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return &TaskStore{Tasks: tasks}, nil
}
