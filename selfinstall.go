package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"appy/appimage"
)

const desktopFileContent = `[Desktop Entry]
Type=Application
Name=Appy
Comment=AppImage installer and manager
Exec=%s %%f
Icon=application-x-appimage
Terminal=false
Categories=Utility;
MimeType=application/x-iso9660-appimage;application/vnd.appimage;application/x-appimage;
`

func installAppy() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	// Get path to current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	targetBinaryPath := filepath.Join(homeDir, "Applications", "appy")
	targetDesktopPath := filepath.Join(homeDir, ".local/share/applications", "appy.desktop")

	// Create ~/Applications directory
	if err := os.MkdirAll(filepath.Dir(targetBinaryPath), 0755); err != nil {
		return fmt.Errorf("cannot create Applications directory: %w", err)
	}

	// Copy binary
	if err := appimage.CopyFile(execPath, targetBinaryPath, 0755); err != nil {
		return fmt.Errorf("cannot copy binary: %w", err)
	}

	// Create desktop file directory
	if err := os.MkdirAll(filepath.Dir(targetDesktopPath), 0755); err != nil {
		return fmt.Errorf("cannot create applications directory: %w", err)
	}

	// Write desktop file
	desktopContent := fmt.Sprintf(desktopFileContent, targetBinaryPath)
	if err := os.WriteFile(targetDesktopPath, []byte(desktopContent), 0644); err != nil {
		_ = os.Remove(targetBinaryPath)
		return fmt.Errorf("cannot write desktop file: %w", err)
	}

	// Update desktop database (ignore errors - command may not exist)
	_ = exec.Command("update-desktop-database", filepath.Dir(targetDesktopPath)).Run()

	// Set as default handler for AppImage MIME types
	for _, mimeType := range []string{
		"application/x-iso9660-appimage",
		"application/vnd.appimage",
		"application/x-appimage",
	} {
		_ = exec.Command("xdg-mime", "default", "appy.desktop", mimeType).Run()
	}

	return nil
}

func uninstallAppy() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	targetBinaryPath := filepath.Join(homeDir, "Applications", "appy")
	targetDesktopPath := filepath.Join(homeDir, ".local/share/applications", "appy.desktop")

	// Check if installed
	if _, err := os.Stat(targetBinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("appy is not installed")
	}

	// Remove binary
	if err := os.Remove(targetBinaryPath); err != nil {
		return fmt.Errorf("cannot remove binary: %w", err)
	}

	// Remove desktop file (ignore error if it doesn't exist)
	_ = os.Remove(targetDesktopPath)

	// Update desktop database (ignore errors - command may not exist)
	_ = exec.Command("update-desktop-database", filepath.Dir(targetDesktopPath)).Run()

	return nil
}
