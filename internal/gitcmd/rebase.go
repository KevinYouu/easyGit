package gitcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// normalizeCommitMessage 规范化注入的提交信息:统一换行为 \n,
// 去掉尾部空行后补一个换行。段内换行与空行原样保留,
// git 侧不再需要二次整理。
func normalizeCommitMessage(message string) string {
	message = strings.ReplaceAll(message, "\r\n", "\n")
	message = strings.TrimRight(message, " \t\n")
	if message == "" {
		return ""
	}
	return message + "\n"
}

// WriteCommitMessageFile 把提交信息整体写入 git 通过 GIT_EDITOR 传入的消息文件。
// 内容不经 shell 解释(不拼接 echo/重定向),多行正文与任意特殊字符原样落盘;
// 空消息不写入,保留 git 生成的原文件内容。
func WriteCommitMessageFile(path, message string) error {
	normalized := normalizeCommitMessage(message)
	if normalized == "" {
		return nil
	}
	return os.WriteFile(path, []byte(normalized), 0o644)
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

// rebaseArgs 构造交互式变基的 git 参数(不含 "git" 自身)。
// injectedMessage 为 true(本工具注入提交信息)时必须关闭默认的
// commit.cleanup=strip:它会删掉以 # 开头的行(Issue 号、模板说明)并压缩空行,
// 与「合并消息保留全部原文」冲突,主题被删空时还会直接提交失败。
func rebaseArgs(baseCommit string, injectedMessage bool) []string {
	args := make([]string, 0, 5)
	if injectedMessage {
		args = append(args, "-c", "commit.cleanup=verbatim")
	}
	args = append(args, "rebase", "-i")
	if baseCommit == "--root" {
		return append(args, "--root")
	}
	return append(args, baseCommit)
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

	injectedMessage := mode == "squash" && newMessage != ""

	if injectedMessage {
		// 提交信息经环境变量交给隐藏子命令写入消息文件:
		// 不生成临时脚本、不经 shell 解释,多行正文不会被折叠
		env = append(env,
			fmt.Sprintf("GIT_EDITOR=%s _internal_commit_message_editor", QuoteForEditorEnv(executable)),
			"EASYGIT_COMMIT_MESSAGE="+newMessage,
		)
	} else if mode == "drop" {
		// drop 不需要编辑器;注入 true 跳过,防止 git 打开默认编辑器
		env = append(env, "GIT_EDITOR=true")
	}

	args := rebaseArgs(baseCommit, injectedMessage)

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
