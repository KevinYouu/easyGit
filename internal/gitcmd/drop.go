package gitcmd

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func Drop() error {
	// 显示开始信息
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Padding(0, 1)

	fmt.Printf("%s\n", headerStyle.Render(i18n.T("drop.title")))

	options, hashes, err := GetRecentCommits()
	if err != nil {
		return err
	}

	selectedHashes, err := form.TableMultiSelectForm(i18n.T("rebase.select.drop_commits"), options)
	if err != nil || len(selectedHashes) == 0 {
		return nil
	}

	confirmed := form.Confirm(fmt.Sprintf(i18n.T("rebase.drop.confirm"), len(selectedHashes)))
	if !confirmed {
		return nil
	}

	// Find the oldest selected commit to determine base
	oldestIndex := -1
	for _, sel := range selectedHashes {
		for i, h := range hashes {
			if sel == h {
				if i > oldestIndex {
					oldestIndex = i
				}
				break
			}
		}
	}

	if oldestIndex == -1 {
		return fmt.Errorf("could not find selected commits in history")
	}

	baseCommit := hashes[oldestIndex]
	parentHash, err := getParentHash(baseCommit)
	if err != nil || parentHash == "" {
		// If no parent (root commit), fallback to the oldest commit itself as base
		parentHash = "--root"
	}

	err = RunInternalRebase(parentHash, "drop", selectedHashes, "")
	if err == nil {
		fmt.Printf("\n%s %s\n",
			theme.SuccessStyle.Render("✓"),
			lipgloss.NewStyle().
				Foreground(theme.SuccessColor).
				Render(i18n.T("drop.success")))
	}
	return err
}
