package config

import (
	"testing"
)

func TestGetConflictEditorEmpty(t *testing.T) {
	newTestDB(t)

	// 未设置时返回空串(调用方按自动检测处理)
	got, err := GetConflictEditor()
	if err != nil {
		t.Fatalf("GetConflictEditor on empty db failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetConflictEditor on empty db = %q, want empty", got)
	}
}

func TestSaveGetConflictEditor(t *testing.T) {
	newTestDB(t)

	// 保存后读取
	if err := SaveConflictEditor("nano"); err != nil {
		t.Fatalf("SaveConflictEditor(nano) failed: %v", err)
	}
	got, err := GetConflictEditor()
	if err != nil {
		t.Fatalf("GetConflictEditor failed: %v", err)
	}
	if got != "nano" {
		t.Errorf("GetConflictEditor after save = %q, want nano", got)
	}

	// UPSERT 覆盖
	if err := SaveConflictEditor("code -w"); err != nil {
		t.Fatalf("SaveConflictEditor(code -w) failed: %v", err)
	}
	got, err = GetConflictEditor()
	if err != nil {
		t.Fatalf("GetConflictEditor failed: %v", err)
	}
	if got != "code -w" {
		t.Errorf("GetConflictEditor after overwrite = %q, want code -w", got)
	}
}

func TestSaveConflictEditorEmptyClears(t *testing.T) {
	newTestDB(t)

	if err := SaveConflictEditor("vim"); err != nil {
		t.Fatalf("SaveConflictEditor(vim) failed: %v", err)
	}

	// 保存空串 = 清除设置(恢复自动检测)
	if err := SaveConflictEditor(""); err != nil {
		t.Fatalf("SaveConflictEditor(empty) failed: %v", err)
	}
	got, err := GetConflictEditor()
	if err != nil {
		t.Fatalf("GetConflictEditor failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetConflictEditor after clear = %q, want empty", got)
	}
}

// TestBuildConfigOptionsConflictEditor 配置中心主列表包含冲突编辑器项
// (BuildConfigOptions 内部读取 settings 表,测试 DB 为空时摘要显示自动检测)
func TestBuildConfigOptionsConflictEditor(t *testing.T) {
	newTestDB(t)

	opts := BuildConfigOptions()
	for _, o := range opts {
		if o.Value == ConfigKeyConflictEditor {
			if o.Label == "" {
				t.Error("冲突编辑器项 Label 为空")
			}
			return
		}
	}
	t.Error("BuildConfigOptions 缺少冲突编辑器项")
}
