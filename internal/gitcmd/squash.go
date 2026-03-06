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
	"github.com/charmbracelet/lipgloss"
)

func Squash() error {
	// 显示开始信息
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Padding(0, 1)

	fmt.Printf("%s\n", headerStyle.Render(i18n.T("squash.title")))

	// 获取提交历史
	cmd := exec.Command("git", "log", "--pretty=format:%h|%s|%ad|%an", "--date=format:%m-%d %H:%M")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("error.git.log")+" %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("not enough commits to squash")
	}

	var options = []config.Option{}
	var commits = []Commit{}

	for i, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			hash := parts[0]
			message := parts[1]
			date := parts[2]
			author := parts[3]

			commit := Commit{
				Hash:    hash,
				Message: message,
				Date:    date,
				Author:  author,
				IsHead:  i == 0,
			}
			commits = append(commits, commit)

			// 限制消息长度
			shortMsg := message
			if len(shortMsg) > 40 {
				shortMsg = shortMsg[:37] + "..."
			}

			prefix := ""
			if i == 0 {
				prefix = "[HEAD] "
			}

			commitLabel := fmt.Sprintf(
				"%s%s %s\n%s • %s",
				prefix,
				hash,
				shortMsg,
				date,
				author,
			)
			options = append(options, config.Option{Label: commitLabel, Value: hash})
		}
	}

	// 选择要合并的最早提交
	fmt.Printf("\n%s\n", lipgloss.NewStyle().Foreground(theme.TextSecondary).Render(i18n.T("squash.select.base")))
	_, choose, err := form.TableSelectForm(options)
	if err != nil {
		return fmt.Errorf(i18n.T("squash.error.select")+" %w", err)
	}

	// 如果没有选择 (例如用户取消了操作)
	if choose == "" {
		fmt.Printf("\n%s %s\n",
			theme.InfoStyle.Render("ℹ️"),
			theme.InfoStyle.Render(i18n.T("squash.cancelled")))
		return nil
	}

	// 获取默认合并信息 (使用被选中的最早提交的信息或 HEAD 信息)
	headMsg := commits[0].Message

	// 输入新的提交信息
	newMsg, err := form.Input(i18n.T("squash.input.message"), headMsg)
	if err != nil || newMsg == "" {
		if err != nil {
			return nil
		}
	}

	// 获取父提交哈希
	parentHash, err := getParentHash(choose)
	if err != nil {
		return fmt.Errorf("failed to get parent commit: %w", err)
	}

	if parentHash != "" {
		// 正常合并: 重置到父提交
		_, err = command.RunCmdWithSpinnerOptions("git", []string{"reset", "--soft", parentHash},
			"Resetting to "+parentHash+"...", "Reset completed", true)
		if err != nil {
			return fmt.Errorf(i18n.T("squash.error.git.reset")+" %w", err)
		}

		// 执行提交
		_, err = command.RunCmdWithSpinnerOptions("git", []string{"commit", "-m", newMsg},
			"Creating squashed commit...", "Commit created", true)
		if err != nil {
			return fmt.Errorf(i18n.T("squash.error.git.commit")+" %w", err)
		}
	} else {
		// 特殊处理: 选择了初始提交 (没有父提交)
		// 1. 重置到该提交 (保留其后的变更在暂存区)
		_, err = command.RunCmdWithSpinnerOptions("git", []string{"reset", "--soft", choose},
			"Resetting to root "+choose+"...", "Reset completed", true)
		if err != nil {
			return fmt.Errorf(i18n.T("squash.error.git.reset")+" %w", err)
		}

		// 2. 使用 --amend 合并暂存区的变更到初始提交中
		_, err = command.RunCmdWithSpinnerOptions("git", []string{"commit", "--amend", "-m", newMsg},
			"Amending root commit...", "Root commit updated", true)
		if err != nil {
			return fmt.Errorf(i18n.T("squash.error.git.commit")+" %w", err)
		}
	}

	// 输出成功信息
	fmt.Printf("\n%s %s\n",
		theme.SuccessStyle.Render("✓"),
		lipgloss.NewStyle().
			Foreground(theme.SuccessColor).
			Render(i18n.T("squash.success")))

	return nil
}

// 获取指定提交的父提交哈希
func getParentHash(hash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%P", hash)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	parents := strings.Fields(string(output))
	if len(parents) == 0 {
		return "", nil // 无父提交 (Root)
	}
	return parents[0], nil // 返回第一个父提交
}
