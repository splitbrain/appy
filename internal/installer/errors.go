package installer

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrAlreadyInstalled = errors.New("application is already installed")
	ErrNotInstalled     = errors.New("application is not installed")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNoIcon           = errors.New("no icon available to install")
)

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
