package appimage

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"os"

	"github.com/KarpelesLab/squashfs"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

func init() {
	// Register XZ decompressor for SquashFS
	squashfs.RegisterDecompressor(squashfs.XZ,
		squashfs.MakeDecompressorErr(func(r io.Reader) (io.ReadCloser, error) {
			rc, err := xz.NewReader(r)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(rc), nil
		}),
	)

	// Register ZSTD decompressor for SquashFS
	squashfs.RegisterDecompressor(squashfs.ZSTD,
		squashfs.MakeDecompressorErr(func(r io.Reader) (io.ReadCloser, error) {
			rc, err := zstd.NewReader(r)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(rc), nil
		}),
	)
}

// Metadata contains extracted AppImage metadata
type Metadata struct {
	Name               string
	Description        string
	Categories         []string
	Icon               *IconData
	FilePath           string
	DesktopFileContent []byte
}

// IconData contains icon image data
type IconData struct {
	Data   []byte
	Format string // "png", "svg", "jpg", "xpm"
}

// Parser handles AppImage parsing
type Parser struct {
	filepath       string
	file           *os.File
	squashfsOffset int64
	sqfs           *squashfs.Superblock
}

// Parse is the main entry point for parsing an AppImage
func Parse(filepath string) (*Metadata, error) {
	p := &Parser{filepath: filepath}
	defer func() {
		if p.file != nil {
			p.file.Close()
		}
		if p.sqfs != nil {
			p.sqfs.Close()
		}
	}()

	// Open file
	var err error
	p.file, err = os.Open(filepath)
	if err != nil {
		return nil, &ParseError{FilePath: filepath, Stage: "open", Err: err}
	}

	// Detect AppImage type
	if err := p.detectType(); err != nil {
		return nil, &ParseError{FilePath: filepath, Stage: "detection", Err: err}
	}

	// Find SquashFS offset
	p.squashfsOffset, err = p.findSquashfsOffset()
	if err != nil {
		return nil, &ParseError{FilePath: filepath, Stage: "offset", Err: err}
	}

	// Open SquashFS filesystem
	if err := p.openSquashfs(); err != nil {
		return nil, &ParseError{FilePath: filepath, Stage: "squashfs", Err: err}
	}

	// Find and parse desktop file
	metadata, err := p.parseDesktopFile()
	if err != nil {
		return nil, &ParseError{FilePath: filepath, Stage: "desktop", Err: err}
	}

	metadata.FilePath = filepath

	// Resolve and extract icon if specified
	if metadata.Icon != nil && len(metadata.Icon.Data) == 0 {
		// Icon name was set by desktop parser, but we need to extract the actual data
		iconName := metadata.Icon.Format // Temporarily stored in Format field
		iconData, err := p.resolveIcon(iconName)
		if err != nil {
			// Don't fail if icon is missing, just log and continue
			metadata.Icon = nil
		} else {
			metadata.Icon = iconData
		}
	}

	return metadata, nil
}

// detectType checks if the file is an AppImage Type 2
func (p *Parser) detectType() error {
	buf := make([]byte, 12)
	_, err := p.file.ReadAt(buf, 0)
	if err != nil {
		return ErrNotAppImage
	}

	// Check for AppImage Type 2 magic bytes at offset 8: "AI\x02"
	if len(buf) >= 11 && buf[8] == 'A' && buf[9] == 'I' && buf[10] == 0x02 {
		return nil // Type 2, proceed
	}

	// Check for Type 1 magic bytes at offset 8: "AI\x01"
	if len(buf) >= 11 && buf[8] == 'A' && buf[9] == 'I' && buf[10] == 0x01 {
		return ErrUnsupportedType
	}

	// Fallback: Check for Type 1 by looking for "ELF" at offset 1 and "CD001" at 32769
	if len(buf) >= 4 && buf[1] == 'E' && buf[2] == 'L' && buf[3] == 'F' {
		// Check for ISO9660 signature
		cd001Buf := make([]byte, 5)
		_, err := p.file.ReadAt(cd001Buf, 32769)
		if err == nil && string(cd001Buf) == "CD001" {
			return ErrUnsupportedType
		}
	}

	return ErrNotAppImage
}

// findSquashfsOffset calculates the SquashFS offset from the ELF header of the runtime.
// The SquashFS image starts directly after the ELF section header table, which ends at
// e_shoff + e_shentsize * e_shnum. Both 32-bit and 64-bit ELF files in either byte order
// are supported. Returns ErrNoSquashFS if the header is invalid or the SquashFS magic
// bytes are not found at the calculated offset.
func (p *Parser) findSquashfsOffset() (int64, error) {
	header := make([]byte, 64)
	if _, err := p.file.ReadAt(header, 0); err != nil {
		return 0, ErrNoSquashFS
	}
	if !bytes.Equal(header[:4], []byte("\x7fELF")) {
		return 0, ErrNoSquashFS
	}

	var order binary.ByteOrder
	switch header[5] {
	case 1:
		order = binary.LittleEndian
	case 2:
		order = binary.BigEndian
	default:
		return 0, ErrNoSquashFS
	}

	var shoff int64
	var shentsize, shnum uint16
	switch header[4] {
	case 1: // ELFCLASS32
		shoff = int64(order.Uint32(header[0x20:]))
		shentsize = order.Uint16(header[0x2E:])
		shnum = order.Uint16(header[0x30:])
	case 2: // ELFCLASS64
		shoff = int64(order.Uint64(header[0x28:]))
		shentsize = order.Uint16(header[0x3A:])
		shnum = order.Uint16(header[0x3C:])
	default:
		return 0, ErrNoSquashFS
	}
	offset := shoff + int64(shentsize)*int64(shnum)

	// SquashFS magic bytes: "hsqs" = 0x68 0x73 0x71 0x73
	magic := make([]byte, 4)
	if _, err := p.file.ReadAt(magic, offset); err != nil || !bytes.Equal(magic, []byte("hsqs")) {
		return 0, ErrNoSquashFS
	}

	return offset, nil
}

// openSquashfs initializes the SquashFS reader at the detected offset
func (p *Parser) openSquashfs() error {
	// Create a section reader starting at the squashfs offset
	fileInfo, err := p.file.Stat()
	if err != nil {
		return err
	}

	size := fileInfo.Size() - p.squashfsOffset
	sectionReader := io.NewSectionReader(p.file, p.squashfsOffset, size)

	// Initialize squashfs reader
	sqfs, err := squashfs.New(sectionReader)
	if err != nil {
		return ErrCorrupted
	}

	p.sqfs = sqfs
	return nil
}

// parseDesktopFile finds and parses the .desktop file
func (p *Parser) parseDesktopFile() (*Metadata, error) {
	// Find .desktop file in root directory
	desktopFileName, err := p.findDesktopFile()
	if err != nil {
		return nil, err
	}

	// Read desktop file contents
	desktopData, err := fs.ReadFile(p.sqfs, desktopFileName)
	if err != nil {
		return nil, err
	}

	// Parse desktop file and extract metadata
	metadata, err := parseDesktop(desktopData)
	if err != nil {
		return nil, err
	}

	// Store raw desktop file content for installation
	metadata.DesktopFileContent = desktopData

	return metadata, nil
}

// findDesktopFile locates the .desktop file in the root directory
func (p *Parser) findDesktopFile() (string, error) {
	entries, err := fs.ReadDir(p.sqfs, ".")
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 8 && entry.Name()[len(entry.Name())-8:] == ".desktop" {
			return entry.Name(), nil
		}
	}

	return "", ErrNoDesktopFile
}
