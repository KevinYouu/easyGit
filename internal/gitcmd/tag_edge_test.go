package gitcmd

import (
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
)

func TestIncrementVersionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		patch    config.Patch
		expected string
	}{
		{
			name:     "version with v prefix",
			version:  "v1.2.3",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected: "v1.2.4",
		},
		{
			name:     "version without prefix",
			version:  "1.2.3",
			patch:    config.Patch{Prefix: "", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected: "1.2.4",
		},
		{
			name:     "version with suffix",
			version:  "v1.2.3-rc1",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 99, Suffix: "-rc1"},
			expected: "v1.2.4-rc1",
		},
		{
			name:     "large patch number",
			version:  "v1.2.99",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 999, Suffix: ""},
			expected: "v1.2.100",
		},
		{
			name:     "patch overflow with small limit",
			version:  "v1.2.5",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 5, Suffix: ""},
			expected: "v1.3.0",
		},
		{
			name:     "minor overflow",
			version:  "v1.5.5",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 5, Patch: 5, Suffix: ""},
			expected: "v2.0.0",
		},
		{
			name:     "zero version",
			version:  "v0.0.0",
			patch:    config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected: "v0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			config.SetTestDBPath(filepath.Join(tmpDir, "test.db"))
			defer config.SetTestDBPath("")

			config.Initialize()
			config.SavePatches([]config.Patch{tt.patch})

			result := incrementVersion(tt.version)
			if result != tt.expected {
				t.Errorf("incrementVersion(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}
