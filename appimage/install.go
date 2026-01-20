package appimage

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
var iconExtensions = []string{".png", ".svg", ".jpg", ".xpm"}

// SanitizeAppName removes spaces and special characters from app name
func SanitizeAppName(name string) string {
	result := strings.ToLower(name)
	result = nonAlphanumeric.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")
	return result
}

// IsInstalled checks if an app is currently installed
func IsInstalled(appName string) bool {
	appImagePath, err := GetInstalledAppImagePath(SanitizeAppName(appName))
	if err != nil {
		return false
	}
	_, err = os.Stat(appImagePath)
	return err == nil
}

// IsInstalledPath checks if the given file path is the installed version
func IsInstalledPath(appName string, filePath string) bool {
	appImagePath, err := GetInstalledAppImagePath(SanitizeAppName(appName))
	if err != nil {
		return false
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	return absFilePath == appImagePath
}

// Install copies the AppImage and creates desktop entry
func Install(metadata *Metadata) error {
	if metadata == nil {
		return &InstallError{Stage: "validation", Err: os.ErrInvalid}
	}

	sanitizedName := SanitizeAppName(metadata.Name)
	if sanitizedName == "" {
		return &InstallError{AppName: metadata.Name, Stage: "validation", Err: os.ErrInvalid}
	}

	targetAppImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	targetDesktopPath, err := GetDesktopFilePath(sanitizedName)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	// Track created files for rollback on error
	var createdFiles []string
	cleanup := func() {
		for _, file := range createdFiles {
			os.Remove(file)
		}
	}

	// Copy AppImage
	if err := os.MkdirAll(filepath.Dir(targetAppImagePath), 0755); err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "mkdir", Err: err}
	}

	srcFile, err := os.Open(metadata.FilePath)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "copy", Err: err}
	}
	defer srcFile.Close()

	dstFile, err := os.Create(targetAppImagePath)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "copy", Err: err}
	}
	createdFiles = append(createdFiles, targetAppImagePath)

	_, err = io.Copy(dstFile, srcFile)
	dstFile.Close()
	if err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "copy", Err: err}
	}

	if err := os.Chmod(targetAppImagePath, 0755); err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "permissions", Err: err}
	}

	// Handle icon installation
	var targetIconPath string
	if metadata.Icon != nil && len(metadata.Icon.Data) > 0 {
		targetIconPath, err = GetIconPath(sanitizedName, "."+metadata.Icon.Format)
		if err == nil {
			if err := os.MkdirAll(filepath.Dir(targetIconPath), 0755); err == nil {
				if err := os.WriteFile(targetIconPath, metadata.Icon.Data, 0644); err == nil {
					createdFiles = append(createdFiles, targetIconPath)
				} else {
					targetIconPath = ""
				}
			} else {
				targetIconPath = ""
			}
		}
	}

	// Create and write desktop file
	modifiedDesktop, err := ModifyDesktopFile(
		metadata.DesktopFileContent,
		targetAppImagePath,
		targetIconPath,
		filepath.Base(metadata.FilePath),
	)
	if err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "desktop", Err: err}
	}

	if err := os.MkdirAll(filepath.Dir(targetDesktopPath), 0755); err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "mkdir", Err: err}
	}

	if err := os.WriteFile(targetDesktopPath, modifiedDesktop, 0644); err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "desktop", Err: err}
	}

	return nil
}

// Uninstall removes all installed files
func Uninstall(appName string) error {
	sanitizedName := SanitizeAppName(appName)

	appImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return err
	}
	if _, err := os.Stat(appImagePath); os.IsNotExist(err) {
		return ErrNotInstalled
	}

	if err := os.Remove(appImagePath); err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "remove", Err: err}
	}

	if desktopPath, err := GetDesktopFilePath(sanitizedName); err == nil {
		os.Remove(desktopPath)
	}

	for _, ext := range iconExtensions {
		if path, err := GetIconPath(sanitizedName, ext); err == nil {
			os.Remove(path)
		}
	}

	return nil
}
