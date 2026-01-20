package appimage

import (
	"errors"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "with file path and stage",
			err: &ParseError{
				FilePath: "/path/to/app.AppImage",
				Stage:    "detection",
				Err:      ErrNotAppImage,
			},
			expected: "failed to parse AppImage at /path/to/app.AppImage (stage: detection): file is not an AppImage",
		},
		{
			name: "with wrapped error",
			err: &ParseError{
				FilePath: "/test.AppImage",
				Stage:    "open",
				Err:      errors.New("file not found"),
			},
			expected: "failed to parse AppImage at /test.AppImage (stage: open): file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	parseErr := &ParseError{
		FilePath: "/test.AppImage",
		Stage:    "parsing",
		Err:      innerErr,
	}

	unwrapped := parseErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, innerErr)
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrNotAppImage", ErrNotAppImage, "file is not an AppImage"},
		{"ErrUnsupportedType", ErrUnsupportedType, "AppImage Type 1 is not supported. Please use Type 2 AppImages"},
		{"ErrNoDesktopFile", ErrNoDesktopFile, "no .desktop file found in AppImage"},
		{"ErrCorrupted", ErrCorrupted, "AppImage file is corrupted"},
		{"ErrInvalidFormat", ErrInvalidFormat, "invalid AppImage format"},
		{"ErrNoSquashFS", ErrNoSquashFS, "could not find SquashFS filesystem in AppImage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.msg)
			}
		})
	}
}
