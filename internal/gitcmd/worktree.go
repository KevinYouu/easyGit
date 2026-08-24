package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// WorktreeInfo 一条 worktree 记录(porcelain 解析)
type WorktreeInfo struct {
	Path   string
	Head   string // 短哈希(空仓库为空)
	Branch string // 本地分支名(refs/heads/main → main)
	IsMain bool   // 是否主工作树(porcelain 首条)
	IsBare bool
}

// worktreeMenu 主菜单选择(测试可注入)
var worktreeMenu = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("worktree.menu.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// worktreePathInput 路径输入(测试可注入)
var worktreePathInput = func() (string, error) {
	return form.InputWithValidate(i18n.T("worktree.path.input"), "", func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s", i18n.T("form.input.empty.error"))
		}
		return nil
	})
}

// worktreeModeSelect 创建模式选择(测试可注入)
var worktreeModeSelect = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("worktree.mode.select"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// worktreeBranchInput 新分支名输入(测试可注入)
var worktreeBranchInput = func(validate func(string) error) (string, error) {
	return form.InputWithValidate(i18n.T("worktree.branch.input"), "", validate)
}

// worktreeBranchSelect 分支选择(测试可注入)
var worktreeBranchSelect = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("worktree.select.branch"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// worktreeSelectRemove 删除多选(测试可注入)
var worktreeSelectRemove = func(options []config.Option) ([]string, error) {
	return form.ListFormColumns(i18n.T("worktree.select.remove"), form.NameDescColumns(), options, form.ListMulti)
}

// worktreeConfirm 确认框(测试可注入)
var worktreeConfirm = form.Confirm

// listWorktrees 获取全部工作树(porcelain 格式解析,首条为主工作树)
func listWorktrees() ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	var infos []WorktreeInfo
	var current *WorktreeInfo
	flush := func() {
		if current != nil {
			infos = append(infos, *current)
			current = nil
		}
	}

	for line := range strings.SplitSeq(string(output), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current = &WorktreeInfo{Path: strings.TrimPrefix(line, "worktree "), IsMain: len(infos) == 0}
		case current == nil:
			continue
		case strings.HasPrefix(line, "HEAD "):
			head := strings.TrimPrefix(line, "HEAD ")
			if len(head) >= 7 {
				current.Head = head[:7]
			}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case strings.TrimSpace(line) == "bare":
			current.IsBare = true
		}
	}
	flush()

	return infos, nil
}

// worktreesInUse 已被任意工作树占用的分支名集合
func worktreesInUse() map[string]bool {
	inUse := make(map[string]bool)
	infos, _ := listWorktrees()
	for _, w := range infos {
		if w.Branch != "" {
			inUse[w.Branch] = true
		}
	}
	return inUse
}

// worktreeAdd 新增工作树:createNew 为 true 时以新分支创建(-b),否则检出已有分支
func worktreeAdd(path, branch string, createNew bool) error {
	args := []string{"worktree", "add"}
	if createNew {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if !createNew && branch != "" {
		args = append(args, branch)
	}

	_, err := command.RunCmdWithSpinnerOptions("git", args,
		fmt.Sprintf(i18n.T("worktree.add.loading"), path),
		fmt.Sprintf(i18n.T("worktree.add.success"), path, branch),
		true)
	return err
}

// worktreeRemovePaths 删除指定工作树(顺序执行,共享仓库元数据不宜并行)
func worktreeRemovePaths(paths []string) error {
	for _, p := range paths {
		if _, err := command.RunCmdWithSpinnerOptions("git",
			[]string{"worktree", "remove", p},
			fmt.Sprintf(i18n.T("worktree.remove.loading"), p),
			fmt.Sprintf(i18n.T("worktree.remove.success"), p),
			false); err != nil {
			return err
		}
	}
	return nil
}

// printWorktrees 输出工作树清单(纯展示)
func printWorktrees(infos []WorktreeInfo) {
	logs.Info(i18n.T("worktree.list.title"))
	for _, w := range infos {
		label := fmt.Sprintf("%s (%s)", w.Path, w.Branch)
		switch {
		case w.IsBare:
			label = fmt.Sprintf("%s %s", w.Path, i18n.T("worktree.tag.bare"))
		case w.IsMain:
			label += " " + i18n.T("worktree.tag.main")
		}
		fmt.Println("  " + label)
	}
}

// addWorktreeFlow 添加工作转子流程:路径 → 模式(新分支/已有分支)→ 执行
func addWorktreeFlow() error {
	path, err := worktreePathInput()
	if err != nil {
		return nil // Esc 取消
	}

	modeOptions := []config.Option{
		{Label: i18n.T("worktree.mode.new"), Description: i18n.T("worktree.mode.new.desc"), Value: "new"},
		{Label: i18n.T("worktree.mode.existing"), Description: i18n.T("worktree.mode.existing.desc"), Value: "existing"},
	}
	mode, err := worktreeModeSelect(modeOptions)
	if err != nil {
		return nil
	}

	if mode == "new" {
		branch, err := worktreeBranchInput(validateBranchName)
		if err != nil || strings.TrimSpace(branch) == "" {
			return nil
		}
		return worktreeAdd(path, branch, true)
	}

	inUse := worktreesInUse()
	var available []config.Option
	if locals, err := GetAllBranches(); err == nil {
		for _, b := range locals {
			if !inUse[b] && !strings.HasPrefix(b, "(") {
				available = append(available, config.Option{Label: b, Value: b})
			}
		}
	}
	if len(available) == 0 {
		logs.Info(i18n.T("worktree.no.available.branches"))
		return nil
	}

	branch, err := worktreeBranchSelect(available)
	if err != nil {
		return nil
	}
	return worktreeAdd(path, branch, false)
}

// removeWorktreeFlow 删除工作转子流程:多选(排除主与裸仓)→ 确认 → 执行
func removeWorktreeFlow() error {
	infos, err := listWorktrees()
	if err != nil {
		return err
	}

	var options []config.Option
	var targets []WorktreeInfo
	for _, w := range infos {
		if w.IsMain || w.IsBare {
			continue
		}
		options = append(options, config.Option{
			Label: fmt.Sprintf("%s (%s)", w.Path, w.Branch),
			Value: w.Path,
		})
		targets = append(targets, w)
	}
	if len(options) == 0 {
		logs.Info(i18n.T("worktree.none.to.remove"))
		return nil
	}

	selected, err := worktreeSelectRemove(options)
	if err != nil || len(selected) == 0 {
		logs.Info(i18n.T("worktree.cancelled"))
		return nil
	}

	if !worktreeConfirm(fmt.Sprintf(i18n.T("worktree.remove.confirm"), len(selected))) {
		logs.Info(i18n.T("worktree.cancelled"))
		return nil
	}
	return worktreeRemovePaths(selected)
}

// Worktree 交互式工作树管理:查看 / 添加 / 删除,循环直到 Esc。
func Worktree() error {
	menuOptions := []config.Option{
		{Label: i18n.T("worktree.menu.list"), Description: i18n.T("worktree.menu.list.desc"), Value: "list"},
		{Label: i18n.T("worktree.menu.add"), Description: i18n.T("worktree.menu.add.desc"), Value: "add"},
		{Label: i18n.T("worktree.menu.remove"), Description: i18n.T("worktree.menu.remove.desc"), Value: "remove"},
	}

	for {
		action, err := worktreeMenu(menuOptions)
		if err != nil {
			return nil // Esc 退出
		}

		switch action {
		case "list":
			infos, err := listWorktrees()
			if err != nil {
				return err
			}
			printWorktrees(infos)
		case "add":
			if err := addWorktreeFlow(); err != nil {
				return err
			}
		case "remove":
			if err := removeWorktreeFlow(); err != nil {
				return err
			}
		}
	}
}
