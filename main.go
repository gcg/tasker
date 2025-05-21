package main

import (
	"fmt"
	"os" // For exiting with a status code and Getwd
	"path/filepath" // For Base

	"todo-cli/internal/store"
	"todo-cli/internal/tui"
	// "todo-cli/internal/model" // Not directly needed in main usually
)

const defaultTodoFile = "todo.yaml" // Define it here or import from store if it's made public there

func main() {
	// Generate dynamic title - REMOVED FOR NOW TO MATCH REVERTED BUBBLE.GO
	// wd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
	// 	// Fallback title or exit, for now, let's use a default
	// 	wd = "Todo" // Default title part if Getwd fails
	// }
	// baseDirName := filepath.Base(wd)
	// dynamicTitle := fmt.Sprintf("%s Task List", baseDirName)

	// Load tasks from the YAML file
	tasks, err := store.LoadTasks(defaultTodoFile) // tasks variable was already defined, removed re-declaration
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading tasks: %v\n", err)
		fmt.Fprintf(os.Stderr, "You can try deleting '%s' if it's corrupted.\n", defaultTodoFile)
		os.Exit(1) // Exit with an error code
	}

	// Start the Bubble Tea program with the dynamic title
	if err := tui.StartTeaProgram(tasks); err != nil { // Call with one argument
		fmt.Fprintf(os.Stderr, "Alas, there's been an error running the TUI: %v\n", err)
		if saveErr := store.SaveTasks(defaultTodoFile, tasks); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Additionally, failed to save tasks: %v\n", saveErr)
		}
		os.Exit(1)
	}

	fmt.Println("Todo CLI exited. Your tasks should be saved.")
}
