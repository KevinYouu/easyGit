package gitcmd

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// cleanItem 一条待清理项:路径 + 类别(已跟踪修改 / 未跟踪文件)
type cleanItem struct {
	Path        string
	IsUntracked bool
}

// cleanSelect 清理项多选菜单(测试可注入)
var cleanSelect = func(options []config.Option) ([]string, error) {
	return form.ListFormColumns(i18n.T("clean.select.title"), form.NameDescColumns(), options, form.ListMulti)
}

// cleanConfirm 清理确认(测试可注入)
var cleanConfirm = form.Confirm

// listModifiedTrackedFiles 已跟踪文件的修改列表(含暂存与未暂存,不含未跟踪):
// `git diff --name-only HEAD` 同时覆盖两者。
func listModifiedTrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// 空仓库无 HEAD:视为无已跟踪修改
		return nil, nil
	}
	return splitNonEmptyLines(string(output)), nil
}

// listUntrackedFiles 未跟踪文件列表(gitignore 忽略项不列出、不删除)
func listUntrackedFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	return splitNonEmptyLines(string(output)), nil
}

// splitNonEmptyLines 按行拆分并去空
func splitNonEmptyLines(output string) []string {
	var lines []string
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// discardTrackedChanges 还原已跟踪文件到 HEAD(工作区+暂存区一并恢复)
func discardTrackedChanges(files []string) error {
	args := append([]string{"checkout", "HEAD", "--"}, files...)
	_, err := command.RunCmdWithSpinnerOptions("git", args,
		i18n.T("clean.discard.loading"),
		fmt.Sprintf(i18n.T("clean.discard.success"), len(files)),
		false)
	return err
}

// deleteUntrackedFiles 删除未跟踪文件(clean -f,不含忽略项与目录递归之外的强制)
func deleteUntrackedFiles(files []string) error {
	args := append([]string{"clean", "-f", "--"}, files...)
	_, err := command.RunCmdWithSpinnerOptions("git", args,
		i18n.T("clean.delete.loading"),
		fmt.Sprintf(i18n.T("clean.delete.success"), len(files)),
		true)
	return err
}

// Clean 交互式工作区清理:
// 单一多选列表混合呈现已跟踪修改与未跟踪文件(标注类别),
// 统一强确认后分组执行还原/删除。
func Clean() error {
	modified, err := listModifiedTrackedFiles()
	if err != nil {
		return err
	}
	untracked, err := listUntrackedFiles()
	if err != nil {
		return err
	}

	if len(modified) == 0 && len(untracked) == 0 {
		logs.Info(i18n.T("clean.empty"))
		return nil
	}

	items := make([]cleanItem, 0, len(modified)+len(untracked))
	for _, p := range modified {
		items = append(items, cleanItem{Path: p})
	}
	for _, p := range untracked {
		items = append(items, cleanItem{Path: p, IsUntracked: true})
	}

	options := make([]config.Option, 0, len(items))
	for i, item := range items {
		tag := i18n.T("clean.tag.modified")
		if item.IsUntracked {
			tag = i18n.T("clean.tag.untracked")
		}
		options = append(options, config.Option{
			Label: fmt.Sprintf("[%s] %s", tag, item.Path),
			Value: strconv.Itoa(i),
		})
	}

	selectedIdx, err := cleanSelect(options)
	if err != nil || len(selectedIdx) == 0 {
		logs.Info(i18n.T("clean.cancelled"))
		return nil
	}

	var toDiscard, toDelete []string
	for _, idxStr := range selectedIdx {
		idx, convErr := strconv.Atoi(idxStr)
		if convErr != nil || idx < 0 || idx >= len(items) {
			continue
		}
		if items[idx].IsUntracked {
			toDelete = append(toDelete, items[idx].Path)
		} else {
			toDiscard = append(toDiscard, items[idx].Path)
		}
	}
	if len(toDiscard)+len(toDelete) == 0 {
		logs.Info(i18n.T("clean.cancelled"))
		return nil
	}

	// 强确认:逐类计数 + 完整清单
	confirmMsg := ""
	if len(toDiscard) > 0 {
		confirmMsg += fmt.Sprintf(i18n.T("clean.confirm.modified"), len(toDiscard)) + "\n"
	}
	if len(toDelete) > 0 {
		confirmMsg += fmt.Sprintf(i18n.T("clean.confirm.untracked"), len(toDelete)) + "\n"
	}
	confirmMsg += "\n" + strings.Join(append(append([]string{}, toDiscard...), toDelete...), "\n")

	if !cleanConfirm(confirmMsg) {
		logs.Info(i18n.T("clean.cancelled"))
		return nil
	}

	cleaned := 0
	if len(toDiscard) > 0 {
		if err := discardTrackedChanges(toDiscard); err != nil {
			return err
		}
		cleaned += len(toDiscard)
	}
	if len(toDelete) > 0 {
		if err := deleteUntrackedFiles(toDelete); err != nil {
			return err
		}
		cleaned += len(toDelete)
	}
	logs.Success(fmt.Sprintf(i18n.T("clean.success"), cleaned))
	return nil
}
