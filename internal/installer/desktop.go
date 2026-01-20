package installer

import (
	"bytes"
	"os"

	"gopkg.in/ini.v1"
)

// ModifyDesktopFile updates Exec and Icon paths, adds X-OriginalFilename
func ModifyDesktopFile(originalContent []byte, execPath, iconPath, originalFilename string) ([]byte, error) {
	cfg, err := ini.Load(originalContent)
	if err != nil {
		return nil, err
	}

	section := cfg.Section("Desktop Entry")
	if section == nil {
		return nil, ErrNotInstalled
	}

	// Update Exec path
	if execPath != "" {
		section.Key("Exec").SetValue(execPath)
	}

	// Update Icon path
	if iconPath != "" {
		section.Key("Icon").SetValue(iconPath)
	}

	// Add X-OriginalFilename
	if originalFilename != "" {
		section.Key("X-OriginalFilename").SetValue(originalFilename)
	}

	// Write to buffer
	var buf bytes.Buffer
	_, err = cfg.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetOriginalFilename reads X-OriginalFilename from installed desktop file
func GetOriginalFilename(desktopFilePath string) (string, error) {
	data, err := os.ReadFile(desktopFilePath)
	if err != nil {
		return "", err
	}

	cfg, err := ini.Load(data)
	if err != nil {
		return "", err
	}

	section := cfg.Section("Desktop Entry")
	if section == nil {
		return "", nil
	}

	if section.HasKey("X-OriginalFilename") {
		return section.Key("X-OriginalFilename").String(), nil
	}

	return "", nil
}
