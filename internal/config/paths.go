package config

import (
	"os"
	"path/filepath"
)

// GetTasksDir returns the path to ~/.claude/tasks/
func GetTasksDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claude", "tasks"), nil
}

// GetProjectDir returns the path to a specific project's tasks directory
func GetProjectDir(projectName string) (string, error) {
	tasksDir, err := GetTasksDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tasksDir, projectName), nil
}

// GetTasksFilePath returns the path to the tasks.json file for a project
func GetTasksFilePath(projectName string) (string, error) {
	projectDir, err := GetProjectDir(projectName)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, "tasks.json"), nil
}

// GetGroupsFilePath returns the path to the _groups.json file for a project
func GetGroupsFilePath(projectName string) (string, error) {
	projectDir, err := GetProjectDir(projectName)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, "_groups.json"), nil
}
