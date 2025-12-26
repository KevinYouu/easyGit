package config

import (
	"path/filepath"
	"testing"
)

func TestGetTagPatch(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	// 初始化数据库
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 获取默认配置
	patch, err := GetTagPatch()
	if err != nil {
		t.Fatalf("GetTagPatch failed: %v", err)
	}

	// 验证默认值
	defaults := GetDefaultTagPatch()[0]
	if patch.Prefix != defaults.Prefix {
		t.Errorf("expected prefix %q, got %q", defaults.Prefix, patch.Prefix)
	}
	if patch.Major != defaults.Major {
		t.Errorf("expected major %d, got %d", defaults.Major, patch.Major)
	}
}

func TestGetDefaultTagPatch(t *testing.T) {
	defaults := GetDefaultTagPatch()

	if len(defaults) != 1 {
		t.Errorf("expected 1 default patch, got %d", len(defaults))
	}

	patch := defaults[0]
	if patch.Major != 999 {
		t.Errorf("expected major 999, got %d", patch.Major)
	}
	if patch.Minor != 9 {
		t.Errorf("expected minor 9, got %d", patch.Minor)
	}
	if patch.Patch != 9 {
		t.Errorf("expected patch 9, got %d", patch.Patch)
	}
}

func TestSavePatchesMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	Initialize()

	// 第一次保存
	patch1 := Patch{Prefix: "v", Major: 1, Minor: 2, Patch: 3, Suffix: ""}
	if err := SavePatches([]Patch{patch1}); err != nil {
		t.Fatalf("first SavePatches failed: %v", err)
	}

	// 第二次保存(应该替换而不是追加)
	patch2 := Patch{Prefix: "r", Major: 4, Minor: 5, Patch: 6, Suffix: "-beta"}
	if err := SavePatches([]Patch{patch2}); err != nil {
		t.Fatalf("second SavePatches failed: %v", err)
	}

	// 验证只有最新的配置
	result, err := GetTagPatch()
	if err != nil {
		t.Fatalf("GetTagPatch failed: %v", err)
	}

	if result.Prefix != "r" {
		t.Errorf("expected prefix 'r', got %q", result.Prefix)
	}
	if result.Major != 4 {
		t.Errorf("expected major 4, got %d", result.Major)
	}
}

func TestSetTestDBPath(t *testing.T) {
	originalPath := getDBPath()

	// 设置测试路径
	testPath := "/tmp/test.db"
	SetTestDBPath(testPath)

	if getDBPath() != testPath {
		t.Errorf("expected path %q, got %q", testPath, getDBPath())
	}

	// 重置
	SetTestDBPath("")
	resetPath := getDBPath()

	if resetPath == testPath {
		t.Error("path should be reset after SetTestDBPath(\"\")")
	}

	// 应该恢复到默认路径(包含 .easyGit.db)
	if resetPath != originalPath {
		t.Logf("path changed from %q to %q", originalPath, resetPath)
	}
}
