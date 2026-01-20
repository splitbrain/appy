package appimage

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNotAppImage     = errors.New("file is not an AppImage")
	ErrUnsupportedType = errors.New("AppImage Type 1 is not supported. Please use Type 2 AppImages")
	ErrNoDesktopFile   = errors.New("no .desktop file found in AppImage")
	ErrCorrupted       = errors.New("AppImage file is corrupted")
	ErrInvalidFormat   = errors.New("invalid AppImage format")
	ErrNoSquashFS      = errors.New("could not find SquashFS filesystem in AppImage")
)

// ParseError provides context for parsing failures
type ParseError struct {
	FilePath string
	Stage    string // "detection", "offset", "desktop", "icon"
	Err      error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse AppImage at %s (stage: %s): %v",
		e.FilePath, e.Stage, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
