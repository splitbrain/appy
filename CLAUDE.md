# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Git Commit Guidelines

- **Do NOT add Claude attributions** to commit messages or Co-Authored-By lines
- Keep commit messages concise and focused on the technical changes

## Project Overview

This is a Go application that parses AppImage files (Type 2) and displays their metadata in a GUI. The tool extracts information from the embedded SquashFS filesystem, parses the .desktop file, resolves icons, and presents the data using the Fyne UI framework.

## Development Commands

### Build
```bash
# Using Make
make build

# Or directly with Go
go build -o appy ./cmd/appy
```

### Run
```bash
./appy <path-to-appimage-file>
# Example: ./appy test.AppImage
```

### Test
```bash
# Run tests (automatically downloads test AppImage if missing)
make test

# Or directly with Go
go test -v ./...
```

### Dependencies
```bash
go mod download
go mod tidy
```

## Architecture

### Package Structure

- `cmd/appy/main.go`: Entry point, handles CLI arguments and orchestrates parsing/display
- `internal/appimage/`: AppImage parsing logic
  - `parser.go`: Core parsing, SquashFS detection and mounting
  - `desktop.go`: .desktop file parsing using INI format
  - `icon.go`: Icon resolution and extraction from SquashFS
  - `errors.go`: Custom error types
- `internal/ui/`: Fyne-based GUI
  - `window.go`: Metadata display window and error dialogs

### AppImage Parsing Flow

1. **Type Detection** (parser.go:94): Checks magic bytes at offset 8 for "AI\x02" (Type 2) or "AI\x01" (Type 1, unsupported)
2. **SquashFS Location** (parser.go:124): Searches first 10MB for SquashFS magic bytes (0x68 0x73 0x71 0x73), collects all candidates, and uses the last one found
3. **SquashFS Mounting** (parser.go:196): Creates io.SectionReader starting at detected offset
4. **Desktop File Parsing** (parser.go:220): Locates .desktop file in root directory, parses using gopkg.in/ini.v1
5. **Icon Resolution** (icon.go:11): Multi-stage icon search:
   - Direct path if extension provided
   - Root directory with common extensions (.png, .svg, .jpg, .xpm)
   - .DirIcon file
   - usr/share/icons/hicolor/ (multiple sizes: 256, 128, 64)
   - usr/share/pixmaps/
   - Recursive walk through icon directories as fallback

### Metadata Structure

```go
type Metadata struct {
    Name        string      // From Name field in .desktop
    Description string      // From Comment field in .desktop
    Version     string      // X-AppImage-Version or Version field
    Categories  []string    // Semicolon-separated from Categories field
    Icon        *IconData   // Resolved icon image data
    FilePath    string      // Original AppImage path
}
```

### Key Dependencies

- `github.com/KarpelesLab/squashfs`: SquashFS filesystem reading
- `github.com/ulikunitz/xz`: XZ decompression for SquashFS
- `github.com/klauspost/compress/zstd`: ZSTD decompression for SquashFS
- `fyne.io/fyne/v2`: Cross-platform GUI framework
- `gopkg.in/ini.v1`: INI file parsing for .desktop files

### Important Implementation Details

- Only AppImage Type 2 is supported (Type 1 returns ErrUnsupportedType)
- Supports both XZ and ZSTD compression for SquashFS filesystems (registered in parser.go:14-34)
- Desktop file parsing prefers non-localized keys but falls back to localized variants (e.g., Name[de])
- Icon search prioritizes larger icon sizes (256x256 > 128x128 > 64x64)
- Image format detection uses magic bytes first, then falls back to file extension
- SquashFS offset search returns the last candidate found (typically the correct one for Type 2 AppImages)
- Debug output is currently present in parser.go (lines 189, 208-216, 223-237) - remove these fmt.Printf calls in production

### Error Handling

Custom errors defined in errors.go:
- `ErrNotAppImage`: File is not an AppImage format
- `ErrUnsupportedType`: AppImage Type 1 encountered
- `ErrNoDesktopFile`: Missing .desktop file in root
- `ErrCorrupted`: SquashFS filesystem unreadable
- `ErrInvalidFormat`: Invalid .desktop file format
- `ErrNoSquashFS`: SquashFS magic bytes not found

`ParseError` wraps errors with file path and parsing stage context.
