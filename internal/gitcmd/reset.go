package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

type Commit struct {
	Hash      string
	Message   string
	Date      string
	Author    string
	Email     string
	IsHead    bool
	Timestamp time.Time
}

// resetMessageMaxWidth 重置模式选择时提交消息的最大显示宽度
const resetMessageMaxWidth = 40

func Reset() error {
	// 显示开始信息 - 简洁的标题
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Padding(0, 1)

	fmt.Printf("%s\n", headerStyle.Render(i18n.T("reset.title")))

	// 使用更详细的git log格式获取提交历史，包含ISO时间戳用于排序
	cmd := exec.Command("git", "log", "--pretty=format:%h|%s|%ad|%an|%ae|%ai", "--date=format:%m-%d %H:%M")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("error.git.log")+" %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var options = []config.Option{}
	var commits = []Commit{}

	// 解析并存储提交信息（不显示历史记录）
	for i, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			hash := parts[0]
			message := parts[1]
			date := parts[2]
			author := parts[3]
			email := parts[4]

			commit := Commit{
				Hash:    hash,
				Message: message,
				Date:    date,
				Author:  author,
				Email:   email,
				IsHead:  i == 0,
			}

			// 解析时间戳用于排序（如果有的话）
			if len(parts) >= 6 {
				if timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", parts[5]); err == nil {
					commit.Timestamp = timestamp
				}
			}

			// 存储提交信息
			commits = append(commits, commit)

			// 限制消息长度，避免过长(按显示宽度截断,正确处理中文/emoji)
			shortMsg := form.SafeTruncate(message, resetMessageMaxWidth)

			// 添加到选择列表，使用纯文本格式以允许背景色正确覆盖
			commitLabel := ""
			if i == 0 {
				// HEAD提交使用纯文本格式，但添加标记以区分
				commitLabel = fmt.Sprintf(
					"[HEAD] %s %s\n%s • %s",
					hash,
					shortMsg,
					date,
					author,
				)
			} else {
				// 普通提交使用纯文本格式
				commitLabel = fmt.Sprintf(
					"%s %s\n%s • %s",
					hash,
					shortMsg,
					date,
					author,
				)
			}
			options = append(options, config.Option{Label: commitLabel, Value: hash})
		}
	}

	// 使用表格选择表单
	_, choose, err := form.TableSelectForm(options)
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.commit")+" %w", err)
	}

	// 获取选择的提交完整信息
	var selectedCommit Commit
	for _, commit := range commits {
		if commit.Hash == choose {
			selectedCommit = commit
			break
		}
	}

	// 选择重置模式，使用更紧凑的格式 - 纯文本格式以确保背景色能正确覆盖
	resetModes := []config.Option{
		{
			Label: i18n.T("reset.mode.soft.label"),
			Value: "--soft",
		},
		{
			Label: i18n.T("reset.mode.mixed.label"),
			Value: "--mixed",
		},
		{
			Label: i18n.T("reset.mode.hard.label"),
			Value: "--hard",
		},
	}

	// 使用表格选择表单选择重置模式
	_, resetMode, err := form.TableSelectForm(resetModes)
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.mode")+" %w", err)
	}

	// 获取可读的重置模式名称
	resetModeReadable := strings.TrimPrefix(resetMode, "--")

	// 重置模式样式：统一使用 PrimaryColor + Bold
	modeStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)

	// 构建更紧凑的确认信息
	shortMsg := form.SafeTruncate(selectedCommit.Message, resetMessageMaxWidth)

	confirmDesc := fmt.Sprintf("%s %s  %s %s %s%s",
		i18n.T("reset.confirm.to"),
		lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(selectedCommit.Hash),
		shortMsg,
		i18n.T("reset.confirm.mode"),
		modeStyle.Render(resetModeReadable),
		getModeDescription(resetMode),
	)

	// 针对 hard 模式添加警告，带黄色警告图标
	if resetMode == "--hard" {
		confirmDesc += "\n" + theme.WarningIconStyle.Render("⚠") + " " + lipgloss.NewStyle().
			Foreground(theme.PrimaryColor).
			Bold(true).
			Render(i18n.T("reset.hard.warning"))
	}

	// 使用自定义确认表单
	confirm := form.Confirm(confirmDesc)

	if confirm {
		// 执行重置操作
		resetArgs := []string{"reset"}
		if resetMode != "--mixed" { // mixed是默认值，不需要显式指定
			resetArgs = append(resetArgs, resetMode)
		}
		resetArgs = append(resetArgs, choose)

		_, err := command.RunCmdWithSpinnerOptions("git", resetArgs,
			fmt.Sprintf(i18n.T("reset.executing.mode"), resetModeReadable),
			fmt.Sprintf(i18n.T("reset.completed.to"), choose, resetModeReadable), true)

		if err != nil {
			return fmt.Errorf(i18n.T("reset.error.git.reset")+" %w", err)
		}

		// 显示简洁的成功信息
		fmt.Printf("\n%s %s\n",
			theme.SuccessIconStyle.Render("✓"),
			lipgloss.NewStyle().
				Foreground(theme.PrimaryColor).
				Render(fmt.Sprintf(i18n.T("reset.success.prefix"), choose)))

		// 简洁的操作提示
		switch resetMode {
		case "--soft":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.soft")))
		case "--mixed":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.mixed")))
		case "--hard":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.hard")))
		}
	} else {
		fmt.Printf("\n%s %s\n",
			theme.InfoIconStyle.Render("ℹ"),
			theme.InfoStyle.Render(i18n.T("reset.cancelled.msg")))
	}
	return nil
}

// 获取重置模式的简短描述
func getModeDescription(mode string) string {
	switch mode {
	case "--soft":
		return lipgloss.NewStyle().
			Foreground(theme.MutedForeground).
			Render(i18n.T("reset.mode.soft.desc"))
	case "--mixed":
		return lipgloss.NewStyle().
			Foreground(theme.MutedForeground).
			Render(i18n.T("reset.mode.mixed.desc"))
	case "--hard":
		return lipgloss.NewStyle().
			Foreground(theme.MutedForeground).
			Render(i18n.T("reset.mode.hard.desc"))
	default:
		return ""
	}
}
