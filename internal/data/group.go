package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jss826/cctasks/internal/config"
)

// TaskGroup represents a task group with styling
type TaskGroup struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
	Color string `json:"color"`
}

// GroupStore handles group persistence
type GroupStore struct {
	ProjectName string
	Groups      []TaskGroup
	filePath    string    // cached file path
	lastModTime time.Time // last modification time
}

// groupsFile represents the JSON structure of _groups.json
type groupsFile struct {
	Groups []TaskGroup `json:"groups"`
}

// DefaultColors provides preset colors for groups
var DefaultColors = []string{
	"#8b5cf6", // purple
	"#3b82f6", // blue
	"#10b981", // green
	"#f59e0b", // amber
	"#ef4444", // red
	"#ec4899", // pink
	"#06b6d4", // cyan
	"#84cc16", // lime
}

// NewGroupStoreForTest creates a GroupStore for testing with a custom directory
func NewGroupStoreForTest(dir string, groups []TaskGroup) (*GroupStore, error) {
	filePath := filepath.Join(dir, "_groups.json")
	store := &GroupStore{
		Groups:   groups,
		filePath: filePath,
	}
	// Assign orders if not set
	for i := range store.Groups {
		if store.Groups[i].Order == 0 && i > 0 {
			store.Groups[i].Order = i
		}
	}
	if err := store.Save(); err != nil {
		return nil, err
	}
	return store, nil
}

// LoadGroups loads groups from a project's _groups.json
func LoadGroups(projectName string) (*GroupStore, error) {
	groupsFilePath, err := config.GetGroupsFilePath(projectName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(groupsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &GroupStore{
				ProjectName: projectName,
				Groups:      []TaskGroup{},
				filePath:    groupsFilePath,
			}, nil
		}
		return nil, err
	}

	// Get file modification time
	fileInfo, err := os.Stat(groupsFilePath)
	var modTime time.Time
	if err == nil {
		modTime = fileInfo.ModTime()
	}

	var gf groupsFile
	if err := json.Unmarshal(data, &gf); err != nil {
		return nil, err
	}

	// Sort by order
	sort.Slice(gf.Groups, func(i, j int) bool {
		return gf.Groups[i].Order < gf.Groups[j].Order
	})

	return &GroupStore{
		ProjectName: projectName,
		Groups:      gf.Groups,
		filePath:    groupsFilePath,
		lastModTime: modTime,
	}, nil
}

// Save saves groups to the project's _groups.json
func (s *GroupStore) Save() error {
	groupsFilePath, err := config.GetGroupsFilePath(s.ProjectName)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(groupsFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	gf := groupsFile{Groups: s.Groups}
	data, err := json.MarshalIndent(gf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(groupsFilePath, data, 0644)
}

// NeedsReload checks if the groups file has been modified since last load
func (s *GroupStore) NeedsReload() bool {
	if s.filePath == "" {
		return false
	}
	fileInfo, err := os.Stat(s.filePath)
	if err != nil {
		return false
	}
	return fileInfo.ModTime().After(s.lastModTime)
}

// GetGroup returns a group by name
func (s *GroupStore) GetGroup(name string) *TaskGroup {
	for i := range s.Groups {
		if s.Groups[i].Name == name {
			return &s.Groups[i]
		}
	}
	return nil
}

// AddGroup adds a new group
func (s *GroupStore) AddGroup(group TaskGroup) {
	// Set order to be at the end
	maxOrder := 0
	for _, g := range s.Groups {
		if g.Order > maxOrder {
			maxOrder = g.Order
		}
	}
	group.Order = maxOrder + 1

	// Set default color if not specified
	if group.Color == "" {
		group.Color = DefaultColors[len(s.Groups)%len(DefaultColors)]
	}

	s.Groups = append(s.Groups, group)
}

// UpdateGroup updates an existing group
func (s *GroupStore) UpdateGroup(name string, updated TaskGroup) bool {
	for i := range s.Groups {
		if s.Groups[i].Name == name {
			s.Groups[i] = updated
			return true
		}
	}
	return false
}

// DeleteGroup removes a group by name
func (s *GroupStore) DeleteGroup(name string) bool {
	for i := range s.Groups {
		if s.Groups[i].Name == name {
			s.Groups = append(s.Groups[:i], s.Groups[i+1:]...)
			return true
		}
	}
	return false
}

// MoveGroupUp moves a group up in order
func (s *GroupStore) MoveGroupUp(name string) bool {
	for i := range s.Groups {
		if s.Groups[i].Name == name && i > 0 {
			// Swap orders
			s.Groups[i].Order, s.Groups[i-1].Order = s.Groups[i-1].Order, s.Groups[i].Order
			// Re-sort
			sort.Slice(s.Groups, func(a, b int) bool {
				return s.Groups[a].Order < s.Groups[b].Order
			})
			return true
		}
	}
	return false
}

// MoveGroupDown moves a group down in order
func (s *GroupStore) MoveGroupDown(name string) bool {
	for i := range s.Groups {
		if s.Groups[i].Name == name && i < len(s.Groups)-1 {
			// Swap orders
			s.Groups[i].Order, s.Groups[i+1].Order = s.Groups[i+1].Order, s.Groups[i].Order
			// Re-sort
			sort.Slice(s.Groups, func(a, b int) bool {
				return s.Groups[a].Order < s.Groups[b].Order
			})
			return true
		}
	}
	return false
}

// GetGroupNames returns just the group names in order
func (s *GroupStore) GetGroupNames() []string {
	names := make([]string, len(s.Groups))
	for i, g := range s.Groups {
		names[i] = g.Name
	}
	return names
}

// GetGroupColor returns the color for a group name
func (s *GroupStore) GetGroupColor(name string) string {
	for _, g := range s.Groups {
		if g.Name == name {
			return g.Color
		}
	}
	// Return a default color for unknown groups
	return "#6b7280" // gray
}

// EnsureGroupExists creates a group if it doesn't exist
func (s *GroupStore) EnsureGroupExists(name string) {
	if name == "" {
		return
	}
	if s.GetGroup(name) == nil {
		s.AddGroup(TaskGroup{Name: name})
	}
}
