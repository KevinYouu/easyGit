package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

// undoReflogLimit 回滚检查点列表长度上限
const undoReflogLimit = 50

// ReflogEntry 一条 reflog 记录(回滚检查点)
type ReflogEntry struct {
	Hash   string // 短哈希(reset 目标)
	Action string // 动作描述(commit:/checkout:/reset: 等)
	Date   string // 发生时间
	Sel    string // 引用标记,如 HEAD@{0}
}

// undoSelectEntry 检查点选择菜单(测试可注入)
var undoSelectEntry = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("undo.select.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// undoModeSelect 重置模式选择(测试可注入,预选上次模式)
var undoModeSelect = func(options []config.Option, preselected string) (string, error) {
	selected, err := form.ListForm(i18n.T("undo.mode.title"), options, form.ListSingle, preselected)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// undoConfirm 恢复确认(测试可注入)
var undoConfirm = form.Confirm

// listReflog 获取最近 limit 条 reflog 记录(最新在前)。
// 格式 %h|%cd|%gs:可变长的 subject 放末段,含 "|" 时字段不错位。
func listReflog(limit int) ([]ReflogEntry, error) {
	cmd := exec.Command("git", "reflog", "-n", fmt.Sprintf("%d", limit),
		"--pretty=format:%h|%cd|%gs", "--date=format:%m-%d %H:%M")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git reflog: %w", err)
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, fmt.Errorf("%s", i18n.T("undo.empty"))
	}

	var entries []ReflogEntry
	for i, line := range strings.Split(trimmed, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) < 3 {
			continue
		}
		entries = append(entries, ReflogEntry{
			Hash:   parts[0],
			Date:   parts[1],
			Action: parts[2],
			Sel:    fmt.Sprintf("HEAD@{%d}", i),
		})
	}
	return entries, nil
}

// executeUndo 执行回滚(核心执行,交互层与测试共用):
// 将当前分支重置到目标检查点,mode 为空串时等同 mixed;
// 成功后记忆本次模式(与 reset 共用同一记忆键)。
func executeUndo(entry ReflogEntry, mode string) error {
	args := []string{"reset"}
	if mode != "" {
		args = append(args, mode)
	}
	args = append(args, entry.Hash)

	_, err := command.RunCmdWithSpinnerOptions("git", args,
		fmt.Sprintf(i18n.T("undo.executing"), entry.Sel),
		fmt.Sprintf(i18n.T("undo.success"), entry.Hash, entry.Action),
		true)
	if err != nil {
		return err
	}
	config.SaveLastChoice(config.LastChoiceResetMode, mode)
	return nil
}

// Undo 交互式回滚(reflog 后悔药):
// 列出最近检查点 → 选重置模式(与 rs 共用记忆)→ 确认后恢复。
func Undo() error {
	entries, err := listReflog(undoReflogLimit)
	if err != nil {
		return err
	}

	options := make([]config.Option, 0, len(entries))
	for _, e := range entries {
		options = append(options, config.Option{
			Label: fmt.Sprintf("%s %s\n%s • %s", e.Hash, e.Action, e.Date, e.Sel),
			Value: e.Sel,
		})
	}

	sel, err := undoSelectEntry(options)
	if err != nil {
		return nil // Esc 取消
	}

	var target ReflogEntry
	for _, e := range entries {
		if e.Sel == sel {
			target = e
			break
		}
	}

	// 重置模式选择:与 reset 共用选项与记忆(默认项在首位,直接 Enter)
	lastMode, _ := config.GetLastChoice(config.LastChoiceResetMode)
	mode, err := undoModeSelect(resetModeOptions(), lastMode)
	if err != nil {
		return nil // Esc 取消
	}

	modeReadable := strings.TrimPrefix(mode, "--")
	if mode == "" {
		modeReadable = i18n.T("reset.option.default.name")
	}

	confirmMsg := fmt.Sprintf("%s [%s] %s (%s)",
		i18n.T("undo.confirm.to"),
		target.Hash,
		target.Action,
		modeReadable) + getModeDescription(mode)
	if mode == "--hard" {
		confirmMsg += "\n" + i18n.T("undo.hard.warning")
	}
	if !undoConfirm(confirmMsg) {
		return nil
	}

	return executeUndo(target, mode)
}
