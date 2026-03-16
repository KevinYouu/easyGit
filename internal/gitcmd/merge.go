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

// MergeStrategy represents different merge strategies
type MergeStrategy struct {
	Name           string
	Args           []string
	NameKey        string // i18n key for strategy name
	DescriptionKey string // i18n key for strategy description
}

var mergeStrategies = []MergeStrategy{
	{
		Name:           "default",
		Args:           []string{},
		NameKey:        "merge.strategy.default.name",
		DescriptionKey: "merge.strategy.default.description",
	},
	{
		Name:           "ff-only",
		Args:           []string{"--ff-only"},
		NameKey:        "merge.strategy.ff.only.name",
		DescriptionKey: "merge.strategy.ff.only.description",
	},
	{
		Name:           "no-ff",
		Args:           []string{"--no-ff"},
		NameKey:        "merge.strategy.no.ff.name",
		DescriptionKey: "merge.strategy.no.ff.description",
	},
	{
		Name:           "squash",
		Args:           []string{"--squash"},
		NameKey:        "merge.strategy.squash.name",
		DescriptionKey: "merge.strategy.squash.description",
	},
}

func MergeIntoCurrent() error {
	// Check if working directory is clean
	if err := checkWorkingDirectoryStatus(); err != nil {
		return fmt.Errorf("working directory check failed: %w", err)
	}

	branches, err := getAllAvailableBranches()
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	if len(branches) == 0 {
		logs.Info(i18n.T("merge.no.branches"))
		return nil
	}

	selectedBranch, err := selectBranchToMerge(branches)
	if err != nil {
		return fmt.Errorf("branch selection failed: %w", err)
	}

	strategy, err := selectMergeStrategy()
	if err != nil {
		return fmt.Errorf("strategy selection failed: %w", err)
	}

	return performMerge(selectedBranch, strategy)
}

// checkWorkingDirectoryStatus checks if the working directory is clean
func checkWorkingDirectoryStatus() error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check git status: %v", err)
	}

	if len(strings.TrimSpace(string(output))) > 0 {
		logs.Warning(i18n.T("merge.warning.dirty.working.directory"))
		confirmed := form.Confirm(i18n.T("merge.confirm.continue.with.changes"))
		if !confirmed {
			return fmt.Errorf("merge cancelled by user")
		}
	}
	return nil
}

// getAllAvailableBranches gets both local and remote branches
func getAllAvailableBranches() ([]config.Option, error) {
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get local branches
	localBranches, err := getLocalBranches(currentBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get local branches: %w", err)
	}

	// Get remote branches
	remoteBranches, err := getRemoteBranches(currentBranch)
	if err != nil {
		logs.Warning("Failed to get remote branches: " + err.Error())
		remoteBranches = []config.Option{} // Continue with local branches only
	}

	// Combine all branches
	allBranches := append(localBranches, remoteBranches...)

	if len(allBranches) == 0 {
		return nil, fmt.Errorf("no branches available for merge")
	}

	return allBranches, nil
}

// getCurrentBranch gets the current branch name
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error getting current branch: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getLocalBranches gets all local branches except current
func getLocalBranches(currentBranch string) ([]config.Option, error) {
	cmd := exec.Command("git", "branch")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting local branches: %v", err)
	}

	var branches []config.Option
	lines := strings.SplitSeq(string(output), "\n")

	for line := range lines {
		branch := strings.TrimSpace(strings.TrimPrefix(line, "* "))
		if branch != "" && branch != "(no branch)" && branch != currentBranch {
			branches = append(branches, config.Option{
				Label: fmt.Sprintf("📍 %s (local)", branch),
				Value: branch,
			})
		}
	}

	return branches, nil
}

// getRemoteBranches gets all remote branches except current
func getRemoteBranches(currentBranch string) ([]config.Option, error) {
	cmd := exec.Command("git", "branch", "-r")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting remote branches: %v", err)
	}

	var branches []config.Option
	lines := strings.SplitSeq(string(output), "\n")

	for line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" && !strings.Contains(branch, "HEAD ->") {
			// Extract branch name from origin/branch-name format
			parts := strings.Split(branch, "/")
			if len(parts) >= 2 {
				branchName := strings.Join(parts[1:], "/")
				if branchName != currentBranch {
					branches = append(branches, config.Option{
						Label: fmt.Sprintf("🌐 %s (remote)", branch),
						Value: branch,
					})
				}
			}
		}
	}

	return branches, nil
}

// selectBranchToMerge shows a selection form for available branches
func selectBranchToMerge(branches []config.Option) (string, error) {
	_, selectedBranch, err := form.SelectForm(i18n.T("merge.select.target"), branches)
	if err != nil {
		return "", fmt.Errorf(i18n.T("error.select.form.detail")+": %w", err)
	}
	return selectedBranch, nil
}

// selectMergeStrategy allows user to choose merge strategy
func selectMergeStrategy() (MergeStrategy, error) {
	var strategies []config.Option
	for _, strategy := range mergeStrategies {
		strategies = append(strategies, config.Option{
			Label: fmt.Sprintf("%s - %s", i18n.T(strategy.NameKey), i18n.T(strategy.DescriptionKey)),
			Value: strategy.Name,
		})
	}

	_, selectedStrategyName, err := form.SelectForm(i18n.T("merge.select.strategy"), strategies)
	if err != nil {
		return MergeStrategy{}, fmt.Errorf(i18n.T("error.select.form.detail")+": %w", err)
	}

	// Find the selected strategy
	for _, strategy := range mergeStrategies {
		if strategy.Name == selectedStrategyName {
			return strategy, nil
		}
	}

	return mergeStrategies[0], nil // Default strategy as fallback
}

// performMerge executes the merge with the selected strategy
func performMerge(branch string, strategy MergeStrategy) error {
	args := append([]string{"merge"}, strategy.Args...)
	args = append(args, branch)

	logs.Info(fmt.Sprintf(i18n.T("merge.starting"), branch, i18n.T(strategy.NameKey)))

	output, err := command.RunCmd("git", args, i18n.T("merge.success.message"))
	if err != nil {
		return handleMergeError(output, err)
	}

	return nil
}

// handleMergeError provides detailed error handling for merge failures
func handleMergeError(output string, err error) error {
	outputStr := strings.TrimSpace(output)

	// Check for common merge issues
	if strings.Contains(outputStr, "CONFLICT") {
		logs.Error(i18n.T("merge.conflict.detected"))
		logs.Info(i18n.T("merge.conflict.instructions"))
		return fmt.Errorf("merge conflict detected: use 'git status' to see conflicted files")
	}

	if strings.Contains(outputStr, "not possible to fast-forward") {
		logs.Error(i18n.T("merge.fast.forward.failed"))
		logs.Info(i18n.T("merge.fast.forward.suggestion"))
		return fmt.Errorf("fast-forward merge not possible")
	}

	if strings.Contains(outputStr, "uncommitted changes") {
		logs.Error(i18n.T("merge.uncommitted.changes"))
		return fmt.Errorf("uncommitted changes prevent merge")
	}

	// Generic error
	logs.Error(i18n.T("merge.failed") + ": " + outputStr)
	return fmt.Errorf("git merge failed: %s", outputStr)
}
