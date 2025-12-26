package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultOptions(t *testing.T) {
	opts := GetDefaultOptions()

	expectedLabels := []string{"fix", "feat", "refactor", "build", "chore", "style", "docs", "revert", "test"}

	if len(opts) != len(expectedLabels) {
		t.Errorf("expected %d options, got %d", len(expectedLabels), len(opts))
	}

	for i, opt := range opts {
		if opt.Label != expectedLabels[i] {
			t.Errorf("option %d: expected label %q, got %q", i, expectedLabels[i], opt.Label)
		}
		if opt.Value != expectedLabels[i] {
			t.Errorf("option %d: expected value %q, got %q", i, expectedLabels[i], opt.Value)
		}
		if opt.Usage != 0 {
			t.Errorf("option %d: expected usage 0, got %d", i, opt.Usage)
		}
	}
}

func TestIncrementUsage(t *testing.T) {
	// 使用临时数据库
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	// 初始化数据库
	if err := Initialize(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// 测试增量更新
	testCases := []struct {
		value         string
		expectedUsage int
	}{
		{"fix", 1},
		{"fix", 2},
		{"fix", 3},
		{"feat", 1},
		{"fix", 4},
	}

	for _, tc := range testCases {
		if err := IncrementUsage(tc.value); err != nil {
			t.Fatalf("IncrementUsage(%q) failed: %v", tc.value, err)
		}

		// 验证使用次数
		opts, err := GetOptions()
		if err != nil {
			t.Fatalf("GetOptions failed: %v", err)
		}

		found := false
		for _, opt := range opts {
			if opt.Value == tc.value {
				found = true
				if opt.Usage != tc.expectedUsage {
					t.Errorf("after IncrementUsage(%q): expected usage %d, got %d",
						tc.value, tc.expectedUsage, opt.Usage)
				}
				break
			}
		}
		if !found {
			t.Errorf("option %q not found in results", tc.value)
		}
	}
}

func TestGetOptionsOrdering(t *testing.T) {
	// 使用临时数据库
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	// 初始化数据库
	if err := Initialize(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// 增加不同的使用次数
	IncrementUsage("feat")
	IncrementUsage("feat")
	IncrementUsage("feat")
	IncrementUsage("fix")
	IncrementUsage("fix")
	IncrementUsage("docs")

	// 获取选项
	opts, err := GetOptions()
	if err != nil {
		t.Fatalf("GetOptions failed: %v", err)
	}

	// 验证排序 (应该按 usage DESC)
	expectedOrder := []struct {
		value string
		usage int
	}{
		{"feat", 3},
		{"fix", 2},
		{"docs", 1},
	}

	for i, expected := range expectedOrder {
		if opts[i].Value != expected.value {
			t.Errorf("position %d: expected value %q, got %q", i, expected.value, opts[i].Value)
		}
		if opts[i].Usage != expected.usage {
			t.Errorf("position %d: expected usage %d, got %d", i, expected.usage, opts[i].Usage)
		}
	}
}

func TestSaveOptions(t *testing.T) {
	// 使用临时数据库
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	// 初始化数据库
	if err := Initialize(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// 创建自定义选项
	customOptions := []Option{
		{Label: "custom1", Value: "custom1", Usage: 5},
		{Label: "custom2", Value: "custom2", Usage: 10},
	}

	// 保存选项
	if err := SaveOptions(customOptions); err != nil {
		t.Fatalf("SaveOptions failed: %v", err)
	}

	// 读取并验证
	opts, err := GetOptions()
	if err != nil {
		t.Fatalf("GetOptions failed: %v", err)
	}

	// 应该包含自定义选项
	found := 0
	for _, opt := range opts {
		if opt.Value == "custom1" || opt.Value == "custom2" {
			found++
		}
	}

	if found != 2 {
		t.Errorf("expected to find 2 custom options, found %d", found)
	}
}

func TestInitialize(t *testing.T) {
	// 使用临时数据库
	tmpDir := t.TempDir()
	testDBPath = filepath.Join(tmpDir, "test.db")
	defer func() { testDBPath = "" }()

	// 初始化应该成功
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 验证数据库文件存在
	if _, err := os.Stat(testDBPath); os.IsNotExist(err) {
		t.Errorf("database file not created at %s", testDBPath)
	}

	// 验证默认选项已插入
	opts, err := GetOptions()
	if err != nil {
		t.Fatalf("GetOptions failed after Initialize: %v", err)
	}

	if len(opts) < 9 {
		t.Errorf("expected at least 9 default options, got %d", len(opts))
	}

	// 多次调用 Initialize 应该是幂等的
	if err := Initialize(); err != nil {
		t.Errorf("second Initialize call failed: %v", err)
	}
}
