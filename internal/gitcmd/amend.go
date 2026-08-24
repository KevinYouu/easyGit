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

// amendActionMenu 操作选择菜单(测试可注入)
var amendActionMenu = func(options []config.Option) (string, error) {
	selected, err := form.ListFormColumns(i18n.T("amend.action.title"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// amendFilesSelect 追加文件多选(测试可注入)
var amendFilesSelect = func(options []config.Option) ([]string, error) {
	return form.ListFormColumns(
		i18n.T("amend.select.files"),
		[]form.ColumnSpec{
			{Kind: form.ColumnAuto, MaxWidth: form.MaxAutoColumnWidth},
			{Kind: form.ColumnFlex},
		},
		options,
		form.ListMulti,
	)
}

// amendMessageInput 新消息输入(预填原消息,测试可注入)
var amendMessageInput = func(current string) (string, error) {
	// 仅校验非空:历史提交可能不含类型前缀,不强加约定
	validate := func(msg string) error {
		if strings.TrimSpace(msg) == "" {
			return fmt.Errorf("%s", i18n.T("form.input.empty.error"))
		}
		return nil
	}
	return form.InputWithSuggestions(i18n.T("amend.message.input"), current, validate, config.GetRecentCommitMessages())
}

// amendConfirm 确认框(测试可注入)
var amendConfirm = form.Confirm

// getHeadSubject 获取当前 HEAD 提交消息主题(%s);无提交时报错
func getHeadSubject() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%s")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s %w", i18n.T("error.git.log"), err)
	}
	return strings.TrimSpace(string(output)), nil
}

// isHeadPushed 判断 HEAD 提交是否已存在于任何远程分支
// (amend 会改写历史,已推送则需 force push 覆盖)
func isHeadPushed() bool {
	cmd := exec.Command("git", "branch", "-r", "--contains", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// getUpstreamRef 返回当前分支上游引用(origin/main 形式),无上游返回空串
func getUpstreamRef() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output))
}

// executeAmend 执行 amend(核心执行,交互层与测试共用):
// files 非空时先暂存这些文件并追加进上次提交;message 非空时改写消息,
// 否则保留原消息(--no-edit)。
func executeAmend(files []string, message string) error {
	if len(files) > 0 {
		args := []string{"add"}
		args = append(args, files...)
		if _, err := command.RunCmdWithSpinnerOptions("git", args,
			i18n.T("git.add.selected.loading"),
			i18n.T("git.add.selected.success"),
			false); err != nil {
			return err
		}
	}

	amendArgs := []string{"commit", "--amend"}
	successMsg := i18n.T("amend.append.success")
	if message != "" {
		amendArgs = append(amendArgs, "-m", message)
		successMsg = i18n.T("amend.message.success")
	} else {
		amendArgs = append(amendArgs, "--no-edit")
	}

	_, err := command.RunCmdWithSpinnerOptions("git", amendArgs,
		i18n.T("amend.executing"),
		successMsg,
		true)
	return err
}

// forcePushAmended amend 后可选强制推送(force-with-lease 安全覆盖):
// 仅在有上游且用户确认时执行。
func forcePushAmended() error {
	upstream := getUpstreamRef()
	if upstream == "" {
		logs.Info(i18n.T("amend.force.no.upstream"))
		return nil
	}
	if !amendConfirm(fmt.Sprintf(i18n.T("amend.force.confirm"), upstream)) {
		logs.Info(i18n.T("amend.force.skipped"))
		return nil
	}

	parts := strings.SplitN(upstream, "/", 2)
	remote := parts[0]
	branch := upstream
	if len(parts) == 2 {
		branch = parts[1]
	}

	_, err := command.RunCmdWithSpinnerOptions("git",
		[]string{"push", "--force-with-lease", remote, branch},
		fmt.Sprintf(i18n.T("amend.force.loading"), remote),
		fmt.Sprintf(i18n.T("amend.force.success"), upstream),
		true)
	return err
}

// Amend 交互式修改上次提交:
// 三选一:仅改消息(预填原消息)/ 追加文件到上次提交 / 两者都做;
// HEAD 已推送时先警告需 force push 并确认,完成后可选用
// force-with-lease 安全覆盖远程。
func Amend() error {
	if _, err := getHeadSubject(); err != nil {
		return fmt.Errorf("%s", i18n.T("amend.no.commits"))
	}

	options := []config.Option{
		{Label: i18n.T("amend.option.message"), Description: i18n.T("amend.option.message.desc"), Value: "message"},
		{Label: i18n.T("amend.option.files"), Description: i18n.T("amend.option.files.desc"), Value: "files"},
		{Label: i18n.T("amend.option.both"), Description: i18n.T("amend.option.both.desc"), Value: "both"},
	}

	action, err := amendActionMenu(options)
	if err != nil {
		return nil // Esc 取消
	}

	changeMessage := action == "message" || action == "both"
	appendFiles := action == "files" || action == "both"

	// 历史改写判定:以改写前 HEAD 是否已在远程为准
	// (amend 产生的新提交必然不在远程,事后检查恒为 false)
	pushedBefore := isHeadPushed()

	// 历史改写警告:HEAD 已推送时需 force push 覆盖远程
	if pushedBefore && !amendConfirm(i18n.T("amend.pushed.warning")) {
		logs.Info(i18n.T("amend.cancelled"))
		return nil
	}

	var files []string
	if appendFiles {
		fileStatus, err := getFileStatuses()
		if err != nil {
			return fmt.Errorf("getFileStatuses: %w", err)
		}
		if len(fileStatus) == 0 {
			logs.Info(i18n.T("amend.no.files"))
			return nil
		}
		var fileOptions []config.Option
		for _, fs := range fileStatus {
			if fs.Status == "" {
				continue
			}
			fileOptions = append(fileOptions, config.Option{
				Label: fs.Path,
				Value: fs.Path,
				Cells: []string{fs.Status, fs.Path},
			})
		}

		files, err = amendFilesSelect(fileOptions)
		if err != nil || len(files) == 0 {
			logs.Info(i18n.T("amend.cancelled"))
			return nil
		}
	}

	message := ""
	if changeMessage {
		subject, err := getHeadSubject()
		if err != nil {
			return err
		}
		message, err = amendMessageInput(subject)
		if err != nil {
			return nil // Esc 取消
		}
	}

	if err := executeAmend(files, message); err != nil {
		return err
	}

	// 改写前已推送:提供 force-with-lease 安全覆盖远程的机会
	if pushedBefore {
		return forcePushAmended()
	}
	return nil
}
