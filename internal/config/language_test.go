package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLanguageSettings(t *testing.T) {
	// 创建临时数据库
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test.db")
	SetTestDBPath(tmpDB)
	defer SetTestDBPath("")

	// 初始化数据库
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// 测试初始状态 - 应该没有语言设置
	lang, err := GetLanguage()
	if err != nil {
		t.Fatalf("GetLanguage() error = %v", err)
	}
	if lang != "" {
		t.Errorf("GetLanguage() = %v, want empty string", lang)
	}

	// 测试保存中文设置
	if err := SaveLanguage("zh"); err != nil {
		t.Fatalf("SaveLanguage(zh) error = %v", err)
	}

	// 验证保存成功
	lang, err = GetLanguage()
	if err != nil {
		t.Fatalf("GetLanguage() after save error = %v", err)
	}
	if lang != "zh" {
		t.Errorf("GetLanguage() = %v, want zh", lang)
	}

	// 测试更新为英文
	if err := SaveLanguage("en"); err != nil {
		t.Fatalf("SaveLanguage(en) error = %v", err)
	}

	// 验证更新成功
	lang, err = GetLanguage()
	if err != nil {
		t.Fatalf("GetLanguage() after update error = %v", err)
	}
	if lang != "en" {
		t.Errorf("GetLanguage() = %v, want en", lang)
	}
}

func TestSaveLanguage_MultipleUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test.db")
	SetTestDBPath(tmpDB)
	defer SetTestDBPath("")

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// 多次更新
	languages := []string{"zh", "en", "zh", "en"}
	for _, lang := range languages {
		if err := SaveLanguage(lang); err != nil {
			t.Fatalf("SaveLanguage(%s) error = %v", lang, err)
		}

		// 验证每次更新
		saved, err := GetLanguage()
		if err != nil {
			t.Fatalf("GetLanguage() error = %v", err)
		}
		if saved != lang {
			t.Errorf("GetLanguage() = %v, want %v", saved, lang)
		}
	}
}

func TestGetLanguage_DatabaseNotExist(t *testing.T) {
	// 使用不存在的数据库路径
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "nonexistent.db")
	SetTestDBPath(tmpDB)
	defer SetTestDBPath("")

	// 不初始化数据库,直接读取
	_, err := GetLanguage()
	if err == nil {
		t.Error("GetLanguage() with non-existent DB should return error")
	}
}

func TestSaveLanguage_EmptyString(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test.db")
	SetTestDBPath(tmpDB)
	defer SetTestDBPath("")

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// 测试保存空字符串
	if err := SaveLanguage(""); err != nil {
		t.Fatalf("SaveLanguage('') error = %v", err)
	}

	// 验证保存成功(即使是空字符串)
	lang, err := GetLanguage()
	if err != nil {
		t.Fatalf("GetLanguage() error = %v", err)
	}
	if lang != "" {
		t.Errorf("GetLanguage() = %v, want empty string", lang)
	}
}

func TestLanguage_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDB := filepath.Join(tmpDir, "test.db")
	SetTestDBPath(tmpDB)
	defer SetTestDBPath("")

	// 初始化并设置语言
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if err := SaveLanguage("zh"); err != nil {
		t.Fatalf("SaveLanguage(zh) error = %v", err)
	}

	// 验证数据库文件存在
	if _, err := os.Stat(tmpDB); os.IsNotExist(err) {
		t.Error("Database file should exist after SaveLanguage")
	}

	// 重新读取验证持久化
	lang, err := GetLanguage()
	if err != nil {
		t.Fatalf("GetLanguage() error = %v", err)
	}
	if lang != "zh" {
		t.Errorf("GetLanguage() = %v, want zh (should persist)", lang)
	}
}
