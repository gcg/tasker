package service

import (
	"fmt"
	"time"
	"todo-cli/internal/model" // Assuming module path
)

// AddTask adds a new task to the list of tasks.
// It returns the updated list of tasks.
func AddTask(tasks []model.Task, description string) []model.Task {
	newTask := model.NewTask(description)
	return append(tasks, newTask)
}

// findTaskRecursive searches for a task by ID in a list of tasks and their subtasks.
// It returns the task, its parent (if any), and the list it belongs to.
func findTaskRecursive(tasks []model.Task, taskID string) (targetTask *model.Task, parentTask *model.Task, taskList []model.Task, found bool) {
	for i, task := range tasks {
		if task.ID == taskID {
			// Create a new slice that is a copy of the original tasks to avoid modifying the input slice directly
			// This is important if the caller expects the original slice to be unchanged
			// However, for our use case, modifying the underlying array of the slice is intended
			return &tasks[i], nil, tasks, true 
		}
		if len(task.Subtasks) > 0 {
			if t, _, subList, f := findTaskRecursive(task.Subtasks, taskID); f {
				// If found in subtasks, the current task is the parent
				return t, &tasks[i], subList, true
			}
		}
	}
	return nil, nil, nil, false
}

// FindTaskByID searches for a task by its ID within the task list (including subtasks).
// Returns the found task, its parent (if any), and an error if not found.
func FindTaskByID(tasks []model.Task, taskID string) (targetTask *model.Task, parentTask *model.Task, err error) {
	t, p, _, found := findTaskRecursive(tasks, taskID)
	if !found {
		return nil, nil, fmt.Errorf("task with ID '%s' not found", taskID)
	}
	return t, p, nil
}

// AddSubtask adds a new subtask to the task with parentID.
// Returns the (potentially modified) root list of tasks and an error if the parent is not found.
func AddSubtask(tasks []model.Task, parentID string, description string) ([]model.Task, error) {
	parentTask, _, err := FindTaskByID(tasks, parentID)
	if err != nil {
		return tasks, fmt.Errorf("parent task for subtask not found: %w", err)
	}

	newSubtask := model.NewTask(description)
	parentTask.Subtasks = append(parentTask.Subtasks, newSubtask)
	
	return tasks, nil
}

// EditTask updates the description of a task or subtask identified by taskID.
// Returns the (potentially modified) root list of tasks and an error if the task is not found.
func EditTask(tasks []model.Task, taskID string, newDescription string) ([]model.Task, error) {
	taskToEdit, _, err := FindTaskByID(tasks, taskID)
	if err != nil {
		return tasks, err
	}

	taskToEdit.Description = newDescription
	return tasks, nil
}

// ToggleTaskStatus changes the status of a task or subtask (Pending <-> Completed).
// It sets/clears the CompletedAt timestamp accordingly.
// Returns the (potentially modified) root list of tasks and an error if the task is not found.
func ToggleTaskStatus(tasks []model.Task, taskID string) ([]model.Task, error) {
	taskToToggle, _, err := FindTaskByID(tasks, taskID)
	if err != nil {
		return tasks, err
	}

	if taskToToggle.Status == model.StatusCompleted {
		taskToToggle.Status = model.StatusPending
		taskToToggle.CompletedAt = nil // Clear completion time
	} else {
		taskToToggle.Status = model.StatusCompleted
		now := time.Now()
		taskToToggle.CompletedAt = &now // Set completion time
	}
	return tasks, nil
}

// RemoveTask removes a task or subtask from its list.
// Note: This is a more complex version that handles subtasks.
// It returns the root task list and a boolean indicating if a modification was made.
func RemoveTask(tasks []model.Task, taskID string) ([]model.Task, bool) {
    // Check top-level tasks first
    for i, task := range tasks {
        if task.ID == taskID {
            return append(tasks[:i], tasks[i+1:]...), true
        }
    }

    // Check subtasks
    for i := range tasks {
        // Make sure to operate on a modifiable copy if Subtasks can be nil or empty
        if tasks[i].Subtasks != nil {
            cleanedSubtasks, modified := RemoveTask(tasks[i].Subtasks, taskID)
            if modified {
                tasks[i].Subtasks = cleanedSubtasks
                return tasks, true
            }
        }
    }
    return tasks, false
}


// Note: MoveTask is a stretch goal and is not implemented here yet.
// func MoveTask(tasks []model.Task, taskID string, newParentID string, newIndex int) []model.Task {
// 	// Implementation would be more complex, involving finding the task, removing it from its current location,
// 	// and inserting it into the new location (either root or as a subtask).
// 	// Careful handling of slices and indices is required.
// 	return tasks
// }
