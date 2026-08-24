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

// SwitchableBranch 可切换分支:本地分支或远程分支(远程切换自动建跟踪分支)
type SwitchableBranch struct {
	Name     string // 本地分支名(远程 origin/a/b 时为 a/b)
	Ref      string // git 引用(本地名或 origin/a/b 完整引用)
	IsRemote bool
}

// branchSwitchSelect 分支选择菜单(测试可注入,默认走自适应双列列表)
var branchSwitchSelect = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("branch.switch.select"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// branchSwitchDirtyAction 脏工作区处理选择菜单(测试可注入)
var branchSwitchDirtyAction = func(options []config.Option) (string, error) {
	selected, err := form.ListForm(i18n.T("branch.switch.dirty.title"), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// listSwitchableBranches 列出可切换分支:本地(排除当前与非法态)+ 远程。
// 远程去重:已有同名本地分支时跳过对应远程项(切本地即达,避免重复选项)。
func listSwitchableBranches() ([]SwitchableBranch, error) {
	current, err := GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("get current branch: %w", err)
	}

	localBranches, err := GetAllBranches()
	if err != nil {
		return nil, fmt.Errorf("get branches: %w", err)
	}

	localNames := make(map[string]bool, len(localBranches))
	var branches []SwitchableBranch
	for _, name := range localBranches {
		// 排除当前分支与 detached 等非法展示项
		if name == current || strings.HasPrefix(name, "(") {
			continue
		}
		localNames[name] = true
		branches = append(branches, SwitchableBranch{Name: name, Ref: name})
	}

	for _, ref := range getRemoteBranchRefs() {
		name := localNameForRemoteBranch(ref)
		if name == "" || localNames[name] || localNames[ref] {
			continue
		}
		branches = append(branches, SwitchableBranch{Name: name, Ref: ref, IsRemote: true})
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("%s", i18n.T("branch.switch.none"))
	}
	return branches, nil
}

// getRemoteBranchRefs 获取全部远程分支引用(origin/dev 形式),排除 HEAD 指向
func getRemoteBranchRefs() []string {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	var refs []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "HEAD") {
			continue
		}
		refs = append(refs, line)
	}
	return refs
}

// localNameForRemoteBranch 远程引用转本地分支名:"origin/a/b" → "a/b";
// 不含 "/"(非 remote/branch 结构)返回空串
func localNameForRemoteBranch(ref string) string {
	idx := strings.Index(ref, "/")
	if idx < 0 || idx == len(ref)-1 {
		return ""
	}
	return ref[idx+1:]
}

// isWorkingDirectoryDirty 工作区是否有未提交修改(含暂存区)
func isWorkingDirectoryDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// switchToBranch 切换到目标分支(核心执行,交互层与测试共用):
//   - 本地分支直接 checkout;远程分支 checkout 自动创建同名跟踪分支(git DWIM)
//   - stashFirst 为 true 时先 stash 当前修改,切换成功后 pop 回新分支;
//     pop 冲突时 git 保留 stash 条目,提示用户手动处理(切换本身已成功)
func switchToBranch(target SwitchableBranch, stashFirst bool) error {
	stashed := false
	if stashFirst {
		from, _ := GetCurrentBranch()
		stashMsg := fmt.Sprintf("easyGit: before switch %s -> %s", from, target.Name)
		_, err := command.RunCmdWithSpinnerOptions("git",
			[]string{"stash", "push", "-m", stashMsg},
			i18n.T("branch.switch.stash.push.loading"),
			fmt.Sprintf(i18n.T("branch.switch.stash.push.success"), target.Name),
			false)
		if err != nil {
			return fmt.Errorf(i18n.T("branch.switch.stash.push.failed")+" %w", err)
		}
		stashed = true
	}

	args := []string{"checkout"}
	if target.IsRemote {
		args = append(args, "--track")
	}
	args = append(args, target.Ref)

	if _, err := command.RunCmdWithSpinnerOptions("git", args,
		fmt.Sprintf(i18n.T("branch.switch.loading"), target.Name),
		fmt.Sprintf(i18n.T("branch.switch.success"), target.Name),
		true); err != nil {
		return err
	}

	if stashed {
		output, err := command.RunCmdWithSpinnerOptions("git",
			[]string{"stash", "pop"},
			i18n.T("branch.switch.stash.pop.loading"),
			i18n.T("branch.switch.stash.pop.success"),
			false)
		if err != nil {
			// 冲突时 git 自动保留 stash 条目;非冲突失败同样保留,由用户手动处理
			logs.Error(i18n.T("branch.switch.stash.pop.conflict"))
			if trimmed := strings.TrimSpace(output); trimmed != "" && !strings.Contains(trimmed, "CONFLICT") {
				logs.Error(trimmed)
			}
		}
	}
	return nil
}

// SwitchBranch 交互式切换分支:
// 脏工作区先询问处理方式(携带修改 / 自动 stash / 取消),
// 随后列出本地 + 远程分支单选(远程自动建跟踪分支),`/` 过滤可用。
func SwitchBranch() error {
	dirty, err := isWorkingDirectoryDirty()
	if err != nil {
		return err
	}

	stashFirst := false
	if dirty {
		options := []config.Option{
			{Label: i18n.T("branch.switch.dirty.carry"), Description: i18n.T("branch.switch.dirty.carry.desc"), Value: "carry"},
			{Label: i18n.T("branch.switch.dirty.stash"), Description: i18n.T("branch.switch.dirty.stash.desc"), Value: "stash"},
			{Label: i18n.T("branch.switch.dirty.cancel"), Description: i18n.T("branch.switch.dirty.cancel.desc"), Value: "cancel"},
		}
		action, err := branchSwitchDirtyAction(options)
		if err != nil {
			return nil // Esc 取消
		}
		switch action {
		case "cancel":
			logs.Info(i18n.T("branch.switch.cancelled"))
			return nil
		case "stash":
			stashFirst = true
		}
	}

	branches, err := listSwitchableBranches()
	if err != nil {
		return err
	}

	options := make([]config.Option, 0, len(branches))
	for _, b := range branches {
		label := fmt.Sprintf("%s (%s)", b.Name, i18n.T("branch.label.local"))
		desc := ""
		if b.IsRemote {
			label = fmt.Sprintf("%s (%s)", b.Ref, i18n.T("branch.label.remote"))
			desc = i18n.T("branch.switch.remote.desc")
		}
		options = append(options, config.Option{Label: label, Description: desc, Value: b.Ref})
	}

	ref, err := branchSwitchSelect(options)
	if err != nil {
		return nil // Esc 取消
	}

	var target SwitchableBranch
	for _, b := range branches {
		if b.Ref == ref {
			target = b
			break
		}
	}

	return switchToBranch(target, stashFirst)
}
