package appimage

import (
	"testing"
)

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		filename string
		expected string
	}{
		{
			name:     "PNG magic bytes",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			filename: "test.png",
			expected: "png",
		},
		{
			name:     "JPEG magic bytes",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
			filename: "test.jpg",
			expected: "jpg",
		},
		{
			name:     "SVG with <svg tag",
			data:     []byte("<svg xmlns=\"http://www.w3.org/2000/svg\">"),
			filename: "test.svg",
			expected: "svg",
		},
		{
			name:     "SVG with <?xml tag",
			data:     []byte("<?xml version=\"1.0\"?>\n<svg>"),
			filename: "test.svg",
			expected: "svg",
		},
		{
			name:     "extension fallback - png",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			filename: "test.png",
			expected: "png",
		},
		{
			name:     "extension fallback - jpg",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			filename: "test.jpg",
			expected: "jpg",
		},
		{
			name:     "extension fallback - jpeg",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			filename: "test.jpeg",
			expected: "jpg",
		},
		{
			name:     "extension fallback - svg",
			data:     []byte("not really svg"),
			filename: "test.svg",
			expected: "svg",
		},
		{
			name:     "extension fallback - xpm",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			filename: "test.xpm",
			expected: "xpm",
		},
		{
			name:     "unknown format",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			filename: "test.bin",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectImageFormat(tt.data, tt.filename)
			if got != tt.expected {
				t.Errorf("detectImageFormat() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHasImageExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.png", true},
		{"test.PNG", true},
		{"test.jpg", true},
		{"test.JPG", true},
		{"test.jpeg", true},
		{"test.JPEG", true},
		{"test.svg", true},
		{"test.SVG", true},
		{"test.xpm", true},
		{"test.XPM", true},
		{"test.txt", false},
		{"test.bin", false},
		{"test", false},
		{"test.png.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := hasImageExtension(tt.filename)
			if got != tt.expected {
				t.Errorf("hasImageExtension(%q) = %v, want %v", tt.filename, got, tt.expected)
			}
		})
	}
}
