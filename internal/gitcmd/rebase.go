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
