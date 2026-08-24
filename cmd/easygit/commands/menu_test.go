package commands

import (
	"testing"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// TestMenuItems 主菜单项完整性:分派键唯一、run 非空、descKey 已定义。
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
		// 描述走 i18n 键(运行时动态取,非包级固化)
		if item.descKey == "" {
			t.Errorf("menu item %q has empty descKey", item.key)
		}
		if !i18n.Has(item.descKey) {
			t.Errorf("menu item %q descKey %q not found in i18n", item.key, item.descKey)
		}
	}

	// 高频命令必须在列表前部(心智负担:入口即所见)
	frontKeys := []string{"push-all", "push-selected", "branch-switch", "stash", "branch-create", "merge"}
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

// TestMenuDescriptionFollowsLanguage 菜单简介随语言切换即时生效
// (回归:包级 var 初始化早于语言加载,i18n.T 会按系统 locale 固化,
// 必须运行时动态取描述)。
func TestMenuDescriptionFollowsLanguage(t *testing.T) {
	orig := i18n.GetCurrentLanguage()
	defer i18n.SetLanguage(orig)

	i18n.SetLanguage(i18n.LangEN)
	enDesc := i18n.T(menuItems[0].descKey)

	i18n.SetLanguage(i18n.LangZH)
	zhDesc := i18n.T(menuItems[0].descKey)

	if enDesc == zhDesc {
		t.Fatalf("desc should differ per language: %q", enDesc)
	}
	if enDesc != "Push all changes to remote repository" {
		t.Errorf("en desc = %q", enDesc)
	}
	if zhDesc != "推送所有更改到远程仓库" {
		t.Errorf("zh desc = %q", zhDesc)
	}
}
