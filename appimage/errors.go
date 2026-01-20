package appimage

import (
	"errors"
	"fmt"
)

// Parsing errors
var (
	ErrNotAppImage     = errors.New("file is not an AppImage")
	ErrUnsupportedType = errors.New("AppImage Type 1 is not supported. Please use Type 2 AppImages")
	ErrNoDesktopFile   = errors.New("no .desktop file found in AppImage")
	ErrCorrupted       = errors.New("AppImage file is corrupted")
	ErrInvalidFormat   = errors.New("invalid AppImage format")
	ErrNoSquashFS      = errors.New("could not find SquashFS filesystem in AppImage")
)

// Installation errors
var (
	ErrAlreadyInstalled = errors.New("application is already installed")
	ErrNotInstalled     = errors.New("application is not installed")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNoIcon           = errors.New("no icon available to install")
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

// InstallError provides context for installation failures
type InstallError struct {
	AppName string
	Stage   string // "copy", "desktop", "icon", "permissions"
	Err     error
}

func (e *InstallError) Error() string {
	return fmt.Sprintf("failed to install %s (stage: %s): %v",
		e.AppName, e.Stage, e.Err)
}

func (e *InstallError) Unwrap() error {
	return e.Err
}
