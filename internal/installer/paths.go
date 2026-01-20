package installer

import (
	"os"
	"path/filepath"
)

const (
	ApplicationsDir = "Applications"
	DesktopFileDir  = ".local/share/applications"
	IconsDir        = ".local/share/icons"
)

// GetApplicationsPath returns ~/Applications
func GetApplicationsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ApplicationsDir), nil
}

// GetDesktopFilesPath returns ~/.local/share/applications
func GetDesktopFilesPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, DesktopFileDir), nil
}

// GetIconsPath returns ~/.local/share/icons
func GetIconsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, IconsDir), nil
}

// GetInstalledAppImagePath returns full path to installed AppImage
func GetInstalledAppImagePath(appName string) (string, error) {
	appsPath, err := GetApplicationsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(appsPath, appName+".AppImage"), nil
}

// GetDesktopFilePath returns full path to desktop file
func GetDesktopFilePath(appName string) (string, error) {
	desktopPath, err := GetDesktopFilesPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(desktopPath, appName+".desktop"), nil
}

// GetIconPath returns full path to icon file
func GetIconPath(appName, extension string) (string, error) {
	iconsPath, err := GetIconsPath()
	if err != nil {
		return "", err
	}
	// Ensure extension starts with a dot
	if extension != "" && extension[0] != '.' {
		extension = "." + extension
	}
	return filepath.Join(iconsPath, appName+extension), nil
}
