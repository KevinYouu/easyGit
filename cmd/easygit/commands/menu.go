package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/update"
	"github.com/KevinYouu/easyGit/internal/version"
)

// menuItem 主菜单项:命令名 + 别名 + i18n 键 + 执行函数。
// 注意:desc 不能包级预取——包初始化早于 main() 的语言加载
// (数据库设置),此时 i18n.T 会按系统 locale 固化语言;
// 必须在 RunMenu 时动态取当前语言的描述。
type menuItem struct {
	key     string
	label   string
	descKey string
	run     func() error
}

// menuItems 主菜单项:顺序按日常使用频率排列,高频在前;value 为分派键。
// run 直接调用底层实现(与 cobra 命令 Run 同一函数体),避免经 cobra 二次分发。
var menuItems = []menuItem{
	{key: "push-all", label: "push-all (pa)", descKey: "push.all.short", run: func() error { return gitcmd.PushAll() }},
	{key: "push-selected", label: "push-selected (ps)", descKey: "push.selected.short", run: func() error { return gitcmd.PushSelected() }},
	{key: "merge", label: "merge (m)", descKey: "merge.short", run: func() error { return gitcmd.MergeIntoCurrent() }},
	{key: "rebase", label: "rebase (r)", descKey: "rebase.short", run: func() error { return gitcmd.RebaseIntoCurrent() }},
	{key: "cherry-pick", label: "cherry-pick (cp)", descKey: "cherry.pick.short", run: func() error { return gitcmd.CherryPick() }},
	{key: "reset", label: "reset (rs)", descKey: "reset.short", run: func() error { return gitcmd.Reset() }},
	{key: "squash", label: "squash (sq)", descKey: "squash.short", run: func() error { return gitcmd.Squash() }},
	{key: "drop", label: "drop (d)", descKey: "drop.short", run: func() error { return gitcmd.Drop() }},
	{key: "tag-create", label: "tag-create (tc)", descKey: "tag.short", run: func() error { return gitcmd.CreateAndPushTag() }},
	{key: "tag-delete", label: "tag-delete (td)", descKey: "tag.delete.short", run: func() error { return gitcmd.DeleteAndPushTag() }},
	{key: "branch-delete", label: "branch-delete (bd)", descKey: "branch.delete.short", run: func() error { return gitcmd.DeleteBranch() }},
	{key: "config", label: "config", descKey: "config.short", run: func() error { runConfigCenter(); return nil }},
	{key: "update", label: "update", descKey: "update.short", run: func() error { return update.UpdateSelf() }},
	{key: "version", label: "version (v)", descKey: "version.short", run: func() error { version.GetVersion(); return nil }},
}

// RunMenu 交互式主菜单:无参数运行 easyGit 时进入。
// 单一入口降低命令记忆负担;选择后执行对应命令并退出,Esc 退出菜单。
// 命令本身仍可独立调用(脚本/高级用户)。
func RunMenu() error {
	options := make([]config.Option, 0, len(menuItems))
	byKey := make(map[string]func() error, len(menuItems))
	for _, item := range menuItems {
		// 描述运行时取当前语言(语言切换后即时生效)
		options = append(options, config.Option{Label: item.label, Description: i18n.T(item.descKey), Value: item.key})
		byKey[item.key] = item.run
	}

	selected, err := form.ListFormColumns(i18n.T("menu.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return nil // Esc / 取消:退出
	}

	if run, ok := byKey[selected[0]]; ok {
		if err := run(); err != nil {
			logs.Error(err.Error())
		}
	}
	return nil
}
