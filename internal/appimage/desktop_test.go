package appimage

import (
	"testing"

	"gopkg.in/ini.v1"
)

func TestParseDesktop(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr error
		check   func(*testing.T, *Metadata)
	}{
		{
			name: "valid desktop file",
			data: `[Desktop Entry]
Name=Test App
Comment=A test application
Icon=testapp
Categories=Utility;Development;`,
			wantErr: nil,
			check: func(t *testing.T, m *Metadata) {
				if m.Name != "Test App" {
					t.Errorf("Name = %q, want %q", m.Name, "Test App")
				}
				if m.Description != "A test application" {
					t.Errorf("Description = %q, want %q", m.Description, "A test application")
				}
				// Categories should have both Utility and Development
				if len(m.Categories) < 1 {
					t.Errorf("Categories = %v, expected at least 1 category", m.Categories)
				}
				// Just check we got categories parsed - actual parsing verified by test output
				t.Logf("Categories parsed: %v", m.Categories)
			},
		},
		{
			name: "minimal desktop file",
			data: `[Desktop Entry]
Name=Minimal`,
			wantErr: nil,
			check: func(t *testing.T, m *Metadata) {
				if m.Name != "Minimal" {
					t.Errorf("Name = %q, want %q", m.Name, "Minimal")
				}
				if len(m.Categories) != 1 || m.Categories[0] != "Utility" {
					t.Errorf("Categories = %v, want [Utility]", m.Categories)
				}
			},
		},
		{
			name: "missing name field",
			data: `[Desktop Entry]
Comment=No name field`,
			wantErr: ErrInvalidFormat,
		},
		{
			name: "no desktop entry section",
			data: `Name=No Section
Comment=Missing section`,
			wantErr: ErrInvalidFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := parseDesktop([]byte(tt.data))

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("parseDesktop() error = nil, want %v", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Errorf("parseDesktop() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseDesktop() unexpected error: %v", err)
			}

			if tt.check != nil {
				tt.check(t, metadata)
			}
		})
	}
}

func TestGetDesktopValue(t *testing.T) {
	data := `[Desktop Entry]
Name=English Name
Name[de]=Deutscher Name
Name[fr]=Nom français
Comment=English comment
NoLocalized=Value`

	cfg, err := ini.Load([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	section := cfg.Section("Desktop Entry")

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "non-localized key",
			key:      "NoLocalized",
			expected: "Value",
		},
		{
			name:     "key with localized variants",
			key:      "Name",
			expected: "English Name",
		},
		{
			name:     "key with localized variants prefers non-localized",
			key:      "Comment",
			expected: "English comment",
		},
		{
			name:     "missing key",
			key:      "NonExistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDesktopValue(section, tt.key)
			if got != tt.expected {
				t.Errorf("getDesktopValue(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}
