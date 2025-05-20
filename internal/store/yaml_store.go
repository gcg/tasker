package store

import (
	"fmt"
	"os"
	"path/filepath" // To ensure cross-platform compatibility for file paths

	"todo-cli/internal/model" // Assuming this is the module path

	"gopkg.in/yaml.v3"
)

const defaultTodoFile = "todo.yaml"

// LoadTasks loads tasks from the specified YAML file.
// If the file does not exist, it creates an empty one and returns an empty slice of tasks.
func LoadTasks(filePath string) ([]model.Task, error) {
	if filePath == "" {
		filePath = defaultTodoFile
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, create it with an empty list
		fmt.Printf("No '%s' found. Creating a new one in the current directory.\n", filePath)
		emptyTasks := make([]model.Task, 0)
		if err := SaveTasks(filePath, emptyTasks); err != nil {
			return nil, fmt.Errorf("failed to create initial tasks file '%s': %w", filePath, err)
		}
		return emptyTasks, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file '%s': %w", filePath, err)
	}

	var tasks []model.Task
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		// Handle case where the file is empty or not valid YAML
		if len(data) == 0 {
			// File is empty, return empty list
			return make([]model.Task, 0), nil
		}
		return nil, fmt.Errorf("failed to unmarshal tasks from YAML file '%s': %w", filePath, err)
	}
	
	// If tasks is nil after unmarshalling (e.g. empty file or just "[]"), make it an empty slice
	if tasks == nil {
		tasks = make([]model.Task, 0)
	}

	return tasks, nil
}

// SaveTasks saves the given tasks to the specified YAML file.
func SaveTasks(filePath string, tasks []model.Task) error {
	if filePath == "" {
		filePath = defaultTodoFile
	}

	// Ensure tasks is not nil to avoid writing "null" to the file for an empty list
	if tasks == nil {
		tasks = make([]model.Task, 0)
	}

	data, err := yaml.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks to YAML: %w", err)
	}

	// Get absolute path for clearer error messages if needed
	absPath, _ := filepath.Abs(filePath)
	err = os.WriteFile(filePath, data, 0644) // 0644 are standard file permissions
	if err != nil {
		return fmt.Errorf("failed to write tasks to file '%s' (abs: '%s'): %w", filePath, absPath, err)
	}
	return nil
}
