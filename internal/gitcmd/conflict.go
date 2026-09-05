package gitcmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// ─── 通用冲突解决闭环(rebase / merge / cherry-pick 共用) ─────────────────────

// errConflictAborted 用户在冲突闭环中主动中止操作(调用方据此停止后续批次)
var errConflictAborted = errors.New("conflict resolution aborted")

// conflictOps 描述一种会产生冲突的 git 操作,由通用闭环驱动
type conflictOps struct {
	displayName func() string // 本地化操作名(菜单标题/提示文案用)
	continueCmd []string      // 继续(必填):如 {"merge", "--continue"}
	skipCmd     []string      // 跳过(nil 表示该操作不支持跳过,如 merge)
	abortCmd    []string      // 中止:如 {"merge", "--abort"}
	inProgress  func() bool   // 是否仍处于该操作状态
	successKey  string        // 完成消息 i18n 键
	abortKey    string        // 中止消息 i18n 键
}

// conflictMenu 操作菜单选择(测试可注入)
var conflictMenu = func(title string, options []config.Option) (string, error) {
	selected, err := form.ListForm(title, options, form.ListSingle)
	if err != nil {
		return "", err
	}
	return selected[0], nil
}

// runConflictResolution 冲突解决闭环:循环展示未合并文件与操作菜单,
// 直到操作完成(continue/skip 后不再处于冲突状态)或用户选择中止/退出。
// 多提交多冲突场景下自动循环。返回值 aborted=true 表示用户主动中止
// (操作已回滚,调用方应停止后续批次且不再报错)。
func runConflictResolution(ops conflictOps) (bool, error) {
	for {
		files := getUnmergedFiles()

		logs.Error(fmt.Sprintf(i18n.T("conflict.loop.detected"), ops.displayName()))
		if len(files) > 0 {
			logs.Info(i18n.T("conflict.loop.files"))
			muted := lipgloss.NewStyle().Foreground(theme.MutedForeground)
			for _, f := range files {
				fmt.Printf("  %s\n", muted.Render(f))
			}
		}

		options := []config.Option{
			{Label: i18n.T("conflict.loop.menu.edit"), Description: i18n.T("conflict.loop.menu.edit.desc"), Value: "edit"},
			{Label: i18n.T("conflict.loop.menu.continue"), Description: i18n.T("conflict.loop.menu.continue.desc"), Value: "continue"},
		}
		if len(ops.skipCmd) > 0 {
			options = append(options,
				config.Option{Label: i18n.T("conflict.loop.menu.skip"), Description: i18n.T("conflict.loop.menu.skip.desc"), Value: "skip"})
		}
		options = append(options,
			config.Option{Label: i18n.T("conflict.loop.menu.abort"), Description: i18n.T("conflict.loop.menu.abort.desc"), Value: "abort"},
			config.Option{Label: i18n.T("conflict.loop.menu.quit"), Description: i18n.T("conflict.loop.menu.quit.desc"), Value: "quit"},
		)

		action, err := conflictMenu(fmt.Sprintf(i18n.T("conflict.loop.menu.title"), ops.displayName()), options)
		if err != nil {
			return false, err
		}

		switch action {
		case "edit":
			openConflictsInEditor(files)
		case "continue":
			done, err := conflictContinue(ops, files)
			if done || err != nil {
				return false, err
			}
		case "skip":
			output, err := runWithEditorTrue(ops.skipCmd)
			if err != nil {
				if trimmed := strings.TrimSpace(output); trimmed != "" {
					logs.Error(trimmed)
				}
			}
			if !ops.inProgress() {
				if err != nil {
					return false, fmt.Errorf("git %s failed: %s", ops.skipCmd[1], strings.TrimSpace(output))
				}
				logs.Info(i18n.T(ops.successKey))
				return false, nil
			}
			logs.Info(i18n.T("conflict.loop.still"))
		case "abort":
			if output, err := commandRunCombined(ops.abortCmd); err != nil {
				return false, fmt.Errorf("git %s failed: %s", ops.abortCmd[1], strings.TrimSpace(output))
			}
			logs.Info(i18n.T(ops.abortKey))
			return true, nil
		case "quit":
			return false, fmt.Errorf("%s", fmt.Sprintf(i18n.T("conflict.loop.quit.error"), ops.displayName()))
		}
	}
}

// conflictContinue 暂存已解决的文件并执行 --continue。
// 返回 done=true 表示整个操作已结束(成功或最终失败)。
func conflictContinue(ops conflictOps, files []string) (done bool, err error) {
	if len(files) > 0 {
		// 仍含冲突标记的文件:确认后按当前内容暂存,避免冲突标记误入提交
		if hasUnresolvedMarkers(files) {
			if !form.Confirm(i18n.T("conflict.loop.unresolved.confirm")) {
				return false, nil
			}
		}
		absFiles := resolveAbsolutePaths(files)
		args := append([]string{"add"}, absFiles...)
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			logs.Error(i18n.T("conflict.loop.git.add.failed") + ": " + strings.TrimSpace(string(out)))
			return false, nil
		}
	}

	output, err := runWithEditorTrue(ops.continueCmd)
	if err != nil {
		// 失败但操作仍在进行(如 pre-commit hook 失败):先展示失败原因,避免被吞
		if trimmed := strings.TrimSpace(output); trimmed != "" {
			logs.Error(trimmed)
		}
	}
	if !ops.inProgress() {
		if err != nil {
			return true, fmt.Errorf("git %s failed: %s", ops.continueCmd[1], strings.TrimSpace(output))
		}
		logs.Info(i18n.T(ops.successKey))
		return true, nil
	}
	logs.Info(i18n.T("conflict.loop.still"))
	return false, nil
}

// ─── 操作状态检测 ────────────────────────────────────────────────────────────

// statePathExists 判断 .git 下相对路径是否存在
func statePathExists(rel ...string) bool {
	gitDir, err := getGitDir()
	if err != nil {
		return false
	}
	_, statErr := os.Stat(filepath.Join(gitDir, filepath.Join(rel...)))
	return !os.IsNotExist(statErr)
}

// isMergeInProgress 合并进行中(MERGE_HEAD 存在)
func isMergeInProgress() bool {
	return statePathExists("MERGE_HEAD")
}

// isCherryPickInProgress 摘取进行中(CHERRY_PICK_HEAD 或 sequencer 目录存在)
func isCherryPickInProgress() bool {
	return statePathExists("CHERRY_PICK_HEAD") || statePathExists("sequencer")
}

// ─── 三种操作的闭环配置 ──────────────────────────────────────────────────────

// rebaseConflictOps 变基冲突配置
var rebaseConflictOps = conflictOps{
	displayName: func() string { return i18n.T("op.rebase.name") },
	continueCmd: []string{"git", "rebase", "--continue"},
	skipCmd:     []string{"git", "rebase", "--skip"},
	abortCmd:    []string{"git", "rebase", "--abort"},
	inProgress:  isRebaseInProgress,
	successKey:  "rebase.success.message",
	abortKey:    "rebase.abort.message",
}

// mergeConflictOps 合并冲突配置(merge 无 --skip)
var mergeConflictOps = conflictOps{
	displayName: func() string { return i18n.T("op.merge.name") },
	continueCmd: []string{"git", "merge", "--continue"},
	abortCmd:    []string{"git", "merge", "--abort"},
	inProgress:  isMergeInProgress,
	successKey:  "merge.success.message",
	abortKey:    "merge.abort.message",
}

// cherryPickConflictOps 摘取冲突配置
var cherryPickConflictOps = conflictOps{
	displayName: func() string { return i18n.T("op.cherry-pick.name") },
	continueCmd: []string{"git", "cherry-pick", "--continue"},
	skipCmd:     []string{"git", "cherry-pick", "--skip"},
	abortCmd:    []string{"git", "cherry-pick", "--abort"},
	inProgress:  isCherryPickInProgress,
	successKey:  "cherry.pick.resolved",
	abortKey:    "cherry.pick.abort.message",
}

// ─── 共享底层工具 ────────────────────────────────────────────────────────────

// commandRunCombined 执行命令返回组合输出
func commandRunCombined(args []string) (string, error) {
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	return string(out), err
}

// runWithEditorTrue 注入 GIT_EDITOR=true 执行命令:
// 非交互场景下 git 会启动编辑器获取提交信息(--continue 等),
// 注入 true 接受预设信息,避免流程挂起等待编辑器输入。
func runWithEditorTrue(args []string) (string, error) {
	env := append(os.Environ(), "GIT_EDITOR=true")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	return string(output), err
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

// getGitWorkTree 获取工作树根绝对路径(Windows 相对路径解析需要)
func getGitWorkTree() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// resolveAbsolutePaths 将相对路径列表转为工作树根下的绝对路径
func resolveAbsolutePaths(files []string) []string {
	if len(files) == 0 {
		return files
	}
	workTree, err := getGitWorkTree()
	if err != nil {
		return files
	}
	absFiles := make([]string, 0, len(files))
	for _, f := range files {
		if filepath.IsAbs(f) {
			absFiles = append(absFiles, f)
		} else {
			absFiles = append(absFiles, filepath.Join(workTree, f))
		}
	}
	return absFiles
}

// openConflictsInEditor 打开冲突文件:配置中心编辑器优先,其次 $EDITOR/$VISUAL,
// 最后回退 vim/vi/nano(Windows 追加 notepad);已知异步编辑器(code/subl/atom)
// 自动补 -w 等待标志;无可用编辑器时展示文件清单,提示手动解决后回到菜单继续
func openConflictsInEditor(files []string) {
	editor := resolveAvailableEditor()
	if editor == "" {
		logs.Error(i18n.T("conflict.editor.none"))
		muted := lipgloss.NewStyle().Foreground(theme.MutedForeground)
		for _, f := range files {
			fmt.Printf("  %s\n", muted.Render(f))
		}
		logs.Info(i18n.T("conflict.editor.manual.hint"))
		return
	}

	files = resolveAbsolutePaths(files)
	program, args, perFile := resolveConflictEditor(editor, files)

	// Windows 记事本:一次只支持一个文件且无等待标志,经 start /wait 逐个打开
	if perFile && runtime.GOOS == "windows" {
		for _, f := range files {
			cmd := exec.Command("cmd", "/c", "start", "/wait", program, f)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				logs.Error(i18n.T("conflict.editor.failed") + ": " + err.Error())
			}
		}
		return
	}

	cmd := exec.Command(program, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logs.Error(i18n.T("conflict.editor.failed") + ": " + err.Error())
	}
}

// resolveAvailableEditor 编辑器解析顺序:配置中心 > $EDITOR > $VISUAL >
// Windows 优先检测 notepad, code, notepad++, subl, atom;其他平台回退 vim/vi/nano
func resolveAvailableEditor() string {
	if configured, err := config.GetConflictEditor(); err == nil && configured != "" {
		return configured
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	candidates := []string{"vim", "vi", "nano"}
	if runtime.GOOS == "windows" {
		candidates = []string{"notepad", "code", "notepad++", "subl", "atom", "vim", "nano"}
	}
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	return ""
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
