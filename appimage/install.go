package appimage

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// InstalledInfo contains information about an installed application
type InstalledInfo struct {
	AppName          string
	InstalledPath    string
	DesktopFilePath  string
	IconPath         string
	OriginalFilename string
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// SanitizeAppName removes spaces and special characters from app name
func SanitizeAppName(name string) string {
	result := strings.ToLower(name)
	result = nonAlphanumeric.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")
	return result
}

// IsInstalled checks if an app is currently installed
func IsInstalled(appName string) bool {
	sanitizedName := SanitizeAppName(appName)
	appImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return false
	}

	// Check if the AppImage exists
	if _, err := os.Stat(appImagePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// IsInstalledPath checks if the given file path is the installed version
func IsInstalledPath(appName string, filePath string) bool {
	sanitizedName := SanitizeAppName(appName)
	appImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return false
	}

	// Resolve both paths to absolute paths for comparison
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	absInstalledPath, err := filepath.Abs(appImagePath)
	if err != nil {
		return false
	}

	return absFilePath == absInstalledPath
}

// GetInstalledInfo retrieves information about an installed app
func GetInstalledInfo(appName string) (*InstalledInfo, error) {
	sanitizedName := SanitizeAppName(appName)

	appImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return nil, err
	}

	desktopFilePath, err := GetDesktopFilePath(sanitizedName)
	if err != nil {
		return nil, err
	}

	// Check if installed
	if _, err := os.Stat(appImagePath); os.IsNotExist(err) {
		return nil, ErrNotInstalled
	}

	// Try to get original filename from desktop file
	originalFilename, _ := GetOriginalFilename(desktopFilePath)

	// Try to find icon path
	iconPath := ""
	iconsDir, err := GetIconsPath()
	if err == nil {
		// Check for common extensions
		for _, ext := range []string{".png", ".svg", ".jpg", ".xpm"} {
			path := filepath.Join(iconsDir, sanitizedName+ext)
			if _, err := os.Stat(path); err == nil {
				iconPath = path
				break
			}
		}
	}

	return &InstalledInfo{
		AppName:          sanitizedName,
		InstalledPath:    appImagePath,
		DesktopFilePath:  desktopFilePath,
		IconPath:         iconPath,
		OriginalFilename: originalFilename,
	}, nil
}

// Install copies the AppImage and creates desktop entry
func Install(metadata *Metadata) error {
	if metadata == nil {
		return &InstallError{Stage: "validation", Err: os.ErrInvalid}
	}

	sanitizedName := SanitizeAppName(metadata.Name)
	if sanitizedName == "" {
		return &InstallError{
			AppName: metadata.Name,
			Stage:   "validation",
			Err:     os.ErrInvalid,
		}
	}

	// Get paths
	appsDir, err := GetApplicationsPath()
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	desktopDir, err := GetDesktopFilesPath()
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	iconsDir, err := GetIconsPath()
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	targetAppImagePath, err := GetInstalledAppImagePath(sanitizedName)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	targetDesktopPath, err := GetDesktopFilePath(sanitizedName)
	if err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "paths", Err: err}
	}

	// Create directories if they don't exist
	if err := os.MkdirAll(appsDir, 0755); err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "mkdir", Err: err}
	}

	if err := os.MkdirAll(desktopDir, 0755); err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "mkdir", Err: err}
	}

	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return &InstallError{AppName: sanitizedName, Stage: "mkdir", Err: err}
	}

	// Track created files for rollback on error
	var createdFiles []string
	cleanup := func() {
		for _, file := range createdFiles {
			os.Remove(file)
		}
	}

	// Copy AppImage
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

	// Set executable permissions
	if err := os.Chmod(targetAppImagePath, 0755); err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "permissions", Err: err}
	}

	// Handle icon installation
	var targetIconPath string
	if metadata.Icon != nil && len(metadata.Icon.Data) > 0 {
		iconExt := "." + metadata.Icon.Format
		targetIconPath, err = GetIconPath(sanitizedName, iconExt)
		if err != nil {
			cleanup()
			return &InstallError{AppName: sanitizedName, Stage: "icon", Err: err}
		}

		if err := os.WriteFile(targetIconPath, metadata.Icon.Data, 0644); err != nil {
			// Don't fail installation if icon fails, just log and continue
			targetIconPath = ""
		} else {
			createdFiles = append(createdFiles, targetIconPath)
		}
	}

	// Create modified desktop file
	originalFilename := filepath.Base(metadata.FilePath)
	modifiedDesktop, err := ModifyDesktopFile(
		metadata.DesktopFileContent,
		targetAppImagePath,
		targetIconPath,
		originalFilename,
	)
	if err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "desktop", Err: err}
	}

	// Write desktop file
	if err := os.WriteFile(targetDesktopPath, modifiedDesktop, 0644); err != nil {
		cleanup()
		return &InstallError{AppName: sanitizedName, Stage: "desktop", Err: err}
	}
	createdFiles = append(createdFiles, targetDesktopPath)

	return nil
}

// Uninstall removes all installed files
func Uninstall(appName string) error {
	sanitizedName := SanitizeAppName(appName)

	// Verify app is installed
	info, err := GetInstalledInfo(sanitizedName)
	if err != nil {
		return err
	}

	// Remove AppImage
	if info.InstalledPath != "" {
		if err := os.Remove(info.InstalledPath); err != nil && !os.IsNotExist(err) {
			return &InstallError{AppName: sanitizedName, Stage: "remove", Err: err}
		}
	}

	// Remove desktop file
	if info.DesktopFilePath != "" {
		if err := os.Remove(info.DesktopFilePath); err != nil && !os.IsNotExist(err) {
			return &InstallError{AppName: sanitizedName, Stage: "remove", Err: err}
		}
	}

	// Remove icon (try all common extensions)
	iconsDir, err := GetIconsPath()
	if err == nil {
		for _, ext := range []string{".png", ".svg", ".jpg", ".xpm"} {
			iconPath := filepath.Join(iconsDir, sanitizedName+ext)
			os.Remove(iconPath) // Ignore errors
		}
	}

	return nil
}
