package appimage

import (
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
	homeDir, err := GetApplicationsPath()
	if err != nil {
		t.Fatalf("Failed to get applications path: %v", err)
	}

	expectedPath := homeDir + "/test-app.AppImage"
	differentPath := "/tmp/test-app.AppImage"

	result = IsInstalledPath("test-app", differentPath)
	if result {
		t.Error("Expected IsInstalledPath to return false for different path")
	}

	// Test that same path matches (even if file doesn't exist)
	result = IsInstalledPath("test-app", expectedPath)
	if !result {
		t.Error("Expected IsInstalledPath to return true for matching path")
	}
}
