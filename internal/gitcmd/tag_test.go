package gitcmd

import (
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
)

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		maxPatch       config.Patch
		expected       string
	}{
		{
			name:           "increment patch",
			currentVersion: "v1.2.3",
			maxPatch:       config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected:       "v1.2.4",
		},
		{
			name:           "patch overflow to minor",
			currentVersion: "v1.2.9",
			maxPatch:       config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 9, Suffix: ""},
			expected:       "v1.3.0",
		},
		{
			name:           "minor overflow to major",
			currentVersion: "v1.9.9",
			maxPatch:       config.Patch{Prefix: "v", Major: 99, Minor: 9, Patch: 9, Suffix: ""},
			expected:       "v2.0.0",
		},
		{
			name:           "with suffix",
			currentVersion: "v1.2.3-beta",
			maxPatch:       config.Patch{Prefix: "v", Major: 99, Minor: 99, Patch: 99, Suffix: "-beta"},
			expected:       "v1.2.4-beta",
		},
		{
			name:           "no prefix or suffix",
			currentVersion: "1.2.3",
			maxPatch:       config.Patch{Prefix: "", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected:       "1.2.4",
		},
		{
			name:           "invalid version format",
			currentVersion: "invalid",
			maxPatch:       config.Patch{Prefix: "", Major: 99, Minor: 99, Patch: 99, Suffix: ""},
			expected:       "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 每个子测试使用独立的临时数据库
			tmpDir := t.TempDir()
			testDBPath := filepath.Join(tmpDir, "test.db")

			// 设置测试数据库路径
			config.SetTestDBPath(testDBPath)
			defer config.SetTestDBPath("")

			// 初始化数据库
			if err := config.Initialize(); err != nil {
				t.Fatalf("failed to initialize: %v", err)
			}

			// 保存配置
			if err := config.SavePatches([]config.Patch{tt.maxPatch}); err != nil {
				t.Fatalf("failed to save patch config: %v", err)
			}

			result := incrementVersion(tt.currentVersion)
			if result != tt.expected {
				t.Errorf("incrementVersion(%q) = %q, want %q", tt.currentVersion, result, tt.expected)
			}
		})
	}
}

func TestVersionRegexParsing(t *testing.T) {
	tests := []struct {
		input      string
		prefix     string
		expected   string
	}{
		{"v1.2.3", "v", "v1.2.4"},
		{"1.2.3", "", "1.2.4"},
		{"v0.0.0", "v", "v0.0.1"},
		{"v10.20.30", "v", "v10.20.31"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// 每个测试使用独立数据库
			tmpDir := t.TempDir()
			config.SetTestDBPath(filepath.Join(tmpDir, "test.db"))
			defer config.SetTestDBPath("")

			// 初始化并设置配置
			config.Initialize()
			config.SavePatches([]config.Patch{{
				Prefix: tt.prefix,
				Major:  99,
				Minor:  99,
				Patch:  99,
				Suffix: "",
			}})

			result := incrementVersion(tt.input)
			if result != tt.expected {
				t.Errorf("incrementVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
