# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**Update this file automatically when making changes to the codebase.**

## Git Commit Guidelines

- **Do NOT add Claude attributions** to commit messages or Co-Authored-By lines
- Keep commit messages concise and focused on the technical changes

## Project Overview

This is a Go application that parses AppImage files (Type 2), displays their metadata in a GUI, and provides installation/uninstallation functionality. The tool extracts information from the embedded SquashFS filesystem, parses the .desktop file, resolves icons, and presents the data using the Fyne UI framework. Users can install AppImages to their local ~/Applications directory with proper desktop integration.

## Development Commands

### Build
```bash
# Using Make
make build

# Or directly with Go
go build -o appy .
```

### Run
```bash
./appy <path-to-appimage-file>
# Example: ./appy test.AppImage

# Self-installation (registers as default AppImage handler)
./appy --install
./appy --uninstall
```

This is a GUI application, you will need to take a screenshot to see the output. You need to kill the application manually after viewing.

### Test
```bash
# Run tests (automatically downloads test AppImage if missing)
make test
```

Update and extend the tests when making changes to the codebase.

### Screenshots

```bash
# take a screenshot of the running application
import -window "$(xdotool search --name '^Appy - ')" /tmp/screenshot.png
```

### Dependencies
```bash
go mod download
go mod tidy
```

## Architecture

### Package Structure

- `main.go`: Entry point, handles CLI arguments and orchestrates parsing/display
- `selfinstall.go`: Self-installation to register as default AppImage handler
- `ui.go`: Fyne-based GUI, metadata display window, install/uninstall UI, dialogs
- `appimage/`: AppImage parsing and installation logic
  - `appimage.go`: Core parsing, SquashFS detection and mounting, type definitions
  - `desktop.go`: .desktop file parsing and modification (read and write)
  - `icon.go`: Icon resolution and extraction from SquashFS
  - `install.go`: Installation and uninstallation functions, file copying, cleanup
  - `paths.go`: Path resolution for installation directories
  - `errors.go`: All error types (parsing and installation)

### AppImage Parsing Flow

1. **Type Detection** (appimage.go): Checks magic bytes at offset 8 for "AI\x02" (Type 2) or "AI\x01" (Type 1, unsupported)
2. **SquashFS Location** (appimage.go): Searches first 10MB for SquashFS magic bytes (0x68 0x73 0x71 0x73), collects all candidates, and uses the last one found
3. **SquashFS Mounting** (appimage.go): Creates io.SectionReader starting at detected offset
4. **Desktop File Parsing** (appimage.go): Locates .desktop file in root directory, parses using gopkg.in/ini.v1
5. **Icon Resolution** (icon.go): Multi-stage icon search:
   - Direct path if extension provided
   - Root directory with common extensions (.png, .svg, .jpg, .xpm)
   - .DirIcon file
   - usr/share/icons/hicolor/ (multiple sizes: 256, 128, 64)
   - usr/share/pixmaps/
   - Recursive walk through icon directories as fallback

### Metadata Structure

```go
type Metadata struct {
    Name               string      // From Name field in .desktop
    Description        string      // From Comment field in .desktop
    Categories         []string    // Semicolon-separated from Categories field
    Icon               *IconData   // Resolved icon image data
    FilePath           string      // Original AppImage path
    DesktopFileContent []byte      // Raw .desktop file content for installation
}
```

### Installation Flow

1. **Path Resolution** (install.go, paths.go): Determines installation directories
   - AppImages: `~/Applications/{sanitized-name}.AppImage`
   - Desktop files: `~/.local/share/applications/{sanitized-name}.desktop`
   - Icons: `~/.local/share/icons/{sanitized-name}.{ext}`
2. **Name Sanitization** (install.go): Converts app names to filesystem-safe identifiers
   - Lowercase conversion
   - Replaces spaces/special characters with hyphens
   - Collapses multiple hyphens
3. **Installation** (install.go):
   - Creates necessary directories (mode 0755)
   - Copies AppImage to installation directory
   - Sets executable permissions (mode 0755)
   - Extracts and saves icon
   - Modifies desktop file with installed paths
   - Tracks created files for rollback on failure
4. **Desktop File Modification** (desktop.go): Updates .desktop file for installation
   - Rewrites Exec= line to point to installed AppImage path
   - Updates Icon= line to point to installed icon path
   - Adds X-OriginalFilename field with original filename
5. **Uninstallation** (install.go):
   - Removes AppImage file
   - Removes desktop file
   - Removes icon files (all common extensions)
6. **UI Integration** (ui.go):
   - Shows Install/Uninstall button based on installation status
   - Displays warning when installation will overwrite existing version
   - Shows original filename of currently installed version
   - Confirmation dialogs for overwrite and uninstall
   - Quits application after successful install/uninstall

### Self-Installation Flow

The `--install` and `--uninstall` flags allow Appy to register itself as the default AppImage handler:

1. **Install** (selfinstall.go):
   - Copies current executable to `~/Applications/appy`
   - Creates `~/.local/share/applications/appy.desktop` with MIME type associations
   - Registers for `application/x-iso9660-appimage`, `application/vnd.appimage`, and `application/x-appimage` MIME types
   - Enables double-click to open AppImages in file managers
2. **Uninstall** (selfinstall.go):
   - Removes `~/Applications/appy` binary
   - Removes desktop file

### Key Dependencies

- `github.com/KarpelesLab/squashfs`: SquashFS filesystem reading
- `github.com/ulikunitz/xz`: XZ decompression for SquashFS
- `github.com/klauspost/compress/zstd`: ZSTD decompression for SquashFS
- `fyne.io/fyne/v2`: Cross-platform GUI framework
- `gopkg.in/ini.v1`: INI file parsing for .desktop files

### Important Implementation Details

**AppImage Parsing:**
- Only AppImage Type 2 is supported (Type 1 returns ErrUnsupportedType)
- Supports both XZ and ZSTD compression for SquashFS filesystems (registered in appimage.go init())
- Desktop file parsing prefers non-localized keys but falls back to localized variants (e.g., Name[de])
- Icon search prioritizes larger icon sizes (256x256 > 128x128 > 64x64)
- Image format detection uses magic bytes first, then falls back to file extension
- SquashFS offset search returns the last candidate found (typically the correct one for Type 2 AppImages)
- Raw desktop file content is preserved in Metadata for installation use

**Installation:**
- Uses XDG Base Directory standard for installation paths
- App names are sanitized to filesystem-safe identifiers (lowercase, alphanumeric with hyphens)
- Installation provides automatic rollback on failure by tracking created files
- Icon extraction fails gracefully - installation continues without icon if extraction fails
- Desktop file is modified to use absolute paths for installed AppImage and icon
- Original filename is preserved in desktop file as X-AppImage-Original-Filename comment

**UI Behavior:**
- Window title shows "Appy - {app_name}" instead of generic "Appy"
- Description field is hidden when empty to reduce clutter
- Displays only the filename (not full path) in the metadata view
- Install button changes to Uninstall when viewing an installed AppImage
- Shows warning with original filename when installation will overwrite existing version
- Application quits automatically after successful install/uninstall operations
- Confirmation dialogs prevent accidental overwrites and uninstalls

### Error Handling

All errors are defined in `appimage/errors.go`:

**Parsing Errors:**
- `ErrNotAppImage`: File is not an AppImage format
- `ErrUnsupportedType`: AppImage Type 1 encountered
- `ErrNoDesktopFile`: Missing .desktop file in root
- `ErrCorrupted`: SquashFS filesystem unreadable
- `ErrInvalidFormat`: Invalid .desktop file format
- `ErrNoSquashFS`: SquashFS magic bytes not found
- `ParseError`: Wraps errors with file path and parsing stage context

**Installation Errors:**
- `ErrAlreadyInstalled`: Application is already installed
- `ErrNotInstalled`: Application is not installed (for uninstall)
- `ErrPermissionDenied`: Insufficient permissions for installation
- `ErrNoIcon`: No icon available to install
- `InstallError`: Wraps installation errors with app name and stage context (validation, paths, mkdir, copy, permissions, icon, desktop, remove)
