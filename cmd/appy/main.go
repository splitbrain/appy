package main

import (
	"fmt"
	"os"

	"appy/internal/appimage"
	"appy/internal/ui"
)

func main() {
	// Check command-line arguments
	if len(os.Args) != 2 {
		fmt.Println("Usage: appy <path-to-appimage>")
		fmt.Println("\nExample:")
		fmt.Println("  appy ~/Downloads/MyApp.AppImage")
		os.Exit(1)
	}

	appImagePath := os.Args[1]

	// Check if file exists
	if _, err := os.Stat(appImagePath); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", appImagePath)
		os.Exit(1)
	}

	// Parse AppImage
	metadata, err := appimage.Parse(appImagePath)
	if err != nil {
		// Show error in GUI
		ui.ShowError(err)
		os.Exit(1)
	}

	// Show metadata in GUI
	ui.ShowMetadata(metadata)
}
