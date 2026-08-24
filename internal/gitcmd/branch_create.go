package gitcmd

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

// branchCreateBaseSelect 基点选择菜单(测试可注入)
var branchCreateBaseSelect = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("branch.create.base.select"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// validateBranchName 分支名校验(表单层):非空且不含 git refname 非法字符。
// 拦截常见错误:空格、以 - 开头(被解析为选项)、".."(路径穿越)、
// ~ ^ : ? * [ \ 控制字符、以 .lock 结尾、以 / 或 . 开头或结尾。
func validateBranchName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s", i18n.T("form.input.empty.error"))
	}
	switch {
	case strings.ContainsAny(name, " \t~^:?*[\\"),
		strings.HasPrefix(name, "-"),
		strings.HasPrefix(name, "."),
		strings.HasPrefix(name, "/"),
		strings.HasSuffix(name, "/"),
		strings.HasSuffix(name, "."),
		strings.HasSuffix(name, ".lock"),
		strings.Contains(name, ".."),
		strings.Contains(name, "//"):
		return fmt.Errorf("%s", i18n.T("branch.create.name.invalid"))
	}
	return nil
}

// branchCreatePushRunner 推送执行器(测试可注入:进度 UI 需 TTY,
// 测试环境注入同步执行版本;生产走并行进度模型)
var branchCreatePushRunner = func(commands []command.CommandInfo) error {
	return command.RunMultipleCommandsParallel(commands, 0)
}

// createBranch 新建分支并切换(核心执行,交互层与测试共用):
// base 为空串时从当前 HEAD 创建,否则从指定基点(本地分支/远程引用)创建;
// push 为 true 时向每个 remote 推送并设 upstream(并行段)。
func createBranch(name, base string, push bool, remotes []string) error {
	args := []string{"checkout", "-b", name}
	if base != "" {
		args = append(args, base)
	}
	if _, err := command.RunCmdWithSpinnerOptions("git", args,
		fmt.Sprintf(i18n.T("branch.create.loading"), name),
		fmt.Sprintf(i18n.T("branch.create.success"), name),
		true); err != nil {
		return err
	}

	if !push || len(remotes) == 0 {
		return nil
	}

	var commands []command.CommandInfo
	for _, remote := range remotes {
		commands = append(commands, command.CommandInfo{
			Command:     "git",
			Args:        []string{"push", "-u", remote, name},
			Description: fmt.Sprintf(i18n.T("git.push.to.remote"), remote),
			LoadingMsg:  fmt.Sprintf(i18n.T("branch.create.push.loading"), name, remote),
			SuccessMsg:  fmt.Sprintf(i18n.T("branch.create.push.success"), name, remote),
		})
	}
	// 新分支推送相互独立,一次性并行启动
	return branchCreatePushRunner(commands)
}

// branchCreateNameInput 名称输入(测试可注入,默认走 form.InputWithValidate)
var branchCreateNameInput = func(validate func(string) error) (string, error) {
	return form.InputWithValidate(i18n.T("branch.create.name.title"), "", validate)
}

// branchCreatePushConfirm 推送确认(测试可注入)
var branchCreatePushConfirm = func(title string) bool {
	return form.Confirm(title)
}

// branchCreateRemotesSelect 远程选择(测试可注入,复用配置持久化逻辑)
var branchCreateRemotesSelect = SelectAndSaveRemotes

// CreateBranch 交互式新建分支:
// 输入名称(表单层校验 refname 合法性)→ 选择基点(当前 HEAD / 本地与远程分支)
// → 可选推送并设置 upstream。
func CreateBranch() error {
	name, err := branchCreateNameInput(validateBranchName)
	if err != nil {
		return nil // Esc 取消
	}

	baseOptions := []config.Option{
		{Label: i18n.T("branch.create.base.head"), Description: i18n.T("branch.create.base.head.desc"), Value: ""},
	}
	for _, b := range getRemoteBranchRefs() {
		if b != "origin/HEAD" && localNameForRemoteBranch(b) != "" {
			baseOptions = append(baseOptions, config.Option{Label: fmt.Sprintf("%s (%s)", b, i18n.T("branch.label.remote")), Value: b})
		}
	}
	if current, err := GetCurrentBranch(); err == nil {
		if locals, err := GetAllBranches(); err == nil {
			for _, l := range locals {
				if l != current && !strings.HasPrefix(l, "(") {
					baseOptions = append(baseOptions, config.Option{Label: fmt.Sprintf("%s (%s)", l, i18n.T("branch.label.local")), Value: l})
				}
			}
		}
	}

	base, err := branchCreateBaseSelect(baseOptions)
	if err != nil {
		return nil // Esc 取消
	}

	push := false
	var remotes []string
	if branchCreatePushConfirm(i18n.T("branch.create.push.confirm")) {
		selected, err := branchCreateRemotesSelect()
		if err != nil {
			return nil // 无远程或取消:降级仅本地创建
		}
		remotes = selected
		push = len(remotes) > 0
	}

	return createBranch(name, base, push, remotes)
}
