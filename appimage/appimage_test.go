package appimage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	testAppImage := filepath.Join("..", "testdata", "test.AppImage")

	// Check if test file exists
	if _, err := os.Stat(testAppImage); os.IsNotExist(err) {
		t.Skipf("Test AppImage not found at %s, skipping test", testAppImage)
	}

	metadata, err := Parse(testAppImage)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Basic validation
	if metadata == nil {
		t.Fatal("Parse() returned nil metadata")
	}

	if metadata.Name == "" {
		t.Error("metadata.Name is empty")
	}

	if metadata.FilePath != testAppImage {
		t.Errorf("metadata.FilePath = %q, want %q", metadata.FilePath, testAppImage)
	}

	// Log the parsed metadata for inspection
	t.Logf("Parsed metadata:")
	t.Logf("  Name: %s", metadata.Name)
	t.Logf("  Description: %s", metadata.Description)
	t.Logf("  Categories: %v", metadata.Categories)
	if metadata.Icon != nil {
		t.Logf("  Icon Format: %s", metadata.Icon.Format)
		t.Logf("  Icon Size: %d bytes", len(metadata.Icon.Data))
	}
}

func TestParseNonExistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/file.AppImage")
	if err == nil {
		t.Fatal("Parse() should fail for non-existent file")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Expected *ParseError, got %T", err)
	} else {
		if parseErr.Stage != "open" {
			t.Errorf("Expected stage 'open', got %q", parseErr.Stage)
		}
	}
}

func TestParseInvalidFile(t *testing.T) {
	// Create a temporary invalid file
	tmpFile, err := os.CreateTemp("", "invalid-*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write some random data that's not an AppImage
	tmpFile.Write([]byte("This is not an AppImage file"))
	tmpFile.Close()

	_, err = Parse(tmpFile.Name())
	if err == nil {
		t.Fatal("Parse() should fail for invalid file")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("Expected *ParseError, got %T", err)
		return
	}

	if parseErr.Stage != "detection" {
		t.Errorf("Expected stage 'detection', got %q", parseErr.Stage)
	}

	if parseErr.Err != ErrNotAppImage {
		t.Errorf("Expected ErrNotAppImage, got %v", parseErr.Err)
	}
}
