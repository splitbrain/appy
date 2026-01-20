package appimage

import (
	"bytes"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

// parseDesktop parses .desktop file contents and extracts metadata
func parseDesktop(data []byte) (*Metadata, error) {
	cfg, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	section := cfg.Section("Desktop Entry")
	if section == nil {
		return nil, ErrInvalidFormat
	}

	metadata := &Metadata{
		Categories: []string{},
	}

	// Extract Name (required)
	metadata.Name = getDesktopValue(section, "Name")
	if metadata.Name == "" {
		return nil, ErrInvalidFormat
	}

	// Extract Description (Comment field)
	metadata.Description = getDesktopValue(section, "Comment")

	// Extract Categories (semicolon-separated)
	categoriesStr := getDesktopValue(section, "Categories")
	if categoriesStr != "" {
		parts := strings.Split(categoriesStr, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				metadata.Categories = append(metadata.Categories, part)
			}
		}
	}
	if len(metadata.Categories) == 0 {
		metadata.Categories = []string{"Utility"}
	}

	// Extract Icon name (we'll resolve it later)
	iconName := getDesktopValue(section, "Icon")
	if iconName != "" {
		// Store icon name temporarily in Format field
		// The actual icon data will be resolved by the parser
		metadata.Icon = &IconData{
			Format: iconName,
		}
	}

	return metadata, nil
}

// getDesktopValue retrieves a value from desktop section, preferring non-localized keys
func getDesktopValue(section *ini.Section, key string) string {
	// First try to get non-localized value
	if section.HasKey(key) {
		return section.Key(key).String()
	}

	// If not found, try to find any localized version (e.g., Name[de])
	for _, k := range section.Keys() {
		if strings.HasPrefix(k.Name(), key+"[") {
			return k.String()
		}
	}

	return ""
}

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
