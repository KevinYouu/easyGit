package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// newTestDB 初始化临时数据库,返回清理函数
func newTestDB(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	t.Cleanup(func() { testDBPath = "" })
	if err := Initialize(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
}

func TestAddCommitType(t *testing.T) {
	newTestDB(t)

	t.Run("新增成功", func(t *testing.T) {
		if err := AddCommitType("ci"); err != nil {
			t.Fatalf("AddCommitType(ci) failed: %v", err)
		}
		opts, err := GetOptions()
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}
		found := false
		for _, opt := range opts {
			if opt.Value == "ci" {
				found = true
				if opt.Label != "ci" {
					t.Errorf("label = %q, want %q", opt.Label, "ci")
				}
				if opt.Usage != 0 {
					t.Errorf("usage = %d, want 0", opt.Usage)
				}
			}
		}
		if !found {
			t.Error("commit type 'ci' not found after add")
		}
	})

	t.Run("重复添加报错", func(t *testing.T) {
		err := AddCommitType("fix") // 默认选项已存在
		if err == nil {
			t.Fatal("expected error for duplicate commit type, got nil")
		}
		if !errors.Is(err, ErrCommitTypeExists) {
			t.Errorf("error = %v, want errors.Is(ErrCommitTypeExists)", err)
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("error = %q, want contains 'already exists'", err)
		}
	})

	t.Run("非法格式报错", func(t *testing.T) {
		for _, bad := range []string{"", "BadType", "bad type", "bad/type", "bad_type", "中文", "aB"} {
			if err := AddCommitType(bad); err == nil {
				t.Errorf("AddCommitType(%q): expected error, got nil", bad)
			}
		}
	})
}

func TestDeleteCommitTypes(t *testing.T) {
	newTestDB(t)

	t.Run("删除部分成功", func(t *testing.T) {
		if err := DeleteCommitTypes([]string{"style", "docs"}); err != nil {
			t.Fatalf("DeleteCommitTypes failed: %v", err)
		}
		opts, err := GetOptions()
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}
		for _, opt := range opts {
			if opt.Value == "style" || opt.Value == "docs" {
				t.Errorf("commit type %q still exists after delete", opt.Value)
			}
		}
	})

	t.Run("删除全部报错", func(t *testing.T) {
		// 先删到只剩 1 个,再尝试删最后一个
		opts, err := GetOptions()
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}
		var values []string
		for _, opt := range opts {
			values = append(values, opt.Value)
		}
		if len(values) < 2 {
			t.Fatal("expected at least 2 commit types")
		}
		if err := DeleteCommitTypes(values[:len(values)-1]); err != nil {
			t.Fatalf("DeleteCommitTypes partial failed: %v", err)
		}
		last := values[len(values)-1]
		if err := DeleteCommitTypes([]string{last}); err == nil {
			t.Fatal("expected error when deleting the last commit type, got nil")
		}
		// 数据未被破坏:最后一个仍存在
		opts, err = GetOptions()
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}
		if len(opts) != 1 || opts[0].Value != last {
			t.Errorf("after failed delete: got %d options, want 1 (%q)", len(opts), last)
		}
	})

	t.Run("空列表无操作", func(t *testing.T) {
		if err := DeleteCommitTypes(nil); err != nil {
			t.Errorf("DeleteCommitTypes(nil) = %v, want nil", err)
		}
	})

	t.Run("不存在的值静默跳过", func(t *testing.T) {
		newTestDB(t)
		if err := DeleteCommitTypes([]string{"nonexistent"}); err != nil {
			t.Errorf("DeleteCommitTypes(nonexistent) = %v, want nil", err)
		}
	})
}

func TestFormatPatch(t *testing.T) {
	tests := []struct {
		name  string
		patch Patch
		want  string
	}{
		{name: "默认值", patch: Patch{Prefix: "", Major: 999, Minor: 9, Patch: 9, Suffix: ""}, want: "999.9.9"},
		{name: "带前缀后缀", patch: Patch{Prefix: "v", Major: 1, Minor: 2, Patch: 3, Suffix: "-beta"}, want: "v1.2.3-beta"},
		{name: "全零", patch: Patch{Prefix: "", Major: 0, Minor: 0, Patch: 0, Suffix: ""}, want: "0.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPatch(tt.patch); got != tt.want {
				t.Errorf("FormatPatch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildConfigOptions(t *testing.T) {
	newTestDB(t)

	opts := BuildConfigOptions()
	if len(opts) != 4 {
		t.Fatalf("expected 4 config options, got %d", len(opts))
	}

	// 分派键唯一且正确
	expectedKeys := []string{ConfigKeyLanguage, ConfigKeyPush, ConfigKeyCommitTypes, ConfigKeyTagPatch}
	seen := make(map[string]bool)
	for i, opt := range opts {
		if opt.Value != expectedKeys[i] {
			t.Errorf("option %d: value = %q, want %q", i, opt.Value, expectedKeys[i])
		}
		if seen[opt.Value] {
			t.Errorf("duplicate config key %q", opt.Value)
		}
		seen[opt.Value] = true
		if opt.Label == "" {
			t.Errorf("option %d: empty label", i)
		}
		if opt.Description == "" {
			t.Errorf("option %d (%s): empty summary", i, opt.Value)
		}
	}

	// 提交类型摘要包含默认类型
	if !strings.Contains(opts[2].Description, "fix") {
		t.Errorf("commit types summary = %q, want contains 'fix'", opts[2].Description)
	}

	// 标签版本上限摘要为默认 999.9.9
	if !strings.Contains(opts[3].Description, "999.9.9") {
		t.Errorf("tag patch summary = %q, want contains '999.9.9'", opts[3].Description)
	}
}
