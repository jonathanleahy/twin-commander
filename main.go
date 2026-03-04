package main

import (
	"fmt"
	"os"
)

func main() {
	// Optional: pass a directory path as the first argument
	startPath := ""
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}

	app := NewApp(startPath)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
