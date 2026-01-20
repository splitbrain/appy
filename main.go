package main

import (
	"fmt"
	"os"

	"appy/appimage"
)

func main() {
	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}

	if handleOption(os.Args[1]) {
		return
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
		showError(err)
		os.Exit(1)
	}

	// Show metadata in GUI
	showMainWindow(metadata)
}

func handleOption(arg string) bool {
	switch arg {
	case "--install":
		if err := installAppy(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Appy installed successfully")
		return true
	case "--uninstall":
		if err := uninstallAppy(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Appy uninstalled successfully")
		return true
	case "--help", "-h":
		printUsage()
		return true
	}
	return false
}

func printUsage() {
	fmt.Println("Usage: appy <path-to-appimage>")
	fmt.Println("       appy --install")
	fmt.Println("       appy --uninstall")
	fmt.Println("\nOptions:")
	fmt.Println("  --install    Install appy as default AppImage handler")
	fmt.Println("  --uninstall  Remove appy installation")
	fmt.Println("\nExample:")
	fmt.Println("  appy ~/Downloads/MyApp.AppImage")
}
