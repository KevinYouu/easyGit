package gitcmd

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/theme"
)

func Squash() error {
	// 显示开始信息
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Padding(0, 1)

	fmt.Printf("%s\n", headerStyle.Render(i18n.T("squash.title")))

	options, hashes, err := GetRecentCommits()
	if err != nil {
		return err
	}

	selectedHashes, err := form.ListForm(i18n.T("rebase.select.squash_commits"), options, form.ListMulti)
	if err != nil {
		return nil
	}

	if len(selectedHashes) < 2 {
		logs.Error(i18n.T("squash.select.min.two"))
		return nil
	}

	// Check if selected commits are contiguous
	var indices []int
	for _, sel := range selectedHashes {
		for i, h := range hashes {
			if sel == h {
				indices = append(indices, i)
				break
			}
		}
	}
	sort.Ints(indices)

	for i := 0; i < len(indices)-1; i++ {
		if indices[i+1]-indices[i] != 1 {
			logs.Error(i18n.T("rebase.squash.not_contiguous"))
			return nil
		}
	}

	// Get new message
	defaultMessage := ""
	// Try to get message from the oldest selected
	oldestHash := hashes[indices[len(indices)-1]]
	for _, opt := range options {
		if opt.Value == oldestHash {
			parts := strings.SplitN(opt.Label, "\n", 2)
			if len(parts) > 0 {
				msgParts := strings.SplitN(parts[0], " ", 2)
				if len(msgParts) > 1 {
					defaultMessage = msgParts[1]
				}
			}
			break
		}
	}

	newMessage, err := form.Input(i18n.T("squash.input.message"), defaultMessage)
	if err != nil || newMessage == "" {
		return nil
	}

	baseCommit := oldestHash
	parentHash, err := getParentHash(baseCommit)
	if err != nil || parentHash == "" {
		parentHash = "--root"
	}

	err = RunInternalRebase(parentHash, "squash", selectedHashes, newMessage)
	if err == nil {
		fmt.Printf("\n%s %s\n",
			theme.SuccessIconStyle.Render("✓"),
			lipgloss.NewStyle().
				Foreground(theme.PrimaryColor).
				Render(i18n.T("squash.success")))
	}
	return err
}
