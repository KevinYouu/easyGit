package gitcmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// StashEntry 一条 stash 记录:列表展示与内部引用
type StashEntry struct {
	Index   string // git 引用,如 stash@{0}
	Message string // stash 消息(默认 "WIP on <branch>: <hash> <subject>")
	Date    string // 创建时间
}

// stashActionMenu stash 条目操作菜单(测试可注入)
var stashActionMenu = func(options []config.Option) (string, error) {
	selected, err := form.ListForm(i18n.T("stash.action.title"), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// stashMainMenu 主菜单选择(测试可注入)
var stashMainMenu = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("stash.menu.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// stashSelectEntry 列表条目选择(测试可注入)
var stashSelectEntry = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("stash.select.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// stashMessageInput 可选消息输入(测试可注入)
var stashMessageInput = func() (string, error) {
	return form.InputOptional(i18n.T("stash.message.input"), "")
}

// stashConfirm 破坏性操作确认(测试可注入)
var stashConfirm = form.Confirm

// listStashes 获取全部 stash 条目(最新在最前,与 git stash list 一致)
func listStashes() ([]StashEntry, error) {
	cmd := exec.Command("git", "stash", "list", "--pretty=format:%gd|%cr|%gs")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git stash list: %w", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}

	var entries []StashEntry
	for line := range strings.SplitSeq(trimmed, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) < 3 {
			continue
		}
		entries = append(entries, StashEntry{Index: parts[0], Date: parts[1], Message: parts[2]})
	}
	return entries, nil
}

// saveStash 保存当前修改到 stash(message 可空,走 git 默认消息);
// 无修改时提示并跳过。
func saveStash(message string) error {
	dirty, err := isWorkingDirectoryDirty()
	if err != nil {
		return err
	}
	if !dirty {
		logs.Info(i18n.T("stash.no.changes"))
		return nil
	}

	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err = command.RunCmdWithSpinnerOptions("git", args,
		i18n.T("stash.push.loading"),
		i18n.T("stash.push.success"),
		true)
	return err
}

// applyStash 恢复指定条目(pop = 恢复后删除条目;apply = 保留条目)。
// 冲突时 pop 的条目由 git 自动保留,提示手动处理。
func applyStash(entry StashEntry, drop bool) error {
	action := "apply"
	loadingKey := "stash.apply.loading"
	successKey := "stash.apply.success"
	if drop {
		action = "pop"
		loadingKey = "stash.pop.loading"
		successKey = "stash.pop.success"
	}

	output, err := command.RunCmdWithSpinnerOptions("git",
		[]string{"stash", action, entry.Index},
		fmt.Sprintf(i18n.T(loadingKey), entry.Index),
		fmt.Sprintf(i18n.T(successKey), entry.Index),
		false)
	if err != nil {
		if strings.Contains(output, "CONFLICT") {
			logs.Error(i18n.T("stash.conflict.hint"))
		}
		return fmt.Errorf("%s: %s", i18n.T("stash.apply.failed"), strings.TrimSpace(output))
	}
	return nil
}

// dropStash 删除单条 stash
func dropStash(entry StashEntry) error {
	_, err := command.RunCmdWithSpinnerOptions("git",
		[]string{"stash", "drop", entry.Index},
		fmt.Sprintf(i18n.T("stash.drop.loading"), entry.Index),
		fmt.Sprintf(i18n.T("stash.drop.success"), entry.Index),
		true)
	return err
}

// clearStashes 清空全部 stash(调用方须先强确认)
func clearStashes() error {
	_, err := command.RunCmdWithSpinnerOptions("git",
		[]string{"stash", "clear"},
		i18n.T("stash.clear.loading"),
		i18n.T("stash.clear.success"),
		false)
	return err
}

// showStashDiff 输出指定条目的改动内容(stash show -p)
func showStashDiff(entry StashEntry) error {
	cmd := exec.Command("git", "stash", "show", "-p", entry.Index)
	output, err := cmd.CombinedOutput()
	titleStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)
	fmt.Printf("\n%s %s\n", titleStyle.Render(entry.Index),
		lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(i18n.T("stash.diff.title")))
	fmt.Println(string(output))
	return err
}

// manageStashes stash 列表管理子流程:选中条目 → 查看 diff / 应用 / 应用并删除 / 删除;
// diff 与删除后回到列表继续(Esc 退出),应用/应用并删除完成即返回。
func manageStashes() error {
	for {
		entries, err := listStashes()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			logs.Info(i18n.T("stash.empty"))
			return nil
		}

		options := make([]config.Option, 0, len(entries))
		for _, e := range entries {
			options = append(options, config.Option{
				Label:       e.Message,
				Description: fmt.Sprintf("%s · %s", e.Index, e.Date),
				Value:       e.Index,
			})
		}

		selected, err := stashSelectEntry(options)
		if err != nil {
			return nil // Esc 返回上级
		}

		var target StashEntry
		for _, e := range entries {
			if e.Index == selected {
				target = e
				break
			}
		}

		actions := []config.Option{
			{Label: i18n.T("stash.action.diff"), Description: i18n.T("stash.action.diff.desc"), Value: "diff"},
			{Label: i18n.T("stash.action.apply"), Description: i18n.T("stash.action.apply.desc"), Value: "apply"},
			{Label: i18n.T("stash.action.pop"), Description: i18n.T("stash.action.pop.desc"), Value: "pop"},
			{Label: i18n.T("stash.action.drop"), Description: i18n.T("stash.action.drop.desc"), Value: "drop"},
		}
		action, err := stashActionMenu(actions)
		if err != nil {
			continue // Esc 回到列表
		}

		switch action {
		case "diff":
			showStashDiff(target)
		case "apply":
			return applyStash(target, false)
		case "pop":
			return applyStash(target, true)
		case "drop":
			if !stashConfirm(fmt.Sprintf(i18n.T("stash.drop.confirm"), target.Index)) {
				logs.Info(i18n.T("stash.cancelled"))
				continue
			}
			if err := dropStash(target); err != nil {
				return err
			}
		}
	}
}

// Stash 交互式 stash 管理:
// 主菜单三选:保存(带消息)/ 列表管理(diff 预览 + 应用/删除)/ 清空全部(强确认)。
func Stash() error {
	menuOptions := []config.Option{
		{Label: i18n.T("stash.menu.save"), Description: i18n.T("stash.menu.save.desc"), Value: "save"},
		{Label: i18n.T("stash.menu.manage"), Description: i18n.T("stash.menu.manage.desc"), Value: "manage"},
		{Label: i18n.T("stash.menu.clear"), Description: i18n.T("stash.menu.clear.desc"), Value: "clear"},
	}

	action, err := stashMainMenu(menuOptions)
	if err != nil {
		return nil // Esc 取消
	}

	switch action {
	case "save":
		message, err := stashMessageInput()
		if err != nil {
			return nil // Esc 取消
		}
		return saveStash(message)
	case "manage":
		return manageStashes()
	case "clear":
		count := stashEntryCount()
		if count == 0 {
			logs.Info(i18n.T("stash.empty"))
			return nil
		}
		if !stashConfirm(fmt.Sprintf(i18n.T("stash.clear.confirm"), count)) {
			logs.Info(i18n.T("stash.cancelled"))
			return nil
		}
		return clearStashes()
	}
	return nil
}

// stashEntryCount 当前 stash 条目数(清空确认文案用)
func stashEntryCount() int {
	entries, err := listStashes()
	if err != nil {
		return 0
	}
	return len(entries)
}

// parseStashRefNumber 从 stash@{N} 提取 N(-1 表示格式非法)
func parseStashRefNumber(ref string) int {
	start := strings.Index(ref, "{")
	end := strings.Index(ref, "}")
	if start < 0 || end <= start {
		return -1
	}
	n, err := strconv.Atoi(ref[start+1 : end])
	if err != nil {
		return -1
	}
	return n
}
