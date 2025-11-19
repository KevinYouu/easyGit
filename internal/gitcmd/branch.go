package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/fastGit/internal/command"
	"github.com/KevinYouu/fastGit/internal/config"
	"github.com/KevinYouu/fastGit/internal/form"
	"github.com/KevinYouu/fastGit/internal/i18n"
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
		fmt.Println(i18n.T("branch.no.deletable.branches"))
		return nil
	}

	var options []config.Option
	for _, branch := range deletableBranches {
		options = append(options, config.Option{
			Label: fmt.Sprintf("🌱 %s", branch),
			Value: branch,
		})
	}

	_, selectedBranch, err := form.SelectForm(i18n.T("branch.delete.select"), options)
	if err != nil {
		return fmt.Errorf("select branch error: %w", err)
	}

	confirmMessage := fmt.Sprintf(i18n.T("branch.delete.confirm"), selectedBranch)
	if !form.Confirm(confirmMessage) {
		fmt.Println(i18n.T("branch.delete.cancelled"))
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
		commands = append(commands, command.CommandInfo{
			Command:     "git",
			Args:        []string{"push", "origin", "--delete", selectedBranch},
			Description: i18n.T("branch.delete.remote"),
			LoadingMsg:  fmt.Sprintf(i18n.T("branch.delete.remote.loading"), selectedBranch),
			SuccessMsg:  fmt.Sprintf(i18n.T("branch.delete.remote.success"), selectedBranch),
		})
	}

	return command.RunMultipleCommands(commands)
}
