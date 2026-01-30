package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jss826/cctasks/internal/config"
)

// Task represents a task item
type Task struct {
	ID          string                 `json:"id"`
	Subject     string                 `json:"subject"`
	Description string                 `json:"description"`
	ActiveForm  string                 `json:"activeForm,omitempty"`
	Status      string                 `json:"status"` // pending, in_progress, completed
	Blocks      []string               `json:"blocks"`
	BlockedBy   []string               `json:"blockedBy"`
	Owner       string                 `json:"owner,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskStore handles task persistence
type TaskStore struct {
	ProjectName string
	Tasks       []Task
	projectDir  string    // cached project directory path
	lastModTime time.Time // last modification time of project directory
}

// NewTaskStoreForTest creates a TaskStore for testing with a custom directory
func NewTaskStoreForTest(dir string, tasks []Task) (*TaskStore, error) {
	store := &TaskStore{
		ProjectName: "test",
		Tasks:       tasks,
		projectDir:  dir,
	}
	// Save each task to file
	for _, task := range tasks {
		data, err := json.MarshalIndent(task, "", "  ")
		if err != nil {
			return nil, err
		}
		filePath := filepath.Join(dir, task.ID+".json")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return nil, err
		}
	}
	return store, nil
}

// Project represents a project with task count
type Project struct {
	Name      string
	TaskCount int
}

// ListProjects returns all projects in the tasks directory
func ListProjects() ([]Project, error) {
	tasksDir, err := config.GetTasksDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Project{}, nil
		}
		return nil, err
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectName := entry.Name()
		projectDir := filepath.Join(tasksDir, projectName)

		// Count task files (*.json except _groups.json)
		taskCount := countTaskFiles(projectDir)
		if taskCount == 0 {
			continue
		}

		projects = append(projects, Project{
			Name:      projectName,
			TaskCount: taskCount,
		})
	}

	// Sort by name
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// countTaskFiles counts the number of task JSON files in a directory
func countTaskFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip _groups.json and other non-task files
		if strings.HasPrefix(name, "_") {
			continue
		}
		if strings.HasSuffix(name, ".json") {
			count++
		}
	}
	return count
}

// LoadTasks loads tasks from individual JSON files in the project directory
func LoadTasks(projectName string) (*TaskStore, error) {
	projectDir, err := config.GetProjectDir(projectName)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskStore{
				ProjectName: projectName,
				Tasks:       []Task{},
				projectDir:  projectDir,
			}, nil
		}
		return nil, err
	}

	// Get directory modification time
	dirInfo, err := os.Stat(projectDir)
	var modTime time.Time
	if err == nil {
		modTime = dirInfo.ModTime()
	}

	var tasks []Task
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip non-task files
		if strings.HasPrefix(name, "_") || !strings.HasSuffix(name, ".json") {
			continue
		}

		filePath := filepath.Join(projectDir, name)
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

	// Sort by ID (numeric)
	sort.Slice(tasks, func(i, j int) bool {
		idI, _ := strconv.Atoi(tasks[i].ID)
		idJ, _ := strconv.Atoi(tasks[j].ID)
		return idI < idJ
	})

	store := &TaskStore{
		ProjectName: projectName,
		Tasks:       tasks,
		projectDir:  projectDir,
		lastModTime: modTime,
	}

	// Backup all task files (only if source is newer)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "_") || !strings.HasSuffix(name, ".json") {
			continue
		}
		store.backupFile(name)
	}

	return store, nil
}

// Save saves all tasks to individual JSON files
func (s *TaskStore) Save() error {
	projectDir, err := config.GetProjectDir(s.ProjectName)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	// Save each task to its own file
	for _, task := range s.Tasks {
		if err := s.saveTask(task); err != nil {
			return err
		}
	}

	return nil
}

// saveTask saves a single task to its JSON file
func (s *TaskStore) saveTask(task Task) error {
	projectDir, err := config.GetProjectDir(s.ProjectName)
	if err != nil {
		return err
	}

	filePath := filepath.Join(projectDir, task.ID+".json")
	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// Backup: write only if content differs
	s.backupTaskData(task.ID+".json", data)
	return nil
}

// backupTaskData backs up task data to backup directory if content differs
func (s *TaskStore) backupTaskData(filename string, data []byte) {
	backupDir, err := config.GetBackupProjectDir(s.ProjectName)
	if err != nil {
		return
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return
	}

	backupPath := filepath.Join(backupDir, filename)

	// Check if backup exists and has same content
	existing, err := os.ReadFile(backupPath)
	if err == nil && string(existing) == string(data) {
		return // Same content, skip write
	}

	os.WriteFile(backupPath, data, 0644)
}

// backupFile copies a file to backup directory if source is newer
func (s *TaskStore) backupFile(filename string) {
	projectDir, err := config.GetProjectDir(s.ProjectName)
	if err != nil {
		return
	}
	backupDir, err := config.GetBackupProjectDir(s.ProjectName)
	if err != nil {
		return
	}

	srcPath := filepath.Join(projectDir, filename)
	dstPath := filepath.Join(backupDir, filename)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return
	}

	// Check if backup is up-to-date
	dstInfo, err := os.Stat(dstPath)
	if err == nil && !srcInfo.ModTime().After(dstInfo.ModTime()) {
		return // Backup is up-to-date, skip
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return
	}

	os.WriteFile(dstPath, data, 0644)
}

// NeedsReload checks if the project directory has been modified since last load
func (s *TaskStore) NeedsReload() bool {
	if s.projectDir == "" {
		return false
	}
	dirInfo, err := os.Stat(s.projectDir)
	if err != nil {
		return false
	}
	return dirInfo.ModTime().After(s.lastModTime)
}

// GetTask returns a task by ID
func (s *TaskStore) GetTask(id string) *Task {
	for i := range s.Tasks {
		if s.Tasks[i].ID == id {
			return &s.Tasks[i]
		}
	}
	return nil
}

// AddTask adds a new task and returns the assigned ID
func (s *TaskStore) AddTask(task Task) string {
	task.ID = s.generateID()
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.Blocks == nil {
		task.Blocks = []string{}
	}
	if task.BlockedBy == nil {
		task.BlockedBy = []string{}
	}
	s.Tasks = append(s.Tasks, task)
	return task.ID
}

// UpdateTask updates an existing task
func (s *TaskStore) UpdateTask(task Task) error {
	for i := range s.Tasks {
		if s.Tasks[i].ID == task.ID {
			s.Tasks[i] = task
			return nil
		}
	}
	return fmt.Errorf("task not found: %s", task.ID)
}

// DeleteTask removes a task by ID
func (s *TaskStore) DeleteTask(id string) error {
	for i := range s.Tasks {
		if s.Tasks[i].ID == id {
			// Remove from blocks/blockedBy of other tasks
			for j := range s.Tasks {
				if j == i {
					continue
				}
				s.Tasks[j].Blocks = removeFromSlice(s.Tasks[j].Blocks, id)
				s.Tasks[j].BlockedBy = removeFromSlice(s.Tasks[j].BlockedBy, id)
			}

			// Remove from memory
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)

			// Delete the file
			projectDir, err := config.GetProjectDir(s.ProjectName)
			if err != nil {
				return err
			}
			filePath := filepath.Join(projectDir, id+".json")
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				return err
			}

			return nil
		}
	}
	return fmt.Errorf("task not found: %s", id)
}

// generateID generates a new unique task ID
func (s *TaskStore) generateID() string {
	maxID := 0
	for _, task := range s.Tasks {
		if id, err := strconv.Atoi(task.ID); err == nil {
			if id > maxID {
				maxID = id
			}
		}
	}
	return strconv.Itoa(maxID + 1)
}

// GetTasksByStatus returns tasks filtered by status
func (s *TaskStore) GetTasksByStatus(status string) []Task {
	if status == "" || status == "all" {
		return s.Tasks
	}

	var filtered []Task
	for _, task := range s.Tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// GetTasksByGroup returns tasks filtered by group (from metadata)
func (s *TaskStore) GetTasksByGroup(group string) []Task {
	var filtered []Task
	for _, task := range s.Tasks {
		taskGroup := GetTaskGroup(task)
		if group == "" || group == "all" || taskGroup == group {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// SearchTasks returns tasks matching the search query in subject or description
func (s *TaskStore) SearchTasks(query string) []Task {
	if query == "" {
		return s.Tasks
	}

	query = strings.ToLower(query)
	var filtered []Task
	for _, task := range s.Tasks {
		if strings.Contains(strings.ToLower(task.Subject), query) ||
			strings.Contains(strings.ToLower(task.Description), query) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// GetTaskGroup returns the group name from task metadata
func GetTaskGroup(task Task) string {
	if task.Metadata == nil {
		return ""
	}
	if group, ok := task.Metadata["group"].(string); ok {
		return group
	}
	return ""
}

// SetTaskGroup sets the group name in task metadata
func SetTaskGroup(task *Task, group string) {
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}
	if group == "" {
		delete(task.Metadata, "group")
	} else {
		task.Metadata["group"] = group
	}
}

// GetAllGroups returns all unique group names from tasks
func (s *TaskStore) GetAllGroups() []string {
	groupSet := make(map[string]bool)
	for _, task := range s.Tasks {
		group := GetTaskGroup(task)
		if group != "" {
			groupSet[group] = true
		}
	}

	var groups []string
	for group := range groupSet {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups
}

// StatusIcon returns the icon for a task status
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

func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
