package appimage

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

// resolveIcon attempts to locate and extract the icon file
func (p *Parser) resolveIcon(iconName string) (*IconData, error) {
	if iconName == "" {
		return nil, errors.New("icon name is empty")
	}

	// Remove leading slash if present (icon paths are relative to AppDir root)
	iconName = strings.TrimPrefix(iconName, "/")

	// If iconName already has an extension, try it directly
	if hasImageExtension(iconName) {
		if data, format, err := p.tryReadIcon(iconName); err == nil {
			return &IconData{Data: data, Format: format}, nil
		}
	}

	// Search locations in priority order
	searchPaths := []string{
		// 1. AppDir root with common extensions
		iconName + ".png",
		iconName + ".svg",
		iconName + ".jpg",
		iconName + ".xpm",
		iconName,
		".DirIcon",

		// 2. Icon directories (hicolor theme, common sizes)
		"usr/share/icons/hicolor/256x256/apps/" + iconName + ".png",
		"usr/share/icons/hicolor/256x256/apps/" + iconName + ".svg",
		"usr/share/icons/hicolor/128x128/apps/" + iconName + ".png",
		"usr/share/icons/hicolor/128x128/apps/" + iconName + ".svg",
		"usr/share/icons/hicolor/64x64/apps/" + iconName + ".png",
		"usr/share/icons/hicolor/64x64/apps/" + iconName + ".svg",

		// 3. Pixmaps directory
		"usr/share/pixmaps/" + iconName + ".png",
		"usr/share/pixmaps/" + iconName + ".svg",
		"usr/share/pixmaps/" + iconName + ".xpm",
		"usr/share/pixmaps/" + iconName,
	}

	// Try each search path
	for _, path := range searchPaths {
		if data, format, err := p.tryReadIcon(path); err == nil {
			return &IconData{Data: data, Format: format}, nil
		}
	}

	// Last resort: walk through icon directories to find any match
	iconData, err := p.searchIconRecursive(iconName)
	if err == nil {
		return iconData, nil
	}

	return nil, errors.New("icon not found")
}

// tryReadIcon attempts to read an icon file at the given path
func (p *Parser) tryReadIcon(path string) ([]byte, string, error) {
	data, err := fs.ReadFile(p.sqfs, path)
	if err != nil {
		return nil, "", err
	}

	format := detectImageFormat(data, path)
	return data, format, nil
}

// searchIconRecursive searches for icon files recursively in icon directories
func (p *Parser) searchIconRecursive(iconName string) (*IconData, error) {
	searchDirs := []string{
		"usr/share/icons",
		"usr/share/pixmaps",
	}

	for _, dir := range searchDirs {
		var found *IconData
		err := fs.WalkDir(p.sqfs, dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue searching
			}
			if d.IsDir() {
				return nil
			}

			// Check if filename matches icon name
			baseName := filepath.Base(path)
			baseNameNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

			if baseNameNoExt == iconName || baseName == iconName {
				if data, format, err := p.tryReadIcon(path); err == nil {
					found = &IconData{Data: data, Format: format}
					return fs.SkipAll // Stop searching
				}
			}

			return nil
		})

		if err == nil && found != nil {
			return found, nil
		}
	}

	return nil, errors.New("icon not found in recursive search")
}

// hasImageExtension checks if a filename has a known image extension
func hasImageExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".svg" || ext == ".jpg" || ext == ".jpeg" || ext == ".xpm"
}

// detectImageFormat detects the image format from magic bytes or file extension
func detectImageFormat(data []byte, filename string) string {
	// Check magic bytes first
	if len(data) >= 4 {
		// PNG magic bytes: 89 50 4E 47
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "png"
		}
		// JPEG magic bytes: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "jpg"
		}
	}

	// Check for SVG (text-based, look for <svg or <?xml)
	if len(data) >= 5 {
		dataStr := string(data[:min(100, len(data))])
		if strings.Contains(dataStr, "<svg") || strings.Contains(dataStr, "<?xml") {
			return "svg"
		}
	}

	// Fallback to file extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "png"
	case ".svg":
		return "svg"
	case ".jpg", ".jpeg":
		return "jpg"
	case ".xpm":
		return "xpm"
	default:
		return "unknown"
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
