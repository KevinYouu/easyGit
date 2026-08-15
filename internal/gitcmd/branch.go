package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// GetCurrentBranch retrieves the name of the currently checked out branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetAllBranches retrieves a list of all local branches.
func GetAllBranches() ([]string, error) {
	cmd := exec.Command("git", "branch")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// The current branch is marked with a '*', remove it
			branches = append(branches, strings.TrimPrefix(line, "* "))
		}
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("%s", i18n.T("branch.no.branches"))
	}

	return branches, nil
}

// DeleteBranch prompts the user to select a branch to delete and optionally deletes the remote branch.
func DeleteBranch() error {
	allBranches, err := GetAllBranches()
	if err != nil {
		return fmt.Errorf("get all branches error: %w", err)
	}

	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("get current branch error: %w", err)
	}

	var deletableBranches []string
	for _, branch := range allBranches {
		if branch != currentBranch {
			deletableBranches = append(deletableBranches, branch)
		}
	}

	if len(deletableBranches) == 0 {
		fmt.Printf("%s %s\n", theme.InfoIconStyle.Render("ℹ"), theme.InfoStyle.Render(i18n.T("branch.no.deletable.branches")))
		return nil
	}

	var options []config.Option
	for _, branch := range deletableBranches {
		options = append(options, config.Option{
			Label: branch,
			Value: branch,
		})
	}

	// 自适应多列:分支名自动宽度不截断
	selectedBranches, err := form.ListFormColumns(i18n.T("branch.delete.select"), form.NameDescColumns(), options, form.ListSingle)
	if err != nil {
		return fmt.Errorf("select branch error: %w", err)
	}
	selectedBranch := selectedBranches[0]

	confirmMessage := fmt.Sprintf(i18n.T("branch.delete.confirm"), selectedBranch)
	if !form.Confirm(confirmMessage) {
		fmt.Printf("%s %s\n", theme.InfoIconStyle.Render("ℹ"), theme.InfoStyle.Render(i18n.T("branch.delete.cancelled")))
		return nil
	}

	deleteRemote := form.Confirm(i18n.T("branch.delete.remote.confirm"))

	var commands []command.CommandInfo
	commands = append(commands, command.CommandInfo{
		Command:     "git",
		Args:        []string{"branch", "-d", selectedBranch},
		Description: i18n.T("branch.delete.local"),
		LoadingMsg:  fmt.Sprintf(i18n.T("branch.delete.local.loading"), selectedBranch),
		SuccessMsg:  fmt.Sprintf(i18n.T("branch.delete.local.success"), selectedBranch),
	})

	if deleteRemote {
		// 选择远程仓库(支持配置持久化和多选,输出保存/使用日志)
		selectedRemotes, err := SelectAndSaveRemotes()
		if err != nil {
			return fmt.Errorf("select remote: %w", err)
		}

		// 添加每个远程的删除推送命令(并行段)
		for _, remote := range selectedRemotes {
			commands = append(commands, command.CommandInfo{
				Command:     "git",
				Args:        []string{"push", remote, "--delete", selectedBranch},
				Description: fmt.Sprintf(i18n.T("git.push.to.remote"), remote),
				LoadingMsg:  fmt.Sprintf(i18n.T("git.push.loading.remote"), remote),
				SuccessMsg:  fmt.Sprintf(i18n.T("branch.delete.remote.success"), selectedBranch),
			})
		}

		// 本地删除串行,随后所有远程删除推送一次性并行启动
		return command.RunMultipleCommandsParallel(commands, 1)
	}

	return command.RunMultipleCommands(commands)
}
