package appimage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeAppName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Application", "my-application"},
		{"App (2024)", "app-2024"},
		{"Test_App!", "test-app"},
		{"Firefox Web Browser", "firefox-web-browser"},
		{"VLC Media Player", "vlc-media-player"},
		{"Simple", "simple"},
		{"___Test___", "test"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"Special!@#$%Chars", "special-chars"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeAppName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeAppName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetInstalledAppImagePath(t *testing.T) {
	path, err := GetInstalledAppImagePath("test-app")
	if err != nil {
		t.Fatalf("GetInstalledAppImagePath failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty path")
	}

	if len(path) < 9 || path[len(path)-9:] != ".AppImage" {
		t.Errorf("Expected path to end with .AppImage, got %s", path)
	}
}

func TestIsInstalled(t *testing.T) {
	// This should return false for a non-existent app
	result := IsInstalled("non-existent-app-12345")
	if result {
		t.Error("Expected IsInstalled to return false for non-existent app")
	}
}

func TestIsInstalledPath(t *testing.T) {
	// Test with a non-existent app and arbitrary path
	result := IsInstalledPath("non-existent-app-12345", "/some/random/path.AppImage")
	if result {
		t.Error("Expected IsInstalledPath to return false for non-existent app")
	}

	// Test that different paths don't match
	result = IsInstalledPath("test-app", "/tmp/test-app.AppImage")
	if result {
		t.Error("Expected IsInstalledPath to return false for different path")
	}

	// Test that same path matches (even if file doesn't exist)
	expectedPath, err := GetInstalledAppImagePath("test-app")
	if err != nil {
		t.Fatalf("Failed to get installed path: %v", err)
	}
	result = IsInstalledPath("test-app", expectedPath)
	if !result {
		t.Error("Expected IsInstalledPath to return true for matching path")
	}
}

func TestGetDesktopFilePath(t *testing.T) {
	path, err := GetDesktopFilePath("test-app")
	if err != nil {
		t.Fatalf("GetDesktopFilePath failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty path")
	}

	if !strings.HasSuffix(path, ".desktop") {
		t.Errorf("Expected path to end with .desktop, got %s", path)
	}

	if !strings.Contains(path, ".local/share/applications") {
		t.Errorf("Expected path to contain .local/share/applications, got %s", path)
	}
}

func TestGetIconPath(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		ext      string
		wantExt  string
	}{
		{"with dot prefix", "test-app", ".png", ".png"},
		{"without dot prefix", "test-app", "svg", ".svg"},
		{"jpg extension", "myapp", "jpg", ".jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetIconPath(tt.appName, tt.ext)
			if err != nil {
				t.Fatalf("GetIconPath failed: %v", err)
			}

			if !strings.HasSuffix(path, tt.wantExt) {
				t.Errorf("Expected path to end with %s, got %s", tt.wantExt, path)
			}

			if !strings.Contains(path, ".local/share/icons") {
				t.Errorf("Expected path to contain .local/share/icons, got %s", path)
			}

			if !strings.Contains(path, tt.appName) {
				t.Errorf("Expected path to contain app name %s, got %s", tt.appName, path)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content for copy")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test successful copy
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := CopyFile(srcPath, dstPath, 0755); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file exists and has correct content
	destContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(destContent) != string(content) {
		t.Errorf("CopyFile content mismatch: got %q, want %q", destContent, content)
	}

	// Verify permissions
	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("CopyFile permissions = %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestCopyFile_Errors(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-existent source
	err := CopyFile("/non/existent/source", filepath.Join(tmpDir, "dest"), 0644)
	if err == nil {
		t.Error("CopyFile expected error for non-existent source")
	}

	// Test invalid destination path
	srcPath := filepath.Join(tmpDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	err = CopyFile(srcPath, "/non/existent/dir/dest.txt", 0644)
	if err == nil {
		t.Error("CopyFile expected error for invalid destination path")
	}
}

func TestInstall(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	tmpDir := t.TempDir()

	// Create a fake AppImage file
	appImagePath := filepath.Join(tmpDir, "TestApp-1.0.AppImage")
	if err := os.WriteFile(appImagePath, []byte("fake appimage content"), 0755); err != nil {
		t.Fatal(err)
	}

	desktopContent := []byte(`[Desktop Entry]
Name=Test Install App
Exec=testapp
Icon=testapp
Categories=Utility;`)

	metadata := &Metadata{
		Name:               "Test Install App",
		FilePath:           appImagePath,
		DesktopFileContent: desktopContent,
		Icon: &IconData{
			Data:   []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
			Format: "png",
		},
	}

	// Install
	if err := Install(metadata); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify files were created
	sanitizedName := SanitizeAppName("Test Install App")

	appImageInstallPath := filepath.Join(homeDir, "Applications", sanitizedName+".AppImage")
	if _, err := os.Stat(appImageInstallPath); os.IsNotExist(err) {
		t.Errorf("AppImage not installed at %s", appImageInstallPath)
	}

	desktopPath := filepath.Join(homeDir, ".local/share/applications", sanitizedName+".desktop")
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		t.Errorf("Desktop file not created at %s", desktopPath)
	}

	iconPath := filepath.Join(homeDir, ".local/share/icons", sanitizedName+".png")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		t.Errorf("Icon not created at %s", iconPath)
	}

	// Verify IsInstalled returns true
	if !IsInstalled("Test Install App") {
		t.Error("IsInstalled should return true after installation")
	}
}

func TestInstall_Validation(t *testing.T) {
	// Test nil metadata
	err := Install(nil)
	if err == nil {
		t.Error("Install should fail with nil metadata")
	}

	// Test empty name
	err = Install(&Metadata{Name: "   ", FilePath: "/some/path"})
	if err == nil {
		t.Error("Install should fail with empty name")
	}
}

func TestUninstall(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	tmpDir := t.TempDir()

	// Create and install a test app
	appImagePath := filepath.Join(tmpDir, "UninstallTest.AppImage")
	if err := os.WriteFile(appImagePath, []byte("fake appimage"), 0755); err != nil {
		t.Fatal(err)
	}

	desktopContent := []byte(`[Desktop Entry]
Name=Uninstall Test App
Exec=testapp
Categories=Utility;`)

	metadata := &Metadata{
		Name:               "Uninstall Test App",
		FilePath:           appImagePath,
		DesktopFileContent: desktopContent,
	}

	if err := Install(metadata); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify it was installed
	if !IsInstalled("Uninstall Test App") {
		t.Fatal("App should be installed")
	}

	// Uninstall
	if err := Uninstall("Uninstall Test App"); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Verify files were removed
	sanitizedName := SanitizeAppName("Uninstall Test App")

	appImageInstallPath := filepath.Join(homeDir, "Applications", sanitizedName+".AppImage")
	if _, err := os.Stat(appImageInstallPath); !os.IsNotExist(err) {
		t.Errorf("AppImage should be removed from %s", appImageInstallPath)
	}

	// Verify IsInstalled returns false
	if IsInstalled("Uninstall Test App") {
		t.Error("IsInstalled should return false after uninstallation")
	}
}

func TestUninstall_NotInstalled(t *testing.T) {
	err := Uninstall("non-existent-app-for-uninstall-test-12345")
	if err != ErrNotInstalled {
		t.Errorf("Uninstall should return ErrNotInstalled for non-existent app, got %v", err)
	}
}
