package main

import (
	"fmt"
	"os" // For exiting with a status code

	"todo-cli/internal/store"
	"todo-cli/internal/tui"
	// "todo-cli/internal/model" // Not directly needed in main usually
)

const defaultTodoFile = "todo.yaml" // Define it here or import from store if it's made public there

func main() {
	// Load tasks from the YAML file
	// Using an empty string for filePath in LoadTasks will make it use defaultTodoFile.
	tasks, err := store.LoadTasks(defaultTodoFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		fmt.Fprintf(os.Stderr, "You can try deleting '%s' if it's corrupted.\n", defaultTodoFile)
		os.Exit(1) // Exit with an error code
	}

	// Start the Bubble Tea program
	if err := tui.StartTeaProgram(tasks); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error running the TUI: %v\n", err)
		// Attempt to save tasks one last time, in case the TUI didn't exit cleanly
		// but had pending changes. This is a best-effort.
		// Note: The TUI's StartTeaProgram already attempts a save on exit.
		// This is mainly if StartTeaProgram itself returns an error before/during setup.
		if saveErr := store.SaveTasks(defaultTodoFile, tasks); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Additionally, failed to save tasks: %v\n", saveErr)
		}
		os.Exit(1)
	}

	// If we reach here, the TUI exited cleanly.
	// Tasks should have been saved by the TUI upon modifications or its own clean exit.
	fmt.Println("Todo CLI exited. Your tasks should be saved.")
}
