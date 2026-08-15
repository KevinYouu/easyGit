package gitcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// rebaseConflictMenu 冲突恢复菜单选择(测试可注入,默认走 form.ListForm)
var rebaseConflictMenu = func(options []config.Option) (string, error) {
	selected, err := form.ListForm(i18n.T("rebase.conflict.menu.title"), options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

func RebaseIntoCurrent() error {
	// Check if rebase is in progress
	if isRebaseInProgress() {
		return handleInProgressRebase()
	}

	// Not in progress, check if working directory is clean
	if err := checkWorkingDirectoryStatus(); err != nil {
		logs.Error(i18n.T("rebase.uncommitted.changes"))
		return fmt.Errorf("working directory is not clean")
	}

	return handleStandardRebase()
}

func isRebaseInProgress() bool {
	gitDir, err := getGitDir()
	if err != nil {
		return false
	}

	rebaseMerge := filepath.Join(gitDir, "rebase-merge")
	rebaseApply := filepath.Join(gitDir, "rebase-apply")

	if _, err := os.Stat(rebaseMerge); !os.IsNotExist(err) {
		return true
	}
	if _, err := os.Stat(rebaseApply); !os.IsNotExist(err) {
		return true
	}
	return false
}

func getGitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func handleInProgressRebase() error {
	options := []config.Option{
		{Label: i18n.T("rebase.action.continue"), Description: i18n.T("rebase.action.continue.desc"), Value: "--continue"},
		{Label: i18n.T("rebase.action.skip"), Description: i18n.T("rebase.action.skip.desc"), Value: "--skip"},
		{Label: i18n.T("rebase.action.abort"), Description: i18n.T("rebase.action.abort.desc"), Value: "--abort"},
	}

	actions, err := form.ListForm(i18n.T("rebase.status.in_progress"), options, form.ListSingle)
	if err != nil {
		return err
	}
	action := actions[0]

	switch action {
	case "--abort":
		output, err := command.RunCmd("git", []string{"rebase", "--abort"}, "")
		if err != nil {
			return handleRebaseError(output, err)
		}
		logs.Info(i18n.T("rebase.abort.message"))
		return nil
	case "--continue":
		output, err := runGitRebaseContinue()
		if err != nil {
			// 失败但变基仍在进行(如 pre-commit hook 失败):先展示失败原因,避免被吞
			if trimmed := strings.TrimSpace(output); trimmed != "" {
				logs.Error(trimmed)
			}
			if !isRebaseInProgress() {
				return handleRebaseError(output, err)
			}
		}
	default: // --skip
		output, err := command.RunCmd("git", []string{"rebase", "--skip"}, "")
		if err != nil {
			if trimmed := strings.TrimSpace(output); trimmed != "" {
				logs.Error(trimmed)
			}
			if !isRebaseInProgress() {
				return handleRebaseError(output, err)
			}
		}
	}

	// continue/skip 后仍处于变基状态(新的冲突):进入冲突解决闭环
	if isRebaseInProgress() {
		return handleRebaseConflict()
	}
	logs.Info(i18n.T("rebase.success.message"))
	return nil
}

func handleStandardRebase() error {
	branches, err := getAllAvailableBranches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	if len(branches) == 0 {
		logs.Info(i18n.T("merge.no.branches"))
		return nil
	}

	selectedBranches, err := form.ListForm(i18n.T("rebase.select.target"), branches, form.ListSingle)
	if err != nil {
		return fmt.Errorf(i18n.T("error.select.form.detail")+": %w", err)
	}
	selectedBranch := selectedBranches[0]

	logs.Info(fmt.Sprintf(i18n.T("rebase.starting"), selectedBranch))
	output, err := command.RunCmd("git", []string{"rebase", selectedBranch}, "")
	if err != nil {
		// 冲突:进入冲突解决闭环(不再直接退出)
		if isRebaseInProgress() {
			return handleRebaseConflict()
		}
		return handleRebaseError(output, err)
	}

	logs.Info(i18n.T("rebase.success.message"))
	return nil
}

// GetRecentCommits 已移至 commits.go(统一数据源 GetCommitsOptions 的薄封装),
// 此处不再重复实现。

func RunInternalRebase(baseCommit, mode string, targets []string, newMessage string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Pass config via env var as JSON
	configData := map[string]any{
		"mode":    mode,
		"targets": targets,
		"message": newMessage,
	}
	configBytes, _ := json.Marshal(configData)

	env := append(os.Environ(),
		fmt.Sprintf("GIT_SEQUENCE_EDITOR=%s _internal_rebase_editor", executable),
		fmt.Sprintf("EASYGIT_REBASE_CONFIG=%s", string(configBytes)),
	)

	if mode == "squash" && newMessage != "" {
		// Create a temporary script to act as GIT_EDITOR to inject our new message
		scriptPath := filepath.Join(os.TempDir(), "easygit-msg-editor.sh")
		// The script echoes the message into the file passed as $1
		scriptContent := fmt.Sprintf("#!/bin/sh\necho \"%s\" > $1\n", newMessage)
		os.WriteFile(scriptPath, []byte(scriptContent), 0755)

		env = append(env, fmt.Sprintf("GIT_EDITOR=%s", scriptPath))
	} else if mode == "squash" {
		// If no message provided, allow default editor to pop up, or just use 'true' to accept default?
		// We'll let git use the user's default editor if they didn't provide a message, so they can edit it.
	} else if mode == "drop" {
		// For drop, we don't need an editor, but just in case git tries to open one, we use 'true' to skip it
		env = append(env, "GIT_EDITOR=true")
	}

	args := []string{"rebase", "-i"}
	if baseCommit == "--root" {
		args = append(args, "--root")
	} else {
		args = append(args, baseCommit)
	}

	logs.Info(i18n.T("rebase.starting"))

	rebaseCmd := exec.Command("git", args...)
	rebaseCmd.Env = env
	rebaseCmd.Stdin = os.Stdin
	rebaseCmd.Stdout = os.Stdout
	rebaseCmd.Stderr = os.Stderr

	err = rebaseCmd.Run()
	if err != nil {
		// 冲突:进入冲突解决闭环(squash/drop 共用)
		if isRebaseInProgress() {
			return handleRebaseConflict()
		}
		return fmt.Errorf("rebase failed: %w", err)
	}

	logs.Info(i18n.T("rebase.success.message"))
	return nil
}

func handleRebaseError(output string, err error) error {
	outputStr := strings.TrimSpace(output)

	if strings.Contains(outputStr, "CONFLICT") && isRebaseInProgress() {
		return handleRebaseConflict()
	}

	logs.Error(i18n.T("rebase.failed") + ": " + outputStr)
	return fmt.Errorf("git rebase failed: %s", outputStr)
}

// ─── 冲突解决闭环 ────────────────────────────────────────────────────────────

// handleRebaseConflict 冲突解决闭环:循环展示未合并文件与操作菜单,
// 直到变基完成(continue/skip 后不再处于冲突状态)或用户选择中止/退出。
// 多提交多冲突场景下自动循环,解决"第一次冲突后即退出"的问题。
func handleRebaseConflict() error {
	for {
		files := getUnmergedFiles()

		logs.Error(i18n.T("rebase.conflict.detected"))
		if len(files) > 0 {
			logs.Info(i18n.T("rebase.conflict.files"))
			for _, f := range files {
				fmt.Println("  " + f)
			}
		}

		options := []config.Option{
			{Label: i18n.T("rebase.conflict.menu.edit"), Description: i18n.T("rebase.conflict.menu.edit.desc"), Value: "edit"},
			{Label: i18n.T("rebase.conflict.menu.continue"), Description: i18n.T("rebase.conflict.menu.continue.desc"), Value: "continue"},
			{Label: i18n.T("rebase.conflict.menu.skip"), Description: i18n.T("rebase.conflict.menu.skip.desc"), Value: "skip"},
			{Label: i18n.T("rebase.conflict.menu.abort"), Description: i18n.T("rebase.conflict.menu.abort.desc"), Value: "abort"},
			{Label: i18n.T("rebase.conflict.menu.quit"), Description: i18n.T("rebase.conflict.menu.quit.desc"), Value: "quit"},
		}

		action, err := rebaseConflictMenu(options)
		if err != nil {
			return err
		}

		switch action {
		case "edit":
			openConflictsInEditor(files)
		case "continue":
			// 暂存已解决的文件后继续
			if len(files) > 0 {
				// 仍含冲突标记的文件:确认后按当前内容暂存,避免冲突标记误入提交
				if hasUnresolvedMarkers(files) {
					confirmed := form.Confirm(i18n.T("rebase.conflict.unresolved.confirm"))
					if !confirmed {
						continue
					}
				}
				args := append([]string{"add"}, files...)
				if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
					logs.Error(i18n.T("rebase.conflict.git.add.failed") + ": " + strings.TrimSpace(string(out)))
					continue
				}
			}
			output, err := runGitRebaseContinue()
			if err != nil {
				// 失败但变基仍在进行(如 pre-commit hook 失败):先展示失败原因,避免被吞
				if trimmed := strings.TrimSpace(output); trimmed != "" {
					logs.Error(trimmed)
				}
			}
			if !isRebaseInProgress() {
				if err != nil {
					return handleRebaseError(output, err)
				}
				logs.Info(i18n.T("rebase.success.message"))
				return nil
			}
			logs.Info(i18n.T("rebase.conflict.still"))
		case "skip":
			output, err := command.RunCmd("git", []string{"rebase", "--skip"}, "")
			if err != nil {
				if trimmed := strings.TrimSpace(output); trimmed != "" {
					logs.Error(trimmed)
				}
			}
			if !isRebaseInProgress() {
				if err != nil {
					return handleRebaseError(output, err)
				}
				logs.Info(i18n.T("rebase.success.message"))
				return nil
			}
			logs.Info(i18n.T("rebase.conflict.still"))
		case "abort":
			if output, err := command.RunCmd("git", []string{"rebase", "--abort"}, ""); err != nil {
				return handleRebaseError(output, err)
			}
			logs.Info(i18n.T("rebase.abort.message"))
			return nil
		case "quit":
			return fmt.Errorf("rebase conflict remains: resolve manually or run 'easyGit rebase' again")
		}
	}
}

// getUnmergedFiles 获取当前未合并(冲突)文件列表
func getUnmergedFiles() []string {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files
}

// runGitRebaseContinue 执行 git rebase --continue。
// 注入 GIT_EDITOR=true:非 TTY 环境下 git 也会启动编辑器获取提交信息,
// pick 提交直接接受原始信息,reword/squash 接受工具内已注入的消息,
// 避免流程挂起等待编辑器输入。
func runGitRebaseContinue() (string, error) {
	env := append(os.Environ(), "GIT_EDITOR=true")
	cmd := exec.Command("git", "rebase", "--continue")
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// openConflictsInEditor 打开冲突文件:配置中心编辑器优先,其次 $EDITOR/$VISUAL,
// 最后回退 vim/vi/nano(Windows 追加 notepad);已知异步编辑器(code/subl/atom)
// 自动补 -w 等待标志;无可用编辑器时展示文件清单,提示手动解决后回到菜单继续
func openConflictsInEditor(files []string) {
	editor := ""
	if configured, err := config.GetConflictEditor(); err == nil && configured != "" {
		editor = configured
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		candidates := []string{"vim", "vi", "nano"}
		if runtime.GOOS == "windows" {
			candidates = append(candidates, "notepad")
		}
		for _, candidate := range candidates {
			if path, err := exec.LookPath(candidate); err == nil {
				editor = path
				break
			}
		}
	}

	if editor == "" {
		logs.Error(i18n.T("rebase.conflict.no.editor"))
		for _, f := range files {
			fmt.Println("  " + f)
		}
		logs.Info(i18n.T("rebase.conflict.manual.hint"))
		return
	}

	program, args, perFile := resolveConflictEditor(editor, files)

	// Windows 记事本:一次只支持一个文件且无等待标志,经 start /wait 逐个打开
	if perFile && runtime.GOOS == "windows" {
		for _, f := range files {
			cmd := exec.Command("cmd", "/c", "start", "/wait", program, f)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				logs.Error(i18n.T("rebase.conflict.editor.failed") + ": " + err.Error())
			}
		}
		return
	}

	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logs.Error(i18n.T("rebase.conflict.editor.failed") + ": " + err.Error())
	}
}

// asyncEditors 启动后立即返回的编辑器:需补等待标志(-w/--wait)才能在编辑完成前阻塞流程
var asyncEditors = map[string]bool{
	"code": true, // VS Code
	"subl": true, // Sublime Text
	"atom": true, // Atom
}

// resolveConflictEditor 解析编辑器命令为可执行参数:
// 1. 按空白拆分(支持 "code -w" 这类带参数配置;引号包裹的 Windows 路径原样保留)
// 2. 已知异步编辑器自动补 -w(已带等待标志不重复)
// 3. 返回 perFile=true 表示该编辑器一次只能打开一个文件(Windows notepad)
func resolveConflictEditor(editor string, files []string) (program string, args []string, perFile bool) {
	fields := splitCommand(editor)
	program = fields[0]
	args = fields[1:]

	base := program
	if idx := strings.LastIndexAny(base, `/\`); idx >= 0 {
		base = base[idx+1:]
	}
	base = strings.TrimSuffix(base, ".exe")
	if asyncEditors[base] && !hasWaitFlag(args) {
		args = append(args, "-w")
	}
	args = append(args, files...)

	if strings.EqualFold(base, "notepad") {
		perFile = true
	}
	return program, args, perFile
}

// hasWaitFlag 检查参数中是否已有等待标志
func hasWaitFlag(args []string) bool {
	for _, a := range args {
		if a == "-w" || a == "--wait" || a == "--wait-for-input" {
			return true
		}
	}
	return false
}

// conflictMarkers 冲突标记:文件含任一时视为未解决(git 不会拦截含标记内容的提交)
var conflictMarkers = [][]byte{[]byte("<<<<<<<"), []byte(">>>>>>>")}

// hasUnresolvedMarkers 检查文件列表是否仍有未解决文件(内容含冲突标记)
func hasUnresolvedMarkers(files []string) bool {
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, marker := range conflictMarkers {
			if bytes.Contains(content, marker) {
				return true
			}
		}
	}
	return false
}

// splitCommand 拆分命令字符串为参数:引号包裹的路径(如 Windows 带空格路径)
// 原样保留,不在引号内的空白作为分隔符。
// TODO: 如需支持 POSIX 单引号与反斜杠转义,在此扩展(当前 Windows 双引号场景已够用)。
func splitCommand(s string) []string {
	var args []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuote = !inQuote
		case unicode.IsSpace(r) && !inQuote:
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}

// getParentHash gets the parent hash of a specific commit
func getParentHash(hash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%P", hash)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	parents := strings.Fields(string(output))
	if len(parents) == 0 {
		return "", nil // No parent (Root commit)
	}
	return parents[0], nil // Return first parent
}
