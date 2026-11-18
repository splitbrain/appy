package appimage

import (
	"encoding/binary"
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

// writeELFTestFile creates a temporary file with a 64-bit little-endian ELF header whose
// section header table ends at shoff + 64*shnum, followed by the given content at the
// given offsets. Returns the opened file, which is closed when the test ends.
func writeELFTestFile(t *testing.T, shoff uint64, shnum uint16, content map[int64][]byte) *os.File {
	t.Helper()

	header := make([]byte, 64)
	copy(header, "\x7fELF")
	header[4] = 2 // ELFCLASS64
	header[5] = 1 // little endian
	copy(header[8:], "AI\x02")
	binary.LittleEndian.PutUint64(header[0x28:], shoff)
	binary.LittleEndian.PutUint16(header[0x3A:], 64)
	binary.LittleEndian.PutUint16(header[0x3C:], shnum)

	f, err := os.CreateTemp(t.TempDir(), "elf-*.AppImage")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })

	if _, err := f.WriteAt(header, 0); err != nil {
		t.Fatal(err)
	}
	for offset, data := range content {
		if _, err := f.WriteAt(data, offset); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

func TestFindSquashfsOffset(t *testing.T) {
	// Section header table ends at 1000 + 64*10 = 1640
	const want = 1640

	// Decoy magic bytes before and after the real SquashFS superblock must be ignored
	f := writeELFTestFile(t, 1000, 10, map[int64][]byte{
		200:      []byte("hsqs"),
		want:     []byte("hsqs"),
		want + 9: []byte("hsqs"),
	})

	p := &Parser{file: f}
	got, err := p.findSquashfsOffset()
	if err != nil {
		t.Fatalf("findSquashfsOffset() failed: %v", err)
	}
	if got != want {
		t.Errorf("findSquashfsOffset() = %d, want %d", got, want)
	}
}

func TestFindSquashfsOffset_NoMagic(t *testing.T) {
	// Magic bytes exist in the file, but not at the calculated offset
	f := writeELFTestFile(t, 1000, 10, map[int64][]byte{
		200:  []byte("hsqs"),
		1700: []byte("hsqs"),
	})

	p := &Parser{file: f}
	if _, err := p.findSquashfsOffset(); err != ErrNoSquashFS {
		t.Errorf("findSquashfsOffset() error = %v, want ErrNoSquashFS", err)
	}
}

func TestFindSquashfsOffset_NotELF(t *testing.T) {
	f := writeELFTestFile(t, 1000, 10, map[int64][]byte{
		0:    []byte("NOPE"),
		1640: []byte("hsqs"),
	})

	p := &Parser{file: f}
	if _, err := p.findSquashfsOffset(); err != ErrNoSquashFS {
		t.Errorf("findSquashfsOffset() error = %v, want ErrNoSquashFS", err)
	}
}
