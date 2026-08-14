package gitcmd

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
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

	selectedHashes, err := form.ListForm(i18n.T("rebase.select.drop_commits"), options, form.ListMulti)
	if err != nil {
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
			theme.SuccessIconStyle.Render("✓"),
			lipgloss.NewStyle().
				Foreground(theme.PrimaryColor).
				Render(i18n.T("drop.success")))
	}
	return err
}
