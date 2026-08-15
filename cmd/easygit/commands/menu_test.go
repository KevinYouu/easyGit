package commands

import (
	"testing"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// TestMenuItems 主菜单项完整性:分派键唯一、run 非空、描述已本地化。
func TestMenuItems(t *testing.T) {
	if len(menuItems) == 0 {
		t.Fatal("menuItems should not be empty")
	}

	seen := make(map[string]bool)
	for _, item := range menuItems {
		if item.key == "" || item.label == "" {
			t.Errorf("menu item missing key/label: %+v", item)
		}
		if seen[item.key] {
			t.Errorf("duplicate menu key %q", item.key)
		}
		seen[item.key] = true
		if item.run == nil {
			t.Errorf("menu item %q has nil run", item.key)
		}
		// 描述走 i18n(非硬编码)
		if item.desc == "" {
			t.Errorf("menu item %q has empty description", item.key)
		}
	}

	// 高频命令必须在列表前部(心智负担:入口即所见)
	frontKeys := []string{"push-all", "push-selected", "merge"}
	for i, front := range frontKeys {
		if menuItems[i].key != front {
			t.Errorf("menuItems[%d] = %q, want %q", i, menuItems[i].key, front)
		}
	}
}

// TestMenuTitleLocalized 菜单标题本地化
func TestMenuTitleLocalized(t *testing.T) {
	if i18n.T("menu.title") == "" {
		t.Error("menu.title should be localized")
	}
}
