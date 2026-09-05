package gitcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

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
		output, err := runWithEditorTrue([]string{"git", "rebase", "--continue"})
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

// squashMessageScript 生成注入提交信息的编辑器脚本路径:
// 创建独立临时目录,Windows 使用 .bat(cmd 语法),其他平台使用 .sh(sh 语法)。
// 脚本与目录由调用方负责清理(RemoveAll)。
func squashMessageScript() (dir, scriptPath string, err error) {
	dir, err = os.MkdirTemp("", "easygit-*")
	if err != nil {
		return "", "", err
	}
	if runtime.GOOS == "windows" {
		return dir, filepath.Join(dir, "easygit-msg-editor.bat"), nil
	}
	return dir, filepath.Join(dir, "easygit-msg-editor.sh"), nil
}

// writeSquashMessageScript 写入消息注入脚本并返回临时目录与脚本路径。
// 脚本内容:回显消息到 git 传入的提交信息文件(%1 / $1)。
// Windows 用 cmd 语法(echo 重定向到已存在的提交信息文件,
// 不会向文件头写入 BOM),其他平台用 sh 语法。
// 消息经转义处理防止 shell 注入(特殊字符 % ^ & | < > \ " $ `)。
func writeSquashMessageScript(newMessage string) (dir, scriptPath string, err error) {
	dir, scriptPath, err = squashMessageScript()
	if err != nil {
		return "", "", err
	}

	var scriptContent string
	if runtime.GOOS == "windows" {
		scriptContent = "@echo off\necho " + escapeBatMessage(newMessage) + "> %1\n"
	} else {
		scriptContent = "#!/bin/sh\necho \"" + escapeShMessage(newMessage) + "\" > \"$1\"\n"
	}

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0o755); err != nil {
		os.RemoveAll(dir)
		return "", "", err
	}
	return dir, scriptPath, nil
}

// escapeBatMessage 转义 Windows cmd 特殊字符防止命令注入:
// % ^ & | < > 需转义为 ^% ^^ ^& ^| ^< ^>
func escapeBatMessage(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '%', '^', '&', '|', '<', '>', '"':
			b.WriteRune('^')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// escapeShMessage 转义 POSIX shell 特殊字符防止命令注入:
// \ " $ ` 需转义为 \\ \" \$ \`
func escapeShMessage(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '"', '$', '`':
			b.WriteRune('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// executablePath 当前进程可执行文件路径(测试可注入,用于验证含空格路径的转义拼接)
var executablePath = func() (string, error) {
	return os.Executable()
}

// runGitCommand 执行 git 命令(测试可注入,用于捕获环境变量而不真正运行变基)
var runGitCommand = func(args, env []string) error {
	cmd := exec.Command("git", args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunInternalRebase(baseCommit, mode string, targets []string, newMessage string) error {
	executable, err := executablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// 通过环境变量传入配置(JSON)
	configData := map[string]any{
		"mode":    mode,
		"targets": targets,
		"message": newMessage,
	}
	configBytes, _ := json.Marshal(configData)

	env := append(os.Environ(),
		fmt.Sprintf("GIT_SEQUENCE_EDITOR=%s _internal_rebase_editor", QuoteForEditorEnv(executable)),
		fmt.Sprintf("EASYGIT_REBASE_CONFIG=%s", string(configBytes)),
	)

	if mode == "squash" && newMessage != "" {
		// 创建临时编辑器脚本注入新提交信息(Windows .bat / 其他 .sh)
		dir, scriptPath, err := writeSquashMessageScript(newMessage)
		if err != nil {
			return fmt.Errorf("failed to create squash message editor script: %w", err)
		}
		defer os.RemoveAll(dir)

		env = append(env, fmt.Sprintf("GIT_EDITOR=%s", QuoteForEditorEnv(scriptPath)))
	} else if mode == "drop" {
		// drop 不需要编辑器;注入 true 跳过,防止 git 打开默认编辑器
		env = append(env, "GIT_EDITOR=true")
	}

	args := []string{"rebase", "-i"}
	if baseCommit == "--root" {
		args = append(args, "--root")
	} else {
		args = append(args, baseCommit)
	}

	logs.Info(i18n.T("rebase.starting"))

	err = runGitCommand(args, env)
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

// ─── 冲突解决闭环(通用实现见 conflict.go) ──────────────────────────────────

// handleRebaseConflict 变基冲突解决闭环:委托通用闭环(rebase 配置)。
// 兼容保留原签名,既有调用点与测试不变。
func handleRebaseConflict() error {
	_, err := runConflictResolution(rebaseConflictOps)
	return err
}

// getUnmergedFiles / runGitRebaseContinue / 冲突编辑器等共享工具
// 已移至 conflict.go(rebase/merge/cherry-pick 三操作共用)。

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
