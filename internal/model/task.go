package model

import (
	"time"
	"github.com/google/uuid" // For generating unique IDs
)

// Status represents the status of a task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
)

// Task represents a single todo item.
type Task struct {
	ID          string    `yaml:"id"`
	Description string    `yaml:"description"`
	Status      Status    `yaml:"status"`
	CreatedAt   time.Time `yaml:"created_at"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty"` // Pointer to allow nullability
	Subtasks    []Task    `yaml:"subtasks,omitempty"`
}

// NewTask creates a new task with a unique ID and current creation time.
func NewTask(description string) Task {
	return Task{
		ID:          uuid.NewString(),
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		Subtasks:    make([]Task, 0), // Initialize with an empty slice, not nil
	}
}
