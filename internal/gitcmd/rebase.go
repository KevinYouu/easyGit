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
		{Label: i18n.T("rebase.action.continue"), Value: "--continue"},
		{Label: i18n.T("rebase.action.skip"), Value: "--skip"},
		{Label: i18n.T("rebase.action.abort"), Value: "--abort"},
	}

	_, action, err := form.SelectForm(i18n.T("rebase.status.in_progress"), options)
	if err != nil {
		return err
	}

	output, err := command.RunCmd("git", []string{"rebase", action}, "")
	if err != nil {
		return handleRebaseError(output, err)
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

	_, selectedBranch, err := form.SelectForm(i18n.T("rebase.select.target"), branches)
	if err != nil {
		return fmt.Errorf(i18n.T("error.select.form.detail")+": %w", err)
	}

	logs.Info(fmt.Sprintf(i18n.T("rebase.starting"), selectedBranch))
	output, err := command.RunCmd("git", []string{"rebase", selectedBranch}, "")
	if err != nil {
		return handleRebaseError(output, err)
	}
	
	logs.Info(i18n.T("rebase.success.message"))
	return nil
}

func GetRecentCommits() ([]config.Option, []string, error) {
	cmd := exec.Command("git", "log", "-n", "50", "--pretty=format:%h|%s|%ad|%an", "--date=format:%m-%d %H:%M")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf(i18n.T("error.git.log")+" %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("not enough commits to rebase")
	}

	var options = []config.Option{}
	var hashes = []string{}
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			hash := parts[0]
			message := parts[1]
			date := parts[2]
			author := parts[3]

			shortMsg := message
			if len(shortMsg) > 50 {
				shortMsg = shortMsg[:47] + "..."
			}

			commitLabel := fmt.Sprintf(
				"%s %s\n%s • %s",
				hash,
				shortMsg,
				date,
				author,
			)
			options = append(options, config.Option{Label: commitLabel, Value: hash})
			hashes = append(hashes, hash)
		}
	}
	return options, hashes, nil
}

func RunInternalRebase(baseCommit, mode string, targets []string, newMessage string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Pass config via env var as JSON
	configData := map[string]interface{}{
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
		// Try to run git status to check if we're in conflict
		if isRebaseInProgress() {
			logs.Error(i18n.T("rebase.conflict.detected"))
			logs.Info(i18n.T("rebase.conflict.instructions"))
			return fmt.Errorf("rebase conflict")
		}
		return fmt.Errorf("rebase failed: %w", err)
	}

	logs.Info(i18n.T("rebase.success.message"))
	return nil
}

func handleRebaseError(output string, err error) error {
	outputStr := strings.TrimSpace(output)

	if strings.Contains(outputStr, "CONFLICT") {
		logs.Error(i18n.T("rebase.conflict.detected"))
		logs.Info(i18n.T("rebase.conflict.instructions"))
		return fmt.Errorf("rebase conflict detected: use 'git status' to see conflicted files")
	}

	logs.Error(i18n.T("rebase.failed") + ": " + outputStr)
	return fmt.Errorf("git rebase failed: %s", outputStr)
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
