package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Todo CLI Application")
	// Later, this will initialize and run the Bubble Tea application
	// For now, just a placeholder
	if len(os.Args) > 1 && os.Args[1] == "test" {
		fmt.Println("Test argument received.")
	}
}
