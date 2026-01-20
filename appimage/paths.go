package appimage

import (
	"os"
	"path/filepath"
	"strings"
)

// GetInstalledAppImagePath returns ~/Applications/<appName>.AppImage
func GetInstalledAppImagePath(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "Applications", appName+".AppImage"), nil
}

// GetDesktopFilePath returns ~/.local/share/applications/<appName>.desktop
func GetDesktopFilePath(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local/share/applications", appName+".desktop"), nil
}

// GetIconPath returns ~/.local/share/icons/<appName>.<ext>
func GetIconPath(appName, ext string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return filepath.Join(homeDir, ".local/share/icons", appName+ext), nil
}
