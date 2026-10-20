package gitcmd

import (
	"fmt"
	"slices"
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

	// selectedHashes 为从旧到新的所选提交,供默认消息拼装与变基计划匹配
	selectedHashes = orderOldestFirst(selectedHashes, hashes)
	baseCommit := selectedHashes[0]
	parentHash, err := getParentHash(baseCommit)
	if err != nil || parentHash == "" {
		parentHash = "--root"
	}

	// Get new message:默认值完整保留所选提交的全部文本(主题 + 正文),
	// 顺序从旧到新、空行分隔,与 git 原生 squash 的拼接结果一致;
	// 只用最新提交的主题会丢掉所有正文与换行。
	defaultMessage, err := buildSquashDefaultMessage(selectedHashes)
	if err != nil {
		return err
	}

	newMessage, err := form.TextInput(i18n.T("squash.input.message"), defaultMessage)
	if err != nil || strings.TrimSpace(newMessage) == "" {
		return nil
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

// orderOldestFirst 按 hashes(新→旧)的下标把所选提交重排为旧→新,
// 保证拼装的默认消息与变基计划的提交顺序一致。
func orderOldestFirst(selected, hashes []string) []string {
	indexOf := make(map[string]int, len(hashes))
	for i, h := range hashes {
		indexOf[h] = i
	}
	ordered := slices.Clone(selected)
	// 下标越大越旧
	slices.SortStableFunc(ordered, func(a, b string) int {
		return indexOf[b] - indexOf[a]
	})
	return ordered
}

// buildSquashDefaultMessage 拼装合并后提交信息的默认值:
// 按旧→新依次写入每条提交的完整消息(主题 + 正文 + 段间空行),
// 与 git 原生 squash 的拼接一致,换行不得被折叠。
func buildSquashDefaultMessage(oldestFirst []string) (string, error) {
	bodies, err := GetCommitMessageBodies(oldestFirst)
	if err != nil {
		return "", err
	}
	blocks := make([]string, 0, len(bodies))
	for _, body := range bodies {
		block := strings.TrimRight(strings.ReplaceAll(body, "\r\n", "\n"), " \t\n")
		if block != "" {
			blocks = append(blocks, block)
		}
	}
	return strings.Join(blocks, "\n\n"), nil
}
